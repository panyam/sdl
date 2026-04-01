package services

import (
	"time"

	"github.com/panyam/sdl/lib/runtime"
)

// SimulationContext provides the execution context that generators and metrics
// need without coupling them to a specific orchestrator (Canvas or DevEnv).
// Both Canvas and DevEnv implement this interface.
type SimulationContext interface {
	// GetTracer returns the tracer used for metric collection during simulation eval.
	GetTracer() runtime.Tracer

	// GetSimulationStartTime returns when the simulation was started.
	GetSimulationStartTime() time.Time

	// IsSimulationStarted returns true if the simulation clock is running.
	IsSimulationStarted() bool

	// GetSimulationTime returns the current virtual simulation time in seconds.
	GetSimulationTime() float64
}
