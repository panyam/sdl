package console

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeasurementTracer_Basic(t *testing.T) {
	// Create temporary time-series store
	tempDir := t.TempDir()
	tsdb, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer tsdb.Close()

	// Create measurement tracer
	tracer := NewMeasurementTracer(tsdb, "test_run_001")
	
	// Add a measurement
	measurement := &MeasurementConfig{
		ID:         "lat1",
		Name:       "Latency Monitor",
		Target:     "server.HandleLookup",
		MetricType: "latency",
		Enabled:    true,
	}
	tracer.AddMeasurement(measurement)

	// Verify measurement was added
	measurements := tracer.GetMeasurements()
	assert.Len(t, measurements, 1)
	assert.Equal(t, measurement, measurements["server.HandleLookup"])
	assert.True(t, tracer.HasMeasurements())
}

func TestMeasurementTracer_MeasurementCapture(t *testing.T) {
	tempDir := t.TempDir()
	tsdb, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer tsdb.Close()

	tracer := NewMeasurementTracer(tsdb, "test_run_capture")
	
	// Add latency measurement
	measurement := &MeasurementConfig{
		ID:         "lat1",
		Name:       "Latency Monitor",
		Target:     "server.HandleLookup",
		MetricType: "latency",
		Enabled:    true,
	}
	tracer.AddMeasurement(measurement)

	// Simulate a trace event pair (Enter + Exit) with current timestamp
	now := time.Now()
	currentTime := float64(now.UnixNano()) / 1e9 // Convert to float seconds
	args := []string{"user123", "phone"}
	
	// Simulate Enter event
	eventID := tracer.Enter(currentTime, runtime.EventEnter, "server.HandleLookup", args...)
	
	// Simulate some processing time
	duration := core.Duration(45.5) // 45.5 simulation time units
	retVal := runtime.Value{} // Mock return value
	retVal.Time = duration
	
	// Simulate Exit event - this should trigger measurement capture
	tracer.Exit(currentTime+float64(duration), duration, retVal, nil)
	
	// Verify measurement was captured
	since := time.Now().Add(-1 * time.Hour)
	points, err := tracer.GetMetrics("server.HandleLookup", since)
	require.NoError(t, err)
	assert.Len(t, points, 1)
	
	point := points[0]
	assert.Equal(t, "server.HandleLookup", point.Target)
	assert.Equal(t, 45.5, point.Duration)
	assert.Equal(t, "test_run_capture", point.RunID)
	assert.Equal(t, args, point.Args)
	
	// Verify no error recorded
	assert.Empty(t, point.Error)
	
	_ = eventID // Silence unused variable warning
}

func TestMeasurementTracer_ErrorCapture(t *testing.T) {
	tempDir := t.TempDir()
	tsdb, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer tsdb.Close()

	tracer := NewMeasurementTracer(tsdb, "test_run_error")
	
	// Add error measurement
	measurement := &MeasurementConfig{
		ID:         "err1",
		Name:       "Error Monitor",
		Target:     "server.HandleLookup",
		MetricType: "errors",
		Enabled:    true,
	}
	tracer.AddMeasurement(measurement)

	// Simulate trace with error
	now := time.Now()
	currentTime := float64(now.UnixNano()) / 1e9
	args := []string{"user456"}
	
	tracer.Enter(currentTime, runtime.EventEnter, "server.HandleLookup", args...)
	
	duration := core.Duration(0) // Zero duration for failed call
	retVal := runtime.Value{}
	testError := errors.New("connection timeout")
	
	tracer.Exit(currentTime, duration, retVal, testError)
	
	// Verify error was captured
	since := time.Now().Add(-1 * time.Hour)
	points, err := tracer.GetMetrics("server.HandleLookup", since)
	require.NoError(t, err)
	assert.Len(t, points, 1)
	
	point := points[0]
	assert.Equal(t, "server.HandleLookup", point.Target)
	assert.Equal(t, 1.0, point.Duration) // Error measurement = 1.0
	assert.Equal(t, "connection timeout", point.Error)
}

func TestMeasurementTracer_ThroughputMeasurement(t *testing.T) {
	tempDir := t.TempDir()
	tsdb, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer tsdb.Close()

	tracer := NewMeasurementTracer(tsdb, "test_run_throughput")
	
	// Add throughput measurement
	measurement := &MeasurementConfig{
		ID:         "tput1",
		Name:       "Throughput Monitor",
		Target:     "server.HandleCreate",
		MetricType: "throughput",
		Enabled:    true,
	}
	tracer.AddMeasurement(measurement)

	// Simulate multiple successful calls
	baseTime := time.Now()
	for i := 0; i < 10; i++ {
		currentTime := float64(baseTime.Add(time.Duration(i)*time.Second).UnixNano()) / 1e9
		args := []string{fmt.Sprintf("request_%d", i)}
		
		tracer.Enter(currentTime, runtime.EventEnter, "server.HandleCreate", args...)
		
		duration := core.Duration(25.0)
		retVal := runtime.Value{}
		retVal.Time = duration
		
		tracer.Exit(currentTime+float64(duration), duration, retVal, nil)
	}
	
	// Verify throughput measurements
	since := time.Now().Add(-1 * time.Hour)
	points, err := tracer.GetMetrics("server.HandleCreate", since)
	require.NoError(t, err)
	assert.Len(t, points, 10)
	
	// Each throughput measurement should be 1.0
	for _, point := range points {
		assert.Equal(t, 1.0, point.Duration) // Throughput uses Duration field
		assert.Equal(t, "server.HandleCreate", point.Target)
	}
}

func TestMeasurementTracer_NonMeasuredTarget(t *testing.T) {
	tempDir := t.TempDir()
	tsdb, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer tsdb.Close()

	tracer := NewMeasurementTracer(tsdb, "test_run_non_measured")
	
	// Add measurement for one target only
	measurement := &MeasurementConfig{
		ID:         "lat1",
		Name:       "Latency Monitor",
		Target:     "server.HandleLookup",
		MetricType: "latency",
		Enabled:    true,
	}
	tracer.AddMeasurement(measurement)

	// Simulate call to non-measured target
	now := time.Now()
	currentTime := float64(now.UnixNano()) / 1e9
	tracer.Enter(currentTime, runtime.EventEnter, "server.HandleDelete", "user789")
	
	duration := core.Duration(30.0)
	retVal := runtime.Value{}
	tracer.Exit(currentTime+float64(duration), duration, retVal, nil)
	
	// Verify no measurement was captured for non-measured target
	since := time.Now().Add(-1 * time.Hour)
	points, err := tracer.GetMetrics("server.HandleDelete", since)
	require.NoError(t, err)
	assert.Len(t, points, 0)
	
	// But measured target should still work
	tracer.Enter(currentTime+100, runtime.EventEnter, "server.HandleLookup", "user789")
	tracer.Exit(currentTime+100+float64(duration), duration, retVal, nil)
	
	points, err = tracer.GetMetrics("server.HandleLookup", since)
	require.NoError(t, err)
	assert.Len(t, points, 1)
}

func TestMeasurementTracer_DisabledMeasurement(t *testing.T) {
	tempDir := t.TempDir()
	tsdb, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer tsdb.Close()

	tracer := NewMeasurementTracer(tsdb, "test_run_disabled")
	
	// Add disabled measurement
	measurement := &MeasurementConfig{
		ID:         "lat1",
		Name:       "Disabled Monitor",
		Target:     "server.HandleLookup",
		MetricType: "latency",
		Enabled:    false, // Disabled
	}
	tracer.AddMeasurement(measurement)

	// Simulate call
	now := time.Now()
	currentTime := float64(now.UnixNano()) / 1e9
	tracer.Enter(currentTime, runtime.EventEnter, "server.HandleLookup", "user999")
	
	duration := core.Duration(40.0)
	retVal := runtime.Value{}
	tracer.Exit(currentTime+float64(duration), duration, retVal, nil)
	
	// Verify no measurement was captured
	since := time.Now().Add(-1 * time.Hour)
	points, err := tracer.GetMetrics("server.HandleLookup", since)
	require.NoError(t, err)
	assert.Len(t, points, 0)
}

func TestMeasurementTracer_Management(t *testing.T) {
	tempDir := t.TempDir()
	tsdb, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer tsdb.Close()

	tracer := NewMeasurementTracer(tsdb, "test_run_mgmt")
	
	// Initially no measurements
	assert.False(t, tracer.HasMeasurements())
	assert.Len(t, tracer.GetMeasurements(), 0)
	
	// Add measurements
	measurement1 := &MeasurementConfig{
		ID:         "lat1",
		Target:     "server.HandleLookup",
		MetricType: "latency",
		Enabled:    true,
	}
	measurement2 := &MeasurementConfig{
		ID:         "tput1",
		Target:     "server.HandleCreate",
		MetricType: "throughput",
		Enabled:    true,
	}
	
	tracer.AddMeasurement(measurement1)
	tracer.AddMeasurement(measurement2)
	
	assert.True(t, tracer.HasMeasurements())
	assert.Len(t, tracer.GetMeasurements(), 2)
	
	// Remove one measurement
	tracer.RemoveMeasurement("server.HandleLookup")
	assert.Len(t, tracer.GetMeasurements(), 1)
	
	// Clear all measurements
	tracer.ClearMeasurements()
	assert.False(t, tracer.HasMeasurements())
	assert.Len(t, tracer.GetMeasurements(), 0)
}

func TestMeasurementTracer_RunIDUpdate(t *testing.T) {
	tempDir := t.TempDir()
	tsdb, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer tsdb.Close()

	tracer := NewMeasurementTracer(tsdb, "initial_run")
	
	measurement := &MeasurementConfig{
		ID:         "lat1",
		Target:     "server.HandleLookup",
		MetricType: "latency",
		Enabled:    true,
	}
	tracer.AddMeasurement(measurement)

	// Update run ID
	tracer.SetRunID("updated_run")
	
	// Simulate call
	now := time.Now()
	currentTime := float64(now.UnixNano()) / 1e9
	tracer.Enter(currentTime, runtime.EventEnter, "server.HandleLookup", "test")
	
	duration := core.Duration(50.0)
	retVal := runtime.Value{}
	tracer.Exit(currentTime+float64(duration), duration, retVal, nil)
	
	// Verify measurement has updated run ID
	since := time.Now().Add(-1 * time.Hour)
	points, err := tracer.GetMetrics("server.HandleLookup", since)
	require.NoError(t, err)
	assert.Len(t, points, 1)
	assert.Equal(t, "updated_run", points[0].RunID)
}