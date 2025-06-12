# Console Package Summary (`console` package)

**Purpose:**

This package provides the core engine for interactive analysis of SDL models. Its primary goal is to create a stateful, long-running session that decouples the interactive user experience (like a REPL or a recipe runner) from the underlying simulation and loading logic.

**Core Components & Files:**

*   **`canvas.go`**: Contains the main `Canvas` struct, which acts as the central session manager.
    *   **`Canvas` Struct:** This is the main entry point for the console API. It holds the session state, including:
        *   An instance of the `loader.Loader` and `runtime.Runtime`.
        *   The currently active loaded file (`*loader.FileStatus`) and instantiated system (`*runtime.SystemInstance`) with its fully populated environment.
        *   A map for `sessionVars` to store the results of `run` commands or other operations, allowing them to be referenced in subsequent commands (e.g., for plotting).
    *   **Canvas API Methods:** The struct exposes a clear, high-level API for scripting system analysis:
        *   `Load(filePath string)`: Loads and validates an SDL file.
        *   `Use(systemName string)`: Sets and initializes the active system for the session.
        *   `Set(path string, value interface{})`: Modifies a component parameter at runtime using a dot-notation path. This is a powerful feature that can traverse nested component dependencies (e.g., "app.cache.HitRate") and works for both native Go components and user-defined SDL components.
        *   `Run(varName string, target string, ...)`: Executes a simulation and stores the results in a session variable.
        *   `Plot(...)`: Generates plots from data stored in session variables.

**Relationship with Other Packages:**

*   **`cmd/sdl`**: The primary consumer of this package. The upcoming `sdl execute` command will instantiate a `Canvas` and call its methods based on the recipe file's contents.
*   **`loader` & `runtime`**: The `Canvas` orchestrates these packages to load, initialize, and run the models.
*   **`viz`**: The `Canvas.Plot` method uses generators from the `viz` package to create visualizations.

**Current Status:**

*   The `Canvas` API is fully functional and production-ready, validated through comprehensive tests including capacity modeling scenarios.
*   The `Set` method correctly handles nested parameter modification for both native and user-defined components, with proper instance management to ensure parameters reach the correct simulation instances.
*   **Capacity Modeling Integration:** The Canvas now supports full capacity modeling workflows, allowing dynamic modification of ResourcePool parameters and execution of realistic load testing scenarios.
*   All core primitives—loading, in-memory modification of parameters, running simulations, and plotting from session variables—are working correctly.
*   The package is ready to be integrated into the user-facing `sdl execute` command and supports sophisticated system analysis workflows.
