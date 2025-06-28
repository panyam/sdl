// +build wasm

package main

import (
	"fmt"
	"github.com/panyam/sdl/runtime"
	"github.com/panyam/sdl/loader"
)

// WASMCanvas is a minimal canvas implementation for WASM
// It doesn't have generators or metrics support, just basic SDL evaluation
type WASMCanvas struct {
	id           string
	runtime      *runtime.Runtime
	activeSystem *runtime.SystemInstance
}

func NewWASMCanvas(id string) *WASMCanvas {
	// Use the global filesystem
	fsResolver := loader.NewFileSystemResolver(fileSystem)
	sdlLoader := loader.NewLoader(nil, fsResolver, 10)
	r := runtime.NewRuntime(sdlLoader)
	
	return &WASMCanvas{
		id:      id,
		runtime: r,
	}
}

func (c *WASMCanvas) ID() string {
	return c.id
}

func (c *WASMCanvas) Runtime() *runtime.Runtime {
	return c.runtime
}

func (c *WASMCanvas) Load(filePath string) error {
	_, err := c.runtime.LoadFile(filePath)
	if err != nil {
		return fmt.Errorf("error loading file '%s': %w", filePath, err)
	}
	return nil
}

func (c *WASMCanvas) Use(systemName string) error {
	// Try to create a new system instance
	sysInstance, err := c.runtime.NewSystem(systemName)
	if err != nil {
		return fmt.Errorf("error creating system instance: %w", err)
	}
	c.activeSystem = sysInstance
	return nil
}

func (c *WASMCanvas) CurrentSystem() *runtime.SystemInstance {
	return c.activeSystem
}

// Simple execution without metrics
func (c *WASMCanvas) ExecuteTrace(componentMethod string) (interface{}, error) {
	if c.activeSystem == nil {
		return nil, fmt.Errorf("no active system")
	}
	
	// TODO: Implement trace execution
	// For now, return empty result
	return map[string]interface{}{
		"events": []interface{}{},
	}, nil
}

// Minimal generator support (no actual execution, just tracking)
type WASMGenerator struct {
	ID        string
	Name      string  
	Component string
	Method    string
	Rate      float64
	Enabled   bool
}

// WASMCanvasWithGenerators extends WASMCanvas with basic generator tracking
type WASMCanvasWithGenerators struct {
	*WASMCanvas
	generators map[string]*WASMGenerator
}

func NewWASMCanvasWithGenerators(id string) *WASMCanvasWithGenerators {
	return &WASMCanvasWithGenerators{
		WASMCanvas: NewWASMCanvas(id),
		generators: make(map[string]*WASMGenerator),
	}
}

func (c *WASMCanvasWithGenerators) AddGenerator(name, component, method string, rate float64) error {
	if _, exists := c.generators[name]; exists {
		return fmt.Errorf("generator %s already exists", name)
	}
	
	c.generators[name] = &WASMGenerator{
		ID:        name,
		Name:      name,
		Component: component,
		Method:    method,
		Rate:      rate,
		Enabled:   true,
	}
	
	return nil
}

func (c *WASMCanvasWithGenerators) UpdateGenerator(name string, rate float64) error {
	gen, exists := c.generators[name]
	if !exists {
		return fmt.Errorf("generator %s not found", name)
	}
	
	gen.Rate = rate
	return nil
}

func (c *WASMCanvasWithGenerators) RemoveGenerator(name string) error {
	if _, exists := c.generators[name]; !exists {
		return fmt.Errorf("generator %s not found", name)
	}
	
	delete(c.generators, name)
	return nil
}

func (c *WASMCanvasWithGenerators) ListGenerators() []*WASMGenerator {
	var gens []*WASMGenerator
	for _, gen := range c.generators {
		gens = append(gens, gen)
	}
	return gens
}