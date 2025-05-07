package decl

import (
	"fmt"
	"strings"

	gfn "github.com/panyam/goutils/fn"
)

type Enum struct {
	Name    string
	Options []string
}

type EnumValue struct {
	Enum   *Enum
	Option int
}

type ValueTypeTag int

const (
	ValueTypeNil ValueTypeTag = iota
	ValueTypeBool
	ValueTypeInt
	ValueTypeFloat // Represents float64
	ValueTypeString
	ValueTypeList     // List Type
	ValueTypeOutcomes // Outcomes - treated like List for now

	// A Tuple value type.  This can be used to a tuple of heterogenous types.
	// The most apt usecase is in modelling Time+Value pairs in Outcome distributions.
	// ie to model what the latency was for a particular value along an execution path.
	// The child types determines the tuple types corresponding to each entry in the tuple type
	ValueTypeTuple

	ValueTypeComponent // Reference to a component that was created.  RuntimeValue.Value holds the component instance
	ValueTypeEnum      // Reference to a component that was created.  RuntimeValue.Value holds enum "case"
	ValueTypeOpNode    // A operator node of operations on outcomes that is collected to be processed later on
)

type ValueType struct {
	Tag        ValueTypeTag
	ChildTypes []*ValueType // Used for List, Outcomes, potentially Func later
}

// --- ValueType Factory Functions ---

var (
	// Use singletons for basic types for efficiency
	NilType       = &ValueType{Tag: ValueTypeNil}
	BoolType      = &ValueType{Tag: ValueTypeBool}
	IntType       = &ValueType{Tag: ValueTypeInt}
	FloatType     = &ValueType{Tag: ValueTypeFloat}
	StrType       = &ValueType{Tag: ValueTypeString}
	EnumType      = &ValueType{Tag: ValueTypeEnum}
	ComponentType = &ValueType{Tag: ValueTypeComponent}
	OpNodeType    = &ValueType{Tag: ValueTypeOpNode}
)

func ListType(elementType *ValueType) *ValueType {
	if elementType == nil {
		panic("List element type cannot be nil")
	}
	return &ValueType{
		Tag:        ValueTypeList,
		ChildTypes: []*ValueType{elementType},
	}
}

func TupleType(elementTypes ...*ValueType) *ValueType {
	if len(elementTypes) == 0 {
		panic("Tuple element type cannot be nil or 0 length")
	}
	return &ValueType{
		Tag:        ValueTypeTuple,
		ChildTypes: elementTypes,
	}
}

func OutcomesType(elementType *ValueType) *ValueType {
	if elementType == nil {
		panic("Outcomes element type cannot be nil")
	}
	return &ValueType{
		Tag:        ValueTypeOutcomes,
		ChildTypes: []*ValueType{elementType},
	}
}

// String representation of the type
func (v *ValueType) String() string {
	switch v.Tag {
	case ValueTypeNil:
		return "nil"
	case ValueTypeBool:
		return "bool"
	case ValueTypeInt:
		return "int"
	case ValueTypeFloat:
		return "float"
	case ValueTypeString:
		return "string"
	case ValueTypeList:
		if len(v.ChildTypes) == 1 && v.ChildTypes[0] != nil {
			return fmt.Sprintf("List[%s]", v.ChildTypes[0].String())
		}
		return "List[?]" // Invalid state
	case ValueTypeTuple:
		return fmt.Sprintf("Tuple[%s]", strings.Join(gfn.Map(v.ChildTypes, func(v *ValueType) string { return v.String() }), ", "))
	case ValueTypeOutcomes:
		if len(v.ChildTypes) == 1 && v.ChildTypes[0] != nil {
			return fmt.Sprintf("Outcomes[%s]", v.ChildTypes[0].String())
		}
		return "Outcomes[?]" // Invalid state
	default:
		// Use reflect to get tag name if possible, otherwise number
		return fmt.Sprintf("UnknownTypeTag(%d)", v.Tag)
	}
}

// Equals checks if two ValueType definitions are equivalent.
func (v *ValueType) Equals(other *ValueType) bool {
	if v == other { // Pointer equality check (useful for singletons)
		return true
	}
	if v == nil || other == nil {
		return false
	}
	if v.Tag != other.Tag {
		return false
	}
	// Tags match, now check children if applicable
	switch v.Tag {
	case ValueTypeList, ValueTypeOutcomes:
		// Must have same number of child types (currently always 1)
		if len(v.ChildTypes) != len(other.ChildTypes) || len(v.ChildTypes) != 1 {
			return false // Should not happen with current factories
		}
		// Recursively check child types
		return v.ChildTypes[0].Equals(other.ChildTypes[0])
	// Add cases for Func, Tuple etc. if they have children
	default:
		// Basic types (or types without children) are equal if tags match
		return true
	}
}

// RuntimeValue wraps a Go value with its type definition.
type RuntimeValue struct {
	Type  *ValueType // The expected runtime type
	Value any        // The underlying Go value
}

// NewRuntimeValue creates a new boxed value, optionally initializing and type-checking.
// If initialValue is provided, Set() is called. Only the first initialValue is used.
func NewRuntimeValue(t *ValueType, initialValue ...any) (*RuntimeValue, error) {
	if t == nil {
		panic("RuntimeValue type cannot be nil")
	}
	rv := &RuntimeValue{
		Type:  t,
		Value: nil, // Default to nil Go value
	}
	// Set Type-specific zero value for Go Value? Or leave as Go nil? Let's leave as Go nil initially.
	// rv.Value = zeroValueForType(t)

	if len(initialValue) > 0 {
		err := rv.Set(initialValue[0])
		if err != nil {
			return nil, fmt.Errorf("failed to initialize RuntimeValue: %w", err)
		}
	}
	return rv, nil
}

// Tries to set the value by enforcing and checking types.
// The input 'v' should be the Go representation corresponding to r.Type.
// For List/Outcomes, 'v' is expected to be '[]*RuntimeValue'.
func (r *RuntimeValue) Set(v any) error {
	if r.Type == nil {
		// Should not happen if constructed properly
		return fmt.Errorf("internal error: RuntimeValue has nil type")
	}

	// Handle general nil case first
	if v == nil {
		if r.Type.Tag == ValueTypeNil {
			r.Value = nil
			return nil
		} else {
			// Allow setting container types to nil Go slice? Maybe. Let's be strict for now.
			return fmt.Errorf("type mismatch: cannot set nil value for expected type %s", r.Type.String())
		}
	}

	// Check type based on expected tag
	switch r.Type.Tag {
	case ValueTypeNil:
		// v is not nil here (handled above)
		return fmt.Errorf("type mismatch: expected nil, got %T", v)

	case ValueTypeBool:
		val, ok := v.(bool)
		if !ok {
			return fmt.Errorf("type mismatch: expected bool, got %T", v)
		}
		r.Value = val
		return nil

	case ValueTypeInt:
		// Allow various Go int types? For simplicity, require 'int' for now.
		val, ok := v.(int)
		if !ok {
			// Could add checks for int32, int64 etc. if needed
			return fmt.Errorf("type mismatch: expected int, got %T", v)
		}
		r.Value = val
		return nil

	case ValueTypeFloat:
		// Use float64 as the standard Go float type
		val, ok := v.(float64)
		if !ok {
			// Could check for float32 if needed
			return fmt.Errorf("type mismatch: expected float64, got %T", v)
		}
		r.Value = val
		return nil

	case ValueTypeString:
		val, ok := v.(string)
		if !ok {
			return fmt.Errorf("type mismatch: expected string, got %T", v)
		}
		r.Value = val
		return nil

	case ValueTypeEnum:
		val, ok := v.(*EnumValue)
		if !ok {
			return fmt.Errorf("type mismatch: expected EnumValue, got %T", v)
		}
		r.Value = val
		return nil

	case ValueTypeComponent:
		val, ok := v.(*ComponentRuntime)
		if !ok {
			return fmt.Errorf("type mismatch: expected ComponentRuntime, got %T", v)
		}
		r.Value = val
		return nil

	case ValueTypeList, ValueTypeOutcomes:
		// Expecting a slice of *RuntimeValue for containers
		listVal, ok := v.([]*RuntimeValue)
		if !ok {
			containerName := "List"
			if r.Type.Tag == ValueTypeOutcomes {
				containerName = "Outcomes"
			}
			return fmt.Errorf("type mismatch: expected %s ([]*RuntimeValue), got %T", containerName, v)
		}

		// Check element types against r.Type.ChildTypes[0]
		expectedElemType := r.Type.ChildTypes[0]
		for i, elem := range listVal {
			if elem == nil {
				// Allow nil elements in lists? Decide based on language semantics.
				// Let's disallow for now unless element type is NilType.
				if expectedElemType.Tag != ValueTypeNil {
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

	default:
		return fmt.Errorf("internal error: unhandled type tag %v in Set", r.Type.Tag)
	}
}

// String representation of the runtime value
func (r *RuntimeValue) String() string {
	if r == nil {
		return "<nil RuntimeValue>"
	}
	valStr := ""
	if r.Value == nil {
		valStr = "<nil>"
	} else {
		// Special handling for list string representation?
		if r.Type.Tag == ValueTypeList || r.Type.Tag == ValueTypeOutcomes {
			if list, ok := r.Value.([]*RuntimeValue); ok {
				var sb strings.Builder
				sb.WriteString("[")
				for i, item := range list {
					if i > 0 {
						sb.WriteString(", ")
					}
					if item == nil {
						sb.WriteString("<nil>") // Or based on Type?
					} else {
						sb.WriteString(item.String()) // Recursive call for element's string
					}
				}
				sb.WriteString("]")
				valStr = sb.String()
			} else {
				// Should not happen if Set worked correctly
				valStr = fmt.Sprintf("invalid list value (%T)", r.Value)
			}
		} else {
			// Default formatting for non-list values
			valStr = fmt.Sprintf("%v", r.Value)
		}
	}
	typeName := "<nil type>"
	if r.Type != nil {
		typeName = r.Type.String()
	}

	return fmt.Sprintf("RV(%s: %s)", typeName, valStr)
}

// --- Custom getter methods
func (r *RuntimeValue) GetInt() (int64, error) {
	if r == nil || r.Type == nil {
		return 0, fmt.Errorf("cannot get Int from nil RuntimeValue")
	}
	if r.Type.Tag != ValueTypeInt {
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

func (r *RuntimeValue) GetBool() (bool, error) {
	if r == nil || r.Type == nil {
		return false, fmt.Errorf("cannot get Bool from nil RuntimeValue")
	}
	if r.Type.Tag != ValueTypeBool {
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

func (r *RuntimeValue) GetFloat() (float64, error) {
	if r == nil || r.Type == nil {
		return 0.0, fmt.Errorf("cannot get Float from nil RuntimeValue")
	}
	if r.Type.Tag != ValueTypeFloat {
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

func (r *RuntimeValue) GetString() (string, error) {
	if r == nil || r.Type == nil {
		return "", fmt.Errorf("cannot get String from nil RuntimeValue")
	}
	if r.Type.Tag != ValueTypeString {
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

func (r *RuntimeValue) GetList() ([]*RuntimeValue, error) {
	if r == nil || r.Type == nil {
		return nil, fmt.Errorf("cannot get List from nil RuntimeValue")
	}
	if r.Type.Tag != ValueTypeList {
		return nil, fmt.Errorf("type mismatch: cannot get List, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		// Return nil slice if internal value is nil (representing empty list)
		return nil, nil
	}
	val, ok := r.Value.([]*RuntimeValue)
	if !ok {
		return nil, fmt.Errorf("internal error: List value is not Go []*RuntimeValue (%T)", r.Value)
	}
	return val, nil
}

func (r *RuntimeValue) GetTuple() ([]*RuntimeValue, error) {
	if r == nil || r.Type == nil {
		return nil, fmt.Errorf("cannot get Tuple from nil RuntimeValue")
	}
	if r.Type.Tag != ValueTypeTuple {
		return nil, fmt.Errorf("type mismatch: cannot get Tuple, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		// Return nil slice if internal value is nil (representing empty list)
		return nil, nil
	}
	val, ok := r.Value.([]*RuntimeValue)
	if !ok {
		return nil, fmt.Errorf("internal error: Tuple value is not Go []*RuntimeValue (%T)", r.Value)
	}
	return val, nil
}

func (r *RuntimeValue) GetOutcomes() ([]*RuntimeValue, error) {
	if r == nil || r.Type == nil {
		return nil, fmt.Errorf("cannot get Outcomes from nil RuntimeValue")
	}
	if r.Type.Tag != ValueTypeOutcomes {
		return nil, fmt.Errorf("type mismatch: cannot get Outcomes, value is type %s", r.Type.String())
	}
	if r.Value == nil {
		// Return nil slice if internal value is nil (representing empty outcomes)
		return nil, nil
	}
	val, ok := r.Value.([]*RuntimeValue)
	if !ok {
		return nil, fmt.Errorf("internal error: Outcomes value is not Go []*RuntimeValue (%T)", r.Value)
	}
	return val, nil
}

// GetNil checks if the value is nil type and holds nil.
// Returns error if type is not NilType or if value is not nil.
func (r *RuntimeValue) GetNil() error {
	if r == nil || r.Type == nil {
		return fmt.Errorf("cannot get Nil from nil RuntimeValue")
	}
	if r.Type.Tag != ValueTypeNil {
		return fmt.Errorf("type mismatch: cannot get Nil, value is type %s", r.Type.String())
	}
	if r.Value != nil {
		return fmt.Errorf("internal error: Nil type has non-nil Go value (%v)", r.Value)
	}
	return nil // Success, it's nil type and holds nil
}

// Helpers to create specific simple values
func StringValue(val string) (out *RuntimeValue) {
	out, _ = NewRuntimeValue(StrType, val)
	return
}

func IntValue(val int64) (out *RuntimeValue) {
	out, _ = NewRuntimeValue(IntType, val)
	return
}

func FloatValue(val float64) (out *RuntimeValue) {
	out, _ = NewRuntimeValue(FloatType, val)
	return
}

func BoolValue(val bool) (out *RuntimeValue) {
	out, _ = NewRuntimeValue(BoolType, val)
	return
}
