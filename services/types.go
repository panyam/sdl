package services

import (
	"time"
)

// Native types that don't depend on protobuf
// These are used internally by Canvas and converted to/from proto at service boundaries

// Generator represents a traffic generator configuration
type Generator struct {
	CreatedAt time.Time
	UpdatedAt time.Time

	// Core fields
	ID        string
	CanvasID  string
	Name      string
	Component string
	Method    string
	Rate      float64
	Duration  float64 // 0 means run forever
	Enabled   bool
}

// Metric represents a metric collection configuration
type Metric struct {
	CreatedAt time.Time
	UpdatedAt time.Time

	// Core fields
	ID                string
	CanvasID          string
	Name              string
	Component         string
	Methods           []string
	Enabled           bool
	MetricType        string  // "count" or "latency"
	Aggregation       string  // "p50", "p99", "avg", etc.
	AggregationWindow float64 // in seconds

	// Result matching
	MatchResult     string
	MatchResultType string

	// Stats
	OldestTimestamp float64
	NewestTimestamp float64
	NumDataPoints   int64
}

// SystemDiagram represents the visual structure of a system
type SystemDiagram struct {
	SystemName string
	Nodes      []DiagramNode
	Edges      []DiagramEdge
}

// DiagramNode represents a component in the system diagram
type DiagramNode struct {
	ID       string
	Name     string
	Type     string // Component type for display
	Methods  []MethodInfo
	Traffic  string // Current traffic flow (e.g., "0 rps")
	FullPath string // Full path from system root
	Icon     string // Icon identifier
}

// MethodInfo represents information about a component method
type MethodInfo struct {
	Name       string
	ReturnType string
	Traffic    float64 // Current traffic rate in RPS
}

// DiagramEdge represents a connection between nodes
type DiagramEdge struct {
	FromID      string
	ToID        string
	FromMethod  string // Source method name
	ToMethod    string // Target method name
	Label       string
	Order       float64 // Execution order
	Condition   string  // Condition expression if conditional
	Probability float64 // Probability of this path
	GeneratorID string  // ID of originating generator
	Color       string  // Visualization color
}

// CanvasState represents the current state of a canvas
type CanvasState struct {
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ID           string
	ActiveSystem string
	LoadedFiles  []string
	Generators   []Generator
	Metrics      []Metric
}
