package runtime

import (
	"sync"
	"time"

	"github.com/panyam/sdl/core"
)

// CircularBuffer stores a fixed number of metric points in a circular fashion
type CircularBuffer struct {
	mu       sync.RWMutex
	points   []MetricPoint
	head     int // Next write position
	size     int // Current number of points
	capacity int // Maximum capacity
}

// NewCircularBuffer creates a new circular buffer with the given capacity
func NewCircularBuffer(capacity int) *CircularBuffer {
	return &CircularBuffer{
		points:   make([]MetricPoint, capacity),
		capacity: capacity,
	}
}

// Add adds a new metric point to the buffer
func (cb *CircularBuffer) Add(point MetricPoint) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.points[cb.head] = point
	cb.head = (cb.head + 1) % cb.capacity
	
	if cb.size < cb.capacity {
		cb.size++
	}
}

// GetLatest returns the most recent n points
func (cb *CircularBuffer) GetLatest(n int) []MetricPoint {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	if n > cb.size {
		n = cb.size
	}
	
	result := make([]MetricPoint, n)
	
	// Calculate starting position
	start := cb.head - n
	if start < 0 {
		start += cb.capacity
	}
	
	// Copy points in chronological order
	for i := 0; i < n; i++ {
		idx := (start + i) % cb.capacity
		result[i] = cb.points[idx]
	}
	
	return result
}

// GetSince returns all points since the given timestamp
func (cb *CircularBuffer) GetSince(since core.Duration) []MetricPoint {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	result := make([]MetricPoint, 0, cb.size)
	
	// Start from the oldest point
	start := cb.head
	if cb.size < cb.capacity {
		start = 0
	}
	
	// Collect points newer than the timestamp
	for i := 0; i < cb.size; i++ {
		idx := (start + i) % cb.capacity
		if cb.points[idx].Timestamp >= since {
			result = append(result, cb.points[idx])
		}
	}
	
	return result
}

// GetInWindow returns all points within the time window from the latest point
func (cb *CircularBuffer) GetInWindow(window time.Duration) []MetricPoint {
	cb.mu.RLock()
	
	if cb.size == 0 {
		cb.mu.RUnlock()
		return nil
	}
	
	// Get the latest timestamp
	latestIdx := (cb.head - 1 + cb.capacity) % cb.capacity
	if cb.size < cb.capacity && cb.head > 0 {
		latestIdx = cb.head - 1
	}
	latestTime := cb.points[latestIdx].Timestamp
	
	// Convert window to virtual time units
	windowDuration := core.Duration(window.Seconds() * 1e9) // Convert to nanoseconds
	since := latestTime - windowDuration
	if since < 0 {
		since = 0
	}
	
	// Unlock before calling GetSince to avoid double lock
	cb.mu.RUnlock()
	return cb.GetSince(since)
}

// Size returns the current number of points in the buffer
func (cb *CircularBuffer) Size() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.size
}

// Clear removes all points from the buffer
func (cb *CircularBuffer) Clear() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.head = 0
	cb.size = 0
	// Clear the slice to help GC
	for i := range cb.points {
		cb.points[i] = MetricPoint{}
	}
}

// GetAll returns all points in chronological order
func (cb *CircularBuffer) GetAll() []MetricPoint {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	if cb.size == 0 {
		return nil
	}
	
	result := make([]MetricPoint, cb.size)
	
	// Start from the oldest point
	start := cb.head
	if cb.size < cb.capacity {
		start = 0
	}
	
	// Copy all points in chronological order
	for i := 0; i < cb.size; i++ {
		idx := (start + i) % cb.capacity
		result[i] = cb.points[idx]
	}
	
	return result
}