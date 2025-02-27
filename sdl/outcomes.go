package sdl

import (
	"log"
)

type Bucket[V any] struct {
	Weight float64
	Value  V
}

type Outcomes[V any] struct {
	Buckets []Bucket[V]
}

func (o *Outcomes[V]) Copy() *Outcomes[V] {
	out := Outcomes[V]{}
	for _, bucket := range o.Buckets {
		out.Buckets = append(out.Buckets, bucket)
	}
	return &out
}

func (o *Outcomes[V]) Len() int {
	if o == nil {
		return 0
	}
	return len(o.Buckets)
}

func (o *Outcomes[V]) TotalWeight() (out float64) {
	if o == nil {
		return 0
	}
	out = 0
	for _, v := range o.Buckets {
		out += v.Weight
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
	if another != nil {
		for _, other := range another.Buckets {
			this.Add(other.Weight, other.Value)
		}
	}
	return this
}

type CombineSequential[A any, B any, C any] = func(A, B) C

func Then[V any, U any, Z any](this *Outcomes[V], that *Outcomes[U], reducer CombineSequential[V, U, Z]) (out *Outcomes[Z]) {
	thisWeight := this.TotalWeight()
	otherWeight := that.TotalWeight()
	if this == nil || that == nil {
		panic("outcomes cannot be nil")
	}
	// log.Println("ThisWeight, otherWeight: ", thisWeight, otherWeight)
	for _, v := range this.Buckets {
		// log.Println("I, This: ", i, v)
		for _, other := range that.Buckets {
			outWeight := other.Weight / otherWeight * (v.Weight / (thisWeight))
			// log.Println("j, other: ", j, other)
			// log.Println("newWeight: ", outWeight)
			result := reducer(v.Value, other.Value)
			out = out.Add(outWeight, result)
		}
	}
	return
}

// Remove outcomes that do not match a filter criteria
// In doing so, the weights are also adjusted.  For example if all outcomes with Latency < X are
// to be removed, and sum of all such outcomes amount to Y then all buckets left behind are scaled
// without Y
func (o *Outcomes[V]) Filter(filter func(v Bucket[V]) bool) (out *Outcomes[V], totalRemovedWeight float64) {
	totalWeight := 0.0
	out = &Outcomes[V]{}
	for _, v := range o.Buckets {
		if filter(v) {
			out.Buckets = append(out.Buckets, v)
			totalWeight = totalWeight + (v.Weight)
		} else {
			totalRemovedWeight = totalRemovedWeight + (v.Weight)
		}
	}
	// Renormalize probabilities - is this needed?
	// for _, v := range out.Buckets { v.Weight /= totalWeight }
	return
}

// Partition into outcomes that match a certain cond vs those that do not
func (o *Outcomes[V]) Partition(matcher func(v V) bool) (matched *Outcomes[V], unmatched *Outcomes[V]) {
	if o != nil {
		for _, v := range o.Buckets {
			if matcher(v.Value) {
				matched = matched.Add(v.Weight, v.Value)
			} else {
				unmatched = unmatched.Add(v.Weight, v.Value)
			}
		}
	}
	return
}

func (o *Outcomes[V]) Add(weight any, value V) *Outcomes[V] {
	if o == nil {
		o = &Outcomes[V]{}
	}
	var fracWeight float64
	if val, ok := weight.(int64); ok {
		fracWeight = float64(val)
	} else if val, ok := weight.(int); ok {
		fracWeight = float64(val)
	} else if val, ok := weight.(float64); ok {
		fracWeight = float64(val)
	} else if fracWeight, ok = weight.(float64); !ok {
		// TODO - caller must check or return error
		log.Fatalf("Invalid weight: %v.  Must be a int or a float64", weight)
	}
	o.Buckets = append(o.Buckets, Bucket[V]{fracWeight, value})
	return o
}
