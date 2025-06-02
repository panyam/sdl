# SDL Runtime Package Summary (`runtime`)

**Purpose:**

This package is responsible for the execution and evaluation of System Design Language (SDL) models that have been parsed and loaded. It provides the mechanisms to instantiate components and systems, interpret their defined logic (methods, initializers), and interact with both native Go components and user-defined SDL components.

**Key Concepts & Components:**

*   **`Runtime` (`runtime.go`):**
    *   The central orchestrator for loading and preparing SDL files for execution.
    *   Holds a reference to an `loader.Loader` to parse and validate SDL source files.
    *   Manages a cache of `FileInstance` objects.
    *   Includes a factory mechanism (`CreateNativeComponent`) to instantiate native Go components that are declared in SDL (e.g., `Disk`, `HashIndex` from the `components` package).

*   **`FileInstance` (`file.go`):**
    *   Represents a loaded and parsed SDL file at runtime.
    *   Holds the `FileDecl` (AST) and a root runtime environment (`Env[Value]`) for the file.
    *   Provides methods to create `SystemInstance` and `ComponentInstance` objects from declarations within the file, resolving imports if necessary.

*   **`SystemInstance` (`system.go`):**
    *   Represents a runtime instance of an SDL `system` declaration.
    *   Contains a reference to its `SystemDecl` (AST) and its own runtime environment (`Env[Value]`) for instances and parameters defined within it.
    *   The `Initializer()` method compiles the system body (instance declarations and their overrides) into a `BlockStmt` that can be evaluated to set up the system.
    *   `GetUninitializedComponents()` helps identify dependencies that haven't been satisfied after initialization.

*   **`ComponentInstance` (`component.go`) & `ObjectInstance` (`object.go`):**
    *   `ObjectInstance`: A base for entities that can have parameters and be interacted with. It distinguishes between native Go objects and user-defined SDL objects.
    *   `ComponentInstance`: Extends `ObjectInstance`. Represents a runtime instance of an SDL `component`.
    *   Manages its own `ComponentDecl` (AST) and an initial environment (`InitialEnv`) which includes parameters and resolved dependencies.
    *   Handles native component interaction through the `NativeObject` interface (which `components/decl/*` wrappers implement).
    *   For user-defined components, its `Initializer()` method prepares a `BlockStmt` to evaluate parameter defaults and initialize `uses` dependencies that have overrides.
    *   Provides `Get()` and `Set()` methods for accessing/modifying its parameters (delegating to `NativeInstance` for native components).

*   **`SimpleEval` (`simpleeval.go`):**
    *   The primary interpreter/evaluator for SDL AST nodes (expressions and statements).
    *   It traverses the AST, evaluating nodes within a given runtime environment (`Env[Value]`).
    *   Implements evaluation logic for various SDL constructs: literals, identifiers, binary/unary operations, `let` and `set` statements, `if`, `for`, `return`, `log`, `delay`, `distribute`, `sample`, `new` (component instantiation), and method calls (`CallExpr`).
    *   `EvalInitSystem()`: Uses `SimpleEval` to execute the initializer block of a `SystemInstance`.
    *   Method calls on native components are delegated via `InvokeMethod` (in `native.go`) which uses reflection.
    *   Crucially, evaluation accumulates latency through the `Value.Time` field. The `currTime *core.Duration` parameter passed during evaluation tracks the total time taken for a sequence of operations.

*   **Native Component Interaction (`native.go`):**
    *   `NativeObject` interface: (`Get`, `Set`). (Note: `InvokeMethod` is a standalone function now).
    *   `InvokeMethod(...)`: Uses Go reflection to call methods on native component instances. It handles argument conversion (SDL `Value` to Go types) and result conversion (Go types back to SDL `Value`). The returned `Value` from a native method call is expected to have its `Time` field set to represent the latency of that native operation.

*   **Runtime Environment (`Env[Value]` from `decl`):**
    *   Used extensively to store and look up runtime values of variables, parameters, and component instances during evaluation.

*   **`Frame` (`frame.go` - currently unused by `SimpleEval` but planned for concurrency):**
    *   Defines a call frame structure intended for managing lexical scopes and potentially tracking asynchronous operations (`go`/`wait` - though these are not yet fully implemented in `SimpleEval`).

*   **Utilities (`utils.go`):**
    *   `RunCallInBatches`: A helper function used by the `plot` command. It takes a `SystemInstance`, component/method names, and batching parameters. It repeatedly calls a target method using `SimpleEval` and collects the resulting `Value` objects (which include latency via `Value.Time`).

**Role in the Project:**

*   Provides the execution engine for SDL.
*   Bridges the declarative SDL models with their actual behavior and performance characteristics.
*   Enables simulation-like execution of specific system operations (as seen in `RunCallInBatches`).

**Current Status & Recent Work:**

*   `SimpleEval` can execute a significant subset of the SDL, including component instantiation, parameter setting, method calls (both SDL-defined and native), and basic control flow.
*   The `Value.Time` mechanism for tracking latency is integrated into `SimpleEval` and native calls.
*   The `plot` command in `cmd/sdl` leverages this runtime to generate performance data.
*   The structure for handling native components via reflection is in place.
*   Concurrency features (`go`/`wait` via `Frame`) are designed but not yet fully integrated into `SimpleEval`'s core loop.

**Key Dependencies:**

*   `decl`: For AST node structures, `Value` type, and `Env`.
*   `loader`: To obtain loaded and validated `FileDecl` objects.
*   `core`: For `core.Duration` and `core.Outcomes` when dealing with probabilistic values (though `SimpleEval` itself primarily deals with single `Value` paths from sampled outcomes).
*   `components/decl`: For the native component wrappers.

**Future Considerations (linking to `NEXTSTEPS.MD`):**

*   **Full Concurrency Support:** Implement evaluation for `go` and `wait` statements, likely leveraging or enhancing the `Frame` concept.
*   **Tracing:** Augment `SimpleEval` (or create a new evaluator) to record execution traces (Feature #1 from our discussion).
*   **Exhaustive/Probabilistic Evaluation:** Develop new evaluators (e.g., an `OutcomesEval` or `ProbabilisticPathEval`) for analytical modeling or exhaustive path tracing (Feature #2).
*   **Enhanced Error Handling:** More robust runtime error reporting.
*   **Debugging Capabilities:** Potential for step-through debugging in the future.
