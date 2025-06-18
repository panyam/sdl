package runtime

import (
	"fmt"
	"sync"
	"time"
)

// MetricStore manages all measurements and processes trace events
type MetricStore struct {
	mu           sync.RWMutex
	measurements map[string]*Measurement
	bufferSize   int
	system       *SystemInstance
}

// Measurement represents an active measurement with its data buffer
type Measurement struct {
	Spec                    MeasurementSpec
	Buffer                  *CircularBuffer
	Matcher                 ResultMatcher
	ResolvedComponentInstance *ComponentInstance // Resolved from Component name
	mu                      sync.RWMutex
	lastUpdate              time.Time
}

// NewMetricStore creates a new metric store with specified buffer size
func NewMetricStore(bufferSize int) *MetricStore {
	return &MetricStore{
		measurements: make(map[string]*Measurement),
		bufferSize:   bufferSize,
	}
}

// SetSystem sets the system instance for component resolution
func (ms *MetricStore) SetSystem(system *SystemInstance) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.system = system
}

// AddMeasurement registers a new measurement specification
func (ms *MetricStore) AddMeasurement(spec MeasurementSpec) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if spec.ID == "" {
		return fmt.Errorf("measurement ID cannot be empty")
	}

	if spec.Component == "" {
		return fmt.Errorf("component cannot be empty")
	}

	if len(spec.Methods) == 0 {
		return fmt.Errorf("at least one method must be specified")
	}

	if spec.Metric != MetricCount && spec.Metric != MetricLatency {
		return fmt.Errorf("invalid metric type: %s", spec.Metric)
	}

	// Resolve the component instance from the system
	var resolvedComponent *ComponentInstance
	if ms.system != nil && ms.system.Env != nil {
		if val, ok := ms.system.Env.Get(spec.Component); ok {
			if comp, ok := val.Value.(*ComponentInstance); ok {
				resolvedComponent = comp
			} else {
				return fmt.Errorf("'%s' is not a component instance", spec.Component)
			}
		} else {
			return fmt.Errorf("component '%s' not found in system", spec.Component)
		}
	}

	// Create the measurement
	measurement := &Measurement{
		Spec:                      spec,
		Buffer:                    NewCircularBuffer(ms.bufferSize),
		Matcher:                   CreateResultMatcher(spec.ResultValue),
		ResolvedComponentInstance: resolvedComponent,
	}

	ms.measurements[spec.ID] = measurement
	return nil
}

// RemoveMeasurement removes a measurement by ID
func (ms *MetricStore) RemoveMeasurement(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.measurements[id]; !exists {
		return fmt.Errorf("measurement %s not found", id)
	}

	delete(ms.measurements, id)
	return nil
}

// ProcessTraceEvent processes a trace event and updates relevant measurements
func (ms *MetricStore) ProcessTraceEvent(event *TraceEvent) {
	ms.mu.RLock()
	measurements := make([]*Measurement, 0)
	
	// Find matching measurements
	for _, m := range ms.measurements {
		if ms.eventMatchesMeasurement(event, m) {
			measurements = append(measurements, m)
		}
	}
	ms.mu.RUnlock()

	// Update matching measurements
	for _, m := range measurements {
		ms.updateMeasurement(m, event)
	}
}

// eventMatchesMeasurement checks if an event matches a measurement
func (ms *MetricStore) eventMatchesMeasurement(event *TraceEvent, measurement *Measurement) bool {
	// Only process exit events with method calls
	if event.Kind != EventExit || event.Method == nil {
		return false
	}

	// Check component match using pointer comparison
	if measurement.ResolvedComponentInstance != nil && event.Component != measurement.ResolvedComponentInstance {
		return false
	}

	// Check method match
	methodName := event.GetMethodName()
	methodMatches := false
	for _, method := range measurement.Spec.Methods {
		if methodName == method {
			methodMatches = true
			break
		}
	}
	if !methodMatches {
		return false
	}

	// Check result value match
	if !measurement.Matcher.Matches(event.ReturnValue) {
		return false
	}

	return true
}

// updateMeasurement adds a metric point for the given event
func (ms *MetricStore) updateMeasurement(m *Measurement, event *TraceEvent) {
	var value float64
	
	switch m.Spec.Metric {
	case MetricCount:
		value = 1.0
	case MetricLatency:
		value = float64(event.Duration)
	}

	point := MetricPoint{
		Timestamp: event.Timestamp,
		Value:     value,
		Component: event.GetComponentName(),
		Method:    event.GetMethodName(),
	}

	m.mu.Lock()
	m.Buffer.Add(point)
	m.lastUpdate = time.Now()
	m.mu.Unlock()
}

// GetMeasurement returns the measurement spec for a given ID
func (ms *MetricStore) GetMeasurement(id string) (*MeasurementSpec, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	m, exists := ms.measurements[id]
	if !exists {
		return nil, fmt.Errorf("measurement %s not found", id)
	}

	return &m.Spec, nil
}

// GetMeasurementInfo returns information about a measurement
func (ms *MetricStore) GetMeasurementInfo(id string) (*MeasurementInfo, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	m, exists := ms.measurements[id]
	if !exists {
		return nil, fmt.Errorf("measurement %s not found", id)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return &MeasurementInfo{
		ID:          m.Spec.ID,
		Name:        m.Spec.Name,
		Component:   m.Spec.Component,
		Methods:     m.Spec.Methods,
		Metric:      m.Spec.Metric,
		Aggregation: m.Spec.Aggregation,
		Window:      m.Spec.Window,
		PointCount:  m.Buffer.Size(),
		LastUpdate:  m.lastUpdate,
	}, nil
}

// ListMeasurements returns all measurement IDs
func (ms *MetricStore) ListMeasurements() []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ids := make([]string, 0, len(ms.measurements))
	for id := range ms.measurements {
		ids = append(ids, id)
	}
	return ids
}

// GetMeasurementData returns the latest n data points for a measurement
func (ms *MetricStore) GetMeasurementData(id string, n int) ([]MetricPoint, error) {
	ms.mu.RLock()
	m, exists := ms.measurements[id]
	if !exists {
		ms.mu.RUnlock()
		return nil, fmt.Errorf("measurement %s not found", id)
	}
	ms.mu.RUnlock()

	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.Buffer.GetLatest(n), nil
}

// GetMeasurementDataInWindow returns data points within the specified time window
func (ms *MetricStore) GetMeasurementDataInWindow(id string, window time.Duration) ([]MetricPoint, error) {
	ms.mu.RLock()
	m, exists := ms.measurements[id]
	if !exists {
		ms.mu.RUnlock()
		return nil, fmt.Errorf("measurement %s not found", id)
	}
	ms.mu.RUnlock()

	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.Buffer.GetInWindow(window), nil
}

// Clear removes all data from all measurements
func (ms *MetricStore) Clear() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, m := range ms.measurements {
		m.mu.Lock()
		m.Buffer.Clear()
		m.mu.Unlock()
	}
}