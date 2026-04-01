package services

import (
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/lib/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testFixturePath returns the absolute path to a test fixture file.
func testFixturePath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "test", "fixtures", name)
}

// mockDevEnvPage records all calls from DevEnv for test assertions.
type mockDevEnvPage struct {
	mu sync.Mutex

	// System panel
	systemChangedCalls []struct {
		SystemName       string
		AvailableSystems []string
	}
	availableSystemsCalls [][]string

	// Diagram panel
	diagramCalls []*SystemDiagram

	// Generator panel
	updateGeneratorCalls []struct {
		Name      string
		Generator *protos.Generator
	}
	removeGeneratorCalls []string

	// Metric panel
	updateMetricCalls []struct {
		Name   string
		Metric *protos.Metric
	}
	removeMetricCalls []string

	// Flow panel
	flowRatesCalls []struct {
		Rates    map[string]float64
		Strategy string
	}

	// Console panel
	logCalls []struct {
		Level, Message, Source string
	}
}

func newMockDevEnvPage() *mockDevEnvPage {
	return &mockDevEnvPage{}
}

func (m *mockDevEnvPage) OnSystemChanged(systemName string, availableSystems []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.systemChangedCalls = append(m.systemChangedCalls, struct {
		SystemName       string
		AvailableSystems []string
	}{systemName, availableSystems})
}

func (m *mockDevEnvPage) OnAvailableSystemsChanged(systemNames []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.availableSystemsCalls = append(m.availableSystemsCalls, systemNames)
}

func (m *mockDevEnvPage) UpdateDiagram(diagram *SystemDiagram) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.diagramCalls = append(m.diagramCalls, diagram)
}

func (m *mockDevEnvPage) UpdateGenerator(name string, generator *protos.Generator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateGeneratorCalls = append(m.updateGeneratorCalls, struct {
		Name      string
		Generator *protos.Generator
	}{name, generator})
}

func (m *mockDevEnvPage) RemoveGenerator(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeGeneratorCalls = append(m.removeGeneratorCalls, name)
}

func (m *mockDevEnvPage) UpdateMetric(name string, metric *protos.Metric) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateMetricCalls = append(m.updateMetricCalls, struct {
		Name   string
		Metric *protos.Metric
	}{name, metric})
}

func (m *mockDevEnvPage) RemoveMetric(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeMetricCalls = append(m.removeMetricCalls, name)
}

func (m *mockDevEnvPage) UpdateFlowRates(rates map[string]float64, strategy string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flowRatesCalls = append(m.flowRatesCalls, struct {
		Rates    map[string]float64
		Strategy string
	}{rates, strategy})
}

func (m *mockDevEnvPage) LogMessage(level string, message string, source string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logCalls = append(m.logCalls, struct {
		Level, Message, Source string
	}{level, message, source})
}

// newTestDevEnv creates a DevEnv with a default file resolver for test fixtures.
func newTestDevEnv() *DevEnv {
	resolver := loader.NewDefaultFileResolver()
	return NewDevEnv(resolver)
}

// TestDevEnvCreate verifies that a DevEnv can be constructed with a FileResolver
// and that its initial state is correct (no active system, no generators, empty
// available systems list).
func TestDevEnvCreate(t *testing.T) {
	dev := newTestDevEnv()
	require.NotNil(t, dev)
	assert.Nil(t, dev.ActiveSystem())
	assert.Empty(t, dev.AvailableSystems())
}

// TestDevEnvLoadAndDiscoverSystems verifies that loading an SDL fixture file
// makes the declared systems available via AvailableSystems(). This tests the
// full pipeline: file resolution -> parsing -> system discovery.
func TestDevEnvLoadAndDiscoverSystems(t *testing.T) {
	dev := newTestDevEnv()
	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)

	systems := dev.AvailableSystems()
	assert.Contains(t, systems, "SimpleAppLoadTest")
}

// TestDevEnvLoadNotifiesPage verifies that loading a file triggers
// OnAvailableSystemsChanged on the attached page handler, so the UI
// can update its system selector.
func TestDevEnvLoadNotifiesPage(t *testing.T) {
	dev := newTestDevEnv()
	page := newMockDevEnvPage()
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)

	assert.Len(t, page.availableSystemsCalls, 1)
	assert.Contains(t, page.availableSystemsCalls[0], "SimpleAppLoadTest")
}

// TestDevEnvUseSystem verifies that Use() activates a system by name,
// making it the active system instance. This is the core lifecycle
// operation that wires up generators, metrics, and flow contexts.
func TestDevEnvUseSystem(t *testing.T) {
	dev := newTestDevEnv()
	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)

	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)
	assert.NotNil(t, dev.ActiveSystem())
}

// TestDevEnvUseSystemNotFound verifies that Use() returns an error
// when the requested system name doesn't exist in any loaded file.
func TestDevEnvUseSystemNotFound(t *testing.T) {
	dev := newTestDevEnv()
	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)

	err = dev.Use("NonExistentSystem")
	assert.Error(t, err)
}

// TestDevEnvPanelNotificationsOnUse verifies that when Use() activates a system,
// the page handler receives a full state push:
// - OnSystemChanged with the system name and available systems
// - UpdateDiagram with the system topology
// - UpdateGenerator for each declared generator (from system block)
// - UpdateMetric for each declared metric (from system block)
// This ensures late-joining UIs get a complete state snapshot.
func TestDevEnvPanelNotificationsOnUse(t *testing.T) {
	dev := newTestDevEnv()
	page := newMockDevEnvPage()
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)

	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)

	// System changed notification
	require.Len(t, page.systemChangedCalls, 1)
	assert.Equal(t, "SimpleAppLoadTest", page.systemChangedCalls[0].SystemName)

	// Diagram notification
	require.Len(t, page.diagramCalls, 1)
	assert.Equal(t, "SimpleAppLoadTest", page.diagramCalls[0].SystemName)

	// Generator notifications: fixture declares "traffic" and "health"
	assert.GreaterOrEqual(t, len(page.updateGeneratorCalls), 2)
	genNames := make(map[string]bool)
	for _, call := range page.updateGeneratorCalls {
		genNames[call.Name] = true
	}
	assert.True(t, genNames["traffic"], "expected 'traffic' generator notification")
	assert.True(t, genNames["health"], "expected 'health' generator notification")
}

// TestDevEnvMetricsNotificationsOnUse verifies that when Use() activates a system
// with declared metrics, the page handler receives UpdateMetric for each one.
// Uses the system_with_metrics.sdl fixture which declares 3 metrics.
func TestDevEnvMetricsNotificationsOnUse(t *testing.T) {
	dev := newTestDevEnv()
	page := newMockDevEnvPage()
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_metrics.sdl"))
	require.NoError(t, err)

	err = dev.Use("SimpleAppTest")
	require.NoError(t, err)

	// Fixture declares 3 metrics: request_latency, throughput, health_latency
	assert.GreaterOrEqual(t, len(page.updateMetricCalls), 3)
	metricNames := make(map[string]bool)
	for _, call := range page.updateMetricCalls {
		metricNames[call.Name] = true
	}
	assert.True(t, metricNames["request_latency"])
	assert.True(t, metricNames["throughput"])
	assert.True(t, metricNames["health_latency"])
}

// TestDevEnvGeneratorLifecycle verifies the full generator lifecycle:
// 1. Use() auto-creates declared generators (not started yet in DevEnv)
// 2. Generators can be looked up by name
// 3. RemoveGenerator removes and notifies page
// This tests the CRUD operations on generators from the DevEnv API.
func TestDevEnvGeneratorLifecycle(t *testing.T) {
	dev := newTestDevEnv()
	page := newMockDevEnvPage()
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)
	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)

	// Generators should exist
	dev.generatorsLock.RLock()
	assert.Len(t, dev.generators, 2)
	_, hasTraffic := dev.generators["traffic"]
	_, hasHealth := dev.generators["health"]
	dev.generatorsLock.RUnlock()
	assert.True(t, hasTraffic)
	assert.True(t, hasHealth)

	// Remove one generator
	initialUpdateCalls := len(page.updateGeneratorCalls)
	err = dev.RemoveGenerator("traffic")
	require.NoError(t, err)

	// Verify removal
	dev.generatorsLock.RLock()
	assert.Len(t, dev.generators, 1)
	dev.generatorsLock.RUnlock()

	// Page should have been notified of removal
	assert.Len(t, page.removeGeneratorCalls, 1)
	assert.Equal(t, "traffic", page.removeGeneratorCalls[0])

	// Removing non-existent should error
	err = dev.RemoveGenerator("traffic")
	assert.Error(t, err)

	// Update rate on remaining generator
	err = dev.UpdateGenerator("health", 5.0)
	require.NoError(t, err)
	assert.Greater(t, len(page.updateGeneratorCalls), initialUpdateCalls)
}

// TestDevEnvMetricLifecycle verifies manual metric add/remove operations:
// 1. After Use(), declared metrics exist in the tracer
// 2. RemoveMetric removes by ID and notifies page
// This tests the CRUD operations on metrics from the DevEnv API.
func TestDevEnvMetricLifecycle(t *testing.T) {
	dev := newTestDevEnv()
	page := newMockDevEnvPage()
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_metrics.sdl"))
	require.NoError(t, err)
	err = dev.Use("SimpleAppTest")
	require.NoError(t, err)

	// Metrics should be in the tracer
	require.NotNil(t, dev.metricTracer)
	specs := dev.metricTracer.ListMetricSpec()
	assert.Len(t, specs, 3)

	// Remove one metric
	err = dev.RemoveMetric("request_latency")
	require.NoError(t, err)

	// Verify removal
	specs = dev.metricTracer.ListMetricSpec()
	assert.Len(t, specs, 2)

	// Page should have been notified
	assert.Len(t, page.removeMetricCalls, 1)
	assert.Equal(t, "request_latency", page.removeMetricCalls[0])
}

// TestDevEnvSystemSwitch verifies that switching between systems via Use()
// cleans up the previous system's generators and sets up the new system.
// Generators from the old system should be cleared, and new ones created.
func TestDevEnvSystemSwitch(t *testing.T) {
	dev := newTestDevEnv()
	page := newMockDevEnvPage()
	dev.SetPage(page)

	// Load both fixtures (they define different systems)
	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)
	err = dev.LoadFile(testFixturePath("system_with_metrics.sdl"))
	require.NoError(t, err)

	// Use first system
	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)
	assert.Len(t, page.systemChangedCalls, 1)

	// Use second system
	err = dev.Use("SimpleAppTest")
	require.NoError(t, err)
	assert.Len(t, page.systemChangedCalls, 2)
	assert.Equal(t, "SimpleAppTest", page.systemChangedCalls[1].SystemName)

	// Generators should be from the new system (SimpleAppTest has "traffic" only)
	dev.generatorsLock.RLock()
	genCount := len(dev.generators)
	dev.generatorsLock.RUnlock()
	assert.Equal(t, 1, genCount, "SimpleAppTest declares 1 generator")
}

// TestDevEnvDetachPage verifies that after ClearPage(), no more notifications
// are sent to the previously attached page handler. This prevents stale
// references and ensures clean page lifecycle management.
func TestDevEnvDetachPage(t *testing.T) {
	dev := newTestDevEnv()
	page := newMockDevEnvPage()
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)

	// Detach page before Use
	dev.ClearPage()

	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)

	// No notifications should have been sent after detach
	assert.Empty(t, page.systemChangedCalls)
	assert.Empty(t, page.diagramCalls)
}

// TestDevEnvGetSystemDiagram verifies that GetSystemDiagram returns a valid
// diagram with nodes and edges representing the system topology. The fixture
// has SimpleApp -> SimpleServer -> SimpleDB, so we expect multiple nodes.
func TestDevEnvGetSystemDiagram(t *testing.T) {
	dev := newTestDevEnv()
	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)
	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)

	diagram, err := dev.GetSystemDiagram()
	require.NoError(t, err)
	assert.NotNil(t, diagram)
	assert.Equal(t, "SimpleAppLoadTest", diagram.SystemName)
	assert.NotEmpty(t, diagram.Nodes, "diagram should have nodes for the component topology")
}

// TestDevEnvClose verifies that Close() stops all generators and cleans up
// the metric tracer without panicking, even when called multiple times.
func TestDevEnvClose(t *testing.T) {
	dev := newTestDevEnv()
	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)
	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)

	err = dev.Close()
	require.NoError(t, err)

	// Double close should not panic
	err = dev.Close()
	require.NoError(t, err)
}
