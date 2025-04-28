package sdl

import (
	// "log"
	"fmt"
	"math"
	"sync"
	// Assuming M/M/c helper functions (factorial, calculateP0, calculateLq etc.) from queue.go are accessible
	// If not, they need to be moved to a common utility package or duplicated.
	// For now, assume they are accessible (e.g., in the same package or imported).
)

// AcquireAttemptResult is no longer needed, Acquire returns AccessResult directly
/*
type AcquireAttemptResult struct {
	Success bool
	Pool    *ResourcePool
}
*/

// ResourcePool models a pool of limited, identical resources (e.g., connections, threads).
// Uses M/M/c analytical model to estimate queueing delay when pool is full.
type ResourcePool struct {
	Name string // Optional identifier
	Size uint   // Maximum number of concurrent users/holders (c)
	Used uint   // Current number of resources in use (State!)

	// --- Configuration for Queuing Model ---
	// Assumed average arrival rate of requests needing this resource (items/sec)
	ArrivalRate float64 // λ (lambda)
	// Assumed average time a resource is held once acquired (service time, seconds/item)
	AvgHoldTime float64 // Ts = 1/μ

	// --- Derived M/M/c Values ---
	serviceRate  float64 // μ (mu) = 1 / AvgHoldTime
	offeredLoad  float64 // a = λ / μ
	utilization  float64 // ρ (rho) = a / c
	isStable     bool    // rho < 1
	avgWaitTimeQ float64 // Average M/M/c waiting time (Wq) when queuing is necessary

	mu sync.Mutex
}

// Init initializes the ResourcePool.
func (rp *ResourcePool) Init(name string, size uint, lambda float64, ts float64) *ResourcePool {
	rp.Name = name
	if size == 0 {
		size = 1 /* Avoid division by zero, maybe log warning */
	}
	rp.Size = size
	rp.Used = 0

	// --- Validate and Store Config ---
	if lambda <= 0 {
		lambda = 1e-9
	}
	if ts <= 0 {
		ts = 1e-9
	}
	rp.ArrivalRate = lambda
	rp.AvgHoldTime = ts

	// --- Calculate Derived M/M/c Metrics ---
	rp.serviceRate = 1.0 / rp.AvgHoldTime              // mu = 1 / Ts
	rp.offeredLoad = rp.ArrivalRate / rp.serviceRate   // a = lambda / mu
	rp.utilization = rp.offeredLoad / float64(rp.Size) // rho = a / c
	rp.isStable = rp.utilization < 1.0

	// Calculate M/M/c average wait time Wq (assuming infinite queue K for simplicity here)
	// We need P0 and Lq for M/M/c (infinite K version)
	if rp.isStable {
		// Calculate P0 for M/M/c (infinite K)
		sum1 := 0.0
		for n := uint(0); n < rp.Size; n++ {
			sum1 += math.Pow(rp.offeredLoad, float64(n)) / factorial(n)
		}
		termC := math.Pow(rp.offeredLoad, float64(rp.Size)) / factorial(rp.Size)
		p0_inf := 1.0 / (sum1 + termC*(1.0/(1.0-rp.utilization)))

		// Calculate Lq for M/M/c (infinite K)
		lq_inf := p0_inf * (math.Pow(rp.offeredLoad, float64(rp.Size)) * rp.utilization) / (factorial(rp.Size) * math.Pow(1.0-rp.utilization, 2))

		// Calculate Wq using Little's Law (Wq = Lq / lambda) - lambda_eff = lambda here
		if rp.ArrivalRate > 1e-9 && !math.IsInf(lq_inf, 0) && !math.IsNaN(lq_inf) && lq_inf >= 0 {
			rp.avgWaitTimeQ = lq_inf / rp.ArrivalRate
		} else {
			rp.avgWaitTimeQ = 0
		}
	} else {
		// Unstable case for infinite queue
		rp.avgWaitTimeQ = math.Inf(1)
	}
	if rp.avgWaitTimeQ < 0 {
		rp.avgWaitTimeQ = 0
	} // Ensure non-negative

	// log.Printf("ResourcePool '%s' Init: Size=%d, lambda=%.2f, Ts=%.4f, mu=%.2f, a=%.3f, rho=%.3f, Wq=%.6f",
	//      rp.Name, rp.Size, rp.ArrivalRate, rp.AvgHoldTime, rp.serviceRate, rp.offeredLoad, rp.utilization, rp.avgWaitTimeQ)

	return rp
}

// NewResourcePool creates and initializes a new ResourcePool component.
func NewResourcePool(name string, size uint, arrivalRate float64, avgHoldTime float64) *ResourcePool {
	rp := &ResourcePool{}
	return rp.Init(name, size, arrivalRate, avgHoldTime)
}

// Acquire attempts to acquire one resource from the pool.
// Returns Outcomes[AccessResult]:
// - Success=true, Latency=0: Acquired immediately.
// - Success=true, Latency=WaitTimeDist: Acquired after queuing delay.
// - Success=false, Latency=0: Rejected (if modelled, not currently implemented).
func (rp *ResourcePool) Acquire() *Outcomes[AccessResult] {
	rp.mu.Lock()
	needsQueueing := rp.Used >= rp.Size
	// Simulate potential state change for caller's benefit if needed? Risky.
	// For now, decision is based purely on state *before* this call.
	rp.mu.Unlock()

	outcomes := &Outcomes[AccessResult]{And: AndAccessResults}

	if !needsQueueing {
		// --- Resource Available ---
		// Success: Acquired immediately with zero latency.
		outcomes.Add(1.0, AccessResult{
			Success: true,
			Latency: 0, // Or small CPU cost? Assume 0 for now.
		})
		// NOTE: State rp.Used is NOT incremented here in the pure model.
		// The caller or simulation loop would need to handle this.
		// log.Printf("Pool '%s': Acquire IMMEDIATE SUCCESS (Used=%d, Size=%d)", rp.Name, rp.Used, rp.Size) // Debugging state *before* hypothetical acquire

	} else {
		// --- Resource Busy - Must Queue ---
		avgWaitTime := rp.avgWaitTimeQ // Use pre-calculated Wq

		if math.IsInf(avgWaitTime, 1) || avgWaitTime > 3600.0*24 {
			// Unstable or excessively long wait - model as rejection?
			// For now, let's return failure if Wq is "infinite".
			// log.Printf("Pool '%s': Acquire FAILURE (infinite Wq) (Used=%d, Size=%d)", rp.Name, rp.Used, rp.Size)
			outcomes.Add(1.0, AccessResult{Success: false, Latency: 0})

		} else if avgWaitTime < 1e-9 {
			// Wait time is negligible, treat as immediate success.
			// This can happen if rho is very low but pool is momentarily full.
			// log.Printf("Pool '%s': Acquire QUEUED but Wq ~ 0 (Used=%d, Size=%d)", rp.Name, rp.Used, rp.Size)
			outcomes.Add(1.0, AccessResult{Success: true, Latency: 0})

		} else {
			// Stable queue, positive wait time. Generate wait time distribution.
			// log.Printf("Pool '%s': Acquire QUEUED (Wq=%.6f) (Used=%d, Size=%d)", rp.Name, avgWaitTime, rp.Used, rp.Size)

			// Approximate distribution using exponential percentiles scaled by calculated Wq
			numBuckets := 5 // Keep consistent with Queue implementation
			totalProb := 1.0
			percentiles := []float64{0.10, 0.30, 0.50, 0.70, 0.90}
			bucketWeights := []float64{0.20, 0.20, 0.20, 0.20, 0.20} // Should sum to 1.0

			if len(percentiles) != numBuckets || len(bucketWeights) != numBuckets {
				panic(fmt.Sprintf("ResourcePool '%s': Mismatch in percentile/weight array lengths for Acquire", rp.Name))
			}

			for i := 0; i < numBuckets; i++ {
				p := percentiles[i]
				waitTime := 0.0
				if p < 0.999999 && avgWaitTime > 1e-12 {
					waitTime = -avgWaitTime * math.Log(1.0-p)
				} else if p >= 0.999999 {
					waitTime = avgWaitTime * 5 // Approximation for P99+
				}

				if waitTime < 0 {
					waitTime = 0
				}
				outcomes.Add(bucketWeights[i]*totalProb, AccessResult{
					Success: true, // Acquired *after* waiting
					Latency: waitTime,
				})
			}
		}
		// NOTE: State rp.Used is NOT incremented here.
	}

	return outcomes
}

// Release returns one resource to the pool. (Direct state modification - limitation)
func (rp *ResourcePool) Release() {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	if rp.Used > 0 {
		rp.Used--
	} else {
		// log.Printf("Warning: Pool '%s': Release called when Used count is zero.", rp.Name)
	}
}
