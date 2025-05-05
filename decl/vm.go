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
func (v *VM) CreateInstance(typeName string, instanceName string, overrides []*AssignmentStmt, frame *Frame) (ComponentRuntime, error) {

	// 1. Get Component Definition (required for both native and DSL)
	compDef, foundDef := v.Entry.Components[typeName]
	if !foundDef {
		// Provide position info if possible - requires passing stmt? Or handling error higher up?
		// Let's return error here, evalInstanceDecl can wrap it.
		return nil, fmt.Errorf("unknown component type definition '%s' for instance '%s'", typeName, instanceName)
	}

	// 2. Check for Native Go Constructor
	constructor, foundConst := v.ComponentRegistry[typeName]
	var runtimeInstance ComponentRuntime

	// --- Branch: Native Go Component ---
	if foundConst {
		// log.Printf("DEBUG: Creating NATIVE instance '%s' of type '%s'", instanceName, typeName)
		overrideParamValues := make(map[string]any)
		dependencyInstances := make(map[string]ComponentRuntime)
		processedOverrides := make(map[string]bool)

		// Evaluate overrides in the provided frame context
		for _, assignStmt := range overrides {
			assignVarName := assignStmt.Var.Name
			processedOverrides[assignVarName] = true

			// Evaluate the RHS of the assignment -> returns OpNode
			valueOpNode, err := Eval(assignStmt.Value, frame, v) // Pass frame and vm
			if err != nil {
				return nil, fmt.Errorf("evaluating override '%s' for native instance '%s': %w", assignVarName, instanceName, err)
			}

			if _, isParam := compDef.Params[assignVarName]; isParam {
				// Parameter assignment: Need to extract raw Go value (Temporary Workaround)
				leaf, ok := valueOpNode.(*LeafNode)
				if !ok {
					return nil, fmt.Errorf("param override '%s' for native instance '%s' requires a literal value (got %T)", assignVarName, instanceName, valueOpNode)
				}
				rawValue, err := extractLeafValue(leaf) // Use Temp Workaround helper
				if err != nil {
					return nil, fmt.Errorf("extracting value for param override '%s' for native instance '%s': %w", assignVarName, instanceName, err)
				}
				overrideParamValues[assignVarName] = rawValue

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
				dependencyInstances[assignVarName] = depRuntime // Store the actual runtime instance

			} else {
				return nil, fmt.Errorf("unknown native override target '%s' for instance '%s'", assignVarName, instanceName)
			}
		}

		// Check 'uses' dependencies satisfied
		for usesName := range compDef.Uses {
			if _, satisfied := dependencyInstances[usesName]; !satisfied {
				return nil, fmt.Errorf("missing override to satisfy 'uses %s: %s' dependency for native instance '%s'", usesName, compDef.Uses[usesName].ComponentRef.Name, instanceName)
			}
		}

		// Instantiate Component via registered Go constructor
		goInstance, err := constructor(instanceName, overrideParamValues)
		if err != nil {
			return nil, fmt.Errorf("failed to construct native component '%s': %w", instanceName, err)
		}

		// Inject Dependencies
		err = injectDependencies(goInstance, dependencyInstances) // Assumes injectDependencies exists
		if err != nil {
			return nil, fmt.Errorf("failed to inject dependencies into native '%s': %w", instanceName, err)
		}

		// Wrap in adapter
		adapter := &NativeComponent{
			InstanceName: instanceName,
			TypeName:     typeName,
			GoInstance:   goInstance,
		}
		runtimeInstance = adapter

	} else { // --- Branch: DSL Component Instance ---
		// log.Printf("DEBUG: Creating DSL instance '%s' of type '%s'", instanceName, typeName)
		dslInstance := &UDComponent{
			Definition:   compDef,
			InstanceName: instanceName,
			Params:       make(map[string]OpNode),
			Dependencies: make(map[string]ComponentRuntime),
		}
		processedOverrides := make(map[string]bool)

		// Evaluate overrides in the provided frame context
		for _, assignStmt := range overrides {
			assignVarName := assignStmt.Var.Name
			processedOverrides[assignVarName] = true

			// Evaluate the RHS -> OpNode
			valueOpNode, err := Eval(assignStmt.Value, frame, v) // Pass frame and vm
			if err != nil {
				return nil, fmt.Errorf("evaluating override '%s' for DSL instance '%s': %w", assignVarName, instanceName, err)
			}

			if _, isParam := compDef.Params[assignVarName]; isParam {
				// Parameter assignment: Store the resulting OpNode directly.
				dslInstance.Params[assignVarName] = valueOpNode

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
				dslInstance.Dependencies[assignVarName] = depRuntime // Store the actual runtime instance

			} else {
				return nil, fmt.Errorf("unknown DSL override target '%s' for instance '%s'", assignVarName, instanceName)
			}
		}

		// Process default parameters for DSL components
		for paramName, paramAST := range compDef.Params {
			if _, overridden := processedOverrides[paramName]; !overridden {
				if paramAST.DefaultValue != nil {
					// Evaluate default value in the *current* frame context
					defaultOpNode, err := Eval(paramAST.DefaultValue, frame, v)
					if err != nil {
						return nil, fmt.Errorf("evaluating default value for param '%s' in DSL instance '%s': %w", paramName, instanceName, err)
					}
					dslInstance.Params[paramName] = defaultOpNode
				} else {
					// Parameter is required but not provided and has no default
					return nil, fmt.Errorf("missing required parameter '%s' for DSL instance '%s'", paramName, instanceName)
				}
			}
		}

		// Check 'uses' dependencies satisfied for DSL components
		for usesName := range compDef.Uses {
			if _, found := dslInstance.Dependencies[usesName]; !found {
				return nil, fmt.Errorf("missing override to satisfy 'uses %s: %s' dependency for DSL instance '%s'", usesName, compDef.Uses[usesName].ComponentRef.Name, instanceName)
			}
		}
		runtimeInstance = dslInstance
	}

	return runtimeInstance, nil
}
