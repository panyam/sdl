# Architecture Update - June 2025

## Major Architectural Changes

### 1. Migration to gRPC-based API

The project has migrated from a hand-coded HTTP server to a gRPC-based architecture with automatic HTTP gateway generation.

#### Service Definition
- **Proto files**: `protos/sdl/v1/canvas.proto` and `models.proto`
- **Generated code**: Auto-generated in `./gen/` folder
- **API styles supported**:
  - Native gRPC (port 9090)
  - REST via grpc-gateway (port 8080)
  - Legacy REST endpoints (being phased out)

#### Key Services
- **CanvasService**: Main service for canvas operations
  - Canvas CRUD operations
  - File loading and system selection
  - Generator management
  - Metric management

### 2. Console Package Restructuring

The console package has been completely refactored:

```
console/
├── service.go         # gRPC CanvasService implementation
├── grpcserver.go      # gRPC server setup
├── webserver.go       # HTTP gateway server
├── api.go             # Gateway configuration
├── canvasinfo.go      # Canvas domain logic
├── generator.go       # Generator management
├── metrictracer.go    # Metric collection via Tracer interface
└── metrics.go         # MetricSpec definitions
```

Key improvements:
- Thread-safe canvas storage with proper locking
- Clean separation between API layer and domain logic
- Proto-embedded structs for consistency

### 3. Tracer Interface Architecture

Introduced a pluggable `Tracer` interface in the runtime:

```go
type Tracer interface {
    Enter(ts core.Duration, kind TraceEventKind, comp *ComponentInstance, method *MethodDecl, args ...string) int64
    Exit(ts core.Duration, duration core.Duration, comp *ComponentInstance, method *MethodDecl, retVal Value, err error)
    PushParentID(id int64)
    PopParent()
}
```

Benefits:
- Runtime remains measurement-agnostic
- Multiple tracer implementations possible
- Clean plugin architecture

### 4. Metrics System Redesign

Metrics have been moved from runtime to console package:
- **MetricTracer**: Implements Tracer interface
- **MetricSpec**: Configuration with proto embedding
- **Channel-based**: Asynchronous event processing
- **System-specific**: New tracer per system selection

### 5. CLI Migration to gRPC

All CLI commands now use gRPC client directly:
- Created `withCanvasClient` helper for connection management
- Added `--canvas` flag for canvas selection (default: "default")
- Renamed `measure` commands to `metric` for consistency
- Better error handling and connection guidance

### 6. Server Architecture

New server structure with clean separation:
- **App**: Manages multiple servers (gRPC + HTTP)
- **GrpcServer**: Native gRPC service
- **WebAppServer**: HTTP gateway with CORS support
- Air configuration for hot reloading during development

## Implementation Status

### Completed ✅
- gRPC service definitions and proto files
- Basic Canvas operations (Create, Load, Use, Get)
- CLI migration to gRPC
- Tracer interface and MetricTracer implementation
- Generator management structure
- Server/client separation

### In Progress 🔄
- Generator start/stop/pause/resume implementation
- Live metric streaming
- Canvas proto field mapping enhancements
- Flow analysis integration

### TODO 📋
- WebSocket support for real-time updates
- Metric aggregation and windowing
- Dashboard UI updates for new API
- Performance profiling tracer
- Distributed tracing support

## Breaking Changes

1. **API Endpoints**: Old REST endpoints replaced with gRPC/REST gateway
2. **CLI Commands**: Now require server to be running
3. **Metric System**: Complete redesign with new data flow
4. **Canvas State**: Now server-managed instead of client-side

## Migration Notes

For existing code:
1. Update API calls to use new gRPC client or REST endpoints
2. Ensure server is running before using CLI commands
3. Update metric collection to use new AddMetric API
4. Canvas operations now require explicit canvas ID