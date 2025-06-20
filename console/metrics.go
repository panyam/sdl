package console

import (
	"log"
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

// AggregationType defines how metrics are aggregated over time
type AggregationType string

const (
	// For count metrics
	AggSum  AggregationType = "sum"  // Total count in window
	AggRate AggregationType = "rate" // Count per second

	// For latency metrics
	AggAvg AggregationType = "avg" // Average
	AggMin AggregationType = "min" // Minimum
	AggMax AggregationType = "max" // Maximum
	AggP50 AggregationType = "p50" // 50th percentile
	AggP90 AggregationType = "p90" // 90th percentile
	AggP95 AggregationType = "p95" // 95th percentile
	AggP99 AggregationType = "p99" // 99th percentile
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
	ms.eventChan = make(chan *runtime.TraceEvent)
	ms.stopChan = make(chan bool)
	defer func() {
		close(ms.stopChan)
		close(ms.eventChan)
		ms.stopChan = nil
		ms.eventChan = nil
		ms.stopped = true
	}()

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ms.stopChan:
			return
		case evt := <-ms.eventChan:
			// Do thigns here
			log.Println("Evt: ", evt)
		case <-ticker.C:
			// Some kind of garbage collection here??
		}
	}
}

// MetricPoint represents a single metric measurement (this could be based on one or more TraceEvents)
type MetricPoint struct {
	Timestamp core.Duration `json:"timestamp"` // Virtual time
	Value     float64       `json:"value"`     // 1.0 for count, duration for latency
	// Component string        `json:"component"` // Source component
	// Method    string        `json:"method"`    // Source method
}
