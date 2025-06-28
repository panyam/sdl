// +build wasm

package main

import (
	"fmt"
	"strings"
	"sync"
	"syscall/js"
)

// FileSystem interface for WASM-compatible file operations
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	ListFiles(dir string) ([]string, error)
	Exists(path string) bool
}

// CompositeFS allows multiple file systems to be composed
type CompositeFS struct {
	mu          sync.RWMutex
	filesystems map[string]FileSystem
}

func NewCompositeFS() *CompositeFS {
	return &CompositeFS{
		filesystems: make(map[string]FileSystem),
	}
}

func (c *CompositeFS) Mount(prefix string, fs FileSystem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.filesystems[prefix] = fs
}

func (c *CompositeFS) findFS(path string) (FileSystem, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Check for exact prefix matches first
	for prefix, fs := range c.filesystems {
		if strings.HasPrefix(path, prefix) {
			return fs, path
		}
	}
	
	// Check for protocol handlers (https://, github://)
	if strings.Contains(path, "://") {
		protocol := strings.Split(path, "://")[0] + "://"
		if fs, exists := c.filesystems[protocol]; exists {
			return fs, path
		}
	}
	
	// Default to memory FS
	if fs, exists := c.filesystems["/"]; exists {
		return fs, path
	}
	
	return nil, path
}

func (c *CompositeFS) ReadFile(path string) ([]byte, error) {
	fs, adjustedPath := c.findFS(path)
	if fs == nil {
		return nil, fmt.Errorf("no filesystem mounted for path: %s", path)
	}
	return fs.ReadFile(adjustedPath)
}

func (c *CompositeFS) WriteFile(path string, data []byte) error {
	fs, adjustedPath := c.findFS(path)
	if fs == nil {
		return fmt.Errorf("no filesystem mounted for path: %s", path)
	}
	return fs.WriteFile(adjustedPath, data)
}

func (c *CompositeFS) ListFiles(dir string) ([]string, error) {
	fs, adjustedPath := c.findFS(dir)
	if fs == nil {
		return nil, fmt.Errorf("no filesystem mounted for path: %s", dir)
	}
	return fs.ListFiles(adjustedPath)
}

func (c *CompositeFS) Exists(path string) bool {
	fs, adjustedPath := c.findFS(path)
	if fs == nil {
		return false
	}
	return fs.Exists(adjustedPath)
}

// MemoryFS implements an in-memory file system
type MemoryFS struct {
	mu    sync.RWMutex
	files map[string][]byte
}

func NewMemoryFS() *MemoryFS {
	return &MemoryFS{
		files: make(map[string][]byte),
	}
}

func (m *MemoryFS) ReadFile(path string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	data, exists := m.files[path]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

func (m *MemoryFS) WriteFile(path string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.files[path] = data
	return nil
}

func (m *MemoryFS) ListFiles(dir string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var files []string
	for path := range m.files {
		if strings.HasPrefix(path, dir) {
			files = append(files, path)
		}
	}
	return files, nil
}

func (m *MemoryFS) Exists(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	_, exists := m.files[path]
	return exists
}

// DevServerFS fetches files from a development server
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

// URLFetcherFS fetches files from URLs
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

// Helper function to create default dev filesystem
func NewDevFS() FileSystem {
	cfs := NewCompositeFS()
	
	// Mount development servers
	cfs.Mount("/examples", &DevServerFS{BaseURL: "http://localhost:8081/examples"})
	cfs.Mount("/lib", &DevServerFS{BaseURL: "http://localhost:8081/lib"})
	cfs.Mount("/demos", &DevServerFS{BaseURL: "http://localhost:8081/demos"})
	
	// Mount memory FS for temporary files
	cfs.Mount("/tmp", NewMemoryFS())
	
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