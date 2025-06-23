package decl

import (
	"fmt"
	"reflect"

	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

type Value = decl.Value

type WrappedComponent interface {
	Init()
}

type NWBase[W WrappedComponent] struct {
	Name     string
	Modified bool
	Wrapped  W
}

func NewNWBase[W WrappedComponent](name string, wrapped W) NWBase[W] {
	return NWBase[W]{Name: name, Modified: true, Wrapped: wrapped}
}

// SetArrivalRate sets the arrival rate for a specific method.
// This forwards the rate to the underlying disk for contention modeling.
func (s *NWBase[W]) SetArrivalRate(method string, rate float64) error {
	if setter, ok := any(s.Wrapped).(interface{ SetArrivalRate(string, float64) error }); ok {
		return setter.SetArrivalRate(method, rate)
	}
	return fmt.Errorf("WrappedComponent %T has no SetArrivalRate", s.Wrapped)
}

// GetArrivalRate returns the arrival rate for a specific method.
func (s *NWBase[W]) GetArrivalRate(method string) float64 {
	if getter, ok := any(s.Wrapped).(interface{ GetArrivalRate(string) float64 }); ok {
		return getter.GetArrivalRate(method)
	}
	return -1
}

// GetTotalArrivalRate returns the sum of all method arrival rates.
func (s *NWBase[W]) GetTotalArrivalRate() float64 {
	if getter, ok := any(s.Wrapped).(interface{ TotalArrivalRate() float64 }); ok {
		return getter.TotalArrivalRate()
	}
	return -1
}

// GetFlowPattern provides flow analysis, delegating to wrapped component if it supports it
func (b *NWBase[W]) GetFlowPattern(method string, inputRate float64) components.FlowPattern {
	// Check if the wrapped component implements FlowAnalyzable
	if flowAnalyzable, ok := any(b.Wrapped).(components.FlowAnalyzable); ok {
		// Delegate to the component's implementation
		return flowAnalyzable.GetFlowPattern(method, inputRate)
	}

	// Default behavior for non-flow-analyzable components:
	// - No outflows (leaf component)
	// - Perfect success rate (infinite capacity)
	// - Minimal service time
	return components.FlowPattern{
		Outflows:      map[string]float64{}, // No downstream flows
		SuccessRate:   1.0,                  // Always succeeds (infinite capacity)
		Amplification: 1.0,                  // No amplification
		ServiceTime:   0.001,                // 1ms default service time
	}
}

// GetUtilizationInfo provides utilization information, delegating to wrapped component if it supports it
func (b *NWBase[W]) GetUtilizationInfo() []components.UtilizationInfo {
	// Check if the wrapped component implements UtilizationProvider
	if provider, ok := any(b.Wrapped).(components.UtilizationProvider); ok {
		// Delegate to the component's implementation
		infos := provider.GetUtilizationInfo()
		// Update component paths to include this component's name
		for i := range infos {
			if b.Name != "" {
				infos[i].ComponentPath = b.Name + "." + infos[i].ComponentPath
			}
		}
		return infos
	}

	// Default: no utilization info for components that don't provide it
	return []components.UtilizationInfo{}
}

func (n *NWBase[W]) Set(name string, value decl.Value) error {
	n.Modified = true

	// Use reflection to set the field on the wrapped component
	objVal := reflect.ValueOf(n.Wrapped)

	// Dereference pointers until we get to a struct
	for objVal.Kind() == reflect.Ptr {
		if objVal.IsNil() {
			return fmt.Errorf("cannot set field on nil pointer")
		}
		objVal = objVal.Elem()
	}

	// Find the field by name
	field := objVal.FieldByName(name)
	if !field.IsValid() {
		return fmt.Errorf("field '%s' not found in component type %s", name, objVal.Type())
	}

	if !field.CanSet() {
		return fmt.Errorf("field '%s' cannot be set (unexported or immutable)", name)
	}

	// Convert the decl.Value to the appropriate Go type and set the field
	switch field.Kind() {
	case reflect.Float64:
		if value.Type == decl.FloatType {
			field.SetFloat(value.Value.(float64))
		} else if value.Type == decl.IntType {
			field.SetFloat(float64(value.Value.(int64)))
		} else {
			return fmt.Errorf("cannot set field '%s': value of type %s is not assignable to field of type float64",
				name, value.Type)
		}
	case reflect.Int, reflect.Int64:
		if value.Type == decl.IntType {
			field.SetInt(value.Value.(int64))
		} else {
			return fmt.Errorf("cannot set field '%s': value of type %s is not assignable to field of type int",
				name, value.Type)
		}
	case reflect.Bool:
		if value.Type == decl.BoolType {
			field.SetBool(value.Value.(bool))
		} else {
			return fmt.Errorf("cannot set field '%s': value of type %s is not assignable to field of type bool",
				name, value.Type)
		}
	case reflect.String:
		if value.Type == decl.StrType {
			field.SetString(value.Value.(string))
		} else {
			return fmt.Errorf("cannot set field '%s': value of type %s is not assignable to field of type string",
				name, value.Type)
		}
	case reflect.Uint, reflect.Uint64:
		if value.Type == decl.IntType {
			field.SetUint(uint64(value.Value.(int64)))
		} else {
			return fmt.Errorf("cannot set field '%s': value of type %s is not assignable to field of type uint",
				name, value.Type)
		}
	default:
		return fmt.Errorf("unsupported field type for field '%s': %s", name, field.Kind())
	}

	return nil
}

func (n *NWBase[W]) Get(name string) (v decl.Value, ok bool) {
	// Use reflection to get the field from the wrapped component
	objVal := reflect.ValueOf(n.Wrapped)

	// Dereference pointers until we get to a struct
	for objVal.Kind() == reflect.Ptr {
		if objVal.IsNil() {
			return decl.Value{}, false
		}
		objVal = objVal.Elem()
	}

	// Find the field by name
	field := objVal.FieldByName(name)
	if !field.IsValid() {
		return decl.Value{}, false
	}

	// Convert the field value to a decl.Value
	var err error
	switch field.Kind() {
	case reflect.Float64:
		v, err = decl.NewValue(decl.FloatType, field.Float())
	case reflect.Int, reflect.Int64:
		v, err = decl.NewValue(decl.IntType, field.Int())
	case reflect.Bool:
		v, err = decl.NewValue(decl.BoolType, field.Bool())
	case reflect.String:
		v, err = decl.NewValue(decl.StrType, field.String())
	case reflect.Uint, reflect.Uint64:
		v, err = decl.NewValue(decl.IntType, int64(field.Uint()))
	default:
		return decl.Value{}, false
	}

	if err != nil {
		return decl.Value{}, false
	}
	return v, true
}
