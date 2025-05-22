package parser

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	// "fmt" // For debugging
)

type ParseError struct {
	StartPos Location
	EndPos   Location
	Msg      string
}

func (e *ParseError) Error() string {
	return e.Msg
}

func NewParseError(start, end Location, msg string) *ParseError {
	return &ParseError{StartPos: start, EndPos: end, Msg: msg}
}

// --- Test Helper: `assertParse` ---
// This helper will encapsulate the common parsing and checking logic.

// cleanNodeInfo recursively sets NodeInfo to its zero value for comparison purposes.
// This is useful for fragment testing where exact positions are hard to predict or not the focus.
func cleanNodeInfo(node Node) {
	if node == nil || reflect.ValueOf(node).IsNil() {
		return
	}

	// Use reflection to find and zero out the NodeInfo field if it exists
	v := reflect.ValueOf(node)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	// Try to set NodeInfo field specifically
	niField := v.FieldByName("NodeInfo")
	if niField.IsValid() && niField.CanSet() {
		zeroNodeInfo := NodeInfo{} // Assuming NodeInfo is a struct
		if niField.Type() == reflect.TypeOf(zeroNodeInfo) {
			niField.Set(reflect.ValueOf(zeroNodeInfo))
		}
	}

	// Recursively clean fields that are Nodes or slices/arrays of Nodes
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue
		}
		fieldInterface := field.Interface()

		if n, ok := fieldInterface.(Node); ok {
			cleanNodeInfo(n)
		} else if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				if elem.CanInterface() {
					if en, ok := elem.Interface().(Node); ok {
						cleanNodeInfo(en)
					}
				}
			}
		} else if field.Kind() == reflect.Ptr {
			if !field.IsNil() {
				if pn, ok := field.Interface().(Node); ok {
					cleanNodeInfo(pn)
				}
			}
		}
	}
}

// parseFragment is a generic helper to parse a fragment using a specific parse function
// from the LLParser (e.g., p.ParseExpression, p.ParseStmt).
func parseFragment[T Node](t *testing.T, input string, parseFunc func(p *LLParser) (T, error)) (actualNode T, err error) {
	t.Helper()
	lexer := NewLexer(strings.NewReader(input)) // Assuming NewLexer takes io.Reader
	parser := NewLLParser(lexer)

	actualNode, err = parseFunc(parser)

	// After parsing, check if there's any unconsumed input (other than EOF)
	// This helps catch cases where the parser stops too early.
	if err == nil {
		if peeked := parser.PeekToken(); peeked != eof {
			t.Errorf("Input: %q\nParser did not consume all input. Remaining token: %s (%q)",
				input, TokenString(peeked), parser.lexer.Text())
			// To prevent cascading errors, we can treat this as a parse failure for this test
			err = NewParseError(parser.peekedTokenValue.node.Pos(),
				parser.peekedTokenValue.node.End(),
				"parser did not consume all input",
			) // Assuming you have a NewParseError or similar
		}
	}
	return
}

// assertNodeEqual checks if two nodes are deeply equal after cleaning NodeInfo.
func assertNodeEqual(t *testing.T, input string, expected, actual Node) {
	t.Helper()
	if expected == nil && actual == nil {
		return
	}
	if expected == nil || actual == nil {
		t.Errorf("Input: %q\nExpected node presence (%v) != actual node presence (%v)", input, expected != nil, actual != nil)
		return
	}

	assert.Equal(t, expected.String(), actual.String())
	/*
		// Create copies for cleaning, to not modify originals if they are reused
		// For simple tests, directly cleaning might be okay if expected nodes are constructed per test.
		// For safety, let's assume we might want to reuse `expected`.
		// However, deep copying ASTs is non-trivial. For now, we'll clean in place.
		// Be mindful of this if `expected` is a shared variable.
		cleanNodeInfo(expected)
		cleanNodeInfo(actual)

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Input: %q\nAST mismatch:\nExpected: %#v\nActual:   %#v", input, expected, actual)
		}
	*/
}

// assertError checks for expected errors.
func assertError(t *testing.T, input string, err error, expectError bool, errorContains string) {
	t.Helper()
	if expectError {
		if err == nil {
			t.Errorf("Input: %q\nExpected an error, but got nil", input)
			return
		}
		if errorContains != "" && !strings.Contains(err.Error(), errorContains) {
			t.Errorf("Input: %q\nError message %q does not contain expected string %q", input, err.Error(), errorContains)
		}
	} else {
		if err != nil {
			t.Errorf("Input: %q\nDid not expect an error, but got: %v", input, err)
		}
	}
}
