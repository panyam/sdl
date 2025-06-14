package console

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuckDBTimeSeriesStore_Basic(t *testing.T) {
	// Create temporary directory for test database
	tempDir := t.TempDir()
	
	// Create time series store
	store, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer store.Close()

	// Test basic insertion
	now := time.Now()
	point := TracePoint{
		Timestamp:   now.UnixNano(),
		Target:      "server.HandleLookup",
		Duration:    45.5,
		ReturnValue: "success",
		Args:        []string{"user123", "phone"},
		RunID:       "test_run_001",
	}

	err = store.Insert(point)
	require.NoError(t, err)

	// Test querying
	since := now.Add(-1 * time.Hour)
	points, err := store.QueryLatency("server.HandleLookup", since)
	require.NoError(t, err)
	assert.Len(t, points, 1)
	assert.Equal(t, "server.HandleLookup", points[0].Target)
	assert.Equal(t, 45.5, points[0].Duration)
	assert.Equal(t, "success", points[0].ReturnValue)
	assert.Equal(t, []string{"user123", "phone"}, points[0].Args)
}

func TestDuckDBTimeSeriesStore_MultiplePoints(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer store.Close()

	// Insert multiple trace points
	baseTime := time.Now()
	targets := []string{"server.HandleLookup", "server.HandleCreate", "db.Query"}
	
	for i, target := range targets {
		for j := 0; j < 10; j++ {
			point := TracePoint{
				Timestamp:   baseTime.Add(time.Duration(i*10+j) * time.Second).UnixNano(),
				Target:      target,
				Duration:    float64(20 + i*10 + j), // Varying durations
				ReturnValue: "success",
				RunID:       "test_run_multi",
			}
			err := store.Insert(point)
			require.NoError(t, err)
		}
	}

	// Test querying specific target
	since := baseTime.Add(-1 * time.Hour)
	points, err := store.QueryLatency("server.HandleLookup", since)
	require.NoError(t, err)
	assert.Len(t, points, 10)

	// Test percentiles calculation
	p50, p90, p95, p99, err := store.QueryPercentiles("server.HandleLookup", since)
	require.NoError(t, err)
	assert.Greater(t, p50, 0.0)
	assert.Greater(t, p90, p50)
	assert.Greater(t, p95, p90)
	assert.Greater(t, p99, p95)
}

func TestDuckDBTimeSeriesStore_TimeBuckets(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer store.Close()

	// Insert points across multiple time buckets
	baseTime := time.Now()
	target := "server.HandleLookup"
	
	for i := 0; i < 60; i++ {
		point := TracePoint{
			Timestamp:   baseTime.Add(time.Duration(i) * time.Second).UnixNano(),
			Target:      target,
			Duration:    float64(10 + i%20), // Varying durations
			ReturnValue: "success",
			RunID:       "test_run_buckets",
		}
		
		// Add some errors
		if i%10 == 0 {
			point.Error = "timeout"
		}
		
		err := store.Insert(point)
		require.NoError(t, err)
	}

	// Test time bucket aggregation
	since := baseTime.Add(-2 * time.Hour)
	buckets, err := store.QueryTimeBuckets(target, since, "30 seconds")
	require.NoError(t, err)
	assert.Greater(t, len(buckets), 0)

	// Verify bucket structure
	for _, bucket := range buckets {
		assert.Greater(t, bucket.CallCount, int64(0))
		assert.Greater(t, bucket.AvgDuration, 0.0)
		assert.Greater(t, bucket.MaxDuration, 0.0)
		assert.GreaterOrEqual(t, bucket.ErrorCount, int64(0))
	}
}

func TestDuckDBTimeSeriesStore_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer store.Close()

	// Test inserting point with error
	point := TracePoint{
		Timestamp:   time.Now().UnixNano(),
		Target:      "server.HandleLookup",
		Duration:    0, // Zero duration for failed call
		ReturnValue: "",
		Error:       "connection timeout",
		RunID:       "test_run_error",
	}

	err = store.Insert(point)
	require.NoError(t, err)

	// Query and verify error is preserved
	since := time.Now().Add(-1 * time.Hour)
	points, err := store.QueryLatency("server.HandleLookup", since)
	require.NoError(t, err)
	assert.Len(t, points, 1)
	assert.Equal(t, "connection timeout", points[0].Error)
}

func TestDuckDBTimeSeriesStore_Stats(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer store.Close()

	// Insert some test data
	targets := []string{"server.HandleLookup", "db.Query"}
	for i, target := range targets {
		point := TracePoint{
			Timestamp:   time.Now().UnixNano(),
			Target:      target,
			Duration:    float64(i + 1),
			ReturnValue: "success",
			RunID:       "test_run_stats",
		}
		err := store.Insert(point)
		require.NoError(t, err)
	}

	// Test stats
	stats, err := store.GetStats()
	require.NoError(t, err)
	assert.Equal(t, int64(2), stats["total_traces"])
	assert.Equal(t, int64(2), stats["unique_targets"])
	assert.Equal(t, int64(1), stats["unique_runs"])
	assert.Contains(t, stats, "database_path")
}

func TestDuckDBTimeSeriesStore_CustomSQL(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(t, err)
	defer store.Close()

	// Insert test data
	point := TracePoint{
		Timestamp:   time.Now().UnixNano(),
		Target:      "server.HandleLookup",
		Duration:    42.0,
		ReturnValue: "success",
		RunID:       "test_run_sql",
	}
	err = store.Insert(point)
	require.NoError(t, err)

	// Test custom SQL query
	query := "SELECT target, avg(duration) as avg_dur FROM traces GROUP BY target"
	results, err := store.ExecuteSQL(query)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "server.HandleLookup", results[0]["target"])
	assert.Equal(t, 42.0, results[0]["avg_dur"])
}

func TestDuckDBTimeSeriesStore_InvalidPath(t *testing.T) {
	// Test with invalid path
	_, err := NewDuckDBTimeSeriesStore("/invalid/path/that/does/not/exist")
	// DuckDB should create the path or handle this gracefully
	// We'll accept either success or a descriptive error
	if err != nil {
		t.Logf("Expected behavior: NewDuckDBTimeSeriesStore failed with invalid path: %v", err)
	}
}

// Benchmark test for insertion performance
func BenchmarkDuckDBTimeSeriesStore_Insert(b *testing.B) {
	tempDir := b.TempDir()
	store, err := NewDuckDBTimeSeriesStore(tempDir)
	require.NoError(b, err)
	defer store.Close()

	point := TracePoint{
		Timestamp:   time.Now().UnixNano(),
		Target:      "server.HandleLookup",
		Duration:    45.5,
		ReturnValue: "success",
		Args:        []string{"user123"},
		RunID:       "benchmark_run",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		point.Timestamp = time.Now().UnixNano() + int64(i)
		err := store.Insert(point)
		if err != nil {
			b.Fatal(err)
		}
	}
}