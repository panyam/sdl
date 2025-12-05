package loader

import (
	"testing"
	"strings"
)

func TestMemoryFSListFiles(t *testing.T) {
	fs := NewMemoryFS()
	
	// Write some test files
	testFiles := map[string]string{
		"/workspace/test.sdl":      "system Test {}",
		"/workspace/foo.sdl":       "system Foo {}",
		"/workspace/sub/bar.sdl":   "system Bar {}",
		"/examples/demo.sdl":       "system Demo {}",
	}
	
	for path, content := range testFiles {
		err := fs.WriteFile(path, []byte(content))
		if err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}
	
	// Test listing /workspace/
	files, err := fs.ListFiles("/workspace/")
	if err != nil {
		t.Fatalf("Failed to list files in /workspace/: %v", err)
	}
	
	t.Logf("Files in /workspace/: %v", files)
	
	// Check that we got the workspace files
	workspaceFiles := 0
	for _, file := range files {
		if strings.HasPrefix(file, "/workspace/") {
			workspaceFiles++
		}
	}
	
	if workspaceFiles != 3 {
		t.Errorf("Expected 3 files in /workspace/, got %d", workspaceFiles)
	}
}

func TestCompositeFSListFiles(t *testing.T) {
	cfs := NewCompositeFS()
	
	// Create memory filesystems
	workspaceFS := NewMemoryFS()
	examplesFS := NewMemoryFS()
	
	// Mount them
	cfs.Mount("/workspace/", workspaceFS)
	cfs.Mount("/examples/", examplesFS)
	
	// Write test files
	err := workspaceFS.WriteFile("test.sdl", []byte("system Test {}"))
	if err != nil {
		t.Fatalf("Failed to write workspace file: %v", err)
	}
	
	err = examplesFS.WriteFile("demo.sdl", []byte("system Demo {}"))
	if err != nil {
		t.Fatalf("Failed to write examples file: %v", err)
	}
	
	// Test listing through composite FS
	files, err := cfs.ListFiles("/workspace/")
	if err != nil {
		t.Fatalf("Failed to list files in /workspace/: %v", err)
	}
	
	t.Logf("CompositeFS files in /workspace/: %v", files)
	
	// The issue might be here - let's see what findFS returns
	fs, adjustedPath := cfs.findFS("/workspace/")
	t.Logf("findFS(/workspace/) returned fs=%v, adjustedPath=%q", fs != nil, adjustedPath)
	
	// Also test the exact same scenario as in WASM
	testScenario(t, cfs)
}

func testScenario(t *testing.T, fs FileSystem) {
	t.Log("=== Testing WASM scenario ===")
	
	// Step 1: Write file (this works in WASM)
	err := fs.WriteFile("/workspace/test.sdl", []byte("system Test {}"))
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	t.Log("✓ File written successfully")
	
	// Step 2: Read file (this works in WASM)
	content, err := fs.ReadFile("/workspace/test.sdl")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	t.Logf("✓ File read successfully: %q", string(content))
	
	// Step 3: List files (this fails in WASM)
	files, err := fs.ListFiles("/workspace/")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	t.Logf("✓ Files in /workspace/: %v", files)
}

func TestWASMFileSystemScenario(t *testing.T) {
	// Recreate the exact WASM filesystem setup
	cfs := NewCompositeFS()
	
	// Add memory filesystem for user edits
	cfs.Mount("/workspace/", NewMemoryFS())
	
	// In production, we'll have bundled files
	bundledFS := NewMemoryFS()
	bundledFS.PreloadFiles(map[string][]byte{
		"/examples/uber.sdl": []byte(`// Uber MVP example
system UberMVP {
    use api APIGateway
    use db Database
}`),
	})
	cfs.Mount("/examples/", bundledFS)
	cfs.Mount("/lib/", bundledFS)
	
	// Now run the exact same test scenario
	testScenario(t, cfs)
}