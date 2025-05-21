package decl

import (
	"fmt"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

// Expr represents an expression node (evaluates to a value/state).
type Expr interface {
	Node
	exprNode() // Marker method for expressions
	InferredType() *Type
	DeclaredType() *Type
	SetInferredType(*Type)
	SetDeclaredType(*Type)
}

type ExprBase struct {
	NodeInfo
	declaredType *Type
	inferredType *Type
}

func (e *ExprBase) SetInferredType(t *Type) {
	e.inferredType = t
}

func (e *ExprBase) InferredType() *Type {
	return e.inferredType
}

func (e *ExprBase) SetDeclaredType(t *Type) {
	e.declaredType = t
}

func (e *ExprBase) DeclaredType() *Type {
	return e.declaredType
}

type ChainedExpr struct {
	ExprBase
	Children  []Expr
	Operators []string

	// Expression after operators have been taken into account
	UnchainedExpr Expr
}

func (b *ChainedExpr) exprNode() {}
func (b *ChainedExpr) stmtNode() {}
func (b *ChainedExpr) String() string {
	// Basic, doesn't handle precedence for parentheses
	return fmt.Sprintf("(%s)", strings.Join(gfn.Map(b.Children, func(e Expr) string { return e.String() }), ", "))
}

// --- Expressions ---
// TupleExpr represents `left operator right`
type TupleExpr struct {
	ExprBase
	Children []Expr
}

func (b *TupleExpr) exprNode() {}
func (b *TupleExpr) String() string {
	// Basic, doesn't handle precedence for parentheses
	return fmt.Sprintf("(%s)", strings.Join(gfn.Map(b.Children, func(e Expr) string { return e.String() }), ", "))
}

// --- Expressions ---
// BinaryExpr represents `left operator right`
type BinaryExpr struct {
	ExprBase
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
	ExprBase
	Operator string // "!", "-"
	Right    Expr
}

func (u *UnaryExpr) exprNode()      {}
func (u *UnaryExpr) String() string { return fmt.Sprintf("(%s%s)", u.Operator, u.Right) }

// LiteralExpr represents literal values
type LiteralExpr struct {
	ExprBase
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
	ExprBase
	Name string
}

func (i *IdentifierExpr) exprNode()           {}
func (i *IdentifierExpr) systemBodyItemNode() {} // Allow bare identifier? Maybe not needed.
func (i *IdentifierExpr) String() string      { return i.Name }

// MemberAccessExpr represents `receiver.member` (accessing parameters/fields)
type MemberAccessExpr struct {
	ExprBase
	Receiver Expr // The object/instance being accessed
	Member   *IdentifierExpr
}

func (m *MemberAccessExpr) exprNode()      {}
func (m *MemberAccessExpr) String() string { return fmt.Sprintf("%s.%s", m.Receiver, m.Member) }

// CallExpr represents `function(arg1, arg2, ...)` (user funcs, component methods, built-ins)
type CallExpr struct {
	ExprBase
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
	ExprBase
	TotalProb Expr // Optional total probability expression
	Cases     []*CaseExpr
	Default   Expr // default can only exist if TotalProb is given
}

func (d *DistributeExpr) exprNode()      {} // Can be expression
func (d *DistributeExpr) stmtNode()      {} // Can be statement
func (d *DistributeExpr) String() string { return "distribute {...}" }

// SampleExpr represents `delay durationExpr;`
type SampleExpr struct {
	ExprBase
	FromExpr Expr // Must evaluate to Outcome[X]
}

func (d *SampleExpr) stmtNode()      {}
func (d *SampleExpr) exprNode()      {}
func (d *SampleExpr) String() string { return fmt.Sprintf("sample %s;", d.FromExpr) }

// CaseExpr represents a single case within a SwitchStmt
type CaseExpr struct {
	ExprBase
	Condition Expr
	Body      Expr
}

func (c *CaseExpr) exprNode()      {}
func (c *CaseExpr) String() string { return fmt.Sprintf("case %s: %s", c.Condition, c.Body) }
