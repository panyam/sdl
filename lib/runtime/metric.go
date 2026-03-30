package runtime

// Metric represents a declared metric bound to a system.
// Created from MetricSpec during system init. The services layer wraps this
// with collection machinery (event channels, stores, aggregation).
type Metric struct {
	// Name is the unique identifier within a system
	Name string

	// Target
	ComponentPath string // Dot-separated component path (e.g., "arch.webserver")
	MethodName    string // Target method name, empty for utilization metrics

	// Configuration
	MetricType  string  // "count", "latency", "utilization"
	Aggregation string  // "sum", "avg", "min", "max", "p50", "p90", "p95", "p99"
	Window      float64 // Aggregation window in seconds

	// Resolved references (populated during system init)
	ResolvedComponent *ComponentInstance
	ResolvedMethod    *MethodDecl // nil for utilization metrics
}

// NewMetricFromSpec creates a Metric from a compile-time MetricSpec.
func NewMetricFromSpec(spec *MetricSpec) *Metric {
	return &Metric{
		Name:          spec.Name,
		ComponentPath: spec.ComponentPath,
		MethodName:    spec.MethodName,
		MetricType:    spec.MetricType,
		Aggregation:   spec.Aggregation,
		Window:        spec.Window,
	}
}
