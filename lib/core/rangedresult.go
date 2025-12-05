package core

import (
	"math"
	"sort"
)

// A result type with latency ranges instead of absolute latency
type RangedResult struct {
	Success     bool
	MinLatency  Duration
	ModeLatency Duration
	MaxLatency  Duration
}

func (a *RangedResult) And(another Outcome) Outcome {
	b := another.(*RangedResult)
	if b == nil {
		return nil
	}
	out := AndRangedResults(*a, *b)
	return &out
}

func AndRangedResults(a RangedResult, b RangedResult) RangedResult {
	return RangedResult{
		a.Success && b.Success,
		a.MinLatency + b.MinLatency,
		a.ModeLatency + b.ModeLatency,
		a.MaxLatency + b.MaxLatency,
	}
}

// Ensure RangedResult implements Metricable (using ModeLatency for point estimate)
// Note: This is an approximation for RangedResult. More sophisticated
// range-based metrics could be added later.
func (r RangedResult) IsSuccess() bool       { return r.Success }
func (rr RangedResult) GetLatency() Duration { return rr.ModeLatency }

func (r *RangedResult) Range() float64 {
	return float64(r.MaxLatency - r.MinLatency)
}

// Returns how much overlap is there with another range.
func (r *RangedResult) Overlap(r2 *RangedResult) float64 {
	if r.Success != r2.Success {
		// We assume client will already group by successes
		return 0
	}
	left := MaxDuration(r.MinLatency, r2.MinLatency)
	right := MinDuration(r.MaxLatency, r2.MaxLatency)
	diff := 0.0
	if right > left {
		diff = float64(right - left)
	}
	return diff / math.Max(r.Range(), r2.Range())
}

// Distance of this range from another
func (r *RangedResult) DistTo(another *RangedResult) float64 {
	minDist := math.Abs(r.MinLatency - another.MinLatency)
	modeDist := math.Abs(r.ModeLatency - another.ModeLatency)
	maxDist := math.Abs(r.MaxLatency - another.MaxLatency)

	// Give more weight to mode, as it's the most representative
	return float64((minDist + 4.0*modeDist + maxDist) / 6.0)
}

func MergeOverlappingRangedResults(input *Outcomes[RangedResult], overlapThreshold float64) (out *Outcomes[RangedResult]) {
	L := input.Len()
	if L <= 1 {
		if input != nil {
			return input.Copy()
		}
		return input
	}

	out = &Outcomes[RangedResult]{And: input.And} // Initialize output
	if L == 0 {
		return out
	}

	// Split by success status FIRST, process groups independently
	successes, failures := input.Split(RangedResult.IsSuccess)

	// Process successes
	if successes != nil && successes.Len() > 0 {
		processedSuccesses := mergeOverlappingGroup(successes, overlapThreshold)
		out = out.Append(processedSuccesses)
	}

	// Process failures
	if failures != nil && failures.Len() > 0 {
		processedFailures := mergeOverlappingGroup(failures, overlapThreshold)
		out = out.Append(processedFailures)
	}

	return out
}

// Helper function to process a group (all success or all failure)
func mergeOverlappingGroup(group *Outcomes[RangedResult], overlapThreshold float64) *Outcomes[RangedResult] {
	L := group.Len()
	if L <= 1 {
		return group
	} // Already reduced or empty

	out := &Outcomes[RangedResult]{And: group.And}
	usedIndices := make(map[int]bool, L) // Use map to track used buckets

	for i := 0; i < L; i++ {
		if usedIndices[i] {
			continue
		} // Skip if already merged into another bucket

		current := group.Buckets[i] // Start with this bucket
		mergedBucket := current     // Initialize the potential merged bucket
		usedIndices[i] = true       // Mark current as used

		// Iterate through the *rest* of the buckets in the group to find overlaps
		for j := i + 1; j < L; j++ {
			if usedIndices[j] {
				continue
			} // Skip if already used

			next := group.Buckets[j]

			// Check for overlap (using the original bucket 'current' or the evolving 'mergedBucket'?)
			// Let's check overlap between the *potentially growing* mergedBucket and the next candidate
			// This might lead to more merging than just checking against the initial 'current'.
			overlap := mergedBucket.Value.Overlap(&next.Value)

			if overlap >= overlapThreshold {
				// Merge 'next' into 'mergedBucket'
				totalWeight := mergedBucket.Weight + next.Weight

				// Merge the range values
				mergedBucket.Value.MinLatency = MinDuration(mergedBucket.Value.MinLatency, next.Value.MinLatency)
				mergedBucket.Value.MaxLatency = MaxDuration(mergedBucket.Value.MaxLatency, next.Value.MaxLatency)
				// Weighted average for ModeLatency
				// Ensure correct weights are used before modification
				prevMergedWeight := mergedBucket.Weight
				mergedBucket.Value.ModeLatency = ((mergedBucket.Value.ModeLatency * prevMergedWeight) + (next.Value.ModeLatency * next.Weight)) / totalWeight
				mergedBucket.Weight = totalWeight // Update weight *after* calculating new mode

				usedIndices[j] = true // Mark 'next' as used because it's merged
			}
		}
		// Add the final merged bucket (which might just be the original 'current' if no merges occurred)
		out = out.Add(mergedBucket.Weight, mergedBucket.Value)
	}
	return out
}

// --- ReduceRangedResults (Restore to original or keep simple merge call) ---
// Let's keep it calling MergeOverlapping for now, as Adaptive is broken/slow.
func ReduceRangedResults(input *Outcomes[RangedResult], numBuckets int) (out *Outcomes[RangedResult]) {
	// Using the restored MergeOverlapping
	overlapThreshold := 0.1 // Default threshold (10%) - can be adjusted
	merged := MergeOverlappingRangedResults(input, overlapThreshold)
	// If merged result still > numBuckets, we need another step (e.g., interpolation) later.
	// For now, just return merged result.
	if merged.Len() > numBuckets {
		// Placeholder: Add interpolation or other reduction here if needed in future
		// return InterpolateRangedResults(merged, numBuckets)
	}
	return merged
}

// --- Keep RangedResultSignificance (even if slow/unused for now) ---
func RangedResultSignificance(o *Outcomes[RangedResult], i int) (importance float64) {
	// ... (original implementation) ...
	// This calculates distance to *all* other points, hence the slowness.
	minDistance := math.Inf(1)
	thisBucket := o.Buckets[i]
	for j, bucket := range o.Buckets {
		if i == j {
			continue
		}
		// Ensure comparison makes sense (e.g., compare only within same success status?)
		if thisBucket.Value.Success == bucket.Value.Success {
			distance := thisBucket.Value.DistTo(&bucket.Value)
			minDistance = math.Min(minDistance, distance) // Use math.Min
		}
	}
	// Handle case where minDistance remains infinite (e.g., only one bucket of a given status)
	if math.IsInf(minDistance, 1) {
		minDistance = 0 // Or some other default?
	}

	width := float64(thisBucket.Value.MaxLatency - thisBucket.Value.MinLatency)
	// Avoid Log(0) or Log(negative) if minDistance is 0
	logMinDistance := 0.0
	if minDistance > 0 {
		logMinDistance = math.Log(1 + minDistance)
	}
	logWidth := 0.0
	if width > 0 {
		logWidth = math.Log(1 + width)
	}

	importance = thisBucket.Weight * logMinDistance * (1 + logWidth*0.2)
	return importance
}

func InterpolateRangedResults(input *Outcomes[RangedResult], targetBuckets int) *Outcomes[RangedResult] {
	L := input.Len()
	if L == 0 || targetBuckets <= 0 {
		return &Outcomes[RangedResult]{And: input.And} // Return empty
	}
	if L <= targetBuckets {
		return input.Copy() // No need to interpolate
	}
	if input.Buckets[0].Value.Success != input.Buckets[L-1].Value.Success {
		// Assume caller split by success status
		return nil // Or panic/error
	}
	successStatus := input.Buckets[0].Value.Success

	// Ensure input is sorted by ModeLatency for predictable interpolation
	sort.SliceStable(input.Buckets, func(i, j int) bool {
		return input.Buckets[i].Value.ModeLatency < input.Buckets[j].Value.ModeLatency
	})

	out := &Outcomes[RangedResult]{And: input.And}
	out.Buckets = make([]Bucket[RangedResult], 0, targetBuckets)

	totalWeight := input.TotalWeight()
	if totalWeight <= 0 {
		totalWeight = 1.0
	}

	targetCumulativeWeights := make([]float64, targetBuckets)
	for i := 0; i < targetBuckets; i++ {
		targetCumulativeWeights[i] = totalWeight * (float64(i) + 0.5) / float64(targetBuckets)
	}

	currentCumulativeWeight := 0.0
	inputIndex := 0

	for targetIdx := 0; targetIdx < targetBuckets; targetIdx++ {
		targetW := targetCumulativeWeights[targetIdx]

		var prevBucket, nextBucket Bucket[RangedResult]
		prevCumulativeWeight := currentCumulativeWeight

		for inputIndex < L && currentCumulativeWeight < targetW {
			prevBucket = input.Buckets[inputIndex]
			prevCumulativeWeight = currentCumulativeWeight
			currentCumulativeWeight += input.Buckets[inputIndex].Weight
			inputIndex++
		}

		if inputIndex >= L {
			nextBucket = input.Buckets[L-1]
			// Avoid adding duplicates if multiple targets land in the last bucket range
			if len(out.Buckets) > 0 && out.Buckets[len(out.Buckets)-1].Value == nextBucket.Value {
				continue
			}
		} else {
			nextBucket = input.Buckets[inputIndex]
			if inputIndex == 0 {
				prevBucket = nextBucket
				prevCumulativeWeight = 0
			}
		}

		// Interpolate Min, Mode, Max Latencies
		interpolatedMinLat, interpolatedModeLat, interpolatedMaxLat := 0.0, 0.0, 0.0
		weightRange := currentCumulativeWeight - prevCumulativeWeight

		prevVal := prevBucket.Value
		nextVal := nextBucket.Value

		if weightRange > 1e-9 { // Avoid division by zero
			proportion := (targetW - prevCumulativeWeight) / weightRange
			// Check if latencies actually differ to avoid interpolating same value
			if math.Abs(nextVal.MinLatency-prevVal.MinLatency) > 1e-12 {
				interpolatedMinLat = prevVal.MinLatency + proportion*(nextVal.MinLatency-prevVal.MinLatency)
			} else {
				interpolatedMinLat = prevVal.MinLatency
			}

			if math.Abs(nextVal.ModeLatency-prevVal.ModeLatency) > 1e-12 {
				interpolatedModeLat = prevVal.ModeLatency + proportion*(nextVal.ModeLatency-prevVal.ModeLatency)
			} else {
				interpolatedModeLat = prevVal.ModeLatency
			}

			if math.Abs(nextVal.MaxLatency-prevVal.MaxLatency) > 1e-12 {
				interpolatedMaxLat = prevVal.MaxLatency + proportion*(nextVal.MaxLatency-prevVal.MaxLatency)
			} else {
				interpolatedMaxLat = prevVal.MaxLatency
			}

		} else { // If weight range is tiny, pick closest bucket's values
			if math.Abs(targetW-prevCumulativeWeight) < math.Abs(targetW-currentCumulativeWeight) || inputIndex == 0 {
				interpolatedMinLat, interpolatedModeLat, interpolatedMaxLat = prevVal.MinLatency, prevVal.ModeLatency, prevVal.MaxLatency
			} else {
				interpolatedMinLat, interpolatedModeLat, interpolatedMaxLat = nextVal.MinLatency, nextVal.ModeLatency, nextVal.MaxLatency
			}
		}

		// Ensure Min <= Mode <= Max after interpolation (can sometimes cross due to independent interpolation)
		if interpolatedModeLat < interpolatedMinLat {
			interpolatedModeLat = interpolatedMinLat
		}
		if interpolatedMaxLat < interpolatedModeLat {
			interpolatedMaxLat = interpolatedModeLat
		}

		newBucketWeight := totalWeight / float64(targetBuckets)

		out.Buckets = append(out.Buckets, Bucket[RangedResult]{
			Weight: newBucketWeight,
			Value: RangedResult{
				Success:     successStatus,
				MinLatency:  interpolatedMinLat,
				ModeLatency: interpolatedModeLat,
				MaxLatency:  interpolatedMaxLat,
			},
		})
	} // end for targetIdx

	// Optional: Normalize weights
	finalTotalWeight := out.TotalWeight()
	if finalTotalWeight > 1e-9 {
		renormFactor := totalWeight / finalTotalWeight
		for i := range out.Buckets {
			out.Buckets[i].Weight *= renormFactor
		}
	}

	return out
}

// --- New Function: TrimToSizeRanged ---
// Applies reduction strategy for RangedResult: Merge high overlap, then interpolate.
func TrimToSizeRanged(lenTrigger, maxLen int, mergeOverlapThreshold float64) (out func(*Outcomes[RangedResult]) *Outcomes[RangedResult]) {
	return func(group *Outcomes[RangedResult]) *Outcomes[RangedResult] {
		if group == nil || group.Len() <= maxLen || group.Len() == 0 {
			return group
		}

		processedGroup := group
		// Apply merging if over trigger threshold
		if group.Len() > lenTrigger {
			// Use the map-based merge (faster)
			processedGroup = MergeOverlappingRangedResults(group, mergeOverlapThreshold)
		}

		// Apply interpolation if STILL over maxLen
		if processedGroup.Len() > maxLen {
			// Interpolate requires sorting; MergeOverlapping doesn't guarantee sort order.
			// Sort by ModeLatency within the group before interpolating.
			// Note: Interpolate function also sorts, but sorting here clarifies the process.
			sort.SliceStable(processedGroup.Buckets, func(i, j int) bool {
				return processedGroup.Buckets[i].Value.ModeLatency < processedGroup.Buckets[j].Value.ModeLatency
			})
			processedGroup = InterpolateRangedResults(processedGroup, maxLen)
		}

		return processedGroup
	}
}
