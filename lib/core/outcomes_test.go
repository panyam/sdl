package core

import (
	"log"
	"math/rand"
	"testing"
	"time"
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

func TestOutcomes_Sample(t *testing.T) {
	o := &Outcomes[AccessResult]{}
	o.Add(90, AccessResult{true, Millis(1)})  // 90%
	o.Add(9, AccessResult{true, Millis(10)})  // 9%
	o.Add(1, AccessResult{false, Millis(50)}) // 1%

	if o.TotalWeight() != 100 {
		t.Fatalf("Total weight should be 100")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	counts := make(map[Duration]int)
	numSamples := 100000

	for i := 0; i < numSamples; i++ {
		sample, ok := o.Sample(rng)
		if !ok {
			t.Fatal("Sample returned ok=false unexpectedly")
		}
		counts[sample.Latency]++ // Count occurrences of each latency
	}

	t.Logf("Sample counts: %v", counts)

	// Check proportions (allow generous tolerance for randomness)
	p1ms := float64(counts[Millis(1)]) / float64(numSamples)
	p10ms := float64(counts[Millis(10)]) / float64(numSamples)
	p50ms := float64(counts[Millis(50)]) / float64(numSamples)

	t.Logf("Sample proportions: P(1ms)=%.4f (exp 0.90), P(10ms)=%.4f (exp 0.09), P(50ms)=%.4f (exp 0.01)", p1ms, p10ms, p50ms)

	tolerance := 0.02 // 2% tolerance
	if !approxEqualTest(p1ms, 0.90, tolerance) {
		t.Errorf("1ms proportion %.4f outside expected range (~0.90)", p1ms)
	}
	if !approxEqualTest(p10ms, 0.09, tolerance) {
		t.Errorf("10ms proportion %.4f outside expected range (~0.09)", p10ms)
	}
	if !approxEqualTest(p50ms, 0.01, tolerance) {
		t.Errorf("50ms proportion %.4f outside expected range (~0.01)", p50ms)
	}
}

func TestOutcomes_Sample_EmptyNil(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var oNil *Outcomes[int]
	oEmpty := &Outcomes[int]{}
	oZeroWeight := (&Outcomes[int]{}).Add(0.0, 10)

	if _, ok := oNil.Sample(rng); ok {
		t.Error("Sample on nil should return ok=false")
	}
	if _, ok := oEmpty.Sample(rng); ok {
		t.Error("Sample on empty should return ok=false")
	}
	if _, ok := oZeroWeight.Sample(rng); ok {
		t.Error("Sample on zero-weight should return ok=false")
	}
	if _, ok := oEmpty.Sample(nil); ok {
		t.Error("Sample with nil RNG should return ok=false")
	}

}

func TestOutcomes_GetValue(t *testing.T) {
	oMulti := (&Outcomes[int]{}).Add(1, 10).Add(1, 20)
	oSingle := (&Outcomes[int]{}).Add(1, 10)
	oEmpty := &Outcomes[int]{}
	var oNil *Outcomes[int]

	if _, ok := oMulti.GetValue(); ok {
		t.Error("GetValue on multi-bucket should return ok=false")
	}
	if _, ok := oEmpty.GetValue(); ok {
		t.Error("GetValue on empty should return ok=false")
	}
	if _, ok := oNil.GetValue(); ok {
		t.Error("GetValue on nil should return ok=false")
	}

	v, ok := oSingle.GetValue()
	if !ok {
		t.Error("GetValue on single bucket should return ok=true")
	}
	if v != 10 {
		t.Errorf("GetValue returned wrong value: exp 10, got %d", v)
	}
}
