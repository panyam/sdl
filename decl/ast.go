package decl

import (
	"fmt"
	"log"
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

// --- Base Struct ---

// NodeInfo embeddable struct for position tracking.
type NodeInfo struct{ StartPos, StopPos int }

func (n *NodeInfo) Pos() int       { return n.StartPos }
func (n *NodeInfo) End() int       { return n.StopPos }
func (n *NodeInfo) String() string { return "{Node}" } // Default stringer

// Declarations are processed at load time unlike statements and expressions
// Used for holding mthod/component and other definitions
type Declaration interface {
	Node

	// Called to resolve specific AST aspects out of the parse tree
	Resolve(file *FileDecl) error
}

// --- Top Level declarations ---

// FileDecl represents the top-level node of a parsed DSL file.
type FileDecl struct {
	NodeInfo
	declarations []Node // ComponentDecl, SystemDecl, OptionsDecl, EnumDecl, ImportDecl

	// Resolved values so we can work with processed/loaded values instead of resolving
	// Identify expressions etc
	Components map[string]*ComponentDecl
	Enums      map[string]*EnumDecl
	Systems    map[string]*SystemDecl
}

// Called to resolve specific AST aspects out of the parse tree
func (f *FileDecl) Resolve(*FileDecl) error {
	if f == nil {
		return fmt.Errorf("cannot load nil file")
	}
	// Initialize maps if they are nil (might happen if VM wasn't Init'd properly)
	if f.Components == nil {
		f.Components = make(map[string]*ComponentDecl)
	}
	if f.Systems == nil {
		f.Systems = make(map[string]*SystemDecl)
	}
	// Add initializers for other registries (Enums, Options) if they exist

	// log.Printf("Loading definitions from File AST...")
	for _, decl := range f.declarations {
		switch node := decl.(type) {
		case *ComponentDecl:
			// Process and register the component definition
			err := node.Resolve(f) // Use a helper function
			if err != nil {
				return fmt.Errorf("error processing component '%s' at pos %d: %w", node.NameNode.Name, node.Pos(), err)
			}
			if err := f.RegisterComponent(node); err != nil {
				return err
			}

		case *SystemDecl:
			// Store the SystemDecl AST by name for later execution
			err := node.Resolve(f) // Use a helper function
			if err != nil {
				return fmt.Errorf("error processing system '%s' at pos %d: %w", node.NameNode.Name, node.Pos(), err)
			}
			if err := f.RegisterSystem(node); err != nil {
				return err
			}
		case *EnumDecl:
			err := node.Resolve(f) // Use a helper function
			if err != nil {
				return fmt.Errorf("error processing enum '%s' at pos %d: %w", node.NameNode.Name, node.Pos(), err)
			}
			if err := f.RegisterEnum(node); err != nil {
				return err
			}

		case *OptionsDecl:
			log.Printf("Found OptionsDecl (TODO: Implement processing)")

		case *ImportDecl:
			log.Printf("Found ImportDecl: %s (TODO: Implement handling)", node.Path)

		default:
			// Ignore other node types at the top level? Or error?
			// log.Printf("Ignoring unsupported top-level declaration type %T at pos %d", node, node.Pos())
		}
	}
	// log.Printf("Finished loading definitions.")
	return nil
}

func (f *FileDecl) RegisterComponent(c *ComponentDecl) error {
	if f.Components == nil {
		f.Components = map[string]*ComponentDecl{}
	}
	if _, exists := f.Components[c.Name]; exists {
		return fmt.Errorf("component definition '%s' already registered", c.Name)
	}
	f.Components[c.Name] = c
	return nil
}

func (f *FileDecl) RegisterSystem(c *SystemDecl) error {
	if f.Systems == nil {
		f.Systems = map[string]*SystemDecl{}
	}
	if _, exists := f.Systems[c.Name]; exists {
		return fmt.Errorf("system definition '%s' already registered", c.Name)
	}
	f.Systems[c.Name] = c
	return nil
}

func (f *FileDecl) RegisterEnum(c *EnumDecl) error {
	if f.Enums == nil {
		f.Enums = map[string]*EnumDecl{}
	}
	if _, exists := f.Enums[c.Name]; exists {
		return fmt.Errorf("enum definition '%s' already registered", c.Name)
	}
	f.Enums[c.Name] = c
	return nil
}

func (f *FileDecl) String() string {
	lines := []string{}
	for _, d := range f.declarations {
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
	Name   string
	Values []string
}

func (d *EnumDecl) Resolve(file *FileDecl) error {
	d.Name = d.NameNode.Name
	d.Values = gfn.Map(d.ValuesNode, func(e *IdentifierExpr) string { return e.Name })
	return nil
}

func (e *EnumDecl) String() string {
	vals := []string{}
	for _, v := range e.ValuesNode {
		vals = append(vals, v.Name)
	}
	return fmt.Sprintf("enum %s { %s };", e.Name, strings.Join(vals, ", "))
}

// ImportDecl represents `import "path";`
type ImportDecl struct {
	NodeInfo
	Path *LiteralExpr // Should be a STRING literal
}

func (i *ImportDecl) String() string { return fmt.Sprintf("import %s;", i.Path) }

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
	Name    string
	Params  map[string]*ParamDecl  // Processed parameters map[name]*ParamDecl
	Uses    map[string]*UsesDecl   // Processed dependencies map[local_name]*UsesDecl
	Methods map[string]*MethodDecl // Processed methods map[method_name]*MethodDef
}

func (d *ComponentDecl) Resolve(file *FileDecl) error {
	d.Name = d.NameNode.Name
	d.Params = map[string]*ParamDecl{} // Processed parameters map[name]*ParamDecl
	d.Uses = map[string]*UsesDecl{}    // Processed dependencies map[local_name]*UsesDecl

	// Process body
	for _, item := range d.Body {
		switch bodyNode := item.(type) {
		case *ParamDecl:
			paramName := bodyNode.Name.Name
			if _, exists := d.Params[paramName]; exists {
				return fmt.Errorf("duplicate parameter '%s'", paramName) // Error relative to component name handled by caller
			}
			d.Params[paramName] = bodyNode
		case *UsesDecl:
			if err := bodyNode.Resolve(file); err != nil {
				return err
			}
			usesName := bodyNode.NameNode.Name
			if _, exists := d.Uses[usesName]; exists {
				return fmt.Errorf("duplicate uses declaration '%s'", usesName)
			}
		case *MethodDecl:
			methodName := bodyNode.NameNode.Name
			if _, exists := d.Methods[methodName]; exists {
				return fmt.Errorf("duplicate method definition '%s'", methodName)
			}
			d.Methods[methodName] = bodyNode
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
	return nil
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
	PrimitiveTypeName string // "int", "float", "bool", "string", "duration", or an EnumDecl identifier
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
	NameNode      *IdentifierExpr
	ComponentNode *IdentifierExpr // Type name of the dependency

	// Resolved values
	Name         string
	ComponentRef ComponentUse
}

func (u *UsesDecl) Resolve(file *FileDecl) error {
	u.Name = u.NameNode.Name
	u.ComponentRef = ComponentUse{
		Name: u.ComponentNode.Name,
	}
	// TODO - Should we add u.Component.Ref into file.UnresolvedRefs so it can be resolved
	// after the entire file is loaded or should it be delayed until runtime for lazy loading?
	return nil
}

func (u *UsesDecl) componentBodyItemNode() {}
func (u *UsesDecl) String() string         { return fmt.Sprintf("uses %s: %s;", u.NameNode, u.ComponentNode) }

// MethodDecl represents `method name(params) [: returnType] { body }`
type MethodDecl struct {
	NodeInfo
	NameNode   *IdentifierExpr
	Parameters []*ParamDecl // Signature parameters (can be empty)
	ReturnType *TypeName    // Optional return type (primitive or enum)
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

	// Resolved values so we can work with processed/loaded values instead of resolving
	// Identify expressions etc
	Name string

	// Note how we can get body etc from the Body decl directly
}

func (s *SystemDecl) Resolve(f *FileDecl) error {
	s.Name = s.NameNode.Name
	for _, item := range s.Body {
		if err := item.Resolve(f); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemDecl) String() string { return fmt.Sprintf("system %s { ... }", s.Name) }

// SystemDeclBodyItem marker interface for items allowed in SystemDecl body.
type SystemDeclBodyItem interface {
	Node
	systemBodyItemNode()
	Resolve(f *FileDecl) error
}

// InstanceDecl represents `instanceName: ComponentType = { overrides };`
type InstanceDecl struct {
	Declaration
	NameNode      *IdentifierExpr
	ComponentType *IdentifierExpr
	Overrides     []*AssignmentStmt

	// Resolved values so we can work with processed/loaded values instead of resolving
	// Identify expressions etc
	Name      string
	Component string // Name of the component being used.  Can be an FQN later if we do "dot" based imports
}

func (i *InstanceDecl) systemBodyItemNode() {}
func (i *InstanceDecl) String() string {
	return fmt.Sprintf("instance %s: %s = { ... };", i.Name, i.ComponentType)
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
