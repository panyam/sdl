package decl

import (
	"testing"

	"github.com/panyam/leetcoach/sdl/core" // For core types
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Helpers ---

// Helper to create a basic VM and Env for testing
func setupTestVM() (*VM, *Env[any]) {
	vm := &VM{}
	vm.Init() // Initialize internal maps etc.
	env := NewEnv[any](nil)
	return vm, env
}

// Helper to create simple AST nodes for testing
func newIntLit(val string) *LiteralExpr {
	return &LiteralExpr{Kind: "INT", Value: val}
}

func newBoolLit(val string) *LiteralExpr {
	return &LiteralExpr{Kind: "BOOL", Value: val}
}

func newIdent(name string) *IdentifierExpr {
	return &IdentifierExpr{Name: name}
}

func newLetStmt(varName string, value Expr) *LetStmt {
	return &LetStmt{Variable: newIdent(varName), Value: value}
}

func newExprStmt(expr Expr) *ExprStmt {
	return &ExprStmt{Expression: expr}
}

func newBlockStmt(stmts ...Stmt) *BlockStmt {
	return &BlockStmt{Statements: stmts}
}

// Helper to assert leaf node properties (simplistic check for now)
func assertLeafInt(t *testing.T, node OpNode, expectedVal int64) {
	t.Helper()
	leaf, ok := node.(*LeafNode)
	require.True(t, ok, "Expected *LeafNode, got %T", node)
	require.NotNil(t, leaf.State, "LeafNode state is nil")
	require.NotNil(t, leaf.State.ValueOutcome, "LeafNode ValueOutcome is nil")

	valOutcome, ok := leaf.State.ValueOutcome.(*core.Outcomes[int64])
	require.True(t, ok, "Expected ValueOutcome *core.Outcomes[int64], got %T", leaf.State.ValueOutcome)
	require.Equal(t, 1, valOutcome.Len(), "Expected 1 bucket in ValueOutcome")
	assert.Equal(t, 1.0, valOutcome.Buckets[0].Weight, "ValueOutcome bucket weight")
	assert.Equal(t, expectedVal, valOutcome.Buckets[0].Value, "ValueOutcome bucket value")

	// Check latency is zero outcome
	require.NotNil(t, leaf.State.LatencyOutcome, "LeafNode LatencyOutcome is nil")
	latOutcome, ok := leaf.State.LatencyOutcome.(*core.Outcomes[core.Duration])
	require.True(t, ok, "Expected LatencyOutcome *core.Outcomes[core.Duration], got %T", leaf.State.LatencyOutcome)
	require.Equal(t, 1, latOutcome.Len(), "Expected 1 bucket in LatencyOutcome")
	assert.Equal(t, 1.0, latOutcome.Buckets[0].Weight, "LatencyOutcome bucket weight")
	assert.Equal(t, 0.0, latOutcome.Buckets[0].Value, "LatencyOutcome bucket value")
}

func assertLeafBool(t *testing.T, node OpNode, expectedVal bool) {
	t.Helper()
	leaf, ok := node.(*LeafNode)
	require.True(t, ok, "Expected *LeafNode, got %T", node)
	require.NotNil(t, leaf.State, "LeafNode state is nil")
	require.NotNil(t, leaf.State.ValueOutcome, "LeafNode ValueOutcome is nil")

	valOutcome, ok := leaf.State.ValueOutcome.(*core.Outcomes[bool])
	require.True(t, ok, "Expected ValueOutcome *core.Outcomes[bool], got %T", leaf.State.ValueOutcome)
	// ... (rest of assertions similar to assertLeafInt)
	assert.Equal(t, expectedVal, valOutcome.Buckets[0].Value, "ValueOutcome bucket value")
	// ... (assert zero latency)
}

func assertNilNode(t *testing.T, node OpNode) {
	t.Helper()
	_, ok := node.(*NilNode)
	assert.True(t, ok, "Expected *NilNode, got %T", node)
}

// --- Actual Tests ---

func TestEvalLiteral(t *testing.T) {
	vm, env := setupTestVM()

	// Test INT
	nodeInt := newIntLit("123")
	resInt, errInt := Eval(nodeInt, env, vm)
	require.NoError(t, errInt)
	assertLeafInt(t, resInt, 123)
	t.Logf("Eval(123): %s", resInt)

	// Test BOOL
	nodeBool := newBoolLit("true")
	resBool, errBool := Eval(nodeBool, env, vm)
	require.NoError(t, errBool)
	assertLeafBool(t, resBool, true)
	t.Logf("Eval(true): %s", resBool)

	// TODO: Test STRING, FLOAT, DURATION
}

func TestEvalIdentifier(t *testing.T) {
	vm, env := setupTestVM()
	// Setup env manually for testing Get
	expectedNode := &LeafNode{State: createNilState()} // Example node
	env.Set("myVar", expectedNode)

	// Test Found
	nodeIdentFound := newIdent("myVar")
	resFound, errFound := Eval(nodeIdentFound, env, vm)
	require.NoError(t, errFound)
	assert.Same(t, expectedNode, resFound, "Identifier lookup returned wrong node")
	t.Logf("Eval(myVar): %s", resFound)

	// Test Not Found
	nodeIdentNotFound := newIdent("noVar")
	_, errNotFound := Eval(nodeIdentNotFound, env, vm)
	require.Error(t, errNotFound)
	assert.ErrorIs(t, errNotFound, ErrNotFound)
	t.Logf("Eval(noVar): %s", errNotFound)

	// Test wrong type in env (shouldn't happen with current Set, but good check)
	env.Set("badVar", 123) // Put non-OpNode
	nodeIdentBad := newIdent("badVar")
	_, errBad := Eval(nodeIdentBad, env, vm)
	require.Error(t, errBad)
	assert.Contains(t, errBad.Error(), "internal error: expected OpNode")
	t.Logf("Eval(badVar): %s", errBad)
}

func TestEvalLetStmt(t *testing.T) {
	vm, env := setupTestVM()

	// let x = 5;
	letNode := newLetStmt("x", newIntLit("5"))
	resLet, errLet := Eval(letNode, env, vm)
	require.NoError(t, errLet)
	assertNilNode(t, resLet) // Let statement itself returns NilNode

	// Verify env
	storedNode, ok := env.Get("x")
	require.True(t, ok, "Variable 'x' not found in env after let stmt")
	assertLeafInt(t, storedNode.(OpNode), 5)
	t.Logf("Env after 'let x = 5': %s", env)
}

func TestEvalBlockStmt(t *testing.T) {
	vm, env := setupTestVM()

	// Test empty block -> NilNode
	blockEmpty := newBlockStmt()
	resEmpty, errEmpty := Eval(blockEmpty, env, vm)
	require.NoError(t, errEmpty)
	assertNilNode(t, resEmpty)
	t.Logf("Eval({}): %s", resEmpty)

	// Test block with only let -> NilNode
	blockLetOnly := newBlockStmt(newLetStmt("a", newIntLit("1")))
	resLetOnly, errLetOnly := Eval(blockLetOnly, env, vm)
	require.NoError(t, errLetOnly)
	assertNilNode(t, resLetOnly)
	t.Logf("Eval({let a=1;}): %s", resLetOnly)

	// Test block with single expression -> LeafNode
	blockSingleExpr := newBlockStmt(newExprStmt(newIntLit("42")))
	resSingleExpr, errSingleExpr := Eval(blockSingleExpr, env, vm)
	require.NoError(t, errSingleExpr)
	assertLeafInt(t, resSingleExpr, 42)
	t.Logf("Eval({42;}): %s", resSingleExpr)

	// Test block with let and expression -> LeafNode (last expr)
	blockLetExpr := newBlockStmt(
		newLetStmt("b", newIntLit("10")),
		newExprStmt(newIntLit("99")),
	)
	resLetExpr, errLetExpr := Eval(blockLetExpr, env, vm)
	require.NoError(t, errLetExpr)
	assertLeafInt(t, resLetExpr, 99)
	// Check env state *outside* the block (should be unchanged)
	_, okOuter := env.Get("b")
	assert.False(t, okOuter, "'b' should not be visible outside the block")
	t.Logf("Eval({let b=10; 99;}): %s", resLetExpr)

	// Test block with multiple expressions -> SequenceNode
	blockMultiExpr := newBlockStmt(
		newExprStmt(newIntLit("1")),
		newExprStmt(newBoolLit("true")),
	)
	resMultiExpr, errMultiExpr := Eval(blockMultiExpr, env, vm)
	require.NoError(t, errMultiExpr)
	seqNode, ok := resMultiExpr.(*SequenceNode)
	require.True(t, ok, "Expected *SequenceNode, got %T", resMultiExpr)
	require.Len(t, seqNode.Steps, 2, "SequenceNode should have 2 steps")
	assertLeafInt(t, seqNode.Steps[0], 1)
	assertLeafBool(t, seqNode.Steps[1], true)
	t.Logf("Eval({1; true;}): %s", resMultiExpr)
}
