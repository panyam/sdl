package decl

import (
	"fmt"
	"strconv"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

// --- Interfaces ---

// Node represents any node in the Abstract Syntax Tree.
type Node interface {
	Pos() int       // Starting position (for error reporting)
	End() int       // Ending position
	String() string // String representation for debugging/printing
}

// Expr represents an expression node (evaluates to a value/state).
type Expr interface {
	Node
	exprNode() // Marker method for expressions
}

// Stmt represents a statement node (performs an action, controls flow).
type Stmt interface {
	Node
	stmtNode() // Marker method for statements
}

// --- Base Struct ---

// NodeInfo embeddable struct for position tracking.
type NodeInfo struct{ StartPos, StopPos int }

func (n *NodeInfo) Pos() int       { return n.StartPos }
func (n *NodeInfo) End() int       { return n.StopPos }
func (n *NodeInfo) String() string { return "{Node}" } // Default stringer

// --- Top Level declarations ---

// File represents the top-level node of a parsed DSL file.
type File struct {
	NodeInfo
	declarations []Node // ComponentDecl, SystemDecl, Options, Enum, Import
}

func (f *File) String() string {
	lines := []string{}
	for _, d := range f.declarations {
		lines = append(lines, d.String())
	}
	return strings.Join(lines, "\n")
}

// Options represents `options { ... }` (structure TBD)
type Options struct {
	NodeInfo
	Body *BlockStmt // Placeholder for options assignments?
}

func (o *Options) systemBodyItemNode() {}
func (o *Options) String() string      { return "options { ... }" }

// Enum represents `enum Name { Val1, Val2, ... };`
type Enum struct {
	NodeInfo
	Name   *IdentifierExpr   // Enum type name
	Values []*IdentifierExpr // List of enum value names
}

func (e *Enum) String() string {
	vals := []string{}
	for _, v := range e.Values {
		vals = append(vals, v.Name)
	}
	return fmt.Sprintf("enum %s { %s };", e.Name, strings.Join(vals, ", "))
}

// Import represents `import "path";`
type Import struct {
	NodeInfo
	Path *LiteralExpr // Should be a STRING literal
}

func (i *Import) String() string { return fmt.Sprintf("import %s;", i.Path) }

// --- ComponentDecl Definition ---

// ComponentDecl represents `component Name { ... }`
type ComponentDecl struct {
	NodeInfo
	Name *IdentifierExpr         // ComponentDecl type name
	Body []ComponentDeclBodyItem // ParamDecl, UsesDecl, MethodDef
}

func (c *ComponentDecl) String() string         { return fmt.Sprintf("component %s { ... }", c.Name) }
func (c *ComponentDecl) componentBodyItemNode() {}

// ComponentDeclBodyItem marker interface for items allowed in ComponentDecl body.
type ComponentDeclBodyItem interface {
	Node
	componentBodyItemNode()
}

// TypeName represents primitive types or registered enum identifiers.
type TypeName struct {
	NodeInfo
	// Can be one of the following
	PrimitiveTypeName string // "int", "float", "bool", "string", "duration", or an Enum identifier
	EnumTypeName      string
	OutcomeTypeName   string
}

func (t *TypeName) Name() string {
	if t.PrimitiveTypeName != "" {
		return t.PrimitiveTypeName
	}
	if t.EnumTypeName != "" {
		return t.EnumTypeName
	}
	if t.OutcomeTypeName != "" {
		return t.OutcomeTypeName
	}
	return ""
}

func (t *TypeName) String() string { return t.Name() }

// ParamDecl represents `param name: TypeName [= defaultExpr];`
type ParamDecl struct {
	NodeInfo
	Name         *IdentifierExpr
	Type         *TypeName
	DefaultValue Expr // Optional
}

func (p *ParamDecl) componentBodyItemNode() {}
func (p *ParamDecl) String() string {
	s := fmt.Sprintf("param %s: %s", p.Name, p.Type)
	if p.DefaultValue != nil {
		s += fmt.Sprintf(" = %s", p.DefaultValue)
	}
	return s + ";"
}

// UsesDecl represents `uses varName: ComponentType [{ overrides }];`
type UsesDecl struct {
	NodeInfo
	Name          *IdentifierExpr
	ComponentType *IdentifierExpr // Type name of the dependency
}

func (u *UsesDecl) componentBodyItemNode() {}
func (u *UsesDecl) String() string         { return fmt.Sprintf("uses %s: %s;", u.Name, u.ComponentType) }

// MethodDef represents `method name(params) [: returnType] { body }`
type MethodDef struct {
	NodeInfo
	Name       *IdentifierExpr
	Parameters []*ParamDecl // Signature parameters (can be empty)
	ReturnType *TypeName    // Optional return type (primitive or enum)
	Body       *BlockStmt
}

func (o *MethodDef) componentBodyItemNode() {}
func (o *MethodDef) String() string {
	retType := ""
	if o.ReturnType != nil {
		retType = fmt.Sprintf(": %s", o.ReturnType)
	}
	return fmt.Sprintf("method %s(...) %s { ... }", o.Name, retType)
}

// --- SystemDecl Definition ---

// SystemDecl represents `system Name { ... }`
type SystemDecl struct {
	NodeInfo
	Name *IdentifierExpr
	Body []SystemDeclBodyItem // InstanceDecl, Analyze, Options, LetStmt
}

func (s *SystemDecl) String() string { return fmt.Sprintf("system %s { ... }", s.Name) }

// SystemDeclBodyItem marker interface for items allowed in SystemDecl body.
type SystemDeclBodyItem interface {
	Node
	systemBodyItemNode()
}

// InstanceDecl represents `instanceName: ComponentType = { overrides };`
type InstanceDecl struct {
	NodeInfo
	Name          *IdentifierExpr
	ComponentType *IdentifierExpr
	Overrides     []*AssignmentStmt // Changed from Params
}

func (i *InstanceDecl) systemBodyItemNode() {}
func (i *InstanceDecl) String() string {
	return fmt.Sprintf("instance %s: %s = { ... };", i.Name, i.ComponentType)
}

// Analyze represents `analyze name = callExpr expect { ... };`
type Analyze struct {
	NodeInfo
	Name         *IdentifierExpr
	Target       *CallExpr    // Changed from Expr
	Expectations *ExpectBlock // Optional
}

func (a *Analyze) systemBodyItemNode() {}
func (a *Analyze) String() string {
	s := fmt.Sprintf("analyze %s = %s", a.Name, a.Target)
	if a.Expectations != nil {
		s += " " + a.Expectations.String()
	}
	return s + ";"
}

// ExpectBlock represents `expect { expectStmt* }`
type ExpectBlock struct {
	NodeInfo
	Expects []*ExpectStmt
}

func (e *ExpectBlock) String() string { return "expect { ... }" }

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

// --- Statements ---

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
	Kind  string // "INT", "FLOAT", "STRING", "BOOL", "DURATION"
	Value string // Raw string value from source
	// Duration specific fields could be added if needed after parsing
	// DurationUnit string
	// NumericValue float64
}

func (l *LiteralExpr) exprNode() {}
func (l *LiteralExpr) String() string {
	if l.Kind == "STRING" {
		// Use strconv.Quote for proper escaping if Value contains raw string content
		return strconv.Quote(l.Value)
	}
	if l.Kind == "DURATION" {
		// Assuming Value includes the unit, e.g., "10ms"
		return l.Value
	}
	return l.Value
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

/** Disable Filter and Repeat for now
// FilterExpr represents filtering buckets
type FilterExpr struct {
	NodeInfo
	Input  Expr
	Filter *FilterParams
}

func (f *FilterExpr) exprNode()      {}
func (f *FilterExpr) String() string { return fmt.Sprintf("filter(%s, %s)", f.Input, f.Filter) }

// FilterParams holds predefined filter criteria
type FilterParams struct {
	NodeInfo
	BySuccess              *bool
	MinLatency, MaxLatency Expr
}

func (f *FilterParams) exprNode()      {}
func (f *FilterParams) String() string {
	// Basic string representation
	return "{FilterParams...}"
}

// RepeatExpr represents Op repeated N times
type RepeatExpr struct {
	NodeInfo
	Input Expr
	Count Expr
	Mode  ExecutionMode
}

func (r *RepeatExpr) exprNode() {}
func (r *RepeatExpr) String() string {
	return fmt.Sprintf("repeat(%s, %s, %s)", r.Input, r.Count, r.Mode)
}

// FanoutExpr represents Op fanned out based on count distribution
type FanoutExpr struct {
	NodeInfo
	CountDist Expr
	OpExpr    Expr
	Mode      ExecutionMode
}

func (f *FanoutExpr) exprNode() {}
func (f *FanoutExpr) String() string {
	return fmt.Sprintf("fanout(%s, %s, %s)", f.CountDist, f.OpExpr, f.Mode)
}
*/

// --- Statement Nodes ---

// AssignmentStmt represents setting a parameter value in an InstanceDecl.
type AssignmentStmt struct {
	NodeInfo
	Var      *IdentifierExpr
	Value    Expr   // The value assigned to the parameter
	IsLet    string // whether this is a let statement
	IsFuture string // whether this is a future
}

func (p *AssignmentStmt) String() string { return fmt.Sprintf("%s = %s", p.Var.Name, p.Value) }

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
