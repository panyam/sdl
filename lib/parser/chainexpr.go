package parser

import (
	"fmt"
	"strings"

	gfn "github.com/panyam/goutils/fn"
	"github.com/panyam/sdl/lib/decl"
)

type PrecedenceInfo struct {
	Precedence int
	AssocType  int // -1 for left, 0 for non associative, 1 for right associative
}

func (c *ChainedExpr) String() string {
	// Basic, doesn't handle precedence for parentheses
	return fmt.Sprintf("(%s)", strings.Join(gfn.Map(c.Children, func(e Expr) string { return e.String() }), ", "))
}

func (e *ChainedExpr) PrettyPrint(cp decl.CodePrinter) {
	panic("Should not be pretty printing this")
	// cp.Print(e.String())
}

// Associativity (assuming these are defined in the same package or accessible)
type Associativity int

const (
	AssocNone Associativity = iota
	AssocLeft
	AssocRight
)

type Precedencer interface {
	PrecedenceFor(operator string) int
	AssociativityFor(operator string) Associativity
}

// defaultPrecedencer provides default behavior if no Precedencer is given to Unchain.
type defaultPrecedencer struct{}

func (dp *defaultPrecedencer) PrecedenceFor(operator string) int {
	switch operator {
	case "*", "/":
		return 2 // Higher precedence
	case "+", "-":
		return 1 // Lower precedence
	default:
		// For other operators like comparisons, assignments, logical ops,
		// assign precedences as per your language design.
		// Defaulting to 0 if not specified.
		return 0
	}
}

func (dp *defaultPrecedencer) AssociativityFor(operator string) Associativity {
	// Define common associativities.
	// This is a simplified example; a real DSL would have more.
	switch operator {
	// Add other right-associative operators if any (e.g., power '^', some assignments)
	// case "=":
	//	 return AssocRight
	default:
		return AssocLeft // Most binary operators are left-associative
	}
}

type ChainedExpr struct {
	ExprBase
	Children  []Expr
	Operators []string

	// Expression after operators have been taken into account
	UnchainedExpr Expr
	Err           error // Stores parsing error
}

// Unchain converts the ChainedExpr into a tree of BinaryExpr nodes.
// It now returns an error if issues are encountered.
func (c *ChainedExpr) Unchain(preceder Precedencer) error {
	if c == nil {
		return fmt.Errorf("cannot unchain a nil ChainedExpr")
	}

	// Reset error and result for this unchain attempt
	c.Err = nil
	c.UnchainedExpr = nil

	if len(c.Children) == 0 {
		// No children, so no expression to form. Not necessarily an error for Unchain itself,
		// but UnchainedExpr remains nil. Could be an error if parser forms such a ChainedExpr.
		return nil
	}

	if len(c.Operators) == 0 {
		if len(c.Children) == 1 {
			c.UnchainedExpr = c.Children[0]
			if c.UnchainedExpr == nil { // If the single child itself was nil
				c.Err = fmt.Errorf("single child in ChainedExpr is nil at Line %d, Col: %d", c.Pos().Line, c.Pos().Col)
				return c.Err
			}
			return nil
		}
		// Malformed: multiple children, no operators.
		c.Err = fmt.Errorf("malformed chain: %d children, 0 operators, starting at pos %d", len(c.Children), c.Pos())
		return c.Err
	}

	if len(c.Children) != len(c.Operators)+1 {
		c.Err = fmt.Errorf("malformed chain: %d children, %d operators, starting at pos %d. Must have one more child than operators", len(c.Children), len(c.Operators), c.Pos())
		return c.Err
	}

	p := preceder
	if p == nil {
		p = &defaultPrecedencer{}
	}

	childIdx := 0
	opIdx := 0
	c.UnchainedExpr = c.parseExpressionRecursive(p, &childIdx, &opIdx, 0)

	if c.Err != nil { // An error was set during recursive parsing
		c.UnchainedExpr = nil // Ensure result is nil on error
		return c.Err
	}

	if c.UnchainedExpr == nil {
		// This case implies parseExpressionRecursive returned nil without setting c.Err.
		// This typically means an issue like a nil operand was encountered.
		// Attempt to create a more specific error if possible.
		if childIdx < len(c.Children) && c.Children[childIdx] == nil {
			c.Err = fmt.Errorf("encountered nil operand during unchaining at approximate child index %d (related to operator index %d) starting at pos %d", childIdx, opIdx, c.Pos())
		} else if opIdx < len(c.Operators) && childIdx >= len(c.Children) {
			// This means an operator was present but its RHS operand was missing.
			errPos := c.OperatorsStartPos(opIdx) // Helper to get approximate operator position
			c.Err = fmt.Errorf("operator '%s' (at operator index %d, approx char pos %d) is missing its right-hand operand in chain starting at pos %d", c.Operators[opIdx], opIdx, errPos, c.Pos())
		} else {
			c.Err = fmt.Errorf("failed to unchain expression starting at pos %d, result is nil without a more specific error (possible structural issue or unhandled nil operand)", c.Pos())
		}
		return c.Err
	}

	// After successful parsing of an initial expression, check if the entire chain was consumed.
	if childIdx != len(c.Children) || opIdx != len(c.Operators) {
		// This means there are leftover tokens that weren't part of the main expression resolved.
		// This is often the case for "a == b == c" where "(a == b)" is formed,
		// and "== c" is leftover. The non-associative check should catch this specifically.
		// If c.Err is already set (e.g., by non-associative check), don't override it.
		if c.Err == nil {
			operatorStr := "<end_of_chain>"
			approxPos := c.End()
			if opIdx < len(c.Operators) {
				operatorStr = c.Operators[opIdx]
				approxPos = c.OperatorsStartPos(opIdx)
			} else if childIdx < len(c.Children) && c.Children[childIdx] != nil {
				approxPos = c.Children[childIdx].Pos()
			}
			c.Err = fmt.Errorf("expression not fully parsed from chain starting at pos %d; operator '%s' (at index %d, approx char pos %d) and subsequent terms were not incorporated into the main expression tree. This may indicate incorrect chaining of operators with different precedences or disallowed non-associative chaining", c.Pos(), operatorStr, opIdx, approxPos)
			c.UnchainedExpr = nil // Invalidate partial result
			return c.Err
		}
	}

	return nil
}

// parseExpressionRecursive (with error detection for non-associative)
func (c *ChainedExpr) parseExpressionRecursive(p Precedencer, childIdx *int, opIdx *int, minPrecedence int) Expr {
	if c.Err != nil { // If an error has already been set, bail out.
		return nil
	}

	if *childIdx >= len(c.Children) {
		c.Err = fmt.Errorf("expected an operand for expression starting at pos %d, but ran out of children (child index %d, operator index %d)", c.Pos(), *childIdx, *opIdx)
		return nil
	}

	lhs := c.Children[*childIdx]
	if lhs == nil {
		c.Err = fmt.Errorf("encountered nil operand at child index %d in chain starting at pos %d (related to operator index %d)", *childIdx, c.Pos(), *opIdx)
		return nil
	}
	*childIdx++

	for {
		if *opIdx >= len(c.Operators) {
			break
		}

		currentOp := c.Operators[*opIdx]
		opPrec := p.PrecedenceFor(currentOp)
		opAssoc := p.AssociativityFor(currentOp)

		if opAssoc == Associativity(99) { // Example of how you might check for an "unknown" enum if your Precedencer can return it
			c.Err = fmt.Errorf("operator '%s' (at operator index %d, approx char pos %d) has unknown associativity in chain starting at pos %d", currentOp, *opIdx, c.OperatorsStartPos(*opIdx), c.Pos())
			return nil
		}

		if opPrec < minPrecedence {
			break
		}

		// --- Start: Error check for non-associative operator chaining ---
		// This check is slightly different: We look *ahead* from the current lhs.
		// If lhs was formed by a non-associative op (not easily known here without more state)
		// and currentOp has same precedence, it's an error.
		// The current precedence climbing handles this implicitly:
		// If lhs's op was non-assoc, its RHS was parsed with opPrec+1.
		// If currentOp here has opPrec, it would have failed the opPrec < minPrecedence check earlier.

		// The check is: if *this* currentOp is non-associative, and after forming its BinaryExpr,
		// the *next* operator in the original chain has the same precedence.
		// This check needs to be done *after* forming the current BinaryExpr with currentOp.

		// --- End: Error check for non-associative operator chaining (moved) ---

		opIdxConsumed := *opIdx // Store opIdx before recursive call for currentOp
		*opIdx++

		var rhs Expr
		var nextMinRecursivePrecedence int

		if opAssoc == AssocLeft {
			nextMinRecursivePrecedence = opPrec + 1
		} else if opAssoc == AssocRight {
			nextMinRecursivePrecedence = opPrec
		} else if opAssoc == AssocNone {
			nextMinRecursivePrecedence = opPrec + 1
		} else { // Should be caught by Associativity(99) check earlier or be a defined default
			c.Err = fmt.Errorf("operator '%s' (at operator index %d, approx char pos %d) has unsupported associativity type %v in chain starting at pos %d", currentOp, opIdxConsumed, c.OperatorsStartPos(opIdxConsumed), opAssoc, c.Pos())
			return nil
		}

		if *childIdx >= len(c.Children) {
			c.Err = fmt.Errorf("operator '%s' (at operator index %d, approx char pos %d) is missing its right-hand operand in chain starting at pos %d", currentOp, opIdxConsumed, c.OperatorsStartPos(opIdxConsumed), c.Pos())
			return nil
		}

		rhs = c.parseExpressionRecursive(p, childIdx, opIdx, nextMinRecursivePrecedence)
		if rhs == nil {
			if c.Err == nil { // If recursive call returned nil without setting c.Err
				c.Err = fmt.Errorf("failed to parse right-hand side for operator '%s' (at operator index %d, approx char pos %d) in chain starting at pos %d", currentOp, opIdxConsumed, c.OperatorsStartPos(opIdxConsumed), c.Pos())
			}
			return nil // Error parsing RHS, or c.Err was already set.
		}

		newExpr := &BinaryExpr{
			Left:     lhs,
			Operator: currentOp,
			Right:    rhs,
		}
		if newExpr.Left != nil && newExpr.Right != nil {
			newExpr.ExprBase.NodeInfo = NodeInfo{StartPos: newExpr.Left.Pos(), StopPos: newExpr.Right.End()}
		} else {
			// Should have been caught by nil checks for lhs or rhs returning nil
			c.Err = fmt.Errorf("internal error: nil operand for BinaryExpr with operator '%s' (at operator index %d, approx char pos %d) in chain starting at pos %d", currentOp, opIdxConsumed, c.OperatorsStartPos(opIdxConsumed), c.Pos())
			return nil
		}

		// --- Moved Error check for non-associative operator chaining ---
		if opAssoc == AssocNone {
			// After forming 'newExpr' with 'currentOp' (which was non-associative),
			// check if there's another operator immediately following in the original chain
			// that has the same precedence. *opIdx now points to this next operator.
			if *opIdx < len(c.Operators) {
				nextToken := c.Operators[*opIdx]
				nextTokenPrec := p.PrecedenceFor(nextToken)
				if nextTokenPrec == opPrec {
					// errPos := newExpr.End() + 1 // Position after the expression just formed
					opApproxPos := c.OperatorsStartPos(*opIdx)
					// if *childIdx < len(c.Children) && c.Children[*childIdx] != nil {
					// errPos = c.Children[*childIdx].Pos() // Position of the operand for the problematic operator
					// }

					c.Err = fmt.Errorf("invalid chaining: non-associative operator '%s' (forming expression ending at char %d) cannot be directly followed by operator '%s' (at operator index %d, approx char pos %d) of the same precedence %d in chain starting at pos %d",
						currentOp, newExpr.End(),
						nextToken, *opIdx, opApproxPos,
						opPrec, c.Pos())
					return nil
				}
			}
		}
		// --- End Moved Error Check ---
		lhs = newExpr
	}
	return lhs
}

// OperatorsStartPos is a helper to estimate operator start positions.
// This is an approximation as ChainedExpr only stores operator strings.
// A more accurate system would store token info for operators.
func (c *ChainedExpr) OperatorsStartPos(opIndex int) Location {
	if opIndex < 0 || opIndex >= len(c.Operators) {
		return c.Pos() // Fallback
	}
	// Operator is between Children[opIndex] and Children[opIndex+1]
	if opIndex < len(c.Children) && c.Children[opIndex] != nil {
		// Approx position is after the end of the left child of this operator
		e := c.Children[opIndex].End()
		e.Pos += 1
		return e
	}
	return c.Pos() // Fallback
}
