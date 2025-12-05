package services

import (
	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto converts the Canvas to its protobuf representation
func (c *Canvas) ToProto() *protos.Canvas {
	c.generatorsLock.RLock()
	defer c.generatorsLock.RUnlock()

	// Collect generators - already proto types
	var generators []*protos.Generator
	for _, genInfo := range c.generators {
		if genInfo.Generator != nil {
			generators = append(generators, genInfo.Generator)
		}
	}

	// Collect metrics - already proto types from ListMetrics()
	var metrics []*protos.Metric
	if c.metricTracer != nil {
		metrics = c.metricTracer.ListMetrics()
	}

	activeSystem := ""
	if c.activeSystem != nil {
		activeSystem = c.activeSystem.System.Name.Value
	}

	return &protos.Canvas{
		CreatedAt:      timestamppb.Now(), // TODO: track actual creation time
		UpdatedAt:      timestamppb.Now(),
		Id:             c.id,
		ActiveSystem:   activeSystem,
		SystemContents: "", // TODO: Track system contents when loaded
		Recipes:        map[string]string{},
		Generators:     generators,
		Metrics:        metrics,
	}
}
