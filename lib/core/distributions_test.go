// sdl/core/distributions_test.go
package core

import (
	"testing"
)

// Helper to check total weight and specific bucket counts
func assertDistributionProperties(t *testing.T, name string, o *Outcomes[AccessResult], expTotalWeight float64, expSuccessBuckets, expFailBuckets int) {
	t.Helper()
	if o == nil {
		t.Errorf("[%s] Expected non-nil Outcomes, got nil", name)
		return
	}
	totalW := o.TotalWeight()
	if !approxEqualTest(totalW, expTotalWeight, 1e-9) {
		t.Errorf("[%s] Expected TotalWeight %.4f, got %.4f", name, expTotalWeight, totalW)
	}

	successCount := 0
	failCount := 0
	for _, b := range o.Buckets {
		if b.Value.Success {
			successCount++
		} else {
			failCount++
		}
	}

	if successCount != expSuccessBuckets {
		t.Errorf("[%s] Expected %d success buckets, got %d", name, expSuccessBuckets, successCount)
	}
	if failCount != expFailBuckets {
		t.Errorf("[%s] Expected %d failure buckets, got %d", name, expFailBuckets, failCount)
	}
}

func TestNewDistributionFromPercentiles_Basic(t *testing.T) {
	// P0=10, P50=20, P100=50
	// FailRate=0.1, FailLat=5
	// 3 Success Buckets
	percentiles := map[float64]Duration{
		0.0:  Millis(10),
		0.50: Millis(20),
		1.0:  Millis(50),
	}
	failLat := (&Outcomes[Duration]{}).Add(1.0, Millis(5))
	numSuccessBuckets := 3

	o := NewDistributionFromPercentiles(percentiles, 0.1, failLat, numSuccessBuckets)

	assertDistributionProperties(t, "Basic", o, 1.0, 3, 1)

	// Check expected latencies (linear interpolation)
	// Target P for success buckets (centers): 1/6, 3/6 (0.5), 5/6
	// Bucket 1 (P=1/6): Between P0(10) and P0.5(20). Prop = (1/6 - 0) / (0.5 - 0) = 1/3. Lat = 10 + 1/3 * (20-10) = 13.33ms
	// Bucket 2 (P=0.5): Exactly P0.5(20). Lat = 20ms
	// Bucket 3 (P=5/6): Between P0.5(20) and P1.0(50). Prop = (5/6 - 0.5) / (1.0 - 0.5) = (1/3) / (0.5) = 2/3. Lat = 20 + 2/3 * (50-20) = 20 + 20 = 40ms
	expSuccessLats := []Duration{Millis(13.333333), Millis(20), Millis(40)}
	expFailLat := Millis(5)
	successBucketWeight := (1.0 - 0.1) / 3.0
	failBucketWeight := 0.1

	foundSuccessLats := make(map[int]bool)
	foundFailLat := false

	for _, b := range o.Buckets {
		if b.Value.Success {
			if !approxEqualTest(b.Weight, successBucketWeight, 1e-6) {
				t.Errorf("[Basic] Unexpected success bucket weight %.6f, expected %.6f", b.Weight, successBucketWeight)
			}
			matchFound := false
			for i, expLat := range expSuccessLats {
				if approxEqualTest(b.Value.Latency, expLat, Millis(0.1)) { // Allow 0.1ms tolerance
					foundSuccessLats[i] = true
					matchFound = true
					break
				}
			}
			if !matchFound {
				t.Errorf("[Basic] Unexpected success latency %.6fs found", b.Value.Latency)
			}
		} else {
			if !approxEqualTest(b.Weight, failBucketWeight, 1e-6) {
				t.Errorf("[Basic] Unexpected failure bucket weight %.6f, expected %.6f", b.Weight, failBucketWeight)
			}
			if approxEqualTest(b.Value.Latency, expFailLat, 1e-9) {
				foundFailLat = true
			} else {
				t.Errorf("[Basic] Unexpected failure latency %.6fs found, expected %.6fs", b.Value.Latency, expFailLat)
			}
		}
	}

	if len(foundSuccessLats) != len(expSuccessLats) {
		t.Errorf("[Basic] Did not find all expected success latencies. Found: %v", foundSuccessLats)
	}
	if !foundFailLat {
		t.Errorf("[Basic] Did not find expected failure latency.")
	}
}

func TestNewDistributionFromPercentiles_NoFailures(t *testing.T) {
	percentiles := map[float64]Duration{
		0.0:  Millis(1),
		0.9:  Millis(10),
		0.99: Millis(100),
		1.0:  Millis(150),
	}
	numSuccessBuckets := 4

	o := NewDistributionFromPercentiles(percentiles, 0.0, nil, numSuccessBuckets) // failRate = 0

	assertDistributionProperties(t, "NoFailures", o, 1.0, 4, 0)

	// Check metrics (optional)
	p99 := PercentileLatency(o, 0.99)
	// P99 target center is (3+0.5)/4 = 0.875. Interpolate between P0.0(1) and P0.9(10)? No.
	// P99 target center is 0.875. This falls between P0.0 and P0.9. Latency should be low.
	// Let's check the P99 *of the generated distribution*. The last bucket center is P=0.875.
	// The *edge* of the last bucket is P=1.0. P99 should fall within the last bucket.
	// Last bucket center P=0.875 -> Latency = 1 + (0.875-0)/(0.9-0) * (10-1) = 1 + 0.972*9 = 9.75ms approx?
	// Let's re-evaluate the interpolation logic slightly.
	// The function maps target P in [0,1] success space to the input P curve.
	// Bucket 4 center P = 0.875. Interpolates between P0.0(1) and P0.9(10). Lat = ~9.75ms
	// P99 falls within the 4th bucket range [0.75, 1.0]. The latency assigned is the center's.
	// This method means P99 might be underestimated if the tail is steep after the last center point.
	// Let's recalculate P99 from the OUTCOME distribution
	// Weights = 0.25 each. Latencies calculated using centers:
	// B1 P=0.125 -> Lat = 1 + (0.125/0.9)*9 = 2.25
	// B2 P=0.375 -> Lat = 1 + (0.375/0.9)*9 = 4.75
	// B3 P=0.625 -> Lat = 1 + (0.625/0.9)*9 = 7.25
	// B4 P=0.875 -> Lat = 1 + (0.875/0.9)*9 = 9.75
	// Cumulative Weights: 0.25, 0.50, 0.75, 1.00
	// P99 target weight = 0.99. Falls into bucket 4 (cum=1.0). Latency = 9.75ms.
	t.Logf("[NoFailures] P99 of generated dist: %.6fs", p99)
	if !approxEqualTest(p99, Millis(9.75), Millis(0.1)) {
		t.Errorf("[NoFailures] Expected P99 around 9.75ms, got %.6fs", p99)
	}
}

func TestNewDistributionFromPercentiles_MissingBounds(t *testing.T) {
	// Missing P0 and P100
	percentiles := map[float64]Duration{
		0.50: Millis(20),
		0.90: Millis(100),
	}
	numSuccessBuckets := 2

	o := NewDistributionFromPercentiles(percentiles, 0.0, nil, numSuccessBuckets)

	// Expect P0 to use min (20ms), P100 to use max (100ms)
	assertDistributionProperties(t, "MissingBounds", o, 1.0, 2, 0)

	// Bucket centers: P=0.25, P=0.75
	// B1 (P=0.25): Interpolate P0(20) and P0.5(20). Prop=(0.25-0)/(0.5-0)=0.5. Lat=20 + 0.5*(20-20)=20ms
	// B2 (P=0.75): Interpolate P0.5(20) and P0.9(100). Prop=(0.75-0.5)/(0.9-0.5)=0.25/0.4=0.625. Lat=20+0.625*(100-20)=20+50=70ms
	expSuccessLats := []Duration{Millis(20), Millis(70)}
	successBucketWeight := 1.0 / 2.0

	foundSuccessLats := make(map[int]bool)
	for _, b := range o.Buckets {
		if !b.Value.Success {
			t.Errorf("[MissingBounds] Unexpected failure bucket")
			continue
		}
		if !approxEqualTest(b.Weight, successBucketWeight, 1e-6) {
			t.Errorf("[MissingBounds] Unexpected success weight %.6f", b.Weight)
		}
		matchFound := false
		for i, expLat := range expSuccessLats {
			if approxEqualTest(b.Value.Latency, expLat, Millis(0.1)) {
				foundSuccessLats[i] = true
				matchFound = true
				break
			}
		}
		if !matchFound {
			t.Errorf("[MissingBounds] Unexpected success latency %.6fs found", b.Value.Latency)
		}
	}
	if len(foundSuccessLats) != len(expSuccessLats) {
		t.Errorf("[MissingBounds] Did not find all expected success latencies.")
	}
}

func TestNewDistributionFromPercentiles_EdgeCases(t *testing.T) {
	// Only failures
	oFail := NewDistributionFromPercentiles(map[float64]Duration{0.5: 1}, 1.0, nil, 1)
	assertDistributionProperties(t, "OnlyFail", oFail, 1.0, 0, 1)
	if oFail.Buckets[0].Value.Latency != 0 { // Default failure latency
		t.Errorf("[OnlyFail] Expected default 0 failure latency, got %.6f", oFail.Buckets[0].Value.Latency)
	}

	// Invalid num buckets
	oInvBuckets := NewDistributionFromPercentiles(map[float64]Duration{0.5: 1}, 0.0, nil, 0)
	if oInvBuckets != nil {
		t.Errorf("[InvBuckets] Expected nil result for 0 buckets")
	}

	// No percentiles provided but success expected
	oNoPoints := NewDistributionFromPercentiles(map[float64]Duration{}, 0.5, nil, 1)
	assertDistributionProperties(t, "NoPoints", oNoPoints, 0.5, 0, 1) // Should only contain failure bucket

	// Identical latencies
	percentilesIdentical := map[float64]Duration{
		0.0: 10,
		0.5: 10,
		1.0: 10,
	}
	oIdentical := NewDistributionFromPercentiles(percentilesIdentical, 0.0, nil, 2)
	assertDistributionProperties(t, "IdenticalLats", oIdentical, 1.0, 2, 0)
	if !approxEqualTest(oIdentical.Buckets[0].Value.Latency, 10, 1e-9) || !approxEqualTest(oIdentical.Buckets[1].Value.Latency, 10, 1e-9) {
		t.Errorf("[IdenticalLats] Expected all latencies to be 10, got %.6f, %.6f", oIdentical.Buckets[0].Value.Latency, oIdentical.Buckets[1].Value.Latency)
	}
}

func TestNewDistributionFromPercentiles_MultiFailLatency(t *testing.T) {
	percentiles := map[float64]Duration{0.5: Millis(10)} // P0=10, P100=10 implicitly
	failLats := (&Outcomes[Duration]{}).
		Add(60, Millis(5)).
		Add(40, Millis(50))
	failRate := 0.2
	numSuccessBuckets := 1

	o := NewDistributionFromPercentiles(percentiles, failRate, failLats, numSuccessBuckets)

	assertDistributionProperties(t, "MultiFail", o, 1.0, 1, 2)

	foundSuccess := false
	foundFail5ms := false
	foundFail50ms := false

	for _, b := range o.Buckets {
		if b.Value.Success {
			foundSuccess = true
			if !approxEqualTest(b.Weight, 1.0-failRate, 1e-9) {
				t.Errorf("[MultiFail] Success weight mismatch")
			}
			if !approxEqualTest(b.Value.Latency, Millis(10), 1e-9) {
				t.Errorf("[MultiFail] Success latency mismatch")
			}
		} else {
			expW5ms := failRate * 0.60
			expW50ms := failRate * 0.40
			if approxEqualTest(b.Value.Latency, Millis(5), 1e-9) {
				foundFail5ms = true
				if !approxEqualTest(b.Weight, expW5ms, 1e-9) {
					t.Errorf("[MultiFail] 5ms Fail weight mismatch: exp %.6f, got %.6f", expW5ms, b.Weight)
				}
			} else if approxEqualTest(b.Value.Latency, Millis(50), 1e-9) {
				foundFail50ms = true
				if !approxEqualTest(b.Weight, expW50ms, 1e-9) {
					t.Errorf("[MultiFail] 50ms Fail weight mismatch: exp %.6f, got %.6f", expW50ms, b.Weight)
				}
			} else {
				t.Errorf("[MultiFail] Unexpected failure latency")
			}
		}
	}
	if !foundSuccess || !foundFail5ms || !foundFail50ms {
		t.Errorf("[MultiFail] Did not find all expected buckets (S:%v, F5:%v, F50:%v)", foundSuccess, foundFail5ms, foundFail50ms)
	}
}
