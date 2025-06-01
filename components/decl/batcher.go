// sdl/decl/components.go (or separate files like decl/disk.go, decl/index.go etc.)
package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

type Batcher struct {
	NWBase
	Wrapped components.Batcher
}

func NewBatcher(name string) *Batcher {
	return &Batcher{NWBase: NewNWBase(name)}
}

// Submit generates  for submitting one item.
func (d *Batcher) Submit() decl.Value {
	return OutcomesToValue(d.Wrapped.Submit())
}
