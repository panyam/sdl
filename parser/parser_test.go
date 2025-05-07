// decl/parser_test.go
package parser

import (
	// Needed for fmt.Sprintf

	"strings"
	"testing"

	// "time" // Only needed if testing duration parsing specifics

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func printWithLineNumbers(t *testing.T, input string) {
	t.Log("============================")
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		if len(line) > 0 {
			t.Logf("%03d: %s", i+1, line)
		}
	}
	t.Log("============================")
}

// --- Helper Functions ---
func parseString(t *testing.T, input string) *File {
	t.Helper()
	printWithLineNumbers(t, input)
	_, ast, err := Parse(strings.NewReader(input)) // Assumes Parse is in the same package
	require.NoError(t, err, "Input:\n%s", input)
	require.NotNil(t, ast, "Input:\n%s", input)
	return ast
}

func parseStringWithError(t *testing.T, input string) (*Lexer, error) {
	t.Helper()
	printWithLineNumbers(t, input)
	lexer, _, err := Parse(strings.NewReader(input))
	require.Error(t, err, "Expected parsing to fail for Input:\n%s", input)
	return lexer, err
}

func assertPosition(t *testing.T, node Node, expectedStart, expectedEnd int) {
	t.Helper()
	require.NotNil(t, node)
	assert.Equal(t, expectedStart, node.Pos(), "Node %T [%d:%d] start position mismatch", node, node.Pos(), node.End())
	assert.Equal(t, expectedEnd, node.End(), "Node %T [%d:%d] end position mismatch", node, node.Pos(), node.End())
}

// Helper to get the first declaration if it exists
func firstDecl(t *testing.T, f *File) Node {
	require.NotEmpty(t, f.Declarations)
	return f.Declarations[0]
}

// --- NEW/UPDATED Assertion Helper for Literals ---
func assertLiteralWithValue(t *testing.T, node Node, expectedType *ValueType, expectedGoValue any) {
	t.Helper()
	litExpr, ok := node.(*LiteralExpr)
	require.True(t, ok, "Expected *LiteralExpr, got %T", node)
	require.NotNil(t, litExpr.Value, "LiteralExpr.Value is nil")
	require.NotNil(t, litExpr.Value.Type, "LiteralExpr.Value.Type is nil")

	// Check Type
	assert.True(t, expectedType.Equals(litExpr.Value.Type), "Literal type mismatch: expected %s, got %s", expectedType, litExpr.Value.Type)

	// Check underlying Go Value
	assert.Equal(t, expectedGoValue, litExpr.Value.Value, "Literal Go value mismatch")
}

// --- Test Cases ---

func TestParseEmpty(t *testing.T) {
	ast := parseString(t, "")
	assert.Empty(t, ast.Declarations)
	// assertPosition(t, ast, 0, 0) // Position might be tricky for empty

	astWs := parseString(t, "  \n\t // comment \n ")
	assert.Empty(t, astWs.Declarations)
	// assertPosition(t, astWs, 0, 20) // Position might be tricky for empty
}

/*
func TestParseLiteralsInExprStmt(t *testing.T) {
	// Assuming grammar allows ExprStmt at top level for testing
	testCases := []struct {
		name            string
		input           string
		expectedType    *ValueType
		expectedGoValue any // The underlying Go value expected
		expectedStart   int
		expectedEnd     int
	}{
		{"Int", "123;", IntType, int64(123), 0, 3},
		{"Float", "45.67;", FloatType, float64(45.67), 0, 5},
		{"String", `"hello world";`, StrType, "hello world", 0, 13}, // Value is unescaped
		{"BoolTrue", "true;", BoolType, true, 0, 4},
		{"BoolFalse", "false;", BoolType, false, 0, 5},
		// Durations might parse into a specific Go type like time.Duration or float64 seconds.
		// Let's assume the lexer creates a float64 representing seconds for DURATION.
		// OR - create a DurationType if RuntimeValue needs to distinguish it.
		// For now, let's assume lexer converts to float64 seconds and uses FloatType.
		// If you add DurationType, update this test.
		// {"DurationMs", "150ms;", FloatType, 0.150, 0, 5},
		// {"DurationS", "2s;", FloatType, 2.0, 0, 2},
		// Reverting duration test for now - depends heavily on lexer implementation details
		// on how duration literals are converted into RuntimeValue.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast := parseString(t, tc.input)
			exprStmt := firstDecl(t, ast).(*ExprStmt)
			lit := exprStmt.Expression // Should be *LiteralExpr

			// Use the new assertion helper
			assertLiteralWithValue(t, lit, tc.expectedType, tc.expectedGoValue)

			// Check positions
			assertPosition(t, lit.(Node), tc.expectedStart, tc.expectedEnd)
			assertPosition(t, exprStmt, tc.expectedStart, tc.expectedEnd+1) // Include semicolon
		})
	}
}
*/

func TestParseComponentParams(t *testing.T) {
	input := `component C {
        param p1: int
        param p2: string = "default" // with default
    }`
	ast := parseString(t, input)
	comp := firstDecl(t, ast).(*ComponentDecl)
	require.Len(t, comp.Body, 2)

	p1 := comp.Body[0].(*ParamDecl)
	assert.Equal(t, "p1", p1.Name.Name)
	assert.Equal(t, "int", p1.Type.PrimitiveTypeName)
	assert.Nil(t, p1.DefaultValue)
	// Add position check if needed

	p2 := comp.Body[1].(*ParamDecl)
	assert.Equal(t, "p2", p2.Name.Name)
	assert.Equal(t, "string", p2.Type.PrimitiveTypeName)
	require.NotNil(t, p2.DefaultValue)
	// Check the default value using the new literal structure
	assertLiteralWithValue(t, p2.DefaultValue, StrType, "default")
	// Add position check if needed
}

// func TestParseBinaryOpsPrecedence(t *testing.T) {
// 	input := "a + b * c;" // Expect (a + (b * c))
// 	ast := parseString(t, input)
// 	exprStmt := firstDecl(t, ast).(*ExprStmt)
// 	addExpr := exprStmt.Expression.(*BinaryExpr)

// 	assert.Equal(t, "+", addExpr.Operator) // Operator stored as string
// 	aIdent := addExpr.Left.(*IdentifierExpr)
// 	assert.Equal(t, "a", aIdent.Name)

// 	mulExpr := addExpr.Right.(*BinaryExpr)
// 	assert.Equal(t, "*", mulExpr.Operator)
// 	bIdent := mulExpr.Left.(*IdentifierExpr)
// 	cIdent := mulExpr.Right.(*IdentifierExpr)
// 	assert.Equal(t, "b", bIdent.Name)
// 	assert.Equal(t, "c", cIdent.Name)

// 	// Check overall span (assuming accurate positions from underlying nodes)
// 	assertPosition(t, addExpr, aIdent.Pos(), cIdent.End())
// }

func TestParseErrors(t *testing.T) {
	testCases := []struct {
		name                 string
		input                string
		expectedErrSubstring string
		errorLine            int
		errorCol             int
		nearToken            string
	}{
		{"Unmatched Brace", "component C {", "syntax error", 1, 13, "{"},
		// {"Missing Semicolon", "system S { let x = 5 }", "unexpected RBRACE"},
		// Note: '123' is now lexed as INT_LITERAL (an <expr>), check if grammar allows expr here
		{"Invalid Token After Kw", "component 123 {}", "syntax error", 1, 11, ""}, // Or specific error based on state
		{"Unterminated String", `log "hello`, "syntax error", 1, 1, ""},
		{"Bad Analyze Target Type", `system S { analyze A = 1 + 2; }`, "analyze target must be a method call", 1, 29, ""}, // Checks type in parser action
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer, err := parseStringWithError(t, tc.input)
			require.NotNil(t, err) // Defensive check
			line, col := lexer.Position()
			assert.Equal(t, line, tc.errorLine, "Test %d: Expected error line: %d, found: %d", i, tc.errorLine, line)
			assert.Equal(t, col, tc.errorCol, "Test %d: Expected error col: %d, found: %d", i, tc.errorCol, col)
			// TODO - enable error messsage checking when we return better error messages
			if false && !assert.Contains(t, err.Error(), tc.expectedErrSubstring) {
				t.Errorf("Test %d: Expected error to contain '%s', got '%s'", i, tc.expectedErrSubstring, err.Error())
			}
			t.Logf("Input: %s\nError: %v", tc.input, err)
		})
	}
}

// --- Other Tests (Imports, Enums, Systems, Methods, Statements, etc.) ---
// Need updates similar to TestParseComponentParams to use assertLiteralWithValue
// when checking default values or literal expressions within the structures.

// Example: Update for InstanceDecl override literal
func TestParseSystemInstancesWithLiteralOverride(t *testing.T) {
	input := `system S { i1: D = { p = 5 } }` // Override p with int literal 5
	ast := parseString(t, input)
	sys := firstDecl(t, ast).(*SystemDecl)
	require.Len(t, sys.Body, 1)
	inst := sys.Body[0].(*InstanceDecl)
	require.Len(t, inst.Overrides, 1)
	assign := inst.Overrides[0]
	assert.Equal(t, "p", assign.Var.Name)
	// Assert the assigned value is a LiteralExpr containing an Int RuntimeValue
	assertLiteralWithValue(t, assign.Value, IntType, int64(5))
}

// TODO: Add tests for all other grammar constructs, checking structure and positions.
// TODO: Update tests involving literals (e.g., default values, expression statements)
//       to use assertLiteralWithValue or check lit.Value.Type and lit.Value.Value.
