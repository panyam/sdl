// +build wasm

package main

import (
	"fmt"
	"strings"
	"sync"
	"syscall/js"
)

// WASMHTTPFileSystem implements loader.FileSystem using browser's fetch API
type WASMHTTPFileSystem struct {
	cache sync.Map // path -> []byte
}

func NewWASMHTTPFS() *WASMHTTPFileSystem {
	return &WASMHTTPFileSystem{}
}

func (w *WASMHTTPFileSystem) ReadFile(url string) ([]byte, error) {
	// Check cache first
	if cached, ok := w.cache.Load(url); ok {
		return cached.([]byte), nil
	}
	
	// Ensure it's a full URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("WASM HTTP filesystem requires full URL, got: %s", url)
	}
	
	// Use fetch API through JS
	promise := js.Global().Call("fetch", url)
	
	// Create channels for async result
	result := make(chan js.Value, 1)
	errChan := make(chan error, 1)
	
	// Success handler
	success := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		result <- args[0]
		return nil
	})
	defer success.Release()
	
	// Error handler
	failure := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errChan <- fmt.Errorf("fetch failed: %v", args[0])
		return nil
	})
	defer failure.Release()
	
	// Execute fetch
	promise.Call("then", success).Call("catch", failure)
	
	// Wait for response
	select {
	case response := <-result:
		// Check response status
		if !response.Get("ok").Bool() {
			status := response.Get("status").Int()
			statusText := response.Get("statusText").String()
			return nil, fmt.Errorf("HTTP %d: %s", status, statusText)
		}
		
		// Get response text
		textPromise := response.Call("text")
		textResult := make(chan string, 1)
		
		textSuccess := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			textResult <- args[0].String()
			return nil
		})
		defer textSuccess.Release()
		
		textPromise.Call("then", textSuccess)
		
		// Wait for text
		text := <-textResult
		data := []byte(text)
		
		// Cache the result
		w.cache.Store(url, data)
		
		return data, nil
		
	case err := <-errChan:
		return nil, err
	}
}

func (w *WASMHTTPFileSystem) WriteFile(path string, data []byte) error {
	return fmt.Errorf("HTTP filesystem is read-only")
}

func (w *WASMHTTPFileSystem) ListFiles(dir string) ([]string, error) {
	return nil, fmt.Errorf("directory listing not supported for HTTP filesystem")
}

func (w *WASMHTTPFileSystem) Exists(path string) bool {
	_, err := w.ReadFile(path)
	return err == nil
}

// ClearCache clears the HTTP cache
func (w *WASMHTTPFileSystem) ClearCache() {
	w.cache = sync.Map{}
}