# SDL Language Reference

## Table of Contents
1. [Introduction](#introduction)
2. [Language Philosophy](#language-philosophy)
3. [Basic Syntax](#basic-syntax)
4. [Type System](#type-system)
5. [Components](#components)
6. [Systems](#systems)
7. [Methods](#methods)
8. [Statements](#statements)
9. [Expressions](#expressions)
10. [Probabilistic Modeling](#probabilistic-modeling)
11. [Import System](#import-system)
12. [Native Extensions](#native-extensions)
13. [Common Patterns](#common-patterns)
14. [Language Limitations](#language-limitations)

## Introduction

SDL (System Design Language) is a domain-specific language for modeling and simulating distributed systems. It focuses on capturing system behavior, capacity constraints, and failure modes rather than implementation details.

## Language Philosophy

SDL is intentionally simple and limited. It is **not** a general-purpose programming language. Instead, it's designed to:

- Model system capacities and performance characteristics
- Simulate distributed system behavior under load
- Analyze bottlenecks and failure modes
- Predict system behavior without implementation

Key principles:
- **Simplicity over features**: Limited operators and control flow
- **Probabilistic by design**: Built-in support for uncertainty
- **Performance focused**: Models latency, throughput, and capacity
- **Declarative**: Describes what systems do, not how

## Basic Syntax

### Comments
```sdl
// Single-line comment

/* 
   Multi-line comment
   Can span multiple lines
*/
```

### Identifiers
- Must start with a letter or underscore
- Can contain letters, digits, and underscores
- Case-sensitive

Valid identifiers: `myComponent`, `_internal`, `Service2`, `MAX_CONNECTIONS`

### Literals
```sdl
// Integers
42
1000
-15

// Floats
3.14
100.0
-0.5

// Strings
"Hello, World!"
"Error message"

// Booleans
true
false

// Durations
10ms    // 10 milliseconds
1s      // 1 second
100us   // 100 microseconds
5m      // 5 minutes
2h      // 2 hours
```

## Type System

### Primitive Types

| Type | Description | Example |
|------|-------------|---------|
| `Int` | Integer values | `42` |
| `Float` | Floating-point values | `3.14` |
| `String` | Text values | `"hello"` |
| `Bool` | Boolean values | `true`, `false` |
| `Duration` | Time durations | `100ms`, `1s` |

### Complex Types

#### Lists
```sdl
List[Int]     // List of integers
List[String]  // List of strings
List[Bool]    // List of booleans
```

#### Tuples
```sdl
(Int, String)           // 2-tuple
(Bool, Int, String)     // 3-tuple
```

#### Outcomes
```sdl
Outcomes[Bool]          // Probabilistic boolean outcome
Outcomes[Int]           // Probabilistic integer outcome
Outcomes[(Bool, Int)]   // Probabilistic tuple outcome
```

## Components

Components are the building blocks of SDL systems. They model individual services, databases, caches, or any system element.

### Basic Component
```sdl
component UserService {
    // Component body
}
```

### Native Component
Native components are implemented in the host language (Go):
```sdl
native component Database {
    param ConnectionPoolSize Int = 10
    param QueryTimeout Duration = 100ms
}
```

### Component Parameters
```sdl
component Cache {
    param Capacity Int = 1000
    param TTL Duration = 5m
    param HitRate Float = 0.8
}
```

### Component Dependencies
```sdl
component AppServer {
    uses db Database
    uses cache Cache
    
    method HandleRequest() Bool {
        // Method implementation
    }
}
```

### Parameterized Dependencies
```sdl
component WebServer {
    uses db Database(ConnectionPoolSize = 50)
    uses cache Cache(Capacity = 10000, HitRate = 0.9)
}
```

## Systems

Systems compose components into complete architectures.

### Basic System
```sdl
system MySystem {
    use web WebServer
    use db Database
}
```

### System with Configuration
```sdl
system EcommerceSystem {
    use frontend WebServer(MaxConnections = 1000)
    use api APIServer(db = database, cache = redis)
    use database Database(ConnectionPoolSize = 100)
    use redis Cache(Capacity = 50000)
}
```

## Methods

Methods define component behavior and interactions.

### Basic Method
```sdl
method ProcessRequest() {
    // Method body
}
```

### Method with Return Type
```sdl
method CheckHealth() Bool {
    return true
}
```

### Method with Complex Return
```sdl
method LookupUser() (Bool, String) {
    // Returns tuple (success, userId)
    return (true, "user123")
}
```

### Method Parameters
**Important**: In SDL, methods model behaviors, not function calls. They don't take runtime parameters:
```sdl
// SDL methods model behaviors
method GetUser() Bool {
    // No parameters - models the behavior of getting a user
}

// NOT: method GetUser(userId String) - SDL doesn't support this
```

## Statements

### Let Statement (Variable Declaration)
```sdl
let result = self.db.Query()
let success, data = self.api.Fetch()  // Tuple unpacking
```

### If Statement
```sdl
if success {
    self.cache.Store(data)
} else {
    self.logger.Error("Fetch failed")
}

// With else-if
if load > 0.9 {
    return false  // Reject request
} else if load > 0.7 {
    delay(10ms)   // Add backpressure
    return true
} else {
    return true   // Normal processing
}
```

### Return Statement
```sdl
return true
return (success, data)
return  // Void return
```

### For Loop
SDL supports simple condition-based loops:
```sdl
for retries < 3 {
    let success = self.db.Connect()
    if success {
        return true
    }
    retries = retries + 1
}
```

### Switch Statement
```sdl
switch statusCode {
    200 => return true
    404 => return false
    500 => self.logger.Error("Server error")
    default => return false
}
```

## Expressions

### Member Access
```sdl
self.cache.Get(key)           // Access component dependency
result.success                // Access struct field
```

### Method Calls
```sdl
// No parameters (SDL pattern)
self.db.Query()
self.cache.Invalidate()

// Native functions can have parameters
delay(100ms)
log("Processing request")
```

### Distributions (Probabilistic Values)
```sdl
// Simple distribution
dist {
    80 => true,   // 80% chance
    20 => false   // 20% chance
}

// Distribution with explicit total
dist 100 {
    70 => "success",
    20 => "retry",
    10 => "failure"
}
```

### Sample Expression
```sdl
// Sample from a distribution
let outcome = sample dist {
    90 => true,
    10 => false
}
```

### Asynchronous Execution

#### Simple Async Operations
```sdl
// Fire and forget
go self.notificationService.Send()

// Async with block
go {
    delay(1s)
    self.cleanup.Run()
}

// Store future for later
let future = go self.slowOperation()
// ... do other work ...
let result = wait future
```

#### Batch Processing
```sdl
// Execute N operations concurrently
gobatch 10 {
    self.worker.Process()
}

// Batch with variable concurrency
component BatchProcessor {
    param Concurrency Int = 5
    
    method ProcessBatch() Bool {
        gobatch self.Concurrency {
            let item = self.queue.Dequeue()
            self.processItem(item)
        }
        return true
    }
}
```

#### Complex Async Patterns
```sdl
component ParallelService {
    uses services List[Service]
    
    method FanOut() Bool {
        // Launch all requests in parallel
        let futures = []
        for i < len(self.services) {
            let f = go self.services[i].Process()
            futures = append(futures, f)
        }
        
        // Wait for all to complete
        let results = wait futures using WaitAll()
        
        // Check if all succeeded
        for result in results {
            if !result {
                return false
            }
        }
        return true
    }
    
    method FirstToRespond() Bool {
        // Launch competing requests
        let f1 = go self.primary.Query()
        let f2 = go self.secondary.Query()
        let f3 = go self.tertiary.Query()
        
        // Return first successful response
        return wait f1, f2, f3 using WaitAny()
    }
}
```

### Wait Expression
```sdl
// Wait for single future
let future = go self.slowOperation()
let result = wait future

// Wait for multiple futures
let f1 = go self.service1.Call()
let f2 = go self.service2.Call()
let results = wait f1, f2 using WaitAll()
```

## Probabilistic Modeling

SDL's strength lies in modeling uncertainty and probabilistic behavior.

### Understanding Distributions

#### Basic Distribution Syntax
```sdl
// Simple distribution - weights are relative
dist {
    80 => true,   // 80% probability
    20 => false   // 20% probability
}

// Distribution with explicit total
dist 100 {
    75 => "success",
    20 => "retry",
    5 => "failure"
}

// Distribution with variables
let successRate = 0.95
dist {
    successRate * 100 => true,
    (1 - successRate) * 100 => false
}
```

### Modeling Failures
```sdl
method QueryDatabase() Bool {
    // Model 1% failure rate
    return sample dist {
        99 => true,
        1 => false
    }
}

// More complex failure modeling
method ResilientQuery() Bool {
    let result = sample dist {
        95 => "success",
        3 => "timeout",
        2 => "error"
    }
    
    switch result {
        "success" => return true
        "timeout" => {
            delay(1s)
            return false
        }
        "error" => return false
        default => return false
    }
}
```

### Modeling Variable Latency
```sdl
method ProcessRequest() Bool {
    // Model variable processing time
    let processingTime = sample dist {
        70 => 10ms,   // Fast path
        25 => 50ms,   // Normal path
        5 => 500ms    // Slow path
    }
    delay(processingTime)
    return true
}

// Latency that degrades with load
component AdaptiveService {
    param Load Float = 0.5
    
    method Process() Bool {
        let baseLatency = 10ms
        let loadMultiplier = 1 + (self.Load * 4)  // Up to 5x slower
        
        delay(baseLatency * loadMultiplier)
        return true
    }
}
```

### Modeling Cache Behavior
```sdl
component CacheLayer {
    param HitRate Float = 0.8
    param Capacity Int = 10000
    
    method Get() Bool {
        return sample dist 100 {
            self.HitRate * 100 => true,   // Cache hit
            (1 - self.HitRate) * 100 => false   // Cache miss
        }
    }
    
    // Cache with degradation
    method GetWithDegradation() Bool {
        // Hit rate decreases as cache fills
        let effectiveHitRate = self.HitRate * 0.9  // 90% effectiveness
        
        return sample dist {
            effectiveHitRate * 100 => true,
            (1 - effectiveHitRate) * 100 => false
        }
    }
}
```

### Modeling Complex Behaviors
```sdl
component NetworkService {
    param PacketLossRate Float = 0.001
    param LatencyDistribution = dist {
        60 => 10ms,   // Low latency
        30 => 50ms,   // Medium latency
        10 => 200ms   // High latency
    }
    
    method Send() Bool {
        // Model packet loss
        let lost = sample dist {
            self.PacketLossRate * 1000 => true,
            (1 - self.PacketLossRate) * 1000 => false
        }
        
        if lost {
            return false
        }
        
        // Model network latency
        let latency = sample self.LatencyDistribution
        delay(latency)
        
        return true
    }
}
```

## Import System

### Basic Import
```sdl
import Database from "./database.sdl"
import Cache, LoadBalancer from "./infrastructure.sdl"
```

### Import with Aliases
```sdl
import Database as DB from "./database.sdl"
import Cache as RedisCache from "./cache.sdl"
```

### Using Imported Components
```sdl
system MySystem {
    use db DB(ConnectionPoolSize = 50)
    use cache RedisCache(Capacity = 10000)
}
```

## Native Extensions

### Native Methods
Declare methods implemented in the host language:
```sdl
native method Calculate(x Int, y Int) Int
native method Random() Float
```

### Native Aggregators
For custom wait strategies:
```sdl
native aggregator Quorum(results List[Bool], required Int) Bool
native aggregator FirstSuccess(results List[Outcomes[Bool]]) Bool
```

## Common Patterns

### Service with Cache
```sdl
component ServiceWithCache {
    uses cache Cache
    uses db Database
    
    method GetData() Bool {
        let cached = self.cache.Get()
        if cached {
            return true
        }
        
        let result = self.db.Query()
        if result {
            self.cache.Store()
        }
        return result
    }
}

// More sophisticated cache-aside pattern
component CacheAsideService {
    uses cache Cache(HitRate = 0.85)
    uses db Database
    uses metrics MetricsCollector
    
    method Get() Bool {
        // Try cache first
        let hit = self.cache.Read()
        if hit {
            self.metrics.RecordCacheHit()
            return true
        }
        
        // Cache miss - record metric
        self.metrics.RecordCacheMiss()
        
        // Load from database
        let data = self.db.Query()
        if data {
            // Update cache for next time
            self.cache.Write()
            return true
        }
        
        return false
    }
}
```

### Retry Pattern
```sdl
method ReliableOperation() Bool {
    let attempts = 0
    for attempts < 3 {
        let success = self.service.Call()
        if success {
            return true
        }
        attempts = attempts + 1
        delay(100ms)
    }
    return false
}
```

### Circuit Breaker (Simplified)
```sdl
component ServiceWithBreaker {
    param FailureThreshold Float = 0.5
    param IsOpen Bool = false
    
    method Call() Bool {
        if self.IsOpen {
            return false  // Fast fail
        }
        
        let result = self.backend.Process()
        // In real implementation, track failures
        return result
    }
}
```

## Language Limitations

SDL has intentional limitations to maintain simplicity:

### No Binary Operators
SDL doesn't support `+`, `-`, `*`, `/`, `%`. Use native functions if needed:
```sdl
// Instead of: x + y
native method plus(a Int, b Int) Int

// Instead of: x * y  
native method multiply(a Int, b Int) Int
```

### No Runtime Method Parameters
Methods model behaviors, not function calls:
```sdl
// SDL way: Methods represent behaviors
method GetUser() Bool

// NOT supported: method GetUser(userId String)
```

### No Complex Data Structures
SDL doesn't support:
- Custom structs/classes
- Maps/dictionaries (except as native components)
- Complex data manipulation

### No State Mutation
SDL models steady-state behavior:
- No global variables
- Limited local state within methods
- Focus on probabilistic outcomes, not state machines

## Best Practices

1. **Keep Components Focused**: Each component should model one system element
2. **Use Meaningful Names**: Components and methods should clearly indicate their purpose
3. **Model Realistic Behavior**: Use actual performance data for distributions
4. **Document Assumptions**: Use comments to explain modeling decisions
5. **Test Edge Cases**: Model both normal and failure scenarios

## Example: Complete System

```sdl
// common.sdl - Shared components
native component Database {
    param ConnectionPoolSize Int = 10
    param QueryLatency Duration = 5ms
}

native component Cache {
    param HitRate Float = 0.8
    param Capacity Int = 1000
}

// service.sdl - Application service
import Database, Cache from "./common.sdl"

component APIService {
    uses cache Cache
    uses db Database
    
    method HandleRequest() Bool {
        // Check cache first
        let hit = sample dist 100 {
            80 => true,   // Cache hit rate
            20 => false
        }
        
        if hit {
            delay(1ms)  // Cache lookup time
            return true
        }
        
        // Cache miss - query database
        delay(self.db.QueryLatency)
        return sample dist {
            99 => true,   // 99% success rate
            1 => false    // 1% failure rate
        }
    }
}

// system.sdl - Complete system
system ProductionSystem {
    use api APIService(
        cache = Cache(HitRate = 0.85, Capacity = 50000),
        db = Database(ConnectionPoolSize = 100)
    )
}
```

This reference covers the complete SDL language syntax and semantics. For specific component documentation and usage examples, see the [Component Library Reference](./COMPONENT_LIBRARY.md).