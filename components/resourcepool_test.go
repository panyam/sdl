// sdl/components/resourcepool_test.go
package components

import (
	"math"
	"testing"

	sc "github.com/panyam/sdl/core"
)

// Tests for Init remain largely the same, checking calculated values
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

	// lambda=15, Ts=0.1(mu=10), c=2. a=1.5, rho=0.75. Stable M/M/2.
	rp4 := NewResourcePool("TestPool4", 2, 15.0, 0.1)
	t.Logf("M/M/2 Pool: rho=%.3f, Wq=%.6fs", rp4.utilization, rp4.avgWaitTimeQ)
	if !rp4.isStable {
		t.Errorf("Pool 4 should be stable (rho=0.75)")
	}
	if rp4.avgWaitTimeQ < 0 || math.IsInf(rp4.avgWaitTimeQ, 0) || math.IsNaN(rp4.avgWaitTimeQ) {
		t.Errorf("Pool 4 Wq calculation failed or invalid: %.4f", rp4.avgWaitTimeQ)
	}
	// M/M/1 with Ts=0.1, lambda=7.5 (so rho=0.75) -> Wq = 0.1*0.75/0.25 = 0.3
	// Expect M/M/2 Wq to be less than 0.3
	if rp4.avgWaitTimeQ >= 0.3 {
		t.Errorf("M/M/2 Wq (%.4f) should be less than equivalent M/M/1 Wq (~0.3)", rp4.avgWaitTimeQ)
	}
}

// Test Acquire directly based on configured rates
func TestResourcePool_Acquire_Analytical(t *testing.T) {
	// --- Case 1: Underutilized Pool (rho < 1, Wq ~ 0) ---
	rpLow := NewResourcePool("LowLoadPool", 2, 5.0, 0.1) // c=2, lambda=5, Ts=0.1 -> mu=10, a=0.5, rho=0.25
	// Calculate expected Wq for M/M/2, rho=0.25 (will be small but non-zero)
	// This is just for setting expectations, the pool calculates it internally
	expectedWqLow := rpLow.avgWaitTimeQ
	acqLowOutcomes := rpLow.Acquire()
	// Analyze
	acqLowExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 1.0), // Should succeed
		// Expect mean latency to be close to the calculated Wq
		sc.ExpectMeanLatency(sc.GTE, expectedWqLow*0.7),
		sc.ExpectMeanLatency(sc.LTE, expectedWqLow*1.3),
	}
	acqLowAnalysis := sc.Analyze("Acquire Low Load (rho=0.25)", func() *sc.Outcomes[sc.AccessResult] { return acqLowOutcomes }, acqLowExpectations...)
	acqLowAnalysis.Assert(t)
	// Check if the mean latency is indeed small, but allow for the calculated Wq
	meanLatencyLow := acqLowAnalysis.Metrics[sc.MeanLatencyMetric]
	t.Logf("Low Load (rho=0.25) calculated Wq: %.6fs, Measured Mean Latency: %.6fs", expectedWqLow, meanLatencyLow)
	if acqLowOutcomes.Len() > 1 && meanLatencyLow > sc.Millis(10) { // Wq should still be small (<10ms?)
		t.Errorf("Low load pool (rho=0.25) Wq seems high: %.6fs", meanLatencyLow)
	}

	// --- Case 2: Utilized Pool (rho < 1, Wq > 0) ---
	rpHigh := NewResourcePool("HighLoadPool", 1, 9.0, 0.1) // c=1, lambda=9, Ts=0.1 -> mu=10, a=0.9, rho=0.9, Wq=0.9
	acqHighOutcomes := rpHigh.Acquire()
	// Analyze
	acqHighExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 1.0),                     // Should succeed after waiting
		sc.ExpectMeanLatency(sc.GTE, rpHigh.avgWaitTimeQ*0.7), // Mean latency should be around Wq
		sc.ExpectMeanLatency(sc.LTE, rpHigh.avgWaitTimeQ*1.3),
	}
	acqHighAnalysis := sc.Analyze("Acquire High Load (rho=0.9)", func() *sc.Outcomes[sc.AccessResult] { return acqHighOutcomes }, acqHighExpectations...)
	acqHighAnalysis.Assert(t)
	if acqHighOutcomes.Len() <= 1 {
		t.Errorf("High load pool should generate multiple buckets for wait time, got %d", acqHighOutcomes.Len())
	}

	// --- Case 3: Unstable Pool (rho >= 1, Wq = Inf) ---
	rpUnstable := NewResourcePool("UnstablePool", 1, 11.0, 0.1) // rho = 1.1
	acqUnstableOutcomes := rpUnstable.Acquire()
	// Analyze
	acqUnstableExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 0.0), // Should fail (rejected)
		// MeanLatency is meaningless for failure case, maybe check P99 is 0? Depends on metrics impl.
	}
	acqUnstableAnalysis := sc.Analyze("Acquire Unstable (rho=1.1)", func() *sc.Outcomes[sc.AccessResult] { return acqUnstableOutcomes }, acqUnstableExpectations...)
	acqUnstableAnalysis.Assert(t) // This will assert that Availability is 0
	if acqUnstableOutcomes.Len() != 1 || acqUnstableOutcomes.Buckets[0].Value.Success {
		t.Errorf("Unstable acquire should result in a single failure outcome")
	}
}

// Test combining Acquire with downstream work
func TestResourcePool_CombinedUsage(t *testing.T) {
	// Use the high load pool from before (rho=0.9, Wq=0.9s)
	pool := NewResourcePool("CombinedPool", 1, 9.0, 0.1)
	workOutcomes := (&Outcomes[sc.AccessResult]{}).
		Add(0.9, sc.AccessResult{true, sc.Millis(10)}).
		Add(0.1, sc.AccessResult{false, sc.Millis(5)})

	// Simulate Operation: Acquire -> Work
	// Acquire() now returns the distribution including Wq
	acqSim := pool.Acquire()
	resCombined := sc.And(acqSim, workOutcomes, sc.AndAccessResults)

	// Analyze
	expMeanCombined := pool.avgWaitTimeQ + sc.MeanLatency(workOutcomes)
	resExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, sc.Availability(workOutcomes)), // Avail dominated by work
		sc.ExpectMeanLatency(sc.GTE, expMeanCombined*0.7),           // Allow wide range due to Wq approx
		sc.ExpectMeanLatency(sc.LTE, expMeanCombined*1.3),
	}
	resAnalysis := sc.Analyze("Combined Acquire+Work (rho=0.9)", func() *sc.Outcomes[sc.AccessResult] { return resCombined }, resExpectations...)
	resAnalysis.Assert(t)

	// Manual check (optional)
	t.Logf("Combined Usage: Expected Mean ~ %.6fs, Actual Mean = %.6fs", expMeanCombined, resAnalysis.Metrics[sc.MeanLatencyMetric])
	if !approxEqualTest(resAnalysis.Metrics[sc.MeanLatencyMetric], expMeanCombined, expMeanCombined*0.3) {
		t.Errorf("Manual Check - Combined mean mismatch (Actual: %.6f vs Expected: %.6f)", resAnalysis.Metrics[sc.MeanLatencyMetric], expMeanCombined)
	}
}
