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
	OriginalDecl Node // Pointer to the original EnumDecl, ComponentDecl, etc.
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
		ChildTypes: gfn.Map(t.Args, func(t *TypeDecl) *Type { return t.Type() }),
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

// TypeUsingScope resolves the TypeDecl to a *Type object within the given scope.
// It handles built-in types, named types (enums, components from scope), and generic-like types.
func (td *TypeDecl) TypeUsingScope(scope *TypeScope) *Type {
	if td == nil {
		return nil
	}
	if td.resolvedType != nil {
		return td.resolvedType
	}

	var resultType *Type

	switch td.Name {
	// Basic known types (can be singletons from types.go)
	case "Int":
		resultType = IntType
	case "Float":
		resultType = FloatType
	case "String": // Assuming StrType is the correct singleton name
		resultType = StrType
	case "Bool":
		resultType = BoolType
	case "Nil": // For void/nil type
		resultType = NilType
	// Duration is often treated as Float or a distinct basic type.
	// If it's just "Duration", it would need to be a known basic type like Int/Float.
	// If it's from an enum or other decl, it'll be caught by the scope.Get below.

	case "List":
		if len(td.Args) == 1 {
			elemTypeDecl := td.Args[0]
			resolvedElemType := elemTypeDecl.TypeUsingScope(scope)
			if resolvedElemType != nil {
				resultType = ListType(resolvedElemType) // From types.go factory
			} else {
				// Error: element type of List could not be resolved
				return nil
			}
		} else {
			// Error: List expects 1 type argument
			return nil
		}
	case "Tuple":
		if len(td.Args) > 0 {
			elemTypes := make([]*Type, len(td.Args))
			for i, argTd := range td.Args {
				resolvedElemType := argTd.TypeUsingScope(scope)
				if resolvedElemType == nil {
					return nil // Error resolving tuple element type
				}
				elemTypes[i] = resolvedElemType
			}
			resultType = TupleType(elemTypes...) // From types.go factory
		} else {
			// Error: Tuple expects at least 1 type argument (as per TupleType factory)
			return nil
		}
	case "Outcomes":
		if len(td.Args) == 1 {
			elemTypeDecl := td.Args[0]
			resolvedElemType := elemTypeDecl.TypeUsingScope(scope)
			if resolvedElemType != nil {
				resultType = OutcomesType(resolvedElemType) // From types.go factory
			} else {
				// Error: element type of Outcomes could not be resolved
				return nil
			}
		} else {
			// Error: Outcomes expects 1 type argument
			return nil
		}
	default:
		// It's a named type (e.g., an enum "MyEnum", a component "MyComponentType").
		// Look it up in the provided scope.
		if scope == nil {
			// Cannot resolve named type without a scope.
			// This might happen if .Type() is called directly on a complex TypeDecl.
			return nil
		}
		foundType, ok := scope.Get(td.Name)
		if ok {
			// scope.Get already returns a *Type, which should have OriginalDecl set
			// if it came from an EnumDecl or ComponentDecl in the env.
			resultType = foundType
		} else {
			// Type name not found in scope
			return nil
		}
	}

	// Cache the resolved type
	if resultType != nil {
		td.resolvedType = resultType
	}
	return resultType
}
