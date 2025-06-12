package decl

import (
	"fmt"

	"github.com/panyam/sdl/core"
)

// Value wraps a Go value with its type definition.
type Value struct {
	Type  *Type
	Value any           // The underlying Go value
	Time  core.Duration // The time duration incurred for this value
}

// NewValue creates a new boxed value, optionally initializing and type-checking.
// If initialValue is provided, Set() is called. Only the first initialValue is used.
func NewValue(t *Type, initialValue ...any) (Value, error) {
	if t == nil {
		panic("Value type cannot be nil")
	}
	rv := Value{
		Type:  t,
		Value: nil, // Default to nil Go value
	}
	if len(initialValue) > 0 {
		err := rv.Set(initialValue[0])
		if err != nil {
			return rv, fmt.Errorf("failed to initialize Value: %w", err)
		}
	}
	return rv, nil
}

var Nil = Value{Type: NilType}

func (r *Value) IsTrue() bool {
	return r.Value == true
}

func (r *Value) IsFalse() bool {
	return r.Value == false
}

func (r *Value) IsNil() bool {
	return r.Value == nil
}

func (v *Value) Equals(another *Value) bool {
	if !v.Type.Equals(another.Type) {
		return false
	}
	// TODO - Full comparison
	return true
}

// Tries to set the value by enforcing and checking types.
// The input 'v' should be the Go representation corresponding to r.Type.
// For List/Outcomes, 'v' is expected to be '[]Value'.
func (r *Value) Set(v any) error {
	if r.Type == nil {
		return fmt.Errorf("internal error: Value has nil type")
	}
	if v == nil {
		if r.Type.Tag == TypeTagNil {
			r.Value = nil
			return nil
		} else {
			// Allow setting container types to nil Go slice? Maybe. Let's be strict for now.
			return fmt.Errorf("type mismatch: cannot set nil value for expected type %s", r.Type.String())
		}
	}

	switch r.Type {
	case NilType:
		return fmt.Errorf("type mismatch: expected nil, got %T", v)
	case BoolType:
		val, ok := v.(bool)
		if !ok {
			return fmt.Errorf("type mismatch: expected bool, got %T", v)
		}
		r.Value = val
		return nil

	case IntType:
		// Allow various Go int types? For simplicity, require 'int' for now.
		if intVal, ok := v.(int64); ok {
			r.Value = intVal
		} else if intVal, ok := v.(int); ok {
			r.Value = int64(intVal)
		} else if intVal, ok := v.(int32); ok {
			r.Value = int64(intVal)
		} else if intVal, ok := v.(int16); ok {
			r.Value = int64(intVal)
		} else if intVal, ok := v.(int8); ok {
			r.Value = int64(intVal)
		} else {
			// Could check for uint types if needed
			return fmt.Errorf("type mismatch: expected int64, got %T", v)
		}
		return nil

	case FloatType:
		// Use float64 as the standard Go float type
		if val, ok := v.(float64); ok {
			r.Value = val
		} else if floatVal, ok := v.(float32); ok {
			r.Value = float64(floatVal)
		} else {
			// Could check for other float types if needed
			return fmt.Errorf("type mismatch: expected float64, got %T", v)
		}
		return nil

	case StrType:
		val, ok := v.(string)
		if !ok {
			return fmt.Errorf("type mismatch: expected string, got %T", v)
		}
		r.Value = val
		return nil

	default:
	}

	if r.Type.Tag == TypeTagExpr {
		if le, ok := v.(ThunkValue); !ok {
			return fmt.Errorf("TypeTagExpr needs a ThunkValue")
		} else {
			r.Value = le
		}
	}

	// Take care of the complex types now
	if r.Type.Tag == TypeTagComponent {
		r.Value = v
		return nil
	}

	if r.Type.Tag == TypeTagEnum {
		val, ok := v.(int)
		if !ok {
			return fmt.Errorf("type mismatch: expected enum value as int, got %T", v)
		}
		r.Value = val
		return nil
	}

	if r.Type.Tag == TypeTagRef {
		refval, ok := v.(*RefValue)
		if !ok {
			return fmt.Errorf("expected RefVal, got %T", v)
		}
		r.Value = refval
		return nil
	}

	if r.Type.Tag == TypeTagMethod {
		refval, ok := v.(*MethodValue)
		if !ok {
			return fmt.Errorf("expected MethodVal, got %T", v)
		}
		r.Value = refval
		return nil
	}

	if r.Type.Tag == TypeTagOutcomes {
		outval, ok := v.(*core.Outcomes[Value])
		if !ok {
			return fmt.Errorf("expected Outcomes[value], got %T", v)
		}
		r.Value = outval
		return nil
	}

	if r.Type.Tag == TypeTagFuture {
		outval, ok := v.(*FutureValue)
		if !ok {
			return fmt.Errorf("expected Future, got %T", v)
		}
		r.Value = outval
		return nil
	}

	if r.Type.Tag == TypeTagTuple {
		listVal, ok := v.([]Value)
		if !ok {
			return fmt.Errorf("type mismatch: expected Tuple ([]Value), got %T", v)
		}
		r.Value = listVal
		return nil
	}

	if r.Type.Tag == TypeTagList {
		// Expecting a slice of Value for containers
		listVal, ok := v.([]Value)
		if !ok {
			containerName := "List"
			if r.Type.Tag == TypeTagOutcomes {
				containerName = "Outcomes"
			}
			return fmt.Errorf("type mismatch: expected %s ([]Value), got %T", containerName, v)
		}

		// Check element types against r.Type.ChildTypes[0]
		expectedElemType := r.Type.Info.(*Type)
		for i, elem := range listVal {
			if elem.IsNil() {
				// Allow nil elements in lists? Decide based on language semantics.
				// Let's disallow for now unless element type is NilType.
				if expectedElemType != nil {
					return fmt.Errorf("type error in list/outcomes element %d: got nil, expected %s", i, expectedElemType.String())
				}
				// If nil is allowed, continue
				continue
			}
			if elem.Type == nil {
				// Should not happen
				return fmt.Errorf("internal error: list/outcomes element %d has nil type", i)
			}
			if !elem.Type.Equals(expectedElemType) {
				return fmt.Errorf("type error in list/outcomes element %d: expected %s, got %s", i, expectedElemType.String(), elem.Type.String())
			}
		}
		// All elements okay, assign the slice
		r.Value = listVal
		return nil
	}

	return fmt.Errorf("internal error: unhandled type tag %v in Set", r.Type.Tag)
}

// String representation of the runtime value
func (r *Value) String() string {
	if r.IsNil() {
		return "<nil>"
	}
	valStr := fmt.Sprintf("%v", r.Value)
	typeName := "<nil type>"
	if r.Type != nil {
		typeName = r.Type.String()
	}
	return fmt.Sprintf("Val(%s: %s)", typeName, valStr)
}

// --- Custom getter methods
func (r *Value) IntVal() int64 {
	out, err := r.GetInt()
	if err != nil {
		panic(err)
	}
	return out
}

func (r *Value) FloatVal() float64 {
	out, err := r.GetFloat()
	if err != nil {
		panic(err)
	}
	return out
}

func (r *Value) OutcomesVal() *core.Outcomes[Value] {
	out, err := r.GetOutcomes()
	if err != nil {
		panic(err)
	}
	return out
}

func (r *Value) GetInt() (int64, error) {
	r, err := r.Deref()
	if r.IsNil() || r.Type == nil {
		return 0, fmt.Errorf("cannot get Int from nil Value")
	}
	if r.Type != IntType {
		return 0, fmt.Errorf("type mismatch: cannot get Int, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		return 0, fmt.Errorf("internal error: Int type has nil Go value")
	}
	val, ok := r.Value.(int64)
	if !ok {
		return 0, fmt.Errorf("internal error: Int value is not Go int (%T)", r.Value)
	}
	return val, err
}

func (r *Value) GetBool() (bool, error) {
	r, err := r.Deref()
	if r.IsNil() || r.Type == nil {
		return false, fmt.Errorf("cannot get Bool from nil Value")
	}
	if r.Type != BoolType {
		return false, fmt.Errorf("type mismatch: cannot get Bool, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		return false, fmt.Errorf("internal error: Bool type has nil Go value")
	}
	val, ok := r.Value.(bool)
	if !ok {
		return false, fmt.Errorf("internal error: Bool value is not Go bool (%T)", r.Value)
	}
	return val, err
}

func (r *Value) GetFloat() (float64, error) {
	r, err := r.Deref()
	if r.IsNil() || r.Type == nil {
		return 0.0, fmt.Errorf("cannot get Float from nil Value")
	}
	if r.Type != FloatType {
		return 0.0, fmt.Errorf("type mismatch: cannot get Float, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		return 0.0, fmt.Errorf("internal error: Float type has nil Go value")
	}
	val, ok := r.Value.(float64)
	if !ok {
		return 0.0, fmt.Errorf("internal error: Float value is not Go float64 (%T)", r.Value)
	}
	return val, err
}

func (r *Value) GetString() (string, error) {
	r, err := r.Deref()
	if r.IsNil() || r.Type == nil {
		return "", fmt.Errorf("cannot get String from nil Value")
	}
	if r.Type != StrType {
		return "", fmt.Errorf("type mismatch: cannot get String, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		// Represent nil for string as empty string? Or error? Let's return empty string.
		// Or maybe it should error, as String type shouldn't have nil Go value after Set.
		return "", fmt.Errorf("internal error: String type has nil Go value")
	}
	val, ok := r.Value.(string)
	if !ok {
		return "", fmt.Errorf("internal error: String value is not Go string (%T)", r.Value)
	}
	return val, err
}

func (r *Value) GetList() ([]Value, error) {
	r, err := r.Deref()
	if r.IsNil() || r.Type == nil {
		return nil, fmt.Errorf("cannot get List from nil Value")
	}
	if r.Type.Tag != TypeTagList {
		return nil, fmt.Errorf("type mismatch: cannot get List, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		// Return nil slice if internal value is nil (representing empty list)
		return nil, nil
	}
	val, ok := r.Value.([]Value)
	if !ok {
		return nil, fmt.Errorf("internal error: List value is not Go []Value (%T)", r.Value)
	}
	return val, err
}

func (r *Value) GetTuple() ([]Value, error) {
	r, err := r.Deref()
	if r.IsNil() || r.Type == nil {
		return nil, fmt.Errorf("cannot get Tuple from nil Value")
	}
	if r.Type.Tag != TypeTagTuple {
		return nil, fmt.Errorf("type mismatch: cannot get Tuple, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		// Return nil slice if internal value is nil (representing empty list)
		return nil, nil
	}
	val, ok := r.Value.([]Value)
	if !ok {
		return nil, fmt.Errorf("internal error: Tuple value is not Go []Value (%T)", r.Value)
	}
	return val, err
}

func (r *Value) GetOutcomes() (*core.Outcomes[Value], error) {
	r, err := r.Deref()
	if r.IsNil() || r.Type == nil {
		return nil, fmt.Errorf("cannot get Outcomes from nil Value")
	}
	if r.Type.Tag != TypeTagOutcomes {
		return nil, fmt.Errorf("type mismatch: cannot get Outcomes, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		// Return nil slice if internal value is nil (representing empty outcomes)
		return nil, nil
	}
	val, ok := r.Value.(*core.Outcomes[Value])
	if !ok {
		return nil, fmt.Errorf("internal error: Outcomes value is not Go []Value (%T)", r.Value)
	}
	return val, err
}

func (r *Value) Deref() (*Value, error) {
	if r.IsNil() || r.Type == nil {
		return nil, fmt.Errorf("cannot Deref from nil Value")
	}
	if r.Type.Tag != TypeTagRef {
		return r, nil
	}
	return r.Value.(*RefValue).Deref(), nil
}

// Helpers to create specific simple values
func StringValue(val string) (out Value) {
	out, _ = NewValue(StrType, val)
	return
}

func IntValue(val int64) (out Value) {
	out, _ = NewValue(IntType, val)
	return
}

func FloatValue(val float64) (out Value) {
	out, _ = NewValue(FloatType, val)
	return
}

func BoolValue(val bool) (out Value) {
	out, _ = NewValue(BoolType, val)
	return
}

// Value specific to references of members inside components
type RefValue struct {
	Receiver Value
	Attrib   string
	Deref    func() *Value
}

type MethodValue struct {
	// tells what this method is bound to (eg ComponentInstance if a comp method else nil)
	BoundInstance any
	Method        *MethodDecl
	SavedEnv      *Env[Value]
	IsNative      bool
}

type ThunkValue struct {
	Stmt     Stmt
	SavedEnv *Env[Value]
}

type FutureValue struct {
	// Posible outcomes distribution of *when* the future was kicked off
	StartedAt core.Duration

	// Value of the loop expr if it is a batch future
	LoopValue Value

	// We track the body of the future as a thunk with a captured environment
	// so we can run it later (at Wait time)
	Body ThunkValue

	TraceID int // ID of the 'go' event in the trace
}

func TupleValue(values ...Value) (out Value) {
	valtypes := make([]*Type, len(values))
	for i, v := range values {
		valtypes[i] = v.Type
	}
	tt := TupleType(valtypes...)
	out, _ = NewValue(tt, values)
	return
}
