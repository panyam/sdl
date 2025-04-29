package core

import "math"

type Duration = float64

func Millis(val float64) Duration {
	return val / 1000.0
}

func Micros(val float64) Duration {
	return val / 1000000.0
}

func Nanos(val float64) Duration {
	return val / 1000000000.0
}

// Returns the larger of two durations
func MaxDuration(d1 Duration, d2 Duration) Duration {
	if d1 >= d2 {
		return d1
	}
	return d2
}

// Returns the smaller of two durations
func MinDuration(d1 Duration, d2 Duration) Duration {
	if d1 <= d2 {
		return d1
	}
	return d2
}

// Helper for float comparison (reuse from metrics_test.go if accessible, or define here)
func approxEqualTest(a, b, tolerance float64) bool {
	if a == b {
		return true
	} // Handle exact equality
	return math.Abs(a-b) < tolerance
}
