package components

import (
	"math"

	sc "github.com/panyam/sdl/core"
)

// CacheWithContention represents a cache with limited throughput capacity.
// Unlike the basic Cache which assumes infinite bandwidth, this models:
// - Maximum requests per second the cache can handle
// - Queueing delays when the cache is under heavy load
// - Performance degradation at high utilization
type CacheWithContention struct {
	// Cache behavior parameters
	HitRate        float64                // Probability of cache hit (0.0 to 1.0)
	HitLatency     *Outcomes[Duration]    // Base latency for hits
	MissLatency    *Outcomes[Duration]    // Base latency for misses
	WriteLatency   *Outcomes[Duration]    // Base latency for writes
	FailureProb    float64                // Probability of operation failure
	FailureLatency *Outcomes[Duration]    // Latency for failures

	// Contention modeling parameters (M/M/1 queue)
	MaxThroughput float64 // Maximum requests per second
	arrivalRate   float64 // Current arrival rate (λ)

	// Pre-calculated outcomes
	readOutcomes  *Outcomes[sc.AccessResult]
	writeOutcomes *Outcomes[sc.AccessResult]
}

// Init initializes the CacheWithContention with defaults
func (c *CacheWithContention) Init() {
	// Cache behavior defaults
	c.HitRate = 0.80      // 80% hit rate
	c.FailureProb = 0.001 // 0.1% failure rate

	// Throughput defaults - typical for a Redis-like cache
	if c.MaxThroughput == 0 {
		c.MaxThroughput = 100000 // 100K ops/sec
	}

	// Latency defaults (base latencies without contention)
	if c.HitLatency == nil {
		c.HitLatency = c.HitLatency.Add(90, Nanos(100)).Add(10, Nanos(500))
	}
	if c.HitLatency.And == nil {
		c.HitLatency.And = func(a, b Duration) Duration { return a + b }
	}

	if c.MissLatency == nil {
		c.MissLatency = c.MissLatency.Add(100, Nanos(200))
	}
	if c.MissLatency.And == nil {
		c.MissLatency.And = func(a, b Duration) Duration { return a + b }
	}

	if c.WriteLatency == nil {
		c.WriteLatency = c.WriteLatency.Add(90, Nanos(120)).Add(10, Nanos(600))
	}
	if c.WriteLatency.And == nil {
		c.WriteLatency.And = func(a, b Duration) Duration { return a + b }
	}

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

// NewCacheWithContention creates a new cache with contention modeling
func NewCacheWithContention(name string) *CacheWithContention {
	c := &CacheWithContention{}
	c.Init()
	return c
}

// SetArrivalRate sets the arrival rate for capacity planning
func (c *CacheWithContention) SetArrivalRate(method string, rate float64) error {
	c.arrivalRate = rate
	// Recalculate outcomes with new arrival rate
	c.calculateReadOutcomes()
	c.calculateWriteOutcomes()
	return nil
}

// GetArrivalRate returns the current arrival rate
func (c *CacheWithContention) GetArrivalRate(method string) float64 {
	return c.arrivalRate
}

// GetTotalArrivalRate returns the total arrival rate
func (c *CacheWithContention) GetTotalArrivalRate() float64 {
	return c.arrivalRate
}

// calculateQueueingDelay calculates M/M/1 queueing delay based on utilization
func (c *CacheWithContention) calculateQueueingDelay() float64 {
	if c.MaxThroughput <= 0 || c.arrivalRate <= 0 {
		return 0
	}

	utilization := c.arrivalRate / c.MaxThroughput
	if utilization >= 1.0 {
		// System is overloaded
		return 3600.0 // 1 hour as "infinity"
	}

	// M/M/1 average queueing time: Wq = ρ / (μ * (1 - ρ))
	// where ρ = λ/μ (utilization), μ = service rate
	avgServiceTime := 1.0 / c.MaxThroughput
	avgQueueTime := (utilization * avgServiceTime) / (1.0 - utilization)

	return avgQueueTime
}

// calculateReadOutcomes generates probabilistic outcomes including contention
func (c *CacheWithContention) calculateReadOutcomes() {
	outcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
	totalProb := 1.0
	queueDelay := c.calculateQueueingDelay()

	// Failures first
	failProb := c.FailureProb
	if failProb > 1e-9 {
		for _, failBucket := range c.FailureLatency.Buckets {
			prob := failProb * (failBucket.Weight / c.FailureLatency.TotalWeight())
			if prob > 1e-9 {
				// Add queueing delay to failure latency
				totalLatency := failBucket.Value + queueDelay
				outcomes.Add(prob, sc.AccessResult{Success: false, Latency: totalLatency})
			}
		}
		totalProb -= failProb
	}

	// Hits
	hitProb := c.HitRate * totalProb
	if hitProb > 1e-9 {
		for _, hitBucket := range c.HitLatency.Buckets {
			prob := hitProb * (hitBucket.Weight / c.HitLatency.TotalWeight())
			if prob > 1e-9 {
				// Add queueing delay to hit latency
				totalLatency := hitBucket.Value + queueDelay
				outcomes.Add(prob, sc.AccessResult{Success: true, Latency: totalLatency})
			}
		}
	}

	// Misses
	missProb := (1.0 - c.HitRate) * totalProb
	if missProb > 1e-9 {
		for _, missBucket := range c.MissLatency.Buckets {
			prob := missProb * (missBucket.Weight / c.MissLatency.TotalWeight())
			if prob > 1e-9 {
				// Add queueing delay to miss latency
				totalLatency := missBucket.Value + queueDelay
				outcomes.Add(prob, sc.AccessResult{Success: false, Latency: totalLatency})
			}
		}
	}

	c.readOutcomes = outcomes
}

// calculateWriteOutcomes generates probabilistic outcomes including contention
func (c *CacheWithContention) calculateWriteOutcomes() {
	outcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
	totalProb := 1.0
	queueDelay := c.calculateQueueingDelay()

	// Failures first
	failProb := c.FailureProb
	if failProb > 1e-9 {
		for _, failBucket := range c.FailureLatency.Buckets {
			prob := failProb * (failBucket.Weight / c.FailureLatency.TotalWeight())
			if prob > 1e-9 {
				totalLatency := failBucket.Value + queueDelay
				outcomes.Add(prob, sc.AccessResult{Success: false, Latency: totalLatency})
			}
		}
		totalProb -= failProb
	}

	// Success
	successProb := totalProb
	if successProb > 1e-9 {
		for _, writeBucket := range c.WriteLatency.Buckets {
			prob := successProb * (writeBucket.Weight / c.WriteLatency.TotalWeight())
			if prob > 1e-9 {
				totalLatency := writeBucket.Value + queueDelay
				outcomes.Add(prob, sc.AccessResult{Success: true, Latency: totalLatency})
			}
		}
	}

	c.writeOutcomes = outcomes
}

// Read simulates a cache read with contention
func (c *CacheWithContention) Read() *Outcomes[sc.AccessResult] {
	if c.readOutcomes == nil {
		c.calculateReadOutcomes()
	}
	return c.readOutcomes
}

// Write simulates a cache write with contention
func (c *CacheWithContention) Write() *Outcomes[sc.AccessResult] {
	if c.writeOutcomes == nil {
		c.calculateWriteOutcomes()
	}
	return c.writeOutcomes
}

// GetUtilization returns current utilization (0.0 to 1.0)
func (c *CacheWithContention) GetUtilization() float64 {
	if c.MaxThroughput <= 0 {
		return 0
	}
	return c.arrivalRate / c.MaxThroughput
}

// GetUtilizationInfo implements UtilizationProvider interface
func (c *CacheWithContention) GetUtilizationInfo() []UtilizationInfo {
	utilization := c.GetUtilization()
	return []UtilizationInfo{
		{
			ResourceName:      "cache",
			ComponentPath:     "", // Will be set by parent
			Utilization:       utilization,
			Capacity:          c.MaxThroughput,
			CurrentLoad:       c.arrivalRate,
			IsBottleneck:      false, // Will be determined by system
			WarningThreshold:  0.7,
			CriticalThreshold: 0.9,
		},
	}
}

// GetFlowPattern implements FlowAnalyzable interface
func (c *CacheWithContention) GetFlowPattern(methodName string, inputRate float64) FlowPattern {
	// Update arrival rate for this analysis
	currentRate := inputRate
	if currentRate <= 0 {
		currentRate = c.arrivalRate
	}

	// Calculate utilization
	utilization := 0.0
	if c.MaxThroughput > 0 {
		utilization = currentRate / c.MaxThroughput
	}

	// Determine success rate based on utilization
	successRate := 1.0 - c.FailureProb
	if utilization >= 1.0 {
		// Overloaded - high failure rate
		successRate = math.Max(0.1, 2.0-utilization)
	} else if utilization > 0.8 {
		// High utilization - some degradation
		successRate *= math.Max(0.8, 1.0-utilization*0.5)
	}

	// Calculate average service time including queueing
	baseServiceTime := 0.0
	queueTime := 0.0

	switch methodName {
	case "Read":
		// Weighted average of hit and miss times
		baseServiceTime = c.HitRate*0.0001 + (1-c.HitRate)*0.0002
	case "Write":
		baseServiceTime = 0.00012
	default:
		baseServiceTime = 0.0001
	}

	// Add queueing delay if under load
	if utilization > 0 && utilization < 1.0 && c.MaxThroughput > 0 {
		avgServiceTime := 1.0 / c.MaxThroughput
		queueTime = (utilization * avgServiceTime) / (1.0 - utilization)
	}

	totalServiceTime := baseServiceTime + queueTime

	return FlowPattern{
		Outflows:      map[string]float64{}, // Cache is a leaf component
		SuccessRate:   successRate,
		Amplification: 1.0,
		ServiceTime:   totalServiceTime,
	}
}