// components/queue_test.go
package components

import (
	"fmt" // Added
	"math"
	"testing"

	sc "github.com/panyam/sdl/core" // Added alias
)

// Tests for Init remain the same...
func TestMMCKQueue_Init_Params(t *testing.T) {
	q := NewMMCKQueue("TestQ", 10, 0.05, 2, 10)
	if q.Servers != 2 {
		t.Errorf("Servers mismatch")
	}
	if q.Capacity != 10 {
		t.Errorf("Capacity mismatch")
	}
	if !approxEqualTest(q.ArrivalRate, 10.0, 1e-9) {
		t.Errorf("Lambda mismatch")
	}
	if !approxEqualTest(q.AvgServiceTime, 0.05, 1e-9) {
		t.Errorf("Ts mismatch")
	}
	if !approxEqualTest(q.serviceRate, 20.0, 1e-9) {
		t.Errorf("Mu mismatch")
	}
	if !approxEqualTest(q.utilization, 0.25, 1e-9) {
		t.Errorf("Rho mismatch: exp 0.25, got %.3f", q.utilization)
	}
	if q.pk < 0 || q.pk > 1.0 {
		t.Errorf("Blocking probability Pk %.4f is out of range [0,1]", q.pk)
	}
	t.Logf("Init Params OK: rho=%.3f, Pk=%.6f, Wq=%.6f", q.utilization, q.pk, q.avgWaitTimeQ)
}
func TestMMCKQueue_Init_Unbounded(t *testing.T) {
	q := NewMMCKQueue("TestQ_Unbounded", 8, 0.1, 1, 0)
	if q.Capacity != 0 {
		t.Errorf("Capacity should be 0 for unbounded")
	}
	if !approxEqualTest(q.utilization, 0.8, 1e-9) {
		t.Errorf("Rho mismatch: exp 0.8, got %.3f", q.utilization)
	}
	if !approxEqualTest(q.pk, 0.0, 1e-9) {
		t.Errorf("Blocking Pk should be 0 for stable unbounded queue, got %.6f", q.pk)
	}
	qUnstable := NewMMCKQueue("TestQ_UnstableUnbounded", 12, 0.1, 1, 0)
	if !approxEqualTest(qUnstable.utilization, 1.2, 1e-9) {
		t.Errorf("Rho mismatch: exp 1.2, got %.3f", qUnstable.utilization)
	}
	if !approxEqualTest(qUnstable.pk, 1.0, 1e-9) {
		t.Errorf("Blocking Pk should be 1.0 for unstable unbounded queue, got %.6f", qUnstable.pk)
	}
	if !math.IsInf(qUnstable.avgWaitTimeQ, 1) {
		t.Errorf("Wq should be infinite for unstable unbounded queue, got %.4f", qUnstable.avgWaitTimeQ)
	}
}

func TestMMCKQueue_Enqueue_Blocking(t *testing.T) {
	// Setup queue
	q := NewMMCKQueue("BlockQ", 10.0, 0.1, 1, 2)
	expectedPk := 1.0 / 3.0
	t.Logf("Blocking Queue Check: rho=%.3f, calc Pk=%.6f (Expected ~%.4f)", q.utilization, q.pk, expectedPk)
	if !approxEqualTest(q.pk, expectedPk, 0.01) {
		t.Errorf("Blocking probability Pk calculation differs from manual check: got %.6f, expected ~%.4f", q.pk, expectedPk)
	}

	enqueueOutcomes := q.Enqueue()
	if enqueueOutcomes == nil || enqueueOutcomes.Len() == 0 {
		t.Fatal("Enqueue returned nil/empty")
	}

	// Manual check
	successW := 0.0
	failureW := 0.0
	for _, b := range enqueueOutcomes.Buckets {
		if b.Value.Success {
			successW += b.Weight
		} else {
			failureW += b.Weight
		}
	}
	t.Logf("Manual Log - Enqueue Blocking: SuccessRate=%.4f, FailureRate=%.4f", successW, failureW)

	// Analyze
	enqueueExpectations := []sc.Expectation{
		// Expect availability to be 1 - Pk
		sc.ExpectAvailability(sc.GTE, (1.0-q.pk)*0.99),
		sc.ExpectAvailability(sc.LTE, (1.0-q.pk)*1.01),
		// Expect MeanLatency to be very small (just the enqueue logic)
		sc.ExpectMeanLatency(sc.LT, sc.Micros(1)),
	}
	enqueueAnalysis := sc.Analyze(fmt.Sprintf("Enqueue Blocking Pk=%.3f", q.pk), func() *sc.Outcomes[sc.AccessResult] { return enqueueOutcomes }, enqueueExpectations...)
	enqueueAnalysis.Assert(t)

	// Manual assertions
	if !approxEqualTest(successW, 1.0-q.pk, 1e-9) {
		t.Errorf("Manual Check - Enqueue success weight %.4f doesn't match expected accept rate %.4f", successW, 1.0-q.pk)
	}
	if !approxEqualTest(failureW, q.pk, 1e-9) {
		t.Errorf("Manual Check - Enqueue failure weight %.4f doesn't match expected blocking rate %.4f", failureW, q.pk)
	}
}

// Dequeue test remains unchanged as it returns Outcomes[Duration]
func TestMMCKQueue_Dequeue_WaitTime(t *testing.T) {
	q := NewMMCKQueue("MM2K5Q", 3.0, 0.5, 2, 5)
	t.Logf("M/M/2/5 Queue Check: rho=%.3f, Pk=%.6f, Wq=%.6fs", q.utilization, q.pk, q.avgWaitTimeQ)
	if q.avgWaitTimeQ < 0 {
		t.Errorf("Average wait time Wq %.4f should be non-negative", q.avgWaitTimeQ)
	}
	if q.avgWaitTimeQ > q.AvgServiceTime*5 && q.isStable {
		t.Errorf("Average wait time Wq %.4f seems unexpectedly high for stable queue", q.avgWaitTimeQ)
	}
	if q.utilization >= 1.0 && q.Capacity == 0 {
		if !math.IsInf(q.avgWaitTimeQ, 1) {
			t.Errorf("Expected infinite Wq for unstable unbounded queue")
		}
	} else if q.utilization >= 1.0 && q.Capacity > 0 {
		t.Logf("Warning: Queue is bounded but rho >= 1 (%.2f). Expect high wait times and blocking.", q.utilization)
		if math.IsInf(q.avgWaitTimeQ, 1) {
			t.Errorf("Expected FINITE Wq for bounded queue, got Inf")
		}
	}

	outcomes := q.Dequeue() // Returns Outcomes[Duration]
	if outcomes == nil || outcomes.Len() == 0 {
		t.Fatal("Dequeue returned nil/empty outcomes")
	}
	// Manual calculation for duration outcomes
	totalW := 0.0
	weightedSumW := 0.0
	for _, b := range outcomes.Buckets {
		totalW += b.Weight
		weightedSumW += b.Weight * b.Value
	}
	calculatedAvgWq := 0.0
	if totalW > 1e-9 {
		calculatedAvgWq = weightedSumW / totalW
	}
	t.Logf("Manual Log - Dequeue Wait Dist: Calculated Avg=%.6fs (Buckets=%d)", calculatedAvgWq, outcomes.Len())
	if !approxEqualTest(calculatedAvgWq, q.avgWaitTimeQ, q.avgWaitTimeQ*0.3+1e-6) {
		t.Errorf("Manual Check - Dequeue distribution average %.6f differs significantly from calculated Wq %.6f", calculatedAvgWq, q.avgWaitTimeQ)
	}
}

// Need approxEqualTest if not accessible globally
// func approxEqualTest(a, b, tolerance float64) bool { ... }
