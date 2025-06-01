package decl

import (
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
	panic("TBD")
	return nil
}

func (n *NWBase[W]) Get(name string) (v decl.Value, ok bool) {
	panic("TBD")
	return
}
