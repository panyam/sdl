package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

// --- ResourcePool (Stateless) ---
type ResourcePool struct {
	NWBase
	Wrapped components.ResourcePool
}

func NewResourcePool(name string) *ResourcePool {
	return &ResourcePool{NWBase: NewNWBase(name)}
}

// Acquire predicts queueing delay based on MMc model.
func (b *ResourcePool) Acquire() decl.Value {
	return OutcomesToValue(b.Wrapped.Acquire())
}
