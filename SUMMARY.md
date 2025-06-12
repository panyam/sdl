# SDL (System Design Language) Project Summary

**Version:** As of conference workshop preparation with Netflix streaming demo and interactive capacity modeling.

**1. Vision & Goal:**

*   **Primary Mission:** Build the ultimate system design interview coaching platform with interactive simulations and real-time performance visualization.
*   **Conference Focus:** "Building an Open Source System Design Interview Coach With Interactive Simulations" - enabling engineers to visualize and communicate complex systems under pressure.
*   **Core Purpose:** The SDL project provides a specialized language and powerful toolchain for modeling, simulating, and analyzing the performance characteristics (e.g., latency, availability, result distribution) of distributed systems through interactive, "Incredible Machine" style experiences.
*   **Use Cases:** System design interview preparation, rapid analysis of system designs, bottleneck identification, SLO evaluation, performance exploration under different configurations, and generating diagrams from the definitions.

**2. Overall Architecture & Key Packages:**

The project is a Go-based system composed of several key packages:

*   **`./core` & `./components`:** Provide a Go API for probabilistic performance modeling (`core`) and a library of pre-built native components like disks, caches, and queues (`components`).
*   **`./decl` & `./parser`:** Define the language's Abstract Syntax Tree (AST) and implement the `goyacc`-based parser to generate the AST from `.sdl` files. The language now supports top-level `native method` declarations.
*   **`./loader`:** Manages the loading of SDL files, resolving imports from various definition types (`component`, `enum`, `method`), and orchestrating type inference and validation.
*   **`./runtime`:** The execution engine for SDL models. It features `SimpleEval`, an interpreter that walks the AST to run simulations, and a factory for instantiating any registered native component.
*   **`./console`**: The workshop engine providing a stateful, API-driven system (`Canvas`) for interactive analysis. This powers the conference demo's real-time parameter modification, scenario execution, and live visualization updates that make system design concepts immediately tangible.
*   **`./viz`:** A top-level library for generating all visualizations, including line charts, bar charts (histograms), and static/dynamic diagrams.
*   **`./cmd/sdl` (Command Line Interface):** The main user-facing tool, built with Cobra. It provides a suite of commands for a comprehensive workflow (`validate`, `run`, `trace`, `plot`, `diagram`), with upcoming workshop-focused commands (`execute`, `dashboard`, `workshop`) for the conference demonstration.
*   **`./examples`:** Contains sample `.sdl` files, Go API usage examples, and the flagship Netflix streaming service demo (`examples/netflix/`) that showcases traffic spike scenarios, capacity modeling, and performance optimization for workshop demonstrations.

**3. Current Status & Conference Workshop Focus:**

*   **Workshop-Ready Foundation:** The project has robust libraries for performance modeling, complete visualization workflows, and interactive analysis capabilities specifically designed for live demonstrations and educational scenarios.
*   **Netflix Demo Scenario (NEW):** A comprehensive Netflix-style streaming service model (`examples/netflix/`) with CDN capacity modeling, database bottlenecks, video encoding pipelines, and traffic spike scenarios. This serves as the flagship demonstration for showing how system design decisions impact real-world performance.
*   **Interactive Analysis Engine (Canvas API):** The `console.Canvas` API is stable and optimized for workshop use. It supports real-time parameter modification during presentations, rapid scenario execution, and immediate visualization updates that create "aha moments" for audiences.
*   **Capacity Modeling with Queuing Theory:** Full M/M/c queuing implementation enables realistic analysis of system capacity limits, demonstrating concepts like cache hit rate impact, database connection pooling, and CDN overload conditions that are crucial for system design interviews.
*   **Next Evolution: Workshop Tooling:** Immediate focus on `sdl execute` (recipe runner), `sdl dashboard` (multi-panel real-time visualization), and workshop-specific commands that enable smooth, professional conference presentations.
*   **Conference Goal:** Transform SDL into the ultimate system design interview preparation platform, providing the tools that every engineering candidate needs to visualize and communicate complex distributed systems concepts effectively.
