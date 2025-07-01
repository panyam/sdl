# SDL (System Definition Language) Project Summary

SDL is a language and runtime for modeling and simulating distributed system performance. It allows engineers to define system architectures, model component behaviors, and simulate traffic patterns to understand system capacity and performance characteristics.

## Core Components

### 1. **SDL Language**
- Domain-specific language for defining systems, components, and their interactions
- Supports latency modeling, capacity constraints, error rates, and traffic distribution
- Import system for modular definitions
- Not a real programming language - focused on performance modeling

### 2. **Parser & Compiler**
- YACC-based parser (`parser/`)
- AST representation (`decl/`)
- Type system and validation
- Import resolution and file loading

### 3. **Runtime Engine**
- **SimpleEval**: Event-driven simulation engine
- **Flow Analysis**: Automatic rate propagation through system
- **Virtual Time**: Deterministic simulation with consistent time tracking
- **Metrics Collection**: Asynchronous event processing and aggregation

### 4. **Console Server**
- gRPC API with HTTP gateway (grpc-gateway)
- Canvas-based session management
- Generator lifecycle (start/stop/pause/resume)
- Real-time metrics with MetricStore architecture
- Web dashboard for visualization

### 5. **CLI Interface**
- Canvas operations (load, use, reset)
- Generator management (add, update, remove)
- Metrics configuration
- Flow evaluation and automatic rate calculation

## Recent Architectural Changes (June 2025)

### gRPC Migration
- Migrated from REST to gRPC with HTTP gateway
- Proto definitions in `protos/sdl/v1/`
- Dual protocol support (gRPC :9090, REST :8080)
- All CLI commands now use gRPC

### Metrics Architecture
- New Tracer interface for pluggable tracing
- MetricTracer implementation for async processing
- RingBufferStore for memory-efficient storage
- Pre-aggregated metrics (count, sum, avg, min, max, percentiles)

### Flow Evaluation
- Runtime-based flow analysis (replacing string-based)
- Automatic downstream rate calculation
- --apply-flows flag on generator commands
- EvaluateFlows and BatchSetParameters RPCs

## WASM Browser Support (June 2025)

### Motivation
- Server-side simulation is CPU/memory intensive and expensive
- Enable free demos without server costs
- Run simulations locally in user's browser
- Instant feedback for learning and experimentation

### Architecture Decision
- Single WASM bundle approach (parser + evaluator together)
- Reuse existing Go codebase instead of rewriting
- FileSystem abstraction for virtual files in browser
- Mirror CLI commands in JavaScript API

### Implementation Strategy
1. **FileSystem Interface** in `loader/` package
   - Shared between server and WASM modes
   - Multiple backends: Local, HTTP, Memory, GitHub
   - Composite pattern for mounting different sources

2. **Canvas Refactoring**
   - Remove proto/gRPC dependencies from core Canvas
   - Create native types for Generator, Metric, SystemDiagram
   - Proto conversion only at service boundaries
   - Single implementation for both server and WASM

3. **Web Integration**
   - Extend existing dashboard with WASM mode
   - Reuse dockview panels, generators, metrics UI
   - Add file explorer and code editor for WASM
   - Toggle between server/WASM modes

### Current WASM Status
- ✅ Basic WASM build structure created
- ✅ FileSystem abstraction implemented and cleaned up
- ✅ Canvas refactoring completed - removed proto dependencies from core Canvas
- ✅ WASM runtime compatibility achieved - 28.6MB binary successfully built
- ✅ Web dashboard unified layout implemented
- ✅ FileClient interface adopted for both server and WASM modes
- ✅ Panel display issue resolved
- ✅ Recipe execution in WASM mode implemented
- ✅ Singleton recipe controls integrated into toolbar
- ⏳ Binary size optimization pending (current: 28.6MB)

## Performance Characteristics

### Traffic Generation
- Simple ticker for <100 RPS
- Batched execution for higher rates (1000+ RPS)
- Virtual time consistency across generators
- Fractional rate handling with accumulator pattern

### Metrics Processing
- 100ms batch processing window
- Circular buffer with configurable retention
- Standard aggregations computed at write time
- Connect-Web streaming for real-time updates

## Testing
- Comprehensive unit tests for runtime
- Flow evaluation tests
- End-to-end metrics validation
- Generator synchronization tests

## Known Limitations
- Only supports latency and count metrics (no utilization/throughput)
- Control flow dependencies not fully represented in path analysis
- No binary operators in SDL (use native functions instead)
- WASM build currently blocked by proto dependencies