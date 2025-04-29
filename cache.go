package sdl

import (
	"math/rand"
	"time"
	// "log"
)

// Cache represents a caching component (e.g., in-memory, Redis).
// It models hit/miss probabilities and associated latencies.
type Cache struct {
	// --- Configurable Parameters ---

	// HitRate is the probability (0.0 to 1.0) that a Read finds the requested item.
	HitRate float64

	// Latency for a successful cache hit.
	HitLatency *Outcomes[Duration]

	// Latency incurred just to determine a cache miss (e.g., key lookup fails).
	MissLatency *Outcomes[Duration]

	// Latency for writing data TO the cache itself.
	WriteLatency *Outcomes[Duration]

	// Probability that a cache operation itself fails (e.g., cache unavailable).
	// Applied to Hits, Misses, and Writes.
	FailureProb float64
	// Latency associated with a cache operation failure.
	FailureLatency *Outcomes[Duration]

	// --- Internal ---
	// Pre-calculated outcomes for efficiency
	readOutcomes  *Outcomes[AccessResult]
	writeOutcomes *Outcomes[AccessResult]
	rng           *rand.Rand
}

// Init initializes the Cache component with provided parameters or defaults.
func (c *Cache) Init() *Cache {
	// --- Set Defaults ---
	c.HitRate = 0.80 // Default: 80% hit rate

	// Default: Fast hit (e.g., local in-memory) with slight variance
	c.HitLatency = c.HitLatency.Add(90, Nanos(100)).Add(10, Nanos(500))
	c.HitLatency.And = func(a, b Duration) Duration { return a + b } // Needed if composed

	// Default: Miss detection is slightly slower than hit
	c.MissLatency = c.MissLatency.Add(100, Nanos(200))
	c.MissLatency.And = func(a, b Duration) Duration { return a + b }

	// Default: Write is similar to hit latency
	c.WriteLatency = c.WriteLatency.Add(90, Nanos(120)).Add(10, Nanos(600))
	c.WriteLatency.And = func(a, b Duration) Duration { return a + b }

	// Default: Cache is generally reliable
	c.FailureProb = 0.001 // 0.1% failure chance
	// Default: Failure is detected relatively quickly
	c.FailureLatency = c.FailureLatency.Add(100, Millis(1))
	c.FailureLatency.And = func(a, b Duration) Duration { return a + b }

	// Initialize RNG
	c.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Pre-calculate outcomes
	c.calculateReadOutcomes()
	c.calculateWriteOutcomes()

	return c
}

// NewCache creates and initializes a new Cache component with defaults.
// Configuration methods can be added later (e.g., ConfigureHitRate(rate float64)).
func NewCache() *Cache {
	c := &Cache{}
	return c.Init()
}

// calculateReadOutcomes generates the probabilistic outcomes for a cache read attempt.
func (c *Cache) calculateReadOutcomes() {
	outcomes := &Outcomes[AccessResult]{And: AndAccessResults}
	totalProb := 1.0

	// --- Failures First ---
	failProb := c.FailureProb
	if failProb > 1e-9 {
		// Add outcomes for cache failure, combining FailureLatency outcomes
		for _, failBucket := range c.FailureLatency.Buckets {
			prob := failProb * (failBucket.Weight / c.FailureLatency.TotalWeight())
			if prob > 1e-9 {
				outcomes.Add(prob, AccessResult{Success: false, Latency: failBucket.Value})
			}
		}
		totalProb -= failProb
	}

	// --- Hits ---
	// Effective Hit Rate = configured HitRate * (1 - FailureProb)
	hitProb := c.HitRate * totalProb
	if hitProb > 1e-9 {
		// Add outcomes for cache hit, combining HitLatency outcomes
		for _, hitBucket := range c.HitLatency.Buckets {
			prob := hitProb * (hitBucket.Weight / c.HitLatency.TotalWeight())
			if prob > 1e-9 {
				// Success = true indicates a HIT
				outcomes.Add(prob, AccessResult{Success: true, Latency: hitBucket.Value})
			}
		}
	}

	// --- Misses ---
	// Effective Miss Rate = (1 - configured HitRate) * (1 - FailureProb)
	missProb := (1.0 - c.HitRate) * totalProb
	if missProb > 1e-9 {
		// Add outcomes for cache miss, combining MissLatency outcomes
		for _, missBucket := range c.MissLatency.Buckets {
			prob := missProb * (missBucket.Weight / c.MissLatency.TotalWeight())
			if prob > 1e-9 {
				// Success = false indicates a MISS (distinguished from failure by latency/caller logic)
				outcomes.Add(prob, AccessResult{Success: false, Latency: missBucket.Value})
			}
		}
	}

	c.readOutcomes = outcomes
	// Optional: Reduce if needed, but likely small number of buckets here
}

// calculateWriteOutcomes generates the probabilistic outcomes for a cache write attempt.
func (c *Cache) calculateWriteOutcomes() {
	outcomes := &Outcomes[AccessResult]{And: AndAccessResults}
	totalProb := 1.0

	// --- Failures First ---
	failProb := c.FailureProb
	if failProb > 1e-9 {
		for _, failBucket := range c.FailureLatency.Buckets {
			prob := failProb * (failBucket.Weight / c.FailureLatency.TotalWeight())
			if prob > 1e-9 {
				outcomes.Add(prob, AccessResult{Success: false, Latency: failBucket.Value})
			}
		}
		totalProb -= failProb
	}

	// --- Success ---
	// Effective Success Rate = 1.0 - FailureProb
	successProb := totalProb
	if successProb > 1e-9 {
		for _, writeBucket := range c.WriteLatency.Buckets {
			prob := successProb * (writeBucket.Weight / c.WriteLatency.TotalWeight())
			if prob > 1e-9 {
				outcomes.Add(prob, AccessResult{Success: true, Latency: writeBucket.Value})
			}
		}
	}

	c.writeOutcomes = outcomes
	// Optional: Reduce if needed
}

// Read simulates attempting to read from the cache.
// Returns outcomes:
// - Success=true: Cache Hit, Latency=HitLatency distribution.
// - Success=false: Cache Miss OR Cache Failure. Caller distinguishes based on context/latency.
// The returned Outcomes should generally not be modified directly.
func (c *Cache) Read() *Outcomes[AccessResult] {
	if c.readOutcomes == nil {
		c.calculateReadOutcomes()
	} // Defensive
	return c.readOutcomes
}

// Write simulates attempting to write to the cache.
// Returns outcomes representing the success/latency of the cache write itself.
// The returned Outcomes should generally not be modified directly.
func (c *Cache) Write() *Outcomes[AccessResult] {
	if c.writeOutcomes == nil {
		c.calculateWriteOutcomes()
	} // Defensive
	return c.writeOutcomes
}
