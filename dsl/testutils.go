package dsl

import "strconv"

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

func LitStr(v string) Expr { return &LiteralExpr{Kind: "STRING", Value: v} }
func LitInt(v int) Expr    { return &LiteralExpr{Kind: "INT", Value: strconv.Itoa(v)} }
func LitFloat(v float64) Expr {
	return &LiteralExpr{Kind: "FLOAT", Value: strconv.FormatFloat(v, 'f', -1, 64)}
}
func LitBool(v bool) Expr            { return &LiteralExpr{Kind: "BOOL", Value: strconv.FormatBool(v)} }
func Ident(n string) Expr            { return &IdentifierExpr{Name: n} }
func Member(r Expr, m string) Expr   { return &MemberAccessExpr{Receiver: r, Member: m} }
func Call(f Expr, args ...Expr) Expr { return &CallExpr{Function: f, Args: args} }
func And(l, r Expr) Expr             { return &AndExpr{Left: l, Right: r} }
func Par(l, r Expr) Expr             { return &ParallelExpr{Left: l, Right: r} }
func InternalCall(fName string, args ...Expr) Expr {
	return &InternalCallExpr{FuncName: fName, Args: args}
}
