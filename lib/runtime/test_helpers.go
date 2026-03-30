package runtime

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/panyam/sdl/lib/core"
	"github.com/panyam/sdl/lib/loader"
	"github.com/stretchr/testify/require"
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

// parseAndLoad parses inline SDL content, loads the first system, and returns it.
// Useful for testing SDL syntax without creating fixture files.
func parseAndLoad(t *testing.T, sdlContent string) *SystemInstance {
	t.Helper()
	sys, err := parseAndLoadSystem(sdlContent)
	require.NoError(t, err)
	require.NotNil(t, sys)
	return sys
}

// parseAndLoadWithError parses inline SDL and returns any error from loading.
func parseAndLoadWithError(t *testing.T, sdlContent string) (*SystemInstance, error) {
	t.Helper()
	return parseAndLoadSystem(sdlContent)
}

func parseAndLoadSystem(sdlContent string) (*SystemInstance, error) {
	l := newTestLoader()
	// Write content to temp file
	tmpFile := filepath.Join(os.TempDir(), "test_sdl_"+fmt.Sprintf("%d", os.Getpid())+".sdl")
	if err := os.WriteFile(tmpFile, []byte(sdlContent), 0644); err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	rt := NewRuntime(l)
	fileInst, err := rt.LoadFile(tmpFile)
	if err != nil {
		return nil, err
	}
	if fileInst == nil {
		return nil, fmt.Errorf("failed to load SDL content")
	}

	// Find the first system
	systems, err := fileInst.Decl.GetSystems()
	if err != nil || len(systems) == 0 {
		return nil, fmt.Errorf("no systems found")
	}
	for name := range systems {
		sys, _ := fileInst.NewSystem(name, true)
		return sys, nil
	}
	return nil, fmt.Errorf("no systems found")
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
