package services

import (
	"context"
	"time"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// MetricPoint represents a single metric measurement
type MetricPoint struct {
	// Timestamp of the measurement
	Timestamp time.Time

	// The actual value
	Value float64

	// Optional tags for additional dimensions (e.g., "cache_hit": "true")
	Tags map[string]string
}

// MetricStore defines the interface for storing and querying metrics
type MetricStore interface {
	// WritePoint stores a single metric point for a specific metric
	WritePoint(ctx context.Context, metric *protos.Metric, point *MetricPoint) error

	// WriteBatch stores multiple metric points for a specific metric
	WriteBatch(ctx context.Context, metric *protos.Metric, points []*MetricPoint) error

	// Query retrieves raw metric points for a specific metric
	Query(ctx context.Context, metric *protos.Metric, opts QueryOptions) (QueryResult, error)

	// QueryMultiple retrieves points for multiple metrics (e.g., for correlation)
	QueryMultiple(ctx context.Context, metrics []*protos.Metric, opts QueryOptions) (map[string]QueryResult, error)

	// Aggregate computes aggregations for a specific metric
	Aggregate(ctx context.Context, metric *protos.Metric, opts AggregateOptions) (AggregateResult, error)

	// Subscribe creates a subscription for real-time metric updates
	Subscribe(ctx context.Context, metricIDs []string) (<-chan *MetricUpdateBatch, error)

	// GetMetricStats returns statistics for a metric
	GetMetricStats(metric *protos.Metric) MetricStats

	// Close cleanly shuts down the store
	Close() error
}

// MetricUpdateBatch contains metric updates that can be sent to subscribers
type MetricUpdateBatch struct {
	Updates []*MetricUpdateItem
}

// MetricUpdateItem represents a single metric update
type MetricUpdateItem struct {
	MetricID string
	Point    *MetricPoint
}

// MetricStats contains summary statistics for a metric
type MetricStats struct {
	TotalPoints     int64
	OldestTimestamp float64
	NewestTimestamp float64
}

// QueryOptions specifies parameters for metric queries
type QueryOptions struct {
	// Time range
	StartTime time.Time
	EndTime   time.Time

	// Additional tag filters
	TagFilters map[string]string

	// Maximum number of results
	Limit int

	// Offset for pagination
	Offset int
}

// QueryResult contains the results of a metric query
type QueryResult struct {
	Points    []*MetricPoint
	TotalRows int64
	HasMore   bool
}

// AggregateOptions specifies parameters for aggregations
type AggregateOptions struct {
	// Time range
	StartTime time.Time
	EndTime   time.Time

	// Aggregation window (e.g., 1s, 5s, 1m)
	Window time.Duration

	// Additional tag filters
	TagFilters map[string]string

	// Aggregation functions to compute
	Functions []AggregateFunc
}

// AggregateFunc represents an aggregation function
type AggregateFunc string

const (
	AggCount  AggregateFunc = "count"
	AggSum    AggregateFunc = "sum"
	AggAvg    AggregateFunc = "avg"
	AggMin    AggregateFunc = "min"
	AggMax    AggregateFunc = "max"
	AggP50    AggregateFunc = "p50"
	AggP90    AggregateFunc = "p90"
	AggP95    AggregateFunc = "p95"
	AggP99    AggregateFunc = "p99"
	AggRate   AggregateFunc = "rate"   // points per second
	AggStdDev AggregateFunc = "stddev" // standard deviation
)

// AggregateResult contains time-series aggregation results
type AggregateResult struct {
	// Time buckets
	Buckets []TimeBucket

	// Reference to the metric this result is for
	Metric *protos.Metric

	// Aggregation window used
	Window time.Duration
}

// TimeBucket represents aggregated metrics for a time window
type TimeBucket struct {
	// Start time of this bucket
	Time time.Time

	// Aggregated values keyed by function name
	Values map[AggregateFunc]float64

	// Number of points in this bucket
	Count int64
}

// MetricStoreFactory creates MetricStore instances
type MetricStoreFactory interface {
	// CreateStore creates a new metric store with the given configuration
	CreateStore(config MetricStoreConfig) (MetricStore, error)
}

// MetricStoreConfig holds configuration for a metric store
type MetricStoreConfig struct {
	// Type of store (e.g., "ringbuffer", "timescaledb", "influxdb")
	Type string

	// Store-specific configuration
	Config map[string]interface{}
}

// Common configuration keys
const (
	// RingBuffer configuration
	ConfigRingBufferSize     = "size"     // number of points per metric
	ConfigRingBufferDuration = "duration" // time window to retain

	// Database configuration
	ConfigDBConnection = "connection" // connection string
	ConfigDBRetention  = "retention"  // data retention period
	ConfigDBBatchSize  = "batch_size" // write batch size
)
