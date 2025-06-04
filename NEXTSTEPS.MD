# SDL Project: Next Steps & Outstanding Work

This document outlines the key areas for future development and refinement in the SDL project, reflecting the current state after static diagramming enhancements and recent discussions on concurrency and tracing.

## 1. Runtime and Evaluation Enhancements (`runtime` package)

*   **Concurrency Model Implementation:**
    *   **`go { <block_returns_T> } => SingleFuture[T]`:** Implement spawning of single asynchronous tasks.
    *   **`gobatch N { <block_returns_T> } => BatchFuture[T]`:** Implement spawning of N parallel identical operations. `T` (type returned by block) can be `Bool`, `Enum`, `Outcomes[Bool]`, or `Outcomes[Enum]`.
    *   **`waitfor <batch_future> using MyAggregator.AggregateMethod(params...)`:** Implement the synchronization primitive. The runtime will pass necessary batch execution data (internally, a structure like `NativeBatchInfo` containing `OriginalCountN` and `SingleOpOutcomeProfile` for `T`) to the specified native Go `AggregateMethod`. The native aggregator is responsible for both functional aggregation and calculating the appropriate completion time profile for its `Outcomes[SummaryEnum]` result.
    *   **`wait <single_future> => T`:** Implement waiting for and retrieving the result of a single future.
    *   **`waitgroup (f1, f2, ...) => (T1, T2, ...), makespan_time`:** Implement waiting for multiple, potentially heterogeneous, single futures, returning their individual results and the overall makespan.
    *   **Core Library Support:** Enhance `sdl/core` with functions to calculate distributions of order statistics (e.g., time to Cth completion from N, makespan of N) needed by native aggregators.

*   **Single Call Tracing (Feature #1):**
    *   Augment `SimpleEval` (or create a `TracingEval`) to record execution paths.
    *   Implement a `Tracer` object callable by the evaluator on method entry/exit, future spawn (`go`, `gobatch`), and synchronization (`waitfor`, `waitgroup`).
    *   Define a structured trace output (e.g., JSON) capable of representing sequential and concurrent operations.
    *   Integrate this into `cmd/sdl trace`.

*   **Exhaustive Tracing / "Travel Maps" (Feature #2):**
    *   Develop a new `ExhaustiveEval` or `ProbabilisticPathEval`.
    *   This evaluator should systematically explore branches from `distribute`, `sample`, `if`, and `switch` statements (when used as expressions with multiple outcomes).
    *   Implement configurable loop unrolling.
    *   The evaluator should build a collapsed path graph (DAG) directly, showing divergences, convergences, and associated probabilities/frequencies.
    *   Define an output format for this graph suitable for visualization.

*   **DSL `analyze` Block Execution:**
    *   Implement runtime support for executing `analyze` blocks defined in SDL systems.
    *   This will involve evaluating the target expression, calculating metrics using `sdl/core`, and comparing against `expect` statements.
    *   This is the backend for the `sdl run` command.

## 2. Language Features & Semantics (DSL & `decl`, `parser`)

*   **Introduce Concurrency Keywords:** Add `go`, `gobatch`, `waitfor`, `using`, `waitgroup` to the grammar (`parser/grammar.y`).
*   **Introduce Future Types:** Define `SingleFuture[T]` and `BatchFuture[T]` as opaque types in the SDL type system (`decl` package). These are primarily handles passed to `waitfor` or `waitgroup` and native aggregators.
*   **`switch` as an Expression:** Allow `switch` statements to be used as expressions, returning a value. Ensure all branches of a switch expression yield type-compatible results. Update grammar and type inference.
*   **Built-in Functions:** Define and implement a standard library of built-in functions.
*   **Global `options` Blocks:** Fully implement parsing and application of global and component-level `options`.

## 3. CLI Command Implementation & Enhancements (`cmd/sdl`)

*   **`sdl run`:**
    *   Integrate with the runtime to execute specified `analyze` blocks.
*   **`sdl trace`:**
    *   Implement fully using the single-call tracing capabilities.
*   **`sdl diagram dynamic` (Feature #4):**
    *   Consume trace output from the `sdl trace` feature.
    *   Generate sequence diagrams or call graphs.
*   **`sdl plot` Enhancements:**
    *   Consider more plot types (e.g., histograms).
*   **Refine Static Diagramming (`sdl diagram static`):**
    *   Improve layout algorithms and styling options.

## 4. Native Component Development (`components`)

*   **Native Aggregators:** Develop a set of standard native aggregator components/methods for common summarization patterns used with `waitfor ... using ...` (e.g., AllSucceeded, AnySucceeded, QuorumReached, based on boolean or simple enum results).

## 5. Testing & CI

*   **Concurrency Runtime Tests:** Extensive tests for `go`, `gobatch`, `waitfor`, `waitgroup`, and native aggregators.
*   **Tracing Tests:** Validate the accuracy and completeness of trace generation.
*   **End-to-End CLI Tests:** Expand for all commands, especially new/updated ones.

## 6. Documentation & Examples

*   **Comprehensive DSL Manual:** Update with concurrency features, `switch` expressions, tracing, etc.
*   **More `.sdl` Examples:** Demonstrate new concurrency features, aggregators, and tracing for realistic scenarios.
*   **Tutorials:** On using the CLI for tracing and analyzing concurrent models.

## Priority Focus (Immediate to Medium Term):

1.  **Core Concurrency Primitives in `runtime` and DSL (`go`, `gobatch`, `waitfor using`, `SingleFuture`, `BatchFuture`).** This includes the necessary `core` library support for latency distribution calculations (e.g., makespan).
2.  **Basic Native Aggregators** for common boolean summarization (all/any/some succeeded).
3.  **Single Call Tracing (`runtime` & `cmd/sdl trace`).**
4.  **Dynamic Diagram Generation (`cmd/sdl diagram dynamic`) based on traces.**
5.  **Implementation of `sdl run` based on DSL `analyze` blocks.**
6.  **`switch` as an Expression.**
7.  **Exhaustive Tracing / "Travel Maps".**
