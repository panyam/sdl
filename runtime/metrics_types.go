package runtime

import (
	"time"

	"github.com/panyam/sdl/core"
)

// MetricType represents the type of metric being measured
type MetricType string

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

// MeasurementSpec defines what to measure and how
type MeasurementSpec struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Component   string          `json:"component"`   // Component instance name
	Methods     []string        `json:"methods"`     // Methods to track
	ResultValue string          `json:"resultValue"` // Expected result value (* for all)
	Metric      MetricType      `json:"metric"`      // count or latency
	Aggregation AggregationType `json:"aggregation"` // How to aggregate
	Window      time.Duration   `json:"window"`      // Time window for aggregation
}

// MetricPoint represents a single metric measurement
type MetricPoint struct {
	Timestamp core.Duration `json:"timestamp"` // Virtual time
	Value     float64       `json:"value"`     // 1.0 for count, duration for latency
	Component string        `json:"component"` // Source component
	Method    string        `json:"method"`    // Source method
}

// AggregatedMetric contains aggregated metric data
type AggregatedMetric struct {
	MeasurementID string          `json:"measurementId"`
	Window        time.Duration   `json:"window"`
	Aggregation   AggregationType `json:"aggregation"`
	Value         float64         `json:"value"`
	Count         int             `json:"count"`
	StartTime     time.Time       `json:"startTime"`
	EndTime       time.Time       `json:"endTime"`
}

// MeasurementInfo provides summary information about a measurement
type MeasurementInfo struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Component   string          `json:"component"`
	Methods     []string        `json:"methods"`
	Metric      MetricType      `json:"metric"`
	Aggregation AggregationType `json:"aggregation"`
	Window      time.Duration   `json:"window"`
	PointCount  int             `json:"pointCount"`
	LastUpdate  time.Time       `json:"lastUpdate"`
}

// ResultMatcher determines if a return value matches the expected result
type ResultMatcher interface {
	Matches(returnValue string) bool
}

// ExactMatcher matches an exact string value
type ExactMatcher struct {
	Value string
}

// Matches returns true if the return value exactly matches the expected value
func (m *ExactMatcher) Matches(returnValue string) bool {
	// Special case: "*" matches everything
	if m.Value == "*" {
		return true
	}
	return returnValue == m.Value
}

// NotMatcher inverts the match result
type NotMatcher struct {
	Inner ResultMatcher
}

// Matches returns true if the inner matcher returns false
func (m *NotMatcher) Matches(returnValue string) bool {
	return !m.Inner.Matches(returnValue)
}

// Helper function to create a result matcher from a string specification
func CreateResultMatcher(spec string) ResultMatcher {
	// Handle special cases
	if spec == "*" {
		return &ExactMatcher{Value: "*"}
	}

	// Handle "!=" prefix for not-equal
	if len(spec) > 2 && spec[:2] == "!=" {
		return &NotMatcher{
			Inner: &ExactMatcher{Value: spec[2:]},
		}
	}

	// Default to exact match
	return &ExactMatcher{Value: spec}
}
