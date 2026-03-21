package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	gohttp "github.com/panyam/servicekit/http"
)

// CanvasWSMessage represents a WebSocket message for the Canvas API
type CanvasWSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// CanvasWSConn implements the servicekit WebSocket connection interface
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

func (c *CanvasWSConn) HandleMessage(msgData any) error {
	message, ok := msgData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid message type")
	}

	msgType, _ := message["type"].(string)
	log.Printf("Received WebSocket message: %s from %s", msgType, c.ConnId())

	switch msgType {
	case "ping":
		c.SendOutput(CanvasWSMessage{Type: "pong", Data: message["data"]})
	default:
		log.Printf("Unknown WebSocket message type: %s", msgType)
	}

	return nil
}

// Validate implements WSHandler - creates a new connection for each upgrade
func (h *CanvasWSHandler) Validate(w http.ResponseWriter, r *http.Request) (out *CanvasWSConn, isValid bool) {
	conn := &CanvasWSConn{
		JSONConn: gohttp.JSONConn{
			Codec:   &gohttp.JSONCodec{},
			NameStr: "CanvasWS",
		},
		handler: h,
	}

	h.mu.Lock()
	h.clients[conn.ConnId()] = conn
	h.mu.Unlock()

	return conn, true
}
