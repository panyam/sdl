package sdl

import (
	"log"
	"testing"
	"time"
)

func TestOutcomes(t *testing.T) {
	var o Outcomes[AccessResult]
	o.
		Add(8, AccessResult{true, 1 * time.Millisecond}).
		Add(2, AccessResult{false, 10 * time.Millisecond})
	log.Println("O: ", o.Buckets)

	o1 := Then(&o, &o, func(a, b AccessResult) AccessResult {
		return AccessResult{a.Success && b.Success, a.Latency + b.Latency}
	})
	log.Println("o1 results: ", len(o1.Buckets), o1.Buckets)

	o2 := Then(o1, o1, func(a, b AccessResult) AccessResult {
		return AccessResult{a.Success && b.Success, a.Latency + b.Latency}
	})
	log.Println("o2 results: ", len(o2.Buckets), o2.Buckets)

	o4 := Then(o2, o2, func(a, b AccessResult) AccessResult {
		return AccessResult{a.Success && b.Success, a.Latency + b.Latency}
	})
	log.Println("o4 results: ", len(o4.Buckets), o4.Buckets)
	log.Println("=============")
	o5 := ReduceAccessResults(o4, 10)
	log.Println("o4 results reduced: ", len(o5.Buckets), o5.Buckets)
}
