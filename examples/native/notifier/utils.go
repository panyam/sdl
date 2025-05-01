package notifier

import "math"

// Helper needed if not global
func approxEqualTest(a, b, tolerance float64) bool {
	if a == b {
		return true
	}
	return math.Abs(a-b) < tolerance
}
