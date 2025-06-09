package decl

import (
	"fmt"
	"testing"

	"github.com/panyam/sdl/core"
	"github.com/stretchr/testify/require"
)

var ZL = Location{}

// ParseLiteralValue converts a LiteralExpr value string to a basic Go type.
func NewNodeInfo(start, end Location) NodeInfo {
	return NodeInfo{StartPos: start, StopPos: end}
}

// Helper to create simple AST nodes for testing
func NewIntLit(val int) *LiteralExpr {
	v, _ := NewValue(IntType, val)
	return &LiteralExpr{Value: v}
}

func NewBoolLit(val bool) *LiteralExpr {
	v, _ := NewValue(BoolType, val)
	return &LiteralExpr{Value: v}
}

func NewStringLit(val string) *LiteralExpr {
	v, _ := NewValue(StrType, val)
	return &LiteralExpr{Value: v}
}

func NewIdent(name string) *IdentifierExpr {
	return NewIdentExpr(name, ZL, ZL)
}

func NewIdentExpr(name string, start, end Location) *IdentifierExpr {
	return &IdentifierExpr{
		ExprBase: ExprBase{NodeInfo: NewNodeInfo(start, end)},
		Value:    name,
	}
}

func NewLetStmt(varName string, value Expr) *LetStmt {
	return &LetStmt{Variables: []*IdentifierExpr{NewIdent(varName)}, Value: value}
}

func NewNewExpr(compDecl *ComponentDecl) *NewExpr {
	return &NewExpr{ComponentDecl: compDecl}
}

func NewBinExpr(left Expr, op string, right Expr) *BinaryExpr {
	return &BinaryExpr{Left: left, Operator: op, Right: right}
}

func NewUnaryExpr(op string, right Expr) *UnaryExpr {
	return &UnaryExpr{Operator: op, Right: right}
}

func NewExprStmt(expr Expr) *ExprStmt {
	return &ExprStmt{Expression: expr}
}

func NewReturnStmt(returnValue Expr) *ReturnStmt {
	// NodeInfo is cleaned
	return &ReturnStmt{ReturnValue: returnValue}
}

func NewBlockStmt(stmts ...Stmt) *BlockStmt {
	return &BlockStmt{Statements: stmts}
}

func NewIfStmt(cond Expr, then *BlockStmt, elseStmt Stmt) *IfStmt { // elseStmt can be nil
	return &IfStmt{Condition: cond, Then: then, Else: elseStmt}
}

func NewSysDecl(name string, body ...SystemDeclBodyItem) *SystemDecl {
	return &SystemDecl{Name: NewIdent(name), Body: body}
}

func NewInstDecl(name, compType string, overrides ...*AssignmentStmt) *InstanceDecl {
	return &InstanceDecl{Name: NewIdent(name), ComponentName: NewIdent(compType), Overrides: overrides}
}

func NewAssignmentStmt(varName string, value Expr) *AssignmentStmt {
	return &AssignmentStmt{Var: NewIdent(varName), Value: value}
}

func NewAssignStmt(varName string, value Expr) *AssignmentStmt {
	return &AssignmentStmt{Var: NewIdent(varName), Value: value}
}

// Helper for ComponentDecl AST
func NewCompDecl(name string, isNative bool, body ...ComponentDeclBodyItem) *ComponentDecl {
	return &ComponentDecl{Name: NewIdent(name), IsNative: isNative, Body: body}
}

// Helper for UsesDecl AST
func NewUsesDecl(varName, compType string) *UsesDecl {
	// Note: AST doesn't have overrides here, matches current struct
	return &UsesDecl{Name: NewIdent(varName), ComponentName: NewIdent(compType)}
}

// Helper function to create a MethodDecl AST node for testing
func NewMethodDecl(name string, params []*ParamDecl, returnType *TypeDecl, body *BlockStmt) *MethodDecl {
	if params == nil {
		params = []*ParamDecl{}
	}
	return &MethodDecl{
		Name:       NewIdent(name),
		Parameters: params,
		ReturnType: returnType, // Can be nil
		Body:       body,
	}
}

func NewTypeDecl(name string, args []*TypeDecl) *TypeDecl {
	// NodeInfo is cleaned
	return &TypeDecl{Name: name, Args: args}
}

// Helper for ParamDecl AST
func NewParamDecl(name string, typeName *TypeDecl, defaultValue Expr) *ParamDecl {
	// NodeInfo is cleaned
	return &ParamDecl{Name: NewIdent(name), TypeDecl: typeName, DefaultValue: defaultValue}
}

func NewMemberAccessExpr(receiver Expr, memberName string) *MemberAccessExpr {
	return &MemberAccessExpr{
		Receiver: receiver,
		Member:   NewIdent(memberName),
	}
}

// Helper to create a CallExpr AST node for testing
func NewCallExpr2(receiver Expr, methodName string, args ...Expr) *CallExpr {
	return &CallExpr{
		Function: &MemberAccessExpr{
			Receiver: receiver,
			Member:   NewIdent(methodName),
		},
		Args: args,
	}
}

func NewCallExpr(fn Expr, args ...Expr) *CallExpr {
	// NodeInfo is cleaned by cleanNodeInfo
	return &CallExpr{Function: fn, Args: args}
}

func NewDistributeExpr(total Expr, defaultCase Expr, cases ...*CaseExpr) *DistributeExpr {
	return &DistributeExpr{TotalProb: total, Cases: cases, Default: defaultCase}
}

func NewGoExpr(stmt Stmt, expr Expr) *GoExpr {
	return &GoExpr{Stmt: stmt, Expr: expr}
}

func NewDelayStmt(duration Expr) *DelayStmt {
	return &DelayStmt{Duration: duration}
}

func NewWaitExpr(idents ...*IdentifierExpr) *WaitExpr {
	return &WaitExpr{FutureNames: idents}
}

// Helper assertion for BinaryOpNode structure
func assertBinaryOpNode(t *testing.T, node OpNode, expectedOp string) *BinaryOpNode {
	t.Helper()
	binOp, ok := node.(*BinaryOpNode)
	require.True(t, ok, "Expected *BinaryOpNode, got %T", node)
	require.Equal(t, expectedOp, binOp.Op, "Binary operator mismatch")
	return binOp
}

// outcomeToVarState converts a result from a native Go component call
// (typically *core.Outcomes[V]) into the dual-track *VarState needed
// by the evaluator.
func outcomeToVarState(outcome any) (*VarState, error) {
	if outcome == nil {
		return createNilState(), nil
	}

	switch o := outcome.(type) {
	case *core.Outcomes[core.AccessResult]:
		// Special case: AccessResult contains both Success (Value) and Latency
		// Need to split this into two separate Outcomes tracks.
		if o == nil || o.Len() == 0 {
			return createNilState(), nil // Or maybe an error state? Nil seems okay.
		}
		valOutcome := &core.Outcomes[bool]{}
		latOutcome := &core.Outcomes[core.Duration]{}
		totalWeight := o.TotalWeight() // Normalize weights if needed? Assume input is normalized.
		if totalWeight == 0 {
			totalWeight = 1.0
		}

		for _, bucket := range o.Buckets {
			// Weight remains the same for both tracks derived from the same source bucket
			weight := bucket.Weight // / totalWeight (If normalization needed)
			valOutcome.Add(weight, bucket.Value.Success)
			latOutcome.Add(weight, bucket.Value.Latency)
		}
		return &VarState{ValueOutcome: valOutcome, LatencyOutcome: latOutcome}, nil

	// case *core.Outcomes[float64]:
	case *core.Outcomes[core.Duration]:
		// Pure duration outcome - maps directly to Latency track
		// Value track becomes neutral identity (e.g., bool true)
		if o == nil || o.Len() == 0 {
			return createNilState(), nil
		}
		return &VarState{ValueOutcome: IdentityValueOutcome(), LatencyOutcome: o}, nil

	case *core.Outcomes[bool]:
		// Pure boolean outcome - maps directly to Value track
		// Latency track becomes zero
		if o == nil || o.Len() == 0 {
			return createNilState(), nil
		}
		return &VarState{ValueOutcome: o, LatencyOutcome: ZeroLatencyOutcome()}, nil

	case *core.Outcomes[int64]:
		if o == nil || o.Len() == 0 {
			return createNilState(), nil
		}
		return &VarState{ValueOutcome: o, LatencyOutcome: ZeroLatencyOutcome()}, nil

		// Add other specific *core.Outcomes[T] types as needed

	default:
		// How to handle unknown types? Error for now.
		// Could potentially try reflection to see if it's *some* *core.Outcomes[T]
		// but that's complex.
		return nil, fmt.Errorf("unsupported outcome type %T for conversion to VarState", outcome)
	}
}
