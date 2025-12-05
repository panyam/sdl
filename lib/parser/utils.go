package parser

import (
	"log"
	"strconv"
	"strings"

	"github.com/panyam/sdl/lib/decl"
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

	if unit == "s" {
		// base unit is seconds - no conversion needed
	} else if unit == "ms" {
		out /= 1000.0 // Convert milliseconds to seconds
	} else if unit == "us" {
		out /= 1000000.0 // Convert microseconds to seconds
	} else if unit == "ns" {
		out /= 1000000000.0 // Convert nanoseconds to seconds
	}
	return
}

// Helper to combine position info - assumes lexer provides positions.
// We'll need to make the lexer accessible, e.g., via yyParseWithLexer.

// Helper function to create a LiteralExpr node
func NewLiteralExpr(value Value, start, end Location) *LiteralExpr {
	// For string literals, the lexer should provide the raw unquoted value.
	// log.Println("Creating Value with: ", value)
	if value.IsNil() {
		panic("Value shoudl not be nil")
	}
	out := &LiteralExpr{
		ExprBase: ExprBase{NodeInfo: NewNodeInfo(start, end)},
		Value:    value,
	}
	return out
}

// Helper to create NodeInfo from a token's SDLSymType value.
// This is the one that was missing.
func NewNodeInfoFromToken(tokenValue *SDLSymType) NodeInfo {
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
func NewNodeInfoFromStartEndNode(startNode Node, endNode Node) NodeInfo {
	if startNode == nil || endNode == nil {
		// Similar to above, handle nil nodes carefully.
		return NodeInfo{}
	}
	return NodeInfo{StartPos: startNode.Pos(), StopPos: endNode.End()}
}

type TokenNode struct {
	NodeInfo
	Text string
}

var e Expr = &ChainedExpr{}

func NewTokenNode(start, end Location, text string) *TokenNode {
	if strings.TrimSpace(text) == "" {
		panic("TOken is empty")
	}
	return &TokenNode{NewNodeInfo(start, end), text}
}

func (tn *TokenNode) Pos() Location                   { return tn.StartPos }
func (tn *TokenNode) End() Location                   { return tn.StopPos }
func (tn *TokenNode) String() string                  { return tn.Text }    // fmt.Sprintf("Token[%d:%d]", tn.StartPos, tn.StopPos) }
func (tn *TokenNode) PrettyPrint(cp decl.CodePrinter) { cp.Print(tn.Text) } // fmt.Sprintf("Token[%d:%d]", tn.StartPos, tn.StopPos) }
