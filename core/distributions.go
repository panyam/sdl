// sdl/core/distributions.go
package core

import (
	"log" // Using log for warnings initially, could be changed
	"math"
	"sort"
)

// PercentilePoint defines a point on the CDF.
type PercentilePoint struct {
	Percentile float64  // Probability (0.0 to 1.0)
	Latency    Duration // Corresponding latency
}

// NewDistributionFromPercentiles creates an Outcomes[AccessResult] distribution
// approximating a profile defined by latency percentiles and a failure rate.
//
// Parameters:
//   - percentiles: Map of percentile points (e.g., {0.5: 50ms, 0.9: 100ms}).
//     Keys must be 0.0 <= p <= 1.0. It's strongly recommended to
//     include entries for 0.0 (P0) and 1.0 (P100) for accurate bounds.
//   - failureRate: Overall probability (0.0 to 1.0) of the operation failing.
//   - failureLatency: Outcomes[Duration] defining latency distribution on failure. Can be nil if failureRate is 0.
//     If failureRate > 0 and this is nil, a default zero latency failure is assumed.
//   - numSuccessBuckets: Number of buckets to generate for the successful outcomes. Must be > 0.
//
// Returns:
//   - *Outcomes[AccessResult]: The generated distribution. Returns nil on invalid input (e.g., numSuccessBuckets <= 0).
//
// Assumptions & Limitations:
//   - Interpolation: Uses linear interpolation between the provided percentile points for latency.
//   - Boundary Handling: If P0 or P100 are missing, it uses the minimum/maximum latency from the provided points.
//   - Accuracy: This is an approximation based on potentially sparse percentile data.
func NewDistributionFromPercentiles(
	percentiles map[float64]Duration,
	failureRate float64,
	failureLatency *Outcomes[Duration],
	numSuccessBuckets int,
) *Outcomes[AccessResult] {

	// --- Input Validation ---
	if numSuccessBuckets <= 0 {
		log.Printf("WARN: NewDistributionFromPercentiles called with numSuccessBuckets <= 0")
		return nil // Or return empty? Returning nil seems clearer for invalid input.
	}
	if failureRate < 0.0 || failureRate > 1.0 {
		log.Printf("WARN: NewDistributionFromPercentiles called with invalid failureRate (%.2f). Clamping to [0,1].", failureRate)
		failureRate = math.Max(0.0, math.Min(1.0, failureRate))
	}
	// --- Prepare Percentile Data ---
	points := make([]PercentilePoint, 0, len(percentiles)+2) // +2 for potential P0/P100
	hasP0 := false
	hasP100 := false
	minLatency := math.Inf(1)
	maxLatency := math.Inf(-1)

	for p, lat := range percentiles {
		if p < 0.0 || p > 1.0 {
			log.Printf("WARN: Skipping invalid percentile %.2f in NewDistributionFromPercentiles.", p)
			continue
		}
		if lat < 0 {
			log.Printf("WARN: Clamping negative latency %.6fs for percentile %.2f to 0.", lat, p)
			lat = 0
		}
		points = append(points, PercentilePoint{Percentile: p, Latency: lat})
		if p == 0.0 {
			hasP0 = true
		}
		if p == 1.0 {
			hasP100 = true
		}
		minLatency = math.Min(minLatency, lat)
		maxLatency = math.Max(maxLatency, lat)
	}

	// Check if we *only* have failures after validation
	totalSuccessProb := 1.0 - failureRate
	if totalSuccessProb <= 1e-9 {
		// Only need to generate failure buckets
		// (We'll do this below, but could optimize here if needed)
	} else if len(percentiles) == 0 {
		// Have success probability but NO data points to define latency.
		// Cannot generate success buckets. Log and proceed to only generate failures.
		log.Printf("WARN: NewDistributionFromPercentiles called with no percentile points but success probability > 0. Cannot generate success distribution.")
		// Effectively treat as 100% failure for generation purposes if no points given
		failureRate = 1.0
		totalSuccessProb = 0.0
	} else {
		// Add boundary points if missing
		if !hasP0 {
			p0Latency := 0.0
			if !math.IsInf(minLatency, 1) {
				p0Latency = minLatency
			}
			points = append(points, PercentilePoint{Percentile: 0.0, Latency: p0Latency})
			// log.Printf("INFO: Adding P0 point with latency %.6fs", p0Latency) // Less noisy logging maybe
		}
		if !hasP100 {
			p100Latency := 0.0
			if !math.IsInf(maxLatency, -1) {
				p100Latency = maxLatency
			}
			points = append(points, PercentilePoint{Percentile: 1.0, Latency: p100Latency})
			// log.Printf("INFO: Adding P100 point with latency %.6fs", p100Latency)
		}
		// Sort and ensure monotonicity only if we added points and have success prob
		if len(points) >= 2 {
			sort.Slice(points, func(i, j int) bool {
				return points[i].Percentile < points[j].Percentile
			})
			for i := 1; i < len(points); i++ {
				if points[i].Latency < points[i-1].Latency {
					log.Printf("WARN: Percentile points not monotonic at P%.2f (%.6fs) < P%.2f (%.6fs). Adjusting.",
						points[i].Percentile, points[i].Latency, points[i-1].Percentile, points[i-1].Latency)
					points[i].Latency = points[i-1].Latency
				}
			}
		}
	}

	// --- Initialize Output ---
	out := &Outcomes[AccessResult]{And: AndAccessResults}

	// Note: failureRate might have been forced to 1.0 above if no percentile points were given
	if failureRate > 1e-9 {
		var failureLatencyDist *Outcomes[Duration]
		if failureLatency != nil && failureLatency.Len() > 0 {
			failureLatencyDist = failureLatency
		} else {
			failureLatencyDist = (&Outcomes[Duration]{}).Add(1.0, 0.0)
			// Only warn if original failureRate was > 0
			if failureRate > 0 && failureLatency == nil && len(percentiles) > 0 { // Avoid warning if we forced failureRate=1 due to no points
				log.Printf("WARN: NewDistributionFromPercentiles using default 0 latency for failureRate %.2f as failureLatency was nil.", failureRate)
			}
		}

		totalFailWeight := failureLatencyDist.TotalWeight()
		if totalFailWeight <= 1e-12 {
			totalFailWeight = 1.0
		} // Avoid division by zero

		for _, bucket := range failureLatencyDist.Buckets {
			weight := failureRate * (bucket.Weight / totalFailWeight)
			if weight > 1e-12 {
				out.Add(weight, AccessResult{Success: false, Latency: bucket.Value})
			}
		}
	}

	// --- Add Success Buckets ---
	// Recalculate based on potentially modified failureRate
	totalSuccessProb = 1.0 - failureRate
	if totalSuccessProb <= 1e-9 {
		// No success probability OR not enough points to interpolate (points includes added bounds)
		return out
	}

	// Proceed with interpolation if we have points and success probability
	if len(points) < 2 {
		// This case *shouldn't* happen now because we add P0/P100,
		// but handle defensively: use average of min/max as latency? Or just 0?
		// Let's use the single point's latency or 0 if truly empty somehow.
		defaultSuccessLatency := 0.0
		if len(points) == 1 {
			defaultSuccessLatency = points[0].Latency
		}
		log.Printf("WARN: NewDistributionFromPercentiles has < 2 points after adding bounds, using default latency %.6fs for success.", defaultSuccessLatency)
		out.Add(totalSuccessProb, AccessResult{Success: true, Latency: defaultSuccessLatency})
		return out
	}

	successBucketWeight := totalSuccessProb / float64(numSuccessBuckets)
	pointIdx := 0

	for i := range numSuccessBuckets {
		// Target percentile *within the success distribution* (0 to 1 range)
		// Use the midpoint of the bucket's probability range
		targetSuccessP := (float64(i) + 0.5) / float64(numSuccessBuckets)

		// Find the input percentile points that bracket this target
		var pLower, pUpper PercentilePoint
		// Find the segment in the *sorted* input points list
		// Reset pointIdx search for each target percentile? No, should continue from previous.
		// Ensure pointIdx doesn't go out of bounds if targetSuccessP is exactly 1.0
		for pointIdx < len(points)-1 && points[pointIdx+1].Percentile < targetSuccessP {
			// Optimization: If target is exactly a point, use it directly? Maybe not needed.
			pointIdx++
		}

		// Ensure we have valid lower and upper bounds for interpolation
		// If targetSuccessP is exactly 1.0, pointIdx might end up as len(points)-1
		if pointIdx >= len(points)-1 {
			// Handle reaching the end (should be P100)
			pLower = points[len(points)-1]
			pUpper = pLower
		} else {
			pLower = points[pointIdx]
			pUpper = points[pointIdx+1]
		}

		var interpolatedLatency Duration
		percentileRange := pUpper.Percentile - pLower.Percentile
		latencyRange := pUpper.Latency - pLower.Latency

		// Handle edge case: target percentile is exactly lower bound percentile
		if targetSuccessP <= pLower.Percentile+1e-12 {
			interpolatedLatency = pLower.Latency
		} else if percentileRange < 1e-12 || latencyRange < 1e-12 {
			// Percentiles are identical or latency doesn't change -> use lower latency
			interpolatedLatency = pLower.Latency
		} else {
			// Linear interpolation
			proportion := (targetSuccessP - pLower.Percentile) / percentileRange
			interpolatedLatency = pLower.Latency + proportion*latencyRange
		}

		if interpolatedLatency < 0 {
			interpolatedLatency = 0
		} // Ensure non-negative
		out.Add(successBucketWeight, AccessResult{Success: true, Latency: interpolatedLatency})
	}

	// Optional: Normalize final weights if needed due to floating point math
	// out.Normalize()

	return out
}
