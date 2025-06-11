# SDL Command Line Interface (CLI) Package Summary (`cmd/sdl`)

**Purpose:**

This package implements the main command-line interface (CLI) tool for the System Design Language (SDL) project, named `sdl`. It allows users to interact with the SDL toolchain to parse, validate, inspect, generate diagrams, and run simulations.

**Core Structure & Files:**

*   **`main.go`**: The entry point for the CLI application. It simply calls `commands.Execute()` to run the root Cobra command.
*   **`commands/` (sub-package):** Contains the definitions for all CLI commands and their logic.
    *   **`root.go`**: Defines the root command (`sdl`) using `github.com/spf13/cobra`. It sets up persistent flags, such as `--file` (`-f`) for specifying the input DSL file path.
    *   **`validate.go`**: Implements `sdl validate <dsl_file_path...>` for parsing and semantic checks using `loader.Loader`.
    *   **`list.go`**: Implements `sdl list <entity_type>` to list defined entities from a DSL file.
    *   **`describe.go`**: Implements `sdl describe <entity_type> <entity_name>` to show detailed information about a specific entity.
    *   **`run.go`**: Implements `sdl run ...` to perform large-scale simulations. It produces a detailed JSON file containing the results (latency, return value, etc.) for each run.
    *   **`trace.go`**: Implements `sdl trace ...` to perform a single-run execution of a method and save the detailed event trace to a JSON file.
    *   **`plot.go`**: A versatile plotting command that consumes the JSON output from `sdl run`. It acts as a client to the `viz` package to generate SVG plots for latency or result counts.
    *   **`diagram.go`**: A command that generates diagrams by calling generators in the `viz` package. It can create `static` diagrams from SDL source or `dynamic` sequence diagrams from the JSON output of `sdl trace`.
    *   **`shared_types.go`**: Defines shared data structures, like `RunResult`, used for serialization between commands.

**Key Functionality & Features:**

*   **Modular Commands:** Uses Cobra for a structured command system.
*   **Separation of Concerns:** The CLI commands are thin wrappers that orchestrate the `loader`, `runtime`, and `viz` packages. Heavy lifting (simulation, tracing, diagram generation) is delegated to the libraries.
*   **Decoupled Workflows:** The CLI promotes a two-step process for analysis:
    1.  **Data Generation:** `sdl run` or `sdl trace` produce JSON data files.
    2.  **Visualization:** `sdl plot` or `sdl diagram dynamic` consume these files to generate visuals.
*   **Rich Analysis:** The commands support statistical performance simulation (`run`/`plot`) and detailed single-run debugging (`trace`/`diagram dynamic`).

**Workflows:**

1.  **Simulation and Plotting:**
    *   `sdl run Twitter tls.GetTimeline --runs=50000 --out=timeline_results.json -f ...`
    *   `sdl plot --from=timeline_results.json --y-axis=latency -o latency.svg`

2.  **Tracing and Dynamic Diagrams:**
    *   `sdl trace Twitter tls.GetTimeline --out=timeline_trace.json -f ...`
    *   `sdl diagram dynamic --from=timeline_trace.json --format=mermaid -o sequence.md`

**Current Status:**

*   The CLI framework is well-established and has been refactored for clarity.
*   The `run`/`plot` and `trace`/`diagram` workflows are fully implemented.
*   The visualization logic has been successfully moved to the new top-level `viz` package.
*   The next major step is to enhance the `runtime` to fully support concurrent execution beyond the current simulation placeholders.
