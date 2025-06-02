# SDL Components Package Summary (`components` package)

**Purpose:**

This package provides concrete Go implementations of common distributed system components, built upon the abstractions and primitives defined in `sdl/core`. These components model the performance characteristics (latency, availability) of specific building blocks like disks, caches, queues, etc. They are intended for use directly via their Go API and also serve as the underlying native implementations for components declared in the SDL DSL.

**Key Concepts & Structure:**

1.  **Core Dependency (`sdl/core`):**
    *   All components fundamentally rely on `sdl/core` for probabilistic representation (`Outcomes`, `AccessResult`, `Duration`) and composition operators (`And`, `Map`, `Split`, `Append`, `TrimToSize`, etc.).
    *   Type aliases for core types are defined in `aliases.go`.

2.  **Component Go Models (e.g., `disk.go`, `cache.go`, `hashindex.go`):**
    *   Each component (e.g., `Disk`, `Cache`, `Queue`, `ResourcePool`, `BTreeIndex`, `HashIndex`, `LSMTree`, `NetworkLink`, `Batcher`, `HeapFile`, `SortedFile`) is represented by a Go struct.
    *   They typically have an `Init()` method and/or a `New...()` constructor to set default parameters and perform initial setup.
    *   Key methods representing operations (`Read()`, `Write()`, `Acquire()`, `Enqueue()`, `Find()`, etc.) return `*core.Outcomes[V]` (usually `*core.Outcomes[AccessResult]` or `*core.Outcomes[Duration]`). These return values encapsulate the probabilistic performance model of that operation.
    *   Models often compose operations on underlying dependencies (e.g., `BTreeIndex.Find` composes multiple `Disk.Read` operations internally).

3.  **Stateless Analytical Models:**
    *   Key flow-control components like `ResourcePool` (`resourcepool.go`) and `Queue` (`queue.go`, `mm1queue.go`) utilize **stateless analytical models** (M/M/c, M/M/c/K, M/M/1 approximations).
    *   Their performance (especially waiting time) is predicted based *solely* on configured parameters (arrival rate `lambda`, service time `Ts`, pool/server size `c`, capacity `K`) rather than tracking internal state. This aligns with SDL's goal of fast, steady-state analysis.

4.  **Native Component Wrappers (`decl/` sub-package, e.g., `components/decl/disk.go`):**
    *   This sub-package contains Go structs that act as **wrappers** around the actual component implementations (e.g., `components.Disk` is wrapped by `components/decl.Disk`).
    *   These wrappers conform to the `runtime.NativeObject` interface (implicitly or explicitly through embedding `NWBase`).
    *   Their methods (e.g., `decl.Disk.Read()`) call the corresponding method on the wrapped component (e.g., `components.Disk.Read()`).
    *   Crucially, they **convert the `*core.Outcomes` returned by the core components into `decl.Value` objects** suitable for the SDL runtime. This involves mapping `core.AccessResult` or `core.Duration` to `decl.Value`, ensuring that latency is transferred to `decl.Value.Time` (often by sampling if the outcome is probabilistic, or by representing the entire distribution as a `Value` of type `Outcomes[T]`).
    *   These wrappers are what the SDL `runtime` instantiates and interacts with when a `native component` is declared in an SDL file.

5.  **Testing:**
    *   Component tests (`*_test.go`) consistently use the `core.Analyze` primitive with relevant `Expectation`s to verify the correctness and plausibility of the component models when used via their direct Go API.

**Current Status:**

*   Provides a reasonably comprehensive suite of standard system components with defined performance models based on `sdl/core`.
*   The `decl/` sub-package provides the necessary wrappers to make these Go components callable from the SDL runtime as native components.

**Relationship with other packages:**

*   `sdl/core`: The foundation for all performance modeling within this package.
*   `sdl/decl`: The `components/decl` wrappers convert to/from `sdl/decl.Value`.
*   `sdl/runtime`: Instantiates and calls methods on the `components/decl` wrappers when executing SDL that uses native components.
