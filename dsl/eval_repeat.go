package dsl

import (
	"errors"
	"fmt"

	"github.com/panyam/leetcoach/sdl/core"
)

var (
	ErrNonDeterministicRepeatCount = errors.New("repeat count must be a deterministic integer outcome")
	ErrInvalidRepeatCount          = errors.New("repeat count must be non-negative")
	ErrUnsupportedRepeatMode       = errors.New("unsupported repeat mode")
)

// evalRepeatExpr handles repeating the evaluation of an expression sequentially.
func (v *VM) evalRepeatExpr(expr *RepeatExpr) error {
	// --- Validate Mode ---
	if expr.Mode != Sequential {
		// TODO: Implement Parallel mode later
		return fmt.Errorf("%w: %s", ErrUnsupportedRepeatMode, expr.Mode)
	}

	// --- Evaluate Count ---
	_, err := v.Eval(expr.Count)
	if err != nil {
		return fmt.Errorf("error evaluating repeat count: %w", err)
	}
	countOutcomeRaw, err := v.pop()
	if err != nil {
		return fmt.Errorf("stack error getting repeat count: %w", err)
	}

	// Ensure count is deterministic integer
	countOutcome, ok := countOutcomeRaw.(*core.Outcomes[int64]) // Assuming INT literal -> int64
	if !ok {
		return fmt.Errorf("%w: expected *core.Outcomes[int64], got %T", ErrNonDeterministicRepeatCount, countOutcomeRaw)
	}
	count, ok := countOutcome.GetValue() // Check if single bucket
	if !ok {
		return fmt.Errorf("%w: count expression yielded multiple outcomes", ErrNonDeterministicRepeatCount)
	}
	if count < 0 {
		return fmt.Errorf("%w: got %d", ErrInvalidRepeatCount, count)
	}
	if count == 0 {
		// If repeating zero times, the result is a zero-effect "identity" outcome.
		// What type should it be? Needs context. For now, push a default success/zero latency.
		// This might need refinement depending on how Repeat is used.
		identity := (&core.Outcomes[core.AccessResult]{}).Add(1.0, core.AccessResult{Success: true, Latency: 0})
		identity.And = core.AndAccessResults // Need to set reducer
		v.push(identity)
		return nil
	}

	// --- Execute Loop ---
	// 1. Evaluate the Input expression ONCE to determine the type and get the base outcome for the first iteration.
	//    This is crucial for establishing the "identity" element and the type for combination.
	_, err = v.Eval(expr.Input)
	if err != nil {
		return fmt.Errorf("error evaluating input expression for repeat: %w", err)
	}
	// The result of the first iteration is now on the stack.
	// If count == 1, we are done.

	// 2. Loop for remaining iterations (count - 1)
	for k := int64(1); k < count; k++ {
		// Get the accumulated result from the previous iteration
		accumulatedOutcome, err := v.pop()
		if err != nil {
			return fmt.Errorf("stack error retrieving accumulated result in repeat loop (iter %d): %w", k, err)
		}

		// Evaluate the input expression AGAIN for this iteration
		_, err = v.Eval(expr.Input)
		if err != nil {
			// Need to push back accumulatedOutcome before returning? Or is stack state invalid now? Assume invalid.
			return fmt.Errorf("error evaluating input expression in repeat loop (iter %d): %w", k, err)
		}
		currentIterationOutcome, err := v.pop()
		if err != nil {
			return fmt.Errorf("stack error retrieving current iteration result in repeat loop (iter %d): %w", k, err)
		}

		// Combine `accumulatedOutcome` THEN `currentIterationOutcome`
		combinedResult, err := v.combineOutcomesAndReduce(accumulatedOutcome, currentIterationOutcome)
		if err != nil {
			return fmt.Errorf("repeat combination failed (iter %d): %w", k, err)
		}

		// Push the result of this iteration back onto the stack for the next loop
		v.push(combinedResult)
	}

	// The final accumulated result is left on the stack after the loop finishes.
	return nil
}
