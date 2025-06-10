# SDL Runtime Package Summary (`runtime`)

**Purpose:**

This package is responsible for the execution and evaluation of System Design Language (SDL) models that have been parsed and loaded. It provides the mechanisms to instantiate components and systems, interpret their defined logic (methods, initializers), and interact with both native Go components and user-defined SDL components.

**Key Concepts & Components:**

*   **`Runtime` (`runtime.go`):**
    *   The central orchestrator for loading and preparing SDL files for execution.
    *   Holds a reference to an `loader.Loader` to parse and validate SDL source files.
    *   Manages a cache of `FileInstance` objects.
    *   Includes a factory mechanism (`CreateNativeComponent`) to instantiate native Go components that are declared in SDL.

*   **`FileInstance` (`file.go`):**
    *   Represents a loaded and parsed SDL file at runtime.
    *   Holds the `FileDecl` (AST) and a root runtime environment (`Env[Value]`) for the file.
    *   Provides methods to create `SystemInstance` and `ComponentInstance` objects.

*   **`SystemInstance` (`system.go`):**
    *   Represents a runtime instance of an SDL `system` declaration.
    *   Contains its `SystemDecl` (AST) and runtime environment (`Env[Value]`).
    *   `Initializer()` compiles the system body into a `BlockStmt` for setup.
    *   `GetUninitializedComponents()` helps identify unsatisfied dependencies.

*   **`ComponentInstance` (`component.go`) & `ObjectInstance` (`object.go`):**
    *   `ObjectInstance`: Base for entities with parameters (native or SDL-defined).
    *   `ComponentInstance`: Extends `ObjectInstance` for SDL `component` instances.
    *   Manages `ComponentDecl` (AST) and an initial environment (`InitialEnv`).
    *   Handles native component interaction via `NativeObject` interface.
    *   `Initializer()` prepares a `BlockStmt` for user-defined components.
    *   `Get()` and `Set()` for parameter access.

*   **`SimpleEval` (`simpleeval.go`):**
    *   The primary interpreter for SDL AST nodes.
    *   Traverses AST, evaluating nodes within an `Env[Value]`.
    *   Handles literals, identifiers, operations, statements (`let`, `set`, `if`, `for`, `return`, `log`, `delay`), `new` (instantiation), and method calls.
    *   Latency accumulation via `Value.Time` and `currTime *core.Duration`.
    *   Async Enhancements: `go`, `gobtach` and `wait using`
    *   *Planned Enhancements:* Integration with tracing mechanisms.

*   **Native Component Interaction (`native.go`):**
    *   `NativeObject` interface (`Get`, `Set`).
    *   `InvokeMethod(...)`: Uses Go reflection for native method calls, handling argument/result conversion between SDL `Value` and Go types. `Value.Time` carries latency.

*   **Runtime Environment (`Env[Value]` from `decl`):**
    *   Stores runtime values of variables, parameters, and instances.

*   **Concurrency Primitives:**
    *   **`gobatch N { <block_returns_T> } => BatchFuture[T]`:** Language construct to spawn `N` parallel identical operations. `T` is the type returned by the block (e.g., `Enum`, `Outcomes[Enum]`). `BatchFuture[T]` is an opaque handle.
    *   **`wait <future1> <future2> ... <futureN> using AggregatorMethod(Params) => AggregatorReturnType`:** SDL statement to synchronize on a `BatchFuture` or N normal futures. The runtime will facilitate passing necessary futures (like the profile of a single operation and `N`) to the native Go `AggregateMethod`. This native method is responsible for both functional aggregation and calculating the appropriate completion time profile for its `AggregatorReturnType` result as a Value object.  For a simulation based eval (like simpleeval) this returns a single value.  For an analysis based eval (like OutcomesEval) it could return a Value that is `Outcomes[AggregatorMethod]`.
    *   **`go { <block_returns_T> } => SingleFuture[T]`:** For spawning a single asynchronous task.
    *   Single Wait - **`wait <single_future> => T`:** To get the result of a single future.
    *   Multi Wait - **`wait (f1, f2, ...) => (T1, T2, ...)`:** For waiting on multiple, potentially heterogeneous, single futures.
    *   Multi Wait With Aggregator - **`wait (f1, f2, ...) using Aggregator(Params) => (AggResult1)`:** For waiting on multiple, potentially heterogeneous, single futures but aggregating them into a single result.

*   **Tracing Infrastructure (Planned):**
    *   The runtime will be augmented to emit trace events (e.g., method entry/exit, future spawn/await) to a `Tracer` object. This will enable `sdl trace` and dynamic diagram generation.

*   **Utilities (`utils.go`):**
    *   `RunCallInBatches`: Helper for the `plot` command.

**Role in the Project:**

*   Provides the execution engine for SDL.
*   Bridges declarative models with behavior and performance.
*   Enables simulation and, with future evaluators, analytical modeling.

**Current Status & Recent Work:**

*   `SimpleEval` executes a significant SDL subset for single-path simulation and latency tracking.
*   Native component integration is functional.
*   Focus is shifting towards implementing concurrency primitives (`gobatch`, `waitfor`) and tracing.

**Key Dependencies:**

*   `decl`: For AST, `Value` type, `Env`.
*   `loader`: For loaded `FileDecl` objects.
*   `core`: For `core.Duration`, `core.Outcomes`.
*   `components/decl`: For native component wrappers.

**Future Considerations (linking to `NEXTSTEPS.MD`):**

*   **Tracing Integration:** Develop and integrate the `Tracer` mechanism.
*   **Exhaustive/Probabilistic Evaluation:** Design and implement new evaluators beyond `SimpleEval`.
*   **Enhanced Error Handling & Debugging.**
