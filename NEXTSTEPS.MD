# SDL Project: Next Steps & Outstanding Work

This document outlines the key areas for future development and refinement in the SDL project, reflecting the current state after static diagramming enhancements.

## 1. Runtime and Evaluation Enhancements (`runtime` package)

*   **Single Call Tracing (Feature #1 from discussion):**
    *   Augment `SimpleEval` or create a `TracingEval` to record execution paths.
    *   Implement a `Tracer` object callable by the evaluator on method entry/exit.
    *   **Concurrency Tracing:** Extend tracing to handle `go` (future spawn) and `wait` (future await) statements, capturing asynchronous flows. Define a structured trace output (e.g., JSON) that can represent concurrency.
    *   Integrate this into `cmd/sdl trace`.

*   **Exhaustive Tracing / "Travel Maps" (Feature #2 from discussion):**
    *   Develop a new `ExhaustiveEval` or `ProbabilisticPathEval`.
    *   This evaluator should systematically explore branches from `distribute`, `sample`, `if`, and `switch` statements.
    *   Implement configurable loop unrolling (`--max-loop-iterations-exhaustive`).
    *   The evaluator should build a collapsed path graph (DAG) directly, showing divergences, convergences, and associated probabilities/frequencies.
    *   Define an output format for this graph suitable for visualization.

*   **Full Concurrency Support (`go`, `wait`):**
    *   Fully implement the evaluation semantics for `go` and `wait` statements in an appropriate evaluator (likely extending `SimpleEval` or in a new concurrent evaluator).
    *   Leverage or refine the `Frame` concept for managing concurrent execution contexts if suitable.

*   **DSL `analyze` Block Execution:**
    *   Implement runtime support for executing `analyze` blocks defined in SDL systems.
    *   This will involve evaluating the target expression (often a method call that produces `Outcomes`), calculating metrics (availability, P99, etc. from the `core` package), and comparing against `expect` statements within the `analyze` block.
    *   This is the backend for the `sdl run` command.

## 2. CLI Command Implementation & Enhancements (`cmd/sdl`)

*   **`sdl run`:**
    *   Integrate with the runtime to execute specified `analyze` blocks (see above).
    *   Display results, including metrics and expectation pass/fail status.

*   **`sdl diagram dynamic` (Feature #4 from discussion):**
    *   Consume trace output from the (to-be-implemented) single-call tracing feature.
    *   Generate sequence diagrams (e.g., Mermaid) or call graphs (e.g., DOT) from the trace data.

*   **`sdl trace`:**
    *   Implement fully using the single-call tracing capabilities developed in the `runtime`.

*   **`sdl plot` Enhancements:**
    *   Consider more plot types (e.g., histograms of latencies from `Outcomes`).
    *   Allow plotting from `sdl run --json-results` if it outputs detailed `Outcomes` or `AnalysisResult` data.

*   **Refine Static Diagramming (`sdl diagram static`):**
    *   Improve layout algorithms for cleaner diagrams, especially for complex systems.
    *   Add more styling options.
    *   Consider edge routing improvements for SVG/DOT to reduce line crossings.

## 3. Language Features & Semantics (DSL & `decl`)

*   **Built-in Functions:** Define and implement a standard library of built-in functions (e.g., math functions, string functions, `len()` for list-like types if they are added).
*   **Error Handling in DSL:** Define error types and potentially try/catch mechanisms if complex error handling within SDL is desired.
*   **List/Map Types:** Consider adding first-class list and map collection types to the DSL, along with associated operations.
*   **Loops (`for`):** Ensure robust semantics for `for` loops (e.g., fixed iterations, iterating over collections if added).
*   **Global `options` Blocks:** Fully implement parsing and application of global and component-level `options` (e.g., for `MaxOutcomeLen`, default profiles).

## 4. Testing & CI

*   **Runtime/Evaluator Test Suite:** Extensive tests for all AST node evaluations, tracing, and new evaluator behaviors (exhaustive, concurrent).
*   **End-to-End CLI Tests:** Expand tests for all CLI commands, especially `run`, `trace`, and dynamic `diagram` as they become functional.
*   **Benchmarking:** For core operations, `SimpleEval`, and any new evaluators.

## 5. Documentation & Examples

*   **Comprehensive DSL Manual:** Update with all new language features, runtime semantics, and CLI command usage.
*   **Go API Documentation:** Continue improving Godoc for all packages.
*   **More `.sdl` Examples:** Create examples demonstrating new features like `go`/`wait`, advanced `distribute`, error handling, and new built-ins.
*   **Tutorials:** For using the CLI for various analysis tasks (validation, plotting, diagramming, tracing).

## 6. Parser Enhancements (`parser` package)

*   **Error Recovery:** Continue to improve error reporting and explore better error recovery strategies in the `goyacc` parser for a more user-friendly experience.
*   **Syntax Refinements:** Address any awkward syntax identified during broader use of new features.

**Priority Focus (Immediate to Medium Term):**

1.  **Single Call Tracing (`runtime` & `cmd/sdl trace`).**
2.  **Dynamic Diagram Generation (`cmd/sdl diagram dynamic`) based on traces.**
3.  **Full `go`/`wait` concurrency implementation in `runtime` and DSL.**
4.  **Implementation of `sdl run` based on DSL `analyze` blocks.**
5.  **Exhaustive Tracing / "Travel Maps".**
