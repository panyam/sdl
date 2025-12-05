# SDL Core Package Summary (`sdl/core`)

**Purpose:**

This package forms the fundamental layer of the SDL library. It provides the generic building blocks for representing probabilistic performance outcomes, composing them, calculating standard metrics, and managing complexity through distribution reduction. It is designed to be independent of specific system components and the DSL execution layer.

**Key Concepts & Components:**

1.  **`Outcomes[V any]` (`outcomes.go`):**
    *   The central data structure representing a discrete probability distribution.
    *   A slice of `Bucket[V]` structs, where each bucket has a `Weight` (probability) and a `Value` of generic type `V`.
    *   Provides core methods: `Add`, `Len`, `TotalWeight`, `Copy`, `Split` (conditional branching based on value predicate), `Partition`, `ScaleWeights`, `GetValue` (for deterministic results), `Sample` (random sampling).

2.  **Composition Operators (`outcomes.go`):**
    *   Generic functions like `And` (sequential composition), `Map` (transform outcome values), `Append` combine `Outcomes` objects.
    *   `And` requires a type-specific `ReducerFunc` (e.g., `AndAccessResults`, `AddDurations`) provided externally to define how inner values combine.

3.  **Result Types (`simpleresult.go`, `rangedresult.go`):**
    *   Defines concrete types for `V` (the value within an `Outcome` bucket) used commonly in performance modeling.
    *   `AccessResult`: `{Success bool, Latency Duration}` - Represents a simple success/failure outcome with a point latency. This is a primary type used by `components`.
    *   `RangedResult`: `{Success bool, MinLatency, ModeLatency, MaxLatency Duration}` - Represents outcomes with latency uncertainty.
    *   `Duration`: Type alias for `float64` representing time in seconds (`utils.go`).

4.  **`Metricable` Interface (`metrics.go`):**
    *   Interface (`IsSuccess() bool`, `GetLatency() Duration`) required by outcome value types (`V`) to be used with generic metric calculations.
    *   Implemented by `AccessResult` and `RangedResult`.
    *   Note: The `decl.Value` type (used by the runtime) also has a `Time` field representing latency, and can be mapped to/from `Metricable` types like `AccessResult`.

5.  **Metrics Calculation (`metrics.go`):**
    *   Provides generic functions to calculate standard performance indicators from `Outcomes[V]` where `V` is `Metricable`.
    *   Includes `Availability`, `MeanLatency` (of successes), `PercentileLatency` (e.g., P50, P99 of successes).

6.  **Reduction / Trimming (`reducer.go`, `simpleresult.go`, `rangedresult.go`):**
    *   Essential mechanisms to manage the combinatorial explosion of buckets during composition.
    *   Strategies include merging (`MergeAdjacentAccessResults`, `MergeOverlappingRangedResults`), interpolation (`InterpolateRangedResults`, `ReduceAccessResultsPercentileAnchor`), and significance-based sampling (`AdaptiveReduce`).
    *   `TrimToSize` / `TrimToSizeRanged`: Higher-level functions orchestrating reduction strategies.

7.  **Converters (`converters.go`):**
    *   Functions to convert between `Outcomes[AccessResult]` and `Outcomes[RangedResult]`.
    *   Includes `SampleWithinRange` for generating point latencies from `RangedResult`.

8.  **Distribution Helpers (`distributions.go`):**
    *   Utilities to create common distributions, e.g., `NewDistributionFromPercentiles` for `Outcomes[AccessResult]`.

9.  **Analysis Primitive (`analyzer.go`):**
    *   (Note: The `analyzer.go` feature within `core` using `Expectation`s is distinct from the DSL `analyze` blocks, which are yet to be fully implemented in the runtime).
    *   Provides a stateless `Analyze` function for the Go API, taking a simulation function `func() *Outcomes[V]`, calculating metrics, checking against `Expectation`s, and returning an `AnalysisResult[V]`.
    *   Heavily used in tests for `components` to verify their Go API behavior.

10. **Utilities (`utils.go`):**
    *   `Duration` type alias, time unit helpers (`Millis`, `Micros`), comparison helpers.

**Current Status:**

*   The core library is mature and provides a robust foundation for probabilistic performance modeling using the Go API.
*   Key composition, reduction, metric, and analysis functions are implemented and well-tested.
*   The design supports extension with new outcome types.

**Relationship with other packages:**

*   `components` package: Concrete components (Disk, Cache, etc.) use `core` types like `Outcomes[AccessResult]` to model their performance via their Go API methods.
*   `decl` package: Defines `Value`, which is the runtime representation of data in the DSL. `decl.Value` has a `Time` field for latency. Converters or mappers are needed to bridge between `core.AccessResult` (from component Go APIs) and `decl.Value` (for the DSL runtime).
*   `runtime` package: The DSL runtime (e.g., `SimpleEval`) operates on `decl.Value`. When it calls native components (which return `core.Outcomes`), these outcomes (or samples from them) are translated into `decl.Value` instances, transferring the latency to `Value.Time`.
