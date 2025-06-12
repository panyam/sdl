# SDL (System Design Language) Project Summary

**Version:** As of the implementation of native methods and histogram plotting.

**1. Vision & Goal:**

*   **Purpose:** The SDL project provides a specialized language and a powerful toolchain for modeling, simulating, and analyzing the performance characteristics (e.g., latency, availability, result distribution) of distributed systems.
*   **Use Cases:** Enable rapid analysis of system designs, bottleneck identification, SLO evaluation, performance exploration under different configurations, and generating diagrams from the definitions.

**2. Overall Architecture & Key Packages:**

The project is a Go-based system composed of several key packages:

*   **`./core` & `./components`:** Provide a Go API for probabilistic performance modeling (`core`) and a library of pre-built native components like disks, caches, and queues (`components`).
*   **`./decl` & `./parser`:** Define the language's Abstract Syntax Tree (AST) and implement the `goyacc`-based parser to generate the AST from `.sdl` files. The language now supports top-level `native method` declarations for defining global functions implemented in Go.
*   **`./loader`:** Manages the loading of SDL files, resolving imports, and orchestrating the crucial type inference and validation phase for all declarations, including native methods.
*   **`./runtime`:** The execution engine for SDL models. It features `SimpleEval`, an interpreter that walks the AST to run simulations. It has been instrumented with a tracer and now includes a registry for calling global native methods (like `delay()` and `log()`).
*   **`./viz`:** A top-level library for generating all visualizations. It contains the logic for creating static diagrams, dynamic sequence diagrams, and time-series plots. The plotter now supports generating both line charts and bar charts (for histograms).
*   **`./cmd/sdl` (Command Line Interface):** The main user-facing tool, `sdl`, built with Cobra. It provides a suite of commands for a comprehensive workflow:
    *   **`validate`**: Parses and type-checks SDL files.
    *   **`list` / `describe`**: Inspects and prints details about components and systems.
    *   **`run`**: A powerful simulation runner that produces a JSON file with detailed per-run statistics.
    *   **`trace`**: A single-run execution tracer that produces a detailed JSON event file.
    *   **`plot`**: A post-simulation analysis tool that consumes data from `sdl run` to generate SVG plots. It can generate time-series latency/count plots and **histograms**.
    *   **`diagram`**: A powerful diagramming tool that consumes data from `sdl trace` (for dynamic diagrams) or directly from `.sdl` files (for static diagrams).
*   **`./examples`:** Contains sample `.sdl` files and Go API usage examples.

**3. Current Status & Future Direction:**

*   **Mature Core & Workflows:** The project has robust libraries for performance modeling and two complete, decoupled workflows: statistical simulation (`run` -> `plot`) and single-run tracing (`trace` -> `diagram`).
*   **Expressive Language:** The grammar has been simplified with the introduction of `native method`s, allowing common operations like `delay` to be standard function calls.
*   **Advanced Visualization:** The `viz` package can generate multi-series SVG line charts, SVG bar charts (for histograms), and multiple static/dynamic diagram formats.
*   **Next Evolution: Interactive Analysis:** To enhance developer experience and enable rapid, iterative analysis, the next major feature will be an **interactive recipe runner**. This will allow users to script a series of actions—loading models, modifying parameters in-memory, running simulations, and generating comparison plots—from a single command, providing a notebook-like experience.
*   **Known Limitations:** The runtime correctly visualizes concurrency (`gobatch`/`wait`), but the underlying execution model for `gobatch` still simulates a single representative path. Fully modeling N parallel executions in the runtime is the next major step.
