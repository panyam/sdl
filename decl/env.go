// sdl/dsl/environment.go
package decl

import (
	"fmt"
)

// References to values
type Ref[T any] struct {
	Value T
}

// Env[T] holds the runtime values for identifiers (variables, functions, components).
// Supports basic scoping via the 'outer' environment.
type Env[T any] struct {
	store map[string]*Ref[T]
	outer *Env[T]
}

// NewEnv[T] creates a new environment nested within an outer one.
// If outer is nil then returns a fresh top-level environment.
// Useful for function calls or block scopes.
func NewEnv[T any](outer *Env[T]) *Env[T] {
	s := make(map[string]*Ref[T])
	return &Env[T]{store: s, outer: outer}
}

// Get retrieves a value by name. It checks the current environment first,
// then recursively checks outer environments.
func (e *Env[T]) GetRef(name string) *Ref[T] {
	ref, ok := e.store[name]
	if (!ok || ref == nil) && e.outer != nil {
		// Not found here, try the outer scope
		ref = e.outer.GetRef(name)
	}
	return ref
}

func (e *Env[T]) Get(name string) (out T, found bool) {
	ref := e.GetRef(name)
	if ref != nil {
		out = ref.Value
		found = true
	}
	return
}

func (e *Env[T]) Set(key string, value T) {
	// Create or update the VarState
	e.store[key] = &Ref[T]{Value: value}
}

// Set multiple key/values at once.
func (e *Env[T]) SetMany(kvpairs map[string]T) {
	for k, v := range kvpairs {
		e.Set(k, v)
	}
}

// Set multiple key/values at once.
func (e *Env[T]) Push() *Env[T] {
	return NewEnv(e)
}

// Copies values from another Env (only top most layer for now)
func (e *Env[T]) CopyFrom(another *Env[T]) *Env[T] {
	for k, v := range another.store {
		e.Set(k, v.Value)
	}
	return e
}

// Extends our environment by creating a new environment and setting values in it
func (e *Env[T]) Extend(kvpairs map[string]T) *Env[T] {
	out := e.Push()
	out.SetMany(kvpairs)
	return out
}

// String representation for debugging
func (e *Env[T]) String() string {
	keys := make([]string, 0, len(e.store))
	for k := range e.store {
		keys = append(keys, k)
	}
	return fmt.Sprintf("Env[T]{store: %v, outer: %v}", keys, e.outer != nil)
}

// Keys returns all keys in this environment (not including outer environments)
func (e *Env[T]) Keys() []string {
	keys := make([]string, 0, len(e.store))
	for k := range e.store {
		keys = append(keys, k)
	}
	return keys
}

// All returns all key-value pairs in this environment (not including outer environments)
func (e *Env[T]) All() map[string]T {
	result := make(map[string]T)
	for k, ref := range e.store {
		if ref != nil {
			result[k] = ref.Value
		}
	}
	return result
}
