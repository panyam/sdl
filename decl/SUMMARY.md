# SDL DSL Package Summary (`sdl/dsl`)

**Purpose:**

This package implements the a declarative of specifying components that maps closer to what a DSL would look like.   Execution is decoupled from the declarative elements (ie the AST elements) by allowing various evaluators that can take a declarative component and run their various method/operators.   This allows us to bring up other runtimes in isolation and not get bogged down in those details (for optimization etc).  This breakdown would evaluators/interpreters/runners to call the primiteves in`sdl/core` and `sdl/decl/components` (declarative versions of primitive components in `sdl/components`) and produces analysis results. It aims to provide a high-level, declarative interface, abstracting direct manipulation of probabilistic outcomes (eg map, trim etc just the way a "garbage collector" would).

**Target Execution Model:**

*   **Implicit Outcomes:** The DSL syntax deals with functional types (`bool`, `enum`, `int`), not `Outcomes` directly.
*   **Dual Track:** Any runtime/evaluator should tracks two distributions per variable/expression result via `VarState { ValueOutcome any /* *Outcomes[V] */, LatencyOutcome any /* *Outcomes[Duration] */ }`.
*   **Discrete Values:** `V` in the value track is restricted to registered discrete types (bool, enums).
*   **Explicit Delay:** Latency is added *usually* by leaf method or explicit `delay` statements.
*   **Probabilistic Branching:** `distribute` statement (future) handles probabilistic control flow. `if` splits based on boolean value outcomes (via `OpCallFunc HandlerIf`).
*   **Flow Only Evaluators**: We are not building a general purpose language.   The goal of the DSL is to model systems
  for performance reasons.  Most complicated behavioral logic is expected to be at the leaf components which will be
  written natively (or via plugins at a later date) so this wont be part of the DSL (beyond calling out to
  foreign-functions).

**Hybrid Bytecode Execution Strategy:**

*   (Future) **Bytecode:** Simple expressions and statements are compiled into bytecode instructions (`bytecode.go`).
*   (Future) **VM Go Helpers:** Complex probabilistic control flow (`if`, `distribute`) and potentially other method (`AndSeq`, `Delay`) are handled by dedicated Go functions (`vm.vm_handle_*`) within the VM, triggered by specific `OpCallFunc` bytecode instructions. These Go helpers orchestrate context splitting, recursive calls (potentially back to the bytecode loop), and result combination using `sdl/core` functions.

**Key Concepts & Components:**

1.  **Abstract Syntax Tree (AST) (`ast.go`):**
    *   Defines Go structs (eg Expr, Node, Stmt, IfStmt etc) representing the parsed DSL (aligned with our grammar in GRAMMAR.md).
    *   Serves as the input to the runtime (EvaluatorVM or Compi
    *   How these differ from the native ones is these define how a model should behave instead of worryign when/how a
        TrimXYZ should be applied.  eg in the BTree implementation each method is responsible for internaly trimming
        Outcomes to prevent size explosion - this feels like garbage collection and should be taken care off by the
        runtime.

2  **Runtime (`env.go`, `state.go`, `stack.go`):**
    *   Define primitives to aid in the "runtime" execution of an expression tree.
    *   Serves as the input to the Compiler.
    *   `state.go` - Stores `*VarState` which is used for tracking variable states/durations.
    *   `env.go`: Handles lexical scoping (though currently only globals are implemented in VM).

3.  **Execution (`exec.go`):**
    *   Initial execution model is through a visitor that executes an Expression by walking down the tree.
    *   In the future we will consdier a ByteCode model.
    *   **Dual Track State:** `VarState` struct used as state vars, globals and locals (via env)
    *   Executor is responsible for invoking "compaction" of outcomes at the right time.
    *   Executor is also responsible for calling the right "mapper" for mapping Outcome[X] -> Outcome[Y] based on types
        at runtime.
    *   **Reduction:** Uses `reduceStateIfNeeded` (currently only checks latency track size and applies reduction via configured `durationReducer`).

9.  **Driver (`driver.go`):**
    *   Initializes VM/Executor, registers built-in components/functions.
    *   Instantiates system components based on AST.
    *   Executes `analyze` targets by calling VM.
    *   Calculates metrics using `core` functions and returns results in `AnalysisResultWrapper`.

**Current Status:**

*   AST (`ast.go`) reasonably defined for our Grammar.
*   VM (`vm.go`) has core loop (`ExecuteChunk`), stack (`*VarState`), globals. Executes `OpPushConst`, `OpDefineGlobal`, `OpGetGlobal`, `OpSetGlobal`, `OpPop`, `OpJumpIfFalsePop`, `OpJump`, `OpReturn`, `OpCallFunc`.
*   VM includes `vm_handle_*` stubs/basic implementations for `Add`, `AndSeq`, `Delay`, `If`, `Distribute`. `vm_handle_add` has basic deterministic logic.
*   Driver (`driver.go`) needs update to use Compiler + `ExecuteChunk`.
*   **Key Missing Pieces:** Full implementation of `vm_handle_*` functions for complex logic (especially `If`, `Distribute`, probabilistic binary ops), local variables/scoping, method/function calls (`OpCall`), parser. `combineOutcomesAndReduce` value propagation is simplified.

**Next Steps:**

*   Implement logic within `vm_handle_*` handlers (Arithmetic, Boolean, `Delay`, `AndSeq`).
*   Implement `core.SplitLatenciesBasedOnValue` and integrate into `vm_handle_if`.
*   Implement `vm_handle_distribute`.
*   Implement locals (`OpGetLocal`, `OpSetLocal`) and call frame management (`OpCall`, `OpReturn`).
*   Update Driver to use compiler + `ExecuteChunk`.
*   Implement Parser.
