package decl

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/panyam/leetcoach/sdl/core"
)

// ComponentRuntime represents an instantiated component, either native Go or DSL-defined.
// It provides a unified interface for the evaluator to interact with instances.
type ComponentRuntime interface {
	// GetInstanceName returns the specific name this instance was given.
	GetInstanceName() string

	// GetComponentTypeName returns the name of the component type (e.g., "Disk", "MyService").
	GetComponentTypeName() string

	// GetParam returns the evaluated parameter value for this instance.
	// For DSL components, this returns the OpNode.
	// For Native components, it needs to retrieve the configured Go value
	// and potentially wrap it in a LeafNode/VarState for consistency? (TBD)
	GetParam(name string) (OpNode, bool) // Returns OpNode, found

	// GetDependency returns the runtime instance (another ComponentRuntime)
	// satisfying a dependency declared via 'uses'.
	GetDependency(name string) (ComponentRuntime, bool) // Returns ComponentRuntime, found

	// InvokeMethod prepares or executes a method call on this component instance.
	// Args are the evaluated OpNodes for the arguments.
	// Returns an OpNode representing the result of the method call (often a LeafNode
	// wrapping the *core.Outcomes from a native call, or the OpNode tree
	// resulting from evaluating a DSL method body).
	InvokeMethod(methodName string, args []OpNode, vm *VM, callEnv *Env[any]) (OpNode, error)

	// --- Potential Helper Methods ---
	// GetDefinition() *ComponentDefinition // Maybe useful?
}

// --- Existing ComponentInstance adapting to ComponentRuntime ---

// ComponentInstance represents a runtime instance of a component
// defined purely within the DSL. It holds resolved parameters (as OpNodes)
// and dependencies.
type ComponentInstance struct {
	Definition   *ComponentDefinition        // Pointer to the blueprint (AST etc.)
	InstanceName string                      // The name given in the InstanceDecl
	Params       map[string]OpNode           // Evaluated parameter OpNodes (override or default)
	Dependencies map[string]ComponentRuntime // *** Unified map ***
}

// Remove isNativeComponent helper (no longer needed here)

// Stringer for debugging - Update to use single map
func (ci *ComponentInstance) String() string {
	paramNames := make([]string, 0, len(ci.Params))
	for k := range ci.Params {
		paramNames = append(paramNames, k)
	}
	sort.Strings(paramNames)

	depNames := make([]string, 0, len(ci.Dependencies))
	for k, v := range ci.Dependencies {
		// Add type info for clarity
		depNames = append(depNames, fmt.Sprintf("%s:%s", k, v.GetComponentTypeName()))
	}
	sort.Strings(depNames)

	return fmt.Sprintf("DSLInstance<%s name=%s, params=%v, deps=%v>",
		ci.Definition.Node.Name.Name,
		ci.InstanceName,
		paramNames,
		depNames,
	)
}

// GetDependency - Simplify to use single map
func (ci *ComponentInstance) GetDependency(name string) (ComponentRuntime, bool) {
	dep, ok := ci.Dependencies[name]
	// Check if ok and dep is not nil, although map lookup should handle nil?
	// Let's return ok directly from map lookup.
	return dep, ok
}

// Implement ComponentRuntime for *ComponentInstance
func (ci *ComponentInstance) GetInstanceName() string {
	return ci.InstanceName
}

func (ci *ComponentInstance) GetComponentTypeName() string {
	return ci.Definition.Node.Name.Name
}

func (ci *ComponentInstance) GetParam(name string) (OpNode, bool) {
	node, ok := ci.Params[name]
	return node, ok
}

func (ci *ComponentInstance) InvokeMethod(methodName string, args []OpNode, vm *VM, callEnv *Env[any]) (OpNode, error) {
	// 1. Find the Method Definition in the ComponentDefinition
	methodDef, found := ci.Definition.Methods[methodName]
	if !found {
		return nil, fmt.Errorf("method '%s' not found on DSL component '%s' (type %s)", methodName, ci.InstanceName, ci.GetComponentTypeName())
	}

	// 2. Create a new environment for the method call.
	//    The outer env should be the env where the *component instance* lives?
	//    Or should it be the env where the *call* is made? Let's use callEnv for now.
	methodEnv := NewEnv(callEnv)

	// 3. Bind parameters (args) to local variables in methodEnv.
	//    (Needs implementation: check arg count, types?, store OpNodes in methodEnv)
	if len(args) != len(methodDef.Parameters) {
		return nil, fmt.Errorf("argument count mismatch for method '%s': expected %d, got %d", methodName, len(methodDef.Parameters), len(args))
	}
	for i, paramDef := range methodDef.Parameters {
		// Store the provided argument OpNode under the parameter's name
		methodEnv.Set(paramDef.Name.Name, args[i])
	}

	// 4. Bind 'self'/'this' maybe? Store ci (*ComponentInstance) itself?
	methodEnv.Set("self", ci) // Allow methods to access instance params/deps via self.

	// 5. Bind dependencies ('uses') to local variables in methodEnv.
	for depName, depInstance := range ci.Dependencies {
		methodEnv.Set(depName, depInstance) // Store the ComponentRuntime dependency
	}

	// 6. Evaluate the method body (BlockStmt) using the methodEnv.
	//    The result of the block is the result of the method.
	resultOpNode, err := Eval(methodDef.Body, methodEnv, vm)
	if err != nil {
		return nil, fmt.Errorf("error executing method '%s' body for instance '%s': %w", methodName, ci.InstanceName, err)
	}

	// TODO: Handle 'return' statements within the block? Eval needs modification.
	// TODO: Check return type compatibility?

	return resultOpNode, nil
}

// NativeComponentAdapter wraps a Go component instance to implement ComponentRuntime.
type NativeComponentAdapter struct {
	InstanceName string
	TypeName     string
	GoInstance   any // The actual *components.Disk, *components.Cache, etc.
	// Maybe store evaluated params/dependencies used for its creation?
	// ParamsMap    map[string]any // Store the raw Go values used at creation?
}

// Implement ComponentRuntime for *NativeComponentAdapter
func (na *NativeComponentAdapter) GetInstanceName() string {
	return na.InstanceName
}

func (na *NativeComponentAdapter) GetComponentTypeName() string {
	return na.TypeName
}

// GetParam for native components is tricky. How do we map a Go field back to an OpNode?
// Option 1: Don't support GetParam directly for native components via the interface.
// Option 2: Use reflection, get the Go field value, wrap it in LeafNode/VarState. (Complex)
// Let's go with Option 2 for now, but only for simple types.
func (na *NativeComponentAdapter) GetParam(name string) (OpNode, bool) {
	instanceVal := reflect.ValueOf(na.GoInstance)
	if instanceVal.Kind() == reflect.Ptr {
		instanceVal = instanceVal.Elem()
	}
	if instanceVal.Kind() != reflect.Struct {
		return nil, false // Cannot get params from non-struct
	}

	fieldVal := instanceVal.FieldByName(name) // Assumes direct mapping Name -> FieldName
	if !fieldVal.IsValid() {
		// Try converting param name (e.g., ProfileName) to field name (ProfileName)
		// This is simple if they match case-sensitively. Often needs tags.
		// For now, assume direct match.
		return nil, false
	}

	// Convert the Go value back to a *VarState -> LeafNode (simplistic)
	goValue := fieldVal.Interface()
	var valueOutcome any
	latencyOutcome := ZeroLatencyOutcome()
	ok := true

	switch v := goValue.(type) {
	case int64:
		valueOutcome = (&core.Outcomes[int64]{}).Add(1.0, v)
	case int:
		valueOutcome = (&core.Outcomes[int64]{}).Add(1.0, int64(v)) // Convert int
	case uint:
		valueOutcome = (&core.Outcomes[int64]{}).Add(1.0, int64(v)) // Convert uint
	case bool:
		valueOutcome = (&core.Outcomes[bool]{}).Add(1.0, v)
	case string:
		valueOutcome = (&core.Outcomes[string]{}).Add(1.0, v)
	// case float64: valueOutcome = (&core.Outcomes[float64]{}).Add(1.0, v)
	case core.Duration:
		valueOutcome = (&core.Outcomes[core.Duration]{}).Add(1.0, v)
	// How to handle *Outcomes[T] fields? Maybe return directly?
	// case *core.Outcomes[core.Duration]: valueOutcome = v // Risky - mutable?
	default:
		ok = false // Unsupported type to wrap back
	}

	if !ok {
		// Cannot represent parameter value as standard outcome
		return nil, false
	}

	state := &VarState{ValueOutcome: valueOutcome, LatencyOutcome: latencyOutcome}
	return &LeafNode{State: state}, true
}

// GetDependency for native components: Native components don't have DSL 'uses' declarations.
// Their dependencies are Go fields injected during construction. This method might not
// be meaningful or implementable in the same way as for DSL components.
// Return false for now.
func (na *NativeComponentAdapter) GetDependency(name string) (ComponentRuntime, bool) {
	// Dependencies are internal Go fields, not exposed via 'uses' name here.
	// We could use reflection based on field names/tags if needed, but complex.
	return nil, false
}

// InvokeMethod uses reflection to call the method on the underlying GoInstance.
func (na *NativeComponentAdapter) InvokeMethod(methodName string, args []OpNode, vm *VM, callEnv *Env[any]) (OpNode, error) {
	// 1. Find the method on the GoInstance using reflection.
	instanceVal := reflect.ValueOf(na.GoInstance)
	methodVal := instanceVal.MethodByName(methodName)
	if !methodVal.IsValid() {
		return nil, fmt.Errorf("method '%s' not found on native component '%s' (type %T)", methodName, na.InstanceName, na.GoInstance)
	}
	methodType := methodVal.Type()

	// 2. Evaluate argument OpNodes to get concrete values.
	//    *** This requires the Tree Evaluator! ***
	//    Temporary Workaround: Assume args are simple LeafNodes.
	goArgs := make([]reflect.Value, methodType.NumIn()) // NumIn includes receiver if method value is bound
	if !methodType.IsVariadic() && len(args) != methodType.NumIn() {
		return nil, fmt.Errorf("argument count mismatch for native method '%s': expected %d, got %d", methodName, methodType.NumIn(), len(args))
	}
	// TODO: Handle variadic methods

	for i := 0; i < methodType.NumIn(); i++ {
		argOpNode := args[i]
		paramType := methodType.In(i)

		// --- TEMPORARY WORKAROUND ---
		leaf, ok := argOpNode.(*LeafNode)
		if !ok {
			return nil, fmt.Errorf("argument %d for native method '%s' did not evaluate to a simple value (got %T)", i, methodName, argOpNode)
		}
		if leaf.State == nil || leaf.State.ValueOutcome == nil {
			return nil, fmt.Errorf("argument %d value for native method '%s' evaluated to nil state", i, methodName)
		}

		var rawValue any
		var extractOk bool
		switch outcome := leaf.State.ValueOutcome.(type) {
		case *core.Outcomes[int64]:
			rawValue, extractOk = outcome.GetValue()
		case *core.Outcomes[float64]:
			rawValue, extractOk = outcome.GetValue()
		case *core.Outcomes[bool]:
			rawValue, extractOk = outcome.GetValue()
		case *core.Outcomes[string]:
			rawValue, extractOk = outcome.GetValue()
		default:
			return nil, fmt.Errorf("unsupported outcome type %T for native arg %d of method '%s'", outcome, i, methodName)
		}
		if !extractOk {
			return nil, fmt.Errorf("native method argument %d value is probabilistic", i)
		}
		// --- END TEMPORARY WORKAROUND ---

		argVal := reflect.ValueOf(rawValue)

		// Check type compatibility (more robust check needed)
		if !argVal.Type().ConvertibleTo(paramType) {
			return nil, fmt.Errorf("type mismatch for argument %d of native method '%s': expected %s, got %s", i, methodName, paramType, argVal.Type())
		}
		// Convert if necessary (e.g., int64 to int)
		goArgs[i] = argVal.Convert(paramType)
	}

	// 3. Call the method using reflection.
	results := methodVal.Call(goArgs)

	// 4. Process the result. Assume native methods return *core.Outcomes[V] or (result, error).
	if len(results) == 0 {
		// Method returns void, treat as NilNode? Or maybe identity state?
		return &LeafNode{State: createIdentityState()}, nil // Return identity
	}

	// Assume first result is the main one (*core.Outcomes or value)
	// Check for (result, error) pattern
	var returnVal any = results[0].Interface()
	var returnErr error = nil
	if len(results) > 1 {
		if errInter, ok := results[len(results)-1].Interface().(error); ok {
			returnErr = errInter // Last value is error
			if len(results) > 1 {
				returnVal = results[0].Interface() // First is main result
			} else {
				returnVal = nil // Only error was returned
			}
		}
	}
	if returnErr != nil {
		return nil, fmt.Errorf("native method '%s' call failed: %w", methodName, returnErr)
	}
	if returnVal == nil {
		return &LeafNode{State: createNilState()}, nil // Return nil
	}

	// Convert the return value (expected *core.Outcomes[V]) to a VarState -> LeafNode
	resultVarState, err := outcomeToVarState(returnVal)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result of native method '%s' (type %T) to VarState: %w", methodName, returnVal, err)
	}

	return &LeafNode{State: resultVarState}, nil
}
