# SDL Declaration DSL Package Summary (`sdl/decl`)

**Purpose:**

This package implements the high-level Domain Specific Language (DSL) for defining and analyzing system performance models. It aims to provide a declarative interface, abstracting the direct manipulation of probabilistic outcomes defined in `sdl/core`. It defines the language structure (AST) and the evaluation mechanism.

**Evaluation Strategy: Operator Tree**

The DSL evaluation employs a two-stage strategy using an intermediate "Operator Tree":

1.  **Stage 1: AST Evaluation to Operator Tree (`Eval` function in `eval.go`)**
    *   The `Eval` function walks the Abstract Syntax Tree (AST) defined in `ast.go`.
    *   It uses the runtime `Env` (`env.go`) for variable scoping.
    *   Instead of immediately calculating combined probabilistic states, it builds an **Operator Tree** (`opnode.go`).
    *   **Leaves** (`LeafNode`) of this tree represent concrete values, holding a `VarState` (`state.go`) which encapsulates the dual-track `ValueOutcome` and `LatencyOutcome`. Leaf nodes are typically produced by evaluating literals or calling methods on Go components (from `sdl/components`).
    *   **Internal Nodes** (`SequenceNode`, `BinaryOpNode`, `ChoiceNode`, etc. - *partially defined*) represent operations (+, &&, sequence, if, distribute) and hold references to their child nodes.
    *   This stage focuses on correctly representing the *structure* of the computation.

2.  **Stage 2: Operator Tree Execution (Future "Tree Evaluator")**
    *   *(Not Yet Implemented)* A separate component (the "Tree Evaluator") will be responsible for processing the `OpNode` tree generated in Stage 1.
    *   It will walk the Operator Tree, recursively evaluating child nodes.
    *   At each internal node, it will perform the actual probabilistic combination of the resulting child `VarState`s, implementing the **V4 Dual-Track Model**:
        *   Combine `LatencyOutcome` tracks (typically using `core.And` with duration addition).
        *   Combine `ValueOutcome` tracks based on the operation (e.g., arithmetic, boolean logic, conditional splitting/merging).
    *   This stage will also apply complexity reduction (`core.TrimToSize` etc.) via the `reducers.go` logic (which needs alignment with `VarState`).
    *   The final result of the Tree Evaluator will be a single, combined `VarState`.

**Key Components & Files:**

*   **`ast.go`:** Defines Go structs representing the parsed DSL grammar (Components, Systems, Statements, Expressions).
*   **`opnode.go`:** Defines the interfaces and structs for the intermediate Operator Tree (`OpNode`, `LeafNode`, `SequenceNode`, `NilNode`).
*   **`eval.go`:** Implements the `Eval` function (Stage 1) which traverses the AST and builds the `OpNode` tree.
    *   *Currently Implemented:* Literals, Identifiers, Let Statements, Expression Statements, Blocks.
*   **`env.go`:** Runtime environment for managing identifier scopes (stores `OpNode`s for variables during Stage 1).
*   **`state.go`:** Defines `VarState` (dual-track outcomes) and helpers (`ZeroLatencyOutcome`, etc.). `VarState` is used within `LeafNode`s.
*   **`vm.go`:** Basic VM structure, holds internal functions registry and reducer registry (currently single-outcome focused).
*   **`analysis.go`:** Defines `AnalysisResultWrapper` and `CalculateAndStoreMetrics` for handling `analyze` block results. *(Needs update for `VarState` and Tree Evaluator integration)*.
*   **`reducers.go`:** Manages outcome combination and reduction. *(Currently operates on single `core.Outcomes`, needs significant update for `VarState` combination within the Tree Evaluator)*.
*   **`SYNTAX.md`, `GRAMMAR.ebnf`, `FUTURES.md`:** DSL design documents.

**Current Status:**

*   AST, environment, and `VarState` structures are defined.
*   The Operator Tree evaluation strategy is chosen. Basic `OpNode` types are defined.
*   Stage 1 `Eval` implementation has begun: Literals, identifiers, let bindings, expression statements, and blocks correctly build `LeafNode`, `SequenceNode`, or `NilNode` results and manage the environment.
*   **Missing:**
    *   `Eval` logic for most operations (binary ops, control flow, component calls, concurrency) to build their respective `OpNode`s.
    *   The entire Stage 2 "Tree Evaluator" component is missing.
    *   Integration with `analysis.go` and updates to `reducers.go` for `VarState` handling are needed.
