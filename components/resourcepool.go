// sdl/components/resourcepool.go
package components

import (
	"fmt"
	"math"

	// No longer need "sync"

	sc "github.com/panyam/sdl/core"
)

// ResourcePool models a pool of limited, identical resources using the M/M/c
// analytical queuing model to predict average steady-state waiting time (Wq).
// It is a stateless component configured with size, average arrival rate,
// and average hold time. Its methods predict steady-state performance
// based solely on these initial configuration parameters.
//
// Limitations:
//   - Analytical Model: Provides steady-state average queueing delay (Wq). It does not
//     capture the variance from bursty arrivals or non-Poisson processes like DES would.
//   - Configuration Driven: Behavior depends entirely on the configured rates; it does
//     not dynamically adapt to observed load or previous calls within a simulation run.
//   - Cold Starts: Does not model system warm-up or initially empty queues.
//   - Stateless: Does not track the instantaneous number of used resources. Acquire/Release
//     semantics are purely statistical based on configured rates.
type ResourcePool struct {
	Name string
	Size uint // Maximum number of concurrent users/holders (c)
	// Removed: Used uint
	// Removed: mu sync.Mutex

	// --- Configuration for Queuing Model ---
	ArrivalRate float64 // λ: Average rate requests for this pool arrive (items/sec)
	AvgHoldTime float64 // Ts: Average time resource is held once acquired (seconds/item)

	// --- Derived M/M/c Values (Calculated in Init) ---
	serviceRate  float64 // μ = 1 / AvgHoldTime
	offeredLoad  float64 // a = λ / μ
	utilization  float64 // ρ = a / c
	isStable     bool    // rho < 1
	avgWaitTimeQ float64 // Average M/M/c waiting time (Wq)
}

// Init initializes the ResourcePool, calculating steady-state M/M/c metrics.
func (rp *ResourcePool) Init(name string, size uint, lambda float64, ts float64) *ResourcePool {
	rp.Name = name
	if size == 0 {
		// log.Printf("Warning: ResourcePool '%s' initialized with size=0. Using size=1.", name)
		size = 1
	}
	rp.Size = size

	// Validate and Store Config
	if lambda <= 0 {
		// log.Printf("Warning: ResourcePool '%s' initialized with lambda <= 0 (%.3f). Using small positive.", name, lambda)
		lambda = 1e-9
	}
	if ts <= 0 {
		// log.Printf("Warning: ResourcePool '%s' initialized with AvgHoldTime <= 0 (%.6f). Using small positive.", name, ts)
		ts = 1e-9
	}
	rp.ArrivalRate = lambda
	rp.AvgHoldTime = ts

	// Calculate Derived M/M/c Metrics
	rp.serviceRate = 1.0 / rp.AvgHoldTime              // mu = 1 / Ts
	rp.offeredLoad = rp.ArrivalRate / rp.serviceRate   // a = lambda / mu
	rp.utilization = rp.offeredLoad / float64(rp.Size) // rho = a / c
	rp.isStable = rp.utilization < 1.0

	// Calculate M/M/c average wait time Wq (assuming infinite queue K for simplicity)
	rp.avgWaitTimeQ = 0 // Default to 0
	if rp.utilization >= 1.0 {
		// Unstable case
		rp.avgWaitTimeQ = math.Inf(1)
	} else if rp.utilization > 1e-12 { // Avoid calculations if load is effectively zero
		// Stable case, calculate P0 and Lq for M/M/c (infinite K)
		sum1 := 0.0
		for n := uint(0); n < rp.Size; n++ {
			termN := math.Pow(rp.offeredLoad, float64(n)) / factorial(n)
			// Check for potential overflow/Inf in termN if offeredLoad is very large
			if math.IsInf(termN, 0) || math.IsNaN(termN) {
				// log.Printf("Warning: Term overflow in P0 calc for ResourcePool '%s'. OfferedLoad likely too high.", rp.Name)
				rp.avgWaitTimeQ = math.Inf(1) // Treat as unstable if terms overflow
				goto EndWqCalc                // Skip rest of calc
			}
			sum1 += termN
		}
		termC := math.Pow(rp.offeredLoad, float64(rp.Size)) / factorial(rp.Size)
		if math.IsInf(termC, 0) || math.IsNaN(termC) {
			// log.Printf("Warning: TermC overflow in P0 calc for ResourcePool '%s'. OfferedLoad likely too high.", rp.Name)
			rp.avgWaitTimeQ = math.Inf(1)
			goto EndWqCalc
		}

		p0_inf_denominator := sum1 + termC*(1.0/(1.0-rp.utilization))
		if p0_inf_denominator <= 1e-12 || math.IsInf(p0_inf_denominator, 0) || math.IsNaN(p0_inf_denominator) {
			// Denominator is effectively zero or Inf, implies instability or P0 near zero
			// Wq could be Inf or 0 depending on exact limit, assume Inf if unstable was missed
			if !rp.isStable {
				rp.avgWaitTimeQ = math.Inf(1)
			} // Otherwise remains 0, which is likely correct if P0 is ~0
			goto EndWqCalc
		}
		p0_inf := 1.0 / p0_inf_denominator

		// Calculate Lq for M/M/c (infinite K)
		lq_inf_numerator := p0_inf * termC * rp.utilization
		lq_inf_denominator := math.Pow(1.0-rp.utilization, 2)

		if lq_inf_denominator <= 1e-12 || math.IsInf(lq_inf_numerator, 0) || math.IsNaN(lq_inf_numerator) {
			// Problem calculating Lq, implies instability or extreme values
			if !rp.isStable {
				rp.avgWaitTimeQ = math.Inf(1)
			} // Otherwise remains 0
			goto EndWqCalc
		}
		lq_inf := lq_inf_numerator / lq_inf_denominator

		// Calculate Wq using Little's Law (Wq = Lq / lambda)
		if rp.ArrivalRate > 1e-9 && !math.IsInf(lq_inf, 0) && !math.IsNaN(lq_inf) && lq_inf >= 0 {
			rp.avgWaitTimeQ = lq_inf / rp.ArrivalRate
		}
		// If Lq or lambda is zero/Inf/NaN, avgWaitTimeQ defaults to 0, which is correct.
	}

EndWqCalc:
	if rp.avgWaitTimeQ < 0 { // Should not happen, but safety check
		rp.avgWaitTimeQ = 0
	}

	return rp
}

// NewResourcePool creates and initializes a new ResourcePool component.
func NewResourcePool(name string, size uint, arrivalRate float64, avgHoldTime float64) *ResourcePool {
	rp := &ResourcePool{}
	return rp.Init(name, size, arrivalRate, avgHoldTime)
}

// Acquire predicts the outcome of attempting to acquire one resource from the pool
// based on the steady-state M/M/c analysis performed during Init.
// Returns Outcomes[sc.AccessResult]:
// - Success=true, Latency=0: Acquired immediately (pool underutilized on average, Wq ~ 0).
// - Success=true, Latency=WaitTimeDist: Acquired after average queuing delay Wq (pool utilized on average).
// - Success=false, Latency=0: Rejected (pool unstable, Wq is infinite).
func (rp *ResourcePool) Acquire() *Outcomes[sc.AccessResult] {

	outcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
	avgWaitTime := rp.avgWaitTimeQ // Use pre-calculated Wq from Init

	if math.IsInf(avgWaitTime, 1) || avgWaitTime > 3600.0*24 { // Treat effectively infinite waits as rejection
		// Unstable (rho >= 1) -> Reject
		outcomes.Add(1.0, sc.AccessResult{Success: false, Latency: 0})
	} else if avgWaitTime < 1e-9 {
		// Stable (rho < 1) and Negligible Wait -> Immediate Success
		outcomes.Add(1.0, sc.AccessResult{Success: true, Latency: 0})
	} else {
		// Stable (rho < 1), Positive Wait -> Generate Wait Distribution
		// Approximate distribution using exponential percentiles scaled by calculated Wq
		numBuckets := 5 // Number of buckets to approximate the distribution
		totalProb := 1.0
		percentiles := []float64{0.10, 0.30, 0.50, 0.70, 0.90}
		bucketWeights := []float64{0.20, 0.20, 0.20, 0.20, 0.20} // Should sum to 1.0

		if len(percentiles) != numBuckets || len(bucketWeights) != numBuckets {
			panic(fmt.Sprintf("ResourcePool '%s': Mismatch in percentile/weight array lengths for Acquire", rp.Name))
		}

		for i := 0; i < numBuckets; i++ {
			p := percentiles[i]
			waitTime := 0.0
			// Use average wait time Wq here for exponential distribution parameterization
			if p < 0.999999 && avgWaitTime > 1e-12 { // Avoid log(0) and ensure Wq is positive
				waitTime = -avgWaitTime * math.Log(1.0-p)
			} else if p >= 0.999999 {
				waitTime = avgWaitTime * 7 // Approximation for P99+, increased multiplier slightly
			}

			if waitTime < 0 {
				waitTime = 0
			}
			outcomes.Add(bucketWeights[i]*totalProb, sc.AccessResult{
				Success: true, // Acquired *after* waiting
				Latency: waitTime,
			})
		}
	}
	return outcomes
}
