// components/network_test.go
package components

import (
	// Added
	"testing"

	sc "github.com/panyam/sdl/lib/core"
)

// Tests for Init remain the same...
func TestNetworkLink_Init_Defaults(t *testing.T) {
	nl := NewNetworkLink()
	if nl.BaseLatency <= 0 {
		t.Error("Default BaseLatency should be > 0")
	}
	if nl.MaxJitter < 0 {
		t.Error("Default MaxJitter should be >= 0")
	}
	if nl.PacketLossProb != 0.0 {
		t.Error("Default PacketLossProb should be 0.0")
	}
	if nl.LatencyBuckets <= 0 {
		t.Error("Default LatencyBuckets should be > 0")
	}
	if nl.transferOutcomes == nil {
		t.Fatal("transferOutcomes should be calculated in Init")
	}
}
func TestNetworkLink_Init_Params(t *testing.T) {
	baseLat := sc.Millis(50)
	jitter := sc.Millis(5)
	loss := 0.01
	nl := &NetworkLink{
		BaseLatency:    baseLat,
		MaxJitter:      jitter,
		PacketLossProb: loss,
	}
	nl.Init()
	if nl.BaseLatency != baseLat {
		t.Errorf("BaseLatency mismatch: expected %f, got %f", baseLat, nl.BaseLatency)
	}
	if nl.MaxJitter != jitter {
		t.Errorf("MaxJitter mismatch: expected %f, got %f", jitter, nl.MaxJitter)
	}
	if !approxEqualTest(nl.PacketLossProb, loss, 1e-9) {
		t.Errorf("PacketLossProb mismatch: expected %f, got %f", loss, nl.PacketLossProb)
	}
}

func TestNetworkLink_Transfer_Metrics_NoLoss(t *testing.T) {
	baseLat := sc.Millis(20)
	jitter := sc.Millis(4)
	loss := 0.0
	nl := &NetworkLink{
		BaseLatency:    baseLat,
		MaxJitter:      jitter,
		PacketLossProb: loss,
	}
	nl.Init()

	outcomes := nl.Transfer()
	if outcomes == nil {
		t.Fatal("Transfer() returned nil outcomes")
	}

	// Manual
	avail := sc.Availability(outcomes)
	meanLat := sc.MeanLatency(outcomes)
	p0 := sc.PercentileLatency(outcomes, 0.0)
	p50 := sc.PercentileLatency(outcomes, 0.5)
	p100 := sc.PercentileLatency(outcomes, 1.0)
	t.Logf("Manual Log - NoLoss Link: Avail=%.4f, Mean=%.6fs, P0=%.6fs, P50=%.6fs, P100=%.6fs", avail, meanLat, p0, p50, p100)

	// Analyze
	expMinLat := baseLat - jitter
	expMaxLat := baseLat + jitter
	noLossExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 1.0),
		sc.ExpectMeanLatency(sc.GTE, baseLat*0.9), // Mean close to base
		sc.ExpectMeanLatency(sc.LTE, baseLat*1.1),
		// P50 should be close to base
		sc.ExpectP50(sc.GTE, baseLat*0.9),
		sc.ExpectP50(sc.LTE, baseLat*1.1),
		// We can't easily check P0/P100 with Analyze yet, could add specific metrics later
	}
	noLossAnalysis := sc.Analyze("Transfer NoLoss", func() *sc.Outcomes[sc.AccessResult] { return outcomes }, noLossExpectations...)
	noLossAnalysis.Assert(t)

	// Manual Assertions
	if !approxEqualTest(avail, 1.0, 1e-9) {
		t.Errorf("Manual Check - Availability mismatch: expected 1.0, got %.4f", avail)
	}
	if !approxEqualTest(meanLat, baseLat, baseLat*0.1) {
		t.Errorf("Manual Check - MeanLatency mismatch: expected near %.6f, got %.6f", baseLat, meanLat)
	}
	latencyStep := (2 * jitter) / float64(nl.LatencyBuckets-1)
	if !approxEqualTest(p0, expMinLat, latencyStep) {
		t.Errorf("Manual Check - P0 Latency mismatch: expected near %.6f, got %.6f", expMinLat, p0)
	}
	if !approxEqualTest(p50, baseLat, latencyStep) {
		t.Errorf("Manual Check - P50 Latency mismatch: expected near %.6f, got %.6f", baseLat, p50)
	}
	if !approxEqualTest(p100, expMaxLat, latencyStep) {
		t.Errorf("Manual Check - P100 Latency mismatch: expected near %.6f, got %.6f", expMaxLat, p100)
	}
	if outcomes.Len() != nl.LatencyBuckets {
		t.Errorf("Manual Check - Expected %d buckets for no loss, got %d", nl.LatencyBuckets, outcomes.Len())
	}
}

func TestNetworkLink_Transfer_Metrics_WithLoss(t *testing.T) {
	baseLat := sc.Millis(100)
	jitter := sc.Millis(20)
	loss := 0.10
	nl := &NetworkLink{
		BaseLatency:    baseLat,
		MaxJitter:      jitter,
		PacketLossProb: loss,
	}
	nl.Init()

	outcomes := nl.Transfer()
	if outcomes == nil {
		t.Fatal("Transfer() returned nil outcomes")
	}

	// Manual
	avail := sc.Availability(outcomes)
	meanSuccLat := sc.MeanLatency(outcomes)
	t.Logf("Manual Log - WithLoss Link: Avail=%.4f, MeanSuccessLatency=%.6fs", avail, meanSuccLat)

	// Analyze
	expectedAvail := 1.0 - loss
	withLossExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, expectedAvail*0.99),
		sc.ExpectAvailability(sc.LTE, expectedAvail*1.01),
		sc.ExpectMeanLatency(sc.GTE, baseLat*0.9), // Mean of successes still around base
		sc.ExpectMeanLatency(sc.LTE, baseLat*1.1),
	}
	withLossAnalysis := sc.Analyze("Transfer WithLoss", func() *sc.Outcomes[sc.AccessResult] { return outcomes }, withLossExpectations...)
	withLossAnalysis.Assert(t)

	// Manual Assertions
	if !approxEqualTest(avail, expectedAvail, 1e-9) {
		t.Errorf("Manual Check - Availability mismatch: expected %.4f, got %.4f", expectedAvail, avail)
	}
	if !approxEqualTest(meanSuccLat, baseLat, baseLat*0.1) {
		t.Errorf("Manual Check - Mean Success Latency mismatch: expected near %.6f, got %.6f", baseLat, meanSuccLat)
	}
	if outcomes.Len() != nl.LatencyBuckets+1 {
		t.Errorf("Manual Check - Expected %d success + 1 failure = %d buckets, got %d", nl.LatencyBuckets, nl.LatencyBuckets+1, outcomes.Len())
	}

	// Manual failure bucket check
	_, failures := outcomes.Split(sc.AccessResult.IsSuccess)
	if failures == nil || failures.Len() != 1 {
		t.Fatalf("Manual Check - Expected exactly one failure bucket, got %d", failures.Len())
	}
	if !approxEqualTest(failures.Buckets[0].Weight, loss, 1e-9) {
		t.Errorf("Manual Check - Failure bucket weight mismatch: expected %.4f, got %.4f", loss, failures.Buckets[0].Weight)
	}
	expectedFailureLatency := baseLat + jitter + sc.Millis(1)
	if !approxEqualTest(failures.Buckets[0].Value.Latency, expectedFailureLatency, 1e-9) {
		t.Errorf("Manual Check - Failure bucket latency mismatch: expected %.6f, got %.6f", expectedFailureLatency, failures.Buckets[0].Value.Latency)
	}
}
