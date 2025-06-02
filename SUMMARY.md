# SDL (System Design Language) Project Summary

**Version:** As of current interactions (Post-Static Diagramming Enhancements)

**1. Vision & Goal:**

*   **Purpose:** The SDL project aims to provide a language and toolchain for modeling and analyzing the performance characteristics (e.g., latency, availability) of distributed systems. It focuses on enabling rapid, interactive "what-if" analysis primarily through analytical models and probabilistic composition for steady-state behavior. It also supports simulation-like execution for single-path analysis and metric gathering.
*   **Use Cases:** Enable rapid analysis of system designs, bottleneck identification, SLO evaluation, performance exploration under different configurations, and potentially educational tools for system design reasoning.

**2. Overall Architecture & Key Packages:**

The project is a Go-based system composed of several key packages:

*   **`./core` (Probabilistic Core):**
    *   Provides the fundamental Go API for probabilistic performance modeling. Key features include `Outcomes[V]` for discrete probability distributions, `AccessResult` for success/latency modeling, composition operators (`And`, `Map`), complexity reduction (`TrimToSize`), and metric calculations (`Availability`, `PercentileLatency`). It defines the `Metricable` interface for types whose performance can be measured.

*   **`./components` (Concrete Analytical Components):**
    *   Offers Go implementations of common system building blocks (Disk, Cache, Queue, ResourcePool, HashIndex, BTreeIndex, LSMTree, etc.) using `./core` primitives.
    *   Many components, like `Queue` and `ResourcePool`, use stateless analytical models (M/M/c, M/M/1) for performance prediction.
    *   Includes a `decl/` sub-package (`./components/decl`) containing Go wrappers for these components. These wrappers make the core Go components callable as "native" components from the SDL runtime, handling the translation between `core.Outcomes` and `decl.Value` (including latency).

*   **`./decl` (DSL Declarations & Types):**
    *   Defines the Go structs for the Abstract Syntax Tree (AST) of the SDL. This includes all language constructs (components, systems, methods, statements, expressions, types).
    *   Defines the core `Value` type system used by the SDL runtime, which notably includes a `Time` field to track latency during evaluation. It also defines the `Type` system for representing SDL types (e.g., `Int`, `ComponentType`, `OutcomesType`).
    *   Provides `Env[T]` for managing scoped symbols.

*   **`./parser` (DSL Parser):**
    *   A `goyacc`-based parser that translates SDL text into the AST defined in the `./decl` package.
    *   Includes a lexer (`lexer.go`) and grammar (`grammar.y`).

*   **`./loader` (File Loading & Validation):**
    *   Manages loading of parsed SDL files, resolving import paths, and handling cyclic dependencies.
    *   Crucially, it orchestrates the validation process, including **type inference** (`infer.go`, `typescope.go`) which populates a type environment (`Env[Node]`) and checks the semantic correctness of the AST.

*   **`./runtime` (DSL Execution Engine):**
    *   Responsible for executing loaded and validated SDL models.
    *   Features `SimpleEval`, an interpreter that traverses the AST to evaluate expressions and statements, managing runtime state in an `Env[Value]`. It supports calls to both SDL-defined methods and native Go components (via reflection and the wrappers in `components/decl`).
    *   The `Value.Time` field is used to accumulate latency during `SimpleEval` execution.
    *   Includes helpers like `RunCallInBatches` (used by the `plot` CLI command) to execute specific methods repeatedly.

*   **`./cmd/sdl` (Command Line Interface):**
    *   The main user-facing tool, `sdl`, built with Cobra.
    *   Provides commands for:
        *   `validate`: Parses and type-checks SDL files.
        *   `list`: Lists entities within an SDL file.
        *   `describe`: Shows details of a specific entity.
        *   `plot`: Executes a system method multiple times (using `./runtime`) and generates SVG latency plots (using `./cmd/sdl/plotter.go`).
        *   `diagram static`: Generates static system diagrams from SDL definitions, showing component instances and their connections. Supports DOT, Mermaid, Excalidraw (JSON), and direct SVG output formats. (Recent work focused heavily on implementing and refining this).
    *   Placeholder commands for `run`, dynamic `diagram`, and `trace` exist for future implementation.

*   **`./examples`:**
    *   Contains Go API examples (`native/`) demonstrating the use of `./core` and `./components`.
    *   Includes sample `.sdl` files used for testing the parser, loader, and other SDL features.

**3. Current Status & Recent Work Summary:**

*   **Core Libraries (`core`, `components` Go API):** Mature and functional for modeling steady-state performance.
*   **DSL Frontend (`parser`, `loader`):** Robust parsing, loading, import resolution, and type inference capabilities are in place. Type inference scoping has been a focus of recent refinement.
*   **Runtime (`runtime`):** `SimpleEval` provides a functional interpreter for a significant subset of SDL, capable of single-path simulation and latency tracking. It powers the `plot` command.
*   **CLI (`cmd/sdl`):** `validate`, `list`, `describe` are functional. The `plot` command provides performance visualization. **Static diagram generation (`diagram static`) is now implemented with support for multiple output formats (DOT, Mermaid, Excalidraw, SVG).**
*   **Known Limitations:** The primary analytical model is for steady-state behavior. Current `SimpleEval` is for single-path simulation rather than full probabilistic outcome calculation across all paths (which would be an `OutcomesEval`). Concurrency features (`go`/`wait`) in the DSL are not yet fully implemented in the runtime.

**4. Build & Test:**

*   `Makefile` provides targets for building the `sdl` CLI tool (`make sdl`), running tests (`make test`), and goyacc generation for the parser.

This summary provides a high-level overview of the SDL project, its architecture, current capabilities, and recent developments.
