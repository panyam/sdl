package components

import (
	"math"
	"testing"

	// Ensure metrics helpers are accessible
	sc "github.com/panyam/leetcoach/sdl/core"
)

// Mock Batch Processor for testing
type MockProcessor struct {
	// Define outcomes based on batch size
	OutcomeMap     map[uint]*Outcomes[sc.AccessResult]
	DefaultOutcome *Outcomes[sc.AccessResult]
}

func (mp *MockProcessor) ProcessBatch(batchSize uint) *Outcomes[sc.AccessResult] {
	if out, ok := mp.OutcomeMap[batchSize]; ok {
		return out
	}
	if mp.DefaultOutcome != nil {
		return mp.DefaultOutcome
	}
	// Default if nothing else matches: simple success
	return (&Outcomes[sc.AccessResult]{}).Add(1.0, sc.AccessResult{true, Millis(5)})
}

func NewMockProcessor() *MockProcessor {
	// Default: Relatively fast processing
	defaultOutcome := (&Outcomes[sc.AccessResult]{}).
		Add(0.98, sc.AccessResult{true, sc.Millis(10)}).
		Add(0.02, sc.AccessResult{false, sc.Millis(5)})
	return &MockProcessor{
		OutcomeMap:     make(map[uint]*Outcomes[sc.AccessResult]),
		DefaultOutcome: defaultOutcome,
	}
}

func TestBatcher_Init(t *testing.T) {
	mockProc := NewMockProcessor()

	// SizeBased
	// func NewBatcher(name string, policy BatchingPolicy, batchSize uint, timeout Duration, arrivalRate float64, downstream BatchProcessor)
	// FIX: Add SizeBased policy, timeout is ignored but required, pass 0? Or a default? Pass 0.
	bSize := NewBatcher("SizeB", SizeBased, 10, 0, 50.0, mockProc)
	expWaitSize := 0.09
	expNSize := 10.0
	if !approxEqualTest(bSize.avgWaitTime, expWaitSize, 1e-9) {
		t.Errorf("SizeBased AvgWait mismatch: exp %.4f, got %.4f", expWaitSize, bSize.avgWaitTime)
	}
	if !approxEqualTest(bSize.avgBatchSize, expNSize, 1e-9) {
		t.Errorf("SizeBased AvgN mismatch: exp %.1f, got %.2f", expNSize, bSize.avgBatchSize)
	}

	// TimeBased
	// FIX: Add TimeBased policy, batchSize is ignored but required? Use 0 or default? Pass 0.
	bTime := NewBatcher("TimeB", TimeBased, 0, sc.Millis(200), 60.0, mockProc)
	expWaitTime := sc.Millis(200) / 2.0 // T/2 = 0.1s
	expNTime := 60.0 * sc.Millis(200)   // lambda * T = 12.0
	if !approxEqualTest(bTime.avgWaitTime, expWaitTime, 1e-9) {
		t.Errorf("TimeBased AvgWait mismatch: exp %.4f, got %.4f", expWaitTime, bTime.avgWaitTime)
	}
	if !approxEqualTest(bTime.avgBatchSize, expNTime, 1e-9) {
		t.Errorf("TimeBased AvgN mismatch: exp %.1f, got %.2f", expNTime, bTime.avgBatchSize)
	}

	// TimeBased with low lambda -> avgN < 1 should become 1
	// FIX: Add TimeBased policy
	bTimeLow := NewBatcher("TimeBLow", TimeBased, 0, sc.Millis(200), 2.0, mockProc)
	expNTimeLow := 1.0 // lambda * T = 0.4 -> avg = 1.0
	// Recalculate based on implementation: avgBatchSize = lambda * T = 0.4, ceil = 1. Correct.
	if !approxEqualTest(bTimeLow.avgBatchSize, expNTimeLow, 1e-9) {
		t.Errorf("TimeBased LowLambda AvgN mismatch: exp %.1f, got %.2f", expNTimeLow, bTimeLow.avgBatchSize)
	}

}

// Update Submit test to potentially test both policies
func TestBatcher_Submit_Metrics_Policies(t *testing.T) {
	mockProc := NewMockProcessor()
	// Outcome for N=8 (used by SizeBased)
	procOutcomeN8 := (&Outcomes[sc.AccessResult]{}).Add(0.95, sc.AccessResult{true, sc.Millis(50)}).Add(0.05, sc.AccessResult{false, sc.Millis(20)})
	mockProc.OutcomeMap[8] = procOutcomeN8
	// Outcome for N=12 (used by TimeBased avg)
	// From prev test: lambda=60, T=0.2 => avgN=12. Need outcome for N=12.
	// Note: Implementation uses Ceil(avgN), so ceil(12.0) = 12.
	procOutcomeN12 := (&Outcomes[sc.AccessResult]{}).Add(0.93, sc.AccessResult{true, sc.Millis(70)}).Add(0.07, sc.AccessResult{false, sc.Millis(30)}) // Slower for larger batch
	mockProc.OutcomeMap[12] = procOutcomeN12

	// --- Test SizeBased ---
	batchSizeN := uint(8)
	arrivalRateN := 100.0
	// FIX: Add SizeBased policy, pass 0 for timeout
	bSize := NewBatcher("SizeSubmit", SizeBased, batchSizeN, 0, arrivalRateN, mockProc)
	submitNOutcomes := bSize.Submit()
	submitNAvail := sc.Availability(submitNOutcomes)
	submitNMean := sc.MeanLatency(submitNOutcomes)
	submitNP99 := sc.PercentileLatency(submitNOutcomes, 0.99)
	t.Logf("Submit SizeBased (N=%d, lambda=%.0f): Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)",
		batchSizeN, arrivalRateN, submitNAvail, submitNMean, submitNP99, submitNOutcomes.Len())
	// Verification (similar to before)
	expAvailN := sc.Availability(procOutcomeN8)
	expMeanN := bSize.avgWaitTime + sc.MeanLatency(procOutcomeN8)
	if !approxEqualTest(submitNAvail, expAvailN, 1e-9) {
		t.Errorf("Size Submit Avail %.4f mismatch: exp %.4f", submitNAvail, expAvailN)
	}
	if !approxEqualTest(submitNMean, expMeanN, expMeanN*0.3) {
		t.Errorf("Size Submit Mean %.6fs mismatch: exp near %.6fs", submitNMean, expMeanN)
	}

	// --- Test TimeBased ---
	timeoutT := sc.Millis(200) // T=0.2s
	arrivalRateT := 60.0       // lambda=60 -> avgN=12
	// FIX: Add TimeBased policy, pass 0 for batchSize
	bTime := NewBatcher("TimeSubmit", TimeBased, 0, timeoutT, arrivalRateT, mockProc)
	// Check avgN calculation inside test for clarity
	calculatedAvgN := bTime.avgBatchSize
	expectedDownstreamN := uint(math.Ceil(calculatedAvgN))
	if expectedDownstreamN != 12 {
		t.Fatalf("Test logic error: Expected downstream N=12, but calculated %.1f", calculatedAvgN)
	}

	submitTOutcomes := bTime.Submit()
	submitTAvail := sc.Availability(submitTOutcomes)
	submitTMean := sc.MeanLatency(submitTOutcomes)
	submitTP99 := sc.PercentileLatency(submitTOutcomes, 0.99)
	t.Logf("Submit TimeBased (T=%.3fs, lambda=%.0f, avgN=%.1f -> Use N=%d): Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)",
		timeoutT, arrivalRateT, calculatedAvgN, expectedDownstreamN, submitTAvail, submitTMean, submitTP99, submitTOutcomes.Len())
	// Verification
	expAvailT := sc.Availability(procOutcomeN12)                   // Use outcome for N=12
	expMeanT := bTime.avgWaitTime + sc.MeanLatency(procOutcomeN12) // Use T/2 wait time + N=12 downstream mean
	if !approxEqualTest(submitTAvail, expAvailT, 1e-9) {
		t.Errorf("Time Submit Avail %.4f mismatch: exp %.4f", submitTAvail, expAvailT)
	}
	if !approxEqualTest(submitTMean, expMeanT, expMeanT*0.3) {
		t.Errorf("Time Submit Mean %.6fs mismatch: exp near %.6fs", submitTMean, expMeanT)
	}
}
