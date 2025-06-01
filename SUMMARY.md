# SDL Project Summary & Onboarding Guide

**1. Vision & Goal:**

*   **Purpose:** The SDL (System Design Language) library models and analyzes the performance characteristics (latency, availability) of distributed system components using **analytical models** and **probabilistic composition**. It prioritizes **speed** for interactive "what-if" analysis over detailed Discrete Event Simulation (DES), focusing primarily on **steady-state** behavior.
*   **Use Cases:** Enable rapid analysis of system designs, bottleneck identification, SLO evaluation, performance exploration under different configurations, and potentially educational tools for system design reasoning.

**2. Overall Architecture:**

*   **Probabilistic Core (`./core`):** The foundation, providing generic `Outcomes[V]` distributions, composition operators (`And`, `Map`, `Split`, `Append`), complexity management via reduction (`Trim...`, `Merge...`, etc.), metric calculations (`Availability`, `PercentileLatency`, etc.), and a standardized `Analyze` primitive for testing the Go API. Relies on external reducers for composition.
*   **Declarative Layer(`./decl`):** This represents the constructs for a parse tree and some runtime values for describing a system.  
*   **Analytical Components (`./components`):** Concrete Go models of system building blocks (Disk, Cache, Queue, Pool, LSM, BTree, etc.) using `./core`. Key components (`ResourcePool`, `Queue`) use stateless analytical formulas (M/M/c, M/M/c/K).
*   **Declarative Layer (`./components/decl`):** Defines component structures and methods (`Read()`, `Find()`, etc.) that generate Abstract Syntax Trees (ASTs) representing operation logic, separating definition from execution.
*   **Parser(`./parser`):** A parser with the grammer (`grammar.y`) that loads a system defined in our DSL and represents using the parse and syntax trees represented in the decl module.
*   **Loader Layer (`./loader`):** The loader takes a grammer parsed by the parser module and performs load time checks (like resolving files, package managing, type inference, missing/valid methods etc).  The output of the loader phase is a well formed expression tree that can be evaluated in various forms.
*   **Runtime Layer (`./runtime`):** The runtime module facilitates execution of loaded system design expression trees.  It holds a loader (for resolving and loading packages) and environments for performing various runs.  Currently a SimpleEval type performs a "simulation" run where a single call end to end is sampled.  Other kinds of Evals will be added (eg an E2E probabilistic evaluator instead of a single E2E call sampler).
*   **Examples (`./examples`):** Demonstrates usage of the **Go API** to build and analyze systems (`bitly`, `gpucaller`, `notifier`).

**3. Code Structure & Package Overview:**

*   **`./` (Root)**
    *   `go.mod`, `go.sum`: Go module definitions.
    *   `Makefile`: Build/test/watch commands.
    *   `DSL.md`, `ROADMAP.MD` (Potentially outdated): Design notes and future plans.
    *   `SUMMARY.md`: (This document) Project overview.
    *   `NEXTSTEPS.md`: Outstanding tasks and future work.
    *   `./SYNTAX.md`: DSL syntax design discussion and decisions.
    *   *(Future)* `README.md`: High-level description and usage examples.

*   **`./core/`**: Fundamental, generic types and operations. (See `./core/SUMMARY.md`).
*   **`./components/`**: Concrete system building blocks using `./core`. Includes `decl/` sub-package for AST generation. (See `./components/SUMMARY.md`).
*   **`./decl/`**: Declarative/Syntax Tree representation of our DSL. (See `./decl/SUMMARY.md`).
*   **`./parser/`**: A parser written in goyacc for parsing systems in our DSL. (See `./parser/SUMMARY.md`).
*   **`./loader/`**: The package loader and name space resolver for our DSL files organized in various ways. The loader also invokes the Parser when loading resolved files and then performs various validations (like Inference - `./loader/infer.go`) (See `./loader/SUMMARY.md`).
*   **`./runtime/`**: The runtime manager for loading (using a Loader) various files/modules and resolving them and preparing environments for kicking off Evaluations (SimpleEval for now). (See `./runtime/SUMMARY.md`).
*   **`./examples/`**: Examples using the Go API. (See `./examples/SUMMARY.md`).

**4. Current Status & Recent Work Summary:**

*   The core analytical library (`./core`, `components`) is functional for modelling steady-state performance via probabilistic composition using the **Go API**. Standardized testing via `core.Analyze` is implemented. Key components are stateless and rate-driven.
*   The declarative layer (`./decl`) generates ASTs for component operations.
*   The **DSL Parser (`./parser`) and Loader (`./loader`)** are now implemented.
*   **Recent Work:** Focused on implementing DSL Runtime and Evaluator.   The plot command (`cmd/sdl/commands/plot.go`) shows how a example SDL file can be loaded a particular system within it can be invoked to obtain various runtime plots of a system in action.

**5. Known Limitations (Explicitly Acknowledged):**

*   **Analytical Steady-State:** Models average performance, not transient bursts/dynamics. Accuracy depends on analytical models (e.g., M/M/c assumptions).
*   **Stateless Components:** `ResourcePool`/`Queue` rely purely on configured rates.
*   **Profile Accuracy:** Depends on realistic input `Outcomes` distributions (especially for leaf components like `Disk` or work profiles like `gpuwork`).
*   **Approximations:** Batcher wait time, parallel execution modeling (pending), fan-out costs in examples.

**6. Next Steps:**

*   See `NEXTSTEPS.md` for a detailed breakdown.
