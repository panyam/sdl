package decl

import (
	"fmt"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

type TypeTag int

const (
	TypeTagUnknown TypeTag = iota
	TypeTagNil
	TypeTagSimple
	TypeTagTuple
	TypeTagList
	TypeTagEnum
	TypeTagComponent
	TypeTagMethod
	TypeTagOutcomes
	TypeTagExpr
)

type Type struct {
	Tag  TypeTag
	Info any
}

// --- ValueType Factory Functions ---

var (
	// Use singletons for basic types for efficiency
	NilType   = SimpleType("Nil")
	BoolType  = SimpleType("Bool")
	IntType   = SimpleType("Int")
	FloatType = SimpleType("Float")
	StrType   = SimpleType("String")
)

func SimpleType(name string) *Type {
	return &Type{Tag: TypeTagSimple, Info: name} // Changed from {} to have a name for clarity
}

type MethodTypeInfo struct {
	Component *ComponentDecl
	Method    *MethodDecl
}

func MethodType(componentDecl *ComponentDecl, methodDecl *MethodDecl) *Type {
	if componentDecl == nil || methodDecl == nil {
		panic("Component and Method Decls cannot be nil")
	}
	return &Type{
		Tag: TypeTagMethod,
		// Name: "Method",
		Info: &MethodTypeInfo{componentDecl, methodDecl},
	}
}

func TupleType(elementTypes ...*Type) *Type {
	if len(elementTypes) == 0 {
		panic("Tuple element type cannot be nil or 0 length")
	}
	return &Type{
		Tag: TypeTagTuple,
		// Name: "Tuple",
		Info: elementTypes,
	}
}

func ExprType(elementType *Type) *Type {
	return &Type{
		Tag:  TypeTagExpr,
		Info: elementType,
	}
}

func ListType(elementType *Type) *Type {
	if elementType == nil {
		panic("Outcomes element type cannot be nil")
	}
	return &Type{
		Tag: TypeTagList,
		// Name: "Outcomes",
		Info: elementType,
	}
}

func OutcomesType(elementType *Type) *Type {
	if elementType == nil {
		panic("Outcomes element type cannot be nil")
	}
	return &Type{
		Tag:  TypeTagOutcomes,
		Info: elementType,
	}
}

// Helper to create a Type instance for an Enum declaration
func EnumType(decl *EnumDecl) *Type {
	if decl == nil || decl.NameNode == nil {
		panic("EnumDecl and its NameNode cannot be nil when creating EnumType")
	}
	return &Type{Tag: TypeTagEnum /*, Name: decl.NameNode.Name*/, Info: decl}
}

// Helper to create a Type instance for a Component declaration (representing the type)
func ComponentType(decl *ComponentDecl) *Type {
	if decl == nil || decl.NameNode == nil {
		panic("ComponentDecl and its NameNode cannot be nil when creating ComponentTypeInstance")
	}
	// This represents the "type" of a component, e.g. "MyComponentType"
	// Distinguish from `ComponentType` singleton which means "any component".
	return &Type{Tag: TypeTagComponent /*Name: decl.NameNode.Name,*/, Info: decl}
}

// String representation of the type
func (t *Type) String() string {
	if t == nil {
		return "<nil_type>"
	} else if t.Tag == TypeTagSimple {
		return t.Info.(string)
	} else if t.Tag == TypeTagTuple {
		return fmt.Sprintf("Tuple(%s)", gfn.Map(t.Info.([]*Type), func(t *Type) string { return t.String() }))
	} else if t.Tag == TypeTagList {
		return fmt.Sprintf("List(%s)", t.Info.(*Type).String())
	} else if t.Tag == TypeTagEnum {
		enumDecl := t.Info.(*EnumDecl)
		return enumDecl.String()
	} else if t.Tag == TypeTagComponent {
		return fmt.Sprintf("Component(%s)", t.Info.(*ComponentDecl).NameNode.Name)
	} else if t.Tag == TypeTagMethod {
	} else if t.Tag == TypeTagOutcomes {
	}
	return "Unknown Type"
}

func (t *Type) PrettyPrint(cp CodePrinter) {
	cp.Print(t.String())
}

// Equals checks if two ValueType definitions are equivalent.
func (v *Type) Equals(other *Type) bool {
	if v == other { // Pointer equality check (useful for singletons)
		return true
	}
	if v == nil || other == nil {
		return false
	}
	// For OriginalDecl, we might not want to compare them for general type equality,
	// as two types can be structurally "MyComponent" even if from different instances of type analysis.
	// However, if Name matches and IsEnum matches, and OriginalDecl *kinds* match (e.g. both are EnumDecl),
	// it implies type equality for named types like enums and components.
	// For now, direct comparison of OriginalDecl is omitted for simplicity in general Equals,
	// but specific checks might need it. The primary check is Name and IsEnum.
	if v.Tag != other.Tag {
		return false
	}

	if v.Tag == TypeTagSimple {
		return v.Info.(string) == other.Info.(string)
	} else if v.Tag == TypeTagTuple {
		c1 := v.Info.([]*Type)
		c2 := other.Info.([]*Type)
		if len(c1) != len(c2) {
			return false
		}
		for i1, t1 := range c1 {
			t2 := c2[i1]
			if !t1.Equals(t2) {
				return false
			}
		}
		return true
	} else if v.Tag == TypeTagEnum {
		e1 := v.Info.(*EnumDecl)
		e2 := other.Info.(*EnumDecl)
		return e1.NameNode.Name == e2.NameNode.Name
	} else if v.Tag == TypeTagList || v.Tag == TypeTagOutcomes {
		return v.Info.(*Type).Equals(other.Info.(*Type))
	} else if v.Tag == TypeTagComponent {
		c1 := v.Info.(*ComponentDecl)
		c2 := other.Info.(*ComponentDecl)
		// TODO - a more deeper check
		return c1.NameNode.Name == c2.NameNode.Name
	}
	panic(fmt.Sprintf("Invalid types... %d, %v, %d, %v", v.Tag, v.Info, other.Tag, other.Info))
}

// IsComponentType checks if the type represents a component (based on OriginalDecl).
func (t *Type) IsComponentType() bool {
	if t == nil {
		return false
	}
	return t.Tag == TypeTagComponent
}

// TypeDecl represents primitive types or registered enum identifiers.

type TypeDecl struct {
	NodeInfo
	Name         string      // e.g., "MyEnum", "List", "Int"
	Args         []*TypeDecl // For generic-like types, e.g., List[Int] -> Args has TypeDecl for "Int"
	resolvedType *Type       // Cache the resolved *Type
}

func (t *TypeDecl) String() string {
	if len(t.Args) == 0 {
		return fmt.Sprintf("Type { %s ", t.Name)
	} else {
		return fmt.Sprintf("Type { %s[%s] } ", t.Name, strings.Join(gfn.Map(t.Args, func(t *TypeDecl) string { return t.String() }), ", "))
	}
}

func (t *TypeDecl) PrettyPrint(cp CodePrinter) {
	cp.Print(t.String())
}

func (td *TypeDecl) SetResolvedType(t *Type) {
	td.resolvedType = t
}

func (td *TypeDecl) ResolvedType() *Type {
	return td.resolvedType
}

/*
func (t *TypeDecl) Type() *Type {
	return &Type{
		Name:       t.Name,
		Info: gfn.Map(t.Args, func(t *TypeDecl) *Type { return t.Type() }),
	}
}
*/

// Type() was the original method, let's make it an alias to TypeUsingScope with a nil scope
// for backward compatibility or simple cases, though using TypeUsingScope is safer.
func (td *TypeDecl) Type() *Type {
	if td.resolvedType != nil {
		return td.resolvedType
	}
	// Fallback for simple types if no scope is needed or for basic resolution.
	// This might be too simple and could be removed if TypeUsingScope is always used.
	// For now, let's imagine it tries to resolve basic built-ins.
	switch td.Name {
	case "Int":
		return IntType
	case "Float":
		return FloatType
	case "String":
		return StrType
	case "Bool":
		return BoolType
	case "Nil":
		return NilType
		// Note: "List", "Tuple", "Outcomes" require args and are better handled by TypeUsingScope.
	}
	// Cannot resolve without a scope if it's a custom name.
	return nil
}

// In decl/ast.go (add this method to TypeDecl)
