package sdl

import (
	"testing"
	// Ensure metrics helpers are accessible
)

func TestMM1Queue_Init_Stable(t *testing.T) {
	// lambda < mu (arrival rate < service rate)
	q := NewMM1Queue("StableQ", 9.0, 0.1) // 9 items/sec arrive, 0.1 sec/item service (mu=10)
	if !q.isStable {
		t.Error("Queue should be stable (rho < 1)")
	}
	if !approxEqualTest(q.utilization, 0.9, 1e-9) {
		t.Errorf("Expected rho=0.9, got %.3f", q.utilization)
	}
}

func TestMM1Queue_Init_Unstable(t *testing.T) {
	// lambda = mu
	qEq := NewMM1Queue("UnstableQ1", 10.0, 0.1) // lambda=10, mu=10
	if qEq.isStable {
		t.Error("Queue should be unstable (rho = 1)")
	}
	if !approxEqualTest(qEq.utilization, 1.0, 1e-9) {
		t.Errorf("Expected rho=1.0, got %.3f", qEq.utilization)
	}

	// lambda > mu
	qGt := NewMM1Queue("UnstableQ2", 11.0, 0.1) // lambda=11, mu=10
	if qGt.isStable {
		t.Error("Queue should be unstable (rho > 1)")
	}
	if !(qGt.utilization > 1.0) {
		t.Errorf("Expected rho > 1.0, got %.3f", qGt.utilization)
	}
}

func TestMM1Queue_Enqueue(t *testing.T) {
	q := NewMM1Queue("TestQ", 5, 0.1)
	outcomes := q.Enqueue()

	if outcomes == nil || outcomes.Len() != 1 {
		t.Fatal("Enqueue should return single outcome")
	}
	res := outcomes.Buckets[0].Value
	if !res.Success {
		t.Error("Enqueue Success should be true")
	}
	// Check latency is very small
	if res.Latency > Millis(0.1) {
		t.Errorf("Enqueue latency %.6f seems too high", res.Latency)
	}
}

func TestMM1Queue_Dequeue_Stable(t *testing.T) {
	// lambda = 9, Ts = 0.1 => mu = 10, rho = 0.9
	// Wq = Ts * rho / (1 - rho) = 0.1 * 0.9 / (1 - 0.9) = 0.09 / 0.1 = 0.9 seconds
	q := NewMM1Queue("StableQ", 9.0, 0.1)
	outcomes := q.Dequeue()

	if outcomes == nil || outcomes.Len() == 0 {
		t.Fatal("Dequeue returned nil/empty outcomes")
	}

	// Calculate expected average wait time
	expectedAvgWq := 0.9
	// Calculate average from the returned distribution
	totalW := 0.0
	weightedSumW := 0.0
	minW := 99999.0
	maxW := 0.0
	for _, b := range outcomes.Buckets {
		totalW += b.Weight
		weightedSumW += b.Weight * b.Value
		if b.Value < minW {
			minW = b.Value
		}
		if b.Value > maxW {
			maxW = b.Value
		}
	}
	calculatedAvgWq := weightedSumW / totalW

	t.Logf("Stable Dequeue: Expected Wq=%.3fs, Calculated Avg Wq=%.3fs (Min=%.3fs, Max=%.3fs, Buckets=%d)",
		expectedAvgWq, calculatedAvgWq, minW, maxW, outcomes.Len())

	// Check if calculated average is close to theoretical M/M/1 average
	if !approxEqualTest(calculatedAvgWq, expectedAvgWq, expectedAvgWq*0.3) { // Allow 30% tolerance due to bucketing
		t.Errorf("Calculated average wait time %.3fs differs significantly from theoretical %.3fs", calculatedAvgWq, expectedAvgWq)
	}
	// Check if bucket count matches expectation
	if outcomes.Len() != 5 {
		t.Errorf("Expected 5 buckets for wait time distribution, got %d", outcomes.Len())
	}
}

func TestMM1Queue_Dequeue_Unstable(t *testing.T) {
	q := NewMM1Queue("UnstableQ", 10.0, 0.1) // rho = 1.0
	outcomes := q.Dequeue()

	if outcomes == nil || outcomes.Len() != 1 {
		t.Fatal("Unstable Dequeue should return single outcome")
	}
	waitTime := outcomes.Buckets[0].Value

	t.Logf("Unstable Dequeue: WaitTime=%.3fs", waitTime)

	// Expect a very large wait time
	if waitTime < 3600.0 { // Less than an hour
		t.Errorf("Unstable queue wait time %.3f is unexpectedly low", waitTime)
	}
}

func TestMM1Queue_Dequeue_LowUtilization(t *testing.T) {
	q := NewMM1Queue("LowUtilQ", 0.1, 0.1)            // rho = 0.01
	if !(q.utilization < 1e-9 || q.utilization > 0) { // Make sure utilization is calculated correctly
		t.Logf("Low Utilization is %.4f", q.utilization) // Log actual value if needed
		// This utilization IS small but > 1e-9, so it WILL go through the bucketing logic.
		// The test assumption that it returns 1 bucket was wrong IF utilization isn't EXACTLY zero.
		// Let's test the *result* instead of the bucket count if utilization is low but non-zero.
	}

	outcomes := q.Dequeue()

	if outcomes == nil || outcomes.Len() == 0 {
		t.Fatal("Low utilization Dequeue returned nil or empty")
	}

	// Calculate average wait time from the outcomes
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

	// Theoretical Wq = Ts * rho / (1 - rho) = 0.1 * 0.01 / (1 - 0.01) = 0.001 / 0.99 = ~0.00101
	expectedAvgWq := q.AvgServiceTime * q.utilization / (1.0 - q.utilization)

	t.Logf("Low Util Dequeue: Expected AvgWq ~ %.6fs, Calculated AvgWq=%.6fs (Buckets: %d)", expectedAvgWq, calculatedAvgWq, outcomes.Len())

	// For very low utilization, the average wait time should be very close to zero.
	if !approxEqualTest(calculatedAvgWq, 0.0, 1e-3) { // Allow up to 1ms tolerance
		t.Errorf("Low utilization queue average wait time %.6fs should be very near zero", calculatedAvgWq)
	}

	// Test the near-zero utilization case specifically
	qZero := NewMM1Queue("ZeroUtilQ", 1e-10, 0.1) // rho near zero
	outcomesZero := qZero.Dequeue()
	if outcomesZero == nil || outcomesZero.Len() != 1 {
		// This case SHOULD hit the optimization and return 1 bucket
		t.Fatalf("Near-zero utilization Dequeue should return single outcome, got %d", outcomesZero.Len())
	}
	waitTimeZero := outcomesZero.Buckets[0].Value
	if !approxEqualTest(waitTimeZero, 0.0, 1e-9) {
		t.Errorf("Near-zero utilization queue wait time %.3f should be zero", waitTimeZero)
	}
}
