
**Phase 4: DSL Implementation Roadmap**

**Goal:** Create a functional DSL parser and interpreter capable of defining basic system components, composing their operations, and triggering analysis, translating DSL constructs into calls to the existing SDL Go library.

**Milestone 4.1: Core DSL Syntax & Parsing Foundation**

*   **Objective:** Establish the basic syntax structure and the ability to parse it into a usable Abstract Syntax Tree (AST).
*   **Tasks:**
    1.  **Finalize v1 Grammar:** Refine the pseudo-BNF grammar for core constructs (components, params, uses, operations (signatures only), system, instance, analyze). Ensure it's reasonably unambiguous for parsing.
    2.  **Choose Parsing Library:** Select the Go library (e.g., `participle`, `goyacc`, `text/scanner`). *Recommendation: `participle` often simplifies mapping grammar to Go structs.*
    3.  **Implement AST Nodes (`sdl/dsl/ast.go`):** Define Go structs for all grammar elements identified in Task 1 (e.g., `File`, `ComponentDecl`, `ParamDecl`, `SystemDecl`, `InstanceDecl`, `AnalysisDecl`, basic `Literal`, `Identifier`). Implement the `Node` interface and basic `String()` methods.
    4.  **Implement Parser (`sdl/dsl/parser.go`):** Write the parser using the chosen library to transform DSL text into the defined AST. Implement basic error reporting (line/column).
    5.  **Parser Unit Tests (`sdl/dsl/parser_test.go`):** Test parsing of simple, valid DSL files containing only declarations (components, systems, instances, params, uses, analysis targets). Test basic syntax error detection. *Do not test parsing of operation bodies yet.*

**Milestone 4.2: Basic Interpreter & Component Instantiation**

*   **Objective:** Create an interpreter capable of loading component definitions, instantiating a system, and resolving dependencies.
*   **Tasks:**
    1.  **Interpreter Structure (`sdl/dsl/interpreter.go`):** Define the main `Interpreter` struct. Implement basic environment/symbol table handling (nested scopes for global, system, component instances).
    2.  **Component Registry:** Implement a mechanism (e.g., a map) within the interpreter to register both built-in Go component types (like `Disk`, `Cache`) and user-defined component types parsed from the DSL (`ComponentDecl` AST nodes).
    3.  **Type System (Basic):** Define how DSL types (`string`, `int`, `duration`, component names) map to Go types. Implement parsing/validation for `duration` literals.
    4.  **`eval(SystemDecl)`:** Implement logic to walk a `SystemDecl` AST node.
    5.  **`eval(InstanceDecl)`:** Implement logic to:
        *   Look up the `ComponentType` in the registry.
        *   If it's a Go type, call its `Init` method (using reflection or predefined factories).
        *   If it's a DSL-defined type, prepare its definition for later use.
        *   Apply `ParamOverrides` (evaluate literal values, resolve instance references using the environment). Store the instantiated component (Go object or DSL definition reference) in the system's environment. Handle dependency resolution order if necessary.
    6.  **Interpreter Tests:** Test instantiating a simple system with built-in Go components (`Disk`, `Cache`) and potentially a simple DSL-defined component (without complex operations yet), verifying parameters are set correctly.

**Milestone 4.3: Interpreting Operation Calls & Sequential Logic**

*   **Objective:** Enable the interpreter to execute simple operation bodies consisting of sequential calls to other component operations.
*   **Tasks:**
    1.  **Refine AST:** Add nodes for `OperationDef` body (`[]Stmt`), `AssignmentStmt`, `CallExpr`, `MemberAccessExpr`, `ReturnStmt`. Update parser.
    2.  **`eval(OperationDef)`:** Set up operation scope (environment) with parameters. Iterate through body statements using `eval`. Handle `ReturnStmt`.
    3.  **`eval(AssignmentStmt)`:** Evaluate RHS expression, store result (which *will be* an `*Outcomes[V]` object from a Go call) in the current environment.
    4.  **`eval(CallExpr)` & `eval(MemberAccessExpr)`:** Implement logic to:
        *   Resolve the receiver (e.g., instance name from environment) and method name.
        *   Evaluate arguments.
        *   Look up the corresponding Go method on the receiver object.
        *   **Invoke Go Method:** Call the method using reflection or type switches/interfaces. Handle argument type conversions if necessary (e.g., DSL `int` to Go `uint`). Receive the `*Outcomes[V]` result.
    5.  **Implicit `And`:** Implement the logic for sequential statements. Maintain the "current accumulated outcome" (`currentOutcome`). When evaluating the next statement (`b = Op2()`), get `op2_Outcomes` and compute `currentOutcome = sdl.And(currentOutcome, op2_Outcomes, selectReducer(...))`. The final statement's result is the block's result. Requires type switching to select the correct `And` reducer.
    6.  **Interpreter Tests:** Test executing simple operations involving sequences of calls to built-in components (e.g., `disk.Read(); cache.Write()`). Verify the final `Outcomes` object looks plausible (e.g., check `Len()`).

**Milestone 4.4: Implementing Conditionals (`if`)**

*   **Objective:** Add support for conditional logic within operation bodies based on outcome properties.
*   **Tasks:**
    1.  **Refine AST:** Add `IfStmt` node. Update parser.
    2.  **`eval(IfStmt)`:** Implement the complex logic described previously:
        *   Evaluate `Condition` (e.g., `var.Success`). Interpreter needs to know this accesses a field within the `Outcomes` values.
        *   Get the `*Outcomes[V]` for `var`.
        *   Use type switching on `V` to select the correct predicate (`func(v V) bool { return v.Success }`).
        *   Call Go `Split` method on the outcomes.
        *   Recursively call `eval` for the `Then` block, passing the `successOutcomes` as context.
        *   Recursively call `eval` for the `Else` block, passing the `failureOutcomes` as context.
        *   Combine results using Go `Append`. Add latency of condition check (if applicable/modelled).
    3.  **Interpreter Tests:** Test operations using `if x.Success { ... } else { ... }`. Verify that the final `Outcomes` distribution correctly reflects the probabilistic combination of both branches. Use the Bitly `Redirect` example as a key test case.

**Milestone 4.5: Analysis Execution & Basic Output**

*   **Objective:** Allow the DSL to trigger an analysis and get basic results.
*   **Tasks:**
    1.  **`eval(AnalysisDecl)`:** Implement logic to:
        *   Evaluate the `Target` expression (an operation call) to get the final `*Outcomes[V]` for the specified scenario entry point.
        *   Store this result associated with `analysisName`.
    2.  **Interpreter Output:** Add functionality for the interpreter to run all `analyze` blocks in a `system` and print basic summary metrics for each result (using the Go `sdl/metrics` functions). E.g., print Availability, MeanLatency, P50, P99.
    3.  **Basic `expect` (Optional Stretch):** Add basic `expect Expr` syntax to `AnalysisDecl` (e.g., `expect Target.P99 < 0.100`). The interpreter evaluates the check after getting the `Target` outcomes and reports pass/fail. Requires adding latency/availability accessors to the DSL expression language (e.g., `Target.P99`, `Target.Availability`).
    4.  **Interpreter Tests:** Test a full `system` definition with `analyze` blocks. Verify that the analysis runs and produces plausible metrics output. Test basic `expect` clauses.

**Milestone 4.6: Integration & Refinement**

*   **Objective:** Clean up implementation, add more tests, document the v1 DSL syntax and interpreter usage.
*   **Tasks:**
    *   **Parameter Handling:** Ensure all component parameters (including rates for Queue/Pool) can be configured correctly in the `system` block and used by the interpreter.
    *   **Error Handling:** Improve error reporting throughout the parser and interpreter (syntax errors, type errors, runtime errors like unknown components/operations).
    *   **Reduction Integration:** Ensure implicit reduction using `Trim...` functions based on `options` hints is reliably applied by the interpreter after composition steps. Test this.
    *   **More Component Tests:** Add tests using DSL definitions for more of the components (`LSM`, `Bitmap`, `NetworkLink`, `Queue`, `Pool`, `Batcher`).
    *   **Documentation:** Write initial documentation for the v1 DSL syntax and how to run the interpreter.

**Future Phases (Post v1 DSL):**

*   **Phase 4 Continued:** Visualization hooks, component libraries, more examples.
*   **Phase 5:** Optimization, advanced validation, looping constructs, `Map` syntax, potential `Bind/Let` for value access, improved error handling, richer type system, modules/imports.
*   **Phase X:** Consider DES integration or more advanced analytical models if limitations become too severe.


