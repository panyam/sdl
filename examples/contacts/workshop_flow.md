# Contacts Service Workshop Flow

**Scenario:** Phone number lookup service demonstrating capacity modeling and performance optimization.

---

## 🎪 **Live Demo Flow**

### **Dashboard Layout:**
```
┌─────────────────────────────────────────────────────────────────────────────┐
│ SDL Workshop - Contacts Lookup Service Demo                                │
├─────────────────────────┬─────────────────────┬─────────────────────────────┤
│ System Architecture     │ Current Metrics     │ Command Input               │
│                         │                     │                             │
│ [User Request]          │ Load: 5.0 RPS       │ > set server.pool.          │
│        ↓                │ P95 Latency: 18ms   │   ArrivalRate 15            │
│ ┌─[AppServer]─────┐     │ Success Rate: 99.2% │                             │
│ │ Pool: 10/10     │     │ Server Util: 75%    │ > run high_load server.     │
│ │ Time: 5ms       │     │ DB Util: 60%        │   HandleLookup 1000         │
│ └─────────────────┘     │                     │                             │
│        ↓                │ Cache Hit: 40%      │ Status: ✓ Running...       │
│ ┌─[Database]──────┐     │ DB Connections: 3/5 │                             │
│ │ Pool: 5/5       │     │                     │                             │
│ │ Cache: 40%      │───┐ │                     │                             │
│ │ Query: 10ms     │   │ │                     │                             │
│ └─────────────────┘   │ │                     │                             │
│        ↓              │ │                     │                             │
│ ┌─[HashIndex]─────┐   │ │                     │                             │
│ │ Lookup: 8ms     │←──┘ │                     │                             │
│ └─────────────────┘     │                     │                             │
├─────────────────────────┴─────────────────────┴─────────────────────────────┤
│ Live Performance Chart                                                      │
│ 50ms ┤                                                                     │
│ 40ms ┤                   ┌─high_load─┐                                     │
│ 30ms ┤                   │           │                                     │
│ 20ms ┤         ┌─────────┘           │                                     │
│ 10ms ┤────┬────┘                     └─┬─cache_opt─┬─scaled──              │
│  0ms └────┴─baseline──────────────────────┴─────────┴─────────────────────  │
│      [0s────15s────30s────45s────60s────75s────90s]                       │
└─────────────────────────────────────────────────────────────────────────────┘
```

### **Demo Narrative:**

#### **Phase 1: "Here's a simple phone lookup service"**
```bash
# Start with baseline
canvas.Set("server.pool.ArrivalRate", 5.0)
canvas.Run("baseline", "server.HandleLookup", 1000)
```
**Shows:** 18ms latency, 99% success rate, everything healthy

#### **Phase 2: "What happens during peak usage?"**
```bash
# 3x load increase
canvas.Set("server.pool.ArrivalRate", 15.0)
canvas.Run("high_load", "server.HandleLookup", 1000)
```
**Shows:** 45ms latency, 85% success rate - system struggling!

#### **Phase 3: "Let's optimize the cache"**
```bash
# Improve cache hit rate
canvas.Set("server.db.CacheHitRate", 0.8)
canvas.Run("cache_opt", "server.HandleLookup", 1000)
```
**Shows:** 25ms latency, 95% success rate - much better!

#### **Phase 4: "What if we scale the server?"**
```bash
# Double server capacity
canvas.Set("server.pool.Size", 20)
canvas.Run("scaled", "server.HandleLookup", 1000)
```
**Shows:** 20ms latency, 99% success rate - back to healthy

#### **Phase 5: "But what's the real bottleneck?"**
```bash
# Push database to limits
canvas.Set("server.db.pool.ArrivalRate", 20.0)
canvas.Run("db_limit", "server.HandleLookup", 1000)
```
**Shows:** Database becomes the constraint - different failure pattern

---

## 💡 **Key Teaching Moments**

### **Moment 1: Capacity Reality**
- **Before:** "Just add more servers to handle load"
- **Demo:** Server scaling helps, but database becomes bottleneck
- **Learning:** "Oh! You can't just scale one component"

### **Moment 2: Cache Impact**
- **Before:** "Cache hit rate is just a nice optimization"
- **Demo:** 40% → 80% hit rate dramatically improves performance
- **Learning:** "Cache effectiveness is critical for system health"

### **Moment 3: Queueing Visualization**
- **Before:** Abstract understanding of queuing theory
- **Demo:** Visual representation of pool utilization and wait times
- **Learning:** "I can see how M/M/c queuing works in practice"

### **Moment 4: System Interactions**
- **Before:** "Components are independent"
- **Demo:** AppServer load affects Database load affects overall performance
- **Learning:** "System design is about understanding interactions"

---

## 🎯 **Workshop Validation Goals**

### **Canvas API Validation:**
1. **Parameter Modification:** ✓ Set arrival rates, pool sizes, cache hit rates
2. **Simulation Execution:** ✓ Run scenarios with different loads
3. **Visualization Generation:** ✓ Latency plots, comparison charts, architecture diagrams
4. **Rapid Iteration:** ✓ Quick parameter changes for live demos
5. **Edge Case Handling:** ✓ Zero capacity, perfect cache, overload conditions

### **Flow Patterns Demonstrated:**
1. **Simple Flow:** User → AppServer → Database → Index
2. **Conditional Flow:** Cache hit vs miss patterns  
3. **Resource Constraints:** Pool exhaustion and queuing
4. **Performance Degradation:** Gradual vs cliff-edge failures

### **Educational Value:**
1. **Relatable Scenario:** Everyone understands phone number lookup
2. **Clear Metrics:** Latency, success rate, utilization easy to understand
3. **Visual Impact:** Charts show dramatic performance changes
4. **Practical Application:** "I can use this for system design interviews"

---

## 🔧 **Technical Requirements Validated**

### **SDL Language Features:**
- Component composition (AppServer uses Database)
- Parameter modification (pool sizes, hit rates)
- Resource modeling (ResourcePool components)  
- Method calls with return types
- Probabilistic modeling (cache hit distributions)

### **Canvas API Features:**
- Load/Use SDL files
- Set component parameters at runtime
- Run simulations with result storage
- Generate plots and diagrams
- Handle parameter validation and edge cases

### **Performance Requirements:**
- Sub-second response for parameter changes
- Reliable simulation execution
- Clean error handling for invalid scenarios
- Consistent results for repeated runs

**This simple contacts service provides the perfect foundation for validating our Canvas API before building the full workshop tooling!** 📱