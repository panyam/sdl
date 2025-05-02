package decl

import (
	"fmt"
	"log"

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

// Constructor function type
// Accepts instance name and a map of *evaluated* override values.
type ComponentConstructor func(instanceName string, params map[string]any) (any, error)

type ComponentDefinition struct {
	Node    *ComponentDecl        // The original AST node
	Params  map[string]*ParamDecl // Processed parameters map[name]*ParamDecl
	Uses    map[string]*UsesDecl  // Processed dependencies map[local_name]*UsesDecl
	Methods map[string]*MethodDef // Processed methods map[method_name]*MethodDef
	// Components map[string]*ComponentDefinition	// Later on add components defined inside a component
	// Add constructor later if needed? Or handle via VM registry only.
}

type VM struct {
	InternalFuncs        map[string]InternalFunction
	SequentialReducers   map[ReducerKey]core.ReducerFunc[any, any, any] // Registry for AND
	MaxOutcomeLen        int
	ComponentRegistry    map[string]ComponentConstructor // Used for currently instantiated components
	ComponentDefRegistry map[string]*ComponentDefinition
}

// NewVM creates a new vm instance.
func (v *VM) Init() {
	if v.MaxOutcomeLen <= 0 {
		v.MaxOutcomeLen = 15 // Default if invalid
	}
	v.InternalFuncs = make(map[string]InternalFunction)
	v.SequentialReducers = make(map[ReducerKey]core.ReducerFunc[any, any, any])
	v.registerDefaultReducers() // Register standard reducers
	v.registerDefaultComponents()
}

// --- Internal Function Registry ---

func (v *VM) RegisterInternalFunc(name string, fn InternalFunction) {
	v.InternalFuncs[name] = fn
}

func (v *VM) RegisterComponent(typeName string, constructor ComponentConstructor) {
	if v.ComponentRegistry == nil {
		v.ComponentRegistry = make(map[string]ComponentConstructor)
	}
	// TODO: Add check for overwrite?
	v.ComponentRegistry[typeName] = constructor
	log.Printf("Registered component type: %s", typeName)
}

func (v *VM) registerDefaultComponents() {
	// Example registration (actual constructor needs to be defined)
	// v.RegisterComponent("Disk", NewDiskComponent)
	// v.RegisterComponent("Cache", NewCacheComponent)
	// ... etc ...
}

// Optional helper method
func (v *VM) RegisterComponentDef(def *ComponentDefinition) error {
	if v.ComponentDefRegistry == nil {
		v.ComponentDefRegistry = make(map[string]*ComponentDefinition)
	}
	name := def.Node.Name.Name
	if _, exists := v.ComponentDefRegistry[name]; exists {
		return fmt.Errorf("component definition '%s' already registered", name)
	}
	v.ComponentDefRegistry[name] = def
	// log.Printf("Registered component definition: %s", name)
	return nil
}
