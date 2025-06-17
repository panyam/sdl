package components

import (
	"fmt"
	"math"

	sc "github.com/panyam/sdl/core"
)

// MM1Queue models a basic FIFO queuing system using analytical approximations (M/M/1).
// Its performance characteristics are calculated dynamically based on its configuration.
type MM1Queue struct {
	Name string

	// --- Configuration ---
	// Assumed average arrival rate (items per second). Used for calculations.
	ArrivalRate float64 // λ (lambda)

	// Assumed average service time (seconds per item). Used for calculations.
	AvgServiceTime float64 // Ts = 1/μ
}

// Init initializes the Queue component with default parameters.
func (q *MM1Queue) Init() {
	// Step 1: No embedded components to initialize

	// Step 2: Set defaults only for uninitialized fields (zero values)
	if q.ArrivalRate == 0 {
		q.ArrivalRate = 1e-9
	}
	if q.AvgServiceTime == 0 {
		q.AvgServiceTime = 1e-9
	}

	// Step 3: No derived values to calculate (computed dynamically in methods)
}

// NewMM1Queue creates and initializes a new Queue component.
func NewMM1Queue(name string) *MM1Queue {
	q := &MM1Queue{Name: name}
	q.Init()
	return q
}

// Enqueue simulates adding an item to the queue.
// For an unbounded M/M/1 queue, enqueue itself is typically modelled as near-instantaneous.
func (q *MM1Queue) Enqueue() *Outcomes[sc.AccessResult] {
	outcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
	enqueueLatency := Nanos(10)
	outcomes.Add(1.0, sc.AccessResult{Success: true, Latency: enqueueLatency})
	return outcomes
}

// Dequeue simulates removing an item from the queue and returns the *waiting time* spent in the queue.
// It dynamically uses the M/M/1 analytical approximation based on the component's current parameters.
func (q *MM1Queue) Dequeue() *Outcomes[Duration] {
	outcomes := &Outcomes[Duration]{
		And: func(a, b Duration) Duration { return a + b }, // Duration outcomes sum up
	}

	// --- Dynamic M/M/1 Calculation ---
	if q.AvgServiceTime < 1e-12 {
		q.AvgServiceTime = 1e-12
	}
	serviceRate := 1.0 / q.AvgServiceTime
	utilization := q.ArrivalRate / serviceRate
	isStable := utilization < 1.0
	// --- End Dynamic Calculation ---

	if !isStable {
		infiniteWait := 3600.0 * 24 // 1 day in seconds
		outcomes.Add(1.0, infiniteWait)
		return outcomes
	}

	if utilization < 1e-9 {
		outcomes.Add(1.0, 0.0) // No waiting if utilization is zero
		return outcomes
	}

	// Wq = Ts * ρ / (1 - ρ)
	avgWaitTime := q.AvgServiceTime * utilization / (1.0 - utilization)

	// Approximate the exponential waiting time distribution with discrete buckets
	numBuckets := 5
	totalProb := 1.0
	percentiles := []float64{0.10, 0.30, 0.50, 0.70, 0.90}
	bucketWeights := []float64{0.20, 0.20, 0.20, 0.20, 0.20}

	if len(percentiles) != numBuckets || len(bucketWeights) != numBuckets {
		panic(fmt.Sprintf("MM1Queue '%s': Mismatch in percentile/weight array lengths", q.Name))
	}

	for i := 0; i < numBuckets; i++ {
		p := percentiles[i]
		waitTime := 0.0
		// t = -Wq_avg * ln(1 - P)
		if p < 0.999999 {
			waitTime = -avgWaitTime * math.Log(1.0-p)
		} else {
			waitTime = avgWaitTime * 5 // Approximation for P99+
		}

		if waitTime < 0 {
			waitTime = 0
		}

		outcomes.Add(bucketWeights[i]*totalProb, waitTime)
	}

	return outcomes
}

// GetFlowPattern implements FlowAnalyzable interface for MM1Queue
func (q *MM1Queue) GetFlowPattern(methodName string, inputRate float64) FlowPattern {
	switch methodName {
	case "Enqueue":
		// Enqueue always succeeds (infinite capacity queue)
		return FlowPattern{
			Outflows:      map[string]float64{}, // No downstream calls
			SuccessRate:   1.0,
			Amplification: 1.0,
			ServiceTime:   0.001, // Very fast enqueue operation
		}

	case "Dequeue":
		// Update arrival rate for this analysis
		currentArrivalRate := inputRate
		if currentArrivalRate <= 0 {
			currentArrivalRate = q.ArrivalRate
		}

		// Calculate utilization and stability
		serviceRate := 1.0 / q.AvgServiceTime
		utilization := currentArrivalRate / serviceRate

		// Determine success rate and service time based on utilization
		successRate := 1.0
		serviceTime := q.AvgServiceTime

		if utilization >= 1.0 {
			// Queue is unstable - very poor performance
			successRate = 0.1                   // Most requests will timeout or fail
			serviceTime = q.AvgServiceTime * 10 // Much longer service times
		} else if utilization > 0.9 {
			// High utilization - some degradation
			degradationFactor := 1.0 + (utilization-0.9)*10 // Linear degradation
			successRate = math.Max(0.5, 1.0-degradationFactor*0.1)
			serviceTime = q.AvgServiceTime * degradationFactor
		}

		return FlowPattern{
			Outflows:      map[string]float64{}, // MM1Queue is typically a leaf node
			SuccessRate:   successRate,
			Amplification: 1.0, // Input rate = output rate for successful dequeues
			ServiceTime:   serviceTime,
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
