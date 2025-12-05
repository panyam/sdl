package decl

import (
	"github.com/panyam/sdl/lib/components"
	"github.com/panyam/sdl/lib/decl"
)

// --- SortedFile (Stateless) ---
type SortedFile struct {
	NWBase[*components.SortedFile]
}

func NewSortedFile(name string) *SortedFile {
	return &SortedFile{NWBase: NewNWBase(name, components.NewSortedFile())}
}

func (b *SortedFile) Scan() decl.Value {
	return OutcomesToValue(b.Wrapped.Scan())
}

func (b *SortedFile) Find() decl.Value {
	return OutcomesToValue(b.Wrapped.Find())
}

func (b *SortedFile) Delete() decl.Value {
	return OutcomesToValue(b.Wrapped.Delete())
}
