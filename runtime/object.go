package runtime

import (
	"errors"
	"fmt"
	"reflect"
)

// Objects are general items with parameters, components and methods
// We have 2 methods in our system so far - Components and Methods
type ObjectInstance struct {
	// Where the component is defined
	File *FileInstance

	// Whether the component is native or not
	// For native components, parameters, dependencies and methods are handled
	// natively via interface methods
	IsNative       bool
	NativeInstance *NativeObject

	// Values for this object's attributes
	Params map[string]Value

	// Initial env for this instance
	InitialEnv *Env[Value]
}

// ===================
// SetParam sets the evaluated parameter value for this instance.
// For DSL components, this sets an Value.
// For Native components, it is upto the component to manage the value of the Value
func (ci *ObjectInstance) Set(name string, value Value) error {
	if ci.IsNative {
		return ci.NativeInstance.Set(name, value)
	} else {
		ci.Params[name] = value
		ci.InitialEnv.Set(name, value)
		return nil
	}
}

func (ci *ObjectInstance) Get(name string) (Value, bool) {
	if ci.IsNative {
		return ci.NativeInstance.Get(name)
	} else {
		node, ok := ci.Params[name]
		return node, ok
	}
}

// NativeObject wraps a Go component instance to implement ComponentRuntime.
type NativeObject struct {
	InstanceName string
	TypeName     string

	// The actual *components.Disk, *components.Cache, etc.
	// Use a goinstance that can already handle some of the runtime methods we need
	// When writing native components, we just enforce these so we can avoid reflection
	GoInstance any
	// GetParam returns the evaluated parameter value for this instance.
	// For DSL components, this returns the Value.
	// For Native components, it needs to retrieve the configured Go value
	// and potentially wrap it in a LeafNode/VarState for consistency? (TBD)
	// GetParam(name string) (Value, bool) // Returns Value, found

	// SetParam sets the evaluated parameter value for this instance.
	// For DSL components, this sets an Value.
	// For Native components, it is upto the component to manage the value of the Value
	// SetParam(name string, value any) error
}

// GetParam for native components is tricky. How do we map a Go field back to an Value?
// Option 1: Don't support GetParam directly for native components via the interface.
// Option 2: Use reflection, get the Go field value, wrap it in LeafNode/VarState. (Complex)
// Let's go with Option 2 for now, but only for simple types.
func (na *NativeObject) Get(name string) (val Value, ok bool) {
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
	/*
		goValue := fieldVal.Interface()
		var valueOutcome any
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
	*/
	return
}

// SetParam sets the evaluated parameter value for this instance.
// For DSL components, this sets an Value.
// For Native components, it is upto the component to manage the value of the Value
func (na *NativeObject) Set(name string, value Value) error {
	instanceVal := reflect.ValueOf(na.GoInstance)
	if instanceVal.Kind() == reflect.Ptr {
		instanceVal = instanceVal.Elem()
	}
	if instanceVal.Kind() != reflect.Struct {
		return errors.New("ErrInvalidNativeObject")
	}

	fieldVal := instanceVal.FieldByName(name) // Assumes direct mapping Name -> FieldName
	if !fieldVal.IsValid() {
		// Try converting param name (e.g., ProfileName) to field name (ProfileName)
		// This is simple if they match case-sensitively. Often needs tags.
		// For now, assume direct match.
		return errors.New("ErrInvalidField")
	}

	// Convert the Go value back to a *VarState -> LeafNode (simplistic)
	if !fieldVal.CanSet() {
		return fmt.Errorf("cannot set value of param '%s'", name)
	}

	/* TODO - Convert Value -> Go Value
	switch value.Type {
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
			// case *BinaryValue:
			// case *IfChoiceNode:
			// case *SequenceNode:
		return fmt.Errorf("node (%s) must be evaluated before setting param ('%s')", reflect.TypeOf(value.Value), name)
	// How to handle *Outcomes[T] fields? Maybe return directly?
	// case *core.Outcomes[core.Duration]: valueOutcome = v // Risky - mutable?
	default:
		return fmt.Errorf("Invalid value type for param ('%s'): '%s')", name, reflect.TypeOf(value.Value))
	}
	*/
	return nil
}

// InvokeMethod uses reflection to call the method on the underlying GoInstance.
/*
func (na *NativeObject) InvokeMethod(methodName string, args []Value, vm *VM, callFrame *Frame) (val Value, err error) {
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

	// Convert Value to go values
	for i := 0; i < methodType.NumIn(); i++ {
		argValue := args[i].Value
		paramType := methodType.In(i)
		argVal := reflect.ValueOf(argValue)

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
*/
