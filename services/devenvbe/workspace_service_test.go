package devenvbe

import (
	"context"
	"path/filepath"
	goruntime "runtime"
	"testing"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/lib/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFixturePath(name string) string {
	_, filename, _, _ := goruntime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "test", "fixtures", name)
}

func newTestService() *WorkspaceService {
	return NewWorkspaceService(loader.NewDefaultFileResolver())
}

func loadAndUse(t *testing.T, svc *WorkspaceService, fixture, system string) {
	t.Helper()
	ctx := context.Background()
	_, err := svc.LoadFile(ctx, &protos.LoadFileRequest{SdlFilePath: testFixturePath(fixture)})
	require.NoError(t, err)
	_, err = svc.UseSystem(ctx, &protos.UseSystemRequest{SystemName: system})
	require.NoError(t, err)
}

// TestDevEnvWorkspaceServiceLoadAndUse verifies that the devenvbe backend
// can load an SDL file and activate a system. This is the core lifecycle
// that all other operations depend on.
func TestDevEnvWorkspaceServiceLoadAndUse(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	_, err := svc.LoadFile(ctx, &protos.LoadFileRequest{
		SdlFilePath: testFixturePath("system_with_generators.sdl"),
	})
	require.NoError(t, err)

	_, err = svc.UseSystem(ctx, &protos.UseSystemRequest{
		SystemName: "SimpleAppLoadTest",
	})
	require.NoError(t, err)

	assert.Equal(t, "SimpleAppLoadTest", svc.DevEnv.GetActiveSystemName())
}

// TestDevEnvWorkspaceServiceGeneratorLifecycle verifies CRUD operations on
// generators through the proto-typed WorkspaceService interface. Tests that
// declared generators appear in ListGenerators after Use(), and that
// manual add/delete works correctly.
func TestDevEnvWorkspaceServiceGeneratorLifecycle(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	loadAndUse(t, svc, "system_with_generators.sdl", "SimpleAppLoadTest")

	// List declared generators
	listResp, err := svc.ListGenerators(ctx, &protos.ListGeneratorsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Generators, 2, "fixture declares 2 generators")

	// Delete one
	_, err = svc.DeleteGenerator(ctx, &protos.DeleteGeneratorRequest{
		GeneratorName: "traffic",
	})
	require.NoError(t, err)

	listResp, err = svc.ListGenerators(ctx, &protos.ListGeneratorsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Generators, 1)
}

// TestDevEnvWorkspaceServiceMetricLifecycle verifies that declared metrics
// appear after Use() and can be deleted through the proto interface.
func TestDevEnvWorkspaceServiceMetricLifecycle(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	loadAndUse(t, svc, "system_with_metrics.sdl", "SimpleAppTest")

	// List declared metrics
	listResp, err := svc.ListMetrics(ctx, &protos.ListMetricsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Metrics, 3, "fixture declares 3 metrics")

	// Delete one
	_, err = svc.DeleteMetric(ctx, &protos.DeleteMetricRequest{
		MetricName: "request_latency",
	})
	require.NoError(t, err)

	listResp, err = svc.ListMetrics(ctx, &protos.ListMetricsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Metrics, 2)
}

// TestDevEnvWorkspaceServiceEvaluateFlows verifies that flow evaluation
// returns component rates for the active system's generators.
func TestDevEnvWorkspaceServiceEvaluateFlows(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	loadAndUse(t, svc, "system_with_generators.sdl", "SimpleAppLoadTest")

	resp, err := svc.EvaluateFlows(ctx, &protos.EvaluateFlowsRequest{
		Strategy: "runtime",
	})
	require.NoError(t, err)
	assert.Equal(t, "runtime", resp.Strategy)
	assert.NotEmpty(t, resp.ComponentRates, "should have flow rates for components")
}

// TestDevEnvWorkspaceServiceGetDiagram verifies that GetSystemDiagram returns
// a valid diagram with the system name and nodes.
func TestDevEnvWorkspaceServiceGetDiagram(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	loadAndUse(t, svc, "system_with_generators.sdl", "SimpleAppLoadTest")

	resp, err := svc.GetSystemDiagram(ctx, &protos.GetSystemDiagramRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.Diagram)
	assert.Equal(t, "SimpleAppLoadTest", resp.Diagram.SystemName)
	assert.NotEmpty(t, resp.Diagram.Nodes)
}

// TestDevEnvWorkspaceServiceExecuteTrace verifies that ExecuteTrace runs a
// single simulated call and returns trace events.
func TestDevEnvWorkspaceServiceExecuteTrace(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	loadAndUse(t, svc, "system_with_generators.sdl", "SimpleAppLoadTest")

	resp, err := svc.ExecuteTrace(ctx, &protos.ExecuteTraceRequest{
		Component: "app.server",
		Method:    "HandleRequest",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.TraceData)
	assert.Equal(t, "SimpleAppLoadTest", resp.TraceData.System)
	assert.NotEmpty(t, resp.TraceData.Events, "should have trace events")
}

// TestDevEnvWorkspaceServiceGetFlowState verifies that GetFlowState returns
// current flow rates after evaluation.
func TestDevEnvWorkspaceServiceGetFlowState(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	loadAndUse(t, svc, "system_with_generators.sdl", "SimpleAppLoadTest")

	// Evaluate first
	svc.EvaluateFlows(ctx, &protos.EvaluateFlowsRequest{Strategy: "runtime"})

	resp, err := svc.GetFlowState(ctx, &protos.GetFlowStateRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.State)
	assert.Equal(t, "runtime", resp.State.Strategy)
	assert.NotEmpty(t, resp.State.Rates)
}
