# SDL Project Summary & Onboarding Guide

**1. Vision & Goal:**

*   **Purpose:** The SDL (System Design Language) library models and analyzes the performance characteristics (latency, availability) of distributed system components using **analytical models** and **probabilistic composition**. It prioritizes **speed** for interactive "what-if" analysis over detailed Discrete Event Simulation (DES), focusing primarily on **steady-state** behavior.
*   **Use Cases:** Enable rapid analysis of system designs, bottleneck identification, SLO evaluation, performance exploration under different configurations, and potentially educational tools for system design reasoning.

**2. Overall Architecture:**

*   **Probabilistic Core (`core` package):** The foundation, providing generic `Outcomes[V]` distributions, composition operators (`And`, `Map`, `Split`, `Append`), complexity management via reduction (`Trim...`, `Merge...`, etc.), metric calculations (`Availability`, `PercentileLatency`, etc.), and a standardized `Analyze` primitive for testing the Go API.
*   **Analytical Components (`components` package):** Concrete Go models of system building blocks (Disk, Cache, Queue, Pool, LSM, BTree, etc.) using `sdl/core`. Key components (`ResourcePool`, `Queue`) use stateless analytical formulas (M/M/c, M/M/c/K).
*   **AST & Declarations (`decl` package):** Defines the Go structures for Abstract Syntax Tree (AST) nodes representing the SDL language (components, systems, operations, statements, expressions). Also includes the SDL type system (`types.go`), type inference logic (`infer.go`), the `TypeScope` helper for inference (`typescope.go`), and the generic environment structure `Env[Node]`.
*   **DSL Parser (`parser` package):** Parses SDL text into the AST defined in the `decl` package. It uses `goyacc` and a custom lexer.
*   **DSL Loader (`loader` package):** Manages loading of SDL source files, including resolving imports, handling cyclic dependencies, and orchestrating parsing and validation (including type inference). It now constructs a comprehensive `decl.Env[Node]` scope for type inference, correctly handling aliased imported symbols.
*   **DSL Execution Engine (`dsl` package - Planned):** This will contain the VM/evaluator to execute ASTs. The target is "Model V4" (implicit outcome handling in DSL syntax, explicit `Outcomes[V]` in Go). This is a major area for future work.
*   **Command Line Interface (`cmd/sdl` package):** A CLI tool (`sdl`) built with Cobra for interacting with the SDL toolchain (validating, listing, describing, running analyses - most are currently placeholders).
*   **Examples (`examples` package):** Demonstrates usage of the **Go API** to build and analyze systems (`bitly`, `gpucaller`, `notifier`) and includes some sample `.sdl` files.

**3. Code Structure & Key Packages:**

*   **`sdl/core/`**: Fundamental probabilistic modeling types and operations.
*   **`sdl/components/`**: Concrete Go component models.
    *   **`sdl/components/decl/`**: Declarative counterparts of components that generate ASTs for the DSL.
*   **`sdl/decl/`**: AST node definitions, type system, type inference logic, and `Env` structure.
*   **`sdl/parser/`**: DSL lexer and parser (goyacc-based).
*   **`sdl/loader/`**: File loading, import resolution, and validation orchestration.
*   **`sdl/dsl/`**: (Largely future work) DSL VM and execution logic.
*   **`sdl/cmd/sdl/`**: CLI application and command definitions.
*   **`sdl/examples/`**: Go API usage examples and sample SDL files.

**4. Current Status & Recent Work Summary:**

*   **Go API**: The `core` and `components` packages provide a functional Go API for probabilistic performance modeling.
*   **DSL Parsing & Loading**: The `parser` can process SDL syntax into an AST. The `loader` can manage file dependencies, parse them, and has recently been significantly improved to prepare the necessary scope (using `decl.Env[Node]` and `decl.TypeScope`) for type inference. This includes correctly resolving aliased imported symbols.
*   **Type Inference (`decl/infer.go`)**: Actively being refactored to work with the new `Env`-based `TypeScope`. The goal is to correctly infer types for all language constructs, including those involving imported and aliased symbols (like enums and components).
*   **DSL VM (`dsl` package)**: Remains largely unimplemented. The design target is "Model V4."
*   **CLI Commands**: Most commands in `cmd/sdl` are placeholders awaiting VM integration. The `validate` command leverages the loader and the developing type inference.

**5. Known Limitations (Explicitly Acknowledged):**

*   **Analytical Steady-State Focus**: Models average performance, not transient dynamics. Accuracy depends on the underlying analytical models (e.g., M/M/c assumptions for queues/pools).
*   **Stateless Analytical Components**: `ResourcePool`/`Queue` in `components` are primarily configuration-driven.
*   **VM Incomplete**: The DSL Execution Engine is the largest piece of outstanding work.
*   **Type Inference WIP**: While significantly improved for scope handling, the `decl/infer.go` logic needs to be completed and thoroughly tested for all language constructs and edge cases.
*   **Limited DSL Value Access**: The DSL currently focuses on composing `Outcomes` containers. Direct access to inner probabilistic values (e.g., a specific latency from an outcome bucket) within DSL logic is generally not supported, aligning with the "Model V4" philosophy.

**6. Next Steps:**

*   Finalize and test the refactored type inference logic in `decl/infer.go`.
*   Begin implementation of the DSL VM (`dsl` package).
*   Implement the actual functionality for the placeholder CLI commands.
*   (See `NEXTSTEPS.MD` for more detail).
