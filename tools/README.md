# SDL Tools & Scripts

This directory contains utility scripts and tools for working with SDL.

## Scripts

### monitor_traces.sh
Monitor measurements stored in DuckDB in real-time. Provides several viewing options:
- Recent traces
- Summary statistics by target
- Traces per second
- Live monitoring mode
- Interactive DuckDB shell

**Usage:**
```bash
./tools/monitor_traces.sh
```

### test_web_stack.sh
Test the complete web stack including dashboard and WebSocket connections.

### livetest.sh
Live testing script for development.

## DuckDB Database

The measurement database is located at:
```
/tmp/sdl_metrics/traces.duckdb
```

You can also use any DuckDB-compatible tool to view the data:
- DuckDB CLI: `duckdb /tmp/sdl_metrics/traces.duckdb`
- DBeaver (GUI)
- TablePlus (macOS)
- Any tool that supports DuckDB connections

## Common Queries

```sql
-- Recent measurements
SELECT * FROM traces ORDER BY timestamp DESC LIMIT 20;

-- Summary by target
SELECT target, COUNT(*) as count, AVG(duration) as avg_latency 
FROM traces GROUP BY target;

-- Live measurements (last 10 seconds)
SELECT * FROM traces 
WHERE timestamp > (unixepoch() - 10) * 1e9 
ORDER BY timestamp DESC;
```