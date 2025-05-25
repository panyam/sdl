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

// --- Base Struct ---

// NodeInfo embeddable struct for position tracking.
type NodeInfo struct{ StartPos, StopPos Location }

func (n *NodeInfo) Pos() Location  { return n.StartPos }
func (n *NodeInfo) End() Location  { return n.StopPos }
func (n *NodeInfo) String() string { return "{Node}" } // Default stringer
func (n *NodeInfo) stmtNode()      {}

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
	NameNode   *IdentifierExpr   // EnumDecl type name
	ValuesNode []*IdentifierExpr // List of enum value names

	// Resolved values so we can work with processed/loaded values instead of resolving
	// Identify expressions etc
	values []string
}

func (d *EnumDecl) Values() []string {
	// TODO - save this
	return gfn.Map(d.ValuesNode, func(e *IdentifierExpr) string { return e.Name })
}

func (e *EnumDecl) String() string {
	vals := []string{}
	for _, v := range e.ValuesNode {
		vals = append(vals, v.Name)
	}
	return fmt.Sprintf("enum %s { %s };", e.NameNode, strings.Join(vals, ", "))
}

func (e *EnumDecl) PrettyPrint(cp CodePrinter) {
	cp.Println("enum {")
	WithIndent(1, cp, func(cp CodePrinter) {
		for _, v := range e.ValuesNode {
			cp.Printf("%s, \n", v.Name)
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
}

func (i *ImportDecl) String() string {
	if i.Alias != nil {
		return fmt.Sprintf("import %s as %s from '%s';", i.ImportedItem.Name, i.Alias.Name, i.Path)
	} else {
		return fmt.Sprintf("import %s from '%s';", i.ImportedItem.Name, i.Path)
	}
}

func (i *ImportDecl) PrettyPrint(cp CodePrinter) {
	cp.Print(i.String())
}

// What the import is imported as if an alias is used
func (i *ImportDecl) ImportedAs() string {
	if i.Alias != nil {
		return i.Alias.Name
	}
	return i.ImportedItem.Name
}

// --- ComponentDecl Definition ---

// Keeps track of a component ref for lazy resolution after file has been loaded
// For example, components in a file can be declared in any order and can use other
// components declared in a file.
type ComponentUse struct {
	Name         string         // Name/FQN of the component being used/referred to
	ResolvedDecl *ComponentDecl // The resolved component when needed
	Error        error          // Errors in resolution
}

// ComponentDecl represents `component Name { ... }`
type ComponentDecl struct {
	NodeInfo
	NameNode *IdentifierExpr         // ComponentDecl type name
	Body     []ComponentDeclBodyItem // ParamDecl, UsesDecl, MethodDecl

	// Marks whether a component is native or not
	// Native components should still be declared if not defined.
	// Method bodies of native components will be ignored (if you need to override we
	// can introduce inheritance or mixins later on)
	IsNative bool

	// Resolved values so we can work with processed/loaded values instead of resolving
	// Identify expressions etc
	resolved bool
	params   map[string]*ParamDecl  // Processed parameters map[name]*ParamDecl
	uses     map[string]*UsesDecl   // Processed dependencies map[local_name]*UsesDecl
	methods  map[string]*MethodDecl // Processed methods map[method_name]*MethodDef
}

func (d *ComponentDecl) GetParams() (out map[string]*ParamDecl, err error) {
	err = d.Resolve()
	out = d.params
	return
}

func (d *ComponentDecl) GetParam(name string) (out *ParamDecl, err error) {
	params, err := d.GetParams()
	if err == nil {
		out = params[name]
	}
	return
}

func (d *ComponentDecl) GetMethods() (out map[string]*MethodDecl, err error) {
	err = d.Resolve()
	out = d.methods
	return
}

func (d *ComponentDecl) GetMethod(name string) (out *MethodDecl, err error) {
	methods, err := d.GetMethods()
	if err == nil {
		out = methods[name]
	}
	return
}

func (d *ComponentDecl) GetDependencies() (out map[string]*UsesDecl, err error) {
	err = d.Resolve()
	out = d.uses
	return
}

func (d *ComponentDecl) GetDependency(name string) (out *UsesDecl, err error) {
	dependencies, err := d.GetDependencies()
	if err == nil {
		out = dependencies[name]
	}
	return
}

func (d *ComponentDecl) Resolve() error {
	if d.resolved {
		return nil
	}
	d.params = map[string]*ParamDecl{}
	d.uses = map[string]*UsesDecl{}      // Processed dependencies map[local_name]*UsesDecl
	d.methods = map[string]*MethodDecl{} // Processed dependencies map[local_name]*UsesDecl

	// Process body
	for _, item := range d.Body {
		switch bodyNode := item.(type) {
		case *ParamDecl:
			paramName := bodyNode.Name.Name
			if _, exists := d.params[paramName]; exists {
				return fmt.Errorf("duplicate parameter '%s'", paramName) // Error relative to component name handled by caller
			}
			d.params[paramName] = bodyNode
		case *UsesDecl:
			usesName := bodyNode.NameNode.Name
			if _, exists := d.uses[usesName]; exists {
				return fmt.Errorf("duplicate uses declaration '%s'", usesName)
			}
			d.uses[usesName] = bodyNode
		case *MethodDecl:
			methodName := bodyNode.NameNode.Name
			if _, exists := d.methods[methodName]; exists {
				return fmt.Errorf("duplicate method definition '%s'", methodName)
			}
			d.methods[methodName] = bodyNode
			/* Disable recursive components for now
			case *ComponentDecl:
				// Handle nested definitions - recursive processing
				nestedCompDef, err := v.processComponentDecl(bodyNode)
				if err != nil {
					return fmt.Errorf("error processing nested component '%s': %w", bodyNode.Name.Name, err)
				}
				// How to register nested? Maybe prefix name? Or store within outer compDef?
				// For now, let's register globally with potentially full name? Needs design.
				// Let's just process it for validation for now, registration TBD.
				// We could register it here:
				err = v.RegisterComponentDef(nestedCompDef)
				if err != nil {
					return fmt.Errorf("error registering nested component '%s': %w", nestedCompDef.Node.Name.Name, err)
				}
			*/

		default:
			// Ignore other items like comments or potentially misplaced nodes
		}
	}
	d.resolved = true
	return nil
}

func (c *ComponentDecl) String() string         { return fmt.Sprintf("component %s { ... }", c.NameNode) }
func (c *ComponentDecl) componentBodyItemNode() {}
func (c *ComponentDecl) PrettyPrint(cp CodePrinter) {
	cp.Printf("component %s {", c.NameNode.Name)
	WithIndent(1, cp, func(cp CodePrinter) {
		for _, item := range c.Body {
			item.PrettyPrint(cp)
			cp.Println("")
		}
	})
	cp.Printf("}")
}

// ComponentDeclBodyItem marker interface for items allowed in ComponentDecl body.
type ComponentDeclBodyItem interface {
	Node
	componentBodyItemNode()
}

// ParamDecl represents `param name: TypeDecl [= defaultExpr];`
type ParamDecl struct {
	NodeInfo
	Name         *IdentifierExpr
	Type         *TypeDecl
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

func (p *ParamDecl) PrettyPrint(cp CodePrinter) {
	cp.Printf("param %s ", p.Name.Name)
	if p.Type != nil {
		p.Type.PrettyPrint(cp)
	}
	if p.DefaultValue != nil {
		cp.Print(" = ")
		p.DefaultValue.PrettyPrint(cp)
	}
}

// UsesDecl represents `uses varName: ComponentType [{ overrides }];`
type UsesDecl struct {
	NodeInfo
	NameNode      *IdentifierExpr
	ComponentNode *IdentifierExpr // Type name of the dependency
}

func (u *UsesDecl) UsedComponent() *ComponentUse {
	return &ComponentUse{
		Name: u.ComponentNode.Name,
	}
}

func (u *UsesDecl) componentBodyItemNode() {}
func (u *UsesDecl) String() string         { return fmt.Sprintf("uses %s: %s;", u.NameNode, u.ComponentNode) }

func (u *UsesDecl) PrettyPrint(cp CodePrinter) {
	cp.Printf("uses %s %s\n", u.NameNode.Name, u.ComponentNode.Name)
}

// MethodDecl represents `method name(params) [: returnType] { body }`
type MethodDecl struct {
	NodeInfo
	NameNode   *IdentifierExpr
	Parameters []*ParamDecl // Signature parameters (can be empty)
	ReturnType *TypeDecl    // Optional return type (primitive or enum)
	Body       *BlockStmt

	Name string
}

func (o *MethodDecl) componentBodyItemNode() {}
func (o *MethodDecl) String() string {
	retType := ""
	if o.ReturnType != nil {
		retType = fmt.Sprintf(": %s", o.ReturnType)
	}
	return fmt.Sprintf("method %s(...) %s { ... }", o.Name, retType)
}

func (m *MethodDecl) PrettyPrint(cp CodePrinter) {
	paramStr := ""
	for idx, param := range m.Parameters {
		if idx > 0 {
			paramStr += ", "
		}
		paramStr += param.Name.Name
		paramStr += " "
		paramStr += param.Type.String()
	}
	if m.ReturnType == nil {
		cp.Printf("method %s(%s) {\n", m.NameNode.Name, paramStr)
	} else {
		cp.Printf("method %s(%s) %s {\n", m.NameNode.Name, paramStr, m.ReturnType.String())
	}
	cp.Indent(1)
	m.Body.PrettyPrint(cp)
	cp.Unindent(1)
	cp.Printf("}")
}

// --- SystemDecl Definition ---

// SystemDecl represents `system Name { ... }`
type SystemDecl struct {
	NodeInfo
	NameNode *IdentifierExpr
	Body     []SystemDeclBodyItem // InstanceDecl, AnalyzeDecl, OptionsDecl, LetStmt
}

func (s *SystemDecl) String() string { return fmt.Sprintf("system %s { ... }", s.NameNode) }
func (s *SystemDecl) PrettyPrint(cp CodePrinter) {
	cp.Printf("system %s {\n", s.NameNode.Name)
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

// InstanceDecl represents `instanceName: ComponentType = { overrides };`
type InstanceDecl struct {
	NodeInfo
	NameNode      *IdentifierExpr
	ComponentType *IdentifierExpr
	Overrides     []*AssignmentStmt
}

func (i *InstanceDecl) systemBodyItemNode() {}
func (i *InstanceDecl) String() string {
	return fmt.Sprintf("instance %s: %s = { ... };", i.NameNode, i.ComponentType)
}

func (i *InstanceDecl) PrettyPrint(cp CodePrinter) {
	if i.Overrides == nil {
		cp.Printf("use %s %s", i.NameNode.Name, i.ComponentType.Name)
	} else {
		cp.Printf("use %s %s (", i.NameNode.Name, i.ComponentType.Name)
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
