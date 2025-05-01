# SDL Project - Next Steps

**Overall Goal:** Implement the core functionality of the DSL based on **Model V4 (Implicit Outcomes + Discrete Values + Explicit Delay)**, including the VM execution logic and a basic parser, to enable declarative definition and analysis of simple systems.

**Phase 1: VM Refactoring & Core Model V4 Implementation**

*Goal: Update the VM's core evaluation logic to support the dual-track (value + latency) implicit model.*

1.  **Modify Environment:** Update `dsl.Environment` to potentially store two entries per variable (e.g., `varName` -> `Outcomes[V]`, `varName_latency` -> `Outcomes[Duration]`) or use a wrapper struct.
2.  **Refactor `vm.Eval` Return:** Standardize how `Eval` returns results. It should probably return *both* the value outcome and the latency outcome.
3.  **Refactor `evalAndExpr` (Sequential Composition):**
    *   Modify to receive/return both value and latency outcomes.
    *   Implement latency combination: `finalLat = And(latA, latB, AddDurations)`.
    *   Implement value combination (Pragmatic "Take Last"): `finalVal = valB`. Associate `finalLat` with `finalVal`. Ensure probability/weights are handled correctly, potentially needing adjustments in `valB` based on combined success derived from `valA` and `valB`, or relying on `finalLat` carrying the combined probability implicitly. *Requires careful design and testing.*
4.  **Refactor `evalIfStmt`:**
    *   Modify to handle the dual tracks.
    *   Split *both* value and latency distributions based on the condition value.
    *   Pass split contexts (`valT`, `latT`) and (`valF`, `latF`) to recursive `evalBlockStmt` calls.
    *   Combine results using `Append` for *both* value and latency tracks.
5.  **Refactor `evalBlockStmt`:**
    *   Modify to accept an initial context containing *both* value and latency outcomes.
    *   Update sequential composition logic within the block to use the refactored `evalAndExpr` principles (combine latency, propagate last value).
    *   Ensure it correctly returns the final value and latency outcomes for the block.
6.  **Refactor `evalCallExpr` & `evalInternalCallExpr`:**
    *   Ensure they correctly handle returning *both* value and latency outcomes from the called functions/methods.
    *   Update argument handling if necessary (though Model V4 still expects deterministic args for Go methods).
7.  **Implement `evalDelay`:**
    *   Evaluate the duration expression (`Outcomes[Duration]`).
    *   Combine this with the *current* implicit latency track using `And`.
    *   Return an identity/void value outcome and the *updated* latency outcome.
8.  **Implement `evalDistribute`:**
    *   Evaluate probabilities.
    *   For each branch:
        *   Calculate effective probability `P_N`.
        *   Split the incoming context (`val_in`, `lat_in`) by `P_N`.
        *   Evaluate the branch block `blockN` with the split context. Get (`val_N`, `lat_N`).
        *   *(Correction: No need to scale results by P_N here, Append handles weights)*.
    *   Combine results from all branches using `Append` for both value and latency tracks.
    *   Handle `default` and `totalExpr`.
9.  **Unit Tests:** Add extensive unit tests for the refactored evaluators, focusing on correct dual-track synchronization and combination.

**Phase 2: DSL Parser Implementation**

*Goal: Create a parser that translates DSL text (following Model V4 syntax) into the AST defined in `dsl/ast.go`.*

1.  **Finalize v1 Grammar:** Formally define the grammar for Model V4, including `component`, `system`, `param`, `uses`, `operation`, `instance`, `analyze`, `distribute`, `delay`, `if`, assignments, returns, expressions, etc.
2.  **Choose Parser Library:** Confirm use of `participle` or select alternative.
3.  **Implement Parser (`sdl/dsl/parser.go`):** Write parser code to generate the AST.
4.  **Parser Unit Tests:** Test parsing of various valid DSL constructs and verify correct error reporting for invalid syntax.

**Phase 3: Integration, Remaining Evaluators & Refinements**

*Goal: Connect parser to VM, implement remaining features, and refine.*

1.  **Integrate Parser & Driver:** Modify `RunDSL` (or create a new entry point) to accept DSL file paths/strings, parse them using the new parser, and execute the resulting AST with the refactored VM.
2.  **Implement Remaining Evaluators:**
    *   `evalParallelExpr` (Requires parallel reducer design).
    *   `evalFanoutExpr` (Complex, may need approximations).
    *   `evalFilterExpr` (Requires core filtering logic).
    *   `evalSwitchExpr` (Requires clear semantics for discrete matching).
3.  **Refine Driver (`RunDSL`):**
    *   Improve component instantiation (handle non-literal params, `uses` dependencies).
    *   Implement `options` handling.
    *   Add `expect` clause parsing and evaluation within `AnalyzeDecl` (requires metric access).
4.  **Implement Metric Access:** Refine `evalMemberAccessExpr` to support accessing calculated metrics like `.P99`, `.Availability` from result variables, returning deterministic `Outcomes[float64]` etc.
5.  **Error Handling:** Improve error reporting throughout the VM and driver, including positional information from the parser.
6.  **Cleanup:** Remove the `.And` field from `core.Outcomes` and update core callers (VM now handles reducer selection).

**Phase 4: Documentation & Examples**

*Goal: Make the DSL usable by others.*

1.  **Write DSL Documentation:** Create documentation explaining the Model V4 syntax, semantics, built-in components, and how to write and analyze systems.
2.  **Rewrite Examples:** Convert at least one example from `sdl/examples` (e.g., `gpucaller` or `bitly`) to use the new DSL.
3.  **Add DSL Tests:** Create test files (`*.sdl`) and use `RunDSL` within Go tests to verify end-to-end DSL execution and analysis results.

**Suggested Order:**

1.  **Phase 1 (VM Refactoring & Core Model V4):** Focus heavily here first, as it's the most fundamental change. Implement `delay` and `distribute` early. Ensure `And`, `If`, `Block`, `Call` handle dual tracks correctly. *(This is the highest priority)*.
2.  **Phase 2 (Parser):** Implement the parser once the target AST and basic V4 semantics are stable.
3.  **Phase 3 (Integration & Remaining Evaluators):** Connect parser/VM. Implement metric access needed for `expect`. Implement `ParallelExpr`. Refine the driver. Defer `FanoutExpr`, `FilterExpr`, `SwitchExpr` if needed to reach a usable v1 faster.
4.  **Phase 4 (Docs & Examples):** Document the usable subset and rewrite an example.

This provides a roadmap focused on delivering the core Model V4 DSL experience.
