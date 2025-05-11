package decl

import (
	"fmt"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

type Type struct {
	Name       string
	ChildTypes []*Type
	// Name        ValueTypeName
	// ChildTypes []*Type // Used for List, Outcomes, potentially Func later
}

// --- ValueType Factory Functions ---

var (
	// Use singletons for basic types for efficiency
	NilType       = &Type{}
	BoolType      = &Type{Name: "Bool"}
	IntType       = &Type{Name: "Int"}
	FloatType     = &Type{Name: "Float"}
	StrType       = &Type{Name: "String"}
	ComponentType = &Type{Name: "Component"}
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

// String representation of the type
func (t *Type) String() string {
	if len(t.ChildTypes) == 0 {
		return fmt.Sprintf("%s", t.Name)
	} else {
		return fmt.Sprintf("%s[%s]", t.Name, strings.Join(gfn.Map(t.ChildTypes, func(t *Type) string { return t.String() }), ", "))
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
	if v.Name != other.Name {
		return false
	}
	if len(v.ChildTypes) != len(other.ChildTypes) {
		return false
	}

	for i, t1 := range v.ChildTypes {
		if !t1.Equals(other.ChildTypes[i]) {
			return false
		}
	}
	return true
}
