# Generator Design and Implementation

## Overview

The generator system provides high-throughput traffic generation for SDL simulations using virtual time and proper eval integration.

## Key Components

### 1. GeneratorInfo Structure
```go
type GeneratorInfo struct {
    *protos.Generator                      // Proto definition
    canvas                    *Canvas      // Parent canvas reference
    System                    *runtime.SystemInstance
    resolvedComponentInstance *runtime.ComponentInstance  
    resolvedMethodDecl        *runtime.MethodDecl
    
    // Virtual time management
    nextVirtualTime core.Duration
    timeMutex       sync.Mutex
    
    // For fractional rate handling
    eventAccumulator float64
    
    GenFunc func(iter int)
}
```

### 2. Virtual Time Implementation

Each generator maintains its own virtual clock that advances deterministically:

```go
func (g *GeneratorInfo) getNextVirtualTime() core.Duration {
    g.timeMutex.Lock()
    defer g.timeMutex.Unlock()
    
    current := g.nextVirtualTime
    // Advance by 1/rate seconds
    g.nextVirtualTime += core.Duration(1.0 / float64(g.Rate))
    return current
}
```

### 3. GenFunc Implementation

The GenFunc creates proper eval contexts with virtual time:

```go
func (g *GeneratorInfo) initializeGenFunc() {
    canvas := g.canvas
    
    g.GenFunc = func(iter int) {
        // Get next virtual time slot
        virtualTime := g.getNextVirtualTime()
        
        // Get tracer from canvas (may be nil)
        var tracer runtime.Tracer
        if canvas != nil && canvas.metricTracer != nil {
            tracer = canvas.metricTracer
        }
        
        // Create evaluator with tracer
        eval := runtime.NewSimpleEval(g.System.File, tracer)
        
        // New environment for isolation
        env := g.System.Env.Push()
        currTime := virtualTime
        
        // Build method call expression
        callExpr := &decl.CallExpr{
            Function: &decl.MemberAccessExpr{
                Receiver: &decl.IdentifierExpr{Value: g.Component},
                Member:   &decl.IdentifierExpr{Value: g.Method},
            },
        }
        
        // Execute with virtual time
        eval.Eval(callExpr, env, &currTime)
    }
}
```

## Design Decisions

### 1. Virtual Time vs Real Time
- **Decision**: Use virtual time for all simulations
- **Rationale**: Ensures deterministic results and allows time-travel debugging
- **Implementation**: Each generator maintains independent virtual clock

### 2. Exit-Only Event Processing
- **Decision**: MetricTracer only processes Exit events
- **Rationale**: Exit events contain duration information needed for metrics
- **Implementation**: Early return in Enter() method

### 3. Buffered Channels
- **Decision**: Use 1000-element buffer for event channels
- **Rationale**: Prevents blocking at high QPS while bounding memory
- **Implementation**: `make(chan *runtime.TraceEvent, 1000)`

### 4. Component Resolution
- **Decision**: Resolve components once during AddGenerator
- **Rationale**: Avoid repeated lookups during high-frequency execution
- **Implementation**: Store resolvedComponentInstance pointer

## Performance Characteristics

### Current Implementation (Tested)
- **Rate**: 2-10 RPS per generator
- **Latency**: Sub-millisecond event processing
- **Memory**: Bounded by channel buffers
- **CPU**: One goroutine per generator

### Target Scale (TODO)
- **Rate**: 10k-100k QPS aggregate
- **Strategy**: Batched execution with eval pooling
- **Concurrency**: Bounded worker pool
- **Memory**: Ring buffer for metrics

## Integration Points

### 1. Canvas Integration
- Canvas creates MetricTracer on Use()
- Generators reference canvas.metricTracer
- Clean lifecycle management

### 2. Flow Analysis
- Generators trigger flow recomputation on add/remove
- Future: Auto-set rates from flow analysis
- Manual rate overrides supported

### 3. Metrics Collection
- All Exit events routed to MetricSpecs
- Filtering by component/method
- Background aggregation

## Testing Results

With ContactsSystem at 2 RPS:
- Consistent 500ms intervals between calls
- Proper virtual time advancement
- Realistic latency variation (5ms cache hit, 15ms cache miss)
- MetricTracer successfully captures all events

## Future Optimizations

### 1. Batched Execution
```go
// Process multiple virtual time slots per tick
batchSize := int(rate * tickInterval)
for i := 0; i < batchSize; i++ {
    go executeWithPooledEval(virtualTime + i*interval)
}
```

### 2. Eval Pooling
```go
evalPool := sync.Pool{
    New: func() interface{} {
        return runtime.NewSimpleEval(system.File, tracer)
    },
}
```

### 3. Fractional Rate Handling
```go
// Already implemented - accumulator pattern
g.eventAccumulator += eventsPerTick
wholeEvents := int(g.eventAccumulator)
g.eventAccumulator -= float64(wholeEvents)
```

This design provides a solid foundation for high-throughput simulation while maintaining accuracy and determinism.