package decl

import "github.com/panyam/sdl/dsl"

type Disk struct {
	Name        string
	ProfileName string // e.g., "SSD", "HDD" - VM uses this to lookup profile
}

func NewDisk(name, profile string) *Disk {
	if profile == "" {
		profile = "SSD" // Default
	}
	return &Disk{Name: name, ProfileName: profile}
}

// Read returns an  expression representing a disk read operation.
// It instructs the VM to look up the read profile based on d.ProfileName.
func (d *Disk) Read() dsl.Expr {
	return &dsl.InternalCallExpr{
		FuncName: "GetDiskReadProfile",
		Args:     []dsl.Expr{&dsl.LiteralExpr{Kind: "STRING", Value: d.ProfileName}},
	}
}

// Write returns an  expression representing a disk write operation.
func (d *Disk) Write() dsl.Expr {
	return &dsl.InternalCallExpr{
		FuncName: "GetDiskWriteProfile",
		Args:     []dsl.Expr{&dsl.LiteralExpr{Kind: "STRING", Value: d.ProfileName}},
	}
}

// ReadProcessWrite returns an  expression for Read -> Process -> Write.
// processingTime is an Expr yielding Outcomes[Duration].
func (d *Disk) ReadProcessWrite(processingTime dsl.Expr) dsl.Expr {
	read := d.Read()
	write := d.Write()

	// Represents Read -> ProcessingTime -> Write sequentially
	step1 := &dsl.AndExpr{Left: read, Right: processingTime}
	step2 := &dsl.AndExpr{Left: step1, Right: write}
	return step2
}
