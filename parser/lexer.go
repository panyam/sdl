// decl/lexer.go
package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
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
	pos             int          // Current byte offset from the beginning of the input
	lastError       error

	// Position tracking for the current token
	tokenStartPos  int    // Byte offset where the current token started
	tokenStartLine int    // Line number (1-based) where the current token started
	tokenStartCol  int    // Column number (rune-based, 1-based) where the current token started
	tokenText      string // Raw text of the current token

	// Current line and column (rune-based) in the input
	line int
	col  int

	parseResult *File // Field to store the final AST root, set by the parser
}

// NewLexer creates a new lexer instance
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		reader: bufio.NewReader(r),
		pos:    0,
		line:   1,
		col:    1,
	}
}

// Error is called by the parser (or lexer itself) on an error.
func (l *Lexer) Error(s string) {
	l.lastError = fmt.Errorf("Error at Line %d, Col %d near '%s': %s", l.tokenStartLine, l.tokenStartCol, l.tokenText, s)
	fmt.Println(l.lastError) // For immediate feedback during development
}

// Pos returns the start byte offset of the most recently lexed token.
func (l *Lexer) Pos() int {
	return l.tokenStartPos
}

// End returns the end byte offset (current position) after lexing the most recent token.
func (l *Lexer) End() int {
	return l.pos
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
	l.pos += width
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
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

	oldPos := l.pos
	oldCol := l.col
	oldLine := l.line
	for i := range nchars {
		if consume {
			// udpate the position if we need to consume (actual consumption will come later when matched)
			l.updatePosition(l.lookaheadRunes[i], l.lookaheadWidths[i])
		}
		if l.lookaheadRunes[i] != rune(prefix[i]) {
			// restore old position
			l.pos = oldPos
			l.col = oldCol
			l.line = oldLine
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
	case "instance":
		return INSTANCE, text
	case "analyze":
		return ANALYZE, text
	case "expect":
		return EXPECT, text
	case "let":
		return LET, text
	case "if":
		return IF, text
	case "else":
		return ELSE, text
	case "distribute":
		return DISTRIBUTE, text
	case "default":
		return DEFAULT, text
	case "return":
		return RETURN, text
	case "delay":
		return DELAY, text
	case "wait":
		return WAIT, text
	case "go":
		return GO, text
	case "log":
		return LOG, text
	case "switch":
		return SWITCH, text
	case "case":
		return CASE, text
	case "enum":
		return ENUM, text
	case "import":
		return IMPORT, text
	case "options":
		return OPTIONS, text
	case "true":
		return BOOL_LITERAL, text
	case "false":
		return BOOL_LITERAL, text
	case "for":
		return FOR, text
	case "int":
		return INT, text
	case "float":
		return FLOAT, text
	case "bool":
		return BOOL, text
	case "string":
		return STRING, text
	case "duration":
		return DURATION, text
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
	l.read() // Consume opening '"'
	for {
		r, _ := l.read()
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
func (l *Lexer) Lex(lval *yySymType) int {
	if l.skipWhitespace() {
		return eof
	}

	l.tokenStartPos = l.pos
	l.tokenStartLine = l.line
	l.tokenStartCol = l.col
	l.tokenText = "" // Reset for current token

	r := l.peek()
	if r == eof {
		return eof
	}

	startPosSnapshot := l.tokenStartPos // Save start for setting NodeInfo on simple tokens

	if unicode.IsLetter(r) || r == '_' {
		tok, text := l.scanIdentifierOrKeyword()
		l.tokenText = text
		endPos := l.pos
		switch tok {
		case IDENTIFIER:
			lval.expr = newIdentifierExpr(text, startPosSnapshot, endPos)
		case BOOL_LITERAL:
			boolVal, _ := NewRuntimeValue(BoolType, text == "true")
			lval.expr = newLiteralExpr(boolVal, startPosSnapshot, endPos)
		case INT, FLOAT, BOOL, STRING, DURATION: // Type keywords
			lval.node = &TokenNode{NodeInfo: newNodeInfo(startPosSnapshot, endPos), Text: text}
		default: // Other keywords
			// Pass position via Node, if grammar expects Node for keywords
			lval.node = &TokenNode{NodeInfo: newNodeInfo(startPosSnapshot, endPos), Text: text}
			lval.sval = text // Keep sval too for compatibility if rules use it for ops
		}
		return tok
	}

	if unicode.IsDigit(r) || (r == '.' && unicode.IsDigit(l.peekN(1))) {
		numTok, numText := l.scanNumber()
		numEndPos := l.pos
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
		} else {
			// Invalid unit, so throw an error or just ignore?
			// For now, just leave unit as empty string
			// l.Error(fmt.Sprintf("Invalid duration unit: %s", unit))
		}
		if nconsume, ok := map[string]int{"ns": 2, "us": 2, "ms": 2, "s": 1}[unit]; ok && nconsume > 0 {
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
				durVal, _ := NewRuntimeValue(FloatType, dur)
				lval.expr = newLiteralExpr(durVal, startPosSnapshot, l.pos)
				return DURATION_LITERAL
			}
		}
		// Still consume the numeric part
		if numTok == INT_LITERAL {
			intVal, err := strconv.ParseInt(numText, 10, 64)
			if err != nil {
				l.Error(fmt.Sprintf("Invalid integer: %s", numText))
			}
			lval.expr = newLiteralExpr(IntValue(intVal), startPosSnapshot, numEndPos)
			return INT_LITERAL
		} else if numTok == FLOAT_LITERAL {
			floatVal, err := strconv.ParseFloat(numText, 64)
			if err != nil {
				l.Error(fmt.Sprintf("Invalid float: %s", numText))
			}
			lval.expr = newLiteralExpr(FloatValue(floatVal), startPosSnapshot, numEndPos)
			return FLOAT_LITERAL
		} else {
			l.Error(fmt.Sprintf("Invalid numeric literal: %s", numText))
		}
		return numTok
	}

	if r == '"' {
		_, content := l.scanString()
		l.tokenText = `"` + content + `"`
		// l.tokenText = content
		strVal, _ := NewRuntimeValue(StrType, content)
		lval.expr = newLiteralExpr(strVal, startPosSnapshot, l.pos)
		return STRING_LITERAL
	}

	// Operators and Punctuation - Default to single character token text
	l.tokenText = string(r)
	currentEndPos := l.pos // End pos for single char token

	// Handle multi-character operators
	switch r {
	case ';', '{', '}', '(', ')', ',', '.':
		l.read()
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return map[rune]int{
			';': SEMICOLON,
			'{': LBRACE,
			'}': RBRACE,
			'(': LPAREN,
			')': RPAREN,
			',': COMMA,
			'.': DOT,
		}[r]
	case ':':
		l.read()
		if l.peek() == '=' {
			l.read()
			l.tokenText = ":="
			currentEndPos = l.pos
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return LET_ASSIGN
		}
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return COLON
	case '=':
		l.read()
		if l.peek() == '=' {
			l.read()
			l.tokenText = "=="
			currentEndPos = l.pos
			lval.sval = "=="
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return EQ
		}
		if l.peek() == '>' {
			l.read()
			l.tokenText = "=>"
			currentEndPos = l.pos
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return ARROW
		}
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return ASSIGN
	case '|':
		if l.peekN(1) == '|' {
			l.read()
			l.read()
			l.tokenText = "||"
			currentEndPos = l.pos
			lval.sval = "||"
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return OR
		}
	case '&':
		if l.peekN(1) == '&' {
			l.read()
			l.read()
			l.tokenText = "&&"
			currentEndPos = l.pos
			lval.sval = "&&"
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return AND
		}
	case '!':
		l.read()
		if l.peek() == '=' {
			l.read()
			l.tokenText = "!="
			currentEndPos = l.pos
			lval.sval = "!="
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return NEQ
		}
		lval.sval = "!"
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return NOT
	case '<':
		l.read()
		if l.peek() == '=' {
			l.read()
			l.tokenText = "<="
			currentEndPos = l.pos
			lval.sval = "<="
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return LTE
		}
		lval.sval = "<"
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return LT
	case '>':
		l.read()
		if l.peek() == '=' {
			l.read()
			l.tokenText = ">="
			currentEndPos = l.pos
			lval.sval = ">="
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return GTE
		}
		lval.sval = ">"
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return GT
	case '+':
		l.read()
		if l.peek() == '=' {
			l.read()
			l.tokenText = "+="
			currentEndPos = l.pos
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return PLUS_ASSIGN
		}
		lval.sval = "+"
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return PLUS
	case '-':
		l.read()
		if l.peek() == '=' {
			l.read()
			l.tokenText = "-="
			currentEndPos = l.pos
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return MINUS_ASSIGN
		}
		lval.sval = "-"
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return MINUS
	case '*':
		l.read()
		if l.peek() == '=' {
			l.read()
			l.tokenText = "*="
			currentEndPos = l.pos
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return MUL_ASSIGN
		}
		lval.sval = "*"
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return MUL
	case '/': // Comments handled in skipWhitespace
		l.read()
		if l.peek() == '=' {
			l.read()
			l.tokenText = "/="
			currentEndPos = l.pos
			lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
			return DIV_ASSIGN
		}
		lval.sval = "/"
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return DIV
	case '%':
		l.read()
		lval.sval = "%"
		lval.node = &TokenNode{newNodeInfo(startPosSnapshot, currentEndPos), l.tokenText}
		return MOD
	}

	l.Error(fmt.Sprintf("unexpected character '%c'", r))
	return eof // Indicate an error that should halt parsing
}
