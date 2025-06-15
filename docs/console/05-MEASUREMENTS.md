# SDL Console & Server Tutorial - Measurements

This chapter covers SDL's measurement system for collecting and analyzing performance metrics in real-time.

## Prerequisites

Complete [Traffic Generation](04-TRAFFIC-GENERATION.md) and have:
- Understanding of traffic generators
- Active traffic generation for measurement collection

## What are Measurements?

SDL measurements automatically collect performance metrics during system execution. They capture:

- **Latency** - Response time distribution
- **Throughput** - Calls per second
- **Success Rate** - Error percentages
- **Custom Metrics** - Application-specific data

Measurements work with both single `run` commands and continuous traffic generators.

## 1. Creating Measurements

### Basic Latency Measurement
```
SDL[ContactsSystem]> measure add lat1 server.HandleLookup latency
✅ Measurement 'lat1' created
🎯 Target: server.HandleLookup
📊 Metric: latency (response time)
💾 Storage: DuckDB time-series database
```

### Different Measurement Types
```
SDL[ContactsSystem]> measure add throughput1 server.HandleLookup throughput
SDL[ContactsSystem]> measure add errors1 server.HandleLookup error_rate
SDL[ContactsSystem]> measure add latency2 server.HandleCreate latency
```

### Measurement Components
Each measurement has:
- **Name** - Unique identifier (`lat1`)
- **Target** - Method being measured (`server.HandleLookup`)
- **Type** - Metric type (`latency`, `throughput`, `error_rate`)
- **Storage** - Automatic persistence in DuckDB

## 2. Viewing Measurements

### List All Measurements
```
SDL[ContactsSystem]> measure list
Active Measurements:
┌─────────────┬─────────────────────┬─────────────┬────────────┐
│ Name        │ Target              │ Type        │ Data Points│
├─────────────┼─────────────────────┼─────────────┼────────────┤
│ lat1        │ server.HandleLookup │ latency     │ 0          │
│ throughput1 │ server.HandleLookup │ throughput  │ 0          │
│ errors1     │ server.HandleLookup │ error_rate  │ 0          │
│ latency2    │ server.HandleCreate │ latency     │ 0          │
└─────────────┴─────────────────────┴─────────────┴────────────┘
```

### Measurement Status
```
SDL[ContactsSystem]> measure status
Measurement Status:
📊 Total Measurements: 4
💾 Database: DuckDB (connected)
🔄 Collection Rate: Real-time
📈 Total Data Points: 0
```

## 3. Collecting Measurement Data

### With Single Runs
Execute runs with active measurements:
```
SDL[ContactsSystem]> run test1 server.HandleLookup 100
✅ Running test1: server.HandleLookup (100 calls)
📊 Collecting measurements: lat1, throughput1, errors1
🎯 Execution completed: test1
📈 Measurements updated with 100 data points
```

### With Traffic Generators
Start generators to collect continuous data:
```
SDL[ContactsSystem]> gen add load1 server.HandleLookup 20
SDL[ContactsSystem]> gen start load1
✅ Generator 'load1' started
📊 Measurements collecting real-time data
🎯 Generating 20 calls/second → lat1, throughput1, errors1
```

Watch the server terminal for measurement collection logs:
```
📊 Measurement lat1: Recorded latency 45.2ms for server.HandleLookup
📊 Measurement lat1: Recorded latency 38.7ms for server.HandleLookup
📊 Measurement throughput1: Current rate 19.8 calls/sec
```

## 4. Querying Measurement Data

### View Recent Data
```
SDL[ContactsSystem]> measure data lat1
Recent data for 'lat1' (last 10 points):
┌─────────────────────┬─────────────────────┬─────────┐
│ Timestamp           │ Target              │ Latency │
├─────────────────────┼─────────────────────┼─────────┤
│ 2024-06-15 10:30:45 │ server.HandleLookup │ 45.2ms  │
│ 2024-06-15 10:30:46 │ server.HandleLookup │ 38.7ms  │
│ 2024-06-15 10:30:47 │ server.HandleLookup │ 42.1ms  │
│ 2024-06-15 10:30:48 │ server.HandleLookup │ 39.9ms  │
│ 2024-06-15 10:30:49 │ server.HandleLookup │ 44.3ms  │
└─────────────────────┴─────────────────────┴─────────┘
```

### Statistical Summary
```
SDL[ContactsSystem]> measure stats lat1
Statistics for 'lat1' (last 1 hour):
📊 Total Samples: 1,250
⏱️  Average Latency: 41.7ms
📈 95th Percentile: 58.3ms  
📉 Min Latency: 22.1ms
📊 Max Latency: 89.4ms
🎯 Standard Deviation: 12.3ms
```

### Time Range Queries
```
SDL[ContactsSystem]> measure data lat1 --last 5m
SDL[ContactsSystem]> measure data lat1 --last 1h
SDL[ContactsSystem]> measure data lat1 --since "2024-06-15 10:00:00"
```

## 5. Advanced Measurement Analysis

### Compare Multiple Measurements
```
SDL[ContactsSystem]> measure compare lat1 latency2
Latency Comparison (last 1 hour):
┌─────────────────────┬─────────────┬─────────────┐
│ Metric              │ lat1        │ latency2    │
├─────────────────────┼─────────────┼─────────────┤
│ Average Latency     │ 41.7ms      │ 67.2ms      │
│ 95th Percentile     │ 58.3ms      │ 89.1ms      │
│ Samples             │ 1,250       │ 485         │
│ Target              │ HandleLookup│ HandleCreate│
└─────────────────────┴─────────────┴─────────────┘
```

### Export Data
```
SDL[ContactsSystem]> measure export lat1 csv latency_data.csv
✅ Exported 1,250 data points to latency_data.csv

SDL[ContactsSystem]> measure export lat1 json latency_data.json
✅ Exported 1,250 data points to latency_data.json
```

## 6. Real-time Monitoring

### Continuous Monitoring
SDL provides several ways to monitor measurements in real-time:

#### 1. Console Updates
```
SDL[ContactsSystem]> measure watch lat1
📊 Watching lat1 (Ctrl+C to stop)
[10:30:45] 45.2ms  [10:30:46] 38.7ms  [10:30:47] 42.1ms
[10:30:48] 39.9ms  [10:30:49] 44.3ms  [10:30:50] 41.8ms
```

#### 2. Web Dashboard
The web dashboard at http://localhost:8080 automatically displays:
- Real-time measurement charts
- Live latency distributions
- Throughput graphs
- Error rate tracking

#### 3. External Database Access
Use external tools to query the DuckDB database:
```bash
# In a separate terminal (read-only access)
./tools/monitor_traces.sh lat1
```

## 7. Database Integration

### DuckDB Storage
All measurements are automatically stored in DuckDB with:
- **Microsecond precision** timestamps
- **Efficient columnar** storage for time-series data
- **Concurrent read access** while measurements are active
- **SQL query support** for advanced analysis

### Direct SQL Queries
```
SDL[ContactsSystem]> sql SELECT AVG(latency_ms) FROM traces WHERE target='server.HandleLookup' AND timestamp > NOW() - INTERVAL 5 MINUTE
Average latency (last 5 min): 42.3ms

SDL[ContactsSystem]> sql SELECT COUNT(*) FROM traces WHERE timestamp > NOW() - INTERVAL 1 HOUR  
Total samples (last hour): 2,847
```

### Schema Information
```
SDL[ContactsSystem]> sql DESCRIBE traces
Table: traces
┌─────────────┬─────────────┬─────────────┐
│ Column      │ Type        │ Description │
├─────────────┼─────────────┼─────────────┤
│ timestamp   │ TIMESTAMP   │ Event time  │
│ target      │ VARCHAR     │ Method name │
│ latency_ms  │ DOUBLE      │ Response ms │
│ success     │ BOOLEAN     │ Success flag│
│ error_msg   │ VARCHAR     │ Error text  │
└─────────────┴─────────────┴─────────────┘
```

## 8. Measurement Scenarios

### Performance Baseline
```
# Establish performance baseline
SDL[ContactsSystem]> measure add baseline server.HandleLookup latency
SDL[ContactsSystem]> run baseline_test server.HandleLookup 1000
SDL[ContactsSystem]> measure stats baseline
# Record baseline metrics for comparison
```

### Load Testing with Measurements
```
# Measure performance under increasing load
SDL[ContactsSystem]> measure add load_test server.HandleLookup latency
SDL[ContactsSystem]> gen add light_load server.HandleLookup 10
SDL[ContactsSystem]> gen start light_load
# ... collect data for 5 minutes ...
SDL[ContactsSystem]> gen set light_load rate 25
# ... observe latency changes ...
SDL[ContactsSystem]> gen set light_load rate 50
```

### Error Rate Monitoring
```
# Monitor error rates during stress testing
SDL[ContactsSystem]> measure add error_tracking server.HandleLookup error_rate
SDL[ContactsSystem]> gen add stress_load server.HandleLookup 100
SDL[ContactsSystem]> gen start stress_load
SDL[ContactsSystem]> measure watch error_tracking
```

## Command Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `measure add <name> <method> <type>` | Create measurement | `measure add lat1 server.HandleLookup latency` |
| `measure list` | Show all measurements | `measure list` |
| `measure data <name>` | View recent data | `measure data lat1` |
| `measure stats <name>` | Statistical summary | `measure stats lat1` |
| `measure watch <name>` | Real-time monitoring | `measure watch lat1` |
| `measure compare <name1> <name2>` | Compare measurements | `measure compare lat1 lat2` |
| `measure export <name> <format> <file>` | Export data | `measure export lat1 csv data.csv` |
| `measure remove <name>` | Delete measurement | `measure remove lat1` |
| `sql <query>` | Direct SQL query | `sql SELECT * FROM traces LIMIT 10` |

## What's Next?

Now that you can collect and analyze measurements, continue to **[Web Dashboard](06-WEB-DASHBOARD.md)** to learn about the powerful real-time visualization capabilities.

## Best Practices

1. **Measure What Matters** - Focus on key performance indicators
2. **Start Before Load** - Create measurements before running traffic generators
3. **Use Descriptive Names** - Name measurements to reflect their purpose
4. **Monitor Continuously** - Use the web dashboard for real-time insights
5. **Export for Analysis** - Save data for historical comparison
6. **Query Efficiently** - Use SQL for complex analysis needs
7. **Watch Error Rates** - Monitor both performance and reliability