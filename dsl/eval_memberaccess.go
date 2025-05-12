package dsl

import (
	"errors"
	"fmt"

	"github.com/panyam/sdl/core"
)

var (
	ErrInvalidMemberAccess = errors.New("invalid member access")
	ErrUnsupportedMember   = errors.New("unsupported member accessed")
)

// evalMemberAccessExpr handles accessing members like 'variable.Member'.
// Crucially, for conditions like 'variable.Success', it evaluates the receiver ('variable')
// and pushes the *receiver's outcome object* itself onto the stack.
// The 'evalIfStmt' function then uses this object along with the knowledge
// that '.Success' was accessed to perform the correct split.
// Other member accesses are currently unsupported.
func (v *VM) evalMemberAccessExpr(expr *MemberAccessExpr) error {
	// 1. Evaluate the Receiver expression
	_, err := v.Eval(expr.Receiver)
	if err != nil {
		return fmt.Errorf("error evaluating receiver for member access '%s': %w", expr.Member, err)
	}
	receiverOutcomeRaw, err := v.pop() // Pop the receiver's outcome object
	if err != nil {
		return fmt.Errorf("stack error retrieving receiver for member access '%s': %w", expr.Member, err)
	}

	// 2. Handle Specific Member Accesses
	switch expr.Member {
	case "Success":
		// Check if the receiver's outcome type supports '.Success'
		// (AccessResult, RangedResult, potentially Bool outcomes?)
		isValidType := false
		switch receiverOutcomeRaw.(type) {
		case *core.Outcomes[core.AccessResult]:
			isValidType = true
		case *core.Outcomes[core.RangedResult]:
			isValidType = true
		case *core.Outcomes[bool]: // Allow accessing ".Success" on bool outcome? Maybe not. Let's restrict for now.
			// isValidType = true // Example if we wanted to allow it
			isValidType = false
		}

		if !isValidType {
			return fmt.Errorf("%w: cannot access '.Success' on type %T", ErrInvalidMemberAccess, receiverOutcomeRaw)
		}

		// For '.Success', push the *original receiver outcome* back onto the stack.
		// evalIfStmt will use this object and the knowledge that '.Success' was the member
		// to select the correct splitting predicate.
		v.push(receiverOutcomeRaw)
		return nil

		// TODO: Add cases for other potential member accesses if needed later
		// e.g., ".Latency", ".P99", ".Availability" ?
		// These would likely need to evaluate the outcome and push a *new* outcome
		// representing just that metric (e.g., Outcomes[float64] or Outcomes[Duration]).
		// case "MeanLatency":
		//    ... calculate mean, wrap in Outcomes[Duration], push ...
		//    return nil

	default:
		// Unsupported member
		// Pop the receiver outcome we evaluated, as it's not used
		// _, _ = v.pop() // Already popped above
		return fmt.Errorf("%w: '%s' on type %T", ErrUnsupportedMember, expr.Member, receiverOutcomeRaw)
	}
}
