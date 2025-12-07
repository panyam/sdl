package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	conc "github.com/panyam/gocurrent"
	gohttp "github.com/panyam/servicekit/http"
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
	clients map[string]*CanvasWSConn
	mu      sync.RWMutex
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

	log.Printf("WebSocket client connected: %s", c.id)

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

	log.Printf("WebSocket client disconnected: %s", c.id)

	// Call the embedded JSONConn's OnClose
	c.JSONConn.OnClose()
}

func (c *CanvasWSConn) OnTimeout() bool {
	log.Printf("WebSocket connection timeout: %s", c.id)
	return true // Close the connection on timeout
}

func (c *CanvasWSConn) HandleMessage(msgData any) error {
	message, ok := msgData.(CanvasWSMessage)
	if !ok {
		return fmt.Errorf("invalid message type")
	}

	log.Printf("Received WebSocket message: %s from %s", message.Type, c.id)

	// Handle different message types
	switch message.Type {
	case "ping":
		c.Writer.Send(conc.Message[any]{Value: CanvasWSMessage{Type: "pong", Data: message.Data}})
	default:
		log.Printf("Unknown WebSocket message type: %s", message.Type)
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
