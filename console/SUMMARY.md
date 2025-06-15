# Console Package Summary (`console` package)

**Purpose:**

This package provides the core engine for interactive workshop demonstrations and system design interview coaching. Its primary goal is to create a stateful, long-running session that enables real-time parameter modification, immediate visualization updates, and seamless live demonstrations that make complex distributed systems concepts immediately tangible for audiences. **Now powers CLI commands, web server, and dashboard interfaces with RESTful API architecture.**

**Core Components & Files:**

*   **`canvas.go`**: Contains the main `Canvas` struct, which acts as the central session manager.
*   **`canvas_web.go`**: Complete web server implementation with RESTful Canvas API endpoints, goutils WebSocket integration, consolidated HTTP/WebSocket handling for dashboard communication, and Canvas state save/restore functionality for session persistence.
    *   **`Canvas` Struct:** This is the main entry point for the console API. It holds the session state, including:
        *   An instance of the `loader.Loader` and `runtime.Runtime`.
        *   The currently active loaded file (`*loader.FileStatus`) and instantiated system (`*runtime.SystemInstance`) with its fully populated environment.
        *   Traffic generators for continuous load simulation with start/stop/pause control.
        *   Measurement configurations for dynamic chart creation in the dashboard.
        *   A map for `sessionVars` to store the results of `run` commands or other operations, allowing them to be referenced in subsequent commands (e.g., for plotting).
    *   **Canvas API Methods:** The struct exposes a workshop-optimized API for live demonstrations:
        *   `Load(filePath string)`: Loads and validates an SDL file for demo scenarios.
        *   `Use(systemName string)`: Sets and initializes the active system for the workshop session.
        *   `Set(path string, value interface{})`: The star of live demos - modifies component parameters in real-time using dot-notation paths (e.g., "videoService.cdn.pool.ArrivalRate") with immediate effect on subsequent simulations. Essential for showing traffic spike scenarios and capacity modeling.
        *   `Run(varName string, target string, ...)`: Executes simulations with results stored for comparison plotting. Optimized for rapid iteration during presentations.
        *   `Plot(...)`: Generates immediate visualizations from session data, enabling before/after comparisons that create "aha moments" for audiences.
        *   **Traffic Generation Methods:** `AddGenerator`, `RemoveGenerator`, `UpdateGenerator`, `PauseGenerator`, `ResumeGenerator`, `StartGenerators`, `StopGenerators` for continuous load simulation.
        *   **Measurement Methods:** `AddMeasurement`, `RemoveMeasurement`, `GetMeasurements` for dynamic dashboard chart configuration.
        *   **Enhanced State Management:** `Save`, `Restore` for complete session persistence including loaded files, active system, generators, measurements, session variables, and system parameters. Enables browser reconnection with full state recovery.

**Relationship with Other Packages:**

*   **`cmd/sdl`**: Powers the serve command hosting Canvas web server and all client CLI commands. Direct commands (`sdl load`, `sdl gen add`, `sdl measure start`) use RESTful Canvas API for clean server/client separation. The consolidated web server architecture delegates all HTTP/WebSocket handling to Canvas.
*   **`loader` & `runtime`**: The `Canvas` orchestrates these packages to seamlessly load workshop scenarios and execute rapid parameter modifications during live demos.
*   **`viz`**: The `Canvas.Plot` method generates immediate visualizations that demonstrate system behavior changes, creating compelling visual storytelling for workshop audiences.

**Current Status:**

*   **Workshop-Ready:** The `Canvas` API is fully functional and conference-presentation ready, validated through comprehensive tests including the Netflix traffic spike demo scenario.
*   **RESTful Canvas API:** Complete implementation of stateless REST endpoints for all Canvas operations. Frontend can control entire Canvas lifecycle through HTTP while receiving real-time updates via WebSocket.
*   **Web Dashboard Integration:** Canvas now fully integrated with TypeScript frontend. Dashboard automatically loads Canvas state on startup, displays system architecture dynamically, and provides real-time traffic generation controls.
*   **System Visualization:** Canvas provides complete component topology data that frontend renders to match `sdl diagram` output. All components, connections, and dependencies are visualized in the dashboard.
*   **Direct CLI Commands:** Shell-native commands (`sdl load`, `sdl gen add`, `sdl measure start`) provide clean, scriptable interface using RESTful Canvas API with environment variable configuration support.
*   **Enhanced State Persistence:** Complete Canvas state tracking including system parameters, enabling perfect session recovery when browser reconnects mid-session.
*   **Real-Time Parameter Modification:** The `Set` method delivers flawless live demo performance, enabling instant parameter changes with immediate simulation impact that creates compelling workshop moments.
*   **Traffic Generation:** Complete generator lifecycle management with start/stop/pause controls. Multiple generators can run simultaneously with different configurations.
*   **DuckDB Measurement System:** Complete call-based measurement architecture with custom MeasurementTracer extending ExecutionTracer. Captures simulation metrics directly from SDL call evaluation points with time-series database storage for advanced analytics.
*   **Measurement Console Commands:** Full `measure add/remove/list/clear/stats/sql` command suite with intelligent tab completion for targets and metric types (latency, throughput, errors). Space-delimited syntax integrates seamlessly with existing console workflow.
*   **Canvas.Run() Auto-Injection:** Smart measurement tracer integration that automatically enables tracing when measurements are registered, with zero-overhead fallback to standard execution when no measurements present. Post-processing architecture extracts data from ExecutionTracer events.
*   **End-to-End Analytics Workflow:** Complete `measure → run → analyze` pipeline with DuckDB storage, percentile calculations (P50/P90/P95/P99), custom SQL queries, and time-series data retrieval. Validated with comprehensive test suite covering multiple simulation runs and data accumulation.
*   **Traffic Generation:** Fully functional generators that execute Canvas.Run() at configured rates, creating actual simulation load with measurement tracing. Supports real-time rate adjustments and pause/resume operations.
*   **Live Measurement Dashboard:** Web dashboard fetches real measurement data from DuckDB via REST API endpoint `/api/measurements/{target}/data`. Charts update every 2 seconds showing actual latency distributions from simulations.
*   **Shared State Synchronization:** CLI commands instantly broadcast to web dashboard via WebSocket, enabling perfect side-by-side demonstrations with clean shell integration.
*   **Capacity Modeling Integration:** Full support for M/M/c queuing demonstrations, allowing audiences to see how traffic spikes overwhelm systems and how capacity scaling and caching optimizations restore performance.
*   **Demo Scenario Validation:** Netflix streaming service model provides complete workshop narrative from baseline performance through traffic spikes, optimization strategies, and failure scenarios.
*   **Ready for Conference:** All primitives tested for rapid iteration, edge case handling, and audience Q&A scenarios. The foundation for "Building an Open Source System Design Interview Coach With Interactive Simulations" is solid and reliable.
