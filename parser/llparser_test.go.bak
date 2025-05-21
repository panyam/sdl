package parser

import (
	"strings"
	"testing"
)

func TestParseIdentifier(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      *IdentifierExpr
		expectError   bool
		errorContains string
	}{
		{"simple identifier", "myVar", newIdent("myVar"), false, ""},
		{"identifier with underscore", "_another_var", newIdent("_another_var"), false, ""},
		{"keyword as identifier (if allowed)", "let_is_var", newIdent("let_is_var"), false, ""}, // Depends on lexer
		{"empty input", "", nil, true, "expected IDENTIFIER"},                                   // PeekToken should see EOF
		{"number as identifier", "123", nil, true, "expected IDENTIFIER"},                       // Lexer provides INT_LITERAL
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (*IdentifierExpr, error) {
				return p.ParseIdentifier()
			})
			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actual)
			}
		})
	}
}

func TestParseLiteralExpression(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      Expr
		expectError   bool
		errorContains string
	}{
		{"integer literal", "12345", newIntLit(12345), false, ""},
		{"string literal", `"hello world"`, newStringLit("hello world"), false, ""},
		{"boolean true", "true", newBoolLit(true), false, ""},
		{"boolean false", "false", newBoolLit(false), false, ""},
		// Add FLOAT_LITERAL, DURATION_LITERAL if you have them
		{"unterminated string", `"abc`, nil, true, "literal"}, // Lexer error, propagated
		{"not a literal", "myIdent", nil, true, "expected a literal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (Expr, error) {
				return p.ParseLiteralExpr() // Or p.ParseExpression() if literals are only parsed via Primary
			})
			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actual)
			}
		})
	}
}

func TestParseBinaryExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Expr
	}{
		{"simple addition", "a + b", newBinaryExpr(newIdent("a"), "+", newIdent("b"))},
		{"simple subtraction", "c - 10", newBinaryExpr(newIdent("c"), "-", newIntLit(10))},
		{"multiplication and addition", "a * b + c",
			newBinaryExpr(
				newBinaryExpr(newIdent("a"), "*", newIdent("b")),
				"+",
				newIdent("c"),
			),
		},
		{"addition and multiplication", "a + b * c",
			newBinaryExpr(
				newIdent("a"),
				"+",
				newBinaryExpr(newIdent("b"), "*", newIdent("c")),
			),
		},
		{"parentheses", "(a + b) * c",
			newBinaryExpr(
				newBinaryExpr(newIdent("a"), "+", newIdent("b")), // This grouping needs to be handled by ParsePrimaryExpr
				"*",
				newIdent("c"),
			),
		},
		// Add tests for OR, AND, EQ, NEQ, LT, LTE, GT, GTE, DIV, MOD
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (Expr, error) {
				return p.ParseExpression()
			})
			assertError(t, tt.input, err, false, "") // Expect no error for these valid expressions
			assertNodeEqual(t, tt.input, tt.expected, actual)
		})
	}
}

func TestParseUnaryExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Expr
	}{
		{"unary minus", "-a", newUnaryExpr("-", newIdent("a"))},
		{"unary not", "!true", newUnaryExpr("!", newBoolLit(true))},
		{"double unary", "--1", newUnaryExpr("-", newUnaryExpr("-", newIntLit(1)))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (Expr, error) {
				return p.ParseExpression()
			})
			assertError(t, tt.input, err, false, "")
			assertNodeEqual(t, tt.input, tt.expected, actual)
		})
	}
}

func TestParseCallAndMemberAccessExpressions(t *testing.T) {
	// Helper for CallExpr
	newCallExpr := func(fn Expr, args ...Expr) *CallExpr {
		return &CallExpr{Function: fn, Args: args}
	}
	// Helper for MemberAccessExpr
	newMemberAccessExpr := func(recv Expr, memberName string) *MemberAccessExpr {
		return &MemberAccessExpr{Receiver: recv, Member: newIdent(memberName)}
	}

	tests := []struct {
		name     string
		input    string
		expected Expr
	}{
		{"simple call no args", "foo()", newCallExpr(newIdent("foo"))},
		{"call with one arg", "bar(1)", newCallExpr(newIdent("bar"), newIntLit(1))},
		{"call with multiple args", `plugh(x, "y", true)`, newCallExpr(newIdent("plugh"), newIdent("x"), newStringLit("y"), newBoolLit(true))},
		{"simple member access", "obj.member", newMemberAccessExpr(newIdent("obj"), "member")},
		{"chained member access", "a.b.c", newMemberAccessExpr(newMemberAccessExpr(newIdent("a"), "b"), "c")},
		{"member access then call", "obj.read()", newCallExpr(newMemberAccessExpr(newIdent("obj"), "read"))},
		{"call then member access", "getObj().field", newMemberAccessExpr(newCallExpr(newIdent("getObj")), "field")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (Expr, error) {
				return p.ParseExpression()
			})
			assertError(t, tt.input, err, false, "")
			assertNodeEqual(t, tt.input, tt.expected, actual)
		})
	}
}

func TestParseLetStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string // Input to ParseStmt should include the semicolon if that's how ExprStmt works
		expected      Stmt
		expectError   bool
		errorContains string
	}{
		{"simple let", "let x = 10", newLetStmt("x", newIntLit(10)), false, ""}, // Assuming ParseLetStmt does not consume semicolon
		{"let with expression", "let y = a + b", newLetStmt("y", newBinaryExpr(newIdent("a"), "+", newIdent("b"))), false, ""},
		{"missing equals", "let x 10", nil, true, "expected '='"},
		{"missing expression", "let x =", nil, true, "expected expression after '='"},
		// {"let with semicolon", "let x = 10;", ...} // Depends on whether ParseLetStmt or ParseStmt handles the semicolon
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (Stmt, error) {
				// If ParseLetStmt is called directly, it might not expect a semicolon.
				// If testing via ParseStmt, it would handle the semicolon for expression statements.
				return p.ParseLetStmt()
			})
			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actual)
			}
		})
	}
}

func TestParseBlockStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      *BlockStmt
		expectError   bool
		errorContains string
	}{
		{"empty block", "{}", newBlockStmt(), false, ""},
		{"block with one let stmt", "{let x = 1}", newBlockStmt(newLetStmt("x", newIntLit(1))), false, ""}, // No semicolon inside block for let in some languages
		{"block with expr stmt", "{ mycall(); }", newBlockStmt(newExprStmt(newCallExpr(newIdent("mycall")))), false, ""},
		{"block with multiple stmts", "{let x = 1; foo();}", // Assuming semicolon separates statements in a block OR is part of ExprStmt
			newBlockStmt(
				newLetStmt("x", newIntLit(1)),             // Assuming LetStmt doesn't include ;
				newExprStmt(newCallExpr(newIdent("foo"))), // Assuming ExprStmt includes ;
			), false, "",
		},
		{"unclosed block", "{ let x = 1", nil, true, "expected '}'"},
		{"block with syntax error", "{ let x = ; }", nil, true, "expected expression after '='"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (*BlockStmt, error) {
				return p.ParseBlockStmt()
			})
			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actual)
			}
		})
	}
}

// TestParseIfStatement, TestParseReturnStatement, etc. would follow a similar pattern.

// --- Top-Level Declaration Tests ---
// For these, we typically parse the whole "file" (fragment) and check the resulting File.Declarations.

func TestParseEnumDeclaration(t *testing.T) {
	newEnumDecl := func(name string, values ...string) *EnumDecl {
		idents := make([]*IdentifierExpr, len(values))
		for i, v := range values {
			idents[i] = newIdent(v)
		}
		return &EnumDecl{NameNode: newIdent(name), ValuesNode: idents}
	}

	tests := []struct {
		name          string
		input         string
		expectedDecls []Node // Expect a slice containing one EnumDecl
		expectError   bool
		errorContains string
	}{
		{"simple enum", "enum Color { RED, GREEN, BLUE }", []Node{newEnumDecl("Color", "RED", "GREEN", "BLUE")}, false, ""},
		{"empty enum", "enum Empty {}", []Node{newEnumDecl("Empty")}, false, ""},
		{"enum missing brace", "enum Bad { RED ", nil, true, "expected '}' to close enum"},
		{"enum bad value list", "enum Bad { RED, 123 }", nil, true, "expected identifier after comma in enum value list for 'Bad', found INT_LITERAL"},
		{"enum trailing comma", "enum Trailing { A, B, }", nil, true, "trailing comma not allowed"}, // If your parser disallows it
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileNode := &File{}
			lexer := NewLexer(strings.NewReader(tt.input))
			parser := NewLLParser(lexer)
			err := parser.Parse(fileNode) // Main Parse method

			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				if len(fileNode.Declarations) != len(tt.expectedDecls) {
					t.Fatalf("Input: %q\nExpected %d declarations, got %d", tt.input, len(tt.expectedDecls), len(fileNode.Declarations))
				}
				if len(tt.expectedDecls) > 0 { // Only compare if there's something to compare
					assertNodeEqual(t, tt.input, tt.expectedDecls[0], fileNode.Declarations[0])
				}
			}
		})
	}
}

// In parser_test.go

// TestParseParamDecl (tests the direct parsing of a 'param' item, usually within a component)
func TestParseParamDecl(t *testing.T) {
	tests := []struct {
		name          string
		input         string // e.g., "param count int" or "param rate float = 0.5"
		expected      *ParamDecl
		expectError   bool
		errorContains string
	}{
		{"simple param", "param count int",
			newParamDecl("count", newTypeName("int", true), nil),
			false, "",
		},
		{"param with custom type", "param user MyUserType",
			newParamDecl("user", newTypeName("MyUserType", false), nil),
			false, "",
		},
		{"param with default value", "param rate float = 0.5",
			newParamDecl("rate", newTypeName("float", true), &LiteralExpr{Value: FloatValue(0.5)}), // Assuming float literal
			false, "",
		},
		{"param with complex default", "param offset int = 10 + c",
			newParamDecl("offset", newTypeName("int", true), newBinaryExpr(newIntLit(10), "+", newIdent("c"))),
			false, "",
		},
		{"missing type", "param count = 5", nil, true, "expected type name for param"},
		{"missing name", "param int", nil, true, "expected identifier for param name"},
		{"missing default expr", "param x int =", nil, true, "expected expression for default value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (*ParamDecl, error) {
				// Need to create a temporary ComponentDecl for context, as ParseParamDecl expects to be called by ParseComponentDecl
				// However, for unit testing ParseParamDecl, we can call it directly.
				// The `out *ParamDecl` argument style might need a wrapper if the parse func signature is strict.
				// Let's assume it can be called directly for the test.
				paramDecl := &ParamDecl{}
				pErr := p.ParseParamDecl(paramDecl) // Call the method that takes `out *ParamDecl`
				return paramDecl, pErr
			})
			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actual)
			}
		})
	}
}

// TestParseUsesDecl (similar to ParamDecl, for component body items)
func TestParseUsesDecl(t *testing.T) {
	tests := []struct {
		name          string
		input         string // e.g., "uses myTimer Timer"
		expected      *UsesDecl
		expectError   bool
		errorContains string
	}{
		{"simple uses", "uses clock GlobalClock",
			newUsesDecl("clock", "GlobalClock"),
			false, "",
		},
		{"missing local name", "uses GlobalClock", nil, true, " expected identifier for component type after uses name"},
		{"missing component type", "uses clock", nil, true, "expected identifier for component type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (*UsesDecl, error) {
				usesDecl := &UsesDecl{}
				pErr := p.ParseUsesDecl(usesDecl)
				return usesDecl, pErr
			})
			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actual)
			}
		})
	}
}

// TestParseMethodDecl
func TestParseMethodDecl(t *testing.T) {
	tests := []struct {
		name            string
		input           string // e.g., "method process() { }"
		isSignatureOnly bool   // To test both native component methods and full methods
		expected        *MethodDecl
		expectError     bool
		errorContains   string
	}{
		{"simple method no params no return with body", "method run() { }", false,
			newMethodDecl("run", nil, newBlockStmt()),
			false, "",
		},
		{"method with params no return with body", "method calculate(a int, b float) { }", false,
			newMethodDecl("calculate", nil, newBlockStmt(),
				newParamDecl("a", newTypeName("int", true), nil), // Note: Method params don't have "param" keyword
				newParamDecl("b", newTypeName("float", true), nil),
			),
			false, "",
		},
		{"method with return type and body", "method getStatus() string { return \"OK\"; }", false,
			newMethodDecl("getStatus", newTypeName("string", true),
				newBlockStmt(&ReturnStmt{ReturnValue: newStringLit("OK")}), // Assuming ReturnStmt constructor
			),
			false, "",
		},
		{"method signature no params no return", "method init()", true,
			newMethodDecl("init", nil, nil), // No body for signature
			false, "",
		},
		{"method signature with params and return", "method query(id string) UserData", true,
			newMethodDecl("query", newTypeName("UserData", false), nil,
				newParamDecl("id", newTypeName("string", true), nil),
			),
			false, "",
		},
		{"missing method name", "method () {}", false, nil, true, "expected identifier for method name"},
		{"missing parens", "method foo {}", false, nil, true, "expected '(' after method name"},
		{"unclosed parens", "method foo( {}", false, nil, true, "Line: 1, Col: 13 - Error near '{' --- error parsing parameters for method 'foo'"},
		{"missing body for non-signature", "method foo() string", false, nil, true, "expected '{' for method body"},
		{"body for signature (should be error or ignored by parser logic)", "method foo() {}", true,
			newMethodDecl("foo", nil, nil), // Parser might parse body but it's not stored for signature
			false,
			`parser did not consume all input`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (*MethodDecl, error) {
				methodDecl := &MethodDecl{}
				pErr := p.ParseMethodDecl(methodDecl, tt.isSignatureOnly)
				return methodDecl, pErr
			})
			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actual)
			}
		})
	}
}

// TestParseComponentDecl (Top-level declaration)
func TestParseComponentDecl(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedDecls []Node // Expects a slice with one ComponentDecl
		expectError   bool
		errorContains string
	}{
		{"simple component", "component MyComp { }",
			[]Node{newComponentDecl("MyComp", false)},
			false, "",
		},
		{"native component", "native component MyNativeComp { }",
			[]Node{newComponentDecl("MyNativeComp", true)}, // isNative = true
			false, "",
		},
		{"component with param", "component WithParam { param count int; }",
			[]Node{newComponentDecl("WithParam", false,
				newParamDecl("count", newTypeName("int", true), nil),
			)},
			false, "",
		},
		{"component with uses", "component WithUses { uses clock Timer; }",
			[]Node{newComponentDecl("WithUses", false,
				newUsesDecl("clock", "Timer"),
			)},
			false, "",
		},
		{"component with method", "component WithMethod { method doit() { } }",
			[]Node{newComponentDecl("WithMethod", false,
				newMethodDecl("doit", nil, newBlockStmt()),
			)},
			false, "",
		},
		{"native component with method signature", "native component NativeAPI { method query(id int) string; }",
			[]Node{newComponentDecl("NativeAPI", true,
				newMethodDecl("query", newTypeName("string", true), nil,
					newParamDecl("id", newTypeName("int", true), nil),
				),
			)},
			false, "",
		},
		{"component missing name", "component { }", nil, true, "expected IDENTIFIER, found: LBRACE"},
		{"component missing open brace", "component MyComp ", nil, true, "expected LBRACE, found: EOF"},
		{"component missing close brace", "component MyComp { param x int;", nil, true, "unexpected eof reading component"}, // Error from your parser,
		{"component invalid body item", "component MyComp { unknown keyword; }", nil, true, "Error near 'unknown' --- Expected 'uses', 'param' or 'method', Found: unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileNode := &File{}
			lexer := NewLexer(strings.NewReader(tt.input))
			parser := NewLLParser(lexer)
			err := parser.Parse(fileNode)

			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				if len(fileNode.Declarations) != len(tt.expectedDecls) {
					t.Fatalf("Input: %q\nExpected %d declarations, got %d", tt.input, len(tt.expectedDecls), len(fileNode.Declarations))
				}
				if len(tt.expectedDecls) > 0 {
					assertNodeEqual(t, tt.input, tt.expectedDecls[0], fileNode.Declarations[0])
				}
			}
		})
	}
}

// TestParseSystemDecl (Top-level declaration)
func TestParseSystemDecl(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedDecls []Node // Expects a slice with one SystemDecl
		expectError   bool
		errorContains string
	}{
		{"simple system", "system MySys { }",
			[]Node{newSystemDecl("MySys")},
			false, "",
		},
		{"system with instance", "system MySys { use i1 CompA; }",
			[]Node{newSystemDecl("MySys",
				newInstanceDecl("i1", "CompA"),
			)},
			false, "",
		},
		{"system with instance and overrides", "system MySys { use i2 CompB = { p1 = 10; }; }",
			[]Node{newSystemDecl("MySys",
				newInstanceDecl("i2", "CompB", newAssignmentStmt("p1", newIntLit(10))),
			)},
			false, "",
		},
		/*
			{"system with options", "system MySys { options { log_level = 1; } }",
				[]Node{newSystemDecl("MySys",
					newOptionsDecl(newBlockStmt(newExprStmt(newBinaryExpr(newIdent("log_level"), "=", newIntLit(1))))), // Assuming assignment is an Expr for StmtList inside options
				)},
				false, "",
			},
		*/
		{"system with let", "system MySys { let global_count = 0; }",
			[]Node{newSystemDecl("MySys",
				newLetStmt("global_count", newIntLit(0)), // Assuming LetStmt implements SystemDeclBodyItem
			)},
			false, "",
		},
		{"system missing name", "system { }", nil, true, "expected identifier for system name"},
		{"system invalid body item", "system MySys { unknown; }", nil, true, "Line: 1, Col: 16 - Error near 'unknown' --- unexpected token 'IDENTIFIER' in system 'MySys' body. Expected 'use' or 'let'."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileNode := &File{}
			lexer := NewLexer(strings.NewReader(tt.input))
			parser := NewLLParser(lexer)
			err := parser.Parse(fileNode)

			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				if len(fileNode.Declarations) != len(tt.expectedDecls) {
					t.Fatalf("Input: %q\nExpected %d declarations, got %d", tt.input, len(tt.expectedDecls), len(fileNode.Declarations))
				}
				if len(tt.expectedDecls) > 0 {
					assertNodeEqual(t, tt.input, tt.expectedDecls[0], fileNode.Declarations[0])
				}
			}
		})
	}
}

// TestParseDistributeStmt
func TestParseDistributeStmt(t *testing.T) {
	tests := []struct {
		name          string
		input         string // Fragment for a single DistributeStmt
		expected      Stmt
		expectError   bool
		errorContains string
	}{
		{"simple distribute", "distribute { 10 => foo(); }",
			newDistributeStmt(nil, nil, newDistributeCase(newIntLit(10), newExprStmt(newCallExpr(newIdent("foo"))))),
			false, "",
		},
		{"distribute with total", "distribute total_prob { 0.5 => bar(); 0.5 => baz(); }",
			newDistributeStmt(newIdent("total_prob"), nil,
				newDistributeCase(&LiteralExpr{Value: FloatValue(0.5)}, newExprStmt(newCallExpr(newIdent("bar")))),
				newDistributeCase(&LiteralExpr{Value: FloatValue(0.5)}, newExprStmt(newCallExpr(newIdent("baz")))),
			),
			false, "",
		},
		{"distribute with default", "distribute { 1 => a(); default => b(); }",
			newDistributeStmt(nil, newDefaultCase(newExprStmt(newCallExpr(newIdent("b")))),
				newDistributeCase(newIntLit(1), newExprStmt(newCallExpr(newIdent("a")))),
			),
			false, "",
		},
		{"empty distribute", "distribute { }", newDistributeStmt(nil, nil), false, ""},
		{"missing arrow", "distribute { 10 foo(); }", nil, true, "expected '=>'"},
		{"missing body after arrow", "distribute { 10 => }", nil, true, "Line: 1, Col: 20 - Error near '}' --- error parsing DISTRIBUTE case body"}, // Or specific error if ParseStmt fails early
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (Stmt, error) {
				return p.ParseDistributeStmt()
			})
			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actual)
			}
		})
	}
}

// TestParseGoStmt
/*
func TestParseGoStmt(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      Stmt
		expectError   bool // Error from the parsing function itself, not the post-check in ParseStmt
		errorContains string
	}{
		{"go block", "go { foo(); }",
			newGoStmt(nil, newBlockStmt(newExprStmt(newCallExpr(newIdent("foo")))), false),
			false, "",
		},
		{"go var assign block", "go h = { bar(); }",
			newGoStmt(newIdent("h"), newBlockStmt(newExprStmt(newCallExpr(newIdent("bar")))), false),
			false, "",
		},
		{"go var assign expr (error case)", "go h = baz();", // Semicolon is handled by ParseStmt if it's an ExprStmt
			newGoStmt(newIdent("h"), newExprStmt(newCallExpr(newIdent("baz"))), true), // IsExprAssignment true
			false, "", // ParseGoStmt itself succeeds, ParseStmt dispatcher flags the error
		},
		{"missing block or var", "go", nil, true, "expected identifier or '{' after GO"},
		{"go var assign missing expr", "go h =", nil, true, "expected block or expression after '='"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For `go var = baz();`, ParseGoStmt will parse it, but the overall ParseStmt will raise the contextual error.
			// Here, we test ParseGoStmt directly.
			actualGoStmt, err := parseFragment(t, tt.input, func(p *LLParser) (Stmt, error) {
				return p.ParseGoStmt()
			})

			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actualGoStmt)

				// Additional check for the error case handled by the dispatcher
				if gs, ok := tt.expected.(*GoStmt); ok && gs.IsExprAssignment {
					// Simulate calling through the main ParseStmt to catch the dispatcher's error
					file := &File{}
					lexer := NewLexer(strings.NewReader(tt.input + ";")) // Add semicolon for ExprStmt if applicable
					parser := NewLLParser(lexer)
					_ = parser.Parse(file) // We expect this to try and parse `go h = baz();` as a statement

					// This part is tricky because the error comes from ParseStmt, not ParseGoStmt.
					// We need a way to inspect the error from the top-level parse if it's the specific 'go expr assign' error.
					// For simplicity in this direct test of ParseGoStmt, we just check IsExprAssignment.
					// A more comprehensive test would call ParseStmt and check its returned error.
					if actualGs, ok := actualGoStmt.(*GoStmt); !ok || !actualGs.IsExprAssignment {
						t.Errorf("Input %q: Expected GoStmt with IsExprAssignment=true", tt.input)
					}
				}
			}
		})
	}
}
*/

// TestParseDelayStmt
func TestParseDelayStmt(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      Stmt
		expectError   bool
		errorContains string
	}{
		{"delay with literal", "delay 100ms", // Assuming 100ms is a DURATION_LITERAL
			newDelayStmt(&LiteralExpr{Value: FloatValue(100)}),
			false, "",
		},
		{"delay with variable", "delay my_timeout",
			newDelayStmt(newIdent("my_timeout")),
			false, "",
		},
		{"delay with expression", "delay base_delay + 10s",
			newDelayStmt(newBinaryExpr(newIdent("base_delay"), "+", &LiteralExpr{Value: FloatValue(10000)})),
			false, "",
		},
		{"missing duration", "delay", nil, true, "expected expression for DELAY duration"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (Stmt, error) {
				return p.ParseDelayStmt()
			})
			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actual)
			}
		})
	}
}

// TestParseWaitStmt
func TestParseWaitStmt(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      Stmt
		expectError   bool
		errorContains string
	}{
		{"wait one ident", "wait handle1",
			newWaitStmt(newIdent("handle1")),
			false, "",
		},
		{"wait multiple idents", "wait h1, h2, h3_thing",
			newWaitStmt(newIdent("h1"), newIdent("h2"), newIdent("h3_thing")),
			false, "",
		},
		{"missing idents", "wait", nil, true, "expected at least one identifier after WAIT"},
		{"trailing comma", "wait h1,", nil, true, "expected identifier after comma"},  // If your parser disallows it
		{"wait non-ident", "wait 123", nil, true, "expected at least one identifier"}, // Or "expected identifier" from ParseIdentifier
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFragment(t, tt.input, func(p *LLParser) (Stmt, error) {
				return p.ParseWaitStmt()
			})
			assertError(t, tt.input, err, tt.expectError, tt.errorContains)
			if !tt.expectError {
				assertNodeEqual(t, tt.input, tt.expected, actual)
			}
		})
	}
}

// TestEOFHandling
func TestEOFHandling(t *testing.T) {
	tests := []struct {
		name          string
		input         string // Incomplete input
		parseFunc     func(p *LLParser) (Node, error)
		errorContains string // Expected part of the error message
	}{
		{"unclosed component", "component MyComp {",
			func(p *LLParser) (Node, error) {
				comp := &ComponentDecl{}
				err := p.ParseComponentDecl(comp)
				return comp, err
			}, "unexpected eof reading component",
		},
		{"unclosed system", "system MySys { use i1 C1",
			func(p *LLParser) (Node, error) {
				sys := &SystemDecl{}
				err := p.ParseSystemDecl(sys)
				return sys, err
			}, "expected '}' to close system",
		},
		{"unclosed enum", "enum MyEnum { A, B",
			func(p *LLParser) (Node, error) {
				enum := &EnumDecl{}
				err := p.ParseEnumDecl(enum)
				return enum, err
			}, "expected '}' to close enum",
		},
		{"unclosed block stmt", "{ let x = 10",
			func(p *LLParser) (Node, error) { return p.ParseBlockStmt() },
			"expected '}' to close block",
		},
		{"expression expecting more", "a + ",
			func(p *LLParser) (Node, error) { return p.ParseExpression() },
			"expected expression after operator 'PLUS_OP', found EOF", // Error from binary expression parsing
		},
		{"let stmt expecting expr", "let x = ",
			func(p *LLParser) (Node, error) { return p.ParseLetStmt() },
			"expected expression after '='",
		},
		{"unclosed distribute", "distribute { 10 -> foo()",
			func(p *LLParser) (Node, error) { return p.ParseDistributeStmt() },
			"expected expression for DISTRIBUTE case condition, found GT_OP",
		},
		/*TODO
		{"unclosed go block", "go {",
			func(p *LLParser) (Node, error) { return p.ParseGoStmt() },
			"expected '}' to close block", // Error from ParseBlockStmt called by ParseGoStmt
		},
		*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFragment(t, tt.input, tt.parseFunc)
			assertError(t, tt.input, err, true, tt.errorContains)
		})
	}
}

// TODO: Add tests for:
// - ParseOptionsDecl
// - ParseImportDecl (if you implement it)
// - Expression statements (e.g. `myFunc();`) via ParseStmt
// - Error recovery (if you implement any beyond simple error reporting)

// Add a placeholder for NewParseError if you need it for the unconsumed input check
