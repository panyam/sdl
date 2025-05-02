## Options for Native Components and Methods

For the DSL to be practical, we absolutely need a way to integrate pre-built, high-performance, or complex components written directly in Go ("native" components) without forcing users to redefine their logic within the DSL.

**Option 1: Convention over Configuration (Registry Determines Nativity)**

*   **How it Works:**
    1.  **Native Component Registration:** During VM initialization (or via some setup phase), we register native Go component constructors with the VM using `vm.RegisterComponent("Disk", NewDiskComponent)` (where `NewDiskComponent` is the Go constructor function that returns a `*components.Disk` or similar).
    2.  **DSL Definition:** Users *might* still define a `component Disk { param ProfileName: string; }` in the DSL. This DSL definition primarily serves to declare the *parameters* expected by the native component for configuration during instantiation. It doesn't need method definitions. Alternatively, for purely native components, the DSL definition might be omitted entirely if parameters are handled differently.
    3.  **Instantiation (`instance myDisk: Disk`) Evaluation (`evalInstanceDecl`):**
        *   When `evalInstanceDecl` sees `instance myDisk: Disk`, it first checks `vm.ComponentRegistry` for "Disk".
        *   **If found:** It uses the registered native Go constructor (`NewDiskComponent`). It evaluates overrides (using the temporary literal workaround for now) and passes them to the constructor. The result is the Go object (e.g., `*components.Disk`), which is stored in the environment.
        *   **If not found:** *(Future)* The evaluator would look for a `component Disk { ... }` definition parsed from the DSL source files and prepare to evaluate its methods using the DSL interpreter.
    4.  **Method Call (`myDisk.Read()`) Evaluation (`evalCallExpr`):**
        *   `evalCallExpr` resolves `myDisk` in the environment. It finds the Go object (`*components.Disk`).
        *   Because the value found is **not** a DSL-defined object representation but a **native Go object**, `evalCallExpr` knows it should call the Go method directly.
        *   It uses reflection (or potentially a pre-cached method map associated with the registered component type) to find and invoke the `.Read()` method on the Go object.
        *   It takes the returned `*core.Outcomes[V]` and converts it to a `VarState` -> `LeafNode`.
*   **Pros:**
    *   No new DSL keywords are needed.
    *   Clear separation: Native components are managed entirely via Go code and registration.
    *   Flexible: Easy to swap a DSL implementation for a native one just by changing registration.
*   **Cons:**
    *   Less explicit *in the DSL* whether `Disk` is native or DSL-defined. Relies on understanding the VM setup.
    *   Requires a robust way to handle parameter overrides evaluated in the DSL and pass them correctly to the native Go constructor.
    *   Doesn't cleanly address mixing native *methods* within a primarily DSL-defined component (though maybe less common for this DSL's purpose).

**Option 2: Explicit `native` Keyword (Less Ideal)**

*   **How it Works:** Add a keyword like `native component Disk;` or `component Disk { native; ... }`.
*   **Pros:** Explicit in the DSL.
*   **Cons:** Awkward syntax. How are parameters defined? Doesn't easily solve native methods within DSL components. Adds grammar complexity.

**Option 3: Explicit `extern` Keyword (FFI Style - More Flexible but Complex)**

*   **How it Works:** Mark specific methods as external implementions.
    *   `component MyDSLComponent { uses n: NativeDisk; method DslLogic() { ... } extern method FastOp(): bool; }`
    *   `component NativeDisk { extern method Read(): AccessResult; extern method Write(): AccessResult; param ProfileName: string; }`
    *   Requires a separate VM registry for external functions (`vm.ExternalFunctions`).
    *   `evalCallExpr` checks an `IsExternal` flag on the `MethodDef` AST node. If true, it calls the registered external Go function; otherwise, it evaluates the DSL body.
*   **Pros:** Most explicit and flexible. Clearly supports mixing native and DSL methods.
*   **Cons:** Adds DSL keyword and significant implementation complexity to the evaluator and registration process. Likely overkill *initially*.

**Recommendation:**

Let's go with **Option 1 (Convention over Configuration)** for now. It provides the necessary mechanism without immediate changes to the DSL syntax or AST.

**How it fits the plan:**

1.  **`decl/vm.go`:** We already added `ComponentRegistry` and `RegisterComponent`. We need `registerDefaultComponents` to register the constructors for `components.Disk`, `components.Cache`, etc.
2.  **`decl/eval.go::evalInstanceDecl`:** The logic described above (check registry first) should be implemented. The temporary override evaluation workaround remains necessary.
3.  **`decl/eval.go::evalCallExpr` (Preview for Phase 4):** The core logic will be:
    *   Evaluate the receiver identifier (e.g., `myDisk`).
    *   Get the value from the environment.
    *   **Type Switch on the value:**
        *   **If `*components.Disk` (or other native Go type):** Use reflection to call the Go method (`Read`, `Write`, etc.). Convert result using `outcomeToVarState` -> `LeafNode`. *(Handle argument evaluation/passing - simplify for now)*.
        *   **If `DSLComponentInstance` (hypothetical type for DSL-defined components):** Look up the `MethodDef` in the DSL AST, evaluate its body -> `OpNode`. *(Logic for DSL methods)*.
        *   **If `OpNode`:** Error (cannot call methods on intermediate trees).
        *   **Else:** Error (unexpected type).
4.  **DSL `component` definitions:** For native components like `Disk`, the DSL `component Disk { param ProfileName: string; }` primarily serves documentation and parameter definition purposes. Its methods don't need to be defined in the DSL.

This approach allows us to treat all the components defined in the `sdl/components` package as "native" from the DSL's perspective, callable directly via the registry and reflection/interface calls. We defer the complexity of evaluating DSL-defined methods until later phases.
