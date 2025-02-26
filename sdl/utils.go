package sdl

import "time"

// Returns the larger of two durations
func MaxDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 >= d2 {
		return d1
	}
	return d2
}

// Returns the smaller of two durations
func MinDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 <= d2 {
		return d1
	}
	return d2
}
