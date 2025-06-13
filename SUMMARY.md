# SDL (System Design Language) Project Summary

**Version:** As of conference workshop preparation with complete web Canvas visualization dashboard.

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
*   **`./console`**: The workshop engine providing a stateful, API-driven system (`Canvas`) for interactive analysis. This powers both the CLI-based analysis and the new web dashboard's real-time parameter modification, scenario execution, and live visualization updates that make system design concepts immediately tangible.
*   **`./cmd/sdl/commands/serve.go`**: Web server providing HTTP API and WebSocket support for the interactive dashboard, enabling browser-based system visualization and real-time parameter manipulation.
*   **`./web/`**: Complete TypeScript + Tailwind frontend providing the "Incredible Machine" style dashboard with 3-panel layout, real-time charts, and interactive parameter controls for workshop demonstrations.
*   **`./types`**: Shared data structures between backend and frontend, ensuring type-safe communication across the web stack.
*   **`./viz`:** A top-level library for generating all visualizations, including line charts, bar charts (histograms), and static/dynamic diagrams.
*   **`./cmd/sdl` (Command Line Interface):** The main user-facing tool, built with Cobra. It provides a suite of commands for a comprehensive workflow (`validate`, `run`, `trace`, `plot`, `diagram`, `serve`) enabling both CLI-based analysis and web-based interactive dashboards.
*   **`./examples`:** Contains sample `.sdl` files, Go API usage examples, and the flagship Netflix streaming service demo (`examples/netflix/`) that showcases traffic spike scenarios, capacity modeling, and performance optimization for workshop demonstrations.

**3. Current Status & Conference Workshop Focus:**

*   **🎉 WEB DASHBOARD COMPLETE:** Full "Incredible Machine" style web interface implemented with TypeScript + Tailwind, providing real-time parameter manipulation, live performance charts, and 3-panel dashboard layout. The web Canvas visualization is production-ready for workshop demonstrations.
*   **Workshop-Ready Foundation:** The project has robust libraries for performance modeling, complete visualization workflows, and both CLI and web-based interactive analysis capabilities specifically designed for live demonstrations and educational scenarios.
*   **Multi-Interface Canvas API:** The `console.Canvas` API powers both CLI commands and the new web dashboard, providing consistent real-time parameter modification, scenario execution, and visualization updates across all interfaces.
*   **Contacts Service Validation (NEW):** Simple 2-tier phone lookup service (`examples/contacts/`) created and fully validated to test Canvas API capabilities, providing a clean foundation for web dashboard development and workshop scenarios.
*   **Netflix Demo Scenario:** A comprehensive Netflix-style streaming service model (`examples/netflix/`) with CDN capacity modeling, database bottlenecks, video encoding pipelines, and traffic spike scenarios available for complex workshop demonstrations.
*   **Capacity Modeling with Queuing Theory:** Full M/M/c queuing implementation enables realistic analysis of system capacity limits, demonstrating concepts like cache hit rate impact, database connection pooling, and overload conditions that are crucial for system design interviews.
*   **Production-Ready Web Stack:** Complete TypeScript frontend with Chart.js visualization, WebSocket real-time updates, HTTP API backend, and comprehensive testing suite. Single-command deployment: `./sdl serve --port 8080`.
*   **Conference Goal ACHIEVED:** SDL now provides the ultimate system design interview preparation platform with both powerful CLI tools and an intuitive web interface that makes complex distributed systems concepts immediately visible and interactive.
