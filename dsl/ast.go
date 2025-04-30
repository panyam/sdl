// sdl/decl/ast/ast.go
package dsl

import (
	"fmt"
	"strings"
)

// Node interface
type Node interface {
	Pos() int
	End() int
	String() string
}

// NodeInfo embeddable struct
type NodeInfo struct{ StartPos, StopPos int }

func (n *NodeInfo) Pos() int       { return n.StartPos }
func (n *NodeInfo) End() int       { return n.StopPos }
func (n *NodeInfo) String() string { return "{Node}" }

// Expr interface
type Expr interface {
	Node
	exprNode()
}

// LiteralExpr struct and String()
type LiteralExpr struct {
	NodeInfo
	Kind, Value string
}

func (l *LiteralExpr) exprNode() {}
func (l *LiteralExpr) String() string {
	if l.Kind == "STRING" {
		return fmt.Sprintf(`"%s"`, l.Value)
	}
	return l.Value
}

// IdentifierExpr struct and String()
type IdentifierExpr struct {
	NodeInfo
	Name string
}

func (i *IdentifierExpr) exprNode()      {}
func (i *IdentifierExpr) String() string { return i.Name }

// MemberAccessExpr struct and String()
type MemberAccessExpr struct {
	NodeInfo
	Receiver Expr
	Member   string
}

func (m *MemberAccessExpr) exprNode()      {}
func (m *MemberAccessExpr) String() string { return fmt.Sprintf("%s.%s", m.Receiver, m.Member) }

// CallExpr struct and String()
type CallExpr struct {
	NodeInfo
	Function Expr
	Args     []Expr
}

func (c *CallExpr) exprNode() {}
func (c *CallExpr) String() string { /* ... implementation ... */
	argsStr := []string{}
	for _, arg := range c.Args {
		argsStr = append(argsStr, arg.String())
	}
	return fmt.Sprintf("%s(%s)", c.Function, strings.Join(argsStr, ", "))
}

// AndExpr struct and String()
type AndExpr struct {
	NodeInfo
	Left, Right Expr
}

func (a *AndExpr) exprNode()      {}
func (a *AndExpr) String() string { return fmt.Sprintf("(%s THEN %s)", a.Left, a.Right) }

// ParallelExpr struct and String()
type ParallelExpr struct {
	NodeInfo
	Left, Right Expr
}

func (p *ParallelExpr) exprNode()      {}
func (p *ParallelExpr) String() string { return fmt.Sprintf("(%s || %s)", p.Left, p.Right) }

// InternalCallExpr struct and String()
type InternalCallExpr struct {
	NodeInfo
	FuncName string
	Args     []Expr
}

func (ic *InternalCallExpr) exprNode() {}
func (ic *InternalCallExpr) String() string { /* ... implementation ... */
	argsStr := []string{}
	for _, arg := range ic.Args {
		argsStr = append(argsStr, arg.String())
	}
	return fmt.Sprintf("Internal.%s(%s)", ic.FuncName, strings.Join(argsStr, ", "))
}

// --- Added missing nodes from previous discussion ---

// SwitchExpr represents conditional branching
type SwitchExpr struct {
	NodeInfo
	Input Expr
	Cases []*CaseExpr /* ; Default *BlockExpr */
}

func (s *SwitchExpr) exprNode() {}
func (s *SwitchExpr) String() string { /* Basic string representation */
	return fmt.Sprintf("switch(%s){...}", s.Input)
}

// CaseExpr represents a single case within a SwitchExpr
type CaseExpr struct {
	NodeInfo
	Condition Expr
	Body      Expr
}

func (c *CaseExpr) exprNode()      {}
func (c *CaseExpr) String() string { return fmt.Sprintf("case %s: %s", c.Condition, c.Body) }

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
func (f *FilterParams) String() string { /* Basic string representation */ return "{FilterParams...}" }

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

type ExecutionMode int

const (
	Sequential ExecutionMode = iota
	Parallel
)

func (e ExecutionMode) String() string {
	if e == Parallel {
		return "Parallel"
	}
	return "Sequential"
}
