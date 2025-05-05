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
type ComponentConstructor func(instanceName string) (ComponentRuntime, error)

type VM struct {
	InternalFuncs      map[string]InternalFunction
	SequentialReducers map[ReducerKey]core.ReducerFunc[any, any, any] // Registry for AND
	MaxOutcomeLen      int
	ComponentRegistry  map[string]ComponentConstructor // Used for currently instantiated components
	Entry              *FileDecl
}

// NewVM creates a new vm instance.
func (v *VM) Init() {
	if v.MaxOutcomeLen <= 0 {
		v.MaxOutcomeLen = 15 // Default if invalid
	}
	if v.Entry == nil {
		v.Entry = &FileDecl{}
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

func (v *VM) RegisterNativeComponent(typeName string, constructor ComponentConstructor) {
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

// CreateInstance handles the logic of instantiating either a native or DSL component.
func (v *VM) CreateInstance(typeName string, instanceName string, overrides []*AssignmentStmt, frame *Frame) (runtimeInstance ComponentRuntime, err error) {

	// 1. Get Component Definition (required for both native and DSL)
	compDef, foundDef := v.Entry.Components[typeName]
	if !foundDef {
		// Provide position info if possible - requires passing stmt? Or handling error higher up?
		// Let's return error here, evalInstanceDecl can wrap it.
		return nil, fmt.Errorf("unknown component type definition '%s' for instance '%s'", typeName, instanceName)
	}

	overriddenDependencies := map[string]ComponentRuntime{}
	overriddenParams := map[string]OpNode{}
	for _, assignStmt := range overrides {
		assignVarName := assignStmt.Var.Name
		valueOpNode, err := Eval(assignStmt.Value, frame, v) // Pass frame and vm
		if err != nil {
			return nil, fmt.Errorf("evaluating override '%s' for DSL instance '%s': %w", assignVarName, instanceName, err)
		}

		if _, isParam := compDef.Params[assignVarName]; isParam {
			// Parameter assignment: Store the resulting OpNode directly.
			overriddenParams[assignVarName] = valueOpNode
		} else if _, isUses := compDef.Uses[assignVarName]; isUses {
			// Dependency assignment: Expect RHS to evaluate to InstanceRefNode
			instanceRef, okRef := valueOpNode.(*InstanceRefNode)
			if !okRef {
				return nil, fmt.Errorf("value for 'uses' override '%s' must resolve to a component instance reference, got %T", assignVarName, valueOpNode)
			}
			depInstanceName := instanceRef.InstanceName
			depInstanceAny, foundDep := frame.Get(depInstanceName) // Look up in frame
			if !foundDep {
				return nil, fmt.Errorf("dependency instance '%s' (for '%s.%s') not found", depInstanceName, instanceName, assignVarName)
			}
			depRuntime, okRuntime := depInstanceAny.(ComponentRuntime)
			if !okRuntime {
				return nil, fmt.Errorf("dependency '%s' resolved to non-runtime type %T", depInstanceName, depInstanceAny)
			}
			overriddenDependencies[assignVarName] = depRuntime
		} else {
			return nil, fmt.Errorf("unknown native override target '%s' for instance '%s'", assignVarName, instanceName)
		}
	}

	if compDef.IsNative {
		// 2. Check for Native Go Constructor
		constructor, foundConst := v.ComponentRegistry[typeName]
		if !foundConst {
			return nil, fmt.Errorf("Could not find native constructor for component: '%s'", typeName)
		}

		// Instantiate Component via registered Go constructor
		var err error
		runtimeInstance, err = constructor(instanceName)
		if err != nil {
			return nil, fmt.Errorf("failed to construct native component '%s': %w", instanceName, err)
		}
	} else { // Branch user defined components
		// log.Printf("DEBUG: Creating DSL instance '%s' of type '%s'", instanceName, typeName
		runtimeInstance = &UDComponent{
			Definition:   compDef,
			InstanceName: instanceName,
			Params:       make(map[string]OpNode),
			Dependencies: make(map[string]ComponentRuntime),
		}
	}

	// Process default parameters for DSL components
	for paramName, paramOpNode := range overriddenParams {
		paramDef := compDef.Params[paramName]
		if paramDef == nil {
			return nil, fmt.Errorf("invalid parameter: %s", paramName)
		}
		if err := runtimeInstance.SetParam(paramName, paramOpNode); err != nil {
			return nil, err
		}
	}

	for depName, depInst := range overriddenDependencies {
		depDef := compDef.Uses[depName]
		if depDef == nil {
			return nil, fmt.Errorf("invalid dependency: %s", depName)
		}
		if err := runtimeInstance.SetDependency(depName, depInst); err != nil {
			return nil, err
		}
	}
	return
}
