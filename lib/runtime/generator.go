package runtime

import "github.com/panyam/sdl/lib/core"

// Generator represents a traffic generator bound to a system.
// This is the runtime representation — created from GeneratorSpec during system init,
// or added dynamically via CLI/API. The services layer wraps this with execution
// machinery (goroutines, stop channels, etc.).
type Generator struct {
	// Name is the unique identifier within a system
	Name string

	// Target
	ComponentPath string // Dot-separated component path (e.g., "arch.webserver")
	MethodName    string // Target method name (e.g., "RequestRide")

	// Rate configuration
	Rate         float64       // Calls per interval
	RateInterval core.Duration // Interval in seconds (default 1.0 = per second)

	// Lifecycle
	Duration core.Duration // Duration in seconds (0 = forever)
	Enabled  bool          // Whether the generator is active

	// Resolved references (populated during system init)
	ResolvedComponent *ComponentInstance
	ResolvedMethod    *MethodDecl
}

// RPS returns the effective requests per second.
func (g *Generator) RPS() float64 {
	if g.RateInterval <= 0 {
		return g.Rate // assume per second
	}
	return g.Rate / g.RateInterval
}

// NewGeneratorFromSpec creates a Generator from a compile-time GeneratorSpec.
func NewGeneratorFromSpec(spec *GeneratorSpec) *Generator {
	return &Generator{
		Name:          spec.Name,
		ComponentPath: spec.ComponentPath,
		MethodName:    spec.MethodName,
		Rate:          spec.Rate,
		RateInterval:  spec.RateInterval,
		Duration:      spec.Duration,
		Enabled:       true,
	}
}
