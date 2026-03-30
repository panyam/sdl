package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/panyam/sdl/tools/systemdetail"
)

var (
	tool *systemdetail.SystemDetailTool
)

func main() {
	// Create a new SystemDetailTool instance
	tool = systemdetail.NewSystemDetailTool()
	
	// Expose SystemDetailTool methods to JavaScript
	js.Global().Set("newSystemDetailTool", js.FuncOf(newSystemDetailTool))
	js.Global().Set("setCallbacks", js.FuncOf(setCallbacks))
	js.Global().Set("initialize", js.FuncOf(initialize))
	js.Global().Set("setSDLContent", js.FuncOf(setSDLContent))
	js.Global().Set("setRecipeContent", js.FuncOf(setRecipeContent))
	js.Global().Set("getSystemInfo", js.FuncOf(getSystemInfo))
	js.Global().Set("getCompileResult", js.FuncOf(getCompileResult))
	js.Global().Set("getExecState", js.FuncOf(getExecState))
	js.Global().Set("useSystem", js.FuncOf(useSystem))
	js.Global().Set("getSystemID", js.FuncOf(getSystemID))
	js.Global().Set("getSDLContent", js.FuncOf(getSDLContent))
	js.Global().Set("getRecipeContent", js.FuncOf(getRecipeContent))
	
	// Keep the program running
	select {}
}

// newSystemDetailTool creates a new SystemDetailTool instance
func newSystemDetailTool(this js.Value, args []js.Value) interface{} {
	tool = systemdetail.NewSystemDetailTool()
	return js.ValueOf(true)
}

// setCallbacks sets up JavaScript callbacks for SystemDetailTool
func setCallbacks(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   "callbacks object required",
		})
	}
	
	callbacks := args[0]
	
	// Extract callback functions
	var onError, onInfo, onSuccess js.Value
	if callbacks.Get("onError").Type() == js.TypeFunction {
		onError = callbacks.Get("onError")
	}
	if callbacks.Get("onInfo").Type() == js.TypeFunction {
		onInfo = callbacks.Get("onInfo")
	}
	if callbacks.Get("onSuccess").Type() == js.TypeFunction {
		onSuccess = callbacks.Get("onSuccess")
	}
	
	// Set up callbacks
	tool.SetCallbacks(&systemdetail.Callbacks{
		OnError: func(msg string) {
			if onError.Type() == js.TypeFunction {
				onError.Invoke(js.ValueOf(msg))
			}
		},
		OnInfo: func(msg string) {
			if onInfo.Type() == js.TypeFunction {
				onInfo.Invoke(js.ValueOf(msg))
			}
		},
		OnSuccess: func(msg string) {
			if onSuccess.Type() == js.TypeFunction {
				onSuccess.Invoke(js.ValueOf(msg))
			}
		},
	})
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
	})
}

// initialize initializes the SystemDetailTool with system data
func initialize(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   "systemId, sdlContent, and recipeContent required",
		})
	}
	
	systemId := args[0].String()
	sdlContent := args[1].String()
	recipeContent := args[2].String()
	
	err := tool.Initialize(systemId, sdlContent, recipeContent)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
	})
}

// setSDLContent sets the SDL content and triggers compilation
func setSDLContent(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   "SDL content required",
		})
	}
	
	content := args[0].String()
	err := tool.SetSDLContent(content)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
	})
}

// setRecipeContent sets the recipe content and parses it
func setRecipeContent(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   "Recipe content required",
		})
	}
	
	content := args[0].String()
	err := tool.SetRecipeContent(content)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
	})
}

// getSystemInfo returns system information as JSON
func getSystemInfo(this js.Value, args []js.Value) interface{} {
	info := tool.GetSystemInfo()
	
	// Convert to JSON and back to ensure proper serialization
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// getCompileResult returns the SDL compilation result
func getCompileResult(this js.Value, args []js.Value) interface{} {
	result := tool.GetCompileResult()
	if result == nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   "No compilation result available",
		})
	}
	
	// Convert to JSON and back to ensure proper serialization
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	var resultMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &resultMap)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
		"data":    resultMap,
	})
}

// getExecState returns the recipe execution state
func getExecState(this js.Value, args []js.Value) interface{} {
	state := tool.GetExecState()
	if state == nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   "No execution state available",
		})
	}
	
	// Convert to JSON and back to ensure proper serialization
	jsonBytes, err := json.Marshal(state)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	var resultMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &resultMap)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
		"data":    resultMap,
	})
}

// useSystem activates a specific system
func useSystem(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   "System name required",
		})
	}
	
	systemName := args[0].String()
	err := tool.UseSystem(systemName)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
	})
}

// getSystemID returns the current system ID
func getSystemID(this js.Value, args []js.Value) interface{} {
	systemID := tool.GetSystemID()
	return js.ValueOf(map[string]interface{}{
		"success": true,
		"data":    systemID,
	})
}

// getSDLContent returns the current SDL content
func getSDLContent(this js.Value, args []js.Value) interface{} {
	content := tool.GetSDLContent()
	return js.ValueOf(map[string]interface{}{
		"success": true,
		"data":    content,
	})
}

// getRecipeContent returns the current recipe content
func getRecipeContent(this js.Value, args []js.Value) interface{} {
	content := tool.GetRecipeContent()
	return js.ValueOf(map[string]interface{}{
		"success": true,
		"data":    content,
	})
}