package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper struct for expected token properties
type expectedToken struct {
	tok        int    // Token type (e.g., IDENTIFIER, INT_LITERAL, COMPONENT, LBRACE)
	text       string // Raw token text as scanned by lexer
	startPos   int    // Expected start byte offset
	endPos     int    // Expected end byte offset
	startLine  int    // Expected start line
	startCol   int    // Expected start column
	literalVal Value  // For LiteralExpr: The parsed value
	identName  string // For IdentifierExpr: The name
}

// Helper function to run lexer tests
func runLexerTest(t *testing.T, input string, expectedTokens []expectedToken, ignoreErrors bool) (lexer *Lexer) {
	t.Helper()
	lexer = NewLexer(strings.NewReader(input))
	lval := &yySymType{} // The semantic value structure

	for i, exp := range expectedTokens {
		tok := lexer.Lex(lval)
		tokStr := tokenString(tok)
		expTokStr := tokenString(exp.tok)
		assert.Equal(t, exp.tok, tok, "Test %d: Token type mismatch. Expected %s, got %s ('%s')", i, expTokStr, tokStr, lexer.Text())
		assert.Equal(t, exp.text, lexer.Text(), "Test %d: Token text mismatch for %s.", i, expTokStr)
		assert.Equal(t, exp.startPos, lexer.Pos(), "Test %d: Token startPos mismatch for %s.", i, expTokStr)
		assert.Equal(t, exp.endPos, lexer.End(), "Test %d: Token endPos mismatch for %s.", i, expTokStr)
		assert.Equal(t, exp.startLine, lexer.tokenStartLine, "Test %d: Token startLine mismatch for %s.", i, expTokStr)
		assert.Equal(t, exp.startCol, lexer.tokenStartCol, "Test %d: Token startCol mismatch for %s.", i, expTokStr)

		// Check literal/identifier specific values if provided in expectation
		if exp.literalVal != nil {
			litExpr, ok := lval.expr.(*LiteralExpr)
			require.True(t, ok, "Test %d: Expected LiteralExpr for token %s, got %T", i, expTokStr, lval.expr)
			assert.Equal(t, exp.literalVal.Type.Tag, litExpr.Value.Type.Tag, "Test %d: Literal value type  mismatch for %s.", i, exp.tok)
			assert.Equal(t, exp.literalVal.Value, litExpr.Value.Value, "Test %d: Literal value type  mismatch for %s.", i, exp.tok)
			assert.Equal(t, exp.startPos, litExpr.Pos(), "Test %d: LiteralExpr startPos mismatch for %s.", i, expTokStr)
			assert.Equal(t, exp.endPos, litExpr.End(), "Test %d: LiteralExpr endPos mismatch for %s.", i, expTokStr)
		}
		if exp.identName != "" {
			identExpr := lval.ident
			require.NotNil(t, identExpr)
			assert.Equal(t, exp.identName, identExpr.Name, "Test %d: Identifier name mismatch for %s.", i, expTokStr)
			assert.Equal(t, exp.startPos, identExpr.Pos(), "Test %d: IdentifierExpr startPos mismatch for %s.", i, expTokStr)
			assert.Equal(t, exp.endPos, identExpr.End(), "Test %d: IdentifierExpr endPos mismatch for %s.", i, expTokStr)
		}

		// Check sval for operators that set it
		switch exp.tok {
		case OR, AND, EQ, NEQ, LT, LTE, GT, GTE, PLUS, MINUS, MUL, DIV, MOD, NOT:
			assert.Equal(t, exp.text, lval.sval, "Test %d: Operator sval mismatch for %s", i, expTokStr)
		}

		if tok == eof { // Should not happen if there are more expected tokens
			require.Equal(t, len(expectedTokens)-1, i, "Lexer returned EOF prematurely at token %d", i)
			break
		}
	}

	// After all expected tokens, Lex should return EOF
	finalTok := lexer.Lex(lval)
	assert.Equal(t, eof, finalTok, "Expected EOF after all tokens, got %s ('%s')", tokenString(finalTok), lexer.Text())
	if !ignoreErrors {
		assert.NoError(t, lexer.lastError, "Expected no lexer error at the end")
	}
	return
}

func TestLexer_KeywordsAndIdentifiers(t *testing.T) {
	input := "component MyComp system _sys1 let foo_bar123"
	expected := []expectedToken{
		{COMPONENT, "component", 0, 9, 1, 1, nil, ""},
		{IDENTIFIER, "MyComp", 10, 16, 1, 11, nil, "MyComp"},
		{SYSTEM, "system", 17, 23, 1, 18, nil, ""},
		{IDENTIFIER, "_sys1", 24, 29, 1, 25, nil, "_sys1"},
		{LET, "let", 30, 33, 1, 31, nil, ""},
		{IDENTIFIER, "foo_bar123", 34, 44, 1, 35, nil, "foo_bar123"},
	}
	runLexerTest(t, input, expected, false)
}

func TestLexer_Literals(t *testing.T) {
	input := `123 45.67 "hello world" true false 10ms 5s 100us 20ns 1.5s`
	expected := []expectedToken{
		{INT_LITERAL, "123", 0, 3, 1, 1, IntValue(123), ""},
		{FLOAT_LITERAL, "45.67", 4, 9, 1, 5, FloatValue(45.67), ""},
		{STRING_LITERAL, `"hello world"`, 10, 23, 1, 11, StringValue("hello world"), ""},
		{BOOL_LITERAL, "true", 24, 28, 1, 25, BoolValue(true), ""},
		{BOOL_LITERAL, "false", 29, 34, 1, 30, BoolValue(false), ""},
		{DURATION_LITERAL, "10ms", 35, 39, 1, 36, FloatValue(parseDuration("10", "ms")), ""},
		{DURATION_LITERAL, "5s", 40, 42, 1, 41, FloatValue(parseDuration("5", "s")), ""},
		{DURATION_LITERAL, "100us", 43, 48, 1, 44, FloatValue(parseDuration("100", "us")), ""},
		{DURATION_LITERAL, "20ns", 49, 53, 1, 50, FloatValue(parseDuration("20", "ns")), ""},
		{DURATION_LITERAL, "1.5s", 54, 58, 1, 55, FloatValue(parseDuration("1.5", "s")), ""},
	}
	runLexerTest(t, input, expected, false)
}

func TestLexer_OperatorsAndPunctuation(t *testing.T) {
	input := `:= = == != < <= > >= + - * / % => . , ; : ( ) { } || && !`
	expected := []expectedToken{
		{LET_ASSIGN, ":=", 0, 2, 1, 1, nil, ""},
		{ASSIGN, "=", 3, 4, 1, 4, nil, ""},
		{EQ, "==", 5, 7, 1, 6, nil, ""},
		{NEQ, "!=", 8, 10, 1, 9, nil, ""},
		{LT, "<", 11, 12, 1, 12, nil, ""},
		{LTE, "<=", 13, 15, 1, 14, nil, ""},
		{GT, ">", 16, 17, 1, 17, nil, ""},
		{GTE, ">=", 18, 20, 1, 19, nil, ""},
		{PLUS, "+", 21, 22, 1, 22, nil, ""},
		{MINUS, "-", 23, 24, 1, 24, nil, ""},
		{MUL, "*", 25, 26, 1, 26, nil, ""},
		{DIV, "/", 27, 28, 1, 28, nil, ""},
		{MOD, "%", 29, 30, 1, 30, nil, ""},
		{ARROW, "=>", 31, 33, 1, 32, nil, ""},
		{DOT, ".", 34, 35, 1, 35, nil, ""},
		{COMMA, ",", 36, 37, 1, 37, nil, ""},
		{SEMICOLON, ";", 38, 39, 1, 39, nil, ""},
		{COLON, ":", 40, 41, 1, 41, nil, ""},
		{LPAREN, "(", 42, 43, 1, 43, nil, ""},
		{RPAREN, ")", 44, 45, 1, 45, nil, ""},
		{LBRACE, "{", 46, 47, 1, 47, nil, ""},
		{RBRACE, "}", 48, 49, 1, 49, nil, ""},
		{OR, "||", 50, 52, 1, 51, nil, ""},
		{AND, "&&", 53, 55, 1, 54, nil, ""},
		{NOT, "!", 56, 57, 1, 57, nil, ""},
	}
	runLexerTest(t, input, expected, false)
}

func TestLexer_Comments(t *testing.T) {
	input := `
// This is a comment
component Abc // another comment
/* multi-line
   comment */ {
	param X: int; /* block
	comment too */
} // final
`
	// Expected positions need to be carefully calculated based on byte offsets
	// For simplicity, only checking token sequence here, then add position checks.
	expected := []expectedToken{
		{COMPONENT, "component", 22, 31, 3, 1, nil, ""}, // Line 3, Col 1 (after newline)
		{IDENTIFIER, "Abc", 32, 35, 3, 11, nil, "Abc"},
		{LBRACE, "{", 83, 84, 5, 15, nil, ""},   // After "/* multi-line \n comment */ "
		{PARAM, "param", 86, 91, 6, 2, nil, ""}, // Line 6, Col 2
		{IDENTIFIER, "X", 92, 93, 6, 8, nil, "X"},
		{COLON, ":", 93, 94, 6, 9, nil, ""},
		{INT, "int", 95, 98, 6, 11, nil, ""}, // Type name "int"
		{SEMICOLON, ";", 98, 99, 6, 14, nil, ""},
		{RBRACE, "}", 125, 126, 8, 1, nil, ""}, // Line 6, Col 1
	}
	runLexerTest(t, input, expected, false)
}

func TestLexer_StringEscapes(t *testing.T) {
	input := `"hello world" "hello\nworld\"\\\t"`
	expected := []expectedToken{
		{STRING_LITERAL, `"hello world"`, 0, 13, 1, 1, StringValue("hello world"), ""},
		{STRING_LITERAL, "\"hello\nworld\"\\\t\"", 14, 34, 1, 15, StringValue("hello\nworld\"\\\t"), ""},
	}
	runLexerTest(t, input, expected, false)
}

func TestLexer_LineColumnTracking(t *testing.T) {
	input := "abc\ndef\n  ghi"
	lexer := NewLexer(strings.NewReader(input))
	lval := &yySymType{}

	// abc
	tok := lexer.Lex(lval)
	assert.Equal(t, IDENTIFIER, tok)
	assert.Equal(t, 1, lexer.tokenStartLine)
	assert.Equal(t, 1, lexer.tokenStartCol)
	assert.Equal(t, 0, lexer.Pos())
	assert.Equal(t, 3, lexer.End())

	// def
	tok = lexer.Lex(lval)
	assert.Equal(t, IDENTIFIER, tok)
	assert.Equal(t, 2, lexer.tokenStartLine) // After '\n'
	assert.Equal(t, 1, lexer.tokenStartCol)
	assert.Equal(t, 4, lexer.Pos()) // 'a'(0) 'b'(1) 'c'(2) '\n'(3) 'd'(4)
	assert.Equal(t, 7, lexer.End())

	// ghi
	tok = lexer.Lex(lval)
	assert.Equal(t, IDENTIFIER, tok)
	assert.Equal(t, 3, lexer.tokenStartLine)
	assert.Equal(t, 3, lexer.tokenStartCol) // After '  '
	assert.Equal(t, 10, lexer.Pos())        // d(4)e(5)f(6)\n(7) (8) (9)g(10)
	assert.Equal(t, 13, lexer.End())

	tok = lexer.Lex(lval)
	assert.Equal(t, eof, tok)
}

func TestLexer_UnterminatedString(t *testing.T) {
	input := `"hello`
	lexer := NewLexer(strings.NewReader(input))
	lval := &yySymType{}
	lexer.Lex(lval) // Should call Error and return eof
	require.Error(t, lexer.lastError)
	assert.Contains(t, lexer.lastError.Error(), "unterminated string literal")
	assert.Equal(t, 1, lexer.tokenStartLine)
	assert.Equal(t, 1, lexer.tokenStartCol)
}

func TestLexer_InvalidEscape(t *testing.T) {
	input := `"hello\xworld"` // \x is not a supported escape
	lexer := NewLexer(strings.NewReader(input))
	lval := &yySymType{}
	tok := lexer.Lex(lval)
	require.Equal(t, STRING_LITERAL, tok) // Still returns a string token
	require.Error(t, lexer.lastError)     // But Error() should have been called
	assert.Contains(t, lexer.lastError.Error(), "invalid escape sequence \\x")
	// The lexer currently writes the 'x' to the buffer.
	litExpr := lval.expr.(*LiteralExpr)
	assert.Equal(t, StringValue("helloxworld"), litExpr.Value)
}

func TestLexer_ComplexDurations(t *testing.T) {
	input := `10 10.5ms 1s2 ` // "1s2" is 1s, then IDENTIFIER "s2"
	expected := []expectedToken{
		{INT_LITERAL, "10", 0, 2, 1, 1, IntValue(10), ""},
		{DURATION_LITERAL, "10.5ms", 3, 9, 1, 4, FloatValue(parseDuration("10.5", "ms")), ""},
		{INT_LITERAL, "1", 10, 11, 1, 11, IntValue(1), ""},
		{IDENTIFIER, "s2", 11, 13, 1, 12, nil, "s2"}, // The '2' from "1s2" should be unread and form a new token
	}
	lexer := runLexerTest(t, input, expected, true)
	require.Error(t, lexer.lastError)
	assert.Contains(t, lexer.lastError.Error(), "invalid character after unit or invalid unit")

	input2 := `1msident` // 1ms then ident
	expected2 := []expectedToken{
		{INT_LITERAL, "1", 0, 1, 1, 1, IntValue(1), ""},
		{IDENTIFIER, "msident", 1, 8, 1, 2, nil, "msident"},
	}
	runLexerTest(t, input2, expected2, true)
}

func TestLexer_DivisionAndMultilineComments(t *testing.T) {
	input := "a / b /* comment * test */ c /**/ d"
	expected := []expectedToken{
		{IDENTIFIER, "a", 0, 1, 1, 1, nil, "a"},
		{DIV, "/", 2, 3, 1, 3, nil, ""}, // Note: DIV token itself doesn't have lval.sval in current setup
		{IDENTIFIER, "b", 4, 5, 1, 5, nil, "b"},
		{IDENTIFIER, "c", 27, 28, 1, 28, nil, "c"}, // After "/* comment * test */ "
		{IDENTIFIER, "d", 34, 35, 1, 35, nil, "d"}, // After "/**/ "
	}
	runLexerTest(t, input, expected, false)

	input2 := "/*unterminated"
	lexer2 := NewLexer(strings.NewReader(input2))
	lval2 := &yySymType{}
	lexer2.Lex(lval2) // Should call Error and return eof
	require.Error(t, lexer2.lastError)
	assert.Contains(t, lexer2.lastError.Error(), "unterminated block comment")

	input3 := "a * /* b */ c" // Ensure '*' operator isn't confused by comment
	expected3 := []expectedToken{
		{IDENTIFIER, "a", 0, 1, 1, 1, nil, "a"},
		{MUL, "*", 2, 3, 1, 3, nil, ""},
		{IDENTIFIER, "c", 12, 13, 1, 13, nil, "c"},
	}
	runLexerTest(t, input3, expected3, false)
}
