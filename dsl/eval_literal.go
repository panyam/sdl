package dsl

import (
	"fmt"
	"strconv"

	"github.com/panyam/leetcoach/sdl/core"
)

// evalLiteral evaluates a LiteralExpr node.
// It parses the literal value based on its Kind, wraps it in a deterministic
// Outcomes[T] object (single bucket, weight 1.0), and pushes the result
// onto the interpreter's stack.
func (i *Interpreter) evalLiteral(expr *LiteralExpr) error {
	var outcome any
	var parseErr error

	switch expr.Kind {
	case "INT":
		val, err := strconv.ParseInt(expr.Value, 10, 64)
		if err != nil {
			parseErr = fmt.Errorf("invalid INT literal '%s': %w", expr.Value, err)
		} else {
			// Create a deterministic outcome: Outcomes[int64]
			// Note: Using int64 for integers
			o := (&core.Outcomes[int64]{}).Add(1.0, val)
			// Set And func? Not strictly needed for literals, but good practice
			o.And = func(a, b int64) int64 { return a + b } // Example reducer
			outcome = o
		}
	case "FLOAT":
		val, err := strconv.ParseFloat(expr.Value, 64)
		if err != nil {
			parseErr = fmt.Errorf("invalid FLOAT literal '%s': %w", expr.Value, err)
		} else {
			// Create a deterministic outcome: Outcomes[float64]
			o := (&core.Outcomes[float64]{}).Add(1.0, val)
			o.And = func(a, b float64) float64 { return a + b }
			outcome = o
		}
	case "STRING":
		// No parsing needed, value is already string
		// Create a deterministic outcome: Outcomes[string]
		o := (&core.Outcomes[string]{}).Add(1.0, expr.Value)
		o.And = func(a, b string) string { return a + b } // Example string reducer (concat)
		outcome = o
	case "BOOL":
		val, err := strconv.ParseBool(expr.Value)
		if err != nil {
			parseErr = fmt.Errorf("invalid BOOL literal '%s': %w", expr.Value, err)
		} else {
			// Create a deterministic outcome: Outcomes[bool]
			o := (&core.Outcomes[bool]{}).Add(1.0, val)
			o.And = func(a, b bool) bool { return a && b } // Example bool reducer (AND)
			outcome = o
		}
		// case "DURATION": // TODO: Add duration parsing later
		// return fmt.Errorf("duration literals not implemented yet")
	default:
		return fmt.Errorf("unknown literal kind: %s", expr.Kind)
	}

	if parseErr != nil {
		return parseErr
	}

	if outcome == nil {
		// Should not happen if parsing succeeded
		return fmt.Errorf("internal error: outcome is nil after parsing literal %s", expr.Value)
	}

	i.push(outcome)
	return nil
}
