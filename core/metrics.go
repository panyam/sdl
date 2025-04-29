package core

import (
	"sort"
	// "log" // Optional logging
)

// Metricable interface (definition can live here or in outcomes.go)
type Metricable interface {
	IsSuccess() bool
	GetLatency() Duration
}

// --- Metric Helper Functions ---

// MeanLatency calculates the weighted average latency for successful outcomes.
// V must implement Metricable.
func MeanLatency[V Metricable](o *Outcomes[V]) Duration {
	totalSuccessWeight := 0.0
	weightedLatencySum := 0.0

	if o == nil || o.Len() == 0 {
		return 0
	}

	// No type assertion needed here due to the [V Metricable] constraint
	for _, bucket := range o.Buckets {
		if bucket.Value.IsSuccess() {
			totalSuccessWeight += bucket.Weight
			weightedLatencySum += bucket.Value.GetLatency() * bucket.Weight
		}
	}

	if totalSuccessWeight == 0 {
		return 0
	}
	return weightedLatencySum / totalSuccessWeight
}

// PercentileLatency calculates the P-th percentile latency for successful outcomes.
// 'p' should be between 0.0 and 1.0 (e.g., 0.99 for P99).
// V must implement Metricable.
func PercentileLatency[V Metricable](o *Outcomes[V], p float64) Duration {
	if o == nil || o.Len() == 0 || p < 0.0 || p > 1.0 {
		return 0
	}

	type successBucket struct {
		Latency Duration
		Weight  float64
	}
	var successBuckets []successBucket
	totalSuccessWeight := 0.0

	// No type assertion needed
	for _, bucket := range o.Buckets {
		if bucket.Value.IsSuccess() {
			successBuckets = append(successBuckets, successBucket{
				Latency: bucket.Value.GetLatency(),
				Weight:  bucket.Weight,
			})
			totalSuccessWeight += bucket.Weight
		}
	} // <--- This closing brace was missing for the for loop

	if len(successBuckets) == 0 || totalSuccessWeight <= 0 {
		return 0
	}

	// Sort buckets by latency
	sort.Slice(successBuckets, func(i, j int) bool {
		return successBuckets[i].Latency < successBuckets[j].Latency
	})

	targetWeight := p * totalSuccessWeight
	cumulativeWeight := 0.0
	for _, sb := range successBuckets {
		cumulativeWeight += sb.Weight
		if cumulativeWeight >= targetWeight {
			return sb.Latency
		}
	}

	// Fallback (shouldn't be reached if p <= 1.0 and len > 0)
	if len(successBuckets) > 0 {
		return successBuckets[len(successBuckets)-1].Latency
	}
	return 0 // Should not happen
} // <--- This closing brace was missing for the function body

// Availability calculates the probability of a successful outcome.
// V must implement Metricable.
func Availability[V Metricable](o *Outcomes[V]) float64 {
	totalSuccessWeight := 0.0
	totalWeight := 0.0

	if o == nil || o.Len() == 0 {
		return 0.0
	}

	// No type assertion needed
	for _, bucket := range o.Buckets {
		totalWeight += bucket.Weight
		if bucket.Value.IsSuccess() {
			totalSuccessWeight += bucket.Weight
		}
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalSuccessWeight / totalWeight
}
