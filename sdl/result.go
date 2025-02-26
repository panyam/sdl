package sdl

import (
	"math"
	"sort"
	"time"
)

type AccessResult struct {
	Success bool
	Latency time.Duration
}

// A basic reducer for AccessResult based outcomes.
// The idea here is that we first group access results based on success/failures
// and then we reduce each grouping and then aggregate it back
func ReduceAccessResults(input *Outcomes[AccessResult], numBuckets int) (out *Outcomes[AccessResult]) {
	successes, failures := input.Partition(func(value AccessResult) bool {
		return value.Success
	})

	successes = AdaptiveReduce(successes, numBuckets, func(i int) float64 {
		prevDelta := math.Abs(float64(successes.Buckets[i].Value.Latency) - float64(successes.Buckets[i-1].Value.Latency))
		nextDelta := math.Abs(float64(successes.Buckets[i].Value.Latency) - float64(successes.Buckets[i+1].Value.Latency))
		return max(prevDelta, nextDelta)
	})
	failures = AdaptiveReduce(failures, numBuckets, func(i int) float64 {
		prevDelta := math.Abs(float64(failures.Buckets[i].Value.Latency) - float64(failures.Buckets[i-1].Value.Latency))
		nextDelta := math.Abs(float64(failures.Buckets[i].Value.Latency) - float64(failures.Buckets[i+1].Value.Latency))
		return max(prevDelta, nextDelta)
	})
	return out.Append(successes).Append(failures)
}

func AdaptiveReduce[V any](o *Outcomes[V], maxBuckets int, sigFunc func(i int) float64) (out *Outcomes[V]) {
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
			sig := sigFunc(i)
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

// Strategy 4: Adaptive reduction based on distribution shape
/*
   static adaptiveReduce(buckets: LatencyBucket[], maxBuckets: number): LatencyBucket[] {
       // First calculate the "importance" of each region
       const deltas = buckets.map((bucket, i) => {
           if (i === 0 || i === buckets.length - 1) return Infinity;
           const prevDelta = Math.abs(bucket.latency - buckets[i-1].latency);
           const nextDelta = Math.abs(buckets[i+1].latency - bucket.latency);
           return Math.max(prevDelta, nextDelta) * bucket.probability;
       });

       // Keep buckets with highest importance
       const withImportance = buckets.map((bucket, i) => ({
           ...bucket,
           importance: deltas[i]
       }));

       const sorted = withImportance
           .sort((a, b) => b.importance - a.importance)
           .slice(0, maxBuckets);

       // Normalize probabilities
       const totalProb = sorted.reduce((sum, b) => sum + b.probability, 0);
       return sorted.map(({latency, probability}) => ({
           latency,
           probability: probability / totalProb
       })).sort((a, b) => a.latency - b.latency);
   }
*/
