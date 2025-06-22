package console

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
	protos "github.com/panyam/sdl/gen/go/sdl/v1"
	"github.com/panyam/sdl/runtime"
)

// MetricType represents the type of metric being measured
type MetricType = string

const (
	// MetricCount tracks the number of events
	MetricCount MetricType = "count"

	// MetricLatency tracks duration values
	MetricLatency MetricType = "latency"
)

// MetricSpec defines what to measure and how
// The tracer will use this to collect and process events and create metric points out of them
// This will corresponding to each "LiveMetric" that can be plotted and will result in a series
// of points
type MetricSpec struct {
	*protos.Metric

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
	batchSize := 100
	batch := make([]*MetricPoint, 0, batchSize)
	batchTicker := time.NewTicker(100 * time.Millisecond) // Flush every 100ms
	defer batchTicker.Stop()

	for {
		select {
		case <-ms.stopChan:
			// Flush remaining batch
			if len(batch) > 0 && ms.store != nil {
				ms.store.WriteBatch(ctx, ms.Metric, batch)
			}
			return
		case evt := <-ms.eventChan:
			// Process the event
			if evt != nil && ms.store != nil {
				// Create metric point based on metric type
				var value float64
				if ms.MetricType == MetricLatency {
					value = float64(evt.Duration)
				} else {
					value = 1.0 // Count metric
				}

				// Convert simulation time to real time
				var timestamp time.Time
				if ms.canvas != nil && ms.canvas.simulationStarted {
					// evt.Timestamp is simulation time in seconds
					// Add it to the real start time
					timestamp = ms.canvas.simulationStartTime.Add(time.Duration(evt.Timestamp * float64(time.Second)))
				} else {
					// Fallback to current time if simulation hasn't started
					timestamp = time.Now()
				}
				point := &MetricPoint{
					Timestamp: timestamp,
					Value:     value,
					Tags:      make(map[string]string),
				}

				// Add to batch
				batch = append(batch, point)

				// Flush if batch is full
				if len(batch) >= batchSize {
					ms.store.WriteBatch(ctx, ms.Metric, batch)
					batch = batch[:0] // Reset batch
				}
			}
		case <-batchTicker.C:
			// Periodic flush
			if len(batch) > 0 && ms.store != nil {
				ms.store.WriteBatch(ctx, ms.Metric, batch)
				batch = batch[:0] // Reset batch
			}
		}
	}
}
