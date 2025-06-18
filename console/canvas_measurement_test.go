package console

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanvas_MeasurementTracer_Integration(t *testing.T) {
	canvas := NewCanvas()
	defer canvas.Close()

	// Test initialization
	tracer, err := canvas.CreateMeasurementTracer("")
	require.NoError(t, err)
	assert.NotNil(t, tracer)

	// Test getting the same tracer
	tracer2, err := canvas.GetMeasurementTracer("")
	require.NoError(t, err)
	assert.Equal(t, tracer, tracer2)
}

func TestCanvas_MeasurementManagement(t *testing.T) {
	canvas := NewCanvas()
	defer canvas.Close()

	// Initially no measurements
	assert.False(t, canvas.HasMeasurements())
	assert.Len(t, canvas.GetCanvasMeasurements(), 0)

	// Add a measurement
	err := canvas.AddCanvasMeasurement("lat1", "Latency Monitor", "server.Lookup", "latency", true)
	require.NoError(t, err)

	// Verify measurement was added
	assert.True(t, canvas.HasMeasurements())
	measurements := canvas.GetCanvasMeasurements()
	assert.Len(t, measurements, 1)
	assert.Contains(t, measurements, "server.Lookup")

	measurement := measurements["server.Lookup"]
	assert.Equal(t, "lat1", measurement.ID)
	assert.Equal(t, "Latency Monitor", measurement.Name)
	assert.Equal(t, "server.Lookup", measurement.Target)
	assert.Equal(t, "latency", measurement.MetricType)
	assert.True(t, measurement.Enabled)

	// Add another measurement
	err = canvas.AddCanvasMeasurement("tput1", "Throughput Monitor", "server.HandleCreate", "throughput", true)
	require.NoError(t, err)
	assert.Len(t, canvas.GetCanvasMeasurements(), 2)

	// Remove a measurement
	err = canvas.RemoveCanvasMeasurement("server.Lookup")
	require.NoError(t, err)
	assert.Len(t, canvas.GetCanvasMeasurements(), 1)

	// Clear all measurements
	canvas.ClearMeasurements()
	assert.False(t, canvas.HasMeasurements())
	assert.Len(t, canvas.GetCanvasMeasurements(), 0)
}

func TestCanvas_MeasurementRunID(t *testing.T) {
	canvas := NewCanvas()
	defer canvas.Close()

	// Create tracer
	_, err := canvas.CreateMeasurementTracer("")
	require.NoError(t, err)

	// Set run ID
	canvas.SetMeasurementRunID("test_run_123")

	// Verify run ID is set (we can't directly test this without accessing internals,
	// but we can verify the method doesn't panic)
	assert.NotPanics(t, func() {
		canvas.SetMeasurementRunID("another_run_456")
	})
}

func TestCanvas_MeasurementStats(t *testing.T) {
	canvas := NewCanvas()
	defer canvas.Close()

	// Test error when tracer not initialized
	_, err := canvas.GetMeasurementStats()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "measurement tracer not initialized")

	// Initialize tracer
	_, err = canvas.CreateMeasurementTracer("")
	require.NoError(t, err)

	// Now stats should work
	stats, err := canvas.GetMeasurementStats()
	require.NoError(t, err)
	assert.Contains(t, stats, "total_traces")
	assert.Contains(t, stats, "database_path")
}

func TestCanvas_MeasurementMetrics_BeforeInit(t *testing.T) {
	canvas := NewCanvas()
	defer canvas.Close()

	// Test error conditions when tracer not initialized
	since := time.Now().Add(-1 * time.Hour)

	_, err := canvas.GetMeasurementMetrics("server.Lookup", since)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "measurement tracer not initialized")

	_, _, _, _, err = canvas.GetMeasurementPercentiles("server.Lookup", since)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "measurement tracer not initialized")

	_, err = canvas.ExecuteMeasurementSQL("SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "measurement tracer not initialized")
}

func TestCanvas_MeasurementSQL(t *testing.T) {
	canvas := NewCanvas()
	defer canvas.Close()

	// Initialize tracer
	_, err := canvas.CreateMeasurementTracer("")
	require.NoError(t, err)

	// Test SQL query
	results, err := canvas.ExecuteMeasurementSQL("SELECT 1 as test_value")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, int32(1), results[0]["test_value"])
}

func TestCanvas_Close(t *testing.T) {
	canvas := NewCanvas()

	// Initialize measurement tracer
	_, err := canvas.CreateMeasurementTracer("")
	require.NoError(t, err)
	assert.NotNil(t, canvas.measurementTracer)
	assert.NotNil(t, canvas.tsdb)

	// Close canvas
	err = canvas.Close()
	require.NoError(t, err)

	// Verify cleanup
	assert.Nil(t, canvas.measurementTracer)
	assert.Nil(t, canvas.tsdb)

	// Close again should not error
	err = canvas.Close()
	require.NoError(t, err)
}

func TestCanvas_MeasurementWithoutInit(t *testing.T) {
	canvas := NewCanvas()
	defer canvas.Close()

	// Test operations without initializing tracer
	err := canvas.RemoveCanvasMeasurement("server.Lookup")
	require.NoError(t, err) // Should not error

	canvas.ClearMeasurements() // Should not panic
	assert.False(t, canvas.HasMeasurements())
	assert.Len(t, canvas.GetCanvasMeasurements(), 0)

	canvas.SetMeasurementRunID("test") // Should not panic
}

func TestCanvas_MeasureCommandWorkflow(t *testing.T) {
	canvas := NewCanvas()
	defer canvas.Close()

	// Test measure add command workflow
	err := canvas.AddCanvasMeasurement("lat1", "Latency Monitor", "server.Lookup", "latency", true)
	require.NoError(t, err)

	err = canvas.AddCanvasMeasurement("tput1", "Throughput Monitor", "server.HandleCreate", "throughput", true)
	require.NoError(t, err)

	err = canvas.AddCanvasMeasurement("err1", "Error Monitor", "server.HandleUpdate", "errors", true)
	require.NoError(t, err)

	// Test measure list functionality
	measurements := canvas.GetCanvasMeasurements()
	assert.Len(t, measurements, 3)
	assert.Contains(t, measurements, "server.Lookup")
	assert.Contains(t, measurements, "server.HandleCreate")
	assert.Contains(t, measurements, "server.HandleUpdate")

	// Verify measurement details
	latencyMeasurement := measurements["server.Lookup"]
	assert.Equal(t, "lat1", latencyMeasurement.ID)
	assert.Equal(t, "Latency Monitor", latencyMeasurement.Name)
	assert.Equal(t, "latency", latencyMeasurement.MetricType)
	assert.True(t, latencyMeasurement.Enabled)

	// Test measure remove command
	err = canvas.RemoveCanvasMeasurement("server.HandleCreate")
	require.NoError(t, err)
	measurements = canvas.GetCanvasMeasurements()
	assert.Len(t, measurements, 2)
	assert.NotContains(t, measurements, "server.HandleCreate")

	// Test measure stats command
	stats, err := canvas.GetMeasurementStats()
	require.NoError(t, err)
	assert.Contains(t, stats, "total_traces")
	assert.Contains(t, stats, "database_path")
	assert.Equal(t, int64(0), stats["total_traces"]) // No data inserted yet

	// Test measure clear command
	canvas.ClearMeasurements()
	assert.False(t, canvas.HasMeasurements())
	assert.Len(t, canvas.GetCanvasMeasurements(), 0)
}

// TestCanvas_EndToEnd_MeasurementWorkflow validates the complete measure → run → analyze workflow
func TestCanvas_EndToEnd_MeasurementWorkflow(t *testing.T) {
	canvas := NewCanvas()
	defer canvas.Close()

	// Step 1: Load SDL file
	err := canvas.Load("../examples/contacts/contacts.sdl")
	require.NoError(t, err, "Should load contacts SDL file")

	// Step 2: Activate system
	err = canvas.Use("ContactsSystem")
	require.NoError(t, err, "Should activate ContactsSystem")

	// Step 3: Add measurements for different targets
	err = canvas.AddCanvasMeasurement("lat1", "Server Latency", "server.Lookup", "latency", true)
	require.NoError(t, err, "Should add latency measurement")

	err = canvas.AddCanvasMeasurement("db1", "Database Latency", "server.db.Query", "latency", true)
	require.NoError(t, err, "Should add database measurement")

	// Verify measurements are registered
	assert.True(t, canvas.HasMeasurements(), "Should have measurements registered")
	measurements := canvas.GetCanvasMeasurements()
	assert.Len(t, measurements, 2, "Should have 2 measurements")
	assert.Contains(t, measurements, "server.Lookup")
	assert.Contains(t, measurements, "server.db.Query")

	// Step 4: Run simulation with measurements (auto-injects tracer)
	t.Logf("About to run simulation. HasMeasurements: %v", canvas.HasMeasurements())
	t.Logf("Measurements: %+v", canvas.GetCanvasMeasurements())
	err = canvas.Run("baseline", "server.Lookup", WithRuns(10))
	require.NoError(t, err, "Should run simulation with measurements")

	// Verify session variable was created
	_, exists := canvas.sessionVars["baseline"]
	assert.True(t, exists, "Should create session variable")

	// Step 5: Verify measurement data was captured in DuckDB
	// Wait a moment for async writes to complete
	time.Sleep(100 * time.Millisecond)

	stats, err := canvas.GetMeasurementStats()
	require.NoError(t, err, "Should get measurement stats")

	t.Logf("Database stats: %+v", stats)

	// Should have traces captured in database
	totalTraces, ok := stats["total_traces"].(int64)
	assert.True(t, ok, "total_traces should be int64")
	assert.Greater(t, totalTraces, int64(0), "Should have captured measurement traces")

	// Step 6: Verify we can query measurement data
	since := time.Now().Add(-1 * time.Minute)
	metrics, err := canvas.GetMeasurementMetrics("server.Lookup", since)
	require.NoError(t, err, "Should get measurement metrics")
	assert.NotEmpty(t, metrics, "Should have measurement data")

	// Step 7: Verify percentile calculations work
	p50, p90, p95, p99, err := canvas.GetMeasurementPercentiles("server.Lookup", since)
	require.NoError(t, err, "Should calculate percentiles")
	assert.GreaterOrEqual(t, p50, 0.0, "P50 should be non-negative")
	assert.GreaterOrEqual(t, p90, p50, "P90 should be >= P50")
	assert.GreaterOrEqual(t, p95, p90, "P95 should be >= P90")
	assert.GreaterOrEqual(t, p99, p95, "P99 should be >= P95")

	// Step 8: Run a second simulation to verify data accumulation
	err = canvas.Run("load_test", "server.Lookup", WithRuns(50))
	require.NoError(t, err, "Should run second simulation")

	// Verify more traces were captured
	newStats, err := canvas.GetMeasurementStats()
	require.NoError(t, err, "Should get updated stats")

	newTotalTraces, ok := newStats["total_traces"].(int64)
	assert.True(t, ok, "new total_traces should be int64")
	assert.Greater(t, newTotalTraces, totalTraces, "Should have more traces after second run")

	// Step 9: Test SQL interface for custom queries
	sqlResult, err := canvas.ExecuteMeasurementSQL("SELECT COUNT(*) as count FROM traces WHERE target = ?", "server.Lookup")
	require.NoError(t, err, "Should execute SQL query")
	assert.NotEmpty(t, sqlResult, "Should return SQL results")

	count, ok := sqlResult[0]["count"]
	assert.True(t, ok, "Should have count field")
	assert.Greater(t, count, int64(0), "Should have positive count")

	// Step 10: Verify workflow without measurements (should not use tracer)
	canvas.ClearMeasurements()
	assert.False(t, canvas.HasMeasurements(), "Should have no measurements")

	// This run should NOT use the measurement tracer
	err = canvas.Run("no_measurements", "server.Lookup", WithRuns(25))
	require.NoError(t, err, "Should run without measurements (no tracer)")

	// Stats should not change (no new traces added)
	finalStats, err := canvas.GetMeasurementStats()
	require.NoError(t, err, "Should get final stats")

	finalTotalTraces, ok := finalStats["total_traces"].(int64)
	assert.True(t, ok, "final total_traces should be int64")
	assert.Equal(t, newTotalTraces, finalTotalTraces, "Trace count should not change when no measurements are active")
}
