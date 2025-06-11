# SDL Visualization Package Summary (`viz` package)

**Purpose:**

This package consolidates all visualization-related logic for the SDL project. It provides a clean, decoupled library for generating various diagrams and plots from different data sources, such as parsed SDL files, trace data, or simulation results. Its primary goal is to separate the concerns of data generation (handled by the `runtime`) and user interaction (handled by `cmd/sdl`) from the specifics of visual presentation.

**Core Structure & Files:**

*   **`interfaces.go`**: Defines the fundamental contracts and common data structures for all generators in the package.
    *   **Interfaces**:
        *   `StaticDiagramGenerator`: For creating static architecture diagrams.
        *   `SequenceDiagramGenerator`: For creating dynamic sequence diagrams.
        *   `Plotter`: For creating time-series plots and charts.
    *   **Data Structures**:
        *   `Node`, `Edge`: For representing static diagrams.
        *   `DataPoint`, `DataSeries`: For representing data for plots.

*   This folder contains a suite of concrete implementations of the `StaticDiagramGenerator` and other interfaces for various formats.
    *   (`dot.go` - `DotGenerator`): Creates Graphviz DOT files.
    *   (`mermaid.go` - `MermaidStaticGenerator`): Creates Mermaid graph diagrams.
    *   (`svgdrawing.go` - `SvgGenerator`): Creates standalone SVG diagrams.
    *   (`excalidraw.go` - `ExcalidrawGenerator`): Creates JSON files for the Excalidraw whiteboarding tool.

*   **`sequence.go`**: Implements the logic for generating dynamic sequence diagrams.
    *   `MermaidSequenceGenerator`: Implements the `SequenceDiagramGenerator` interface to create Mermaid sequence diagrams from `runtime.TraceData`. It correctly visualizes participants, call flows, and concurrency primitives like `loop` (`gobatch`) and `note` (`wait`).

*   **`plot.go`**: Implements the logic for generating plots.
    *   `SVGPlotter`: Implements the `Plotter` interface to create rich, multi-series SVG charts with legends, axes, and grids.

**Key Functionality:**

*   **Static Diagrams:** Can generate system architecture diagrams from a list of instances (`Node`s) and their connections (`Edge`s).
*   **Dynamic Diagrams:** Can generate sequence diagrams from a detailed execution trace (`runtime.TraceData`), providing insight into the dynamic behavior of a system.
*   **Time-Series Plots:** Can generate sophisticated SVG line charts to visualize performance metrics over time, such as latency percentiles or result counts.

**Relationship with Other Packages:**

*   **`cmd/sdl`**: The primary consumer of this package. The `diagram` and `plot` commands are thin clients that orchestrate calls to the generators in this package.
*   **`runtime`**: The `SequenceDiagramGenerator` consumes `runtime.TraceData` produced by the tracer in the runtime package.

**Current Status:**

*   The package is fully functional and contains a robust set of generators for the project's visualization needs.
*   The clear interface-based design allows for easy addition of new output formats in the future (e.g., a `PlantUMLGenerator` or an interactive D3.js plotter).
