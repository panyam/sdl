// sdl/dsl/eval_stmt.go
package dsl

import (
	"errors"
	"fmt"

	"github.com/panyam/sdl/core" // For identity outcome
)

// --- Return Value Handling ---
// We use a custom error type to signal a return statement was encountered.
// This avoids needing a separate return channel or modifying Eval's signature drastically.
var ErrReturnSignal = errors.New("return signal")

type ReturnValue struct {
	Value any // The actual return value (an Outcome object)
}

// Implement the error interface for ReturnValue
func (rv *ReturnValue) Error() string {
	return fmt.Sprintf("ReturnSignal(%v)", rv.Value)
}

// --- Statement Evaluators ---

// evalBlockStmt executes a sequence of statements within a given environment.
// It handles the implicit sequential 'And' composition.
// Returns the outcome of the *last* executed statement/expression in the block.
func (v *VM) evalBlockStmt(block *BlockStmt, env *Environment, initialContext any) (any, error) {
	// Use a new environment enclosed by the current one if needed for scoping?
	// For now, execute in the provided environment.
	// previousEnv := v.env
	// v.env = env // Or NewEnclosedEnvironment(env)? Let's use current for simplicity first.
	// defer func() { v.env = previousEnv }() // Restore env
	var blockResult = initialContext            // Start with the provided context
	isFirstStatement := (initialContext == nil) // Only true if no context was given

	// fmt.Printf("DEBUG evalBlockStmt: Start. Context: %T, isFirst: %t\n", initialContext, isFirstStatement) // Debug
	for _, stmt := range block.Statements {
		var currentStmtOutcome any
		evalErr := error(nil)

		// Evaluate the statement
		switch s := stmt.(type) {
		case *AssignmentStmt:
			evalErr = v.evalAssignmentStmt(s)
			if evalErr == nil {
				// Assignment result is stored in env, pop it from stack for composition chain
				currentStmtOutcome, evalErr = v.pop() // Value is left on stack by evalAssignmentStmt's Eval call
			}
		case *ExprStmt:
			// Evaluate the expression, result is left on stack
			_, evalErr = v.Eval(s.Expression)
			if evalErr == nil {
				currentStmtOutcome, evalErr = v.pop() // Pop the result
			}
		case *ReturnStmt:
			// Evaluate the return expression
			evalErr = v.evalReturnStmt(s) // evalReturnStmt pushes result and returns ReturnValue error wrapper
			// Error handling below will catch the ReturnValue signal
		// TODO: Add IfStmt, other statement types later
		default:
			return nil, fmt.Errorf("unsupported statement type in block: %T", s)
		}

		// --- Error Handling ---
		if evalErr != nil {
			// Check if it's our special return signal
			var retVal *ReturnValue
			if errors.As(evalErr, &retVal) {
				// It was a return statement. Propagate the return value up.
				return retVal.Value, ErrReturnSignal // Return the captured value and the signal
			}
			// Otherwise, it's a real error
			return nil, evalErr
		}

		// --- Implicit Sequential Composition ---
		// fmt.Printf("DEBUG evalBlockStmt: After Stmt %T. currentStmtOutcome: %T, blockResult: %T\n", stmt, currentStmtOutcome, blockResult) // Debug
		if currentStmtOutcome == nil {
			// Statement didn't produce an outcome to compose (e.g., just an assignment might conceptually do this, though ours pops)
			// Or it was the return statement handled above.
			// If blockResult is also nil, there's nothing to compose yet.
			// If blockResult is not nil, we just continue with the existing blockResult.
			// fmt.Printf("DEBUG evalBlockStmt: No outcome from stmt %T, continuing.\n", stmt) // Debug
			continue
		}

		if isFirstStatement {
			blockResult = currentStmtOutcome
			isFirstStatement = false
		} else {
			// Ensure blockResult is not nil before combining
			if blockResult == nil {
				// This means we had no initial context and the first statement didn't produce an outcome.
				// This shouldn't happen based on current logic, but handle defensively.
				return nil, fmt.Errorf("internal error: trying implicit AND with nil blockResult")
			}

			// Combine previous blockResult THEN currentStmtOutcome
			// Call helper to combine and reduce
			combinedResult, err := v.combineOutcomesAndReduce(blockResult, currentStmtOutcome)
			if err != nil {
				return nil, fmt.Errorf("implicit AND failed in block: %w", err)
			}
			// Update blockResult directly, don't push/pop here
			blockResult = combinedResult // Update block result directly
		}
	} // End loop over statements

	// If the block was empty or only contained assignments/non-outcome statements,
	// return a default "identity" outcome (Success=true, Latency=0)?
	if blockResult == nil {
		// Create default AccessResult outcome
		identity := (&core.Outcomes[core.AccessResult]{}).Add(1.0, core.AccessResult{Success: true, Latency: 0})
		identity.And = core.AndAccessResults
		blockResult = identity
		// It's debatable whether an empty block should have a result. Pushing identity for now.
		// This might need adjustment based on how function calls expect results.
	}

	// The final accumulated result of the block
	return blockResult, nil
}

// evalAssignmentStmt evaluates the RHS expression, stores the resulting
// Outcome object in the environment under the variable name, and leaves the
// result on the stack (for potential use in implicit AND).
func (v *VM) evalAssignmentStmt(stmt *AssignmentStmt) error {
	// Evaluate the RHS expression
	_, err := v.Eval(stmt.Value)
	if err != nil {
		return fmt.Errorf("error evaluating assignment value for '%s': %w", stmt.Variable.Name, err)
	}

	// Peek the result from the stack (don't pop yet, needed for implicit AND)
	valueOutcome, ok := v.peek()
	if !ok {
		// Should not happen if Eval succeeded without error
		return fmt.Errorf("stack empty after evaluating assignment value for '%s'", stmt.Variable.Name)
	}

	// Store the value (the Outcome object) in the environment
	v.env.Set(stmt.Variable.Name, valueOutcome)

	return nil
}

// evalReturnStmt evaluates the return expression, pushes the result onto the stack,
// and returns a special ReturnValue error wrapper containing the result.
func (v *VM) evalReturnStmt(stmt *ReturnStmt) error {
	// Evaluate the return value expression
	_, err := v.Eval(stmt.ReturnValue)
	if err != nil {
		return fmt.Errorf("error evaluating return value: %w", err)
	}

	// Get the result from the stack
	retValOutcome, ok := v.peek() // Peek, don't pop, the Eval caller will handle it
	if !ok {
		// Should not happen if Eval succeeded without error
		return fmt.Errorf("stack empty after evaluating return value")
	}

	// Wrap the result in our special error type to signal return
	return &ReturnValue{Value: retValOutcome} // Return the wrapper
}

// evalExprStmt evaluates an expression statement. The result is pushed onto the stack
// and becomes the input for the next step in implicit sequential composition.
func (v *VM) evalExprStmt(stmt *ExprStmt) error {
	// Evaluate the expression, result is left on stack
	_, err := v.Eval(stmt.Expression)
	if err != nil {
		return fmt.Errorf("error evaluating expression statement: %w", err)
	}
	// Result remains on the stack for evalBlockStmt to handle composition
	return nil
}

var (
	ErrInvalidConditionType = errors.New("if condition must evaluate to *core.Outcomes[bool] or access .Success field")
	ErrConditionSplitFailed = errors.New("failed to split outcome based on condition")
)

// selectSplitPredicate dynamically creates the predicate function for core.Split
// based on the Condition expression. Currently supports direct boolean outcomes
// or accessing the '.Success' field of an AccessResult outcome.
func selectSplitPredicate(conditionExpr Expr, conditionOutcome any) (func(v any) bool, error) {
	// Case 1: Condition is direct boolean outcome
	if _, ok := conditionOutcome.(*core.Outcomes[bool]); ok {
		return func(v any) bool {
			b, _ := v.(bool) // Assume type assertion works if we got here
			return b
		}, nil
	}

	// Case 2: Condition is accessing '.Success' (e.g., myVar.Success)
	if memExpr, ok := conditionExpr.(*MemberAccessExpr); ok && memExpr.Member == "Success" {
		// We need to check the *type* of the outcome associated with the receiver (myVar)
		// The `conditionOutcome` here IS that outcome.
		if _, ok := conditionOutcome.(*core.Outcomes[core.AccessResult]); ok {
			return func(v any) bool {
				ar, _ := v.(core.AccessResult) // Assume type assertion works
				return ar.Success
			}, nil
		}
		// Add checks for RangedResult .Success?
		// if _, ok := conditionOutcome.(*core.Outcomes[core.RangedResult]); ok { ... }
	}

	// If neither matches, the condition isn't supported yet
	return nil, fmt.Errorf("%w: expression %q (%T outcome) not supported", ErrInvalidConditionType, conditionExpr.String(), conditionOutcome)
}

// evalIfStmt handles conditional execution.
// It evaluates the condition, splits the relevant outcome distribution,
// evaluates the 'then' and 'else' branches with the split distributions,
// and appends the results.
func (v *VM) evalIfStmt(stmt *IfStmt) error {

	// 1. Evaluate the Condition Expression
	_, err := v.Eval(stmt.Condition)
	if err != nil {
		return fmt.Errorf("error evaluating if condition: %w", err)
	}
	conditionOutcomeRaw, err := v.pop() // Pop the result of the condition expression
	if err != nil {
		return fmt.Errorf("stack error getting if condition result: %w", err)
	}

	// 2. Determine Predicate and Split the Distribution
	// The condition itself might be complex (e.g., `a.Success`). We need the *outcome*
	// that the condition depends on (e.g., the outcome stored in `a`).
	// For now, assume the condition evaluation *directly* yields the outcome to be split.
	// Example: If `Condition` is `myVar.Success`, `Eval(Condition)` should push `myVar`'s outcome.
	// This requires `evalMemberAccess` to handle `.Success` specially or assumes
	// the DSL structure enforces `if someOutcome { ... }` where someOutcome is bool.
	// --- Let's refine: Assume Eval(stmt.Condition) pushes the relevant OUTCOME to split ---
	// This implies `evalMemberAccess` for `.Success` should push the receiver's outcome.
	// Now that evalMemberAccess is implemented, this assumption should hold.
	// For *this phase*, let's assume `conditionOutcomeRaw` IS the outcome distribution to split.

	splitPredicate, err := selectSplitPredicate(stmt.Condition, conditionOutcomeRaw)
	if err != nil {
		return err // Unsupported condition type
	}

	// Perform the split using type assertion (needs refinement for more types)
	var thenInputOutcome, elseInputOutcome any // Use any for split results

	switch condOutcome := conditionOutcomeRaw.(type) {
	case *core.Outcomes[core.AccessResult]:
		// Need to wrap the generic predicate
		specificPredicate := func(v core.AccessResult) bool {
			return splitPredicate(v) // Call the generic predicate
		}
		thenInputOutcome, elseInputOutcome = condOutcome.Split(specificPredicate)
	case *core.Outcomes[bool]:
		// Wrap the generic predicate
		specificPredicate := func(v bool) bool { // Expects bool
			return splitPredicate(v) // Call generic func(v any) bool
		}
		thenInputOutcome, elseInputOutcome = condOutcome.Split(specificPredicate)
	// Add cases for other splittable types (e.g., RangedResult if needed)
	default:
		return fmt.Errorf("%w: cannot split outcome of type %T based on condition %q", ErrConditionSplitFailed, conditionOutcomeRaw, stmt.Condition.String())
	}

	// 3. Evaluate Then Branch
	// Need to execute the Then block, providing `thenInputOutcome` as its starting context.
	// How to pass this context? The implicit 'And' in evalBlockStmt needs modification.
	// --- Simplification for now: ---
	// Assume the block implicitly starts with the split outcome. Push it onto the stack before calling evalBlockStmt.
	// The block's first operation will implicitly 'And' with it.
	var thenBranchResultOutcome any // This will hold the result *returned by* evalBlockStmt
	var thenErr error
	// Check if the split result is valid before proceeding
	// --- Use OutcomeContainer interface to check Len ---
	thenOutcomeContainer, thenIsContainer := thenInputOutcome.(core.OutcomeContainer)
	if !thenIsContainer && thenInputOutcome != nil {
		return fmt.Errorf("internal error: 'then' split result (%T) does not implement OutcomeContainer", thenInputOutcome)
	}
	if thenIsContainer && thenOutcomeContainer != nil && thenOutcomeContainer.Len() > 0 { // Check validity and length via interface
		// --- Pass context as argument, DON'T PUSH ---
		thenBranchResultOutcome, thenErr = v.evalBlockStmt(stmt.Then, v.env, thenInputOutcome) // Pass context
		if thenErr != nil && !errors.Is(thenErr, ErrReturnSignal) {                            // Handle errors (ignore Return signal here)
			return fmt.Errorf("error in 'then' branch: %w", thenErr)
		}
		if errors.Is(thenErr, ErrReturnSignal) {
			// If the 'then' branch returned, we need to propagate that
			// The actual return value is in thenBranchResultOutcome
			return &ReturnValue{Value: thenBranchResultOutcome} // Re-wrap
		}
		// Result is returned directly by evalBlockStmt, pop it if successful
		// No, evalBlockStmt doesn't leave result on stack, it returns it.
	} else {
		// If the 'then' path has zero probability, result is nil (or an empty outcome?)
		thenBranchResultOutcome = nil // Or create empty outcome of expected type?
	}

	// 4. Evaluate Else Branch (if it exists)
	var elseBranchResultOutcome any
	elseOutcomeContainer, elseIsContainer := elseInputOutcome.(core.OutcomeContainer)
	if !elseIsContainer && elseInputOutcome != nil {
		return fmt.Errorf("internal error: 'else' split result (%T) does not implement OutcomeContainer", elseInputOutcome)
	}
	var elseErr error
	if stmt.Else != nil && elseIsContainer && elseOutcomeContainer != nil && elseOutcomeContainer.Len() > 0 { // Check validity and length
		// --- Pass context as argument, DON'T PUSH ---
		elseBranchResultOutcome, elseErr = v.evalBlockStmt(stmt.Else, v.env, elseInputOutcome)
		// NOTE: evalBlockStmt should NOT leave its result on the stack. It returns it.
		if elseErr != nil && !errors.Is(elseErr, ErrReturnSignal) {
			return fmt.Errorf("error in 'else' branch: %w", elseErr)
		}
		if errors.Is(elseErr, ErrReturnSignal) {
			// Propagate return from 'else'
			return &ReturnValue{Value: elseBranchResultOutcome}
		}
		// Result is returned by evalBlockStmt
	} else {
		elseBranchResultOutcome = nil
	}

	// 5. Combine Results using Append
	var finalCombinedOutcome any

	// --- DEBUG ---
	fmt.Printf("DEBUG evalIfStmt Combine: ThenResult=%T, ElseResult=%T\n", thenBranchResultOutcome, elseBranchResultOutcome)

	// Use type assertions and core.Append
	switch tb := thenBranchResultOutcome.(type) {
	case *core.Outcomes[core.AccessResult]:
		eb, ok := elseBranchResultOutcome.(*core.Outcomes[core.AccessResult])
		if elseBranchResultOutcome == nil || (ok && eb != nil) { // Allow nil else or matching type
			// Need to handle case where tb itself might be nil if path had 0 prob
			if tb == nil {
				tb = &core.Outcomes[core.AccessResult]{And: core.AndAccessResults}
			} // Empty if no then path
			finalCombinedOutcome = tb.Append(eb) // Append handles nil input
		} else {
			return fmt.Errorf("%w: cannot combine then (%T) and else (%T)", ErrTypeMismatch, tb, elseBranchResultOutcome)
		}
	case *core.Outcomes[core.Duration]:
		eb, ok := elseBranchResultOutcome.(*core.Outcomes[core.Duration])
		if elseBranchResultOutcome == nil || (ok && eb != nil) {
			if tb == nil {
				tb = &core.Outcomes[core.Duration]{And: func(a, b core.Duration) core.Duration { return a + b }}
			}
			finalCombinedOutcome = tb.Append(eb)
		} else {
			return fmt.Errorf("%w: cannot combine then (%T) and else (%T)", ErrTypeMismatch, tb, elseBranchResultOutcome)
		}
	// Add cases for other types (bool, int, float, string...) if blocks can return them
	case nil: // Only else branch executed (or neither)
		// Simply use the else branch result (which might also be nil)
		finalCombinedOutcome = elseBranchResultOutcome // Handles nil/nil -> nil
	default:
		return fmt.Errorf("%w: cannot combine branches with type %T", ErrUnsupportedType, tb)
	}

	// --- Sanity Check: Ensure stack is empty before pushing final result ---
	if len(v.stack) != 0 {
		fmt.Printf("WARNING: Stack not empty before final push in evalIfStmt. Len=%d. Contents: %s\n", len(v.stack), v.stackString())
		// Optionally clear it? This might hide the real bug.
		// v.ClearStack()
	}

	// 6. Push the final combined result onto the stack
	// If both branches were nil/empty, finalCombinedOutcome might be nil. Push identity?
	if finalCombinedOutcome == nil {
		// Push a default identity outcome? This seems problematic.
		// If an if/else results in no possible outcomes, maybe that's okay?
		// Let's push nil for now, caller needs to handle.
		fmt.Println("Warning: IfStmt resulted in nil combined outcome.")
		v.push(nil) // Or push an empty outcome of a default type?
	} else {
		v.push(finalCombinedOutcome)
	}

	// TODO: Apply reduction to the final combined outcome? Yes.
	// Need to pop, check type, trim, push back.
	if finalCombinedOutcome != nil {
		finalPopped, _ := v.pop() // Pop the combined result we just pushed
		reductionNeeded := false
		// Check length based on type
		switch fc := finalPopped.(type) {
		case *core.Outcomes[core.AccessResult]:
			reductionNeeded = fc.Len() > v.maxOutcomeLen
			// Add other types if they need reduction
		}

		if reductionNeeded {
			switch fc := finalPopped.(type) {
			case *core.Outcomes[core.AccessResult]:
				trimmerFuncGen := core.TrimToSize(v.maxOutcomeLen+50, v.maxOutcomeLen)
				successes, failures := fc.Split(core.AccessResult.IsSuccess)
				trimmedSuccesses := trimmerFuncGen(successes)
				trimmedFailures := trimmerFuncGen(failures)
				finalTrimmed := (&core.Outcomes[core.AccessResult]{And: fc.And}).Append(trimmedSuccesses, trimmedFailures)
				v.push(finalTrimmed) // Push trimmed result
				// fmt.Printf("Applied TrimToSize after IF, len %d -> %d\n", finalLen, finalTrimmed.Len()) // Debug log
			// Add other trimmable types
			default:
				v.push(finalPopped) // Push original back if no trimmer
			}
		} else {
			v.push(finalPopped) // Push original back if no reduction needed
		}
	}

	return nil
}
