// +build wasm

package main

import (
	"syscall/js"
	"fmt"
	"strings"
	"time"
	"github.com/panyam/sdl/console"
	"github.com/panyam/sdl/loader"
)

// Global canvas manager - reusing existing Canvas type
var canvases map[string]*console.Canvas
var fileSystem loader.FileSystem

func init() {
	canvases = make(map[string]*console.Canvas)
	// Initialize with default canvas
	canvases["default"] = console.NewCanvas("default")
	
	// Initialize filesystem for WASM environment
	fileSystem = createWASMFileSystem()
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
	
	// Support for external URLs
	cfs.Mount("https://", NewWASMHTTPFS())
	cfs.Mount("http://", NewWASMHTTPFS())
	cfs.Mount("github.com/", loader.NewGitHubFS())
	
	return cfs
}

func main() {
	fmt.Println("SDL WASM initialized")
	
	// Create global SDL object that mirrors CLI commands
	sdl := js.ValueOf(map[string]interface{}{
		"version": "0.1.0",
		
		// Canvas commands (sdl canvas ...)
		"canvas": map[string]interface{}{
			"load":   js.FuncOf(canvasLoad),
			"use":    js.FuncOf(canvasUse),
			"info":   js.FuncOf(canvasInfo),
			"list":   js.FuncOf(canvasList),
			"reset":  js.FuncOf(canvasReset),
			"remove": js.FuncOf(canvasRemove),
		},
		
		// Generator commands (sdl gen ...)
		"gen": map[string]interface{}{
			"add":    js.FuncOf(genAdd),
			"remove": js.FuncOf(genRemove),
			"update": js.FuncOf(genUpdate),
			"list":   js.FuncOf(genList),
			"start":  js.FuncOf(genStart),
			"stop":   js.FuncOf(genStop),
		},
		
		// Metrics commands (sdl metrics ...)
		"metrics": map[string]interface{}{
			"add":    js.FuncOf(metricsAdd),
			"remove": js.FuncOf(metricsRemove),
			"update": js.FuncOf(metricsUpdate),
			"list":   js.FuncOf(metricsList),
			"query":  js.FuncOf(metricsQuery),
		},
		
		// Execution commands
		"run":   js.FuncOf(run),
		"trace": js.FuncOf(trace),
		"flows": js.FuncOf(flows),
		
		// File system access
		"fs": map[string]interface{}{
			"readFile":  js.FuncOf(fsReadFile),
			"writeFile": js.FuncOf(fsWriteFile),
			"listFiles": js.FuncOf(fsListFiles),
			"mount":     js.FuncOf(fsMount), // For dev server mounting
		},
		
		// Configuration
		"config": map[string]interface{}{
			"setDevMode": js.FuncOf(setDevMode),
		},
	})
	
	js.Global().Set("SDL", sdl)
	
	// Keep the program running
	select {}
}

// Helper to get or create canvas
func getCanvas(id string) *console.Canvas {
	if id == "" {
		id = "default"
	}
	
	canvas, exists := canvases[id]
	if !exists {
		canvas = console.NewCanvas(id)
		canvases[id] = canvas
	}
	return canvas
}

// Canvas commands implementation

func canvasLoad(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("canvas.load requires recipe file path")
	}
	
	recipePath := args[0].String()
	canvasID := "default"
	if len(args) > 1 {
		canvasID = args[1].String()
	}
	
	canvas := getCanvas(canvasID)
	
	// Canvas doesn't expose runtime directly anymore
	// We'll need to load through the canvas directly
	
	// Load the recipe
	err := canvas.Load(recipePath)
	if err != nil {
		return jsError(fmt.Sprintf("Failed to load recipe: %v", err))
	}
	
	// Get available systems
	systems := canvas.GetAvailableSystemNames()
	
	return jsSuccess(map[string]interface{}{
		"canvasId": canvasID,
		"recipe": recipePath,
		"systems": len(systems),
		"message": fmt.Sprintf("Loaded recipe %s into canvas %s", recipePath, canvasID),
	})
}

func canvasUse(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("canvas.use requires system name")
	}
	
	systemName := args[0].String()
	canvasID := "default"
	if len(args) > 1 {
		canvasID = args[1].String()
	}
	
	canvas := getCanvas(canvasID)
	
	err := canvas.Use(systemName)
	if err != nil {
		return jsError(fmt.Sprintf("Failed to use system: %v", err))
	}
	
	return jsSuccess(map[string]interface{}{
		"canvasId": canvasID,
		"system": systemName,
		"message": fmt.Sprintf("Using system %s in canvas %s", systemName, canvasID),
	})
}

func canvasInfo(this js.Value, args []js.Value) interface{} {
	canvasID := "default"
	if len(args) > 0 {
		canvasID = args[0].String()
	}
	
	canvas := getCanvas(canvasID)
	currentSystem := canvas.CurrentSystem()
	
	info := map[string]interface{}{
		"canvasId": canvasID,
		"hasActiveSystem": currentSystem != nil,
	}
	
	if currentSystem != nil {
		info["activeSystem"] = currentSystem.System.Name.Value
		// Count components from the system
		// TODO: Get component count from system
		info["components"] = "unknown"
	}
	
	// Add generators info
	generators := canvas.ListGenerators()
	info["generators"] = len(generators)
	
	return jsSuccess(info)
}

func canvasList(this js.Value, args []js.Value) interface{} {
	var canvasIDs []string
	for id := range canvases {
		canvasIDs = append(canvasIDs, id)
	}
	
	return jsSuccess(map[string]interface{}{
		"canvases": canvasIDs,
	})
}

func canvasReset(this js.Value, args []js.Value) interface{} {
	canvasID := "default"
	if len(args) > 0 {
		canvasID = args[0].String()
	}
	
	// Create a new canvas to replace the old one
	canvases[canvasID] = console.NewCanvas(canvasID)
	
	return jsSuccess(map[string]interface{}{
		"canvasId": canvasID,
		"message": fmt.Sprintf("Canvas %s reset", canvasID),
	})
}

func canvasRemove(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("canvas.remove requires canvas ID")
	}
	
	canvasID := args[0].String()
	if canvasID == "default" {
		return jsError("Cannot remove default canvas")
	}
	
	delete(canvases, canvasID)
	
	return jsSuccess(map[string]interface{}{
		"canvasId": canvasID,
		"message": fmt.Sprintf("Canvas %s removed", canvasID),
	})
}

// Generator commands implementation

func genAdd(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return jsError("gen.add requires: name, component.method, rate")
	}
	
	name := args[0].String()
	target := args[1].String()
	rate := args[2].Float()
	
	options := parseOptions(args, 3)
	canvasID := options["canvas"]
	if canvasID == "" {
		canvasID = "default"
	}
	
	canvas := getCanvas(canvasID)
	
	applyFlows := options["applyFlows"] == "true"
	
	// Parse component.method
	component, method := parseComponentMethod(target)
	if component == "" || method == "" {
		return jsError("Invalid target format. Expected: component.method")
	}
	
	// Create GeneratorInfo
	gen := &console.GeneratorInfo{
		Generator: &console.Generator{
			ID:        name,
			Name:      name,
			Component: component,
			Method:    method,
			Rate:      rate,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	err := canvas.AddGenerator(gen)
	if err != nil {
		return jsError(fmt.Sprintf("Failed to add generator: %v", err))
	}
	
	// Apply flows if requested
	if applyFlows {
		// TODO: Implement flow evaluation in WASM
	}
	
	return jsSuccess(map[string]interface{}{
		"name": name,
		"target": target,
		"rate": rate,
		"canvasId": canvasID,
		"applyFlows": applyFlows,
		"message": fmt.Sprintf("Added generator %s -> %s at %v RPS", name, target, rate),
	})
}

// Placeholder implementations for remaining commands
func genRemove(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"message": "Generator removed"})
}

func genUpdate(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"message": "Generator updated"})
}

func genList(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"generators": []interface{}{}})
}

func genStart(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"message": "Generators started"})
}

func genStop(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"message": "Generators stopped"})
}

func metricsAdd(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"message": "Metric added"})
}

func metricsRemove(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"message": "Metric removed"})
}

func metricsUpdate(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"message": "Metric updated"})
}

func metricsList(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"metrics": []interface{}{}})
}

func metricsQuery(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"results": []interface{}{}})
}

func run(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"message": "Simulation complete"})
}

func trace(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"trace": []interface{}{}})
}

func flows(this js.Value, args []js.Value) interface{} {
	return jsSuccess(map[string]interface{}{"flows": map[string]interface{}{}})
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
		"path": path,
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
		"path": path,
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
	
	return jsSuccess(map[string]interface{}{
		"files": files,
		"directory": dir,
	})
}

// Configuration commands
func setDevMode(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("setDevMode requires boolean argument")
	}
	
	devMode := args[0].Bool()
	
	if devMode {
		// Switch to development filesystem with HTTP backend
		fileSystem = loader.CreateDevelopmentFileSystem("http://localhost:8081")
	} else {
		// Switch to production filesystem (bundled/memory)
		fileSystem = createWASMFileSystem()
	}
	
	return jsSuccess(map[string]interface{}{
		"devMode": devMode,
		"message": fmt.Sprintf("Development mode set to %v", devMode),
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
			"prefix": prefix,
			"url": url,
			"message": fmt.Sprintf("Mounted %s to %s", url, prefix),
		})
	}
	
	return jsError("Mounting only supported with composite filesystem")
}

// Helper functions

func jsError(message string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error": message,
	}
}

func jsSuccess(data map[string]interface{}) map[string]interface{} {
	data["success"] = true
	return data
}

func parseOptions(args []js.Value, startIdx int) map[string]string {
	options := make(map[string]string)
	
	// Parse remaining arguments as options object
	if len(args) > startIdx && args[startIdx].Type() == js.TypeObject {
		optObj := args[startIdx]
		
		// Common options
		if canvas := optObj.Get("canvas"); !canvas.IsUndefined() {
			options["canvas"] = canvas.String()
		}
		if applyFlows := optObj.Get("applyFlows"); !applyFlows.IsUndefined() {
			options["applyFlows"] = fmt.Sprintf("%v", applyFlows.Bool())
		}
		if duration := optObj.Get("duration"); !duration.IsUndefined() {
			options["duration"] = duration.String()
		}
	}
	
	return options
}

// Placeholder for embedded files (will be generated at build time)
func getEmbeddedFiles() map[string][]byte {
	return map[string][]byte{
		"/examples/uber.sdl": []byte(`// Uber MVP example
system UberMVP {
    use api APIGateway
    use db Database
}`),
	}
}

// parseComponentMethod splits "component.method" into component and method
func parseComponentMethod(target string) (component, method string) {
	parts := strings.Split(target, ".")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}