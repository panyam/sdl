// sdl/dsl/vm.go
package dsl

import (
	"errors" // Import errors package
	"fmt"

	"github.com/panyam/leetcoach/sdl/decl"
	// We will need core later, but not for phase 0 structure
	// "github.com/panyam/leetcoach/sdl/core"
)

var (
	ErrStackUnderflow = errors.New("vm stack underflow")
	// Add more specific errors as needed
)

// --- Reducer Registry ---

// VM holds the state for evaluating an AST.
type VM struct {
	decl.VM
	stack []any
}

// NewVM creates a new vm instance.
func (v *VM) Init() {
	v.stack = make([]any, 0, 10) // Initial capacity
	v.VM.Init()
}

// --- Stack Operations ---

func (v *VM) push(val any) {
	v.stack = append(v.stack, val)
}

func (v *VM) pop() (any, error) {
	if len(v.stack) == 0 {
		return nil, ErrStackUnderflow
	}
	lastIndex := len(v.stack) - 1
	val := v.stack[lastIndex]
	v.stack = v.stack[:lastIndex] // Slice off the last element
	return val, nil
}

// peek returns the top element without removing it (useful for debugging)
func (v *VM) peek() (any, bool) {
	if len(v.stack) == 0 {
		return nil, false
	}
	return v.stack[len(v.stack)-1], true
}

// stackString provides a string representation of the stack for debugging
func (v *VM) stackString() string {
	L := len(v.stack)
	items := make([]string, L)
	for idx, item := range v.stack {
		items[idx] = fmt.Sprintf("%v", item)
	}
	return fmt.Sprintf("Stack(len=%d): %v", L, items)
}

// --- Evaluation (Stub) ---

// Eval is the main entry point for evaluating an AST node.
// It uses a type switch to delegate to specific eval methods.
// The result of an evaluation is typically left on the vm's stack.
// Returns the final result (often the top of the stack after full eval) and any error.
func (v *VM) Eval(node decl.Node) (any, error) {
	// fmt.Printf("Eval entry: %T - %s\n", node, node) // Debug entry
	var err error
	switch n := node.(type) {
	// We'll add cases here in subsequent phases
	case *decl.LiteralExpr:
		err = v.evalLiteral(n)
	case *decl.IdentifierExpr:
		err = v.evalIdentifier(n)
	case *decl.InternalCallExpr:
		err = v.evalInternalCall(n)
	case *decl.AndExpr:
		err = v.evalAndExpr(n)
	case *decl.CallExpr:
		err = v.evalCallExpr(n)
	case *decl.MemberAccessExpr:
		// Member access is often handled *within* evalCallExpr,
		// but we might need a stub if it can be evaluated alone.
		err = v.evalMemberAccessExpr(n) // <-- Call the actual implementation
	case *decl.RepeatExpr:
		err = v.evalRepeatExpr(n)

	// --- Statement Nodes ---
	case *decl.BlockStmt:
		// When Eval is called directly on a BlockStmt (e.g., top level),
		// there is no initial context from an outer structure like IfStmt.
		// evalBlockStmt now returns the result directly, doesn't leave on stack implicitly
		blockResult, evalErr := v.evalBlockStmt(n, v.env, nil) // Pass nil context
		// Only push if no error and result is non-nil (avoids pushing nil return signal value)
		if evalErr == nil && blockResult != nil {
			v.push(blockResult) // Push the final result of the block evaluation
		}
		err = evalErr // Assign any error from the block execution
	case *decl.AssignmentStmt:
		err = v.evalAssignmentStmt(n)
	case *decl.ReturnStmt:
		err = v.evalReturnStmt(n) // Return signals via special error/value
	case *decl.ExprStmt:
		err = v.evalExprStmt(n)
	case *decl.IfStmt: // <-- Will be implemented now
		err = v.evalIfStmt(n)

	default:
		return nil, fmt.Errorf("Eval not implemented for node type %T", node)
	}

	// fmt.Printf("Eval exit: %T - Err: %v, Stack: %s\n", node, err, v.stackString()) // Debug exit

	if err != nil {
		return nil, err
	}

	// After evaluating a top-level node, the final result should be on the stack.
	// However, for recursive calls, intermediate results are left.
	// Let's return the top of the stack only if the stack has exactly one item
	// after the Eval completes for the *top-level* node. The caller of the
	// top-level Eval can decide what to do with the final stack state.
	// For now, just return nil result (caller can inspect stack).
	// Final result retrieval might be a separate method like `GetResult()`.
	return nil, nil // Result is on the stack
}

// GetFinalResult attempts to retrieve the single final result from the stack.
// Returns an error if the stack is empty or contains more than one item.
func (v *VM) GetFinalResult() (any, error) {
	if len(v.stack) == 0 {
		return nil, fmt.Errorf("cannot get final result: stack is empty")
	}
	if len(v.stack) > 1 {
		return nil, fmt.Errorf("cannot get final result: stack contains multiple items (%d)", len(v.stack))
	}
	return v.stack[0], nil
}

// ClearStack resets the stack for a new evaluation run.
func (v *VM) ClearStack() {
	v.stack = v.stack[:0]
}

// Add stubs for other eval functions as needed...
