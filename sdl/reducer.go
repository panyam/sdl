package sdl

import (
	"math"
	"sort"
)

// Sometimes in our outcome list we want to manage the explosion of buckets (eg consider a method
// call that has 10 buckets.  If method call is called twice then the resultant
// distribution would have 10 * 10 = 100 outcomes.  Doing just 5 times gets this to 10^5
// outcomes which is extrememly large.  Instead our caller could chose to just reduce this
// to 10 after say 200-300 outcomes.  Especially for large topologies this would help us
// manage outcome sizes and afterall we just need a distribution rather than something
// highly accurate.  There are several strategies here.

type SignificanceFunction[V any] = func(o *Outcomes[V], index int) float64

// Adaptively reduce an outcome to fit into a set number of buckets.
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
