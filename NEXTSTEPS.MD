# SDL Project: Next Steps & Outstanding Work

This document outlines the key areas for future development and refinement in the SDL project.

## 1. Implement Full Functionality for CLI Commands (`cmd/sdl`)

Once the VM is functional, implement the actual logic for:
*   **`sdl run`**: Use the DSL Driver to execute specified analyses and display results (metrics from the final `Outcomes`).
*   **`sdl diagram`**: Generate diagrams (e.g., DOT, Mermaid) by:
    *   **Static:** Traversing the AST of a `SystemDecl` to find instances and their connections (from `Overrides`).
    *   **Dynamic:** Requires VM execution tracing or analysis of `Outcomes` paths from an `analyze` run.
*   **`sdl plot`**: Integrate with a plotting library. Take results from `sdl run --json-results` (which should output `AnalysisResultWrapper` or `Outcomes` data) and generate specified plots (CDFs, histograms).
*   **`sdl trace`**: Implement VM execution tracing to capture operation call sequences and their (probabilistic) latencies.

## 2. Enhance Language Features & Semantics (DSL & `decl`)

*   **Built-in Functions:** Define and implement a standard library of built-in functions (e.g., `len()`, math functions, type conversions, `log()`).
*   **Error Handling in DSL:** Define error types and try/catch mechanisms if desired.
*   **Advanced `distribute` / `sample`:** Ensure full expressiveness for defining probabilistic branches and sampling from distributions.
*   **Loops (`for`):** Define semantics for `for` loops (e.g., fixed iterations, iterating over lists if lists are added as a first-class type).
*   **Concurrency (`go`, `wait`):** Finalize semantics for concurrent operations. `go` likely spawns an operation whose `Outcomes` are stored. `wait` would compose these outcomes.
*   **System Parameters & Overrides:** Ensure robust handling of parameter overrides in `SystemDecl` and their propagation to component instances.
*   **Global `options`:** Full implementation for `options` blocks (global and component-level) to control things like `MaxOutcomeLen`, default profiles, etc.

## 3. Testing & CI

*   **VM Test Suite:** Extensive tests for all AST node evaluations and VM behaviors.
*   **End-to-End CLI Tests:** Test CLI commands with various `.sdl` files.
*   **Benchmarking:** For core operations and VM execution.

## 4. Documentation & Examples

*   **Comprehensive DSL Manual:** Full syntax, semantics, and examples for all language features.
*   **Go API Documentation:** Improve Godoc for `core`, `components`, and `decl` packages.
*   **More `.sdl` Examples:** Create a suite of examples demonstrating various modeling techniques and DSL features.

## 5. Parser Enhancements (`parser` package)

*   **Error Recovery:** Improve error reporting and recovery in the `goyacc` parser.
*   **Syntax Refinements:** Address any awkward syntax identified during broader use.

This provides a roadmap. The immediate focus should be on completing the type inference refactoring, followed by starting the VM implementation.
