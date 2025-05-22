package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Helpers ---

// MockExpr is a simple implementation of Expr for testing.
type MockExpr struct {
	ExprBase
	ID string // To identify the mock expression
}

func (me *MockExpr) String() string { return me.ID }

// Helper to create a MockExpr with NodeInfo
func newMockExpr(id string, start, end int) *MockExpr {
	return &MockExpr{
		ExprBase: ExprBase{NodeInfo: NodeInfo{StartPos: start, StopPos: end}},
		ID:       id,
	}
}

// Helper to create a ChainedExpr for testing
func newTestChainedExpr(children []Expr, operators []string) *ChainedExpr {
	// Calculate overall NodeInfo for the ChainedExpr
	ni := NodeInfo{}
	if len(children) > 0 {
		if children[0] != nil {
			ni.StartPos = children[0].Pos()
		}
		if children[len(children)-1] != nil {
			ni.StopPos = children[len(children)-1].End()
		}
	}

	return &ChainedExpr{
		ExprBase:  ExprBase{NodeInfo: ni},
		Children:  children,
		Operators: operators,
	}
}

// Helper to verify BinaryExpr structure and NodeInfo
func assertBinaryExpr(t *testing.T, expr Expr, expectedOp string, expectedLeft Expr, expectedRight Expr) *BinaryExpr {
	t.Helper()
	binExpr, ok := expr.(*BinaryExpr)
	require.True(t, ok, "Expected UnchainedExpr to be *BinaryExpr, got %T", expr)
	assert.Equal(t, expectedOp, binExpr.Operator, "Operator mismatch")
	assert.Same(t, expectedLeft, binExpr.Left, "Left operand mismatch")
	assert.Same(t, expectedRight, binExpr.Right, "Right operand mismatch")

	// Verify NodeInfo for the BinaryExpr
	if expectedLeft != nil && expectedRight != nil {
		assert.Equal(t, expectedLeft.Pos(), binExpr.Pos(), "BinaryExpr StartPos should match left operand's StartPos")
		assert.Equal(t, expectedRight.End(), binExpr.End(), "BinaryExpr StopPos should match right operand's EndPos")
	}
	return binExpr
}

// --- Test Cases ---

func TestChainedExpr_Unchain_NilReceiver(t *testing.T) {
	var c *ChainedExpr
	c.Unchain(nil) // Should not panic
	assert.Nil(t, c, "Receiver should still be nil")
}

func TestChainedExpr_Unchain_EmptyOrSingleChild(t *testing.T) {
	t.Run("EmptyChildren", func(t *testing.T) {
		c := newTestChainedExpr(nil, nil)
		c.Unchain(nil)
		assert.Nil(t, c.UnchainedExpr)

		c2 := newTestChainedExpr([]Expr{}, nil)
		c2.Unchain(nil)
		assert.Nil(t, c2.UnchainedExpr)
	})

	t.Run("SingleChildNoOperators", func(t *testing.T) {
		child := newMockExpr("a", 0, 1)
		c := newTestChainedExpr([]Expr{child}, nil)
		c.Unchain(nil)
		assert.Same(t, child, c.UnchainedExpr)
	})

	t.Run("SingleChildWithOperator_Malformed", func(t *testing.T) {
		child := newMockExpr("a", 0, 1)
		c := newTestChainedExpr([]Expr{child}, []string{"+"}) // Malformed
		c.Unchain(nil)
		assert.Nil(t, c.UnchainedExpr, "UnchainedExpr should be nil for malformed chain (1 child, 1 op)")
	})
}

func TestChainedExpr_Unchain_MalformedChains(t *testing.T) {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)

	t.Run("MultipleChildrenNoOperators", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{a, b}, nil) // Malformed
		c.Unchain(nil)
		assert.Nil(t, c.UnchainedExpr)
	})

	t.Run("ChildrenOperatorsMismatch_TooFewOps", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{a, b, newMockExpr("c", 4, 5)}, []string{"+"}) // Malformed
		c.Unchain(nil)
		assert.Nil(t, c.UnchainedExpr)
	})

	t.Run("ChildrenOperatorsMismatch_TooManyOps", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{a, b}, []string{"+", "-"}) // Malformed
		c.Unchain(nil)
		assert.Nil(t, c.UnchainedExpr)
	})
}

func TestChainedExpr_Unchain_NilOperands(t *testing.T) {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)

	t.Run("FirstChildNil_LeftAssoc", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{nil, b}, []string{"+"})
		c.Unchain(nil) // Defaults to left-associative
		assert.Nil(t, c.UnchainedExpr)
	})

	t.Run("MiddleChildNil_LeftAssoc", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{a, nil, b}, []string{"+", "-"})
		c.Unchain(nil)
		assert.Nil(t, c.UnchainedExpr)
	})

	t.Run("LastChildNil_RightAssoc", func(t *testing.T) {
		precedences := []PrecedenceInfo{{AssocType: 1}} // Right-associative
		c := newTestChainedExpr([]Expr{a, nil}, []string{"+"})
		c.Unchain(precedences)
		assert.Nil(t, c.UnchainedExpr)
	})

	t.Run("MiddleChildNil_RightAssoc", func(t *testing.T) {
		precedences := []PrecedenceInfo{{AssocType: 1}} // Right-associative
		c := newTestChainedExpr([]Expr{a, nil, b}, []string{"+", "-"})
		c.Unchain(precedences)
		assert.Nil(t, c.UnchainedExpr)
	})

	t.Run("ChildNil_NonAssoc", func(t *testing.T) {
		precedences := []PrecedenceInfo{{AssocType: 0}} // Non-associative
		c := newTestChainedExpr([]Expr{a, nil}, []string{"=="})
		c.Unchain(precedences)
		assert.Nil(t, c.UnchainedExpr)
	})
}

func TestChainedExpr_Unchain_LeftAssociative(t *testing.T) {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)
	c_expr := newMockExpr("c", 4, 5) // Renamed to avoid conflict with ChainedExpr variable 'c'

	precedences := []PrecedenceInfo{{AssocType: -1}} // Explicit Left-associative

	t.Run("a+b", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b}, []string{"+"})
		chain.Unchain(precedences) // Explicit left
		assertBinaryExpr(t, chain.UnchainedExpr, "+", a, b)

		chainDefault := newTestChainedExpr([]Expr{a, b}, []string{"+"})
		chainDefault.Unchain(nil) // Default left
		assertBinaryExpr(t, chainDefault.UnchainedExpr, "+", a, b)
	})

	t.Run("a+b-c", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"+", "-"})
		chain.Unchain(precedences)

		// Expected: ((a+b)-c)
		// Inner: (a+b)
		// Outer: Inner - c
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "((a + b) - c)", "UnchainedExpr should be ((a + b) - c)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "(a + b)", "Left operand should be (a + b)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "c", "Right operand should be c")

		// Verify NodeInfo for the outer expression (spans a to c_expr)
		assert.Equal(t, a.Pos(), chain.UnchainedExpr.Pos())
		assert.Equal(t, c_expr.End(), chain.UnchainedExpr.End())
	})
}

func TestChainedExpr_Unchain_RightAssociative(t *testing.T) {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)
	c_expr := newMockExpr("c", 4, 5)

	precedences := []PrecedenceInfo{{AssocType: 1}} // Right-associative

	t.Run("a=b", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b}, []string{"="})
		chain.Unchain(precedences)
		assertBinaryExpr(t, chain.UnchainedExpr, "=", a, b)
	})

	t.Run("a=b=c", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"=", "="})
		chain.Unchain(precedences)

		// Expected: (a=(b=c))
		// Inner: (b=c)
		// Outer: a = Inner
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "(a = (b = c))", "UnchainedExpr should be (a = (b = c))")
		outerAssign := chain.UnchainedExpr.(*BinaryExpr)
		innerAssign := outerAssign.Right.(*BinaryExpr)
		assert.Equal(t, outerAssign.Left.String(), "a", "Outer left operand should be a")
		assert.Equal(t, outerAssign.Right.String(), "(b = c)", "Outer right operand should be (b = c)")
		assert.Equal(t, innerAssign.Left.String(), "b", "Inner left operand should be b")
		assert.Equal(t, innerAssign.Right.String(), "c", "Inner right operand should be c")

		// Verify NodeInfo for the outer expression (spans a to c_expr)
		assert.Equal(t, a.Pos(), outerAssign.Pos())
		assert.Equal(t, c_expr.End(), outerAssign.End())
	})
}

func TestChainedExpr_Unchain_NonAssociative(t *testing.T) {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)
	c_expr := newMockExpr("c", 4, 5)

	precedences := []PrecedenceInfo{{AssocType: 0}} // Non-associative

	t.Run("a==b", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b}, []string{"=="})
		chain.Unchain(precedences)
		assertBinaryExpr(t, chain.UnchainedExpr, "==", a, b)
	})

	t.Run("a==b==c_Invalid", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"==", "=="})
		chain.Unchain(precedences)
		assert.Nil(t, chain.UnchainedExpr, "UnchainedExpr should be nil for chained non-associative operators")
	})

	t.Run("TwoChildrenOneOp_NonAssociative_Valid", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b}, []string{"<"}) // e.g., a < b
		chain.Unchain(precedences)                               // Should treat as non-associative
		assertBinaryExpr(t, chain.UnchainedExpr, "<", a, b)
	})
}

func TestChainedExpr_Unchain_UnsupportedAssociativity(t *testing.T) {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)
	precedences := []PrecedenceInfo{{AssocType: 99}} // Unknown type

	chain := newTestChainedExpr([]Expr{a, b}, []string{"+"})
	chain.Unchain(precedences)
	assert.Nil(t, chain.UnchainedExpr, "UnchainedExpr should be nil for unsupported associativity")
}

// Example to show a print of the unchained expression
func ExampleChainedExpr_Unchain() {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)
	c_expr := newMockExpr("c", 4, 5)
	d_expr := newMockExpr("d", 6, 7)

	// Left-associative: (a + b - c) * d
	// First chain: a + b - c
	chain1 := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"+", "-"})
	chain1.Unchain(nil) // Default left
	// chain1.UnchainedExpr is now ((a+b)-c)

	// Assume a parser would then create a new ChainedExpr for the '*'
	// For this example, we manually construct the next level:
	chain2 := newTestChainedExpr([]Expr{chain1.UnchainedExpr, d_expr}, []string{"*"})
	chain2.Unchain(nil)

	// To print the structure, we need a String() method on BinaryExpr that shows its children.
	// Let's assume BinaryExpr.String() prints like "(Left Op Right)"
	// and MockExpr.String() prints its ID.

	// Manually crafting the expected string output for ((a+b)-c)*d
	// ((a + b) - c) * d
	// BinaryExpr: Left=((a+b)-c), Op=*, Right=d
	//   BinaryExpr: Left=(a+b), Op=-, Right=c
	//     BinaryExpr: Left=a, Op=+, Right=b

	// We need a more detailed stringer for BinaryExpr to verify this easily.
	// For now, let's just print what we have for chain2.UnchainedExpr.
	// A real test would use assertions on the structure.

	fmt.Println(chain2.UnchainedExpr.String()) // Will print based on BinaryExpr's String method

	// Output: (((a + b) - c) * d)
}
