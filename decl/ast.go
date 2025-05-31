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

	ResolvedFullPath string // Full path to the imported item, resolved after loading
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

// --- ComponentDecl Definition ---

// Keeps track of a component ref for lazy resolution after file has been loaded
// For example, components in a file can be declared in any order and can use other
// components declared in a file.

// ComponentDecl represents `component Name { ... }`
type ComponentDecl struct {
	NodeInfo
	Name *IdentifierExpr         // ComponentDecl type name
	Body []ComponentDeclBodyItem // ParamDecl, UsesDecl, MethodDecl

	// Marks whether a component is native or not
	// Native components should still be declared if not defined.
	// Method bodies of native components will be ignored (if you need to override we
	// can introduce inheritance or mixins later on)
	IsNative bool

	// Resolved values so we can work with processed/loaded values instead of resolving
	// Identify expressions etc
	resolved bool

	// Parameters and Dependencies are in the order in which they appear.  This is important unlike
	// methods parameters can only reer to other parameters if they have been defined first.
	paramList []*ParamDecl
	usesList  []*UsesDecl

	params  map[string]*ParamDecl  // Processed parameters map[name]*ParamDecl
	uses    map[string]*UsesDecl   // Processed dependencies map[local_name]*UsesDecl
	methods map[string]*MethodDecl // Processed methods map[method_name]*MethodDef
}

func (d *ComponentDecl) Params() (out []*ParamDecl, err error) {
	err = d.Resolve()
	out = d.paramList
	return
}

func (d *ComponentDecl) GetParam(name string) (out *ParamDecl, err error) {
	err = d.Resolve()
	if err == nil {
		out = d.params[name]
	}
	return
}

func (d *ComponentDecl) Methods() (out map[string]*MethodDecl, err error) {
	err = d.Resolve()
	out = d.methods
	return
}

func (d *ComponentDecl) GetMethod(name string) (out *MethodDecl, err error) {
	methods, err := d.Methods()
	if err == nil {
		out = methods[name]
	}
	return
}

func (d *ComponentDecl) Dependencies() (out []*UsesDecl, err error) {
	err = d.Resolve()
	out = d.usesList
	return
}

func (d *ComponentDecl) GetDependency(name string) (out *UsesDecl, err error) {
	err = d.Resolve()
	if err == nil {
		out = d.uses[name]
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
			paramName := bodyNode.Name.Value
			if _, exists := d.params[paramName]; exists {
				return fmt.Errorf("duplicate parameter '%s'", paramName) // Error relative to component name handled by caller
			}
			d.params[paramName] = bodyNode
			d.paramList = append(d.paramList, bodyNode)
		case *UsesDecl:
			usesName := bodyNode.Name.Value
			if _, exists := d.uses[usesName]; exists {
				return fmt.Errorf("duplicate uses declaration '%s'", usesName)
			}
			d.uses[usesName] = bodyNode
			d.usesList = append(d.usesList, bodyNode)
		case *MethodDecl:
			methodName := bodyNode.Name.Value
			if _, exists := d.methods[methodName]; exists {
				return fmt.Errorf("duplicate method definition '%s'", methodName)
			}
			d.methods[methodName] = bodyNode
			/* Disable recursive components for now
			case *ComponentDecl:
				// Handle nested definitions - recursive processing
				nestedCompDef, err := v.processComponentDecl(bodyNode)
				if err != nil {
					return fmt.Errorf("error processing nested component '%s': %w", bodyNode.Name.Value, err)
				}
				// How to register nested? Maybe prefix name? Or store within outer compDef?
				// For now, let's register globally with potentially full name? Needs design.
				// Let's just process it for validation for now, registration TBD.
				// We could register it here:
				err = v.RegisterComponentDef(nestedCompDef)
				if err != nil {
					return fmt.Errorf("error registering nested component '%s': %w", nestedCompDef.Node.Name.Value, err)
				}
			*/

		default:
			// Ignore other items like comments or potentially misplaced nodes
		}
	}
	d.resolved = true
	return nil
}

func (c *ComponentDecl) String() string         { return fmt.Sprintf("component %s { ... }", c.Name) }
func (c *ComponentDecl) componentBodyItemNode() {}
func (c *ComponentDecl) PrettyPrint(cp CodePrinter) {
	cp.Printf("component %s {", c.Name.Value)
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
	TypeDecl     *TypeDecl
	DefaultValue Expr // Optional
}

func (p *ParamDecl) componentBodyItemNode() {}
func (p *ParamDecl) String() string {
	s := fmt.Sprintf("param %s: %s", p.Name, p.TypeDecl)
	if p.DefaultValue != nil {
		s += fmt.Sprintf(" = %s", p.DefaultValue)
	}
	return s + ";"
}

func (p *ParamDecl) PrettyPrint(cp CodePrinter) {
	cp.Printf("param %s ", p.Name.Value)
	if p.TypeDecl != nil {
		p.TypeDecl.PrettyPrint(cp)
	}
	if p.DefaultValue != nil {
		cp.Print(" = ")
		p.DefaultValue.PrettyPrint(cp)
	}
}

// UsesDecl represents `uses varName: ComponentType [{ overrides }];`
type UsesDecl struct {
	NodeInfo
	Name          *IdentifierExpr
	ComponentName *IdentifierExpr   // Type name of the dependency
	Overrides     []*AssignmentStmt // When overrides are specified - the component is also instantiated otherwise it must be set in a system or in the enclosing component's uses decl

	// Resolved ComponentDecl after type checking
	ResolvedComponent *ComponentDecl
}

func (u *UsesDecl) componentBodyItemNode() {}
func (u *UsesDecl) String() string         { return fmt.Sprintf("uses %s: %s;", u.Name, u.ComponentName) }

func (u *UsesDecl) PrettyPrint(cp CodePrinter) {
	cp.Printf("uses %s %s\n", u.Name.Value, u.ComponentName.Value)
}

// MethodDecl represents `method name(params) [: returnType] { body }`
type MethodDecl struct {
	NodeInfo
	Name       *IdentifierExpr
	Parameters []*ParamDecl // Signature parameters (can be empty)
	ReturnType *TypeDecl    // Optional return type (primitive or enum)
	Body       *BlockStmt
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
		paramStr += param.Name.Value
		paramStr += " "
		paramStr += param.TypeDecl.String()
	}
	if m.ReturnType == nil {
		cp.Printf("method %s(%s) {\n", m.Name.Value, paramStr)
	} else {
		cp.Printf("method %s(%s) %s {\n", m.Name.Value, paramStr, m.ReturnType.String())
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
	Name *IdentifierExpr
	Body []SystemDeclBodyItem // InstanceDecl, AnalyzeDecl, OptionsDecl, LetStmt
}

func (s *SystemDecl) String() string { return fmt.Sprintf("system %s { ... }", s.Name) }
func (s *SystemDecl) PrettyPrint(cp CodePrinter) {
	cp.Printf("system %s {\n", s.Name.Value)
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

// InstanceDecl represents `instanceName ComponentType ( overrides );`
type InstanceDecl struct {
	NodeInfo
	Name          *IdentifierExpr
	ComponentName *IdentifierExpr
	Overrides     []*AssignmentStmt
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
