# SDL (System Design Language) Project Summary

**Version:** RESTful Canvas API with goutils WebSocket integration - Production-ready system design platform.

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
*   **`./console`**: The core engine providing both the Canvas API and web server implementation. Features include stateful Canvas sessions, RESTful traffic generation APIs, measurement management, and production-grade WebSocket handling using goutils library for real-time updates.
*   **`./console/canvas_web.go`**: Complete web server implementation with RESTful Canvas API endpoints, goutils WebSocket integration, and consolidated HTTP/WebSocket handling for dashboard communication.
*   **`./web/`**: Complete TypeScript + Tailwind frontend providing the "Incredible Machine" style dashboard with 3-panel layout, real-time charts, and interactive parameter controls for workshop demonstrations.
*   **`./types`**: Shared data structures between backend and frontend, ensuring type-safe communication across the web stack.
*   **`./viz`:** A top-level library for generating all visualizations, including line charts, bar charts (histograms), and static/dynamic diagrams.
*   **`./cmd/sdl` (Command Line Interface):** The main user-facing tool, built with Cobra. It provides a suite of commands for a comprehensive workflow (`validate`, `run`, `trace`, `plot`, `diagram`, `serve`, `console`) enabling CLI-based analysis, interactive REPL sessions, and web-based dashboards.
*   **`./examples`:** Contains sample `.sdl` files, Go API usage examples, and the flagship Netflix streaming service demo (`examples/netflix/`) that showcases traffic spike scenarios, capacity modeling, and performance optimization for workshop demonstrations.

**3. Current Status & Architecture:**

*   **RESTful Canvas API:** Complete implementation of stateless, RESTful endpoints for traffic generation and measurement management. Eliminates WebSocket brittleness by using HTTP for control operations and WebSocket only for live data updates.
*   **goutils WebSocket Integration:** Production-grade WebSocket implementation using the goutils library with proper lifecycle hooks (OnStart, OnClose, OnTimeout, HandleMessage) and automatic connection management.
*   **Consolidated Web Server:** All HTTP/WebSocket functionality moved to `console/canvas_web.go` for better architecture. Simple serve command delegates to Canvas web router.
*   **Enhanced Canvas API:** Extended Canvas with 8 new methods for traffic generation (AddGenerator, RemoveGenerator, UpdateGenerator, PauseGenerator, ResumeGenerator, StartGenerators, StopGenerators) and measurement management (AddMeasurement, RemoveMeasurement, GetMeasurements).
*   **Canvas State Management:** Complete save/restore functionality enables stateless operations and session recovery. Canvas state includes loaded files, active system, generators, measurements, and session variables.
*   **Frontend Canvas Integration:** Web dashboard now fully integrated with RESTful Canvas API. Dashboard starts empty and automatically loads Canvas state. All control operations (load, use, set, run) work through REST endpoints with real-time WebSocket updates for metrics.
*   **System Architecture Visualization:** Frontend displays full component topology matching the `sdl diagram` command output. Shows complete system structure with all components, connections, and dependencies in the prominent left panel.
*   **2-Row Web Dashboard:** TypeScript + Tailwind frontend with clean layout - Row 1 (System Architecture + Controls), Row 2 (Dynamic Metrics Grid). Real-time updates via WebSocket, control operations via REST API. Removed all hardcoded service-specific references for true generic operation.
*   **Interactive REPL Console:** New `sdl console` command provides interactive REPL with shared Canvas state. Features command history, real-time web dashboard synchronization, and eliminates need for curl commands - perfect for side-by-side demonstrations.
*   **Enhanced Canvas State Persistence:** Complete Canvas state management with parameter tracking. Includes loaded files, active system, generators, measurements, session variables, and system parameters. Enables perfect session recovery and browser reconnection.
*   **Real-time Synchronization:** Console commands instantly update web dashboard via WebSocket broadcasting. Browser can reconnect mid-session and automatically sync with current REPL state.
*   **Production Deployment:** Single-command deployment with `./sdl serve --port 8080` or interactive mode with `./sdl console --port 8080`. Serves RESTful Canvas API with proper CORS handling and real-time WebSocket updates.
*   **Demonstration Examples:** Netflix streaming service and contacts lookup service provide comprehensive examples for system design interview scenarios with traffic generation and performance analysis.
