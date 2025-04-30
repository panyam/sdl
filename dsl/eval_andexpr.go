package dsl

import (
	"errors"
	"fmt"
	// Needed for more generic type handling later, maybe basic switch now
)

var (
	ErrTypeMismatch    = errors.New("type mismatch in operation")
	ErrUnsupportedType = errors.New("unsupported type for operation")
)

// evalAndExpr evaluates the Left and Right expressions, then combines their
// resulting outcomes sequentially using core.And and the appropriate reducer.
// It also applies outcome reduction (trimming) if the result exceeds the limit.
func (v *VM) evalAndExpr(expr *AndExpr) error {
	// 1. Evaluate Left
	_, err := v.Eval(expr.Left)
	if err != nil {
		return fmt.Errorf("error evaluating left side of AND: %w", err)
	}

	// 2. Evaluate Right
	_, err = v.Eval(expr.Right)
	if err != nil {
		// Pop the left result before returning error to keep stack clean
		_, _ = v.pop() // Ignore error on cleanup pop
		return fmt.Errorf("error evaluating right side of AND: %w", err)
	}

	// 3. Pop results (Right first, then Left)
	rightOutcome, rErr := v.pop()
	leftOutcome, lErr := v.pop()
	if rErr != nil || lErr != nil {
		// Should not happen if Eval succeeded
		return fmt.Errorf("stack underflow after evaluating AND operands (L:%v, R:%v)", lErr, rErr)
	}

	// 4. Combine using the helper
	combinedResult, err := v.combineOutcomesAndReduce(leftOutcome, rightOutcome)
	if err != nil {
		return err // Propagate combination error
	}

	// 5. Push final result
	v.push(combinedResult)
	return nil
}
