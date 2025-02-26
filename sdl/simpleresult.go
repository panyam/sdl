package sdl

import (
	"time"
)

// A result with absolute latency
type AccessResult struct {
	Success bool
	Latency time.Duration
}

func CombineSequentialAccessResults(a AccessResult, b AccessResult) AccessResult {
	return AccessResult{a.Success && b.Success, a.Latency + b.Latency}
}

func AccessResultSignificance(o *Outcomes[AccessResult], i int) float64 {
	prevDelta := float64((o.Buckets[i].Value.Latency - o.Buckets[i-1].Value.Latency).Abs())
	nextDelta := float64((o.Buckets[i].Value.Latency - o.Buckets[i+1].Value.Latency).Abs())
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
