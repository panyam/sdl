package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

type NetworkLink struct {
	NWBase
	Wrapped components.NetworkLink
}

func NewNetworkLink(name string) *NetworkLink {
	return &NetworkLink{NWBase: NewNWBase(name)}
}

// Insert builds for NetworkLink Insert.
func (b *NetworkLink) Transfer() decl.Value {
	return OutcomesToValue(b.Wrapped.Transfer())
}
