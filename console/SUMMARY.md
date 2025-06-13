# Console Package Summary (`console` package)

**Purpose:**

This package provides the core engine for interactive workshop demonstrations and system design interview coaching. Its primary goal is to create a stateful, long-running session that enables real-time parameter modification, immediate visualization updates, and seamless live demonstrations that make complex distributed systems concepts immediately tangible for audiences. **Now powers both CLI and web interfaces.**

**Core Components & Files:**

*   **`canvas.go`**: Contains the main `Canvas` struct, which acts as the central session manager.
*   **`canvas_web.go`**: Complete web server implementation with RESTful Canvas API endpoints, goutils WebSocket integration, and consolidated HTTP/WebSocket handling for dashboard communication.
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
        *   **State Management:** `SaveState`, `RestoreState` for session persistence and recovery.

**Relationship with Other Packages:**

*   **`cmd/sdl`**: Powers the serve command which provides the complete web dashboard experience. The consolidated web server architecture delegates all HTTP/WebSocket handling to Canvas.
*   **`loader` & `runtime`**: The `Canvas` orchestrates these packages to seamlessly load workshop scenarios and execute rapid parameter modifications during live demos.
*   **`viz`**: The `Canvas.Plot` method generates immediate visualizations that demonstrate system behavior changes, creating compelling visual storytelling for workshop audiences.

**Current Status:**

*   **Workshop-Ready:** The `Canvas` API is fully functional and conference-presentation ready, validated through comprehensive tests including the Netflix traffic spike demo scenario.
*   **RESTful Canvas API:** Complete implementation of stateless REST endpoints for all Canvas operations. Frontend can control entire Canvas lifecycle through HTTP while receiving real-time updates via WebSocket.
*   **Web Dashboard Integration:** Canvas now fully integrated with TypeScript frontend. Dashboard automatically loads Canvas state on startup, displays system architecture dynamically, and provides real-time traffic generation controls.
*   **System Visualization:** Canvas provides complete component topology data that frontend renders to match `sdl diagram` output. All components, connections, and dependencies are visualized in the dashboard.
*   **Real-Time Parameter Modification:** The `Set` method delivers flawless live demo performance, enabling instant parameter changes with immediate simulation impact that creates compelling workshop moments.
*   **Traffic Generation:** Complete generator lifecycle management with start/stop/pause controls. Multiple generators can run simultaneously with different configurations.
*   **Measurement Management:** Dynamic chart creation based on Canvas measurements. Dashboard automatically creates and updates charts as new measurements are added.
*   **Capacity Modeling Integration:** Full support for M/M/c queuing demonstrations, allowing audiences to see how traffic spikes overwhelm systems and how capacity scaling and caching optimizations restore performance.
*   **Demo Scenario Validation:** Netflix streaming service model provides complete workshop narrative from baseline performance through traffic spikes, optimization strategies, and failure scenarios.
*   **Ready for Conference:** All primitives tested for rapid iteration, edge case handling, and audience Q&A scenarios. The foundation for "Building an Open Source System Design Interview Coach With Interactive Simulations" is solid and reliable.
