package dsl

import (
	"fmt"
)

// evalIdentifier looks up the identifier's name in the current environment
// and pushes the found value onto the stack.
func (v *VM) evalIdentifier(expr *IdentifierExpr) error {
	val, ok := v.env.Get(expr.Name)
	if !ok {
		return fmt.Errorf("%w: '%s'", ErrNotFound, expr.Name)
	}
	v.push(val)
	return nil
}
