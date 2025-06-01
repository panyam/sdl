package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

type Disk struct {
	NWBase
	Wrapped components.Disk
}

func NewDisk(name string) *Disk {
	return &Disk{NWBase: NewNWBase(name)}
}

// Read returns an  expression representing a disk read operation.
// It instructs the VM to look up the read profile based on d.ProfileName.
func (d *Disk) Read() decl.Value {
	return OutcomesToValue(h.Wrapped.Read())
}

// Write returns an  expression representing a disk write operation.
func (d *Disk) Write() decl.Value {
	return OutcomesToValue(h.Wrapped.Write())
}

// ReadProcessWrite returns an  expression for Read -> Process -> Write.
// processingTime is an Expr yielding Outcomes[Duration].
func (d *Disk) ReadProcessWrite(processingTimeVal decl.Value) decl.Value {
	processingTime := processingTimeVal.FloatVal()

	// Represents Read -> ProcessingTime -> Write sequentially
	result := d.Wrapped.ReadProcessWrite(processingTime)
	return OutcomesToValue(result)
}
