package sdl

import (
	"log"
	"testing"
)

func TestOutcomes(t *testing.T) {
	type V struct {
		Success bool
		Latency TimeValue
	}
	var o Outcomes[V]
	o.
		Add(.8, V{true, Val(1, MilliSeconds)}).
		Add(0.2, V{false, Val(10, MilliSeconds)})
	log.Println("O: ", o.Values)

	o1 := Then(&o, &o, func(a, b V) V {
		return V{a.Success && b.Success, a.Latency.Add(b.Latency)}
	})
	log.Println("o1 results: ", len(o1.Values), o1.Values)
}
