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

func (b *BlockStmt) stmtNode()      {}
func (b *BlockStmt) String() string { return "{ ...statements... }" } // Simplified

// LetStmt represents `let var = expr;`
type LetStmt struct {
	NodeInfo
	Variable *IdentifierExpr
	Value    Expr
}

func (l *LetStmt) stmtNode()           {}
func (l *LetStmt) systemBodyItemNode() {} // Allow let at system level
func (l *LetStmt) String() string      { return fmt.Sprintf("let %s = %s;", l.Variable, l.Value) }

// ExprStmt represents an expression used as a statement (e.g., a call)
type ExprStmt struct {
	NodeInfo
	Expression Expr
}

func (e *ExprStmt) stmtNode()      {}
func (e *ExprStmt) String() string { return e.Expression.String() + ";" }

// ReturnStmt represents `return expr;`
type ReturnStmt struct {
	NodeInfo
	ReturnValue Expr // Optional? Can be `return;`? Let's require a value for now.
}

func (r *ReturnStmt) stmtNode()      {}
func (r *ReturnStmt) String() string { return fmt.Sprintf("return %s;", r.ReturnValue) }

// IfStmt represents `if cond { ... } else { ... }`
type IfStmt struct {
	NodeInfo
	Condition Expr // Must evaluate to bool outcome
	Then      *BlockStmt
	Else      Stmt // Can be another IfStmt or a BlockStmt
}

func (i *IfStmt) stmtNode()      {}
func (i *IfStmt) String() string { return fmt.Sprintf("if (%s) { ... } else { ... }", i.Condition) }

// DistributeStmt represents probabilistic control flow
type DistributeStmt struct {
	NodeInfo
	Total       Expr // Optional total probability expression (float outcome)
	Cases       []*DistributeCase
	DefaultCase *DefaultCase // Optional
}

func (d *DistributeStmt) stmtNode()      {}
func (d *DistributeStmt) String() string { return "distribute { ... }" }

// DistributeCase represents `probExpr => { block }`
type DistributeCase struct {
	NodeInfo
	Probability Expr // Must evaluate to float outcome
	Body        *BlockStmt
}

func (d *DistributeCase) String() string { return fmt.Sprintf("%s => { ... }", d.Probability) }

// DefaultCase represents `default => { block }`
type DefaultCase struct {
	NodeInfo
	Body *BlockStmt
}

func (d *DefaultCase) String() string { return "default => { ... }" }

// DelayStmt represents `delay durationExpr;`
type DelayStmt struct {
	NodeInfo
	Duration Expr // Must evaluate to Duration outcome
}

func (d *DelayStmt) stmtNode()      {}
func (d *DelayStmt) String() string { return fmt.Sprintf("delay %s;", d.Duration) }

// WaitStmt represents `delay durationExpr;`
type WaitStmt struct {
	NodeInfo
	Idents []*IdentifierExpr // Must evaluate to Duration outcome
}

func (d *WaitStmt) stmtNode() {}
func (d *WaitStmt) String() string {
	return fmt.Sprintf("wait %s;", strings.Join(gfn.Map(d.Idents, func(i *IdentifierExpr) string { return i.Name }), ", "))
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
	Stmt *BlockStmt
	Expr *Expr
}

func (p *GoStmt) stmtNode()      {}
func (p *GoStmt) String() string { return "go { ... }" }

// LogStmt represents `log "message", expr1, expr2;`
type LogStmt struct {
	NodeInfo
	Args []Expr // First arg often StringLiteral, others are values to log
}

func (l *LogStmt) stmtNode()      {}
func (l *LogStmt) String() string { return "log ... ;" }

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

// SwitchStmt represents conditional branching
type SwitchStmt struct {
	NodeInfo
	Input   Expr
	Cases   []*CaseExpr /* ; Default *BlockExpr */
	Default Expr
}

func (s *SwitchStmt) exprNode() {}
func (s *SwitchStmt) String() string { /* Basic string representation */
	return fmt.Sprintf("switch(%s){...}", s.Input)
}

// CaseExpr represents a single case within a SwitchStmt
type CaseExpr struct {
	NodeInfo
	Condition Expr
	Body      Expr
}

func (c *CaseExpr) exprNode()      {}
func (c *CaseExpr) String() string { return fmt.Sprintf("case %s: %s", c.Condition, c.Body) }

// AssignmentStmt represents setting a parameter value in an InstanceDecl.
type AssignmentStmt struct {
	NodeInfo
	Var      *IdentifierExpr
	Value    Expr   // The value assigned to the parameter
	IsLet    string // whether this is a let statement
	IsFuture string // whether this is a future
}

func (p *AssignmentStmt) String() string { return fmt.Sprintf("%s = %s", p.Var.Name, p.Value) }
