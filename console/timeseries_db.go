package console

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

// TracePoint represents a single trace measurement point
type TracePoint struct {
	Timestamp   int64    `json:"timestamp"`   // Unix timestamp in nanoseconds
	Target      string   `json:"target"`      // e.g., "server.HandleLookup"
	Duration    float64  `json:"duration"`    // Duration in simulation time units
	ReturnValue string   `json:"return_value"`// Method return value as string
	Error       string   `json:"error,omitempty"` // Error message if any
	Args        []string `json:"args,omitempty"`  // Method arguments
	RunID       string   `json:"run_id"`      // Simulation run identifier
}

// DuckDBTimeSeriesStore provides time-series storage using DuckDB
type DuckDBTimeSeriesStore struct {
	mu       sync.RWMutex
	db       *sql.DB
	dbPath   string
	insertStmt *sql.Stmt
}

// NewDuckDBTimeSeriesStore creates a new DuckDB-based time-series store
func NewDuckDBTimeSeriesStore(dataDir string) (*DuckDBTimeSeriesStore, error) {
	// Create data directory if it doesn't exist
	if dataDir == "" {
		dataDir = "/tmp/sdl_metrics"
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", dataDir, err)
	}

	dbPath := filepath.Join(dataDir, "traces.duckdb")
	
	// Connect to DuckDB
	// Default settings should allow concurrent reads
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB at %s: %w", dbPath, err)
	}
	
	// Set connection pool settings for better concurrency
	db.SetMaxOpenConns(1)    // DuckDB works best with single writer
	db.SetMaxIdleConns(1)    // Keep connection ready
	db.SetConnMaxLifetime(0) // Connection doesn't expire

	store := &DuckDBTimeSeriesStore{
		db:     db,
		dbPath: dbPath,
	}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Prepare insert statement
	if err := store.prepareStatements(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return store, nil
}

// initSchema creates the time-series optimized schema
func (ts *DuckDBTimeSeriesStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS traces (
		timestamp TIMESTAMP,      -- Timestamp of the trace event
		target VARCHAR,           -- Target method/component (e.g., "server.HandleLookup")
		duration DOUBLE,          -- Execution duration in simulation time units
		return_value VARCHAR,     -- Return value as string
		error VARCHAR,            -- Error message (NULL if no error)
		args JSON,                -- Method arguments as JSON array
		run_id VARCHAR            -- Simulation run identifier
	);

	-- Optimized indexes for time-series queries
	CREATE INDEX IF NOT EXISTS idx_target_time ON traces(target, timestamp);
	CREATE INDEX IF NOT EXISTS idx_run_time ON traces(run_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON traces(timestamp);
	`

	_, err := ts.db.Exec(schema)
	return err
}

// prepareStatements prepares commonly used SQL statements
func (ts *DuckDBTimeSeriesStore) prepareStatements() error {
	insertSQL := `
	INSERT INTO traces (timestamp, target, duration, return_value, error, args, run_id)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	stmt, err := ts.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	
	ts.insertStmt = stmt
	return nil
}

// Insert stores a trace point in the database
func (ts *DuckDBTimeSeriesStore) Insert(point TracePoint) error {
	// Use RWMutex for reads, only lock for writes
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Convert timestamp to time.Time
	timestamp := time.Unix(0, point.Timestamp)
	
	// Convert args to JSON
	var argsJSON interface{}
	if len(point.Args) > 0 {
		argsBytes, err := json.Marshal(point.Args)
		if err != nil {
			return fmt.Errorf("failed to marshal args: %w", err)
		}
		argsJSON = string(argsBytes)
	}

	// Handle null error
	var errorStr interface{}
	if point.Error != "" {
		errorStr = point.Error
	}

	// Execute insert without explicit transaction - DuckDB will handle it
	_, err := ts.insertStmt.Exec(
		timestamp,
		point.Target,
		point.Duration,
		point.ReturnValue,
		errorStr,
		argsJSON,
		point.RunID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to insert trace point: %w", err)
	}

	return nil
}

// QueryLatency retrieves latency data for a specific target within a time range
func (ts *DuckDBTimeSeriesStore) QueryLatency(target string, since time.Time) ([]TracePoint, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	query := `
	SELECT timestamp, target, duration, return_value, error, args, run_id
	FROM traces 
	WHERE target = ? AND timestamp >= ?
	ORDER BY timestamp
	`

	rows, err := ts.db.Query(query, target, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query latency data: %w", err)
	}
	defer rows.Close()

	var points []TracePoint
	for rows.Next() {
		var point TracePoint
		var timestamp time.Time
		var errorStr, argsJSON sql.NullString

		err := rows.Scan(
			&timestamp,
			&point.Target,
			&point.Duration,
			&point.ReturnValue,
			&errorStr,
			&argsJSON,
			&point.RunID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		point.Timestamp = timestamp.UnixNano()
		
		if errorStr.Valid {
			point.Error = errorStr.String
		}
		
		if argsJSON.Valid {
			var args []string
			if err := json.Unmarshal([]byte(argsJSON.String), &args); err == nil {
				point.Args = args
			}
		}

		points = append(points, point)
	}

	return points, nil
}

// QueryPercentiles calculates percentiles for a target within a time range
func (ts *DuckDBTimeSeriesStore) QueryPercentiles(target string, since time.Time) (p50, p90, p95, p99 float64, err error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	query := `
	SELECT 
		percentile_cont(0.50) WITHIN GROUP (ORDER BY duration) as p50,
		percentile_cont(0.90) WITHIN GROUP (ORDER BY duration) as p90,
		percentile_cont(0.95) WITHIN GROUP (ORDER BY duration) as p95,
		percentile_cont(0.99) WITHIN GROUP (ORDER BY duration) as p99
	FROM traces 
	WHERE target = ? AND timestamp >= ?
	`

	row := ts.db.QueryRow(query, target, since)
	err = row.Scan(&p50, &p90, &p95, &p99)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to query percentiles: %w", err)
	}

	return p50, p90, p95, p99, nil
}

// QueryTimeBuckets returns time-bucketed aggregations
func (ts *DuckDBTimeSeriesStore) QueryTimeBuckets(target string, since time.Time, bucketSize string) ([]TimeBucketResult, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	query := `
	SELECT 
		time_bucket(INTERVAL '%s', timestamp) as bucket,
		avg(duration) as avg_duration,
		max(duration) as max_duration,
		count(*) as call_count,
		count(*) FILTER (WHERE error IS NOT NULL) as error_count
	FROM traces 
	WHERE target = ? AND timestamp >= ?
	GROUP BY bucket
	ORDER BY bucket
	`

	formattedQuery := fmt.Sprintf(query, bucketSize)
	rows, err := ts.db.Query(formattedQuery, target, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query time buckets: %w", err)
	}
	defer rows.Close()

	var results []TimeBucketResult
	for rows.Next() {
		var result TimeBucketResult
		var bucket time.Time

		err := rows.Scan(
			&bucket,
			&result.AvgDuration,
			&result.MaxDuration,
			&result.CallCount,
			&result.ErrorCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bucket row: %w", err)
		}

		result.Bucket = bucket.UnixNano()
		results = append(results, result)
	}

	return results, nil
}

// TimeBucketResult represents aggregated metrics for a time bucket
type TimeBucketResult struct {
	Bucket      int64   `json:"bucket"`       // Bucket timestamp in nanoseconds
	AvgDuration float64 `json:"avg_duration"` // Average duration in bucket
	MaxDuration float64 `json:"max_duration"` // Maximum duration in bucket
	CallCount   int64   `json:"call_count"`   // Number of calls in bucket
	ErrorCount  int64   `json:"error_count"`  // Number of errors in bucket
}

// ExecuteSQL runs a custom SQL query and returns results as JSON
func (ts *DuckDBTimeSeriesStore) ExecuteSQL(query string, args ...interface{}) ([]map[string]interface{}, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	rows, err := ts.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Create a slice to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}

// GetStats returns basic statistics about the database
func (ts *DuckDBTimeSeriesStore) GetStats() (map[string]interface{}, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	query := `
	SELECT 
		count(*) as total_traces,
		count(DISTINCT target) as unique_targets,
		count(DISTINCT run_id) as unique_runs,
		min(timestamp) as earliest_trace,
		max(timestamp) as latest_trace
	FROM traces
	`

	row := ts.db.QueryRow(query)
	
	var totalTraces, uniqueTargets, uniqueRuns int64
	var earliestTrace, latestTrace sql.NullTime

	err := row.Scan(&totalTraces, &uniqueTargets, &uniqueRuns, &earliestTrace, &latestTrace)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_traces":    totalTraces,
		"unique_targets":  uniqueTargets,
		"unique_runs":     uniqueRuns,
		"database_path":   ts.dbPath,
	}

	if earliestTrace.Valid {
		stats["earliest_trace"] = earliestTrace.Time.Format(time.RFC3339)
	}
	if latestTrace.Valid {
		stats["latest_trace"] = latestTrace.Time.Format(time.RFC3339)
	}

	return stats, nil
}

// Close closes the database connection and cleans up resources
func (ts *DuckDBTimeSeriesStore) Close() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.insertStmt != nil {
		ts.insertStmt.Close()
	}
	
	return ts.db.Close()
}