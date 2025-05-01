// sdl/dsl/vm.go
package dsl

import (
	"errors" // Import errors package
	"fmt"
	"reflect"

	"github.com/panyam/leetcoach/sdl/core"
	// We will need core later, but not for phase 0 structure
	// "github.com/panyam/leetcoach/sdl/core"
)

var (
	ErrStackUnderflow       = errors.New("vm stack underflow")
	ErrNotImplemented       = errors.New("evaluation for this node type not implemented")
	ErrNotFound             = errors.New("identifier not found")
	ErrInternalFuncNotFound = errors.New("internal function not found")
	// Add more specific errors as needed
)

// InternalFunction defines the signature for built-in functions callable by the VM.
// It receives the vm instance (for potential access to env/stack) and evaluated arguments.
type InternalFunction func(vm *VM, args []any) (any, error)

// --- Reducer Registry ---

// ReducerKey identifies a pair of types for dispatching reducer functions.
// Using reflect.Type directly might be problematic with generics. Let's try string representation first.
type ReducerKey struct {
	LeftType  string // e.g., "*core.Outcomes[core.AccessResult]"
	RightType string
}

// VM holds the state for evaluating an AST.
type VM struct {
	stack              []any
	env                *Environment
	internalFuncs      map[string]InternalFunction
	maxOutcomeLen      int
	sequentialReducers map[ReducerKey]core.ReducerFunc[any, any, any] // Registry for AND
}

// NewVM creates a new vm instance.
func NewVM(maxBuckets int) *VM {
	if maxBuckets <= 0 {
		maxBuckets = 15 // Default if invalid
	}
	interp := &VM{
		stack:              make([]any, 0, 10), // Initial capacity
		env:                NewEnvironment(),   // Start with a global environment
		internalFuncs:      make(map[string]InternalFunction),
		sequentialReducers: make(map[ReducerKey]core.ReducerFunc[any, any, any]),
		maxOutcomeLen:      maxBuckets,
	}
	interp.registerDefaultReducers() // Register standard reducers
	return interp
}

// Env returns the current environment the vm is using.
func (v *VM) Env() *Environment {
	return v.env
}

// Helper to provide access to VM environment during setup
func (v *VM) SetEnv(e *Environment) {
	v.env = e
}

// Helper to access outer environment
func (e *Environment) Outer() *Environment {
	return e.outer
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

// --- Internal Function Registry ---

func (v *VM) RegisterInternalFunc(name string, fn InternalFunction) {
	v.internalFuncs[name] = fn
}

// --- Evaluation (Stub) ---

// Eval is the main entry point for evaluating an AST node.
// It uses a type switch to delegate to specific eval methods.
// The result of an evaluation is typically left on the vm's stack.
// Returns the final result (often the top of the stack after full eval) and any error.
func (v *VM) Eval(node Node) (any, error) {
	// fmt.Printf("Eval entry: %T - %s\n", node, node) // Debug entry
	var err error
	switch n := node.(type) {
	// We'll add cases here in subsequent phases
	case *LiteralExpr:
		err = v.evalLiteral(n)
	case *IdentifierExpr:
		err = v.evalIdentifier(n)
	case *InternalCallExpr:
		err = v.evalInternalCall(n)
	case *AndExpr:
		err = v.evalAndExpr(n)
	case *CallExpr:
		err = v.evalCallExpr(n)
	case *MemberAccessExpr:
		// Member access is often handled *within* evalCallExpr,
		// but we might need a stub if it can be evaluated alone.
		err = v.evalMemberAccessExpr(n) // <-- Call the actual implementation
	case *RepeatExpr:
		err = v.evalRepeatExpr(n)

	// --- Statement Nodes ---
	case *BlockStmt:
		// When Eval is called directly on a BlockStmt (e.g., top level),
		// there is no initial context from an outer structure like IfStmt.
		// evalBlockStmt now returns the result directly, doesn't leave on stack implicitly
		blockResult, evalErr := v.evalBlockStmt(n, v.env, nil) // Pass nil context
		// Only push if no error and result is non-nil (avoids pushing nil return signal value)
		if evalErr == nil && blockResult != nil {
			v.push(blockResult) // Push the final result of the block evaluation
		}
		err = evalErr // Assign any error from the block execution
	case *AssignmentStmt:
		err = v.evalAssignmentStmt(n)
	case *ReturnStmt:
		err = v.evalReturnStmt(n) // Return signals via special error/value
	case *ExprStmt:
		err = v.evalExprStmt(n)
	case *IfStmt: // <-- Will be implemented now
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

// --- Combination & Reduction Helper ---

// combineAndReduceImplicit takes two outcome objects, determines
// the correct sequential reducer from the registry, calls core.And (via the registered func),
// applies reduction if necessary, and returns the final combined outcome.
func (v *VM) combineOutcomesAndReduce(leftOutcome, rightOutcome any) (any, error) {
	var combinedOutcome any // Use any to hold the result
	var reductionNeeded bool = false
	// var inputLenForLog int = 0 // For logging reduction - maybe add later if needed

	// --- Use Reducer Registry ---
	key := ReducerKey{
		LeftType:  getOutcomeTypeString(leftOutcome),
		RightType: getOutcomeTypeString(rightOutcome),
	}

	// The registered functions now directly perform the core.And call
	registeredReducer, found := v.sequentialReducers[key]
	if !found {
		return nil, fmt.Errorf("%w: no sequential reducer registered for combination %T THEN %T (key: %+v)", ErrUnsupportedType, leftOutcome, rightOutcome, key)
	}

	// The registered function takes any, performs type assertions, calls core.And, returns any
	combinedOutcome = registeredReducer(leftOutcome, rightOutcome)

	// Determine if reduction is needed based on the combined result type and length
	reductionNeeded = v.needsReduction(combinedOutcome) // Use helper method

	// --- Apply Reduction if needed ---
	if reductionNeeded {
		trimmedOutcome := combinedOutcome // Start with the combined outcome
		switch co := combinedOutcome.(type) {
		case *core.Outcomes[core.AccessResult]:
			trimmerFuncGen := core.TrimToSize(v.maxOutcomeLen+50, v.maxOutcomeLen)
			successes, failures := co.Split(core.AccessResult.IsSuccess)
			trimmedSuccesses := trimmerFuncGen(successes)
			trimmedFailures := trimmerFuncGen(failures)
			finalTrimmed := (&core.Outcomes[core.AccessResult]{And: co.And}).Append(trimmedSuccesses, trimmedFailures)
			trimmedOutcome = finalTrimmed // Update the outcome
			// fmt.Printf("Applied TrimToSize, len %d -> %d\n", co.Len(), finalTrimmed.Len()) // Debug log
		// Add cases for other types that need trimming (e.g., RangedResult)
		default:
			// Type doesn't have a defined trimmer, or reduction wasn't deemed necessary earlier
		}
		combinedOutcome = trimmedOutcome // Use the (potentially) trimmed result
	}

	return combinedOutcome, nil // Return the final result
}

// needsReduction checks if an outcome object requires trimming based on its type and length.
func (v *VM) needsReduction(outcome any) bool {
	// Check length based on type implementing OutcomeContainer
	container, ok := outcome.(core.OutcomeContainer)
	if !ok {
		return false // Not a container we can check length on
	}

	// Check specific types that support trimming
	switch outcome.(type) {
	case *core.Outcomes[core.AccessResult]:
		return container.Len() > v.maxOutcomeLen
	case *core.Outcomes[core.RangedResult]:
		return container.Len() > v.maxOutcomeLen
	// Add other types that support reduction
	default:
		return false // By default, types don't need reduction
	}
}

// --- Reducer Registry ---

func getOutcomeTypeString(outcome any) string {
	// Use reflection to get a string representation like "*core.Outcomes[core.AccessResult]"
	// This is somewhat fragile but avoids complex generic type reflection for now.
	if outcome == nil {
		return "<nil>"
	}
	return reflect.TypeOf(outcome).String()
}

func (v *VM) RegisterSequentialReducer(leftExample, rightExample any, reducer core.ReducerFunc[any, any, any]) error {
	if reducer == nil {
		return fmt.Errorf("reducer function cannot be nil")
	}
	// Use type strings as keys for simplicity
	key := ReducerKey{
		LeftType:  getOutcomeTypeString(leftExample),
		RightType: getOutcomeTypeString(rightExample),
	}
	if key.LeftType == "<nil>" || key.RightType == "<nil>" {
		return fmt.Errorf("cannot register reducer with nil example types")
	}

	// fmt.Printf("DEBUG: Registering Reducer for Key: %+v\n", key) // Debug registration
	v.sequentialReducers[key] = reducer
	return nil
}

// registerDefaultReducers populates the registry with standard combinations.
func (v *VM) registerDefaultReducers() {
	// --- AccessResult Reducers ---
	accRes := &core.Outcomes[core.AccessResult]{} // Example instance
	dur := &core.Outcomes[core.Duration]{}        // Example instance
	boolean := &core.Outcomes[bool]{}             // Example instance

	// AccessResult + AccessResult
	// The registered function needs to handle the any conversion and call core.And
	_ = v.RegisterSequentialReducer(accRes, accRes, func(a, b any) any {
		// Type assertion happens here inside the registered func
		return core.And(a.(*core.Outcomes[core.AccessResult]), b.(*core.Outcomes[core.AccessResult]), core.AndAccessResults)
	})
	// AccessResult + Duration
	_ = v.RegisterSequentialReducer(accRes, dur, func(a, b any) any {
		reducer := func(vA core.AccessResult, vB core.Duration) core.AccessResult {
			return core.AccessResult{Success: vA.Success, Latency: vA.Latency + vB}
		}
		return core.And(a.(*core.Outcomes[core.AccessResult]), b.(*core.Outcomes[core.Duration]), reducer)
	})
	// Duration + AccessResult
	_ = v.RegisterSequentialReducer(dur, accRes, func(a, b any) any {
		reducer := func(vA core.Duration, vB core.AccessResult) core.AccessResult {
			return core.AccessResult{Success: vB.Success, Latency: vA + vB.Latency}
		}
		return core.And(a.(*core.Outcomes[core.Duration]), b.(*core.Outcomes[core.AccessResult]), reducer)
	})

	// --- Duration Reducers ---
	// Duration + Duration
	_ = v.RegisterSequentialReducer(dur, dur, func(a, b any) any {
		reducer := func(vA, vB core.Duration) core.Duration { return vA + vB }
		return core.And(a.(*core.Outcomes[core.Duration]), b.(*core.Outcomes[core.Duration]), reducer)
	})

	// --- Bool Reducers ---
	// Bool + AccessResult
	_ = v.RegisterSequentialReducer(boolean, accRes, func(a, b any) any {
		reducer := func(vA bool, vB core.AccessResult) core.AccessResult { return vB } // Bool acts as filter via weights
		return core.And(a.(*core.Outcomes[bool]), b.(*core.Outcomes[core.AccessResult]), reducer)
	})
	// Bool + Bool
	_ = v.RegisterSequentialReducer(boolean, boolean, func(a, b any) any {
		reducer := func(vA, vB bool) bool { return vA && vB } // Example: logical AND
		return core.And(a.(*core.Outcomes[bool]), b.(*core.Outcomes[bool]), reducer)
	})

	// Add more combinations as needed (e.g., involving RangedResult)
}
