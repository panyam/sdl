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
*   **`./runtime`:** The execution engine for SDL models. It features `SimpleEval`, an interpreter that walks the AST to run simulations, and a factory for instantiating any registered native component. Now includes clean Tracer interface for pluggable measurement strategies.
*   **`./console`**: The core engine providing Canvas API with gRPC-first architecture. Features include stateful Canvas sessions, gRPC service with HTTP gateway via grpc-gateway, traffic generation with virtual time support, and measurement management via Tracer interface.
*   **`./protos`**: Protocol buffer definitions for all gRPC services and messages. Enables dual-protocol support (gRPC + REST) with auto-generated code via buf.
*   **`./gen`**: Auto-generated Go code from protocol buffers including gRPC clients, servers, and gateway code.
*   **`./web/`**: Complete TypeScript + Tailwind frontend providing the "Incredible Machine" style dashboard with 4-panel DockView layout, real-time charts, compact UI controls, and interactive parameter controls for workshop demonstrations.
*   **`./types`**: Shared data structures between backend and frontend, ensuring type-safe communication across the web stack.
*   **`./viz`:** A top-level library for generating all visualizations, including line charts, bar charts (histograms), and static/dynamic diagrams.
*   **`./cmd/sdl` (Command Line Interface):** The main user-facing tool, built with Cobra. All commands now use gRPC client for communication with Canvas server. Provides comprehensive shell-native commands (`validate`, `run`, `trace`, `plot`, `diagram`, `serve`, `load`, `use`, `gen`, `measure`) with full server/client separation.
*   **`./examples`:** Contains sample `.sdl` files, Go API usage examples, and the flagship Netflix streaming service demo (`examples/netflix/`) that showcases traffic spike scenarios, capacity modeling, and performance optimization for workshop demonstrations.

**3. Current Status & Architecture (June 2025):**

*   **gRPC Migration Complete:** Entire console package migrated from REST to gRPC-first architecture with HTTP gateway support. All CLI commands now use gRPC client for server communication. Clean separation between protocol (gRPC) and business logic (Canvas).
*   **Tracer Interface Architecture:** New pluggable Tracer interface in runtime package enables multiple measurement strategies. MetricTracer implementation processes only Exit events for efficiency. System-specific tracers created on Canvas.Use() for optimal performance.
*   **CLI Commands Unified:** All commands migrated to use gRPC with shared client helper pattern. Canvas selection via --canvas flag with environment variable support. Better error messages and connection handling.
*   **ExecuteTrace RPC:** New trace command for single execution debugging. Captures full call graph with virtual timestamps, parent-child relationships, and return values. Essential for understanding system behavior before high-volume generation.

*   **RESTful Canvas API:** Complete implementation of stateless, RESTful endpoints for traffic generation and measurement management. Eliminates WebSocket brittleness by using HTTP for control operations and WebSocket only for live data updates.
*   **goutils WebSocket Integration:** Production-grade WebSocket implementation using the goutils library with proper lifecycle hooks (OnStart, OnClose, OnTimeout, HandleMessage) and automatic connection management.
*   **Consolidated Web Server:** All HTTP/WebSocket functionality moved to `console/canvas_web.go` for better architecture. Simple serve command delegates to Canvas web router.
*   **Enhanced Canvas API:** Extended Canvas with 8 new methods for traffic generation (AddGenerator, RemoveGenerator, UpdateGenerator, PauseGenerator, ResumeGenerator, StartGenerators, StopGenerators) and measurement management (AddMeasurement, RemoveMeasurement, GetMeasurements).
*   **Canvas State Management:** Complete save/restore functionality enables stateless operations and session recovery. Canvas state includes loaded files, active system, generators, measurements, and session variables.
*   **Frontend Canvas Integration:** Web dashboard now fully integrated with RESTful Canvas API. Dashboard starts empty and automatically loads Canvas state. All control operations (load, use, set, run) work through REST endpoints with real-time WebSocket updates for metrics.
*   **System Architecture Visualization:** Frontend displays full component topology matching the `sdl diagram` command output. Shows complete system structure with all components, connections, and dependencies in the prominent left panel.
*   **4-Panel DockView Dashboard:** Professional TypeScript + Tailwind frontend with DockView layout system - System Architecture, Traffic Generation, Measurements, and Live Metrics panels. Improved bounding box calculations, compact traffic generator controls with numeric input and increment/decrement buttons, real-time updates via WebSocket, and initial file loading support via `--load` flag. Removed all hardcoded service-specific references for true generic operation.
*   **Simplified CLI Architecture:** Replaced complex REPL console with direct CLI commands that leverage native shell features. Commands like `sdl load`, `sdl gen add`, and `sdl measure start` provide clean, scriptable interface with full shell integration (history, completion, piping).
*   **Server/Client Separation:** Clear separation between server (`sdl serve`) hosting Canvas engine and client commands using REST API. Enables distributed usage, scripting, and eliminates client-side state complexity.
*   **Enhanced Canvas State Persistence:** Complete Canvas state management with parameter tracking. Includes loaded files, active system, generators, measurements, session variables, and system parameters. Enables perfect session recovery and browser reconnection.
*   **Real-time Synchronization:** CLI commands instantly update web dashboard via WebSocket broadcasting. Browser can reconnect mid-session and automatically sync with current Canvas state.
*   **Production Deployment:** Single-command deployment with `./sdl serve --port 8080` provides RESTful Canvas API with proper CORS handling and real-time WebSocket updates. Support for initial file loading via `--load file1 file2 ... fileN` flag eliminates manual setup on server restart. All operations accessible via direct CLI commands with `--server` flag.
*   **Demonstration Examples:** Netflix streaming service and contacts lookup service provide comprehensive examples for system design interview scenarios with traffic generation and performance analysis.
*   **Server Stability:** Fixed critical nil pointer dereference in measurement data endpoints by implementing lazy initialization of DuckDB time-series database when first accessed.
*   **Development Infrastructure:** Complete Playwright-based test suite for dashboard validation with automated system loading, API testing, and visual regression capabilities via screenshot comparison.
*   **Advanced Flow Analysis Engine:** Complete implementation of iterative back-pressure and convergence modeling. Original string-based implementation in `runtime/floweval.go` being replaced by runtime-based implementation in `runtime/flowrteval.go`. Features include fixed-point solver with damping, FlowAnalyzable interface for native components, realistic capacity constraint modeling, and validation of multi-component dependency scenarios (A→C, B→C load aggregation).
*   **Runtime-Based Flow Analysis:** New architecture uses actual ComponentInstance objects from SimpleEval's instantiated system graph. Key components include RateMap for type-safe rate tracking, FlowScope for runtime traversal context, and smart NWBase wrapper providing default flow behavior for non-flow-analyzable components. Eliminates duplicate component instances and parameter tracking.
*   **Interactive Flow Visualization:** Real-time traffic flow visualization with numbered execution paths and conditional labels. Features cache-aside pattern implementation, variable outcome tracking for conditional evaluation, nested call ordering (e.g., 3.1 for calls within method 3), and consistent amber edge styling for all flow paths. Dashboard displays actual RPS calculations based on conditional probabilities.
