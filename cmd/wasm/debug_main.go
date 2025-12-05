// +build wasm

package main

import (
	"syscall/js"
	"fmt"
)

func init() {
	// Add a debug function to test slice marshaling
	js.Global().Set("debugTest", js.FuncOf(debugTest))
}

func debugTest(this js.Value, args []js.Value) interface{} {
	fmt.Println("Debug test called")
	
	// Test 1: Return with populated slice
	test1 := map[string]interface{}{
		"test": "populated",
		"files": []string{"file1.txt", "file2.txt"},
		"success": true,
	}
	
	// Test 2: Return with empty slice
	test2 := map[string]interface{}{
		"test": "empty",
		"files": []string{},
		"success": true,
	}
	
	// Test 3: Return with nil slice
	var nilSlice []string
	test3 := map[string]interface{}{
		"test": "nil",
		"files": nilSlice,
		"success": true,
	}
	
	// Return based on argument
	if len(args) > 0 {
		switch args[0].String() {
		case "populated":
			return test1
		case "empty":
			return test2
		case "nil":
			return test3
		default:
			return map[string]interface{}{
				"error": "unknown test type",
				"success": false,
			}
		}
	}
	
	return test1
}