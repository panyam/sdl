// sdl/dsl/interpreter.go
package dsl

import (
	"errors" // Import errors package
	"fmt"
	// We will need core later, but not for phase 0 structure
	// "github.com/panyam/leetcoach/sdl/core"
)

var (
	ErrStackUnderflow       = errors.New("interpreter stack underflow")
	ErrNotImplemented       = errors.New("evaluation for this node type not implemented")
	ErrNotFound             = errors.New("identifier not found")
	ErrInternalFuncNotFound = errors.New("internal function not found")
	// Add more specific errors as needed
)

// InternalFunction defines the signature for built-in functions callable by the VM.
// It receives the interpreter instance (for potential access to env/stack) and evaluated arguments.
type InternalFunction func(interpreter *Interpreter, args []any) (any, error)

// Interpreter holds the state for evaluating an AST.
type Interpreter struct {
	stack         []any
	env           *Environment
	internalFuncs map[string]InternalFunction
	maxOutcomeLen int
}

// NewInterpreter creates a new interpreter instance.
func NewInterpreter(maxBuckets int) *Interpreter {
	if maxBuckets <= 0 {
		maxBuckets = 15 // Default if invalid
	}
	return &Interpreter{
		stack:         make([]any, 0, 10), // Initial capacity
		env:           NewEnvironment(),   // Start with a global environment
		internalFuncs: make(map[string]InternalFunction),
		maxOutcomeLen: maxBuckets,
	}
}

// Env returns the current environment the interpreter is using.
func (i *Interpreter) Env() *Environment {
	return i.env
}

// --- Stack Operations ---

func (i *Interpreter) push(val any) {
	i.stack = append(i.stack, val)
}

func (i *Interpreter) pop() (any, error) {
	if len(i.stack) == 0 {
		return nil, ErrStackUnderflow
	}
	lastIndex := len(i.stack) - 1
	val := i.stack[lastIndex]
	i.stack = i.stack[:lastIndex] // Slice off the last element
	return val, nil
}

// peek returns the top element without removing it (useful for debugging)
func (i *Interpreter) peek() (any, bool) {
	if len(i.stack) == 0 {
		return nil, false
	}
	return i.stack[len(i.stack)-1], true
}

// stackString provides a string representation of the stack for debugging
func (i *Interpreter) stackString() string {
	L := len(i.stack)
	items := make([]string, L)
	for idx, item := range i.stack {
		items[idx] = fmt.Sprintf("%v", item)
	}
	return fmt.Sprintf("Stack(len=%d): %v", L, items)
}

// --- Internal Function Registry ---

func (i *Interpreter) RegisterInternalFunc(name string, fn InternalFunction) {
	i.internalFuncs[name] = fn
}

// --- Evaluation (Stub) ---

// Eval is the main entry point for evaluating an AST node.
// It uses a type switch to delegate to specific eval methods.
// The result of an evaluation is typically left on the interpreter's stack.
// Returns the final result (often the top of the stack after full eval) and any error.
func (i *Interpreter) Eval(node Node) (any, error) {
	// fmt.Printf("Eval entry: %T - %s\n", node, node) // Debug entry
	var err error
	switch n := node.(type) {
	// We'll add cases here in subsequent phases
	case *LiteralExpr:
		err = i.evalLiteral(n)
	case *IdentifierExpr:
		err = i.evalIdentifier(n)
	case *InternalCallExpr:
		err = i.evalInternalCall(n)
	case *AndExpr:
		err = i.evalAndExpr(n)
	case *CallExpr:
		err = i.evalCallExpr(n)
	case *MemberAccessExpr:
		// Member access is often handled *within* evalCallExpr,
		// but we might need a stub if it can be evaluated alone.
		err = ErrNotImplemented // Placeholder
	case *RepeatExpr:
		err = i.evalRepeatExpr(n)
	// ... other node types ...

	default:
		return nil, fmt.Errorf("Eval not implemented for node type %T", node)
	}

	// fmt.Printf("Eval exit: %T - Err: %v, Stack: %s\n", node, err, i.stackString()) // Debug exit

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
func (i *Interpreter) GetFinalResult() (any, error) {
	if len(i.stack) == 0 {
		return nil, fmt.Errorf("cannot get final result: stack is empty")
	}
	if len(i.stack) > 1 {
		return nil, fmt.Errorf("cannot get final result: stack contains multiple items (%d)", len(i.stack))
	}
	return i.stack[0], nil
}

// ClearStack resets the stack for a new evaluation run.
func (i *Interpreter) ClearStack() {
	i.stack = i.stack[:0]
}

// --- Placeholder Eval methods (will be implemented in separate files/phases) ---

func (i *Interpreter) evalAndExpr(expr *AndExpr) error {
	// To be implemented in Phase 3
	return fmt.Errorf("evalAndExpr %w", ErrNotImplemented)
}

func (i *Interpreter) evalCallExpr(expr *CallExpr) error {
	// To be implemented in Phase 4
	return fmt.Errorf("evalCallExpr %w", ErrNotImplemented)
}

func (i *Interpreter) evalRepeatExpr(expr *RepeatExpr) error {
	// To be implemented in Phase 6
	return fmt.Errorf("evalRepeatExpr %w", ErrNotImplemented)
}

// Add stubs for other eval functions as needed...
