package console

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	gohttp "github.com/panyam/goutils/http"
	"github.com/panyam/goutils/conc"
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
	canvas    *Canvas
	wsHandler *CanvasWSHandler
}

// NewWebServer creates a new web server instance
func NewWebServer() *WebServer {
	ws := &WebServer{
		canvas: NewCanvas(),
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
	
	log.Printf("🔌 WebSocket client connected: %s", c.id)
	
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
	
	log.Printf("🔌 WebSocket client disconnected: %s", c.id)
	
	// Call the embedded JSONConn's OnClose
	c.JSONConn.OnClose()
}

func (c *CanvasWSConn) OnTimeout() bool {
	log.Printf("⏰ WebSocket connection timeout: %s", c.id)
	return true // Close the connection on timeout
}

func (c *CanvasWSConn) HandleMessage(msgData any) error {
	message, ok := msgData.(CanvasWSMessage)
	if !ok {
		return fmt.Errorf("invalid message type")
	}
	
	log.Printf("📡 Received WebSocket message: %s from %s", message.Type, c.id)
	
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

// GetRouter returns a configured HTTP router with all Canvas API routes
func (ws *WebServer) GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(corsMiddleware)

	// Legacy API routes (for backward compatibility)
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/load", ws.handleLoad).Methods("POST")
	api.HandleFunc("/use", ws.handleUse).Methods("POST")
	api.HandleFunc("/set", ws.handleSet).Methods("POST")
	api.HandleFunc("/run", ws.handleRun).Methods("POST")
	api.HandleFunc("/plot", ws.handlePlot).Methods("POST")
	
	// WebSocket endpoint using goutils
	api.HandleFunc("/live", gohttp.WSServe(ws.wsHandler, nil))

	// RESTful Canvas API routes
	ws.setupCanvasAPIRoutes(r)

	// Static file serving (will serve the built frontend)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist/")))

	return r
}

// API Request/Response types
type LoadRequest struct {
	FilePath string `json:"filePath"`
}

type UseRequest struct {
	SystemName string `json:"systemName"`
}

type SetRequest struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type RunRequest struct {
	VarName string `json:"varName"`
	Target  string `json:"target"`
	Runs    int    `json:"runs"`
}

type PlotRequest struct {
	Series     []SeriesInfo `json:"series"`
	OutputFile string       `json:"outputFile"`
	Title      string       `json:"title"`
}


type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Legacy HTTP Handlers (for backward compatibility)
func (ws *WebServer) handleLoad(w http.ResponseWriter, r *http.Request) {
	var req LoadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := ws.canvas.Load(req.FilePath)
	if err != nil {
		ws.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "loaded", "file": req.FilePath})
	ws.broadcast("fileLoaded", map[string]interface{}{
		"file": req.FilePath,
	})
}

func (ws *WebServer) handleUse(w http.ResponseWriter, r *http.Request) {
	var req UseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := ws.canvas.Use(req.SystemName)
	if err != nil {
		ws.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "system activated", "system": req.SystemName})
	ws.broadcast("systemActivated", map[string]interface{}{
		"system": req.SystemName,
	})
}

func (ws *WebServer) handleSet(w http.ResponseWriter, r *http.Request) {
	var req SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := ws.canvas.Set(req.Path, req.Value)
	if err != nil {
		ws.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.sendSuccess(w, map[string]interface{}{
		"status": "parameter set",
		"path":   req.Path,
		"value":  req.Value,
	})
	ws.broadcast("parameterChanged", map[string]interface{}{
		"path":  req.Path,
		"value": req.Value,
	})
}

func (ws *WebServer) handleRun(w http.ResponseWriter, r *http.Request) {
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	runs := req.Runs
	if runs <= 0 {
		runs = 1000 // Default
	}

	err := ws.canvas.Run(req.VarName, req.Target, WithRuns(runs))
	if err != nil {
		ws.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.sendSuccess(w, map[string]interface{}{
		"status":  "simulation completed",
		"varName": req.VarName,
		"target":  req.Target,
		"runs":    runs,
	})
	ws.broadcast("simulationCompleted", map[string]interface{}{
		"varName": req.VarName,
		"target":  req.Target,
		"runs":    runs,
	})
}

func (ws *WebServer) handlePlot(w http.ResponseWriter, r *http.Request) {
	var req PlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build plot options
	opts := []PlotOption{
		WithOutput(req.OutputFile),
	}
	for _, series := range req.Series {
		opts = append(opts, WithSeries(series.Name, series.From))
	}

	err := ws.canvas.Plot(opts...)
	if err != nil {
		ws.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.sendSuccess(w, map[string]interface{}{
		"status":     "plot generated",
		"outputFile": req.OutputFile,
		"series":     req.Series,
	})
	ws.broadcast("plotGenerated", map[string]interface{}{
		"outputFile": req.OutputFile,
		"series":     req.Series,
	})
}


// RESTful Canvas API Routes
func (ws *WebServer) setupCanvasAPIRoutes(router *mux.Router) {
	// Canvas state management
	router.HandleFunc("/api/canvas/state", ws.handleGetState).Methods("GET")
	router.HandleFunc("/api/canvas/state", ws.handleSaveState).Methods("POST")
	router.HandleFunc("/api/canvas/state/restore", ws.handleRestoreState).Methods("POST")
	
	// System diagram
	router.HandleFunc("/api/canvas/diagram", ws.handleGetDiagram).Methods("GET")

	// Generator management (RESTful)
	router.HandleFunc("/api/canvas/generators", ws.handleGetGenerators).Methods("GET")
	router.HandleFunc("/api/canvas/generators", ws.handleAddGenerator).Methods("POST")
	router.HandleFunc("/api/canvas/generators/{id}", ws.handleGetGenerator).Methods("GET")
	router.HandleFunc("/api/canvas/generators/{id}", ws.handleUpdateGenerator).Methods("PUT")
	router.HandleFunc("/api/canvas/generators/{id}", ws.handleRemoveGenerator).Methods("DELETE")
	router.HandleFunc("/api/canvas/generators/{id}/pause", ws.handlePauseGenerator).Methods("POST")
	router.HandleFunc("/api/canvas/generators/{id}/resume", ws.handleResumeGenerator).Methods("POST")
	router.HandleFunc("/api/canvas/generators/start", ws.handleStartGenerators).Methods("POST")
	router.HandleFunc("/api/canvas/generators/stop", ws.handleStopGenerators).Methods("POST")

	// Measurement management (RESTful)
	router.HandleFunc("/api/canvas/measurements", ws.handleGetMeasurements).Methods("GET")
	router.HandleFunc("/api/canvas/measurements", ws.handleAddMeasurement).Methods("POST")
	router.HandleFunc("/api/canvas/measurements/{id}", ws.handleGetMeasurement).Methods("GET")
	router.HandleFunc("/api/canvas/measurements/{id}", ws.handleUpdateMeasurement).Methods("PUT")
	router.HandleFunc("/api/canvas/measurements/{id}", ws.handleRemoveMeasurement).Methods("DELETE")
}

// Generator handlers
func (ws *WebServer) handleGetGenerators(w http.ResponseWriter, r *http.Request) {
	generators := ws.canvas.GetGenerators()
	ws.sendSuccess(w, generators)
}

func (ws *WebServer) handleAddGenerator(w http.ResponseWriter, r *http.Request) {
	var config GeneratorConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ws.canvas.AddGenerator(&config); err != nil {
		ws.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success", "id": config.ID})
	ws.broadcast("generatorAdded", map[string]interface{}{
		"generator": config,
	})
}

func (ws *WebServer) handleGetGenerator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	generators := ws.canvas.GetGenerators()
	if gen, ok := generators[id]; ok {
		ws.sendSuccess(w, gen)
	} else {
		ws.sendError(w, fmt.Sprintf("Generator '%s' not found", id), http.StatusNotFound)
	}
}

func (ws *WebServer) handleUpdateGenerator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var config GeneratorConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config.ID = id // Ensure ID matches URL
	if err := ws.canvas.UpdateGenerator(&config); err != nil {
		ws.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success"})
	ws.broadcast("generatorUpdated", map[string]interface{}{
		"generator": config,
	})
}

func (ws *WebServer) handleRemoveGenerator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := ws.canvas.RemoveGenerator(id); err != nil {
		ws.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success"})
	ws.broadcast("generatorRemoved", map[string]interface{}{
		"generatorId": id,
	})
}

func (ws *WebServer) handlePauseGenerator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := ws.canvas.PauseGenerator(id); err != nil {
		ws.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success"})
	ws.broadcast("generatorPaused", map[string]interface{}{
		"generatorId": id,
	})
}

func (ws *WebServer) handleResumeGenerator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := ws.canvas.ResumeGenerator(id); err != nil {
		ws.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success"})
	ws.broadcast("generatorResumed", map[string]interface{}{
		"generatorId": id,
	})
}

func (ws *WebServer) handleStartGenerators(w http.ResponseWriter, r *http.Request) {
	if err := ws.canvas.StartGenerators(); err != nil {
		ws.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success"})
	ws.broadcast("generatorsStarted", map[string]interface{}{})
}

func (ws *WebServer) handleStopGenerators(w http.ResponseWriter, r *http.Request) {
	if err := ws.canvas.StopGenerators(); err != nil {
		ws.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success"})
	ws.broadcast("generatorsStopped", map[string]interface{}{})
}

// Measurement handlers
func (ws *WebServer) handleGetMeasurements(w http.ResponseWriter, r *http.Request) {
	measurements := ws.canvas.GetMeasurements()
	ws.sendSuccess(w, measurements)
}

func (ws *WebServer) handleAddMeasurement(w http.ResponseWriter, r *http.Request) {
	var config MeasurementConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ws.canvas.AddMeasurement(&config); err != nil {
		ws.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success", "id": config.ID})
	ws.broadcast("measurementAdded", map[string]interface{}{
		"measurement": config,
	})
}

func (ws *WebServer) handleGetMeasurement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	measurements := ws.canvas.GetMeasurements()
	if meas, ok := measurements[id]; ok {
		ws.sendSuccess(w, meas)
	} else {
		ws.sendError(w, fmt.Sprintf("Measurement '%s' not found", id), http.StatusNotFound)
	}
}

func (ws *WebServer) handleUpdateMeasurement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var config MeasurementConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config.ID = id // Ensure ID matches URL

	// For now, update is remove + add
	if err := ws.canvas.RemoveMeasurement(id); err != nil && !strings.Contains(err.Error(), "not found") {
		ws.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ws.canvas.AddMeasurement(&config); err != nil {
		ws.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success"})
	ws.broadcast("measurementUpdated", map[string]interface{}{
		"measurement": config,
	})
}

func (ws *WebServer) handleRemoveMeasurement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := ws.canvas.RemoveMeasurement(id); err != nil {
		ws.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success"})
	ws.broadcast("measurementRemoved", map[string]interface{}{
		"measurementId": id,
	})
}

// State management handlers
func (ws *WebServer) handleGetState(w http.ResponseWriter, r *http.Request) {
	state, err := ws.canvas.Save()
	if err != nil {
		ws.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.sendSuccess(w, state)
}

func (ws *WebServer) handleSaveState(w http.ResponseWriter, r *http.Request) {
	state, err := ws.canvas.Save()
	if err != nil {
		ws.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// In a real implementation, this might save to a file or database
	// For now, just return the state
	ws.sendSuccess(w, state)
}

func (ws *WebServer) handleRestoreState(w http.ResponseWriter, r *http.Request) {
	var state CanvasState
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ws.canvas.Restore(&state); err != nil {
		ws.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.sendSuccess(w, map[string]string{"status": "success"})
	ws.broadcast("stateRestored", map[string]interface{}{
		"state": state,
	})
}

func (ws *WebServer) handleGetDiagram(w http.ResponseWriter, r *http.Request) {
	diagram, err := ws.canvas.GetSystemDiagram()
	if err != nil {
		ws.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.sendSuccess(w, diagram)
}

// Utility methods
func (ws *WebServer) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: data})
}

func (ws *WebServer) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{Success: false, Error: message})
}

func (ws *WebServer) broadcast(messageType string, data interface{}) {
	message := CanvasWSMessage{
		Type: messageType,
		Data: data,
	}
	
	// Broadcast to all connected WebSocket clients
	ws.wsHandler.mu.RLock()
	defer ws.wsHandler.mu.RUnlock()
	
	for _, client := range ws.wsHandler.clients {
		client.Writer.Send(conc.Message[any]{Value: message})
	}
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}