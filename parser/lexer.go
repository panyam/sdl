package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Ensure EOF is defined
const eof = 0

// Lexer structure
type Lexer struct {
	lookaheadRunes  []rune
	lookaheadWidths []int
	reader          *bufio.Reader
	buf             bytes.Buffer // Temporary buffer for scanned text
	lastError       error

	// Precedecences, associativity of operators
	Precedences map[int]PrecedenceInfo

	// Position tracking for the current token
	tokenStart    Location // Byte offset where the current token started
	tokenText     string   // Raw text of the current token
	lastTokenCode int      // <-- Added: Store the last token returned by Lex

	// Current line and column (rune-based) in the input
	location Location

	parseResult *FileDecl // Field to store the final AST root, set by the parser
}

// NewLexer creates a New lexer instance
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		reader: bufio.NewReader(r),
		location: Location{
			Pos:  0,
			Line: 1,
			Col:  1,
		},
	}
}

// Error is called by the parser (or lexer itself) on an error.
func (l *Lexer) Error(s string) {
	if l.Text() != "" {
		l.lastError = fmt.Errorf("Line: %d, Col: %d - Error near '%s' --- %s", l.tokenStart.Line, l.tokenStart.Col, l.Text(), s)
	} else {
		l.lastError = fmt.Errorf("Line: %d, Col: %d - %s", l.tokenStart.Line, l.tokenStart.Col, s)
	}
	// fmt.Println(s) // For immediate feedback during development
}

// Pos returns the start byte offset of the most recently lexed token.
func (l *Lexer) Pos() Location {
	return l.tokenStart
}

func (l *Lexer) End() Location {
	return l.location
}

// LastToken returns the code of the last token successfully lexed
func (l *Lexer) LastToken() int {
	return l.lastTokenCode
}

// Text returns the raw text of the most recently lexed token.
func (l *Lexer) Text() string {
	return l.tokenText
}

// --- Rune Reading Helpers (with line/col tracking) ---
func (l *Lexer) read() (r rune, width int) {
	if l.peek() == eof {
		return eof, 0
	}
	r, width = l.lookaheadRunes[0], l.lookaheadWidths[0]
	l.lookaheadRunes, l.lookaheadWidths = l.lookaheadRunes[1:], l.lookaheadWidths[1:]
	l.updatePosition(r, width)
	return r, width
}

func (l *Lexer) updatePosition(r rune, width int) {
	l.location.Pos += width
	if r == '\n' {
		l.location.Line++
		l.location.Col = 1
	} else {
		l.location.Col++
	}
}

func (l *Lexer) peekN(nthchar int) rune {
	if nthchar <= len(l.lookaheadRunes) {
		l.ensureLookAhead(nthchar + 1)
	}
	if nthchar >= len(l.lookaheadRunes) {
		return eof
	}
	return l.lookaheadRunes[nthchar]
}

func (l *Lexer) peek() rune {
	if len(l.lookaheadRunes) == 0 {
		r, width, err := l.reader.ReadRune()
		if err != nil {
			return eof
		}
		l.lookaheadRunes = []rune{r}
		l.lookaheadWidths = []int{width}
	}
	return l.lookaheadRunes[0]
}

func (l *Lexer) ensureLookAhead(numchars int) (numread int) {
	for len(l.lookaheadRunes) <= numchars {
		r, width, err := l.reader.ReadRune()
		if err != nil {
			return len(l.lookaheadRunes)
		}
		l.lookaheadRunes = append(l.lookaheadRunes, r)
		l.lookaheadWidths = append(l.lookaheadWidths, width)
		numread += width
	}
	return len(l.lookaheadRunes)
}

func (l *Lexer) hasPrefix(prefix string, consume bool) bool {
	nchars := len(prefix)
	nlookahead := l.ensureLookAhead(nchars)
	if nchars > nlookahead {
		return false
	}

	oldPos := l.location.Pos
	oldCol := l.location.Col
	oldLine := l.location.Line
	for i := range nchars {
		if consume {
			// udpate the position if we need to consume (actual consumption will come later when matched)
			l.updatePosition(l.lookaheadRunes[i], l.lookaheadWidths[i])
		}
		if l.lookaheadRunes[i] != rune(prefix[i]) {
			// restore old position
			l.location.Pos = oldPos
			l.location.Col = oldCol
			l.location.Line = oldLine
			return false
		}
	}

	if consume {
		l.lookaheadRunes = l.lookaheadRunes[nchars:]
		l.lookaheadWidths = l.lookaheadWidths[nchars:]
	}
	return true
}

func (l *Lexer) readTill(stop rune, skip bool) (foundeof bool) {
	for {
		r := l.peek()
		if r == eof {
			return true
		}
		if r == stop {
			if skip {
				// Found the stop character, but skip it
				l.read()
			}
			return false
		}
		// Not the stop character, so consume it
		l.read()
	}
}

// --- Scanning Functions ---
func (l *Lexer) skipWhitespace() bool {
	for {
		firstChar := l.peek()
		if firstChar == eof {
			return true
		}
		if unicode.IsSpace(firstChar) {
			// consume it
			l.read()
		} else if l.hasPrefix("//", true) {
			// Skip whitespace and comments
			l.readTill('\n', true)
		} else if l.hasPrefix("/*", true) {
			// Skip whitespace and comments
			expectSlash := false
			for {
				nextCh, _ := l.read()
				if nextCh == eof {
					l.Error("unterminated block comment")
					return true
				}
				if expectSlash {
					if nextCh == '/' {
						break // done with comment
					} else {
						expectSlash = false // not a slash, so reset
					}
				}
				if nextCh == '*' {
					expectSlash = true
				}
			}
		} else {
			// Not whitespace or comment, so stop
			return false
		}
	}
}

func (l *Lexer) scanIdentifierOrKeyword() (tok int, text string) {
	l.buf.Reset()
	for r := l.peek(); r != eof && (unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'); r = l.peek() {
		l.read()
		l.buf.WriteRune(r)
	}
	text = l.buf.String()
	switch text {
	case "native":
		return NATIVE, text
	case "use":
		return USE, text
	case "component":
		return COMPONENT, text
	case "system":
		return SYSTEM, text
	case "param":
		return PARAM, text
	case "uses":
		return USES, text
	case "method":
		return METHOD, text
	case "analyze":
		return ANALYZE, text
	case "expect":
		return EXPECT, text
	case "let":
		return LET, text
	case "if":
		return IF, text
	// case "tup": return TUP, text
	case "else":
		return ELSE, text
	case "sample":
		return SAMPLE, text
	case "dist":
		return DISTRIBUTE, text
	case "default":
		return DEFAULT, text
	case "return":
		return RETURN, text
	case "wait":
		return WAIT, text
	case "go":
		return GO, text
	case "gobatch":
		return GOBATCH, text
	case "aggregator":
		return AGGREGATOR, text
	case "using":
		return USING, text
	// case "log": return LOG, text
	case "switch":
		return SWITCH, text
	case "case":
		return CASE, text
	case "enum":
		return ENUM, text
	case "import":
		return IMPORT, text
	case "from":
		return FROM, text
	case "as":
		return AS, text
	case "options":
		return OPTIONS, text
	case "true":
		return BOOL_LITERAL, text
	case "false":
		return BOOL_LITERAL, text
	case "not":
		return UNARY_OP, text
	case "for":
		return FOR, text
	default:
		return IDENTIFIER, text
	}
}

func (l *Lexer) scanNumber() (tok int, text string) {
	l.buf.Reset()
	hasDecimal := false
	for r := l.peek(); r != eof; r = l.peek() {
		if unicode.IsDigit(r) {
			l.read() // consume it
			l.buf.WriteRune(r)
		} else if r == '.' && !hasDecimal {
			if !unicode.IsDigit(l.peekN(1)) {
				break
			}
			l.read() // consume it
			hasDecimal = true
			l.buf.WriteRune(r)
		} else {
			break
		}
	}
	text = l.buf.String()
	if hasDecimal {
		return FLOAT_LITERAL, text
	}
	return INT_LITERAL, text
}

func (l *Lexer) scanString() (tok int, content string) {
	l.buf.Reset()
	l.tokenText = "\""
	l.read() // Consume opening '"'
	for {
		r, _ := l.read()
		l.tokenText += string(r)
		if r == eof {
			l.Error("unterminated string literal")
			return eof, ""
		}
		if r == '"' {
			break
		}
		if r == '\\' {
			esc, _ := l.read()
			if esc == eof {
				l.Error("unterminated string literal after escape")
				return eof, ""
			}
			l.tokenText += string(esc)
			switch esc {
			case 'n':
				l.buf.WriteRune('\n')
			case 't':
				l.buf.WriteRune('\t')
			case '\\':
				l.buf.WriteRune('\\')
			case '"':
				l.buf.WriteRune('"')
			default:
				l.Error(fmt.Sprintf("invalid escape sequence \\%c", esc))
				l.buf.WriteRune(esc)
			}
		} else {
			l.buf.WriteRune(r)
		}
	}
	return STRING_LITERAL, l.buf.String()
}

// Lex is the main lexing function called by the parser.
func (l *Lexer) Lex(lval *SDLSymType) int {
	if l.skipWhitespace() {
		l.lastTokenCode = eof
		return eof
	}

	l.tokenStart.Pos = l.location.Pos
	l.tokenStart.Line = l.location.Line
	l.tokenStart.Col = l.location.Col
	l.tokenText = "" // Reset for current token

	r := l.peek()
	if r == eof {
		l.lastTokenCode = eof
		return eof
	}

	startPosSnapshot := l.tokenStart

	if unicode.IsLetter(r) || r == '_' {
		tok, text := l.scanIdentifierOrKeyword()
		l.tokenText = text
		endPos := l.location

		switch tok {
		case IDENTIFIER:
			lval.ident = NewIdentExpr(text, startPosSnapshot, endPos)
		case BOOL_LITERAL:
			boolVal, _ := NewValue(BoolType, text == "true")
			lval.expr = NewLiteralExpr(boolVal, startPosSnapshot, endPos)
		case INT, FLOAT, BOOL, STRING, DURATION: // Type keywords
			lval.node = NewTokenNode(startPosSnapshot, endPos, text)
		default: // Other keywords
			// Pass position via Node, if grammar expects Node for keywords
			lval.node = NewTokenNode(startPosSnapshot, endPos, text)
			lval.sval = text // Keep sval too for compatibility if rules use it for ops
		}
		return tok
	}

	if unicode.IsDigit(r) || (r == '.' && unicode.IsDigit(l.peekN(1))) {
		numTok, numText := l.scanNumber()
		numEndPos := l.location
		l.tokenText = numText

		unit := ""
		if l.hasPrefix("ms", false) {
			unit = "ms"
		} else if l.hasPrefix("s", false) {
			unit = "s"
		} else if l.hasPrefix("us", false) {
			unit = "us"
		} else if l.hasPrefix("ns", false) {
			unit = "ns"
		} else if l.hasPrefix("min", false) {
			unit = "min"
		} else if l.hasPrefix("hr", false) {
			unit = "hr"
		} else {
			// Invalid unit, so throw an error or just ignore?
			// For now, just leave unit as empty string
			// l.Error(fmt.Sprintf("Invalid duration unit: %s", unit))
		}
		if nconsume, ok := map[string]int{"ns": 2, "us": 2, "ms": 2, "s": 1, "hr": 2, "min": 3}[unit]; ok && nconsume > 0 {
			// Make sure we have a non digit and non ident char *after* the unit
			peekedAfter := l.peekN(nconsume)
			if peekedAfter == '_' || unicode.IsDigit(peekedAfter) || unicode.IsLetter(peekedAfter) {
				l.Error("invalid character after unit or invalid unit")
			} else {
				for range nconsume {
					l.read()
				}
				l.tokenText += unit
				dur := parseDuration(numText, unit)
				durVal, _ := NewValue(FloatType, dur)
				lval.expr = NewLiteralExpr(durVal, startPosSnapshot, l.location)
				return DURATION_LITERAL
			}
		}
		// Still consume the numeric part
		if numTok == INT_LITERAL {
			intVal, err := strconv.ParseInt(numText, 10, 64)
			if err != nil {
				l.Error(fmt.Sprintf("Invalid integer: %s", numText))
			}
			lval.expr = NewLiteralExpr(IntValue(intVal), startPosSnapshot, numEndPos)
			return INT_LITERAL
		} else if numTok == FLOAT_LITERAL {
			floatVal, err := strconv.ParseFloat(numText, 64)
			if err != nil {
				l.Error(fmt.Sprintf("Invalid float: %s", numText))
			}
			lval.expr = NewLiteralExpr(FloatValue(floatVal), startPosSnapshot, numEndPos)
			return FLOAT_LITERAL
		} else {
			l.Error(fmt.Sprintf("Invalid numeric literal: %s", numText))
		}
		return numTok
	}

	if r == '"' {
		_, content := l.scanString()
		// l.tokenText = content
		strVal, _ := NewValue(StrType, content)
		lval.expr = NewLiteralExpr(strVal, startPosSnapshot, l.location)
		return STRING_LITERAL
	}

	// Operators and Punctuation - Default to single character token text
	l.tokenText = string(r)
	currentEndPos := l.location

	// Handle multi-character operators
	switch r {
	case ';', '{', '}', '(', ')', ',', '.', '[', ']':
		l.read()
		lval.node = NewTokenNode(startPosSnapshot, currentEndPos, l.tokenText)
		return map[rune]int{
			';': SEMICOLON,
			'{': LBRACE,
			'}': RBRACE,
			'[': LSQUARE,
			']': RSQUARE,
			'(': LPAREN,
			')': RPAREN,
			',': COMMA,
			'.': DOT,
		}[r]
	default:
	}

	// Collect all operator characters
	opchars := "<>&^%$#@!*~=/|:+-"
	var out []rune
	for l.peek() > 0 && strings.IndexRune(opchars, l.peek()) >= 0 {
		out = append(out, l.peek())
		l.read()
	}
	optoken := string(out)
	if optoken != "" {
		lval.node = NewTokenNode(startPosSnapshot, currentEndPos, string(out))
		lval.sval = optoken
		l.tokenText = optoken
		tokenCode := BINARY_OP
		if optoken == "-" {
			tokenCode = MINUS
			// } else if optoken == "==" { tokenCode = EQ
		} else if optoken == "=" {
			tokenCode = ASSIGN
		} else if optoken == ":=" {
			tokenCode = LET_ASSIGN
		} else if optoken == "=>" {
			tokenCode = ARROW
		} else if optoken == ":" {
			tokenCode = COLON
		}
		return tokenCode
	}
	l.Error(fmt.Sprintf("unexpected character '%c'", r))
	return eof // Indicate an error that should halt parsing
}

// TokenString helper needs yyToknames from generated sdl.go
// For testing, we can define a minimal version or mock it.
// Minimal version for testing:
var testTokenNames = map[int]string{
	eof:              "EOF",
	IDENTIFIER:       "IDENTIFIER",
	INT_LITERAL:      "INT_LITERAL",
	FLOAT_LITERAL:    "FLOAT_LITERAL",
	STRING_LITERAL:   "STRING_LITERAL",
	BOOL_LITERAL:     "BOOL_LITERAL",
	DURATION_LITERAL: "DURATION_LITERAL",
	NATIVE:           "NATIVE",
	USE:              "USE",
	COMPONENT:        "COMPONENT",
	SYSTEM:           "SYSTEM",
	PARAM:            "PARAM",
	USES:             "USES",
	METHOD:           "METHOD",
	ANALYZE:          "ANALYZE",
	EXPECT:           "EXPECT",
	LET:              "LET",
	IF:               "IF",
	ELSE:             "ELSE",
	DISTRIBUTE:       "DISTRIBUTE",
	DEFAULT:          "DEFAULT",
	RETURN:           "RETURN",
	WAIT:             "WAIT",
	GO:               "GO",
	GOBATCH:          "GOBATCH",
	AGGREGATOR:       "AGGREGATOR",
	USING:            "USING",
	// LOG:              "LOG",
	SWITCH:     "SWITCH",
	CASE:       "CASE",
	ENUM:       "ENUM",
	IMPORT:     "IMPORT",
	FROM:       "FROM",
	AS:         "AS",
	OPTIONS:    "OPTIONS",
	FOR:        "FOR",
	INT:        "int", // Keyword for type
	FLOAT:      "float",
	BOOL:       "bool",
	STRING:     "string",
	DURATION:   "Duration",
	ASSIGN:     "ASSIGN",
	COLON:      "COLON",
	SEMICOLON:  "SEMICOLON",
	LBRACE:     "LBRACE",
	RBRACE:     "RBRACE",
	LSQUARE:    "LSQUARE",
	RSQUARE:    "RSQUARE",
	LPAREN:     "LPAREN",
	RPAREN:     "RPAREN",
	COMMA:      "COMMA",
	DOT:        "DOT",
	ARROW:      "ARROW",
	LET_ASSIGN: "LET_ASSIGN",
	BINARY_OP:  "BINARY_OP",
	MINUS:      "MINUS",
}

func TokenString(tok int) string {
	if name, ok := testTokenNames[tok]; ok {
		return name
	}
	// For single char punct tokens that are not in the map
	if tok > 0 && tok < 256 && !unicode.IsLetter(rune(tok)) && !unicode.IsDigit(rune(tok)) {
		return fmt.Sprintf("'%c'", rune(tok))
	}
	return fmt.Sprintf("TOKEN<%d>", tok)
}
