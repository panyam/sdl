package services

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/panyam/sdl/lib/core"
	"github.com/panyam/sdl/lib/decl"
	"github.com/panyam/sdl/lib/runtime"
)

// MetricType represents the type of metric being measured
type MetricType = string

const (
	// MetricCount tracks the number of events
	MetricCount MetricType = "count"

	// MetricLatency tracks duration values
	MetricLatency MetricType = "latency"

	// MetricUtilization tracks resource utilization (0.0 to 1.0)
	MetricUtilization MetricType = "utilization"
)

// MetricSpec defines what to measure and how
// The tracer will use this to collect and process events and create metric points out of them
// This will corresponding to each "LiveMetric" that can be plotted and will result in a series
// of points
type MetricSpec struct {
	*Metric // Use native type instead of proto

	// The system this metric spec applies to
	System                    *runtime.SystemInstance
	Matcher                   ResultMatcher
	resolvedComponentInstance *runtime.ComponentInstance // Resolved from Component name
	stopped                   bool
	stopChan                  chan bool
	eventChan                 chan *runtime.TraceEvent
	idCounter                 atomic.Int64
	store                     MetricStore // Reference to metric store
	canvas                    *Canvas     // Reference to canvas for simulation time
}

// Handles the next trace event.  Returns true if event accepted, false otherwise
func (ms *MetricSpec) ProcessTraceEvent(ts core.Duration, duration core.Duration, comp *runtime.ComponentInstance, method *decl.MethodDecl, retVal decl.Value, err error) bool {
	// Only process exit events with method calls
	if method == nil || comp == nil {
		return false
	}

	// If we have a resolved component instance, use pointer comparison
	if ms.resolvedComponentInstance != nil && comp != ms.resolvedComponentInstance {
		return false
	}

	// Check method match
	methodName := method.Name.Value
	methodMatches := false
	for _, method := range ms.Methods {
		if methodName == method {
			methodMatches = true
			break
		}
	}
	if !methodMatches {
		return false
	}

	// TODO - Check result value match
	if ms.Matcher != nil && ms.Matcher.Matches(retVal) {
		return false
	}

	nextId := ms.idCounter.Add(1)
	event := &runtime.TraceEvent{
		Kind:      runtime.EventExit,
		ID:        nextId,
		Timestamp: ts,
		Duration:  duration,
		Component: comp,
		Method:    method,
	}

	ms.eventChan <- event
	return true
}

func (ms *MetricSpec) Stop() {
	if ms.stopped || ms.stopChan == nil {
		return
	}
	ms.stopped = true
	ms.stopChan <- true
}

func (ms *MetricSpec) Start() {
	if ms.stopChan != nil {
		return
	}
	ms.stopped = false
	ms.eventChan = make(chan *runtime.TraceEvent, 1000) // Buffered channel
	ms.stopChan = make(chan bool)

	// Run processing in background
	go ms.run()
}

func (ms *MetricSpec) run() {
	defer func() {
		close(ms.stopChan)
		close(ms.eventChan)
		ms.stopChan = nil
		ms.eventChan = nil
		ms.stopped = true
	}()

	ctx := context.Background()

	// For utilization metrics, use periodic sampling
	if ms.MetricType == MetricUtilization {
		ms.runUtilizationCollection(ctx)
		return
	}

	// Time window aggregation for event-based metrics
	window := time.Duration(ms.AggregationWindow) * time.Second
	aggregationTicker := time.NewTicker(window)
	defer aggregationTicker.Stop()

	// Collect events within time windows for pre-aggregation
	currentWindow := make([]float64, 0)
	var currentWindowStart time.Time

	for {
		select {
		case <-ms.stopChan:
			// Flush current window if it has data
			if len(currentWindow) > 0 && ms.store != nil {
				ms.flushAggregatedWindow(ctx, currentWindow, currentWindowStart)
			}
			return
		case evt := <-ms.eventChan:
			// Process the event
			if evt != nil && ms.store != nil {
				// Extract value based on metric type
				var value float64
				if ms.MetricType == MetricLatency {
					value = float64(evt.Duration)
				} else {
					value = 1.0 // Count metric - each event counts as 1
				}

				// Set window start time if this is the first event
				if len(currentWindow) == 0 {
					if ms.canvas != nil && ms.canvas.simulationStarted {
						currentWindowStart = ms.canvas.simulationStartTime.Add(time.Duration(evt.Timestamp * float64(time.Second)))
					} else {
						currentWindowStart = time.Now()
					}
				}

				// Add value to current window
				currentWindow = append(currentWindow, value)
			}
		case <-aggregationTicker.C:
			// Time to aggregate and flush current window
			if len(currentWindow) > 0 && ms.store != nil {
				ms.flushAggregatedWindow(ctx, currentWindow, currentWindowStart)
				currentWindow = currentWindow[:0] // Reset window
			}
		}
	}
}

// flushAggregatedWindow computes aggregation over collected values and writes to store
func (ms *MetricSpec) flushAggregatedWindow(ctx context.Context, values []float64, windowStart time.Time) {
	if len(values) == 0 {
		return
	}

	// Compute aggregated value based on aggregation function
	aggregatedValue := ms.computeAggregation(values)

	// Create aggregated metric point
	point := &MetricPoint{
		Timestamp: windowStart,
		Value:     aggregatedValue,
		Tags:      make(map[string]string),
	}

	// Write single aggregated point
	ms.store.WritePoint(ctx, ms.Metric, point)
}

// computeAggregation applies the configured aggregation function to values
func (ms *MetricSpec) computeAggregation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	switch ms.Aggregation {
	case "sum":
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum
	case "avg":
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum / float64(len(values))
	case "min":
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min
	case "max":
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max
	case "count":
		return float64(len(values))
	case "p50", "p90", "p95", "p99":
		// Sort values for percentile calculation
		sorted := make([]float64, len(values))
		copy(sorted, values)
		// Simple bubble sort for small arrays
		for i := 0; i < len(sorted); i++ {
			for j := 0; j < len(sorted)-1-i; j++ {
				if sorted[j] > sorted[j+1] {
					sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
				}
			}
		}

		var percentile float64
		switch ms.Aggregation {
		case "p50":
			percentile = 0.50
		case "p90":
			percentile = 0.90
		case "p95":
			percentile = 0.95
		case "p99":
			percentile = 0.99
		}

		idx := int(float64(len(sorted)-1) * percentile)
		return sorted[idx]
	default:
		// Default to sum if aggregation type is unknown
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum
	}
}

// runUtilizationCollection periodically samples utilization from components
func (ms *MetricSpec) runUtilizationCollection(ctx context.Context) {
	// Use aggregation window for sampling interval, default to 10s
	interval := time.Duration(ms.AggregationWindow) * time.Second
	if interval <= 0 {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ms.stopChan:
			return
		case <-ticker.C:
			ms.collectUtilizationMetrics(ctx)
		}
	}
}

// collectUtilizationMetrics samples current utilization and writes to store
func (ms *MetricSpec) collectUtilizationMetrics(ctx context.Context) {
	if ms.resolvedComponentInstance == nil || ms.store == nil {
		return
	}

	// Get utilization info from component
	infos := ms.resolvedComponentInstance.GetUtilizationInfo()

	// For now, collect the first/primary utilization value
	// In future, we might want to support multiple utilization metrics per component
	if len(infos) > 0 {
		// Find the bottleneck or first utilization
		var utilValue float64
		found := false

		for _, info := range infos {
			if info.IsBottleneck || !found {
				utilValue = info.Utilization
				found = true
				if info.IsBottleneck {
					break // Prefer bottleneck utilization
				}
			}
		}

		// Create metric point
		var timestamp time.Time
		if ms.canvas != nil && ms.canvas.simulationStarted {
			// Use simulation time if available
			simTime := ms.canvas.GetSimulationTime()
			timestamp = ms.canvas.simulationStartTime.Add(time.Duration(simTime * float64(time.Second)))
		} else {
			timestamp = time.Now()
		}

		point := &MetricPoint{
			Timestamp: timestamp,
			Value:     utilValue,
			Tags:      make(map[string]string),
		}

		// Write to store
		ms.store.WritePoint(ctx, ms.Metric, point)
	}
}
