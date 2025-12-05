package services

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/panyam/sdl/lib/core"
	"github.com/panyam/sdl/lib/decl"
	"github.com/panyam/sdl/lib/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The main tracer that processes TraceEvents and generates metrics from them
type MetricTracer struct {
	seriesLock sync.RWMutex
	seriesMap  map[string]*MetricSpec
	system     *runtime.SystemInstance
	store      MetricStore
	canvas     *Canvas // Reference to parent canvas for simulation time
}

func NewMetricTracer(system *runtime.SystemInstance, canvas *Canvas) *MetricTracer {
	// Create default ring buffer store
	store, _ := NewRingBufferStore(MetricStoreConfig{
		Type: "ringbuffer",
		Config: map[string]interface{}{
			ConfigRingBufferSize:     10000,
			ConfigRingBufferDuration: 5 * time.Minute,
		},
	})

	return &MetricTracer{
		seriesMap: map[string]*MetricSpec{},
		system:    system,
		store:     store,
		canvas:    canvas,
	}
}

// SetMetricStore sets a custom metric store
func (mt *MetricTracer) SetMetricStore(store MetricStore) {
	mt.seriesLock.Lock()
	defer mt.seriesLock.Unlock()

	// Close old store if it exists
	if mt.store != nil {
		mt.store.Close()
	}
	mt.store = store
}

func (mt *MetricTracer) ListMetricSpec() []*MetricSpec {
	mt.seriesLock.RLock()
	defer mt.seriesLock.RUnlock()
	return slices.Collect(maps.Values(mt.seriesMap))
}

func (mt *MetricTracer) GetMetricSpec(specId string) (spec *MetricSpec) {
	mt.seriesLock.RLock()
	defer mt.seriesLock.RUnlock()
	return mt.seriesMap[specId]
}

func (mt *MetricTracer) AddMetricSpec(spec *MetricSpec) error {
	mt.seriesLock.Lock()
	defer mt.seriesLock.Unlock()

	if spec.ID == "" {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("measurement ID cannot be empty"))
	}

	if mt.seriesMap[spec.ID] != nil {
		return status.Error(codes.AlreadyExists, fmt.Sprintf("Metric already exists"))
	}

	if spec.Component == "" {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("component cannot be empty"))
	}

	// For utilization metrics, methods are optional
	if spec.MetricType != MetricUtilization && len(spec.Methods) == 0 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("at least one method must be specified for %s metrics", spec.MetricType))
	}

	if spec.MetricType != MetricCount && spec.MetricType != MetricLatency && spec.MetricType != MetricUtilization {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid metric type: %s", spec.MetricType))
	}

	// Set the system reference
	spec.System = mt.system

	// Resolve the component instance from the system
	if mt.system == nil || mt.system.Env == nil {
		return status.Error(codes.FailedPrecondition, "system or its env not defined")
	}
	resolvedComponent := mt.system.FindComponent(spec.Component)
	if resolvedComponent == nil {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("component '%s' not found in system", spec.Component))
	}

	if spec.Metric.AggregationWindow < 0 {
		spec.Metric.AggregationWindow = 10
	}

	// Create the measurement
	spec.resolvedComponentInstance = resolvedComponent
	spec.store = mt.store
	spec.canvas = mt.canvas
	mt.seriesMap[spec.ID] = spec
	spec.Start()
	return nil
}

func (mt *MetricTracer) Clear() {
	mt.seriesLock.Lock()
	defer mt.seriesLock.Unlock()
	for _, ms := range mt.seriesMap {
		ms.Stop()
	}
	mt.seriesMap = map[string]*MetricSpec{}
}

func (mt *MetricTracer) RemoveMetricSpec(specId string) {
	mt.seriesLock.Lock()
	defer mt.seriesLock.Unlock()
	if spec, ok := mt.seriesMap[specId]; ok && spec != nil {
		spec.Stop()
		delete(mt.seriesMap, specId)
	}
}

// ListMetrics returns all configured metrics with statistics
func (mt *MetricTracer) ListMetrics() []*Metric {
	mt.seriesLock.RLock()
	defer mt.seriesLock.RUnlock()

	metrics := make([]*Metric, 0, len(mt.seriesMap))
	for _, spec := range mt.seriesMap {
		if spec.Metric != nil {
			// Create a copy of the metric with updated statistics
			metricCopy := &Metric{
				ID:                spec.Metric.ID,
				CanvasID:          spec.Metric.CanvasID,
				Name:              spec.Metric.Name,
				Component:         spec.Metric.Component,
				Methods:           spec.Metric.Methods,
				Enabled:           spec.Metric.Enabled,
				MetricType:        spec.Metric.MetricType,
				Aggregation:       spec.Metric.Aggregation,
				AggregationWindow: spec.Metric.AggregationWindow,
				MatchResult:       spec.Metric.MatchResult,
				MatchResultType:   spec.Metric.MatchResultType,
			}

			// Get statistics from the store
			if mt.store != nil {
				stats := mt.store.GetMetricStats(spec.Metric)
				metricCopy.NumDataPoints = stats.TotalPoints
				metricCopy.OldestTimestamp = stats.OldestTimestamp
				metricCopy.NewestTimestamp = stats.NewestTimestamp
			}

			metrics = append(metrics, metricCopy)
		}
	}
	return metrics
}

// GetMetricByID finds a metric by its ID
func (mt *MetricTracer) GetMetricByID(id string) *Metric {
	mt.seriesLock.RLock()
	defer mt.seriesLock.RUnlock()

	if spec, ok := mt.seriesMap[id]; ok && spec.Metric != nil {
		return spec.Metric
	}
	return nil
}

// Main Tracer interface methods
func (mt *MetricTracer) Exit(ts core.Duration, duration core.Duration, comp *runtime.ComponentInstance, method *decl.MethodDecl, retVal decl.Value, err error) {
	// Only care about events on exit
	mt.seriesLock.RLock()
	defer mt.seriesLock.RUnlock()

	// Find matching measurements
	for _, m := range mt.seriesMap {
		m.ProcessTraceEvent(ts, duration, comp, method, retVal, err)
	}
}

func (mt *MetricTracer) Enter(ts core.Duration, kind runtime.TraceEventKind, comp *runtime.ComponentInstance, method *decl.MethodDecl, args ...string) int64 {
	return 0
}

// PushParentID manually pushes a parent ID onto the stack.
// Used by the aggregator to set the context for evaluating futures.
func (mt *MetricTracer) PushParentID(id int64) {}

// PopParent removes the most recent event ID from the stack.
func (mt *MetricTracer) PopParent() {}

// GetMetricStore returns the current metric store
func (mt *MetricTracer) GetMetricStore() MetricStore {
	mt.seriesLock.RLock()
	defer mt.seriesLock.RUnlock()
	return mt.store
}

// QueryMetrics queries metrics from the store
func (mt *MetricTracer) QueryMetrics(ctx context.Context, specId string, opts QueryOptions) (QueryResult, error) {
	mt.seriesLock.RLock()
	spec := mt.seriesMap[specId]
	store := mt.store
	mt.seriesLock.RUnlock()

	if spec == nil {
		return QueryResult{}, fmt.Errorf("metric spec %s not found", specId)
	}

	if store == nil {
		return QueryResult{}, fmt.Errorf("no metric store configured")
	}

	return store.Query(ctx, spec.Metric, opts)
}

// AggregateMetrics computes aggregations for a metric
func (mt *MetricTracer) AggregateMetrics(ctx context.Context, specId string, opts AggregateOptions) (AggregateResult, error) {
	mt.seriesLock.RLock()
	spec := mt.seriesMap[specId]
	store := mt.store
	mt.seriesLock.RUnlock()

	if spec == nil {
		return AggregateResult{}, fmt.Errorf("metric spec %s not found", specId)
	}

	if store == nil {
		return AggregateResult{}, fmt.Errorf("no metric store configured")
	}

	return store.Aggregate(ctx, spec.Metric, opts)
}

// ResultMatcher determines if a return value matches the expected result
type ResultMatcher interface {
	Matches(returnValue decl.Value) bool
}

// ExactMatcher matches an exact string value
type ExactMatcher struct {
	Value decl.Value
}

// Matches returns true if the return value exactly matches the expected value
func (m *ExactMatcher) Matches(returnValue decl.Value) bool {
	// Special case: "*" matches everything
	if m.Value == decl.Nil {
		return true
	}
	return returnValue.Equals(&m.Value)
}

// NotMatcher inverts the match result
type NotMatcher struct {
	Inner ResultMatcher
}

// Matches returns true if the inner matcher returns false
func (m *NotMatcher) Matches(returnValue decl.Value) bool {
	return !m.Inner.Matches(returnValue)
}

// Helper function to create a result matcher from a string specification
func CreateResultMatcher(spec string) ResultMatcher {
	// Handle special cases
	if spec == "*" {
		return &ExactMatcher{Value: decl.Nil}
	}

	// Handle "!=" prefix for not-equal
	if len(spec) > 2 && spec[:2] == "!=" {
		return &NotMatcher{
			Inner: &ExactMatcher{Value: decl.StringValue(spec[2:])},
		}
	}

	// Default to exact match
	return &ExactMatcher{Value: decl.StringValue(spec)}
}
