# SDL Runtime Package Summary (`runtime`)

**Purpose:**

This package is responsible for the execution and evaluation of System Design Language (SDL) models that have been parsed and loaded. It provides the mechanisms to instantiate components and systems, interpret their defined logic, and interact with both native Go components and user-defined SDL components. Its primary role is to power the simulation-based features of the CLI.

**Key Concepts & Components:**

*   **`Runtime` (`runtime.go`):**
    *   The central orchestrator for loading and preparing SDL files for execution. It holds a reference to a `loader.Loader` and manages a cache of loaded `FileInstance` objects.

*   **`FileInstance` (`file.go`):**
    *   Represents a loaded and parsed SDL file at runtime, holding the AST (`FileDecl`) and a root environment. It provides methods to create runtime instances of systems and components.

*   **`SystemInstance` & `ComponentInstance` (`system.go`, `component.go`):**
    *   Runtime representations of `system` and `component` declarations. They manage their respective AST nodes and runtime environments (`Env[Value]`). Their `Initializer()` methods are crucial for compiling declarative bodies into executable statements.

*   **`SimpleEval` (`simpleeval.go`):**
    *   The primary interpreter (evaluator) for SDL AST nodes. It walks the AST to evaluate expressions and statements.
    *   It manages runtime state, including variables and component instances, within an `Env[Value]`.
    *   It handles all language constructs: literals, operations, statements (`let`, `if`, `for`, `return`, `delay`), `new` (instantiation), and method calls.
    *   **Latency Accumulation**: A key feature is its tracking of simulated time. The `Time` field on the `decl.Value` struct is used to accumulate the latency of operations within a single simulation run.
    *   **Error Handling**: The evaluator has been improved to provide clear, user-friendly error messages for common runtime issues, such as calling a non-existent method.

*   **Concurrency Primitives (`aggregator.go`, `simpleeval.go`):**
    *   The runtime now has placeholder implementations for concurrency constructs like `gobatch` and `wait using <Aggregator>`.
    *   The `WaitAll` aggregator has a basic simulation-focused implementation that correctly models the "makespan" latency of parallel operations and returns a value, allowing simulations of concurrent systems to complete successfully.

*   **Utilities (`utils.go`):**
    *   **`RunCallInBatches`**: A powerful helper function that drives the `sdl run` and `sdl plot` commands. It efficiently executes a method thousands of times across multiple concurrent workers. It has been updated to correctly track and provide the per-run latency for creating accurate time-series data.

*   **Native Component Interaction (`native.go`):**
    *   Defines the bridge to native Go components (from the `components` package). It uses reflection to invoke methods on Go objects, transparently converting between the runtime's `decl.Value` and standard Go types.

**Role in the Project:**

The runtime package is the engine that brings SDL models to life. It enables the powerful simulation capabilities of the toolchain by executing the user-defined logic and providing the raw data (latency, results) needed for performance analysis.

**Current Status & Recent Work:**

*   `SimpleEval` can execute a significant SDL subset, making it suitable for single-path, simulation-based analysis.
*   The `RunCallInBatches` utility is robust and supports the CLI's simulation features.
*   The placeholder implementation of the `WaitAll` aggregator has been fixed to unblock the simulation of simple concurrent patterns.
*   Error reporting within the evaluator has been improved.

**Key Dependencies:**

*   `decl`: For all AST nodes, the `Value` type system, and `Env`.
*   `loader`: To get validated `FileDecl` objects.
*   `core`: For core types like `core.Duration`.
*   `components/decl`: For the native component wrappers.
