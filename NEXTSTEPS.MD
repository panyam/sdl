# SDL Project: Next Steps & Outstanding Work

This document outlines the key areas for future development and refinement in the SDL project.

---
## Completed

*   **`[COMPLETED]` Native Methods & Grammar Simplification:**
    *   Added support for top-level `native method` declarations.
    *   Refactored the `delay` statement into a standard native method call, simplifying the grammar.
    *   Implemented a native method registry in the runtime for functions like `delay` and `log`.
*   **`[COMPLETED]` Histogram Plotting:** The `sdl plot` command now supports a `--y-axis=histogram` option to generate bar chart visualizations of metric distributions (e.g., latency). The `viz` package was enhanced to generate bar charts.
*   **`[COMPLETED]` Simulation & Plotting Workflow:** The `sdl run` and `sdl plot` commands are fully implemented, providing a powerful workflow for statistical performance analysis.
*   **`[COMPLETED]` Visualization Package Refactoring:** All diagramming and plotting logic has been successfully refactored into a new top-level `viz` package, decoupling it from the command-line interface.
*   **`[COMPLETED]` Tracing & Dynamic Diagram Workflow:**
    *   The `runtime` is now instrumented with a tracer to capture detailed execution events.
    *   The `sdl trace` command successfully generates a JSON trace file from a single method execution.
    *   The `sdl diagram dynamic` command consumes this trace file to generate a Mermaid sequence diagram, correctly visualizing concurrency and call flows.

---
## 1. Runtime and Evaluation Enhancements (`runtime` package)

*   **Full Concurrency Model Implementation:** This is the next major priority.
    *   The `gobatch` and `wait using` constructs are visualized correctly but the runtime model is a placeholder.
    *   **`gobatch N { ... }`**: The runtime should properly model the execution of N parallel operations, not just a single representative one. This is crucial for accurate simulation of fan-out scenarios.
    *   **Native Aggregators**: Develop a set of standard native aggregator methods (`WaitAll`, `WaitAny`, `Quorum`) that can properly calculate the resulting latency distribution and success probability from a batch of futures.
    *   **Core Library Support:** Enhance `sdl/core` with functions to calculate distributions of order statistics (e.g., time to C-th completion from N, makespan of N) needed by the native aggregators.

*   **Exhaustive Tracing / Analytical Evaluation:**
    *   This is a longer-term goal to move beyond single-path simulation.
    *   Develop a new `ExhaustiveEval` that systematically explores all probabilistic branches (`distribute`, `sample`) to build a complete outcome graph, enabling true analytical results directly from the DSL.

## 2. Language Features & Semantics (DSL & `decl`, `parser`)

*   **`switch` as an Expression:** Allow `switch` statements to be used as expressions, returning a value. Ensure all branches of a switch expression yield type-compatible results. Update the grammar and type inference logic.
*   **More Built-in Functions:** Define and implement a broader standard library of built-in functions (e.g., math functions, string manipulation).
*   **Global `options` Blocks:** Fully implement parsing and application of global and component-level `options` to configure the runtime or components.

## 3. Visualization Enhancements (`viz` package)

*   **Improve Dynamic Diagrams:**
    *   Add support for more output formats beyond Mermaid (e.g., SVG, Excalidraw).
    *   Show return values and latencies on the sequence diagram arrows.
*   **Refine Static Diagramming:**
    *   Improve layout algorithms and styling options.

## 4. Testing & Documentation

*   **Concurrency Runtime Tests:** Add extensive tests for the full concurrency model once it is implemented.
*   **Documentation:** Update the project's main documentation and examples to cover the new `trace` and `diagram dynamic` workflow in detail.

## Priority Focus (Immediate to Medium Term):

1.  **Full Concurrency Model (`runtime` & `core`).** This is the highest priority as it's the largest remaining gap in the simulation engine's capabilities.
2.  **Language Enhancements.** Implementing `switch` as an expression and adding more built-in functions will significantly improve the expressiveness of the DSL.
3.  **Visualization Polish.** Adding more output formats and details to the dynamic diagrams.
