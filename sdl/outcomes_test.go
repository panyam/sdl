package sdl

import (
	"log"
	"testing"
)

func TestOutcomes(t *testing.T) {
	var o Outcomes[AccessResult]
	o.
		Add(8, AccessResult{true, Millis(1)}).
		Add(2, AccessResult{false, Millis(10)})

	o1 := And(&o, &o, AndAccessResults)

	o2 := And(o1, o1, AndAccessResults)

	o4 := And(o2, o2, AndAccessResults)
	o5 := ReduceAccessResults(o4, 10)
	log.Println("o4 simple results reduced to 10 buckets: ", len(o5.Buckets), o5.Buckets)
}

func TestOutcomes_RangedResult(t *testing.T) {
	var o Outcomes[RangedResult]
	o.
		Add(8, RangedResult{true, Millis(1), Millis(5), Millis(10)}).
		Add(2, RangedResult{false, Millis(10), Millis(50), Millis(100)})

	o1 := And(&o, &o, AndRangedResults)

	o2 := And(o1, o1, AndRangedResults)

	o4 := And(o2, o2, AndRangedResults)
	o5 := MergeOverlappingRangedResults(o4, 0.9)
	o5 = MergeOverlappingRangedResults(o4, 0.5)
	o5 = MergeOverlappingRangedResults(o4, 0.99)

	o6 := ReduceRangedResults(o5, 10)
	log.Println("o4 ranged results reduced to 10 buckets: ", len(o6.Buckets), o6.Buckets)
}
