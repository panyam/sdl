package console

import (
	"fmt"
	"sync"
	"time"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/runtime"
)

// MeasurementTracer extends ExecutionTracer to capture measurements
type MeasurementTracer struct {
	*runtime.ExecutionTracer
	tsdb         *DuckDBTimeSeriesStore
	measurements map[string]*MeasurementConfig // target -> measurement config
	runID        string                        // current simulation run ID
	mu           sync.RWMutex
}

// NewMeasurementTracer creates a new measurement tracer
func NewMeasurementTracer(tsdb *DuckDBTimeSeriesStore, runID string) *MeasurementTracer {
	return &MeasurementTracer{
		ExecutionTracer: runtime.NewExecutionTracer(),
		tsdb:            tsdb,
		measurements:    make(map[string]*MeasurementConfig),
		runID:           runID,
	}
}

// AddMeasurement registers a new measurement target
func (mt *MeasurementTracer) AddMeasurement(config *MeasurementConfig) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.measurements[config.Target] = config
}

// RemoveMeasurement unregisters a measurement target
func (mt *MeasurementTracer) RemoveMeasurement(target string) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	delete(mt.measurements, target)
}

// GetMeasurements returns all registered measurements
func (mt *MeasurementTracer) GetMeasurements() map[string]*MeasurementConfig {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*MeasurementConfig)
	for k, v := range mt.measurements {
		result[k] = v
	}
	return result
}

// ClearMeasurements removes all measurement targets
func (mt *MeasurementTracer) ClearMeasurements() {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.measurements = make(map[string]*MeasurementConfig)
}

// HasMeasurements returns true if any measurements are registered
func (mt *MeasurementTracer) HasMeasurements() bool {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return len(mt.measurements) > 0
}

// SetRunID updates the current run identifier
func (mt *MeasurementTracer) SetRunID(runID string) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.runID = runID
}

// Exit overrides the base tracer to capture measurements
func (mt *MeasurementTracer) Exit(ts core.Duration, duration core.Duration, comp *runtime.ComponentInstance, method *runtime.MethodDecl, retVal runtime.Value, err error) {
	// Check if we have any measurements to capture before calling base tracer
	var target string
	var args []string
	var shouldCapture bool
	
	if mt.HasMeasurements() {
		// Get the corresponding Enter event to find the target BEFORE calling base tracer
		events := mt.ExecutionTracer.Events
		if len(events) >= 1 {
			// Find the most recent Enter event (should be the last event since Exit hasn't been called yet)
			for i := len(events) - 1; i >= 0; i-- {
				if events[i].Kind == runtime.EventEnter {
					target = events[i].Target()
					args = events[i].Arguments
					shouldCapture = true
					break
				}
			}
		}
	}

	// Always call the base tracer to maintain proper event recording
	mt.ExecutionTracer.Exit(ts, duration, comp, method, retVal, err)

	// Now capture measurement if we found a target
	if shouldCapture {
		// Check if this target is being measured
		mt.mu.RLock()
		measurement, exists := mt.measurements[target]
		mt.mu.RUnlock()

		if exists && measurement.Enabled {
			// Extract metric based on measurement type and store
			mt.storeMeasurement(measurement, ts, duration, retVal, err, args)
		}
	}
}

// storeMeasurement extracts the appropriate metric and stores it in the time-series database
func (mt *MeasurementTracer) storeMeasurement(measurement *MeasurementConfig, ts float64, duration core.Duration, retVal runtime.Value, err error, args []string) {
	// Convert simulation timestamp to nanoseconds
	timestampNs := int64(ts * 1e9)

	// Extract the metric value based on measurement type
	var metricValue float64
	var returnValue string
	var errorStr string

	switch measurement.MetricType {
	case "latency":
		metricValue = float64(duration) // Duration is already in simulation time units
		returnValue = retVal.String()
	case "throughput":
		metricValue = 1.0 // Each call represents one unit of throughput
		returnValue = retVal.String()
	case "errors":
		if err != nil {
			metricValue = 1.0 // Error occurred
			errorStr = err.Error()
		} else {
			metricValue = 0.0 // No error
		}
		returnValue = retVal.String()
	default:
		// Default to latency measurement
		metricValue = float64(duration)
		returnValue = retVal.String()
	}

	if err != nil {
		errorStr = err.Error()
	}

	// Create trace point
	point := TracePoint{
		Timestamp:   timestampNs,
		Target:      measurement.Target,
		Duration:    metricValue,
		ReturnValue: returnValue,
		Error:       errorStr,
		Args:        args,
		RunID:       mt.runID,
	}

	// Store in time-series database
	if insertErr := mt.tsdb.Insert(point); insertErr != nil {
		// Log error but don't fail the simulation
		fmt.Printf("Warning: Failed to store measurement for %s: %v\n", measurement.Target, insertErr)
	}
}

// GetMetrics retrieves recent metrics for a target
func (mt *MeasurementTracer) GetMetrics(target string, since time.Time) ([]TracePoint, error) {
	return mt.tsdb.QueryLatency(target, since)
}

// GetPercentiles calculates percentiles for a target
func (mt *MeasurementTracer) GetPercentiles(target string, since time.Time) (p50, p90, p95, p99 float64, err error) {
	return mt.tsdb.QueryPercentiles(target, since)
}

// GetTimeBuckets returns time-bucketed aggregations
func (mt *MeasurementTracer) GetTimeBuckets(target string, since time.Time, bucketSize string) ([]TimeBucketResult, error) {
	return mt.tsdb.QueryTimeBuckets(target, since, bucketSize)
}

// ExecuteSQL runs a custom SQL query on the measurement data
func (mt *MeasurementTracer) ExecuteSQL(query string, args ...interface{}) ([]map[string]interface{}, error) {
	return mt.tsdb.ExecuteSQL(query, args...)
}

// GetStats returns statistics about stored measurements
func (mt *MeasurementTracer) GetStats() (map[string]interface{}, error) {
	return mt.tsdb.GetStats()
}

// Close closes the underlying time-series database
func (mt *MeasurementTracer) Close() error {
	if mt.tsdb != nil {
		return mt.tsdb.Close()
	}
	return nil
}