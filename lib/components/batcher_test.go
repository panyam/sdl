// components/batcher_test.go
package components

import (
	"fmt" // Added
	"math"
	"testing"

	sc "github.com/panyam/sdl/lib/core"
)

// MockProcessor and NewMockProcessor remain the same...
type MockProcessor struct {
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
	return (&Outcomes[sc.AccessResult]{}).Add(1.0, sc.AccessResult{true, sc.Millis(5)})
}
func NewMockProcessor() *MockProcessor {
	defaultOutcome := (&Outcomes[sc.AccessResult]{}).
		Add(0.98, sc.AccessResult{true, sc.Millis(10)}).
		Add(0.02, sc.AccessResult{false, sc.Millis(5)})
	return &MockProcessor{
		OutcomeMap:     make(map[uint]*Outcomes[sc.AccessResult]),
		DefaultOutcome: defaultOutcome,
	}
}

// Test Init remains the same...
func TestBatcher_Init(t *testing.T) {
	mockProc := NewMockProcessor()
	// Use standardized pattern: configure struct, then call Init()
	bSize := &Batcher{
		Name:        "SizeB",
		Policy:      SizeBased,
		BatchSize:   10,
		ArrivalRate: 50.0,
		Downstream:  mockProc,
	}
	bSize.Init()
	expWaitSize := 0.09
	expNSize := 10.0
	if !approxEqualTest(bSize.avgWaitTime, expWaitSize, 1e-9) {
		t.Errorf("SizeBased AvgWait mismatch: exp %.4f, got %.4f", expWaitSize, bSize.avgWaitTime)
	}
	if !approxEqualTest(bSize.avgBatchSize, expNSize, 1e-9) {
		t.Errorf("SizeBased AvgN mismatch: exp %.1f, got %.2f", expNSize, bSize.avgBatchSize)
	}
	bTime := &Batcher{
		Name:        "TimeB",
		Policy:      TimeBased,
		Timeout:     sc.Millis(200),
		ArrivalRate: 60.0,
		Downstream:  mockProc,
	}
	bTime.Init()
	expWaitTime := sc.Millis(200) / 2.0
	expNTime := 60.0 * sc.Millis(200)
	if !approxEqualTest(bTime.avgWaitTime, expWaitTime, 1e-9) {
		t.Errorf("TimeBased AvgWait mismatch: exp %.4f, got %.4f", expWaitTime, bTime.avgWaitTime)
	}
	if !approxEqualTest(bTime.avgBatchSize, expNTime, 1e-9) {
		t.Errorf("TimeBased AvgN mismatch: exp %.1f, got %.2f", expNTime, bTime.avgBatchSize)
	}
	bTimeLow := &Batcher{
		Name:        "TimeBLow",
		Policy:      TimeBased,
		Timeout:     sc.Millis(200),
		ArrivalRate: 2.0,
		Downstream:  mockProc,
	}
	bTimeLow.Init()
	expNTimeLow := 1.0
	if !approxEqualTest(bTimeLow.avgBatchSize, expNTimeLow, 1e-9) {
		t.Errorf("TimeBased LowLambda AvgN mismatch: exp %.1f, got %.2f", expNTimeLow, bTimeLow.avgBatchSize)
	}
}

func TestBatcher_Submit_Metrics_Policies(t *testing.T) {
	mockProc := NewMockProcessor()
	// Setup mock outcomes
	procOutcomeN8 := (&Outcomes[sc.AccessResult]{}).Add(0.95, sc.AccessResult{true, sc.Millis(50)}).Add(0.05, sc.AccessResult{false, sc.Millis(20)})
	mockProc.OutcomeMap[8] = procOutcomeN8
	procOutcomeN12 := (&Outcomes[sc.AccessResult]{}).Add(0.93, sc.AccessResult{true, sc.Millis(70)}).Add(0.07, sc.AccessResult{false, sc.Millis(30)})
	mockProc.OutcomeMap[12] = procOutcomeN12

	// --- Test SizeBased ---
	batchSizeN := uint(8)
	arrivalRateN := 100.0
	bSize := &Batcher{
		Name:        "SizeSubmit",
		Policy:      SizeBased,
		BatchSize:   batchSizeN,
		ArrivalRate: arrivalRateN,
		Downstream:  mockProc,
	}
	bSize.Init()
	submitNOutcomes := bSize.Submit()
	// Manual
	submitNAvail := sc.Availability(submitNOutcomes)
	submitNMean := sc.MeanLatency(submitNOutcomes)
	submitNP99 := sc.PercentileLatency(submitNOutcomes, 0.99)
	t.Logf("Manual Log - Submit SizeBased (N=%d, lambda=%.0f): Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)",
		batchSizeN, arrivalRateN, submitNAvail, submitNMean, submitNP99, submitNOutcomes.Len())
	// Analyze
	expAvailN := sc.Availability(procOutcomeN8)
	expMeanN := bSize.avgWaitTime + sc.MeanLatency(procOutcomeN8)
	submitNExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, expAvailN*0.99),
		sc.ExpectAvailability(sc.LTE, expAvailN*1.01),
		sc.ExpectMeanLatency(sc.GTE, expMeanN*0.8), // Allow wider range due to wait time approx
		sc.ExpectMeanLatency(sc.LTE, expMeanN*1.2),
	}
	submitNAnalysis := sc.Analyze(fmt.Sprintf("Submit SizeBased N=%d", batchSizeN), func() *sc.Outcomes[sc.AccessResult] { return submitNOutcomes }, submitNExpectations...)
	submitNAnalysis.Assert(t)

	// Manual checks
	if !approxEqualTest(submitNAvail, expAvailN, 1e-9) {
		t.Errorf("Manual Check - Size Submit Avail %.4f mismatch: exp %.4f", submitNAvail, expAvailN)
	}
	if !approxEqualTest(submitNMean, expMeanN, expMeanN*0.3) {
		t.Errorf("Manual Check - Size Submit Mean %.6fs mismatch: exp near %.6fs", submitNMean, expMeanN)
	}

	// --- Test TimeBased ---
	timeoutT := sc.Millis(200)
	arrivalRateT := 60.0
	bTime := &Batcher{
		Name:        "TimeSubmit",
		Policy:      TimeBased,
		Timeout:     timeoutT,
		ArrivalRate: arrivalRateT,
		Downstream:  mockProc,
	}
	bTime.Init()
	calculatedAvgN := bTime.avgBatchSize
	expectedDownstreamN := uint(math.Ceil(calculatedAvgN))
	if expectedDownstreamN != 12 {
		t.Fatalf("Test logic error: Expected downstream N=12, but calculated %.1f", calculatedAvgN)
	}
	submitTOutcomes := bTime.Submit()
	// Manual
	submitTAvail := sc.Availability(submitTOutcomes)
	submitTMean := sc.MeanLatency(submitTOutcomes)
	submitTP99 := sc.PercentileLatency(submitTOutcomes, 0.99)
	t.Logf("Manual Log - Submit TimeBased (T=%.3fs, lambda=%.0f, avgN=%.1f -> Use N=%d): Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)",
		timeoutT, arrivalRateT, calculatedAvgN, expectedDownstreamN, submitTAvail, submitTMean, submitTP99, submitTOutcomes.Len())
	// Analyze
	expAvailT := sc.Availability(procOutcomeN12)
	expMeanT := bTime.avgWaitTime + sc.MeanLatency(procOutcomeN12)
	submitTExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, expAvailT*0.99),
		sc.ExpectAvailability(sc.LTE, expAvailT*1.01),
		sc.ExpectMeanLatency(sc.GTE, expMeanT*0.8),
		sc.ExpectMeanLatency(sc.LTE, expMeanT*1.2),
	}
	submitTAnalysis := sc.Analyze(fmt.Sprintf("Submit TimeBased T=%.3f", timeoutT), func() *sc.Outcomes[sc.AccessResult] { return submitTOutcomes }, submitTExpectations...)
	submitTAnalysis.Assert(t)

	// Manual checks
	if !approxEqualTest(submitTAvail, expAvailT, 1e-9) {
		t.Errorf("Manual Check - Time Submit Avail %.4f mismatch: exp %.4f", submitTAvail, expAvailT)
	}
	if !approxEqualTest(submitTMean, expMeanT, expMeanT*0.3) {
		t.Errorf("Manual Check - Time Submit Mean %.6fs mismatch: exp near %.6fs", submitTMean, expMeanT)
	}
}
