package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

type BTreeIndex struct {
	NWBase[*components.BTreeIndex]
}

func NewBTreeIndex(name string) *BTreeIndex {
	return &BTreeIndex{NWBase: NewNWBase(name, components.NewBTreeIndex())}
}

// Find builds  for BTreeIndex Find.
func (b *BTreeIndex) Find() decl.Value {
	return OutcomesToValue(b.Wrapped.Find())
}

// Insert builds  for BTreeIndex Insert.
func (b *BTreeIndex) Insert() decl.Value {
	return OutcomesToValue(b.Wrapped.Insert())
}

// Delete builds  for BTreeIndex Delete.
func (b *BTreeIndex) Delete() decl.Value {
	return OutcomesToValue(b.Wrapped.Delete())
}
