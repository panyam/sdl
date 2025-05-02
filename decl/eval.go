package decl

import (
	"errors"
	"fmt"
)

type Value = any

var (
	ErrNotImplemented       = errors.New("evaluation for this node type not implemented")
	ErrNotFound             = errors.New("identifier not found")
	ErrInternalFuncNotFound = errors.New("internal function not found")
	ErrUnsupportedType      = errors.New("unsupported type for operation")
)

// Starts the execution of a single expression
func Eval(node Node, env *Env, v *VM) (val Value, err error) {
	// fmt.Printf("Eval entry: %T - %s\n", node, node) // Debug entry
	switch n := node.(type) {
	// We'll add cases here in subsequent phases
	case *LiteralExpr:
		return evalLiteral(n, env, v)
	case *IdentifierExpr:
		return evalIdentifier(n, env, v)
	case *CallExpr:
		return evalCall(n, env, v)
	case *MemberAccessExpr:
		// Member access is often handled *within* evalCallExpr,
		// but we might need a stub if it can be evaluated alone.
		return evalMemberAccess(n, env, v) // <-- Call the actual implementation

	// --- Statement Nodes ---
	case *BlockStmt:
		// When Eval is called directly on a BlockStmt (e.g., top level),
		// there is no initial context from an outer structure like IfStmt.
		// evalBlockStmt now returns the result directly, doesn't leave on stack implicitly
		return evalBlockStmt(n, NewEnv(env), nil) // Pass nil context
	case *AssignmentStmt:
		return evalAssignmentStmt(n, env, v)
	case *ExprStmt:
		return evalExprStmt(n, env, v)
	case *IfStmt: // <-- Will be implemented now
		return evalIfStmt(n, env, v)

	default:
		return nil, fmt.Errorf("Eval not implemented for node type %T", node)
	}
}

/** Evaluate a literal and return its value */
func evalLiteral(expr *LiteralExpr, env *Env, v *VM) (val Value, err error) {
	return
}

/** Evaluate a Identifier and return its value */
func evalIdentifier(expr *IdentifierExpr, env *Env, v *VM) (val Value, err error) {
	return
}

/** Evaluate a Call and return its value */
func evalCall(expr *CallExpr, env *Env, v *VM) (val Value, err error) {
	return
}

/** Evaluate a MemberAccess and return its value */
func evalMemberAccess(expr *MemberAccessExpr, env *Env, v *VM) (val Value, err error) {
	return
}

/** Evaluate a Block and return its value */
func evalBlockStmt(stmt *BlockStmt, env *Env, v *VM) (val Value, err error) {
	return
}

/** Evaluate a If and return its value */
func evalIfStmt(stmt *IfStmt, env *Env, v *VM) (val Value, err error) {
	return
}

/** Evaluate a Expr as a statement and return its value */
func evalExprStmt(stmt *ExprStmt, env *Env, v *VM) (val Value, err error) {
	return
}

/** Evaluate a Assignment as a statement and return its value */
func evalAssignmentStmt(stmt *AssignmentStmt, env *Env, v *VM) (val Value, err error) {
	return
}
