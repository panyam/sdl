package parser

import (
	"testing"

	"github.com/panyam/sdl/decl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Helpers ---

// MockExpr is a simple implementation of Expr for testing.
type MockExpr struct {
	ExprBase
	ID string // To identify the mock expression
}

func (me *MockExpr) String() string                  { return me.ID }
func (me *MockExpr) PrettyPrint(cp decl.CodePrinter) {}

// Helper to create a MockExpr with NodeInfo
func newMockExpr(id string, start, end int) *MockExpr {
	return &MockExpr{
		ExprBase: ExprBase{NodeInfo: NodeInfo{StartPos: Location{Pos: start}, StopPos: Location{Pos: end}}},
		ID:       id,
	}
}

// --- MockPrecedencer for Testing ---
type MockPrecedencer struct {
	Precedences     map[string]int
	Associativities map[string]Associativity
	DefaultPrec     int
	DefaultAssoc    Associativity
}

func NewMockPrecedencer() *MockPrecedencer {
	return &MockPrecedencer{
		Precedences:     make(map[string]int),
		Associativities: make(map[string]Associativity),
		DefaultPrec:     0,
		DefaultAssoc:    AssocLeft, // Default to left for unspecified ops
	}
}

func (mp *MockPrecedencer) PrecedenceFor(operator string) int {
	if prec, ok := mp.Precedences[operator]; ok {
		return prec
	}
	return mp.DefaultPrec
}

func (mp *MockPrecedencer) AssociativityFor(operator string) Associativity {
	if assoc, ok := mp.Associativities[operator]; ok {
		return assoc
	}
	return mp.DefaultAssoc
}

// Helper to create a ChainedExpr for testing
func newTestChainedExpr(children []Expr, operators []string) *ChainedExpr {
	ni := NodeInfo{}
	if len(children) > 0 {
		if children[0] != nil {
			ni.StartPos = children[0].Pos()
		}
		// Find the last non-nil child for StopPos
		lastValidChildEnd := ni.StartPos // Default to start if all are nil or only one child
		for i := len(children) - 1; i >= 0; i-- {
			if children[i] != nil {
				lastValidChildEnd = children[i].End()
				break
			}
		}
		ni.StopPos = lastValidChildEnd
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
	c.Unchain(nil)
	assert.Nil(t, c)
}

func TestChainedExpr_Unchain_EmptyOrSingleChild(t *testing.T) {
	preceder := NewMockPrecedencer()

	t.Run("EmptyChildren", func(t *testing.T) {
		c := newTestChainedExpr(nil, nil)
		c.Unchain(preceder)
		assert.Nil(t, c.UnchainedExpr)

		c2 := newTestChainedExpr([]Expr{}, nil)
		c2.Unchain(preceder)
		assert.Nil(t, c2.UnchainedExpr)
	})

	t.Run("SingleChildNoOperators", func(t *testing.T) {
		child := newMockExpr("a", 0, 1)
		c := newTestChainedExpr([]Expr{child}, nil)
		c.Unchain(preceder)
		assert.Same(t, child, c.UnchainedExpr)
	})

	t.Run("SingleChildWithOperator_Malformed", func(t *testing.T) {
		child := newMockExpr("a", 0, 1)
		c := newTestChainedExpr([]Expr{child}, []string{"+"}) // Malformed
		c.Unchain(preceder)
		assert.Nil(t, c.UnchainedExpr, "UnchainedExpr should be nil for malformed chain (1 child, 1 op)")
	})
}

func TestChainedExpr_Unchain_MalformedChains(t *testing.T) {
	preceder := NewMockPrecedencer()
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)

	t.Run("MultipleChildrenNoOperators", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{a, b}, nil) // Malformed
		c.Unchain(preceder)
		assert.Nil(t, c.UnchainedExpr)
	})

	t.Run("ChildrenOperatorsMismatch_TooFewOps", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{a, b, newMockExpr("c", 4, 5)}, []string{"+"}) // Malformed
		c.Unchain(preceder)
		assert.Nil(t, c.UnchainedExpr)
	})

	t.Run("ChildrenOperatorsMismatch_TooManyOps", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{a, b}, []string{"+", "-"}) // Malformed
		c.Unchain(preceder)
		assert.Nil(t, c.UnchainedExpr)
	})
}

func TestChainedExpr_Unchain_NilOperands(t *testing.T) {
	preceder := NewMockPrecedencer() // Using default (left-assoc, prec 0)
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)

	t.Run("FirstChildNil", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{nil, b}, []string{"+"})
		c.Unchain(preceder)
		assert.Nil(t, c.UnchainedExpr, "Unchain should result in nil if first operand is nil")
	})

	t.Run("MiddleChildNil", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{a, nil, b}, []string{"+", "-"})
		c.Unchain(preceder)
		assert.Nil(t, c.UnchainedExpr, "Unchain should result in nil if a middle operand is nil")
	})

	t.Run("LastChildNil", func(t *testing.T) {
		c := newTestChainedExpr([]Expr{a, nil}, []string{"+"})
		c.Unchain(preceder)
		assert.Nil(t, c.UnchainedExpr, "Unchain should result in nil if last operand is nil for binary op")
	})
}

func TestChainedExpr_Unchain_Associativity(t *testing.T) {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)
	c_expr := newMockExpr("c", 4, 5)

	t.Run("LeftAssociative_DefaultPreceder", func(t *testing.T) {
		// defaultPreceder in Unchain is left-associative, precedence based on common ops
		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"+", "-"})
		chain.Unchain(nil) // Use default preceder

		// Expected: ((a+b)-c)
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "((a + b) - c)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "(a + b)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "c")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Operator, "-")
	})

	t.Run("LeftAssociative_Explicit", func(t *testing.T) {
		preceder := NewMockPrecedencer()
		preceder.Precedences["+"] = 1
		preceder.Associativities["+"] = AssocLeft
		preceder.Precedences["-"] = 1
		preceder.Associativities["-"] = AssocLeft

		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"+", "-"})
		chain.Unchain(preceder)
		// Expected: ((a+b)-c)
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "((a + b) - c)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "(a + b)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "c")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Operator, "-")
	})

	t.Run("RightAssociative_Explicit", func(t *testing.T) {
		preceder := NewMockPrecedencer()
		preceder.Precedences["="] = 1
		preceder.Associativities["="] = AssocRight

		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"=", "="})
		chain.Unchain(preceder)
		// Expected: (a=(b=c))
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "(a = (b = c))")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "a")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "(b = c)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Operator, "=")
	})

	t.Run("NonAssociative_SingleOp", func(t *testing.T) {
		preceder := NewMockPrecedencer()
		preceder.Precedences["=="] = 1
		preceder.Associativities["=="] = AssocNone

		chain := newTestChainedExpr([]Expr{a, b}, []string{"=="})
		chain.Unchain(preceder)
		// Expected: (a==b)
		assertBinaryExpr(t, chain.UnchainedExpr, "==", a, b)
	})

	t.Run("NonAssociative_Chained_Error", func(t *testing.T) {
		// Although Unchain might produce a result for a == b == c with non-assoc by parsing (a == b)
		// and then trying to use that result with '== c', the Precedencer combined with
		// parseExpressionRecursive's nextMinRecursivePrecedence logic for AssocNone (opPrec + 1)
		// should prevent chaining of same-precedence non-associative operators.
		// The result will be the first valid binary expr, and the rest of the chain unconsumed by this specific Unchain call.
		// A higher-level parser would typically signal this as a syntax error.
		// The current Unchain will parse (a == b) and leave "== c" unconsumed by the initial call depth.
		// If the expectation is for Unchain to error out *itself*, that logic needs to be more explicit.
		// Current Unchain will produce (a == b) and childIdx/opIdx will not be at the end.
		// Let's test the actual behavior: it should parse the first valid part.
		preceder := NewMockPrecedencer()
		preceder.Precedences["=="] = 1
		preceder.Associativities["=="] = AssocNone

		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"==", "=="})
		err := chain.Unchain(preceder)

		require.Error(t, err, "Expected an error for chaining non-associative operators")
		assert.Nil(t, chain.UnchainedExpr, "UnchainedExpr should be nil on error")
		// fmt.Println(err.Error()) // For debugging the error message
		assert.Contains(t, err.Error(), "invalid chaining: non-associative operator '=='")
		assert.Contains(t, err.Error(), "cannot be directly followed by operator '==' (at operator index 1")
		assert.Contains(t, err.Error(), "of the same precedence 1")
	})
}

func TestChainedExpr_Unchain_ComplexPrecedence(t *testing.T) {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)
	c_expr := newMockExpr("c", 4, 5)
	d_expr := newMockExpr("d", 6, 7)
	e_expr := newMockExpr("e", 8, 9)

	preceder := NewMockPrecedencer()
	// Standard mathematical precedences
	preceder.Precedences["*"] = 3
	preceder.Associativities["*"] = AssocLeft
	preceder.Precedences["/"] = 3
	preceder.Associativities["/"] = AssocLeft
	preceder.Precedences["+"] = 2
	preceder.Associativities["+"] = AssocLeft
	preceder.Precedences["-"] = 2
	preceder.Associativities["-"] = AssocLeft
	preceder.Precedences[">"] = 1
	preceder.Associativities[">"] = AssocNone
	preceder.Precedences["=="] = 1
	preceder.Associativities["=="] = AssocNone

	t.Run("a + b * c", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"+", "*"})
		chain.Unchain(preceder)
		// Expected: (a + (b*c))
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "(a + (b * c))")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "a")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "(b * c)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Operator, "+")
	})

	t.Run("a * b + c", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"*", "+"})
		chain.Unchain(preceder)
		// Expected: ((a*b) + c)
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "((a * b) + c)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "(a * b)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "c")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Operator, "+")
	})

	t.Run("a + b * c - d / e", func(t *testing.T) {
		// Chain: a + b * c - d / e
		// Ops:      +   *   -   /
		// Children: a   b   c   d   e
		chain := newTestChainedExpr([]Expr{a, b, c_expr, d_expr, e_expr}, []string{"+", "*", "-", "/"})
		chain.Unchain(preceder)

		// Expected: ((a + (b*c)) - (d/e))
		// Root will be "-"
		// Left of "-": (a + (b*c))
		// Right of "-": (d/e)
		require.NotNil(t, chain.UnchainedExpr)

		assert.Equal(t, chain.UnchainedExpr.String(), "((a + (b * c)) - (d / e))")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "(a + (b * c))")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "(d / e)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Operator, "-")
	})

	t.Run("a > b + c", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{">", "+"})
		chain.Unchain(preceder)
		// Expected: (a > (b+c)) because + is higher precedence than >
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "(a > (b + c))")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "a")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "(b + c)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Operator, ">")
	})

	t.Run("a * b == c - d", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b, c_expr, d_expr}, []string{"*", "==", "-"})
		chain.Unchain(preceder)
		// Expected: ((a*b) == (c-d))
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "((a * b) == (c - d))")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "(a * b)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "(c - d)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Operator, "==")
	})
}

func TestChainedExpr_Unchain_DefaultPrecederSimple(t *testing.T) {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)
	c_expr := newMockExpr("c", 4, 5)

	// defaultPreceder treats +,- as prec 1, *,/ as prec 2. All left-associative.
	t.Run("a + b * c (default)", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"+", "*"})
		chain.Unchain(nil) // Use default preceder
		// Expected: (a + (b*c))
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "(a + (b * c))")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "a")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "(b * c)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Operator, "+")
	})

	t.Run("a * b + c (default)", func(t *testing.T) {
		chain := newTestChainedExpr([]Expr{a, b, c_expr}, []string{"*", "+"})
		chain.Unchain(nil) // Use default preceder
		// Expected: ((a*b) + c)
		require.NotNil(t, chain.UnchainedExpr)
		assert.Equal(t, chain.UnchainedExpr.String(), "((a * b) + c)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Left.String(), "(a * b)")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Right.String(), "c")
		assert.Equal(t, chain.UnchainedExpr.(*BinaryExpr).Operator, "+")
	})
}

func TestChainedExpr_Unchain_UnknownAssociativity(t *testing.T) {
	a := newMockExpr("a", 0, 1)
	b := newMockExpr("b", 2, 3)
	preceder := NewMockPrecedencer()
	preceder.Precedences["$"] = 1
	preceder.Associativities["$"] = Associativity(99) // Unknown

	chain := newTestChainedExpr([]Expr{a, b}, []string{"$"})
	chain.Unchain(preceder)
	assert.Nil(t, chain.UnchainedExpr, "UnchainedExpr should be nil for operator with unknown associativity")
}
