# SDL Project: Next Steps & Outstanding Work

This document outlines the key areas for future development and refinement in the SDL project.

---
## Completed

*   **`[COMPLETED]` Interactive Session Proof-of-Concept:**
    *   Created the `console` package and the `Canvas` API, which provides a stateful engine for interactive analysis.
    *   Successfully implemented and tested the core workflow (`Load`, `Set`, `Run`, `Plot`) in a unit test (`TestExecutionRecipe`), proving the feasibility of the recipe runner concept.
*   **`[COMPLETED]` Native Methods & Grammar Simplification:**
    *   Added support for top-level `native method` declarations.
    *   Refactored the `delay` statement into a standard native method call, simplifying the grammar.
*   **`[COMPLETED]` Histogram Plotting:** The `sdl plot` command now supports a `--y-axis=histogram` option to generate bar chart visualizations of metric distributions.
*   **`[COMPLETED]` Simulation & Plotting Workflow:** The `sdl run` and `sdl plot` commands are fully implemented.
*   **`[COMPLETED]` Tracing & Dynamic Diagram Workflow:** The `sdl trace` and `sdl diagram dynamic` commands are fully implemented.

---
## 1. Implement the Interactive Recipe Runner

With the `Canvas` API proven, the next step is to build the user-facing frontend for it.

*   **Create the `sdl execute` command:**
    *   Create the file `cmd/sdl/commands/execute.go`.
    *   The command will take one argument: the path to a `.recipe` file.
*   **Build the Recipe Parser:**
    *   The `Run` function for the command will read the recipe file line by line.
    *   It will parse each line into a command and its arguments (e.g., using `strings.Fields`).
*   **Integrate the `Canvas` Engine:**
    *   The command will instantiate a `console.Canvas` object.
    *   Based on the parsed command from the recipe file, it will call the corresponding `Canvas` method (e.g., `canvas.Load(...)`, `canvas.Set(...)`, `canvas.Run(...)`).

## 2. Model System Capacity

*   **Enhance Queuing Primitives:** Refactor `ResourcePool.Acquire` and `Queue.Dequeue` to return an `Outcomes[Duration]` representing the distribution of time spent waiting in the queue.
*   **Create Capacity-Aware Models:** Update the `.sdl` examples (like `Disk`) to embed `ResourcePool` components to model contention and queuing delays under load.
*   **Validate with a Test:** Create a new test to verify that increasing the `ArrivalRate` parameter on a component's internal `ResourcePool` correctly increases its end-to-end latency.

## 3. Runtime and Evaluation Enhancements (`runtime` package)

*   **Full Concurrency Model Implementation:**
    *   **`gobatch N { ... }`**: Implement the runtime logic to properly model the execution of N parallel operations.
    *   **Native Aggregators**: Develop standard aggregator methods (`WaitAll`, `WaitAny`, `Quorum`).

## Priority Focus (Immediate to Medium Term):

1.  **`sdl execute` Command.** This is the highest priority as it delivers the interactive analysis workflow planned for the conference.
2.  **Model System Capacity.** This is a critical modeling feature that builds directly on the interactive capabilities of the recipe runner.
3.  **Full Concurrency Model.** This remains the most significant feature gap in the simulation engine.
