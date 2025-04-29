package components

import (
	"fmt"
	"math"

	sc "github.com/panyam/leetcoach/sdl/core"
	// "log"
)

// --- Helper functions for M/M/c/K formulas ---

// Calculates n!
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

// Calculates P0 for M/M/c/K
func calculateP0(a float64, c uint, k uint, rho float64) float64 {
	if c == 0 {
		return 0
	} // Invalid configuration

	sum1 := 0.0 // Sum from n=0 to c-1
	for n := uint(0); n < c; n++ {
		sum1 += math.Pow(a, float64(n)) / factorial(n)
	}

	termC := math.Pow(a, float64(c)) / factorial(c)

	sum2 := 0.0                   // Sum from n=c to k
	if math.Abs(1.0-rho) < 1e-9 { // Handle rho == 1 case for sum2
		sum2 = float64(k - c + 1)
	} else {
		// Geometric series sum adjusted
		sum2 = (1.0 - math.Pow(rho, float64(k-c+1))) / (1.0 - rho)
	}

	denominator := sum1 + termC*sum2
	if denominator == 0 {
		return 0
	} // Avoid division by zero
	return 1.0 / denominator
}

// Calculates Pk (blocking probability) for M/M/c/K
func calculatePk(p0 float64, a float64, c uint, k uint, rho float64) float64 {
	if c == 0 || k < c {
		return 1.0
	} // Invalid or always blocked if K < c
	if math.IsInf(p0, 0) || math.IsNaN(p0) || p0 == 0 {
		return 0
	} // If P0 is zero, PK is likely zero

	termC := math.Pow(a, float64(c)) / factorial(c)
	pk := p0 * termC * math.Pow(rho, float64(k-c))

	// Clamp pk between 0 and 1 due to potential floating point inaccuracies
	if pk < 0 {
		return 0
	}
	if pk > 1.0 {
		return 1.0
	}
	return pk
}

// Calculates Lq (average queue length) for M/M/c/K
func calculateLq(p0 float64, a float64, c uint, k uint, rho float64) float64 {
	if c == 0 || k < c {
		return 0
	} // No queue if K<c
	if rho >= 1.0-1e-9 {
		return math.Inf(1)
	} // Infinite queue length if unstable/rho=1
	if math.IsInf(p0, 0) || math.IsNaN(p0) || p0 == 0 {
		return 0
	}

	termC := math.Pow(a, float64(c)) / factorial(c)
	rhoTerm := rho / math.Pow(1.0-rho, 2)
	kMinusC := float64(k - c)
	rhoPowKMinC := math.Pow(rho, kMinusC)

	bracket := 1.0 - rhoPowKMinC - (kMinusC * rhoPowKMinC * (1.0 - rho))

	lq := p0 * termC * rhoTerm * bracket
	if lq < 0 {
		return 0
	} // Queue length cannot be negative
	return lq
}

// Queue models a multi-server queue with optional capacity (M/M/c/K).
type Queue struct {
	Name string

	// --- Configuration ---
	ArrivalRate    float64 // λ (lambda)
	AvgServiceTime float64 // Ts = 1/μ
	Servers        uint    // c (Number of parallel servers)
	Capacity       uint    // K (Max items IN SYSTEM: queue + servers), 0 for infinite

	// --- Derived Values (Calculated in Init) ---
	serviceRate  float64 // μ (mu) = 1 / Ts
	offeredLoad  float64 // a = λ / μ
	utilization  float64 // ρ (rho) = a / c
	isStable     bool    // Requires rho < 1 for infinite K, always stable if K is finite
	p0           float64 // Probability system is empty
	pk           float64 // Probability system is full (blocking probability)
	lambdaEff    float64 // Effective arrival rate λ_eff = λ * (1 - pk)
	avgWaitTimeQ float64 // Average waiting time in queue (Wq) for non-blocked items
}

// Init initializes the M/M/c/K Queue component.
func (q *Queue) Init(name string, lambda float64, ts float64, c uint, k uint) *Queue {
	q.Name = name

	// --- Validate Inputs ---
	if lambda <= 0 {
		lambda = 1e-9
	}
	if ts <= 0 {
		ts = 1e-9
	}
	if c == 0 {
		// log.Printf("Error: Queue '%s' initialized with c=0 servers.", name)
		c = 1 // Default to 1 server to avoid errors, log warning
	}
	isBounded := k > 0
	if isBounded && k < c {
		// log.Printf("Warning: Queue '%s' has Capacity K (%d) < Servers c (%d). Effective capacity is c. Blocking is high.", name, k, c)
		// Effective capacity is 'c', blocking calculations need care.
		// For simplicity, we might assume K=c in this edge case for formulas? Or handle carefully.
		// Let's proceed but note formulas might be approximate here. Or should K just be >= c?
		// For now, let K be K, formulas handle K<c somewhat (e.g., sum2 range becomes empty).
	}

	// --- Basic Calculations ---
	q.ArrivalRate = lambda
	q.AvgServiceTime = ts
	q.Servers = c
	q.Capacity = k // 0 means infinite K for calculations
	q.serviceRate = 1.0 / q.AvgServiceTime
	q.offeredLoad = q.ArrivalRate / q.serviceRate      // a = lambda / mu
	q.utilization = q.offeredLoad / float64(q.Servers) // rho = a / c

	// --- Stability ---
	q.isStable = q.utilization < 1.0
	// Note: Finite K queues are always "stable" in the sense they don't grow infinitely,
	// but high rho still means high blocking and waiting times. The stability check
	// mainly matters for Wq calculations when K is infinite.

	// --- Calculate M/M/c/K Probabilities and Metrics ---
	// Use a very large number for K if unbounded (0) for calculations
	calcK := float64(k)
	if !isBounded {
		// Need a large K for approximation, but too large causes Pow overflow
		// Choose K large enough relative to c and rho. Heuristic:
		calcK = float64(c) + 20.0/math.Max(1e-6, (1.0-q.utilization)) // Needs ~20/(1-rho) slots past c for stability?
		if calcK > 10000 {
			calcK = 10000
		} // Limit to avoid overflow
		if calcK < float64(c+5) {
			calcK = float64(c + 5)
		} // Ensure minimum reasonable K
	}
	uintK := uint(math.Ceil(calcK)) // Use this K for P0, Pk, Lq calculations

	q.p0 = calculateP0(q.offeredLoad, q.Servers, uintK, q.utilization)
	q.pk = calculatePk(q.p0, q.offeredLoad, q.Servers, uintK, q.utilization)

	// For infinite queue (k=0), blocking prob is 0 if stable, 1 if unstable
	if !isBounded {
		if q.isStable {
			q.pk = 0.0
		} else {
			q.pk = 1.0
		}
	}

	q.lambdaEff = q.ArrivalRate * (1.0 - q.pk)

	lq := 0.0
	// Calculate Lq only if stable or bounded (otherwise infinite)
	if q.isStable || isBounded {
		lq = calculateLq(q.p0, q.offeredLoad, q.Servers, uintK, q.utilization)
	} else {
		lq = math.Inf(1) // Unstable unbounded queue
	}

	// Calculate Wq using Little's Law (Wq = Lq / lambda_eff)
	if q.lambdaEff > 1e-9 && !math.IsInf(lq, 0) {
		q.avgWaitTimeQ = lq / q.lambdaEff
	} else if math.IsInf(lq, 1) {
		q.avgWaitTimeQ = math.Inf(1) // Propagate infinite wait time
	} else {
		q.avgWaitTimeQ = 0 // No waiting if no effective arrivals or zero queue length
	}
	if q.avgWaitTimeQ < 0 {
		q.avgWaitTimeQ = 0
	} // Ensure non-negative

	// log.Printf("Queue '%s' Init: lambda=%.2f, Ts=%.4f, c=%d, k=%d(eff:%d), mu=%.2f, a=%.3f, rho=%.3f, P0=%.4f, PK=%.4f, Lq=%.3f, lambdaEff=%.3f, Wq=%.4f",
	//      q.Name, q.ArrivalRate, q.AvgServiceTime, q.Servers, q.Capacity, uintK, q.serviceRate, q.offeredLoad, q.utilization, q.p0, q.pk, lq, q.lambdaEff, q.avgWaitTimeQ)

	return q
}

// NewMMCKQueue creates and initializes a new M/M/c/K Queue component.
func NewMMCKQueue(name string, arrivalRate float64, avgServiceTime float64, servers uint, capacity uint) *Queue {
	q := &Queue{}
	return q.Init(name, arrivalRate, avgServiceTime, servers, capacity)
}

// Enqueue simulates adding an item to the queue.
// Returns Success=true if item is accepted, Success=false if blocked (for K>0).
func (q *Queue) Enqueue() *Outcomes[sc.AccessResult] {
	outcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
	enqueueLatency := Nanos(10) // Assume near-instantaneous logic

	blockingProb := q.pk // Use pre-calculated blocking probability

	acceptProb := 1.0 - blockingProb
	if acceptProb > 1e-9 {
		outcomes.Add(acceptProb, sc.AccessResult{Success: true, Latency: enqueueLatency})
	}
	if blockingProb > 1e-9 {
		outcomes.Add(blockingProb, sc.AccessResult{Success: false, Latency: enqueueLatency}) // Blocked is immediate failure
	}
	if outcomes.Len() == 0 { // Handle case where probabilities might be exactly 0 or 1
		if acceptProb >= 0.5 {
			outcomes.Add(1.0, sc.AccessResult{Success: true, Latency: enqueueLatency})
		} else {
			outcomes.Add(1.0, sc.AccessResult{Success: false, Latency: enqueueLatency})
		}
	}

	return outcomes
}

// Dequeue simulates removing an item from the queue and returns the *waiting time* spent in the queue.
// Uses M/M/c/K analytical approximation for average wait time (Wq) and approximates
// the distribution using percentiles of an exponential distribution scaled by Wq.
func (q *Queue) Dequeue() *Outcomes[Duration] {
	outcomes := &Outcomes[Duration]{
		And: func(a, b Duration) Duration { return a + b },
	}

	avgWaitTime := q.avgWaitTimeQ // Use pre-calculated average wait time

	if math.IsInf(avgWaitTime, 1) || avgWaitTime > 3600.0*24 { // Treat very large waits as "infinite"
		infiniteWait := 3600.0 * 24
		outcomes.Add(1.0, infiniteWait)
		// log.Printf("Queue '%s': Dequeue returning infinite wait (Wq=%.3f)", q.Name, avgWaitTime)
		return outcomes
	}

	if avgWaitTime < 1e-9 { // Essentially zero wait time
		outcomes.Add(1.0, 0.0)
		return outcomes
	}

	// Approximate distribution using exponential percentiles scaled by calculated Wq
	// (Same logic as before, but now uses the M/M/c/K Wq)
	numBuckets := 5
	totalProb := 1.0
	percentiles := []float64{0.10, 0.30, 0.50, 0.70, 0.90}
	bucketWeights := []float64{0.20, 0.20, 0.20, 0.20, 0.20}

	if len(percentiles) != numBuckets || len(bucketWeights) != numBuckets {
		panic(fmt.Sprintf("Queue '%s': Mismatch in percentile/weight array lengths for Dequeue", q.Name))
	}

	for i := 0; i < numBuckets; i++ {
		p := percentiles[i]
		waitTime := 0.0
		// Use average wait time Wq here
		if p < 0.999999 && avgWaitTime > 1e-12 { // Avoid log(0) and ensure Wq is positive
			waitTime = -avgWaitTime * math.Log(1.0-p)
		} else if p >= 0.999999 {
			waitTime = avgWaitTime * 5 // Approximation for P99+
		} // else waitTime remains 0

		if waitTime < 0 {
			waitTime = 0
		}
		outcomes.Add(bucketWeights[i]*totalProb, waitTime)
	}

	return outcomes
}
