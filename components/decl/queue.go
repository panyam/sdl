package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

// --- Queue (MMCK) ---
type Queue struct {
	NWBase[*components.Queue]
}

func NewQueue(name string) *Queue {
	return &Queue{NWBase: NewNWBase(name, components.NewMMCKQueue(name))}
}

// Inserts into the queue
func (b *Queue) Enqueue() decl.Value {
	return OutcomesToValue(b.Wrapped.Enqueue())
}

// Removes from the queue
func (b *Queue) Dequeue() decl.Value {
	return OutcomesOfDurationsToValue(b.Wrapped.Dequeue())
}
