package decl

import "github.com/panyam/leetcoach/sdl/dsl"

// --- HeapFile ---
type HeapFile struct {
	IndexBase // Embeds DiskName, NumRecords etc.
}

func NewHeapFile(base IndexBase) *HeapFile {
	return &HeapFile{IndexBase: base}
}

// Scan builds  for HeapFile Scan (Read N pages)
func (h *HeapFile) Scan() dsl.Expr {
	numPagesExpr := &dsl.InternalCallExpr{
		FuncName: "CalculateNumPages",
		Args:     []dsl.Expr{h.NumRecords, h.RecordSize, h.PageSize},
	}
	// Cost of reading one page + processing records on it
	recsPerPageExpr := &dsl.InternalCallExpr{FuncName: "CalculateRecsPerPage", Args: []dsl.Expr{h.RecordSize, h.PageSize}}
	procPageCost := &dsl.InternalCallExpr{FuncName: "ScaleLatency", Args: []dsl.Expr{h.RecordProcessingTime, recsPerPageExpr}} // Latency = ProcTime * RecsPerPage
	readAndProcessPage := &dsl.AndExpr{Left: h.diskRead(), Right: procPageCost}

	// Repeat readAndProcessPage N times sequentially
	return &dsl.RepeatExpr{ // Use RepeatExpr  node
		Input: readAndProcessPage,
		Count: numPagesExpr,
		Mode:  dsl.Sequential, // Scan is sequential
	}
}

// Insert for HeapFile (Simplified: RMW on one page?)
// Actual heap insert finds space, might be complex. Approximation: 1 Read + 1 Write.
func (h *HeapFile) Insert() dsl.Expr {
	// Simplification: Cost is like one Read + Proc + Write
	proc := &dsl.InternalCallExpr{FuncName: "GetRecordProcessingTime", Args: []dsl.Expr{h.RecordProcessingTime}}
	step1 := &dsl.AndExpr{Left: h.diskRead(), Right: proc}
	step2 := &dsl.AndExpr{Left: step1, Right: h.diskWrite()}
	return step2
}

// Find for HeapFile (Simplified: Scan ~half the pages)
func (h *HeapFile) Find() dsl.Expr {
	numPagesExpr := &dsl.InternalCallExpr{
		FuncName: "CalculateNumPages",
		Args:     []dsl.Expr{h.NumRecords, h.RecordSize, h.PageSize},
	}
	// Approx half the pages
	halfPagesExpr := &dsl.InternalCallExpr{FuncName: "DivideInt", Args: []dsl.Expr{numPagesExpr, &dsl.LiteralExpr{Kind: "INT", Value: "2"}}}

	recsPerPageExpr := &dsl.InternalCallExpr{FuncName: "CalculateRecsPerPage", Args: []dsl.Expr{h.RecordSize, h.PageSize}}
	procPageCost := &dsl.InternalCallExpr{FuncName: "ScaleLatency", Args: []dsl.Expr{h.RecordProcessingTime, recsPerPageExpr}}
	readAndProcessPage := &dsl.AndExpr{Left: h.diskRead(), Right: procPageCost}

	// Repeat readAndProcessPage N/2 times sequentially
	return &dsl.RepeatExpr{
		Input: readAndProcessPage,
		Count: halfPagesExpr,
		Mode:  dsl.Sequential,
	}
}

// Delete for HeapFile (Simplified: same as Find + Write?)
func (h *HeapFile) Delete() dsl.Expr {
	findCost := h.Find()       // Find the record/page
	writeCost := h.diskWrite() // Write the change
	return &dsl.AndExpr{Left: findCost, Right: writeCost}
}
