package sdl

import (
	"math"
	"math/rand"
)

// ConvertToRanged converts Outcomes[AccessResult] to Outcomes[RangedResult].
// It maps each AccessResult to a RangedResult.
// 'defaultRangeFactor' (e.g., 0.1) determines the Min/Max range around the ModeLatency,
// such that Range = ModeLatency * defaultRangeFactor. Use 0 for Min=Mode=Max.
func ConvertToRanged(input *Outcomes[AccessResult], defaultRangeFactor float64) *Outcomes[RangedResult] {
	if input == nil {
		return nil
	}

	if defaultRangeFactor < 0 {
		defaultRangeFactor = 0
	}

	output := &Outcomes[RangedResult]{And: AndRangedResults} // Use RangedResult And func

	for _, bucket := range input.Buckets {
		ar := bucket.Value
		modeLat := ar.Latency
		minLat := modeLat
		maxLat := modeLat

		// Apply range only if factor > 0 and mode latency is positive
		if defaultRangeFactor > 1e-9 && modeLat > 1e-12 {
			halfRange := modeLat * defaultRangeFactor / 2.0
			minLat = modeLat - halfRange
			maxLat = modeLat + halfRange
			if minLat < 0 {
				minLat = 0
			} // Ensure non-negative
		}

		rr := RangedResult{
			Success:     ar.Success,
			MinLatency:  minLat,
			ModeLatency: modeLat,
			MaxLatency:  maxLat,
		}
		output.Add(bucket.Weight, rr)
	}
	return output
}

// ConvertToAccess converts Outcomes[RangedResult] to Outcomes[AccessResult].
// It maps each RangedResult to an AccessResult using the ModeLatency.
func ConvertToAccess(input *Outcomes[RangedResult]) *Outcomes[AccessResult] {
	if input == nil {
		return nil
	}

	output := &Outcomes[AccessResult]{And: AndAccessResults} // Use AccessResult And func

	for _, bucket := range input.Buckets {
		rr := bucket.Value
		ar := AccessResult{
			Success: rr.Success,
			Latency: rr.ModeLatency, // Use ModeLatency as the representative point
		}
		output.Add(bucket.Weight, ar)
	}
	return output
}

// --- Sampling within a RangedResult ---

// SampleWithinRange generates a single latency value based on a Min/Mode/Max range.
// Uses a simple triangular distribution approximation.
// Requires a seeded RNG.
func SampleWithinRange(minLat, modeLat, maxLat Duration, rng *rand.Rand) Duration {
	if rng == nil {
		return modeLat
	} // Cannot sample without RNG
	if maxLat <= minLat+1e-12 {
		return modeLat
	} // Avoid division by zero / No range

	// Simulate a triangular distribution: P(x) increases linearly from min to mode,
	// then decreases linearly from mode to max.
	// Area under first triangle = 0.5 * (mode-min) * height
	// Area under second triangle = 0.5 * (max-mode) * height
	// Total Area = 0.5 * (max-min) * height = 1 (for normalized PDF) => height = 2 / (max-min)
	// Probability threshold for choosing first vs second part = Area1 / TotalArea
	// Area1 = 0.5 * (mode-min) * (2 / (max-min)) = (mode-min) / (max-min)

	threshold := 0.0
	totalRange := maxLat - minLat
	if totalRange > 1e-12 {
		threshold = (modeLat - minLat) / totalRange
	}
	// Handle edge case where mode=min or mode=max
	if threshold < 1e-9 {
		threshold = 0.0
	}
	if threshold > 1.0-1e-9 {
		threshold = 1.0
	}

	r := rng.Float64() // Random number 0.0 to 1.0

	var sampledLatency Duration
	if r < threshold {
		// Sample from the first part (min to mode)
		// Use inverse transform sampling for triangle: x = min + sqrt(rand * base * height) * base / height
		// Simpler: x = min + sqrt(rand * (mode-min) * (max-min))
		sampledLatency = minLat + math.Sqrt(r*(modeLat-minLat)*totalRange)
	} else {
		// Sample from the second part (mode to max)
		// Simpler: x = max - sqrt((1-rand) * (max-mode) * (max-min))
		sampledLatency = maxLat - math.Sqrt((1.0-r)*(maxLat-modeLat)*totalRange)
	}

	// Clamp result just in case floating point issues push it out slightly
	if sampledLatency < minLat {
		return minLat
	}
	if sampledLatency > maxLat {
		return maxLat
	}
	return sampledLatency
}
