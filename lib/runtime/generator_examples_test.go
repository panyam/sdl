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

// TestContactsSystemDeclarations verifies the contacts example loads with
// generators and metrics resolved against its cache-aside architecture.
func TestContactsSystemDeclarations(t *testing.T) {
	sys, _ := loadSystem(t, "../../examples/contacts/contacts.sdl", "ContactsSystem")
	require.NotNil(t, sys)
	require.Len(t, sys.Generators, 2)
	require.Len(t, sys.Metrics, 1)
	assert.Equal(t, "lookups", sys.Generators[0].Name)
	assert.Equal(t, "inserts", sys.Generators[1].Name)
	assert.Equal(t, "lookup_latency", sys.Metrics[0].Name)
	require.NotNil(t, sys.Generators[0].ResolvedComponent)
	require.NotNil(t, sys.Metrics[0].ResolvedComponent)
}

// TestTwitterDeclarations verifies the Twitter example loads with
// timeline and tweet generators resolved against the fan-out architecture.
func TestTwitterDeclarations(t *testing.T) {
	sys, _ := loadSystem(t, "../../examples/twitter/services.sdl", "Twitter")
	require.NotNil(t, sys)
	require.Len(t, sys.Generators, 2)
	require.Len(t, sys.Metrics, 2)
	assert.Equal(t, "timelines", sys.Generators[0].Name)
	assert.Equal(t, 50.0, sys.Generators[0].RPS())
	assert.Equal(t, "tweets", sys.Generators[1].Name)
	assert.Equal(t, "timeline_latency", sys.Metrics[0].Name)
	assert.Equal(t, "tweet_latency", sys.Metrics[1].Name)
}

// TestNetflixDeclarations verifies the Netflix example loads with
// streaming and upload generators against the CDN/encoder architecture.
func TestNetflixDeclarations(t *testing.T) {
	sys, _ := loadSystem(t, "../../examples/netflix/netflix.sdl", "NetflixSystem")
	require.NotNil(t, sys)
	require.Len(t, sys.Generators, 2)
	require.Len(t, sys.Metrics, 2)
	assert.Equal(t, "streaming", sys.Generators[0].Name)
	assert.Equal(t, 100.0, sys.Generators[0].RPS())
	assert.Equal(t, "stream_latency", sys.Metrics[0].Name)
	assert.Equal(t, "cdn_latency", sys.Metrics[1].Name)
}

// TestDelayExamplesDeclarations verifies both delay examples load with
// generators and metrics for queue buildup and cascading delay analysis.
func TestDelayExamplesDeclarations(t *testing.T) {
	// Queue buildup
	sys1, _ := loadSystem(t, "../../examples/delays/queue_buildup.sdl", "QueueBuildupDemo")
	require.NotNil(t, sys1)
	require.Len(t, sys1.Generators, 1)
	require.Len(t, sys1.Metrics, 1)
	assert.Equal(t, "traffic", sys1.Generators[0].Name)
	assert.Equal(t, 50.0, sys1.Generators[0].RPS())

	// Cascading delays
	sys2, _ := loadSystem(t, "../../examples/delays/cascading_delays.sdl", "CascadingDelayDemo")
	require.NotNil(t, sys2)
	require.Len(t, sys2.Generators, 1)
	require.Len(t, sys2.Metrics, 2)
	assert.Equal(t, "traffic", sys2.Generators[0].Name)
	assert.Equal(t, "api_latency", sys2.Metrics[0].Name)
	assert.Equal(t, "db_latency", sys2.Metrics[1].Name)
}
