package decl

import (
	"fmt"

	"github.com/panyam/sdl/core"
)

// A future that was spawned in side a frame
type Future struct {
	// The frame from which this future was called
	CallerFrame *Frame

	// Has the Await been called on this
	Awaited bool

	// The OpNode representing the computation performed by the future's body.
	// This is set when the 'go' statement/expression is evaluated.
	// The Tree Evaluator will process this node when 'wait' is encountered.
	ResultOpNode OpNode
}

type Frame struct {
	outer *Frame

	// Locals in this frame
	locals map[string]Value

	// Futures that were started in this frame the variable names they are bound to
	futures map[string]*Future

	// Keeps track of the curret time within the frame.  Note this is distribution
	// based on the all the paths taken so far
	timeSoFar *core.Outcomes[float64]
}

// NewFrame creates a new frame nested within an outer one.
func NewFrame(outer *Frame) *Frame {
	return &Frame{
		locals:  make(map[string]Value),
		futures: make(map[string]*Future), // Initialize the futures map
		outer:   outer,
	}
}

// Get retrieves a value by name, checking the current frame first,
// then recursively checking outer frames.
func (f *Frame) Get(name string) (Value, bool) {
	obj, ok := f.locals[name]
	if !ok && f.outer != nil {
		// Not found here, try the outer scope
		obj, ok = f.outer.Get(name)
	}
	return obj, ok
}

// Set defines or updates a value in the *current* frame's locals map.
// Use this for function parameters, let bindings, etc., within the current scope.
func (f *Frame) Set(name string, value Value) {
	f.locals[name] = value
}

// GetFuture retrieves future metadata associated with a name, checking current
// and outer frames.
func (f *Frame) GetFuture(name string) (*Future, bool) {
	future, ok := f.futures[name]
	if !ok && f.outer != nil {
		future, ok = f.outer.GetFuture(name)
	}
	return future, ok
}

// AddFuture registers metadata for a future started in this frame.
// Called when evaluating a 'go var = ...' statement/expression.
func (f *Frame) AddFuture(name string, future *Future) {
	// Consider error handling: What if 'name' already exists as a local or another future?
	// For now, allow overwriting, but this might need refinement.
	// if _, exists := f.locals[name]; exists { ... }
	// if _, exists := f.futures[name]; exists { ... }
	f.futures[name] = future
}

// String representation for debugging (optional but helpful).
func (f *Frame) String() string {
	localKeys := make([]string, 0, len(f.locals))
	for k := range f.locals {
		localKeys = append(localKeys, k)
	}
	futureKeys := make([]string, 0, len(f.futures))
	for k := range f.futures {
		futureKeys = append(futureKeys, k)
	}
	return fmt.Sprintf("Frame{locals: %v, futures: %v, outer: %v}", localKeys, futureKeys, f.outer != nil)
}
