package decl

import (
	"github.com/panyam/sdl/lib/components"
	"github.com/panyam/sdl/lib/decl"
)

type HashIndex struct {
	NWBase[*components.HashIndex]
}

func NewHashIndex(name string) *HashIndex {
	return &HashIndex{NWBase: NewNWBase(name, components.NewHashIndex())}
}

func (h *HashIndex) Find() (v decl.Value) {
	return OutcomesToValue(h.Wrapped.Find())
}

func (h *HashIndex) Insert() (v decl.Value) {
	return OutcomesToValue(h.Wrapped.Insert())
}

func (h *HashIndex) Delete() (v decl.Value) {
	return OutcomesToValue(h.Wrapped.Delete())
}
