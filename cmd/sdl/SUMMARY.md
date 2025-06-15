# SDL Command Line Interface (CLI) Package Summary (`cmd/sdl`)

**Purpose:**

This package implements the main command-line interface (CLI) tool for the System Design Language (SDL) project, with a primary focus on **workshop demonstrations and system design interview coaching**. The CLI enables seamless conference presentations through interactive simulations, real-time parameter modification, and immediate visualization updates that make distributed systems concepts tangible for audiences.

**Core Structure & Files:**

*   **`main.go`**: The entry point for the CLI application. It simply calls `commands.Execute()` to run the root Cobra command.
*   **`commands/` (sub-package):** Contains the definitions for all CLI commands and their logic.
    *   **`root.go`**: Defines the root command (`sdl`) using `github.com/spf13/cobra`. Sets up persistent flags for server/client configuration (`--server`, `--host`, `--port`) with environment variable support.
    *   **`validate.go`**: Implements `sdl validate <dsl_file_path...>` for parsing and semantic checks using `loader.Loader`.
    *   **`list.go`**: Implements `sdl list <entity_type>` to list defined entities from a DSL file.
    *   **`describe.go`**: Implements `sdl describe <entity_type> <entity_name>` to show detailed information about a specific entity.
    *   **`run.go`**: Implements `sdl run ...` to perform large-scale simulations. It produces a detailed JSON file containing the results (latency, return value, etc.) for each run.
    *   **`trace.go`**: Implements `sdl trace ...` to perform a single-run execution of a method and save the detailed event trace to a JSON file.
    *   **`plot.go`**: A versatile plotting command that generates immediate visualizations for workshop demonstrations. Creates comparison plots showing before/after performance that generate "aha moments" for audiences.
    *   **`diagram.go`**: A command that generates system architecture diagrams essential for workshop presentations. Creates static diagrams from SDL source and dynamic sequence diagrams from execution traces.
    *   **`serve.go`**: Starts the SDL Canvas server hosting simulation engine, web dashboard, and REST API endpoints. Supports graceful shutdown and statistics display.
    *   **`api.go`**: Unified API client providing server connection handling and environment variable configuration (CANVAS_SERVER_URL, CANVAS_SERVE_HOST, CANVAS_SERVE_PORT).
    *   **`canvas.go`**: Direct Canvas management commands (`load`, `use`, `set`, `get`, `run`, `info`, `execute`) using REST API instead of local Canvas instance.
    *   **`generators.go`**: Traffic generator management commands (`gen add/list/start/stop/pause/resume/remove`) as direct CLI operations using REST API.
    *   **`measurements.go`**: Measurement system commands (`measure add/list/start/stop/remove/data/stats/sql/clear`) as direct CLI operations using REST API.
    *   **`shared_types.go`**: Defines shared data structures, like `RunResult`, used for serialization between commands.

**🎪 Upcoming Workshop Commands:**
    *   **`execute.go`** (Planned): Recipe runner for scripted workshop demonstrations, enabling smooth progression through demo scenarios with automatic parameter changes and visualization updates.
    *   **`dashboard.go`** (Planned): Multi-panel real-time visualization interface showing latency plots, architecture diagrams, and system metrics simultaneously during live demos.
    *   **`workshop.go`** (Planned): Specialized workshop management commands for scenario loading, guided progression, and audience interaction features.

**Key Functionality & Features:**

*   **Direct CLI Architecture:** Shell-native commands (`sdl load`, `sdl gen add`, `sdl measure start`) provide clean, scriptable interface leveraging native shell features (history, completion, piping).
*   **Server/Client Separation:** Clear separation between server (`sdl serve`) hosting Canvas engine and client commands using REST API for distributed usage and scripting.
*   **Workshop-Optimized Commands:** Cobra-based command system designed for smooth conference presentations and live demonstrations.
*   **Canvas Integration:** CLI commands orchestrate the `console.Canvas` API via REST endpoints for real-time parameter modification and immediate simulation execution.
*   **Live Visualization Pipeline:** Seamless integration between simulation execution and visualization generation, enabling instant before/after comparisons that create educational moments.
*   **Educational Workflows:** Commands designed specifically for system design interview coaching:
    1.  **Interactive Analysis:** Real-time parameter changes with immediate performance impact visualization.
    2.  **Scenario Progression:** Scripted workshop flows from baseline → traffic spike → optimization → scaling.
    3.  **Visual Storytelling:** Architecture diagrams, latency plots, and comparison charts that make abstract concepts concrete.

**Workshop Workflows:**

1.  **Server/Client Demo (Current):**
    *   Terminal 1: `sdl serve --port 8080`
    *   Terminal 2: `sdl load examples/netflix/netflix.sdl`
    *   Terminal 2: `sdl use NetflixSystem`
    *   Terminal 2: `sdl gen add load1 videoService.StreamVideo 10`
    *   Terminal 2: `sdl gen start load1`
    *   Real-time dashboard updates at http://localhost:8080

2.  **Traditional CLI Demo (Current):**
    *   `sdl run examples/netflix/netflix.sdl NetflixSystem videoService.StreamVideo --count=1000`
    *   `sdl plot baseline,traffic_spike --type=comparison --title="Before vs After Traffic Spike"`
    *   `sdl diagram examples/netflix/netflix.sdl NetflixSystem --type=static`

3.  **Scripted Demo (Current):**
    *   `sdl serve --port 8080 &`
    *   `sdl execute examples/netflix/traffic_spike_demo.recipe`
    *   Shell scripting: `for rate in 10 50 100; do sdl set arrival_rate $rate; done`

4.  **Legacy Analysis Workflows:**
    *   `sdl run Twitter tls.GetTimeline --runs=50000 --out=timeline_results.json`
    *   `sdl trace Twitter tls.GetTimeline --out=timeline_trace.json`

**Current Status:**

*   **Workshop Foundation Ready:** Core CLI framework supports all essential workshop operations with existing `run`, `plot`, and `diagram` commands fully functional.
*   **Netflix Demo Validated:** Complete command workflows tested with Netflix streaming service demo, ready for conference presentation.
*   **Canvas Integration Complete:** CLI commands successfully orchestrate the `console.Canvas` API for interactive demonstrations.
*   **Visualization Pipeline Mature:** The `viz` package integration provides immediate, high-quality plots and diagrams essential for workshop impact.

**Conference Preparation Priority:**
*   **Critical Path:** Implement `sdl execute` and `sdl dashboard` commands to enable scripted demos and multi-panel live visualization.
*   **Success Criteria:** Seamless workshop presentation with real-time parameter modification, immediate visual feedback, and compelling audience "aha moments" about distributed systems performance and capacity modeling.
