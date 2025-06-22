# Traffic Generator Design

## Overview
The SDL traffic generator system is designed to simulate realistic load patterns with support for both low and high-throughput scenarios. Generators execute actual SDL code through SimpleEval, ensuring realistic performance characteristics.

## Key Design Decisions

### 1. Dual Execution Strategies
- **Low Rate (<100 RPS)**: Simple ticker-based execution with precise timing
- **High Rate (≥100 RPS)**: Batched execution with bounded concurrency

### 2. Virtual Time Management
Each generator maintains its own virtual clock:
- Starts at 0
- Advances by 1/rate seconds per event
- Ensures deterministic results independent of wall-clock time
- Passed to SimpleEval for consistent timestamps

### 3. Integration with SimpleEval
```go
// Each execution creates a fresh evaluator
eval := sdlruntime.NewSimpleEval(g.System.File, g.canvas.metricTracer)

// New environment for isolation
env := g.System.Env.Push()

// Execute with virtual time
result, _ := eval.Eval(callExpr, env, &currTime)
```

### 4. Fractional Rate Handling
For rates that don't divide evenly into batch intervals:
- Accumulate fractional events across batches
- Convert to integer batch sizes
- Ensures accurate long-term rates

### 5. Lifecycle Management
- Atomic state management with `atomic.Bool`
- Channel-based stop signaling
- Proper cleanup on generator stop
- Concurrent-safe start/stop operations

## Performance Characteristics

### Low Rate Mode
- Precise timing with individual timers
- Minimal overhead
- Suitable for debugging and low-volume testing
- Logs every 20 iterations

### High Rate Mode  
- 10ms batch intervals
- Bounded concurrency (2 * CPU cores)
- Pre-calculated virtual times per batch
- Logs every 100 batches (1 second)
- Tested at 1000+ RPS

## Example Usage
```go
// Create generator
gen := &GeneratorInfo{
    Generator: &protos.Generator{
        Id:        "gen1",
        Component: "server",
        Method:    "HandleRequest",
        Rate:      1000, // 1000 RPS
        Enabled:   false,
    },
    canvas: canvas,
    System: activeSystem,
}

// Start generation
gen.Start()

// Stop generation
gen.Stop()
```

## Future Enhancements
1. **Rate Ramping**: Gradual increase/decrease of rates
2. **Burst Patterns**: Configurable traffic spikes
3. **Distribution Models**: Poisson, normal, custom distributions
4. **Coordinated Generators**: Multiple generators with dependencies
5. **Load Profiles**: Time-based rate changes