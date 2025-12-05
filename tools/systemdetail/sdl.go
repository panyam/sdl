package systemdetail

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/panyam/sdl/lib/decl"
	"github.com/panyam/sdl/lib/loader"
	"github.com/panyam/sdl/lib/runtime"
	"github.com/panyam/sdl/services"
)

// MemoryResolver implements FileResolver for in-memory content
type MemoryResolver struct {
	content  string              // The main SDL content
	stdlibFS loader.FileResolver // For @stdlib imports
}

func NewMemoryResolver(content string, stdlibFS loader.FileResolver) *MemoryResolver {
	return &MemoryResolver{
		content:  content,
		stdlibFS: stdlibFS,
	}
}

func (m *MemoryResolver) Resolve(importerPath, importPath string, open bool) (io.ReadCloser, string, error) {
	// Handle @stdlib imports
	if strings.HasPrefix(importPath, "@stdlib/") {
		if m.stdlibFS != nil {
			return m.stdlibFS.Resolve(importerPath, importPath, open)
		}
		return nil, "", fmt.Errorf("@stdlib import not supported: %s", importPath)
	}

	// For main system file
	if importPath == "/system.sdl" || importPath == "system.sdl" {
		if !open {
			return nil, "/system.sdl", nil
		}
		return io.NopCloser(strings.NewReader(m.content)), "/system.sdl", nil
	}

	// Reject all other imports (local files)
	return nil, "", fmt.Errorf("local imports not allowed: %s", importPath)
}

// validateNoLocalImports checks that SDL content only uses @stdlib imports
func (t *SystemDetailTool) validateNoLocalImports(content string) error {
	// Regex to find import statements
	importRegex := regexp.MustCompile(`(?m)^\s*import\s+.*from\s+["']([^"']+)["']`)
	matches := importRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			importPath := match[1]
			// Allow @stdlib imports, reject everything else
			if !strings.HasPrefix(importPath, "@stdlib/") {
				return fmt.Errorf("local imports not allowed: %s. Only @stdlib imports are permitted", importPath)
			}
		}
	}

	return nil
}

// SetSDLContent sets the SDL content and validates/compiles it
func (t *SystemDetailTool) SetSDLContent(content string) error {
	// Store content
	t.sdlContent = content

	// Reset state
	t.compileResult = nil
	t.diagram = nil

	// Validate no local imports
	if err := t.validateNoLocalImports(content); err != nil {
		result := &CompileResult{
			Success: false,
			Errors:  []string{err.Error()},
		}
		t.compileResult = result
		t.emitError(err.Error())
		return err
	}

	// Compile SDL
	return t.compileSDL()
}

// StdlibPrefixFS wraps a filesystem to handle @stdlib prefix stripping
type StdlibPrefixFS struct {
	fs loader.FileSystem
}

func (s *StdlibPrefixFS) Resolve(importerPath, importPath string, open bool) (io.ReadCloser, string, error) {
	// Strip @stdlib/ prefix if present
	adjustedPath := importPath
	if strings.HasPrefix(importPath, "@stdlib/") {
		adjustedPath = strings.TrimPrefix(importPath, "@stdlib/")
	}

	// Check if file exists
	if !s.fs.Exists(adjustedPath) {
		return nil, "", fmt.Errorf("file not found: %s", importPath)
	}

	canonicalPath := importPath // Keep original path as canonical

	if !open {
		return nil, canonicalPath, nil
	}

	// Read the file
	data, err := s.fs.ReadFile(adjustedPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file %s: %w", importPath, err)
	}

	return io.NopCloser(strings.NewReader(string(data))), canonicalPath, nil
}

// createStdlibFileSystem creates a memory filesystem with @stdlib files
func (t *SystemDetailTool) createStdlibFileSystem() loader.FileResolver {
	// Create memory filesystem for @stdlib files
	stdlibFS := loader.NewMemoryFS()

	// Load stdlib files from examples/stdlib directory
	// Try multiple possible paths to find the stdlib directory
	possiblePaths := []string{
		"examples/stdlib",
		"../../examples/stdlib",    // From tools/systemdetail when running tests
		"../../../examples/stdlib", // In case of deeper nesting
	}

	var stdlibDir string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			stdlibDir = path
			break
		}
	}

	if stdlibDir != "" {
		if files, err := os.ReadDir(stdlibDir); err == nil {
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".sdl") {
					filePath := filepath.Join(stdlibDir, file.Name())
					if data, err := os.ReadFile(filePath); err == nil {
						// Store without the @stdlib/ prefix - the mount handles that
						stdlibFS.WriteFile(file.Name(), data)
					}
				}
			}
		}
	}

	// Wrap with prefix handling
	return &StdlibPrefixFS{fs: stdlibFS}
}

// compileSDL compiles the current SDL content using in-memory resolution
func (t *SystemDetailTool) compileSDL() error {
	if t.sdlContent == "" {
		err := fmt.Errorf("no SDL content to compile")
		t.emitError(err.Error())
		return err
	}

	t.emitInfo("Compiling SDL content...")

	// Create stdlib resolver
	stdlibResolver := t.createStdlibFileSystem()

	// Create memory resolver with current content and stdlib support
	memoryResolver := NewMemoryResolver(t.sdlContent, stdlibResolver)

	// Create new loader with memory resolver and default parser
	loader := loader.NewLoader(nil, memoryResolver, 10)

	// Create runtime and canvas
	runtime := runtime.NewRuntime(loader)
	canvas := services.NewCanvas(t.canvasID, runtime)

	// Update tool with new runtime components
	t.runtime = runtime
	t.canvas = canvas

	// Load the main system file from memory
	fileInstance, err := runtime.LoadFile("/system.sdl")
	if err != nil {
		result := &CompileResult{
			Success: false,
			Errors:  []string{err.Error()},
		}
		t.compileResult = result
		t.emitError(fmt.Sprintf("SDL compilation failed: %s", err.Error()))
		return err
	}

	// Extract available systems from the file instance
	systems := []string{}
	if fileInstance != nil && fileInstance.Decl != nil {
		defs, _ := fileInstance.Decl.AllDefinitions()
		for _, def := range defs {
			if systemDecl, ok := def.(*decl.SystemDecl); ok {
				systems = append(systems, systemDecl.Name.Value)
			}
		}
	}

	result := &CompileResult{
		Success: true,
		Systems: systems,
	}

	t.compileResult = result
	t.emitSuccess(fmt.Sprintf("SDL compiled successfully. Found %d systems: %v", len(systems), systems))

	return nil
}

// CompileSDL compiles the current SDL content and returns the result
func (t *SystemDetailTool) CompileSDL() (*CompileResult, error) {
	err := t.compileSDL()
	return t.compileResult, err
}

// UseSystem activates a specific system for simulation
func (t *SystemDetailTool) UseSystem(systemName string) error {
	if t.compileResult == nil || !t.compileResult.Success {
		return fmt.Errorf("SDL must be compiled successfully before using a system")
	}

	// Check if system exists
	found := false
	for _, sys := range t.compileResult.Systems {
		if sys == systemName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("system '%s' not found. Available systems: %v", systemName, t.compileResult.Systems)
	}

	t.emitInfo(fmt.Sprintf("Using system: %s", systemName))

	// Use the system in the canvas
	if err := t.canvas.Use(systemName); err != nil {
		err = fmt.Errorf("failed to use system '%s': %w", systemName, err)
		t.emitError(err.Error())
		return err
	}

	t.emitSuccess(fmt.Sprintf("Successfully activated system: %s", systemName))
	return nil
}

// ValidateSDL validates the SDL content without fully compiling
func (t *SystemDetailTool) ValidateSDL() error {
	if t.sdlContent == "" {
		return fmt.Errorf("no SDL content to validate")
	}

	// Validate no local imports
	if err := t.validateNoLocalImports(t.sdlContent); err != nil {
		return err
	}

	// For now, validation is the same as compilation
	// In the future, we could implement faster syntax-only validation
	return t.compileSDL()
}

// GetSystemInfo returns information about the loaded systems
func (t *SystemDetailTool) GetSystemInfo() map[string]interface{} {
	if t.compileResult == nil {
		return map[string]interface{}{
			"compiled": false,
		}
	}

	info := map[string]interface{}{
		"compiled":    t.compileResult.Success,
		"systems":     t.compileResult.Systems,
		"systemCount": len(t.compileResult.Systems),
	}

	if !t.compileResult.Success {
		info["errors"] = t.compileResult.Errors
		info["warnings"] = t.compileResult.Warnings
	}

	// Add current system info if available
	currentSystem := t.canvas.CurrentSystem()
	if currentSystem != nil {
		info["activeSystem"] = currentSystem.GetSystemName()
	}

	return info
}
