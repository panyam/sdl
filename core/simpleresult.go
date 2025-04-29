package core

import (
	"math"
	"sort"
)

// A result with absolute latency
type AccessResult struct {
	Success bool
	Latency Duration
}

func AndAccessResults(a AccessResult, b AccessResult) AccessResult {
	return AccessResult{a.Success && b.Success, a.Latency + b.Latency}
}

// Ensure AccessResult implements Metricable
func (a AccessResult) IsSuccess() bool       { return a.Success }
func (ar AccessResult) GetLatency() Duration { return ar.Latency }

func (a AccessResult) AddLatency(latency Duration) AccessResult {
	return AccessResult{a.Success, a.Latency + latency}
}

// A basic reducer for AccessResult based outcomes.
// The idea here is that we first group access results based on success/failures
// and then we reduce each grouping and then aggregate it back
func ReduceAccessResults(input *Outcomes[AccessResult], numBuckets int) (out *Outcomes[AccessResult]) {
	significanceFunc := AccessResultSignificanceWeightedLatencyDelta
	successes, failures := input.Split(func(value AccessResult) bool {
		return value.Success
	})

	// Reduce success group - needs sorting by latency for significance calculation if delta-based
	// (Weight-based doesn't strictly need pre-sorting, but AdaptiveReduce sorts internally by importance)
	if successes != nil && successes.Len() > numBuckets {
		sort.SliceStable(successes.Buckets, func(i, j int) bool { return successes.Buckets[i].Value.Latency < successes.Buckets[j].Value.Latency })
		successes = AdaptiveReduce(successes, numBuckets, significanceFunc)
	}

	// Reduce failure group (often fewer buckets, but apply same logic)
	if failures != nil && failures.Len() > numBuckets {
		sort.SliceStable(failures.Buckets, func(i, j int) bool { return failures.Buckets[i].Value.Latency < failures.Buckets[j].Value.Latency })

		failures = AdaptiveReduce(failures, numBuckets, significanceFunc)
	}

	out = &Outcomes[AccessResult]{And: input.And} // Create new empty outcome list
	return out.Append(successes).Append(failures) // Append handles nil inputs
}

// MergeAdjacentAccessResults merges adjacent buckets if they have the same
// success status and their latency difference is within a relative threshold.
// Input MUST be sorted first (e.g., by Success status then Latency).
func MergeAdjacentAccessResults(input *Outcomes[AccessResult], relativeLatencyThreshold float64) (out *Outcomes[AccessResult]) {
	L := input.Len()
	if L <= 1 {
		// Return a copy to avoid modifying original if L==1
		if input != nil {
			return input.Copy()
		}
		return input
	}

	out = &Outcomes[AccessResult]{And: input.And} // Create new empty outcome list
	if L == 0 {
		return out
	} // Handle empty input

	// Add the first bucket unconditionally
	out.Buckets = append(out.Buckets, input.Buckets[0])

	for i := 1; i < L; i++ {
		current := input.Buckets[i]
		previous := &out.Buckets[len(out.Buckets)-1] // Pointer to last bucket IN THE OUTPUT

		// Check merge conditions:
		// 1. Same success status
		// 2. Previous latency is not zero (to avoid division by zero)
		// 3. Relative latency difference is within threshold
		canMerge := false
		if current.Value.Success == previous.Value.Success && previous.Value.Latency > 1e-12 { // Avoid division by zero/tiny numbers
			latencyDiff := math.Abs(current.Value.Latency - previous.Value.Latency)
			relativeDiff := latencyDiff / previous.Value.Latency
			if relativeDiff <= relativeLatencyThreshold {
				canMerge = true
			}
		}
		// Handle case where previous latency is zero/very small - merge if current is also zero/very small
		// Note: This might need refinement based on typical latency scales.
		// else if current.Value.Success == previous.Value.Success && previous.Value.Latency <= 1e-12 {
		//    if math.Abs(current.Value.Latency) <= 1e-12 {
		//        canMerge = true
		//    }
		//}

		if canMerge {
			// Merge current into previous bucket in the output list
			newWeight := previous.Weight + current.Weight
			// Weighted average for the new latency
			mergedLatency := (previous.Value.Latency*previous.Weight + current.Value.Latency*current.Weight) / newWeight

			// Update the last bucket in the output directly
			previous.Weight = newWeight
			previous.Value.Latency = mergedLatency
			// Success status remains the same
		} else {
			// Cannot merge, append current bucket as a new one to the output
			out.Buckets = append(out.Buckets, current)
		}
	}
	return out
}

func TrimToSize2(lenTrigger, maxLen int) (out func(*Outcomes[AccessResult]) *Outcomes[AccessResult]) {
	significanceFunc := AccessResultSignificanceWeightedLatencyDelta
	return func(group *Outcomes[AccessResult]) *Outcomes[AccessResult] {
		// If group is nil or empty, or already small enough, return it
		if group == nil || group.Len() <= maxLen || group.Len() == 0 {
			return group
		}
		// Sort by Success (Failures first), then Latency for MergeAdjacent
		sort.SliceStable(group.Buckets, func(i, j int) bool {
			bi, bj := group.Buckets[i].Value, group.Buckets[j].Value
			if bi.Success != bj.Success {
				return !bi.Success
			}
			return bi.Latency < bj.Latency
		})

		processedGroup := group
		// Apply merging if over trigger threshold
		if group.Len() > lenTrigger {
			// Use a refined threshold? Let's try 5% for now.
			processedGroup = MergeAdjacentAccessResults(group, 0.05)
		}

		// Apply adaptive if STILL over maxLen
		if processedGroup.Len() > maxLen {
			// *** Use the new significance function here ***
			processedGroup = AdaptiveReduce(processedGroup, maxLen, significanceFunc)
		}

		return processedGroup
	}
}

// --- Update TrimToSize ---
// It should now potentially call ReduceAccessResultsPercentileAnchor
func TrimToSize(lenTrigger, maxLen int) (out func(*Outcomes[AccessResult]) *Outcomes[AccessResult]) {
	return func(group *Outcomes[AccessResult]) *Outcomes[AccessResult] {
		if group == nil || group.Len() <= maxLen || group.Len() == 0 {
			return group
		}

		// Sort by Success then Latency (Needed for MergeAdjacent)
		sort.SliceStable(group.Buckets, func(i, j int) bool {
			bi, bj := group.Buckets[i].Value, group.Buckets[j].Value
			if bi.Success != bj.Success {
				return !bi.Success
			}
			return bi.Latency < bj.Latency
		})

		processedGroup := group
		// Apply merging if over trigger threshold
		if group.Len() > lenTrigger {
			processedGroup = MergeAdjacentAccessResults(group, 0.05) // Keep 5%
		}

		// Apply Percentile Anchoring if STILL over maxLen
		if processedGroup.Len() > maxLen {
			// *** Use the new Percentile Anchor reduction ***
			// Note: This function handles splitting success/failure internally if needed,
			// but here 'group' is already either successes or failures.
			// We need to adapt the call slightly or the function. Let's assume
			// we call the helper directly.
			processedGroup = reduceGroupPercentileAnchor(processedGroup, maxLen, []float64{0.01, 0.05, 0.25, 0.50, 0.75, 0.95, 0.99, 0.999})
		}

		return processedGroup
	}
}

// --- Significance Function: Weighted Latency Delta ---
func AccessResultSignificanceWeightedLatencyDelta(o *Outcomes[AccessResult], i int) float64 {
	if i < 0 || i >= o.Len() {
		return 0
	} // Bounds check

	bucketWeight := o.Buckets[i].Weight
	currentLatency := o.Buckets[i].Value.Latency

	var prevDelta float64 = 0
	var nextDelta float64 = 0
	epsilon := 1e-12 // To avoid zero significance if deltas are zero

	// Calculate delta to previous
	if i > 0 {
		prevDelta = math.Abs(currentLatency - o.Buckets[i-1].Value.Latency)
	}

	// Calculate delta to next
	if i < o.Len()-1 {
		nextDelta = math.Abs(currentLatency - o.Buckets[i+1].Value.Latency)
	}

	// If only one bucket, use weight as significance (or a small constant)
	if o.Len() == 1 {
		return bucketWeight + epsilon
	}

	// For first/last bucket, use the single delta available
	if i == 0 {
		return bucketWeight * (nextDelta + epsilon) // Add epsilon in case nextDelta is 0
	}
	if i == o.Len()-1 {
		return bucketWeight * (prevDelta + epsilon) // Add epsilon in case prevDelta is 0
	}

	// For intermediate buckets, use max delta
	maxDelta := math.Max(prevDelta, nextDelta)

	// Significance = Weight * MaxLatencyDifference (plus epsilon)
	// This prioritizes high-probability points that mark significant latency jumps.
	return bucketWeight * (maxDelta + epsilon)
}

// Simple significance function based on prev and next buckets
func AccessResultSignificanceLatencyDelta(o *Outcomes[AccessResult], i int) float64 {
	// Handle edge cases (first and last bucket)
	if i <= 0 || i >= o.Len()-1 {
		// Give highest significance to endpoints? Or handle differently?
		// For now, let's base it only on the available delta.
		if i == 0 && o.Len() > 1 {
			return math.Abs(o.Buckets[i].Value.Latency-o.Buckets[i+1].Value.Latency) + 1e-9 // Add epsilon
		}
		if i == o.Len()-1 && o.Len() > 1 {
			return math.Abs(o.Buckets[i].Value.Latency-o.Buckets[i-1].Value.Latency) + 1e-9 // Add epsilon
		}
		// Only one bucket or invalid index
		return 1e-9 // Minimal significance if only one point
	}
	// Calculate delta to previous and next
	prevDelta := math.Abs(o.Buckets[i].Value.Latency - o.Buckets[i-1].Value.Latency)
	nextDelta := math.Abs(o.Buckets[i].Value.Latency - o.Buckets[i+1].Value.Latency)
	// Significance could be max delta, avg delta, etc. Add epsilon for non-zero.
	return math.Max(prevDelta, nextDelta) + 1e-9
}

// --- Alternative Significance Function: Weight Based ---
func AccessResultSignificanceWeightBased(o *Outcomes[AccessResult], i int) float64 {
	// Higher weight means more significant. Add small epsilon for stability.
	if i < 0 || i >= o.Len() {
		return 0
	} // Basic bounds check
	return o.Buckets[i].Weight + 1e-12
}

func ReduceAccessResultsPercentileAnchor(input *Outcomes[AccessResult], maxBuckets int) *Outcomes[AccessResult] {
	if input == nil || input.Len() <= maxBuckets || input.Len() == 0 {
		return input // Or copy? Needs clarification. Return input for now.
	}

	successes, failures := input.Split(AccessResult.IsSuccess)

	// Define target percentiles (adjust as needed)
	targetPercentiles := []float64{0.01, 0.05, 0.25, 0.50, 0.75, 0.95, 0.99, 0.999}

	// Reduce each group
	reducedSuccesses := reduceGroupPercentileAnchor(successes, maxBuckets, targetPercentiles)
	reducedFailures := reduceGroupPercentileAnchor(failures, maxBuckets, targetPercentiles) // Use same logic/max size

	out := &Outcomes[AccessResult]{And: input.And}
	return out.Append(reducedSuccesses).Append(reducedFailures)
}

// Helper to reduce a single group (e.g., all successes or all failures)
func reduceGroupPercentileAnchor(group *Outcomes[AccessResult], maxBuckets int, percentiles []float64) *Outcomes[AccessResult] {
	if group == nil || group.Len() <= maxBuckets || group.Len() == 0 {
		return group
	}

	// 1. Sort group by latency (essential for finding percentiles)
	sort.SliceStable(group.Buckets, func(i, j int) bool {
		return group.Buckets[i].Value.Latency < group.Buckets[j].Value.Latency
	})

	// 2. Find buckets closest to target percentiles
	totalWeight := group.TotalWeight()
	anchorIndices := make(map[int]bool) // Use map to store unique indices of anchor buckets
	cumulativeWeight := 0.0
	percentileIdx := 0

	// Add first and last buckets as anchors automatically? Good practice.
	if group.Len() > 0 {
		anchorIndices[0] = true
		if group.Len() > 1 {
			anchorIndices[group.Len()-1] = true
		}
	}

	for i, bucket := range group.Buckets {
		// Avoid recalculating if already an anchor (first/last)
		if anchorIndices[i] {
			cumulativeWeight += bucket.Weight // Still need to track cumulative weight
			continue
		}

		currentPercentile := cumulativeWeight / totalWeight // Percentile *before* adding this bucket

		// Check if this bucket crosses the next target percentile
		for percentileIdx < len(percentiles) && percentiles[percentileIdx] <= currentPercentile {
			percentileIdx++ // Skip percentiles already passed
		}

		// If we haven't exhausted target percentiles, check if this bucket is the one
		if percentileIdx < len(percentiles) {
			targetWeight := percentiles[percentileIdx] * totalWeight
			// Is the target percentile within this bucket's weight range?
			if cumulativeWeight < targetWeight && (cumulativeWeight+bucket.Weight) >= targetWeight {
				// This bucket contains the percentile. Choose this or the previous one based on proximity?
				// Simplest: choose this one.
				anchorIndices[i] = true
				percentileIdx++ // Move to next target percentile
			}
		}
		cumulativeWeight += bucket.Weight
	}

	// 3. Fill remaining slots with highest weight buckets
	numAnchors := len(anchorIndices)
	numToAdd := maxBuckets - numAnchors

	outputBuckets := []Bucket[AccessResult]{}
	usedIndices := make(map[int]bool) // Track indices already added

	// Add anchor buckets first
	for idx := range anchorIndices {
		outputBuckets = append(outputBuckets, group.Buckets[idx])
		usedIndices[idx] = true
	}

	if numToAdd > 0 {
		// Find remaining buckets and sort by weight (descending)
		remainingBuckets := []struct {
			bucket Bucket[AccessResult]
			index  int
		}{}
		for i, bucket := range group.Buckets {
			if !usedIndices[i] {
				remainingBuckets = append(remainingBuckets, struct {
					bucket Bucket[AccessResult]
					index  int
				}{bucket, i})
			}
		}
		sort.Slice(remainingBuckets, func(i, j int) bool {
			return remainingBuckets[i].bucket.Weight > remainingBuckets[j].bucket.Weight // Descending weight
		})

		// Add the top N remaining buckets
		for i := 0; i < numToAdd && i < len(remainingBuckets); i++ {
			outputBuckets = append(outputBuckets, remainingBuckets[i].bucket)
			// No need to mark used again, just taking top N
		}
	}

	// 4. Final sort by latency for the output
	sort.Slice(outputBuckets, func(i, j int) bool {
		return outputBuckets[i].Value.Latency < outputBuckets[j].Value.Latency
	})

	out := &Outcomes[AccessResult]{And: group.And}
	out.Buckets = outputBuckets
	return out
}
