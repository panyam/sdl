# DuckDB Viewer Setup Guide

## Concurrent Access Solutions

DuckDB has limitations with concurrent access, but here are solutions:

### 1. Command Line (Read-Only Mode)

Always use `-readonly` flag when querying while SDL is running:

```bash
# Connect in read-only mode
duckdb -readonly /tmp/sdl_metrics/traces.duckdb

# Or use the monitor script
./tools/monitor_traces.sh
```

### 2. TablePlus Configuration

1. Create connection with path: `/tmp/sdl_metrics/traces.duckdb`
2. In connection settings, add to "Additional Parameters":
   ```
   access_mode=read_only
   ```

### 3. DBeaver Configuration

1. Create DuckDB connection
2. Go to "Driver Properties" tab
3. Add property:
   - Name: `access_mode`
   - Value: `read_only`

### 4. Python Script (Read-Only)

```python
import duckdb

# Connect in read-only mode
conn = duckdb.connect('/tmp/sdl_metrics/traces.duckdb', read_only=True)

# Now you can query while SDL is writing
result = conn.execute("SELECT * FROM traces ORDER BY timestamp DESC LIMIT 10").fetchall()
```

### 5. Alternative: Copy Database for Analysis

If you need full access:

```bash
# Copy the database while SDL is running
cp /tmp/sdl_metrics/traces.duckdb ~/traces_snapshot.duckdb

# Open the copy with full access
duckdb ~/traces_snapshot.duckdb
```

## Why This Happens

DuckDB uses file-level locking for consistency. When SDL console is writing:
- Writers need exclusive access during transactions
- Readers can access between transactions
- Read-only mode ensures no lock conflicts

## Best Practices

1. **Use read-only mode** for all monitoring queries
2. **Keep transactions short** - SDL now uses minimal lock time
3. **Use the monitor script** - it's pre-configured for read-only access
4. **For heavy analysis** - copy the database file first

## Quick Commands

```bash
# Monitor recent traces (read-only)
duckdb -readonly /tmp/sdl_metrics/traces.duckdb -c "SELECT * FROM traces ORDER BY timestamp DESC LIMIT 20"

# Get summary stats (read-only)
duckdb -readonly /tmp/sdl_metrics/traces.duckdb -c "SELECT target, COUNT(*) as count, AVG(duration) as avg FROM traces GROUP BY target"

# Export data for analysis
duckdb -readonly /tmp/sdl_metrics/traces.duckdb -c "COPY (SELECT * FROM traces) TO 'traces.csv' (HEADER, DELIMITER ',')"
```