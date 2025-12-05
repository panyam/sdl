//go:build wasm
// +build wasm

package main

import (
	"fmt"
	"strings"
	"sync"
	"syscall/js"

	"github.com/panyam/sdl/lib/loader"
)

// DevServerFS fetches files from a development server using browser's fetch API
type DevServerFS struct {
	BaseURL string
	cache   sync.Map // path -> []byte
}

func (d *DevServerFS) ReadFile(path string) ([]byte, error) {
	// Check cache first
	if cached, ok := d.cache.Load(path); ok {
		return cached.([]byte), nil
	}

	// Fetch from server
	url := d.BaseURL + path

	// Use fetch API through JS
	promise := js.Global().Call("fetch", url)

	// Wait for promise to resolve
	result := make(chan js.Value, 1)
	errChan := make(chan error, 1)

	success := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		result <- args[0]
		return nil
	})
	defer success.Release()

	failure := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errChan <- fmt.Errorf("fetch failed: %v", args[0])
		return nil
	})
	defer failure.Release()

	promise.Call("then", success).Call("catch", failure)

	select {
	case response := <-result:
		if !response.Get("ok").Bool() {
			return nil, fmt.Errorf("HTTP %d: %s", response.Get("status").Int(), response.Get("statusText").String())
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

		text := <-textResult
		data := []byte(text)

		// Cache the result
		d.cache.Store(path, data)

		return data, nil

	case err := <-errChan:
		return nil, err
	}
}

func (d *DevServerFS) WriteFile(path string, data []byte) error {
	// Dev server is read-only
	return fmt.Errorf("dev server filesystem is read-only")
}

func (d *DevServerFS) ListFiles(dir string) ([]string, error) {
	// Could implement directory listing if server supports it
	return nil, fmt.Errorf("directory listing not supported for dev server")
}

func (d *DevServerFS) Exists(path string) bool {
	_, err := d.ReadFile(path)
	return err == nil
}

func (d *DevServerFS) IsReadOnly() bool {
	return true // DevServerFS is read-only
}

// BundledFS uses embedded files (will be generated at build time)
type BundledFS struct {
	files map[string][]byte
}

func NewBundledFS() *BundledFS {
	// This will be populated at build time with embedded SDL files
	return &BundledFS{
		files: bundledFiles,
	}
}

func (b *BundledFS) ReadFile(path string) ([]byte, error) {
	data, exists := b.files[path]
	if !exists {
		return nil, fmt.Errorf("file not found in bundle: %s", path)
	}
	return data, nil
}

func (b *BundledFS) WriteFile(path string, data []byte) error {
	return fmt.Errorf("bundled filesystem is read-only")
}

func (b *BundledFS) ListFiles(dir string) ([]string, error) {
	var files []string
	for path := range b.files {
		if strings.HasPrefix(path, dir) {
			files = append(files, path)
		}
	}
	return files, nil
}

func (b *BundledFS) Exists(path string) bool {
	_, exists := b.files[path]
	return exists
}

func (b *BundledFS) IsReadOnly() bool {
	return true // BundledFS is read-only
}

// URLFetcherFS fetches files from URLs using browser's fetch API
type URLFetcherFS struct{}

func (u *URLFetcherFS) ReadFile(url string) ([]byte, error) {
	// Similar to DevServerFS but for arbitrary URLs
	promise := js.Global().Call("fetch", url)

	result := make(chan js.Value, 1)
	errChan := make(chan error, 1)

	success := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		result <- args[0]
		return nil
	})
	defer success.Release()

	failure := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errChan <- fmt.Errorf("fetch failed: %v", args[0])
		return nil
	})
	defer failure.Release()

	promise.Call("then", success).Call("catch", failure)

	select {
	case response := <-result:
		if !response.Get("ok").Bool() {
			return nil, fmt.Errorf("HTTP %d", response.Get("status").Int())
		}

		textPromise := response.Call("text")
		textResult := make(chan string, 1)

		textSuccess := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			textResult <- args[0].String()
			return nil
		})
		defer textSuccess.Release()

		textPromise.Call("then", textSuccess)

		return []byte(<-textResult), nil

	case err := <-errChan:
		return nil, err
	}
}

func (u *URLFetcherFS) WriteFile(path string, data []byte) error {
	return fmt.Errorf("URL filesystem is read-only")
}

func (u *URLFetcherFS) ListFiles(dir string) ([]string, error) {
	return nil, fmt.Errorf("directory listing not supported for URLs")
}

func (u *URLFetcherFS) Exists(path string) bool {
	// Can't check existence without fetching
	return false
}

func (u *URLFetcherFS) IsReadOnly() bool {
	return true // URLFetcherFS is read-only
}

// Helper function to create default dev filesystem
func NewDevFS() loader.FileSystem {
	cfs := loader.NewCompositeFS()

	// Mount development servers
	cfs.Mount("/examples", &DevServerFS{BaseURL: "http://localhost:8081/examples"})
	cfs.Mount("/lib", &DevServerFS{BaseURL: "http://localhost:8081/lib"})
	cfs.Mount("/demos", &DevServerFS{BaseURL: "http://localhost:8081/demos"})

	// Mount memory FS for temporary files
	cfs.Mount("/tmp", loader.NewMemoryFS())

	// Mount URL fetchers
	cfs.Mount("https://", &URLFetcherFS{})
	cfs.Mount("http://", &URLFetcherFS{})

	return cfs
}

// Placeholder for bundled files (will be generated at build time)
var bundledFiles = map[string][]byte{
	"/examples/uber.sdl": []byte(`// Uber MVP example
system UberMVP {
    use api APIGateway
    use db Database
}`),
}
