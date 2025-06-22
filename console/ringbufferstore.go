package console

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	protos "github.com/panyam/sdl/gen/go/sdl/v1"
)

// RingBufferStore implements MetricStore using in-memory ring buffers
type RingBufferStore struct {
	// Configuration
	maxPointsPerMetric int
	maxDuration        time.Duration

	// Ring buffers per metric ID
	buffers map[string]*ringBuffer
	mu      sync.RWMutex

	// Closed flag
	closed bool
}

// ringBuffer holds metric points in a circular buffer
type ringBuffer struct {
	points    []*MetricPoint
	size      int
	writePos  int
	readStart int
	count     int
	mu        sync.RWMutex
}

// NewRingBufferStore creates a new ring buffer metric store
func NewRingBufferStore(config MetricStoreConfig) (*RingBufferStore, error) {
	size := 10000 // default
	if v, ok := config.Config[ConfigRingBufferSize]; ok {
		if s, ok := v.(int); ok {
			size = s
		}
	}

	duration := 5 * time.Minute // default
	if v, ok := config.Config[ConfigRingBufferDuration]; ok {
		if d, ok := v.(time.Duration); ok {
			duration = d
		}
	}

	return &RingBufferStore{
		maxPointsPerMetric: size,
		maxDuration:        duration,
		buffers:            make(map[string]*ringBuffer),
	}, nil
}

// WritePoint stores a single metric point
func (s *RingBufferStore) WritePoint(ctx context.Context, metric *protos.Metric, point *MetricPoint) error {
	if s.closed {
		return fmt.Errorf("store is closed")
	}

	s.mu.Lock()
	rb, ok := s.buffers[metric.Id]
	if !ok {
		rb = newRingBuffer(s.maxPointsPerMetric)
		s.buffers[metric.Id] = rb
	}
	s.mu.Unlock()

	rb.add(point)
	return nil
}

// WriteBatch stores multiple metric points efficiently
func (s *RingBufferStore) WriteBatch(ctx context.Context, metric *protos.Metric, points []*MetricPoint) error {
	if s.closed {
		return fmt.Errorf("store is closed")
	}

	s.mu.Lock()
	rb, ok := s.buffers[metric.Id]
	if !ok {
		rb = newRingBuffer(s.maxPointsPerMetric)
		s.buffers[metric.Id] = rb
	}
	s.mu.Unlock()

	for _, point := range points {
		rb.add(point)
	}
	return nil
}

// Query retrieves raw metric points
func (s *RingBufferStore) Query(ctx context.Context, metric *protos.Metric, opts QueryOptions) (QueryResult, error) {
	if s.closed {
		return QueryResult{}, fmt.Errorf("store is closed")
	}

	s.mu.RLock()
	rb, ok := s.buffers[metric.Id]
	s.mu.RUnlock()

	if !ok {
		return QueryResult{Points: []*MetricPoint{}}, nil
	}

	points := rb.query(opts.StartTime, opts.EndTime, opts.TagFilters)
	
	// Apply limit and offset
	total := len(points)
	if opts.Offset >= total {
		return QueryResult{Points: []*MetricPoint{}, TotalRows: int64(total)}, nil
	}

	points = points[opts.Offset:]
	hasMore := false
	if opts.Limit > 0 && len(points) > opts.Limit {
		points = points[:opts.Limit]
		hasMore = true
	}

	return QueryResult{
		Points:    points,
		TotalRows: int64(total),
		HasMore:   hasMore,
	}, nil
}

// QueryMultiple retrieves points for multiple metrics
func (s *RingBufferStore) QueryMultiple(ctx context.Context, metrics []*protos.Metric, opts QueryOptions) (map[string]QueryResult, error) {
	results := make(map[string]QueryResult)
	for _, metric := range metrics {
		result, err := s.Query(ctx, metric, opts)
		if err != nil {
			return nil, err
		}
		results[metric.Id] = result
	}
	return results, nil
}

// Aggregate computes aggregations over time windows
func (s *RingBufferStore) Aggregate(ctx context.Context, metric *protos.Metric, opts AggregateOptions) (AggregateResult, error) {
	if s.closed {
		return AggregateResult{}, fmt.Errorf("store is closed")
	}

	queryResult, err := s.Query(ctx, metric, QueryOptions{
		StartTime:  opts.StartTime,
		EndTime:    opts.EndTime,
		TagFilters: opts.TagFilters,
	})
	if err != nil {
		return AggregateResult{}, err
	}

	// Group points into time buckets
	buckets := computeTimeBuckets(queryResult.Points, opts)

	return AggregateResult{
		Buckets: buckets,
		Metric:  metric,
		Window:  opts.Window,
	}, nil
}

// Close shuts down the store
// Subscribe creates a subscription for real-time metric updates
func (s *RingBufferStore) Subscribe(ctx context.Context, metricIDs []string) (<-chan *MetricUpdateBatch, error) {
	if s.closed {
		return nil, fmt.Errorf("store is closed")
	}

	// For now, we'll implement a simple polling mechanism
	// In a production system, this would use a proper pub/sub pattern
	updateChan := make(chan *MetricUpdateBatch, 10)

	// Start a goroutine to poll for updates
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // Poll every 100ms
		defer ticker.Stop()
		defer close(updateChan)

		// Track last seen positions for each metric
		lastPositions := make(map[string]int)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Check for new data in each subscribed metric
				batch := &MetricUpdateBatch{
					Updates: make([]*MetricUpdateItem, 0),
				}

				s.mu.RLock()
				for _, metricID := range metricIDs {
					if rb, ok := s.buffers[metricID]; ok {
						rb.mu.RLock()
						currentCount := rb.count
						lastPos := lastPositions[metricID]
						
						// If we have new data since last check
						if currentCount > lastPos {
							// Get the latest point
							var latestPoint *MetricPoint
							if rb.count > 0 {
								// Get the most recent point
								idx := (rb.writePos - 1 + rb.size) % rb.size
								if rb.points[idx] != nil {
									latestPoint = &MetricPoint{
										Timestamp: rb.points[idx].Timestamp,
										Value:     rb.points[idx].Value,
										Tags:      rb.points[idx].Tags,
									}
								}
							}
							
							if latestPoint != nil {
								batch.Updates = append(batch.Updates, &MetricUpdateItem{
									MetricID: metricID,
									Point:    latestPoint,
								})
							}
							
							lastPositions[metricID] = currentCount
						}
						rb.mu.RUnlock()
					}
				}
				s.mu.RUnlock()

				// Send batch if we have updates
				if len(batch.Updates) > 0 {
					select {
					case updateChan <- batch:
					case <-ctx.Done():
						return
					default:
						// Channel full, skip this update
					}
				}
			}
		}
	}()

	return updateChan, nil
}

func (s *RingBufferStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.buffers = nil
	return nil
}

// newRingBuffer creates a new ring buffer
func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		points: make([]*MetricPoint, size),
		size:   size,
	}
}

// add adds a point to the ring buffer
func (rb *ringBuffer) add(point *MetricPoint) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.points[rb.writePos] = point
	rb.writePos = (rb.writePos + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	} else {
		// Buffer is full, advance read position
		rb.readStart = (rb.readStart + 1) % rb.size
	}
}

// query retrieves points within a time range
func (rb *ringBuffer) query(startTime, endTime time.Time, tagFilters map[string]string) []*MetricPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	var results []*MetricPoint
	for i := 0; i < rb.count; i++ {
		idx := (rb.readStart + i) % rb.size
		point := rb.points[idx]
		
		// Time range filter
		if !point.Timestamp.Before(startTime) && !point.Timestamp.After(endTime) {
			// Tag filter
			match := true
			for k, v := range tagFilters {
				if point.Tags[k] != v {
					match = false
					break
				}
			}
			if match {
				results = append(results, point)
			}
		}
	}

	// Sort by timestamp in descending order (most recent first)
	sort.Slice(results, func(i, j int) bool {
		return results[j].Timestamp.Before(results[i].Timestamp)
	})

	return results
}

// computeTimeBuckets groups points into time buckets and computes aggregations
func computeTimeBuckets(points []*MetricPoint, opts AggregateOptions) []TimeBucket {
	if len(points) == 0 {
		return []TimeBucket{}
	}

	// Determine bucket boundaries
	bucketStart := opts.StartTime.Truncate(opts.Window)
	bucketEnd := opts.EndTime.Truncate(opts.Window).Add(opts.Window)

	// Create buckets
	buckets := make(map[time.Time]*TimeBucket)
	for t := bucketStart; t.Before(bucketEnd); t = t.Add(opts.Window) {
		buckets[t] = &TimeBucket{
			Time:   t,
			Values: make(map[AggregateFunc]float64),
		}
	}

	// Group points into buckets
	bucketPoints := make(map[time.Time][]float64)
	for _, point := range points {
		bucketTime := point.Timestamp.Truncate(opts.Window)
		bucketPoints[bucketTime] = append(bucketPoints[bucketTime], point.Value)
		if bucket, ok := buckets[bucketTime]; ok {
			bucket.Count++
		}
	}

	// Compute aggregations for each bucket
	for bucketTime, values := range bucketPoints {
		bucket := buckets[bucketTime]
		for _, fn := range opts.Functions {
			bucket.Values[fn] = computeAggregation(values, fn, opts.Window)
		}
	}

	// Convert to sorted slice
	var result []TimeBucket
	for _, bucket := range buckets {
		result = append(result, *bucket)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time.Before(result[j].Time)
	})

	return result
}

// computeAggregation computes a single aggregation function
func computeAggregation(values []float64, fn AggregateFunc, window time.Duration) float64 {
	if len(values) == 0 {
		return 0
	}

	switch fn {
	case AggCount:
		return float64(len(values))
	case AggSum:
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum
	case AggAvg:
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum / float64(len(values))
	case AggMin:
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min
	case AggMax:
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max
	case AggRate:
		// Events per second
		return float64(len(values)) / window.Seconds()
	case AggP50, AggP90, AggP95, AggP99:
		// Sort values for percentile calculation
		sorted := make([]float64, len(values))
		copy(sorted, values)
		sort.Float64s(sorted)
		
		var percentile float64
		switch fn {
		case AggP50:
			percentile = 0.50
		case AggP90:
			percentile = 0.90
		case AggP95:
			percentile = 0.95
		case AggP99:
			percentile = 0.99
		}
		
		idx := int(float64(len(sorted)-1) * percentile)
		return sorted[idx]
	default:
		return 0
	}
}