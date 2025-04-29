**Prompt for Next LLM Task (Phase 4 - DSL Implementation):**

```text
**Project Context:**

You are continuing development on the SDL (System Design Language) Go library. This library enables performance modelling (latency, availability) of system components using probabilistic `Outcomes` distributions and analytical approximations, avoiding full discrete-event simulation for speed and focusing on steady-state analysis.

**Current State:**
- The core Go library (conceptually `core`, `results`, `metrics`, `components`, `examples`) provides:
    - `Outcomes[V any]`, composition operators (`And`, `If`, `Map`, `Split`, `Append`).
    - `AccessResult`, `RangedResult` types with robust reduction (`TrimToSize...`).
    - Metric calculations (`Availability`, `MeanLatency`, `PercentileLatency`).
    - Sampling and distribution generation (`Sample`, `NewDistributionFromPercentiles`).
    - A stateless `Analyze` primitive (`core.Analyze`, `AnalysisResult`) for standardized simulation execution and expectation checking, used throughout tests.
    - A suite of component models (`Disk`, `NetworkLink`, `Queue`, `ResourcePool`, `Cache`, `LSMTree`, `BTreeIndex`, etc.). Notably, `ResourcePool` is now fully stateless and analytical.
    - Examples (`bitly`, `gpucaller`) demonstrating usage.
- Known limitations of the analytical approach (steady-state focus, stateless pool, profile accuracy dependence) are documented.
- Design documents (`DSL.md`, `ROADMAP.MD`) outline the plan for a "Hybrid" DSL (explicit `Outcomes` in Go, implicit/cleaner syntax in DSL).

**Phase 4 Goal:** Focus on Usability & System Composition via the DSL.

**Next Task (Phase 4, Task 1): Design and Implement the DSL Parser**

Your primary task is to **finalize the v1 syntax for the SDL DSL** based on the "Hybrid" approach and **implement the parser** using `participle` (or another suitable Go library).

**Requirements:**

1.  **DSL Syntax Design (v1):**
    *   Define a clear, HCL-like syntax for *declarations* and *signatures*:
        *   `component "Name" { ... }`
        *   `param name: type = defaultValue;` (string, int, float, bool, duration)
        *   `uses name: ComponentType;`
        *   `operation name(params...): ReturnValueType { /* Body parsing deferred */ }` (Parse only the signature. ReturnValueType like `AccessResult`, `Duration`, `bool`).
        *   `system "Name" { ... }`
        *   `instance name: ComponentType = { paramOverrides... };`
        *   `analyze analysisName = instance.Operation(args);`
        *   `options { MaxOutcomeLen = 10; DefaultDisk = "SSD"; /* etc */ }`
    *   Refer to `DSL.md` for the "Hybrid" approach philosophy. Focus only on the structure needed for parsing these top-level elements and signatures.

2.  **Parser Implementation:**
    *   Create the `sdl/dsl` directory.
    *   Use `participle` to define Go structs in `sdl/dsl/ast.go` that directly map to the DSL grammar elements above (e.g., `ast.File`, `ast.ComponentDecl`, `ast.ParamDecl`, `ast.OperationSignature`, `ast.SystemDecl`, `ast.InstanceDecl`, `ast.AnalyzeDecl`, `ast.OptionsDecl`, `ast.Literal`, `ast.Ident`, `ast.TypeSpec`). Implement a basic `Node` interface (e.g., providing position info).
    *   Implement the parser in `sdl/dsl/parser.go` (e.g., `func Parse(filename string, src io.Reader) (*ast.File, error)`) using `participle.New`.
    *   Add unit tests in `sdl/dsl/parser_test.go`. Test parsing of simple DSL snippets containing *only* the declarative elements (no complex operation bodies yet). Verify the correct AST structure is generated. Test basic syntax error reporting from `participle`.

**Focus:** Create a parser capable of understanding the structure of component/system definitions, parameter/dependency declarations, operation *signatures*, and analysis targets. **Do not** parse or define the syntax for the *logic inside operation bodies* yet.

**Deliverable:** Go code for `sdl/dsl/ast.go`, `sdl/dsl/parser.go`, and `sdl/dsl/parser_test.go`. Include examples of the designed v1 DSL syntax used in the tests.
```
