package decl

import (
	"fmt"
	"log"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

// --- Interfaces ---

type Location struct {
	Pos  int
	Line int
	Col  int
}

// Node represents any node in the Abstract Syntax Tree.
type Node interface {
	Pos() Location  // Starting position (for error reporting)
	End() Location  // Ending position
	String() string // String representation for debugging/printing
}

// --- Base Struct ---

// NodeInfo embeddable struct for position tracking.
type NodeInfo struct{ StartPos, StopPos Location }

func (n *NodeInfo) Pos() Location  { return n.StartPos }
func (n *NodeInfo) End() Location  { return n.StopPos }
func (n *NodeInfo) String() string { return "{Node}" } // Default stringer
func (n *NodeInfo) stmtNode()      {}

// --- Top Level declarations ---

// FileDecl represents the top-level node of a parsed DSL file.
type FileDecl struct {
	NodeInfo
	Declarations []Node // ComponentDecl, SystemDecl, OptionsDecl, EnumDecl, ImportDecl

	// Resolved values so we can work with processed/loaded values instead of resolving
	// Identify expressions etc
	resolved   bool
	components map[string]*ComponentDecl
	enums      map[string]*EnumDecl
	imports    map[string]*ImportDecl
	importList []*ImportDecl // Keep original list for iteration order if needed
	systems    map[string]*SystemDecl
}

func (f *FileDecl) GetComponents() (out map[string]*ComponentDecl, err error) {
	err = f.Resolve()
	out = f.components
	return
}

func (f *FileDecl) GetComponent(name string) (out *ComponentDecl, err error) {
	components, err := f.GetComponents()
	if err == nil {
		out = components[name]
	}
	return
}

func (f *FileDecl) GetEnums() (out map[string]*EnumDecl, err error) {
	err = f.Resolve()
	out = f.enums
	return
}

func (f *FileDecl) GetEnum(name string) (out *EnumDecl, err error) {
	enums, err := f.GetEnums()
	if err == nil {
		out = enums[name]
	}
	return
}

func (f *FileDecl) GetSystems() (out map[string]*SystemDecl, err error) {
	err = f.Resolve()
	out = f.systems
	return
}

func (f *FileDecl) GetSystem(name string) (out *SystemDecl, err error) {
	systems, err := f.GetSystems()
	if err == nil {
		out = systems[name]
	}
	return
}

func (f *FileDecl) Imports() (map[string]*ImportDecl, error) {
	if err := f.Resolve(); err != nil {
		return nil, err
	}
	return f.imports, nil
}

// Called to resolve specific AST aspects out of the parse tree
func (f *FileDecl) Resolve() error {
	if f == nil {
		return fmt.Errorf("cannot load nil file")
	}
	if f.resolved {
		return nil
	}
	// Add initializers for other registries (Enums, Options) if they exist

	// log.Printf("Loading definitions from File AST...")
	for _, decl := range f.Declarations {
		switch node := decl.(type) {
		case *ComponentDecl:
			// Process and register the component definition
			err := node.Resolve() // Use a helper function
			if err != nil {
				return fmt.Errorf("error processing component '%s' at pos %d: %w", node.NameNode.Name, node.Pos(), err)
			}
			if err := f.RegisterComponent(node); err != nil {
				return err
			}

		case *SystemDecl:
			// Store the SystemDecl AST by name for later execution
			if err := f.RegisterSystem(node); err != nil {
				return err
			}
		case *EnumDecl:
			if err := f.RegisterEnum(node); err != nil {
				return err
			}

		case *OptionsDecl:
			log.Printf("Found OptionsDecl (TODO: Implement processing)")

		case *ImportDecl:
			if err := f.RegisterImport(node); err != nil {
				return err
			}

		default:
			// Ignore other node types at the top level? Or error?
			// log.Printf("Ignoring unsupported top-level declaration type %T at pos %d", node, node.Pos())
		}
	}
	// log.Printf("Finished loading definitions.")
	f.resolved = true
	return nil
}

func (f *FileDecl) RegisterComponent(c *ComponentDecl) error {
	if f.components == nil {
		f.components = map[string]*ComponentDecl{}
	}
	if _, exists := f.components[c.NameNode.Name]; exists {
		return fmt.Errorf("component definition '%s' already registered", c.NameNode.Name)
	}
	f.components[c.NameNode.Name] = c
	return nil
}

func (f *FileDecl) RegisterSystem(c *SystemDecl) error {
	if f.systems == nil {
		f.systems = map[string]*SystemDecl{}
	}
	if _, exists := f.systems[c.NameNode.Name]; exists {
		return fmt.Errorf("system definition '%s' already registered", c.NameNode.Name)
	}
	f.systems[c.NameNode.Name] = c
	return nil
}

func (f *FileDecl) RegisterEnum(c *EnumDecl) error {
	if f.enums == nil {
		f.enums = map[string]*EnumDecl{}
	}
	if _, exists := f.enums[c.NameNode.Name]; exists {
		return fmt.Errorf("enum definition '%s' already registered", c.NameNode.Name)
	}
	f.enums[c.NameNode.Name] = c
	return nil
}

func (f *FileDecl) RegisterImport(c *ImportDecl) error {
	if f.imports == nil {
		f.imports = map[string]*ImportDecl{}
		f.importList = []*ImportDecl{}
	}
	if _, exists := f.imports[c.ImportedAs()]; exists {
		err := fmt.Errorf("import definition '%s' already registered", c.ImportedAs())
		panic(err)
	}
	f.imports[c.ImportedAs()] = c
	f.importList = append(f.importList, c)
	return nil
}

func (f *FileDecl) String() string {
	lines := []string{}
	for _, d := range f.Declarations {
		lines = append(lines, d.String())
	}
	return strings.Join(lines, "\n")
}

// OptionsDecl represents `options { ... }` (structure TBD)
type OptionsDecl struct {
	NodeInfo
	Body *BlockStmt // Placeholder for options assignments?
}

func (o *OptionsDecl) systemBodyItemNode() {}
func (o *OptionsDecl) String() string      { return "options { ... }" }

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

// ComponentDeclBodyItem marker interface for items allowed in ComponentDecl body.
type ComponentDeclBodyItem interface {
	Node
	componentBodyItemNode()
}

// TypeDecl represents primitive types or registered enum identifiers.
type TypeDecl struct {
	NodeInfo
	// Can be one of the following
	Name string
	Args []*TypeDecl
}

func (t *TypeDecl) Type() *Type {
	return &Type{
		Name:       t.Name,
		ChildTypes: gfn.Map(t.Args, func(t *TypeDecl) *Type { return t.Type() }),
	}
}

func (t *TypeDecl) String() string {
	if len(t.Args) == 0 {
		return fmt.Sprintf("Type { %s ", t.Name)
	} else {
		return fmt.Sprintf("Type { %s[%s] } ", t.Name, strings.Join(gfn.Map(t.Args, func(t *TypeDecl) string { return t.String() }), ", "))
	}
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

// --- SystemDecl Definition ---

// SystemDecl represents `system Name { ... }`
type SystemDecl struct {
	NodeInfo
	NameNode *IdentifierExpr
	Body     []SystemDeclBodyItem // InstanceDecl, AnalyzeDecl, OptionsDecl, LetStmt
}

func (s *SystemDecl) String() string { return fmt.Sprintf("system %s { ... }", s.NameNode) }

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
