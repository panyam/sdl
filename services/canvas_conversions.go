package services

import (
	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto returns the canvas proto with runtime state merged in.
// Metadata comes from c.Proto; runtime state (generators, metrics,
// active system, loaded systems) is computed from live objects.
func (c *Canvas) ToProto() *protos.Canvas {
	c.generatorsLock.RLock()
	defer c.generatorsLock.RUnlock()

	// Start with a clone of the stored proto (has id, name, description, system_contents)
	out := proto.Clone(c.Proto).(*protos.Canvas)
	out.UpdatedAt = timestamppb.Now()

	// Merge runtime state
	if c.activeSystem != nil {
		out.ActiveSystem = c.activeSystem.System.Name.Value
	}

	out.LoadedSystemNames = c.GetAvailableSystemNames()

	// Collect generators
	var generators []*protos.Generator
	for _, genInfo := range c.generators {
		if genInfo.Generator != nil {
			generators = append(generators, genInfo.Generator)
		}
	}
	out.Generators = generators

	// Collect metrics
	if c.metricTracer != nil {
		out.Metrics = c.metricTracer.ListMetrics()
	}

	return out
}
