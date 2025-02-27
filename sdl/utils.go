package sdl

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
