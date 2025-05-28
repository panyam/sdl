package decl

import (
	"fmt"
)

type Enum struct {
	Name    string
	Options []string
}

type EnumValue struct {
	Enum   *Enum
	Option int
}

// Value wraps a Go value with its type definition.
type Value struct {
	Type  *Type
	Value any // The underlying Go value
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
	// Set Type-specific zero value for Go Value? Or leave as Go nil? Let's leave as Go nil initially.
	// rv.Value = zeroValueForType(t)

	if len(initialValue) > 0 {
		err := rv.Set(initialValue[0])
		if err != nil {
			return rv, fmt.Errorf("failed to initialize Value: %w", err)
		}
	}
	return rv, nil
}

func (r Value) IsNil() bool {
	return r.Value == nil
}

// Tries to set the value by enforcing and checking types.
// The input 'v' should be the Go representation corresponding to r.Type.
// For List/Outcomes, 'v' is expected to be '[]Value'.
func (r Value) Set(v any) error {
	if r.Type == nil {
		// Should not happen if constructed properly
		return fmt.Errorf("internal error: Value has nil type")
	}

	// Handle general nil case first
	if v == nil {
		if r.Type.Name == "" {
			r.Value = nil
			return nil
		} else {
			// Allow setting container types to nil Go slice? Maybe. Let's be strict for now.
			return fmt.Errorf("type mismatch: cannot set nil value for expected type %s", r.Type.String())
		}
	}

	// Check type based on expected tag
	switch r.Type {
	case NilType:
		// v is not nil here (handled above)
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

	case ComponentType:
		val, ok := v.(*ComponentRuntime)
		if !ok {
			return fmt.Errorf("type mismatch: expected ComponentRuntime, got %T", v)
		}
		r.Value = val
		return nil

	default:
	}

	/* TODO - How to handle unions/enums?
	if r.Type.IsUnion {
		val, ok := v.(*EnumValue)
		if !ok {
			return fmt.Errorf("type mismatch: expected EnumValue, got %T", v)
		}
		r.Value = val
		return nil
	}
	*/

	if r.Type.Name == "List" || r.Type.Name == "Outcomes" {
		// Expecting a slice of Value for containers
		listVal, ok := v.([]Value)
		if !ok {
			containerName := "List"
			if r.Type.Name == "Outcomes" {
				containerName = "Outcomes"
			}
			return fmt.Errorf("type mismatch: expected %s ([]Value), got %T", containerName, v)
		}

		// Check element types against r.Type.ChildTypes[0]
		expectedElemType := r.Type.ChildTypes[0]
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

	return fmt.Errorf("internal error: unhandled type tag %v in Set", r.Type.Name)
}

// String representation of the runtime value
func (r Value) String() string {
	if r.IsNil() {
		return "<nil Value>"
	}
	valStr := fmt.Sprintf("%v", r.Value)
	typeName := "<nil type>"
	if r.Type != nil {
		typeName = r.Type.String()
	}
	return fmt.Sprintf("RV(%s: %s)", typeName, valStr)
}

// --- Custom getter methods
func (r Value) GetInt() (int64, error) {
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
	return val, nil
}

func (r Value) GetBool() (bool, error) {
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
	return val, nil
}

func (r Value) GetFloat() (float64, error) {
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
	return val, nil
}

func (r Value) GetString() (string, error) {
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
	return val, nil
}

func (r Value) GetList() ([]Value, error) {
	if r.IsNil() || r.Type == nil {
		return nil, fmt.Errorf("cannot get List from nil Value")
	}
	if r.Type.Name != "List" {
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
	return val, nil
}

func (r Value) GetTuple() ([]Value, error) {
	if r.IsNil() || r.Type == nil {
		return nil, fmt.Errorf("cannot get Tuple from nil Value")
	}
	if r.Type.Name != "Tuple" {
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
	return val, nil
}

func (r Value) GetOutcomes() ([]Value, error) {
	if r.IsNil() || r.Type == nil {
		return nil, fmt.Errorf("cannot get Outcomes from nil Value")
	}
	if r.Type.Name != "Outcome" {
		return nil, fmt.Errorf("type mismatch: cannot get Outcomes, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		// Return nil slice if internal value is nil (representing empty outcomes)
		return nil, nil
	}
	val, ok := r.Value.([]Value)
	if !ok {
		return nil, fmt.Errorf("internal error: Outcomes value is not Go []Value (%T)", r.Value)
	}
	return val, nil
}

// GetNil checks if the value is nil type and holds nil.
// Returns error if type is not NilType or if value is not nil.
func (r Value) GetNil() error {
	if r.IsNil() || r.Type == nil {
		return fmt.Errorf("cannot get Nil from nil Value")
	}
	if r.Type != nil {
		return fmt.Errorf("type mismatch: cannot get Nil, value is type %s", r.Type.String())
	}
	if r.Value != nil {
		return fmt.Errorf("internal error: Nil type has non-nil Go value (%v)", r.Value)
	}
	return nil // Success, it's nil type and holds nil
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
