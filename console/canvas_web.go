package console

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/panyam/goutils/conc"
	gohttp "github.com/panyam/goutils/http"
	"github.com/panyam/sdl/loader"
	"github.com/panyam/templar"
)

// CanvasWSMessage represents a WebSocket message for the Canvas API
type CanvasWSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// CanvasWSConn implements the goutils WebSocket connection interface
type CanvasWSConn struct {
	gohttp.JSONConn
	handler *CanvasWSHandler
	id      string
}

// CanvasWSHandler manages WebSocket connections for the Canvas API
type CanvasWSHandler struct {
	webServer *WebServer
	clients   map[string]*CanvasWSConn
	mu        sync.RWMutex
}

// WebServer wraps the Canvas API for HTTP access
type WebServer struct {
	api            *SDLApi
	wsHandler      *CanvasWSHandler
	canvasService  *CanvasService
	systemsHandler *SystemsHandler
	templateGroup  *templar.TemplateGroup
}

// NewWebServer creates a new web server instance
func NewWebServer(grpcAddress string, canvasService *CanvasService) *WebServer {
	ws := &WebServer{
		api:           NewSDLApi(grpcAddress, canvasService),
		canvasService: canvasService,
	}

	// Initialize WebSocket handler
	ws.wsHandler = &CanvasWSHandler{
		webServer: ws,
		clients:   make(map[string]*CanvasWSConn),
	}

	// Initialize template engine
	templateGroup, err := SetupTemplates("console/templates")
	if err != nil {
		Warn("Failed to setup templates: %v", err)
	}
	ws.templateGroup = templateGroup

	// Initialize systems handler
	ws.systemsHandler = NewSystemsHandler(templateGroup)

	return ws
}

// WebSocket connection lifecycle methods
func (c *CanvasWSConn) OnStart(conn *websocket.Conn) error {
	// First, initialize the embedded JSONConn
	if err := c.JSONConn.OnStart(conn); err != nil {
		return err
	}

	c.id = fmt.Sprintf("conn_%s", conn.RemoteAddr().String())

	c.handler.mu.Lock()
	c.handler.clients[c.id] = c
	c.handler.mu.Unlock()

	Info("WebSocket client connected: %s", c.id)

	// Send initial connection message
	message := CanvasWSMessage{
		Type: "connected",
		Data: map[string]interface{}{
			"status": "connected",
			"server": "SDL Canvas API",
			"id":     c.id,
		},
	}

	// Now the Writer should be properly initialized
	c.Writer.Send(conc.Message[any]{Value: message})
	return nil
}

func (c *CanvasWSConn) OnClose() {
	// Clean up our client tracking
	c.handler.mu.Lock()
	delete(c.handler.clients, c.id)
	c.handler.mu.Unlock()

	Info("WebSocket client disconnected: %s", c.id)

	// Call the embedded JSONConn's OnClose
	c.JSONConn.OnClose()
}

func (c *CanvasWSConn) OnTimeout() bool {
	Warn("⏰ WebSocket connection timeout: %s", c.id)
	return true // Close the connection on timeout
}

func (c *CanvasWSConn) HandleMessage(msgData any) error {
	message, ok := msgData.(CanvasWSMessage)
	if !ok {
		return fmt.Errorf("invalid message type")
	}

	Debug("Received WebSocket message: %s from %s", message.Type, c.id)

	// Handle different message types
	switch message.Type {
	case "ping":
		c.Writer.Send(conc.Message[any]{Value: CanvasWSMessage{Type: "pong", Data: message.Data}})
	default:
		Warn("Unknown WebSocket message type: %s", message.Type)
	}

	return nil
}

// WebSocket handler interface implementation
func (h *CanvasWSHandler) Validate(w http.ResponseWriter, r *http.Request) (out *CanvasWSConn, isValid bool) {
	// Allow all connections for now - can add authentication here
	conn := &CanvasWSConn{
		handler: h,
	}
	return conn, true
}

// GetRouter returns a configured HTTP router with all Canvas API routes
func (ws *WebServer) Handler() http.Handler {
	r := http.NewServeMux()
	
	// API routes
	r.Handle("/api/", http.StripPrefix("/api", ws.api.Handler()))
	
	// Filesystem API endpoints
	r.HandleFunc("/api/filesystems", ws.handleFilesystemsAPI)
	r.HandleFunc("/api/filesystems/", ws.handleFilesystemOperations)
	
	// Serve examples directory for WASM demos
	r.Handle("/examples/", http.StripPrefix("/examples", http.FileServer(http.Dir("./examples/"))))
	
	// System showcase routes (server-rendered)
	if ws.systemsHandler != nil {
		r.Handle("/systems", ws.systemsHandler.Handler())
		r.Handle("/systems/", ws.systemsHandler.Handler())
		r.Handle("/system/", ws.systemsHandler.Handler())
	}
	
	// Canvas-specific routes
	r.HandleFunc("/canvases/", ws.handleCanvasRoute)
	
	// Root redirect to systems listing
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			http.Redirect(w, req, "/systems", http.StatusFound)
			return
		}
		// Serve static files for other root-level paths
		http.FileServer(http.Dir("./web/dist/")).ServeHTTP(w, req)
	})

	return r
}

// handleCanvasRoute serves the dashboard for a specific canvas
func (ws *WebServer) handleCanvasRoute(w http.ResponseWriter, r *http.Request) {
	// Extract canvas ID from URL path
	path := r.URL.Path
	prefix := "/canvases/"
	
	if !strings.HasPrefix(path, prefix) {
		http.NotFound(w, r)
		return
	}
	
	// Remove prefix and find canvas ID
	remaining := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remaining, "/")
	
	if len(parts) == 0 || parts[0] == "" {
		// Redirect /canvases/ to /canvases/default/
		http.Redirect(w, r, "/canvases/default/", http.StatusFound)
		return
	}
	
	// canvasID := parts[0] // Available if we need to use it later
	
	// For any path under /canvases/{canvasId}/, serve the index.html
	// This allows the frontend router to handle sub-paths
	if strings.HasSuffix(path, "/") || len(parts) == 1 {
		// Serve index.html for canvas root
		http.ServeFile(w, r, "./web/dist/index.html")
	} else if len(parts) > 1 {
		// For assets and other files, strip the canvas prefix and serve from dist
		// e.g., /canvases/mycanvas/assets/index.js -> /assets/index.js
		assetPath := "/" + strings.Join(parts[1:], "/")
		r.URL.Path = assetPath
		http.FileServer(http.Dir("./web/dist/")).ServeHTTP(w, r)
	}
}

// FileSystemInfo represents information about a mounted filesystem
type FileSystemInfo struct {
	Prefix     string `json:"prefix"`
	Type       string `json:"type"`
	ReadOnly   bool   `json:"readOnly"`
	BasePath   string `json:"basePath,omitempty"`
}

// handleFilesystemsAPI handles requests to the /api/filesystems endpoint
func (ws *WebServer) handleFilesystemsAPI(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers for browser access
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the default canvas
	canvas := ws.getDefaultCanvas()
	if canvas == nil {
		http.Error(w, "Default canvas not found", http.StatusInternalServerError)
		return
	}

	// Get the loader from the runtime
	if canvas.runtime == nil || canvas.runtime.Loader == nil {
		http.Error(w, "Canvas runtime not initialized", http.StatusInternalServerError)
		return
	}

	// Get filesystem information
	filesystems := ws.getFileSystemInfo(canvas.runtime.Loader)

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filesystems); err != nil {
		Error("Failed to encode filesystems response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getDefaultCanvas retrieves the default canvas
func (ws *WebServer) getDefaultCanvas() *Canvas {
	if ws.canvasService == nil {
		return nil
	}

	ws.canvasService.storeMutex.RLock()
	defer ws.canvasService.storeMutex.RUnlock()

	return ws.canvasService.store["default"]
}

// getFileSystemInfo extracts filesystem information from the loader
func (ws *WebServer) getFileSystemInfo(l *loader.Loader) []FileSystemInfo {
	var filesystems []FileSystemInfo

	// The loader has a resolver field which is private, so we can't access it directly
	// For now, we'll return the default filesystem configuration
	// In a production scenario, you'd want to expose this information through the loader

	// Return the standard filesystem configuration
	filesystems = append(filesystems, FileSystemInfo{
		Prefix:   "/",
		Type:     "local",
		ReadOnly: false,
		BasePath: ".",
	})

	// Examples directory is typically mounted
	filesystems = append(filesystems, FileSystemInfo{
		Prefix:   "/examples/",
		Type:     "local",
		ReadOnly: true,
		BasePath: "./examples",
	})

	// GitHub filesystem for remote imports
	filesystems = append(filesystems, FileSystemInfo{
		Prefix:   "github.com/",
		Type:     "github",
		ReadOnly: true,
	})

	// HTTP/HTTPS support
	filesystems = append(filesystems, FileSystemInfo{
		Prefix:   "https://",
		Type:     "http",
		ReadOnly: true,
	})

	return filesystems
}

// extractCompositeFileSystemInfo extracts info from a CompositeFS
func (ws *WebServer) extractCompositeFileSystemInfo(composite *loader.CompositeFS) []FileSystemInfo {
	var filesystems []FileSystemInfo

	// Note: CompositeFS doesn't expose its internal filesystems map,
	// so we'll return a generic representation
	// In a real implementation, you might want to add a method to CompositeFS
	// to expose this information

	// For now, return common filesystem types that might be mounted
	filesystems = append(filesystems, FileSystemInfo{
		Prefix:   "/",
		Type:     "local",
		ReadOnly: false,
		BasePath: ".",
	})

	// Check for examples mount
	filesystems = append(filesystems, FileSystemInfo{
		Prefix:   "/examples/",
		Type:     "local",
		ReadOnly: true,
		BasePath: "./examples",
	})

	return filesystems
}

// getFileSystemType returns a string representation of the filesystem type
func (ws *WebServer) getFileSystemType(fs loader.FileSystem) string {
	switch fs.(type) {
	case *loader.LocalFS:
		return "local"
	case *loader.MemoryFS:
		return "memory"
	case *loader.HTTPFileSystem:
		return "http"
	case *loader.GitHubFS:
		return "github"
	case *loader.CompositeFS:
		return "composite"
	default:
		return "unknown"
	}
}

// FileInfo represents file information for API responses
type FileInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"isDirectory"`
	Size        int64  `json:"size,omitempty"`
	ModTime     string `json:"modTime,omitempty"`
}

// ListFilesResponse represents the response for file listing
type ListFilesResponse struct {
	Files []FileInfo `json:"files"`
}

// FileSystemConfig represents configuration for a filesystem
type FileSystemConfig struct {
	ID         string   `json:"id"`
	BasePath   string   `json:"basePath"`
	ReadOnly   bool     `json:"readOnly"`
	Extensions []string `json:"extensions"` // Allowed file extensions
}

// Default filesystem configurations
var defaultFileSystems = map[string]FileSystemConfig{
	"examples": {
		ID:         "examples",
		BasePath:   "./examples",
		ReadOnly:   false,
		Extensions: []string{".sdl", ".recipe"},
	},
	"demos": {
		ID:         "demos",
		BasePath:   "./demos",
		ReadOnly:   true,
		Extensions: []string{".sdl", ".recipe"},
	},
}

// handleFilesystemOperations handles file operations for a specific filesystem
func (ws *WebServer) handleFilesystemOperations(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract filesystem ID and path from URL
	// URL format: /api/filesystems/{fsId}/{path...}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/filesystems/"), "/")
	if len(pathParts) < 1 {
		http.Error(w, "Invalid filesystem path", http.StatusBadRequest)
		return
	}

	fsID := pathParts[0]
	filePath := "/"
	if len(pathParts) > 1 {
		filePath = "/" + strings.Join(pathParts[1:], "/")
	}

	// Get filesystem configuration
	fsConfig, exists := defaultFileSystems[fsID]
	if !exists {
		http.Error(w, "Filesystem not found", http.StatusNotFound)
		return
	}

	// Resolve the full path and ensure it's within the base path
	fullPath := filepath.Join(fsConfig.BasePath, filePath)
	fullPath = filepath.Clean(fullPath)

	// Security check: ensure the path is within the base path
	absBase, _ := filepath.Abs(fsConfig.BasePath)
	absFull, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFull, absBase) {
		http.Error(w, "Access denied: path outside filesystem root", http.StatusForbidden)
		return
	}

	// Handle different HTTP methods
	switch r.Method {
	case http.MethodGet:
		ws.handleFileGet(w, r, fullPath, filePath, fsConfig)
	case http.MethodPut:
		if fsConfig.ReadOnly {
			http.Error(w, "Filesystem is read-only", http.StatusForbidden)
			return
		}
		ws.handleFilePut(w, r, fullPath, filePath, fsConfig)
	case http.MethodDelete:
		if fsConfig.ReadOnly {
			http.Error(w, "Filesystem is read-only", http.StatusForbidden)
			return
		}
		ws.handleFileDelete(w, r, fullPath, filePath, fsConfig)
	case http.MethodPost:
		if fsConfig.ReadOnly {
			http.Error(w, "Filesystem is read-only", http.StatusForbidden)
			return
		}
		ws.handleFilePost(w, r, fullPath, filePath, fsConfig)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleFileGet handles GET requests for files or directories
func (ws *WebServer) handleFileGet(w http.ResponseWriter, r *http.Request, fullPath, requestPath string, fsConfig FileSystemConfig) {
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error accessing file", http.StatusInternalServerError)
		}
		return
	}

	if info.IsDir() {
		// List directory contents
		files, err := ws.listDirectory(fullPath, requestPath, fsConfig)
		if err != nil {
			http.Error(w, "Error listing directory", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListFilesResponse{Files: files})
	} else {
		// Check file extension
		if !ws.isAllowedFile(info.Name(), fsConfig) {
			http.Error(w, "File type not allowed", http.StatusForbidden)
			return
		}

		// Serve file content
		http.ServeFile(w, r, fullPath)
	}
}

// handleFilePut handles PUT requests to write files
func (ws *WebServer) handleFilePut(w http.ResponseWriter, r *http.Request, fullPath, requestPath string, fsConfig FileSystemConfig) {
	// Check file extension
	if !ws.isAllowedFile(filepath.Base(fullPath), fsConfig) {
		http.Error(w, "File type not allowed", http.StatusForbidden)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Ensure parent directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	// Write file
	if err := os.WriteFile(fullPath, body, 0644); err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File saved successfully"))
}

// handleFileDelete handles DELETE requests to remove files
func (ws *WebServer) handleFileDelete(w http.ResponseWriter, r *http.Request, fullPath, requestPath string, fsConfig FileSystemConfig) {
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error accessing file", http.StatusInternalServerError)
		}
		return
	}

	// Only allow deletion of files, not directories (for safety)
	if info.IsDir() {
		http.Error(w, "Cannot delete directories", http.StatusForbidden)
		return
	}

	// Check file extension
	if !ws.isAllowedFile(info.Name(), fsConfig) {
		http.Error(w, "File type not allowed", http.StatusForbidden)
		return
	}

	if err := os.Remove(fullPath); err != nil {
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File deleted successfully"))
}

// handleFilePost handles POST requests to create directories
func (ws *WebServer) handleFilePost(w http.ResponseWriter, r *http.Request, fullPath, requestPath string, fsConfig FileSystemConfig) {
	var body struct {
		Type string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Type != "directory" {
		http.Error(w, "Only directory creation is supported", http.StatusBadRequest)
		return
	}

	// Create directory
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Directory created successfully"))
}

// listDirectory lists the contents of a directory with filtering
func (ws *WebServer) listDirectory(dirPath, requestPath string, fsConfig FileSystemConfig) ([]FileInfo, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// For directories, always include them
		if entry.IsDir() {
			files = append(files, FileInfo{
				Name:        entry.Name(),
				Path:        path.Join(requestPath, entry.Name()) + "/",
				IsDirectory: true,
				ModTime:     info.ModTime().Format("2006-01-02T15:04:05Z"),
			})
		} else if ws.isAllowedFile(entry.Name(), fsConfig) {
			// For files, only include if they match allowed extensions
			files = append(files, FileInfo{
				Name:        entry.Name(),
				Path:        path.Join(requestPath, entry.Name()),
				IsDirectory: false,
				Size:        info.Size(),
				ModTime:     info.ModTime().Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	return files, nil
}

// isAllowedFile checks if a file matches the allowed extensions
func (ws *WebServer) isAllowedFile(filename string, fsConfig FileSystemConfig) bool {
	// If no extensions specified, allow all files
	if len(fsConfig.Extensions) == 0 {
		return true
	}

	ext := filepath.Ext(filename)
	for _, allowed := range fsConfig.Extensions {
		if ext == allowed {
			return true
		}
	}

	return false
}

/*
type PlotRequest struct {
	Series     []SeriesInfo `json:"series"`
	OutputFile string       `json:"outputFile"`
	Title      string       `json:"title"`
}
*/
