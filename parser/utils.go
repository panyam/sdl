package parser

import (
	"log"
	"strconv"
	"strings"

	"github.com/panyam/sdl/decl"
)

func parseDuration(numText, unit string) (out float64) {
	numText = strings.TrimSpace(numText)
	if v, err := strconv.ParseInt(numText, 10, 64); err == nil {
		out = float64(v)
	} else if v, err := strconv.ParseFloat(numText, 64); err == nil {
		out = v
	} else {
		log.Println("Invalid duration: ", numText, unit)
	}

	if unit == "ms" {
		// base unit is ms
	} else if unit == "s" {
		out *= 1000.0
	} else if unit == "us" {
		out /= 1000
	} else if unit == "ns" {
		out /= 1000000
	}
	return
}

// Helper to combine position info - assumes lexer provides positions.
// We'll need to make the lexer accessible, e.g., via yyParseWithLexer.
func newNodeInfo(start, end Location) NodeInfo {
	return NodeInfo{StartPos: start, StopPos: end}
}

// Helper function to create a LiteralExpr node
func newLiteralExpr(value Value, start, end Location) *LiteralExpr {
	// For string literals, the lexer should provide the raw unquoted value.
	// log.Println("Creating Value with: ", value)
	if value.IsNil() {
		panic("Value shoudl not be nil")
	}
	out := &LiteralExpr{
		ExprBase: ExprBase{NodeInfo: newNodeInfo(start, end)},
		Value:    value,
	}
	return out
}

// Helper function to create an IdentifierExpr node
func newIdentifierExpr(name string, start, end Location) *IdentifierExpr {
	return &IdentifierExpr{
		ExprBase: ExprBase{NodeInfo: newNodeInfo(start, end)},
		Name:     name,
	}
}

// Helper to create NodeInfo from a token's SDLSymType value.
// This is the one that was missing.
func newNodeInfoFromToken(tokenValue *SDLSymType) NodeInfo {
	if tokenValue == nil || tokenValue.node == nil {
		// This case should ideally be an error or handled carefully.
		// Returning zero NodeInfo might hide parsing issues.
		// For robustness, you might want to log or panic here if tokenValue or tokenValue.node is nil,
		// as it indicates an issue in how the lexer or parser is handling token values.
		// For now, returning an empty NodeInfo to avoid panics during development.
		return NodeInfo{}
	}
	return NodeInfo{StartPos: tokenValue.node.Pos(), StopPos: tokenValue.node.End()}
}

// Helper to create NodeInfo spanning from a start Node to an end Node.
func newNodeInfoFromStartEndNode(startNode Node, endNode Node) NodeInfo {
	if startNode == nil || endNode == nil {
		// Similar to above, handle nil nodes carefully.
		return NodeInfo{}
	}
	return NodeInfo{StartPos: startNode.Pos(), StopPos: endNode.End()}
}

func newIdent(name string) *IdentifierExpr {
	// NodeInfo will be zeroed out by cleanNodeInfo for comparison
	return &IdentifierExpr{Name: name}
}

func newIntLit(val int64) *LiteralExpr {
	return &LiteralExpr{Value: IntValue(val)}
}
func newStringLit(val string) *LiteralExpr { // val is the content without quotes
	return &LiteralExpr{Value: StringValue(val)}
}
func newBoolLit(val bool) *LiteralExpr {
	return &LiteralExpr{Value: BoolValue(val)}
}

func newBinaryExpr(left Expr, op string, right Expr) *BinaryExpr {
	if strings.TrimSpace(op) == "" {
		panic("Invalid op")
	}
	return &BinaryExpr{Left: left, Operator: op, Right: right}
}

func newUnaryExpr(op string, right Expr) *UnaryExpr {
	return &UnaryExpr{Operator: op, Right: right}
}

func newLetStmt(varName string, value Expr) *LetStmt {
	return &LetStmt{Variables: []*IdentifierExpr{newIdent(varName)}, Value: value}
}

func newBlockStmt(stmts ...Stmt) *BlockStmt {
	return &BlockStmt{Statements: stmts}
}

func newCallExpr(fn Expr, args ...Expr) *CallExpr {
	// NodeInfo is cleaned by cleanNodeInfo
	return &CallExpr{Function: fn, Args: args}
}

func newReturnStmt(returnValue Expr) *ReturnStmt {
	// NodeInfo is cleaned
	return &ReturnStmt{ReturnValue: returnValue}
}

func newExprStmt(expr Expr) *ExprStmt {
	return &ExprStmt{Expression: expr}
}

func newComponentDecl(name string, isNative bool, body ...ComponentDeclBodyItem) *ComponentDecl {
	return &ComponentDecl{NameNode: newIdent(name), IsNative: isNative, Body: body}
}

func newSystemDecl(name string, body ...SystemDeclBodyItem) *SystemDecl {
	return &SystemDecl{NameNode: newIdent(name), Body: body}
}

func newUsesDecl(localName, componentType string) *UsesDecl {
	return &UsesDecl{NameNode: newIdent(localName), ComponentNode: newIdent(componentType)}
}

func newMethodDecl(name string, returnTypeDecl *TypeDecl, body *BlockStmt, params ...*ParamDecl) *MethodDecl {
	// NodeInfo is cleaned
	if params == nil {
		params = []*ParamDecl{}
	}
	return &MethodDecl{NameNode: newIdent(name), Parameters: params, ReturnType: returnTypeDecl, Body: body}
}

func newParamDecl(name string, typeName *TypeDecl, defaultValue Expr) *ParamDecl {
	// NodeInfo is cleaned
	return &ParamDecl{Name: newIdent(name), Type: typeName, DefaultValue: defaultValue}
}

func newTypeDecl(name string, args []*TypeDecl) *TypeDecl {
	// NodeInfo is cleaned
	return &TypeDecl{Name: name, Args: args}
}

func newInstanceDecl(instanceName, componentType string, overrides ...*AssignmentStmt) *InstanceDecl {
	return &InstanceDecl{
		NameNode:      newIdent(instanceName),
		ComponentType: newIdent(componentType),
		Overrides:     overrides,
	}
}

func newAssignmentStmt(varName string, value Expr) *AssignmentStmt {
	return &AssignmentStmt{Var: newIdent(varName), Value: value}
}

func newOptionsDecl(body *BlockStmt) *OptionsDecl {
	return &OptionsDecl{Body: body}
}

func newDistributeExpr(total Expr, defaultCase Expr, cases ...*CaseExpr) *DistributeExpr {
	return &DistributeExpr{TotalProb: total, Cases: cases, Default: defaultCase}
}

func newGoStmt(varName *IdentifierExpr, stmt Stmt, expr Expr) *GoStmt {
	return &GoStmt{VarName: varName, Stmt: stmt, Expr: expr}
}

func newDelayStmt(duration Expr) *DelayStmt {
	return &DelayStmt{Duration: duration}
}

func newWaitStmt(idents ...*IdentifierExpr) *WaitStmt {
	return &WaitStmt{Idents: idents}
}

type TokenNode struct {
	NodeInfo
	Text string
}

var e Expr = &ChainedExpr{}

func newTokenNode(start, end Location, text string) *TokenNode {
	if strings.TrimSpace(text) == "" {
		panic("TOken is empty")
	}
	return &TokenNode{newNodeInfo(start, end), text}
}

func (tn *TokenNode) Pos() Location                   { return tn.StartPos }
func (tn *TokenNode) End() Location                   { return tn.StopPos }
func (tn *TokenNode) String() string                  { return tn.Text }    // fmt.Sprintf("Token[%d:%d]", tn.StartPos, tn.StopPos) }
func (tn *TokenNode) PrettyPrint(cp decl.CodePrinter) { cp.Print(tn.Text) } // fmt.Sprintf("Token[%d:%d]", tn.StartPos, tn.StopPos) }
