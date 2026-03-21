# Services Package Summary

## Overview

The services package provides the core business logic for SDL (System Detail Language) canvas management. It handles canvas state, generator management, metrics tracking, flow evaluation, and system diagram generation.

## Architecture

### Package Structure (December 2024)

The package has been reorganized to support both server-side and WASM modes:

```
services/
├── canvas.go               # Core Canvas struct and operations
├── canvas_conversions.go   # Canvas proto conversion helpers
├── canvas_view_presenter.go # MVP Presenter for WASM (js/wasm build only)
├── conversions.go          # Proto<->Native type conversions (exported)
├── generator.go            # Generator management logic
├── logger.go               # Logging utilities
├── metrics.go              # Metric collection and aggregation
├── metricstore.go          # Metric storage interface
├── metrictracer.go         # Metric tracing implementation
├── ringbufferstore.go      # In-memory ring buffer storage
├── service.go              # CanvasService gRPC implementation
├── system_catalog.go       # System discovery and catalog
├── systems_service.go      # SystemsService implementation
├── types.go                # Shared type definitions
└── singleton/              # WASM singleton implementations
    ├── canvas_service.go   # SingletonCanvasService for WASM
    └── systems_service.go  # SingletonSystemsService for WASM
```

### Proto Organization

Protos are split into models and services following weewar conventions:

```
protos/sdl/v1/
├── models/
│   ├── models.proto           # Core data types (Canvas, Generator, Metric, etc.)
│   ├── canvas_service.proto   # Request/Response messages for CanvasService
│   ├── presenter.proto        # Presenter initialization messages
│   ├── dashboardpage.proto    # Browser callback messages
│   └── systems.proto          # SystemInfo and related types
└── services/
    ├── canvas.proto           # CanvasService RPC definitions
    ├── presenter.proto        # CanvasViewPresenter and SingletonInitializer services
    ├── dashboardpage.proto    # CanvasDashboardPage browser service (browser_provided)
    └── systems.proto          # SystemsService RPC definitions
```

### WASM Architecture (MVP Pattern)

For WASM mode, we use Model-View-Presenter:

1. **Model**: `SingletonCanvasService` - manages canvas state in-memory
2. **View**: Browser dashboard (implements `CanvasDashboardPage` interface)
3. **Presenter**: `CanvasViewPresenter` - orchestrates between views and services

The presenter receives user interactions from the browser, calls service methods, and pushes updates back to the browser via the `CanvasDashboardPage` client.

### Key Components

#### CanvasService (Server Mode)
- Multi-tenant canvas management
- gRPC/Connect service implementation
- Persistent storage support

#### SingletonCanvasService (WASM Mode)
- Single in-memory canvas
- No gRPC dependencies (uses generated WASM interfaces)
- Direct method calls from JavaScript

#### CanvasViewPresenter (WASM Mode)
- Handles UI logic and orchestration
- Receives user interactions (FileSelected, AddGenerator, etc.)
- Pushes updates to browser (UpdateDiagram, UpdateFlowRates, etc.)

### Conversion Functions

All proto<->native conversion functions are now exported for use by singleton package:

- `ToProtoGenerator` / `FromProtoGenerator`
- `ToProtoMetric` / `FromProtoMetric`
- `ToProtoSystemDiagram` / `FromProtoSystemDiagram`
- `ToProtoCanvas` / `FromProtoCanvas`

## Streaming Metrics & Utilization Tracking

### Metric Collection & Pre-Aggregation
- **MetricSpec**: Collects events and applies pre-aggregation in time windows
- **Multiple Metric Types**: count, latency, and utilization metrics
- **Time Window Aggregation**: Events collected within configurable time windows

### Streaming Flow
```
gRPC Client → StreamMetrics → MetricStore.Subscribe → Channel → Real-time Updates
```

For WASM, streaming uses callback-based pattern instead of gRPC streams.

## Current Status

- ✅ Services moved from web/services to services/
- ✅ Proto reorganization (models/services split)
- ✅ Singleton implementations for WASM
- ✅ CanvasViewPresenter MVP pattern
- ✅ Exported conversion functions
- ✅ Streaming metrics with proper WASM signatures
- 🔲 Browser adapter implementations
- 🔲 Frontend TypeScript client integration

## Testing

```bash
go test -v ./services -run TestGenerator
go test -v ./services -run TestMetric
```

## Related Packages

- **`cmd/wasm`**: WASM entry point using generated exports
- **`gen/wasm/go/sdl/v1/services`**: Generated WASM service interfaces
- **`web/`**: Frontend dashboard
- **`protos/`**: Protocol definitions

## Recent Changes (March 2026)

- `fsbe/canvas_service.go`: Switched storage import from `turnforge/turnengine/engine/storage` to `goutils/storage`
