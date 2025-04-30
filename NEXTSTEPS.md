# SDL Project - Next Steps

**Overall Goal:** Complete the initial functional version of the DSL VM/VM and the DSL Parser, enabling users to define, simulate, and analyze simple systems declaratively.

**Phase 4: DSL Implementation (Continued)**

**A. Complete Core VM Functionality:**

1.  **`ParallelExpr` Evaluator:**
    *   Implement `evalParallelExpr`.
    *   Design and implement a parallel reducer registry (similar to `sequentialReducers`).
    *   Register default parallel reducers (e.g., Success=AND, Latency=MAX).
    *   Apply reduction logic.
    *   Add tests.

2.  **`FanoutExpr` Evaluator:**
    *   Implement `evalFanoutExpr`.
    *   Evaluate the count distribution (`*Outcomes[int]`).
    *   Iterate through count buckets.
    *   For each count `N` and probability `P`:
        *   Evaluate the `OpExpr` `N` times (likely using `RepeatExpr` with Parallel mode).
        *   Handle success probability (approx `singleSuccess^N`?). Needs careful modeling.
        *   Scale the resulting outcome by `P`.
    *   Combine (Append) results from all count buckets.
    *   Apply reduction.
    *   Add tests. *(Note: This is complex and may require approximations)*.

3.  **`FilterExpr` Evaluator:**
    *   Implement `evalFilterExpr`.
    *   Evaluate the `Input` expression.
    *   Evaluate filter parameters (MinLatency, MaxLatency, BySuccess). Parameters must be deterministic.
    *   Implement or refine `core.Filter` function if needed (current one might be basic).
    *   Call `core.Filter` (or equivalent logic) using type assertion on the input outcome.
    *   Push the filtered outcome.
    *   Add tests.

4.  **`SwitchExpr` Evaluator:**
    *   Implement `evalSwitchExpr`.
    *   Evaluate the `Input` expression.
    *   Iterate through `Cases`. For each case:
        *   Evaluate `case.Condition`. Must yield a deterministic value of a comparable type to the Input's inner value.
        *   Compare the Input outcome's value(s) against the case condition. This is complex with distributions - maybe only allow deterministic Input for Switch? Or match based on buckets? *Decision needed.*
        *   If match, evaluate `case.Body` and use that as the result (signal exit like `ReturnStmt`).
    *   Handle default case.
    *   Add tests. *(Note: Semantics with probabilistic input need careful design)*.

5.  **Refine Top-Level Driver (`driver.go`):**
    *   Handle `ast.File` input (iterate through declarations).
    *   Process `ComponentDecl` - store definitions for instantiation.
    *   Refine `InstanceDecl` evaluation: Handle non-literal parameters (e.g., references), resolve `uses` dependencies.
    *   Improve component constructor registration and parameter handling.
    *   Implement parsing/handling of `OptionsDecl`.
    *   Add support for `expect` clauses in `AnalyzeDecl` (requires extending `evalMemberAccessExpr` for metrics like `.P99`, `.Availability`).

6.  **Refine `evalMemberAccessExpr`:**
    *   Add support for accessing metrics directly (e.g., `myResult.P99`, `myResult.Availability`). This would involve calculating the metric on the receiver outcome and returning a new deterministic `Outcomes[float64]` or `Outcomes[Duration]`.

7.  **Error Handling & Stack Management:**
    *   Improve error messages with position info (requires passing `NodeInfo` through calls).
    *   Ensure stack is consistently managed, especially on error paths.
    *   Add more robust type checking during combinations and calls.

8.  **Refinements & Cleanup:**
    *   Remove the `.And` field from `core.Outcomes` and update callers.
    *   Optimize Reducer Registry lookup (e.g., using type IDs instead of strings).
    *   Add more internal function implementations needed by `decl` components (math helpers, etc.).


**B. Implement DSL Parser (Milestone 4.1 from Roadmap):**

1.  **Finalize v1 Grammar:** Solidify the syntax based on implemented AST nodes.
2.  **Implement Parser (`sdl/dsl/parser.go`):** Use `participle` (or chosen library) to parse DSL text into the defined AST (`ast.go`).
3.  **Parser Unit Tests:** Test parsing of various declaration structures. Test error reporting.

**C. Integration & Documentation:**

1.  **Integrate Parser & Driver:** Modify `RunDSL` (or add a new entry point) to accept DSL file input, parse it, and then execute the resulting AST.
2.  **Examples:** Rewrite one of the `sdl/examples` (e.g., `gpucaller`) using the DSL.
3.  **Documentation:** Write initial user documentation for the v1 DSL syntax and how to run analyses.


**Suggested Next Steps Order:**

1.  **`evalMemberAccessExpr` (for metrics):** Implement access for `.Availability`, `.MeanLatency`, `.P50`, `.P99` to enable `expect` clauses later.
2.  **`ParallelExpr`:** Implement parallel composition (relatively contained, builds on registry pattern).
3.  **Refine Driver:** Improve component instantiation, parameter handling in `RunDSL`.
4.  **Parser:** Implement the DSL parser (`parser.go`).
5.  **Integration:** Connect parser to driver.
6.  **Remaining Evaluators:** Implement `FanoutExpr`, `FilterExpr`, `SwitchExpr` (potentially deferring complex ones).
7.  **Error Handling/Refinement:** Continuous improvement.
8.  **Remove `.And` field:** Cleanup `core.Outcomes`.
