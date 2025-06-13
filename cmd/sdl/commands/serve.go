package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/panyam/sdl/console"
	"github.com/spf13/cobra"
)

var (
	servePort = 8080
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
)

// WebServer wraps the Canvas API for HTTP access
type WebServer struct {
	canvas *console.Canvas
	clients map[*websocket.Conn]bool
}

func NewWebServer() *WebServer {
	return &WebServer{
		canvas:  console.NewCanvas(),
		clients: make(map[*websocket.Conn]bool),
	}
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

type SeriesInfo struct {
	Name string `json:"name"`
	From string `json:"from"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HTTP Handlers
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
	ws.broadcast(map[string]interface{}{
		"type": "fileLoaded",
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
	ws.broadcast(map[string]interface{}{
		"type":   "systemActivated",
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
	ws.broadcast(map[string]interface{}{
		"type":  "parameterChanged",
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

	err := ws.canvas.Run(req.VarName, req.Target, console.WithRuns(runs))
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
	ws.broadcast(map[string]interface{}{
		"type":    "simulationCompleted",
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
	opts := []console.PlotOption{
		console.WithOutput(req.OutputFile),
	}
	for _, series := range req.Series {
		opts = append(opts, console.WithSeries(series.Name, series.From))
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
	ws.broadcast(map[string]interface{}{
		"type":       "plotGenerated",
		"outputFile": req.OutputFile,
		"series":     req.Series,
	})
}

// WebSocket handler for real-time updates
func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	ws.clients[conn] = true
	defer delete(ws.clients, conn)

	log.Printf("WebSocket client connected")

	// Keep connection alive and handle any incoming messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
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

func (ws *WebServer) broadcast(message map[string]interface{}) {
	for client := range ws.clients {
		err := client.WriteJSON(message)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(ws.clients, client)
		}
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

// Serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the SDL web visualization server",
	Long: `Start an HTTP server that provides a web interface for interactive SDL system visualization.
	
The server provides:
- REST API for Canvas operations (load, use, set, run, plot)
- WebSocket connection for real-time updates
- Static file serving for the frontend dashboard

Example:
  sdl serve --port 8080`,
	Run: func(cmd *cobra.Command, args []string) {
		webServer := NewWebServer()

		// Create router
		r := mux.NewRouter()
		r.Use(corsMiddleware)

		// API routes
		api := r.PathPrefix("/api").Subrouter()
		api.HandleFunc("/load", webServer.handleLoad).Methods("POST")
		api.HandleFunc("/use", webServer.handleUse).Methods("POST")
		api.HandleFunc("/set", webServer.handleSet).Methods("POST")
		api.HandleFunc("/run", webServer.handleRun).Methods("POST")
		api.HandleFunc("/plot", webServer.handlePlot).Methods("POST")
		api.HandleFunc("/live", webServer.handleWebSocket)

		// Static file serving (will serve the built frontend)
		r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist/")))

		addr := fmt.Sprintf(":%d", servePort)
		fmt.Printf("🚀 SDL Web Server starting on http://localhost%s\n", addr)
		fmt.Printf("📊 Dashboard: http://localhost%s\n", addr)
		fmt.Printf("🔌 API: http://localhost%s/api\n", addr)
		fmt.Printf("📡 WebSocket: ws://localhost%s/api/live\n", addr)

		log.Fatal(http.ListenAndServe(addr, r))
	},
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port to serve the web interface on")
	rootCmd.AddCommand(serveCmd)
}