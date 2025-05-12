package decl

import (
	"fmt"
	"testing"

	"github.com/panyam/sdl/core"
	"github.com/stretchr/testify/require"
)

// ParseLiteralValue converts a LiteralExpr value string to a basic Go type.
/*
func ParseLiteralValue(lit string) (any, error) {
	switch lit.Value.Type.Tag {
	case ValueTypeString:
		return lit.Value, nil
	case ValueTypeInt:
		return strconv.ParseInt(lit.Value, 10, 64)
	case ValueTypeFloat:
		return strconv.ParseFloat(lit.Value, 64)
	case ValueTypeBool:
		return strconv.ParseBool(lit.Value)
	// TODO: case "DURATION":
	default:
		return nil, fmt.Errorf("cannot parse literal kind %s yet", lit.Kind)
	}
}
*/

// Helper to create simple AST nodes for testing
func newIntLit(val int) *LiteralExpr {
	v, _ := NewRuntimeValue(IntType, val)
	return &LiteralExpr{Value: v}
}

func newBoolLit(val bool) *LiteralExpr {
	v, _ := NewRuntimeValue(BoolType, val)
	return &LiteralExpr{Value: v}
}

func newIdent(name string) *IdentifierExpr {
	return &IdentifierExpr{Name: name}
}

func newStringLit(val string) *LiteralExpr {
	v, _ := NewRuntimeValue(StrType, val)
	return &LiteralExpr{Value: v}
}

func newLetStmt(varName string, value Expr) *LetStmt {
	return &LetStmt{Variables: []*IdentifierExpr{newIdent(varName)}, Value: value}
}

func newBinExpr(left Expr, op string, right Expr) *BinaryExpr {
	return &BinaryExpr{Left: left, Operator: op, Right: right}
}

func newExprStmt(expr Expr) *ExprStmt {
	return &ExprStmt{Expression: expr}
}

func newBlockStmt(stmts ...Stmt) *BlockStmt {
	return &BlockStmt{Statements: stmts}
}

func newIfStmt(cond Expr, then *BlockStmt, elseStmt Stmt) *IfStmt { // elseStmt can be nil
	return &IfStmt{Condition: cond, Then: then, Else: elseStmt}
}

func newSysDecl(name string, body ...SystemDeclBodyItem) *SystemDecl {
	return &SystemDecl{NameNode: newIdent(name), Body: body}
}

func newInstDecl(name, compType string, overrides ...*AssignmentStmt) *InstanceDecl {
	return &InstanceDecl{NameNode: newIdent(name), ComponentType: newIdent(compType), Overrides: overrides}
}

func newAssignStmt(varName string, value Expr) *AssignmentStmt {
	return &AssignmentStmt{Var: newIdent(varName), Value: value}
}

// Helper for ComponentDecl AST
func newCompDecl(name string, body ...ComponentDeclBodyItem) *ComponentDecl {
	return &ComponentDecl{NameNode: newIdent(name), Body: body}
}

// Helper for UsesDecl AST
func newUsesDecl(varName, compType string) *UsesDecl {
	// Note: AST doesn't have overrides here, matches current struct
	return &UsesDecl{NameNode: newIdent(varName), ComponentNode: newIdent(compType)}
}

// Helper for ParamDecl with default value
func newParamDeclWithDefault(varName, typeName string, defaultVal Expr) *ParamDecl {
	p := newParamDecl(varName, typeName) // Use existing helper to create base param
	p.DefaultValue = defaultVal          // Set the default value expression
	return p
}

// Helper for ParamDecl AST (without default for now)
func newParamDecl(varName, typeName string) *ParamDecl {
	// Assuming TypeDecl handling can be simple for now
	tn := &TypeDecl{Name: typeName}
	return &ParamDecl{Name: newIdent(varName), Type: tn}
}

// Helper function to create a MethodDecl AST node for testing
func newMethodDecl(name string, params []*ParamDecl, returnType *TypeDecl, body *BlockStmt) *MethodDecl {
	return &MethodDecl{
		NameNode:   newIdent(name),
		Parameters: params,
		ReturnType: returnType, // Can be nil
		Body:       body,
	}
}

func newMemberAccessExpr(receiver Expr, memberName string) *MemberAccessExpr {
	return &MemberAccessExpr{
		Receiver: receiver,
		Member:   newIdent(memberName),
	}
}

// Helper to create a CallExpr AST node for testing
func newCallExpr(receiver Expr, methodName string, args ...Expr) *CallExpr {
	return &CallExpr{
		Function: &MemberAccessExpr{
			Receiver: receiver,
			Member:   newIdent(methodName),
		},
		Args: args,
	}
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
