package runtime

import (
	"testing"

	"github.com/panyam/sdl/lib/core"
	"github.com/panyam/sdl/lib/loader"
)

// loadSystem is a test helper that loads an SDL file and initializes a system
func loadSystem(t *testing.T, sdlFile string, systemName string) (*SystemInstance, core.Duration) {
	// Silence logs during loading
	defer QuietTest(t)()

	l := loader.NewLoader(nil, nil, 10)
	rt := NewRuntime(l)

	// Get file instance
	fileInst, err := rt.LoadFile(sdlFile)
	if fileInst == nil || err != nil {
		t.Fatalf("Failed to get file instance for %s, err: %v", sdlFile, err)
	}

	// Create system instance
	sys, currTime := fileInst.NewSystem(systemName, true)
	if sys == nil {
		t.Fatalf("System %s not found in %s", systemName, sdlFile)
	}
	return sys, currTime
}
