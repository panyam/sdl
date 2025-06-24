# SDL Core Concepts and Design Philosophy

## Table of Contents

1. [What is SDL?](#what-is-sdl)
2. [Core Philosophy](#core-philosophy)
3. [Key Concepts](#key-concepts)
4. [Modeling Approach](#modeling-approach)
5. [Simulation Engine](#simulation-engine)
6. [Performance Analysis](#performance-analysis)
7. [Design Patterns](#design-patterns)
8. [Best Practices](#best-practices)

## What is SDL?

SDL (System Design Language) is a specialized language for modeling the **performance characteristics** of distributed systems. Unlike general-purpose programming languages that focus on implementation, SDL focuses on:

- **Capacity modeling**: How much load can a system handle?
- **Performance prediction**: What will latencies look like under load?
- **Bottleneck identification**: Where will the system break first?
- **Failure analysis**: How do failures cascade through the system?

### SDL is NOT:

- A programming language for building systems
- A configuration language for deployment
- A testing framework for existing code
- A monitoring or observability tool

### SDL IS:

- A modeling language for system behavior
- A simulation framework for performance analysis
- A tool for capacity planning and sizing
- A way to validate architectural decisions before building

## Core Philosophy

### 1. Simplicity Through Constraints

SDL intentionally limits features to maintain focus on system modeling:

```sdl
// No complex arithmetic - use native functions
native method plus(a Int, b Int) Int

// No runtime parameters - methods model behaviors
method GetUser() Bool {
    // Models the behavior, not the implementation
}
```

### 2. Probabilistic by Design

Real systems have uncertainty. SDL embraces this:

```sdl
// Model variable outcomes
method ProcessRequest() Bool {
    return sample dist {
        95 => true,   // 95% success
        5 => false    // 5% failure
    }
}

// Model variable latency
method QueryDatabase() Bool {
    let latency = sample dist {
        70 => 5ms,    // Fast queries
        25 => 50ms,   // Normal queries
        5 => 500ms    // Slow queries
    }
    delay(latency)
    return true
}
```

### 3. Composition Over Complexity

Systems are built from simple, composable components:

```sdl
component Cache {
    param HitRate Float = 0.8
}

component Database {
    uses pool ResourcePool(Size = 10)
}

component Service {
    uses cache Cache
    uses db Database
    
    // Compose behaviors
    method HandleRequest() Bool {
        if self.cache.Read() {
            return true
        }
        return self.db.Query()
    }
}
```

## Key Concepts

### 1. Components

Components are the building blocks that model system elements:

```sdl
component WebServer {
    // Parameters define configuration
    param MaxConnections Int = 1000
    param RequestTimeout Duration = 30s
    
    // Dependencies model relationships
    uses backend APIServer
    uses cache RedisCache
    
    // Methods model behaviors
    method ServeRequest() Bool {
        // Implementation
    }
}
```

**Key Points:**
- Components represent services, databases, caches, queues, etc.
- Parameters configure behavior without code changes
- Dependencies create system topology
- Methods define interaction patterns

### 2. Systems

Systems compose components into architectures:

```sdl
system EcommerceSystem {
    // Instantiate components with specific configurations
    use web WebServer(MaxConnections = 5000)
    use api APIServer(
        db = postgres,
        cache = redis
    )
    use postgres Database(ConnectionPoolSize = 100)
    use redis Cache(HitRate = 0.9)
}
```

**Key Points:**
- Systems define complete architectures
- Component instances can be configured
- Dependencies are wired explicitly
- Multiple systems can model different scenarios

### 3. Probabilistic Outcomes

SDL models uncertainty through distributions:

```sdl
// Simple probability
let success = sample dist {
    99 => true,
    1 => false
}

// Weighted outcomes with explicit total
let response = sample dist 100 {
    80 => "success",
    15 => "retry",
    5 => "error"
}

// Model varying behavior
method GetFromCache() Outcomes[Bool] {
    return dist {
        self.HitRate * 100 => true,
        (1 - self.HitRate) * 100 => false
    }
}
```

### 4. Flow Control

SDL supports limited, purposeful control flow:

```sdl
// Conditional logic
if cacheHit {
    return true
} else {
    return self.db.Query()
}

// Simple loops for retry logic
let attempts = 0
for attempts < 3 {
    if self.service.Call() {
        return true
    }
    attempts = attempts + 1
    delay(100ms)
}

// Switch for multiple outcomes
switch responseCode {
    200 => return true
    429 => delay(1s)
    500 => return false
    default => return false
}
```

### 5. Virtual Time

SDL simulations run in virtual time, enabling:

```sdl
method ProcessBatch() Bool {
    // These happen in virtual time
    delay(100ms)  // Processing time
    
    go {
        delay(1s)
        self.notifier.Send()
    }
    
    return true
}
```

**Benefits:**
- Deterministic simulations
- Fast execution (1 hour simulated in milliseconds)
- Reproducible results
- Time-travel debugging

## Modeling Approach

### 1. Start with the Architecture

Define your system structure:

```sdl
system UserService {
    use api APIGateway
    use auth AuthService
    use users UserDatabase
    use cache UserCache
}
```

### 2. Model Key Behaviors

Focus on performance-critical paths:

```sdl
component APIGateway {
    uses auth AuthService
    uses users UserDatabase
    uses cache UserCache
    
    method GetUser() Bool {
        // Check auth first
        if !self.auth.Validate() {
            return false
        }
        
        // Try cache
        let cached = self.cache.Get()
        if cached {
            return true
        }
        
        // Fall back to database
        return self.users.Lookup()
    }
}
```

### 3. Add Realistic Parameters

Use real-world measurements:

```sdl
component UserDatabase {
    // Based on actual measurements
    param QueryLatency = dist {
        70 => 5ms,
        25 => 20ms,
        5 => 100ms
    }
    
    param ConnectionPoolSize = 50
    param FailureRate = 0.001  // 0.1% failure rate
}
```

### 4. Simulate Under Load

Use traffic generators to test scenarios:

```sdl
// Test normal load
sdl gen add normal api.GetUser 100  // 100 RPS

// Test peak load
sdl gen add peak api.GetUser 1000   // 1000 RPS

// Test failure scenarios
sdl set auth.FailureRate 0.1        // 10% auth failures
```

## Simulation Engine

### 1. Execution Model

SDL uses discrete event simulation:

1. **Event Queue**: All method calls become events
2. **Virtual Clock**: Time advances to next event
3. **Deterministic**: Same inputs produce same outputs
4. **Concurrent**: Models parallel execution

### 2. Queueing Theory

Built-in components use proper queueing models:

```sdl
component DatabasePool {
    uses pool ResourcePool(
        Size = 20,              // Number of connections
        ArrivalRate = 100.0,    // Requests per second
        AvgHoldTime = 50ms      // Average query time
    )
    
    method Query() Bool {
        // Uses M/M/c queueing model
        return self.pool.Acquire()
    }
}
```

### 3. Flow Analysis

SDL automatically calculates flow rates:

```sdl
// With 100 RPS at entry point and 80% cache hit rate:
// - Cache sees 100 RPS
// - Database sees 20 RPS (cache misses)
// - Notification service sees 100 RPS

component Service {
    uses cache Cache(HitRate = 0.8)
    uses db Database
    uses notifier NotificationService
    
    method Process() Bool {
        let hit = self.cache.Get()
        if !hit {
            self.db.Query()
        }
        self.notifier.Send()
        return true
    }
}
```

## Performance Analysis

### 1. Latency Analysis

Understand response time distributions:

```sdl
// Model realistic latency
component ServiceWithLatency {
    method Process() Bool {
        // Network latency
        delay(5ms)
        
        // Processing time varies with load
        let utilization = self.GetUtilization()
        if utilization > 0.8 {
            delay(100ms)  // Degraded performance
        } else {
            delay(10ms)   // Normal performance
        }
        
        return true
    }
}
```

### 2. Capacity Analysis

Find system limits:

```sdl
// ResourcePool automatically models capacity
component ConnectionLimitedService {
    uses pool ResourcePool(Size = 100)
    
    method Handle() Bool {
        let conn = self.pool.Acquire()
        if !conn {
            return false  // Rejected due to capacity
        }
        
        delay(50ms)  // Hold connection
        return true
    }
}
```

### 3. Failure Analysis

Model cascading failures:

```sdl
component ServiceWithDependency {
    uses critical CriticalService
    uses fallback FallbackService
    
    method Process() Bool {
        let result = self.critical.Call()
        if !result {
            // Critical service failed, try fallback
            return self.fallback.Call()
        }
        return true
    }
}
```

## Design Patterns

### 1. Cache-Aside Pattern

```sdl
component CacheAsideService {
    uses cache Cache
    uses db Database
    
    method Get(key String) Bool {
        // Check cache first
        let value = self.cache.Get(key)
        if value {
            return true
        }
        
        // Cache miss - load from database
        let data = self.db.Query(key)
        if data {
            self.cache.Set(key, data)
        }
        return data
    }
}
```

### 2. Circuit Breaker (Simplified)

```sdl
component CircuitBreakerService {
    param FailureThreshold Float = 0.5
    param IsOpen Bool = false
    
    uses backend BackendService
    
    method Call() Bool {
        if self.IsOpen {
            return false  // Fast fail
        }
        
        let result = self.backend.Process()
        // In real implementation, track failures
        // and open circuit if threshold exceeded
        return result
    }
}
```

### 3. Retry with Backoff

```sdl
component RetryService {
    uses unreliable UnreliableService
    
    method CallWithRetry() Bool {
        let attempts = 0
        let backoff = 100ms
        
        for attempts < 3 {
            let result = self.unreliable.Call()
            if result {
                return true
            }
            
            delay(backoff)
            backoff = backoff * 2  // Exponential backoff
            attempts = attempts + 1
        }
        
        return false
    }
}
```

### 4. Load Balancing

```sdl
component LoadBalancer {
    uses servers List[Server]
    
    method Route() Bool {
        // Simple round-robin
        let index = sample dist {
            33 => 0,
            33 => 1,
            34 => 2
        }
        
        return self.servers[index].Handle()
    }
}
```

## Best Practices

### 1. Model What Matters

Focus on performance-critical paths:

```sdl
// Good: Model the slow database query
method GetUserProfile() Bool {
    delay(50ms)  // Known slow query
    return sample dist { 99 => true, 1 => false }
}

// Avoid: Don't model trivial operations
method ValidateInteger() Bool {
    // Don't model CPU-bound operations unless they matter
    return true
}
```

### 2. Use Realistic Parameters

Base parameters on actual measurements:

```sdl
component RealisticDatabase {
    // Bad: Made-up numbers
    // param Latency = 10ms
    
    // Good: Based on P50/P95/P99 measurements
    param Latency = dist {
        50 => 5ms,    // P50
        45 => 20ms,   // P95  
        5 => 100ms    // P99
    }
}
```

### 3. Start Simple, Add Complexity

Begin with basic models and refine:

```sdl
// Version 1: Simple model
component SimpleCache {
    param HitRate = 0.8
    
    method Get() Bool {
        return sample dist {
            80 => true,
            20 => false
        }
    }
}

// Version 2: Add capacity constraints
component BetterCache {
    param HitRate = 0.8
    uses pool ResourcePool(Size = 1000)  // Connection limit
    
    method Get() Bool {
        if !self.pool.Acquire() {
            return false  // Overloaded
        }
        
        return sample dist {
            80 => true,
            20 => false
        }
    }
}
```

### 4. Validate Against Reality

Compare simulation results with production:

1. Start with known scenarios
2. Tune parameters to match reality
3. Use model to predict new scenarios
4. Validate predictions against outcomes

### 5. Document Assumptions

Make modeling decisions explicit:

```sdl
component PaymentService {
    // Assumption: Payment provider SLA guarantees 99.9% availability
    param AvailabilityRate = 0.999
    
    // Assumption: Timeout set to P99 latency + buffer
    param Timeout = 1s
    
    method ProcessPayment() Bool {
        // Model based on provider's documented latencies
        delay(sample dist {
            90 => 100ms,  // Normal
            9 => 500ms,   // Slow
            1 => 1s       // Timeout
        })
        
        return sample dist {
            999 => true,  // Success (99.9%)
            1 => false    // Failure (0.1%)
        }
    }
}
```

## Summary

SDL provides a powerful yet simple way to model distributed system performance. By focusing on key behaviors, realistic parameters, and probabilistic outcomes, you can:

- Predict system behavior before building
- Identify bottlenecks early
- Validate architectural decisions
- Plan capacity with confidence

Remember: SDL models **what** systems do, not **how** they do it. Keep models simple, focused, and grounded in reality.