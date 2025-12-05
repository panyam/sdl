//go:build js && wasm
// +build js,wasm

// SDL WASM Module - Service Injection Architecture
// This module provides a thin dependency injection layer that wires existing
// service implementations into the generated WASM exports.

package main

import (
	"context"
	"embed"
	"fmt"
	"strings"
	"syscall/js"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	wasmservices "github.com/panyam/sdl/gen/wasm/go/sdl/v1/services"
	"github.com/panyam/sdl/lib/loader"
	"github.com/panyam/sdl/services"
	"github.com/panyam/sdl/services/singleton"
)

// Embed stdlib SDL files
//
//go:embed stdlib/*.sdl
var stdlibFiles embed.FS

// Global filesystem for WASM
var fileSystem loader.FileSystem

func init() {
	// Initialize filesystem for WASM environment
	fileSystem = createWASMFileSystem()
}

// SingletonInitializerService bootstraps the WASM singleton with initial data
type SingletonInitializerService struct {
	CanvasService     *singleton.SingletonCanvasService
	CanvasPresenter   *services.CanvasViewPresenter
	InitialSDLContent string
}

func (s *SingletonInitializerService) InitializeSingleton(ctx context.Context, req *protos.InitializeSingletonRequest) (*protos.InitializeSingletonResponse, error) {
	// Load initial SDL content if provided
	if req.SdlContent != "" {
		// Write to memory filesystem and load
		err := fileSystem.WriteFile("/workspace/init.sdl", []byte(req.SdlContent))
		if err != nil {
			return &protos.InitializeSingletonResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to write initial SDL: %v", err),
			}, nil
		}

		_, err = s.CanvasService.LoadFile(ctx, &protos.LoadFileRequest{
			CanvasId:    "default",
			SdlFilePath: "/workspace/init.sdl",
		})
		if err != nil {
			return &protos.InitializeSingletonResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to load initial SDL: %v", err),
			}, nil
		}
	}

	// Use the specified system if provided
	if req.SystemName != "" {
		_, err := s.CanvasService.UseSystem(ctx, &protos.UseSystemRequest{
			CanvasId:   "default",
			SystemName: req.SystemName,
		})
		if err != nil {
			return &protos.InitializeSingletonResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to use system: %v", err),
			}, nil
		}
	}

	// Initialize the presenter
	initResp, err := s.CanvasPresenter.Initialize(ctx, &protos.InitializePresenterRequest{})
	if err != nil {
		return &protos.InitializeSingletonResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to initialize presenter: %v", err),
		}, nil
	}

	return &protos.InitializeSingletonResponse{
		Success:          true,
		CanvasId:         initResp.CanvasId,
		AvailableSystems: initResp.AvailableSystems,
	}, nil
}

func main() {
	fmt.Println("SDL WASM module loading...")

	// Create singleton services
	canvasService := singleton.NewSingletonCanvasService(fileSystem)
	systemsService := singleton.NewSingletonSystemsService(fileSystem)

	// Create presenter and wire dependencies
	canvasPresenter := services.NewCanvasViewPresenter()
	canvasPresenter.CanvasService = canvasService
	canvasPresenter.SystemsService = systemsService

	// Create initializer service
	initializerService := &SingletonInitializerService{
		CanvasService:   canvasService,
		CanvasPresenter: canvasPresenter,
	}

	// Wire service implementations to generated WASM exports
	exports := &wasmservices.Sdl_v1ServicesExports{
		CanvasService:               canvasService,
		SystemsService:              systemsService,
		CanvasViewPresenter:         canvasPresenter,
		SingletonInitializerService: initializerService,
		CanvasDashboardPage:         wasmservices.NewCanvasDashboardPageClient(),
	}

	// Wire the dashboard page client to the presenter for callbacks
	canvasPresenter.DashboardPage = exports.CanvasDashboardPage

	// Register the JavaScript API using generated exports
	exports.RegisterAPI()

	// Add additional filesystem utilities to the SDL object
	sdlObj := js.Global().Get("sdl")
	if !sdlObj.Truthy() {
		fmt.Println("Warning: sdl object not found after RegisterAPI(), creating it")
		sdlObj = js.ValueOf(map[string]any{})
		js.Global().Set("sdl", sdlObj)
	}

	// Add filesystem utilities
	fsObj := map[string]any{
		"readFile":   js.FuncOf(fsReadFile),
		"writeFile":  js.FuncOf(fsWriteFile),
		"listFiles":  js.FuncOf(fsListFiles),
		"mount":      js.FuncOf(fsMount),
		"isReadOnly": js.FuncOf(fsIsReadOnly),
	}
	sdlObj.Set("fs", js.ValueOf(fsObj))

	// Add configuration utilities
	configObj := map[string]any{
		"setDevMode": js.FuncOf(setDevMode),
	}
	sdlObj.Set("config", js.ValueOf(configObj))

	fmt.Println("SDL WASM module loaded successfully")

	// Keep the WASM module running
	select {}
}

func createWASMFileSystem() loader.FileSystem {
	// Start with a composite filesystem
	cfs := loader.NewCompositeFS()

	// Add memory filesystem for user edits
	cfs.Mount("/workspace/", loader.NewMemoryFS())

	// In production, we'll have bundled files
	// For now, use empty bundles
	bundledFS := loader.NewMemoryFS()
	bundledFS.PreloadFiles(getEmbeddedFiles())
	cfs.Mount("/examples/", bundledFS)
	cfs.Mount("/lib/", bundledFS)

	// Mount stdlib files at @stdlib prefix
	stdlibFS := loader.NewMemoryFS()
	stdlibFS.PreloadFiles(getStdlibFiles())
	cfs.Mount("@stdlib/", stdlibFS)

	// Support for external URLs using WASM fetch API
	cfs.Mount("https://", &URLFetcherFS{})
	cfs.Mount("http://", &URLFetcherFS{})
	cfs.Mount("github.com/", loader.NewGitHubFS())

	return cfs
}

// File system commands
func fsReadFile(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("fs.readFile requires file path")
	}

	path := args[0].String()
	content, err := fileSystem.ReadFile(path)
	if err != nil {
		return jsError(fmt.Sprintf("Failed to read file: %v", err))
	}

	return jsSuccess(map[string]interface{}{
		"content": string(content),
		"path":    path,
	})
}

func fsWriteFile(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return jsError("fs.writeFile requires path and content")
	}

	path := args[0].String()
	content := args[1].String()

	err := fileSystem.WriteFile(path, []byte(content))
	if err != nil {
		return jsError(fmt.Sprintf("Failed to write file: %v", err))
	}

	return jsSuccess(map[string]interface{}{
		"path":    path,
		"message": "File written successfully",
	})
}

func fsListFiles(this js.Value, args []js.Value) interface{} {
	dir := "/"
	if len(args) > 0 {
		dir = args[0].String()
	}

	files, err := fileSystem.ListFiles(dir)
	if err != nil {
		return jsError(fmt.Sprintf("Failed to list files: %v", err))
	}

	// Ensure files is not nil
	if files == nil {
		files = []string{}
	}

	// Convert []string to []interface{} for JavaScript compatibility
	jsFiles := make([]interface{}, len(files))
	for i, f := range files {
		jsFiles[i] = f
	}

	return jsSuccess(map[string]interface{}{
		"files":     jsFiles,
		"directory": dir,
	})
}

func fsMount(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return jsError("fs.mount requires: prefix, url")
	}

	prefix := args[0].String()
	url := args[1].String()

	// Mount the URL to the prefix in our composite filesystem
	if cfs, ok := fileSystem.(*loader.CompositeFS); ok {
		cfs.Mount(prefix, loader.NewHTTPFileSystem(url))
		return jsSuccess(map[string]interface{}{
			"prefix":  prefix,
			"url":     url,
			"message": fmt.Sprintf("Mounted %s to %s", url, prefix),
		})
	}

	return jsError("Mounting only supported with composite filesystem")
}

func fsIsReadOnly(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("fs.isReadOnly requires file path")
	}

	path := args[0].String()

	// Check if the path would be readonly based on mount points
	if _, ok := fileSystem.(*loader.CompositeFS); ok {
		readonlyPrefixes := []string{"/examples/", "/lib/", "/demos/", "https://", "http://", "github.com/"}
		for _, prefix := range readonlyPrefixes {
			if strings.HasPrefix(path, prefix) {
				return jsSuccess(map[string]interface{}{
					"path":       path,
					"isReadOnly": true,
				})
			}
		}
		// Workspace is writable
		if strings.HasPrefix(path, "/workspace/") {
			return jsSuccess(map[string]interface{}{
				"path":       path,
				"isReadOnly": false,
			})
		}
	}

	return jsSuccess(map[string]interface{}{
		"path":       path,
		"isReadOnly": fileSystem.IsReadOnly(),
	})
}

// Configuration commands
func setDevMode(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("setDevMode requires boolean argument")
	}

	devMode := args[0].Bool()

	if devMode {
		// Switch to development filesystem with WASM fetch backend
		fileSystem = NewDevFS()
	} else {
		// Switch to production filesystem (bundled/memory)
		fileSystem = createWASMFileSystem()
	}

	return jsSuccess(map[string]interface{}{
		"devMode": devMode,
		"message": fmt.Sprintf("Development mode set to %v", devMode),
	})
}

// Helper functions

func jsError(message string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error":   message,
	}
}

func jsSuccess(data map[string]interface{}) map[string]interface{} {
	data["success"] = true
	return data
}

// Load embedded SDL library files (placeholder for now)
func getEmbeddedFiles() map[string][]byte {
	return map[string][]byte{}
}

// getStdlibFiles returns the standard library SDL files from embedded FS
func getStdlibFiles() map[string][]byte {
	files := make(map[string][]byte)

	// Read all files from the embedded stdlib directory
	entries, err := stdlibFiles.ReadDir("stdlib")
	if err != nil {
		fmt.Printf("Failed to read stdlib directory: %v\n", err)
		return files
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sdl") {
			data, err := stdlibFiles.ReadFile("stdlib/" + entry.Name())
			if err != nil {
				fmt.Printf("Failed to read stdlib file %s: %v\n", entry.Name(), err)
				continue
			}
			files[entry.Name()] = data
		}
	}

	return files
}
