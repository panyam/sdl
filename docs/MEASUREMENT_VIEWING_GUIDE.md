# SDL Measurement Viewing Guide

## DuckDB Database Location

The SDL console stores all measurement data in a DuckDB database at:
```
/tmp/sdl_metrics/traces.duckdb
```

## Viewing Measurements with External Tools

### 1. Using DuckDB CLI

Install DuckDB CLI:
```bash
# macOS
brew install duckdb

# Or download from https://duckdb.org/docs/installation/
```

Connect to the database:
```bash
duckdb /tmp/sdl_metrics/traces.duckdb
```

Common queries to monitor measurements:

```sql
-- View recent traces (last 20)
SELECT timestamp/1e9 as time_sec, target, duration, run_id 
FROM traces 
ORDER BY timestamp DESC 
LIMIT 20;

-- Summary statistics by target
SELECT target, 
       COUNT(*) as count,
       AVG(duration) as avg_duration,
       MIN(duration) as min_duration,
       MAX(duration) as max_duration,
       PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY duration) as p50,
       PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration) as p95
FROM traces
GROUP BY target;

-- Real-time monitoring (run periodically)
SELECT datetime(timestamp/1e9, 'unixepoch') as time,
       target,
       duration,
       run_id
FROM traces
WHERE timestamp > (unixepoch() - 60) * 1e9  -- Last 60 seconds
ORDER BY timestamp DESC;

-- Traces per second by target
SELECT target,
       strftime('%Y-%m-%d %H:%M:%S', datetime(timestamp/1e9, 'unixepoch')) as second,
       COUNT(*) as requests_per_second
FROM traces
GROUP BY target, second
ORDER BY second DESC
LIMIT 10;
```

### 2. Using DBeaver (GUI)

1. Download DBeaver from https://dbeaver.io/
2. Create new connection:
   - Database type: DuckDB
   - Path: `/tmp/sdl_metrics/traces.duckdb`
3. Browse the `traces` table and run queries

### 3. Using TablePlus (macOS)

1. Download TablePlus from https://tableplus.com/
2. Create new connection:
   - Choose "DuckDB" 
   - Database path: `/tmp/sdl_metrics/traces.duckdb`
3. View data and create custom queries

### 4. Python Script for Live Monitoring

Create a file `monitor_traces.py`:

```python
import duckdb
import time
from datetime import datetime

# Connect to DuckDB
conn = duckdb.connect('/tmp/sdl_metrics/traces.duckdb')

while True:
    # Get traces from last 5 seconds
    result = conn.execute("""
        SELECT 
            datetime(timestamp/1e9, 'unixepoch') as time,
            target,
            duration,
            run_id
        FROM traces
        WHERE timestamp > (unixepoch() - 5) * 1e9
        ORDER BY timestamp DESC
        LIMIT 10
    """).fetchall()
    
    print(f"\n=== Traces at {datetime.now()} ===")
    for row in result:
        print(f"{row[0]} | {row[1]:30} | {row[2]:8.3f} | {row[3]}")
    
    time.sleep(1)
```

## Understanding the Data

### Table Schema

The `traces` table contains:
- `timestamp` (BIGINT): Unix timestamp in nanoseconds
- `target` (VARCHAR): The SDL method being measured (e.g., "server.HandleLookup")
- `duration` (DOUBLE): Execution duration in simulation time units
- `return_value` (TEXT): JSON-encoded return value
- `error` (TEXT): Error message if any
- `args` (TEXT): JSON-encoded method arguments
- `run_id` (VARCHAR): Unique identifier for each simulation run

### Measurement Flow

1. When you run `measure add lat1 server.HandleLookup latency`, it registers the target
2. When traffic generators call `Canvas.Run()`, the MeasurementTracer captures:
   - Entry/exit events from the ExecutionTracer
   - Extracts timing data for registered targets
   - Stores in DuckDB with wall-clock timestamps
3. Data accumulates in real-time as simulations run

### Tips for Monitoring

1. **Keep DuckDB CLI open** in another terminal:
   ```bash
   duckdb /tmp/sdl_metrics/traces.duckdb
   .timer on  -- Show query execution time
   ```

2. **Create a view for easy monitoring**:
   ```sql
   CREATE VIEW recent_metrics AS
   SELECT 
       datetime(timestamp/1e9, 'unixepoch') as time,
       target,
       duration,
       CASE WHEN error != '' THEN 'ERROR' ELSE 'OK' END as status
   FROM traces
   WHERE timestamp > (unixepoch() - 300) * 1e9;  -- Last 5 minutes
   ```

3. **Export data for analysis**:
   ```sql
   COPY (SELECT * FROM traces) TO 'traces_export.csv' (HEADER, DELIMITER ',');
   ```

## Troubleshooting

If you don't see data:
1. Check if the database exists: `ls -la /tmp/sdl_metrics/`
2. Verify measurements are registered: `measure list` in SDL console
3. Ensure generators are running: `gen list` should show "running" status
4. Check for errors in the console output