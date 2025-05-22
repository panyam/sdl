package parser

import (
	"fmt"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

type ChainedExpr struct {
	ExprBase
	Children  []Expr
	Operators []string

	// Expression after operators have been taken into account
	UnchainedExpr Expr
}

type PrecedenceInfo struct {
	Precedence int
	AssocType  int // -1 for left, 0 for non associative, 1 for right associative
}

func (c *ChainedExpr) String() string {
	// Basic, doesn't handle precedence for parentheses
	return fmt.Sprintf("(%s)", strings.Join(gfn.Map(c.Children, func(e Expr) string { return e.String() }), ", "))
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

// Unchain converts the ChainedExpr into a tree of BinaryExpr nodes,
// respecting operator precedence and associativity provided by the Precedencer.
// The result is stored in c.UnchainedExpr.
func (c *ChainedExpr) Unchain(preceder Precedencer) {
	if c == nil {
		return // Should not happen if called on a valid object instance
	}
	if len(c.Children) == 0 {
		c.UnchainedExpr = nil
		return
	}

	if len(c.Operators) == 0 {
		if len(c.Children) == 1 {
			c.UnchainedExpr = c.Children[0]
		} else {
			// Malformed: multiple children, no operators. Parser should catch this.
			c.UnchainedExpr = nil
		}
		return
	}

	// A valid chain must have one more child than operators.
	if len(c.Children) != len(c.Operators)+1 {
		c.UnchainedExpr = nil // Malformed chain
		return
	}

	// Use provided preceder or a default if nil.
	p := preceder
	if p == nil {
		p = &defaultPrecedencer{}
	}

	// Initialize indices for stepping through children and operators.
	childIdx := 0
	opIdx := 0

	// Start parsing with the lowest possible precedence (0).
	// The parseExpressionRecursive function will build the tree.
	c.UnchainedExpr = c.parseExpressionRecursive(p, &childIdx, &opIdx, 0)

	// Post-parsing check: If not all children or operators were consumed,
	// it might indicate an issue with the input ChainedExpr or parsing logic,
	// potentially meaning the chain contained multiple independent expressions
	// or trailing tokens not part of a single valid expression structure.
	// For now, we assume the ChainedExpr is intended to form one cohesive expression.
	// If c.UnchainedExpr is nil, an error occurred during the recursive parse.
	if c.UnchainedExpr != nil {
		if childIdx != len(c.Children) || opIdx != len(c.Operators) {
			// This condition suggests that the ChainedExpr might have been formed
			// with more tokens than constitute a single expression parsable by
			// precedence climbing from the start. This should ideally be handled
			// by the parser that creates ChainedExpr instances.
			// For robustness, one might consider this an error state and set c.UnchainedExpr = nil.
			// log.Printf("Warning: Unchain did not consume all children/operators. Child: %d/%d, Op: %d/%d",
			// 	childIdx, len(c.Children), opIdx, len(c.Operators))
		}
	}
}

// parseExpressionRecursive implements the core precedence climbing logic.
// It consumes operands and operators from the ChainedExpr's lists,
// starting from the current *childIdx and *opIdx.
// It only processes operators whose precedence is >= minPrecedence.
func (c *ChainedExpr) parseExpressionRecursive(p Precedencer, childIdx *int, opIdx *int, minPrecedence int) Expr {
	// Check if there's an initial operand available.
	if *childIdx >= len(c.Children) {
		return nil // Error: Expected an operand, but ran out of children.
	}

	lhs := c.Children[*childIdx]
	if lhs == nil {
		return nil // Error: Encountered a nil operand in the chain.
	}
	*childIdx++ // Consume the lhs operand.

	for {
		// Check if there are more operators to process.
		if *opIdx >= len(c.Operators) {
			break // No more operators, lhs is the result for this level.
		}

		currentOp := c.Operators[*opIdx]
		opPrec := p.PrecedenceFor(currentOp)
		opAssoc := p.AssociativityFor(currentOp)

		// If the current operator's precedence is less than minPrecedence,
		// we don't process it at this level; return the lhs accumulated so far.
		// For left-associative operators of the same precedence:
		//   If opPrec == minPrecedence, it should NOT be processed by the recursive call for RHS,
		//   but by this loop iteration. This is handled by nextMinPrecedence = opPrec + 1.
		// For right-associative operators of the same precedence:
		//   If opPrec == minPrecedence, it SHOULD be processed by the recursive call for RHS.
		//   This is handled by nextMinPrecedence = opPrec.
		if opPrec < minPrecedence {
			break
		}

		// Specific check for non-associative operators:
		// If current operator is non-associative, and its precedence is equal to minPrecedence,
		// it implies an attempt to chain non-associative operators of the same precedence level
		// (e.g., a < b < c). This is an error.
		// This loop structure means we've already processed an 'lhs' and are considering 'currentOp'.
		// If 'minPrecedence' came from a previous non-associative op of the same precedence,
		// 'nextMinPrecedence' would have been 'opPrec + 1', causing this 'opPrec < minPrecedence' to break.
		// The only way to hit 'opPrec == minPrecedence' for a non-associative operator in the loop
		// implies an issue or that the `minPrecedence` wasn't correctly bumped by a prior non-associative op.
		// However, the core logic is: when we encounter `op`, we parse `rhs` with a potentially higher `minPrecedence`.
		// If `op` is non-associative, `rhs` is parsed with `opPrec + 1`. If `rhs` itself starts with an operator
		// of `opPrec`, that inner operator will fail `innerOpPrec < (opPrec + 1)` and the inner `parseExpressionRecursive`
		// will return just its `lhs`. This correctly prevents chaining.

		*opIdx++ // Consume the operator.

		var rhs Expr
		var nextMinRecursivePrecedence int

		if opAssoc == AssocLeft {
			nextMinRecursivePrecedence = opPrec + 1
		} else if opAssoc == AssocRight {
			// For right-associative, recurse with the same precedence to allow chaining on the right.
			nextMinRecursivePrecedence = opPrec
		} else if opAssoc == AssocNone {
			// For non-associative, the RHS should not contain another operator of the same precedence.
			// So, the recursive call for RHS must look for strictly higher precedence.
			nextMinRecursivePrecedence = opPrec + 1
		} else {
			// Unknown associativity - treat as an error.
			c.UnchainedExpr = nil // Mark error on the ChainedExpr
			return nil
		}

		// Check if there's an operand for the RHS.
		if *childIdx >= len(c.Children) {
			c.UnchainedExpr = nil // Error: Operator exists, but no RHS operand.
			return nil
		}

		rhs = c.parseExpressionRecursive(p, childIdx, opIdx, nextMinRecursivePrecedence)
		if rhs == nil {
			// Error parsing RHS (e.g., encountered nil operand or malformed structure deeper).
			c.UnchainedExpr = nil
			return nil
		}

		// Construct the new BinaryExpr node.
		// The NodeInfo should span from the start of the original lhs of this operation
		// to the end of the parsed rhs.
		newExpr := &BinaryExpr{
			// NodeInfo will be set after lhs_for_new_expr is determined
			Left:     lhs, // This is the lhs accumulated so far at the current level
			Operator: currentOp,
			Right:    rhs,
		}
		if newExpr.Left != nil && newExpr.Right != nil {
			newExpr.NodeInfo = NodeInfo{StartPos: newExpr.Left.Pos(), StopPos: newExpr.Right.End()}
		}
		// else: NodeInfo remains zero if operands are problematic, though nil checks should prevent this.

		lhs = newExpr // The new binary expression becomes the lhs for the next iteration.
	}
	return lhs
}
