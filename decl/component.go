package decl

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/panyam/leetcoach/sdl/core"
)

var (
	ErrInvalidField           = errors.New("invalid field")
	ErrInvalidFieldType       = errors.New("invalid field type")
	ErrInvalidNativeComponent = errors.New("invalid native component")
)

// ComponentRuntime represents an instantiated component, either native Go or DSL-defined.
// It provides a unified interface for the evaluator to interact with instances.
type ComponentRuntime interface {
	// GetInstanceName returns the specific name this instance was given.
	GetInstanceName() string

	// GetComponentTypeName returns the name of the component type (e.g., "Disk", "MyService").
	GetComponentTypeName() string

	// GetParam returns the evaluated parameter value for this instance.
	// For DSL components, this returns the RuntimeValue.
	// For Native components, it needs to retrieve the configured Go value
	// and potentially wrap it in a LeafNode/VarState for consistency? (TBD)
	GetParam(name string) (RuntimeValue, bool) // Returns RuntimeValue, found

	// GetDependency returns the runtime instance (another ComponentRuntime)
	// satisfying a dependency declared via 'uses'.
	GetDependency(name string) (ComponentRuntime, bool) // Returns ComponentRuntime, found

	// SetParam sets the evaluated parameter value for this instance.
	// For DSL components, this sets an RuntimeValue.
	// For Native components, it is upto the component to manage the value of the RuntimeValue
	SetParam(name string, value RuntimeValue) error // Returns RuntimeValue, found

	// SetDependency sets the runtime instance (another ComponentRuntime)
	// satisfying a dependency declared via 'uses'.
	SetDependency(name string, comp ComponentRuntime) error // Returns ComponentRuntime, found

	// InvokeMethod prepares or executes a method call on this component instance.
	// Args are the evaluated RuntimeValues for the arguments.
	// Returns an RuntimeValue representing the result of the method call (often a LeafNode
	// wrapping the *core.Outcomes from a native call, or the RuntimeValue tree
	// resulting from evaluating a DSL method body).
	InvokeMethod(methodName string, args []Value, vm *VM, callFrame *Frame) (Value, error)

	// --- Potential Helper Methods ---
	// GetDefinition() *ComponentDefinition // Maybe useful?
}

// --- Existing UDComponent adapting to ComponentRuntime ---

// UDComponent represents a runtime instance of a component
// defined purely within the DSL. It holds resolved parameters (as RuntimeValues)
// and dependencies.
type UDComponent struct {
	Definition   *ComponentDecl              // Pointer to the blueprint (AST etc.)
	InstanceName string                      // The name given in the InstanceDecl
	Params       map[string]RuntimeValue     // Evaluated parameter RuntimeValues (override or default)
	Dependencies map[string]ComponentRuntime // *** Unified map ***
}

// SetParam sets the evaluated parameter value for this instance.
// For DSL components, this sets an RuntimeValue.
// For Native components, it is upto the component to manage the value of the RuntimeValue
func (ci *UDComponent) SetParam(name string, value RuntimeValue) error {
	ci.Params[name] = value
	return nil
}

// SetDependency sets the runtime instance (another ComponentRuntime)
// satisfying a dependency declared via 'uses'.
func (ci *UDComponent) SetDependency(name string, comp ComponentRuntime) error {
	ci.Dependencies[name] = comp
	return nil
}

// Remove isNativeComponent helper (no longer needed here)

// Stringer for debugging - Update to use single map
func (ci *UDComponent) String() string {
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

	if ci.Definition.NameNode.Name == "" {
		panic("component not yet resolved")
	}

	return fmt.Sprintf("DSLInstance<%s name=%s, params=%v, deps=%v>",
		ci.Definition.NameNode.Name,
		ci.InstanceName,
		paramNames,
		depNames,
	)
}

// GetDependency - Simplify to use single map
func (ci *UDComponent) GetDependency(name string) (ComponentRuntime, bool) {
	dep, ok := ci.Dependencies[name]
	// Check if ok and dep is not nil, although map lookup should handle nil?
	// Let's return ok directly from map lookup.
	return dep, ok
}

// Implement ComponentRuntime for *UDComponent
func (ci *UDComponent) GetInstanceName() string {
	return ci.InstanceName
}

func (ci *UDComponent) GetComponentTypeName() string {
	return ci.Definition.NameNode.Name
}

func (ci *UDComponent) GetParam(name string) (RuntimeValue, bool) {
	node, ok := ci.Params[name]
	return node, ok
}

func (ci *UDComponent) InvokeMethod(methodName string, args []Value, vm *VM, callFrame *Frame) (val Value, err error) {
	// 1. Find the Method Definition in the ComponentDefinition
	methodDef, err := ci.Definition.GetMethod(methodName)
	if err != nil {
		return
	}
	if methodDef == nil {
		err = fmt.Errorf("method not found")
		return
	}

	// 2. Create a new frame for the method call.
	//    The outer frame should be the frame where the *component instance* lives?
	//    Or should it be the frame where the *call* is made? Let's use callFrame for now.
	methodFrame := NewFrame(nil)

	// 3. Bind parameters (args) to local variables in methodFrame.
	//    (Needs implementation: check arg count, types?, store RuntimeValues in methodFrame)
	if len(args) != len(methodDef.Parameters) {
		err = fmt.Errorf("argument count mismatch for method '%s': expected %d, got %d", methodName, len(methodDef.Parameters), len(args))
		return
	}
	for i, paramDef := range methodDef.Parameters {
		// Store the provided argument RuntimeValue under the parameter's name
		methodFrame.Set(paramDef.Name.Name, args[i])
	}

	// 4. Bind 'self'/'this' maybe? Store ci (*UDComponent) itself?
	val, err = NewRuntimeValue(ComponentType, ci)
	if err != nil {
		return
	}
	methodFrame.Set("self", val) // Allow methods to access instance params/deps via self.

	// 5. Bind dependencies ('uses') to local variables in methodFrame.
	for depName, depInstance := range ci.Dependencies {
		depVal, err := NewRuntimeValue(ComponentType, depInstance)
		if err != nil {
			return nil, err
		}
		methodFrame.Set(depName, depVal) // Store the ComponentRuntime dependency
	}

	// 6. Evaluate the method body (BlockStmt) using the methodFrame.
	//    The result of the block is the result of the method.
	resultRuntimeValue, err := Eval(methodDef.Body, methodFrame, vm)
	if err != nil {
		err = fmt.Errorf("error executing method '%s' body for instance '%s': %w", methodName, ci.InstanceName, err)
		return
	}

	// TODO: Handle 'return' statements within the block? Eval needs modification.
	// TODO: Check return type compatibility?

	return resultRuntimeValue, nil
}

// NativeComponent wraps a Go component instance to implement ComponentRuntime.
type NativeComponent struct {
	InstanceName string
	TypeName     string

	// The actual *components.Disk, *components.Cache, etc.
	// Use a goinstance that can already handle some of the runtime methods we need
	// When writing native components, we just enforce these so we can avoid reflection
	GoInstance interface {
		// GetParam returns the evaluated parameter value for this instance.
		// For DSL components, this returns the RuntimeValue.
		// For Native components, it needs to retrieve the configured Go value
		// and potentially wrap it in a LeafNode/VarState for consistency? (TBD)
		// GetParam(name string) (RuntimeValue, bool) // Returns RuntimeValue, found

		// GetDependency returns the runtime instance (another ComponentRuntime)
		// satisfying a dependency declared via 'uses'.
		// GetDependency(name string) (ComponentRuntime, bool) // Returns ComponentRuntime, found

		// SetParam sets the evaluated parameter value for this instance.
		// For DSL components, this sets an RuntimeValue.
		// For Native components, it is upto the component to manage the value of the RuntimeValue
		// SetParam(name string, value any) error

		// SetDependency sets the runtime instance (another ComponentRuntime)
		// satisfying a dependency declared via 'uses'.
		// SetDependency(name string, comp ComponentRuntime) error
	}
}

// Implement ComponentRuntime for *NativeComponent
func (na *NativeComponent) GetInstanceName() string {
	return na.InstanceName
}

func (na *NativeComponent) GetComponentTypeName() string {
	return na.TypeName
}

// GetParam for native components is tricky. How do we map a Go field back to an RuntimeValue?
// Option 1: Don't support GetParam directly for native components via the interface.
// Option 2: Use reflection, get the Go field value, wrap it in LeafNode/VarState. (Complex)
// Let's go with Option 2 for now, but only for simple types.
func (na *NativeComponent) GetParam(name string) (val RuntimeValue, ok bool) {
	instanceVal := reflect.ValueOf(na.GoInstance)
	if instanceVal.Kind() == reflect.Ptr {
		instanceVal = instanceVal.Elem()
	}
	if instanceVal.Kind() != reflect.Struct {
		return
	}

	fieldVal := instanceVal.FieldByName(name) // Assumes direct mapping Name -> FieldName
	if !fieldVal.IsValid() {
		// Try converting param name (e.g., ProfileName) to field name (ProfileName)
		// This is simple if they match case-sensitively. Often needs tags.
		// For now, assume direct match.
		return
	}

	// Convert the Go value back to a *VarState -> LeafNode (simplistic)
	goValue := fieldVal.Interface()
	var valueOutcome any
	latencyOutcome := ZeroLatencyOutcome()
	ok = true

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
		return
	}

	state := &VarState{ValueOutcome: valueOutcome, LatencyOutcome: latencyOutcome}
	val.Type = OpNodeType
	val.Value = &LeafNode{State: state}
	return
}

// GetDependency for native components: Native components don't have DSL 'uses' declarations.
// Their dependencies are Go fields injected during construction. This method might not
// be meaningful or implementable in the same way as for DSL components.
// Return false for now.
func (na *NativeComponent) GetDependency(name string) (ComponentRuntime, bool) {
	// Dependencies are internal Go fields, not exposed via 'uses' name here.
	// We could use reflection based on field names/tags if needed, but complex.
	return nil, false
}

// SetParam sets the evaluated parameter value for this instance.
// For DSL components, this sets an RuntimeValue.
// For Native components, it is upto the component to manage the value of the RuntimeValue
func (na *NativeComponent) SetParam(name string, value RuntimeValue) error {
	instanceVal := reflect.ValueOf(na.GoInstance)
	if instanceVal.Kind() == reflect.Ptr {
		instanceVal = instanceVal.Elem()
	}
	if instanceVal.Kind() != reflect.Struct {
		return ErrInvalidNativeComponent
	}

	fieldVal := instanceVal.FieldByName(name) // Assumes direct mapping Name -> FieldName
	if !fieldVal.IsValid() {
		// Try converting param name (e.g., ProfileName) to field name (ProfileName)
		// This is simple if they match case-sensitively. Often needs tags.
		// For now, assume direct match.
		return ErrInvalidField
	}

	// Convert the Go value back to a *VarState -> LeafNode (simplistic)
	if !fieldVal.CanSet() {
		return fmt.Errorf("cannot set value of param '%s'", name)
	}

	switch value.Type.Tag {
	case ValueTypeNil:
		fieldVal.SetPointer(nil)
		break
	case ValueTypeInt:
		if val, err := value.GetInt(); err != nil {
			return err
		} else {
			fieldVal.SetInt(val)
		}
		break
	case ValueTypeBool:
		if val, err := value.GetBool(); err != nil {
			return err
		} else {
			fieldVal.SetBool(val)
		}
		break
	case ValueTypeFloat:
		if val, err := value.GetFloat(); err != nil {
			return err
		} else {
			fieldVal.SetFloat(val)
		}
		break
	case ValueTypeString:
		if val, err := value.GetString(); err != nil {
			return err
		} else {
			fieldVal.SetString(val)
		}
		break
	case ValueTypeOpNode:
		/*
			case *BinaryRuntimeValue:
			case *IfChoiceNode:
			case *SequenceNode:
		*/
		return fmt.Errorf("node (%s) must be evaluated before setting param ('%s')", reflect.TypeOf(value.Value), name)
	// How to handle *Outcomes[T] fields? Maybe return directly?
	// case *core.Outcomes[core.Duration]: valueOutcome = v // Risky - mutable?
	default:
		return fmt.Errorf("Invalid value type for param ('%s'): '%s')", name, reflect.TypeOf(value.Value))
	}
	return nil
}

// SetDependency sets the runtime instance (another ComponentRuntime)
// satisfying a dependency declared via 'uses'.
func (ci *NativeComponent) SetDependency(name string, comp ComponentRuntime) error {
	// Cannot override dependeencies in native cmponents for now
	return nil
}

// InvokeMethod uses reflection to call the method on the underlying GoInstance.
func (na *NativeComponent) InvokeMethod(methodName string, args []Value, vm *VM, callFrame *Frame) (val Value, err error) {
	// 1. Find the method on the GoInstance using reflection.
	instanceVal := reflect.ValueOf(na.GoInstance)
	methodVal := instanceVal.MethodByName(methodName)
	if !methodVal.IsValid() {
		err = fmt.Errorf("method '%s' not found on native component '%s' (type %T)", methodName, na.InstanceName, na.GoInstance)
		return
	}
	methodType := methodVal.Type()

	// 2. Here we assume that the caller already has all the values evaluated before entering here.
	// Temporary Workaround - Only caveat is the OpNode types - we will assume those are not what are being passed around for now.
	goArgs := make([]reflect.Value, methodType.NumIn()) // NumIn includes receiver if method value is bound
	if !methodType.IsVariadic() && len(args) != methodType.NumIn() {
		err = fmt.Errorf("argument count mismatch for native method '%s': expected %d, got %d", methodName, methodType.NumIn(), len(args))
		return
	}
	// TODO: Handle variadic methods

	// Convert RuntimeValue to go values
	for i := 0; i < methodType.NumIn(); i++ {
		argRuntimeValue := args[i].Value
		paramType := methodType.In(i)
		argVal := reflect.ValueOf(argRuntimeValue)

		// Check type compatibility (more robust check needed)
		if !argVal.Type().ConvertibleTo(paramType) {
			err = fmt.Errorf("type mismatch for argument %d of native method '%s': expected %s, got %s", i, methodName, paramType, argVal.Type())
			return
		}
		// Convert if necessary (e.g., int64 to int)
		goArgs[i] = argVal.Convert(paramType)
	}

	// 3. Call the method using reflection.
	results := methodVal.Call(goArgs)

	// 4. Process the result. Assume native methods return *core.Outcomes[V] or (result, error).
	if len(results) == 0 {
		// Method returns void, treat as NilNode? Or maybe identity state?
		val.Type = OpNodeType
		val.Value = &LeafNode{State: createIdentityState()}
		return
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
		err = fmt.Errorf("native method '%s' call failed: %w", methodName, returnErr)
		return
	}
	if returnVal == nil {
		val.Type = OpNodeType
		val.Value = &LeafNode{State: createNilState()}
		return
	}

	// Convert the return value (expected *core.Outcomes[V]) to a VarState -> LeafNode
	resultVarState, err := outcomeToVarState(returnVal)
	if err != nil {
		err = fmt.Errorf("failed to convert result of native method '%s' (type %T) to VarState: %w", methodName, returnVal, err)
		return
	}

	val.Type = OpNodeType
	val.Value = &LeafNode{State: resultVarState}
	return
}
