package loader

import (
	"fmt"
	"log"
	"reflect"

	"github.com/panyam/sdl/decl"
	// "log" // Uncomment for debugging
)

// TypeScope manages type information for identifiers within a scope.
// It uses an Env[Node] for storing named declarations (globals, imports, lexical vars)
// and retains direct references for contextual lookups (self, method params, component params).
type TypeScope struct {
	env           *Env[Node] // Stores LetStmt vars, global decls (Enums, Components), and imported symbols.
	outer         *TypeScope
	currComponent *ComponentDecl
	currMethod    *MethodDecl
}

// NewRootTypeScope creates a top-level scope, using the provided Env.
// The env should be pre-populated by the loader with global and imported declarations.
func NewRootTypeScope(env *Env[Node]) *TypeScope {
	if env == nil {
		env = decl.NewEnv[Node](nil)
		// panic("NewRootTypeScope requires a non-nil Env")
	}
	return &TypeScope{
		env:           env,
		currComponent: nil,
		currMethod:    nil,
	}
}

// Push creates a new nested lexical scope for a new block.
func (ts *TypeScope) Push() *TypeScope {
	return &TypeScope{
		env:           ts.env.Push(),
		outer:         ts,
		currComponent: nil,
		currMethod:    nil,
	}
}

// PushComponent creates a new nested scope when entering a component
// The component's param and method declarations are added to the env
func (ts *TypeScope) PushComponent(comp *ComponentDecl) *TypeScope {
	ts = ts.Push()
	ts.currComponent = comp
	ts.currMethod = nil
	ts.env.Set("self", comp)

	params, _ := comp.Params()
	for _, def := range params {
		ts.env.Set(def.Name.Value, def)
	}

	deps, _ := comp.Dependencies()
	for _, def := range deps {
		ts.env.Set(def.Name.Value, def)
	}

	methods, _ := comp.Methods()
	for _, def := range methods {
		ts.env.Set(def.Name.Value, def)
	}
	return ts
}

// PushComponent creates a new nested scope when entering a component's methods
// The methods's params  are added to the env
func (ts *TypeScope) PushMethod(currComponent *ComponentDecl, currentMethod *MethodDecl) *TypeScope {
	ts = ts.Push()
	ts.currComponent = currComponent
	ts.currMethod = currentMethod
	for _, param := range currentMethod.Parameters {
		ts.env.Set(param.Name.Value, param)
	}
	return ts
}

func (ts *TypeScope) Method() *MethodDecl {
	for ; ts != nil; ts = ts.outer {
		if ts.currMethod != nil {
			return ts.currMethod
		}
	}
	return nil
}

func (ts *TypeScope) Component() *ComponentDecl {
	for ; ts != nil; ts = ts.outer {
		if ts.currComponent != nil {
			return ts.currComponent
		}
	}
	return nil
}

// Get retrieves the type of an identifier.
func (ts *TypeScope) Get(name string) (*Type, bool) {
	// 1. 'self' keyword
	node, _ := ts.env.Get(name)

	if node != nil {
		switch n := node.(type) {
		case *EnumDecl:
			return EnumType(n), true
		case *ComponentDecl:
			return ComponentType(n), true
		case *AggregatorDecl:
			return AggregatorType(n), true
		case *IdentifierExpr: // This is how 'let' bound variables are stored
			if n.InferredType() == nil {
				// log.Printf("Warning: Identifier '%s' (let variable) found in env but its type is not yet inferred.", name)
				return nil, false
			}
			return n.InferredType(), true
		case *InstanceDecl:
			if ts.env != nil {
				compDeclNode, compFound := ts.env.Get(n.ComponentName.Value)
				if compFound {
					if compDecl, ok := compDeclNode.(*ComponentDecl); ok {
						return ComponentType(compDecl), true
					}
				}
			}
			return nil, false
		case *MethodDecl:
			if ts.env != nil {
				methodDeclNode, methodFound := ts.env.Get(n.Name.Value)
				if methodFound {
					if methodDecl, ok := methodDeclNode.(*MethodDecl); ok {
						return MethodType(methodDecl), true
					}
				}
			}
			return nil, false
		default:
			log.Println("not found: node, nodetype: ", node, reflect.TypeOf(node))
			panic("TBD - we have to see if there is a failure else where or just a standard undefined error")
			// return nil, false
		}
	}

	return nil, false
}

// Set is used to define the type of a lexically scoped variable (from a LetStmt).
// It stores the LHS IdentifierExpr node in the current lexical environment (env).
// The inferred type `t` is set on the IdentifierExpr itself.
func (ts *TypeScope) Set(name string, identNode *IdentifierExpr, t *Type) error {
	if ts.env.GetRef(name) != nil { // Check current lexical scope only for 'let' shadowing/redefinition
		// Distinguish between GetRef on ts.env (for lexical) vs ts.env.Outer().GetRef (for parent scopes)
		// For 'let', we only care if 'name' is already defined in the *innermost* scope.
		// Env.GetRef() checks current then outer. We need a way to check current only.
		// Let's assume for now: if ts.env.store[name] exists.
		// This requires Env to expose its store or a "GetLocalRef" method.
		// For simplicity, let's assume Env.Set will overwrite if we allow shadowing,
		// or if Env.Set itself handles collision in the *same* scope level.
		// A proper check requires: `if _, definedInCurrentStore := ts.env.store[name]; definedInCurrentStore`
		// This is a placeholder for proper shadowing/redefinition error handling.
		// return fmt.Errorf("identifier '%s' already defined in the current lexical scope", name)
	}
	if identNode == nil {
		return fmt.Errorf("cannot set type for nil IdentifierExpr node (name: %s)", name)
	}
	identNode.SetInferredType(t)
	ts.env.Set(name, identNode) // Store the IdentifierExpr node itself.
	return nil
}

// Resolves resolves the TypeDecl to a *Type object within the given scope.
// It handles built-in types, named types (enums, components from scope), and generic-like types.
func (scope *TypeScope) ResolveType(td *TypeDecl) *Type {
	if td == nil {
		return nil
	}
	if td.ResolvedType() != nil {
		return td.ResolvedType()
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
			resolvedElemType := scope.ResolveType(elemTypeDecl)
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
				resolvedElemType := scope.ResolveType(argTd)
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
			resolvedElemType := scope.ResolveType(elemTypeDecl)
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
		td.SetResolvedType(resultType)
	}
	return resultType
}
