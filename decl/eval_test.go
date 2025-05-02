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

func TestEvalBinaryExpr(t *testing.T) {
	vm, env := setupTestVM()

	// --- Test Arithmetic ---
	// 1 + 2
	exprAdd := newBinExpr(newIntLit("1"), "+", newIntLit("2"))
	resAdd, errAdd := Eval(exprAdd, env, vm)
	require.NoError(t, errAdd)

	binOpAdd, okAdd := resAdd.(*BinaryOpNode)
	require.True(t, okAdd, "Expected *BinaryOpNode for add, got %T", resAdd)
	assert.Equal(t, "+", binOpAdd.Op)
	assertLeafInt(t, binOpAdd.Left, 1)
	assertLeafInt(t, binOpAdd.Right, 2)
	t.Logf("Eval(1 + 2): %s", resAdd)

	// --- Test Boolean ---
	// true && false
	exprAnd := newBinExpr(newBoolLit("true"), "&&", newBoolLit("false"))
	resAnd, errAnd := Eval(exprAnd, env, vm)
	require.NoError(t, errAnd)

	binOpAnd, okAnd := resAnd.(*BinaryOpNode)
	require.True(t, okAnd, "Expected *BinaryOpNode for and, got %T", resAnd)
	assert.Equal(t, "&&", binOpAnd.Op)
	assertLeafBool(t, binOpAnd.Left, true)
	assertLeafBool(t, binOpAnd.Right, false)
	t.Logf("Eval(true && false): %s", resAnd)

	// --- Test Comparison ---
	// 10 > 5
	exprGt := newBinExpr(newIntLit("10"), ">", newIntLit("5"))
	resGt, errGt := Eval(exprGt, env, vm)
	require.NoError(t, errGt)

	binOpGt, okGt := resGt.(*BinaryOpNode)
	require.True(t, okGt, "Expected *BinaryOpNode for gt, got %T", resGt)
	assert.Equal(t, ">", binOpGt.Op)
	assertLeafInt(t, binOpGt.Left, 10)
	assertLeafInt(t, binOpGt.Right, 5)
	t.Logf("Eval(10 > 5): %s", resGt)

	// --- Test with Variables ---
	// let x = 20; x * 3;
	block := newBlockStmt(
		newLetStmt("x", newIntLit("20")),
		newExprStmt(newBinExpr(newIdent("x"), "*", newIntLit("3"))),
	)
	resBlock, errBlock := Eval(block, env, vm) // Use fresh env for block
	require.NoError(t, errBlock)

	binOpMul, okMul := resBlock.(*BinaryOpNode)
	require.True(t, okMul, "Expected *BinaryOpNode from block, got %T", resBlock)
	assert.Equal(t, "*", binOpMul.Op)
	// Left operand should be the LeafNode associated with 'x'
	identNode, okIdent := binOpMul.Left.(*LeafNode) // Assuming identifier resolves directly for now
	require.True(t, okIdent, "Expected identifier 'x' to resolve to *LeafNode, got %T", binOpMul.Left)
	assertLeafInt(t, identNode, 20)
	// Right operand is the literal 3
	assertLeafInt(t, binOpMul.Right, 3)
	t.Logf("Eval({let x=20; x*3;}): %s", resBlock)

	// --- Test Nested Expressions ---
	// (1 + 2) * 3
	exprNested := newBinExpr(
		newBinExpr(newIntLit("1"), "+", newIntLit("2")), // Left operand is another BinaryExpr
		"*",
		newIntLit("3"), // Right operand
	)
	resNested, errNested := Eval(exprNested, env, vm) // Use same env is fine
	require.NoError(t, errNested)

	outerMul, okOuter := resNested.(*BinaryOpNode)
	require.True(t, okOuter, "Expected outer *BinaryOpNode, got %T", resNested)
	assert.Equal(t, "*", outerMul.Op)

	// Check right operand of outer '*'
	assertLeafInt(t, outerMul.Right, 3)

	// Check left operand of outer '*' (should be the inner '+')
	innerAdd, okInner := outerMul.Left.(*BinaryOpNode)
	require.True(t, okInner, "Expected inner node to be *BinaryOpNode, got %T", outerMul.Left)
	assert.Equal(t, "+", innerAdd.Op)
	assertLeafInt(t, innerAdd.Left, 1)
	assertLeafInt(t, innerAdd.Right, 2)
	t.Logf("Eval((1 + 2) * 3): %s", resNested)
}

func TestEvalIfStmt(t *testing.T) {
	vm, env := setupTestVM()

	// --- Test Basic If-Then ---
	// if true { 1 }
	ifStmt1 := newIfStmt(
		newBoolLit("true"),
		newBlockStmt(newExprStmt(newIntLit("1"))),
		nil, // No else
	)
	resIf1, errIf1 := Eval(ifStmt1, env, vm)
	require.NoError(t, errIf1)

	ifNode1, ok1 := resIf1.(*IfChoiceNode)
	require.True(t, ok1, "Expected *IfChoiceNode, got %T", resIf1)

	// Check condition node
	assertLeafBool(t, ifNode1.Condition, true)
	// Check then node
	assertLeafInt(t, ifNode1.Then, 1)
	// Check else node (should be NilNode)
	assertNilNode(t, ifNode1.Else)
	t.Logf("Eval(if true { 1 }): %s", resIf1)

	// --- Test Basic If-Then-Else ---
	// if 5 < 3 { 10 } else { 20 }
	ifStmt2 := newIfStmt(
		newBinExpr(newIntLit("5"), "<", newIntLit("3")), // Condition
		newBlockStmt(newExprStmt(newIntLit("10"))),      // Then
		newBlockStmt(newExprStmt(newIntLit("20"))),      // Else (BlockStmt)
	)
	resIf2, errIf2 := Eval(ifStmt2, env, vm)
	require.NoError(t, errIf2)

	ifNode2, ok2 := resIf2.(*IfChoiceNode)
	require.True(t, ok2, "Expected *IfChoiceNode, got %T", resIf2)

	// Check condition node (should be BinaryOpNode)
	condBinOp, okCond := ifNode2.Condition.(*BinaryOpNode)
	require.True(t, okCond, "Expected Condition to be *BinaryOpNode, got %T", ifNode2.Condition)
	assert.Equal(t, "<", condBinOp.Op)
	assertLeafInt(t, condBinOp.Left, 5)
	assertLeafInt(t, condBinOp.Right, 3)

	// Check then node
	assertLeafInt(t, ifNode2.Then, 10)
	// Check else node
	assertLeafInt(t, ifNode2.Else, 20)
	t.Logf("Eval(if 5 < 3 { 10 } else { 20 }): %s", resIf2)

	// --- Test If with Sequence in Branches ---
	// let x = 5;
	// if x == 5 { let y = 1; y+1; } else { 99; }
	mainBlock := newBlockStmt(
		newLetStmt("x", newIntLit("5")), // Outer scope let
		newIfStmt( // If statement using outer var
			newBinExpr(newIdent("x"), "==", newIntLit("5")), // Condition
			newBlockStmt( // Then block
				newLetStmt("y", newIntLit("1")),                             // Inner scope let
				newExprStmt(newBinExpr(newIdent("y"), "+", newIntLit("1"))), // Use inner var
			),
			newBlockStmt(newExprStmt(newIntLit("99"))), // Else block
		),
	)

	resBlock, errBlock := Eval(mainBlock, env, vm)
	require.NoError(t, errBlock)

	// The result of the block is the result of the if statement
	ifNode3, ok3 := resBlock.(*IfChoiceNode)
	require.True(t, ok3, "Expected *IfChoiceNode from block, got %T", resBlock)

	// Check condition node (accesses outer 'x')
	condBinOp3, okCond3 := ifNode3.Condition.(*BinaryOpNode)
	require.True(t, okCond3, "Expected Condition to be *BinaryOpNode")
	assert.Equal(t, "==", condBinOp3.Op)
	assert.IsType(t, &LeafNode{}, condBinOp3.Left, "Condition left operand should be leaf 'x'")
	assertLeafInt(t, condBinOp3.Right, 5)

	// Check then node (should be BinaryOpNode for y+1)
	thenBinOp, okThen := ifNode3.Then.(*BinaryOpNode)
	require.True(t, okThen, "Expected Then branch result to be *BinaryOpNode, got %T", ifNode3.Then)
	assert.Equal(t, "+", thenBinOp.Op)
	assert.IsType(t, &LeafNode{}, thenBinOp.Left, "Then '+' left operand should be leaf 'y'")
	assertLeafInt(t, thenBinOp.Right, 1)

	// Check else node
	assertLeafInt(t, ifNode3.Else, 99)
	t.Logf("Eval(complex if block): %s", resBlock)
}
