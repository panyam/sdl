package services

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/panyam/sdl/lib/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testFixturePath returns the absolute path to a test fixture file.
func testFixturePath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "test", "fixtures", name)
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
// OnAvailableSystemsChanged on the attached workspace page, so the UI
// can update its system selector.
func TestDevEnvLoadNotifiesPage(t *testing.T) {
	dev := newTestDevEnv()
	page := NewConsoleWorkspacePage(false)
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)

	assert.Contains(t, page.AvailableSystems, "SimpleAppLoadTest")
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
// the workspace page receives a full state push:
// - OnSystemChanged with the system name
// - UpdateDiagram with the system topology
// - UpdateGenerator for each declared generator (from system block)
// This ensures late-joining UIs get a complete state snapshot.
func TestDevEnvPanelNotificationsOnUse(t *testing.T) {
	dev := newTestDevEnv()
	page := NewConsoleWorkspacePage(false)
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)

	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)

	// System changed
	assert.Equal(t, "SimpleAppLoadTest", page.ActiveSystem)

	// Diagram pushed
	require.NotNil(t, page.Diagram)
	assert.Equal(t, "SimpleAppLoadTest", page.Diagram.SystemName)

	// Generators: fixture declares "traffic" and "health"
	assert.Len(t, page.Generators, 2)
	assert.Contains(t, page.Generators, "traffic")
	assert.Contains(t, page.Generators, "health")
}

// TestDevEnvMetricsNotificationsOnUse verifies that when Use() activates a system
// with declared metrics, the workspace page receives UpdateMetric for each one.
// Uses the system_with_metrics.sdl fixture which declares 3 metrics.
func TestDevEnvMetricsNotificationsOnUse(t *testing.T) {
	dev := newTestDevEnv()
	page := NewConsoleWorkspacePage(false)
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_metrics.sdl"))
	require.NoError(t, err)

	err = dev.Use("SimpleAppTest")
	require.NoError(t, err)

	// Fixture declares 3 metrics: request_latency, throughput, health_latency
	assert.Len(t, page.Metrics, 3)
	assert.Contains(t, page.Metrics, "request_latency")
	assert.Contains(t, page.Metrics, "throughput")
	assert.Contains(t, page.Metrics, "health_latency")
}

// TestDevEnvGeneratorLifecycle verifies the full generator lifecycle:
// 1. Use() auto-creates declared generators
// 2. Generators appear in the page's recorded state
// 3. RemoveGenerator removes from page state
// This tests the CRUD operations on generators from the DevEnv API.
func TestDevEnvGeneratorLifecycle(t *testing.T) {
	dev := newTestDevEnv()
	page := NewConsoleWorkspacePage(false)
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)
	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)

	// Generators should exist in page state
	assert.Len(t, page.Generators, 2)
	assert.Contains(t, page.Generators, "traffic")
	assert.Contains(t, page.Generators, "health")

	// Remove one generator
	err = dev.RemoveGenerator("traffic")
	require.NoError(t, err)

	// Page should reflect removal
	assert.Len(t, page.Generators, 1)
	assert.NotContains(t, page.Generators, "traffic")
	assert.Contains(t, page.Generators, "health")

	// Removing non-existent should error
	err = dev.RemoveGenerator("traffic")
	assert.Error(t, err)

	// Update rate on remaining generator
	err = dev.UpdateGenerator("health", 5.0)
	require.NoError(t, err)
	assert.Equal(t, float64(5.0), page.Generators["health"].Rate)
}

// TestDevEnvMetricLifecycle verifies manual metric add/remove operations:
// 1. After Use(), declared metrics appear in page state
// 2. RemoveMetric removes from page state
// This tests the CRUD operations on metrics from the DevEnv API.
func TestDevEnvMetricLifecycle(t *testing.T) {
	dev := newTestDevEnv()
	page := NewConsoleWorkspacePage(false)
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_metrics.sdl"))
	require.NoError(t, err)
	err = dev.Use("SimpleAppTest")
	require.NoError(t, err)

	// Metrics should be in page state
	assert.Len(t, page.Metrics, 3)

	// Remove one metric
	err = dev.RemoveMetric("request_latency")
	require.NoError(t, err)

	// Page should reflect removal
	assert.Len(t, page.Metrics, 2)
	assert.NotContains(t, page.Metrics, "request_latency")
}

// TestDevEnvSystemSwitch verifies that switching between systems via Use()
// cleans up the previous system's generators and sets up the new system.
// Generators from the old system should be cleared, and new ones created.
func TestDevEnvSystemSwitch(t *testing.T) {
	dev := newTestDevEnv()
	page := NewConsoleWorkspacePage(false)
	dev.SetPage(page)

	// Load both fixtures (they define different systems)
	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)
	err = dev.LoadFile(testFixturePath("system_with_metrics.sdl"))
	require.NoError(t, err)

	// Use first system
	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)
	assert.Equal(t, "SimpleAppLoadTest", page.ActiveSystem)

	// Use second system
	err = dev.Use("SimpleAppTest")
	require.NoError(t, err)
	assert.Equal(t, "SimpleAppTest", page.ActiveSystem)

	// Generators should be from the new system (SimpleAppTest has "traffic" only)
	assert.Len(t, page.Generators, 1, "SimpleAppTest declares 1 generator")
}

// TestDevEnvDetachPage verifies that after ClearPage(), no more notifications
// are sent to the previously attached workspace page. This prevents stale
// references and ensures clean page lifecycle management.
func TestDevEnvDetachPage(t *testing.T) {
	dev := newTestDevEnv()
	page := NewConsoleWorkspacePage(false)
	dev.SetPage(page)

	err := dev.LoadFile(testFixturePath("system_with_generators.sdl"))
	require.NoError(t, err)

	// Detach page before Use
	dev.ClearPage()

	err = dev.Use("SimpleAppLoadTest")
	require.NoError(t, err)

	// No notifications should have been sent after detach
	assert.Empty(t, page.ActiveSystem)
	assert.Nil(t, page.Diagram)
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
