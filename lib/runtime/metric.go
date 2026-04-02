package runtime

import (
	"context"
	"sync/atomic"
	"time"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/lib/core"
	"github.com/panyam/sdl/lib/decl"
)

// MetricType constants
const (
	MetricCount       = "count"
	MetricLatency     = "latency"
	MetricUtilization = "utilization"
)

// Metric represents a metric bound to a system.
// Embeds the proto Metric for transport and adds runtime collection state.
// This consolidates the old runtime.Metric + services.MetricSpec into one type.
type Metric struct {
	*protos.Metric // Embed proto (Id, Name, Component, Methods, MetricType, Aggregation, etc.)

	// Resolved references
	System                    *SystemInstance
	Matcher                   ResultMatcher
	ResolvedComponent         *ComponentInstance
	ResolvedMethod            *MethodDecl

	// Runtime collection state
	stopped   bool
	stopChan  chan bool
	eventChan chan *TraceEvent
	idCounter atomic.Int64
	store     MetricStore
	simCtx    SimulationContext
}

// NewMetricFromSpec creates a Metric from a compile-time MetricSpec.
func NewMetricFromSpec(spec *MetricSpec) *Metric {
	methods := []string{}
	if spec.MethodName != "" {
		methods = []string{spec.MethodName}
	}
	return &Metric{
		Metric: &protos.Metric{
			
			Name:              spec.Name,
			Component:         spec.ComponentPath,
			Methods:           methods,
			MetricType:        spec.MetricType,
			Aggregation:       spec.Aggregation,
			AggregationWindow: spec.Window,
			Enabled:           true,
		},
	}
}

// ProcessTraceEvent handles a trace event. Returns true if accepted.
func (m *Metric) ProcessTraceEvent(ts core.Duration, duration core.Duration, comp *ComponentInstance, method *decl.MethodDecl, retVal decl.Value, err error) bool {
	if method == nil || comp == nil {
		return false
	}

	if m.ResolvedComponent != nil && comp != m.ResolvedComponent {
		return false
	}

	methodName := method.Name.Value
	methodMatches := false
	for _, mn := range m.Methods {
		if methodName == mn {
			methodMatches = true
			break
		}
	}
	if !methodMatches {
		return false
	}

	if m.Matcher != nil && m.Matcher.Matches(retVal) {
		return false
	}

	nextId := m.idCounter.Add(1)
	event := &TraceEvent{
		Kind:      EventExit,
		ID:        nextId,
		Timestamp: ts,
		Duration:  duration,
		Component: comp,
		Method:    method,
	}

	m.eventChan <- event
	return true
}

func (m *Metric) Stop() {
	if m.stopped || m.stopChan == nil {
		return
	}
	m.stopped = true
	m.stopChan <- true
}

func (m *Metric) Start() {
	if m.stopChan != nil {
		return
	}
	m.stopped = false
	m.eventChan = make(chan *TraceEvent, 1000)
	m.stopChan = make(chan bool)
	go m.run()
}

func (m *Metric) run() {
	defer func() {
		close(m.stopChan)
		close(m.eventChan)
		m.stopChan = nil
		m.eventChan = nil
		m.stopped = true
	}()

	ctx := context.Background()

	if m.MetricType == MetricUtilization {
		m.runUtilizationCollection(ctx)
		return
	}

	window := time.Duration(m.AggregationWindow) * time.Second
	aggregationTicker := time.NewTicker(window)
	defer aggregationTicker.Stop()

	currentWindow := make([]float64, 0)
	var currentWindowStart time.Time

	for {
		select {
		case <-m.stopChan:
			if len(currentWindow) > 0 && m.store != nil {
				m.flushAggregatedWindow(ctx, currentWindow, currentWindowStart)
			}
			return
		case evt := <-m.eventChan:
			if evt != nil && m.store != nil {
				var value float64
				if m.MetricType == MetricLatency {
					value = float64(evt.Duration)
				} else {
					value = 1.0
				}

				if len(currentWindow) == 0 {
					if m.simCtx != nil && m.simCtx.IsSimulationStarted() {
						currentWindowStart = m.simCtx.GetSimulationStartTime().Add(time.Duration(evt.Timestamp * float64(time.Second)))
					} else {
						currentWindowStart = time.Now()
					}
				}

				currentWindow = append(currentWindow, value)
			}
		case <-aggregationTicker.C:
			if len(currentWindow) > 0 && m.store != nil {
				m.flushAggregatedWindow(ctx, currentWindow, currentWindowStart)
				currentWindow = currentWindow[:0]
			}
		}
	}
}

func (m *Metric) flushAggregatedWindow(ctx context.Context, values []float64, windowStart time.Time) {
	if len(values) == 0 {
		return
	}
	aggregatedValue := m.computeAggregation(values)
	point := &MetricPoint{
		Timestamp: windowStart,
		Value:     aggregatedValue,
		Tags:      make(map[string]string),
	}
	m.store.WritePoint(ctx, m.Metric, point)
}

func (m *Metric) computeAggregation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	switch m.Aggregation {
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
		sorted := make([]float64, len(values))
		copy(sorted, values)
		for i := range sorted {
			for j := 0; j < len(sorted)-1-i; j++ {
				if sorted[j] > sorted[j+1] {
					sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
				}
			}
		}
		var percentile float64
		switch m.Aggregation {
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
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum
	}
}

func (m *Metric) runUtilizationCollection(ctx context.Context) {
	interval := time.Duration(m.AggregationWindow) * time.Second
	if interval <= 0 {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.collectUtilizationMetrics(ctx)
		}
	}
}

func (m *Metric) collectUtilizationMetrics(ctx context.Context) {
	if m.ResolvedComponent == nil || m.store == nil {
		return
	}

	infos := m.ResolvedComponent.GetUtilizationInfo()

	if len(infos) > 0 {
		var utilValue float64
		found := false

		for _, info := range infos {
			if info.IsBottleneck || !found {
				utilValue = info.Utilization
				found = true
				if info.IsBottleneck {
					break
				}
			}
		}

		var timestamp time.Time
		if m.simCtx != nil && m.simCtx.IsSimulationStarted() {
			simTime := m.simCtx.GetSimulationTime()
			timestamp = m.simCtx.GetSimulationStartTime().Add(time.Duration(simTime * float64(time.Second)))
		} else {
			timestamp = time.Now()
		}

		point := &MetricPoint{
			Timestamp: timestamp,
			Value:     utilValue,
			Tags:      make(map[string]string),
		}
		m.store.WritePoint(ctx, m.Metric, point)
	}
}

// ResultMatcher determines if a return value matches the expected result
type ResultMatcher interface {
	Matches(returnValue decl.Value) bool
}

// ExactMatcher matches an exact value
type ExactMatcher struct {
	Value decl.Value
}

func (m *ExactMatcher) Matches(returnValue decl.Value) bool {
	if m.Value == decl.Nil {
		return true
	}
	return returnValue.Equals(&m.Value)
}

// NotMatcher inverts the match result
type NotMatcher struct {
	Inner ResultMatcher
}

func (m *NotMatcher) Matches(returnValue decl.Value) bool {
	return !m.Inner.Matches(returnValue)
}

// CreateResultMatcher creates a result matcher from a string specification
func CreateResultMatcher(spec string) ResultMatcher {
	if spec == "*" {
		return &ExactMatcher{Value: decl.Nil}
	}
	if len(spec) > 2 && spec[:2] == "!=" {
		return &NotMatcher{
			Inner: &ExactMatcher{Value: decl.StringValue(spec[2:])},
		}
	}
	return &ExactMatcher{Value: decl.StringValue(spec)}
}
