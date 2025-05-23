# SDL Project: Next Steps & Outstanding Work

This document outlines the key areas for future development and refinement in the SDL project.

## 1. Finalize Type Inference & Static Analysis (`decl` package)

*   **Complete `decl/infer.go` Refactoring:**
    *   Ensure all statement types (`SwitchStmt`, `ForStmt`, `DelayStmt`, `WaitStmt`, `GoStmt`, `LogStmt`) are correctly handled in `InferTypesForStmt`.
    *   Verify type checking for `uses` dependencies in `InferTypesForSystemDeclBodyItem` for `InstanceDecl` overrides, especially when the assigned value is an instance identifier.
    *   Thoroughly test `TypeDecl.TypeUsingScope()` for all kinds of type declarations (primitives, generics like List/Outcomes, and named types resolved from scope).
    *   Ensure `TypeDecl.SetResolvedType()` is correctly used and `ResolvedType()` is consistently available after the first inference pass over signatures.
*   **Comprehensive Testing for Type Inference:**
    *   Create targeted `.sdl` files that test various type checking scenarios, including:
        *   Correct type usages.
        *   Type mismatches in assignments, operations, function calls, returns.
        *   Resolution of imported and aliased enums and components.
        *   Correct usage of enum values (e.g., `Http.StatusOk`).
        *   Scope resolution for `let` variables in nested blocks and methods.
        *   `self` usage within component methods.
    *   Ensure `loader_test.go` (e.g., `TestBitly`) passes with the new type inference logic.
*   **Semantic Checks Beyond Basic Type Inference:**
    *   Unused variable detection (`let` bindings).
    *   Checks for `return` statements in all paths of non-void functions/methods.
    *   Validation of `analyze` block targets (e.g., ensuring it calls a valid operation).
    *   More specific checks for `distribute` expressions (e.g., sum of probabilities if `TotalProb` is not given).

## 2. Implement DSL Execution Engine (VM - `dsl` package)

This is the largest piece of outstanding work and the core of making the DSL executable.

*   **Design & Implement Core VM Structures (`dsl/vm.go`, `dsl/environment.go` - `Env` is now in `decl`):
    *   **VM/Evaluator:** Engine to walk the AST (`*decl.FileDecl`).
    *   **Runtime Environment:** Manage scopes for instances, `let` variables, and potentially function call stacks. `decl.Env[Node]` can be a basis, but runtime might need `decl.Env[RuntimeValue]` or similar for evaluated expression results.
    *   **Runtime Values:** Define how DSL types (Int, Float, String, Bool, Duration, Enum values, Component instances, Outcomes) are represented at runtime.
    *   **Stack (if needed):** For expression evaluation or function call management.
*   **Implement AST Node Evaluation (Model V4 - Implicit Outcomes):
    *   **Literals, Identifiers:** Resolve to runtime values (potentially wrapped in deterministic `Outcomes`).
    *   **Expressions (`BinaryExpr`, `UnaryExpr`, `TupleExpr`, `SampleExpr`):** Evaluate operands and perform operations. If operands are `Outcomes`, these operations translate to `core.Map` or custom composition logic.
    *   **`DistributeExpr`:** Translate to `core.Outcomes` construction based on weights and case bodies.
    *   **`CallExpr` (Component Methods):**
        *   Resolve method on the target component instance.
        *   Set up a new scope for the method call, binding arguments.
        *   Recursively evaluate the method's body (`BlockStmt`).
        *   The Go component methods (from `components` package) will return `*core.Outcomes[V]`. The VM needs to handle this.
    *   **`CallExpr` (Internal/Built-in Functions):** Implement calls to helper functions (e.g., `CalculateBTreeHeight`, `ScaleLatency`, `MMCKEnqueue`, `PoolAcquireMMc` referenced in `components/decl`). These map to specific Go functions in the `components` or `core` package.
    *   **Statements (`LetStmt`, `ReturnStmt`, `IfStmt`, `ExprStmt`, `BlockStmt`):**
        *   `LetStmt`: Evaluate RHS, store result in runtime environment.
        *   `IfStmt`: Evaluate condition (which might be an `Outcomes[bool]`). This requires a way to branch execution probabilistically based on the `Outcomes` of the condition, or to split the current context's `Outcomes` and evaluate branches. (Model V4 detail).
        *   `ReturnStmt`: Propagate the return value (an `Outcomes` object).
    *   **Control Flow (`ForStmt`, `SwitchStmt` - if kept in DSL V1):** Define their evaluation semantics.
*   **Implicit `Outcomes` Handling:**
    *   The VM automatically wraps/unwraps `core.Outcomes` when transitioning between DSL expressions and calls to Go component methods.
    *   Sequential DSL statements imply `core.And` composition on their underlying `Outcomes` results.
    *   Manage reduction (`TrimToSize`) automatically based on `options` or defaults.
*   **Driver (`dsl/driver.go` - Planned):** Entry point to execute a system's `analyze` block. It would instantiate the system, find the target operation, and kick off VM evaluation.

## 3. Implement Full Functionality for CLI Commands (`cmd/sdl`)

Once the VM is functional, implement the actual logic for:
*   **`sdl run`**: Use the DSL Driver to execute specified analyses and display results (metrics from the final `Outcomes`).
*   **`sdl diagram`**: Generate diagrams (e.g., DOT, Mermaid) by:
    *   **Static:** Traversing the AST of a `SystemDecl` to find instances and their connections (from `Overrides`).
    *   **Dynamic:** Requires VM execution tracing or analysis of `Outcomes` paths from an `analyze` run.
*   **`sdl plot`**: Integrate with a plotting library. Take results from `sdl run --json-results` (which should output `AnalysisResultWrapper` or `Outcomes` data) and generate specified plots (CDFs, histograms).
*   **`sdl trace`**: Implement VM execution tracing to capture operation call sequences and their (probabilistic) latencies.

## 4. Enhance Language Features & Semantics (DSL & `decl`)

*   **Built-in Functions:** Define and implement a standard library of built-in functions (e.g., `len()`, math functions, type conversions, `log()`).
*   **Error Handling in DSL:** Define error types and try/catch mechanisms if desired.
*   **Advanced `distribute` / `sample`:** Ensure full expressiveness for defining probabilistic branches and sampling from distributions.
*   **Loops (`for`):** Define semantics for `for` loops (e.g., fixed iterations, iterating over lists if lists are added as a first-class type).
*   **Concurrency (`go`, `wait`):** Finalize semantics for concurrent operations. `go` likely spawns an operation whose `Outcomes` are stored. `wait` would compose these outcomes.
*   **System Parameters & Overrides:** Ensure robust handling of parameter overrides in `SystemDecl` and their propagation to component instances.
*   **Global `options`:** Full implementation for `options` blocks (global and component-level) to control things like `MaxOutcomeLen`, default profiles, etc.

## 5. Documentation & Examples

*   **Comprehensive DSL Manual:** Full syntax, semantics, and examples for all language features.
*   **Go API Documentation:** Improve Godoc for `core`, `components`, and `decl` packages.
*   **More `.sdl` Examples:** Create a suite of examples demonstrating various modeling techniques and DSL features.

## 6. Testing & CI

*   **VM Test Suite:** Extensive tests for all AST node evaluations and VM behaviors.
*   **End-to-End CLI Tests:** Test CLI commands with various `.sdl` files.
*   **Benchmarking:** For core operations and VM execution.

## 7. Parser Enhancements (`parser` package)

*   **Error Recovery:** Improve error reporting and recovery in the `goyacc` parser.
*   **Syntax Refinements:** Address any awkward syntax identified during broader use.

This provides a roadmap. The immediate focus should be on completing the type inference refactoring, followed by starting the VM implementation.
