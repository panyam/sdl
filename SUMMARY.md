## SDL Project Summary & Onboarding Guide

**1. Vision & Goal:**

*   **Purpose:** The SDL (System Design Language) library aims to model and simulate the performance characteristics (latency, availability, error rates) of distributed system components using analytical models and probabilistic composition rather than full discrete-event simulation.
*   **Use Cases:** Enable rapid "what-if" analysis of system designs, bottleneck identification, SLO evaluation, interactive performance exploration, and potentially educational tools by modelling component interactions and their probabilistic outcomes.

**2. Overall Architecture:**

*   The core simulation engine relies on composing `Outcomes` distributions.
*   Component models (primitives, storage, etc.) define their performance profiles as `Outcomes`.
*   Operations are modelled by combining `Outcomes` using operators (`And`, `If`, `Map`, `Split`, `Append`).
*   Complexity (combinatorial explosion) is managed via reduction strategies (Merge+Interpolate/Anchor).
*   **Stateless Analytical Core:** Components like `ResourcePool` now operate based on configured steady-state rates, improving analytical consistency but not modeling instantaneous state or bursts. State evolution in a system context is handled externally by modifying component parameters between simulation runs.
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
        *   `utils.go`: Defines `Duration` type (`float64`) and time unit helpers (`Millis`, `Nanos`, etc.), `approxEqualTest`.
        *   `metricable.go` (conceptual): Defines the `Metricable` interface (`IsSuccess() bool`, `GetLatency() Duration`) used by metric calculations.
        *   `converters.go`: Contains `ConvertToRanged`, `ConvertToAccess`, and `SampleWithinRange`.
        *   `reducer.go`: Defines generic `AdaptiveReduce` (currently less used) and related types. Core reduction *strategies* are often implemented alongside their specific result types.
        *   `distributions.go`: **(Recent Work)** Implements distribution generation helpers, notably `NewDistributionFromPercentiles` which creates an `Outcomes` profile from percentile data (e.g., P50, P90, P99) and failure rates. Includes `distributions_test.go`.
        *   `analyzer.go`: **(Recent Work)** Implements the **stateless `Analyze` primitive**. Takes a simulation function and expectations, returns an `AnalysisResult` struct containing calculated metrics and expectation check outcomes. Includes helper types (`MetricType`, `OperatorType`, `Expectation`) and result reporting/assertion methods (`LogResults`, `Assert`, `AssertFailure`). This standardizes test result verification. Includes `analyzer_test.go`.

*   **`sdl/core/results/` (Conceptual - files currently live directly in `core/`)**: Defines concrete result types (`V`) that can be used within `Outcomes[V]`, along with type-specific helper functions and reduction strategies.
    *   **Key Files & Concepts:**
        *   `simpleresult.go`: Defines `AccessResult{Success, Latency}`, its methods (`IsSuccess`, `GetLatency`), `AndAccessResults`, `MergeAdjacentAccessResults`, `ReduceAccessResultsPercentileAnchor`, and the composite `TrimToSize`.
        *   `rangedresult.go`: Defines `RangedResult{Success, Min, Mode, Max}`, its methods (`IsSuccess`, `GetLatency` via Mode), `AndRangedResults`, `Overlap()`, `MergeOverlappingRangedResults` (map-based), `InterpolateRangedResults`, and the composite `TrimToSizeRanged`.
        *   Associated `_test.go` files verify converters and reduction strategies.

*   **`sdl/core/metrics/` (Conceptual - files currently live directly in `core/`)**: Provides functions to analyze and calculate standard performance metrics from `Outcomes` distributions.
    *   **Key Files & Concepts:**
        *   `metrics.go`: Implements generic helper functions `Availability[V Metricable]`, `MeanLatency[V Metricable]`, `PercentileLatency[V Metricable]`.
        *   `metrics_test.go`: Verifies the metric calculations for different result types and edge cases.

*   **`sdl/components/`**: Defines concrete system building block components, leveraging the `core` library.
    *   **Key Files & Concepts:**
        *   `aliases.go`: Provides convenient type aliases (`Duration`, `Outcomes`) for use within the components package.
        *   **Primitives:** `disk.go` (SSD/HDD profiles), `network.go` (latency, jitter, loss), `queue.go` (M/M/c/K analytical model), `resourcepool.go` (**Now Stateless:** Uses M/M/c analytical model based *only* on configured rates `lambda`, `Ts`, `Size`; `Used` state and `Release` method removed).
        *   **Storage:** `cache.go` (hit/miss logic), `index.go` (base struct), `btree.go`, `hashindex.go` (heuristic probabilities), `lsm.go` (simplified model), `bitmap.go`, `heapfile.go`, `sortedfile.go`.
        *   **Orchestration:** `batcher.go` (size/time based batching using analytical wait time approx).
        *   Associated `_test.go` files verify component logic and performance characteristics, **now consistently using the `core.Analyze` primitive** (`result.Assert()`, `result.AssertFailure()`) for standardized assertions.

*   **`sdl/examples/`**: Contains examples demonstrating how to wire components together to model a system using the Go API.
    *   **Key Files & Concepts:**
        *   `bitly/`: Models a URL shortener (`IDGenerator`, `Cache`, `DatabaseComponent`, `BitlyService`). Tests use `Analyze(...).Assert()`.
        *   `gpucaller/`: **(Recent Work)** Models an App Server using a Batcher to interact with a pool of GPU resources (`AppServer`, `Batcher`, `GpuBatchProcessor`, `ResourcePool`). Demonstrates parameter sweeping in tests (`TestGpuCaller_Scenarios`) using the stateless `ResourcePool` and `Analyze`.

*   **`sdl/dsl/` (Future - Phase 4)**
    *   **Purpose:** Will contain the parser, interpreter, Abstract Syntax Tree (AST) nodes, and runtime environment for the user-facing System Design Language.

**4. Current Status & Recent Work Summary:**

*   The core Go library providing the probabilistic modelling engine (`Outcomes`, composition, reduction, metrics) and a suite of component models is stable and well-tested.
*   **Recent Work:**
    *   Refactored `core.Analyze` into a stateless primitive returning `AnalysisResult`, separating analysis execution from test assertion. Added `Assert`/`AssertFailure` helpers.
    *   Updated all component tests (`sdl/components/*_test.go`) and examples tests (`sdl/examples/*/*_test.go`) to use the new `Analyze(...).Assert()` pattern for standardized verification.
    *   Added `core.NewDistributionFromPercentiles` helper function to create `Outcomes` distributions from percentile data, with tests.
    *   Refactored `components.ResourcePool` to be fully stateless and analytical, removing the `Used` field and `Release` method, aligning it with its M/M/c steady-state model. Updated its tests.
    *   Created the `sdl/examples/gpucaller` example demonstrating `Batcher` and the stateless `ResourcePool`.
*   **Known Limitations (Explicitly Acknowledged):**
    *   **Analytical vs. DES:** The library models *steady-state averages*. It doesn't capture transient dynamics (bursts, precise queue buildup/drain) like DES. Use for capacity planning and average performance analysis.
    *   **Stateless `ResourcePool`:** Cannot model instantaneous contention based on exact usage; relies purely on configured average rates (`lambda`, `Ts`).
    *   **Profile Accuracy:** Simulation accuracy depends on realistic `Outcomes` profiles (esp. for components like the GPU defined by SLOs). Use real data or `NewDistributionFromPercentiles`.
    *   **Batcher/Queue Approximations:** Use analytical formulas for average wait times.
    *   **Cold Starts:** Steady-state models don't represent initial empty system state.
    *   **Scaling/Topology:** Models focus on single-instance paths. Extrapolate total throughput externally using parameter sweeping to find per-instance capacity.
    *   **Implicit Network:** Add `NetworkLink` explicitly for higher fidelity.

**5. Next Steps (Phase 4):**

*   Focus on **Usability & System Composition** via the DSL.
*   **Primary Task:** Implement the DSL parser and interpreter as outlined in `DSL.md` and `ROADMAP.MD`, translating DSL constructs into calls to this refined Go library.

