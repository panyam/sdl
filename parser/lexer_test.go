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
	lval := &SDLSymType{} // The semantic value structure

	for i, exp := range expectedTokens {
		tok := lexer.Lex(lval)
		// Get position info *after* lexing the token
		tokenStart := lexer.Pos()

		tokStr := TokenString(tok)
		expTokStr := TokenString(exp.tok)
		assert.Equal(t, exp.tok, tok, "Test %d: Token type mismatch. Expected %s, got %s ('%s')", i, expTokStr, tokStr, lexer.Text())
		assert.Equal(t, exp.text, lexer.Text(), "Test %d: Token text mismatch for %s.", i, expTokStr)
		assert.Equal(t, exp.startPos, lexer.Pos().Pos, "Test %d: Token startPos mismatch for %s.", i, expTokStr)
		assert.Equal(t, exp.endPos, lexer.End().Pos, "Test %d: Token endPos mismatch for %s.", i, expTokStr)
		assert.Equal(t, exp.startLine, tokenStart.Line, "Test %d: Token startLine mismatch for %s.", i, expTokStr)
		assert.Equal(t, exp.startCol, tokenStart.Col, "Test %d: Token startCol mismatch for %s.", i, expTokStr)

		// Check literal/identifier specific values if provided in expectation
		if !exp.literalVal.IsNil() {
			litExpr, ok := lval.expr.(*LiteralExpr)
			require.True(t, ok, "Test %d: Expected LiteralExpr for token %s, got %T", i, expTokStr, lval.expr)
			assert.Equal(t, exp.literalVal.Type.Tag, litExpr.Value.Type.Tag, "Test %d: Literal value type  mismatch for %s.", i, exp.tok)
			assert.Equal(t, exp.literalVal.Value, litExpr.Value.Value, "Test %d: Literal value type  mismatch for %s.", i, exp.tok)
			assert.Equal(t, exp.startPos, litExpr.Pos().Pos, "Test %d: LiteralExpr startPos mismatch for %s.", i, expTokStr)
			assert.Equal(t, exp.endPos, litExpr.End().Pos, "Test %d: LiteralExpr endPos mismatch for %s.", i, expTokStr)
		}
		if exp.identName != "" {
			identExpr := lval.ident
			require.NotNil(t, identExpr)
			assert.Equal(t, exp.identName, identExpr.Name, "Test %d: Identifier name mismatch for %s.", i, expTokStr)
			assert.Equal(t, exp.startPos, identExpr.Pos().Pos, "Test %d: IdentifierExpr startPos mismatch for %s.", i, expTokStr)
			assert.Equal(t, exp.endPos, identExpr.End().Pos, "Test %d: IdentifierExpr endPos mismatch for %s.", i, expTokStr)
		}

		// Check sval for operators that set it
		// This check is adjusted because the lexer now groups all operator chars into BINARY_OP
		if exp.tok == BINARY_OP {
			assert.Equal(t, exp.text, lval.sval, "Test %d: Operator sval mismatch for BINARY_OP (%s)", i, exp.text)
		} else {
			// Original check for specific operators, might be less relevant now if all are BINARY_OP
			switch exp.tok {
			case OR, AND, EQ, NEQ, LT, LTE, GT, GTE, PLUS, MINUS, MUL, DIV, MOD:
				assert.Equal(t, exp.text, lval.sval, "Test %d: Operator sval mismatch for specific op %s", i, expTokStr)
			}
		}

		if tok == eof { // Should not happen if there are more expected tokens
			require.Equal(t, len(expectedTokens)-1, i, "Lexer returned EOF prematurely at token %d", i)
			break
		}
	}

	// After all expected tokens, Lex should return EOF
	finalTok := lexer.Lex(lval)
	assert.Equal(t, eof, finalTok, "Expected EOF after all tokens, got %s ('%s')", TokenString(finalTok), lexer.Text())
	if !ignoreErrors {
		assert.NoError(t, lexer.lastError, "Expected no lexer error at the end")
	}
	return
}

func TestLexer_KeywordsAndIdentifiers(t *testing.T) {
	input := "component MyComp system _sys1 let foo_bar123"
	expected := []expectedToken{
		{COMPONENT, "component", 0, 9, 1, 1, Nil(), ""},
		{IDENTIFIER, "MyComp", 10, 16, 1, 11, Nil(), "MyComp"},
		{SYSTEM, "system", 17, 23, 1, 18, Nil(), ""},
		{IDENTIFIER, "_sys1", 24, 29, 1, 25, Nil(), "_sys1"},
		{LET, "let", 30, 33, 1, 31, Nil(), ""},
		{IDENTIFIER, "foo_bar123", 34, 44, 1, 35, Nil(), "foo_bar123"},
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
	// Adjusted to reflect that sequences of opchars are lexed as a single BINARY_OP token.
	// The punctuation characters . , ; : ( ) { } are still separate tokens.
	// :=, =>, ==, !=, <=, >=, ||, && will be single BINARY_OP tokens with that text.
	// Single ops like +, -, *, /, %, <, > will also be BINARY_OP.
	input := `:= = == != < <= > >= + - * / % => . , ; : ( ) { } || && !`
	expected := []expectedToken{
		{LET_ASSIGN, ":=", 0, 2, 1, 1, Nil(), ""}, // Keep as is, it has a specific token type in grammar.y
		{ASSIGN, "=", 3, 4, 1, 4, Nil(), ""},
		// All subsequent operators are now BINARY_OP
		{BINARY_OP, "==", 5, 7, 1, 6, Nil(), ""},
		{BINARY_OP, "!=", 8, 10, 1, 9, Nil(), ""},
		{BINARY_OP, "<", 11, 12, 1, 12, Nil(), ""},
		{BINARY_OP, "<=", 13, 15, 1, 14, Nil(), ""},
		{BINARY_OP, ">", 16, 17, 1, 17, Nil(), ""},
		{BINARY_OP, ">=", 18, 20, 1, 19, Nil(), ""},
		{BINARY_OP, "+", 21, 22, 1, 22, Nil(), ""},
		{MINUS, "-", 23, 24, 1, 24, Nil(), ""},
		{BINARY_OP, "*", 25, 26, 1, 26, Nil(), ""},
		{BINARY_OP, "/", 27, 28, 1, 28, Nil(), ""},
		{BINARY_OP, "%", 29, 30, 1, 30, Nil(), ""},
		{ARROW, "=>", 31, 33, 1, 32, Nil(), ""}, // Keep as is, it has a specific token type
		// Punctuation remains the same
		{DOT, ".", 34, 35, 1, 35, Nil(), ""},
		{COMMA, ",", 36, 37, 1, 37, Nil(), ""},
		{SEMICOLON, ";", 38, 39, 1, 39, Nil(), ""},
		{COLON, ":", 40, 41, 1, 41, Nil(), ""},
		{LPAREN, "(", 42, 43, 1, 43, Nil(), ""},
		{RPAREN, ")", 44, 45, 1, 45, Nil(), ""},
		{LBRACE, "{", 46, 47, 1, 47, Nil(), ""},
		{RBRACE, "}", 48, 49, 1, 49, Nil(), ""},
		// Logical operators are also BINARY_OP now
		{BINARY_OP, "||", 50, 52, 1, 51, Nil(), ""},
		{BINARY_OP, "&&", 53, 55, 1, 54, Nil(), ""},
		{BINARY_OP, "!", 56, 57, 1, 57, Nil(), ""}, // NOT could be UNARY_OP if lexer distinguishes
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
	expected := []expectedToken{
		{COMPONENT, "component", 22, 31, 3, 1, Nil(), ""},
		{IDENTIFIER, "Abc", 32, 35, 3, 11, Nil(), "Abc"},
		{LBRACE, "{", 83, 84, 5, 15, Nil(), ""},
		{PARAM, "param", 86, 91, 6, 2, Nil(), ""},
		{IDENTIFIER, "X", 92, 93, 6, 8, Nil(), "X"},
		{COLON, ":", 93, 94, 6, 9, Nil(), ""},
		{IDENTIFIER, "int", 95, 98, 6, 11, Nil(), "int"}, // Corrected: INT keyword is lexed as IDENTIFIER
		{SEMICOLON, ";", 98, 99, 6, 14, Nil(), ""},
		{RBRACE, "}", 125, 126, 8, 1, Nil(), ""},
	}
	runLexerTest(t, input, expected, false)
}

func TestLexer_StringEscapes(t *testing.T) {
	input := `"hello world" "hello\nworld\"\\\t"`
	expected := []expectedToken{
		{STRING_LITERAL, `"hello world"`, 0, 13, 1, 1, StringValue("hello world"), ""},
		{STRING_LITERAL, `"hello\nworld\"\\\t"`, 14, 34, 1, 15, StringValue("hello\nworld\"\\\t"), ""},
	}
	runLexerTest(t, input, expected, false)
}

func TestLexer_LineColumnTracking(t *testing.T) {
	input := "abc\ndef\n  ghi"
	lexer := NewLexer(strings.NewReader(input))
	lval := &SDLSymType{}

	// abc
	tok := lexer.Lex(lval)
	tokenStart := lexer.Pos() // Get position *after* Lex
	assert.Equal(t, IDENTIFIER, tok)
	assert.Equal(t, 1, tokenStart.Line)
	assert.Equal(t, 1, tokenStart.Col)
	assert.Equal(t, 0, lexer.Pos().Pos)
	assert.Equal(t, 3, lexer.End().Pos)

	// def
	tok = lexer.Lex(lval)
	tokenStart = lexer.Pos() // Get position *after* Lex
	assert.Equal(t, IDENTIFIER, tok)
	assert.Equal(t, 2, tokenStart.Line) // After '\n'
	assert.Equal(t, 1, tokenStart.Col)
	assert.Equal(t, 4, lexer.Pos().Pos) // 'a'(0) 'b'(1) 'c'(2) '\n'(3) 'd'(4)
	assert.Equal(t, 7, lexer.End().Pos)

	// ghi
	tok = lexer.Lex(lval)
	tokenStart = lexer.Pos() // Get position *after* Lex
	assert.Equal(t, IDENTIFIER, tok)
	assert.Equal(t, 3, tokenStart.Line)
	assert.Equal(t, 3, tokenStart.Col)   // After '  '
	assert.Equal(t, 10, lexer.Pos().Pos) // d(4)e(5)f(6)\n(7) (8) (9)g(10)
	assert.Equal(t, 13, lexer.End().Pos)

	tok = lexer.Lex(lval)
	assert.Equal(t, eof, tok)
}

func TestLexer_UnterminatedString(t *testing.T) {
	input := `"hello`
	lexer := NewLexer(strings.NewReader(input))
	lval := &SDLSymType{}
	lexer.Lex(lval) // Should call Error and return eof
	require.Error(t, lexer.lastError)
	assert.Contains(t, lexer.lastError.Error(), "unterminated string literal")
}

func TestLexer_InvalidEscape(t *testing.T) {
	input := `"hello\xworld"` // \x is not a supported escape
	lexer := NewLexer(strings.NewReader(input))
	lval := &SDLSymType{}
	tok := lexer.Lex(lval)
	require.Equal(t, STRING_LITERAL, tok) // Still returns a string token
	require.Error(t, lexer.lastError)     // But Error() should have been called
	assert.Contains(t, lexer.lastError.Error(), "invalid escape sequence \\x")
	litExpr := lval.expr.(*LiteralExpr)
	assert.Equal(t, StringValue("helloxworld"), litExpr.Value)
}

func TestLexer_ComplexDurations(t *testing.T) {
	input := `10 10.5ms 1s2 `
	expected := []expectedToken{
		{INT_LITERAL, "10", 0, 2, 1, 1, IntValue(10), ""},
		{DURATION_LITERAL, "10.5ms", 3, 9, 1, 4, FloatValue(parseDuration("10.5", "ms")), ""},
		{INT_LITERAL, "1", 10, 11, 1, 11, IntValue(1), ""},
		{IDENTIFIER, "s2", 11, 13, 1, 12, Nil(), "s2"},
	}
	lexer := runLexerTest(t, input, expected, true)
	require.Error(t, lexer.lastError)
	assert.Contains(t, lexer.lastError.Error(), "invalid character after unit or invalid unit")

	input2 := `1msident`
	expected2 := []expectedToken{
		{INT_LITERAL, "1", 0, 1, 1, 1, IntValue(1), ""},
		{IDENTIFIER, "msident", 1, 8, 1, 2, Nil(), "msident"},
	}
	runLexerTest(t, input2, expected2, true)
}

func TestLexer_DivisionAndMultilineComments(t *testing.T) {
	input := "a / b /* comment * test */ c /**/ d"
	expected := []expectedToken{
		{IDENTIFIER, "a", 0, 1, 1, 1, Nil(), "a"},
		// Since / is an opchar, it will be BINARY_OP
		{BINARY_OP, "/", 2, 3, 1, 3, Nil(), ""},
		{IDENTIFIER, "b", 4, 5, 1, 5, Nil(), "b"},
		{IDENTIFIER, "c", 27, 28, 1, 28, Nil(), "c"},
		{IDENTIFIER, "d", 34, 35, 1, 35, Nil(), "d"},
	}
	runLexerTest(t, input, expected, false)

	input2 := "/*unterminated"
	lexer2 := NewLexer(strings.NewReader(input2))
	lval2 := &SDLSymType{}
	lexer2.Lex(lval2)
	require.Error(t, lexer2.lastError)
	assert.Contains(t, lexer2.lastError.Error(), "unterminated block comment")

	input3 := "a * /* b */ c"
	expected3 := []expectedToken{
		{IDENTIFIER, "a", 0, 1, 1, 1, Nil(), "a"},
		// Since * is an opchar, it will be BINARY_OP
		{BINARY_OP, "*", 2, 3, 1, 3, Nil(), ""},
		{IDENTIFIER, "c", 12, 13, 1, 13, Nil(), "c"},
	}
	runLexerTest(t, input3, expected3, false)
}
