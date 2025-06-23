package components

import (
	"fmt"
	"math"

	sc "github.com/panyam/sdl/core"
	// "log"
)

// --- Helper functions for M/M/c/K formulas ---
// Note: These are now used dynamically by the Queue component.

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
// Its performance characteristics are calculated dynamically based on its configuration.
type Queue struct {
	Name string

	// --- Configuration ---
	ArrivalRate    float64 // λ (lambda)
	AvgServiceTime float64 // Ts = 1/μ
	Servers        uint    // c (Number of parallel servers)
	Capacity       uint    // K (Max items IN SYSTEM: queue + servers), 0 for infinite

	// --- Computed Metrics (for observability) ---
	lastUtilization float64 // Last calculated utilization (ρ)
}

// Init initializes the M/M/c/K Queue component with default configuration.
func (q *Queue) Init() {
	// Step 1: No embedded components to initialize
	
	// Step 2: Set defaults only for uninitialized fields (zero values)
	if q.ArrivalRate == 0 {
		q.ArrivalRate = 1e-9
	}
	if q.AvgServiceTime == 0 {
		q.AvgServiceTime = 1e-9
	}
	if q.Servers == 0 {
		q.Servers = 1
	}
	// Capacity defaults to 0 (infinite) which is the zero value, so no need to check
	
	// Step 3: No derived values to calculate (computed dynamically in methods)
}

// NewMMCKQueue creates and initializes a new M/M/c/K Queue component.
func NewMMCKQueue(name string) *Queue {
	q := &Queue{Name: name}
	q.Init()
	return q
}

// calculateMMCKMetrics performs the dynamic M/M/c/K calculations.
// It returns the blocking probability (Pk) and the average queue wait time (Wq).
func (q *Queue) calculateMMCKMetrics() (float64, float64) {
	isBounded := q.Capacity > 0
	if isBounded && q.Capacity < q.Servers {
		// log.Printf("Warning: Queue '%s' has Capacity K (%d) < Servers c (%d).", q.Name, q.Capacity, q.Servers)
	}
	if q.AvgServiceTime < 1e-12 {
		q.AvgServiceTime = 1e-12
	}

	serviceRate := 1.0 / q.AvgServiceTime
	offeredLoad := q.ArrivalRate / serviceRate
	utilization := offeredLoad / float64(q.Servers)
	q.lastUtilization = utilization // Store for observability
	isStable := utilization < 1.0

	calcK := float64(q.Capacity)
	if !isBounded {
		calcK = float64(q.Servers) + 20.0/math.Max(1e-6, (1.0-utilization))
		if calcK > 10000 {
			calcK = 10000
		}
		if calcK < float64(q.Servers+5) {
			calcK = float64(q.Servers + 5)
		}
	}
	uintK := uint(math.Ceil(calcK))

	p0 := calculateP0(offeredLoad, q.Servers, uintK, utilization)
	pk := calculatePk(p0, offeredLoad, q.Servers, uintK, utilization)

	if !isBounded {
		if isStable {
			pk = 0.0
		} else {
			pk = 1.0
		}
	}

	var avgWaitTimeQ float64
	lambdaEff := q.ArrivalRate * (1.0 - pk)
	lq := 0.0
	if isStable || isBounded {
		lq = calculateLq(p0, offeredLoad, q.Servers, uintK, utilization)
	} else {
		lq = math.Inf(1)
	}

	if lambdaEff > 1e-9 && !math.IsInf(lq, 0) {
		avgWaitTimeQ = lq / lambdaEff
	} else if math.IsInf(lq, 1) {
		avgWaitTimeQ = math.Inf(1)
	} else {
		avgWaitTimeQ = 0
	}
	if avgWaitTimeQ < 0 {
		avgWaitTimeQ = 0
	}

	return pk, avgWaitTimeQ
}

// Enqueue simulates adding an item to the queue.
// Returns Success=true if item is accepted, Success=false if blocked (for K>0).
func (q *Queue) Enqueue() *Outcomes[sc.AccessResult] {
	outcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
	enqueueLatency := Nanos(10)

	blockingProb, _ := q.calculateMMCKMetrics()

	acceptProb := 1.0 - blockingProb
	if acceptProb > 1e-9 {
		outcomes.Add(acceptProb, sc.AccessResult{Success: true, Latency: enqueueLatency})
	}
	if blockingProb > 1e-9 {
		outcomes.Add(blockingProb, sc.AccessResult{Success: false, Latency: enqueueLatency})
	}
	if outcomes.Len() == 0 {
		if acceptProb >= 0.5 {
			outcomes.Add(1.0, sc.AccessResult{Success: true, Latency: enqueueLatency})
		} else {
			outcomes.Add(1.0, sc.AccessResult{Success: false, Latency: enqueueLatency})
		}
	}
	return outcomes
}

// Dequeue simulates removing an item and returns the *waiting time* distribution.
func (q *Queue) Dequeue() *Outcomes[Duration] {
	outcomes := &Outcomes[Duration]{
		And: func(a, b Duration) Duration { return a + b },
	}

	_, avgWaitTime := q.calculateMMCKMetrics()

	if math.IsInf(avgWaitTime, 1) || avgWaitTime > 3600.0*24 {
		infiniteWait := 3600.0 * 24
		outcomes.Add(1.0, infiniteWait)
		return outcomes
	}

	if avgWaitTime < 1e-9 {
		outcomes.Add(1.0, 0.0)
		return outcomes
	}

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
		if p < 0.999999 && avgWaitTime > 1e-12 {
			waitTime = -avgWaitTime * math.Log(1.0-p)
		} else if p >= 0.999999 {
			waitTime = avgWaitTime * 5
		}
		if waitTime < 0 {
			waitTime = 0
		}
		outcomes.Add(bucketWeights[i]*totalProb, waitTime)
	}
	return outcomes
}

// SetArrivalRate sets the arrival rate for a specific method.
// For Queue, we use a single rate since it has one queue.
func (q *Queue) SetArrivalRate(method string, rate float64) error {
	q.ArrivalRate = rate
	return nil
}

// GetArrivalRate returns the arrival rate for a specific method.
func (q *Queue) GetArrivalRate(method string) float64 {
	return q.ArrivalRate
}

// GetTotalArrivalRate returns the total arrival rate.
func (q *Queue) GetTotalArrivalRate() float64 {
	return q.ArrivalRate
}

// GetUtilization returns the current utilization (ρ) of the queue.
// Values close to 1.0 indicate the system is approaching instability.
func (q *Queue) GetUtilization() float64 {
	if q.AvgServiceTime < 1e-12 || q.Servers == 0 {
		return 0
	}
	serviceRate := 1.0 / q.AvgServiceTime
	offeredLoad := q.ArrivalRate / serviceRate
	return offeredLoad / float64(q.Servers)
}

// GetUtilizationInfo implements UtilizationProvider interface
func (q *Queue) GetUtilizationInfo() []UtilizationInfo {
	utilization := q.GetUtilization()
	capacity := float64(q.Servers)
	if q.Capacity > 0 {
		// For bounded queues, effective capacity is the system capacity
		capacity = float64(q.Capacity)
	}
	
	return []UtilizationInfo{
		{
			ResourceName:      "queue",
			ComponentPath:     q.Name,
			Utilization:       utilization,
			Capacity:          capacity,
			CurrentLoad:       q.ArrivalRate,
			IsBottleneck:      true, // Single resource, always the bottleneck
			WarningThreshold:  0.8,
			CriticalThreshold: 0.95,
		},
	}
}
