package runtime

import (
	"io"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/panyam/sdl/lib/core"
	"github.com/panyam/sdl/lib/loader"
)

// projectRoot returns the SDL project root by walking up from this source file.
func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// lib/runtime/ -> project root is 2 levels up
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// testResolver wraps the default file resolver with @stdlib support.
type testResolver struct {
	defaultResolver *loader.DefaultFileResolver
	fsResolver      *loader.FileSystemResolver
}

func (r *testResolver) Resolve(importerPath, importPath string, open bool) (io.ReadCloser, string, error) {
	// Use filesystem resolver for @-prefixed imports
	if len(importPath) > 0 && importPath[0] == '@' {
		return r.fsResolver.Resolve(importerPath, importPath, open)
	}
	// Fall back to default for regular file paths
	return r.defaultResolver.Resolve(importerPath, importPath, open)
}

// newTestLoader creates a loader with @stdlib mounted for tests that need imports.
func newTestLoader() *loader.Loader {
	cfs := loader.NewCompositeFS()
	stdlibPath := filepath.Join(projectRoot(), "examples", "stdlib")
	cfs.Mount("@stdlib/", loader.NewLocalFS(stdlibPath))
	resolver := &testResolver{
		defaultResolver: loader.NewDefaultFileResolver(),
		fsResolver:      loader.NewFileSystemResolver(cfs),
	}
	return loader.NewLoader(nil, resolver, 10)
}

// loadSystem is a test helper that loads an SDL file and initializes a system.
// Supports @stdlib imports via a CompositeFS with the stdlib mounted.
func loadSystem(t *testing.T, sdlFile string, systemName string) (*SystemInstance, core.Duration) {
	// Silence logs during loading
	defer QuietTest(t)()

	l := newTestLoader()
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
