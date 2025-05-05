# SDL Project: Next Steps

This document outlines the next steps for developing the SDL project, focusing primarily on completing the DSL evaluator based on the "Operator Tree" strategy using the `ComponentRuntime` interface.

**Recap of Current State:**

*   `core` and `components` Go libraries are functional.
*   `decl` package defines the DSL AST, `VarState`, `OpNode` types, and the `ComponentRuntime` interface (implemented by `NativeComponentAdapter` and `ComponentInstance`).
*   Stage 1 `Eval` (`decl/eval.go`) correctly processes component definitions, instantiates both native and DSL components (storing `ComponentRuntime` in `Env`), handles basic expressions/statements (Literals, Let, Identifiers, Blocks, Binary Ops), and basic control flow (`if`), building the corresponding `OpNode` tree structure.
*   Stage 2 "Tree Evaluator" is not implemented.
*   `VarState` combination logic (V4 model) and reduction need implementation within the Tree Evaluator.
*   A temporary workaround exists in `evalInstanceDecl` to evaluate native component parameters immediately.

**Prioritized Next Steps:**

1.  **Implement `evalCallExpr` (Stage 1):**
    *   **(Next)** This is the most crucial next step for making the DSL functional.
    *   Evaluate the `Function` part of the `CallExpr`. Expect it to resolve to:
        *   A `MemberAccessExpr` (e.g., `instanceName.methodName`).
        *   Potentially a simple `IdentifierExpr` (for standalone functions - less common initially).
    *   If `MemberAccessExpr`:
        *   Evaluate the `Receiver` (e.g., `instanceName`) -> expect IdentifierExpr.
        *   Look up the identifier in the `Env`.
        *   Assert the result is a `ComponentRuntime`.
        *   Get the `Member` name (`methodName`).
    *   Evaluate Arguments (`Args`) -> `[]OpNode`.
    *   **Call `InvokeMethod`:** Use the `ComponentRuntime.InvokeMethod(methodName, argOpNodes, vm, env)` method. This call leverages the polymorphism:
        *   For `NativeComponentAdapter`, `InvokeMethod` uses reflection to call the Go method (currently requires temp workaround for args).
        *   For `ComponentInstance`, `InvokeMethod` looks up the DSL method definition and prepares to evaluate its body (needs parameter/dependency binding logic implemented).
    *   Return the `OpNode` result from `InvokeMethod`.

2.  **Complete Remaining Stage 1 `Eval` Functions:**
    *   **Unary Operations:** Implement `evalUnaryExpr` -> `UnaryOpNode`.
    *   **Control Flow (`distribute`):** Implement `evalDistributeStmt`/`evalDistributeExpr` -> `ChoiceNode`. Assume deterministic probabilities initially.
    *   **Explicit Latency:** Implement `evalDelayStmt` -> `DelayOpNode` or similar.
    *   **Concurrency Placeholders:** Implement `evalGoStmt`/`evalWaitStmt` -> `GoOpNode`/`WaitOpNode`.

3.  **Develop Stage 2 "Tree Evaluator":**
    *   Create the core walking function/struct.
    *   Implement `combineVarStates` (V4 dual-track logic).
    *   Implement logic for walking/evaluating each `OpNode` type (`LeafNode`, `BinaryOpNode`, `SequenceNode`, `IfChoiceNode`, `ChoiceNode`, etc.), calling `combineVarStates` as needed.
    *   **Remove Temporary Workarounds:** Update `NativeComponentAdapter.InvokeMethod` (for args) and `evalInstanceDecl` (for native params) to call the Tree Evaluator instead of extracting simple values.
    *   Integrate Reducer Logic: Update `reducers.go` for `VarState`. Call reduction functions within the Tree Evaluator.

4.  **Implement `ComponentInstance.InvokeMethod` Fully:**
    *   Add logic for binding arguments and dependencies into the `methodEnv`.
    *   Handle `return` statements (may require modification to `Eval`/`evalBlockStmt` or specific handling in `InvokeMethod`).

5.  **Integrate Analysis (`decl/analysis.go`):**
    *   Connect `analyze` blocks to run the Tree Evaluator.
    *   Update `AnalysisResultWrapper`/`CalculateAndStoreMetrics` for `VarState`.
    *   Implement `ExpectBlock` evaluation.

6.  **Implement Concurrency (within Tree Evaluator):**
    *   Handle `GoOpNode`/`WaitOpNode` with actual goroutines and synchronization, combining `VarState`s correctly.

**Longer Term:**

*   Implement a DSL Parser.
*   Add more components and refine existing ones.
*   Enhance error reporting.
*   Develop analysis/visualization tools.
