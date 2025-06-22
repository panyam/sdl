# New Metrics Architecture

## Overview

The metrics system has been redesigned with a clean separation between the runtime execution engine and measurement concerns. The new architecture introduces a `Tracer` interface that allows pluggable implementations for different kinds of execution tracking.

## Key Components

### 1. Tracer Interface (runtime/simpleeval.go)

```go
type Tracer interface {
    Enter(ts core.Duration, kind TraceEventKind, comp *ComponentInstance, method *MethodDecl, args ...string) int64
    Exit(ts core.Duration, duration core.Duration, comp *ComponentInstance, method *MethodDecl, retVal Value, err error)
    PushParentID(id int64)
    PopParent()
}
```

The interface is intentionally minimal with just 4 methods:
- `Enter`: Called when entering a method, returns a trace ID
- `Exit`: Called when exiting a method with results and duration
- `PushParentID/PopParent`: Stack-based parent tracking for nested calls

### 2. MetricTracer Implementation (console/metrictracer.go)

The `MetricTracer` is a concrete implementation that:
- Only processes `Exit` events (where actual results and duration are available)
- Thread-safe with read/write locks
- Manages multiple `MetricSpec` configurations
- Tied to a specific `SystemInstance` (created when Canvas.Use() is called)

```go
type MetricTracer struct {
    seriesLock sync.RWMutex
    seriesMap  map[string]*MetricSpec
    system     *runtime.SystemInstance
}
```

### 3. MetricSpec (console/metrics.go)

```go
type MetricSpec struct {
    *protos.Metric  // Embedded proto definition
    System          *runtime.SystemInstance
    Matcher         ResultMatcher
    resolvedComponentInstance *runtime.ComponentInstance
    eventChan       chan *runtime.TraceEvent
}
```

Key features:
- Embeds the proto definition directly for API consistency
- Runs in separate goroutine for non-blocking metric collection
- Uses channels for real-time event processing
- Supports result value matching via `ResultMatcher` interface

### 4. Integration Flow

1. **Canvas Creation**: Each canvas manages its own metrics
2. **System Selection**: When `Use(systemName)` is called:
   - Old MetricTracer is cleared
   - New MetricTracer is created for the selected system
   - MetricTracer is passed to SimpleEval as the tracer
3. **Metric Addition**: Via gRPC `AddMetric` call
4. **Event Flow**: 
   - SimpleEval calls tracer.Exit()
   - MetricTracer checks all MetricSpecs for matches
   - Matching specs receive events via channels
   - MetricPoints are recorded asynchronously

## Benefits of the New Architecture

### 1. Clean Separation of Concerns
- Runtime remains measurement-agnostic
- Console handles all metric-specific logic
- Easy to add new tracer implementations

### 2. Proto Integration
- MetricSpec embeds `*protos.Metric` directly
- No translation layers needed
- Consistent data model across API boundaries

### 3. Performance
- Minimal overhead on execution hot path
- Asynchronous metric processing
- Channel-based architecture prevents blocking

### 4. Extensibility
- Multiple tracers can be composed
- New measurement strategies don't require runtime changes
- Support for different metric types (count, latency, custom)

### 5. Real-time Capabilities
- Channel-based processing enables live metrics
- Natural support for streaming APIs
- Built-in backpressure handling

## Implementation Notes

### System-Specific MetricTracer
The MetricTracer is created per-system because:
- Component instances are system-specific
- Allows clean metric isolation between systems
- Simplifies component resolution and matching

### Generator Integration
Generators now resolve their target component when added:
- Validates component and method exist
- Stores resolved component instance for efficient execution
- Enables proper flow analysis integration

### Thread Safety
- MetricTracer uses RWMutex for concurrent access
- Each MetricSpec runs in its own goroutine
- Event channels provide natural synchronization

## Current Limitations

1. **Metric Types**: Only `latency` and `count` types are implemented. No support for `utilization`, `throughput`, or resource usage metrics.
2. **Manual Arrival Rates**: ResourcePool queueing simulation requires manual setting of arrival rates - generators don't automatically update them.
3. **Aggregation at Collection**: Metrics are pre-aggregated when collected, not at query time.

## Future Considerations

1. **Metric Persistence**: Currently in-memory only
2. **Additional Metric Types**: Utilization, throughput, resource usage
3. **Automatic Arrival Rate Updates**: Generators could update component arrival rates
4. **Composite Tracers**: Ability to chain multiple tracers
5. **Performance Profiling**: Dedicated tracer for performance analysis