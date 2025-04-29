package sdl

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// Reuse approxEqualTest if needed
// func approxEqualTest(a, b, tolerance float64) bool { ... }

func TestConvert_AccessToRanged(t *testing.T) {
	oAcc := (&Outcomes[AccessResult]{And: AndAccessResults}).
		Add(80, AccessResult{true, Millis(10)}). // 0.01
		Add(20, AccessResult{false, Millis(5)})  // 0.005

	rangeFactor := 0.2 // 20% range

	oRng := ConvertToRanged(oAcc, rangeFactor)

	if oRng == nil {
		t.Fatal("ConvertToRanged returned nil")
	}
	if oRng.Len() != oAcc.Len() {
		t.Fatalf("Length mismatch: exp %d, got %d", oAcc.Len(), oRng.Len())
	}

	// Check first bucket (success)
	b0Acc := oAcc.Buckets[0]
	b0Rng := oRng.Buckets[0]
	if b0Rng.Value.Success != b0Acc.Value.Success {
		t.Error("Bucket 0 Success mismatch")
	}
	if !approxEqualTest(b0Rng.Weight, b0Acc.Weight, 1e-9) {
		t.Error("Bucket 0 Weight mismatch")
	}
	expMode0 := Millis(10)
	expMin0 := expMode0 - (expMode0 * rangeFactor / 2.0) // 0.01 - 0.001 = 0.009
	expMax0 := expMode0 + (expMode0 * rangeFactor / 2.0) // 0.01 + 0.001 = 0.011
	if !approxEqualTest(b0Rng.Value.ModeLatency, expMode0, 1e-9) {
		t.Errorf("Bucket 0 Mode mismatch: exp %.6f, got %.6f", expMode0, b0Rng.Value.ModeLatency)
	}
	if !approxEqualTest(b0Rng.Value.MinLatency, expMin0, 1e-9) {
		t.Errorf("Bucket 0 Min mismatch: exp %.6f, got %.6f", expMin0, b0Rng.Value.MinLatency)
	}
	if !approxEqualTest(b0Rng.Value.MaxLatency, expMax0, 1e-9) {
		t.Errorf("Bucket 0 Max mismatch: exp %.6f, got %.6f", expMax0, b0Rng.Value.MaxLatency)
	}

	// Check second bucket (failure)
	b1Acc := oAcc.Buckets[1]
	b1Rng := oRng.Buckets[1]
	if b1Rng.Value.Success != b1Acc.Value.Success {
		t.Error("Bucket 1 Success mismatch")
	}
	if !approxEqualTest(b1Rng.Weight, b1Acc.Weight, 1e-9) {
		t.Error("Bucket 1 Weight mismatch")
	}
	expMode1 := Millis(5)
	expMin1 := expMode1 - (expMode1 * rangeFactor / 2.0) // 0.005 - 0.0005 = 0.0045
	expMax1 := expMode1 + (expMode1 * rangeFactor / 2.0) // 0.005 + 0.0005 = 0.0055
	if !approxEqualTest(b1Rng.Value.ModeLatency, expMode1, 1e-9) {
		t.Errorf("Bucket 1 Mode mismatch: exp %.6f, got %.6f", expMode1, b1Rng.Value.ModeLatency)
	}
	if !approxEqualTest(b1Rng.Value.MinLatency, expMin1, 1e-9) {
		t.Errorf("Bucket 1 Min mismatch: exp %.6f, got %.6f", expMin1, b1Rng.Value.MinLatency)
	}
	if !approxEqualTest(b1Rng.Value.MaxLatency, expMax1, 1e-9) {
		t.Errorf("Bucket 1 Max mismatch: exp %.6f, got %.6f", expMax1, b1Rng.Value.MaxLatency)
	}

	// Test zero range factor
	oRngZero := ConvertToRanged(oAcc, 0)
	if !approxEqualTest(oRngZero.Buckets[0].Value.MinLatency, oRngZero.Buckets[0].Value.ModeLatency, 1e-9) ||
		!approxEqualTest(oRngZero.Buckets[0].Value.MaxLatency, oRngZero.Buckets[0].Value.ModeLatency, 1e-9) {
		t.Error("Zero range factor should result in Min=Mode=Max")
	}

}

func TestConvert_RangedToAccess(t *testing.T) {
	oRng := (&Outcomes[RangedResult]{And: AndRangedResults}).
		Add(70, RangedResult{true, Millis(8), Millis(10), Millis(15)}).
		Add(30, RangedResult{false, Millis(4), Millis(5), Millis(7)})

	oAcc := ConvertToAccess(oRng)

	if oAcc == nil {
		t.Fatal("ConvertToAccess returned nil")
	}
	if oAcc.Len() != oRng.Len() {
		t.Fatalf("Length mismatch: exp %d, got %d", oRng.Len(), oAcc.Len())
	}

	// Check first bucket
	b0Rng := oRng.Buckets[0]
	b0Acc := oAcc.Buckets[0]
	if b0Acc.Value.Success != b0Rng.Value.Success {
		t.Error("Bucket 0 Success mismatch")
	}
	if !approxEqualTest(b0Acc.Weight, b0Rng.Weight, 1e-9) {
		t.Error("Bucket 0 Weight mismatch")
	}
	if !approxEqualTest(b0Acc.Value.Latency, b0Rng.Value.ModeLatency, 1e-9) {
		t.Errorf("Bucket 0 Latency mismatch: exp ModeLatency %.6f, got %.6f", b0Rng.Value.ModeLatency, b0Acc.Value.Latency)
	}

	// Check second bucket
	b1Rng := oRng.Buckets[1]
	b1Acc := oAcc.Buckets[1]
	if b1Acc.Value.Success != b1Rng.Value.Success {
		t.Error("Bucket 1 Success mismatch")
	}
	if !approxEqualTest(b1Acc.Weight, b1Rng.Weight, 1e-9) {
		t.Error("Bucket 1 Weight mismatch")
	}
	if !approxEqualTest(b1Acc.Value.Latency, b1Rng.Value.ModeLatency, 1e-9) {
		t.Errorf("Bucket 1 Latency mismatch: exp ModeLatency %.6f, got %.6f", b1Rng.Value.ModeLatency, b1Acc.Value.Latency)
	}
}

func TestSampleWithinRange(t *testing.T) {
	minL, modeL, maxL := Millis(10), Millis(15), Millis(30) // Asymmetric range
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	numSamples := 10000
	samples := make([]Duration, numSamples)
	sum := 0.0
	countBelowMode := 0

	fmt.Printf("Sampling within range [%.3f, %.3f, %.3f] (ms)\n", minL*1000, modeL*1000, maxL*1000)

	for i := 0; i < numSamples; i++ {
		s := SampleWithinRange(minL, modeL, maxL, rng)
		samples[i] = s
		sum += s
		if s < modeL {
			countBelowMode++
		}
	}

	meanSample := sum / float64(numSamples)
	propBelowMode := float64(countBelowMode) / float64(numSamples)

	// Theoretical threshold for sampling below mode = (mode-min)/(max-min)
	expThreshold := (modeL - minL) / (maxL - minL) // (15-10)/(30-10) = 5/20 = 0.25

	t.Logf("Sampled Mean: %.6fs, Proportion < Mode: %.4f (Expected Threshold ~%.4f)", meanSample, propBelowMode, expThreshold)

	// Mean of triangular distribution = (min + mode + max) / 3
	expMean := (minL + modeL + maxL) / 3.0
	if !approxEqualTest(meanSample, expMean, expMean*0.1) { // Allow 10% tolerance for sampling noise
		t.Errorf("Sampled mean %.6f differs significantly from expected triangular mean %.6f", meanSample, expMean)
	}

	// Proportion below mode should be near the calculated threshold
	if !approxEqualTest(propBelowMode, expThreshold, 0.05) { // Allow 5% tolerance
		t.Errorf("Proportion below mode %.4f differs significantly from expected threshold %.4f", propBelowMode, expThreshold)
	}

	// Test zero range
	sZero := SampleWithinRange(Millis(10), Millis(10), Millis(10), rng)
	if !approxEqualTest(sZero, Millis(10), 1e-9) {
		t.Error("Sampling zero range failed")
	}
}
