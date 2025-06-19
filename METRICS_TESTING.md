# SDL Metrics System Testing Guide

This guide shows how to test the SDL metrics system from the command line.

## Prerequisites

1. Build and Start SDL:
   ```bash
   air
   ```
   
or to do so manually:

  ```bash
  make
  
  sdl serve --port 8080
  ```
   
This will continuously build it (into the $GOBIN/ folder)

2. Load a system (in another terminal):
   ```bash
   
   SDL=$GOBIN/sdl
   
   $SDL load examples/contacts/contacts.sdl
   $SDL use ContactsSystem
   ```

## Testing with SDL CLI Commands

The SDL CLI provides comprehensive `measure` commands for managing metrics:

### 1. List All Measurements

```bash
$SDL measure list
```

### 2. Add a Measurement

**Basic syntax:**
```bash
$SDL measure add <id> <component.method> <metric> [options]
```

**Options:**
- `--aggregation <type>` - Aggregation type (rate, sum, avg, p50, p90, p95, p99)
- `--window <duration>` - Time window (default: 10s)
- `--result-value <filter>` - Filter by result value (default: *)

**Examples:**

Count/Throughput measurement:
```bash
$SDL measure add server_qps server.Lookup count --aggregation rate
```

Latency measurement:
```bash
$SDL measure add server_p95 server.Lookup latency --aggregation p95 --window 30s
```

Error rate measurement:
```bash
$SDL measure add server_errors server.Lookup count --aggregation rate --result-value "Val(Bool: false)"
```

### 3. Generate Traffic to Create Metrics

Using traffic generators:
```bash
$SDL gen add lookup server.Lookup 20
$SDL gen start
# Let it run for a while
$SDL gen stop
```

### 4. View Measurement Data

Get raw data points:
```bash
$SDL measure data server_qps
```

Get aggregated statistics:
```bash
$SDL measure stats server_qps
```

### 5. Remove Measurements

Remove specific measurements:
```bash
$SDL measure remove server_qps server_p95
```

Clear all measurements:
```bash
$SDL measure clear
```

## Testing with curl (Advanced)

### 1. List All Measurements

```bash
curl http://localhost:8080/api/measurements
```

### 2. Add a Measurement

**Count/Throughput Measurement:**
```bash
curl -X POST http://localhost:8080/api/measurements \
  -H "Content-Type: application/json" \
  -d '{
    "id": "server_qps",
    "name": "Server QPS",
    "component": "server",
    "methods": ["Lookup"],
    "resultValue": "*",
    "metric": "count",
    "aggregation": "rate",
    "window": "10s"
  }'
```

**Latency Measurement:**
```bash
curl -X POST http://localhost:8080/api/measurements \
  -H "Content-Type: application/json" \
  -d '{
    "id": "server_latency",
    "name": "Server Latency",
    "component": "server",
    "methods": ["Lookup"],
    "resultValue": "*",
    "metric": "latency",
    "aggregation": "p95",
    "window": "30s"
  }'
```

### 3. Generate Traffic to Create Metrics

**Using sdl run:**
```bash
$SDL run results server.Lookup --runs 100 --workers 10
```

**Using traffic generators:**
```bash
$SDL gen add lookup server.Lookup 20
$SDL gen start
# Let it run for a while
$SDL gen stop
```

### 4. View Measurement Data

**Get raw data points:**
```bash
curl http://localhost:8080/api/measurements/server_qps/data?limit=10
```

**Get aggregated statistics:**
```bash
curl http://localhost:8080/api/measurements/server_qps/aggregated
```

### 5. Get Measurement Details

```bash
curl http://localhost:8080/api/measurements/server_qps
```

### 6. Remove a Measurement

```bash
curl -X DELETE http://localhost:8080/api/measurements/server_qps
```

## Measurement Configuration

### Metric Types
- `count` - Number of events (used for throughput, error counts)
- `latency` - Duration measurements (response times)

### Aggregation Types
For `count` metrics:
- `sum` - Total count in window
- `rate` - Count per second (throughput)

For `latency` metrics:
- `avg` - Average latency
- `min` - Minimum latency
- `max` - Maximum latency
- `p50` - 50th percentile (median)
- `p90` - 90th percentile
- `p95` - 95th percentile
- `p99` - 99th percentile

### Result Value Filters
- `*` - Match all results
- `Val(Bool: true)` - Match successful operations (for SDL bool returns)
- `Val(Bool: false)` - Match failed operations
- `!=false` - Match non-false results (coming soon)

## Example Scenarios

### Monitor Cache Performance

Using CLI commands:
```bash
# Cache hit rate
$SDL measure add cache_hits cache.Get count --aggregation rate --window 60s --result-value "Val(Bool: true)"

# Cache miss rate  
$SDL measure add cache_misses cache.Get count --aggregation rate --window 60s --result-value "Val(Bool: false)"

# View statistics
$SDL measure stats cache_hits
$SDL measure stats cache_misses
```

### Monitor Database Performance

```bash
# Database query latency (P95)
$SDL measure add db_p95 database.Query latency --aggregation p95 --window 5m

# Database throughput
$SDL measure add db_qps database.Query count --aggregation rate

# View performance
$SDL measure stats db_p95
$SDL measure stats db_qps
```

### Complete System Monitoring

```bash
# Set up comprehensive monitoring
$SDL measure add server_qps server.Lookup count --aggregation rate
$SDL measure add server_p99 server.Lookup latency --aggregation p99
$SDL measure add cache_hit_rate cache.Get count --aggregation rate --result-value "Val(Bool: true)"
$SDL measure add db_qps database.Query count --aggregation rate

# Generate load
$SDL gen add traffic server.Lookup 50
$SDL gen start

# Monitor in real-time
watch -n 2 '$SDL measure list'

# Check statistics
$SDL measure stats server_qps
$SDL measure stats server_p99

# Stop and clean up
$SDL gen stop
$SDL measure clear
```

## Using with jq for Better Output

Install jq for prettier JSON output:
```bash
# macOS
brew install jq

# Linux
apt-get install jq
```

Then pipe curl output through jq:
```bash
curl http://localhost:8080/api/measurements | jq '.'
curl http://localhost:8080/api/measurements/server_qps/data?limit=5 | jq '.'
```

## Automated Test Scripts

Three test scripts are provided:

1. **metrics_quickstart.sh** - Shows exact commands to copy/paste
2. **test_metrics_simple.sh** - Basic test covering common scenarios
3. **test_metrics.sh** - Comprehensive test with error cases

Run them:
```bash
# Quick reference
./metrics_quickstart.sh

# Simple test
./test_metrics_simple.sh

# Full test
./test_metrics.sh
```

All scripts now use the SDL CLI commands instead of curl.

## Troubleshooting

### No data points appearing
- Ensure the system is loaded and active
- Check that component names match exactly (e.g., "server" not "Server")
- Verify methods exist on the component
- Make sure you're generating traffic after adding measurements

### Component not found errors
- Component names are instance names from the system definition, not type names
- Example: use "server" not "ContactAppServer"

### Metrics show very small latency values
- Latency is stored in nanoseconds internally
- The API converts to milliseconds for display
- SDL simulations often have sub-millisecond latencies
