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

	o1 := Then(&o, &o, CombineSequentialAccessResults)
	log.Println("o1 results: ", len(o1.Buckets), o1.Buckets)

	o2 := Then(o1, o1, CombineSequentialAccessResults)
	log.Println("o2 results: ", len(o2.Buckets), o2.Buckets)

	o4 := Then(o2, o2, CombineSequentialAccessResults)
	log.Println("o4 results: ", len(o4.Buckets), o4.Buckets)
	log.Println("=============")
	o5 := ReduceAccessResults(o4, 10)
	log.Println("o4 results reduced: ", len(o5.Buckets), o5.Buckets)
}

func TestOutcomes_RangedResult(t *testing.T) {
	var o Outcomes[RangedResult]
	MS := time.Millisecond
	o.
		Add(8, RangedResult{true, 1 * MS, 5 * MS, 10 * MS}).
		Add(2, RangedResult{false, 10 * MS, 50 * MS, 100 * MS})
	log.Println("O: ", o.Buckets)

	o1 := Then(&o, &o, CombineSequentialRangedResults)
	log.Println("o1 results: ", len(o1.Buckets), o1.Buckets)

	o2 := Then(o1, o1, CombineSequentialRangedResults)
	log.Println("o2 results: ", len(o2.Buckets), o2.Buckets)

	o4 := Then(o2, o2, CombineSequentialRangedResults)
	log.Println("=============")
	log.Println("o4 results: ", len(o4.Buckets), o4.Buckets)
	log.Println("=============")
	o5 := MergeOverlappingRangedResults(o4, 0.9)
	log.Println("After overlap merges with 0.9: ", len(o5.Buckets), o5.Buckets)
	log.Println("=============")
	o5 = MergeOverlappingRangedResults(o4, 0.5)
	log.Println("After overlap merges with 0.5: ", len(o5.Buckets), o5.Buckets)
	log.Println("=============")
	o5 = MergeOverlappingRangedResults(o4, 0.99)
	log.Println("After overlap merges with 0.99: ", len(o5.Buckets), o5.Buckets)
	log.Println("=============")

	o6 := ReduceRangedResults(o5, 10)
	log.Println("o4 results reduced to 10 buckets: ", len(o6.Buckets), o6.Buckets)
}
