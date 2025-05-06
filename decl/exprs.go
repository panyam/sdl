package decl

import (
	"fmt"
	"strings"
)

// Expr represents an expression node (evaluates to a value/state).
type Expr interface {
	Node
	exprNode() // Marker method for expressions
}

// --- Expressions ---
// BinaryExpr represents `left operator right`
type BinaryExpr struct {
	NodeInfo
	Left     Expr
	Operator string // "||", "&&", "==", "!=", "<", "<=", ">", ">=", "+", "-", "*", "/", "%"
	Right    Expr
}

func (b *BinaryExpr) exprNode() {}
func (b *BinaryExpr) String() string {
	// Basic, doesn't handle precedence for parentheses
	return fmt.Sprintf("(%s %s %s)", b.Left, b.Operator, b.Right)
}

// UnaryExpr represents `operator operand`
type UnaryExpr struct {
	NodeInfo
	Operator string // "!", "-"
	Right    Expr
}

func (u *UnaryExpr) exprNode()      {}
func (u *UnaryExpr) String() string { return fmt.Sprintf("(%s%s)", u.Operator, u.Right) }

// LiteralExpr represents literal values
type LiteralExpr struct {
	NodeInfo
	Value Value
	// Duration specific fields could be added if needed after parsing
	// DurationUnit string
	// NumericValue float64
}

func (l *LiteralExpr) exprNode() {}
func (l *LiteralExpr) String() string {
	return l.Value.String()
}

// IdentifierExpr represents variable or function names
type IdentifierExpr struct {
	NodeInfo
	Name string
}

func (i *IdentifierExpr) exprNode()           {}
func (i *IdentifierExpr) systemBodyItemNode() {} // Allow bare identifier? Maybe not needed.
func (i *IdentifierExpr) String() string      { return i.Name }

// MemberAccessExpr represents `receiver.member` (accessing parameters/fields)
type MemberAccessExpr struct {
	NodeInfo
	Receiver Expr // The object/instance being accessed
	Member   *IdentifierExpr
}

func (m *MemberAccessExpr) exprNode()      {}
func (m *MemberAccessExpr) String() string { return fmt.Sprintf("%s.%s", m.Receiver, m.Member) }

// CallExpr represents `function(arg1, arg2, ...)` (user funcs, component methods, built-ins)
type CallExpr struct {
	NodeInfo
	Function Expr   // Typically IdentifierExpr or MemberAccessExpr
	Args     []Expr // Argument expressions
}

func (c *CallExpr) exprNode() {}
func (c *CallExpr) String() string {
	argsStr := []string{}
	for _, arg := range c.Args {
		argsStr = append(argsStr, arg.String())
	}
	return fmt.Sprintf("%s(%s)", c.Function, strings.Join(argsStr, ", "))
}

// DistributeExpr represents the probabilistic choice expression/statement
type DistributeExpr struct {
	NodeInfo
	TotalProb Expr // Optional total probability expression
	Cases     []*DistributeExprCase
	Default   Expr
}

func (d *DistributeExpr) exprNode()      {} // Can be expression
func (d *DistributeExpr) stmtNode()      {} // Can be statement
func (d *DistributeExpr) String() string { return "distribute {...}" }

// DistributeExprnCase represents `probExpr => { block }`
type DistributeExprCase struct {
	NodeInfo
	Probability Expr // Must evaluate to float outcome
	Body        Expr
}

func (d *DistributeExprCase) String() string { return fmt.Sprintf("%s => { ... }", d.Probability) }
