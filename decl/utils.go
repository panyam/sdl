package decl

import (
	"fmt"
	"strconv"
)

// ParseLiteralValue converts a LiteralExpr value string to a basic Go type.
func ParseLiteralValue(lit *LiteralExpr) (any, error) {
	switch lit.Kind {
	case "STRING":
		return lit.Value, nil
	case "INT":
		return strconv.ParseInt(lit.Value, 10, 64)
	case "FLOAT":
		return strconv.ParseFloat(lit.Value, 64)
	case "BOOL":
		return strconv.ParseBool(lit.Value)
	// TODO: case "DURATION":
	default:
		return nil, fmt.Errorf("cannot parse literal kind %s yet", lit.Kind)
	}
}

// Helper to create simple AST nodes for testing
func newIntLit(val string) *LiteralExpr {
	return &LiteralExpr{Kind: "INT", Value: val}
}

func newBoolLit(val string) *LiteralExpr {
	return &LiteralExpr{Kind: "BOOL", Value: val}
}

func newIdent(name string) *IdentifierExpr {
	return &IdentifierExpr{Name: name}
}

func newLetStmt(varName string, value Expr) *LetStmt {
	return &LetStmt{Variable: newIdent(varName), Value: value}
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
