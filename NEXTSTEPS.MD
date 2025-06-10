# SDL Project: Next Steps & Outstanding Work

This document outlines the key areas for future development and refinement in the SDL project.

**`[COMPLETED]`** The `sdl run` and `sdl plot` commands have been implemented to provide a powerful simulation-based analysis workflow.

---

## 1. Runtime and Evaluation Enhancements (`runtime` package)

*   **Full Concurrency Model Implementation:**
    *   The `gobatch` and `wait using` constructs have placeholder implementations that are sufficient for basic simulation but need to be fully realized.
    *   **`gobatch N { ... }`**: The runtime should properly model the execution of N parallel operations, not just a single representative one.
    *   **Native Aggregators**: Develop a set of standard native aggregator methods (`WaitAll`, `WaitAny`, `Quorum`) that can properly calculate the resulting latency distribution and success probability from a batch of futures.
    *   **`go { ... }`**: Implement the spawning and execution of single asynchronous tasks.
    *   **`wait <future>`**: Implement waiting for single and multiple heterogeneous futures.
    *   **Core Library Support:** Enhance `sdl/core` with functions to calculate distributions of order statistics (e.g., time to C-th completion from N, makespan of N) needed by the native aggregators.

*   **Single Call Tracing (The "How" and "Why"):**
    *   This is the next major feature. Augment `SimpleEval` (or create a new `TracingEval`) to record the execution path of a *single* method call.
    *   Implement a `Tracer` object that the evaluator can call on method entry/exit, future spawn (`go`, `gobatch`), and synchronization (`wait`).
    *   Define a structured trace output (e.g., JSON) capable of representing sequential and concurrent operations.
    *   This will be the backend for the `sdl trace` and `sdl diagram dynamic` commands.

*   **Exhaustive Tracing / Analytical Evaluation:**
    *   This is a longer-term goal to move beyond single-path simulation.
    *   Develop a new `ExhaustiveEval` that systematically explores all probabilistic branches (`distribute`, `sample`) to build a complete outcome graph, enabling true analytical results directly from the DSL.

## 2. Language Features & Semantics (DSL & `decl`, `parser`)

*   **`switch` as an Expression:** Allow `switch` statements to be used as expressions, returning a value. Ensure all branches of a switch expression yield type-compatible results. Update the grammar and type inference logic.
*   **Built-in Functions:** Define and implement a standard library of built-in functions (e.g., math functions, string manipulation).
*   **Global `options` Blocks:** Fully implement parsing and application of global and component-level `options` to configure the runtime or components.

## 3. CLI Command Implementation & Enhancements (`cmd/sdl`)

*   **`sdl trace`:**
    *   Implement this command fully using the single-call tracing infrastructure described above. It should produce a detailed JSON trace file.
*   **`sdl diagram dynamic`:**
    *   Implement this command to consume the trace output from `sdl trace` and generate a sequence diagram or call graph (e.g., in Mermaid format).
*   **Refine Static Diagramming (`sdl diagram static`):**
    *   Improve layout algorithms and styling options.

## 4. Testing & Documentation

*   **Concurrency Runtime Tests:** Add extensive tests for the full concurrency model once implemented.
*   **Tracing Tests:** Validate the accuracy and completeness of trace generation.
*   **Documentation:** Update the project's documentation and examples to cover the new `run`/`plot` workflow and the upcoming `trace` command.

## Priority Focus (Immediate to Medium Term):

1.  **Single Call Tracing (`runtime` & `cmd/sdl trace`).** This is the highest priority as it unlocks debugging and dynamic diagrams.
2.  **Dynamic Diagram Generation (`cmd/sdl diagram dynamic`).** This provides a powerful visualization for the new trace data.
3.  **Full Concurrency Model.** This involves fleshing out the `gobatch`/`wait` logic and creating robust native aggregators.
4.  **`switch` as an Expression.** This is a valuable language enhancement for more concise models.
