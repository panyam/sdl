// components/resourcepool_test.go
package components

import (
	// Added
	"math"
	"testing"

	sc "github.com/panyam/leetcoach/sdl/core"
)

// Tests for Init remain the same...
func TestResourcePool_Init_MMc(t *testing.T) {
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
	rp2 := NewResourcePool("TestPool2", 1, 5.0, 0.1)
	if !rp2.isStable {
		t.Errorf("Pool 2 should be stable (rho=0.5)")
	}
	if !approxEqualTest(rp2.avgWaitTimeQ, 0.1, 0.001) {
		t.Errorf("Pool 2 Wq mismatch: exp ~0.1, got %.4f", rp2.avgWaitTimeQ)
	}
	rp3 := NewResourcePool("TestPool3", 1, 10.0, 0.1)
	if rp3.isStable {
		t.Errorf("Pool 3 should be unstable (rho=1.0)")
	}
	if !math.IsInf(rp3.avgWaitTimeQ, 1) {
		t.Errorf("Pool 3 Wq mismatch: exp Inf, got %.4f", rp3.avgWaitTimeQ)
	}
	rp4 := NewResourcePool("TestPool4", 2, 15.0, 0.1)
	t.Logf("M/M/2 Pool: rho=%.3f, Wq=%.6fs", rp4.utilization, rp4.avgWaitTimeQ)
	if !rp4.isStable {
		t.Errorf("Pool 4 should be stable (rho=0.75)")
	}
	if rp4.avgWaitTimeQ < 0 || math.IsInf(rp4.avgWaitTimeQ, 0) || math.IsNaN(rp4.avgWaitTimeQ) {
		t.Errorf("Pool 4 Wq calculation failed or invalid: %.4f", rp4.avgWaitTimeQ)
	}
	if rp4.avgWaitTimeQ >= 0.3 {
		t.Errorf("M/M/2 Wq (%.4f) should be less than equivalent M/M/1 Wq (~0.3)", rp4.avgWaitTimeQ)
	}
}

func TestResourcePool_Acquire_WithQueueing(t *testing.T) {
	rp := NewResourcePool("QueuePool", 1, 9.0, 0.1) // Size 1, High load -> Wq=0.9s

	// --- Attempt 1: Acquire when empty ---
	rp.Used = 0 // Ensure empty
	acq1Outcomes := rp.Acquire()
	if acq1Outcomes == nil || acq1Outcomes.Len() != 1 {
		t.Fatal("Acquire(1) failed to return single outcome")
	}
	// Manual
	res1 := acq1Outcomes.Buckets[0].Value
	t.Logf("Manual Log - Acquire(1): Success=%v, Latency=%.6fs", res1.Success, res1.Latency)
	// Analyze
	acq1Expectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 1.0), // Should succeed
		sc.ExpectMeanLatency(sc.EQ, 0.0),  // Should be immediate
	}
	acq1Analysis := sc.Analyze("Acquire Immediate", func() *sc.Outcomes[sc.AccessResult] { return acq1Outcomes }, acq1Expectations...)
	acq1Analysis.Assert(t)

	// Manual checks
	if !res1.Success {
		t.Errorf("Manual Check - Acquire(1) should succeed")
	}
	if !approxEqualTest(res1.Latency, 0.0, 1e-9) {
		t.Errorf("Manual Check - Acquire(1) latency should be 0, got %.6f", res1.Latency)
	}
	rp.Used++ // Manual state update for next step

	// --- Attempt 2: Acquire when full ---
	acq2Outcomes := rp.Acquire()
	if acq2Outcomes == nil || acq2Outcomes.Len() == 0 {
		t.Fatal("Acquire(2) failed to return outcomes")
	}
	t.Logf("Manual Log - Acquire(2) when full resulted in %d buckets", acq2Outcomes.Len())
	// Manual calculations
	calculatedAvgWq := 0.0
	totalW := 0.0
	nonZeroLatency := false
	weightedSumW := 0.0
	for _, b := range acq2Outcomes.Buckets {
		if !b.Value.Success {
			t.Errorf("Manual Check - Acquire(2) outcome should have Success=true (acquired after wait)")
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
		t.Error("Manual Check - Zero total weight in outcomes")
	}
	t.Logf("Manual Log - Acquire(2) wait time distribution avg = %.6fs (Expected Wq ~ %.6fs)", calculatedAvgWq, rp.avgWaitTimeQ)

	// Analyze
	acq2Expectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 1.0),                 // Still succeeds, just after wait
		sc.ExpectMeanLatency(sc.GTE, rp.avgWaitTimeQ*0.7), // Mean wait should be around Wq
		sc.ExpectMeanLatency(sc.LTE, rp.avgWaitTimeQ*1.3),
	}
	acq2Analysis := sc.Analyze("Acquire Queued", func() *sc.Outcomes[sc.AccessResult] { return acq2Outcomes }, acq2Expectations...)
	acq2Analysis.Assert(t)

	// Manual checks
	if acq2Outcomes.Len() <= 1 {
		t.Errorf("Manual Check - Acquire(2) should return multiple buckets for wait time dist, got %d", acq2Outcomes.Len())
	}
	if !nonZeroLatency {
		t.Errorf("Manual Check - Acquire(2) outcomes should have non-zero latency for wait time")
	}
	if !approxEqualTest(calculatedAvgWq, rp.avgWaitTimeQ, rp.avgWaitTimeQ*0.3+1e-6) {
		t.Errorf("Manual Check - Acquire(2) distribution average %.6f differs significantly from calculated Wq %.6f", calculatedAvgWq, rp.avgWaitTimeQ)
	}
	if rp.Used != 1 {
		t.Errorf("Manual Check - Pool Used count should remain 1, got %d", rp.Used)
	}

	// --- Attempt 3: Acquire when unstable ---
	rpUnstable := NewResourcePool("UnstablePool", 1, 11.0, 0.1)
	rpUnstable.Used = 1 // Assume full
	acq3Outcomes := rpUnstable.Acquire()
	if acq3Outcomes == nil || acq3Outcomes.Len() != 1 {
		t.Fatal("Acquire(Unstable) failed to return single outcome")
	}
	// Manual
	res3 := acq3Outcomes.Buckets[0].Value
	t.Logf("Manual Log - Acquire(Unstable): Success=%v, Latency=%.6fs", res3.Success, res3.Latency)
	// Analyze
	acq3Expectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 0.0), // Should fail (rejected)
		sc.ExpectMeanLatency(sc.EQ, 0.0),  // Failure is immediate
	}
	acq3Analysis := sc.Analyze("Acquire Unstable", func() *sc.Outcomes[sc.AccessResult] { return acq3Outcomes }, acq3Expectations...)
	acq3Analysis.Assert(t)

	// Manual checks
	if res3.Success {
		t.Errorf("Manual Check - Acquire(Unstable) should return Success=false (rejection)")
	}
	if !approxEqualTest(res3.Latency, 0.0, 1e-9) {
		t.Errorf("Manual Check - Acquire(Unstable) rejection latency should be 0, got %.6f", res3.Latency)
	}
}

func TestResourcePool_SimplifiedUsage_WithQueueing(t *testing.T) {
	pool := NewResourcePool("TestPool", 1, 9.0, 0.1)
	workOutcomes := (&Outcomes[sc.AccessResult]{}).
		Add(0.9, sc.AccessResult{true, sc.Millis(10)}).
		Add(0.1, sc.AccessResult{false, sc.Millis(5)})
	workFunc := func() *Outcomes[sc.AccessResult] { return workOutcomes }

	// --- Simulate Op 1 (Pool Empty) ---
	acq1Sim := pool.Acquire()
	if acq1Sim.Buckets[0].Value.Latency != 0 {
		t.Error("Op1 Acquire should have 0 latency")
	}
	pool.Used++ // Manual state update FOR TEST ONLY
	res1 := sc.And(acq1Sim, workFunc(), sc.AndAccessResults)
	// Manual
	expMean1 := 0.0 + sc.MeanLatency(workOutcomes)
	t.Logf("Manual Log - Op1 Result (AcqLat=0): Mean=%.6fs (Exp~%.6fs)", sc.MeanLatency(res1), expMean1)
	// Analyze
	res1Expectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, sc.Availability(workOutcomes)), // Avail dominated by work
		sc.ExpectMeanLatency(sc.GTE, expMean1*0.99),
		sc.ExpectMeanLatency(sc.LTE, expMean1*1.01),
	}
	res1Analysis := sc.Analyze("Combined Op1 (Immediate)", func() *sc.Outcomes[sc.AccessResult] { return res1 }, res1Expectations...)
	res1Analysis.Assert(t)

	// Manual check
	if !approxEqualTest(sc.MeanLatency(res1), expMean1, expMean1*0.01) {
		t.Errorf("Manual Check - Op1 mean mismatch")
	}

	// --- Simulate Op 2 (Pool Full -> Queuing) ---
	acq2Sim := pool.Acquire()
	// Manual calc
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
	t.Logf("Manual Log - Op2 Acquire Avg Latency: %.6fs (Pool Wq=%.6fs)", acq2AvgLat, pool.avgWaitTimeQ)
	if acq2AvgLat < pool.avgWaitTimeQ*0.5 {
		t.Error("Manual Check - Op2 Acquire latency seems too low")
	}
	// Combine
	res2 := sc.And(acq2Sim, workFunc(), sc.AndAccessResults)
	// Manual
	expMean2 := pool.avgWaitTimeQ + sc.MeanLatency(workOutcomes)
	t.Logf("Manual Log - Op2 Result (AcqLat~Wq): Mean=%.6fs (Exp~%.6fs)", sc.MeanLatency(res2), expMean2)
	// Analyze
	res2Expectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, sc.Availability(workOutcomes)), // Avail still dominated by work
		sc.ExpectMeanLatency(sc.GTE, expMean2*0.7),                  // Allow wide range due to Wq approx
		sc.ExpectMeanLatency(sc.LTE, expMean2*1.3),
	}
	res2Analysis := sc.Analyze("Combined Op2 (Queued)", func() *sc.Outcomes[sc.AccessResult] { return res2 }, res2Expectations...)
	res2Analysis.Assert(t)

	// Manual check
	if !approxEqualTest(sc.MeanLatency(res2), expMean2, expMean2*0.3) {
		t.Errorf("Manual Check - Op2 mean mismatch")
	}
}

func TestResourcePool_Release_Empty(t *testing.T) {
	rp := NewResourcePool("TestPool", 2, 1, 1)
	rp.Release()
	if rp.Used != 0 {
		t.Errorf("Release on empty pool should leave Used=0, got %d", rp.Used)
	}
}

// Need approxEqualTest if not accessible globally
// func approxEqualTest(a, b, tolerance float64) bool { ... }
