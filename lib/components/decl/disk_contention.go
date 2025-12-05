package decl

import (
	"github.com/panyam/sdl/lib/components"
	"github.com/panyam/sdl/lib/decl"
)

type DiskWithContention struct {
	NWBase[*components.DiskWithContention]
}

func NewDiskWithContention() *DiskWithContention {
	return &DiskWithContention{NWBase: NewNWBase("DiskWithContention", components.NewDiskWithContention())}
}

// Read returns an expression representing a contention-aware disk read operation.
func (d *DiskWithContention) Read() decl.Value {
	return OutcomesToValue(d.Wrapped.Read())
}

// Write returns an expression representing a contention-aware disk write operation.
func (d *DiskWithContention) Write() decl.Value {
	return OutcomesToValue(d.Wrapped.Write())
}
