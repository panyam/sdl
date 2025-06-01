package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

type HeapFile struct {
	NWBase
	Wrapped components.HeapFile
}

func NewHeapFile(name string) *HeapFile {
	return &HeapFile{NWBase: NewNWBase(name)}
}

func (b *HeapFile) Scan() decl.Value {
	return OutcomesToValue(b.Wrapped.Scan())
}

// Find builds for HeapFile Find.
func (b *HeapFile) Find() decl.Value {
	return OutcomesToValue(b.Wrapped.Find())
}

// Insert builds for HeapFile Insert.
func (b *HeapFile) Insert() decl.Value {
	return OutcomesToValue(b.Wrapped.Insert())
}

// Delete builds  for HeapFile Delete.
func (b *HeapFile) Delete() decl.Value {
	return OutcomesToValue(b.Wrapped.Delete())
}
