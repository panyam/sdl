package bitly

import "log"

type WValue[V any] struct {
	Weight Fraction
	Value  V
}

type Outcomes[V any] struct {
	Values []WValue[V]
}

func (o Outcomes[V]) Len() int {
	return len(o.Values)
}

func (o Outcomes[V]) TotalWeight() (out Fraction) {
	for _, v := range o.Values {
		out = out.Plus(v.Weight).Factorized()
	}
	return
}

func ToOutcomesAny(input any) *Outcomes[any] {
	return (input).(*Outcomes[any])
}

func Convert[V any, U any](this Outcomes[V], mapper func(v V) U) (out Outcomes[U]) {
	for _, v := range this.Values {
		out.Add(v.Weight, mapper(v.Value))
	}
	return
}

func (this *Outcomes[V]) Append(another *Outcomes[V]) *Outcomes[V] {
	for _, other := range another.Values {
		this.Add(other.Weight, other.Value)
	}
	return this
}

func (this *Outcomes[V]) Then(that *Outcomes[any], reducer func(v V, other any) V) (out Outcomes[V]) {
	thisWeight := this.TotalWeight()
	otherWeight := that.TotalWeight()
	for _, v := range this.Values {
		for _, other := range that.Values {
			result := reducer(v.Value, other.Value)
			out.Add(other.Weight.DivBy(otherWeight).Times(thisWeight), result)
		}
	}
	return
}

func Then[V any, U any, Z any](this Outcomes[V], that Outcomes[U], reducer func(v V, other U) Z) (out Outcomes[Z]) {
	thisWeight := this.TotalWeight()
	otherWeight := that.TotalWeight()
	for _, v := range this.Values {
		for _, other := range that.Values {
			result := reducer(v.Value, other.Value)
			out.Add(other.Weight.DivBy(otherWeight).Times(thisWeight), result)
		}
	}
	return
}

func (o Outcomes[V]) Partition(matcher func(v V) bool) (matched Outcomes[V], unmatched Outcomes[V]) {
	for _, v := range o.Values {
		if matcher(v.Value) {
			matched.Add(v.Weight, v.Value)
		} else {
			unmatched.Add(v.Weight, v.Value)
		}
	}
	return
}

func (o Outcomes[V]) Add(weight any, value V) {
	fracWeight, ok := weight.(Fraction)
	if !ok {
		// TODO - caller must check or return error
		log.Fatalf("Invalid weight: %v.  Must be a int or a Fraction", weight)
	}
	o.Values = append(o.Values, WValue[V]{fracWeight, value})
}
