package decl

// TypeScope manages type information for identifiers within a scope.
type TypeScope struct {
	store            map[string]*Type
	outer            *TypeScope
	file             *FileDecl // Access to global declarations (enums, components)
	currentComponent *ComponentDecl
	currentMethod    *MethodDecl // Useful for 'self' and method params/return
}

// NewRootTypeScope creates a top-level scope for a file.
func NewRootTypeScope(file *FileDecl) *TypeScope {
	return &TypeScope{
		store: make(map[string]*Type),
		outer: nil,
		file:  file,
	}
}

// Push creates a new nested scope.
func (ts *TypeScope) Push(currentComponent *ComponentDecl, currentMethod *MethodDecl) *TypeScope {
	// Inherit component/method context if not overridden
	effectiveCurrentComponent := ts.currentComponent
	if currentComponent != nil {
		effectiveCurrentComponent = currentComponent
	}
	effectiveCurrentMethod := ts.currentMethod
	if currentMethod != nil {
		effectiveCurrentMethod = currentMethod
	}

	return &TypeScope{
		store:            make(map[string]*Type),
		outer:            ts,
		file:             ts.file, // Propagate file
		currentComponent: effectiveCurrentComponent,
		currentMethod:    effectiveCurrentMethod,
	}
}

// Get retrieves the type of an identifier.
// Order: local store -> 'self' -> method params -> outer scopes -> file globals.
func (ts *TypeScope) Get(name string) (*Type, bool) {
	// 1. Local store
	if t, ok := ts.store[name]; ok {
		return t, true
	}

	// 2. 'self' keyword (if in a method context of a component)
	if ts.currentComponent != nil && name == "self" {
		return &Type{Name: ts.currentComponent.NameNode.Name /* IsComponent: true */}, true
	}

	// 3. Method parameters (if in a method context)
	if ts.currentMethod != nil {
		for _, param := range ts.currentMethod.Parameters {
			if param.Name.Name == name {
				if param.Type == nil { // Should be caught by an earlier validation pass
					return nil, false // Or return an error type marker
				}
				return param.Type.Type(), true
			}
		}
	}
	// Note: Component-level params and 'uses' deps are typically accessed via 'self.member',
	// which is handled by MemberAccessExpr inference, not direct lookup here.

	// 4. Outer scopes
	if ts.outer != nil {
		return ts.outer.Get(name)
	}

	// 5. File-level declarations (Enums, Components as type names, Instances in SystemDecl)
	if ts.file != nil {
		if enumDecl, _ := ts.file.GetEnum(name); enumDecl != nil {
			// This identifier refers to an enum type itself
			return &Type{Name: enumDecl.NameNode.Name, IsEnum: true}, true
		}
		// An identifier might refer to a component type (e.g. in `instance x: MyComponent;`)
		// if compDecl, _ := ts.file.GetComponent(name); compDecl != nil {
		//    return &Type{Name: compDecl.NameNode.Name /* IsComponentType: true */}, true
		// }
		// Note: Instance names declared in SystemDecl are added to the scope by inferTypesForSystemDeclBodyItem.
	}

	return nil, false
}

// Set defines the type of an identifier in the current local scope.
func (ts *TypeScope) Set(name string, t *Type) {
	// log.Printf("Scope (comp: %s, meth: %s): Setting type for '%s' to %s", ts.currentComponent != nil, ts.currentMethod != nil, name, t)
	ts.store[name] = t
}
