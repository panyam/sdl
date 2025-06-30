// +build wasm

package main

import (
	"syscall/js"
	"fmt"
	"strings"
	"time"
	"github.com/panyam/sdl/console"
	"github.com/panyam/sdl/loader"
	"github.com/panyam/sdl/runtime"
)

// Global canvas manager - reusing existing Canvas type
var canvases map[string]*console.Canvas
var fileSystem loader.FileSystem

func init() {
	canvases = make(map[string]*console.Canvas)
	
	// Initialize filesystem for WASM environment
	fileSystem = createWASMFileSystem()
	
	// Initialize default canvas with WASM filesystem
	fsResolver := loader.NewFileSystemResolver(fileSystem)
	sdlLoader := loader.NewLoader(nil, fsResolver, 10)
	r := runtime.NewRuntime(sdlLoader)
	canvases["default"] = console.NewCanvas("default", r)
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
	
	// Support for external URLs using WASM fetch API
	cfs.Mount("https://", &URLFetcherFS{})
	cfs.Mount("http://", &URLFetcherFS{})
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
			"readFile":   js.FuncOf(fsReadFile),
			"writeFile":  js.FuncOf(fsWriteFile),
			"listFiles":  js.FuncOf(fsListFiles),
			"mount":      js.FuncOf(fsMount), // For dev server mounting
			"isReadOnly": js.FuncOf(fsIsReadOnly),
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
		// Create a loader with our WASM filesystem
		fsResolver := loader.NewFileSystemResolver(fileSystem)
		sdlLoader := loader.NewLoader(nil, fsResolver, 10)
		r := runtime.NewRuntime(sdlLoader)
		
		canvas = console.NewCanvas(id, r)
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
	fsResolver := loader.NewFileSystemResolver(fileSystem)
	sdlLoader := loader.NewLoader(nil, fsResolver, 10)
	r := runtime.NewRuntime(sdlLoader)
	canvases[canvasID] = console.NewCanvas(canvasID, r)
	
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

func genRemove(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("gen.remove requires generator name")
	}
	
	name := args[0].String()
	options := parseOptions(args, 1)
	canvasID := options["canvas"]
	if canvasID == "" {
		canvasID = "default"
	}
	
	canvas := getCanvas(canvasID)
	applyFlows := options["applyFlows"] == "true"
	
	err := canvas.RemoveGenerator(name)
	if err != nil {
		return jsError(fmt.Sprintf("Failed to remove generator: %v", err))
	}
	
	// Apply flows if requested
	if applyFlows {
		// TODO: Implement flow evaluation in WASM
	}
	
	return jsSuccess(map[string]interface{}{
		"name": name,
		"canvasId": canvasID,
		"message": fmt.Sprintf("Removed generator %s", name),
	})
}

func genUpdate(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return jsError("gen.update requires: name, rate")
	}
	
	name := args[0].String()
	rate := args[1].Float()
	
	options := parseOptions(args, 2)
	canvasID := options["canvas"]
	if canvasID == "" {
		canvasID = "default"
	}
	
	canvas := getCanvas(canvasID)
	applyFlows := options["applyFlows"] == "true"
	
	// Find and update the generator
	gen := canvas.GetGenerator(name)
	if gen == nil {
		return jsError(fmt.Sprintf("Generator not found: %s", name))
	}
	
	gen.Generator.Rate = rate
	gen.Generator.UpdatedAt = time.Now()
	
	// Apply flows if requested
	if applyFlows {
		// TODO: Implement flow evaluation in WASM
	}
	
	return jsSuccess(map[string]interface{}{
		"name": name,
		"rate": rate,
		"canvasId": canvasID,
		"applyFlows": applyFlows,
		"message": fmt.Sprintf("Updated generator %s to %v RPS", name, rate),
	})
}

func genList(this js.Value, args []js.Value) interface{} {
	options := parseOptions(args, 0)
	canvasID := options["canvas"]
	if canvasID == "" {
		canvasID = "default"
	}
	
	canvas := getCanvas(canvasID)
	generators := canvas.ListGenerators()
	
	// Convert to JS-friendly format
	var genList []interface{}
	for _, gen := range generators {
		genList = append(genList, map[string]interface{}{
			"id": gen.Generator.ID,
			"name": gen.Generator.Name,
			"component": gen.Generator.Component,
			"method": gen.Generator.Method,
			"rate": gen.Generator.Rate,
			"enabled": gen.Generator.Enabled,
			"createdAt": gen.Generator.CreatedAt.Format(time.RFC3339),
			"updatedAt": gen.Generator.UpdatedAt.Format(time.RFC3339),
		})
	}
	
	return jsSuccess(map[string]interface{}{
		"generators": genList,
		"canvasId": canvasID,
	})
}

func genStart(this js.Value, args []js.Value) interface{} {
	options := parseOptions(args, 0)
	canvasID := options["canvas"]
	if canvasID == "" {
		canvasID = "default"
	}
	
	canvas := getCanvas(canvasID)
	applyFlows := options["applyFlows"] == "true"
	
	// Parse generator names if provided
	var names []string
	if len(args) > 0 && args[0].Type() == js.TypeObject {
		// Names might be in the first argument as an array
		namesValue := args[0]
		if namesValue.Length() > 0 {
			for i := 0; i < namesValue.Length(); i++ {
				names = append(names, namesValue.Index(i).String())
			}
		}
	}
	
	// Start specified generators or all if none specified
	if len(names) == 0 {
		// Start all generators
		_, err := canvas.StartAllGenerators()
		if err != nil {
			return jsError(fmt.Sprintf("Failed to start generators: %v", err))
		}
	} else {
		// Start specific generators
		for _, name := range names {
			err := canvas.StartGenerator(name)
			if err != nil {
				return jsError(fmt.Sprintf("Failed to start generator %s: %v", name, err))
			}
		}
	}
	
	// Apply flows if requested
	if applyFlows {
		// TODO: Implement flow evaluation in WASM
	}
	
	return jsSuccess(map[string]interface{}{
		"names": names,
		"canvasId": canvasID,
		"message": "Generators started",
	})
}

func genStop(this js.Value, args []js.Value) interface{} {
	options := parseOptions(args, 0)
	canvasID := options["canvas"]
	if canvasID == "" {
		canvasID = "default"
	}
	
	canvas := getCanvas(canvasID)
	applyFlows := options["applyFlows"] == "true"
	
	// Parse generator names if provided
	var names []string
	if len(args) > 0 && args[0].Type() == js.TypeObject {
		// Names might be in the first argument as an array
		namesValue := args[0]
		if namesValue.Length() > 0 {
			for i := 0; i < namesValue.Length(); i++ {
				names = append(names, namesValue.Index(i).String())
			}
		}
	}
	
	// Stop specified generators or all if none specified
	if len(names) == 0 {
		// Stop all generators
		_, err := canvas.StopAllGenerators()
		if err != nil {
			return jsError(fmt.Sprintf("Failed to stop generators: %v", err))
		}
	} else {
		// Stop specific generators
		for _, name := range names {
			err := canvas.StopGenerator(name)
			if err != nil {
				return jsError(fmt.Sprintf("Failed to stop generator %s: %v", name, err))
			}
		}
	}
	
	// Apply flows if requested
	if applyFlows {
		// TODO: Implement flow evaluation in WASM
	}
	
	return jsSuccess(map[string]interface{}{
		"names": names,
		"canvasId": canvasID,
		"message": "Generators stopped",
	})
}

func metricsAdd(this js.Value, args []js.Value) interface{} {
	if len(args) < 4 {
		return jsError("metrics.add requires: name, target, type, aggregation")
	}
	
	name := args[0].String()
	target := args[1].String()
	metricType := args[2].String()
	aggregation := args[3].String()
	
	options := parseOptions(args, 4)
	canvasID := options["canvas"]
	if canvasID == "" {
		canvasID = "default"
	}
	
	// canvas := getCanvas(canvasID)
	
	// Parse component.method from target
	component, method := parseComponentMethod(target)
	if component == "" {
		return jsError("Invalid target format. Expected: component.method or component")
	}
	
	// Create MetricSpec
	metric := &console.MetricSpec{
		Metric: &console.Metric{
			ID:          name,
			CanvasID:    canvasID,
			Component:   component,
			Methods:     []string{method},
			MetricType:  metricType,
			Aggregation: aggregation,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	
	// Parse optional window
	if window := options["window"]; window != "" {
		if w, err := time.ParseDuration(window); err == nil {
			metric.Metric.AggregationWindow = float64(w.Seconds())
		}
	}
	
	// TODO: Implement metric management in Canvas
	// For now, metrics are managed through the metric tracer
	// err := canvas.AddMetric(metric)
	// if err != nil {
	// 	return jsError(fmt.Sprintf("Failed to add metric: %v", err))
	// }
	
	return jsSuccess(map[string]interface{}{
		"name": name,
		"target": target,
		"type": metricType,
		"aggregation": aggregation,
		"canvasId": canvasID,
		"message": fmt.Sprintf("Added metric %s", name),
	})
}

func metricsRemove(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("metrics.remove requires metric name")
	}
	
	name := args[0].String()
	options := parseOptions(args, 1)
	canvasID := options["canvas"]
	if canvasID == "" {
		canvasID = "default"
	}
	
	// canvas := getCanvas(canvasID)
	
	// TODO: Implement metric management in Canvas
	// err := canvas.RemoveMetric(name)
	// if err != nil {
	// 	return jsError(fmt.Sprintf("Failed to remove metric: %v", err))
	// }
	
	return jsSuccess(map[string]interface{}{
		"name": name,
		"canvasId": canvasID,
		"message": fmt.Sprintf("Removed metric %s", name),
	})
}

func metricsUpdate(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("metrics.update requires metric name")
	}
	
	// name := args[0].String()
	// options := parseOptions(args, 1)
	// canvasID := options["canvas"]
	// if canvasID == "" {
	// 	canvasID = "default"
	// }
	
	// canvas := getCanvas(canvasID)
	
	// TODO: Implement metric management in Canvas
	// Find the metric
	// metric := canvas.GetMetric(name)
	// if metric == nil {
	// 	return jsError(fmt.Sprintf("Metric not found: %s", name))
	// }
	return jsSuccess(map[string]interface{}{
		"message": "Metric update not yet implemented in WASM",
	})
	
}

func metricsList(this js.Value, args []js.Value) interface{} {
	options := parseOptions(args, 0)
	canvasID := options["canvas"]
	if canvasID == "" {
		canvasID = "default"
	}
	
	// canvas := getCanvas(canvasID)
	// TODO: Implement metric management in Canvas
	// metrics := canvas.ListMetrics()
	var metrics []*console.MetricSpec
	
	// Convert to JS-friendly format
	var metricList []interface{}
	for _, metric := range metrics {
		metricList = append(metricList, map[string]interface{}{
			"id": metric.Metric.ID,
			"component": metric.Metric.Component,
			"methods": metric.Metric.Methods,
			"metricType": metric.Metric.MetricType,
			"aggregation": metric.Metric.Aggregation,
			"aggregationWindow": metric.Metric.AggregationWindow,
			"enabled": metric.Metric.Enabled,
			"createdAt": metric.Metric.CreatedAt.Format(time.RFC3339),
			"updatedAt": metric.Metric.UpdatedAt.Format(time.RFC3339),
		})
	}
	
	return jsSuccess(map[string]interface{}{
		"metrics": metricList,
		"canvasId": canvasID,
	})
}

func metricsQuery(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("metrics.query requires metric name")
	}
	
	metricName := args[0].String()
	options := parseOptions(args, 1)
	canvasID := options["canvas"]
	if canvasID == "" {
		canvasID = "default"
	}
	
	// For now, return empty results as we don't have metric storage in WASM yet
	// TODO: Implement metric storage and querying
	
	return jsSuccess(map[string]interface{}{
		"metric": metricName,
		"canvasId": canvasID,
		"results": []interface{}{},
		"message": "Metric querying not yet implemented in WASM",
	})
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
		"files": jsFiles,
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

func fsIsReadOnly(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("fs.isReadOnly requires file path")
	}
	
	path := args[0].String()
	
	// Check if the path would be readonly based on mount points
	// For composite filesystem, we'll check based on path prefix
	if cfs, ok := fileSystem.(*loader.CompositeFS); ok {
		// Check common readonly prefixes
		readonlyPrefixes := []string{"/examples/", "/lib/", "/demos/", "https://", "http://", "github.com/"}
		for _, prefix := range readonlyPrefixes {
			if strings.HasPrefix(path, prefix) {
				return jsSuccess(map[string]interface{}{
					"path": path,
					"isReadOnly": true,
				})
			}
		}
		// Workspace is writable
		if strings.HasPrefix(path, "/workspace/") {
			return jsSuccess(map[string]interface{}{
				"path": path,
				"isReadOnly": false,
			})
		}
	}
	
	// Fall back to checking the main filesystem
	return jsSuccess(map[string]interface{}{
		"path": path,
		"isReadOnly": fileSystem.IsReadOnly(),
	})
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