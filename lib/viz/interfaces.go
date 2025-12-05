// package viz defines common interfaces and data structures for generating visualizations.
package viz

import (
	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/lib/runtime"
)

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
	Generate(diagram *protos.SystemDiagram) (string, error)
}

// SequenceDiagramGenerator defines the interface for creating dynamic sequence diagrams from a trace.
type SequenceDiagramGenerator interface {
	Generate(trace *runtime.TraceData) (string, error)
}

// Plotter defines the interface for creating plots and charts.
type Plotter interface {
	Generate(series []DataSeries, title, xLabel, yLabel string) (string, error)
}
