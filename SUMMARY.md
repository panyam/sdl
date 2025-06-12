# SDL (System Design Language) Project Summary

**Version:** As of the implementation of native methods and the `console.Canvas` proof-of-concept.

**1. Vision & Goal:**

*   **Purpose:** The SDL project provides a specialized language and a powerful toolchain for modeling, simulating, and analyzing the performance characteristics (e.g., latency, availability, result distribution) of distributed systems.
*   **Use Cases:** Enable rapid analysis of system designs, bottleneck identification, SLO evaluation, performance exploration under different configurations, and generating diagrams from the definitions.

**2. Overall Architecture & Key Packages:**

The project is a Go-based system composed of several key packages:

*   **`./core` & `./components`:** Provide a Go API for probabilistic performance modeling (`core`) and a library of pre-built native components like disks, caches, and queues (`components`).
*   **`./decl` & `./parser`:** Define the language's Abstract Syntax Tree (AST) and implement the `goyacc`-based parser to generate the AST from `.sdl` files. The language now supports top-level `native method` declarations.
*   **`./loader`:** Manages the loading of SDL files, resolving imports, and orchestrating type inference and validation.
*   **`./runtime`:** The execution engine for SDL models. It features `SimpleEval`, an interpreter that walks the AST to run simulations, and a registry for calling native methods.
*   **`./console`**: A new package that provides a stateful, API-driven engine (`Canvas`) for interactive analysis. This forms the backend for tools like the upcoming recipe runner, allowing for in-memory modification of models and iterative simulation.
*   **`./viz`:** A top-level library for generating all visualizations, including line charts, bar charts (histograms), and static/dynamic diagrams.
*   **`./cmd/sdl` (Command Line Interface):** The main user-facing tool, built with Cobra. It provides a suite of commands for a comprehensive workflow (`validate`, `run`, `trace`, `plot`, `diagram`), and will soon include a recipe runner (`execute`).
*   **`./examples`:** Contains sample `.sdl` files and Go API usage examples.

**3. Current Status & Future Direction:**

*   **Mature Core & Workflows:** The project has robust libraries for performance modeling and two complete, decoupled workflows: statistical simulation (`run` -> `plot`) and single-run tracing (`trace` -> `diagram`).
*   **Interactive Analysis Engine (Canvas API):** A proof-of-concept for the new `console.Canvas` API is complete. This proves the feasibility of an interactive workflow where a user can load a model, programmatically modify component parameters, run multiple simulations, and generate comparison plots, all within a single, stateful session.
*   **Next Evolution: Recipe Runner:** The next major feature will be the `sdl execute` command, a user-facing tool that runs "recipe" files. This will provide a powerful, scriptable, and notebook-like experience for system analysis, built directly on top of the `Canvas` API.
*   **Known Limitations:** The runtime correctly visualizes concurrency (`gobatch`/`wait`), but the underlying execution model for `gobatch` still simulates a single representative path. Fully modeling N parallel executions in the runtime is the next major step.
