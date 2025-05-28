package decl

import (
	"fmt"
	"reflect"

	"github.com/panyam/sdl/core"
)

// --- Reducer Registry ---

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
	v.SequentialReducers[key] = reducer
	return nil
}

func getOutcomeTypeString(outcome any) string {
	// Use reflection to get a string representation like "*core.Outcomes[core.AccessResult]"
	// This is somewhat fragile but avoids complex generic type reflection for now.
	if outcome == nil {
		return "<nil>"
	}
	return reflect.TypeOf(outcome).String()
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
		return container.Len() > v.MaxOutcomeLen
	case *core.Outcomes[core.RangedResult]:
		return container.Len() > v.MaxOutcomeLen
	// Add other types that support reduction
	default:
		return false // By default, types don't need reduction
	}
}

// --- Combination & Reduction Helper ---

// combineAndReduceImplicit takes two outcome objects, determines
// the correct sequential reducer from the registry, calls core.And (via the registered func),
// applies reduction if necessary, and returns the final combined outcome.
// Think of this as the "garbage collector" where as outcomes across executions is growing the compactor
// will kick in without the user worrying about it
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
	registeredReducer, found := v.SequentialReducers[key]
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
			trimmerFuncGen := core.TrimToSize(v.MaxOutcomeLen+50, v.MaxOutcomeLen)
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
