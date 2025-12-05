package decl

import (
	"github.com/panyam/sdl/lib/components"
)

type CacheWithContention struct {
	NWBase[*components.CacheWithContention]
}

func NewCacheWithContention(name string) *CacheWithContention {
	wrapped := components.NewCacheWithContention(name)
	return &CacheWithContention{
		NWBase: NewNWBase(name, wrapped),
	}
}

func (c *CacheWithContention) Read() any {
	return OutcomesToValue(c.Wrapped.Read())
}

func (c *CacheWithContention) Write() any {
	return OutcomesToValue(c.Wrapped.Write())
}

func (c *CacheWithContention) Init() {
	c.Wrapped.Init()
}
