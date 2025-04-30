package dsl

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
func (d *Disk) Read() Expr {
	return &InternalCallExpr{
		FuncName: "GetDiskReadProfile",
		Args:     []Expr{&LiteralExpr{Kind: "STRING", Value: d.ProfileName}},
	}
}

// Write returns an  expression representing a disk write operation.
func (d *Disk) Write() Expr {
	return &InternalCallExpr{
		FuncName: "GetDiskWriteProfile",
		Args:     []Expr{&LiteralExpr{Kind: "STRING", Value: d.ProfileName}},
	}
}

// ReadProcessWrite returns an  expression for Read -> Process -> Write.
// processingTime is an Expr yielding Outcomes[Duration].
func (d *Disk) ReadProcessWrite(processingTime Expr) Expr {
	read := d.Read()
	write := d.Write()

	// Represents Read -> ProcessingTime -> Write sequentially
	step1 := &AndExpr{Left: read, Right: processingTime}
	step2 := &AndExpr{Left: step1, Right: write}
	return step2
}
