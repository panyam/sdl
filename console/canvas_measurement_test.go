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
	err := canvas.AddCanvasMeasurement("lat1", "Latency Monitor", "server.HandleLookup", "latency", true)
	require.NoError(t, err)

	// Verify measurement was added
	assert.True(t, canvas.HasMeasurements())
	measurements := canvas.GetCanvasMeasurements()
	assert.Len(t, measurements, 1)
	assert.Contains(t, measurements, "server.HandleLookup")

	measurement := measurements["server.HandleLookup"]
	assert.Equal(t, "lat1", measurement.ID)
	assert.Equal(t, "Latency Monitor", measurement.Name)
	assert.Equal(t, "server.HandleLookup", measurement.Target)
	assert.Equal(t, "latency", measurement.MetricType)
	assert.True(t, measurement.Enabled)

	// Add another measurement
	err = canvas.AddCanvasMeasurement("tput1", "Throughput Monitor", "server.HandleCreate", "throughput", true)
	require.NoError(t, err)
	assert.Len(t, canvas.GetCanvasMeasurements(), 2)

	// Remove a measurement
	err = canvas.RemoveCanvasMeasurement("server.HandleLookup")
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

	_, err := canvas.GetMeasurementMetrics("server.HandleLookup", since)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "measurement tracer not initialized")

	_, _, _, _, err = canvas.GetMeasurementPercentiles("server.HandleLookup", since)
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
	err := canvas.RemoveCanvasMeasurement("server.HandleLookup")
	require.NoError(t, err) // Should not error

	canvas.ClearMeasurements() // Should not panic
	assert.False(t, canvas.HasMeasurements())
	assert.Len(t, canvas.GetCanvasMeasurements(), 0)

	canvas.SetMeasurementRunID("test") // Should not panic
}