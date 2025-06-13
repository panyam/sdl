# Flow Evaluation System Design (`FlowEval`)

**Purpose:** Analytical computation of traffic flow patterns and system performance metrics without running simulations. This is the foundation for real-time dashboard updates in the workshop presentation system.

---

## 🎯 **Core Requirements**

### **Primary Function:**
Given a component.method receiving a specific input rate, compute all outbound traffic rates to downstream components.

### **Key Design Goals:**
1. **Real-time Performance:** Support live parameter changes during conference presentations
2. **Mathematical Accuracy:** Handle retry patterns, conditional flows, and resource constraints
3. **Cycle Robustness:** Gracefully handle recursive calls and feedback loops
4. **Native Component Integration:** Support both SDL and Go-implemented components

---

## 🏗️ **Architecture Overview**

### **Core FlowEval Signature:**
```go
// Given: component.method receiving inputRate
// Return: map of {downstreamComponent.method: outputRate}
func FlowEval(component, method string, inputRate float64, context *FlowContext) map[string]float64

type FlowContext struct {
    System *SystemDecl                    // SDL system definition
    Parameters map[string]interface{}     // Current parameter values (hitRate, poolSize, etc.)
    SuccessRates map[string]float64       // Component success rates for retry analysis
    ServiceTimes map[string]time.Duration // Component service times for queueing
    ResourceLimits map[string]int         // Pool sizes, capacities
    
    // Cycle handling configuration
    MaxRetries int                        // Limit exponential growth (recommended: 50)
    ConvergenceThreshold float64          // Fixed-point iteration threshold (recommended: 0.01)
    MaxIterations int                     // Maximum fixed-point iterations (recommended: 10)
    CallStack []string                    // Detect infinite recursion
}
```

---

## 🌊 **Flow Analysis Examples**

### **Simple 1:1 Flow:**
```sdl
method Read() Bool {
    return self.OurDisk.Read()
}
```
```go
// FlowEval("DiskServer", "Read", 5.0, ctx) returns:
// {"DiskWithCapacity.Read": 5.0}
```

### **Conditional Flow (Cache Pattern):**
```sdl
method Read() Bool {
    let cached = self.cache.Read()      // Always called: 10.0 RPS
    if cached { return true }
    return self.disk.Read()             // Called on cache miss: 2.0 RPS (20% miss rate)
}
```
```go
// FlowEval("CacheServer", "Read", 10.0, ctx) returns:
// {"Cache.Read": 10.0, "Disk.Read": 2.0}  // Assuming 80% cache hit rate
```

### **Retry Flow (Exponential Pattern):**
```sdl
method ReadWithRetry() Bool {
    for i in 3 {
        let result = self.disk.Read()   // May fail and retry
        if result { return true }
    }
    return false
}
```
```go
// FlowEval("RetryServer", "ReadWithRetry", 5.0, ctx) returns:
// {"Disk.Read": 6.25}  // Expected calls = 1 + 0.3 + 0.09 = 1.39 (assuming 70% success rate)
```

### **Batch Processing:**
```sdl
method ProcessBatch() Bool {
    for item in self.batchSize {  // batchSize = 10
        self.processor.Process(item)
    }
}
```
```go
// FlowEval("BatchProcessor", "ProcessBatch", 5.0, ctx) returns:
// {"Processor.Process": 50.0}  // 5 batches/sec × 10 items/batch = 50 items/sec
```

---

## 🧮 **Mathematical Flow Computation**

### **Core Algorithm:**
```go
func (fc *FlowContext) FlowEval(component, method string, inputRate float64) map[string]float64 {
    outflows := make(map[string]float64)
    
    // Get method definition from SDL AST
    methodDecl := fc.getMethodDecl(component, method)
    
    // Analyze method body for call patterns
    for _, stmt := range methodDecl.Body.Statements {
        flows := fc.analyzeStatement(stmt, inputRate)
        fc.mergeFlows(outflows, flows)
    }
    
    return outflows
}
```

### **Conditional Analysis:**
```go
func (fc *FlowContext) analyzeConditional(ifStmt *IfStmt, inputRate float64) map[string]float64 {
    // Determine condition probability from parameters or heuristics
    conditionProb := fc.evaluateConditionProbability(ifStmt.Condition)
    
    // Analyze both branches with probability-weighted rates
    thenFlows := fc.analyzeStatement(ifStmt.Then, inputRate * conditionProb)
    elseFlows := fc.analyzeStatement(ifStmt.Else, inputRate * (1 - conditionProb))
    
    return fc.mergeFlows(thenFlows, elseFlows)
}
```

### **Retry Analysis with Exponential Limiting:**
```go
func (fc *FlowContext) computeExpectedRetries(successRate float64, maxRetries int) (float64, string) {
    // Option 1: Reasonable cap on max retries to prevent absurd inputs
    warning := ""
    if maxRetries > 50 {
        maxRetries = 50
        warning = "Capped max retries at 50"
    }
    
    // Geometric series: 1 + (1-p) + (1-p)² + ... + (1-p)^(n-1)
    if successRate >= 1.0 {
        return 1.0, warning
    }
    
    failureRate := 1.0 - successRate
    var expectedCalls float64
    
    if maxRetries == 0 {
        expectedCalls = 1.0 / successRate  // Infinite retries
    } else {
        numerator := 1.0 - math.Pow(failureRate, float64(maxRetries))
        expectedCalls = numerator / successRate
    }
    
    // Option 2: Detect pathological results and cap them
    if expectedCalls > 20.0 {
        return 20.0, "Pathological retry pattern detected, capped at 20x"
    }
    
    return expectedCalls, warning
}
```

---

## 🔄 **Cycle Resolution Strategies**

### **Problem: Recursive Calls and Feedback Loops**

**Direct Cycle Example:**
```sdl
component A {
    uses b B
    method Process() Bool {
        if (failed) {
            return self.b.Retry()  // A → B
        }
        return true
    }
}

component B {  
    uses a A
    method Retry() Bool {
        return self.a.Process()  // B → A (cycle!)
    }
}
```

### **Solution: Fixed-Point Iteration (Primary)**
```go
func (fc *FlowContext) SolveSystemFlows(entryPoints map[string]float64) map[string]float64 {
    arrivalRates := make(map[string]float64)
    
    // Initialize with entry points
    for component, rate := range entryPoints {
        arrivalRates[component] = rate
    }
    
    // Iterate until convergence
    for iteration := 0; iteration < fc.MaxIterations; iteration++ {
        oldRates := copyMap(arrivalRates)
        
        // Update success rates based on current loads (affects retry patterns)
        fc.updateSuccessRates(arrivalRates)
        
        // Recompute outgoing flows for each component
        newRates := make(map[string]float64)
        for component, inRate := range arrivalRates {
            outflows := fc.FlowEval(component, "method", inRate)
            for target, outRate := range outflows {
                newRates[target] += outRate
            }
        }
        
        // Check convergence (all rates changed by < 1%)
        maxChange := 0.0
        for component := range arrivalRates {
            change := math.Abs(newRates[component] - oldRates[component])
            if change > maxChange {
                maxChange = change
            }
        }
        
        if maxChange < fc.ConvergenceThreshold {
            return newRates  // Converged!
        }
        
        // Apply damping to prevent oscillation
        for component := range arrivalRates {
            arrivalRates[component] = oldRates[component] + 0.5*(newRates[component] - oldRates[component])
        }
    }
    
    log.Warn("Fixed-point iteration did not converge after %d iterations", fc.MaxIterations)
    return arrivalRates  // Best approximation
}
```

### **Solution: Call Stack Limiting (Fallback)**
```go
func (fc *FlowContext) FlowEvalWithCallStack(component, method string, inputRate float64) map[string]float64 {
    callKey := fmt.Sprintf("%s.%s", component, method)
    
    // Detect direct cycles
    if fc.isInCallStack(callKey) {
        log.Debug("Cycle detected: %s, breaking recursion", callKey)
        return map[string]float64{}  // Break cycle
    }
    
    // Prevent infinite recursion
    if len(fc.CallStack) >= 20 {
        log.Warn("Maximum call depth reached, stopping recursion")
        return map[string]float64{}
    }
    
    // Add to call stack and proceed
    fc.CallStack = append(fc.CallStack, callKey)
    defer func() { fc.CallStack = fc.CallStack[:len(fc.CallStack)-1] }()
    
    return fc.FlowEval(component, method, inputRate)
}
```

### **When to Use Each Approach:**
- **Fixed-Point Iteration:** Retry loops, feedback systems, complex interdependencies
- **Call Stack Limiting:** Rapid prototyping, debugging, when fixed-point fails to converge

---

## 🔧 **Parameter Dependencies**

### **Three Approaches Analyzed:**

#### **Strategy 1: Parameter Snapshot (Recommended for Initial Implementation)**
```go
func (fe *FlowEvaluator) ComputeFlows() map[string]float64 {
    snapshot := fe.freezeParameters()  // {hitRate: 0.8, poolSize: 10, etc.}
    flows := fe.analyzeWithFrozenParams(snapshot)
    return flows
}
```

**Pros:**
- Simple to implement and debug
- Predictable behavior - same inputs always give same outputs
- Fast computation - single pass through the system
- Perfect for demo scenarios where parameters change every 1-5 seconds

**Cons:**
- Ignores feedback loops (e.g., load affecting cache hit rates)
- Less realistic than systems with dynamic parameters

**When to Use:** Initial implementation, educational demos, systems without strong parameter interdependencies

#### **Strategy 2: Dependency Graph (For Specific Relationships)**
```go
type ParameterDependency struct {
    parameter string      // "cache.effectiveHitRate"
    dependsOn []string   // ["cache.load", "cache.configuredHitRate"]  
    computeFunc func(deps map[string]interface{}) interface{}
}

// Example: effectiveHitRate = configuredHitRate * (1.0 - overloadPenalty(load))
```

**When to Use:** When you have specific, well-understood parameter relationships to model

#### **Strategy 3: Multi-Pass Resolution (For Complex Feedback)**
```go
func (fe *FlowEvaluator) ResolveSystem() {
    for iteration := 0; iteration < 5; iteration++ {
        fe.updateParametersBasedOnLoad()    // Hit rates, success rates
        fe.recomputeFlows()                 // Traffic patterns
        if fe.parametersConverged() { break }
    }
}
```

**When to Use:** Complex systems with multiple feedback loops, advanced modeling scenarios

### **Migration Path:**
1. Start with Strategy 1 (Parameter Snapshot)
2. Add Strategy 2 for specific parameter relationships  
3. Upgrade to Strategy 3 when modeling complex feedback systems

---

## 🎭 **Native Component Integration**

### **Problem:**
Native Go components are "black boxes" to FlowEval. We need visibility into their traffic patterns without running simulations.

### **Solution: FlowAnalyzable Interface**
```go
type FlowAnalyzable interface {
    GetFlowPattern(methodName string, inputRate float64, params map[string]interface{}) FlowPattern
}

type FlowPattern struct {
    // Output traffic to other components
    Outflows map[string]float64     // {"disk.Read": 5.0, "cache.Read": 10.0}
    
    // Method behavior characteristics  
    SuccessRate float64             // 0.0 - 1.0
    Amplification float64           // outputRate / inputRate
    
    // Optional: detailed conditional flows
    ConditionalFlows []ConditionalFlow
}

type ConditionalFlow struct {
    Condition string                // "cache_miss", "retry", "batch_processing"
    Probability float64             // 0.0 - 1.0  
    Outflows map[string]float64
}
```

### **Example Implementations:**

#### **ResourcePool (Leaf Component):**
```go
func (rp *ResourcePool) GetFlowPattern(methodName string, inputRate float64, params map[string]interface{}) FlowPattern {
    switch methodName {
    case "Acquire":
        utilization := inputRate * rp.serviceTime.Seconds() / float64(rp.size)
        successRate := computeMMcSuccessRate(utilization)
        
        return FlowPattern{
            Outflows: map[string]float64{},  // ResourcePool is a leaf node
            SuccessRate: successRate,
            Amplification: 1.0,
        }
    }
    return FlowPattern{}
}
```

#### **Cache (Conditional Component):**
```go
func (cache *Cache) GetFlowPattern(methodName string, inputRate float64, params map[string]interface{}) FlowPattern {
    switch methodName {
    case "Read":
        hitRate := cache.hitRate
        
        return FlowPattern{
            Outflows: map[string]float64{
                "disk.Read": inputRate * (1.0 - hitRate),  // Only on cache miss
            },
            SuccessRate: 1.0,
            Amplification: 1.0,
            ConditionalFlows: []ConditionalFlow{
                {
                    Condition: "cache_hit", 
                    Probability: hitRate,
                    Outflows: map[string]float64{},
                },
                {
                    Condition: "cache_miss",
                    Probability: 1.0 - hitRate, 
                    Outflows: map[string]float64{"disk.Read": inputRate},
                },
            },
        }
    }
    return FlowPattern{}
}
```

### **Integration with FlowEvaluator:**
```go
func (fe *FlowEvaluator) analyzeNativeCall(component, method string, inputRate float64) map[string]float64 {
    nativeComp := fe.getNativeComponent(component)
    
    if flowAnalyzable, ok := nativeComp.(FlowAnalyzable); ok {
        pattern := flowAnalyzable.GetFlowPattern(method, inputRate, fe.getComponentParams(component))
        return pattern.Outflows
    }
    
    // Fallback: assume no outflows (leaf component)
    log.Warn("Component %s doesn't implement FlowAnalyzable, assuming leaf node", component)
    return map[string]float64{}
}
```

---

## 🎪 **Workshop Integration**

### **Real-time Dashboard Updates:**
```javascript
// User changes load generation in web UI
apiCall('/api/generate', {target: 'server.Read', rate: 10})

// FlowEval computes system-wide impact instantly:
flowResults = FlowEval("server", "Read", 10.0, context)
// Returns: {"disk.Read": 10.0, "pool.Acquire": 10.0}

// All downstream metrics update immediately:
// - Disk utilization: 50% → 100%  
// - Queue length: 0.5 → ∞ (overloaded!)
// - Wait time: 2ms → ∞
```

### **Parameter Sensitivity Analysis:**
```javascript
// User modifies cache hit rate
apiCall('/api/set', {path: 'cache.HitRate', value: 0.9})

// FlowEval recomputes traffic flows:
oldFlows = {"cache": 10.0, "disk": 2.0}  // 80% hit rate
newFlows = {"cache": 10.0, "disk": 1.0}  // 90% hit rate

// Dashboard shows immediate impact:
// - Disk load drops 50%
// - System-wide latency improves
```

---

## 🚀 **Performance Characteristics**

### **Computational Complexity:**
- **Simple Systems:** O(components × methods) - typically 10-50ms
- **Complex Systems:** O(components² × call_depth) - typically 50-200ms  
- **Cyclic Systems:** O(iterations × components × methods) - typically 100-500ms

### **Target Performance (Workshop Use Case):**
- **Parameter changes:** Every 1-5 seconds (user-paced)
- **Acceptable latency:** < 500ms for real-time feedback
- **Current estimates:** Well within acceptable range for 50-component systems

### **Optimization Strategies:**
```go
type FlowCache struct {
    parameterHash string                           // Hash of current parameters
    cachedFlows   map[string]map[string]float64   // Cached flow results
    dirty         bool                            // Invalidation flag
}

func (fc *FlowCache) GetFlowsIfValid(paramHash string) (map[string]map[string]float64, bool) {
    if fc.parameterHash == paramHash && !fc.dirty {
        return fc.cachedFlows, true
    }
    return nil, false
}
```

---

## 🎯 **Implementation Roadmap**

### **Phase 1: Basic FlowEval (Week 1)**
- [ ] Implement core FlowEval function for SDL components
- [ ] Handle simple patterns: 1:1 calls, conditional flows
- [ ] Basic retry analysis with exponential limiting
- [ ] Parameter snapshot approach

### **Phase 2: Cycle Resolution (Week 2)**  
- [ ] Fixed-point iteration solver
- [ ] Call stack limiting as fallback
- [ ] Convergence detection and damping

### **Phase 3: Native Component Integration (Week 3)**
- [ ] FlowAnalyzable interface definition
- [ ] Update existing native components (ResourcePool, Cache, Disk)
- [ ] Integration with FlowEvaluator

### **Phase 4: Advanced Features (Week 4+)**
- [ ] Parameter dependency tracking
- [ ] Confidence bounds for uncertain parameters
- [ ] Performance optimization and caching
- [ ] Comprehensive testing with Netflix demo scenario

---

**Last Updated:** Conference Workshop Development Phase  
**Next Review:** After Phase 1 implementation completion