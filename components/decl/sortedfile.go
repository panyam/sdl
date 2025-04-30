package decl

import "github.com/panyam/leetcoach/sdl/dsl"

// --- SortedFile ---
type SortedFile struct {
	IndexBase
}

func NewSortedFile(base IndexBase) *SortedFile {
	return &SortedFile{IndexBase: base}
}

// Scan for SortedFile (same as HeapFile Scan)
func (sf *SortedFile) Scan() dsl.Expr {
	numPagesExpr := &dsl.InternalCallExpr{FuncName: "CalculateNumPages", Args: []dsl.Expr{sf.NumRecords, sf.RecordSize, sf.PageSize}}
	recsPerPageExpr := &dsl.InternalCallExpr{FuncName: "CalculateRecsPerPage", Args: []dsl.Expr{sf.RecordSize, sf.PageSize}}
	procPageCost := &dsl.InternalCallExpr{FuncName: "ScaleLatency", Args: []dsl.Expr{sf.RecordProcessingTime, recsPerPageExpr}}
	readAndProcessPage := &dsl.AndExpr{Left: sf.diskRead(), Right: procPageCost}
	return &dsl.RepeatExpr{Input: readAndProcessPage, Count: numPagesExpr, Mode: dsl.Sequential}
}

// Find for SortedFile (Binary Search - Log N page reads + CPU per page)
func (sf *SortedFile) Find() dsl.Expr {
	numPagesExpr := &dsl.InternalCallExpr{FuncName: "CalculateNumPages", Args: []dsl.Expr{sf.NumRecords, sf.RecordSize, sf.PageSize}}
	logPagesExpr := &dsl.InternalCallExpr{FuncName: "Log2", Args: []dsl.Expr{numPagesExpr}} // VM needs Log2 function
	logRecsExpr := &dsl.InternalCallExpr{FuncName: "Log2", Args: []dsl.Expr{&dsl.InternalCallExpr{FuncName: "CalculateRecsPerPage", Args: []dsl.Expr{sf.RecordSize, sf.PageSize}}}}
	pageSearchCpu := &dsl.InternalCallExpr{FuncName: "ScaleLatency", Args: []dsl.Expr{sf.RecordProcessingTime, logRecsExpr}}
	readAndSearchPage := &dsl.AndExpr{Left: sf.diskRead(), Right: pageSearchCpu}

	// Repeat LogN times
	return &dsl.RepeatExpr{
		Input: readAndSearchPage,
		Count: logPagesExpr,
		Mode:  dsl.Sequential,
	}
}

// InsertAST/Delete for SortedFile (Expensive - Find + Shift ~ N/2 reads/writes?)
// Approximation: Find Cost + Significant additional IO cost related to N
func (sf *SortedFile) Insert() dsl.Expr {
	find := sf.Find()
	// Estimate shift cost as, say, reading/writing 1/4 of pages? Very heuristic.
	numPagesExpr := &dsl.InternalCallExpr{FuncName: "CalculateNumPages", Args: []dsl.Expr{sf.NumRecords, sf.RecordSize, sf.PageSize}}
	shiftPagesExpr := &dsl.InternalCallExpr{FuncName: "DivideInt", Args: []dsl.Expr{numPagesExpr, &dsl.LiteralExpr{Kind: "INT", Value: "4"}}}
	readWritePair := &dsl.AndExpr{Left: sf.diskRead(), Right: sf.diskWrite()}
	shiftCost := &dsl.RepeatExpr{Input: readWritePair, Count: shiftPagesExpr, Mode: dsl.Sequential}

	return &dsl.AndExpr{Left: find, Right: shiftCost}
}
func (sf *SortedFile) Delete() dsl.Expr {
	// Similar cost to Insert due to shifting
	return sf.Insert()
}
