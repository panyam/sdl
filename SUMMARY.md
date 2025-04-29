## SDL Project Summary & Onboarding Guide (Organized by Proposed Folders)

**1. Vision & Goal:**

*   **Purpose:** The SDL (System Design Language) library aims to model and simulate the performance characteristics (latency, availability, error rates) of distributed system components using analytical models and probabilistic composition rather than full discrete-event simulation.
*   **Use Cases:** Enable rapid "what-if" analysis of system designs, bottleneck identification, SLO evaluation, interactive performance exploration, and potentially educational tools by modelling component interactions and their probabilistic outcomes.

**2. Overall Architecture:**

*   The core simulation engine relies on composing `Outcomes` distributions.
*   Component models (primitives, storage, etc.) define their performance profiles as `Outcomes`.
*   Operations are modelled by combining `Outcomes` using operators (`And`, `If`, `Map`, `Split`, `Append`).
*   Complexity (combinatorial explosion) is managed via reduction strategies (Merge+Interpolate/Anchor).
*   State evolution is handled externally by modifying component parameters between simulation runs, not automatically within operations (known limitation for components like `ResourcePool`).
*   Future Goal (Phase 4): Implement a user-friendly DSL to define components and orchestrate simulations, translating DSL definitions into calls to this Go library.

**3. Code Structure & Package Overview:**

**(Proposed Folder Structure)**

*   **`sdl/` (Root)**
    *   `go.mod`, `go.sum`: Go module definitions.
    *   `Makefile`: Build/test/watch commands.
    *   `DSL.md`, `ROADMAP.MD`: Design notes and future plans for the DSL.
    *   `SUMMARY.md`: (This document) Project overview.
    *   *(Future)* `README.md`: High-level project description and usage examples.

*   **`sdl/core/`**: Contains the absolute fundamental, generic types and operations for probabilistic distributions and analysis, independent of specific result types or components.
    *   **Key Files & Concepts:**
        *   `outcomes.go`: Defines `Bucket[V]` and the core `Outcomes[V any]` struct. Includes fundamental methods like `Add`, `Len`, `TotalWeight`, `Copy`, `Split`, `Partition`, `ScaleWeights`, `GetValue`, and the generic `Sample(rng *rand.Rand)`. Defines generic operators `And`, `If`, `Map`.
        *   `utils.go`, (`duration.go` in summary): Defines `Duration` type (`float64`) and time unit helpers (`Millis`, `Nanos`, etc.), `approxEqualTest`.
        *   `metricable.go` (conceptual): Defines the `Metricable` interface (`IsSuccess() bool`, `GetLatency() Duration`) used by metric calculations.
        *   `converters.go`: Contains `ConvertToRanged`, `ConvertToAccess`, and `SampleWithinRange`.
        *   `reducer.go`: Defines generic `AdaptiveReduce` (currently less used) and related types. Core reduction *strategies* are often implemented alongside their specific result types.
        *   `analyzer.go`: **(Recent Work)** Implements the stateless `Analyze` primitive. Takes a simulation function and expectations, returns an `AnalysisResult` struct containing calculated metrics and expectation check outcomes. Includes helper types (`MetricType`, `OperatorType`, `Expectation`) and result reporting/assertion methods (`LogResults`, `Assert`, `AssertFailure`). This standardizes test result verification.
        *   `analyzer_test.go`: Unit tests verifying the `Analyze` function's logic and reporting.

*   **`sdl/core/results/` (Conceptual - currently files live directly in `core/`)**: Defines concrete result types (`V`) that can be used within `Outcomes[V]`, along with type-specific helper functions and reduction strategies.
    *   **Key Files & Concepts:**
        *   `simpleresult.go`: Defines `AccessResult{Success, Latency}`, its methods (`IsSuccess`, `GetLatency`), `AndAccessResults`, `MergeAdjacentAccessResults`, `ReduceAccessResultsPercentileAnchor`, and the composite `TrimToSize`.
        *   `rangedresult.go`: Defines `RangedResult{Success, Min, Mode, Max}`, its methods (`IsSuccess`, `GetLatency` via Mode), `AndRangedResults`, `Overlap()`, `MergeOverlappingRangedResults` (map-based), `InterpolateRangedResults`, and the composite `TrimToSizeRanged`.
        *   Associated `_test.go` files verify converters and reduction strategies.

*   **`sdl/core/metrics/` (Conceptual - currently files live directly in `core/`)**: Provides functions to analyze and calculate standard performance metrics from `Outcomes` distributions.
    *   **Key Files & Concepts:**
        *   `metrics.go`: Implements generic helper functions `Availability[V Metricable]`, `MeanLatency[V Metricable]`, `PercentileLatency[V Metricable]`.
        *   `metrics_test.go`: Verifies the metric calculations for different result types and edge cases.

*   **`sdl/components/`**: Defines concrete system building block components, leveraging the `core` library.
    *   **Key Files & Concepts:**
        *   `aliases.go`: Provides convenient type aliases (`Duration`, `Outcomes`) for use within the components package.
        *   **Primitives:** `disk.go` (SSD/HDD profiles), `network.go` (latency, jitter, loss), `queue.go` (M/M/c/K analytical model), `resourcepool.go` (M/M/c analytical model, state limitations).
        *   **Storage:** `cache.go` (hit/miss logic), `index.go` (base struct), `btree.go`, `hashindex.go` (heuristic probabilities), `lsm.go` (simplified model), `bitmap.go`, `heapfile.go`, `sortedfile.go`.
        *   **Orchestration:** `batcher.go` (size/time based batching).
        *   Associated `_test.go` files verify component logic and performance characteristics, **now primarily using the `core.Analyze` primitive for assertions**.

*   **`sdl/examples/`**: Contains examples demonstrating how to wire components together to model a system using the Go API.
    *   **Key Files & Concepts:**
        *   `bitly/`: Models a URL shortener with `IDGenerator`, `Cache`, `DatabaseComponent` (using `HashIndex`), and `BitlyService`.
        *   `bitly_test.go`: Demonstrates end-to-end analysis of the `BitlyService` operations (`Redirect`, `ShortenURL`) using the `core.Analyze` primitive and expectations.

*   **`sdl/dsl/` (Future - Phase 4)**
    *   **Purpose:** Will contain the parser, interpreter, Abstract Syntax Tree (AST) nodes, and runtime environment for the user-facing System Design Language.

**4. Current Status & Recent Work:**

*   Phase 1 (Core & Primitives) and Phase 2 (Component Models & Basic Concurrency concepts) are largely complete, focusing on analytical modelling. Phase 3 (Advanced Concurrency/Queuing/State) was partially addressed by adding analytical Queue/Pool models.
*   Core `Outcomes` system with composition operators and robust reduction strategies is stable and tested.
*   A good suite of primitive and storage components exist, primarily using `Outcomes[AccessResult]` and analytical approximations. Metric calculation and sampling are available.
*   **Recent Work:** The `Analyze` function was refactored into a **stateless primitive** in `sdl/core/analyzer.go`. It now returns a structured `AnalysisResult` instead of directly asserting within tests. Helper methods (`Assert`, `AssertFailure`, `LogResults`) were added to `AnalysisResult` for convenience. Component tests (`disk_test`, `btree_test`, `bitmap_test`, `batcher_test`, etc.) and the `bitly_test` example were updated to use `Analyze(...).Assert()` for standardized result verification, removing manual metric calculations and checks from the test bodies.
*   **Known Limitations:** Accurate modelling of stateful interactions (esp. `ResourcePool.Release`), complex asynchronous patterns, and time-based batching variance remains challenging without moving towards DES. The current models favour analytical speed over detailed simulation accuracy in these areas.

**5. Next Steps (Phase 4):**

*   Focus on **Usability & System Composition** via the DSL.
*   **Primary Task:** Implement the DSL parser and interpreter as outlined in `DSL.md` and `ROADMAP.MD`.
