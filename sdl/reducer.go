package sdl

// Sometimes in our outcome list we want to manage the explosion of buckets (eg consider a method
// call that has 10 buckets.  If method call is called twice then the resultant
// distribution would have 10 * 10 = 100 outcomes.  Doing just 5 times gets this to 10^5
// outcomes which is extrememly large.  Instead our caller could chose to just reduce this
// to 10 after say 200-300 outcomes.  Especially for large topologies this would help us
// manage outcome sizes and afterall we just need a distribution rather than something
// highly accurate.  There are several strategies here.

type Reducer[V any] interface {
	Reduce(input *Outcomes[V], numBuckets int) *Outcomes[V]
}
