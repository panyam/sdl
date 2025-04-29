package components

import (
	"testing"

	sc "github.com/panyam/leetcoach/sdl/core"
	// Ensure metrics and approxEqualTest are available
)

func TestNetworkLink_Init_Defaults(t *testing.T) {
	// Test with zero values, should apply defaults
	nl := NewNetworkLink(0, 0, 0)

	if nl.BaseLatency <= 0 {
		t.Error("Default BaseLatency should be > 0")
	}
	if nl.MaxJitter < 0 { // Allow zero jitter
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
	baseLat := Millis(50)
	jitter := Millis(5)
	loss := 0.01 // 1% loss

	nl := NewNetworkLink(baseLat, jitter, loss)

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
	baseLat := Millis(20)
	jitter := Millis(4) // Latency range: 16ms to 24ms
	loss := 0.0

	nl := NewNetworkLink(baseLat, jitter, loss)
	outcomes := nl.Transfer()

	if outcomes == nil {
		t.Fatal("Transfer() returned nil outcomes")
	}

	// Calculate metrics
	avail := sc.Availability(outcomes)
	meanLat := sc.MeanLatency(outcomes)
	p0 := sc.PercentileLatency(outcomes, 0.0)   // Should be near Base - Jitter
	p50 := sc.PercentileLatency(outcomes, 0.5)  // Should be near Base
	p100 := sc.PercentileLatency(outcomes, 1.0) // Should be near Base + Jitter

	t.Logf("NoLoss Link:sc.Avail=%.4f,sc.Mean=%.6fs, P0=%.6fs, P50=%.6fs, P100=%.6fs", avail, meanLat, p0, p50, p100)

	// Assertions
	if !approxEqualTest(avail, 1.0, 1e-9) {
		t.Errorf("Availability mismatch: expected 1.0, got %.4f", avail)
	}
	//sc.Mean should be close to BaseLatency for symmetric jitter distribution
	if !approxEqualTest(meanLat, baseLat, baseLat*0.1) { // Allow 10% tolerance for discretization
		t.Errorf("MeanLatency mismatch: expected near %.6f, got %.6f", baseLat, meanLat)
	}
	// Check P0, P50, P100 against expected bounds (allow tolerance for bucket steps)
	latencyStep := (2 * jitter) / float64(nl.LatencyBuckets-1)
	if !approxEqualTest(p0, baseLat-jitter, latencyStep) {
		t.Errorf("P0 Latency mismatch: expected near %.6f, got %.6f", baseLat-jitter, p0)
	}
	if !approxEqualTest(p50, baseLat, latencyStep) {
		t.Errorf("P50 Latency mismatch: expected near %.6f, got %.6f", baseLat, p50)
	}
	// P100 might be slightly less than Base+Jitter depending on bucket count
	if !approxEqualTest(p100, baseLat+jitter, latencyStep) {
		t.Errorf("P100 Latency mismatch: expected near %.6f, got %.6f", baseLat+jitter, p100)
	}
	// Check number of buckets (should be LatencyBuckets if no loss)
	if outcomes.Len() != nl.LatencyBuckets {
		t.Errorf("Expected %d buckets for no loss, got %d", nl.LatencyBuckets, outcomes.Len())
	}
}

func TestNetworkLink_Transfer_Metrics_WithLoss(t *testing.T) {
	baseLat := Millis(100)
	jitter := Millis(20)
	loss := 0.10 // 10% loss

	nl := NewNetworkLink(baseLat, jitter, loss)
	outcomes := nl.Transfer()

	if outcomes == nil {
		t.Fatal("Transfer() returned nil outcomes")
	}

	// Calculate metrics
	avail := sc.Availability(outcomes)
	meanSuccLat := sc.MeanLatency(outcomes) //sc.Mean latency *of successful transfers*

	t.Logf("WithLoss Link:sc.Avail=%.4f,sc.MeanSuccessLatency=%.6fs", avail, meanSuccLat)

	// Assertions
	expectedAvail := 1.0 - loss
	if !approxEqualTest(avail, expectedAvail, 1e-9) {
		t.Errorf("Availability mismatch: expected %.4f, got %.4f", expectedAvail, avail)
	}
	//sc.Mean success latency should still be around BaseLatency
	if !approxEqualTest(meanSuccLat, baseLat, baseLat*0.1) {
		t.Errorf("Mean Success Latency mismatch: expected near %.6f, got %.6f", baseLat, meanSuccLat)
	}
	// Check number of buckets (should be LatencyBuckets + 1 failure bucket)
	if outcomes.Len() != nl.LatencyBuckets+1 {
		t.Errorf("Expected %d success + 1 failure = %d buckets, got %d", nl.LatencyBuckets, nl.LatencyBuckets+1, outcomes.Len())
	}

	// Verify failure bucket weight and latency (optional)
	_, failures := outcomes.Split(sc.AccessResult.IsSuccess)
	if failures == nil || failures.Len() != 1 {
		t.Fatalf("Expected exactly one failure bucket, got %d", failures.Len())
	}
	if !approxEqualTest(failures.Buckets[0].Weight, loss, 1e-9) {
		t.Errorf("Failure bucket weight mismatch: expected %.4f, got %.4f", loss, failures.Buckets[0].Weight)
	}
	expectedFailureLatency := baseLat + jitter + Millis(1) // Matches calculation in code
	if !approxEqualTest(failures.Buckets[0].Value.Latency, expectedFailureLatency, 1e-9) {
		t.Errorf("Failure bucket latency mismatch: expected %.6f, got %.6f", expectedFailureLatency, failures.Buckets[0].Value.Latency)
	}
}
