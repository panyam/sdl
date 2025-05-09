package parser

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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
		out /= 1000.0
	} else if unit == "us" {
		out *= 1000
	} else if unit == "ns" {
		out *= 1000000
	}
	return
}

// Helper to combine position info - assumes lexer provides positions.
// We'll need to make the lexer accessible, e.g., via yyParseWithLexer.
func newNodeInfo(startPos, endPos int) NodeInfo {
	return NodeInfo{StartPos: startPos, StopPos: endPos}
}

// Helper function to create a LiteralExpr node
func newLiteralExpr(value Value, startPos, endPos int) *LiteralExpr {
	// For string literals, the lexer should provide the raw unquoted value.
	out := &LiteralExpr{
		NodeInfo: newNodeInfo(startPos, endPos),
		Value:    value,
	}
	return out
}

// Helper function to create an IdentifierExpr node
func newIdentifierExpr(name string, startPos, endPos int) *IdentifierExpr {
	return &IdentifierExpr{
		NodeInfo: newNodeInfo(startPos, endPos),
		Name:     name,
	}
}

// Helper to create NodeInfo from a token's yySymType value.
// This is the one that was missing.
func newNodeInfoFromToken(tokenValue *yySymType) NodeInfo {
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

type TokenNode struct {
	NodeInfo
	Text string
}

func (tn *TokenNode) Pos() int       { return tn.StartPos }
func (tn *TokenNode) End() int       { return tn.StopPos }
func (tn *TokenNode) String() string { return fmt.Sprintf("Token[%d:%d]", tn.StartPos, tn.StopPos) }
func (tn *TokenNode) exprNode()      {} // If needed to satisfy Expr for some rules
func (tn *TokenNode) stmtNode()      {} // If needed to satisfy Stmt
