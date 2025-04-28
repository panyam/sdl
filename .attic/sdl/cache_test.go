package sdl

import (
	"testing"
	// Ensure metrics helpers are accessible
)

func TestCache_Init_Defaults(t *testing.T) {
	c := NewCache()

	if c.HitRate <= 0 || c.HitRate >= 1.0 {
		t.Errorf("Default HitRate invalid: %.2f", c.HitRate)
	}
	if c.HitLatency.Len() == 0 {
		t.Error("Default HitLatency is empty")
	}
	if c.MissLatency.Len() == 0 {
		t.Error("Default MissLatency is empty")
	}
	if c.WriteLatency.Len() == 0 {
		t.Error("Default WriteLatency is empty")
	}
	if c.FailureProb < 0 || c.FailureProb >= 1.0 {
		t.Errorf("Default FailureProb invalid: %.4f", c.FailureProb)
	}
	if c.FailureLatency.Len() == 0 {
		t.Error("Default FailureLatency is empty")
	}

	if c.readOutcomes == nil || c.writeOutcomes == nil {
		t.Fatal("Outcomes not calculated in Init")
	}
	if c.readOutcomes.Len() == 0 || c.writeOutcomes.Len() == 0 {
		t.Fatal("Calculated outcomes are empty")
	}
}

func TestCache_Read_Metrics(t *testing.T) {
	c := NewCache()
	// Use specific values for easier testing
	c.HitRate = 0.75
	c.HitLatency = (&Outcomes[Duration]{}).Add(1.0, Nanos(100))    // Single hit latency 100ns
	c.MissLatency = (&Outcomes[Duration]{}).Add(1.0, Nanos(300))   // Single miss latency 300ns
	c.FailureProb = 0.01                                           // 1% failure
	c.FailureLatency = (&Outcomes[Duration]{}).Add(1.0, Millis(2)) // Single failure latency 2ms
	// Recalculate outcomes after setting params
	c.calculateReadOutcomes()

	outcomes := c.Read()
	if outcomes == nil || outcomes.Len() == 0 {
		t.Fatal("Read() returned nil or empty outcomes")
	}

	// We need to analyze the buckets directly or split carefully
	hits, missesAndFailures := outcomes.Split(AccessResult.IsSuccess)

	totalWeight := outcomes.TotalWeight()
	hitWeight := hits.TotalWeight()
	missFailWeight := missesAndFailures.TotalWeight()

	t.Logf("Cache Read: TotalW=%.4f, HitW=%.4f, MissFailW=%.4f (Buckets=%d)", totalWeight, hitWeight, missFailWeight, outcomes.Len())

	// Expected weights
	expectedFailProb := c.FailureProb
	expectedHitProb := c.HitRate * (1.0 - c.FailureProb)
	expectedMissProb := (1.0 - c.HitRate) * (1.0 - c.FailureProb)

	// Verify total weight sums to ~1.0
	if !approxEqualTest(totalWeight, 1.0, 1e-9) {
		t.Errorf("Total weight mismatch: expected ~1.0, got %.4f", totalWeight)
	}

	// Verify hit weight
	if !approxEqualTest(hitWeight, expectedHitProb, 1e-9) {
		t.Errorf("Hit weight mismatch: expected %.4f, got %.4f", expectedHitProb, hitWeight)
	}

	// Verify combined miss/failure weight
	if !approxEqualTest(missFailWeight, expectedMissProb+expectedFailProb, 1e-9) {
		t.Errorf("Miss+Failure weight mismatch: expected %.4f, got %.4f", expectedMissProb+expectedFailProb, missFailWeight)
	}

	// Check latencies (assuming single latency outcomes for simplicity in test)
	if hits != nil && hits.Len() > 0 {
		meanHitLat := MeanLatency(hits) // MeanLatency works as Success=true here means Hit
		expHitLat := Nanos(100)
		if !approxEqualTest(meanHitLat, expHitLat, 1e-9) {
			t.Errorf("Mean Hit Latency mismatch: expected %.6f, got %.6f", expHitLat, meanHitLat)
		}
	} else if expectedHitProb > 0 {
		t.Error("Expected hit outcomes but found none")
	}

	// How to check miss vs failure in missesAndFailures? Need to look at latency.
	missCount := 0
	failCount := 0
	expMissLat := Nanos(300)
	expFailLat := Millis(2)
	if missesAndFailures != nil {
		for _, b := range missesAndFailures.Buckets {
			if approxEqualTest(b.Value.Latency, expMissLat, 1e-9) {
				missCount++
				// Check weight contribution
				if !approxEqualTest(b.Weight, expectedMissProb, 1e-9) {
					t.Errorf("Miss bucket weight mismatch: exp %.4f, got %.4f", expectedMissProb, b.Weight)
				}
			} else if approxEqualTest(b.Value.Latency, expFailLat, 1e-9) {
				failCount++
				// Check weight contribution
				if !approxEqualTest(b.Weight, expectedFailProb, 1e-9) {
					t.Errorf("Failure bucket weight mismatch: exp %.4f, got %.4f", expectedFailProb, b.Weight)
				}
			} else {
				t.Errorf("Unexpected latency %.6f in miss/failure outcomes", b.Value.Latency)
			}
		}
	}
	if expectedMissProb > 0 && missCount == 0 {
		t.Error("Expected miss outcome but found none")
	}
	if expectedFailProb > 0 && failCount == 0 {
		t.Error("Expected failure outcome but found none")
	}
	if expectedMissProb == 0 && missCount > 0 {
		t.Error("Found unexpected miss outcome")
	}
	if expectedFailProb == 0 && failCount > 0 {
		t.Error("Found unexpected failure outcome")
	}

}

func TestCache_Write_Metrics(t *testing.T) {
	c := NewCache()
	// Use specific values for easier testing
	c.WriteLatency = (&Outcomes[Duration]{}).Add(1.0, Nanos(150))  // Single write latency 150ns
	c.FailureProb = 0.02                                           // 2% failure
	c.FailureLatency = (&Outcomes[Duration]{}).Add(1.0, Millis(3)) // Single failure latency 3ms
	// Recalculate outcomes
	c.calculateWriteOutcomes()

	outcomes := c.Write()
	if outcomes == nil || outcomes.Len() == 0 {
		t.Fatal("Write() returned nil or empty outcomes")
	}

	avail := Availability(outcomes) // Success = successful write to cache
	meanSuccLat := MeanLatency(outcomes)

	t.Logf("Cache Write: Avail=%.4f, MeanSuccessLatency=%.6fs (Buckets: %d)", avail, meanSuccLat, outcomes.Len())

	expectedAvail := 1.0 - c.FailureProb
	if !approxEqualTest(avail, expectedAvail, 1e-9) {
		t.Errorf("Write Availability mismatch: expected %.4f, got %.4f", expectedAvail, avail)
	}

	expectedMeanSuccLat := Nanos(150)
	if !approxEqualTest(meanSuccLat, expectedMeanSuccLat, 1e-9) {
		t.Errorf("Write Mean Success Latency mismatch: expected %.6f, got %.6f", expectedMeanSuccLat, meanSuccLat)
	}
}
