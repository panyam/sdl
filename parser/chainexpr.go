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

func (c *ChainedExpr) exprNode() {}
func (c *ChainedExpr) stmtNode() {}
func (c *ChainedExpr) String() string {
	// Basic, doesn't handle precedence for parentheses
	return fmt.Sprintf("(%s)", strings.Join(gfn.Map(c.Children, func(e Expr) string { return e.String() }), ", "))
}

// Note: This implementation assumes that all operators within a single ChainedExpr
// instance share the same precedence and associativity. This is typical if ChainedExpr
// is built by a parser rule that groups operators of the same precedence level.
// The `precedences` argument is assumed to provide the PrecedenceInfo for this
// shared level, or defaults are used if it's empty.
func (c *ChainedExpr) Unchain(precedences []PrecedenceInfo) {
	if c == nil {
		return
	}

	// Handle cases with no children or no operators first.
	if len(c.Children) == 0 {
		c.UnchainedExpr = nil
		return
	}

	if len(c.Operators) == 0 {
		if len(c.Children) == 1 {
			// If there's one child and no operators, the unchained expression is just that child.
			c.UnchainedExpr = c.Children[0]
		} else {
			// Malformed: e.g., multiple children but no operators to combine them.
			c.UnchainedExpr = nil
		}
		return
	}

	// Ensure the number of children and operators is consistent for a chain.
	// Expected: len(Children) == len(Operators) + 1
	if len(c.Children) != len(c.Operators)+1 {
		c.UnchainedExpr = nil // Malformed chain
		return
	}

	// Determine associativity. Default to left-associative if no precedence info is provided.
	assocType := -1 // Default: Left-associative
	if len(precedences) > 0 {
		// Assuming the first (or only) PrecedenceInfo applies to all operators in this chain.
		assocType = precedences[0].AssocType
	}

	var currentExpr Expr

	switch assocType {
	case -1: // Left-associative (e.g., a + b + c -> ((a+b)+c))
		currentExpr = c.Children[0]
		if currentExpr == nil { // First child should not be nil
			c.UnchainedExpr = nil
			return
		}

		for i := 0; i < len(c.Operators); i++ {
			rightOperand := c.Children[i+1]
			if rightOperand == nil { // Subsequent children should not be nil
				c.UnchainedExpr = nil
				return
			}
			opStr := c.Operators[i]

			newExprBase := ExprBase{}
			newExprBase.NodeInfo.StartPos = currentExpr.Pos()
			newExprBase.NodeInfo.StopPos = rightOperand.End()

			currentExpr = &BinaryExpr{
				ExprBase: newExprBase,
				Left:     currentExpr,
				Operator: opStr,
				Right:    rightOperand,
			}
		}
		c.UnchainedExpr = currentExpr

	case 1: // Right-associative (e.g., a = b = c -> (a=(b=c)) or a ^ b ^ c -> (a^(b^c)))
		currentExpr = c.Children[len(c.Children)-1]
		if currentExpr == nil { // Last child should not be nil
			c.UnchainedExpr = nil
			return
		}

		for i := len(c.Operators) - 1; i >= 0; i-- {
			leftOperand := c.Children[i]
			if leftOperand == nil { // Earlier children should not be nil
				c.UnchainedExpr = nil
				return
			}
			opStr := c.Operators[i]

			newExprBase := ExprBase{}
			newExprBase.NodeInfo.StartPos = leftOperand.Pos()
			newExprBase.NodeInfo.StopPos = currentExpr.End()

			currentExpr = &BinaryExpr{
				ExprBase: newExprBase,
				Left:     leftOperand,
				Operator: opStr,
				Right:    currentExpr,
			}
		}
		c.UnchainedExpr = currentExpr

	case 0: // Non-associative (e.g., a == b)
		if len(c.Operators) == 1 {
			// Must have exactly two children for one non-associative operator
			if len(c.Children) != 2 {
				c.UnchainedExpr = nil // Malformed
				return
			}
			leftOperand := c.Children[0]
			rightOperand := c.Children[1]
			if leftOperand == nil || rightOperand == nil {
				c.UnchainedExpr = nil
				return
			}
			opStr := c.Operators[0]

			newExprBase := ExprBase{}
			newExprBase.NodeInfo.StartPos = leftOperand.Pos()
			newExprBase.NodeInfo.StopPos = rightOperand.End()

			c.UnchainedExpr = &BinaryExpr{
				ExprBase: newExprBase,
				Left:     leftOperand,
				Operator: opStr,
				Right:    rightOperand,
			}
		} else {
			// Non-associative operators cannot be chained (e.g., a == b == c).
			// This implies a parsing error if ChainedExpr has multiple non-associative ops.
			c.UnchainedExpr = nil
		}

	default:
		// Unknown or unsupported associativity type.
		c.UnchainedExpr = nil
	}
}
