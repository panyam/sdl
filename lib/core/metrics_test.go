package core

import (
	"math"
	"testing"
)

// approxEqual helper remains the same
func approxEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestOutcomes_Metrics_AccessResult(t *testing.T) {
	// Ensure AccessResult has GetLatency() defined in simpleresult.go
	_ = AccessResult{}.GetLatency // Compile-time check

	o := &Outcomes[AccessResult]{}
	o.Add(80, AccessResult{true, Millis(1)})  // 0.001s
	o.Add(15, AccessResult{true, Millis(10)}) // 0.010s
	o.Add(3, AccessResult{false, Millis(5)})  // Failure
	o.Add(2, AccessResult{true, Millis(100)}) // 0.100s
	// Total Weight = 100
	// Success Weight = 80 + 15 + 2 = 97
	// Failure Weight = 3
	// Weighted Success Latency = (80 * 0.001) + (15 * 0.010) + (2 * 0.100)
	//                        = 0.080 + 0.150 + 0.200 = 0.430
	// Mean Latency = 0.430 / 97 = 0.00443...

	expectedAvailability := 0.97
	// Use helper function: Availability(o)
	calculatedAvailability := Availability(o)
	if !approxEqual(calculatedAvailability, expectedAvailability, 1e-9) {
		t.Errorf("Availability mismatch: expected %f, got %f", expectedAvailability, calculatedAvailability)
	}

	expectedMeanLatency := 0.430 / 97.0
	// Use helper function: MeanLatency(o)
	calculatedMeanLatency := MeanLatency(o)
	if !approxEqual(calculatedMeanLatency, expectedMeanLatency, 1e-9) {
		t.Errorf("MeanLatency mismatch: expected %f, got %f", expectedMeanLatency, calculatedMeanLatency)
	}

	// P50: Cumulative Weights: 80 (at 1ms), 80+15=95 (at 10ms), 80+15+2=97 (at 100ms)
	// Total Success Weight = 97. Target weight for P50 = 0.50 * 97 = 48.5. Falls into the 1ms bucket.
	expectedP50 := Millis(1)
	// Use helper function: PercentileLatency(o, p)
	calculatedP50 := PercentileLatency(o, 0.50)
	if !approxEqual(calculatedP50, expectedP50, 1e-9) {
		t.Errorf("P50 Latency mismatch: expected %f, got %f", expectedP50, calculatedP50)
	}

	// P95: Target weight = 0.95 * 97 = 92.15. Falls into the 10ms bucket (cumulative weight 95).
	expectedP95 := Millis(10)
	calculatedP95 := PercentileLatency(o, 0.95)
	if !approxEqual(calculatedP95, expectedP95, 1e-9) {
		t.Errorf("P95 Latency mismatch: expected %f, got %f", expectedP95, calculatedP95)
	}

	// P99: Target weight = 0.99 * 97 = 96.03. Falls into the 100ms bucket (cumulative weight 97).
	expectedP99 := Millis(100)
	calculatedP99 := PercentileLatency(o, 0.99)
	if !approxEqual(calculatedP99, expectedP99, 1e-9) {
		t.Errorf("P99 Latency mismatch: expected %f, got %f", expectedP99, calculatedP99)
	}

	// P100: Target weight = 1.00 * 97 = 97. Falls into the 100ms bucket.
	expectedP100 := Millis(100)
	calculatedP100 := PercentileLatency(o, 1.0)
	if !approxEqual(calculatedP100, expectedP100, 1e-9) {
		t.Errorf("P100 Latency mismatch: expected %f, got %f", expectedP100, calculatedP100)
	}

	// P0: Should return lowest latency
	expectedP0 := Millis(1)
	calculatedP0 := PercentileLatency(o, 0.0)
	if !approxEqual(calculatedP0, expectedP0, 1e-9) {
		t.Errorf("P0 Latency mismatch: expected %f, got %f", expectedP0, calculatedP0)
	}
}

func TestOutcomes_Metrics_RangedResult(t *testing.T) {
	// Ensure RangedResult has GetLatency() defined in rangedresult.go
	_ = RangedResult{}.GetLatency // Compile-time check

	// Test that metrics work with RangedResult using ModeLatency
	o := &Outcomes[RangedResult]{}
	o.Add(80, RangedResult{true, Millis(0.5), Millis(1), Millis(2)})   // Mode 1ms
	o.Add(15, RangedResult{true, Millis(8), Millis(10), Millis(15)})   // Mode 10ms
	o.Add(3, RangedResult{false, Millis(4), Millis(5), Millis(8)})     // Failure
	o.Add(2, RangedResult{true, Millis(90), Millis(100), Millis(120)}) // Mode 100ms

	expectedAvailability := 0.97
	// Use helper function: Availability(o)
	calculatedAvailability := Availability(o)
	if !approxEqual(calculatedAvailability, expectedAvailability, 1e-9) {
		t.Errorf("Ranged Availability mismatch: expected %f, got %f", expectedAvailability, calculatedAvailability)
	}

	// Using ModeLatency for calculation (same as AccessResult test)
	expectedMeanLatency := 0.430 / 97.0
	// Use helper function: MeanLatency(o)
	calculatedMeanLatency := MeanLatency(o)
	if !approxEqual(calculatedMeanLatency, expectedMeanLatency, 1e-9) {
		t.Errorf("Ranged MeanLatency mismatch: expected %f, got %f", expectedMeanLatency, calculatedMeanLatency)
	}

	expectedP99 := Millis(100)
	// Use helper function: PercentileLatency(o, p)
	calculatedP99 := PercentileLatency(o, 0.99)
	if !approxEqual(calculatedP99, expectedP99, 1e-9) {
		t.Errorf("Ranged P99 Latency mismatch: expected %f, got %f", expectedP99, calculatedP99)
	}
}

/*
// This test is no longer valid as it would cause a compile-time error
// trying to pass Outcomes[any] to functions expecting Outcomes[Metricable].
func TestOutcomes_Metrics_MixedTypes_Fails(t *testing.T) {
    o := &Outcomes[any]{}
    o.Add(50, AccessResult{true, Millis(1)})
    o.Add(50, "not metricable") // Add a string outcome

    // These calls would fail compilation:
    // Availability(o)
    // MeanLatency(o)
    // PercentileLatency(o, 0.5)
}
*/

func TestOutcomes_Metrics_EmptyOrNil(t *testing.T) {
	// Need to specify the type parameter for nil Outcomes when passing to generic function
	var oNilAccess *Outcomes[AccessResult]
	var oNilRanged *Outcomes[RangedResult]
	// It doesn't matter which Metricable type we use for nil, result is 0
	oEmpty := &Outcomes[AccessResult]{}

	if Availability(oNilAccess) != 0.0 || Availability(oEmpty) != 0.0 {
		t.Errorf("Availability should be 0 for nil/empty outcomes")
	}
	if Availability(oNilRanged) != 0.0 { // Test with another type just in case
		t.Errorf("Availability should be 0 for nil/empty outcomes (ranged)")
	}

	if MeanLatency(oNilAccess) != 0.0 || MeanLatency(oEmpty) != 0.0 {
		t.Errorf("MeanLatency should be 0 for nil/empty outcomes")
	}
	if PercentileLatency(oNilAccess, 0.5) != 0.0 || PercentileLatency(oEmpty, 0.5) != 0.0 {
		t.Errorf("PercentileLatency should be 0 for nil/empty outcomes")
	}
}

func TestOutcomes_Metrics_AllFailures(t *testing.T) {
	// Using AccessResult which implements Metricable
	o := &Outcomes[AccessResult]{}
	o.Add(70, AccessResult{false, Millis(5)})
	o.Add(30, AccessResult{false, Millis(50)})

	// Now we call the functions which expect Metricable type
	if Availability(o) != 0.0 {
		t.Errorf("Availability should be 0 for all-failure outcomes, got %f", Availability(o))
	}
	// Mean latency is calculated only for successes, should be 0
	if MeanLatency(o) != 0.0 {
		t.Errorf("MeanLatency should be 0 for all-failure outcomes, got %f", MeanLatency(o))
	}
	// Percentile latency is calculated only for successes, should be 0
	if PercentileLatency(o, 0.5) != 0.0 {
		t.Errorf("PercentileLatency should be 0 for all-failure outcomes, got %f", PercentileLatency(o, 0.5))
	}
}

func TestOutcomes_Metrics_AllSuccess(t *testing.T) {
	// Added test case specifically for only successes
	o := &Outcomes[AccessResult]{}
	o.Add(70, AccessResult{true, Millis(5)})  // 0.005
	o.Add(30, AccessResult{true, Millis(50)}) // 0.050
	// Total Weight = 100
	// Weighted Latency = (70 * 0.005) + (30 * 0.050) = 0.35 + 1.5 = 1.85
	// Mean = 1.85 / 100 = 0.0185

	expectedAvailability := 1.0
	calculatedAvailability := Availability(o)
	if !approxEqual(calculatedAvailability, expectedAvailability, 1e-9) {
		t.Errorf("AllSuccess Availability mismatch: expected %f, got %f", expectedAvailability, calculatedAvailability)
	}

	expectedMeanLatency := 0.0185
	calculatedMeanLatency := MeanLatency(o)
	if !approxEqual(calculatedMeanLatency, expectedMeanLatency, 1e-9) {
		t.Errorf("AllSuccess MeanLatency mismatch: expected %f, got %f", expectedMeanLatency, calculatedMeanLatency)
	}

	// P50: Target Weight = 0.50 * 100 = 50. Falls in 5ms bucket (cumulative 70).
	expectedP50 := Millis(5)
	calculatedP50 := PercentileLatency(o, 0.50)
	if !approxEqual(calculatedP50, expectedP50, 1e-9) {
		t.Errorf("AllSuccess P50 mismatch: expected %f, got %f", expectedP50, calculatedP50)
	}

	// P90: Target Weight = 0.90 * 100 = 90. Falls in 50ms bucket (cumulative 100).
	expectedP90 := Millis(50)
	calculatedP90 := PercentileLatency(o, 0.90)
	if !approxEqual(calculatedP90, expectedP90, 1e-9) {
		t.Errorf("AllSuccess P90 mismatch: expected %f, got %f", expectedP90, calculatedP90)
	}
}
