package components

import (
	"math"

	sc "github.com/panyam/sdl/core"
)

// BatchProcessor defines the interface for a component that can process a batch of items.
type BatchProcessor interface {
	ProcessBatch(batchSize uint) *Outcomes[sc.AccessResult]
}

// BatchingPolicy defines the batch formation strategy.
type BatchingPolicy int

const (
	SizeBased BatchingPolicy = iota // Batch when N items arrive
	TimeBased                       // Batch every T seconds
	// Combined                    // TODO: Batch when N items arrive OR T seconds pass
)

// Batcher collects items and processes them based on a policy.
// approximations to model the average waiting time for batch formation.
//
// Limitations:
//   - Wait Time Approximation: Uses analytical formulas (e.g., (N-1)/(2*lambda) for SizeBased,
//     T/2 for TimeBased) to estimate average wait time. Does not capture the full variance
//     or impact of non-Poisson arrival patterns.
//   - Cold Starts: Represents steady-state behavior, doesn't model initially empty batches.
type Batcher struct {
	Name string

	// --- Configuration ---
	Policy      BatchingPolicy // SizeBased or TimeBased
	BatchSize   uint           // N: Target/Max size (used by SizeBased, potentially caps TimeBased avg)
	Timeout     Duration       // T: Time window (used by TimeBased)
	ArrivalRate float64        // λ: Assumed average arrival rate (items/sec)

	Downstream BatchProcessor

	// Internal derived values
	avgWaitTime  float64 // Estimated average time an item waits
	avgBatchSize float64 // Estimated average batch size (esp. for TimeBased)
}

// Init initializes the Batcher component with defaults and calculates derived values.
func (b *Batcher) Init() {
	// Step 1: No embedded components to initialize
	
	// Step 2: Set defaults only for uninitialized fields (zero values)
	if b.BatchSize == 0 {
		b.BatchSize = 1
	}
	if b.Timeout == 0 {
		b.Timeout = 1.0
	}
	if b.ArrivalRate == 0 {
		b.ArrivalRate = 1e-9
	}
	
	// Step 3: Always calculate derived values based on current parameters
	b.calculateMetrics()
	
	// log.Printf("Batcher '%s' Init: Policy=%v, N=%d, T=%.3fs, lambda=%.2f, AvgWait=%.6fs, AvgN=%.2f",
	//      b.Name, b.Policy, b.BatchSize, b.Timeout, b.ArrivalRate, b.avgWaitTime, b.avgBatchSize)
}

// calculateMetrics computes the estimated average values based on current parameters.
// Call this after changing Policy, BatchSize, Timeout, or ArrivalRate.
func (b *Batcher) calculateMetrics() {
	b.avgWaitTime = 0
	b.avgBatchSize = float64(b.BatchSize) // Default for SizeBased

	switch b.Policy {
	case SizeBased:
		if b.ArrivalRate > 1e-9 && b.BatchSize >= 1 {
			b.avgWaitTime = float64(b.BatchSize-1) / (2.0 * b.ArrivalRate)
		}
		// avgBatchSize remains b.BatchSize
	case TimeBased:
		// Average wait time heuristic: T/2
		b.avgWaitTime = b.Timeout / 2.0
		// Average batch size heuristic: lambda * T (capped by BatchSize if treated as max)
		b.avgBatchSize = b.ArrivalRate * b.Timeout
		if b.avgBatchSize < 1.0 {
			b.avgBatchSize = 1.0
		} // Process at least 1 on average if T>0
	}
	if b.avgWaitTime < 0 {
		b.avgWaitTime = 0
	}
}

// NewBatcher creates and initializes a new Batcher component.
func NewBatcher(name string) (out *Batcher) {
	out = &Batcher{Name: name}
	out.Init()
	return
}

// Submit simulates submitting a single item to the batcher.
// Returns the estimated end-to-end outcomes for this *individual* item,
// including waiting time for batch formation and the outcome of the downstream batch processing.
func (b *Batcher) Submit() *Outcomes[sc.AccessResult] {
	// 1. Model Waiting Time for Batch Formation (uses pre-calculated avgWaitTime)
	//    Note: This uses an analytical approximation for the average wait time and
	//    approximates the distribution around that average. It doesn't track
	//    individual arrivals or the current batch fill level.
	waitOutcomes := &Outcomes[Duration]{And: func(a, b Duration) Duration { return a + b }}
	avgWait := b.avgWaitTime
	if avgWait < 1e-9 {
		waitOutcomes.Add(1.0, 0.0) // No wait time
	} else {
		// Use percentile approximation
		numBuckets := 5
		totalProb := 1.0
		percentiles := []float64{0.10, 0.30, 0.50, 0.70, 0.90}
		bucketWeights := []float64{0.20, 0.20, 0.20, 0.20, 0.20}

		for i := range numBuckets {
			p := percentiles[i]
			waitTime := 0.0
			if p < 0.999999 && avgWait > 1e-12 {
				waitTime = -avgWait * math.Log(1.0-p) // Exponential approx
			} else if p >= 0.999999 {
				waitTime = avgWait * 5
			}
			if waitTime < 0 {
				waitTime = 0
			}
			waitOutcomes.Add(bucketWeights[i]*totalProb, waitTime)
		}
	}
	// Map wait time Outcomes[Duration] to Outcomes[sc.AccessResult] (assuming waiting itself succeeds)
	waitAccessOutcomes := sc.Map(waitOutcomes, func(d Duration) sc.AccessResult { return sc.AccessResult{Success: true, Latency: d} })

	// 2. Model Downstream Batch Processing
	// Use the average batch size for the downstream call.
	// Important: Convert avgBatchSize float to uint, ceiling might be appropriate.
	downstreamBatchSize := uint(math.Ceil(b.avgBatchSize))
	if downstreamBatchSize == 0 {
		downstreamBatchSize = 1
	} // Ensure at least 1

	downstreamOutcomes := b.Downstream.ProcessBatch(downstreamBatchSize)
	if downstreamOutcomes == nil {
		return (&Outcomes[sc.AccessResult]{}).Add(1.0, sc.AccessResult{Success: false, Latency: 0})
	}

	// 3. Combine Waiting Time + Downstream Processing Time
	finalOutcomes := sc.And(waitAccessOutcomes, downstreamOutcomes, func(waitRes, downstreamRes sc.AccessResult) sc.AccessResult {
		return sc.AccessResult{Success: downstreamRes.Success, Latency: waitRes.Latency + downstreamRes.Latency}
	})

	// 4. Apply Reduction
	maxLen := 10 // Default max len
	// TODO: Add MaxOutcomeLen to Batcher struct?
	trimmer := sc.TrimToSize(100, maxLen)
	successes, failures := finalOutcomes.Split(sc.AccessResult.IsSuccess)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalTrimmedOutcomes := (&Outcomes[sc.AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalTrimmedOutcomes
}
