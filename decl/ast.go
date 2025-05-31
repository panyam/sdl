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

// --- SystemDecl Definition ---

// SystemDecl represents `system Name { ... }`
type SystemDecl struct {
	NodeInfo
	Name *IdentifierExpr
	Body []SystemDeclBodyItem // InstanceDecl, AnalyzeDecl, OptionsDecl, LetStmt

	// File declaration this Component is declared in
	ParentFileDecl *FileDecl
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
