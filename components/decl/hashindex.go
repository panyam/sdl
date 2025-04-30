package decl

import "github.com/panyam/leetcoach/sdl/dsl"

// --- HashIndex ---

type HashIndex struct {
	IndexBase
	// Hash Specific Params (as Expr)
	AvgOverflowReads dsl.Expr // Expr yielding FLOAT (avg extra reads)
	ResizeCostFactor dsl.Expr // Expr yielding FLOAT (multiplier for resize R/W)
}

func NewHashIndex(base IndexBase, overflowReads, resizeFactor dsl.Expr) *HashIndex {
	return &HashIndex{IndexBase: base, AvgOverflowReads: overflowReads, ResizeCostFactor: resizeFactor}
}

// Helper  for average overflow read cost
func (h *HashIndex) avgOverflowReadCost() dsl.Expr {
	return &dsl.InternalCallExpr{
		FuncName: "ScaleLatency", // VM primitive
		Args: []dsl.Expr{
			h.diskRead(),       // Base is one disk read
			h.AvgOverflowReads, // Factor
		},
	}
}

// Find builds  for HashIndex Find.
// Simplified: HashCPU -> ReadPrimary -> AvgOverflowReadCost
func (h *HashIndex) Find() dsl.Expr {
	hashCpu := &dsl.InternalCallExpr{FuncName: "GetRecordProcessingTime", Args: []dsl.Expr{h.RecordProcessingTime}} // Simplification: Use base processing time
	readPrimary := h.diskRead()
	avgOverflow := h.avgOverflowReadCost()

	step1 := &dsl.AndExpr{Left: hashCpu, Right: readPrimary}
	step2 := &dsl.AndExpr{Left: step1, Right: avgOverflow}
	return step2
}

// Helper  for average resize cost
func (h *HashIndex) avgResizeCost() dsl.Expr {
	numPagesExpr := &dsl.InternalCallExpr{
		FuncName: "CalculateNumPages",
		Args:     []dsl.Expr{h.NumRecords, h.RecordSize, h.PageSize},
	}
	// Cost = (Read+Write) * NumPages * Factor
	// Represent as VM call: CalculateResizeCost(ReadOutcome, WriteOutcome, NumPagesExpr, FactorExpr)
	return &dsl.InternalCallExpr{
		FuncName: "CalculateResizeCost",
		Args: []dsl.Expr{
			h.diskRead(),
			h.diskWrite(),
			numPagesExpr,
			h.ResizeCostFactor,
		},
	}
}

// Insert builds  for HashIndex Insert.
// Simplified: Find -> ModifyCPU -> Write -> AvgResizeCost (Probabilistic choice omitted for  simplicity)
// VM would ideally handle the probabilistic addition of resize cost based on internal calculation.
// Here, we just add the *average* cost sequentially.
func (h *HashIndex) Insert() dsl.Expr {
	find := h.Find()
	modifyCpu := &dsl.InternalCallExpr{FuncName: "GetRecordProcessingTime", Args: []dsl.Expr{h.RecordProcessingTime}}
	write := h.diskWrite()
	avgResize := h.avgResizeCost() // Represents the *average* cost added by potential resize

	step1 := &dsl.AndExpr{Left: find, Right: modifyCpu}
	step2 := &dsl.AndExpr{Left: step1, Right: write}
	step3 := &dsl.AndExpr{Left: step2, Right: avgResize} // Add average resize cost
	return step3
}

// Delete builds  for HashIndex Delete.
// Simplified: Find -> ModifyCPU -> Write
func (h *HashIndex) Delete() dsl.Expr {
	find := h.Find()
	modifyCpu := &dsl.InternalCallExpr{FuncName: "GetRecordProcessingTime", Args: []dsl.Expr{h.RecordProcessingTime}}
	write := h.diskWrite()

	step1 := &dsl.AndExpr{Left: find, Right: modifyCpu}
	step2 := &dsl.AndExpr{Left: step1, Right: write}
	return step2
}
