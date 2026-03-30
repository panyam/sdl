package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBitlyGenerators(t *testing.T) {
	sys, _ := loadSystem(t, "../../examples/bitly/mvp.sdl", "Bitly")
	require.NotNil(t, sys)
	require.Len(t, sys.Generators, 1)

	baseline := sys.Generators[0]
	assert.Equal(t, "baseline", baseline.Name)
	assert.Equal(t, "arch.app", baseline.ComponentPath)
	assert.Equal(t, "Shorten", baseline.MethodName)
	assert.Equal(t, 5.0, baseline.RPS())
	require.NotNil(t, baseline.ResolvedComponent)
	require.NotNil(t, baseline.ResolvedMethod)

	// Metrics
	require.Len(t, sys.Metrics, 2)
	assert.Equal(t, "shorten_latency", sys.Metrics[0].Name)
	assert.Equal(t, "latency", sys.Metrics[0].MetricType)
	assert.Equal(t, "redirect_latency", sys.Metrics[1].Name)
}

func TestUberMVPGenerators(t *testing.T) {
	sys, _ := loadSystem(t, "../../examples/uber/mvp.sdl", "UberMVP")
	require.NotNil(t, sys)
	require.Len(t, sys.Generators, 2)

	baseline := sys.Generators[0]
	assert.Equal(t, "baseline", baseline.Name)
	assert.Equal(t, "arch.webserver", baseline.ComponentPath)
	assert.Equal(t, "RequestRide", baseline.MethodName)
	assert.Equal(t, 5.0, baseline.RPS())
	require.NotNil(t, baseline.ResolvedComponent)
	require.NotNil(t, baseline.ResolvedMethod)

	drivers := sys.Generators[1]
	assert.Equal(t, "drivers", drivers.Name)
	assert.Equal(t, 10.0, drivers.RPS())

	// Metrics
	require.Len(t, sys.Metrics, 3)
	assert.Equal(t, "request_latency", sys.Metrics[0].Name)
	assert.Equal(t, "p90", sys.Metrics[0].Aggregation)
	assert.Equal(t, "maps_latency", sys.Metrics[1].Name)
	assert.Equal(t, "db_query", sys.Metrics[2].Name)
}

func TestUberIntermediateGenerators(t *testing.T) {
	sys, _ := loadSystem(t, "../../examples/uber/intermediate.sdl", "UberIntermediate")
	require.NotNil(t, sys)
	require.Len(t, sys.Generators, 2)

	assert.Equal(t, "baseline", sys.Generators[0].Name)
	assert.Equal(t, 20.0, sys.Generators[0].RPS())
	require.NotNil(t, sys.Generators[0].ResolvedComponent)

	assert.Equal(t, "drivers", sys.Generators[1].Name)
	assert.Equal(t, 50.0, sys.Generators[1].RPS())
	require.NotNil(t, sys.Generators[1].ResolvedComponent)
}

func TestUberModernGenerators(t *testing.T) {
	sys, _ := loadSystem(t, "../../examples/uber/modern.sdl", "UberModern")
	require.NotNil(t, sys)
	require.Len(t, sys.Generators, 2)

	assert.Equal(t, "baseline", sys.Generators[0].Name)
	assert.Equal(t, 100.0, sys.Generators[0].RPS())
	require.NotNil(t, sys.Generators[0].ResolvedComponent)

	assert.Equal(t, "drivers", sys.Generators[1].Name)
	assert.Equal(t, 1000.0, sys.Generators[1].RPS())
	require.NotNil(t, sys.Generators[1].ResolvedComponent)
}
