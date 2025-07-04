// sdl/components/resourcepool.go
package components

import (
	"fmt"
	"math"

	"github.com/panyam/sdl/core"
)

// factorial calculates n!
// A helper needed for the M/M/c queuing formulas.
func factorial(n uint) float64 {
	if n == 0 {
		return 1.0
	}
	res := 1.0
	for i := uint(2); i <= n; i++ {
		res *= float64(i)
	}
	return res
}

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

	// --- Configuration for Queuing Model ---
	ArrivalRate float64 // λ: Average rate requests for this pool arrive (items/sec)
	AvgHoldTime float64 // Ts: Average time resource is held once acquired (seconds/item)

	// --- Computed Metrics (for observability) ---
	lastUtilization float64 // Last calculated utilization (ρ)
}

// Init initializes the ResourcePool with default parameters.
// The core queuing calculations are now done in Acquire() to allow for dynamic changes.
func (rp *ResourcePool) Init() {
	// Step 1: No embedded components to initialize

	// Step 2: Set defaults only for uninitialized fields (zero values)
	if rp.Size == 0 {
		rp.Size = 1
	}
	if rp.ArrivalRate == 0 {
		rp.ArrivalRate = 1e-9
	}
	if rp.AvgHoldTime == 0 {
		rp.AvgHoldTime = 1e-9
	}

	// Step 3: No derived values to calculate (computed dynamically in methods)
}

// NewResourcePool creates and initializes a new ResourcePool component.
func NewResourcePool(name string) *ResourcePool {
	rp := &ResourcePool{Name: name}
	rp.Init()
	return rp
}

// calculateMMCMetrics performs the M/M/c calculation and returns (isStable, avgWaitTimeQ).
// This is extracted from the Acquire method for testing purposes.
func (rp *ResourcePool) calculateMMCMetrics() (bool, float64) {
	if rp.AvgHoldTime < 1e-12 {
		rp.AvgHoldTime = 1e-12 // Avoid division by zero
	}
	serviceRate := 1.0 / rp.AvgHoldTime
	offeredLoad := rp.ArrivalRate / serviceRate
	utilization := offeredLoad / float64(rp.Size)
	rp.lastUtilization = utilization // Store for observability
	isStable := utilization < 1.0

	var avgWaitTimeQ float64 = 0 // Wq
	if !isStable {
		avgWaitTimeQ = math.Inf(1)
	} else if utilization > 1e-12 {
		// Calculate P0 and Lq for M/M/c (infinite K)
		sum1 := 0.0
		for n := uint(0); n < rp.Size; n++ {
			sum1 += math.Pow(offeredLoad, float64(n)) / factorial(n)
		}
		termC := math.Pow(offeredLoad, float64(rp.Size)) / factorial(rp.Size)
		p0_denominator := sum1 + termC*(1.0/(1.0-utilization))

		if p0_denominator > 1e-12 {
			p0 := 1.0 / p0_denominator
			lq_numerator := p0 * termC * utilization
			lq_denominator := math.Pow(1.0-utilization, 2)
			if lq_denominator > 1e-12 {
				lq := lq_numerator / lq_denominator
				if rp.ArrivalRate > 1e-9 {
					avgWaitTimeQ = lq / rp.ArrivalRate
				}
			}
		}
	}
	if avgWaitTimeQ < 0 {
		avgWaitTimeQ = 0
	}

	return isStable, avgWaitTimeQ
}

// Acquire predicts the queuing delay for acquiring one resource from the pool.
// It dynamically calculates the steady-state M/M/c queuing delay (Wq) based on
// the component's current ArrivalRate and AvgHoldTime parameters.
//
// Returns Outcomes[AccessResult]:
//   - Success=true with queuing delay if the pool is stable (utilization < 1).
//   - Success=false with high latency if the pool is unstable (utilization >= 1).
func (rp *ResourcePool) Acquire() *core.Outcomes[core.AccessResult] {
	// Use the extracted calculation method
	isStable, avgWaitTimeQ := rp.calculateMMCMetrics()

	outcomes := &core.Outcomes[core.AccessResult]{}

	// If unstable (rho >= 1), return failure with high latency
	if !isStable || avgWaitTimeQ > 3600.0*24 {
		outcomes.Add(1.0, core.AccessResult{Success: false, Latency: 3600.0 * 24}) // 1 day timeout
		return outcomes
	}

	if avgWaitTimeQ < 1e-9 {
		outcomes.Add(1.0, core.AccessResult{Success: true, Latency: 0.0}) // No waiting if utilization is negligible
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
			if p < 0.999999 && avgWaitTimeQ > 1e-12 { // Avoid log(0) and ensure Wq is positive
				waitTime = -avgWaitTimeQ * math.Log(1.0-p)
			} else if p >= 0.999999 {
				waitTime = avgWaitTimeQ * 7 // Approximation for P99+, increased multiplier slightly
			}

			if waitTime < 0 {
				waitTime = 0
			}
			outcomes.Add(bucketWeights[i]*totalProb, core.AccessResult{Success: true, Latency: waitTime})
		}
	}
	return outcomes
}

// SetArrivalRate sets the arrival rate for a specific method.
// For ResourcePool, we aggregate all method rates since the pool is shared.
func (rp *ResourcePool) SetArrivalRate(method string, rate float64) error {
	// For simplicity, we'll just update the total arrival rate
	// In a more sophisticated implementation, you might track per-method rates
	rp.ArrivalRate = rate
	return nil
}

// GetArrivalRate returns the arrival rate for a specific method.
// For ResourcePool, we return the total rate since it's shared.
func (rp *ResourcePool) GetArrivalRate(method string) float64 {
	return rp.ArrivalRate
}

// GetTotalArrivalRate returns the total arrival rate across all methods.
func (rp *ResourcePool) GetTotalArrivalRate() float64 {
	return rp.ArrivalRate
}

// GetUtilization returns the current utilization (ρ) of the resource pool.
// Values close to 1.0 indicate the system is approaching instability.
func (rp *ResourcePool) GetUtilization() float64 {
	if rp.AvgHoldTime < 1e-12 || rp.Size == 0 {
		return 0
	}
	serviceRate := 1.0 / rp.AvgHoldTime
	offeredLoad := rp.ArrivalRate / serviceRate
	return offeredLoad / float64(rp.Size)
}

// GetUtilizationInfo implements UtilizationProvider interface
func (rp *ResourcePool) GetUtilizationInfo() []UtilizationInfo {
	utilization := rp.GetUtilization()
	return []UtilizationInfo{
		{
			ResourceName:      "pool",
			ComponentPath:     rp.Name,
			Utilization:       utilization,
			Capacity:          float64(rp.Size),
			CurrentLoad:       rp.ArrivalRate,
			IsBottleneck:      true, // Single resource, always the bottleneck
			WarningThreshold:  0.8,
			CriticalThreshold: 0.95,
		},
	}
}

// GetFlowPattern implements FlowAnalyzable interface for ResourcePool
func (rp *ResourcePool) GetFlowPattern(methodName string, inputRate float64) FlowPattern {
	switch methodName {
	case "Acquire":
		// Update arrival rate for this analysis
		currentArrivalRate := inputRate
		if currentArrivalRate <= 0 {
			currentArrivalRate = rp.ArrivalRate
		}

		// Calculate utilization and stability
		serviceRate := 1.0 / rp.AvgHoldTime
		utilization := currentArrivalRate / (serviceRate * float64(rp.Size))

		// Determine success rate based on utilization
		successRate := 1.0
		if utilization >= 1.0 {
			// System is overloaded - success rate drops significantly
			successRate = math.Max(0.1, 2.0-utilization) // Drops to 0.1 at utilization = 1.9
		} else if utilization > 0.8 {
			// High utilization - some degradation
			successRate = math.Max(0.9, 1.0-utilization*0.25)
		}

		return FlowPattern{
			Outflows:      map[string]float64{}, // ResourcePool is typically a leaf node
			SuccessRate:   successRate,
			Amplification: 1.0, // Input rate = output rate for successful acquisitions
			ServiceTime:   rp.AvgHoldTime,
		}

	case "Release":
		// Release operations typically succeed
		return FlowPattern{
			Outflows:      map[string]float64{}, // No downstream calls
			SuccessRate:   1.0,
			Amplification: 1.0,
			ServiceTime:   0.001, // Very fast operation
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
