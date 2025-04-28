package sdl

import (
	"math"
	// "log"
)

// Queue models a basic FIFO queue with analytical waiting time estimation.
type Queue struct {
	Name string

	// --- Queue Parameters (for Analytical Modelling) ---

	// AvgArrivalRate (lambda): Average number of items arriving per unit time (e.g., items/second).
	// This needs to be configured or estimated by the user based on expected load.
	AvgArrivalRate float64

	// AvgServiceRate (mu): Average number of items the *single server* processing items
	// from this queue can handle per unit time (e.g., items/second).
	// This is often derived from 1 / MeanLatency of the servicing component.
	AvgServiceRate float64

	// NumberOfServers (c): For M/M/c model - how many parallel servers process items
	// from this queue. Defaults to 1 for M/M/1.
	NumberOfServers int

	// Optional: MaxQueueSize (K): For bounded queue models (M/M/1/K). 0 means infinite.
	MaxQueueSize uint

	// --- Latency/Failure of Queue Operations Themselves ---
	EnqueueLatency  Outcomes[Duration] // Latency to simply add an item
	EnqueueFailProb float64            // Probability enqueue fails (e.g., if bounded queue is full)
	DequeueLatency  Outcomes[Duration] // Latency of the dequeue operation itself (finding item)

	// Failure of queue mechanism itself (distinct from enqueue full / dequeue empty)
	QueueFailProb    float64
	QueueFailLatency Outcomes[Duration]

	// --- Precomputed? ---
	// Enqueue outcomes can be precomputed. Dequeue needs dynamic wait time.
	enqueueOutcomes *Outcomes[AccessResult]
}

// Init initializes the Queue component.
func (q *Queue) Init() *Queue {
	q.Name = "Queue" // Default name

	// Default Rates (User MUST configure these for meaningful results)
	q.AvgArrivalRate = 100.0 // Default: 100 items/sec
	q.AvgServiceRate = 110.0 // Default: Service slightly faster than arrival
	q.NumberOfServers = 1    // Default: Single server model (M/M/1)

	q.MaxQueueSize = 0 // Default: Infinite queue

	// Default latencies for queue operations
	q.EnqueueLatency.Add(1.0, Nanos(50)) // Fast enqueue
	q.EnqueueLatency.And = func(a, b Duration) Duration { return a + b }
	q.EnqueueFailProb = 0.0 // Assume enqueue to infinite queue doesn't fail

	q.DequeueLatency.Add(1.0, Nanos(50)) // Fast dequeue op
	q.DequeueLatency.And = func(a, b Duration) Duration { return a + b }

	q.QueueFailProb = 0.0001               // 0.01% chance queue mechanism fails
	q.QueueFailLatency.Add(1.0, Millis(5)) // Failure detected relatively quickly
	q.QueueFailLatency.And = func(a, b Duration) Duration { return a + b }

	q.calculateEnqueueOutcomes()

	return q
}

// NewQueue creates and initializes a new Queue component with defaults.
func NewQueue() *Queue {
	q := &Queue{}
	return q.Init()
}

// ConfigureRates allows setting the key analytical parameters.
func (q *Queue) ConfigureRates(arrivalRate, serviceRate float64, numServers int) *Queue {
	if arrivalRate >= 0 {
		q.AvgArrivalRate = arrivalRate
	}
	if serviceRate > 0 {
		q.AvgServiceRate = serviceRate
	} // Must be > 0
	if numServers > 0 {
		q.NumberOfServers = numServers
	} else {
		q.NumberOfServers = 1
	}
	// Note: Changing rates might invalidate cached outcomes if we cached Dequeue later
	return q
}

// calculateEnqueueOutcomes generates the outcomes for trying to enqueue.
func (q *Queue) calculateEnqueueOutcomes() {
	outcomes := &Outcomes[AccessResult]{And: AndAccessResults}
	totalProb := 1.0

	// --- Queue Mechanism Failure ---
	failProb := q.QueueFailProb
	if failProb > 1e-9 {
		for _, bucket := range q.QueueFailLatency.Buckets {
			prob := failProb * (bucket.Weight / q.QueueFailLatency.TotalWeight())
			if prob > 1e-9 {
				outcomes.Add(prob, AccessResult{false, bucket.Value})
			}
		}
		totalProb -= failProb
	}

	// --- Enqueue Full Failure (Only for bounded queues - simplified for now) ---
	if q.MaxQueueSize > 0 {
		enqueueFailProb := q.EnqueueFailProb * totalProb // Effective probability
		// Need a model here based on AvgArrivalRate, AvgServiceRate, MaxQueueSize
		// Simplified: Use configured EnqueueFailProb directly for now.
		if q.MaxQueueSize > 0 && enqueueFailProb > 1e-9 {
			// Use QueueFailLatency for now, could define separate EnqueueFailLatency
			for _, bucket := range q.QueueFailLatency.Buckets {
				prob := enqueueFailProb * (bucket.Weight / q.QueueFailLatency.TotalWeight())
				if prob > 1e-9 {
					outcomes.Add(prob, AccessResult{false, bucket.Value})
				}
			}
			totalProb -= enqueueFailProb
		}
	}

	// --- Enqueue Success ---
	successProb := totalProb
	if successProb > 1e-9 {
		for _, bucket := range q.EnqueueLatency.Buckets {
			prob := successProb * (bucket.Weight / q.EnqueueLatency.TotalWeight())
			if prob > 1e-9 {
				outcomes.Add(prob, AccessResult{true, bucket.Value})
			}
		}
	}
	q.enqueueOutcomes = outcomes
}

// Enqueue simulates adding an item to the queue.
// Returns outcomes for the enqueue operation itself (fast).
func (q *Queue) Enqueue() *Outcomes[AccessResult] {
	if q.enqueueOutcomes == nil {
		q.calculateEnqueueOutcomes()
	}
	return q.enqueueOutcomes
}

// Dequeue simulates removing an item from the queue.
// Returns Outcomes[Duration] representing the *WAITING TIME* in the queue,
// calculated using analytical models (M/M/c currently).
// Returns nil or specific error outcome if queue is determined to be unstable.
func (q *Queue) Dequeue() *Outcomes[Duration] {
	outcomes := &Outcomes[Duration]{And: func(a, b Duration) Duration { return a + b }}

	lambda := q.AvgArrivalRate
	mu := q.AvgServiceRate
	c := float64(q.NumberOfServers)
	uint_c := uint64(q.NumberOfServers) // Need uint for factorial helper

	if mu <= 0 {
		return outcomes
	} // Invalid service rate

	rho := lambda / (c * mu) // Utilization per server

	if rho >= 1.0 {
		return outcomes
	} // Unstable queue

	// --- Calculate M/M/c Waiting Time (Wq) ---

	// Calculate P0 (Probability of 0 customers in system)
	sumTerm := 0.0
	for k := 0; k < q.NumberOfServers; k++ { // Iterate up to c-1
		fact_k := factorial(uint64(k))
		if math.IsInf(fact_k, 1) { /* Handle potential overflow */
			return outcomes
		}
		sumTerm += math.Pow(lambda/mu, float64(k)) / fact_k
	}

	fact_c := factorial(uint_c)
	if math.IsInf(fact_c, 1) { /* Handle potential overflow */
		return outcomes
	}

	lastTermNum := math.Pow(lambda/mu, c)
	lastTermDenom := fact_c * (1.0 - rho)
	if math.Abs(lastTermDenom) < 1e-12 { /* Handle potential division by zero */
		return outcomes
	} // Should not happen if rho < 1

	p0 := 1.0 / (sumTerm + (lastTermNum / lastTermDenom))
	if math.IsNaN(p0) || math.IsInf(p0, 0) || p0 < 0 { /* Handle invalid P0 calculation */
		return outcomes
	}

	// Calculate Average Queue Length (Lq)
	lq_numerator := (p0 * math.Pow(lambda/mu, c) * rho)
	lq_denominator := (fact_c * math.Pow(1.0-rho, 2))
	if math.Abs(lq_denominator) < 1e-12 { /* Handle potential division by zero */
		return outcomes
	}
	lq := lq_numerator / lq_denominator

	// Calculate Average Waiting Time in Queue (Wq) using Little's Law (Lq = lambda * Wq)
	wq := 0.0
	if lambda > 1e-9 {
		wq = lq / lambda
	}
	if wq < 0 {
		wq = 0
	} // Waiting time cannot be negative

	// --- Model Waiting Time Distribution (Exponential Approximation) ---
	numWaitBuckets := 20
	maxWaitTimeFactor := 5.0
	expRate := c*mu - lambda // Rate parameter for exponential distribution

	if wq < 1e-12 || expRate <= 0 { // Handle Wq=0 or invalid rate
		outcomes.Add(1.0, 0.0)
		return outcomes
	}

	maxT := maxWaitTimeFactor * wq
	timeStep := maxT / float64(numWaitBuckets)
	cumulativeProb := 0.0

	for i := 0; i < numWaitBuckets; i++ {
		t_start := float64(i) * timeStep
		t_end := float64(i+1) * timeStep
		if i == numWaitBuckets-1 {
			t_end = math.Inf(1)
		}

		prob_end := 1.0
		if !math.IsInf(t_end, 1) {
			prob_end = 1.0 - math.Exp(-expRate*t_end)
		}
		prob_start := 1.0 - math.Exp(-expRate*t_start)
		bucketProb := prob_end - prob_start

		bucketLatency := t_start + timeStep*0.5
		if math.IsInf(t_end, 1) {
			bucketLatency = t_start + 1.0/expRate
		}
		if bucketLatency < 0 {
			bucketLatency = 0
		} // Ensure non-negative

		if bucketProb > 1e-9 {
			outcomes.Add(bucketProb, bucketLatency)
			cumulativeProb += bucketProb
		}
	}

	// Renormalize
	if cumulativeProb > 1e-9 {
		renormFactor := 1.0 / cumulativeProb
		for i := range outcomes.Buckets {
			outcomes.Buckets[i].Weight *= renormFactor
		}
	} else if outcomes.Len() == 0 { // If no buckets added, add zero wait time
		outcomes.Add(1.0, 0.0)
	}

	return outcomes
}

// --- Add Factorial Helper Function ---
// Calculates factorial(n) for non-negative integers.
// Returns 1 for factorial(0).
// Handles potential overflow by returning +Inf if result exceeds float64 max,
// although unlikely for typical 'c' values in M/M/c.
func factorial(n uint64) float64 {
	if n == 0 {
		return 1.0
	}
	if n > 170 { // Factorial(171) exceeds float64 max
		return math.Inf(1)
	}
	result := 1.0
	for i := uint64(2); i <= n; i++ {
		result *= float64(i)
		if math.IsInf(result, 1) { // Check for overflow during calculation
			return math.Inf(1)
		}
	}
	return result
}
