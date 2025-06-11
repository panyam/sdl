# SDL (System Design Language) Project Summary

**Version:** As of the implementation of the tracing and dynamic diagramming workflow.

**1. Vision & Goal:**

*   **Purpose:** The SDL project provides a specialized language and a powerful toolchain for modeling, simulating, and analyzing the performance characteristics (e.g., latency, availability, result distribution) of distributed systems.
*   **Use Cases:** Enable rapid analysis of system designs, bottleneck identification, SLO evaluation, performance exploration under different configurations, and generating diagrams from the definitions.

**2. Overall Architecture & Key Packages:**

The project is a Go-based system composed of several key packages:

*   **`./core` & `./components`:** Provide a Go API for probabilistic performance modeling (`core`) and a library of pre-built native components like disks, caches, and queues (`components`).
*   **`./decl` & `./parser`:** Define the language's Abstract Syntax Tree (AST) and implement the `goyacc`-based parser to generate the AST from `.sdl` files.
*   **`./loader`:** Manages the loading of SDL files, resolving imports, and orchestrating the crucial type inference and validation phase.
*   **`./runtime`:** The execution engine for SDL models. It features `SimpleEval`, an interpreter that walks the AST to run simulations. It has been instrumented with a tracer to capture detailed execution events.
*   **`./viz`:** A new top-level library for generating all visualizations. It contains the logic for creating static diagrams (DOT, Mermaid, etc.), dynamic sequence diagrams, and time-series plots. It decouples the presentation logic from the CLI.
*   **`./cmd/sdl` (Command Line Interface):** The main user-facing tool, `sdl`, built with Cobra. It provides a suite of commands for a comprehensive workflow:
    *   **`validate`**: Parses and type-checks SDL files.
    *   **`list` / `describe`**: Inspects and prints details about components and systems.
    *   **`run`**: A powerful simulation runner that produces a JSON file with detailed per-run statistics.
    *   **`trace`**: A single-run execution tracer that produces a detailed JSON event file.
    *   **`plot`**: A post-simulation analysis tool that consumes data from `sdl run` to generate SVG plots.
    *   **`diagram`**: A powerful diagramming tool that consumes data from `sdl trace` (for dynamic diagrams) or directly from `.sdl` files (for static diagrams).
*   **`./examples`:** Contains sample `.sdl` files and Go API usage examples.

**3. Current Status & Key Features:**

*   **Mature Core & Frontend:** The project has robust libraries for performance modeling, a mature parser, and a powerful type inference system.
*   **Complete Analysis Workflows:** Two complete, decoupled workflows are now implemented:
    1.  **Performance Simulation:** `sdl run` -> `sdl plot` for statistical analysis.
    2.  **Execution Tracing:** `sdl trace` -> `sdl diagram dynamic` for visualizing single-run behavior as a sequence diagram.
*   **Advanced Visualization:** The `viz` package can generate multi-series SVG charts, multiple static diagram formats (DOT, Mermaid, SVG, Excalidraw), and Mermaid sequence diagrams that correctly represent concurrency.
*   **Known Limitations:** The runtime correctly visualizes concurrency (`gobatch`/`wait`), but the underlying execution model for `gobatch` still simulates a single representative path. Fully modeling N parallel executions in the runtime is the next major step.

This summary provides a high-level overview of the SDL project, its architecture, and its powerful simulation and analysis capabilities.
