package sdl

import (
	"math"
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

func RangedResultSignificance(o *Outcomes[RangedResult], i int) (importance float64) {
	minDistance := math.Inf(1)
	thisBucket := o.Buckets[i]
	for j, bucket := range o.Buckets {
		if i == j {
			continue
		}
		distance := thisBucket.Value.DistTo(&bucket.Value)
		minDistance = min(minDistance, distance)
	}

	width := float64(thisBucket.Value.MaxLatency - thisBucket.Value.MinLatency)
	importance = thisBucket.Weight * math.Log(1+minDistance) * (1 + math.Log(1+width)*0.2)
	return
}

// A basic reducer for RangedResult based outcomes.
// The idea here is that we first group access results based on success/failures
// and then we reduce each grouping and then aggregate it back
func ReduceRangedResults(input *Outcomes[RangedResult], numBuckets int) (out *Outcomes[RangedResult]) {
	successes, failures := input.Split(func(value RangedResult) bool {
		return value.Success
	})
	successes = AdaptiveReduce(successes, numBuckets, RangedResultSignificance)
	failures = AdaptiveReduce(failures, numBuckets, RangedResultSignificance)
	return out.Append(successes).Append(failures)
}

func MergeOverlappingRangedResults(input *Outcomes[RangedResult], overlapThreshold float64) (out *Outcomes[RangedResult]) {
	if input == nil || input.Len() <= 1 {
		return input.Copy()
	}

	// Can sort instead
	usedIndices := map[int]bool{}
	for i, current := range input.Buckets {
		if usedIndices[i] {
			continue
		}
		usedIndices[i] = true
		mergedBucket := current
		for j, next := range input.Buckets {
			if usedIndices[j] {
				continue
			}
			overlap := current.Value.Overlap(&next.Value)
			if overlap >= overlapThreshold {
				// Merge the ranges
				totalWeight := mergedBucket.Weight + next.Weight

				// Weight the merged range by probability
				mergedBucket.Value.MinLatency = MinDuration(mergedBucket.Value.MinLatency, next.Value.MinLatency)
				mergedBucket.Value.MaxLatency = MaxDuration(mergedBucket.Value.MaxLatency, next.Value.MaxLatency)
				mergedBucket.Value.ModeLatency = ((mergedBucket.Value.ModeLatency)*mergedBucket.Weight + (next.Value.ModeLatency)*next.Weight) / totalWeight
				mergedBucket.Weight = totalWeight
				usedIndices[j] = true
			}
		}
		out = out.Add(mergedBucket.Weight, mergedBucket.Value)
	}
	return
}
