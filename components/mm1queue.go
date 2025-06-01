package components

import (
	"fmt"
	"math"

	sc "github.com/panyam/sdl/core"
)

// Queue models a basic FIFO queuing system using analytical approximations (M/M/1).
type MM1Queue struct {
	Name string

	// --- Configuration ---
	// Assumed average arrival rate (items per second). Used for calculations.
	ArrivalRate float64 // λ (lambda)

	// Assumed average service time (seconds per item). Used for calculations.
	// This could potentially be derived from the MeanLatency of the component
	// that processes items dequeued from this queue.
	AvgServiceTime float64 // Ts = 1/μ

	// Optional: Max Queue Length (0 for unbounded)
	// MaxLength uint64 // TODO: Implement logic for bounded queues later

	// Internal derived values
	serviceRate float64 // μ (mu) = 1 / AvgServiceTime
	utilization float64 // ρ (rho) = λ / μ
	isStable    bool
}

// Init initializes the Queue component.
// Requires average arrival rate (lambda, items/sec) and average service time (Ts, seconds/item).
func (q *MM1Queue) Init() {
	q.ArrivalRate = 1e-9
	q.AvgServiceTime = 1e-9
	q.serviceRate = 1.0 / q.AvgServiceTime        // mu = 1 / Ts
	q.utilization = q.ArrivalRate / q.serviceRate // rho = lambda / mu
	q.isStable = q.utilization < 1.0

	if !q.isStable {
		// log.Printf("Warning: Queue '%s' is unstable (Utilization rho = %.3f >= 1.0). Waiting times will be infinite.", name, q.utilization)
	}

	// Note: Enqueue/Dequeue outcomes depend on these parameters, calculated on the fly.
}

// NewMM1Queue creates and initializes a new Queue component.
func NewMM1Queue(name string) *MM1Queue {
	q := &MM1Queue{}
	q.Init()
	return q
}

// Enqueue simulates adding an item to the queue.
// For an unbounded M/M/1 queue, enqueue itself is typically modelled as near-instantaneous.
// Failures could occur if the queue had a max length (TODO).
func (q *MM1Queue) Enqueue() *Outcomes[sc.AccessResult] {
	// Simple model: Enqueue is fast and always succeeds (for unbounded queue)
	outcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
	// Small CPU cost for enqueue logic?
	enqueueLatency := Nanos(10)
	outcomes.Add(1.0, sc.AccessResult{Success: true, Latency: enqueueLatency})
	return outcomes
}

// Dequeue simulates removing an item from the queue and returns the *waiting time* spent in the queue.
// Uses M/M/1 analytical approximation.
func (q *MM1Queue) Dequeue() *Outcomes[Duration] {
	outcomes := &Outcomes[Duration]{
		And: func(a, b Duration) Duration { return a + b }, // Duration outcomes sum up
	}

	if !q.isStable {
		// Queue is unstable, waiting time is theoretically infinite.
		// How to represent this? Return a huge duration? Special error value?
		// Let's return a large duration indicating instability.
		infiniteWait := 3600.0 * 24 // 1 day in seconds - effectively infinite
		outcomes.Add(1.0, infiniteWait)
		// log.Printf("Queue '%s': Dequeue returning infinite wait due to instability (rho=%.3f)", q.Name, q.utilization)
		return outcomes
	}

	if q.utilization < 1e-9 {
		// If utilization is virtually zero, waiting time is zero.
		outcomes.Add(1.0, 0.0)
		return outcomes
	}

	// Calculate average waiting time in queue (Wq) using M/M/1 formula
	// Wq = Ts * ρ / (1 - ρ)
	avgWaitTime := q.AvgServiceTime * q.utilization / (1.0 - q.utilization)

	// The M/M/1 waiting time distribution is also exponential (with parameter mu*(1-rho)).
	// To represent this with discrete buckets:
	// We can create buckets representing percentiles of the exponential distribution.
	// CDF: P(Wait <= t) = 1 - exp(-t / Wq_avg)  =>  t = -Wq_avg * ln(1 - P)
	// P = probability (e.g., 0.5 for median, 0.9 for P90)

	numBuckets := 5 // Number of buckets to approximate the distribution
	totalProb := 1.0

	// Add buckets for P10, P30, P50, P70, P90 (or similar percentiles)
	percentiles := []float64{0.10, 0.30, 0.50, 0.70, 0.90}
	bucketWeights := []float64{0.20, 0.20, 0.20, 0.20, 0.20} // Equal weight per quantile range

	if len(percentiles) != numBuckets || len(bucketWeights) != numBuckets {
		panic(fmt.Sprintf("Queue '%s': Mismatch in percentile/weight array lengths", q.Name))
	}

	for i := 0; i < numBuckets; i++ {
		p := percentiles[i]
		// Calculate the waiting time corresponding to this percentile
		// Use the middle of the quantile range for representative latency?
		// E.g., for 0-20%, use P10. For 20-40%, use P30.
		waitTime := 0.0
		// Avoid log(0) by using 1-p, ensure p is not exactly 1.0
		if p < 0.999999 {
			waitTime = -avgWaitTime * math.Log(1.0-p)
		} else {
			// For high percentiles, use a large multiple of avgWaitTime
			waitTime = avgWaitTime * 5 // Approximation for P99+
		}

		if waitTime < 0 {
			waitTime = 0
		} // Ensure non-negative

		outcomes.Add(bucketWeights[i]*totalProb, waitTime)
	}

	return outcomes
}
