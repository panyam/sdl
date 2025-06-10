# SDL Command Line Interface (CLI) Package Summary (`cmd/sdl`)

**Purpose:**

This package implements the main command-line interface (CLI) tool for the System Design Language (SDL) project, named `sdl`. It allows users to interact with the SDL toolchain to parse, validate, inspect, generate diagrams, and run powerful performance simulations.

**Core Structure & Files:**

*   **`main.go`**: The entry point for the CLI application. It simply calls `commands.Execute()` to run the root Cobra command.
*   **`commands/` (sub-package):** Contains the definitions for all CLI commands and their logic.
    *   **`root.go`**: Defines the root command (`sdl`) using `github.com/spf13/cobra`. It sets up persistent flags, such as `--file` (`-f`) for specifying the input DSL file path.
    *   **`validate.go`**: Implements `sdl validate <dsl_file_path...>` for parsing and semantic checks using `loader.Loader`.
    *   **`list.go`**: Implements `sdl list <entity_type>` to list defined entities from a DSL file.
    *   **`describe.go`**: Implements `sdl describe <entity_type> <entity_name>` to show detailed information about a specific entity.
    *   **`run.go`**: Implements `sdl run <system> <instance> <method>` to perform large-scale simulations. It executes the specified method thousands of times and captures detailed results (latency, return value, errors) for each run into a JSON file.
    *   **`plot.go`**: A versatile plotting command. It can either run a live simulation (similar to its original function) or, more powerfully, read the JSON output from `sdl run` to generate sophisticated plots. It can create latency percentile charts or multi-series count charts (e.g., showing the rate of different `HttpStatusCode` results per second).
    *   **`plotter.go`**: A flexible SVG plotting library that now supports rendering both single-series and multi-series line charts with legends, used by the `plot` command.
    *   **`diagram.go`**: Implements `sdl diagram static ...` to generate static architecture diagrams in various formats (DOT, Mermaid, Excalidraw, SVG). Dynamic diagram generation is a placeholder.
    *   **`trace.go`**: (Placeholder) Implements `sdl trace ...` for detailed single-run execution tracing.
    *   **`shared_types.go`**: Defines shared data structures, like `RunResult`, used for serialization between commands (e.g., between `run` and `plot`).

**Key Functionality & Features:**

*   **Modular Commands:** Uses Cobra for a structured command system.
*   **Separation of Concerns:** A powerful `run` command for expensive simulations and a flexible `plot` command for post-facto analysis and visualization.
*   **Performance Simulation:** `sdl run` allows for large-scale Monte Carlo style simulations to gather statistical data on performance and behavior.
*   **Rich Plotting:** `sdl plot` can visualize not just latency percentiles but also the distribution of outcomes (like return codes) over time from simulation data.
*   **Static Diagram Generation:** `sdl diagram static` can generate diagrams in DOT, Mermaid, Excalidraw (JSON), and SVG formats.
*   **Placeholders for Advanced Features:** Dynamic `diagram` and `trace` commands indicate future capabilities.

**Workflow for Simulation and Plotting:**

1.  **Run Simulation:** A user runs a large-scale simulation and saves the results.
    *   `sdl run Twitter tls GetTimeline --runs=50000 --out=timeline_results.json -f examples/twitter/services.sdl`
2.  **Generate Plots:** The user then uses the `plot` command to analyze the results from different angles without re-running the simulation.
    *   **Plot Latency:** `sdl plot --from=timeline_results.json --y-axis=latency -o latency.svg`
    *   **Plot Result Counts:** `sdl plot --from=timeline_results.json --y-axis=count --group-by=result -o counts.svg`

**Current Status:**

*   The CLI framework is well-established.
*   `validate`, `list`, `describe`, and static `diagram` commands are functional.
*   The `run` and `plot` commands provide a powerful, two-step workflow for performance analysis through simulation.
*   The `plotter` is now capable of multi-series rendering.
*   Dynamic diagrams and full execution tracing via `sdl trace` are the next major features to be implemented.
