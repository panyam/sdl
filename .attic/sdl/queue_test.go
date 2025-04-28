package sdl

import (
	"math"
	"testing"
)

// --- Keep approxEqualTest if not globally available ---
// func approxEqualTest(a, b, tolerance float64) bool { ... }

func TestMMCKQueue_Init_Params(t *testing.T) {
	// Test basic parameter assignment
	q := NewMMCKQueue("TestQ", 10, 0.05, 2, 10) // lambda=10, Ts=0.05(mu=20), c=2, K=10
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
	// rho = lambda / (c*mu) = 10 / (2*20) = 10 / 40 = 0.25
	if !approxEqualTest(q.utilization, 0.25, 1e-9) {
		t.Errorf("Rho mismatch: exp 0.25, got %.3f", q.utilization)
	}
	// Check if internal calcs ran (spot check Pk)
	// Note: Exact Pk calculation is complex, just check if it's plausible (0 to 1)
	if q.pk < 0 || q.pk > 1.0 {
		t.Errorf("Blocking probability Pk %.4f is out of range [0,1]", q.pk)
	}
	t.Logf("Init Params OK: rho=%.3f, Pk=%.6f, Wq=%.6f", q.utilization, q.pk, q.avgWaitTimeQ)
}

func TestMMCKQueue_Init_Unbounded(t *testing.T) {
	// Test K=0 for unbounded
	q := NewMMCKQueue("TestQ_Unbounded", 8, 0.1, 1, 0) // M/M/1 case (rho=0.8)
	if q.Capacity != 0 {
		t.Errorf("Capacity should be 0 for unbounded")
	}
	if !approxEqualTest(q.utilization, 0.8, 1e-9) {
		t.Errorf("Rho mismatch: exp 0.8, got %.3f", q.utilization)
	}
	if !approxEqualTest(q.pk, 0.0, 1e-9) {
		t.Errorf("Blocking Pk should be 0 for stable unbounded queue, got %.6f", q.pk)
	}

	qUnstable := NewMMCKQueue("TestQ_UnstableUnbounded", 12, 0.1, 1, 0) // M/M/1 unstable (rho=1.2)
	if !approxEqualTest(qUnstable.utilization, 1.2, 1e-9) {
		t.Errorf("Rho mismatch: exp 1.2, got %.3f", qUnstable.utilization)
	}
	if !approxEqualTest(qUnstable.pk, 1.0, 1e-9) {
		t.Errorf("Blocking Pk should be 1.0 for unstable unbounded queue, got %.6f", qUnstable.pk)
	} // Effectively blocks internally
	if !math.IsInf(qUnstable.avgWaitTimeQ, 1) {
		t.Errorf("Wq should be infinite for unstable unbounded queue, got %.4f", qUnstable.avgWaitTimeQ)
	}
}

func TestMMCKQueue_Enqueue_Blocking(t *testing.T) {
	// Setup a queue likely to block: High load, small K relative to c
	// lambda=10, Ts=0.1(mu=10), c=1, K=2. rho=1.0. 'a'=1.0
	// P0 = [ a^0/0! + a^1/1! * (Sum_{n=1 to 2} rho^(n-1)) ]^-1
	//    = [ 1 + 1 * (rho^0 + rho^1) ]^-1 = [ 1 + 1*(1+1) ]^-1 = 1/3
	// PK = P0 * (a^K / (c! * c^(K-c))) = (1/3) * (1^2 / (1! * 1^(2-1))) = (1/3) * 1 = 1/3
	q := NewMMCKQueue("BlockQ", 10.0, 0.1, 1, 2)
	expectedPk := 1.0 / 3.0
	t.Logf("Blocking Queue Check: rho=%.3f, calc Pk=%.6f (Expected ~%.4f)", q.utilization, q.pk, expectedPk)
	// Allow some tolerance due to float calcs or K approx for infinite case in P0 helper
	if !approxEqualTest(q.pk, expectedPk, 0.01) {
		t.Errorf("Blocking probability Pk calculation differs from manual check: got %.6f, expected ~%.4f", q.pk, expectedPk)
	}

	outcomes := q.Enqueue()
	if outcomes == nil || outcomes.Len() == 0 {
		t.Fatal("Enqueue returned nil/empty")
	}

	successW := 0.0
	failureW := 0.0
	for _, b := range outcomes.Buckets {
		if b.Value.Success {
			successW += b.Weight
		} else {
			failureW += b.Weight
		}
	}

	t.Logf("Enqueue Blocking: SuccessRate=%.4f, FailureRate=%.4f", successW, failureW)
	if !approxEqualTest(successW, 1.0-q.pk, 1e-9) {
		t.Errorf("Enqueue success weight %.4f doesn't match expected accept rate %.4f", successW, 1.0-q.pk)
	}
	if !approxEqualTest(failureW, q.pk, 1e-9) {
		t.Errorf("Enqueue failure weight %.4f doesn't match expected blocking rate %.4f", failureW, q.pk)
	}
}

func TestMMCKQueue_Dequeue_WaitTime(t *testing.T) {
	// M/M/2/5 Example: lambda=3, Ts=0.5(mu=2), c=2, K=5
	// a = lambda/mu = 3/2 = 1.5
	// rho = a/c = 1.5/2 = 0.75 (Stable)
	q := NewMMCKQueue("MM2K5Q", 3.0, 0.5, 2, 5)
	t.Logf("M/M/2/5 Queue Check: rho=%.3f, Pk=%.6f, Wq=%.6fs", q.utilization, q.pk, q.avgWaitTimeQ)

	// Manual calc for Wq is very tedious. Let's check plausibility.
	// Wq for M/M/2 (infinite K) = P0*(a^c/(c!*(1-rho)^2))*(1/lambda) * rho ... complex Erlang C based.
	// Wq for M/M/c/K should be LESS than M/M/c (infinite K) due to blocking reducing effective load.
	// Wq for M/M/1 = Ts*rho/(1-rho) = 0.5*0.75/(0.25) = 1.5s (This is NOT correct for MMc)
	// Expect Wq to be positive but likely < Ts.
	if q.avgWaitTimeQ < 0 {
		t.Errorf("Average wait time Wq %.4f should be non-negative", q.avgWaitTimeQ)
	}
	if q.avgWaitTimeQ > q.AvgServiceTime*5 && q.isStable { // Heuristic check
		t.Errorf("Average wait time Wq %.4f seems unexpectedly high for stable queue", q.avgWaitTimeQ)
	}
	// Check stability again
	if q.utilization >= 1.0 && q.Capacity == 0 { // Unbounded AND unstable
		if !math.IsInf(q.avgWaitTimeQ, 1) {
			t.Errorf("Expected infinite Wq for unstable unbounded queue")
		}
	} else if q.utilization >= 1.0 && q.Capacity > 0 { // Bounded but unstable load
		t.Logf("Warning: Queue is bounded but rho >= 1 (%.2f). Expect high wait times and blocking.", q.utilization)
		if math.IsInf(q.avgWaitTimeQ, 1) {
			t.Errorf("Expected FINITE Wq for bounded queue, got Inf")
		}
	}

	outcomes := q.Dequeue()
	if outcomes == nil || outcomes.Len() == 0 {
		t.Fatal("Dequeue returned nil/empty outcomes")
	}

	// Check average wait time from distribution matches calculated avgWaitTimeQ
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

	t.Logf("Dequeue Wait Dist: Calculated Avg=%.6fs (Buckets=%d)", calculatedAvgWq, outcomes.Len())
	if !approxEqualTest(calculatedAvgWq, q.avgWaitTimeQ, q.avgWaitTimeQ*0.3+1e-6) { // Allow 30% tolerance + epsilon
		t.Errorf("Dequeue distribution average %.6f differs significantly from calculated Wq %.6f", calculatedAvgWq, q.avgWaitTimeQ)
	}
}

// Reuse approxEqualTest if not globally available
// func approxEqualTest(a, b, tolerance float64) bool { ... }
