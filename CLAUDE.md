Read and learn all about this project by looking at all the SUMMARY.md and NEXTSTEPS.md.  

I am continuing with a previous project.  You will find the summaries in SUMMARY.md files located in the top level as various sub folders.  NEXTSTEPS.md is used to note what has been completed and what are next steps in our roadmap.
Thorougly understand it and give me a recap so we can continue where we left off.

## Major Architecture Update (June 2025)
The project has undergone significant architectural changes:

### 1. gRPC Migration
- **API Layer**: Migrated from REST to gRPC with HTTP gateway (grpc-gateway)
- **Proto definitions**: `protos/sdl/v1/canvas.proto` and `models.proto`
- **Dual protocol**: gRPC on :9090, REST gateway on :8080
- **Generated code**: Auto-generated in `./gen/` folder

### 2. New Metrics Architecture
- **Tracer Interface**: Clean 4-method interface in runtime for pluggable tracing
- **MetricTracer**: Implementation in console package, processes Exit events only
- **System-specific**: New MetricTracer created on each Canvas.Use() call
- **Channel-based**: Asynchronous event processing with thread-safety
- **Proto embedding**: MetricSpec embeds proto definition directly
- **Documentation**: See `console/NEWMETRICS.md` for details

### 3. Console Package Restructuring
- **service.go**: gRPC CanvasService implementation
- **grpcserver.go/webserver.go**: Server setup and management
- **canvasinfo.go**: Canvas domain logic with thread-safe storage
- **metrictracer.go**: Tracer interface implementation
- **generator.go**: Generator management with flow integration

### 4. CLI Migration
- **All commands use gRPC**: Direct client usage, no more REST calls
- **Helper pattern**: `withCanvasClient` for connection reuse
- **Canvas selection**: `--canvas` flag (default: "default")
- **Better errors**: Connection guidance when server is down

### 5. Current Status (Updated June 2025)
- **Working**: 
  - All Canvas operations via gRPC (load, use, info, etc.)
  - ExecuteTrace RPC for single execution debugging
  - Traffic generation with SimpleEval integration
  - Virtual time tracking for deterministic simulations
  - Batched execution for high-rate generators (1000+ RPS)
  - MetricStore architecture with RingBufferStore
  - Asynchronous metric processing with MetricTracer
  - Automatic flow calculation with --apply-flows flag
  - Generator rate updates via `sdl gen update`
  - Flow evaluation showing both current and calculated rates
- **Completed Recent Work**:
  - Generator lifecycle management (start/stop/pause/resume)
  - MetricStore interface with pluggable backends
  - Memory-efficient metrics with circular buffers
  - Standard aggregations (count, sum, avg, min, max, percentiles)
  - EvaluateFlows and BatchSetParameters RPCs for automatic flow calculation
  - GetFlowState RPC for retrieving current flow rates
  - Batch generator operations (StartAllGenerators, StopAllGenerators)
  - Fixed component name resolution for nested components in flow analysis
- **TODO**: 
  - Human-readable trace command output
  - Dashboard integration with new gRPC endpoints
  - Server-streaming for real-time metrics

## Traffic Generation & MetricStore Architecture (June 2025)
When working with traffic generation and metrics:
- **Generator Pattern**: Simple ticker for <100 RPS, batched execution for higher rates
- **Virtual Time**: Each generator maintains consistent virtual clock (1/rate seconds per event)
- **MetricStore Interface**: Clean abstraction with WritePoint/WriteBatch/Query/Aggregate
- **RingBufferStore**: Default implementation with configurable retention (e.g., 1 hour)
- **Async Processing**: MetricTracer writes to store in 100ms batches
- **Testing**: Use `go test -v ./console -run TestGenerator` for generator tests

## FlowEval Runtime Migration (June 2025)
When continuing work on FlowEval, note that we're in the middle of migrating from string-based to runtime-based flow analysis:
- **New code location**: runtime/flowrteval.go (runtime-based) replacing runtime/floweval.go (string-based)
- **Key types**: RateMap (runtime/ratemap.go), FlowScope (runtime/flowscope.go), GeneratorEntryPointRuntime
- **Architecture**: Uses actual ComponentInstance objects from SimpleEval, no duplicate instances
- **Pattern**: NWBase wrapper provides smart defaults for non-flow-analyzable components
- **Status**: Migration complete, Canvas integration working with runtime strategy
- **Test with**: `go test -v ./runtime -run "TestFlowEvalRuntime|TestSolveSystemFlowsRuntime"`

## Automatic Flow Calculation (June 2025)
- **--apply-flows flag**: Available on all generator commands (add, update, start, stop, remove)
- **Flow evaluation**: Automatically calculates downstream component rates based on generators
- **Batch application**: Uses BatchSetParameters RPC to atomically update all arrival rates
- **sdl gen update**: Efficiently updates generator rates without creating new goroutines
- **Demo scripts**: Updated to use --apply-flows instead of manual arrival rate settings

## Important Implementation Notes
- **Metrics**: Pre-aggregated at collection time, not query time
- **Bug Fix**: When removing metrics, use `delete(map, key)` not `map[key] = nil` to avoid panics
- **Testing**: Run `./test_metrics_e2e.sh` for comprehensive end-to-end validation
- **TraceAllPaths**: Static analysis tool that enumerates all possible execution paths without runtime execution
- **Path Analysis Limitation**: Control flow dependencies not fully represented - conditional branches shown as siblings
- **Self References**: Use pattern `self.component.method` for accessing component dependencies in SDL   
- **SDL Language Note**: SDL is not a real programming language. It is a language for modelling capacities and system performance. So methods in SDL will not need parameters. For example in a real language, an API method GetDriver would take a driverId parameter. In SDL GetDriver is just GetDriver(). It will still have output types to denote outcomes of result types.
- **SDL Binary Operators**: SDL does not have binary operators. Again - remember it is not a real programming language, it is simple so as to enable performance and error modelling. If you need extra behaviors you CAN suggest native functions to be implemented and called. Eg if you absolutely want binary arithmetic then suggest a "plus" NATIVE function.
- **Enums must be comma seperated in SDL**

## SDL System Declaration Notes
- In SDL system declaration you can declare the components in any order. There are no "set" statements. You pass the dependencies in the constructor of a "use" keyword.  For example:
```system Twitter {
    use app AppServer(db = database)
    use db Database
}```
- Here the AppServer component has a "db" dependency that is set by the "database" component declared in the next line.

Also be conservative on how many comments are you are adding or modifying unless it is absolutely necessary (for example a comment could be contradicting what is going on - in which case it is prudent to modify it).  
When modifying files just focus on areas where the change is required instead of diving into a full fledged refactor.
Make sure you ignore 'gen' and 'node_modules' as it has a lot of files you wont need for most things and are either auto generated or just package dependencies
When updating .md files and in commit messages use emojis and flowerly languages sparingly.  We dont want to be too grandios or overpromising.
Make sure the playwright tool is setup so you can inspect the browser when we are implementing and testing the Dashboard features.
Do not refer to claude or anthropic in your commit messages
Do not rebuild the server - it will be continuosly be rebuilt and run by the air configs.  Output of the server will be written to /tmp/sdlserver.log.  Build errors will also be shown in this log file.
Find the root cause of an issue before figuring out a solution.  Fix problems.
Do not create workarounds for issues without asking.  Always find the root cause of an issue and fix it.
The web module automatically builds when files are changed - DO NOT run npm build or npm run build commands.
Proto files are automatically regenerated when changed - DO NOT run buf generate commands.

## Available commands

When generating protos run the command `buf generate`

To build the SDL binary use `make`.  The binary is now already in path and can be run with `sdl`

## Key Architecture Decisions (June 2025)

### WASM Support Architecture (June 2025)
- **Native Types**: Created in `console/types.go` - Generator, Metric, SystemDiagram, etc.
- **Proto Types**: Only used at gRPC service boundaries (service.go)
- **Conversions**: Bidirectional converters in `console/conversions.go` and `console/canvas_conversions.go`
- **Clean Core**: Canvas, MetricTracer, and other core components now use native types exclusively
- **WASM Compatible**: No proto dependencies in core logic enables WASM compilation
- **DuckDB Removed**: Eliminated unused time-series database dependency
- **WASM Binary**: Successfully builds at 28.6MB (optimization pending)

### Metrics System
- Only supports `latency` and `count` metric types (no `utilization` or `throughput`)
- Metrics are pre-aggregated at collection time using specified aggregation function
- Connect-Web streaming used for real-time metric updates (no WebSocket)

### Canvas Management
- Multi-canvas support with unique IDs (e.g., ubermvp, uberv2, uberv3)
- Canvas reset requires explicit ID: `sdl canvas reset <canvasId>`
- Dashboard URL pattern: `/canvases/{canvasId}/`
- CLI defaults to SDL_CANVAS_ID env var, then "default"

### Dashboard Updates
- Generators polled every 5 seconds (no WebSocket needed)
- System diagram auto-updates when generators change
- Manual refresh button (🔄) for immediate updates
- Charts properly destroyed before recreation to prevent errors

### Demo Recipes
- Uber demos show MVP → Intermediate → Modern evolution
- Manual arrival rate updates required for proper queueing simulation
- All demos constrained to latency/count metrics only

# Summary instructions

When you are using compact, please focus on test output and code changes

- For the NEXTSTEPS.md always use the top-level ./NEXTSTEPS.md so we have a global view of the roadmap instead of being fragemented in various folders.

## SDL Demo Guidelines
- Make sure when you create SDL demos they are not as markdown but as .recipe files that are executable with pause points that print out what is going to be come next before the SDL command is executed.
```

**Session Workflow Memories:**
- When you checkpoint update all relevant .md files with our latest understanding, statuses and progress in the current session and then commit.
```