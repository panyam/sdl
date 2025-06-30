package console

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/panyam/goutils/conc"
	gohttp "github.com/panyam/goutils/http"
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
	api       *SDLApi
	wsHandler *CanvasWSHandler
}

// NewWebServer creates a new web server instance
func NewWebServer(grpcAddress string, canvasService *CanvasService) *WebServer {
	ws := &WebServer{
		api: NewSDLApi(grpcAddress, canvasService),
	}

	// Initialize WebSocket handler
	ws.wsHandler = &CanvasWSHandler{
		webServer: ws,
		clients:   make(map[string]*CanvasWSConn),
	}

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
	
	// Serve examples directory for WASM demos
	r.Handle("/examples/", http.StripPrefix("/examples", http.FileServer(http.Dir("./examples/"))))
	
	// Canvas-specific routes
	r.HandleFunc("/canvases/", ws.handleCanvasRoute)
	
	// Root redirect to default canvas
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			http.Redirect(w, req, "/canvases/default/", http.StatusFound)
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

/*
type PlotRequest struct {
	Series     []SeriesInfo `json:"series"`
	OutputFile string       `json:"outputFile"`
	Title      string       `json:"title"`
}
*/
