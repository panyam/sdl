package decl

import (
	"fmt"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

// --- Interfaces ---

type Location struct {
	Pos  int
	Line int
	Col  int
}

func (l Location) LineColStr() string {
	return fmt.Sprintf("Line %d, Col %d", l.Line, l.Col)
}

// Node represents any node in the Abstract Syntax Tree.
type Node interface {
	Pos() Location  // Starting position (for error reporting)
	End() Location  // Ending position
	String() string // String representation for debugging/printing
	PrettyPrint(cp CodePrinter)
}

// Annotations are tings that can be added to other constructs via @ syntax
type Annotation struct {
	Node
	Key *IdentifierExpr
}

// --- Base Struct ---

// NodeInfo embeddable struct for position tracking.
type NodeInfo struct{ StartPos, StopPos Location }

func (n *NodeInfo) Pos() Location { return n.StartPos }
func (n *NodeInfo) End() Location { return n.StopPos }
func (n *NodeInfo) stmtNode()     {}

// --- Top Level declarations ---

// OptionsDecl represents `options { ... }` (structure TBD)
type OptionsDecl struct {
	NodeInfo
	Body *BlockStmt // Placeholder for options assignments?
}

func (o *OptionsDecl) systemBodyItemNode()        {}
func (o *OptionsDecl) String() string             { return "options { ... }" }
func (o *OptionsDecl) PrettyPrint(cp CodePrinter) { cp.Println("options { ... }") }

// EnumDecl represents `enum Name { Val1, Val2, ... };`
type EnumDecl struct {
	NodeInfo
	Name   *IdentifierExpr   // EnumDecl type name
	Values []*IdentifierExpr // List of enum value names

	// Resolved values so we can work with processed/loaded values instead of resolving
	// Identify expressions etc
	values []string
}

func (d *EnumDecl) IndexOfVariant(variant string) int {
	// TODO - save this
	for idx, v := range d.Values {
		if v.Value == variant {
			return idx
		}
	}
	return -1 // Not found
}

func (d *EnumDecl) Variants() []string {
	// TODO - save this
	return gfn.Map(d.Values, func(e *IdentifierExpr) string { return e.Value })
}

func (e *EnumDecl) String() string {
	vals := []string{}
	for _, v := range e.Values {
		vals = append(vals, v.Value)
	}
	return fmt.Sprintf("enum %s { %s };", e.Name, strings.Join(vals, ", "))
}

func (e *EnumDecl) PrettyPrint(cp CodePrinter) {
	cp.Println("enum {")
	WithIndent(1, cp, func(cp CodePrinter) {
		for _, v := range e.Values {
			cp.Printf("%s, \n", v.Value)
		}
	})
	cp.Print("}")
}

// ImportDecl represents `import "path";`
type ImportDecl struct {
	NodeInfo
	Path         *LiteralExpr // Should be a STRING literal
	Alias        *IdentifierExpr
	ImportedItem *IdentifierExpr

	ResolvedFullPath string // Full path to the imported item, resolved after loading (recursively if needed)
	ResolvedItem     Node   // Resolves to a ComponentDecl or Method or Param in the ResolvedFullPath module
}

func (i *ImportDecl) String() string {
	if i.Alias != nil {
		return fmt.Sprintf("import %s as %s from '%s';", i.ImportedItem.Value, i.Alias.Value, i.Path)
	} else {
		return fmt.Sprintf("import %s from '%s';", i.ImportedItem.Value, i.Path)
	}
}

func (i *ImportDecl) PrettyPrint(cp CodePrinter) {
	cp.Print(i.String())
}

// What the import is imported as if an alias is used
func (i *ImportDecl) ImportedAs() string {
	if i.Alias != nil {
		return i.Alias.Value
	}
	return i.ImportedItem.Value
}

// --- SystemDecl Definition ---

// SystemDecl represents a system declaration with typed parameters:
//
//	system Name(param1: Type1, param2: Type2) { body }
//
// Systems declare which component instances participate in a simulation.
// Parameters are typed references to component types. The body is reserved
// for generator and metric declarations (future issues #24/#25).
// The legacy syntax `system Name { ... }` (no parameters) is still supported.
type SystemDecl struct {
	NodeInfo
	Name       *IdentifierExpr
	Parameters []*ParamDecl         // Typed parameters: (name: ComponentType, ...)
	Body       []SystemDeclBodyItem // ExprStmt (generator/metric calls)

	// Resolved during inference from generator(...) calls in Body
	Generators []*GeneratorSpec

	// File declaration this System is declared in
	ParentFileDecl *FileDecl
}

// GeneratorSpec is the resolved form of a generator(...) call in a system body.
// Populated during the inference phase so errors are caught at compile time.
type GeneratorSpec struct {
	NodeInfo
	Name          string  // Generator identifier
	ComponentPath string  // Dot-separated component path (e.g., "arch.webserver")
	MethodName    string  // Target method name (e.g., "RequestRide")
	Rate          float64 // Calls per interval
	RateInterval  float64 // Interval in seconds (default 1.0 = per second)
	Duration      float64 // Duration in seconds (0 = forever)
}

func (s *SystemDecl) String() string {
	if len(s.Parameters) == 0 {
		return fmt.Sprintf("system %s { ... }", s.Name)
	}
	params := []string{}
	for _, p := range s.Parameters {
		params = append(params, fmt.Sprintf("%s %s", p.Name.Value, p.TypeDecl.Name))
	}
	return fmt.Sprintf("system %s(%s) { ... }", s.Name, strings.Join(params, ", "))
}

func (s *SystemDecl) PrettyPrint(cp CodePrinter) {
	if len(s.Parameters) > 0 {
		cp.Printf("system %s(", s.Name.Value)
		for i, p := range s.Parameters {
			if i > 0 {
				cp.Print(", ")
			}
			cp.Printf("%s %s", p.Name.Value, p.TypeDecl.Name)
		}
		cp.Print(") {\n")
	} else {
		cp.Printf("system %s {\n", s.Name.Value)
	}
	WithIndent(1, cp, func(cp CodePrinter) {
		for _, b := range s.Body {
			b.PrettyPrint(cp)
			cp.Println("")
		}
	})
	cp.Printf("}")
}

// SystemDeclBodyItem marker interface for items allowed in SystemDecl body.
type SystemDeclBodyItem interface {
	Node
	systemBodyItemNode()
}

// SplitMemberAccessTarget walks a MemberAccessExpr tree and splits off the rightmost
// member as the method name, with everything else joined as the component path.
// For "arch.webserver.RequestRide", returns ("arch.webserver", "RequestRide").
func SplitMemberAccessTarget(expr Expr) (componentPath string, methodName string) {
	switch e := expr.(type) {
	case *MemberAccessExpr:
		methodName = e.Member.Value
		componentPath = MemberAccessToString(e.Receiver)
	default:
		// Shouldn't happen for well-formed generator targets
		methodName = fmt.Sprintf("%s", expr)
	}
	return
}

// MemberAccessToString converts a MemberAccessExpr chain to a dotted string.
func MemberAccessToString(expr Expr) string {
	switch e := expr.(type) {
	case *IdentifierExpr:
		return e.Value
	case *MemberAccessExpr:
		return MemberAccessToString(e.Receiver) + "." + e.Member.Value
	default:
		return fmt.Sprintf("%s", expr)
	}
}

// InstanceDecl represents `instanceName ComponentType ( overrides );`
type InstanceDecl struct {
	NodeInfo
	Name          *IdentifierExpr
	ComponentName *IdentifierExpr
	Overrides     []*AssignmentStmt

	// Resolved during type inference/checking
	ResolvedComponentDecl *ComponentDecl // The resolved component declaration (handles imports)
}

func (i *InstanceDecl) systemBodyItemNode() {}
func (i *InstanceDecl) String() string {
	return fmt.Sprintf("instance %s: %s = { ... };", i.Name, i.ComponentName)
}

func (i *InstanceDecl) PrettyPrint(cp CodePrinter) {
	if i.Overrides == nil {
		cp.Printf("use %s %s", i.Name.Value, i.ComponentName.Value)
	} else {
		cp.Printf("use %s %s (", i.Name.Value, i.ComponentName.Value)
		for idx, o := range i.Overrides {
			if idx > 0 {
				cp.Print(", ")
			}
			o.PrettyPrint(cp)
		}
		cp.Print(" )")
	}
}

// AnalyzeDecl represents `analyze name = callExpr expect { ... };`
type AnalyzeDecl struct {
	NodeInfo
	Name         *IdentifierExpr
	Target       *CallExpr         // Changed from Expr
	Expectations *ExpectationsDecl // Optional
}

func (a *AnalyzeDecl) systemBodyItemNode() {}
func (a *AnalyzeDecl) String() string {
	s := fmt.Sprintf("analyze %s = %s", a.Name, a.Target)
	if a.Expectations != nil {
		s += " " + a.Expectations.String()
	}
	return s + ";"
}

// ExpectationsDecl represents `expect { expectStmt* }`
type ExpectationsDecl struct {
	NodeInfo
	Expects []*ExpectStmt
}

func (e *ExpectationsDecl) String() string { return "expect { ... }" }

// Aggregator declarations - For now they can only be native
// Aggregators are used to select from a set of futures being waited on
type AggregatorDecl struct {
	NodeInfo
	Name       *IdentifierExpr
	Parameters []*ParamDecl // Signature parameters (can be empty)
	ReturnType *TypeDecl    // Optional return type (primitive or enum)

	// Once resolved
	ParentFileDecl *FileDecl
}

func (a *AggregatorDecl) String() string {
	s := fmt.Sprintf("aggregator %s (%s) %s", a.Name, "ParamTypesTbd", a.ReturnType)
	return s + ";"
}
func (a *AggregatorDecl) PrettyPrint(cp CodePrinter) {
	cp.Print(a.String())
}
