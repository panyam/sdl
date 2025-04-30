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

---

**Next Steps (Phase 4 - DSL Interpreter/VM):**

**Project Context:**

You are continuing development on the SDL (System Design Language) Go library. This library models steady-state system performance using probabilistic `Outcomes` distributions and analytical components (`sdl/core`, `sdl/components`). It prioritizes analytical speed over discrete-event simulation. Key components (`ResourcePool`) are now stateless and rate-driven. A standardized testing primitive (`core.Analyze`) is used extensively.

**Current State:**
- A new declarative layer (`sdl/decl`) has been implemented.
    - `sdl/decl/ast/ast.go` defines AST nodes (`LiteralExpr`, `IdentifierExpr`, `CallExpr`, `AndExpr`, `ParallelExpr`, `InternalCallExpr`, `RepeatExpr`, `FanoutExpr`, etc.).
    - `sdl/decl/components.go` defines declarative component structs holding configuration and methods (`*AST()`) that generate these `ast.Expr` trees, representing operations rather than executing them directly. Tests verify correct AST generation.
- The core library (`sdl/core`) provides the building blocks: `Outcomes[V]`, composition functions (`And`, `Map`, `Split`), result types (`AccessResult`), reduction (`TrimToSize`), metrics (`Availability`, etc.), and distribution helpers (`NewDistributionFromPercentiles`).
- Known limitations (steady-state focus, stateless pool, parallel composition approximations, missing `Filter`/`Repeat`/`FanoutAnd` primitives) are documented. Design documents (`DSL.md`, `ROADMAP.MD`) outline the overall plan.

**Phase 4 Goal:** Focus on Usability & System Composition.

**Next Task (Phase 4, Task 2): Implement the DSL Interpreter / VM**

Your primary task is to **implement the initial version of the interpreter/VM** that can execute the Abstract Syntax Trees (ASTs) generated by the `sdl/decl` package. This VM will orchestrate calls to the `sdl/core` library functions to perform the actual probabilistic calculations.

**Requirements:**

1.  **Location:** Implement the core interpreter logic likely in `sdl/dsl/interpreter.go` (or `vm.go`).
2.  **Input:** The interpreter should take an `ast.Expr` as input (typically the target expression from an `analyze` block or a component operation's body).
3.  **Execution Model:** Implement a **stack-based execution model**. Expressions are evaluated, pushing their results (primarily `*core.Outcomes[V]` objects) onto the VM stack. Operators consume operands from the stack and push results back.
4.  **Environment:** Implement a basic environment (symbol table) for resolving identifiers (`ast.IdentifierExpr`) to runtime values (e.g., component instances created during system setup, intermediate results stored in variables). Handle basic scoping if necessary later.
5.  **Instruction Handlers (Eval Logic):** Implement evaluation logic for key AST nodes:
    *   **`ast.LiteralExpr`:** Parse value, wrap in a deterministic `*core.Outcomes[T]`, push onto stack.
    *   **`ast.IdentifierExpr`:** Look up name in environment, push value onto stack.
    *   **`ast.CallExpr`:** Evaluate function expression, evaluate args, resolve Go method/function, call it (it might return `*core.Outcomes[V]` directly or another `ast.Expr` to evaluate recursively), push result. Handle calls to `sdl/decl` component `*AST()` methods by recursively evaluating the returned AST.
    *   **`ast.MemberAccessExpr`:** Evaluate receiver, use member name for method lookup (in `CallExpr`) or potentially field access later.
    *   **`ast.AndExpr`:** Evaluate `Left`, evaluate `Right`. Pop both results (`*Outcomes[V1]`, `*Outcomes[V2]`). **Inspect types V1/V2** (using type assertions/reflection initially). Select appropriate **sequential reducer** (e.g., `core.AndAccessResults`). Call `core.And(LeftResult, RightResult, reducer)`. **Apply intermediate trimming** based on context/options. Push result.
    *   **`ast.ParallelExpr`:** Similar to `AndExpr`, but select/implement a **parallel reducer** (approximate: Max Latency, AND Success for known types). Call `core.And` with this parallel reducer. Apply trimming. Push result.
    *   **`ast.InternalCallExpr`:** Lookup `FuncName` in a registry of internal VM helper functions (e.g., `GetDiskReadProfile`, `CalculateBTreeHeight`, `ScaleLatency`). Evaluate args, execute the helper Go function, push result.
6.  **Intermediate Trimming:** Implement logic within the VM handlers for composition operators (`AndExpr`, `ParallelExpr`, `RepeatExpr`, `FanoutExpr`) to check the length of the resulting `Outcomes` object and apply the appropriate `core.TrimToSize` (or similar) function if it exceeds a configured threshold (`MaxOutcomeLen` from options/context). Type inspection is needed here too.
7.  **Type Handling:** The VM *must* handle type heterogeneity. Use type assertions (`switch v := value.(type)`) or reflection (`reflect` package) on the `*core.Outcomes[V]` objects popped from the stack to determine the underlying type `V` when selecting reducers, trimmers, or specific logic.
8.  **Initial Focus:** Implement handlers for `LiteralExpr`, `IdentifierExpr`, `CallExpr` (calling simple Go funcs returning `Outcomes`), `AndExpr`, and `InternalCallExpr` (with a few basic helpers like `GetDiskReadProfile`). Implement the basic stack machine loop and intermediate trimming after `AndExpr`.

**Do Not Implement Yet:** `RepeatExpr`, `FanoutExpr`, `SwitchExpr`, `FilterExpr`, complex environment scoping, full DSL parsing integration (assume AST input is given).

**Deliverable:** Go code for the initial stack-based VM/Interpreter (`interpreter.go` and supporting types/environment struct). Include basic tests (`interpreter_test.go`) that construct simple ASTs manually and verify the interpreter produces the expected final `*core.Outcomes[V]` result on the stack (check `Len()`, `TotalWeight()`, maybe `Availability()` if applicable).
