package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

type LSMTree struct {
	NWBase[*components.LSMTree]
}

func NewLSMTree(name string) *LSMTree {
	return &LSMTree{NWBase: NewNWBase(name, components.NewLSMTree())}
}

// Insert builds for LSMTree Insert.
func (b *LSMTree) Read() decl.Value {
	return OutcomesToValue(b.Wrapped.Read())
}

// Delete builds  for LSMTree Delete.
func (b *LSMTree) Write() decl.Value {
	return OutcomesToValue(b.Wrapped.Write())
}
