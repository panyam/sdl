package sdl

import (
	"math"
	"testing"
	// Ensure metrics helpers are accessible
)

func TestResourcePool_Init_MMc(t *testing.T) {
	// lambda=8, Ts=0.1(mu=10), c=1. rho=0.8. Stable M/M/1. Wq = Ts*rho/(1-rho)=0.1*0.8/0.2 = 0.4
	rp1 := NewResourcePool("TestPool1", 1, 8.0, 0.1)
	if rp1.Size != 1 {
		t.Errorf("Size mismatch")
	}
	if !rp1.isStable {
		t.Errorf("Pool 1 should be stable (rho=0.8)")
	}
	if !approxEqualTest(rp1.avgWaitTimeQ, 0.4, 0.01) {
		t.Errorf("Pool 1 Wq mismatch: exp ~0.4, got %.4f", rp1.avgWaitTimeQ)
	}

	// lambda=5, Ts=0.1(mu=10), c=1. rho=0.5. Stable M/M/1. Wq = 0.1*0.5/0.5 = 0.1
	rp2 := NewResourcePool("TestPool2", 1, 5.0, 0.1)
	if !rp2.isStable {
		t.Errorf("Pool 2 should be stable (rho=0.5)")
	}
	if !approxEqualTest(rp2.avgWaitTimeQ, 0.1, 0.001) {
		t.Errorf("Pool 2 Wq mismatch: exp ~0.1, got %.4f", rp2.avgWaitTimeQ)
	}

	// lambda=10, Ts=0.1(mu=10), c=1. rho=1.0. Unstable M/M/1. Wq = Inf
	rp3 := NewResourcePool("TestPool3", 1, 10.0, 0.1)
	if rp3.isStable {
		t.Errorf("Pool 3 should be unstable (rho=1.0)")
	}
	if !math.IsInf(rp3.avgWaitTimeQ, 1) {
		t.Errorf("Pool 3 Wq mismatch: exp Inf, got %.4f", rp3.avgWaitTimeQ)
	}

	// lambda=15, Ts=0.1(mu=10), c=2. a=1.5, rho=0.75. Stable M/M/2. Wq should be calculated.
	rp4 := NewResourcePool("TestPool4", 2, 15.0, 0.1)
	t.Logf("M/M/2 Pool: rho=%.3f, Wq=%.6fs", rp4.utilization, rp4.avgWaitTimeQ)
	if !rp4.isStable {
		t.Errorf("Pool 4 should be stable (rho=0.75)")
	}
	if rp4.avgWaitTimeQ < 0 || math.IsInf(rp4.avgWaitTimeQ, 0) || math.IsNaN(rp4.avgWaitTimeQ) {
		t.Errorf("Pool 4 Wq calculation failed or invalid: %.4f", rp4.avgWaitTimeQ)
	}
	// Wq M/M/2 should be < Wq M/M/1 for same offered load per server? Check logic.
	// M/M/1 with Ts=0.1, lambda=7.5 (so rho=0.75) -> Wq = 0.1*0.75/0.25 = 0.3
	// Expect M/M/2 Wq to be less than 0.3
	if rp4.avgWaitTimeQ >= 0.3 {
		t.Errorf("M/M/2 Wq (%.4f) should be less than equivalent M/M/1 Wq (~0.3)", rp4.avgWaitTimeQ)
	}

}

func TestResourcePool_Acquire_WithQueueing(t *testing.T) {
	// Pool Size 1, Stable but high load (rho=0.9) -> Wq=0.9s
	rp := NewResourcePool("QueuePool", 1, 9.0, 0.1)

	// --- Attempt 1: Acquire when empty ---
	rp.Used = 0 // Ensure empty
	out1 := rp.Acquire()
	if out1 == nil || out1.Len() != 1 {
		t.Fatal("Acquire(1) failed to return single outcome")
	}
	res1 := out1.Buckets[0].Value
	if !res1.Success {
		t.Errorf("Acquire(1) should succeed")
	}
	if !approxEqualTest(res1.Latency, 0.0, 1e-9) {
		t.Errorf("Acquire(1) latency should be 0, got %.6f", res1.Latency)
	}
	// Manually update state for next step
	rp.Used++

	// --- Attempt 2: Acquire when full ---
	out2 := rp.Acquire()
	if out2 == nil || out2.Len() == 0 {
		t.Fatal("Acquire(2) failed to return outcomes")
	}
	t.Logf("Acquire(2) when full resulted in %d buckets", out2.Len())

	// Expect multiple buckets representing wait time distribution
	if out2.Len() <= 1 {
		t.Errorf("Acquire(2) should return multiple buckets for wait time dist, got %d", out2.Len())
	}

	// Check results are Success=true but with latency > 0
	calculatedAvgWq := 0.0
	totalW := 0.0
	nonZeroLatency := false
	weightedSumW := 0.0
	for _, b := range out2.Buckets {
		if !b.Value.Success {
			t.Errorf("Acquire(2) outcome should have Success=true (acquired after wait)")
		}
		if b.Value.Latency > 1e-9 {
			nonZeroLatency = true
		}
		totalW += b.Weight
		weightedSumW += b.Weight * b.Value.Latency
	}
	if totalW > 1e-9 {
		calculatedAvgWq = weightedSumW / totalW
	} else {
		t.Error("Zero total weight in outcomes")
	}

	if !nonZeroLatency {
		t.Errorf("Acquire(2) outcomes should have non-zero latency for wait time")
	}

	// Check average wait time matches pool's calculated Wq
	t.Logf("Acquire(2) wait time distribution avg = %.6fs (Expected Wq ~ %.6fs)", calculatedAvgWq, rp.avgWaitTimeQ)
	if !approxEqualTest(calculatedAvgWq, rp.avgWaitTimeQ, rp.avgWaitTimeQ*0.3+1e-6) { // Allow 30% tolerance + epsilon
		t.Errorf("Acquire(2) distribution average %.6f differs significantly from calculated Wq %.6f", calculatedAvgWq, rp.avgWaitTimeQ)
	}
	// State check: Used count still 1 because Acquire doesn't change it
	if rp.Used != 1 {
		t.Errorf("Pool Used count should remain 1, got %d", rp.Used)
	}

	// --- Attempt 3: Acquire when unstable ---
	rpUnstable := NewResourcePool("UnstablePool", 1, 11.0, 0.1) // rho = 1.1
	rpUnstable.Used = 1                                         // Assume full
	out3 := rpUnstable.Acquire()
	if out3 == nil || out3.Len() != 1 {
		t.Fatal("Acquire(Unstable) failed to return single outcome")
	}
	res3 := out3.Buckets[0].Value
	if res3.Success {
		t.Errorf("Acquire(Unstable) should return Success=false (rejection)")
	}
	if !approxEqualTest(res3.Latency, 0.0, 1e-9) {
		t.Errorf("Acquire(Unstable) rejection latency should be 0, got %.6f", res3.Latency)
	}

}

// --- Test Simplified Usage ---
// (Keep TestResourcePool_SimplifiedUsage from previous step - it should still work,
// but now the Acquire mapping might include non-zero latency if the pool was full)
func TestResourcePool_SimplifiedUsage_WithQueueing(t *testing.T) {
	pool := NewResourcePool("TestPool", 1, 9.0, 0.1) // Size 1, High load -> Wq=0.9s
	workOutcomes := (&Outcomes[AccessResult]{}).
		Add(0.9, AccessResult{true, Millis(10)}).
		Add(0.1, AccessResult{false, Millis(5)})
	workFunc := func() *Outcomes[AccessResult] { return workOutcomes }

	// --- Simulate Op 1 (Pool Empty) ---
	acq1Sim := pool.Acquire()
	if acq1Sim.Buckets[0].Value.Latency != 0 {
		t.Error("Op1 Acquire should have 0 latency")
	}
	// Manual state update FOR TEST ONLY
	pool.Used++
	res1 := And(acq1Sim, workFunc(), AndAccessResults)
	expMean1 := 0.0 + MeanLatency(workOutcomes) // Expected mean = acquire lat + work mean
	t.Logf("Op1 Result (AcqLat=0): Mean=%.6fs (Exp~%.6fs)", MeanLatency(res1), expMean1)
	if !approxEqualTest(MeanLatency(res1), expMean1, expMean1*0.01) {
		t.Errorf("Op1 mean mismatch")
	}

	// --- Simulate Op 2 (Pool Full -> Queuing) ---
	acq2Sim := pool.Acquire()
	// Check average latency of acquire outcome
	acq2AvgLat := 0.0
	totalW := 0.0
	weightedSumW := 0.0
	for _, b := range acq2Sim.Buckets {
		totalW += b.Weight
		weightedSumW += b.Weight * b.Value.Latency
	}
	if totalW > 0 {
		acq2AvgLat = weightedSumW / totalW
	}
	t.Logf("Op2 Acquire Avg Latency: %.6fs (Pool Wq=%.6fs)", acq2AvgLat, pool.avgWaitTimeQ)
	if acq2AvgLat < pool.avgWaitTimeQ*0.5 {
		t.Error("Op2 Acquire latency seems too low")
	} // Should be near Wq

	// Combine acquire (with wait time) and work
	res2 := And(acq2Sim, workFunc(), AndAccessResults)
	// Expected mean = Avg Pool Wq + Work Mean
	expMean2 := pool.avgWaitTimeQ + MeanLatency(workOutcomes)
	t.Logf("Op2 Result (AcqLat~Wq): Mean=%.6fs (Exp~%.6fs)", MeanLatency(res2), expMean2)
	// Allow larger tolerance because Wq itself is an average of a distribution
	if !approxEqualTest(MeanLatency(res2), expMean2, expMean2*0.3) {
		t.Errorf("Op2 mean mismatch")
	}

	// State after Op2: Used=1 (because we only incremented manually once)
}

// --- Keep TestResourcePool_Release_Empty ---
func TestResourcePool_Release_Empty(t *testing.T) {
	rp := NewResourcePool("TestPool", 2, 1, 1) // Rates don't matter here
	rp.Release()
	if rp.Used != 0 {
		t.Errorf("Release on empty pool should leave Used=0, got %d", rp.Used)
	}
}

// Reuse approxEqualTest if needed
// func approxEqualTest(a, b, tolerance float64) bool { ... }
