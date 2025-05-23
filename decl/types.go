package decl

import (
	"fmt"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

type Type struct {
	Name         string
	ChildTypes   []*Type
	ChildNames   []string // when we are a record type
	IsEnum       bool
	OriginalDecl Node     // Pointer to the original EnumDecl, ComponentDecl, etc.
}

// --- ValueType Factory Functions ---

var (
	// Use singletons for basic types for efficiency
	NilType       = &Type{Name: "Nil"} // Changed from {} to have a name for clarity
	BoolType      = &Type{Name: "Bool"}
	IntType       = &Type{Name: "Int"}
	FloatType     = &Type{Name: "Float"}
	StrType       = &Type{Name: "String"}
	ComponentType = &Type{Name: "Component"} // Represents the type "Component" itself, not an instance
	OpNodeType    = &Type{Name: "OpNode"}
)

func ListType(elementType *Type) *Type {
	if elementType == nil {
		panic("List element type cannot be nil")
	}
	return &Type{
		Name:       "List",
		ChildTypes: []*Type{elementType},
	}
}

func TupleType(elementTypes ...*Type) *Type {
	if len(elementTypes) == 0 {
		panic("Tuple element type cannot be nil or 0 length")
	}
	return &Type{
		Name:       "Tuple",
		ChildTypes: elementTypes,
	}
}

func OutcomesType(elementType *Type) *Type {
	if elementType == nil {
		panic("Outcomes element type cannot be nil")
	}
	return &Type{
		Name:       "Outcomes",
		ChildTypes: []*Type{elementType},
	}
}

// Helper to create a Type instance for an Enum declaration
func EnumType(decl *EnumDecl) *Type {
	if decl == nil || decl.NameNode == nil {
		panic("EnumDecl and its NameNode cannot be nil when creating EnumType")
	}
	return &Type{Name: decl.NameNode.Name, IsEnum: true, OriginalDecl: decl}
}

// Helper to create a Type instance for a Component declaration (representing the type)
func ComponentTypeInstance(decl *ComponentDecl) *Type {
	if decl == nil || decl.NameNode == nil {
		panic("ComponentDecl and its NameNode cannot be nil when creating ComponentTypeInstance")
	}
	// This represents the "type" of a component, e.g. "MyComponentType"
	// Distinguish from `ComponentType` singleton which means "any component".
	return &Type{Name: decl.NameNode.Name, OriginalDecl: decl}
}


// String representation of the type
func (t *Type) String() string {
	if t == nil {
		return "<nil_type>"
	}
	if t.Name == "Nil" && len(t.ChildTypes) == 0 { // Handle NilType specifically
		return "Nil"
	}
	if len(t.ChildTypes) == 0 {
		return fmt.Sprintf("%s", t.Name)
	} else {
		return fmt.Sprintf("%s[%s]", t.Name, strings.Join(gfn.Map(t.ChildTypes, func(ct *Type) string { return ct.String() }), ", "))
	}
}

// Equals checks if two ValueType definitions are equivalent.
func (v *Type) Equals(other *Type) bool {
	if v == other { // Pointer equality check (useful for singletons)
		return true
	}
	if v == nil || other == nil {
		return false
	}
	if v.Name != other.Name || v.IsEnum != other.IsEnum {
		return false
	}
	// For OriginalDecl, we might not want to compare them for general type equality,
	// as two types can be structurally "MyComponent" even if from different instances of type analysis.
	// However, if Name matches and IsEnum matches, and OriginalDecl *kinds* match (e.g. both are EnumDecl),
	// it implies type equality for named types like enums and components.
	// For now, direct comparison of OriginalDecl is omitted for simplicity in general Equals,
	// but specific checks might need it. The primary check is Name and IsEnum.

	if len(v.ChildTypes) != len(other.ChildTypes) {
		return false
	}
	if len(v.ChildNames) != len(other.ChildNames) {
		return false
	}

	for i, t1 := range v.ChildTypes {
		if !t1.Equals(other.ChildTypes[i]) {
			return false
		}
	}
	for i, n1 := range v.ChildNames {
		if n1 != other.ChildNames[i] {
			return false
		}
	}
	return true
}

// IsComponentType checks if the type represents a component (based on OriginalDecl).
func (t *Type) IsComponentType() bool {
    if t == nil {
        return false
    }
    _, ok := t.OriginalDecl.(*ComponentDecl)
    return ok
}

