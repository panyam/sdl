package console

/*
// MetricStore manages all measurements and processes trace events
type MetricStore struct {
	mu           sync.RWMutex
	measurements map[string]*Measurement
	bufferSize   int
	system       *runtime.SystemInstance
}

// Measurement represents an active measurement with its data buffer
type Measurement struct {
	Spec       MeasurementSpec
	Buffer     *CircularBuffer
	Matcher    ResultMatcher
	mu         sync.RWMutex
	lastUpdate time.Time
}

// NewMetricStore creates a new metric store with specified buffer size
func NewMetricStore(bufferSize int) *MetricStore {
	return &MetricStore{
		measurements: make(map[string]*Measurement),
		bufferSize:   bufferSize,
	}
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
		ResultValue: m.Spec.ResultValue,
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
*/
