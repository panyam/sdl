## **SDL DSL Design & Implementation Plan Summary**

**1. Overall Goal & Philosophy:**

*   Create a user-friendly, declarative Domain Specific Language (DSL) for defining and analyzing system performance models.
*   The DSL serves as a high-level interface to the underlying SDL Go library.
*   The Go library remains the canonical implementation of probabilistic logic (`Outcomes`, composition operators, component models).
*   The DSL **compiles/interprets down to calls** within the Go library. It should not introduce simulation semantics unavailable in the library.
*   Prioritize analytical modelling speed over detailed Discrete Event Simulation (DES) for interactive "what-if" analysis. State evolution is primarily handled externally by changing component parameters between simulation runs.

**2. Chosen DSL Approach: "Hybrid" (Explicit Outcomes in Go, Implicit in DSL Syntax)**

*   **Go Library:** Component methods (`disk.Read`, `cache.Write`, etc.) **always** return `Outcomes[V]` (e.g., `Outcomes[AccessResult]`, `Outcomes[Duration]`). The core engine explicitly manipulates these distributions.
*   **DSL Syntax:**
    *   Operation signatures omit the `Outcomes<>` wrapper for brevity (e.g., `operation Read(): AccessResult`).
    *   Operation logic primarily manipulates variables that *conceptually* hold concrete types, but the interpreter maps these to the underlying `Outcomes` objects.
*   **Interpreter Role:** Translate the cleaner DSL syntax into explicit calls to the Go library's `Outcomes` manipulation functions (`And`, `If`, `Map`, `Split`, `Append`, `Trim...`), managing the composition of distributions behind the scenes.

**3. Proposed DSL Syntax (Based on HCL-like Structure + Script-like Logic):**

*   **Declarations:** Use HCL-like blocks:
    *   `component "Name" { ... }`
    *   `system "Name" { ... }`
    *   `options { ... }` (Global or component-level hints like `MaxOutcomeLen`, `DefaultDiskProfile`)
*   **Component Members:**
    *   `param name: type = defaultValue;` (Types: `string`, `int`, `float`, `bool`, `duration`)
    *   `uses name: ComponentType;` (Dependencies)
    *   `operation name(params): ReturnValueType { ... }` (e.g., `AccessResult`, `Duration`, `bool`)
*   **Operation Logic Block (`{ ... }`):**
    *   **Implicit `self`:** Unqualified calls (`cache.Read()`) refer to component's dependencies.
    *   **Assignment:** `var = component.Operation(args)` (Interpreter stores resulting `Outcomes` object in `var`).
    *   **Sequential Composition (Implicit `And`):** `a = Op1(); b = Op2(); return b;` -> Interpreter generates `And(Op1_Outcomes, Op2_Outcomes, ...)`. Needs default reducer logic based on types. Result of sequence is outcome of last step's composition.
    *   **Conditional (`If`):** `if var.Success { <then_block> } else { <else_block> }`. Interpreter translates to logic using `Split` on `var`'s `Outcomes` based on the `.Success` field (or other boolean field), evaluates branches with split distributions, and combines results (`Append`) adding conditional check latency.
    *   **Return:** `return var` (Interpreter returns the `Outcomes` object associated with `var`). If returning a literal, interpreter wraps it in a deterministic `Outcomes` distribution.
    *   **Map (Future):** May need explicit `Map(var, func...)` syntax.
    *   **Value Access:** **Disallowed for now.** Logic cannot directly access fields like `var.Latency` or use non-boolean fields in conditions. Focus is on composing the `Outcomes` containers based on boolean success/failure paths.
    *   **Randomness:** No explicit `rand()` in DSL. Probability originates from components' `dist { ... }` blocks (if that syntax is adopted for primitives) or from the composition of `Outcomes`.
*   **System Definition:**
    *   `instanceName: ComponentType = { paramOverrides... };` (Overrides use literals or references to other instances).
    *   `analyze analysisName = instance.Operation(args);` (Triggers simulation/analysis of the target operation call).
    *   *(Future):* `expect` clauses for SLO checks. Context blocks for scenario parameters.
*   **Parameter Passing:** Methods *can* take parameters (e.g., `Acquire(lambda float64)`). The interpreter handles passing arguments during calls.

**4. Interpreter Design & Challenges:**

*   **AST:** Requires nodes for all DSL constructs (Components, Params, Uses, Ops, Systems, Instances, Analyses, Options, Statements: Assign, If, Return, Call; Expressions: MemberAccess, Call, Literal, Identifier). (AST draft defined previously).
*   **Environment/Symbol Table:** Manages scopes and maps names to runtime objects (Go components, `Outcomes` objects).
*   **Evaluation (`eval`):** Recursively walks the AST.
    *   Translates DSL control flow (`if`, sequence) into Go library calls (`Split`, `Append`, `And`, `Map`).
    *   Manages the "current accumulated outcome" during sequential execution.
    *   Handles type checking/dispatching (using runtime type switches initially) to call appropriate Go reducers/trimmers for different `Outcomes[V]` types.
    *   Handles the implicit `Outcomes` mapping (DSL `AccessResult` -> Go `*Outcomes[AccessResult]`).
    *   Applies implicit reduction based on `options` hints.
*   **Stateful Components (`ResourcePool`):** The `Release` operation remains a known limitation needing direct state modification, breaking the pure functional model. Acquisition (`Acquire`) relies on passed-in parameters (`lambda`) rather than internal state (`Used`) for calculations.

**5. Next Steps (Phase 4):**

1.  **Implement Parser:** Choose a library (e.g., `participle`) and implement the parser in `sdl/dsl/parser.go` to generate the AST defined in `sdl/dsl/ast.go` based on the Hybrid DSL grammar. Include basic parser tests.
2.  **Implement Interpreter:** Create the interpreter (`sdl/dsl/interpreter.go`) that walks the AST, manages the environment, evaluates expressions/statements, and orchestrates calls to the core SDL Go library functions to execute the simulation logic and produce final `Outcomes`. Start with core features (sequence, assignment, calls, basic `if`).
3.  **Integration & Testing:** Test the interpreter with simple DSL examples parsed into ASTs.
4.  **Refine & Extend:** Add support for more complex expressions, mapping, potentially looping, scenario context, `expect` clauses, etc.

This summary captures the refined DSL direction, emphasizing usability while grounding the execution in the existing Go library's explicit `Outcomes` manipulation.

## 'Analyze' Feature

Currently, in the Go library, analysis is done **explicitly within Go test files** or potentially a separate Go application. We:

1.  Instantiate the Go component structs (e.g., `bs := setupBitlyService(t)`).
2.  Call the operation method on the top-level component (`redirectOutcomes := bs.Redirect(...)`). This returns the final `*Outcomes[V]` object.
3.  Call the metric helper functions from `sdl/metrics/` on the result (`avail := Availability(redirectOutcomes)`, `p99 := PercentileLatency(redirectOutcomes, 0.99)`).
4.  Print the results or assert against expected values (`t.Logf(...)`, `if p99 > target {...}`).

The `analyze` block in the DSL aims to **automate and standardize this process**, making it part of the system definition itself rather than requiring separate Go code for each analysis run.

**Mapping `analyze` to the Go Library:**

The `analyze analysisName = instance.Operation(args);` syntax in the DSL maps to the following actions performed by the **Interpreter**:

1.  **Trigger Simulation (`instance.Operation(args)` evaluation):**
    *   This part maps directly to **calling the corresponding Go method** on the target component instance (e.g., `theService` instance's `Redirect` method).
    *   The interpreter performs all the necessary composition (`And`, `If`, `Map`, `Trim...`) by calling the Go library functions as it evaluates the operation's logic defined in the DSL.
    *   The **final result** of evaluating this expression is the `*Outcomes[V]` object returned by the top-level Go method call (e.g., the `*Outcomes[AccessResult]` from `bs.Redirect`).

2.  **Store Result:**
    *   The interpreter stores the resulting `*Outcomes[V]` object, associating it with the `analysisName` (e.g., "Redirect"). This makes the result available for potential later inspection or use in `expect` clauses.

3.  **Implicit Metric Calculation & Output (Interpreter Feature):**
    *   This part **doesn't map to a single Go library function** but is a feature of the **DSL interpreter itself**.
    *   After obtaining the `*Outcomes[V]` result for an `analyze` block, the interpreter automatically calls the Go library's metric functions:
        *   It checks if the result type `V` implements `Metricable` (using type switching/reflection).
        *   If yes, it calls `metrics.Availability(resultOutcomes)`, `metrics.MeanLatency(resultOutcomes)`, `metrics.PercentileLatency(resultOutcomes, 0.50)`, `metrics.PercentileLatency(resultOutcomes, 0.99)`, etc.
        *   It then formats and prints these calculated metrics to the user's console or another output stream.

4.  **`expect` Clause Evaluation (Optional Interpreter Feature):**
    *   If `expect Target.P99 < 0.100` syntax is added:
        *   The interpreter parses this check associated with the `analyze` block.
        *   After getting `resultOutcomes` for `Target`, it calculates the specific metric (`PercentileLatency(resultOutcomes, 0.99)`).
        *   It compares the calculated value to the threshold (0.100).
        *   It reports "PASS" or "FAIL" for the expectation. This also uses the Go `metrics` library functions.

**In Summary:**

*   The **expression part** of `analyze` (`instance.Operation(args)`) directly maps to **executing the simulation logic** by calling the Go component methods and composition functions.
*   The **analysis/reporting part** (calculating and displaying metrics, evaluating `expect` clauses) maps to the **interpreter automatically calling functions from the `sdl/metrics/` package** on the resulting `Outcomes` object.

The `analyze` block essentially bundles the Go library execution trigger and the standard metric reporting into a convenient DSL construct. It makes running standard performance analyses declarative within the system definition.

## Visuals in UX

While the core SDL engine focuses on the analytical/probabilistic outcome of a *single* operation flow, visualizing the *implications* of that distribution under sustained load (like 10k calls/sec) is crucial for user understanding and provides a much richer UX feedback loop.

We can achieve this **without** actually running a 10k QPS discrete-event simulation, by using the results of our analytical model to *parameterize* visualizations or *derive* system-level metrics under load.

Here's how the UX feedback loop could evolve, integrating visualization derived from the analytical model:

**1. Core SDL Analysis (As Planned):**

*   User defines the system in DSL (components, params, wiring).
*   User defines an `analyze` block targeting a key operation (e.g., `analyze ReadPerf = service.ReadUser("id")`).
*   Interpreter runs the analysis by composing `Outcomes` via the Go library.
*   Interpreter calculates and displays basic metrics from the final `Outcomes[AccessResult]` distribution:
    *   Availability: `Availability(ReadPerf)` -> e.g., 0.9995
    *   Mean Latency: `MeanLatency(ReadPerf)` -> e.g., 0.015s (15ms)
    *   P50 Latency: `PercentileLatency(ReadPerf, 0.50)` -> e.g., 0.008s (8ms)
    *   P99 Latency: `PercentileLatency(ReadPerf, 0.99)` -> e.g., 0.045s (45ms)
    *   P99.9 Latency: `PercentileLatency(ReadPerf, 0.999)` -> e.g., 0.090s (90ms)

**2. Enhanced UX - Visualization & Load Extrapolation (Post-Analysis):**

After the core analysis provides the `Outcomes` distribution for the single operation, the interpreter or a separate visualization tool could offer:

*   **A. Latency Distribution Plot (Histogram/CDF):**
    *   **How:** Generate this directly from the final `Outcomes[AccessResult]` distribution. Iterate through the buckets (`Weight`, `Latency`).
    *   **Histogram:** Create bins on the X-axis (latency ranges). Sum the `Weight` of buckets falling into each bin. Plot bars representing probability density.
    *   **CDF (Cumulative Distribution Function):** Sort buckets by latency. Plot cumulative weight (Y-axis, 0 to 1.0) against latency (X-axis). This clearly shows percentiles (e.g., the latency at Y=0.99 is the P99).
    *   **UX:** Show a plot alongside the numerical metrics. Users can visually grasp the shape of the distribution, identify long tails, or multi-modal behavior.

*   **B. Estimated Queuing Delay Plot (Requires Load Input):**
    *   **How:**
        1.  Ask the user for an **input Load (QPS)**, e.g., "10,000 calls/sec". This is our target `lambda`.
        2.  Identify potential bottlenecks using the component models involved in the `analyze` target (specifically `Queue` and `ResourcePool` components).
        3.  For each `Queue` and `ResourcePool` in the path:
            *   Use its configured `Servers` (c) and `AvgServiceTime`/`AvgHoldTime` (Ts).
            *   Calculate utilization `rho = input_lambda / (c * mu)`.
            *   Calculate average waiting time `Wq` using the M/M/c(/K) formulas **based on the user-provided `input_lambda`**.
        4.  **Visualize:** Plot the calculated `Wq` for each queue/pool against the component name, or potentially against varying `input_lambda` values (showing how wait time increases with load). Show utilization (`rho`) as well.
    *   **UX:** This plot directly answers the "what happens at 10k QPS?" question in terms of *average waiting time* at key contention points. It highlights which queues/pools become saturated (`rho >= 1`) or experience high delays.

*   **C. Estimated Overall Latency Under Load Plot:**
    *   **How:** This combines A and B.
        1.  Get the base latency distribution `Outcomes` from the core analysis (like in A).
        2.  Get the estimated *average* queuing delays (`Wq`) for relevant queues/pools at the target QPS (like in B).
        3.  **Approximate:** Add the calculated *average* queuing delays (`Wq`) to the latency of *each bucket* in the base `Outcomes` distribution. This effectively shifts the entire latency distribution plot (from A) to the right based on the average queuing impact.
        4.  Plot the *shifted* latency distribution (Histogram/CDF). Recalculate P99 etc. from this shifted distribution.
    *   **UX:** Shows the user the predicted *end-to-end* latency distribution (including estimated average queuing) at their specified load level. They can directly see if the P99 under load exceeds SLOs. *Crucially, this avoids simulating 10k individual requests.*

*   **D. Sampling for Scatter Plot / Jitter Visualization:**
    *   **How:**
        1.  Use the `ConvertToRanged` function on the final `Outcomes[AccessResult]` to add synthetic jitter/range.
        2.  Generate N samples using `Sample()` on the `Outcomes[RangedResult]`.
        3.  For each sampled `RangedResult`, use `SampleWithinRange()` to get a single latency point reflecting the range.
        4.  Plot these N points on a scatter plot (Y-axis = latency, X-axis = sample number or time).
    *   **UX:** Provides a visual sense of the variability and potential jitter in the response times, even though it's derived from the analytical model plus synthetic range.

**Implementation:**

*   The core Go library needs the `Sample()` and `ConvertToRanged`/`SampleWithinRange` functions (which we added).
*   The `sdl/metrics` package might need functions to generate data suitable for plotting (e.g., return histogram bins/counts, or CDF points).
*   The **Interpreter** (or a wrapper tool around it) becomes responsible for:
    *   Running the core analysis.
    *   Generating data for the Latency Distribution Plot (A).
    *   Prompting for target QPS.
    *   Identifying Queues/Pools in the execution path (requires tracing during evaluation or analyzing the AST).
    *   Calculating `Wq` based on user QPS (B).
    *   Generating data for the Estimated Queuing Delay Plot (B).
    *   Combining base outcomes with average delays to generate data for the Estimated Overall Latency Plot (C).
    *   Optionally generating samples for the Scatter Plot (D).
    *   Ideally, integrating with a simple plotting library (Go or external via data export) to display these visualizations.

This layered approach provides increasing levels of insight: starting with the core analytical distribution, then estimating queuing impact under load, and finally visualizing the potential variability, all without departing from the faster analytical foundation.

## DSL Challenges

Addressing the interpreter's core challenges *before* locking down the parser/AST makes a lot of sense.
If the interpreter logic fundamentally changes how we need to structure or represent things, it's better to know that *before* building the parser around potentially flawed assumptions.

**Interpreter Challenges & Potential Solutions for Hybrid DSL with Explicit Outcome<T> in Go and implicit in DSL:**

1.  **Mapping DSL Logic to `Outcomes` Composition:** This is the central challenge. How do `if`, sequence, and variable usage in the DSL translate to `And`, `If`, `Map`, `Split`, `Append` calls on `Outcomes` objects?

    *   **Challenge:** A simple DSL sequence `a = Op1(); b = Op2(); return b` needs to become roughly `And(Op1Outcomes, Op2Outcomes, DefaultReducer)`. An `if a.Success { OpT() } else { OpF() }` needs to become something involving `Split(a_Outcomes)`, `And(success_split, OpT_Outcomes)`, `And(failure_split, OpF_Outcomes)`, and `Append`. The interpreter needs to manage this translation.
    *   **Solution Approach:**
        *   **Environment Holds `Outcomes`:** When `a = Op1()` is evaluated, the interpreter stores the `*Outcomes[T]` returned by the Go `Op1` method in the environment, associated with the name `a`.
        *   **Sequential `And`:** The interpreter maintains a "current accumulated outcome" for the execution block. When it evaluates the next statement (`b = Op2()`), it fetches the required inputs (if any) from the environment, evaluates `Op2()` to get `Op2_Outcomes`, and then calls the Go `sdl.And(currentOutcome, Op2_Outcomes, selectReducer(...))` function. The result becomes the new `currentOutcome`.
        *   **`If` Statement:** This is the most complex. When evaluating `if condExpr { ThenStmts } else { ElseStmts }`:
            1.  Evaluate `condExpr`. This involves getting the `*Outcomes[V]` associated with the variable in the condition (e.g., `a` in `a.Success`).
            2.  **Split Distribution:** Call `a_Outcomes.Split(func(v V) bool { return v.Success /* or other field */ })` to get `successOutcomes` and `failureOutcomes`.
            3.  **Evaluate Then Branch:** Recursively evaluate `ThenStmts` *using `successOutcomes` as the starting context/input distribution*. Let the result be `thenResultOutcomes`.
            4.  **Evaluate Else Branch:** Recursively evaluate `ElseStmts` *using `failureOutcomes` as the starting context/input distribution*. Let the result be `elseResultOutcomes`.
            5.  **Combine:** Append `thenResultOutcomes` and `elseResultOutcomes`. `finalResult = thenResultOutcomes.Append(elseResultOutcomes)`.
            6.  **Latency:** The latency of evaluating the *condition* itself (if non-negligible, e.g., if `condExpr` was an operation call) needs to be added to *both* `thenResultOutcomes` and `elseResultOutcomes` *before* appending. This requires careful management during the recursive calls.
        *   **`Return` Statement:** Evaluate the return expression to get `retOutcomes` and signal this as the result of the current operation block evaluation.

2.  **Handling Different `Outcomes[V]` Types:** How does the interpreter know which reducer to use for `And` or how to `Split` based on `if x.SomeField` when `x` could be `Outcomes[AccessResult]`, `Outcomes[Duration]`, `Outcomes[bool]` etc.?

    *   **Challenge:** The Go library relies on type information for selecting reducers (`AndAccessResults`, `AndRangedResults`) and predicates for `Split`. The interpreter operates on AST nodes and runtime values (which will be `interface{}` holding `*Outcomes[?]`).
    *   **Solution Approaches:**
        *   **Runtime Type Reflection/Switches:** The interpreter's `And` or `If` logic uses Go's type reflection (`reflect` package) or type switches on the actual `*Outcomes[V]` objects held in the environment to determine the types `V` involved and select the appropriate Go library reducer function or build the correct `Split` predicate. This can be slow and complex.
        *   **AST Type Information:** Enhance the AST/parsing phase to include resolved type information. When parsing `a = Op1()`, store the fact that `a` has type `Outcomes[AccessResult]`. The interpreter then uses this stored type info to select Go functions without runtime reflection. Requires a type-checking/resolution phase after parsing.
        *   **Interface-Based:** Define interfaces for common operations needed by the interpreter (e.g., `SplittableBySuccess`, `Mappable`, `CombineableWith`). The Go `Outcomes[V]` methods or associated helper functions would implement these. The interpreter calls methods via these interfaces. This requires careful interface design.
        *   **Recommendation:** Start with **Runtime Type Switches** as it's often the most direct way in Go, accepting the potential performance cost. If it becomes a bottleneck, move to AST Type Information.

3.  **Accessing Fields within `Outcomes` (e.g., `if cacheRead.Success`):**

    *   **Challenge:** `cacheRead` in the interpreter holds `*Outcomes[AccessResult]`. The DSL condition `.Success` refers to a field *inside* the `AccessResult` values within the distribution's buckets.
    *   **Solution Approach:** The interpreter's `eval` for `IfStmt` (and potentially other expressions accessing fields) needs to understand this. When it evaluates `cacheRead.Success`, it doesn't return a single bool. Instead, it triggers the **Split** operation on the `cacheRead` `Outcomes` object, using a predicate based on the accessed field (`func(v AccessResult) bool { return v.Success }`). The `If` evaluation then proceeds with the two resulting distributions as described in point 1. Accessing non-boolean fields (e.g., `cacheRead.Latency`) directly in expressions is problematic and should likely be disallowed initially unless within a `Map` construct.

4.  **Value Passing / Accessing Inner Values (e.g., using generated ID):**

    *   **Challenge:** As identified before, DSL `id=Gen(); Save(id)` doesn't work if `id` holds `Outcomes`.
    *   **Solution Approach (Sticking to composing Outcomes):** Disallow direct value access for now. Operations that logically need a value from a previous probabilistic step need to be designed differently in the DSL/Go library.
        *   Example: `SaveMapping(idOutcomes *Outcomes[string], ...)` - The `SaveMapping` Go method itself would take the *distribution* of possible IDs and handle the composition internally (applying the save logic for each possible ID, weighted by probability). This pushes complexity into the Go component methods.
        *   Example: Introduce explicit `Bind` or `FlatMap` operation in the DSL/library.
    *   **Recommendation:** Keep the initial DSL simple and **do not support** direct value extraction from `Outcomes` variables within the operation logic. Focus purely on composing the `Outcomes` containers.

5.  **Implicit Reduction:**

    *   **Challenge:** When and how to apply `TrimToSize` / `TrimToSizeRanged`?
    *   **Solution Approach:**
        *   The interpreter, after key composition steps (end of an `And` sequence, after combining `If/Else` branches, before `Return`), checks the `Len()` of the resulting `Outcomes` object.
        *   It fetches the relevant `MaxOutcomeLen` value from the current component's `options` or global `options`.
        *   If `Len() > MaxOutcomeLen`, it uses runtime type switching/reflection to determine the type `V` and calls the appropriate Go library function (`TrimToSize` for `AccessResult`, `TrimToSizeRanged` for `RangedResult`, etc.).

6.  **Parameterizing Queues/Pools:**

    *   **Challenge:** How does `myQueue.Dequeue()` get the `lambda`/`Ts` it needs if they aren't stored in the `Queue` struct itself (because they are scenario-dependent)?
    *   **Solution Approach (Matching Recommendation):** The Go methods for `Queue.Dequeue` and `Pool.Acquire` *must* take the necessary rates (`lambda`, `Ts` or `avgHoldTime`) as **arguments**.
        *   **DSL:** `qWait = myQ.Dequeue(lambda=currentLambda, avgServiceTime=workerTs)` or `acquired = myPool.Acquire(lambda=currentLambda)`.
        *   **Interpreter:** When evaluating these calls, it needs to find the values for `currentLambda`, `workerTs` etc., from the current execution environment (which might have received them as parameters to the operation or from a scenario context block).

**Revised Interpreter Plan:**

*   Focus on the **Hybrid DSL** execution model.
*   Use **runtime type switches/reflection** initially to handle different `Outcomes[V]` types and select appropriate Go library functions (reducers, trimmers).
*   Implement **sequential execution as implicit `And`**, maintaining a current `Outcomes` result.
*   Implement **`If Stmt` by using `Split`** based on the condition field, recursively evaluating branches with the split outcomes, and appending results (adding condition latency).
*   **Disallow direct value access** from `Outcomes` variables in expressions (except for boolean fields used in `If`).
*   Implement **implicit reduction** after composition steps based on `options`.
*   Ensure methods like `Queue.Dequeue` and `Pool.Acquire` take necessary **rate parameters as arguments** in their Go signatures, and the DSL syntax/interpreter supports passing these arguments.

This plan addresses the core challenges while staying aligned with the explicit `Outcomes` manipulation in the Go library and avoiding the full complexity of the "Implicit Outcomes" VM. It defines a clear path for the interpreter's responsibilities.
