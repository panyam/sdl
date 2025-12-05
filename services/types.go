package services

// This package uses proto types directly from github.com/panyam/sdl/gen/go/sdl/v1/models
// for Generator, Metric, and Canvas. The diagram types below are kept as native
// types for now as they are built dynamically from system runtime state.

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
