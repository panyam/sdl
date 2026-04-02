package runtime

import (
	"testing"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeclarativeGeneratorsParsed(t *testing.T) {
	sys, _ := loadSystem(t, "../../test/fixtures/system_with_generators.sdl", "SimpleAppLoadTest")
	require.NotNil(t, sys)

	// System should have 2 generators resolved from the SDL
	require.Len(t, sys.Generators, 2, "Expected 2 generators from system body")

	// First generator: traffic at 100 rps
	traffic := sys.Generators[0]
	assert.Equal(t, "traffic", traffic.Name)
	assert.Equal(t, "app.server", traffic.Component)
	assert.Equal(t, "HandleRequest", traffic.Method)
	assert.Equal(t, 100.0, traffic.RPS())
	assert.Equal(t, 0.0, traffic.Duration)
	assert.True(t, traffic.Enabled)

	// Second generator: health at 1 per 5s = 0.2 rps
	health := sys.Generators[1]
	assert.Equal(t, "health", health.Name)
	assert.Equal(t, "app.server", health.Component)
	assert.Equal(t, "HealthCheck", health.Method)
	assert.InDelta(t, 0.2, health.RPS(), 0.001)
	assert.Equal(t, 0.0, health.Duration)
}

func TestDeclarativeGeneratorsResolved(t *testing.T) {
	sys, _ := loadSystem(t, "../../test/fixtures/system_with_generators.sdl", "SimpleAppLoadTest")
	require.NotNil(t, sys)
	require.Len(t, sys.Generators, 2)

	// Generators should have resolved component and method references
	traffic := sys.Generators[0]
	require.NotNil(t, traffic.ResolvedComponent, "traffic generator should have resolved component")
	assert.Equal(t, "SimpleServer", traffic.ResolvedComponent.ComponentDecl.Name.Value)
	require.NotNil(t, traffic.ResolvedMethod, "traffic generator should have resolved method")
	assert.Equal(t, "HandleRequest", traffic.ResolvedMethod.Name.Value)

	health := sys.Generators[1]
	require.NotNil(t, health.ResolvedComponent, "health generator should have resolved component")
	require.NotNil(t, health.ResolvedMethod, "health generator should have resolved method")
	assert.Equal(t, "HealthCheck", health.ResolvedMethod.Name.Value)
}

func TestGeneratorRPS(t *testing.T) {
	// rate(100) — 100 per second (default interval)
	g1 := &Generator{Generator: &protos.Generator{Rate: 100}, RateInterval: 1.0}
	assert.Equal(t, 100.0, g1.RPS())

	// rate(1, 5s) — 1 per 5 seconds = 0.2 rps
	g2 := &Generator{Generator: &protos.Generator{Rate: 1}, RateInterval: 5.0}
	assert.InDelta(t, 0.2, g2.RPS(), 0.001)

	// rate(50, 0.1) — 50 per 100ms = 500 rps
	g3 := &Generator{Generator: &protos.Generator{Rate: 50}, RateInterval: 0.1}
	assert.Equal(t, 500.0, g3.RPS())

	// Edge case: zero interval defaults to treating Rate as rps
	g4 := &Generator{Generator: &protos.Generator{Rate: 42}, RateInterval: 0}
	assert.Equal(t, 42.0, g4.RPS())
}
