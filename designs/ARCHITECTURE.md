# SDL Architecture Overview

## Core Design Principles

SDL is designed as a system modeling and simulation platform with:
- **gRPC-first architecture** for efficient client-server communication
- **Pluggable measurement strategies** via the Tracer interface
- **High-throughput traffic generation** with virtual time support
- **Flexible metric storage** with multiple backend options

## Key Architectural Decisions

### 1. gRPC-First Design
- **Decision**: Use gRPC as the primary protocol with HTTP gateway
- **Rationale**: Better performance, type safety, and streaming support
- **Implementation**: Proto files define the contract, grpc-gateway provides REST

### 2. Tracer Interface Pattern
- **Decision**: Abstract tracing behind a simple 4-method interface
- **Rationale**: Allows multiple measurement strategies without runtime changes
- **Implementation**: Enter/Exit for lifecycle, Push/Pop for context management

### 3. Exit-Only Metric Processing
- **Decision**: MetricTracer only processes Exit events
- **Rationale**: Exit events contain duration info needed for latency metrics
- **Implementation**: Enter() returns immediately with ID 0

### 4. Canvas State Management
- **Decision**: All state managed by Canvas, services are thin wrappers
- **Rationale**: Single source of truth, easier testing and debugging
- **Implementation**: CanvasService delegates all logic to Canvas methods

### 5. MetricStore Abstraction
- **Decision**: Separate metric storage from collection with pluggable backends
- **Rationale**: Different use cases need different storage strategies
- **Implementation**: MetricStore interface with RingBuffer, TimeSeries, etc.

### 6. Batched High-Throughput Execution
- **Decision**: Use different strategies for low vs high rate generators
- **Rationale**: Optimize for both accuracy (low rate) and throughput (high rate)
- **Implementation**: Simple ticker for <100 RPS, batched execution for higher rates

### 7. Virtual Time Management
- **Decision**: Each generator maintains its own virtual clock
- **Rationale**: Deterministic simulation results independent of wall-clock time
- **Implementation**: Virtual time advances by 1/rate seconds per event

### 8. Metric-Centric Storage Design
- **Decision**: MetricPoint doesn't embed Component/Method/MetricType
- **Rationale**: These are static properties of the Metric, not the data point
- **Implementation**: MetricPoint only contains timestamp and value

### 9. Pre-Aggregated Metrics Architecture
- **Decision**: Metrics are aggregated at collection time, not query time
- **Rationale**: Predictable performance, simplified query logic, bounded storage
- **Implementation**: MetricTracer calculates aggregations based on MetricSpec during event processing
- **Implementation**: MetricStore methods accept Metric reference + points

### 10. Parameter Management Design (June 21, 2025)
- **Decision**: Use SDL expression syntax for all parameter values
- **Rationale**: Consistent with the language, supports complex types, leverages existing parser
- **Implementation**: SetParameter accepts strings like "42", "true", "'string'" that are parsed as SDL expressions
- **Design**: Runtime.GetParam/SetParam handle nested component paths with SystemInstance.FindComponent

## Component Architecture

### Runtime Layer
- **SimpleEval**: AST-based interpreter for SDL execution
- **Tracer Interface**: Pluggable measurement and debugging
- **Virtual Time**: Deterministic simulation support
- **Flow Analysis**: Component dependency and rate calculation

### Console Layer
- **Canvas**: Core engine managing systems, generators, and metrics
- **CanvasService**: gRPC service wrapping Canvas operations
- **MetricTracer**: Asynchronous metric collection implementation
- **GeneratorInfo**: Traffic generation with batched execution
- **MetricStore**: Abstraction for metric persistence

### Protocol Layer
- **Proto Definitions**: Single source of truth for all APIs
- **gRPC Server**: Primary service endpoint on port 9090
- **HTTP Gateway**: REST compatibility on port 8080
- **WebSocket**: Real-time metric streaming

### CLI Layer
- **Direct Commands**: Shell-native interface (load, use, gen, etc.)
- **gRPC Client**: Consistent connection handling
- **Environment Support**: Canvas selection and server configuration

## Data Flow

1. **Command Reception**: CLI sends gRPC request to server
2. **Canvas Processing**: Request routed to appropriate Canvas method
3. **System Execution**: SimpleEval runs with attached Tracer
4. **Metric Collection**: Exit events processed by MetricTracer
5. **Storage**: MetricPoints written to MetricStore in batches
6. **Query**: Clients retrieve metrics via gRPC/REST APIs

## Performance Characteristics

- **Single Execution**: ~1-10ms latency with full tracing
- **Low Rate Generation**: <100 RPS with accurate timing
- **High Rate Generation**: 1000+ RPS with batched execution
- **Metric Processing**: Asynchronous with 100ms batch windows
- **Storage**: Memory-bounded with configurable retention

## Future Enhancements

- Additional MetricStore backends (TimescaleDB, InfluxDB)
- Server-side streaming for real-time metrics
- Distributed execution for even higher scale
- Advanced flow analysis with ML-based predictions