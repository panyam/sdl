// package viz defines common interfaces and data structures for generating visualizations.
package viz

import "github.com/panyam/sdl/runtime"

// --- Common Data Structures ---

// Node represents a component or instance in a static diagram.
type Node struct {
	ID      string       // Unique identifier for the node
	Name    string       // Display name
	Type    string       // Component type for display
	Methods []MethodInfo // Methods provided by this component
	Traffic string       // Current traffic flow (e.g., "0 rps")
}

// MethodInfo represents information about a component method
type MethodInfo struct {
	Name       string  // Method name
	ReturnType string  // Return type (e.g., "Bool", "Int", etc.)
	Traffic    float64 // Current traffic rate in RPS (calculated by FlowEval)
}

// Edge represents a connection between nodes in a static diagram.
type Edge struct {
	FromID      string
	ToID        string
	FromMethod  string  // Source method name (for flow edges)
	ToMethod    string  // Target method name (for flow edges)
	Label       string
	Order       float64 // Execution order (supports decimals for conditional paths)
	Condition   string  // Condition expression if this is a conditional path
	Probability float64 // Probability of this path being taken
	GeneratorID string  // ID of the generator that originated this flow
	Color       string  // Color for visualization (based on generator)
}

// DataPoint represents a single plot point for time-series charts.
type DataPoint struct {
	X int64   // Typically Unix timestamp in milliseconds
	Y float64 // The value (e.g., latency, count)
}

// DataSeries represents a single named series of data points.
type DataSeries struct {
	Name   string
	Points []DataPoint
}

// --- Interfaces for Generators ---

// StaticDiagramGenerator defines the interface for creating static architecture diagrams.
type StaticDiagramGenerator interface {
	Generate(systemName string, nodes []Node, edges []Edge) (string, error)
}

// SequenceDiagramGenerator defines the interface for creating dynamic sequence diagrams from a trace.
type SequenceDiagramGenerator interface {
	Generate(trace *runtime.TraceData) (string, error)
}

// Plotter defines the interface for creating plots and charts.
type Plotter interface {
	Generate(series []DataSeries, title, xLabel, yLabel string) (string, error)
}
