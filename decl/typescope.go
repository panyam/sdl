package decl

import (
	"fmt"
	// "log" // Uncomment for debugging
)

// TypeScope manages type information for identifiers within a scope.
// It uses an Env[Node] for storing named declarations (globals, imports, lexical vars)
// and retains direct references for contextual lookups (self, method params, component params).
type TypeScope struct {
	env              *Env[Node] // Stores LetStmt vars, global decls (Enums, Components), and imported symbols.
	currentComponent *ComponentDecl
	currentMethod    *MethodDecl
}

// NewRootTypeScope creates a top-level scope, using the provided Env.
// The env should be pre-populated by the loader with global and imported declarations.
func NewRootTypeScope(env *Env[Node]) *TypeScope {
	if env == nil {
		env = NewEnv[Node](nil)
		// panic("NewRootTypeScope requires a non-nil Env")
	}
	return &TypeScope{
		env:              env,
		currentComponent: nil,
		currentMethod:    nil,
	}
}

// Push creates a new nested lexical scope (e.g., for a block or method).
// It pushes the environment for lexical scoping of variables defined with 'let'.
func (ts *TypeScope) Push(currentComponent *ComponentDecl, currentMethod *MethodDecl) *TypeScope {
	effectiveCurrentComponent := ts.currentComponent
	if currentComponent != nil {
		effectiveCurrentComponent = currentComponent
	}
	effectiveCurrentMethod := ts.currentMethod
	if currentMethod != nil {
		effectiveCurrentMethod = currentMethod
	}

	return &TypeScope{
		env:              ts.env.Push(), // New lexical scope for 'let' variables
		currentComponent: effectiveCurrentComponent,
		currentMethod:    effectiveCurrentMethod,
	}
}

// Get retrieves the type of an identifier.
// Lookup Order:
// 1. 'self' (if in component method context).
// 2. Method parameters (if in method context).
// 3. Component parameters (if in component context, typically accessed via 'self.param').
// 4. 'uses' dependencies (if in component context, typically accessed via 'self.dep').
// 5. Lexically scoped variables (let bindings) and global/imported declarations from the Env.
func (ts *TypeScope) Get(name string) (*Type, bool) {
	// 1. 'self' keyword
	if ts.currentComponent != nil && name == "self" {
		// 'self' refers to the current component type
		return ComponentTypeInstance(ts.currentComponent), true
	}

	// 2. Method parameters
	if ts.currentMethod != nil {
		for _, param := range ts.currentMethod.Parameters {
			if param.Name.Name == name {
				if param.Type == nil { // Should be caught by earlier validation
					// log.Printf("Warning: Parameter '%s' in method '%s' has no TypeDecl.", name, ts.currentMethod.NameNode.Name)
					return nil, false
				}
				// The Type field of ParamDecl is a TypeDecl node.
				// Type.Type() resolves it to a *Type, which should include OriginalDecl.
				return param.Type.Type(), true
			}
		}
	}

	// 3. Component parameters (only relevant if we are in a component's direct scope, not method)
	//    Typically accessed via `self.param_name`, handled by MemberAccessExpr.
	//    If direct access like `param_name` is allowed from a method, this could be added.
	//    For now, assuming params are primarily via `self`.

	// 4. 'uses' dependencies (similar to component params, typically via `self.dependency_name`)

	// 5. Lexically scoped variables ('let' bindings) and global/imported declarations from Env.
	//    The Env stores the actual decl.Node.
	declNode, foundNode := ts.env.Get(name)
	if foundNode {
		switch node := declNode.(type) {
		case *EnumDecl:
			return EnumType(node), true
		case *ComponentDecl: // This is for referring to a component *as a type*
			return ComponentTypeInstance(node), true
		case *IdentifierExpr: // This is how 'let' bound variables are stored (the LHS Identifier)
			if node.InferredType() == nil {
				// This implies a 'let' variable was used before its type could be fully inferred,
				// or there was an error during its inference.
				// log.Printf("Warning: Identifier '%s' found in env but its type is not yet inferred.", name)
				return nil, false // Or return a special "inference pending" type / error
			}
			return node.InferredType(), true
		case *InstanceDecl: // If instance names from SystemDecl are in scope
			// The type of an instance is its component type.
			// We need to find the ComponentDecl for node.ComponentType.Name.
			// This requires the Env to also hold ComponentDecls by their type names.
			if ts.env != nil { // Check if env is available
				compDeclNode, compFound := ts.env.Get(node.ComponentType.Name)
				if compFound {
					if compDecl, ok := compDeclNode.(*ComponentDecl); ok {
						return ComponentTypeInstance(compDecl), true
					}
				}
				// log.Printf("Warning: Could not resolve component type '%s' for instance '%s'", node.ComponentType.Name, name)
			}
			return nil, false // Could not fully resolve instance type
			// Add cases for other declarable/typeable nodes if they can be directly named and looked up.
		default:
			// log.Printf("Warning: Found node of unexpected type %T for name '%s' in env.", node, name)
			return nil, false
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
