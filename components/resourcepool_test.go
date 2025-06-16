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
	rp1 := &ResourcePool{
		Name:        "TestPool1",
		Size:        1,
		ArrivalRate: 8.0,
		AvgHoldTime: 0.1,
	}
	rp1.Init()
	
	serviceRate1 := 1.0 / rp1.AvgHoldTime
	utilization1 := rp1.ArrivalRate / (serviceRate1 * float64(rp1.Size))
	isStable1 := utilization1 < 1.0
	_, avgWaitTimeQ1 := rp1.calculateMMCMetrics()
	
	if rp1.Size != 1 {
		t.Errorf("Size mismatch")
	}
	if !isStable1 {
		t.Errorf("Pool 1 should be stable (rho=0.8)")
	}
	if !approxEqualTest(avgWaitTimeQ1, 0.4, 0.01) {
		t.Errorf("Pool 1 Wq mismatch: exp ~0.4, got %.4f", avgWaitTimeQ1)
	}

	// lambda=5, Ts=0.1(mu=10), c=1. rho=0.5. Stable M/M/1. Wq = 0.1*0.5/0.5 = 0.1
	rp2 := NewResourcePool("TestPool2")
	rp2.Size = 1
	rp2.ArrivalRate = 5.0
	rp2.AvgHoldTime = 0.1
	
	serviceRate2 := 1.0 / rp2.AvgHoldTime
	utilization2 := rp2.ArrivalRate / (serviceRate2 * float64(rp2.Size))
	isStable2 := utilization2 < 1.0
	_, avgWaitTimeQ2 := rp2.calculateMMCMetrics()
	
	if !isStable2 {
		t.Errorf("Pool 2 should be stable (rho=0.5)")
	}
	if !approxEqualTest(avgWaitTimeQ2, 0.1, 0.001) {
		t.Errorf("Pool 2 Wq mismatch: exp ~0.1, got %.4f", avgWaitTimeQ2)
	}

	// lambda=10, Ts=0.1(mu=10), c=1. rho=1.0. Unstable M/M/1. Wq = Inf
	rp3 := NewResourcePool("TestPool3")
	rp3.Size = 1
	rp3.ArrivalRate = 10.0
	rp3.AvgHoldTime = 0.1
	
	serviceRate3 := 1.0 / rp3.AvgHoldTime
	utilization3 := rp3.ArrivalRate / (serviceRate3 * float64(rp3.Size))
	isStable3 := utilization3 < 1.0
	_, avgWaitTimeQ3 := rp3.calculateMMCMetrics()
	
	if isStable3 {
		t.Errorf("Pool 3 should be unstable (rho=1.0)")
	}
	if !math.IsInf(avgWaitTimeQ3, 1) {
		t.Errorf("Pool 3 Wq mismatch: exp Inf, got %.4f", avgWaitTimeQ3)
	}

	// lambda=15, Ts=0.1(mu=10), c=2. a=1.5, rho=0.75. Stable M/M/2.
	rp4 := NewResourcePool("TestPool4")
	rp4.Size = 2
	rp4.ArrivalRate = 15.0
	rp4.AvgHoldTime = 0.1
	
	serviceRate4 := 1.0 / rp4.AvgHoldTime
	utilization4 := rp4.ArrivalRate / (serviceRate4 * float64(rp4.Size))
	isStable4 := utilization4 < 1.0
	_, avgWaitTimeQ4 := rp4.calculateMMCMetrics()
	
	t.Logf("M/M/2 Pool: rho=%.3f, Wq=%.6fs", utilization4, avgWaitTimeQ4)
	if !isStable4 {
		t.Errorf("Pool 4 should be stable (rho=0.75)")
	}
	if avgWaitTimeQ4 < 0 || math.IsInf(avgWaitTimeQ4, 0) || math.IsNaN(avgWaitTimeQ4) {
		t.Errorf("Pool 4 Wq calculation failed or invalid: %.4f", avgWaitTimeQ4)
	}
	// M/M/1 with Ts=0.1, lambda=7.5 (so rho=0.75) -> Wq = 0.1*0.75/0.25 = 0.3
	// Expect M/M/2 Wq to be less than 0.3
	if avgWaitTimeQ4 >= 0.3 {
		t.Errorf("M/M/2 Wq (%.4f) should be less than equivalent M/M/1 Wq (~0.3)", avgWaitTimeQ4)
	}
}

// Test Acquire directly based on configured rates
func TestResourcePool_Acquire_Analytical(t *testing.T) {
	// --- Case 1: Underutilized Pool (rho < 1, Wq ~ 0) ---
	rpLow := NewResourcePool("LowLoadPool")
	rpLow.Size = 2
	rpLow.ArrivalRate = 5.0
	rpLow.AvgHoldTime = 0.1 // c=2, lambda=5, Ts=0.1 -> mu=10, a=0.5, rho=0.25
	// Calculate expected Wq for M/M/2, rho=0.25 (will be small but non-zero)
	// This is just for setting expectations, the pool calculates it internally
	_, expectedWqLow := rpLow.calculateMMCMetrics()
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
	rpHigh := NewResourcePool("HighLoadPool")
	rpHigh.Size = 1
	rpHigh.ArrivalRate = 9.0
	rpHigh.AvgHoldTime = 0.1 // c=1, lambda=9, Ts=0.1 -> mu=10, a=0.9, rho=0.9, Wq=0.9
	
	_, avgWaitTimeQHigh := rpHigh.calculateMMCMetrics()
	acqHighOutcomes := rpHigh.Acquire()
	// Analyze
	acqHighExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 1.0),                     // Should succeed after waiting
		sc.ExpectMeanLatency(sc.GTE, avgWaitTimeQHigh*0.7), // Mean latency should be around Wq
		sc.ExpectMeanLatency(sc.LTE, avgWaitTimeQHigh*1.3),
	}
	acqHighAnalysis := sc.Analyze("Acquire High Load (rho=0.9)", func() *sc.Outcomes[sc.AccessResult] { return acqHighOutcomes }, acqHighExpectations...)
	acqHighAnalysis.Assert(t)
	if acqHighOutcomes.Len() <= 1 {
		t.Errorf("High load pool should generate multiple buckets for wait time, got %d", acqHighOutcomes.Len())
	}

	// --- Case 3: Unstable Pool (rho >= 1, Wq = Inf) ---
	rpUnstable := NewResourcePool("UnstablePool")
	rpUnstable.Size = 1
	rpUnstable.ArrivalRate = 11.0
	rpUnstable.AvgHoldTime = 0.1 // rho = 1.1
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
	pool := NewResourcePool("CombinedPool")
	pool.Size = 1
	pool.ArrivalRate = 9.0
	pool.AvgHoldTime = 0.1
	workOutcomes := (&Outcomes[sc.AccessResult]{}).
		Add(0.9, sc.AccessResult{true, sc.Millis(10)}).
		Add(0.1, sc.AccessResult{false, sc.Millis(5)})

	// Simulate Operation: Acquire -> Work
	// Acquire() now returns the distribution including Wq
	acqSim := pool.Acquire()
	resCombined := sc.And(acqSim, workOutcomes, sc.AndAccessResults)

	// Analyze
	_, poolAvgWaitTimeQ := pool.calculateMMCMetrics()
	expMeanCombined := poolAvgWaitTimeQ + sc.MeanLatency(workOutcomes)
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
