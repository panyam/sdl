package decl

import "github.com/panyam/sdl/dsl"

// --- BitmapIndex ---

type BitmapIndex struct {
	IndexBase
	// Bitmap Specific Params (as Expr)
	Cardinality      dsl.Expr // Expr yielding INT
	IsCompressed     dsl.Expr // Expr yielding BOOL
	UpdateCostFactor dsl.Expr // Expr yielding FLOAT
	QuerySelectivity dsl.Expr // Expr yielding FLOAT (0-1)
}

func NewBitmapIndex(base IndexBase, card, isComp, updateFactor, selectivity dsl.Expr) *BitmapIndex {
	return &BitmapIndex{IndexBase: base, Cardinality: card, IsCompressed: isComp, UpdateCostFactor: updateFactor, QuerySelectivity: selectivity}
}

// Find builds  for BitmapIndex Find.
func (bmi *BitmapIndex) Find() dsl.Expr {
	loadIndex := bmi.diskRead()

	// Assume avg 3 ops, VM calculates cost based on compression flag
	bitwiseOp := &dsl.InternalCallExpr{
		FuncName: "CalculateBitwiseOpCost",
		Args: []dsl.Expr{
			&dsl.LiteralExpr{Kind: "INT", Value: "3"}, // Num ops
			bmi.IsCompressed,
		},
	}

	// VM calculates result processing cost based on selectivity
	resultProcessing := &dsl.InternalCallExpr{
		FuncName: "CalculateResultProcessingCost",
		Args: []dsl.Expr{
			bmi.NumRecords,
			bmi.QuerySelectivity,
			bmi.RecordProcessingTime,
		},
	}

	// Load -> Bitwise -> Result Processing
	step1 := &dsl.AndExpr{Left: loadIndex, Right: bitwiseOp}
	step2 := &dsl.AndExpr{Left: step1, Right: resultProcessing}
	return step2
}

// modifyBitmapCost builds  for the core R-M-W cycle used in updates.
func (bmi *BitmapIndex) modifyBitmapCost() dsl.Expr {
	read := bmi.diskRead()
	modifyCpu := &dsl.InternalCallExpr{FuncName: "GetRecordProcessingTime", Args: []dsl.Expr{bmi.RecordProcessingTime}} // Base cost
	write := bmi.diskWrite()

	// R -> M -> W
	step1 := &dsl.AndExpr{Left: read, Right: modifyCpu}
	step2 := &dsl.AndExpr{Left: step1, Right: write}

	// Scale the result by the UpdateCostFactor
	return &dsl.InternalCallExpr{
		FuncName: "ScaleLatency",
		Args: []dsl.Expr{
			step2,                // Input outcome expr
			bmi.UpdateCostFactor, // Factor expr
		},
	}
}

// Insert builds  for BitmapIndex Insert.
func (bmi *BitmapIndex) Insert() dsl.Expr {
	return bmi.modifyBitmapCost()
}

// Delete builds  for BitmapIndex Delete.
func (bmi *BitmapIndex) Delete() dsl.Expr {
	return bmi.modifyBitmapCost()
}

// Update builds  for BitmapIndex Update.
func (bmi *BitmapIndex) Update() dsl.Expr {
	// Often slightly more than insert/delete, but UpdateCostFactor handles this average.
	return bmi.modifyBitmapCost()
}
