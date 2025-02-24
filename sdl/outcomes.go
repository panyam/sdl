package sdl

import (
	"log"
)

type Bucket[V any] struct {
	Weight Fraction
	Value  V
}

type Outcomes[V any] struct {
	Buckets []Bucket[V]
}

func (o *Outcomes[V]) Len() int {
	return len(o.Buckets)
}

func (o *Outcomes[V]) TotalWeight() (out Fraction) {
	if o == nil {
		return FracZero
	}
	out = 0
	for _, v := range o.Buckets {
		out = out.Plus(v.Weight)
		out = out.Factorized()
	}
	return
}

func ToOutcomesAny(input any) *Outcomes[any] {
	return (input).(*Outcomes[any])
}

func Convert[V any, U any](this *Outcomes[V], mapper func(v V) U) (out *Outcomes[U]) {
	for _, v := range this.Buckets {
		out = out.Add(v.Weight, mapper(v.Value))
	}
	return
}

func (this *Outcomes[V]) Append(another *Outcomes[V]) *Outcomes[V] {
	if this == nil {
		this = &Outcomes[V]{}
	}
	for _, other := range another.Buckets {
		this.Add(other.Weight, other.Value)
	}
	return this
}

/*
func (this *Outcomes[V]) Then(that *Outcomes[any], reducer func(v V, other any) V) (out *Outcomes[V]) {
	thisWeight := this.TotalWeight()
	otherWeight := that.TotalWeight()
	for _, v := range this.Buckets {
		for _, other := range that.Buckets {
			result := reducer(v.Value, other.Value)
			out = out.Add(other.Weight.DivBy(otherWeight).Times(thisWeight), result)
		}
	}
	return
}
*/

func Then[V any, U any, Z any](this *Outcomes[V], that *Outcomes[U], reducer func(v V, other U) Z) (out *Outcomes[Z]) {
	thisWeight := this.TotalWeight()
	otherWeight := that.TotalWeight()
	if this == nil || that == nil {
		panic("outcomes cannot be nil")
	}
	// log.Println("ThisWeight, otherWeight: ", thisWeight, otherWeight)
	for _, v := range this.Buckets {
		// log.Println("I, This: ", i, v)
		for _, other := range that.Buckets {
			outWeight := other.Weight.DivBy(otherWeight).Times(v.Weight.DivBy(thisWeight))
			// log.Println("j, other: ", j, other)
			// log.Println("newWeight: ", outWeight)
			result := reducer(v.Value, other.Value)
			out = out.Add(outWeight, result)
		}
	}
	return
}

func (o *Outcomes[V]) Partition(matcher func(v V) bool) (matched *Outcomes[V], unmatched *Outcomes[V]) {
	for _, v := range o.Buckets {
		if matcher(v.Value) {
			matched = matched.Add(v.Weight, v.Value)
		} else {
			unmatched = unmatched.Add(v.Weight, v.Value)
		}
	}
	return
}

func (o *Outcomes[V]) Add(weight any, value V) *Outcomes[V] {
	if o == nil {
		o = &Outcomes[V]{}
	}
	var fracWeight Fraction
	if val, ok := weight.(int64); ok {
		fracWeight = FracN(val)
	} else if val, ok := weight.(int); ok {
		fracWeight = FracN(int64(val))
	} else if val, ok := weight.(float64); ok {
		fracWeight = Fraction(val)
	} else if fracWeight, ok = weight.(Fraction); !ok {
		// TODO - caller must check or return error
		log.Fatalf("Invalid weight: %v.  Must be a int or a Fraction", weight)
	}
	o.Buckets = append(o.Buckets, Bucket[V]{fracWeight.Factorized(), value})
	return o
}

// Remove outcomes that do not match a filter criteria
func (o *Outcomes[V]) Filter(filter func(v Bucket[V]) bool) *Outcomes[V] {
	out := &Outcomes[V]{}
	for _, v := range o.Buckets {
		if filter(v) {
			out.Buckets = append(out.Buckets, v)
		}
	}
	// Note - re-normalizing probabilities may not be needed?
	return out
}
