package console

import (
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The main tracer that processes TraceEvents and generates metrics from them
type MetricTracer struct {
	seriesLock sync.RWMutex
	seriesMap  map[string]*MetricSpec
	system     *runtime.SystemInstance
}

func NewMetricTracer(system *runtime.SystemInstance) *MetricTracer {
	return &MetricTracer{
		seriesMap: map[string]*MetricSpec{},
		system:    system,
	}
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

	if spec.Id == "" {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("measurement ID cannot be empty"))
	}

	if mt.seriesMap[spec.Id] != nil {
		return status.Error(codes.AlreadyExists, fmt.Sprintf("Metric already exists"))
	}

	if spec.Component == "" {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("component cannot be empty"))
	}

	if len(spec.Methods) == 0 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("at least one method must be specified"))
	}

	if spec.MetricType != MetricCount && spec.MetricType != MetricLatency {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid metric type: %s", spec.Metric))
	}

	// Resolve the component instance from the system
	var resolvedComponent *runtime.ComponentInstance
	if spec.System != nil && spec.System.Env != nil {
		if val, ok := spec.System.Env.Get(spec.Component); ok {
			if comp, ok := val.Value.(*runtime.ComponentInstance); ok {
				resolvedComponent = comp
			} else {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("'%s' is not a component instance", spec.Component))
			}
		} else {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("component '%s' not found in system", spec.Component))
		}
	}

	if spec.Metric.AggregationWindow < 0 {
		spec.Metric.AggregationWindow = 10
	}

	// Create the measurement
	spec.resolvedComponentInstance = resolvedComponent
	mt.seriesMap[spec.Id] = spec
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
	if mt.seriesMap[specId] != nil {
		mt.seriesMap[specId].Stop()
		mt.seriesMap[specId] = nil
	}
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

func (mt *MetricTracer) Enter(ts core.Duration, kind runtime.TraceEventKind, comp *runtime.ComponentInstance, method *decl.MethodDecl, args ...string) int {
	return 0
}

// PushParentID manually pushes a parent ID onto the stack.
// Used by the aggregator to set the context for evaluating futures.
func (mt *MetricTracer) PushParentID(id int) {}

// PopParent removes the most recent event ID from the stack.
func (mt *MetricTracer) PopParent() {}

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
