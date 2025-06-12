# SDL Project: Next Steps & Outstanding Work

This document outlines the key areas for future development and refinement in the SDL project.

---
## Completed

*   **`[COMPLETED]` Interactive Session Engine (`console.Canvas`):**
    *   The `console.Canvas` API is now a stable and robust engine for interactive analysis.
    *   Core workflow (`Load`, `Use`, `Set`, `Run`, `Plot`) is fully functional and tested.
    *   The `Set` command is particularly powerful, supporting runtime modification of parameters on both top-level instances and their nested dependencies (e.g., `app.cache.HitRate`).
*   **`[COMPLETED]` Loader & Runtime Enhancements:**
    *   The `loader` now correctly handles imports of all declaration types, including `native method`.
    *   The `runtime` can now instantiate all available native components from the `components` library.
*   **`[COMPLETED]` Native Methods & Grammar Simplification:**
    *   Added support for top-level `native method` declarations.
    *   Refactored the `delay` statement into a standard native method call, simplifying the grammar.
*   **`[COMPLETED]` Plotting & Visualization:**
    *   The `sdl plot` command supports latency, count, and histogram plots.
    *   The `sdl diagram` command supports static (architecture) and dynamic (sequence) diagrams.
*   **`[COMPLETED]` Capacity Modeling:**
    *   The `ResourcePool` component now supports full M/M/c queuing calculations.
    *   Canvas API correctly handles capacity modeling workflows with proper instance management.
    *   Boolean evaluation enhanced to handle `Outcomes[Bool]` types with embedded latency.
    *   Duration parsing fixed to use correct unit conversion (seconds as base unit).
    *   All Canvas tests pass, demonstrating realistic queuing delays under different loads.

---
## 1. Implement the Interactive Recipe Runner

With the `Canvas` API complete and stable, the next step is to build the user-facing frontend for it.

*   **Create the `sdl execute` command:**
    *   Create the file `cmd/sdl/commands/execute.go`.
    *   The command will take one argument: the path to a `.recipe` file.
*   **Build the Recipe Parser:**
    *   The `Run` function for the command will read the recipe file line by line.
    *   It will parse each line into a command and its arguments (e.g., using `strings.Fields`).
*   **Integrate the `Canvas` Engine:**
    *   The command will instantiate a `console.Canvas` object.
    *   Based on the parsed command from the recipe file, it will call the corresponding `Canvas` method (e.g., `canvas.Load(...)`, `canvas.Set(...)`, `canvas.Run(...)`).

## 2. ~~Model System Capacity~~ `[COMPLETED]`

*   ~~**Enhance Queuing Primitives:** Refactor `ResourcePool.Acquire` and `Queue.Dequeue` to return an `Outcomes[Duration]` representing the distribution of time spent waiting in the queue.~~
*   ~~**Create Capacity-Aware Models:** Update the `.sdl` examples (like `Disk`) to embed `ResourcePool` components to model contention and queuing delays under load.~~
*   ~~**Validate with a Test:** Create a new test to verify that increasing the `ArrivalRate` parameter on a component's internal `ResourcePool` correctly increases its end-to-end latency.~~

## 3. Runtime and Evaluation Enhancements (`runtime` package)

*   **Full Concurrency Model Implementation:**
    *   **`gobatch N { ... }`**: Implement the runtime logic to properly model the execution of N parallel operations.
    *   **Native Aggregators**: Develop standard aggregator methods (`WaitAll`, `WaitAny`, `Quorum`).

## Priority Focus (Immediate to Medium Term):

1.  **`sdl execute` Command.** This is the highest priority as it delivers the interactive analysis workflow planned for the conference.
2.  ~~**Model System Capacity.**~~ `[COMPLETED]` ~~This is a critical modeling feature that builds directly on the interactive capabilities of the recipe runner.~~
3.  **Full Concurrency Model.** This remains the most significant feature gap in the simulation engine.
