package loader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileSystem interface provides an abstraction for file operations
// that can work in different environments (local disk, HTTP, memory, etc.)
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	ListFiles(dir string) ([]string, error)
	Exists(path string) bool
}

// CompositeFS allows multiple file systems to be composed with different mount points
type CompositeFS struct {
	mu          sync.RWMutex
	filesystems map[string]FileSystem
	fallback    FileSystem
}

func NewCompositeFS() *CompositeFS {
	return &CompositeFS{
		filesystems: make(map[string]FileSystem),
	}
}

func (c *CompositeFS) SetFallback(fs FileSystem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fallback = fs
}

func (c *CompositeFS) Mount(prefix string, fs FileSystem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.filesystems[prefix] = fs
}

func (c *CompositeFS) findFS(path string) (FileSystem, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Check for longest prefix match
	var bestMatch string
	var bestFS FileSystem
	
	for prefix, fs := range c.filesystems {
		if strings.HasPrefix(path, prefix) && len(prefix) > len(bestMatch) {
			bestMatch = prefix
			bestFS = fs
		}
	}
	
	if bestFS != nil {
		return bestFS, path
	}
	
	// Check for protocol handlers (https://, github://)
	if strings.Contains(path, "://") {
		protocol := strings.Split(path, "://")[0] + "://"
		if fs, exists := c.filesystems[protocol]; exists {
			return fs, path
		}
	}
	
	// Use fallback if available
	if c.fallback != nil {
		return c.fallback, path
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

// LocalFS implements FileSystem using the local disk
type LocalFS struct {
	basePath string
}

func NewLocalFS(basePath string) *LocalFS {
	return &LocalFS{basePath: basePath}
}

func (l *LocalFS) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(l.basePath, path)
}

func (l *LocalFS) ReadFile(path string) ([]byte, error) {
	fullPath := l.resolvePath(path)
	return os.ReadFile(fullPath)
}

func (l *LocalFS) WriteFile(path string, data []byte) error {
	fullPath := l.resolvePath(path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

func (l *LocalFS) ListFiles(dir string) ([]string, error) {
	fullPath := l.resolvePath(dir)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

func (l *LocalFS) Exists(path string) bool {
	fullPath := l.resolvePath(path)
	_, err := os.Stat(fullPath)
	return err == nil
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
	return append([]byte(nil), data...), nil // Return a copy
}

func (m *MemoryFS) WriteFile(path string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.files[path] = append([]byte(nil), data...) // Store a copy
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

// PreloadFiles adds files to the memory filesystem
func (m *MemoryFS) PreloadFiles(files map[string][]byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for path, content := range files {
		m.files[path] = append([]byte(nil), content...)
	}
}

// HTTPFileSystem fetches files over HTTP
type HTTPFileSystem struct {
	baseURL string
	client  *http.Client
	cache   sync.Map // path -> []byte
}

func NewHTTPFileSystem(baseURL string) *HTTPFileSystem {
	return &HTTPFileSystem{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{},
	}
}

func (h *HTTPFileSystem) ReadFile(path string) ([]byte, error) {
	// Check cache first
	if cached, ok := h.cache.Load(path); ok {
		return cached.([]byte), nil
	}
	
	// Construct URL
	url := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		url = h.baseURL + "/" + strings.TrimPrefix(path, "/")
	}
	
	// Fetch from server
	resp, err := h.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Cache the result
	h.cache.Store(path, data)
	
	return data, nil
}

func (h *HTTPFileSystem) WriteFile(path string, data []byte) error {
	return fmt.Errorf("HTTP filesystem is read-only")
}

func (h *HTTPFileSystem) ListFiles(dir string) ([]string, error) {
	// Could implement if server provides directory listing
	return nil, fmt.Errorf("directory listing not supported for HTTP filesystem")
}

func (h *HTTPFileSystem) Exists(path string) bool {
	// Would need to make a HEAD request to check
	_, err := h.ReadFile(path)
	return err == nil
}

// ClearCache clears the HTTP cache
func (h *HTTPFileSystem) ClearCache() {
	h.cache = sync.Map{}
}

// GitHubFS provides access to GitHub raw files
type GitHubFS struct {
	httpFS *HTTPFileSystem
}

func NewGitHubFS() *GitHubFS {
	return &GitHubFS{
		httpFS: NewHTTPFileSystem("https://raw.githubusercontent.com"),
	}
}

func (g *GitHubFS) transformPath(path string) string {
	// Transform github.com/user/repo/file.sdl to raw URL
	if strings.HasPrefix(path, "github.com/") {
		parts := strings.SplitN(path[11:], "/", 3)
		if len(parts) >= 3 {
			// Assume main branch if not specified
			return fmt.Sprintf("/%s/%s/main/%s", parts[0], parts[1], parts[2])
		}
	}
	return path
}

func (g *GitHubFS) ReadFile(path string) ([]byte, error) {
	return g.httpFS.ReadFile(g.transformPath(path))
}

func (g *GitHubFS) WriteFile(path string, data []byte) error {
	return fmt.Errorf("GitHub filesystem is read-only")
}

func (g *GitHubFS) ListFiles(dir string) ([]string, error) {
	return nil, fmt.Errorf("directory listing not supported for GitHub filesystem")
}

func (g *GitHubFS) Exists(path string) bool {
	_, err := g.ReadFile(path)
	return err == nil
}