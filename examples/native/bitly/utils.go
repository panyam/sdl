package bitly

import "math"

// Helper for float comparison (reuse from metrics_test.go if accessible, or define here)
func approxEqualTest(a, b, tolerance float64) bool {
	if a == b {
		return true
	} // Handle exact equality
	return math.Abs(a-b) < tolerance
}
