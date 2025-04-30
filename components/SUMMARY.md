# SDL Components Package Summary (`sdl/components`)

**Purpose:**

This package provides concrete implementations of common distributed system components, built upon the abstractions and primitives defined in `sdl/core`. These components model the performance characteristics (latency, availability) of specific building blocks like disks, caches, queues, etc.

**Key Concepts & Components:**

1.  **Foundation:** Relies entirely on `sdl/core` for probabilistic representation (`Outcomes`, `AccessResult`) and composition. Defines type aliases (`Duration`, `Outcomes`) in `aliases.go`.
2.  **Component Models:**
    *   Each component (e.g., `Disk`, `Cache`, `Queue`) is represented by a Go struct.
    *   They typically have an `Init()` or `New...()` constructor to set default parameters or accept configuration.
    *   Key methods representing operations (`Read`, `Write`, `Acquire`, `Submit`, `Find`, `Insert`, etc.) return `*core.Outcomes[AccessResult]` or `*core.Outcomes[Duration]`, encapsulating the performance model of that operation.
    *   Models often compose operations on underlying dependencies (e.g., `BTreeIndex.Find` composes multiple `Disk.Read` operations).
3.  **Stateless Analytical Models:**
    *   Key flow-control components like `ResourcePool` (`resourcepool.go`) and `Queue` (`queue.go`, `mm1queue.go`) have been refactored to use **stateless analytical models** (M/M/c, M/M/c/K, M/M/1 approximations).
    *   Their performance (especially waiting time) is predicted based *solely* on configured parameters (arrival rate `lambda`, service time `Ts`, pool/server size `c`, capacity `K`) rather than tracking internal state like used resources or current queue length. This aligns with the project's goal of fast, steady-state analysis.
4.  **Specific Components Implemented:**
    *   **Primitives:** `Disk` (SSD/HDD profiles), `NetworkLink` (latency, jitter, loss), `Queue` (M/M/c/K), `MM1Queue`, `ResourcePool` (M/M/c).
    *   **Storage/Indexing:** `Cache` (hit/miss), `Index` (base), `BTreeIndex`, `HashIndex`, `LSMTree`, `BitmapIndex`, `HeapFile`, `SortedFile`.
    *   **Orchestration:** `Batcher` (size/time based, uses `BatchProcessor` interface).
5.  **Declarative Layer (`decl/` sub-package):**
    *   Contains parallel structures for each component (e.g., `decl.Disk`, `decl.BTreeIndex`).
    *   These structs hold configuration parameters, often as `dsl.Expr` types.
    *   They provide methods (`ReadAST`, `FindAST`, etc.) that **generate Abstract Syntax Trees (ASTs)** defined in `sdl/dsl/ast.go`, representing the intended operation logic rather than executing it directly.
    *   This layer separates the *definition* of component interactions from their *execution* (which is handled by the `sdl/dsl` interpreter). Includes `components_test.go` verifying AST generation.
6.  **Testing:**
    *   Component tests (`*_test.go`) consistently use the `core.Analyze` primitive with relevant expectations (`ExpectAvailability`, `ExpectP99`, etc.) and assertions (`Assert`, `AssertFailure`) to verify the correctness and plausibility of the component models.

**Current Status:**

*   Provides a reasonably comprehensive suite of standard system components.
*   Key components model steady-state behavior analytically.
*   The `decl` sub-package provides the foundation for the DSL by generating ASTs for component operations.
*   Testing is standardized via `core.Analyze`.
