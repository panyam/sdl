package decl

import (
	"fmt"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

// --- Statements ---

// Stmt represents a statement node (performs an action, controls flow).
type Stmt interface {
	Node
	stmtNode() // Marker method for statements
}

// BlockStmt represents a sequence of statements `{ stmt1; stmt2; ... }`
type BlockStmt struct {
	NodeInfo
	Statements []Stmt
}

func (b *BlockStmt) String() string { return "{ ...statements... }" } // Simplified
func (b *BlockStmt) PrettyPrint(cp CodePrinter) {
	if b != nil && b.Statements != nil {
		if len(b.Statements) == 1 {
			b.Statements[0].PrettyPrint(cp)
		} else {
			cp.Println("{")
			cp.Indent(1)
			for _, stmt := range b.Statements {
				stmt.PrettyPrint(cp)
			}
			cp.Unindent(1)
			cp.Print("}")
		}
	}
}

// ForStmt represents `for expression { stmt }`
type ForStmt struct {
	NodeInfo
	Condition Expr
	Body      Stmt
}

func (l *ForStmt) systemBodyItemNode() {} // Allow let at system level
func (l *ForStmt) String() string {
	return fmt.Sprintf("for %s { %s }", l.Condition.String(), l.Body.String())
}

func (f *ForStmt) PrettyPrint(cp CodePrinter) {
	cp.Print("for ")
	f.Condition.PrettyPrint(cp)
	cp.Println(" {")
	cp.Indent(1)
	f.Body.PrettyPrint(cp)
	cp.Unindent(1)
	cp.Print("}")
}

// LetStmt represents `let var = expr;`
type LetStmt struct {
	NodeInfo
	Variables []*IdentifierExpr
	Value     Expr
}

func (l *LetStmt) systemBodyItemNode() {} // Allow let at system level

func (l *LetStmt) String() string {
	return fmt.Sprintf("let %s = %s;", strings.Join(gfn.Map(l.Variables, func(i *IdentifierExpr) string { return i.String() }), ", "), l.Value)
}

func (l *LetStmt) PrettyPrint(cp CodePrinter) {
	cp.Print(l.String())
}

// SetStmt represents `MemberAccessExpr = value`
type SetStmt struct {
	NodeInfo
	TargetExpr Expr
	Value      Expr
}

func (l *SetStmt) systemBodyItemNode() {} // Allow let at system level

func (l *SetStmt) String() string {
	return fmt.Sprintf("set %s = %s;", l.TargetExpr, l.Value)
}

func (l *SetStmt) PrettyPrint(cp CodePrinter) {
	cp.Print(l.String())
}

// ExprStmt represents an expression used as a statement (e.g., a call)
type ExprStmt struct {
	NodeInfo
	Expression Expr
}

func (e *ExprStmt) String() string { return e.Expression.String() + ";" }

func (e *ExprStmt) PrettyPrint(cp CodePrinter) {
	cp.Print(e.String())
}

// ReturnStmt represents `return expr;`
type ReturnStmt struct {
	NodeInfo
	ReturnValue Expr // Optional? Can be `return;`? Let's require a value for now.
}

func (r *ReturnStmt) String() string { return fmt.Sprintf("return %s;", r.ReturnValue) }
func (r *ReturnStmt) PrettyPrint(cp CodePrinter) {
	cp.Print(r.String())
}

// IfStmt represents `if cond { ... } else { ... }`
type IfStmt struct {
	NodeInfo
	Condition Expr // Must evaluate to bool outcome
	Then      *BlockStmt
	Else      Stmt // Can be another IfStmt or a BlockStmt
}

func (i *IfStmt) String() string { return fmt.Sprintf("if (%s) { ... } else { ... }", i.Condition) }
func (i *IfStmt) PrettyPrint(cp CodePrinter) {
	cp.Print("if ")
	i.Condition.PrettyPrint(cp)
	cp.Print(" {")
	cp.Indent(1)
	i.Then.PrettyPrint(cp)
	cp.Unindent(1)
	if i.Else == nil {
		cp.Print("}")
	} else {
		cp.Println("} else {")
		cp.Indent(1)
		i.Else.PrettyPrint(cp)
		cp.Unindent(1)
		cp.Print("}")
	}
}

// DefaultCase represents `default => { block }`
type DefaultCase struct {
	NodeInfo
	Body Stmt
}

func (d *DefaultCase) String() string { return "default => { ... }" }

// DelayStmt represents `delay durationExpr;`
type DelayStmt struct {
	NodeInfo
	Duration Expr // Must evaluate to Duration outcome
}

func (d *DelayStmt) String() string { return fmt.Sprintf("delay %s;", d.Duration) }

func (d *DelayStmt) PrettyPrint(cp CodePrinter) {
	cp.Print("delay ")
	d.Duration.PrettyPrint(cp)
	cp.Print("")
}

// WaitStmt represents `delay durationExpr;`
type WaitStmt struct {
	NodeInfo
	Idents []*IdentifierExpr // Must evaluate to Duration outcome
}

func (d *WaitStmt) String() string {
	return fmt.Sprintf("wait %s;", strings.Join(gfn.Map(d.Idents, func(i *IdentifierExpr) string { return i.Value }), ", "))
}

func (w *WaitStmt) PrettyPrint(cp CodePrinter) {
	cp.Print(w.String())
}

// ExecutionMode determines sequential or parallel execution.
type ExecutionMode string // Use string for simplicity

const (
	Sequential ExecutionMode = "Sequential"
	Go         ExecutionMode = "Go"
)

// GoStmt represents `parallel { stmt* }`
type GoStmt struct {
	NodeInfo
	VarName *IdentifierExpr
	// Can call a async/parallel on a statement or an expression
	Stmt Stmt
	Expr Expr
}

func (p *GoStmt) String() string { return "go { ... }" }

// LogStmt represents `log "message", expr1, expr2;`
type LogStmt struct {
	NodeInfo
	Args []Expr // First arg often StringLiteral, others are values to log
}

func (l *LogStmt) String() string { return "log ... ;" }

func (l *LogStmt) PrettyPrint(cp CodePrinter) {
	cp.Print("log ")
	for idx, arg := range l.Args {
		if idx > 0 {
			cp.Print(", ")
		}
		arg.PrettyPrint(cp)
	}
}

// ExpectStmt represents `targetMetric operator threshold;` (e.g., `result.P99 < 100ms;`)
type ExpectStmt struct {
	NodeInfo
	Target    *MemberAccessExpr // e.g., result.P99 (requires VM resolution)
	Operator  string            // e.g., "<", ">=", "=="
	Threshold Expr
}

func (e *ExpectStmt) String() string {
	return fmt.Sprintf("%s %s %s;", e.Target, e.Operator, e.Threshold)
}

// AssignmentStmt represents setting a parameter value in an InstanceDecl.
type AssignmentStmt struct {
	NodeInfo
	Var      *IdentifierExpr
	Value    Expr   // The value assigned to the parameter
	IsLet    string // whether this is a let statement
	IsFuture string // whether this is a future
}

func (p *AssignmentStmt) String() string { return fmt.Sprintf("%s = %s", p.Var.Value, p.Value) }
func (p *AssignmentStmt) PrettyPrint(cp CodePrinter) {
	cp.Print(p.String())
}

// SwitchStmt represents the probabilistic choice expression/statement
type SwitchStmt struct {
	NodeInfo
	Expr    Expr
	Cases   []*CaseStmt
	Default Stmt // default can only exist if TotalProb is given
}

func (s *SwitchStmt) exprNode()      {} // Can be expression
func (s *SwitchStmt) String() string { return "switch {...}" }
func (s *SwitchStmt) PrettyPrint(cp CodePrinter) {
	cp.Print("switch ")
	s.Expr.PrettyPrint(cp)
	cp.Println(" {")
	cp.Indent(1)
	for _, cse := range s.Cases {
		cse.PrettyPrint(cp)
		cp.Println("")
	}
	if s.Default != nil {
		s.Default.PrettyPrint(cp)
		cp.Println("")
	}
	cp.Unindent(1)
	cp.Print("}")
}

// CaseStmt represents a single case within a SwitchStmt
type CaseStmt struct {
	NodeInfo
	Condition Expr
	Body      Stmt
}

func (c *CaseStmt) exprNode()      {}
func (c *CaseStmt) String() string { return fmt.Sprintf("case %s: %s", c.Condition, c.Body) }
func (c *CaseStmt) PrettyPrint(cp CodePrinter) {
	cp.Print("case ")
	c.Condition.PrettyPrint(cp)
	cp.Print(" => ")
	c.Body.PrettyPrint(cp)
}
