package console

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/panyam/sdl/runtime"
)

// MetricsAPI provides REST endpoints for metrics management
type MetricsAPI struct {
	canvas *Canvas
}

// NewMetricsAPI creates a new metrics API handler
func NewMetricsAPI(canvas *Canvas) *MetricsAPI {
	return &MetricsAPI{
		canvas: canvas,
	}
}

// RegisterRoutes registers all metrics-related routes
func (m *MetricsAPI) RegisterRoutes(router *mux.Router) {
	// Measurement management endpoints
	router.HandleFunc("/api/measurements", m.ListMeasurements).Methods("GET")
	router.HandleFunc("/api/measurements", m.AddMeasurement).Methods("POST")
	router.HandleFunc("/api/measurements/{id}", m.GetMeasurement).Methods("GET")
	router.HandleFunc("/api/measurements/{id}", m.RemoveMeasurement).Methods("DELETE")
	
	// Data retrieval endpoints
	router.HandleFunc("/api/measurements/{id}/data", m.GetMeasurementData).Methods("GET")
	router.HandleFunc("/api/measurements/{id}/aggregated", m.GetAggregatedData).Methods("GET")
}

// MeasurementRequest represents the JSON request for adding a measurement
type MeasurementRequest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Component   string   `json:"component"`
	Methods     []string `json:"methods"`
	ResultValue string   `json:"resultValue,omitempty"`
	Metric      string   `json:"metric"`
	Aggregation string   `json:"aggregation"`
	Window      string   `json:"window,omitempty"`
}

// MeasurementResponse represents the JSON response for a measurement
type MeasurementResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Component   string   `json:"component"`
	Methods     []string `json:"methods"`
	ResultValue string   `json:"resultValue"`
	Metric      string   `json:"metric"`
	Aggregation string   `json:"aggregation"`
	Window      string   `json:"window"`
	PointCount  int      `json:"pointCount,omitempty"`
	LastUpdate  string   `json:"lastUpdate,omitempty"`
}

// MetricDataPoint represents a single data point in the response
type MetricDataPoint struct {
	Timestamp float64 `json:"timestamp"` // Seconds since simulation start
	Value     float64 `json:"value"`
	Component string  `json:"component,omitempty"`
	Method    string  `json:"method,omitempty"`
}

// AggregatedDataResponse represents the aggregated data response
type AggregatedDataResponse struct {
	ID          string  `json:"id"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Window      string  `json:"window"`
	PointCount  int     `json:"pointCount"`
	LastUpdate  string  `json:"lastUpdate,omitempty"`
}

// ListMeasurements returns all active measurements
func (m *MetricsAPI) ListMeasurements(w http.ResponseWriter, r *http.Request) {
	// Get the current system's runtime
	currentSystem := m.canvas.CurrentSystem()
	if currentSystem == nil {
		http.Error(w, "No system loaded", http.StatusBadRequest)
		return
	}

	rt := currentSystem.File.Runtime
	if rt == nil || rt.GetMetricStore() == nil {
		// Return empty list if metrics not enabled
		json.NewEncoder(w).Encode([]MeasurementResponse{})
		return
	}

	store := rt.GetMetricStore()
	ids := store.ListMeasurements()
	
	measurements := make([]MeasurementResponse, 0, len(ids))
	for _, id := range ids {
		if info, err := store.GetMeasurementInfo(id); err == nil {
			measurements = append(measurements, MeasurementResponse{
				ID:          info.ID,
				Name:        info.Name,
				Component:   info.Component,
				Methods:     info.Methods,
				ResultValue: info.ResultValue,
				Metric:      string(info.Metric),
				Aggregation: string(info.Aggregation),
				Window:      info.Window.String(),
				PointCount:  info.PointCount,
				LastUpdate:  info.LastUpdate.Format(time.RFC3339),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(measurements)
}

// AddMeasurement adds a new measurement
func (m *MetricsAPI) AddMeasurement(w http.ResponseWriter, r *http.Request) {
	var req MeasurementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ID == "" {
		http.Error(w, "measurement ID is required", http.StatusBadRequest)
		return
	}
	if req.Component == "" {
		http.Error(w, "component is required", http.StatusBadRequest)
		return
	}
	if len(req.Methods) == 0 {
		http.Error(w, "at least one method is required", http.StatusBadRequest)
		return
	}

	// Get the current system's runtime
	currentSystem := m.canvas.CurrentSystem()
	if currentSystem == nil {
		http.Error(w, "No system loaded", http.StatusBadRequest)
		return
	}

	rt := currentSystem.File.Runtime
	rt.EnableMetrics(10000) // Ensure metrics are enabled with default buffer size
	store := rt.GetMetricStore()

	// Parse window duration
	window := 10 * time.Second // Default
	if req.Window != "" {
		if parsed, err := time.ParseDuration(req.Window); err == nil {
			window = parsed
		}
	}

	// Set defaults
	if req.ResultValue == "" {
		req.ResultValue = "*"
	}

	// Create measurement spec
	spec := runtime.MeasurementSpec{
		ID:          req.ID,
		Name:        req.Name,
		Component:   req.Component,
		Methods:     req.Methods,
		ResultValue: req.ResultValue,
		Metric:      runtime.MetricType(req.Metric),
		Aggregation: runtime.AggregationType(req.Aggregation),
		Window:      window,
	}

	// Add the measurement
	if err := store.AddMeasurement(spec); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the created measurement
	resp := MeasurementResponse{
		ID:          spec.ID,
		Name:        spec.Name,
		Component:   spec.Component,
		Methods:     spec.Methods,
		ResultValue: spec.ResultValue,
		Metric:      string(spec.Metric),
		Aggregation: string(spec.Aggregation),
		Window:      spec.Window.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetMeasurement returns details about a specific measurement
func (m *MetricsAPI) GetMeasurement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	currentSystem := m.canvas.CurrentSystem()
	if currentSystem == nil {
		http.Error(w, "No system loaded", http.StatusBadRequest)
		return
	}

	rt := currentSystem.File.Runtime
	if rt == nil || rt.GetMetricStore() == nil {
		http.Error(w, "Metrics not enabled", http.StatusBadRequest)
		return
	}

	store := rt.GetMetricStore()
	info, err := store.GetMeasurementInfo(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := MeasurementResponse{
		ID:          info.ID,
		Name:        info.Name,
		Component:   info.Component,
		Methods:     info.Methods,
		ResultValue: info.ResultValue,
		Metric:      string(info.Metric),
		Aggregation: string(info.Aggregation),
		Window:      info.Window.String(),
		PointCount:  info.PointCount,
		LastUpdate:  info.LastUpdate.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RemoveMeasurement removes a measurement
func (m *MetricsAPI) RemoveMeasurement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	currentSystem := m.canvas.CurrentSystem()
	if currentSystem == nil {
		http.Error(w, "No system loaded", http.StatusBadRequest)
		return
	}

	rt := currentSystem.File.Runtime
	if rt == nil || rt.GetMetricStore() == nil {
		http.Error(w, "Metrics not enabled", http.StatusBadRequest)
		return
	}

	store := rt.GetMetricStore()
	if err := store.RemoveMeasurement(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMeasurementData returns raw data points for a measurement
func (m *MetricsAPI) GetMeasurementData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse query parameters
	limit := 1000 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	currentSystem := m.canvas.CurrentSystem()
	if currentSystem == nil {
		http.Error(w, "No system loaded", http.StatusBadRequest)
		return
	}

	rt := currentSystem.File.Runtime
	if rt == nil || rt.GetMetricStore() == nil {
		http.Error(w, "Metrics not enabled", http.StatusBadRequest)
		return
	}

	store := rt.GetMetricStore()
	points, err := store.GetMeasurementData(id, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Convert to response format
	dataPoints := make([]MetricDataPoint, len(points))
	for i, p := range points {
		dataPoints[i] = MetricDataPoint{
			Timestamp: float64(p.Timestamp) / 1e9, // Convert to seconds
			Value:     p.Value,
			Component: p.Component,
			Method:    p.Method,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataPoints)
}

// GetAggregatedData returns aggregated data for a measurement
func (m *MetricsAPI) GetAggregatedData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	currentSystem := m.canvas.CurrentSystem()
	if currentSystem == nil {
		http.Error(w, "No system loaded", http.StatusBadRequest)
		return
	}

	rt := currentSystem.File.Runtime
	if rt == nil || rt.GetMetricStore() == nil {
		http.Error(w, "Metrics not enabled", http.StatusBadRequest)
		return
	}

	store := rt.GetMetricStore()
	
	// Get aggregated value
	agg, err := store.GetAggregatedData(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to aggregate data: %v", err), http.StatusInternalServerError)
		return
	}

	// Determine unit based on metric type and aggregation
	unit := ""
	if agg.Aggregation == runtime.AggRate {
		unit = "per second"
	} else if strings.Contains(string(agg.Aggregation), "p") || agg.Aggregation == runtime.AggAvg {
		unit = "ms"
	} else {
		unit = "count"
	}

	resp := AggregatedDataResponse{
		ID:         id,
		Value:      agg.Value,
		Unit:       unit,
		Window:     agg.Window.String(),
		PointCount: agg.Count,
		LastUpdate: agg.EndTime.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}