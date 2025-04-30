package decl

import "github.com/panyam/leetcoach/sdl/dsl"

// --- Base Index Struct (For common fields) ---
type IndexBase struct {
	Name string
	// References to dependencies/config - VM resolves these names
	DiskName             string
	RecordProcessingTime dsl.Expr // Expr yielding Outcomes[Duration] for base CPU work per record
	NumRecords           dsl.Expr // Expr yielding INT
	RecordSize           dsl.Expr // Expr yielding INT
	PageSize             dsl.Expr // Expr yielding INT
	NodeFanout           dsl.Expr // Expr yielding INT (used by BTree)
}

// Helper to create an  reference to the Disk dependency
func (ib *IndexBase) diskRead() dsl.Expr {
	return &dsl.CallExpr{
		Function: &dsl.MemberAccessExpr{
			Receiver: &dsl.IdentifierExpr{Name: ib.DiskName}, // VM resolves this
			Member:   "Read",
		},
	}
}
func (ib *IndexBase) diskWrite() dsl.Expr {
	return &dsl.CallExpr{
		Function: &dsl.MemberAccessExpr{
			Receiver: &dsl.IdentifierExpr{Name: ib.DiskName}, // VM resolves this
			Member:   "Write",
		},
	}
}
