# SDL Runtime Metrics System

## Overview

The SDL Runtime Metrics system provides real-time performance monitoring during system execution. It processes trace events to calculate latency, throughput, and error metrics without requiring external dependencies.

## Design Decisions

### 1. Runtime Package Ownership
- Metrics are a core runtime feature, not a UI concern
- Enables testing without HTTP/WebSocket infrastructure  
- Console package only provides API/UI layer on top
- Other tools can directly use runtime metrics

### 2. Simple Component + Method Targeting
- No regex patterns - just component name + method list
- Leverages SDL's system structure
- Component names resolve to ComponentInstance objects
- Efficient matching during trace processing

### 3. Two Fundamental Metric Types
- **count**: How many times something happened
- **latency**: How long something took
- Everything else is derived through aggregation

### 4. In-Memory Storage
- CircularBuffer for recent data points
- No external database dependency for live metrics
- Configurable retention (e.g., last 10,000 points)
- Can add persistence later if needed

### 5. Result Value Filtering
- Use ReturnValue field from TraceEvent for error detection
- Bool returns: false = error, true = success
- Enum returns: Different values for different outcomes
- Start with exact matching, extend to patterns later

## Architecture

```
runtime/
├── metrics_types.go      # Core types and interfaces
├── metrics_processor.go  # Event processing logic
├── metrics_data.go       # Data access and aggregation
└── metrics_test.go       # Unit tests

console/
├── metrics_api.go        # REST endpoints
└── metrics_api_test.go   # API tests
```

### Integration Points

1. **SimpleEval** hooks:
   ```go
   // When trace event is generated
   if s.runtime.metricStore != nil {
       s.runtime.metricStore.ProcessTraceEvent(event)
   }
   ```

2. **Runtime** struct:
   ```go
   type Runtime struct {
       // existing fields...
       metricStore *MetricStore
   }
   ```

## Metric Types

### Base Metrics
1. **count** - Number of events
   - Used for: throughput, error counts, success counts
   - Value: Always 1.0 per event

2. **latency** - Duration measurements  
   - Used for: response times, processing delays
   - Value: Event duration in nanoseconds

### Aggregations

For **count** metrics:
- `sum`: Total count in window
- `rate`: Count per second (throughput)

For **latency** metrics:
- `avg`: Average latency
- `min`: Minimum latency
- `max`: Maximum latency
- `p50`: 50th percentile (median)
- `p90`: 90th percentile
- `p95`: 95th percentile
- `p99`: 99th percentile

## API Design

### Measurement Specification
```yaml
measurement:
  id: "server_errors"
  name: "Server Lookup Errors"
  component: "server"           # Simple component name
  methods: ["Lookup", "Insert"] # Methods to track
  resultValue: "false"          # What counts as error
  metric: "count"               # Base metric type
  aggregation: "rate"           # How to aggregate
  window: "10s"                 # Time window
```

### Common Patterns

```yaml
# Throughput
- id: "server_throughput"
  component: "server"
  methods: ["Lookup"]
  resultValue: "*"      # All results
  metric: "count"
  aggregation: "rate"
  window: "10s"

# Error Rate
- id: "db_errors"
  component: "database"
  methods: ["Query", "Insert"]
  resultValue: "false"
  metric: "count"
  aggregation: "rate"
  window: "30s"

# Latency P95
- id: "server_latency_p95"
  component: "server"
  methods: ["Lookup"]
  resultValue: "*"
  metric: "latency"
  aggregation: "p95"
  window: "60s"

# Cache Hit Rate
- id: "cache_hits"
  component: "cache"
  methods: ["Get"]
  resultValue: "true"   # Hits return true
  metric: "count"
  aggregation: "rate"
  window: "30s"
```

## Implementation Plan

### Phase 1: Core Runtime Metrics ✅
- [ ] Create metrics_types.go with core types
- [ ] Implement CircularBuffer for data storage
- [ ] Create MetricStore with measurement management
- [ ] Add ProcessTraceEvent for event filtering
- [ ] Implement basic aggregations (sum, rate, avg)
- [ ] Add percentile calculations
- [ ] Integrate with SimpleEval
- [ ] Unit tests for all components

### Phase 2: Console API Layer
- [ ] Add metrics_api.go with REST handlers
- [ ] Create measurement CRUD endpoints
- [ ] Add data retrieval endpoint
- [ ] Wire up routes in canvas_web.go
- [ ] API integration tests

### Phase 3: Advanced Features
- [ ] WebSocket streaming for live updates
- [ ] Compound result filters (!=, in, ranges)
- [ ] Multi-component measurements
- [ ] Metric persistence to file
- [ ] Alerting thresholds

## Example Usage

```go
// Creating a measurement
spec := MeasurementSpec{
    ID:          "server_errors",
    Component:   "server",
    Methods:     []string{"Lookup"},
    ResultValue: "false",
    Metric:      MetricCount,
    Aggregation: AggRate,
    Window:      10 * time.Second,
}

store := runtime.GetMetricStore()
store.AddMeasurement(spec)

// Getting data
points := store.GetMeasurementData("server_errors", 100)
aggregated := store.GetAggregatedData("server_errors")
```

## Evolution from Original Vision

The original LIVE_METRICS_DESIGN.md envisioned a comprehensive monitoring system with:
- 6 categories of metrics (latency, throughput, utilization, flow, resource, business)
- Advanced visualizations (heatmaps, correlation plots, distribution histograms)
- Resource metrics (CPU, memory, I/O)
- Business metrics (SLA compliance, cost analysis)

### What We're Building Now

Based on SDL's current capabilities and immediate needs, we've focused on:
1. **Two fundamental metrics**: count and latency (directly from TraceEvent)
2. **Simple filtering**: component + method + result value matching
3. **In-memory storage**: No external dependencies
4. **Basic aggregations**: sum, rate, percentiles
5. **REST API**: Simple data access (WebSocket streaming deferred)

### What We're Deferring

1. **Utilization & Resource Metrics** - SDL doesn't track component capacity or OS resources
2. **Queue Metrics** - Components don't expose queue depth or wait times
3. **Advanced Visualizations** - Starting with simple time-series plots
4. **Business Metrics** - Focusing on technical metrics first
5. **Anomaly Detection** - Requires historical baselines
6. **Correlation Analysis** - Needs more sophisticated analytics

### Path to the Future

As SDL evolves, we could add:

1. **Phase 1 Extensions** (Near-term)
   - WebSocket streaming for live updates
   - Flow rate metrics from FlowEval
   - Compound result filters (!=, ranges)
   - Simple alerting thresholds

2. **Phase 2 Extensions** (Medium-term)
   - Component capacity modeling
   - Queue depth exposure from native components
   - Heatmap visualizations
   - Multi-system comparisons

3. **Phase 3 Extensions** (Long-term)
   - Resource utilization tracking
   - ML-based anomaly detection
   - Cost modeling integration
   - Predictive scaling recommendations

## Future Enhancements

1. **Pattern Matching**
   - Support for "4xx" status codes
   - Numeric comparisons ">= 400"
   - Set membership ["NotFound", "Forbidden"]

2. **Derived Metrics**
   - Error percentage (errors/total)
   - Success rate (1 - error_rate)
   - Relative metrics (cache_hits/total_requests)

3. **Advanced Aggregations**
   - Exponential moving average
   - Rate of change
   - Standard deviation

4. **System-wide Metrics**
   - Total system throughput
   - Component utilization estimates
   - Flow distribution analysis

## Progress Tracking

**Last Updated**: December 2024
**Current Status**: Phase 1 Core Runtime Metrics Complete ✅

### Completed Items:
- ✅ Created metrics_types.go with core types using virtual time (core.Duration)
- ✅ Implemented CircularBuffer for efficient in-memory storage
- ✅ Created MetricStore with system-aware measurement management
- ✅ Added ProcessTraceEvent with component instance resolution
- ✅ Implemented aggregation functions (sum, rate, avg, min, max, percentiles)
- ✅ Integrated with SimpleEval through ExecutionTracer
- ✅ Added comprehensive unit tests using real SDL systems

### Key Implementation Decisions:
1. **Virtual Time**: All timestamps use `core.Duration` starting from 0 for deterministic simulation
2. **Instance-based Tracking**: Components are tracked by instance name (e.g., "server") not type name
3. **Efficient Matching**: Direct pointer comparison between TraceEvent.Component and resolved ComponentInstance
4. **System Integration**: MetricStore is system-aware and resolves component names from the environment

### Next Steps:
- Phase 2: Console API Layer
  - Add metrics_api.go with REST handlers
  - Create measurement CRUD endpoints
  - Add data retrieval endpoint
  - Wire up routes in canvas_web.go

## Virtual Time Design

The metrics system uses virtual simulation time (`core.Duration`) instead of wall clock time:

- **Consistency**: All components use the same time type as SimpleEval's currTime
- **Determinism**: Tests produce identical results regardless of execution speed
- **Simplicity**: No timezone or clock synchronization issues
- **Integration**: Natural fit with SDL's simulation model

Time starts at 0 when a simulation begins and advances based on simulated operations.

## Component Instance Resolution

When creating measurements, users specify component instance names as they appear in the system definition:

```yaml
system ContactsSystem {
    use server ContactAppServer(...)  # "server" is the instance name
    use database ContactDatabase(...)  # "database" is the instance name
}
```

The MetricStore resolves these names to actual ComponentInstance pointers:
1. When adding a measurement, look up the component in the system's environment
2. Store the resolved ComponentInstance pointer in the measurement
3. During event processing, use direct pointer comparison for efficient matching

This approach provides intuitive naming while maintaining high performance.