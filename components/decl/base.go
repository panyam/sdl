package decl

import "github.com/panyam/sdl/decl"

type NWBase struct {
	Name     string
	Modified bool
}

func NewNWBase(name string) NWBase {
	return NWBase{Name: name, Modified: true}
}

func (n *NWBase) Set(name string, value decl.Value) error {
	n.Modified = true
	panic("TBD")
	return nil
}

func (n *NWBase) Get(name string) (v decl.Value, ok bool) {
	panic("TBD")
	return
}
