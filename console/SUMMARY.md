# Console Package Summary

## Streaming Metrics & Utilization Tracking Implementation (June 2025)

This document summarizes the major streaming metrics and utilization tracking implementation completed in the console package.

### Architecture Overview

The console package provides real-time metric streaming capabilities and comprehensive utilization tracking through gRPC server-side streaming. The package serves as the core gRPC service layer implementing the CanvasService protocol with full support for performance monitoring and capacity analysis.

### Key Components

#### 1. Metric Collection & Pre-Aggregation (`console/metrics.go`)
- **MetricSpec**: Collects events and applies pre-aggregation in time windows
- **Multiple Metric Types**: Support for count, latency, and utilization metrics with different collection strategies
- **Time Window Aggregation**: Events collected within configurable time windows (e.g., 1s) and aggregated before storage
- **Periodic Sampling**: Utilization metrics use periodic sampling instead of event-based collection
- **Simulation Time**: Uses simulation time from Canvas for consistent timestamps across high-frequency events
- **Aggregation Functions**: Supports sum, avg, min, max, count, p50, p90, p95, p99

#### 2. Metric Storage (`console/metricstore.go`, `console/ringbufferstore.go`)
- **MetricStore Interface**: Defines storage and subscription contract with Subscribe method
- **RingBufferStore**: In-memory implementation with subscription polling
- **Real-time Updates**: Subscribe method enables metric updates via channels

#### 3. Utilization Tracking (`console/service.go`)
- **GetUtilization RPC**: Component utilization queries with hierarchical support
- **Component Resolution**: Handles nested component paths for utilization queries
- **UtilizationInfo Protocol**: Complete proto definitions for utilization data

#### 4. gRPC Service Layer (`console/service.go`)
- **StreamMetrics RPC**: Server-side streaming for multiple metrics in single connection
- **Multiple Metric IDs**: Handles multiple metrics efficiently to avoid N connections
- **MetricUpdateBatch**: Efficient batched updates for real-time streaming

#### 5. Enhanced CLI Commands
- **Metrics Commands** (`cmd/sdl/commands/metrics.go`): Rich metrics list with utilization type support
- **Utilization Command** (`cmd/sdl/commands/utilization.go`): Color-coded utilization display with bottleneck identification
- **Float Rate Support**: Generator commands accept decimal rates for fine-grained control

### Technical Implementation

#### Pre-Aggregation Process
```
Events → Time Windows → Aggregation Functions → MetricStore → Streaming
```

1. **Event Collection**: MetricSpec collects events in time windows (e.g., 1s intervals)
2. **Pre-Aggregation**: Applies configured aggregation function to collected events  
3. **Storage**: Stores single aggregated value per time window
4. **Streaming**: Real-time updates pushed to subscribers

#### Streaming Flow
```
gRPC Client → StreamMetrics → MetricStore.Subscribe → Channel → Real-time Updates
```

### Key Benefits

1. **Real-time Updates**: No polling overhead, immediate metric updates
2. **Efficient Network Usage**: Single connection for multiple metrics  
3. **Pre-aggregated Data**: Metrics show meaningful aggregated values (e.g., 100 RPS → 100.000/second)
4. **Consistent Timestamps**: Simulation time prevents timestamp collisions in high-frequency scenarios
5. **Rich CLI Display**: Enhanced metrics list with comprehensive statistics

#### Utilization Tracking Process
```
Component.GetUtilization() → UtilizationInfo → GetUtilization RPC → CLI/Dashboard
```

1. **Component Utilization**: Each queueing component calculates ρ = λ/(μ×c)
2. **Hierarchical Reporting**: Components delegate to children via UtilizationProvider interface
3. **Bottleneck Identification**: Automatic identification of performance bottlenecks
4. **Real-time Sampling**: Periodic collection for utilization metrics

### Current Status

- ✅ Complete streaming metrics infrastructure
- ✅ Pre-aggregation with time windows
- ✅ Multi-metric streaming support
- ✅ Utilization tracking for all queueing components
- ✅ Hierarchical utilization reporting
- ✅ Performance cliff visualization
- ✅ Dashboard integration
- ✅ Enhanced CLI commands  
- ✅ Simulation time tracking
- ✅ Float rate support for generators

### Testing

Run end-to-end metrics validation:
```bash
./test_metrics_e2e.sh
```

Test specific components:
```bash
go test -v ./console -run TestGenerator
go test -v ./console -run TestMetric
```

### Related Packages

- **`cmd/sdl`**: CLI commands use gRPC client to interact with CanvasService
- **`web/`**: Frontend dashboard consumes streaming metrics via Connect-Web
- **`protos/`**: Protocol definitions for gRPC service contract
