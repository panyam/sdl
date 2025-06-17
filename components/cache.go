package components

import (
	sc "github.com/panyam/sdl/core"
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
	readOutcomes  *Outcomes[sc.AccessResult]
	writeOutcomes *Outcomes[sc.AccessResult]
}

// Init initializes the Cache component with provided parameters or defaults.
func (c *Cache) Init() {
	// --- Set Defaults ---
	c.HitRate = 0.80 // Default: 80% hit rate

	// Default: Fast hit (e.g., local in-memory) with slight variance
	if c.HitLatency == nil {
		c.HitLatency = c.HitLatency.Add(90, Nanos(100)).Add(10, Nanos(500))
	}
	if c.HitLatency.And == nil {
		c.HitLatency.And = func(a, b Duration) Duration { return a + b } // Needed if composed
	}

	// Default: Miss detection is slightly slower than hit
	if c.MissLatency == nil {
		c.MissLatency = c.MissLatency.Add(100, Nanos(200))
	}
	if c.MissLatency.And == nil {
		c.MissLatency.And = func(a, b Duration) Duration { return a + b }
	}

	// Default: Write is similar to hit latency
	if c.WriteLatency == nil {
		c.WriteLatency = c.WriteLatency.Add(90, Nanos(120)).Add(10, Nanos(600))
	}
	if c.WriteLatency.And == nil {
		c.WriteLatency.And = func(a, b Duration) Duration { return a + b }
	}

	// Default: Cache is generally reliable
	c.FailureProb = 0.001 // 0.1% failure chance
	// Default: Failure is detected relatively quickly
	if c.FailureLatency == nil {
		c.FailureLatency = c.FailureLatency.Add(100, Millis(1))
	}
	if c.FailureLatency.And == nil {
		c.FailureLatency.And = func(a, b Duration) Duration { return a + b }
	}

	// Pre-calculate outcomes
	c.calculateReadOutcomes()
	c.calculateWriteOutcomes()
}

// NewCache creates and initializes a new Cache component with defaults.
// Configuration methods can be added later (e.g., ConfigureHitRate(rate float64)).
func NewCache() *Cache {
	c := &Cache{}
	c.Init()
	return c
}

// calculateReadOutcomes generates the probabilistic outcomes for a cache read attempt.
func (c *Cache) calculateReadOutcomes() {
	outcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
	totalProb := 1.0

	// --- Failures First ---
	failProb := c.FailureProb
	if failProb > 1e-9 {
		// Add outcomes for cache failure, combining FailureLatency outcomes
		for _, failBucket := range c.FailureLatency.Buckets {
			prob := failProb * (failBucket.Weight / c.FailureLatency.TotalWeight())
			if prob > 1e-9 {
				outcomes.Add(prob, sc.AccessResult{Success: false, Latency: failBucket.Value})
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
				outcomes.Add(prob, sc.AccessResult{Success: true, Latency: hitBucket.Value})
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
				outcomes.Add(prob, sc.AccessResult{Success: false, Latency: missBucket.Value})
			}
		}
	}

	c.readOutcomes = outcomes
	// Optional: Reduce if needed, but likely small number of buckets here
}

// calculateWriteOutcomes generates the probabilistic outcomes for a cache write attempt.
func (c *Cache) calculateWriteOutcomes() {
	outcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
	totalProb := 1.0

	// --- Failures First ---
	failProb := c.FailureProb
	if failProb > 1e-9 {
		for _, failBucket := range c.FailureLatency.Buckets {
			prob := failProb * (failBucket.Weight / c.FailureLatency.TotalWeight())
			if prob > 1e-9 {
				outcomes.Add(prob, sc.AccessResult{Success: false, Latency: failBucket.Value})
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
				outcomes.Add(prob, sc.AccessResult{Success: true, Latency: writeBucket.Value})
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
func (c *Cache) Read() *Outcomes[sc.AccessResult] {
	if c.readOutcomes == nil {
		c.calculateReadOutcomes()
	} // Defensive
	return c.readOutcomes
}

// Write simulates attempting to write to the cache.
// Returns outcomes representing the success/latency of the cache write itself.
// The returned Outcomes should generally not be modified directly.
func (c *Cache) Write() *Outcomes[sc.AccessResult] {
	if c.writeOutcomes == nil {
		c.calculateWriteOutcomes()
	} // Defensive
	return c.writeOutcomes
}

// GetFlowPattern implements FlowAnalyzable interface for Cache
func (c *Cache) GetFlowPattern(methodName string, inputRate float64) FlowPattern {
	switch methodName {
	case "Read":
		// Cache reads are leaf operations with no downstream calls
		// Service time based on hit rate (hits are fast, misses are slower)
		avgServiceTime := c.HitRate*0.0001 + (1-c.HitRate)*0.0002 // 0.1ms hit, 0.2ms miss

		return FlowPattern{
			Outflows:      map[string]float64{}, // No downstream calls
			SuccessRate:   c.HitRate,            // Success = hit rate
			Amplification: 1.0,                  // Input rate = output rate
			ServiceTime:   avgServiceTime,
		}

	case "Write":
		// Cache writes are also leaf operations
		return FlowPattern{
			Outflows:      map[string]float64{}, // No downstream calls
			SuccessRate:   1.0 - c.FailureProb,  // Success based on failure probability
			Amplification: 1.0,
			ServiceTime:   0.00012, // ~0.12ms for cache write
		}
	}

	// Unknown method
	return FlowPattern{
		Outflows:      map[string]float64{},
		SuccessRate:   1.0,
		Amplification: 1.0,
		ServiceTime:   0.001,
	}
}
