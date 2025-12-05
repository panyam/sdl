package services

import (
	"time"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// ToProto converts the Canvas to its protobuf representation
func (c *Canvas) ToProto() *protos.Canvas {
	c.generatorsLock.RLock()
	defer c.generatorsLock.RUnlock()

	// Get current state
	state := c.GetState()

	// Convert to proto
	return ToProtoCanvas(state)
}

// GetState returns the current state of the canvas as a native type
func (c *Canvas) GetState() *CanvasState {
	c.generatorsLock.RLock()
	defer c.generatorsLock.RUnlock()

	// Collect generators
	var generators []Generator
	for _, genInfo := range c.generators {
		if genInfo.Generator != nil {
			generators = append(generators, *genInfo.Generator)
		}
	}

	// Collect metrics
	var metrics []Metric
	if c.metricTracer != nil {
		// Use the public ListMetrics method to get metrics
		metricsList := c.metricTracer.ListMetrics()
		for _, m := range metricsList {
			if m != nil {
				metrics = append(metrics, *m)
			}
		}
	}

	// Collect loaded files
	var loadedFiles []string
	// TODO: Track loaded files properly
	if c.runtime != nil && c.runtime.Loader != nil {
		// For now, just return empty - we'll need to track this separately
		loadedFiles = []string{}
	}

	activeSystem := ""
	if c.activeSystem != nil {
		activeSystem = c.activeSystem.System.Name.Value
	}

	return &CanvasState{
		CreatedAt:    time.Now(), // TODO: track actual creation time
		UpdatedAt:    time.Now(),
		ID:           c.id,
		ActiveSystem: activeSystem,
		LoadedFiles:  loadedFiles,
		Generators:   generators,
		Metrics:      metrics,
	}
}
