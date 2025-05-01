# SDL Core Package Summary (`sdl/core`)

**Purpose:**

This package forms the fundamental layer of the SDL library. It provides the generic building blocks for representing probabilistic performance outcomes, composing them, calculating standard metrics, and managing complexity through distribution reduction. It is designed to be independent of specific system components and the DSL execution layer.

**Key Concepts & Components:**

1.  **`Outcomes[V any]` (`outcomes.go`):**
    *   The central data structure representing a discrete probability distribution.
    *   A slice of `Bucket[V]` structs, where each bucket has a `Weight` (probability) and a `Value` of generic type `V`.
    *   Provides core methods: `Add`, `Len`, `TotalWeight`, `Copy`, `Split` (conditional branching based on value predicate), `Partition`, `ScaleWeights`, `GetValue` (for deterministic results), `Sample` (random sampling).
    *   *(Note: The `And` field on this struct is planned for removal, composition handled externally)*.

2.  **Composition Operators (`outcomes.go`):**
    *   Generic functions like `And` (sequential composition), `Map` (transform outcome values), `Append` combine `Outcomes` objects.
    *   `And` requires a type-specific `ReducerFunc` (e.g., `AndAccessResults`, `AddDurations`) provided externally (often via a registry in the VM) to define how inner values combine.

3.  **Result Types (`simpleresult.go`, `rangedresult.go`):**
    *   Defines concrete types for `V` used commonly in performance modeling.
    *   `AccessResult`: `{Success bool, Latency Duration}` - Represents a simple success/failure outcome with a point latency.
    *   `RangedResult`: `{Success bool, MinLatency, ModeLatency, MaxLatency Duration}` - Represents outcomes with latency uncertainty or jitter.
    *   `Duration`: Type alias for `float64` representing time in seconds (`utils.go`).

4.  **`Metricable` Interface (`metrics.go`):**
    *   Interface (`IsSuccess() bool`, `GetLatency() Duration`) required by outcome types (`V`) to be used with generic metric calculations.
    *   Implemented by `AccessResult` and `RangedResult`.

5.  **Metrics Calculation (`metrics.go`, `metrics_test.go`):**
    *   Provides generic functions to calculate standard performance indicators from `Outcomes[V]` where `V` is `Metricable`.
    *   Includes `Availability`, `MeanLatency` (of successes), `PercentileLatency` (e.g., P50, P99, P999 of successes).

6.  **Reduction / Trimming (`reducer.go`, `simpleresult.go`, `rangedresult.go`, `reduction_benchmark_test.go`):**
    *   Essential mechanisms to manage the combinatorial explosion of buckets during composition.
    *   Strategies include:
        *   Merging: `MergeAdjacentAccessResults`, `MergeOverlappingRangedResults`.
        *   Interpolation: `InterpolateRangedResults`, `ReduceAccessResultsPercentileAnchor`.
        *   Significance/Sampling: `AdaptiveReduce` (generic base).
    *   `TrimToSize` / `TrimToSizeRanged`: Higher-level functions orchestrating reduction strategies (often split -> merge -> interpolate/anchor) to meet a target bucket count (`maxLen`). Benchmarks exist to evaluate performance and accuracy.

7.  **Converters (`converters.go`, `converters_test.go`):**
    *   Functions to convert between `Outcomes[AccessResult]` and `Outcomes[RangedResult]`.
    *   Includes `SampleWithinRange` for generating point latencies from `RangedResult`.

8.  **Distribution Helpers (`distributions.go`, `distributions_test.go`):**
    *   Utilities to create common distributions.
    *   `NewDistributionFromPercentiles`: Creates `Outcomes[AccessResult]` from latency percentiles (P0, P50, P99, etc.), failure rate, and failure latency profile.

9.  **Analysis Primitive (`analyzer.go`, `analyzer_test.go`):**
    *   **Stateless `Analyze` function:** Takes a simulation function `func() *Outcomes[V]`, calculates standard metrics, checks against `Expectation`s (e.g., `ExpectP99(LT, 0.100)`), and returns a `AnalysisResult[V]` struct.
    *   `AnalysisResult[V]`: Holds raw outcomes, calculated metrics, expectation check results, and helper methods (`Assert`, `AssertFailure`, `LogResults`).
    *   Provides a standardized way to execute simulations defined via the **Go API** and assert performance characteristics, used heavily in tests across the project.

10. **Utilities (`utils.go`):**
    *   `Duration` type alias (`float64`).
    *   Time unit helpers (`Millis`, `Micros`, `Nanos`).
    *   `approxEqualTest`, `MaxDuration`, `MinDuration` helpers.

**Current Status:**

*   The core library is mature and provides a robust foundation for probabilistic performance modeling using the Go API.
*   Key composition, reduction, metric, and analysis functions are implemented and tested.
*   Generic design allows extension with new outcome types.
*   Ready to support the execution logic required by the DSL VM (Model V4).
