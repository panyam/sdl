# SDL Command Line Interface (CLI) Package Summary (`cmd/sdl`)

**Purpose:**

This package implements the main command-line interface (CLI) tool for the System Design Language (SDL) project, named `sdl`. It allows users to interact with the SDL toolchain to parse, validate, inspect, generate diagrams, plot performance, and eventually analyze/execute SDL model files.

**Core Structure & Files:**

*   **`main.go`**: The entry point for the CLI application. It simply calls `commands.Execute()` to run the root Cobra command.
*   **`commands/` (sub-package):** Contains the definitions for all CLI commands and their logic.
    *   **`root.go`**: Defines the root command (`sdl`) using the `github.com/spf13/cobra` library. It sets up persistent flags, such as `--file` (`-f`) for specifying the input DSL file path (global `dslFilePath` variable). Provides `AddCommand` for subcommands.
    *   **`validate.go`**: Implements `sdl validate <dsl_file_path...>` for parsing and semantic checks (including type inference) using `loader.Loader`.
    *   **`list.go`**: Implements `sdl list <entity_type>` to list defined entities from a DSL file.
    *   **`describe.go`**: Implements `sdl describe <entity_type> <entity_name>` to show detailed information about a specific entity.
    *   **`plot.go`**: Implements `sdl plot <system> <component> <method>` command. It uses the `runtime.SimpleEval` to run a specified method multiple times in batches, collecting latency data, and then uses `plotter.go` to generate SVG plots (Avg, P50, P90, P99 latency over time).
    *   **`diagram.go`**: Implements `sdl diagram <diagram_type> <system_name>`. 
        *   For `static` diagrams, it parses the SDL file, analyzes the `SystemDecl` for `InstanceDecl`s and their `Overrides` to identify connections, and then generates diagram definitions.
        *   Contains format-specific generators:
            *   **`dot.go`**: Generates DOT language output for Graphviz.
            *   **`mermaid.go`**: Generates Mermaid syntax for flowcharts/graphs.
            *   **`excalidraw.go`**: Generates JSON output compatible with the Excalidraw whiteboard tool. Contains an `ExcalidrawScene` abstraction for building diagrams.
            *   **`svg.go`**: Generates direct SVG vector graphics for static diagrams.
            *   **`diagram_common.go`**: Defines shared structs like `DiagramNode` and `DiagramEdge` used by the static diagram generators.
        *   Dynamic diagram generation is currently a placeholder.
    *   **`plotter.go`**: Contains an SVG plotting library used by the `plot` command to generate line charts from performance data. It includes logic for scaling, axis generation, and SVG template rendering.
    *   **`run.go`**: (Placeholder) Implements `sdl run <system_name> [<analysis_name>]`. Intended to execute analyses.
    *   **`trace.go`**: (Placeholder) Implements `sdl trace <system_name> <method_call_string>`. Intended to trace execution.
    *   **`utils.go`**: Contains utility code, like `SDLParserAdapter` for the `loader.Loader`.

**Key Functionality & Features:**

*   **Modular Commands:** Uses Cobra for a structured command system.
*   **Global File Flag:** Persistent `-f/--file` flag for specifying the primary DSL input file.
*   **Validation:** Comprehensive validation via `sdl validate` leveraging the `loader` and type inference.
*   **AST Inspection:** `list` and `describe` commands.
*   **Performance Plotting:** `sdl plot` executes methods using the `runtime` and generates latency plots.
*   **Static Diagram Generation:** `sdl diagram static` can generate diagrams in DOT, Mermaid, Excalidraw (JSON), and SVG formats by analyzing system instance declarations and their interconnections.
*   **Placeholders for Advanced Features:** `run`, dynamic `diagram`, and `trace` indicate future capabilities dependent on a more complete DSL execution engine.

**Workflow for Static Diagram Generation (`sdl diagram static ...`):**

1.  User runs `sdl diagram static MySystem -f my_system.sdl -o out.dot --format dot`.
2.  `diagramCmd.Run` is executed.
3.  An `loader.NewLoader` is used to load and parse `my_system.sdl` into an AST (`*decl.FileDecl`).
4.  The specified `SystemDecl` (e.g., `MySystem`) is retrieved from the AST.
5.  The command iterates through `InstanceDecl`s in the system to identify nodes (`DiagramNode`).
6.  It then re-iterates, examining `Overrides` in each `InstanceDecl` to find connections between instances, creating `DiagramEdge` entries.
7.  Based on the `--format` flag, the corresponding generator function (e.g., `generateDotOutput`) is called with the collected nodes and edges.
8.  The generated string (DOT, Mermaid, Excalidraw JSON, or SVG) is written to the output file or stdout.

**Current Status:**

*   CLI framework is well-established.
*   `validate`, `list`, `describe`, `plot`, and static `diagram` commands provide significant functionality.
*   The `plot` command demonstrates integration with the `runtime` for basic simulation.
*   Static diagram generation supports multiple output formats.
*   Dynamic diagrams, full `run`, and `trace` capabilities are pending further DSL execution engine development.

**Next Steps (for this package, aligning with project `NEXTSTEPS.MD`):**

*   Implement the actual logic for `run`, dynamic `diagram`, and `trace` by integrating a more complete DSL execution engine/VM.
*   Enhance error reporting for all commands.
*   Add more sophisticated output formatting and layout options for diagrams.
*   Further refine plotting capabilities.
