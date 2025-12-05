package decl

import (
	"github.com/panyam/sdl/lib/components"
	"github.com/panyam/sdl/lib/decl"
)

type BitmapIndex struct {
	NWBase[*components.BitmapIndex]
}

func NewBitmapIndex(name string) *BitmapIndex {
	return &BitmapIndex{NWBase: NewNWBase(name, components.NewBitmapIndex())}
}

// Find builds  for BitmapIndex Find.
func (d *BitmapIndex) Find() decl.Value {
	return OutcomesToValue(d.Wrapped.Find())
}

// Insert builds  for BitmapIndex Insert.
func (d *BitmapIndex) Insert() decl.Value {
	return OutcomesToValue(d.Wrapped.Insert())
}

// Delete builds  for BitmapIndex Delete.
func (d *BitmapIndex) Delete() decl.Value {
	return OutcomesToValue(d.Wrapped.Delete())
}

// Update builds  for BitmapIndex Update.
func (d *BitmapIndex) Update() decl.Value {
	return OutcomesToValue(d.Wrapped.Update())
}
