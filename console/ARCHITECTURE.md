# Console Architecture: gRPC Migration and Design Decisions

## Overview

The console package has been migrated from a REST-based architecture to a gRPC-first design with HTTP gateway support. This document captures the key architectural decisions, implementation details, and current status.

## Architecture Layers

### 1. Protocol Buffers (protos/sdl/v1/)
- **canvas.proto**: Defines all Canvas service RPCs and messages
- **models.proto**: Shared data models (Canvas, Generator, Metric)
- **buf.yaml**: Configuration for buf code generation

### 2. gRPC Service Implementation (console/)
- **grpcserver.go**: gRPC server setup with reflection
- **service.go**: CanvasService implementation of all RPCs
- **webserver.go**: HTTP gateway using grpc-gateway

### 3. Canvas Core (console/)
- **canvas.go**: Core Canvas state management and business logic
- **generator.go**: Traffic generator implementation
- **metrictracer.go**: Performance measurement via Tracer interface

## Key Design Decisions

### 1. Tracer Interface Architecture
Instead of tightly coupling metrics to the runtime, we introduced a clean Tracer interface:

```go
type Tracer interface {
    Enter(ts Duration, kind TraceEventKind, comp *ComponentInstance, method *MethodDecl, args ...string) int64
    Exit(ts Duration, duration Duration, comp *ComponentInstance, method *MethodDecl, retVal Value, err error)
    PushParentID(id int64)
    PopParent()
}
```

**Benefits:**
- Runtime remains measurement-agnostic
- Multiple tracer implementations possible (ExecutionTracer, MetricTracer)
- Clean separation of concerns
- Easy to add new measurement strategies

### 2. System-Specific MetricTracer
The MetricTracer is created per-system when Canvas.Use() is called:

```go
func (c *Canvas) Use(systemName string) error {
    // ... load system ...
    c.metricTracer = NewMetricTracer(c.activeSystem)
}
```

**Design Rationale:**
- Each system has its own components and methods
- Tracer can efficiently filter events for registered metrics
- Clean lifecycle management

### 3. Channel-Based Async Processing
MetricTracer uses channels for asynchronous event processing:

```go
type MetricTracer struct {
    system       *runtime.SystemInstance
    metricSpecs  map[string]*MetricSpec
    metricChan   chan *MetricPoint
    resultsChan  chan *MetricResult
}
```

**Benefits:**
- Non-blocking trace event processing
- Decoupled collection from aggregation
- Natural backpressure handling
- Easy to scale processing

### 4. gRPC with HTTP Gateway
Dual protocol support via grpc-gateway:

```
Client → HTTP/REST (8080) → grpc-gateway → gRPC (9090) → Service
Client → gRPC (9090) → Service
```

**Advantages:**
- RESTful API for web clients
- Native gRPC for CLI/SDK clients
- Single service implementation
- Auto-generated OpenAPI docs

## Current Implementation Status

### ✅ Completed
1. **Proto Definitions**: Complete service and message definitions
2. **gRPC Server**: Full implementation with reflection
3. **HTTP Gateway**: RESTful API via grpc-gateway
4. **Canvas Service**: All core operations implemented
5. **CLI Migration**: All commands use gRPC client
6. **Trace Command**: New ExecuteTrace RPC for debugging

### 🚧 In Progress
1. **Generator Implementation**:
   - GenFunc needs proper eval integration
   - Virtual time tracking required
   - Batched execution for efficiency

2. **MetricStore Design**:
   - Abstract interface for pluggable storage
   - RingBuffer for real-time monitoring
   - DuckDB for historical analysis

### 📋 TODO
1. **Live Metrics Streaming**: Implement server-side streaming for real-time updates
2. **Parameter Management**: SetParameter/GetParameter RPCs
3. **Flow Analysis Integration**: Connect generators to flow rate calculations
4. **Dashboard Migration**: Update web UI to use gRPC gateway

## Generator Design (In Progress)

### Virtual Time Management
Each generator maintains its own virtual clock:
```go
type GeneratorInfo struct {
    *protos.Generator
    nextVirtualTime core.Duration
    timeMutex      sync.Mutex
}
```

### Batched Execution Strategy
For efficiency at high QPS (10k-100k):
```go
// Instead of: 1 goroutine per eval
// Use: Batched execution with bounded concurrency
batchSize := calculateBatchSize(rate, interval)
evalPool := make(chan *SimpleEval, numCPU * 2)
```

### MetricStore Interface (Proposed)
```go
type MetricStore interface {
    Write(point MetricPoint) error
    WriteBatch(points []MetricPoint) error
    Query(timeRange TimeRange, filters ...Filter) ([]MetricPoint, error)
    Aggregate(timeRange TimeRange, window Duration, agg AggregateFunc) ([]AggregateResult, error)
}
```

**Implementations:**
- **RingBufferStore**: Fixed memory, O(1) writes, real-time focus
- **DuckDBStore**: Columnar storage, complex analytics, persistence
- **HybridStore**: Hot/cold architecture combining both

## Migration Guide

### For CLI Users
```bash
# Old: Complex REPL commands
SDL> load file.sdl
SDL> use System

# New: Direct CLI commands
sdl load file.sdl
sdl use System
sdl trace server.Lookup -o trace.json
```

### For API Clients
```go
// Old: REST client
resp, err := http.Post("/api/canvas/load", ...)

// New: gRPC client
client := v1.NewCanvasServiceClient(conn)
resp, err := client.LoadFile(ctx, &v1.LoadFileRequest{...})
```

## Next Steps

1. **Complete Generator Implementation**
   - Integrate SimpleEval for actual method execution
   - Add virtual time propagation
   - Implement batched execution for scale

2. **Design MetricStore**
   - Define clean interface boundaries
   - Implement RingBufferStore first
   - Add DuckDB integration later

3. **Update MetricTracer**
   - Use MetricStore for persistence
   - Add aggregation windows
   - Implement result streaming

4. **Dashboard Integration**
   - Migrate to gRPC gateway endpoints
   - Use streaming for live updates
   - Update generator controls

This architecture provides a solid foundation for scaling SDL to handle high-throughput simulations while maintaining clean separation of concerns and extensibility.