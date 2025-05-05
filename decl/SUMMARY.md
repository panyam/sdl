# SDL Declaration DSL Package Summary (`sdl/decl`)

**Purpose:**

This package implements the high-level Domain Specific Language (DSL) for defining and analyzing system performance models. It aims to provide a declarative interface, abstracting the direct manipulation of probabilistic outcomes defined in `sdl/core`. It defines the language structure (AST) and the evaluation mechanism.

**Evaluation Strategy: Operator Tree**

The DSL evaluation employs a two-stage strategy using an intermediate "Operator Tree":

1.  **Stage 1: AST Evaluation to Operator Tree (`Eval` function in `eval.go`)**
    *   The `Eval` function walks the Abstract Syntax Tree (AST) defined in `ast.go`.
    *   It uses the runtime `Env` (`env.go`) for variable scoping.
    *   It builds an **Operator Tree** (`opnode.go`) representing the computation symbolically.
    *   **Leaves** (`LeafNode`) hold a `VarState` (`state.go`) (Value + Latency outcomes), created from literals or component calls.
    *   **Internal Nodes** (`SequenceNode`, `BinaryOpNode`, `IfChoiceNode`, etc.) represent operations and structure.
    *   Component *definitions* (`ComponentDecl`) are processed to populate a `ComponentDefRegistry` in the `VM`.
    *   Component *instantiations* (`InstanceDecl`) use the definition registry and potentially a native Go constructor registry (`VM.ComponentRegistry`) to create runtime instances.
        *   Native Go components (from `sdl/components`) are wrapped in a `NativeComponentAdapter`.
        *   DSL-defined components are represented by `ComponentInstance`.
        *   Both implement the `ComponentRuntime` interface (`instance.go`) and are stored in the `Env`.
    *   Focuses on building the correct symbolic tree and setting up the runtime environment with component instances.

2.  **Stage 2: Operator Tree Execution (Future "Tree Evaluator")**
    *   *(Not Yet Implemented)* A separate component will process the `OpNode` tree.
    *   It will walk the tree, performing actual `VarState` combinations (using the V4 dual-track model) and reduction (`reducers.go`).
    *   It will interact with `ComponentRuntime` instances (via `InvokeMethod`) to handle method calls.
    *   The final result will be a single, combined `VarState`.

**Key Components & Files:**

*   **`ast.go`:** Defines the parsed DSL grammar structs.
*   **`opnode.go`:** Defines the Operator Tree nodes (`OpNode`, `LeafNode`, `BinaryOpNode`, `IfChoiceNode`, etc.).
*   **`instance.go`:** Defines the `ComponentRuntime` interface and the `ComponentInstance` struct (for DSL components). Includes the `NativeComponentAdapter` for wrapping Go components.
*   **`eval.go`:** Implements the Stage 1 `Eval` function (builds `OpNode` tree).
    *   *Currently Implemented:* Literals, Identifiers, Let Stmts, Expr Stmts, Blocks, Binary Ops, If Stmts, Component Defs, System Defs, Instance Decls (Native & DSL).
*   **`env.go`:** Runtime environment (stores `OpNode`s for variables, `ComponentRuntime`s for instances).
*   **`state.go`:** Defines `VarState` (dual-track outcomes).
*   **`vm.go`:** VM structure, holds `ComponentDefRegistry`, `ComponentRegistry` (for native constructors), internal funcs, reducers.
*   **`analysis.go`:** Defines `AnalysisResultWrapper`. *(Needs update for `VarState` and Tree Evaluator integration)*.
*   **`reducers.go`:** Manages outcome combination/reduction. *(Needs update for `VarState`)*.
*   **`SYNTAX.md`, `GRAMMAR.ebnf`, `FUTURES.md`, `EXAMPLES.md`:** Design and example documents.

**Current Status:**

*   AST, environment, `VarState`, `ComponentRuntime`, `ComponentInstance`, `NativeComponentAdapter`, and basic `OpNode` types are defined.
*   The Operator Tree evaluation strategy is established.
*   Stage 1 `Eval` implementation handles key structural elements: definitions (Component), setup (System, Instance - native & DSL, including dependency injection structure), and basic expressions/statements (Literals, Let, Identifiers, Blocks, Binary Ops, If).
*   **Missing:**
    *   `Eval` logic for method calls (`evalCallExpr`), unary ops, `distribute`, `delay`, concurrency (`go`, `wait`), etc.
    *   The entire Stage 2 "Tree Evaluator" component.
    *   Integration with `analysis.go` and updates to `reducers.go` for `VarState` / Tree Evaluator.
    *   Temporary workarounds for evaluating native component parameters need replacement.
