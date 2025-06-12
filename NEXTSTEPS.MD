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
## 1. Interactive Recipe Runner

To provide a more fluid, notebook-like experience for iterative analysis and storytelling (e.g., for conference demos), we will build an interactive "recipe runner."

*   **Phase 1: Proof of Concept (Unit Test):** Before building the full `execute` command, create a unit test (`TestExecutionRecipe`) that validates the core primitives:
    1.  Load an SDL model via the Go API.
    2.  Instantiate a system and get a handle to a native component instance (e.g., a `components.Cache`).
    3.  **Run Scenario A:** Modify a parameter on the component struct (e.g., `cache.HitRate = 0.95`), run a simulation, and store the results in memory.
    4.  **Run Scenario B:** Modify the *same* component instance again (`cache.HitRate = 0.60`), run a new simulation, and store its results.
    5.  **Generate Comparison:** Combine the results from both scenarios into a single multi-series plot and save it to a file. This proves the end-to-end concept works.

*   **Phase 2: `sdl execute` Command:**
    *   Create the `cmd/sdl/commands/execute.go` file.
    *   Implement a simple parser for the `.recipe` file format.
    *   Build a `Session` object to manage the loaded state (loader, variables).
    *   Implement the recipe commands (`load`, `use`, `set`, `run`, `plot`, `diagram`, `echo`).

## 2. Runtime and Evaluation Enhancements (`runtime` package)

*   **Full Concurrency Model Implementation:** This is the next major priority after the recipe runner PoC.
    *   **`gobatch N { ... }`**: The runtime should properly model the execution of N parallel operations, not just a single representative one.
    *   **Native Aggregators**: Develop standard aggregator methods (`WaitAll`, `WaitAny`, `Quorum`) to calculate latency and success from futures.
    *   **Core Library Support:** Enhance `sdl/core` with functions for order statistics (e.g., makespan of N).

*   **Exhaustive Tracing / Analytical Evaluation:**
    *   A longer-term goal to move beyond single-path simulation by exploring all probabilistic branches (`distribute`, `sample`) to build a complete outcome graph.

## 3. Language Features & Semantics (DSL & `decl`, `parser`)

*   **`switch` as an Expression:** Allow `switch` to be used as an expression that returns a value.
*   **More Built-in Functions:** Define and implement a broader standard library.

## Priority Focus (Immediate to Medium Term):

1.  **Interactive Recipe Runner (Phase 1).** The PoC test is the highest priority as it directly enables the creation of compelling, story-driven demos for the upcoming conference.
2.  **Full Concurrency Model (`runtime` & `core`).** This remains the most significant feature gap in the simulation engine.
3.  **Language Enhancements.** Implementing `switch` as an expression and adding more built-in functions.
