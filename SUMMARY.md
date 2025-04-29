
## SDL Project Summary & Onboarding Guide (Organized by Proposed Folders)

**1. Vision & Goal:**

*   **Purpose:** The SDL (System Design Language) library aims to model and simulate the performance characteristics (latency, availability, error rates) of distributed system components using analytical models and probabilistic composition rather than full discrete-event simulation.
*   **Use Cases:** Enable rapid "what-if" analysis of system designs, bottleneck identification, SLO evaluation, interactive performance exploration, and potentially educational tools by modelling component interactions and their probabilistic outcomes.

**2. Overall Architecture:**

*   The core simulation engine relies on composing `Outcomes` distributions.
*   Component models (primitives, storage, etc.) define their performance profiles as `Outcomes`.
*   Operations are modelled by combining `Outcomes` using operators (`And`, `If`, `Map`).
*   Complexity (combinatorial explosion) is managed via reduction strategies (Merge+Interpolate).
*   State evolution is handled externally by modifying component parameters between simulation runs, not automatically within operations.
*   Future Goal (Phase 4): Implement a user-friendly DSL to define components and orchestrate simulations, translating DSL definitions into calls to this Go library.

**3. Code Structure & Package Overview:**

**(Proposed Folder Structure)**

*   **`sdl/` (Root)**
    *   `go.mod`, `go.sum`: Go module definitions.
    *   `README.md`: Project overview, usage examples. (To be created/updated)

*   **`sdl/core/`**
    *   **Core Files:** Contains the absolute fundamental, generic types and operations for probabilistic distributions, independent of specific result types or components.
    *   **Key Files:**
        *   `outcomes.go`: Defines `Bucket[V]` and the core `Outcomes[V any]` struct. Includes fundamental methods like `Add`, `Len`, `TotalWeight`, `Copy`, `Split`, `Partition`, `ScaleWeights`, `GetValue`, and the generic `Sample(rng *rand.Rand)`. Defines generic operators `And`, `If`, `Map`.
        *   `duration.go`: Defines `Duration` type (`float64`) and time unit helpers (`Millis`, `Nanos`, etc.).
        *   `metricable.go`: Defines the `Metricable` interface (`IsSuccess() bool`, `GetLatency() Duration`) used by metric calculations.
        *   `reducers.go`: *(Potential)* Could hold the generic base logic for reduction strategies like `AdaptiveReduce` if it were reused (currently not recommended). Interpolation logic might live here or in result-specific packages.
    *   **Results:** Defines concrete result types (`V`) that can be used within `Outcomes[V]`, along with type-specific helper functions and reduction strategies.
    *   **Key Files:**
        *   `access_result.go`: Defines `AccessResult{Success, Latency}`, its `GetLatency()` (for `Metricable`), and the specific `AndAccessResults` reducer.
        *   `ranged_result.go`: Defines `RangedResult{Success, Min, Mode, Max}`, its `GetLatency()` (using Mode), `AndRangedResults`, `Overlap()`, `DistTo()`.
        *   `access_result_reducers.go`: Implements `MergeAdjacentAccessResults`, `InterpolateAccessResults`, and the composite `TrimToSize`.
        *   `ranged_result_reducers.go`: Implements `MergeOverlappingRangedResults` (map-based), `InterpolateRangedResults`, the composite `TrimToSizeRanged`, and potentially `RangedResultSignificance` (though adaptive is currently unused).
        *   `converters.go`: Contains `ConvertToRanged`, `ConvertToAccess`, and `SampleWithinRange`.
    *   **Metrics:** Provides functions to analyze and calculate standard performance metrics from `Outcomes` distributions.
    *   **Key Files:**
        *   `metrics.go`: Implements generic helper functions `Availability[V Metricable]`, `MeanLatency[V Metricable]`, `PercentileLatency[V Metricable]`.

*   **`sdl/components/`**
    *   **Primitives:** Defines basic, often indivisible building block components.
    *   **Key Files:**
        *   `disk.go`: Defines `Disk` struct, `ProfileSSD`, `ProfileHDD`, `Init()`, `Read()`, `Write()`, `ReadProcessWrite()`. Returns `Outcomes[AccessResult]`.
        *   `network.go`: Defines `NetworkLink` struct, `Init()`, `Transfer()`. Models latency, jitter (via multiple buckets), and loss. Returns `Outcomes[AccessResult]`.
        *   `queue.go`: Defines `Queue` struct (M/M/c/K analytical model), `Init()`, `Enqueue()` (returns `AccessResult` with Success=Blocked?), `Dequeue()` (returns `Outcomes[Duration]` representing wait time).
        *   `resourcepool.go`: Defines `ResourcePool` struct (M/M/c analytical model), `Init()`, `Acquire()` (takes `lambda`, returns `AccessResult` including wait time), `Release()` (direct state change - known limitation).
    *   **Storage:** Defines more complex components, often representing data storage and indexing structures, built using primitives and core types.
    *   **Key Files:**
        *   `index.go`: Defines the base `Index` struct (embeds `Disk`, holds `NumRecords`, `RecordSize`, `PageSize`, `MaxOutcomeLen`, `RecordProcessingTime`).
        *   `heapfile.go`: `HeapFile` implementation.
        *   `btree.go`: Refined `BTreeIndex` implementation.
        *   `hashindex.go`: Refined `HashIndex` implementation (heuristic probabilities).
        *   `sortedfile.go`: `SortedFile` implementation.
        *   `lsm.go`: Simplified `LSMTree` implementation.
        *   `bitmap.go`: Simplified `BitmapIndex` implementation.
        *   `cache.go`: `Cache` component implementation (standalone component approach).

*   **`sdl/dsl/` (Future - Phase 4)**
    *   **Purpose:** Will contain the parser, interpreter, Abstract Syntax Tree (AST) nodes, and runtime environment for the user-facing System Design Language.

*   **`sdl/examples/` (Future - Phase 4/5)**
    *   **Purpose:** Will contain complete examples demonstrating how to define systems using the DSL (or Go API) and analyze them.

**4. Current Status:**

*   Phase 1 (Core & Primitives) and Phase 2 (Component Models & Basic Concurrency concepts) are largely complete, focusing on analytical modelling. Phase 3 (Advanced Concurrency/Queuing/State) was partially addressed by adding analytical Queue/Pool models.
*   Core `Outcomes` system with composition operators (`And`, `If`, `Map`) is stable.
*   Effective reduction strategies (Merge+Interpolate via `TrimToSize`/`TrimToSizeRanged`) are implemented and tested for `AccessResult` and `RangedResult`.
*   A good suite of primitive (`Disk`, `NetworkLink`, `Queue`, `ResourcePool`) and storage (`HeapFile`, `BTree`, `Hash`, `SortedFile`, `LSM`, `Bitmap`, `Cache`) components exist, modelled primarily using `Outcomes[AccessResult]` and analytical approximations.
*   Metric calculation and sampling capabilities are available.
*   **Known Limitations:** Accurate modelling of stateful interactions (esp. `ResourcePool.Release`), complex asynchronous patterns, and time-based batching variance remains challenging without moving towards DES. The current models favour analytical speed over detailed simulation accuracy in these areas.

**5. Next Steps (Phase 4):**

*   Focus on **Usability & System Composition**.
*   **Primary Task:** Implement the DSL parser and interpreter.
*   Other Tasks: Visualization hooks, build a component library (pre-defined compositions), create full system examples.

---

**Prompt for Next LLM Task (Phase 4 - DSL Implementation):**

```text
**Project Context:**

You are continuing development on the SDL (System Design Language) Go library. This library enables performance modelling (latency, availability) of system components using probabilistic `Outcomes` distributions and analytical approximations, avoiding full discrete-event simulation for speed.

**Current State:**
- The core Go library (organized conceptually into `core`, `results`, `metrics`, `primitives`, `storage` packages) provides:
    - `Outcomes[V any]` for probability distributions.
    - Composition operators (`And`, `If`, `Map`).
    - Result types (`AccessResult`, `RangedResult`) and converters.
    - Robust reduction strategies (`TrimToSize`, `TrimToSizeRanged` using Merge+Interpolate).
    - Metric calculations (`Availability`, `MeanLatency`, `PercentileLatency`).
    - Sampling (`Sample`, `SampleWithinRange`).
    - A suite of component models (`Disk`, `NetworkLink`, `Queue`, `ResourcePool`, `Cache`, `LSMTree`, `BTreeIndex`, `HashIndex`, `BitmapIndex`, etc.) primarily using `Outcomes[AccessResult]`.
- State evolution is handled externally by configuring component parameters between simulation runs. Complex async/stateful interactions have known limitations in the analytical model.
- The project documentation includes a detailed summary outlining the vision, concepts, components, status, and proposed code organization.

**Phase 4 Goal:** Focus on Usability & System Composition.

**Next Task (Phase 4, Task 1): Design and Start Implementing the DSL**

Your primary task is to **design the syntax for the user-facing System Design Language (SDL DSL)** and begin implementing the **parser** for it using a suitable Go parsing library (e.g., `participle`, `goyacc`, or even standard library tools if simple enough).

**Requirements:**

1.  **DSL Syntax Design:**
    *   Define a clear, intuitive, text-based syntax for users to:
        *   **Declare Components:** Define new component types (e.g., `component MyDatabase { ... }`).
        *   **Define Parameters:** Specify configurable parameters within components (e.g., `diskProfile: string = "SSD"; poolSize: int = 10`). Include support for basic types (string, int, float, bool).
        *   **Define Dependencies:** Declare instances of other components used internally (e.g., `cache: RedisCache; db: PostgresDB`).
        *   **Define Operations:** Declare methods/operations the component provides (e.g., `operation ReadUser(userId: string): Outcomes<UserResult> { ... }`). Specify input parameters and the *type* of `Outcomes` returned (e.g., `Outcomes<AccessResult>`, `Outcomes<Duration>`).
        *   **Define Operation Logic:** Specify the sequence of internal calls and compositions within an operation using a syntax that maps clearly to the underlying Go library operators (`And`, `If`, `Map`, calls to dependent components). Consider syntax for sequential calls, basic conditionals (if/else based on outcome properties like `Success`), mapping results, and potentially loops (e.g., `for i := 0; i < height; i++`).
        *   **Instantiate Components:** Create instances of defined components (e.g., `myDB: MyDatabase = { poolSize: 20 }`).
        *   **Orchestrate Analysis:** Define entry points or scenarios to trigger analysis (e.g., `analyze myDB.ReadUser("user123")`).
    *   **Reference Existing Ideas:** Refer to the high-level DSL ideas previously mentioned in the project documentation (e.g., the Database/ApiServer example) but refine and formalize the syntax.
    *   **Keep it Simple Initially:** Focus on core functionality; advanced features like complex types, modules, or imports can come later.

2.  **Parser Implementation (Initial):**
    *   Choose a suitable Go parsing library/technique. `participle` is often good for defining grammar directly in Go structs.
    *   Define the Abstract Syntax Tree (AST) node types in Go structs corresponding to the designed DSL syntax (e.g., `ComponentDecl`, `ParamDecl`, `OperationDef`, `CallExpr`, `SequenceExpr`, `IfExpr`).
    *   Implement the parser that takes DSL text as input and produces an AST.
    *   Add basic unit tests for the parser, verifying that simple DSL snippets parse correctly into the expected AST structure. Handle syntax errors gracefully.

**Focus:** Design a workable initial DSL syntax and implement the parser to generate an AST. **Do not implement the interpreter/evaluator** that executes the AST yet. That will be the subsequent step.

**Deliverable:** Go code for the AST node definitions and the parser implementation, along with basic parser tests. Documentation/examples of the designed DSL syntax.
