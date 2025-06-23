package runtime

import (
	"github.com/panyam/sdl/decl"
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
	NativeInstance NativeObject

	// Initial env for this instance
	Env *Env[Value]

	// Values for this object's attributes
	params map[string]Value
}

func NewObjectInstance(file *FileInstance, nativeValue NativeObject) ObjectInstance {
	out := ObjectInstance{
		File:           file,
		IsNative:       nativeValue != nil,
		NativeInstance: nativeValue,
		Env:            decl.NewEnv[Value](nil),
	}
	// TODO:
	// 1. SHould Env exists for native objects?
	// 2. Should parent env be that of the file where it is declared?
	if !out.IsNative {
		out.params = make(map[string]Value) // Evaluated parameter Values (override or default)
	}
	return out
}

// ===================
// SetParam sets the evaluated parameter value for this instance.
// For DSL components, this sets an Value.
// For Native components, it is upto the component to manage the value of the Value
func (ci *ObjectInstance) Set(name string, value Value) error {
	if ci.IsNative {
		return ci.NativeInstance.Set(name, value)
	} else {
		ci.params[name] = value
		ci.Env.Set(name, value)
		return nil
	}
}

func (ci *ObjectInstance) Get(name string) (Value, bool) {
	if ci.IsNative {
		return ci.NativeInstance.Get(name)
	} else {
		node, ok := ci.params[name]
		return node, ok
	}
}

type NativeObject interface {
	Set(name string, value Value) error
	Get(name string) (Value, bool)
}
