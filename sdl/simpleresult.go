package sdl

import (
	"math"
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

func MergeAdjacentAccessResults(input *Outcomes[AccessResult], maxError float64) (out *Outcomes[AccessResult]) {
	L := input.Len()
	if L <= 1 {
		return input
	}
	out = out.Add(input.Buckets[0].Weight, input.Buckets[0].Value)
	for i := 1; i < L; i++ {
		current := input.Buckets[i]
		previous := out.Buckets[len(out.Buckets)-1]
		if current.Value.Success != previous.Value.Success {
			out = out.Add(current.Weight, current.Value)
			continue
		}

		mergedLatency := (float64(previous.Value.Latency)*previous.Weight +
			float64(current.Value.Latency)*current.Weight) /
			(previous.Weight + current.Weight)
		err := math.Abs(mergedLatency-float64(previous.Value.Latency))*previous.Weight +
			math.Abs(mergedLatency-float64(current.Value.Latency))*current.Weight
		if err <= maxError {
			out.Buckets[len(out.Buckets)-1].Weight = previous.Weight + current.Weight
			out.Buckets[len(out.Buckets)-1].Value = AccessResult{
				Success: current.Value.Success,
				Latency: time.Duration(mergedLatency * float64(time.Second)),
			}
		} else {
			out = out.Add(current.Weight, current.Value)
		}
	}
	return
}
