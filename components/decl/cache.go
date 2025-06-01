package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

type Cache struct {
	NWBase[*components.Cache]
}

func NewCache(name string) *Cache {
	return &Cache{NWBase: NewNWBase(name, components.NewCache())}
}

func (h *Cache) Read() (v decl.Value) {
	return OutcomesToValue(h.Wrapped.Read())
}

func (h *Cache) Write() (v decl.Value) {
	return OutcomesToValue(h.Wrapped.Write())
}
