package decl

import (
	"github.com/panyam/sdl/lib/components"
	"github.com/panyam/sdl/lib/decl"
)

type MM1Queue struct {
	NWBase[*components.MM1Queue]
}

func NewMM1Queue(name string) *MM1Queue {
	return &MM1Queue{NWBase: NewNWBase(name, components.NewMM1Queue(name))}
}

// Insert builds for MM1Queue Insert.
func (b *MM1Queue) Enqueue() decl.Value {
	return OutcomesToValue(b.Wrapped.Enqueue())
}

// Delete builds  for MM1Queue Delete.
func (b *MM1Queue) Dequeue() decl.Value {
	return OutcomesOfDurationsToValue(b.Wrapped.Dequeue())
}
