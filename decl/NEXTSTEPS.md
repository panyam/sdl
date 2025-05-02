# SDL Project: Next Steps

This document outlines the next steps for developing the SDL project, focusing primarily on completing the DSL evaluator based on the "Operator Tree" strategy.

**Recap of Current State:**

*   `core` and `components` Go libraries are functional.
*   `decl` package defines the DSL AST and the Operator Tree evaluation approach.
*   Stage 1 `Eval` (`decl/eval.go`) can handle Literals, Let, Identifiers, Blocks, and Expr Stmts, building basic `OpNode` trees (`LeafNode`, `SequenceNode`, `NilNode`).
*   Stage 2 "Tree Evaluator" is not implemented.
*   `VarState` combination logic (V4 model) and reduction need implementation within the Tree Evaluator.

**Prioritized Next Steps:**

1.  **Complete Stage 1 `Eval` (Build Full Operator Tree):**
    *   **(Next)** **Binary Operations:** Implement `evalBinaryExpr` to return `BinaryOpNode`. Handle arithmetic (`+`, `-`, `*`, `/`, `%`), boolean (`&&`, `||`), and comparison (`==`, `!=`, `<`, `<=`, `>`, `>=`) operators. Requires defining `BinaryOpNode` struct.
    *   **Unary Operations:** Implement `evalUnaryExpr` for `!` and `-`. Requires `UnaryOpNode`.
    *   **Control Flow (`if`):**
        *   Implement `evalIfStmt` to return an `IfChoiceNode(ConditionTree, ThenTree, ElseTree)`. This pushes the condition evaluation complexity to the Tree Evaluator, aligning with the Operator Tree philosophy for now. Requires `IfChoiceNode`. *(Deferring probabilistic condition evaluation within Eval)*.
    *   **Control Flow (`distribute`):**
        *   Implement `evalDistributeStmt`/`evalDistributeExpr` to return a `ChoiceNode` representing the probabilistic branches. Assume probabilities evaluate deterministically for now. Requires `ChoiceNode` and `WeightedBranch`.
    *   **Component Method Calls:**
        *   Implement `evalCallExpr` to handle calls like `instance.Method(args)`.
        *   Requires looking up the component instance (from `components`) in the `Env`.
        *   Requires calling the Go method via reflection or a registry.
        *   Requires converting the returned `*core.Outcomes[V]` to a `VarState` using the `outcomeToVarState` helper (which needs full implementation).
        *   Returns a `LeafNode` containing the resulting `VarState`.
    *   **Explicit Latency:** Implement `evalDelayStmt` to likely return a specific `DelayOpNode(DurationTree)` or modify the context? TBD, but needs representation.
    *   **Concurrency Placeholders:** Implement `evalGoStmt` and `evalWaitStmt` to return placeholder `GoOpNode` / `WaitOpNode` or similar, deferring actual concurrency logic until the Tree Evaluator.

2.  **Develop Stage 2 "Tree Evaluator":**
    *   Create a new struct/set of functions responsible for walking the `OpNode` tree.
    *   Implement `combineVarStates` function according to the **V4 Dual-Track Model**.
    *   Implement walking logic for each `OpNode` type (`LeafNode`, `BinaryOpNode`, `SequenceNode`, `ChoiceNode`, `IfChoiceNode`, etc.).
        *   `LeafNode`: Returns its `VarState`.
        *   `BinaryOpNode`: Recursively evaluates children, gets `VarState`s, applies `combineVarStates` and the specific value track operation.
        *   `SequenceNode`: Evaluates steps sequentially, using `combineVarStates` at each step.
        *   `ChoiceNode`/`IfChoiceNode`: Evaluates branches, splits/scales context `VarState`s, merges results using `appendVarStates` helper (needs implementation). Requires evaluating condition tree for `IfChoiceNode`.
    *   Integrate Reducer Logic: Update `reducers.go` to provide functions that operate on `VarState`s (or extract tracks). Call these reduction functions appropriately within the Tree Evaluator (e.g., after combinations in sequences or merges in choices).
    *   Error Handling: Implement robust error handling during tree evaluation.

3.  **Integrate Analysis (`decl/analysis.go`):**
    *   Modify `evalAnalyzeDecl` (or have the Tree Evaluator handle it) to run the Tree Evaluator on the target expression's `OpNode` tree.
    *   Update `AnalysisResultWrapper` and `CalculateAndStoreMetrics` to accept the final `VarState` from the Tree Evaluator and extract the necessary `*core.Outcomes` for metric calculation.
    *   Implement evaluation of `ExpectBlock`.

4.  **Implement Concurrency (within Tree Evaluator):**
    *   Refine the `GoOpNode`/`WaitOpNode` handling.
    *   The Tree Evaluator needs to manage goroutines launched by `go` and synchronization logic for `wait`, combining the resulting `VarState`s appropriately considering parallel execution time.

5.  **Complete Remaining AST Nodes:** Implement `Eval` and Tree Evaluator logic for any remaining nodes (`ReturnStmt`, `LogStmt`, etc.).

**Longer Term:**

*   Implement a DSL Parser (to replace manual AST creation in tests).
*   Add more components to the `components` library.
*   Refine analytical models in components.
*   Enhance error reporting with source code positions.
*   Develop more sophisticated analysis and visualization tools.
