# SDL Components Package Summary (`sdl/components`)

**Purpose:**

This package provides concrete Go implementations of common distributed system components, built upon the abstractions and primitives defined in `sdl/core`. These components model the performance characteristics (latency, availability) of specific building blocks like disks, caches, queues, etc., intended for use both directly via the Go API and as the underlying implementation for the DSL.

**Key Concepts & Components:**

1.  **Foundation:** Relies entirely on `sdl/core` for probabilistic representation (`Outcomes`, `AccessResult`, `Duration`) and composition operators (`And`, `Map`, `Split`, `Append`, `TrimToSize`, etc.). Defines type aliases in `aliases.go`.
2.  **Component Models:**
    *   Each component (e.g., `Disk`, `Cache`, `Queue`, `ResourcePool`, `BTreeIndex`, `HashIndex`, `LSMTree`, `NetworkLink`, `Batcher`, `HeapFile`, `SortedFile`) is represented by a Go struct.
    *   They typically have an `Init()` or `New...()` constructor to set default parameters or accept configuration.
    *   Key methods representing operations (`Read`, `Write`, `Acquire`, `Enqueue`, `Dequeue`, `Find`, `Insert`, `Submit`, `Transfer`, etc.) return `*core.Outcomes[V]` (e.g., `*core.Outcomes[AccessResult]`, `*core.Outcomes[Duration]`, potentially `*core.Outcomes[bool]` in future). These return values encapsulate the probabilistic performance model of that operation.
    *   Models often compose operations on underlying dependencies (e.g., `BTreeIndex.Find` composes multiple `Disk.Read` operations).

3.  **Stateless Analytical Models:**
    *   Key flow-control components like `ResourcePool` (`resourcepool.go`) and `Queue` (`queue.go`, `mm1queue.go`) utilize **stateless analytical models** (M/M/c, M/M/c/K, M/M/1 approximations).
    *   Their performance (especially waiting time) is predicted based *solely* on configured parameters (arrival rate `lambda`, service time `Ts`, pool/server size `c`, capacity `K`) rather than tracking internal state like used resources or current queue length. This aligns with the project's goal of fast, steady-state analysis.

4.  **Declarative Layer (`decl/` sub-package):**
    *   Contains parallel Go structs for many components (e.g., `decl.Disk`, `decl.BTreeIndex`).
    *   These structs hold configuration parameters, often as `dsl.Expr` types (defined in `sdl/dsl/ast.go`).
    *   They provide methods (`Read()`, `Find()`, etc.) that **generate Abstract Syntax Trees (ASTs)**, representing the intended operation logic rather than executing it directly.
    *   This layer separates the *definition* of component interactions from their *execution* (which is handled by the `sdl/dsl` vm). Includes `components_test.go` verifying AST generation.

5.  **Testing:**
    *   Component tests (`*_test.go`) consistently use the `core.Analyze` primitive with relevant `Expectation`s (`ExpectAvailability`, `ExpectP99`, etc.) and assertions (`Assert`, `AssertFailure`) to verify the correctness and plausibility of the component models when used via the Go API.

**Current Status:**

*   Provides a reasonably comprehensive suite of standard system components with defined performance models based on `sdl/core`.
*   Key components model steady-state behavior analytically (Queues, Pools) or via probabilistic composition (Indexes, Cache, Disk).
*   The `decl` sub-package provides the AST generation capabilities needed for the DSL front-end.
*   Testing via `core.Analyze` validates the Go API usage and component behavior.
*   The components are ready to be instantiated and called by the DSL VM as it executes ASTs.
