package runtime

import (
	"testing"

	"github.com/panyam/sdl/loader"
)

// Test utility to load a system from a sdl file and initialize it
func loadSystem(t *testing.T, sdlfile, systemName string) (*SystemInstance, *Env[Value]) {
	// Load and test the cascading delays example
	l := loader.NewLoader(nil, nil, 10)
	r := NewRuntime(l)

	fileInstance := r.LoadFile(sdlfile)
	if fileInstance == nil {
		t.Skip("Skipping - cascading_delays.sdl not found")
	}

	system := fileInstance.NewSystem(systemName)
	if system == nil {
		t.Fatalf("System '%s' not found", systemName)
	}

	// Initialize system
	var currTime Duration
	se := NewSimpleEval(fileInstance, nil)
	env := fileInstance.Env()
	se.EvalInitSystem(system, env, &currTime)
	return system, env
}
