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

func newStringLit(val string) *LiteralExpr {
	return &LiteralExpr{Kind: "STRING", Value: val}
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

func newSysDecl(name string, body ...SystemDeclBodyItem) *SystemDecl {
	return &SystemDecl{Name: newIdent(name), Body: body}
}

func newInstDecl(name, compType string, overrides ...*AssignmentStmt) *InstanceDecl {
	return &InstanceDecl{Name: newIdent(name), ComponentType: newIdent(compType), Overrides: overrides}
}

func newAssignStmt(varName string, value Expr) *AssignmentStmt {
	return &AssignmentStmt{Var: newIdent(varName), Value: value}
}

// Helper for ComponentDecl AST
func newCompDecl(name string, body ...ComponentDeclBodyItem) *ComponentDecl {
	return &ComponentDecl{Name: newIdent(name), Body: body}
}

// Helper for UsesDecl AST
func newUsesDecl(varName, compType string) *UsesDecl {
	// Note: AST doesn't have overrides here, matches current struct
	return &UsesDecl{Name: newIdent(varName), ComponentType: newIdent(compType)}
}

// Helper for ParamDecl AST (without default for now)
func newParamDecl(varName, typeName string) *ParamDecl {
	// Assuming TypeName handling can be simple for now
	tn := &TypeName{}
	switch typeName {
	case "string":
		tn.PrimitiveTypeName = typeName
	case "int":
		tn.PrimitiveTypeName = typeName
	case "float":
		tn.PrimitiveTypeName = typeName
	case "bool":
		tn.PrimitiveTypeName = typeName
	case "duration":
		tn.PrimitiveTypeName = typeName
	default:
		tn.EnumTypeName = typeName // Assume others are enums/custom
	}
	return &ParamDecl{Name: newIdent(varName), Type: tn}
}
