// sdl/dsl/environment.go
package decl

import (
	"fmt"
)

// Env holds the runtime values for identifiers (variables, functions, components).
// Supports basic scoping via the 'outer' environment.
type Env struct {
	store map[string]Node
	outer *Env
}

// NewEnv creates a new environment nested within an outer one.
// If outer is nil then returns a fresh top-level environment.
// Useful for function calls or block scopes.
func NewEnv(outer *Env) *Env {
	s := make(map[string]Node)
	return &Env{store: s, outer: outer}
}

// Get retrieves a value by name. It checks the current environment first,
// then recursively checks outer environments.
func (e *Env) Get(name string) (Node, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		// Not found here, try the outer scope
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set defines or updates a VarState in the *current* environment.
// It takes separate value and latency outcomes and creates/updates the VarState.
/*
func (e *Env) Set(name string, valueOutcome any, latencyOutcome any) {
	// Ensure latency is always *Outcomes[Duration] or nil
	if latencyOutcome != nil {
		if _, ok := latencyOutcome.(*core.Outcomes[core.Duration]); !ok {
			// This indicates an internal error, should not happen if logic is correct
			panic(fmt.Sprintf("internal error: attempting to set non-Duration outcome (%T) as latency for '%s'", latencyOutcome, name))
		}
	}

	// Create or update the VarState
	e.store[name] = &VarState{
		ValueOutcome:   valueOutcome,
		LatencyOutcome: latencyOutcome,
	}
}
*/

func (e *Env) Set(name string, node Node) {
	// Create or update the VarState
	e.store[name] = node
}

// String representation for debugging
func (e *Env) String() string {
	keys := make([]string, 0, len(e.store))
	for k := range e.store {
		keys = append(keys, k)
	}
	return fmt.Sprintf("Env{store: %v, outer: %v}", keys, e.outer != nil)
}
