package sdl

import (
	"log"
)

type Outcome = any

type Bucket[V Outcome] struct {
	Weight float64
	Value  V
}

type Outcomes[V Outcome] struct {
	Buckets []Bucket[V]
	And     func(a, b V) V
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

func Convert[V Outcome, U Outcome](this *Outcomes[V], mapper func(v V) U) (out *Outcomes[U]) {
	for _, v := range this.Buckets {
		out = out.Add(v.Weight, mapper(v.Value))
	}
	return
}

// Append another list of outcomes to ourselves.
func (this *Outcomes[V]) Append(rest ...*Outcomes[V]) *Outcomes[V] {
	for _, another := range rest {
		if another != nil {
			for _, other := range another.Buckets {
				this = this.Add(other.Weight, other.Value)
			}
		}
	}
	return this
}

// Remove outcomes that do not match a filter criteria In doing so, the weights are also
// adjusted.  For example if all outcomes with Latency < X are to be removed, and sum of
// all such outcomes amount to Y then all buckets left behind are scaled without Y
func (o *Outcomes[V]) Filter(filter func(v Bucket[V]) bool) (out *Outcomes[V], totalRemovedWeight float64) {
	totalWeight := 0.0
	out = &Outcomes[V]{And: o.And}
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
func (o *Outcomes[V]) Split(matcher func(v V) bool) (matched *Outcomes[V], unmatched *Outcomes[V]) {
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

func (o *Outcomes[V]) Partition(matchers ...func(v V) bool) (groups []*Outcomes[V], unmatched *Outcomes[V]) {
	if o != nil {
		groups = make([]*Outcomes[V], len(matchers)+1)
		for _, v := range o.Buckets {
			matched := false
			for index, matcher := range matchers {
				if matcher(v.Value) {
					groups[index] = groups[index].Add(v.Weight, v.Value)
					matched = true
					break
				}
			}
			if !matched {
				unmatched = unmatched.Add(v.Weight, v.Value)
			}
		}
	}
	return
}

// A sequential caller of X outcomes in sequence as long as they are the same
// type so they can use the same reducer
func (this *Outcomes[V]) Then(others ...*Outcomes[V]) (out *Outcomes[V]) {
	and := this.And
	for _, that := range others {
		this = And(this, that, and)
	}
	this.And = and
	return this
}

// Like an If statement but with multiple branches - like Switch/Case
/*
func (this *Outcomes[V]) When(cond func(V) bool, then *Outcomes[V], rest ...any) (out *Outcomes[V]) {
	var groups []*Outcomes[V]
	for i := -2; i < len(rest); i += 2 {
		if i >= 0 {
			cond = rest[i]
			cond, then = rest[i], rest[i+1]
		}
	}
	return out.Append(groups...)
}
*/

func Map[V Outcome, U Outcome](this *Outcomes[V], mapper func(v V) U) (out *Outcomes[U]) {
	for _, b := range this.Buckets {
		out = out.Add(b.Weight, mapper(b.Value))
	}
	return
}

func (this *Outcomes[V]) If(cond func(V) bool, then *Outcomes[V], otherwise *Outcomes[V], reducer ReducerFunc[V, V, V]) (out *Outcomes[V]) {
	// thisWeight := this.TotalWeight()
	thenWeight := then.TotalWeight()
	otherwiseWeight := otherwise.TotalWeight()
	if this == nil {
		panic("outcomes cannot be nil")
	}

	for _, bucket := range this.Buckets {
		matches := cond(bucket.Value)
		var other *Outcomes[V] = nil
		otherWeight := 0.0
		if matches && then != nil {
			other = then
			otherWeight = thenWeight
		} else if !matches && otherwise != nil {
			other = otherwise
			otherWeight = otherwiseWeight
		}
		if other == nil {
			// bodies are nil - so just add the bucket as is
			out = out.Add(bucket.Weight, bucket.Value)
		} else {
			for _, other := range other.Buckets {
				outWeight := other.Weight / otherWeight * bucket.Weight
				// outWeight := (other.Weight / otherWeight) * bucket.Weight / (thisWeight)
				// log.Println("j, other: ", j, other)
				// log.Println("newWeight: ", outWeight)
				result := reducer(bucket.Value, other.Value)
				out = out.Add(outWeight, result)
			}
		}
	}
	return
}

// Call two outcomes in sequence and return the outcomes of doing so
func And[V Outcome, U Outcome, Z Outcome](this *Outcomes[V], that *Outcomes[U], reducer ReducerFunc[V, U, Z]) (out *Outcomes[Z]) {
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
			// log.Printf("j, other: %p", reducer)
			// log.Println("newWeight: ", outWeight)
			result := reducer(v.Value, other.Value)
			out = out.Add(outWeight, result)
		}
	}
	return
}
