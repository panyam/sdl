// sdl/examples/gpucaller/utils.go
package gpucaller

import "math"

// Helper for float comparison
func approxEqualTest(a, b, tolerance float64) bool {
	if a == b {
		return true
	}
	return math.Abs(a-b) < tolerance
}
