# SDL DSL Syntax Design Evolution

Also see dsl/GRAMMAR.md

**1. Goal:**

Create a user-friendly, declarative Domain Specific Language (DSL) for defining and analyzing system performance models (and capacity planning) using the underlying SDL Go library (`sdl/core`, `sdl/components`, `sdl/decl`, `sdl/decl/components`). The DSL should abstract away some of the complexities of direct `Outcomes` manipulation while enabling powerful probabilistic modeling.

**2. Initial "Hybrid" Model (Implemented in the simple Evaluator):**

*   **Concept:** DSL method signatures declare conceptual return types like `AccessResult`, `Duration`. The VM internally maps these to the Go library's `*core.Outcomes[AccessResult]`, `*core.Outcomes[Duration]`, etc. Composition relies on the VM implicitly calling `core.And` with registered type-specific reducers.
*   **Example Syntax:**

```dsl
    component Disk {
        method Read(): AccessResult {
          // Go returns Outcomes[AccessResult]
          // Given tihs is a primitive leaf level -
          // we would initially implement this via a foreign func call.
        }
    }
    
    component Index {
        uses d: Disk;
        
        method ReadBoth(): AccessResult {
            a = d.Read(); // a -> Outcomes[AccessResult] implicitly
            b = d.Read(); // b -> Outcomes[AccessResult] implicitly
            // Implicit sequential 'And' combines a and b using AccessResult reducer
            return b;
        }
        method CheckRead(): bool { // Returns bool conceptually
             res = d.Read();
             if res.Success { return true; } // If splits Outcomes[AccessResult]
             else { return false; }
        }
    }
```
*   **Pros:** Relatively direct mapping to Go library; VM manages outcome containers.
*   **Cons:** DSL users still need awareness of types like `AccessResult`; defining primitives requires Go code returning `Outcomes`; `.Success` access for `if` requires special VM handling; passing values between probabilistic steps is difficult. Ergonomics of latency handling are awkward (latency always bundled with value).

**3. "Implicit Outcomes V1" Proposal (Discussion Only):**

*   **Concept:** Hide `Outcomes` completely. Operations return base types (`bool`, `int`). Probability introduced via `distribute { prob => expr }`. Latency added via `Delay()` or tuple `(value, latency)`.
*   **Example Syntax:**
    ```dsl
    method Read(): bool {
        distribute {
            0.9 => Delay(10ms); return true;
            0.1 => (false, 100ms);
        }
    }
    ```
*   **Pros:** Potentially more intuitive syntax for simple cases.
*   **Cons:** VM complexity increases significantly (needs to track implicit distributions); unclear latency accumulation rules; `distribute` block overloaded.

**4. "Implicit Outcomes V2 + Discrete Values" Refinement (Discussion Only):**

*   **Concept:** Like V1, but restrict probabilistic *value* returns (`V` in `Outcomes[V]`) to **finite, discrete sets** (enums, bool, status codes). Latency (`Outcomes[Duration]`) remains continuous. VM tracks value + latency distributions implicitly. Introduces `distribute V with latency { prob => value ; latency; }`.
*   **Pros:** Enables a feasible **Value Reducer Registry** for combining discrete types sequentially; better ergonomics than Hybrid.
*   **Cons:** Still bundles latency definition with value in `distribute`; VM complexity high (dual tracking). Restriction on return types.

**5. "Model V3: + Explicit Delay Statement" Refinement (Discussion Only):**

*   **Concept:** Like V2, but separate probabilistic *choice* (`distribute`) from *latency addition* (`Internal.Delay`). `distribute` selects an expression branch; latency comes only from leaf methods or `Delay`.
*   **Example Syntax:**
    ```dsl
    // Using 'distribute with latency' for primitives
    method Read(): bool {
         return distribute bool with latency { /* ... */ };
    }
    // Using pure 'distribute' for choice
    method ReadPreferred(): bool {
        return distribute bool {
            0.5 => primary.Read();
            // ...
        };
    }
    // Adding delay in logic
    if !d.Write() { delay 10ms; return false; }
    ```
*   **Pros:** Clear separation of concerns (choice vs. timing); composable.
*   **Cons:** Primitives still need a way to define their base latency associated with their value outcome (leading back to `distribute with latency` or similar). VM complexity remains high.

**6. Chosen Model V4: Implicit Outcomes + Discrete Values + Explicit Delay + Optional Total/Default (Current Target):**

*   **Implicit Tracking:** VM manages `Outcomes[V]` (discrete value) and `Outcomes[Duration]` (latency) implicitly for variables and expressions.
*   **Discrete Value Sets:** `V` must be a type with a finite set of values (e.g., `bool`, registered enums like `HttpStatusCode`). Structs containing continuous data are allowed if the branching logic relies on a discrete member.
*   **`delay durationExpr` Statement:** The **only** way to explicitly add latency within a logic block. It modifies the implicit latency track.
*   **`distribute [totalExpr] { probExpr => { block }; ... [default => { block }] }` Statement/Expression:**
    *   Purely for **probabilistic control flow**. Selects a block to execute based on probabilities.
    *   Does **not** add latency itself. Latency comes from methods *within* the chosen block or preceding methods.
    *   Evaluates the chosen block in a context split according to the branch probability.
    *   Combines the value and latency results from all executed branches using `Append`.
    *   Can be used as an expression if all branches return the same required type.
    *   `totalExpr` and `default` are optional conveniences.
*   **Primitives:** Base component methods (like `Disk.Read`) are expected to internally use `delay` within their logic (potentially within a `distribute` block) to associate their inherent methodal latency with their success/failure/value outcomes before returning the final functional value (e.g., `bool`).
*   **Composition:**
    *   **Sequence (`a=Op1(); b=Op2()`):** VM implicitly combines latency tracks using `And(latA, latB, AddDurations)`. The value track effectively becomes that of the last method (`valB`), associated with the combined latency.
    *   **`if condition { then } else { else }`:** VM evaluates `condition` (which must yield `Outcomes[bool]` or access a `.Success`/discrete field), splits *both* value and latency tracks, executes branches with split contexts, and `Append`s the resulting value and latency distributions.

*   **Example (Disk Read in V4):**
    ```dsl
    component Disk {
        param ReadLatency: duration = 10ms;
        param FailureProb: float = 0.01;

        method Read(): bool { // Functional return type
            distribute { // Default total = 1.0
                self.FailureProb => {
                    delay 10s;     // Add explicit failure latency
                    return false;  // Return value for this path
                }
                default => { // Prob = 1.0 - FailureProb
                    delay self.ReadLatency; // Add explicit success latency
                    return true;   // Return value for this path
                }
            }
            // VM combines branches into:
            // Value: Outcomes[bool] { (1-FP)=>true, FP=>false }
            // Latency: Outcomes[Duration] { (1-FP)=>ReadLatency, FP=>10s }
        }
    }
    ```

Couple of notes:
* distribute is a statement.
* There is also an expression version of it `distribution` which can be passed as values:
  ```
  // x will be of type Enum<val1, val2, val3, val4>
  x := distribute Total {
    a => val1
    b => val2
    c => val3
    _ => val4
  }
  
  // Add the latency given by above
  delay x
  ```
* Total, a, b and c should eval to a number.  `_` is the "default" case and is used for Total - (a + b + c).  If this
  resolves to <= 0 then default is ignored.
* val1, val2, val3 ... should all be members of the SAME enum set.
* Using this we could do `Delay(x)` or `latency x` 
* The above can be written using a `distribute` statement:
  ```
  distribute Total {
    a => disk.Write()
    b => delay 20
    c => disk.Read()
    _ => ;
  }
  ```
* Advantage of the distribute statement is it is simpler sometimes instead of having to calculate and save a local.
  Also dist and distribute are similar to a `random` being called (ie it is a point where we could sample if needed).
  
Can we perform the above with a switch?

```
switch x {
  case a => Stmt1
  case b => Stmt2
  case c => Stmt3
}
```

What are the types `x` can take?

* If x is a primitive - eg duration, int, float, string then all the cases in the switch except where case == x will be
ignored and this becomes just another "linear" step.
* If x is an enum then it is treated as a distribution where each value has equal probability.
* If x is Outcome[enum] then a/b/c/ are split by and combined

**Conclusion on V4:**

Model V4 provides the best balance found so far:

*   **User Experience:** Clean syntax, implicit outcome handling, natural expression of logic based on functional values.
*   **Clarity:** Clear separation between probabilistic choice (`distribute`) and latency addition (`delay`).
*   **Feasibility:** While demanding on the VM implementation (dual-track synchronization), the constraint to discrete value sets makes value combination manageable (often "take last" or potentially using value reducers if needed).
*   **Composability:** Base methods define their intrinsic value/latency profiles, which are then composed implicitly by the VM during sequence and conditional execution, with explicit `delay` for adjustments.
