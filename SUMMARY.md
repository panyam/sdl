## SDL Project Summary & Onboarding Guide

**1. Vision & Goal:**

*   **Purpose:** The SDL (System Design Language) library models and analyzes the performance characteristics (latency, availability) of distributed system components using **analytical models** and **probabilistic composition**, prioritizing speed for interactive "what-if" analysis over detailed Discrete Event Simulation (DES).
*   **Use Cases:** Enable rapid analysis of system designs, bottleneck identification, SLO evaluation, performance exploration, and potentially educational tools.

**2. Overall Architecture:**

*   **Probabilistic Core:** Relies on composing `Outcomes[V]` distributions defined in `sdl/core`.
*   **Analytical Components:** Models (`sdl/components`) define performance profiles as `Outcomes` based on configuration and analytical formulas (e.g., M/M/c for queues/pools).
*   **Stateless Core:** Key components like `ResourcePool` are now stateless, relying purely on configuration parameters (`lambda`, `Ts`, `Size`) for steady-state predictions, enhancing analytical consistency. State evolution is handled externally between analysis runs.
*   **Composition:** Operations are modelled by composing `Outcomes` using operators (`And`, `Map`, `Split`, `Append`) defined in `sdl/core`.
*   **Complexity Management:** Combinatorial explosion is managed via reduction strategies (`TrimToSize`, `TrimToSizeRanged`) combining merging and interpolation/anchoring, implemented alongside result types (`simpleresult.go`, `rangedresult.go`).
*   **Declarative Layer (`sdl/decl`):** A new layer defines component configurations and methods (`*AST()`) that generate Abstract Syntax Trees (ASTs) representing the intended computation, separating definition from execution.
*   **Future Execution:** A DSL Interpreter/VM (`sdl/dsl`) will execute these ASTs, orchestrating calls to the `sdl/core` primitives, managing type-specific logic, and applying intermediate reduction.

**3. Code Structure & Package Overview:**

*   **`sdl/` (Root)**
    *   `go.mod`, `go.sum`: Go module definitions.
    *   `Makefile`: Build/test/watch commands.
    *   `DSL.md`, `ROADMAP.MD`: Design notes and future plans for the DSL and VM/Interpreter.
    *   `SUMMARY.md`: (This document) Project overview.
    *   *(Future)* `README.md`: High-level project description and usage examples.

*   **`sdl/core/`**: Fundamental, generic types and operations for probabilistic distributions and analysis.
    *   `outcomes.go`: `Bucket[V]`, `Outcomes[V any]`, core methods (`Add`, `Len`, `TotalWeight`, `Copy`, `Split`, `Partition`, `ScaleWeights`, `GetValue`, `Sample`). Generic operators `And`, `Map`. (Missing `Filter`, `Repeat`, `FanoutAnd`).
    *   `utils.go`: `Duration`, time unit helpers, `approxEqualTest`.
    *   `metricable.go` (conceptual): `Metricable` interface.
    *   `converters.go`: `ConvertToRanged`, `ConvertToAccess`, `SampleWithinRange`.
    *   `reducer.go`: Generic `AdaptiveReduce` base. Specific strategies are in result type files.
    *   `distributions.go`: **(New)** `NewDistributionFromPercentiles` helper. Includes `distributions_test.go`.
    *   `analyzer.go`: **(New/Refactored)** **Stateless `Analyze` primitive**, `AnalysisResult`, `Expectation` types, `Assert`/`AssertFailure`/`LogResults` helpers. Standardizes testing. Includes `analyzer_test.go`.
    *   `simpleresult.go`: `AccessResult`, `AndAccessResults`, `MergeAdjacent...`, `Reduce...PercentileAnchor`, `TrimToSize`. Includes `_test.go`.
    *   `rangedresult.go`: `RangedResult`, `AndRangedResults`, `MergeOverlapping...`, `Interpolate...`, `TrimToSizeRanged`. Includes `_test.go`.
    *   `metrics.go`: Generic `Availability`, `MeanLatency`, `PercentileLatency`. Includes `metrics_test.go`.

*   **`sdl/components/`**: Concrete system building block components using `sdl/core`.
    *   `aliases.go`: Type aliases (`Duration`, `Outcomes`).
    *   **Primitives:** `disk.go`, `network.go`, `queue.go` (M/M/c/K), `resourcepool.go` (**Stateless**, M/M/c).
    *   **Storage:** `cache.go`, `index.go`, `btree.go`, `hashindex.go`, `lsm.go`, `bitmap.go`, `heapfile.go`, `sortedfile.go`.
    *   **Orchestration:** `batcher.go`.
    *   `*_test.go`: Component tests consistently use `core.Analyze(...).Assert()` for verification.

*   **`sdl/components/decl/` (New Package):** Defines declarative component representations generating ASTs.
    *   Defines declarative structs (e.g., `decl.Disk`, `decl.BTreeIndex`) holding configuration (`ast.Expr` where needed) and methods (`*AST()`) that construct and return `ast.Expr` trees representing operations. Covers all components previously in `sdl/components` with similar file structures.
    *   `components_test.go`: Tests verify that the `*AST()` methods generate the correct Expr nodes.

*   **`sdl/examples/`**: Examples using the Go API (or potentially `sdl/decl` + VM later).
    *   `bitly/`: URL shortener model (`IDGenerator`, `Cache`, `DatabaseComponent`, `BitlyService`). Tests use `Analyze`.
    *   `gpucaller/`: Models GPU processing pool (`AppServer`, `Batcher`, `GpuBatchProcessor`, stateless `ResourcePool`). Tests demonstrate parameter sweeping with `Analyze`.

*   **`sdl/dsl/` (Future - Phase 4)**
    *   **Purpose:** Will contain the **DSL Interpreter/VM**, parser (if needed separately from AST generation), environment management, and potentially the user-facing DSL parser entry point.
    *   `ast/ast.go`: Defines AST node types (`LiteralExpr`, `CallExpr`, `AndExpr`, `ParallelExpr`, `InternalCallExpr`, `RepeatExpr`, `FanoutExpr`, etc.) with `String()` methods.

**4. Current Status & Recent Work Summary:**

*   The core analytical library (`sdl/core`, `sdl/components`) is functionally complete for modelling steady-state performance via probabilistic composition. Key components are stateless and rate-driven. Standardized testing via `core.Analyze` is implemented.
*   **Recent Work:**
    *   Refactored `Analyze` to be stateless.
    *   Added `NewDistributionFromPercentiles`.
    *   Refactored `ResourcePool` to be stateless.
    *   Updated all tests to use `Analyze(...).Assert()`.
    *   Created the `gpucaller` example.
    *   Created the **`sdl/decl` package** with AST nodes and declarative component versions (`*AST()` methods) generating these ASTs. Added tests for this layer.
*   **Known Limitations (Explicitly Acknowledged):**
    *   **Analytical Steady-State:** Models average performance, not transient bursts/dynamics.
    *   **Stateless Components:** `ResourcePool` relies purely on configured rates.
    *   **Profile Accuracy:** Depends on realistic input `Outcomes` distributions.
    *   **Approximations:** Batcher wait time, parallel execution (`Repeat`/`FanoutAnd` logic needed).
    *   **Cold Starts/Scaling/Network:** Not inherently modeled, require external handling or explicit components.
    *   **Missing Core Primitives:** `Filter`, `Repeat`, `FanoutAnd` needed for advanced patterns.

**5. Next Steps (Phase 4):**

*   Focus on **Usability & System Composition** via the DSL and its execution engine.
*   **Primary Task:** Implement the **DSL Interpreter/VM** (`sdl/dsl/interpreter.go` or `vm.go`) that executes the ASTs generated by `sdl/decl`.
