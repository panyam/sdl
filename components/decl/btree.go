package decl

import "github.com/panyam/sdl/dsl"

// --- BTreeIndex ---

type BTreeIndex struct {
	IndexBase
	// BTree Specific Params (as Expr for flexibility, though likely Literals)
	Occupancy        dsl.Expr // Expr yielding FLOAT (0-1)
	AvgSplitPropCost dsl.Expr // Expr yielding FLOAT (avg extra R/W pairs)
	AvgMergePropCost dsl.Expr // Expr yielding FLOAT (avg extra R/W pairs)
}

func NewBTreeIndex(base IndexBase, occupancy, splitCost, mergeCost dsl.Expr) *BTreeIndex {
	return &BTreeIndex{IndexBase: base, Occupancy: occupancy, AvgSplitPropCost: splitCost, AvgMergePropCost: mergeCost}
}

// Find builds the  for a BTree find operation.
func (bt *BTreeIndex) Find() dsl.Expr {
	// Note: Height calculation needs NumPages -> NumRecords, RecordSize, PageSize.
	// This calculation is complex to represent purely in  without VM helpers.
	// We represent it as an internal call the VM resolves.
	heightExpr := &dsl.InternalCallExpr{
		FuncName: "CalculateBTreeHeight",
		Args:     []dsl.Expr{bt.NumRecords, bt.RecordSize, bt.PageSize, bt.NodeFanout},
	}

	// Cost of searching within a node (CPU)
	nodeSearchCpu := &dsl.InternalCallExpr{
		FuncName: "CalculateNodeSearchCPU",
		Args:     []dsl.Expr{bt.RecordProcessingTime, bt.NodeFanout},
	}

	// Cost of one level traversal (Read + Node Search CPU)
	read := bt.diskRead()
	levelCost := &dsl.AndExpr{Left: read, Right: nodeSearchCpu}

	//  instructing VM to repeat levelCost 'heightExpr' times sequentially
	return &dsl.InternalCallExpr{
		FuncName: "RepeatSequence", // VM primitive for repeating an expression
		Args: []dsl.Expr{
			levelCost,  // The expression to repeat
			heightExpr, // The number of times to repeat (as an expression)
		},
	}
}

// Helper  for one Read+Write pair cost
func (bt *BTreeIndex) readWritePair() dsl.Expr {
	return &dsl.AndExpr{Left: bt.diskRead(), Right: bt.diskWrite()}
}

// Helper  for average propagation cost (scaled R/W pair)
func (bt *BTreeIndex) avgPropagationCost(costFactorExpr dsl.Expr) dsl.Expr {
	return &dsl.InternalCallExpr{
		FuncName: "ScaleLatency", // VM primitive to scale latency
		Args: []dsl.Expr{
			bt.readWritePair(), // Input outcome expression
			costFactorExpr,     // Factor expression (e.g., bt.AvgSplitPropCost)
		},
	}
}

// Insert builds  for BTree Insert.
func (bt *BTreeIndex) Insert() dsl.Expr {
	find := bt.Find()
	modifyLeafCpu := &dsl.InternalCallExpr{FuncName: "GetRecordProcessingTime", Args: []dsl.Expr{bt.RecordProcessingTime}} // Base processing
	writeLeaf := bt.diskWrite()
	avgPropCost := bt.avgPropagationCost(bt.AvgSplitPropCost)

	// Find -> ModifyCPU -> WriteLeaf -> AvgPropCost
	step1 := &dsl.AndExpr{Left: find, Right: modifyLeafCpu}
	step2 := &dsl.AndExpr{Left: step1, Right: writeLeaf}
	step3 := &dsl.AndExpr{Left: step2, Right: avgPropCost}
	return step3
}

// Delete builds  for BTree Delete.
func (bt *BTreeIndex) Delete() dsl.Expr {
	find := bt.Find()
	modifyLeafCpu := &dsl.InternalCallExpr{FuncName: "GetRecordProcessingTime", Args: []dsl.Expr{bt.RecordProcessingTime}}
	writeLeaf := bt.diskWrite()
	avgPropCost := bt.avgPropagationCost(bt.AvgMergePropCost) // Use merge cost factor

	// Find -> ModifyCPU -> WriteLeaf -> AvgPropCost
	step1 := &dsl.AndExpr{Left: find, Right: modifyLeafCpu}
	step2 := &dsl.AndExpr{Left: step1, Right: writeLeaf}
	step3 := &dsl.AndExpr{Left: step2, Right: avgPropCost}
	return step3
}
