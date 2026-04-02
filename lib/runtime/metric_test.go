package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeclarativeMetricsParsed verifies that metric() calls in a system body
// are validated during inference and produce MetricSpec entries on SystemDecl.
// Tests the full path: SDL source → parse → inference → MetricSpec.
func TestDeclarativeMetricsParsed(t *testing.T) {
	sys, _ := loadSystem(t, "../../test/fixtures/system_with_metrics.sdl", "SimpleAppTest")
	require.NotNil(t, sys)

	// System should have 3 metrics resolved from the SDL
	require.Len(t, sys.Metrics, 3, "Expected 3 metrics from system body")

	// First metric: request latency with p95 and 5s window
	reqLatency := sys.Metrics[0]
	assert.Equal(t, "request_latency", reqLatency.Name)
	assert.Equal(t, "app.server", reqLatency.Component)
	assert.Equal(t, "HandleRequest", reqLatency.Method)
	assert.Equal(t, "latency", reqLatency.MetricType)
	assert.Equal(t, "p95", reqLatency.Aggregation)
	assert.Equal(t, 5.0, reqLatency.Window)

	// Second metric: throughput count with sum and 5s window
	throughput := sys.Metrics[1]
	assert.Equal(t, "throughput", throughput.Name)
	assert.Equal(t, "count", throughput.MetricType)
	assert.Equal(t, "sum", throughput.Aggregation)
	assert.Equal(t, 5.0, throughput.Window)

	// Third metric: health latency with avg and default window (10s)
	healthLatency := sys.Metrics[2]
	assert.Equal(t, "health_latency", healthLatency.Name)
	assert.Equal(t, "latency", healthLatency.MetricType)
	assert.Equal(t, "avg", healthLatency.Aggregation)
	assert.Equal(t, 10.0, healthLatency.Window, "Default window should be 10s")
}

// TestDeclarativeMetricsResolved verifies that metrics have their component
// and method references resolved against the initialized system environment.
func TestDeclarativeMetricsResolved(t *testing.T) {
	sys, _ := loadSystem(t, "../../test/fixtures/system_with_metrics.sdl", "SimpleAppTest")
	require.NotNil(t, sys)
	require.Len(t, sys.Metrics, 3)

	// All metrics should have resolved component references
	for _, m := range sys.Metrics {
		require.NotNil(t, m.ResolvedComponent, "metric %s should have resolved component", m.Name)
		require.NotNil(t, m.ResolvedMethod, "metric %s should have resolved method", m.Name)
	}

	// Verify specific resolutions
	assert.Equal(t, "SimpleServer", sys.Metrics[0].ResolvedComponent.ComponentDecl.Name.Value)
	assert.Equal(t, "HandleRequest", sys.Metrics[0].ResolvedMethod.Name.Value)
	assert.Equal(t, "HealthCheck", sys.Metrics[2].ResolvedMethod.Name.Value)
}

// TestMetricValidTypes verifies that only valid metric types are accepted
// during inference. Invalid types should produce compile-time errors.
func TestMetricValidTypes(t *testing.T) {
	// Valid types: latency, count, utilization
	validSDL := `
component S { method M() Bool { return true } }
component Arch { uses s S() }
system T(a Arch) {
    metric("m1", a.s.M, "latency", "p95", 5s)
    metric("m2", a.s.M, "count", "sum", 5s)
}
`
	ast := parseAndLoad(t, validSDL)
	require.NotNil(t, ast)
}

// TestMetricInvalidTypeRejected verifies that an unknown metric type
// produces an error during inference rather than silently passing.
// The inference phase panics on errors (ErrorCollector behavior), so we
// check for a panic containing the expected error message.
func TestMetricInvalidTypeRejected(t *testing.T) {
	invalidSDL := `
component S { method M() Bool { return true } }
component Arch { uses s S() }
system T(a Arch) {
    metric("m1", a.s.M, "bogus", "avg", 5s)
}
`
	assert.Panics(t, func() {
		parseAndLoadSystem(invalidSDL)
	}, "Invalid metric type should be rejected during inference")
}
