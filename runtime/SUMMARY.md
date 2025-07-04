# SDL Runtime Package Summary (`runtime`)

**Purpose:**

This package is responsible for the execution and evaluation of System Design Language (SDL) models that have been parsed and loaded. It provides the mechanisms to instantiate components and systems, interpret their defined logic, and interact with both native Go components and user-defined SDL components. Its primary role is to power the simulation-based features of the CLI.

**Key Concepts & Components:**

*   **`Runtime` (`runtime.go`):**
    *   The central orchestrator for loading and preparing SDL files for execution. It holds a reference to a `loader.Loader` and manages a cache of loaded `FileInstance` objects.
    *   It contains the `CreateNativeComponent` factory method, which can instantiate any of the registered native Go components (e.g., `Cache`, `Disk`, `LSMTree`) by name.

*   **`FileInstance` (`file.go`):**
    *   Represents a loaded and parsed SDL file at runtime, holding the AST (`FileDecl`) and a root environment. It provides methods to create runtime instances of systems and components.

*   **`SystemInstance` & `ComponentInstance` (`system.go`, `component.go`):**
    *   Runtime representations of `system` and `component` declarations. They manage their respective AST nodes and runtime environments (`Env[Value]`). Their `Initializer()` methods are crucial for compiling declarative bodies into executable statements.
    *   **Utilization Integration**: ComponentInstance now implements GetUtilizationInfo() for hierarchical utilization reporting from native components. SystemInstance provides AllComponents() for system-wide utilization queries.

*   **`SimpleEval` (`simpleeval.go`):**
    *   The primary interpreter (evaluator) for SDL AST nodes. It walks the AST to evaluate expressions and statements.
    *   It manages runtime state, including variables and component instances, within an `Env[Value]`.
    *   It handles all language constructs: literals, operations, statements (`let`, `if`, `for`, `return`), `new` (instantiation), and method calls.
    *   **Latency Accumulation**: A key feature is its tracking of simulated time. The `Time` field on the `decl.Value` struct is used to accumulate the latency of operations within a single simulation run.
    *   **Enhanced Boolean Evaluation**: Now correctly handles `Outcomes[Bool]` types in unary operations (like `not`), sampling from probabilistic outcomes and applying boolean operations while preserving latency information.

*   **Concurrency Primitives (`aggregator.go`, `simpleeval.go`):**
    *   The runtime now has placeholder implementations for concurrency constructs like `gobatch` and `wait using <Aggregator>`.
    *   The `WaitAll` aggregator has a basic simulation-focused implementation that correctly models the "makespan" latency of parallel operations and returns a value, allowing simulations of concurrent systems to complete successfully.

*   **Utilities (`utils.go`):**
    *   **`RunCallInBatches`**: A powerful helper function that drives the `sdl run` and `sdl plot` commands. It efficiently executes a method thousands of times across multiple concurrent workers. It has been updated to correctly track and provide the per-run latency for creating accurate time-series data.
    *   **Instance Management**: Now ensures proper instance isolation by reusing existing system environments rather than creating new ones for each batch, ensuring Canvas.Set() parameter modifications reach the correct simulation instances.

*   **Native Component Interaction (`native.go`):**
    *   Defines the bridge to native Go components (from the `components` package). It uses reflection to invoke methods on Go objects, transparently converting between the runtime's `decl.Value` and standard Go types.
    *   **Enhanced InvokeMethod**: The `InvokeMethod` function now includes a `shouldSample` parameter to support future extensibility for `OutcomesEval`, preparing for more sophisticated probabilistic evaluation strategies.

*   **Flow Evaluation System (`floweval.go`, `flowrteval.go`):**
    *   **Dual Implementation**: String-based (legacy) and Runtime-based (current) flow evaluation
    *   **Runtime-Based FlowEval**: Uses actual ComponentInstance objects from SimpleEval, avoiding duplication
    *   **`SolveSystemFlowsRuntime()`**: Iterative fixed-point solver for system-wide flow analysis
    *   **`FlowAnalyzable` Interface**: Native components report traffic patterns; NWBase provides smart defaults
    *   **Known Limitations**: Overestimates flows for early return patterns, no capacity-based backpressure yet
    *   **Convergence**: Typically converges in 7-12 iterations with 0.5 damping factor

*   **Live Metrics System (`metrics_*.go`, `trace.go`):**
    *   **MetricStore**: System-aware measurement management with component instance resolution
    *   **Virtual Time**: All timestamps use `core.Duration` for deterministic simulation
    *   **Component Resolution**: Maps instance names (e.g., "server") to actual ComponentInstance pointers
    *   **Efficient Matching**: Direct pointer comparison between TraceEvent and resolved components
    *   **CircularBuffer**: Memory-efficient storage for recent metric points
    *   **Aggregations**: Comprehensive support for sum, rate, percentiles (p50, p90, p95, p99)
    *   **Enhanced Tracer**: TraceEvent carries Component and Method references directly

**Role in the Project:**

The runtime package is the engine that brings SDL models to life. It enables the powerful simulation capabilities of the toolchain by executing the user-defined logic and providing the raw data (latency, results) needed for performance analysis.

**Current Status & Recent Work:**

*   `SimpleEval` can execute a significant SDL subset, making it suitable for single-path, simulation-based analysis.
*   The `RunCallInBatches` utility is robust and supports the CLI's simulation features.
*   The placeholder implementation of the `WaitAll` aggregator has been fixed to unblock the simulation of simple concurrent patterns.
*   The native component factory `CreateNativeComponent` has been updated to support all available native components.
*   **Capacity Modeling Support**: The runtime now fully supports capacity modeling workflows with proper boolean evaluation for `Outcomes[Bool]` types, correct instance management for parameter modification, and enhanced native method invocation for probabilistic components.
*   **Runtime-Based Flow Analysis**: Migrated from string-based to ComponentInstance-based flow evaluation, using actual runtime components. Current implementation provides reasonable estimates for simple flows but overestimates in presence of early returns. Suitable for initial system sizing where overestimation is conservative.
*   **Live Metrics System**: Complete implementation of real-time performance monitoring with virtual time support, component instance resolution, and comprehensive aggregation functions. Seamlessly integrated with SimpleEval through enhanced ExecutionTracer.

**Key Dependencies:**

*   `decl`: For all AST nodes, the `Value` type system, and `Env`.
*   `loader`: To get validated `FileDecl` objects.
*   `core`: For core types like `core.Duration`.
*   `components/decl`: For the native component wrappers.
