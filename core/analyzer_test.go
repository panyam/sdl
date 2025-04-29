package core

import (
	"testing"
)

// Helper function to create simple AccessResult outcomes for testing
// Ensures Metricable interface is satisfied.
func createTestOutcomes(buckets ...struct {
	W float64
	S bool
	L Duration
}) *Outcomes[AccessResult] {
	o := &Outcomes[AccessResult]{And: AndAccessResults}
	for _, b := range buckets {
		o.Add(b.W, AccessResult{Success: b.S, Latency: b.L})
	}
	return o
}

// Helper function to run Analyze within a subtest and assert pass/fail
func runAndAssert(t *testing.T, name string, shouldFail bool, simFunc func() *Outcomes[AccessResult], expectations ...Expectation) {
	t.Helper()
	t.Run(name, func(subT *testing.T) {
		subT.Helper()
		Analyze(name, simFunc, expectations...)
		actualDidFail := subT.Failed()
		if shouldFail && !actualDidFail {
			subT.Error("Test Error: Expected Analyze to fail due to unmet expectations, but it passed.")
		}
		if !shouldFail && actualDidFail {
			subT.Error("Test Error: Expected Analyze to pass, but it failed.")
		}
	})
}

// --- Test Cases ---

func TestAnalyze_Stateless_Passing(t *testing.T) {
	simFunc := func() *Outcomes[AccessResult] {
		return createTestOutcomes(
			struct {
				W float64
				S bool
				L Duration
			}{90, true, Millis(10)},
			struct {
				W float64
				S bool
				L Duration
			}{10, false, Millis(5)},
		)
	}
	expectations := []Expectation{
		ExpectAvailability(GTE, 0.85),
		ExpectMeanLatency(LTE, Millis(11)),
		ExpectP99(EQ, Millis(10)),
	}

	result := Analyze("Passing", simFunc, expectations...)

	// Use the Assert helper
	result.Assert(t)

	// Or assert manually
	if !result.AnalysisPerformed {
		t.Errorf("Analysis should have been performed")
	}
	if !result.AllPassed {
		t.Errorf("Expected all expectations to pass, but AllPassed is false")
	}
	if len(result.ExpectationChecks) != len(expectations) {
		t.Errorf("Expected %d expectation results, got %d", len(expectations), len(result.ExpectationChecks))
	}
	for _, check := range result.ExpectationChecks {
		if !check.Passed {
			t.Errorf("Expected check for %v %v %.f to pass, but it failed (Actual: %.f)",
				metricTypeToString(check.Expectation.Metric),
				operatorTypeToString(check.Expectation.Operator),
				check.Expectation.Threshold,
				check.ActualValue)
		}
	}
}

func TestAnalyze_Stateless_Failing(t *testing.T) {
	simFunc := func() *Outcomes[AccessResult] {
		return createTestOutcomes(
			struct {
				W float64
				S bool
				L Duration
			}{90, true, Millis(10)},
			struct {
				W float64
				S bool
				L Duration
			}{10, false, Millis(5)},
		)
	}
	expectations := []Expectation{
		ExpectAvailability(GTE, 0.85),     // Pass
		ExpectMeanLatency(GT, Millis(10)), // Fail
	}

	result := Analyze("Failing", simFunc, expectations...)

	// Use the AssertFailure helper (note: this itself might fail the overall test run, which is ok)
	result.AssertFailure(t)

	// Or assert manually
	if !result.AnalysisPerformed {
		t.Errorf("Analysis should have been performed")
	}
	if result.AllPassed {
		t.Errorf("Expected AllPassed to be false, but it was true")
	}
	if len(result.ExpectationChecks) != len(expectations) {
		t.Errorf("Expected %d expectation results, got %d", len(expectations), len(result.ExpectationChecks))
	}
	// Check specific failure
	if len(result.ExpectationChecks) > 1 && result.ExpectationChecks[1].Passed {
		t.Errorf("Expected second expectation (MeanLatency > 10ms) to fail, but it passed")
	}
}

func TestAnalyze_Stateless_NilOutcomes(t *testing.T) {
	nilFunc := func() *Outcomes[AccessResult] { return nil }
	expectations := []Expectation{ExpectAvailability(GTE, 0.9)}

	result := Analyze("Nil", nilFunc, expectations...)

	result.LogResults(t) // Log the skipped message

	if result.AnalysisPerformed {
		t.Error("AnalysisPerformed should be false for nil outcomes")
	}
	if result.AllPassed { // Should fail if expectations were provided
		t.Error("AllPassed should be false when expectations exist but analysis couldn't be performed")
	}
	if len(result.ExpectationChecks) != 0 {
		t.Errorf("ExpectationChecks should be empty for nil outcomes, got %d", len(result.ExpectationChecks))
	}

	// Test Assert helper
	// Need a way to capture Assert's failure without failing the main test
	// For now, just test the direct fields. The Assert/AssertFailure helpers are for *using* Analyze.
}

func TestAnalyze_Stateless_EmptyOutcomes(t *testing.T) {
	emptyFunc := func() *Outcomes[AccessResult] { return &Outcomes[AccessResult]{} }
	expectations := []Expectation{ExpectAvailability(GTE, 0.9)}

	result := Analyze("Empty", emptyFunc, expectations...)

	result.LogResults(t) // Log the skipped message

	if result.AnalysisPerformed {
		t.Error("AnalysisPerformed should be false for empty outcomes")
	}
	if result.AllPassed { // Should fail if expectations were provided
		t.Error("AllPassed should be false when expectations exist but analysis couldn't be performed")
	}
	if len(result.ExpectationChecks) != 0 {
		t.Errorf("ExpectationChecks should be empty for empty outcomes, got %d", len(result.ExpectationChecks))
	}
}

func TestAnalyze_Stateless_NoExpectations(t *testing.T) {
	simFunc := func() *Outcomes[AccessResult] {
		return createTestOutcomes(
			struct {
				W float64
				S bool
				L Duration
			}{100, true, Millis(10)},
		)
	}
	result := Analyze("NoExpectations", simFunc) // No expectations

	result.Assert(t) // Should pass

	if !result.AnalysisPerformed {
		t.Error("Analysis should have been performed")
	}
	if !result.AllPassed {
		t.Error("AllPassed should be true when no expectations are given")
	}
	if len(result.ExpectationChecks) != 0 {
		t.Errorf("ExpectationChecks should be empty when no expectations are given, got %d", len(result.ExpectationChecks))
	}
	// Check metrics were calculated
	if _, ok := result.Metrics[AvailabilityMetric]; !ok {
		t.Error("Availability metric missing")
	}
}
