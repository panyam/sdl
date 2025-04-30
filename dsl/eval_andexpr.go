package dsl

import (
	"errors"
	"fmt"
	"reflect" // Needed for more generic type handling later, maybe basic switch now

	"github.com/panyam/leetcoach/sdl/core"
)

var (
	ErrTypeMismatch    = errors.New("type mismatch in operation")
	ErrUnsupportedType = errors.New("unsupported type for operation")
)

// selectAndReducer selects the appropriate core.ReducerFunc based on the types
// of the two outcomes being combined sequentially.
// TODO: Extend this with more type combinations (e.g., AccessResult + Duration)
func selectAndReducer(left, right any) (core.ReducerFunc[any, any, any], error) {
	// We expect outcomes[T] interfaces{}
	leftOutcome, lok := left.(interface{ GetCoreType() reflect.Type })   // Placeholder interface
	rightOutcome, rok := right.(interface{ GetCoreType() reflect.Type }) // Placeholder interface

	if !lok || !rok {
		// Fallback to reflection if placeholder interface isn't implemented
		// This is less safe and assumes the any holds *core.Outcomes[T]
		leftType := reflect.TypeOf(left)
		rightType := reflect.TypeOf(right)

		if leftType.Kind() != reflect.Ptr || leftType.Elem().Name() != "Outcomes" ||
			rightType.Kind() != reflect.Ptr || rightType.Elem().Name() != "Outcomes" {
			return nil, fmt.Errorf("%w: AND expects two *core.Outcomes[T], got %T and %T", ErrTypeMismatch, left, right)
		}

		// Crude check for common case: AccessResult + AccessResult
		// A better approach would involve registering reducers per type combo
		// or having Outcomes expose its underlying type V.
		// This string check is fragile.
		if leftType.String() == "*core.Outcomes[core.AccessResult]" && rightType.String() == "*core.Outcomes[core.AccessResult]" {
			// Cast the generic reducer func to the expected any, any, any signature
			reducer := func(a, b any) any {
				return core.AndAccessResults(a.(core.AccessResult), b.(core.AccessResult))
			}
			return reducer, nil
		}
		// Handle AccessResult + Duration (maps duration to AccessResult)
		if leftType.String() == "*core.Outcomes[core.AccessResult]" && rightType.String() == "*core.Outcomes[float64]" {
			reducer := func(a, b any) any {
				return core.AndAccessResults(a.(core.AccessResult), core.AccessResult{Success: true, Latency: b.(core.Duration)})
			}
			return reducer, nil
		}
		if leftType.String() == "*core.Outcomes[float64]" && rightType.String() == "*core.Outcomes[core.AccessResult]" {
			reducer := func(a, b any) any {
				return core.AndAccessResults(core.AccessResult{Success: true, Latency: a.(core.Duration)}, b.(core.AccessResult))
			}
			return reducer, nil
		}

		// Add more combinations here...

		return nil, fmt.Errorf("%w: no AND reducer found for types %T and %T", ErrUnsupportedType, left, right)

	}

	// Example using hypothetical GetCoreType() method on Outcomes
	lt := leftOutcome.GetCoreType()
	rt := rightOutcome.GetCoreType()

	// Example placeholder logic (replace with actual type checking/registry)
	if lt == reflect.TypeOf(core.AccessResult{}) && rt == reflect.TypeOf(core.AccessResult{}) {
		reducer := func(a, b any) any {
			return core.AndAccessResults(a.(core.AccessResult), b.(core.AccessResult))
		}
		return reducer, nil
	}

	// Add more rules...

	return nil, fmt.Errorf("no sequential reducer found for types %v and %v", lt, rt)
}

// evalAndExpr evaluates the Left and Right expressions, then combines their
// resulting outcomes sequentially using core.And and the appropriate reducer.
// It also applies outcome reduction (trimming) if the result exceeds the limit.
func (i *Interpreter) evalAndExpr(expr *AndExpr) error {
	// 1. Evaluate Left
	_, err := i.Eval(expr.Left)
	if err != nil {
		return fmt.Errorf("error evaluating left side of AND: %w", err)
	}

	// 2. Evaluate Right
	_, err = i.Eval(expr.Right)
	if err != nil {
		// Pop the left result before returning error to keep stack clean
		_, _ = i.pop() // Ignore error on cleanup pop
		return fmt.Errorf("error evaluating right side of AND: %w", err)
	}

	// 3. Pop results (Right first, then Left)
	rightOutcome, rErr := i.pop()
	leftOutcome, lErr := i.pop()
	if rErr != nil || lErr != nil {
		// Should not happen if Eval succeeded
		return fmt.Errorf("stack underflow after evaluating AND operands (L:%v, R:%v)", lErr, rErr)
	}

	// 4. Select Reducer based on types
	// --- Using basic type switch for common cases ---
	var combinedOutcome any // Use any to hold the result
	var reductionNeeded bool = false

	switch lo := leftOutcome.(type) {
	case *core.Outcomes[core.AccessResult]:
		switch ro := rightOutcome.(type) {
		case *core.Outcomes[core.AccessResult]:
			reducer := core.AndAccessResults
			combined := core.And(lo, ro, reducer) // core.And returns *Outcomes[Z]
			combinedOutcome = combined            // Store the specific type
			reductionNeeded = combined.Len() > i.maxOutcomeLen
		case *core.Outcomes[core.Duration]: // Handle AccessResult + Duration
			reducer := func(a core.AccessResult, b core.Duration) core.AccessResult {
				// Combine: Add duration latency, success depends on original AccessResult
				return core.AccessResult{Success: a.Success, Latency: a.Latency + b}
			}
			combined := core.And(lo, ro, reducer)
			combinedOutcome = combined
			reductionNeeded = combined.Len() > i.maxOutcomeLen
		default:
			return fmt.Errorf("%w: cannot AND %T with %T", ErrTypeMismatch, lo, ro)
		}
	case *core.Outcomes[core.Duration]:
		switch ro := rightOutcome.(type) {
		case *core.Outcomes[core.AccessResult]: // Handle Duration + AccessResult
			reducer := func(a core.Duration, b core.AccessResult) core.AccessResult {
				return core.AccessResult{Success: b.Success, Latency: a + b.Latency}
			}
			combined := core.And(lo, ro, reducer)
			combinedOutcome = combined
			reductionNeeded = combined.Len() > i.maxOutcomeLen
		case *core.Outcomes[core.Duration]: // Handle Duration + Duration
			reducer := func(a core.Duration, b core.Duration) core.Duration { return a + b }
			combined := core.And(lo, ro, reducer)
			combinedOutcome = combined
			// Reduction for Duration outcomes? Maybe not needed or defined differently.
			reductionNeeded = false // Assume no trimming for duration for now
		default:
			return fmt.Errorf("%w: cannot AND %T with %T", ErrTypeMismatch, lo, ro)
		}
	// Add more cases for RangedResult, etc.
	default:
		return fmt.Errorf("%w: cannot AND %T with %T", ErrUnsupportedType, leftOutcome, rightOutcome)
	}

	// 5. Apply Reduction if needed
	if reductionNeeded {
		// --- Apply specific reduction based on the combined type ---
		switch co := combinedOutcome.(type) {
		case *core.Outcomes[core.AccessResult]:
			// Use the appropriate reduction strategy from core
			// TrimToSize returns a function, we need to call it
			trimmerFuncGen := core.TrimToSize(i.maxOutcomeLen+50, i.maxOutcomeLen)
			// The trimmerFuncGen expects a group (success/failure). We need to split.
			successes, failures := co.Split(core.AccessResult.IsSuccess)
			trimmedSuccesses := trimmerFuncGen(successes)
			trimmedFailures := trimmerFuncGen(failures) // Apply to failures too
			// Re-combine
			finalTrimmed := (&core.Outcomes[core.AccessResult]{And: co.And}).Append(trimmedSuccesses, trimmedFailures)
			combinedOutcome = finalTrimmed                                                 // Update the outcome to be pushed
			fmt.Printf("Applied TrimToSize, len %d -> %d\n", co.Len(), finalTrimmed.Len()) // Debug log
		// Add cases for other types that need trimming (e.g., RangedResult)
		// case *core.Outcomes[core.RangedResult]:
		// trimmerFuncGen := core.TrimToSizeRanged(i.maxOutcomeLen+50, i.maxOutcomeLen, 0.9) // Example threshold
		// ... split, trim, append ...
		default:
			// Type doesn't have a defined trimmer, or reduction wasn't deemed necessary earlier
			// Do nothing or log a warning
		}
	}

	// 6. Push final result
	i.push(combinedOutcome)
	return nil
}
