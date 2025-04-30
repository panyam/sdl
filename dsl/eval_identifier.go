package dsl

import (
	"fmt"
)

// evalIdentifier looks up the identifier's name in the current environment
// and pushes the found value onto the stack.
func (i *Interpreter) evalIdentifier(expr *IdentifierExpr) error {
	val, ok := i.env.Get(expr.Name)
	if !ok {
		return fmt.Errorf("%w: '%s'", ErrNotFound, expr.Name)
	}
	i.push(val)
	return nil
}
