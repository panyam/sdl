//go:build !wasm
// +build !wasm

package main

import (
	"testing"

	"github.com/panyam/sdl/lib/loader"
)

// Test the exact filesystem setup used in WASM
func TestWASMFileSystemSetup(t *testing.T) {
	// Simulate the WASM filesystem setup
	cfs := loader.NewCompositeFS()

	// Add memory filesystem for user edits
	cfs.Mount("/workspace/", loader.NewMemoryFS())

	// In production, we'll have bundled files
	bundledFS := loader.NewMemoryFS()
	bundledFS.PreloadFiles(map[string][]byte{
		"/examples/uber.sdl": []byte(`// Uber MVP example
system UberMVP {
    use api APIGateway
    use db Database
}`),
	})
	cfs.Mount("/examples/", bundledFS)
	cfs.Mount("/lib/", bundledFS)

	// Test the exact sequence from JavaScript
	t.Log("=== Simulating JavaScript test sequence ===")

	// 1. Write file
	err := cfs.WriteFile("/workspace/test.sdl", []byte("system Test {}"))
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	t.Log("✓ WriteFile succeeded")

	// 2. Read file
	content, err := cfs.ReadFile("/workspace/test.sdl")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	t.Logf("✓ ReadFile succeeded: %q", string(content))

	// 3. List files - this is where WASM fails
	files, err := cfs.ListFiles("/workspace/")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	t.Logf("✓ ListFiles succeeded: %v", files)

	// Also test the result that would be returned to JavaScript
	result := map[string]interface{}{
		"success":   true,
		"files":     files,
		"directory": "/workspace/",
	}
	t.Logf("Result that would be sent to JS: %+v", result)
}

// Test potential panic scenarios
func TestPotentialPanics(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Caught panic: %v", r)
		}
	}()

	cfs := loader.NewCompositeFS()
	cfs.Mount("/workspace/", loader.NewMemoryFS())

	// Test with nil files slice
	var nilFiles []string
	result := map[string]interface{}{
		"files":     nilFiles,
		"directory": "/workspace/",
	}
	t.Logf("Result with nil files: %+v", result)

	// Test with empty files slice
	emptyFiles := []string{}
	result2 := map[string]interface{}{
		"files":     emptyFiles,
		"directory": "/workspace/",
	}
	t.Logf("Result with empty files: %+v", result2)
}
