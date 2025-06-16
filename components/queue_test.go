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
	q := &Queue{
		Name:           "TestQ",
		ArrivalRate:    10.0,
		AvgServiceTime: 0.05,
		Servers:        2,
		Capacity:       10,
	}
	q.Init()
	
	// Calculate metrics manually since they're computed dynamically
	serviceRate := 1.0 / q.AvgServiceTime
	utilization := q.ArrivalRate / (serviceRate * float64(q.Servers))
	pk, avgWaitTimeQ := q.calculateMMCKMetrics()
	
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
	if !approxEqualTest(serviceRate, 20.0, 1e-9) {
		t.Errorf("Mu mismatch")
	}
	if !approxEqualTest(utilization, 0.25, 1e-9) {
		t.Errorf("Rho mismatch: exp 0.25, got %.3f", utilization)
	}
	if pk < 0 || pk > 1.0 {
		t.Errorf("Blocking probability Pk %.4f is out of range [0,1]", pk)
	}
	t.Logf("Init Params OK: rho=%.3f, Pk=%.6f, Wq=%.6f", utilization, pk, avgWaitTimeQ)
}
func TestMMCKQueue_Init_Unbounded(t *testing.T) {
	q := &Queue{
		Name:           "TestQ_Unbounded",
		ArrivalRate:    8.0,
		AvgServiceTime: 0.1,
		Servers:        1,
		Capacity:       0,
	}
	q.Init()
	
	serviceRate := 1.0 / q.AvgServiceTime
	utilization := q.ArrivalRate / (serviceRate * float64(q.Servers))
	pk, _ := q.calculateMMCKMetrics()
	
	if q.Capacity != 0 {
		t.Errorf("Capacity should be 0 for unbounded")
	}
	if !approxEqualTest(utilization, 0.8, 1e-9) {
		t.Errorf("Rho mismatch: exp 0.8, got %.3f", utilization)
	}
	if !approxEqualTest(pk, 0.0, 1e-9) {
		t.Errorf("Blocking Pk should be 0 for stable unbounded queue, got %.6f", pk)
	}
	
	qUnstable := &Queue{
		Name:           "TestQ_UnstableUnbounded",
		ArrivalRate:    12.0,
		AvgServiceTime: 0.1,
		Servers:        1,
		Capacity:       0,
	}
	qUnstable.Init()
	
	serviceRateUnstable := 1.0 / qUnstable.AvgServiceTime
	utilizationUnstable := qUnstable.ArrivalRate / (serviceRateUnstable * float64(qUnstable.Servers))
	pkUnstable, avgWaitTimeQUnstable := qUnstable.calculateMMCKMetrics()
	
	if !approxEqualTest(utilizationUnstable, 1.2, 1e-9) {
		t.Errorf("Rho mismatch: exp 1.2, got %.3f", utilizationUnstable)
	}
	if !approxEqualTest(pkUnstable, 1.0, 1e-9) {
		t.Errorf("Blocking Pk should be 1.0 for unstable unbounded queue, got %.6f", pkUnstable)
	}
	if !math.IsInf(avgWaitTimeQUnstable, 1) {
		t.Errorf("Wq should be infinite for unstable unbounded queue, got %.4f", avgWaitTimeQUnstable)
	}
}

func TestMMCKQueue_Enqueue_Blocking(t *testing.T) {
	// Setup queue
	q := NewMMCKQueue("BlockQ")
	q.ArrivalRate = 10.0
	q.AvgServiceTime = 0.1
	q.Servers = 1
	q.Capacity = 2
	
	serviceRate := 1.0 / q.AvgServiceTime
	utilization := q.ArrivalRate / (serviceRate * float64(q.Servers))
	pk, _ := q.calculateMMCKMetrics()
	
	expectedPk := 1.0 / 3.0
	t.Logf("Blocking Queue Check: rho=%.3f, calc Pk=%.6f (Expected ~%.4f)", utilization, pk, expectedPk)
	if !approxEqualTest(pk, expectedPk, 0.01) {
		t.Errorf("Blocking probability Pk calculation differs from manual check: got %.6f, expected ~%.4f", pk, expectedPk)
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
		sc.ExpectAvailability(sc.GTE, (1.0-pk)*0.99),
		sc.ExpectAvailability(sc.LTE, (1.0-pk)*1.01),
		// Expect MeanLatency to be very small (just the enqueue logic)
		sc.ExpectMeanLatency(sc.LT, sc.Micros(1)),
	}
	enqueueAnalysis := sc.Analyze(fmt.Sprintf("Enqueue Blocking Pk=%.3f", pk), func() *sc.Outcomes[sc.AccessResult] { return enqueueOutcomes }, enqueueExpectations...)
	enqueueAnalysis.Assert(t)

	// Manual assertions
	if !approxEqualTest(successW, 1.0-pk, 1e-9) {
		t.Errorf("Manual Check - Enqueue success weight %.4f doesn't match expected accept rate %.4f", successW, 1.0-pk)
	}
	if !approxEqualTest(failureW, pk, 1e-9) {
		t.Errorf("Manual Check - Enqueue failure weight %.4f doesn't match expected blocking rate %.4f", failureW, pk)
	}
}

// Dequeue test remains unchanged as it returns Outcomes[Duration]
func TestMMCKQueue_Dequeue_WaitTime(t *testing.T) {
	q := NewMMCKQueue("MM2K5Q")
	q.ArrivalRate = 3.0
	q.AvgServiceTime = 0.5
	q.Servers = 2
	q.Capacity = 5
	
	serviceRate := 1.0 / q.AvgServiceTime
	utilization := q.ArrivalRate / (serviceRate * float64(q.Servers))
	pk, avgWaitTimeQ := q.calculateMMCKMetrics()
	isStable := utilization < 1.0
	
	t.Logf("M/M/2/5 Queue Check: rho=%.3f, Pk=%.6f, Wq=%.6fs", utilization, pk, avgWaitTimeQ)
	if avgWaitTimeQ < 0 {
		t.Errorf("Average wait time Wq %.4f should be non-negative", avgWaitTimeQ)
	}
	if avgWaitTimeQ > q.AvgServiceTime*5 && isStable {
		t.Errorf("Average wait time Wq %.4f seems unexpectedly high for stable queue", avgWaitTimeQ)
	}
	if utilization >= 1.0 && q.Capacity == 0 {
		if !math.IsInf(avgWaitTimeQ, 1) {
			t.Errorf("Expected infinite Wq for unstable unbounded queue")
		}
	} else if utilization >= 1.0 && q.Capacity > 0 {
		t.Logf("Warning: Queue is bounded but rho >= 1 (%.2f). Expect high wait times and blocking.", utilization)
		if math.IsInf(avgWaitTimeQ, 1) {
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
	if !approxEqualTest(calculatedAvgWq, avgWaitTimeQ, avgWaitTimeQ*0.3+1e-6) {
		t.Errorf("Manual Check - Dequeue distribution average %.6f differs significantly from calculated Wq %.6f", calculatedAvgWq, avgWaitTimeQ)
	}
}

// Need approxEqualTest if not accessible globally
// func approxEqualTest(a, b, tolerance float64) bool { ... }
