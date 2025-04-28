## SDL Project Summary & Onboarding Guide

**1. Vision & Goal:**

*   **Purpose:** The SDL (System Design Language) library aims to model and simulate the performance characteristics (latency, availability, error rates) of distributed system components.
*   **Use Cases:**
    *   Allow developers/designers to define system architectures (databases, caches, APIs, etc.).
    *   Simulate the probabilistic outcomes of operations (e.g., database reads, API calls) without executing real code.
    *   Analyze end-to-end system performance, SLOs (Service Level Objectives), and bottlenecks based on component interactions.
    *   Enable interactive "what-if" analysis by changing component parameters (e.g., disk type, network latency, cache size) and observing the impact on overall system behavior (visualized via plots).
    *   Potential for educational tools or gamified system design challenges ("Design a system that meets X Availability and Y P99 Latency").

**2. Core Concepts:**

*   **`Outcomes[V any]`:** The central data structure representing a probability distribution of possible results for an operation.
    *   It holds a slice of `Bucket[V]`, where each bucket has a `Weight` (probability/frequency) and a `Value V` (the specific outcome).
    *   Uses Go generics (`V any`) for flexibility, allowing distributions of different outcome types.
*   **Composition:** Simulating system interactions by combining `Outcomes`:
    *   **`And(o1, o2, reducer)`:** Models sequential operations. Calculates the Cartesian product of outcomes from `o1` and `o2`, combining weights and values using a `reducer` function. *Crucially, this leads to combinatorial explosion of buckets.*
    *   **`If(cond, then, otherwise, reducer)`:** Models conditional logic based on the outcome of a preceding step.
    *   **`Map(o, mapper)`:** Transforms outcome values within a distribution.
*   **Outcome Types:** Concrete types representing results:
    *   **`AccessResult`:** Simple `{ Success bool; Latency Duration }`.
    *   **`RangedResult`:** More detailed `{ Success bool; MinLatency, ModeLatency, MaxLatency Duration }`.
    *   Other types (like `Duration`, `int`, `string`) can exist during intermediate steps.
*   **Metric Calculation:**
    *   `Metricable` interface: Defines `IsSuccess()` and `GetLatency()` methods.
    *   Helper functions (`Availability`, `MeanLatency`, `PercentileLatency`) operate on `Outcomes[V]` where `V` implements `Metricable`, providing compile-time safety. These are located in `metrics.go`.
*   **Reduction (Complexity Management):** Essential for managing the combinatorial explosion from `And`. Various strategies exist to reduce the number of buckets while trying to preserve key distributional characteristics.

**3. Key Components (Current):**

*   **Primitives:**
    *   `Disk`: Models disk I/O with basic configurable `ReadOutcomes` and `WriteOutcomes` (currently `AccessResult` based).
*   **Base Types:**
    *   `Index`: Embeddable struct holding common properties for on-disk data structures (`NumRecords`, `RecordSize`, `PageSize`, `Disk` dependency, `RecordProcessingTime`, `MaxOutcomeLen` for trimming).
*   **Storage Models:** Implementations of data structures built using `Index` and `Outcomes` composition:
    *   `HeapFile`
    *   `BTreeIndex`
    *   `HashIndex`
    *   `SortedFile`
    *   *(These models are currently simplified approximations)*.

**4. Current Status & Recent Work (Reduction Refinement - End of Phase 1, Step 2):**

*   The core `Outcomes` distribution and composition logic (`And`, `If`, `Map`) is implemented.
*   Basic components (`Disk`, `Index`, several storage models) exist.
*   **Extensive work was just completed on refining bucket reduction strategies:**
    *   **`AdaptiveReduce`:** Found to be either too slow (RangedResult) or inaccurate (AccessResult) with tested significance functions. It's currently **not recommended** for use.
    *   **`AccessResult` Reduction Strategy:**
        *   Use `MergeAdjacentAccessResults` (with a small relative latency threshold, e.g., 5%) for initial, accurate reduction.
        *   Follow with `InterpolateAccessResults` to deterministically reduce to a specific `maxLen` while preserving distribution shape well.
        *   This logic is encapsulated in `TrimToSize`.
    *   **`RangedResult` Reduction Strategy:**
        *   Use `MergeOverlappingRangedResults` (map-based version, with a high overlap threshold, e.g., 90%) for initial, fast reduction.
        *   Follow with `InterpolateRangedResults` to reduce to `maxLen`.
        *   This logic is encapsulated in `TrimToSizeRanged`.
*   Metric calculation helpers (`Availability`, `MeanLatency`, `PercentileLatency`) are available.
*   Benchmarks and accuracy tests exist for reduction strategies.

**5. Known Issues & Limitations:**

*   **RangedResult Merge Memory:** `MergeOverlappingRangedResults` (map-based) still uses significant memory (~20MB/78k allocs for ~78k input buckets in benchmarks).
*   **Model Fidelity:** Current component models (Disk, Indexes) are simplified and lack features like caching, detailed B-Tree costs, LSM specifics, concurrency effects etc.
*   **Concurrency/Asynchrony:** No built-in support for modelling parallel execution, queues, asynchronous callbacks, batching, or resource contention (Phase 2/3 goal).
*   **DSL Disconnect:** The high-level declarative System Design Language (envisioned in `sdl.go` and markdown) is not yet implemented or connected to the simulation engine. Modelling requires writing Go code using `And`/`If`.
*   **State Evolution:** The library primarily models single operations based on initial state; modelling how system state *changes* over time (affecting future performance) is not yet supported.

**6. Code Structure (Current):**

*   All Go files currently reside in a single top-level `sdl` folder.

**7. Roadmap Recap:**

*   **Phase 1 (Partially Complete):** Solidify Core & Enhance Primitives.
    *   ✅ Basic Validation Hooks (Metrics).
    *   ✅ Benchmark & Refine Reduction Strategies.
    *   *Next:* Enhance `Disk` (SSD/HDD).
    *   *Next:* Add `NetworkLink`.
    *   *Next:* (Optional) Improve Outcome Type (Retries).
*   **Phase 2:** Expand Component Models & Basic Concurrency (More indexes, Caching, `ParallelAnd`, Resource Pools).
*   **Phase 3:** Advanced Concurrency, Queuing & State (Queues, Async, Batching, State Evolution).
*   **Phase 4:** Usability & System Composition (DSL Implementation, Visualization, Component Library, Examples).
*   **Phase 5:** Refinement & Ecosystem (Optimization, Validation, Docs, Community).

---

**Proposed Code Organization:**

To improve maintainability and avoid future circular dependencies as the project grows, consider restructuring into sub-packages:

```
sdl/
├── core/                 # Fundamental types & logic (NOT specific components)
│   ├── outcomes.go         # Outcomes[V], Bucket[V] definitions
│   ├── duration.go         # Duration type, helpers (Millis, etc.)
│   ├── metricable.go       # Metricable interface definition
│   └── reducers.go         # Generic reduction logic (e.g., AdaptiveReduce base, Interpolation base?)
│
├── results/              # Concrete outcome types + their specific reducers/helpers
│   ├── access_result.go    # AccessResult definition, GetLatency, AndAccessResults
│   ├── ranged_result.go    # RangedResult definition, GetLatency, AndRangedResults, Overlap, DistTo
│   ├── access_result_reducers.go # MergeAdjacent, InterpolateAccess, TrimToSize
│   └── ranged_result_reducers.go # MergeOverlapping, InterpolateRanged, TrimToSizeRanged, RangedSignificance (if kept)
│
├── metrics/              # Metric calculation logic
│   └── metrics.go          # Availability, MeanLatency, PercentileLatency functions
│
├── primitives/           # Basic, indivisible components
│   ├── disk.go             # Disk struct, Init, Read, Write profiles (SSD/HDD)
│   └── network.go          # (Future) NetworkLink struct, latency, loss models
│
├── storage/              # More complex storage components/data structures
│   ├── index.go            # Base Index struct definition
│   ├── heapfile.go         # HeapFile implementation
│   ├── btree.go            # BTreeIndex implementation
│   ├── hashindex.go        # HashIndex implementation
│   ├── sortedfile.go       # SortedFile implementation
│   └── lsm.go              # (Future) LSM Tree implementation
│
├── dsl/                  # (Future) DSL parser, interpreter, AST nodes
│   └── ...
│
├── examples/             # (Future) Usage examples showing composition
│   └── ...
│
└── go.mod                # Go module definition
└── go.sum
└── (Other top-level files like README.md, LICENSE)
```

**Benefits:**

*   Clear separation of concerns.
*   Dependencies flow logically (e.g., `storage` depends on `primitives`, `results`, `core`; `results` depends on `core`). Reduces likelihood of circular imports.
*   Easier navigation and contribution.

**Action:** This refactoring can be done incrementally, perhaps starting by moving `core` and `results` types first.

---

**Prompt for Next LLM Task:**

```text
**Project Context:**

You are continuing development on the SDL (System Design Language) Go library. This library simulates the performance (latency, availability) of system components using probabilistic `Outcomes` distributions. Key concepts include:
- `Outcomes[V any]`: Represents a probability distribution (slice of `Bucket[V]`).
- `AccessResult`, `RangedResult`: Concrete outcome types.
- Composition: `And` (sequential), `If` (conditional) combine `Outcomes`.
- Reduction: Strategies are needed to manage the combinatorial explosion from `And`. Recent work established the following preferred reduction strategies:
    - For `Outcomes[AccessResult]`: Use `TrimToSize` (located in `sdl/results/access_result_reducers.go`) which first calls `MergeAdjacentAccessResults` (low relative threshold) then `InterpolateAccessResults` to reach a target bucket count (`maxLen`).
    - For `Outcomes[RangedResult]`: Use `TrimToSizeRanged` (located in `sdl/results/ranged_result_reducers.go`) which first calls `MergeOverlappingRangedResults` (map-based, high overlap threshold) then `InterpolateRangedResults` to reach `maxLen`.
    - `AdaptiveReduce` is generally avoided due to performance/accuracy issues found previously.
- Components: `Disk`, `Index`, `HeapFile`, `BTreeIndex`, etc. model system parts.
- Metrics: Helper functions (`Availability`, `MeanLatency`, `PercentileLatency` in `sdl/metrics/metrics.go`) analyze results.
- Code Structure: Currently flat (`sdl/`), but a proposed structure (core/, results/, primitives/, storage/, metrics/) exists (see summary above for details).

**Current Status:**

Phase 1, Step 2 (Reduction Refinement) is complete. The chosen reduction strategies are implemented and tested for accuracy.

**Next Task (Phase 1, Step 3): Enhance Disk Primitive**

Your task is to enhance the existing `Disk` primitive component (currently in `sdl/disk.go`) to model different disk types, specifically SSDs and HDDs, with distinct performance profiles.

**Requirements:**

1.  **Modify `Disk` Struct:** Add a field to the `Disk` struct (e.g., `Type string` or `Profile DiskProfile`) to indicate whether it represents an "SSD" or an "HDD".
2.  **Update `Disk.Init()`:** Modify the `Init` method (or create new constructors like `NewSSD()`, `NewHDD()`) to accept the disk type and initialize `ReadOutcomes` and `WriteOutcomes` with appropriate, distinct latency/error distributions for SSDs vs HDDs.
    *   **SSD Profile:** Should generally have lower latency (e.g., mostly sub-millisecond reads/writes), tighter distribution (less variance), potentially lower failure rates, but maybe different failure modes.
    *   **HDD Profile:** Should have higher average latency (e.g., several milliseconds), wider distribution (more variability, maybe a distinct "slow tail"), potentially different read vs. write speeds, and distinct failure characteristics (maybe higher chance of slow operations vs. outright failure compared to SSD). Use `AccessResult` for the outcomes for now. Define reasonable, distinct probability buckets for each profile.
3.  **Unit Tests:** Add new tests (e.g., in `sdl/primitives/disk_test.go` if reorganizing, or `sdl/disk_test.go` if not) specifically for the `Disk` component. These tests should:
    *   Create instances of SSD and HDD disks.
    *   Verify that their initialized `ReadOutcomes` and `WriteOutcomes` differ significantly and plausibly reflect SSD vs HDD characteristics (e.g., check mean latency, P99 latency, availability using the metric helpers).
4.  **Code Style:** Follow existing Go conventions and patterns used in the codebase (e.g., `Init` method returning receiver).
5.  **Code Location:** Place the modified `Disk` code and new tests in the appropriate location. If following the proposed reorganization, use `sdl/primitives/disk.go` and `sdl/primitives/disk_test.go`. If keeping the flat structure, update `sdl/disk.go` and `sdl/disk_test.go`. *(Assume flat structure for now unless instructed otherwise)*.

**Focus:** Implement only the Disk enhancement described above. Do not refactor the entire codebase structure in this step unless specifically asked.
```
