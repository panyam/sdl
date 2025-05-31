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
	PrettyPrint(cp CodePrinter)
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

func (me *ExprBase) exprNode() {}

// LazyExpr is a thunk/holder of another expression that is evaluated when needed.
// It is usually reference and is shared
type LazyExpr struct {
	ExprBase
	TargetExpr Expr
	Evaluated  bool
}

// --- Expressions ---
// TupleExpr represents `left operator right`
type TupleExpr struct {
	ExprBase
	Children []Expr
}

func (b *TupleExpr) String() string {
	// Basic, doesn't handle precedence for parentheses
	return fmt.Sprintf("(%s)", strings.Join(gfn.Map(b.Children, func(e Expr) string { return e.String() }), ", "))
}
func (e *TupleExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(e.String())
}

// --- Expressions ---
// BinaryExpr represents `left operator right`
type BinaryExpr struct {
	ExprBase
	Left     Expr
	Operator string // "||", "&&", "==", "!=", "<", "<=", ">", ">=", "+", "-", "*", "/", "%"
	Right    Expr
}

func (b *BinaryExpr) String() string {
	leftStr := "nil"
	if b.Left != nil {
		leftStr = b.Left.String()
	}
	rightStr := "nil"
	if b.Right != nil {
		rightStr = b.Right.String()
	}
	return fmt.Sprintf("(%s %s %s)", leftStr, b.Operator, rightStr)
}
func (e *BinaryExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(e.String())
}

// An expression to create an instance of a component
// InstanceDecls are compiled down to a combination of NewExpr and SetExpr expression types
type NewExpr struct {
	ExprBase
	ComponentExpr *IdentifierExpr
}

func (n *NewExpr) String() string { return fmt.Sprintf("(new %s)", n.ComponentExpr.Value) }
func (n *NewExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(n.String())
}

// UnaryExpr represents `operator operand`
type UnaryExpr struct {
	ExprBase
	Operator string // "!", "-"
	Right    Expr
}

func (u *UnaryExpr) String() string { return fmt.Sprintf("(%s %s)", u.Operator, u.Right) }
func (e *UnaryExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(e.String())
}

// LiteralExpr represents literal values
type LiteralExpr struct {
	ExprBase
	Value Value
	// Duration specific fields could be added if needed after parsing
	// DurationUnit string
	// NumericValue float64
}

func (l *LiteralExpr) String() string {
	return l.Value.String()
}
func (l *LiteralExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(l.String())
}

// IdentifierExpr represents variable or function names
type IdentifierExpr struct {
	ExprBase
	Value string
}

func (i *IdentifierExpr) systemBodyItemNode() {} // Allow bare identifier? Maybe not needed.
func (i *IdentifierExpr) String() string      { return i.Value }
func (e *IdentifierExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(e.String())
}

// MemberAccessExpr represents `receiver.member` (accessing parameters/fields)
type MemberAccessExpr struct {
	ExprBase
	Receiver Expr // The object/instance being accessed
	Member   *IdentifierExpr
}

func (m *MemberAccessExpr) String() string { return fmt.Sprintf("%s.%s", m.Receiver, m.Member) }
func (e *MemberAccessExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(e.String())
}

// CallExpr represents `function(arg1, arg2, ...)` (user funcs, component methods, built-ins)
type CallExpr struct {
	ExprBase
	Function Expr   // Typically IdentifierExpr or MemberAccessExpr
	Args     []Expr // Argument expressions
}

func (c *CallExpr) String() string {
	argsStr := []string{}
	for _, arg := range c.Args {
		argsStr = append(argsStr, arg.String())
	}
	return fmt.Sprintf("%s(%s)", c.Function, strings.Join(argsStr, ", "))
}
func (e *CallExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(e.String())
}

// DistributeExpr represents the probabilistic choice expression/statement
type DistributeExpr struct {
	ExprBase
	TotalProb Expr // Optional total probability expression
	Cases     []*CaseExpr
	Default   Expr // default can only exist if TotalProb is given
}

func (d *DistributeExpr) stmtNode() {} // Can be statement
func (d *DistributeExpr) String() string {
	out := "dist {"
	for _, cse := range d.Cases {
		out += "\n" + cse.String()
	}
	if d.Default != nil {
		out += "\n" + d.Default.String()
	}
	return out
}
func (e *DistributeExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(e.String())
}

// SampleExpr represents `delay durationExpr;`
type SampleExpr struct {
	ExprBase
	FromExpr Expr // Must evaluate to Outcome[X]
}

func (d *SampleExpr) stmtNode()      {}
func (d *SampleExpr) String() string { return fmt.Sprintf("sample %s;", d.FromExpr) }
func (e *SampleExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(e.String())
}

// CaseExpr represents a single case within a SwitchStmt
type CaseExpr struct {
	ExprBase
	Condition Expr
	Body      Expr
}

func (c *CaseExpr) String() string { return fmt.Sprintf("case %s => %s", c.Condition, c.Body) }
func (e *CaseExpr) PrettyPrint(cp CodePrinter) {
	cp.Print(e.String())
}
