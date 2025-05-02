package decl

import (
	"github.com/panyam/leetcoach/sdl/core"
)

// ReducerKey identifies a pair of types for dispatching reducer functions.
// Using reflect.Type directly might be problematic with generics. Let's try string representation first.
type ReducerKey struct {
	LeftType  string // e.g., "*core.Outcomes[core.AccessResult]"
	RightType string
}

// InternalFunction defines the signature for built-in functions callable by the VM.
// It receives the vm instance (for potential access to env/stack) and evaluated arguments.
type InternalFunction func(vm *VM, args []any) (any, error)

type VM struct {
	InternalFuncs      map[string]InternalFunction
	SequentialReducers map[ReducerKey]core.ReducerFunc[any, any, any] // Registry for AND
	MaxOutcomeLen      int
}

// NewVM creates a new vm instance.
func (v *VM) Init() {
	if v.MaxOutcomeLen <= 0 {
		v.MaxOutcomeLen = 15 // Default if invalid
	}
	v.InternalFuncs = make(map[string]InternalFunction)
	v.SequentialReducers = make(map[ReducerKey]core.ReducerFunc[any, any, any])
	v.registerDefaultReducers() // Register standard reducers
}

// --- Internal Function Registry ---

func (v *VM) RegisterInternalFunc(name string, fn InternalFunction) {
	v.InternalFuncs[name] = fn
}
