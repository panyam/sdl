// +build wasm

package main

import (
	"testing"
	"github.com/panyam/sdl/loader"
)

func TestWASMFileSystem(t *testing.T) {
	// Create the same filesystem as in WASM
	fs := createWASMFileSystem()
	
	// Test write
	err := fs.WriteFile("/workspace/test.sdl", []byte("system Test {}"))
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	
	// Test read
	content, err := fs.ReadFile("/workspace/test.sdl")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	t.Logf("Read content: %s", content)
	
	// Test list - this is where it might fail
	files, err := fs.ListFiles("/workspace/")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	t.Logf("Files: %v", files)
}

// Also test the helper functions
func TestJSHelpers(t *testing.T) {
	// Test jsSuccess
	result := jsSuccess(map[string]interface{}{
		"files": []string{"test1.sdl", "test2.sdl"},
		"directory": "/workspace/",
	})
	
	t.Logf("jsSuccess result: %+v", result)
	
	// Test jsError
	errResult := jsError("test error")
	t.Logf("jsError result: %+v", errResult)
}