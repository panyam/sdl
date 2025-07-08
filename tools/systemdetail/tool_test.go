package systemdetail

import (
	"strings"
	"testing"
)

// Test SDL content that should compile successfully
const testBitlySDL = `
component AppServer {
	param RetryCount = 3
	
	method Shorten() String {
		return "ok"
	}
}

system Bitly {
	use app AppServer
}
`

// Test SDL content with forbidden imports
const testSDLWithLocalImports = `
import Database from "./db.sdl"

system TestSystem {
	use db Database
}
`

// Test SDL content with allowed @stdlib imports
const testSDLWithStdlibImports = `
import Cache from "@stdlib/cache.sdl"

system TestSystem {
	use cache Cache
}
`

func TestNewSystemDetailTool(t *testing.T) {
	tool := NewSystemDetailTool()
	
	if tool == nil {
		t.Fatal("NewSystemDetailTool returned nil")
	}
	
	if tool.GetSDLContent() != "" {
		t.Error("Expected empty SDL content initially")
	}
	
	if tool.GetRecipeContent() != "" {
		t.Error("Expected empty recipe content initially")
	}
	
	if tool.GetSystemID() != "" {
		t.Error("Expected empty system ID initially")
	}
}

func TestSetSDLContent_Valid(t *testing.T) {
	tool := NewSystemDetailTool()
	
	// Set up callback to capture output
	var lastError string
	var lastInfo string
	tool.SetCallbacks(&Callbacks{
		OnError: func(msg string) { lastError = msg },
		OnInfo:  func(msg string) { lastInfo = msg },
	})
	
	// Use variables to avoid unused warnings
	_ = lastError
	_ = lastInfo
	
	err := tool.SetSDLContent(testBitlySDL)
	if err != nil {
		t.Fatalf("SetSDLContent failed: %v", err)
	}
	
	if tool.GetSDLContent() != testBitlySDL {
		t.Error("SDL content not stored correctly")
	}
	
	// Check that compilation was triggered
	result := tool.GetCompileResult()
	if result == nil {
		t.Fatal("Expected compile result after SetSDLContent")
	}
	
	if !result.Success {
		t.Errorf("Expected successful compilation, got errors: %v", result.Errors)
	}
	
	if len(result.Systems) == 0 {
		t.Error("Expected at least one system to be found")
	}
	
	// Should find the Bitly system
	found := false
	for _, sys := range result.Systems {
		if sys == "Bitly" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find 'Bitly' system, got: %v", result.Systems)
	}
}

func TestSetSDLContent_LocalImportsRejected(t *testing.T) {
	tool := NewSystemDetailTool()
	
	err := tool.SetSDLContent(testSDLWithLocalImports)
	if err == nil {
		t.Fatal("Expected error for local imports, got nil")
	}
	
	if !strings.Contains(err.Error(), "local imports not allowed") {
		t.Errorf("Expected 'local imports not allowed' error, got: %v", err)
	}
	
	// Check that compilation result shows failure
	result := tool.GetCompileResult()
	if result == nil || result.Success {
		t.Error("Expected failed compilation result for local imports")
	}
}

func TestSetSDLContent_StdlibImportsAllowed(t *testing.T) {
	tool := NewSystemDetailTool()
	
	// This should not fail validation (though it might fail compilation if @stdlib doesn't exist)
	err := tool.validateNoLocalImports(testSDLWithStdlibImports)
	if err != nil {
		t.Errorf("stdlib imports should be allowed, got error: %v", err)
	}
}

func TestValidateNoLocalImports(t *testing.T) {
	tool := NewSystemDetailTool()
	
	tests := []struct {
		name     string
		content  string
		shouldErr bool
	}{
		{
			name:     "no imports",
			content:  "system Test { }",
			shouldErr: false,
		},
		{
			name:     "stdlib import",
			content:  `import Cache from "@stdlib/cache.sdl"`,
			shouldErr: false,
		},
		{
			name:     "multiple stdlib imports",
			content:  `import Cache from "@stdlib/cache.sdl"\nimport DB from "@stdlib/db.sdl"`,
			shouldErr: false,
		},
		{
			name:     "local import",
			content:  `import DB from "./db.sdl"`,
			shouldErr: true,
		},
		{
			name:     "relative import",
			content:  `import Config from "../config.sdl"`,
			shouldErr: true,
		},
		{
			name:     "mixed imports",
			content:  `import Cache from "@stdlib/cache.sdl"\nimport DB from "./db.sdl"`,
			shouldErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.validateNoLocalImports(tt.content)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	tool := NewSystemDetailTool()
	
	err := tool.Initialize("bitly", testBitlySDL, "")
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	
	if tool.GetSystemID() != "bitly" {
		t.Errorf("Expected system ID 'bitly', got '%s'", tool.GetSystemID())
	}
	
	if tool.GetSDLContent() != testBitlySDL {
		t.Error("SDL content not set by Initialize")
	}
	
	if tool.GetRecipeContent() != "" {
		t.Error("Recipe content should be empty")
	}
}

func TestGetSystemInfo(t *testing.T) {
	tool := NewSystemDetailTool()
	
	// Before compilation
	info := tool.GetSystemInfo()
	if info["compiled"].(bool) != false {
		t.Error("Expected compiled=false before compilation")
	}
	
	// After successful compilation
	err := tool.SetSDLContent(testBitlySDL)
	if err != nil {
		t.Fatalf("SetSDLContent failed: %v", err)
	}
	
	info = tool.GetSystemInfo()
	if info["compiled"].(bool) != true {
		t.Error("Expected compiled=true after successful compilation")
	}
	
	if info["systemCount"].(int) == 0 {
		t.Error("Expected systemCount > 0")
	}
	
	systems, ok := info["systems"].([]string)
	if !ok || len(systems) == 0 {
		t.Error("Expected non-empty systems array")
	}
}

func TestSetSDLContent_WithStdlibImports(t *testing.T) {
	tool := NewSystemDetailTool()
	
	// SDL content that uses @stdlib imports from actual common.sdl
	sdlContent := `
import Cache, HashIndex, HttpStatusCode from "@stdlib/common.sdl";

component TestServer {
  uses cache Cache()
  uses index HashIndex()
  
  method Process() HttpStatusCode {
    if cache.Read() {
      return HttpStatusCode.Ok
    }
    return HttpStatusCode.InternalError
  }
}

system TestSystem {
  use server TestServer
}
`
	
	err := tool.SetSDLContent(sdlContent)
	if err != nil {
		t.Fatalf("SDL with @stdlib imports should compile successfully, got error: %v", err)
	}
	
	// Check compilation result
	result := tool.GetCompileResult()
	if result == nil || !result.Success {
		t.Error("Expected successful compilation")
	}
	
	if len(result.Systems) != 1 || result.Systems[0] != "TestSystem" {
		t.Errorf("Expected 1 system [TestSystem], got %v", result.Systems)
	}
}

func TestUseSystem(t *testing.T) {
	tool := NewSystemDetailTool()
	
	// Set up SDL content with @stdlib imports
	sdlContent := `
import HttpStatusCode from "@stdlib/common.sdl";

component SimpleServer {
  method GetStatus() HttpStatusCode {
    return HttpStatusCode.Ok
  }
}

system WebSystem {
  use server SimpleServer
}
`
	
	err := tool.SetSDLContent(sdlContent)
	if err != nil {
		t.Fatalf("SDL compilation failed: %v", err)
	}
	
	// Test using the system
	err = tool.UseSystem("WebSystem")
	if err != nil {
		t.Fatalf("Failed to use system: %v", err)
	}
	
	// Check system info includes active system
	info := tool.GetSystemInfo()
	if info["activeSystem"] != "WebSystem" {
		t.Errorf("Expected activeSystem=WebSystem, got %v", info["activeSystem"])
	}
	
	// Test using non-existent system
	err = tool.UseSystem("NonExistentSystem")
	if err == nil {
		t.Error("Expected error when using non-existent system")
	}
}