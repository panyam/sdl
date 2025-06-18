package runtime

import (
	"testing"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/loader"
)

// loadSystem is a test helper that loads an SDL file and initializes a system
func loadSystem(t *testing.T, sdlFile string, systemName string) (*SystemInstance, *Env[Value], core.Duration) {
	// Silence logs during loading
	defer QuietTest(t)()

	l := loader.NewLoader(nil, nil, 10)
	rt := NewRuntime(l)

	// Load the file
	fileStatus, err := l.LoadFile(sdlFile, "", 0)
	if err != nil {
		t.Fatalf("Failed to load file %s: %v", sdlFile, err)
	}

	// Validate
	if !l.Validate(fileStatus) {
		t.Fatalf("Validation failed for %s", sdlFile)
	}

	// Get file instance
	fi := rt.LoadFile(fileStatus.FullPath)
	if fi == nil {
		t.Fatalf("Failed to get file instance for %s", sdlFile)
	}

	// Create system instance
	sys := fi.NewSystem(systemName)
	if sys == nil {
		t.Fatalf("System %s not found in %s", systemName, sdlFile)
	}

	// Initialize the system
	var currTime core.Duration
	se := NewSimpleEval(fi, nil)
	env := fi.Env().Push() // Create new environment for system

	se.EvalInitSystem(sys, env, &currTime)

	sys.Env = env
	return sys, env, currTime
}
