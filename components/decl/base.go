package decl

import (
	"fmt"
	"reflect"
	
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
