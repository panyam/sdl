package parser

import (
	"fmt"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

type LLParser struct {
	lexer            *Lexer
	peekedTokenValue *SDLSymType
	peekedToken      int

	PanicOnError bool
	Errors       []error
	// file *FileDecl // You might want to store the top-level AST node being built here.
}

func NewLLParser(lexer *Lexer) *LLParser {
	return &LLParser{lexer: lexer}
}
func (p *LLParser) Parse(file *FileDecl) (err error) {
	// p.file = file // Store the file being parsed
	for {
		peekedToken := p.PeekToken()
		if peekedToken == eof {
			return nil
		}
		if peekedToken == COMPONENT || peekedToken == NATIVE {
			node := &ComponentDecl{}
			if err = p.ParseComponentDecl(node); err != nil {
				return err
			}
			file.Declarations = append(file.Declarations, node)
		} else if peekedToken == IMPORT {
			imports, err := p.ParseImportDecl()
			if err != nil {
				return err
			}
			for _, imp := range imports {
				file.Declarations = append(file.Declarations, imp)
			}
		} else if peekedToken == ENUM {
			node := &EnumDecl{}
			if err = p.ParseEnumDecl(node); err != nil {
				return err
			}
			file.Declarations = append(file.Declarations, node)
		} else if peekedToken == SYSTEM {
			node := &SystemDecl{}
			if err = p.ParseSystemDecl(node); err != nil {
				return err
			}
			file.Declarations = append(file.Declarations, node)
		} else if peekedToken == SEMICOLON {
			p.Advance()
			continue
		} else {
			// No more top-level declarations or unexpected token
			return p.Errorf("expected 'component', 'enum' or 'system', found: %s (%s)",
				TokenString(p.PeekToken()), p.lexer.Text())
		}
	}
}

func (p *LLParser) Errorf(format string, args ...any) error {
	s := fmt.Sprintf(format, args...)
	p.lexer.Error(s)
	if p.PanicOnError {
		panic(p.lexer.lastError)
	}
	return p.lexer.lastError
}

func (p *LLParser) Advance() int {
	p.PeekToken()
	last := p.peekedToken
	p.peekedTokenValue = nil
	p.peekedToken = -1
	return last
}

func (p *LLParser) PeekToken() int {
	if p.peekedTokenValue == nil {
		p.peekedTokenValue = &SDLSymType{}
		p.peekedToken = p.lexer.Lex(p.peekedTokenValue)
	}
	return p.peekedToken
}

// Expect checks if the current peeked token is one of the expected tokens.
// It does NOT advance.
func (p *LLParser) Expect(tokensIn ...int) (foundToken int, err error) {
	peekedToken := p.PeekToken()
	for _, tok := range tokensIn {
		if tok == peekedToken {
			return tok, nil
		}
	}
	expectedStrings := gfn.Map(tokensIn, func(t int) string { return TokenString(t) })
	errMsg := "expected"
	if len(tokensIn) == 1 {
		errMsg = fmt.Sprintf("expected %s, found: %s", TokenString(tokensIn[0]), TokenString(peekedToken))
	} else {
		errMsg = fmt.Sprintf("expected one of: [%s], found: %s", strings.Join(expectedStrings, ", "), TokenString(peekedToken))
	}
	if p.lexer.Text() != "" {
		errMsg = fmt.Sprintf("%s (%s)", errMsg, p.lexer.Text())
	}
	return -1, p.Errorf(errMsg, "")
}

// AdvanceIf expects one of the given tokens and advances if found.
// Returns the matched token type and its semantic value.
func (p *LLParser) AdvanceIf(tokensIn ...int) (foundToken int, tokenValue *SDLSymType, err error) {
	if _, err = p.Expect(tokensIn...); err != nil {
		return -1, nil, err
	}
	// At this point, p.PeekToken() is one of tokensIn.
	foundToken = p.peekedToken
	tokenValue = p.peekedTokenValue
	p.Advance()
	return
}

// Extract a single identifier
// Ensure it populates NodeInfo correctly using p.peekedTokenValue before advancing.
func (p *LLParser) ParseIdentifier() (out *IdentifierExpr, err error) {
	if _, err = p.Expect(IDENTIFIER); err != nil {
		return nil, err
	}
	// Identifier token is currently peeked.
	tokenVal := p.peekedTokenValue
	p.Advance() // Consume IDENTIFIER

	// Assuming lexer populates tokenVal.ident for IDENTIFIER tokens.
	// If not, construct IdentifierExpr here:
	if tokenVal.ident == nil { // Fallback if lexer doesn't directly create *IdentifierExpr
		out = &IdentifierExpr{
			ExprBase: ExprBase{NodeInfo: newNodeInfoFromToken(tokenVal)},
			Name:     tokenVal.sval, // Or tokenVal.node.(TokenNode).String() if sval is not for idents
		}
	} else {
		out = tokenVal.ident
	}
	return out, nil
}

func (p *LLParser) ParseComponentDecl(out *ComponentDecl) (err error) {
	if _, err = p.Expect(COMPONENT, NATIVE); err != nil {
		return
	}

	peeked := p.PeekToken()
	out.NodeInfo = newNodeInfo(p.peekedTokenValue.node.Pos(), p.peekedTokenValue.node.End())
	isNative := peeked == NATIVE
	if isNative {
		TokenString(p.Advance())
		TokenString(p.Advance())
	} else {
		TokenString(p.Advance())
	}

	if out.NameNode, err = p.ParseIdentifier(); err != nil {
		return
	}
	if _, _, err = p.AdvanceIf(LBRACE); err != nil {
		return
	}

	// Parse body items
	for {
		peekedToken := p.PeekToken()
		if peekedToken == eof {
			err = fmt.Errorf("unexpected eof reading component: %s", out.NameNode.Name)
			return
		}
		if peekedToken == SEMICOLON {
			p.Advance()
			continue
		}
		if peekedToken == PARAM {
			node := &ParamDecl{}
			if err = p.ParseParamDecl(node); err != nil {
				return err
			}
			out.Body = append(out.Body, node)
		} else if peekedToken == METHOD {
			node := &MethodDecl{}
			if err = p.ParseMethodDecl(node, isNative); err != nil {
				return err
			}
			out.Body = append(out.Body, node)
		} else if peekedToken == USES {
			node := &UsesDecl{}
			if err = p.ParseUsesDecl(node); err != nil {
				return err
			}
			out.Body = append(out.Body, node)
		} else if peekedToken == RBRACE {
			out.NodeInfo.StopPos = p.peekedTokenValue.node.End()
			p.Advance()
			return
		} else {
			return p.Errorf("Expected 'uses', 'param' or 'method', Found: %s", p.lexer.Text())
		}
	}
}

// ParseImportDecl parses an enumeration declaration.
// Grammar:
//
//	IMPORT from STRING
//	IMPORT ImportedItemList STRING
//
// ImportedItem := IDENTIFIER
// ImportedItem := IDENTIFIER ( "as IDENTIFIER )
// ImportedItemList := ImportedItem (COMMA ImportedItemList) *
func (p *LLParser) ParseImportDecl() (out []*ImportDecl, err error) {
	if _, _, err = p.AdvanceIf(IMPORT); err != nil {
		return nil, err
	}

	for {
		if p.PeekToken() == eof {
			return nil, p.Errorf("Unexpected end of input.  Expected IDENTIFIER")
		}
		if p.PeekToken() != IDENTIFIER {
			return nil, p.Errorf("Expected IDENTIFER, Found: %s", p.lexer.Text())
		}

		imported, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		importDecl := &ImportDecl{
			NodeInfo:     newNodeInfo(imported.Pos(), imported.End()),
			ImportedItem: imported,
		}

		if p.PeekToken() == AS {
			// we have an alias
			p.Advance()
			alias, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			importDecl.Alias = alias
		}
		out = append(out, importDecl)

		if p.PeekToken() == FROM {
			break
		} else if p.PeekToken() == COMMA {
			// all good
			p.Advance()
		} else {
			return nil, p.Errorf("expected 'from' or ',' in import declaration, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
		}
	}

	if _, _, err = p.AdvanceIf(FROM); err != nil {
		return nil, p.Errorf("expected 'from' import declaration, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
	}

	if _, err := p.Expect(STRING_LITERAL); err != nil {
		return nil, err
	}

	source, err := p.ParseLiteralExpr()
	if err != nil {
		return nil, err
	}
	// set path on all imports
	for _, imp := range out {
		imp.Path = source
	}
	return out, nil
}

// ParseEnumDecl parses an enumeration declaration.
// Grammar: ENUM IDENTIFIER LBRACE IdentifierList RBRACE
// IdentifierList: IDENTIFIER (COMMA IDENTIFIER)*
func (p *LLParser) ParseEnumDecl(out *EnumDecl) (err error) {
	var enumTokenVal *SDLSymType
	if _, enumTokenVal, err = p.AdvanceIf(ENUM); err != nil {
		return err
	}

	out.NameNode, err = p.ParseIdentifier()
	if err != nil {
		return p.Errorf("expected identifier for enum name after ENUM, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	if _, _, err = p.AdvanceIf(LBRACE); err != nil {
		return p.Errorf("expected '{' after enum name '%s', found %s (%s)", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text())
	}

	// Parse IdentifierList
	out.ValuesNode = []*IdentifierExpr{}
	if p.PeekToken() != RBRACE { // Check if list is not empty
		firstIdent, err := p.ParseIdentifier()
		if err != nil {
			return p.Errorf("expected identifier in enum value list for '%s', found %s (%s): %v", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text(), err)
		}
		out.ValuesNode = append(out.ValuesNode, firstIdent)

		for p.PeekToken() == COMMA {
			p.Advance() // Consume COMMA

			// Optional: Check for trailing comma if not allowed by grammar before RBRACE
			if p.PeekToken() == RBRACE {
				return p.Errorf("trailing comma not allowed in enum '%s' value list before '}'", out.NameNode.Name)
			}

			nextIdent, err := p.ParseIdentifier()
			if err != nil {
				return p.Errorf("expected identifier after comma in enum value list for '%s', found %s (%s): %v", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text(), err)
			}
			out.ValuesNode = append(out.ValuesNode, nextIdent)
		}
	}
	if len(out.ValuesNode) == 0 {
		// Depending on grammar, empty enum values might be an error or allowed.
		// Your yacc grammar `IdentifierList: IDENTIFIER | IdentifierList COMMA IDENTIFIER`
		// implies at least one identifier is required if the list is not empty.
		// If LBRACE RBRACE is allowed, this is fine. If not, add an error check here.
		// For now, let's assume an empty enum `enum E {}` is valid and results in an empty ValuesNode.
	}

	var rbraceTokenVal *SDLSymType
	if _, rbraceTokenVal, err = p.AdvanceIf(RBRACE); err != nil {
		return p.Errorf("expected '}' to close enum '%s', found %s (%s)", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text())
	}

	out.NodeInfo = newNodeInfo(enumTokenVal.node.Pos(), rbraceTokenVal.node.End())
	return nil
}

// ParseSystemDecl parses a system declaration.
// Grammar: SYSTEM IDENTIFIER LBRACE SystemBodyItemOptList RBRACE
// SystemBodyItem: InstanceDecl | OptionsDecl | LetStmt
func (p *LLParser) ParseSystemDecl(out *SystemDecl) (err error) {
	var systemTokenVal *SDLSymType
	if _, systemTokenVal, err = p.AdvanceIf(SYSTEM); err != nil {
		return err
	}

	out.NameNode, err = p.ParseIdentifier()
	if err != nil {
		return p.Errorf("expected identifier for system name after SYSTEM, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	if _, _, err = p.AdvanceIf(LBRACE); err != nil {
		return p.Errorf("expected '{' after system name '%s', found %s (%s)", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text())
	}

	out.Body = []SystemDeclBodyItem{}
	for p.PeekToken() != RBRACE && p.PeekToken() != eof {
		var item SystemDeclBodyItem
		peekedItemStart := p.PeekToken()
		if peekedItemStart == SEMICOLON {
			p.Advance()
			continue
		}
		switch peekedItemStart {
		case USE: // Start of InstanceDecl
			instDecl := &InstanceDecl{}
			if err = p.ParseInstanceDecl(instDecl); err != nil {
				return err
			}
			item = instDecl
		case OPTIONS: // Start of OptionsDecl
			optsDecl := &OptionsDecl{}
			if err = p.ParseOptionsDecl(optsDecl); err != nil {
				return err
			}
			item = optsDecl
		case LET: // Start of LetStmt (which is a Stmt, but can be a SystemBodyItem here)
			letStmt, err := p.ParseLetStmt() // ParseLetStmt returns Stmt
			if err != nil {
				return err
			}
			// We need to ensure LetStmt can be a SystemDeclBodyItem.
			// This requires LetStmt to implement the SystemDeclBodyItem interface.
			// For now, let's assume it does or wrap it if necessary.
			// If LetStmt itself implements systemDeclBodyItemNode():
			var ok bool
			item, ok = letStmt.(SystemDeclBodyItem)
			if !ok {
				// This would be an internal error or require LetStmt to conform.
				// For now, this shows the type assertion challenge.
				// A common approach is for LetStmt to embed a type that implements it, or:
				// type SystemLetItem struct { LetStmt }
				// func (s *SystemLetItem) systemDeclBodyItemNode() {}
				// item = &SystemLetItem{LetStmt: letStmt.(*LetStmt)}
				return p.Errorf("LetStmt is not a valid SystemBodyItem (type assertion failed)")
			}

		default:
			return p.Errorf("unexpected token '%s' in system '%s' body. Expected 'use' or 'let'.",
				TokenString(peekedItemStart), out.NameNode.Name)
		}
		out.Body = append(out.Body, item)
	}

	var rbraceTokenVal *SDLSymType
	if _, rbraceTokenVal, err = p.AdvanceIf(RBRACE); err != nil {
		return p.Errorf("expected '}' to close system '%s', found %s (%s)", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text())
	}

	out.NodeInfo = newNodeInfo(systemTokenVal.node.Pos(), rbraceTokenVal.node.End())
	return nil
}

// ParseInstanceDecl parses an instance declaration (SystemBodyItem).
// Grammar: USE IDENTIFIER IDENTIFIER
//
//	| USE IDENTIFIER IDENTIFIER ASSIGN LBRACE AssignListOpt RBRACE
func (p *LLParser) ParseInstanceDecl(out *InstanceDecl) (err error) {
	var useTokenVal *SDLSymType
	if _, useTokenVal, err = p.AdvanceIf(USE); err != nil {
		return err // Should not happen if called correctly by ParseSystemDecl
	}

	out.NameNode, err = p.ParseIdentifier()
	if err != nil {
		return p.Errorf("expected identifier for instance name after USE, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	out.ComponentType, err = p.ParseIdentifier()
	if err != nil {
		return p.Errorf("expected identifier for component type after instance name '%s', found %s (%s): %v", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	endPos := out.ComponentType.End() // End position if no overrides

	if p.PeekToken() == ASSIGN {
		p.Advance() // Consume ASSIGN

		if _, _, err = p.AdvanceIf(LBRACE); err != nil {
			return p.Errorf("expected '{' for instance '%s' overrides, found %s (%s)", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text())
		}

		// Parse AssignListOpt
		out.Overrides = []*AssignmentStmt{}
		if p.PeekToken() != RBRACE { // Check if list is not empty
			firstAssign, err := p.ParseAssignment() // Helper to parse one IDENTIFIER ASSIGN Expression
			if err != nil {
				return err // Error from ParseAssignment
			}
			out.Overrides = append(out.Overrides, firstAssign)

			// Your grammar for AssignList is `Assignment | AssignList Assignment`.
			// It doesn't explicitly show commas or semicolons.
			// Let's assume assignments are just listed one after another without separators,
			// until RBRACE. If they need separators (like ';'), adjust here.
			for p.PeekToken() != RBRACE && p.PeekToken() != eof {
				// Before parsing next assignment, ensure it's an IDENTIFIER.
				// If it's RBRACE, the loop condition will handle it.
				// If it's something else, ParseAssignment will error.
				if p.PeekToken() == SEMICOLON {
					p.Advance()
					continue
				}
				if p.PeekToken() != IDENTIFIER {
					// This might be too strict if other tokens could start an assignment
					// or if RBRACE is the only valid non-identifier.
					// For `IDENTIFIER ASSIGN Expr`, this check is fine.
					break
				}
				nextAssign, err := p.ParseAssignment()
				if err != nil {
					return err
				}
				out.Overrides = append(out.Overrides, nextAssign)
			}
		}
		var rbraceTokenVal *SDLSymType
		if _, rbraceTokenVal, err = p.AdvanceIf(RBRACE); err != nil {
			return p.Errorf("expected '}' to close instance '%s' overrides, found %s (%s)", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text())
		}
		endPos = rbraceTokenVal.node.End()
	} else {
		out.Overrides = []*AssignmentStmt{} // Ensure it's initialized even if no overrides block
	}

	out.NodeInfo = newNodeInfo(useTokenVal.node.Pos(), endPos)
	return nil
}

// ParseAssignment parses a single `IDENTIFIER ASSIGN Expression`
// Used by ParseInstanceDecl for overrides.
func (p *LLParser) ParseAssignment() (*AssignmentStmt, error) {
	out := &AssignmentStmt{}
	var err error

	// Start position is from the IDENTIFIER
	if _, err = p.Expect(IDENTIFIER); err != nil {
		return nil, p.Errorf("expected identifier for assignment variable, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
	}
	startPos := p.peekedTokenValue.ident.Pos()

	out.Var, err = p.ParseIdentifier()
	if err != nil {
		// This error path should ideally not be hit if Expect passed, but defensive.
		return nil, err
	}

	if _, _, err = p.AdvanceIf(ASSIGN); err != nil {
		return nil, p.Errorf("expected '=' after variable '%s' in assignment, found %s (%s)", out.Var.Name, TokenString(p.PeekToken()), p.lexer.Text())
	}

	out.Value, err = p.ParseExpression()
	if err != nil {
		return nil, p.Errorf("expected expression after '=' for variable '%s' in assignment, found %s (%s): %v", out.Var.Name, TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	out.NodeInfo = newNodeInfo(startPos, out.Value.End())
	return out, nil
}

// ParseOptionsDecl parses an options block (SystemBodyItem or TopLevelDeclaration).
// Grammar: OPTIONS LBRACE StmtList RBRACE
func (p *LLParser) ParseOptionsDecl(out *OptionsDecl) (err error) {
	var optionsTokenVal *SDLSymType
	if _, optionsTokenVal, err = p.AdvanceIf(OPTIONS); err != nil {
		return err // Should not happen if called correctly
	}

	// The body of OptionsDecl is a BlockStmt in the AST
	out.Body, err = p.ParseBlockStmt()
	if err != nil {
		return p.Errorf("error parsing body of OPTIONS block: %v", err)
	}

	out.NodeInfo = newNodeInfo(optionsTokenVal.node.Pos(), out.Body.End())
	return nil
}

// ParseUsesDecl parses a uses declaration (ComponentBodyItem).
// Grammar: USES IDENTIFIER IDENTIFIER
func (p *LLParser) ParseUsesDecl(out *UsesDecl) (err error) {
	var usesTokenVal *SDLSymType
	if _, usesTokenVal, err = p.AdvanceIf(USES); err != nil {
		return err // Should not happen if called correctly
	}

	out.NameNode, err = p.ParseIdentifier()
	if err != nil {
		return p.Errorf("expected identifier for uses name after USES, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	out.ComponentNode, err = p.ParseIdentifier()
	if err != nil {
		return p.Errorf("expected identifier for component type after uses name '%s', found %s (%s): %v", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	out.NodeInfo = newNodeInfo(usesTokenVal.node.Pos(), out.ComponentNode.End())
	return nil
}

// ParseMethodDecl parses a method declaration or signature (ComponentBodyItem).
// isSignatureOnly = true for NATIVE component methods (no body).
// Grammar: METHOD IDENTIFIER LPAREN MethodParamListOpt RPAREN [TypeDecl] [BlockStmt]
func (p *LLParser) ParseMethodDecl(out *MethodDecl, isSignatureOptional bool) (err error) {
	var methodTokenVal *SDLSymType
	if _, methodTokenVal, err = p.AdvanceIf(METHOD); err != nil {
		return err // Should not happen if called correctly
	}

	out.NameNode, err = p.ParseIdentifier()
	if err != nil {
		return p.Errorf("expected identifier for method name after METHOD, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	if _, _, err = p.AdvanceIf(LPAREN); err != nil {
		return p.Errorf("expected '(' after method name '%s', found %s (%s)", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text())
	}

	out.Parameters, err = p.ParseMethodParamListOpt() // Assumes this function is defined (from previous response)
	if err != nil {
		return p.Errorf("error parsing parameters for method '%s': %v", out.NameNode.Name, err)
	}

	var rparenTokenVal *SDLSymType
	if _, rparenTokenVal, err = p.AdvanceIf(RPAREN); err != nil {
		return p.Errorf("expected ')' after parameters for method '%s', found %s (%s)", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text())
	}

	endNodePos := rparenTokenVal.node.End() // End position if no return type or body

	// Check for optional ReturnType
	// A TypeDecl can start with IDENTIFIER or a primitive type keyword (INT, FLOAT, etc.)
	peekedAfterParen := p.PeekToken()
	if peekedAfterParen == IDENTIFIER ||
		peekedAfterParen == INT || peekedAfterParen == FLOAT ||
		peekedAfterParen == STRING || peekedAfterParen == BOOL || peekedAfterParen == DURATION {

		// It *could* be a TypeDecl . But if it's `isSignatureOnly == false` and the next token
		// is LBRACE (start of BlockStmt), then this IDENTIFIER/keyword is NOT a return type.
		// Example: method foo() int {}  vs method foo() {}
		// If it's a signature, any valid TypeDecl start means it IS a return type.
		// If it's a full decl, we need to ensure it's not the LBRACE of the body.

		isPotentiallyBlockStart := (peekedAfterParen == LBRACE && !isSignatureOptional)

		if !isPotentiallyBlockStart { // If it's not LBRACE (for full decl) or if it IS a signature decl
			// Then it must be a return type if it looks like one
			out.ReturnType, err = p.ParseTypeDecl() // Assumes this function is defined
			if err != nil {
				// This error means it looked like a type but wasn't valid.
				return p.Errorf("error parsing return type for method '%s': %v", out.NameNode.Name, err)
			}
			endNodePos = out.ReturnType.End()
		}
	}

	if !isSignatureOptional {
		// Parse BlockStmt for the method body
		if p.PeekToken() != LBRACE {
			return p.Errorf("expected '{' for method body of '%s', found %s (%s)", out.NameNode.Name, TokenString(p.PeekToken()), p.lexer.Text())
		}
		out.Body, err = p.ParseBlockStmt()
		if err != nil {
			return p.Errorf("error parsing body for method '%s': %v", out.NameNode.Name, err)
		}
		endNodePos = out.Body.End()
	}

	out.NodeInfo = newNodeInfo(methodTokenVal.node.Pos(), endNodePos)
	return nil
}

// ParseParamDecl parses a parameter declaration (ComponentBodyItem).
// Grammar: PARAM IDENTIFIER TypeDecl [ASSIGN Expression]
func (p *LLParser) ParseParamDecl(out *ParamDecl) (err error) {
	var paramTokenVal *SDLSymType
	if _, paramTokenVal, err = p.AdvanceIf(PARAM); err != nil {
		return err // Should not happen if called correctly
	}

	out.Name, err = p.ParseIdentifier()
	if err != nil {
		return p.Errorf("expected identifier for param name after PARAM, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	out.TypeDecl, err = p.ParseTypeDecl() // Assumes this function is defined
	if err != nil {
		return p.Errorf("expected type name for param '%s', found %s (%s): %v", out.Name.Name, TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	endPos := out.TypeDecl.End() // End position if no default value

	// Check for optional default value
	if p.PeekToken() == ASSIGN {
		p.Advance() // Consume ASSIGN
		out.DefaultValue, err = p.ParseExpression()
		if err != nil {
			return p.Errorf("expected expression for default value of param '%s', found %s (%s): %v", out.Name.Name, TokenString(p.PeekToken()), p.lexer.Text(), err)
		}
		endPos = out.DefaultValue.End()
	}

	out.NodeInfo = newNodeInfo(paramTokenVal.node.Pos(), endPos)
	return nil
}

// ParseBlockStmt parses a block of statements enclosed in braces.
// Grammar: LBRACE StmtList RBRACE
func (p *LLParser) ParseBlockStmt() (out *BlockStmt, err error) {
	out = &BlockStmt{}
	var lbraceTokenVal *SDLSymType
	if _, lbraceTokenVal, err = p.AdvanceIf(LBRACE); err != nil {
		return nil, err
	}

	// Parse statements until RBRACE or EOF
	out.Statements, err = p.ParseStmtList(RBRACE)
	if err != nil {
		return nil, err // Error from parsing statements
	}

	var rbraceTokenVal *SDLSymType
	if _, rbraceTokenVal, err = p.AdvanceIf(RBRACE); err != nil {
		return nil, p.Errorf("expected '}' to close block, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
	}
	out.NodeInfo = newNodeInfo(lbraceTokenVal.node.Pos(), rbraceTokenVal.node.End())
	return out, nil
}

// **2. Parsing Lists **

// ParseStmtList parses a sequence of statements until a closing token or EOF.
// It's used by ParseBlockStmt. The closing token itself is NOT consumed by this function.
func (p *LLParser) ParseStmtList(closingTokens ...int) (stmts []Stmt, err error) {
	for {
		peeked := p.PeekToken()
		if peeked == eof {
			break // End of file
		}
		if peeked == SEMICOLON {
			p.Advance()
			continue
		}
		isClosing := false
		for _, closer := range closingTokens {
			if peeked == closer {
				isClosing = true
				break
			}
		}
		if isClosing {
			break // Found one of the closing tokens
		}

		stmt, err := p.ParseStmt() // You'll implement/flesh out ParseStmt
		if err != nil {
			return nil, err // Propagate error from statement parsing
		}
		if stmt == nil { // ParseStmt might return nil if it's an optional construct it didn't find
			// This indicates an issue or an empty statement if allowed.
			// For now, let's assume ParseStmt always returns a Stmt or an error.
			// If ParseStmt can return nil, decide if that's an error here.
			return nil, p.Errorf("ParseStmt returned nil unexpectedly at %s", TokenString(p.PeekToken()))
		}
		stmts = append(stmts, stmt)
	}
	return stmts, nil
}

// ParseArgList parses a comma-separated list of one or more expressions.
// Grammar: Expression (COMMA Expression)*
// Used in CallExpr.
func (p *LLParser) ParseArgList() (args []Expr, err error) {
	firstArg, err := p.ParseExpression()
	if err != nil {
		return nil, p.Errorf("expected expression for argument list, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}
	args = append(args, firstArg)

	for p.PeekToken() == COMMA {
		p.Advance() // Consume COMMA
		nextArg, err := p.ParseExpression()
		if err != nil {
			return nil, p.Errorf("expected expression after ',' in argument list, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
		}
		args = append(args, nextArg)
	}
	return args, nil
}

// **3. Parsing Optionals and Optional Lists **

// ParseMethodParamListOpt parses an optional comma-separated list of parameter declarations
// typically found within parentheses, e.g., in a method signature.
// Grammar: ( MethodParamDecl (COMMA MethodParamDecl)* )?
// Assumes LPAREN was consumed by caller, caller will consume RPAREN.
func (p *LLParser) ParseMethodParamListOpt() (params []*ParamDecl, err error) {
	params = []*ParamDecl{} // Initialize to empty list

	// If the next token is the closing token for the list (e.g., RPAREN), it's an empty list.
	if p.PeekToken() == RPAREN {
		return params, nil
	}

	// Parse the first parameter
	// ParseMethodParamDecl is specific to your grammar: IDENTIFIER TypeDecl [ASSIGN Expression]
	firstParam, err := p.ParseMethodParamDecl()
	if err != nil {
		return nil, err // Error parsing the first parameter
	}
	params = append(params, firstParam)

	// Parse subsequent parameters separated by COMMA
	for p.PeekToken() == COMMA {
		p.Advance() // Consume COMMA

		// Optional: Check for trailing comma if not allowed (e.g. if next is RPAREN)
		if p.PeekToken() == RPAREN {
			return nil, p.Errorf("trailing comma not allowed in parameter list at %s", TokenString(RPAREN))
		}

		nextParam, err := p.ParseMethodParamDecl()
		if err != nil {
			return nil, err // Error parsing subsequent parameter
		}
		params = append(params, nextParam)
	}
	return params, nil
}

// ParseMethodParamDecl parses a single parameter declaration for a method.
// Grammar: IDENTIFIER TypeDecl [ASSIGN Expression]
func (p *LLParser) ParseMethodParamDecl() (out *ParamDecl, err error) {
	out = &ParamDecl{}
	var identNode *IdentifierExpr

	identNode, err = p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	out.Name = identNode

	// Store start position from the first token of the declaration
	startPos := identNode.Pos()

	out.TypeDecl, err = p.ParseTypeDecl() // You'll need to implement ParseTypeDecl
	if err != nil {
		return nil, err
	}

	endPos := out.TypeDecl.End() // End position so far

	// Optional: Parse default value if your grammar supports it here
	// if p.PeekToken() == ASSIGN {
	//     p.Advance() // Consume ASSIGN
	//     defaultValue, err := p.ParseExpression()
	//     if err != nil {
	//         return nil, err
	//     }
	//     out.DefaultValue = defaultValue
	//     endPos = defaultValue.End()
	// }
	out.NodeInfo = newNodeInfo(startPos, endPos)
	return out, nil
}

// ParseTypeDecl parses a type name.
// Grammar: IDENTIFIER | PrimitiveType (INT | FLOAT | STRING | BOOL | DURATION)
func (p *LLParser) ParseTypeDecl() (out *TypeDecl, err error) {
	peeked := p.PeekToken()
	typeNameTokenVal := p.peekedTokenValue
	p.Advance() // Consume the type token

	out = &TypeDecl{NodeInfo: newNodeInfoFromToken(typeNameTokenVal)}

	switch peeked {
	case IDENTIFIER:
		// If lexer puts IdentifierExpr in .ident for IDENTIFIER token
		out.Name = typeNameTokenVal.ident.Name
	default:
		// Revert advance if token was not a valid type
		// This is tricky. PeekToken/Advance model makes this hard.
		// Better to use p.Expect(IDENTIFIER, INT, ...) and then p.AdvanceIf
		// For now, we assume the advance was correct and raise error.
		return nil, p.Errorf("expected type name (identifier or primitive), found %s (%s)",
			TokenString(peeked), p.lexer.Text())
	}
	return out, nil
}

// --- Expression Parsing (Recursive Descent with Precedence) ---

// ParseExpression is the entry point for parsing any expression.
// It starts with the lowest precedence operator (OR in your grammar).
func (p *LLParser) ParseExpression() (Expr, error) {
	// return p.ParseOrExpr() // Start with the lowest precedence level
	return p.ParseChainedExpr()
}

// ChainedExpr : UnaryExpr ( BINARY_OP UnaryExpr ) *
// Idea is to collect them and then sort out precedences dynamically
func (p *LLParser) ParseChainedExpr() (Expr, error) {
	left, err := p.ParseUnaryExpr()
	if err != nil {
		return nil, err
	}
	out := &ChainedExpr{Children: []Expr{left}}

	for {
		currentPeekedToken := p.PeekToken()
		if currentPeekedToken != BINARY_OP {
			break
		}
		opToken := currentPeekedToken
		opTokenVal := p.peekedTokenValue // Capture semantic value of operator before advancing
		p.Advance()                      // Consume the operator

		next, err := p.ParseUnaryExpr()
		if err != nil {
			return nil, p.Errorf("expected unary expression after operator '%s', found %s (%s): %v",
				TokenString(opToken), TokenString(p.PeekToken()), p.lexer.Text(), err)
		}
		out.Children = append(out.Children, next)
		out.Operators = append(out.Operators, opTokenVal.node.(*TokenNode).Text)
	}
	return out, nil
}

// Generic helper for parsing left-associative binary expressions for a given precedence level.
func (p *LLParser) parseBinaryExpr(
	parseHigherPrecedenceOperand func() (Expr, error), // Function to parse the operands (or next higher precedence)
	operators ...int) (Expr, error) {

	left, err := parseHigherPrecedenceOperand()
	if err != nil {
		return nil, err
	}

	for {
		currentPeekedToken := p.PeekToken()
		isCurrentLevelOperator := false
		for _, opToken := range operators {
			if currentPeekedToken == opToken {
				isCurrentLevelOperator = true
				break
			}
		}

		if !isCurrentLevelOperator {
			break // No more operators at this precedence level
		}

		// Matched an operator for the current precedence level
		opToken := currentPeekedToken
		opTokenVal := p.peekedTokenValue // Capture semantic value of operator before advancing
		p.Advance()                      // Consume the operator

		right, err := parseHigherPrecedenceOperand()
		if err != nil {
			return nil, p.Errorf("expected expression after operator '%s', found %s (%s): %v",
				TokenString(opToken), TokenString(p.PeekToken()), p.lexer.Text(), err)
		}

		// Create binary expression node
		left = &BinaryExpr{
			ExprBase: ExprBase{NodeInfo: newNodeInfo(left.Pos(), right.End())},
			Left:     left,
			Operator: opTokenVal.node.String(), // Assumes operator token's text from Node via TokenNode
			Right:    right,
		}
		// Loop continues to handle left-associativity (e.g., a + b + c)
	}
	return left, nil
}

// OrExpr: AndBoolExpr ( OR AndBoolExpr )*
/*
func (p *LLParser) ParseOrExpr() (Expr, error) {
	return p.parseBinaryExpr(p.ParseAndBoolExpr, OR)
}

// AndBoolExpr: CmpExpr ( AND CmpExpr )*
func (p *LLParser) ParseAndBoolExpr() (Expr, error) {
	return p.parseBinaryExpr(p.ParseCmpExpr, AND)
}

// CmpExpr: AddExpr ( (EQ|NEQ|LT|LTE|GT|GTE) AddExpr )?
// Comparison operators are non-associative. This means `a == b == c` is an error.
// We parse one comparison.
func (p *LLParser) ParseCmpExpr() (Expr, error) {
	left, err := p.ParseAddExpr()
	if err != nil {
		return nil, err
	}

	peeked := p.PeekToken()
	switch peeked {
	case EQ, NEQ, LT, LTE, GT, GTE:
		opToken := peeked
		opTokenVal := p.peekedTokenValue
		p.Advance() // Consume operator

		right, err := p.ParseAddExpr() // Parse the right-hand operand
		if err != nil {
			return nil, p.Errorf("expected expression after comparison operator '%s', found %s (%s): %v",
				TokenString(opToken), TokenString(p.PeekToken()), p.lexer.Text(), err)
		}

		// Optional: Check for chaining if strict non-associativity is required.
		// If p.PeekToken() is another comparison operator, it's an error.
		// peekedAfterRight := p.PeekToken()
		// if peekedAfterRight == EQ || ... {
		//    return nil, p.Errorf("comparison operators are non-associative and cannot be chained at %s", TokenString(peekedAfterRight))
		// }

		return &BinaryExpr{
			NodeInfo: newNodeInfo(left.Pos(), right.End()),
			Left:     left,
			Operator: opTokenVal.node.String(),
			Right:    right,
		}, nil
	default:
		return left, nil // No comparison operator, just return the AddExpr
	}
}

// AddExpr: MulExpr ( (PLUS|MINUS) MulExpr )*
func (p *LLParser) ParseAddExpr() (Expr, error) {
	return p.parseBinaryExpr(p.ParseMulExpr, PLUS, MINUS)
}

// MulExpr: UnaryExpr ( (MUL|DIV|MOD) UnaryExpr )*
func (p *LLParser) ParseMulExpr() (Expr, error) {
	return p.parseBinaryExpr(p.ParseUnaryExpr, MUL, DIV, MOD)
}
*/

// UnaryExpr: (NOT | MINUS) UnaryExpr | PrimaryExpr
// Unary operators are right-associative (e.g., !!a or --a).
func (p *LLParser) ParseUnaryExpr() (Expr, error) {
	peeked := p.PeekToken()
	if peeked == UNARY_OP || peeked == MINUS { // MINUS for unary negation (%prec UMINUS)
		opToken := peeked
		opTokenVal := p.peekedTokenValue
		p.Advance() // Consume operator

		// Recursively call ParseUnaryExpr for right-associativity
		operand, err := p.ParseUnaryExpr()
		if err != nil {
			return nil, p.Errorf("expected expression after unary operator '%s', found %s (%s): %v",
				TokenString(opToken), TokenString(p.PeekToken()), p.lexer.Text(), err)
		}
		return &UnaryExpr{
			ExprBase: ExprBase{NodeInfo: newNodeInfo(opTokenVal.node.Pos(), operand.End())},
			Operator: opTokenVal.node.String(),
			Right:    operand,
		}, nil
	}
	// If not a recognized unary operator, parse the next higher precedence (PrimaryExpr)
	return p.ParsePrimaryExpr()
}

// PrimaryExpr: Literal | IDENTIFIER | LPAREN Expression RPAREN | (PrimaryExpr PostfixOps)
// PostfixOps: DOT IDENTIFIER | LPAREN ArgListOpt RPAREN
func (p *LLParser) ParsePrimaryExpr() (expr Expr, err error) {
	peeked := p.PeekToken()
	// startTokenVal := p.peekedTokenValue // For NodeInfo if it's a simple token

	switch peeked {
	case INT_LITERAL, FLOAT_LITERAL, STRING_LITERAL, DURATION_LITERAL, BOOL_LITERAL:
		expr, err = p.ParseLiteralExpr()
		if err != nil {
			return nil, err
		}

	case IDENTIFIER:
		// ParseIdentifier returns *IdentifierExpr
		expr, err = p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

	case LPAREN:
		// lparenTokenVal := p.peekedTokenValue
		p.Advance()                           // Consume LPAREN
		innerExpr, err := p.ParseExpression() // Parse grouped expression
		if err != nil {
			return nil, err
		}
		// var rparenTokenVal *SDLSymType
		if _ /*rparenTokenVal*/, _, err = p.AdvanceIf(RPAREN); err != nil {
			return nil, p.Errorf("expected ')' to close parenthesized expression, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
		}
		// The yacc grammar `LPAREN Expression RPAREN { $$ = $2 }` means the resulting node IS the inner expression.
		// If you want the parens to contribute to NodeInfo, you'd wrap it or adjust NodeInfo.
		// For now, follow yacc: $$ = $2. The NodeInfo is that of innerExpr.
		expr = innerExpr
		// If you want to wrap it, e.g., in a GroupedExpr or update its NodeInfo:
		// expr = &GroupedExpr{NodeInfo: newNodeInfo(lparenTokenVal.node.Pos(), rparenTokenVal.node.End()), Expr: innerExpr}

	default:
		return nil, p.Errorf("unexpected token at start of primary expression: %s (%s)",
			TokenString(peeked), p.lexer.Text())
	}

	// After parsing a base primary (like IDENTIFIER, literal, or parenthesized expr),
	// check for postfix operators like member access ('.') or function call ('(').
	// This loop handles left-recursive grammar rules like:
	// PrimaryExpr: PrimaryExpr DOT IDENTIFIER
	// PrimaryExpr: PrimaryExpr LPAREN ArgListOpt RPAREN
	for {
		peekedSuffix := p.PeekToken()
		if peekedSuffix == DOT {
			p.Advance() // Consume DOT
			// dotTokenVal := p.peekedTokenValue // For NodeInfo of DOT if needed

			memberIdent, err := p.ParseIdentifier()
			if err != nil {
				return nil, p.Errorf("expected identifier after '.', found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
			}
			expr = &MemberAccessExpr{
				ExprBase: ExprBase{NodeInfo: newNodeInfo(expr.Pos(), memberIdent.End())}, // Spans from start of receiver to end of member
				Receiver: expr,
				Member:   memberIdent,
			}
		} else if peekedSuffix == LPAREN {
			// lparenTokenVal := p.peekedTokenValue // For NodeInfo of LPAREN
			p.Advance() // Consume LPAREN

			var args []Expr
			// ArgListOpt: /* empty */ | ArgList
			// Call ParseArgList only if not immediately followed by RPAREN
			if p.PeekToken() != RPAREN {
				args, err = p.ParseArgList() // ParseArgList parses one or more
				if err != nil {
					return nil, err
				}
			} else {
				args = []Expr{} // Empty argument list
			}
			var rparenTokenVal *SDLSymType
			if _, rparenTokenVal, err = p.AdvanceIf(RPAREN); err != nil {
				return nil, p.Errorf("expected ')' to close function call arguments, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
			}
			expr = &CallExpr{
				ExprBase: ExprBase{NodeInfo: newNodeInfo(expr.Pos(), rparenTokenVal.node.End())},
				Function: expr,
				Args:     args,
			}
		} else {
			break // No more postfix operators applicable to 'expr'
		}
	}
	return expr, nil
}

// ParseLiteralExpr parses a literal value.
// Grammar: INT_LITERAL | FLOAT_LITERAL | STRING_LITERAL | BOOL_LITERAL | DURATION_LITERAL
func (p *LLParser) ParseLiteralExpr() (*LiteralExpr, error) {
	peeked := p.PeekToken()
	if p.lexer.lastError != nil {
		return nil, p.lexer.lastError
	}

	// Expect one of the literal tokens
	switch peeked {
	case INT_LITERAL, FLOAT_LITERAL, STRING_LITERAL, DURATION_LITERAL, BOOL_LITERAL:
		// Valid literal token
	default:
		return nil, p.Errorf("expected a literal, found %s (%s)", TokenString(peeked), p.lexer.Text())
	}

	litTokenVal := p.peekedTokenValue
	p.Advance() // Consume the literal token

	// In a yacc setup, the lexer populates SDLSymType.expr or a specific literal field.
	// In a handwritten parser, we construct the LiteralExpr AST node here.
	out := &LiteralExpr{
		ExprBase: ExprBase{NodeInfo: newNodeInfoFromToken(litTokenVal)},
		Value:    litTokenVal.expr.(*LiteralExpr).Value,
		// Or use litTokenVal.sval if populated by lexer for literals
	}

	// You might parse out.Value into a Go native type (int, float64, bool) here or later.
	return out, nil
}

// --- Statement Parsing ---

// ParseStmt (you have a stub, this is a more fleshed-out dispatcher)
// This is a crucial part where you decide which specific statement parser to call.
func (p *LLParser) ParseStmt() (out Stmt, err error) {
	peeked := p.PeekToken()

	switch peeked {
	case LBRACE: // Start of a BlockStmt
		return p.ParseBlockStmt()
	case LET:
		return p.ParseLetStmt() // You'll implement this
	case RETURN:
		return p.ParseReturnStmt() // You'll implement this
	case IF:
		return p.ParseIfStmt() // You'll implement this
	case WAIT:
		return p.ParseWaitStmt() // You'll implement this
	case DELAY:
		return p.ParseDelayStmt() // You'll implement this
	case DISTRIBUTE: // Added
		return p.ParseDistributeExpr()
		/*
			case GO: // Added
				goStmt, goErr := p.ParseGoStmt()
				if goErr != nil {
					return nil, goErr
				}
				// Handle the error case for `go var = Expr;`
				if gs, ok := goStmt.(*GoStmt); ok && gs.IsExprAssignment {
					// Report error as per grammar for `go IDENTIFIER ASSIGN Expression `
					// You can return the node + an error, or just the error,
					// or have a separate error collection mechanism.
					// For now, let's return an explicit error that the caller of ParseStmt must check.
					return gs, p.Errorf("`go` with variable assignment currently only supports assigning blocks, not expressions (at pos %d near '%s')",
						gs.Pos(), p.lexer.Text()) // Use gs.Pos() for more accuracy
				}
				return goStmt, nil
		*/
	// ... Add cases for other statement-starting keywords:
	// case LOG: return p.ParseLogStmt()
	// case SWITCH: return p.ParseSwitchStmt()

	default:
		// If no specific statement keyword, it might be an ExpressionStatement.
		// ExpressionStatement: Expression (as per your grammar)
		// This needs care: ensure an expression can indeed start with the current token.
		expr, exprErr := p.ParseExpression()
		if exprErr != nil {
			// If parsing an expression fails, it's not a valid statement start.
			return nil, p.Errorf("expected statement, found %s (%s). Attempting to parse as expression failed: %v",
				TokenString(peeked), p.lexer.Text(), exprErr)
		}

		// Successfully parsed an expression, now expect a SEMICOLON.
		var semiTokenVal *SDLSymType
		if _, semiTokenVal, err = p.AdvanceIf(SEMICOLON); err != nil {
			return nil, p.Errorf("expected ';' after expression statement, found %s (%s) (expression was: %T)",
				TokenString(p.PeekToken()), p.lexer.Text(), expr)
		}
		return &ExprStmt{
			NodeInfo:   newNodeInfo(expr.Pos(), semiTokenVal.node.End()),
			Expression: expr,
		}, nil
	}
}

// Example: ParseLetStmt (you can create similar ones for other statements)
// Grammar: LET IDENTIFIER ASSIGN Expression
func (p *LLParser) ParseLetStmt() (Stmt, error) {
	var letTokenVal *SDLSymType
	var err error

	if _, letTokenVal, err = p.AdvanceIf(LET); err != nil {
		return nil, err
	}

	varIdent, err := p.ParseIdentifier()
	if err != nil {
		return nil, p.Errorf("expected identifier after LET, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	if _, _, err = p.AdvanceIf(ASSIGN); err != nil {
		return nil, p.Errorf("expected '=' after identifier in LET statement, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
	}

	valExpr, err := p.ParseExpression()
	if err != nil {
		return nil, p.Errorf("expected expression after '=' in LET statement, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	// Your grammar does not show a semicolon for LetStmt itself.
	// It's a Stmt directly. If a semicolon is required, add AdvanceIf(SEMICOLON).
	return &LetStmt{
		NodeInfo:  newNodeInfo(letTokenVal.node.Pos(), valExpr.End()),
		Variables: []*IdentifierExpr{varIdent},
		Value:     valExpr,
	}, nil
}

// Example: ParseReturnStmt
// Grammar: RETURN Expression | RETURN SEMICOLON
func (p *LLParser) ParseReturnStmt() (Stmt, error) {
	var returnTokenVal *SDLSymType
	var err error

	if _, returnTokenVal, err = p.AdvanceIf(RETURN); err != nil {
		return nil, err
	}

	// Check if it's `RETURN SEMICOLON` (return without value)
	if p.PeekToken() == SEMICOLON {
		semiTokenVal := p.peekedTokenValue
		p.Advance() // Consume SEMICOLON
		return &ReturnStmt{
			NodeInfo:    newNodeInfo(returnTokenVal.node.Pos(), semiTokenVal.node.End()),
			ReturnValue: nil,
		}, nil
	}

	// Otherwise, parse `RETURN Expression` (semicolon might be handled by ExprStmt or required after)
	// Your grammar shows `ReturnStmt` itself does not consume a final semicolon.
	// If `RETURN Expression ;` is required, the semicolon is part of an `ExprStmt` wrapping `ReturnStmt`
	// or this function should consume it. Let's assume `ReturnStmt` rule handles the expression only.
	returnValue, err := p.ParseExpression()
	if err != nil {
		return nil, p.Errorf("expected expression or ';' after RETURN, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	return &ReturnStmt{
		NodeInfo:    newNodeInfo(returnTokenVal.node.Pos(), returnValue.End()),
		ReturnValue: returnValue,
	}, nil
}

// Example: ParseIfStmt
// Grammar: IF Expression BlockStmt IfStmtElseOpt
// IfStmtElseOpt: /* empty */ | ELSE IfStmt | ELSE BlockStmt
func (p *LLParser) ParseIfStmt() (Stmt, error) {
	var ifTokenVal *SDLSymType
	var err error

	if _, ifTokenVal, err = p.AdvanceIf(IF); err != nil {
		return nil, err
	}

	condition, err := p.ParseExpression()
	if err != nil {
		return nil, p.Errorf("expected condition expression after IF, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	thenBlock, err := p.ParseBlockStmt()
	if err != nil {
		return nil, p.Errorf("expected block statement for THEN branch of IF, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	var elseStmt Stmt // Can be *IfStmt, *BlockStmt, or nil
	endPos := thenBlock.End()

	if p.PeekToken() == ELSE {
		p.Advance() // Consume ELSE

		// Check for 'else if' vs 'else { ... }'
		if p.PeekToken() == IF {
			elseStmt, err = p.ParseIfStmt() // Recursive call for 'else if'
			if err != nil {
				return nil, err
			}
		} else if p.PeekToken() == LBRACE {
			elseStmt, err = p.ParseBlockStmt() // 'else { ... }'
			if err != nil {
				return nil, err
			}
		} else {
			return nil, p.Errorf("expected IF or '{' after ELSE, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
		}
		if elseStmt != nil {
			endPos = elseStmt.End()
		}
	}

	return &IfStmt{
		NodeInfo:  newNodeInfo(ifTokenVal.node.Pos(), endPos),
		Condition: condition,
		Then:      thenBlock,
		Else:      elseStmt,
	}, nil
}

// ParseDistributeExpr parses a distribute statement.
// Grammar: DISTRIBUTE TotalClauseOpt LBRACE CaseExprListOpt DefaultCaseOpt RBRACE
// TotalClauseOpt: /* empty */ | Expression
// CaseExprListOpt: /* empty */ | CaseExprListOpt CaseExpr
// CaseExpr: Expression ARROW Stmt
// DefaultCaseOpt: /* empty */ | DefaultCase
// DefaultCase: DEFAULT ARROW Stmt
func (p *LLParser) ParseDistributeExpr() (Stmt, error) {
	out := &DistributeExpr{}
	var distributeTokenVal *SDLSymType
	var err error

	if _, distributeTokenVal, err = p.AdvanceIf(DISTRIBUTE); err != nil {
		return nil, err
	}
	out.NodeInfo.StartPos = distributeTokenVal.node.Pos()

	// Parse TotalClauseOpt: Optional Expression
	// If the next token is not LBRACE, it could be an expression for TotalClause.
	if p.PeekToken() != LBRACE {
		// Need to be careful: An expression can start with LBRACE (e.g. map literal, if supported)
		// However, for `DISTRIBUTE Total { ... }`, the LBRACE is a clear delimiter.
		// We rely on the fact that an expression parser (ParseExpression) would consume tokens
		// and leave LBRACE if it's not part of that expression.
		// A simpler check is if it's not an LBRACE, try to parse an expression.
		// This assumes expressions don't usually end right before an LBRACE that isn't part of them,
		// which is generally true unless expressions themselves can be blocks.
		// Given the grammar TotalClauseOpt LBRACE, the expression *must* terminate before LBRACE.

		// Let's try to parse an expression. If it fails and next is LBRACE, it was optional.
		// If it fails and next is NOT LBRACE, it's an error.
		// If it succeeds, we must then see LBRACE.

		// A robust way: peek. If not LBRACE, attempt to parse Expr.
		// If successful, out.Total = expr. Then expect LBRACE.
		// If LBRACE directly, out.Total remains nil.
		out.TotalProb, err = p.ParseExpression()
		if err != nil {
			// If parsing expression failed, but the next token IS LBRACE,
			// then the TotalClause was indeed empty, and the failure was likely
			// because the current token was not the start of a valid expression.
			// This can happen if the "Expression" rule is too greedy or has ambiguities.
			// For now, assume ParseExpression correctly fails if no expression is present.
			return nil, p.Errorf("expected expression for total clause or '{' in DISTRIBUTE statement, found %s (%s): %v",
				TokenString(p.PeekToken()), p.lexer.Text(), err)
		}
	}

	if _, _, err = p.AdvanceIf(LBRACE); err != nil {
		return nil, p.Errorf("expected '{' to start DISTRIBUTE statement body, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
	}

	// Parse CaseExprListOpt
	out.Cases = []*CaseExpr{}
	for p.PeekToken() != RBRACE && p.PeekToken() != DEFAULT && p.PeekToken() != eof {
		caseExpr, err := p.ParseExpression()
		if err != nil {
			return nil, p.Errorf("expected expression for DISTRIBUTE case condition, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
		}

		// var arrowTokenVal *SDLSymType
		if _, _, err = p.AdvanceIf(ARROW); err != nil {
			return nil, p.Errorf("expected '=>' after DISTRIBUTE case condition, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
		}

		// A case body is a single Stmt, not necessarily a BlockStmt.
		// Stmt: LetStmt | ExprStmt | ReturnStmt | IfStmt | BlockStmt etc.
		caseBody, err := p.ParseExpression()
		if err != nil {
			return nil, p.Errorf("error parsing DISTRIBUTE case body: %v", err)
		}

		distCase := &CaseExpr{
			ExprBase:  ExprBase{NodeInfo: newNodeInfo(caseExpr.Pos(), caseBody.End())},
			Condition: caseExpr,
			Body:      caseBody,
		}
		out.Cases = append(out.Cases, distCase)
	}

	// Parse DefaultCaseOpt
	if p.PeekToken() == DEFAULT {
		p.Advance() // Consume DEFAULT

		if _, _, err = p.AdvanceIf(ARROW); err != nil {
			return nil, p.Errorf("expected '=>' after DEFAULT in DISTRIBUTE statement, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
		}

		defaultBody, err := p.ParseExpression()
		if err != nil {
			return nil, p.Errorf("error parsing DEFAULT case body in DISTRIBUTE statement: %v", err)
		}
		out.Default = defaultBody
	}

	var rbraceTokenVal *SDLSymType
	if _, rbraceTokenVal, err = p.AdvanceIf(RBRACE); err != nil {
		return nil, p.Errorf("expected '}' to close DISTRIBUTE statement, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
	}
	out.NodeInfo.StopPos = rbraceTokenVal.node.End()

	return out, nil
}

// ParseGoStmt parses a go statement.
// Grammar:
//
//	GO IDENTIFIER ASSIGN BlockStmt
//	GO BlockStmt
//	GO IDENTIFIER ASSIGN Expression SEMICOLON (This is an error case in yacc)
func (p *LLParser) ParseGoStmt() (Stmt, error) {
	/*
			out := &GoStmt{}
			var goTokenVal *SDLSymType
			var err error

			if _, goTokenVal, err = p.AdvanceIf(GO); err != nil {
				return nil, err
			}
			out.NodeInfo.StartPos = goTokenVal.node.Pos()

			if p.PeekToken() == IDENTIFIER {
				// Could be `GO IDENTIFIER ASSIGN BlockStmt` or `GO IDENTIFIER ASSIGN Expression SEMICOLON`
				// identPeek := p.peekedTokenValue    // Peek at identifier before parsing it
				// nextTokenAfterIdent := p.lexer.Lex // This is tricky, we need lookahead(2)
				// For a simple LL(1) parser, this is harder.
				// Let's parse the identifier first.

				var varName *IdentifierExpr
				varName, err = p.ParseIdentifier() // Consumes IDENTIFIER
				if err != nil {
					// This shouldn't happen if PeekToken was IDENTIFIER, but defensive.
					return nil, err
				}
				out.VarName = varName

				if _, _, err = p.AdvanceIf(ASSIGN); err != nil {
					return nil, p.Errorf("expected '=' after identifier '%s' in GO statement, found %s (%s)", out.VarName.Name, TokenString(p.PeekToken()), p.lexer.Text())
				}

				if p.PeekToken() == LBRACE { // `GO IDENTIFIER ASSIGN BlockStmt`
					block, blockErr := p.ParseBlockStmt()
					if blockErr != nil {
						return nil, blockErr
					}
					out.Stmt = block
					out.NodeInfo.StopPos = block.End()
				} else { // `GO IDENTIFIER ASSIGN Expression SEMICOLON` (error case)
					expr, exprErr := p.ParseExpression()
					if exprErr != nil {
						return nil, p.Errorf("expected block or expression after '=' in GO statement, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), exprErr)
					}
					out.Stmt = expr // Store the expression, but mark as error type
					out.IsExprAssignment = true

					var semiTokenVal *SDLSymType
					if _, semiTokenVal, err = p.AdvanceIf(SEMICOLON); err != nil {
						return nil, p.Errorf("expected ';' after expression in GO statement assignment, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
					}
					out.NodeInfo.StopPos = semiTokenVal.node.End()

					// As per grammar's yyerror, this form is an error.
					// The parser can construct the node, but should signal an error.
					// One way is to return the node AND an error, or have a separate error reporting.
					// For now, let the caller (or a later semantic pass) handle the IsExprAssignment flag.
					// Alternatively, return an error directly here:
					// return out, p.Errorf("`go` with variable assignment currently only supports assigning blocks, not expressions, at pos %d", goTokenVal.node.Pos())
				}
			} else if p.PeekToken() == LBRACE { // `GO BlockStmt`
				block, blockErr := p.ParseBlockStmt()
				if blockErr != nil {
					return nil, blockErr
				}
				out.Stmt = block
				out.NodeInfo.StopPos = block.End()
			} else {
				return nil, p.Errorf("expected identifier or '{' after GO, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
			}
		return out, nil
	*/
	return nil, nil
}

// ParseDelayStmt parses a delay statement.
// Grammar: DELAY Expression
// The semicolon is handled by the statement list context, not here.
func (p *LLParser) ParseDelayStmt() (Stmt, error) {
	out := &DelayStmt{}
	var delayTokenVal *SDLSymType
	var err error

	if _, delayTokenVal, err = p.AdvanceIf(DELAY); err != nil {
		return nil, err
	}

	out.Duration, err = p.ParseExpression()
	if err != nil {
		return nil, p.Errorf("expected expression for DELAY duration, found %s (%s): %v", TokenString(p.PeekToken()), p.lexer.Text(), err)
	}

	out.NodeInfo = newNodeInfo(delayTokenVal.node.Pos(), out.Duration.End())
	return out, nil
}

// ParseWaitStmt parses a wait statement.
// Grammar: WAIT IdentifierList
// IdentifierList: IDENTIFIER (COMMA IDENTIFIER)*
// The semicolon is handled by the statement list context.
func (p *LLParser) ParseWaitStmt() (Stmt, error) {
	out := &WaitStmt{}
	var waitTokenVal *SDLSymType
	var err error

	if _, waitTokenVal, err = p.AdvanceIf(WAIT); err != nil {
		return nil, err
	}

	// Parse IdentifierList
	out.Idents = []*IdentifierExpr{}
	if p.PeekToken() != IDENTIFIER {
		return nil, p.Errorf("expected at least one identifier after WAIT, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
	}

	firstIdent, err := p.ParseIdentifier()
	if err != nil {
		// Should be caught by the PeekToken check above, but defensive.
		return nil, err
	}
	out.Idents = append(out.Idents, firstIdent)

	for p.PeekToken() == COMMA {
		p.Advance() // Consume COMMA

		// Optional: Check for trailing comma if not allowed.
		// If the next token is not IDENTIFIER, it's an error (e.g. WAIT id1, ;)
		if p.PeekToken() != IDENTIFIER {
			return nil, p.Errorf("expected identifier after comma in WAIT statement, found %s (%s)", TokenString(p.PeekToken()), p.lexer.Text())
		}

		nextIdent, err := p.ParseIdentifier()
		if err != nil {
			return nil, err // Error parsing subsequent identifier
		}
		out.Idents = append(out.Idents, nextIdent)
	}

	if len(out.Idents) == 0 {
		// This case should ideally be caught earlier, but as a final check.
		return nil, p.Errorf("WAIT statement requires at least one identifier")
	}

	out.NodeInfo = newNodeInfo(waitTokenVal.node.Pos(), out.Idents[len(out.Idents)-1].End())
	return out, nil
}
