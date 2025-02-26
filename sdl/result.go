package sdl

import (
	"log"
	"math"
	"sort"
	"time"
)

type SignificanceFunction[V any] = func(o *Outcomes[V], index int) float64

func MaxDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 >= d2 {
		return d1
	}
	return d2
}

func MinDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 <= d2 {
		return d1
	}
	return d2
}

// A result with absolute latency
type AccessResult struct {
	Success bool
	Latency time.Duration
}

func CombineSequentialAccessResults(a AccessResult, b AccessResult) AccessResult {
	return AccessResult{a.Success && b.Success, a.Latency + b.Latency}
}

func AccessResultSignificance(o *Outcomes[AccessResult], i int) float64 {
	prevDelta := math.Abs(float64(o.Buckets[i].Value.Latency) - float64(o.Buckets[i-1].Value.Latency))
	nextDelta := math.Abs(float64(o.Buckets[i].Value.Latency) - float64(o.Buckets[i+1].Value.Latency))
	return max(prevDelta, nextDelta)
}

// A basic reducer for AccessResult based outcomes.
// The idea here is that we first group access results based on success/failures
// and then we reduce each grouping and then aggregate it back
func ReduceAccessResults(input *Outcomes[AccessResult], numBuckets int) (out *Outcomes[AccessResult]) {
	successes, failures := input.Partition(func(value AccessResult) bool {
		return value.Success
	})

	successes = AdaptiveReduce(successes, numBuckets, AccessResultSignificance)
	failures = AdaptiveReduce(failures, numBuckets, AccessResultSignificance)
	return out.Append(successes).Append(failures)
}

func AdaptiveReduce[V any](o *Outcomes[V], maxBuckets int, sigFunc SignificanceFunction[V]) (out *Outcomes[V]) {
	type BucketWithImportance[V any] struct {
		Bucket[V]
		Importance float64
	}

	if len(o.Buckets) <= maxBuckets {
		return o
	}

	var buckets []BucketWithImportance[V]

	for i, bucket := range o.Buckets {
		delta := 0.0
		if i == 0 || i == len(o.Buckets)-1 {
			delta = math.Inf(1)
		} else {
			sig := sigFunc(o, i)
			delta = bucket.Weight * sig
		}
		buckets = append(buckets, BucketWithImportance[V]{
			Bucket:     bucket,
			Importance: delta,
		})
	}

	// Now we have importance of each bucket as well as all the buckets
	// sort buckets by importance
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Importance < buckets[j].Importance
	})
	// log.Println("2. Buckets: ", buckets, len(o.Buckets), maxBuckets)
	// and take the N buckets
	for _, bucket := range buckets[:maxBuckets] {
		out = out.Add(bucket.Weight, bucket.Value)
	}
	return
}

// A result type with latency ranges instead of absolute latency
type RangedResult struct {
	Success     bool
	MinLatency  time.Duration
	ModeLatency time.Duration
	MaxLatency  time.Duration
}

func CombineSequentialRangedResults(a RangedResult, b RangedResult) RangedResult {
	return RangedResult{
		a.Success && b.Success,
		a.MinLatency + b.MinLatency,
		a.ModeLatency + b.ModeLatency,
		a.MaxLatency + b.MaxLatency,
	}
}

func (r *RangedResult) Range() float64 {
	return float64(r.MaxLatency - r.MinLatency)
}

// Returns how much overlap is there with another range.
func (r *RangedResult) Overlap(r2 *RangedResult) float64 {
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
	minDist := (r.MinLatency - another.MinLatency).Abs()
	modeDist := (r.ModeLatency - another.ModeLatency).Abs()
	maxDist := (r.MaxLatency - another.MaxLatency).Abs()

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
	successes, failures := input.Partition(func(value RangedResult) bool {
		return value.Success
	})
	log.Println("Success: ", successes)
	log.Println("Failures: ", failures)
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
				mergedBucket.Value.ModeLatency = time.Duration((float64(mergedBucket.Value.ModeLatency)*mergedBucket.Weight + float64(next.Value.ModeLatency)*next.Weight) / totalWeight)
				mergedBucket.Weight = totalWeight
				usedIndices[j] = true
			}
		}
		out = out.Add(mergedBucket.Weight, mergedBucket.Value)
	}
	return
}

/**
static mergeOverlapping(buckets: RangeBucket[], overlapThreshold: number): RangeBucket[] {
        if (buckets.length <= 1) return buckets;

        const result: RangeBucket[] = [];
        const usedIndices = new Set<number>();

        for (let i = 0; i < buckets.length; i++) {
            if (usedIndices.has(i)) continue;

            const current = buckets[i];
            const mergedBucket = { ...current };
            usedIndices.add(i);

            // Find overlapping ranges
            for (let j = i + 1; j < buckets.length; j++) {
                if (usedIndices.has(j)) continue;

                const next = buckets[j];
                const overlap = this.rangeOverlap(current.range, next.range);

                if (overlap >= overlapThreshold) {
                    // Merge the ranges
                    const totalWeight = mergedBucket.probability + next.probability;

                    // Weight the merged range by probability
                    mergedBucket.range = {
                        min: math.Min(mergedBucket.range.min, next.range.min),
                        mode: (mergedBucket.range.mode * mergedBucket.probability +
                               next.range.mode * next.probability) / totalWeight,
                        max: math.Max(mergedBucket.range.max, next.range.max)
                    };

                    mergedBucket.probability = totalWeight;
                    usedIndices.add(j);
                }
            }

            result.push(mergedBucket);
        }

        return result;
    }
*/
