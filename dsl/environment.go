// sdl/dsl/environment.go
package dsl

import "fmt"

// Environment holds the runtime values for identifiers (variables, functions, components).
// Supports basic scoping via the 'outer' environment.
type Environment struct {
	store map[string]any
	outer *Environment
}

// NewEnvironment creates a fresh top-level environment.
func NewEnvironment() *Environment {
	s := make(map[string]any)
	return &Environment{store: s, outer: nil}
}

// NewEnclosedEnvironment creates a new environment nested within an outer one.
// Useful for function calls or block scopes.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get retrieves a value by name. It checks the current environment first,
// then recursively checks outer environments.
func (e *Environment) Get(name string) (any, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		// Not found here, try the outer scope
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set defines or updates a value in the *current* environment.
// It does not modify outer environments.
func (e *Environment) Set(name string, value any) {
	e.store[name] = value
}

// String representation for debugging (optional)
func (e *Environment) String() string {
	// Simplified representation
	keys := make([]string, 0, len(e.store))
	for k := range e.store {
		keys = append(keys, k)
	}
	return fmt.Sprintf("Env{store: %v, outer: %v}", keys, e.outer != nil)
}
