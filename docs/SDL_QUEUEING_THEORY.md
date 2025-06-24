# Queueing Theory in SDL

## Introduction

SDL's power comes from its foundation in queueing theory - the mathematical study of waiting lines. This document explains how SDL uses queueing theory to model real distributed systems accurately.

## Why Queueing Theory Matters

In distributed systems, components don't have infinite capacity. When requests arrive faster than they can be processed:
- Queues form
- Wait times increase  
- Systems eventually overflow
- Performance degrades non-linearly

SDL models these effects automatically using proven mathematical models.

## Basic Concepts

### The Universal Queue Model

Every system component can be modeled as:
```
Arrivals → Queue → Servers → Departures
```

In SDL terms:
```sdl
component ServiceWithQueue {
    param ArrivalRate Float = 100      // λ (lambda) - requests per second
    param ServiceTime Duration = 10ms   // 1/μ (mu) - time per request  
    param NumServers Int = 5            // c - number of servers
    param QueueCapacity Int = 100       // K - max queue size
}
```

### Key Metrics

1. **Utilization (ρ)**: How busy the system is
   - ρ = λ / (c × μ)
   - ρ < 1: Stable system
   - ρ ≥ 1: Unstable (queue grows forever)

2. **Queue Length (L)**: Average number in system
3. **Wait Time (W)**: Average time in system
4. **Service Time (S)**: Time being served

## SDL's Queue Models

### 1. M/M/1 Queue (Single Server)

The simplest queue: one server, exponential arrivals and service times.

```sdl
component SingleServerQueue {
    uses queue MM1Queue
    
    param ArrivalRate Float = 10       // 10 requests/second
    param ServiceTime Duration = 80ms   // 80ms per request
    
    method Process() Bool {
        // Utilization = 10 × 0.08 = 0.8 (80%)
        // Average wait time ≈ 320ms
        return self.queue.Enqueue()
    }
}
```

**Key Properties:**
- Wait time increases dramatically as utilization approaches 100%
- At 50% utilization: wait ≈ service time
- At 80% utilization: wait ≈ 4 × service time
- At 90% utilization: wait ≈ 9 × service time

### 2. M/M/c Queue (Multiple Servers)

Models thread pools, connection pools, and parallel processors.

```sdl
component ConnectionPool {
    uses pool ResourcePool(
        Size = 20,              // 20 connections
        ArrivalRate = 100.0,    // 100 requests/second
        AvgHoldTime = 150ms     // 150ms per query
    )
    
    method Query() Bool {
        // Utilization = 100 × 0.15 / 20 = 0.75 (75%)
        let conn = self.pool.Acquire()
        if !conn {
            return false  // Pool exhausted
        }
        
        delay(150ms)  // Simulate query
        return true
    }
}
```

**Erlang-C Formula** (used internally):
- Probability of waiting = C(c, a) where a = λ/μ
- Average wait time = C(c, a) / (c × μ - λ)

### 3. M/M/c/K Queue (Finite Capacity)

Models real systems with limits.

```sdl
component BoundedQueue {
    uses queue MMCKQueue
    
    param NumServers Int = 5
    param QueueCapacity Int = 50
    param ArrivalRate Float = 100
    param ServiceTime Duration = 40ms
    
    method Handle() Bool {
        let queued = self.queue.Enqueue()
        if !queued {
            // Rejected - queue full
            return false
        }
        return true
    }
}
```

**Key Behaviors:**
- Rejects requests when full
- Provides backpressure
- Prevents system overload

## Real-World Examples

### Example 1: Database Connection Pool

```sdl
component DatabaseService {
    uses pool ResourcePool(
        Size = 30,               // 30 connections
        ArrivalRate = 200.0,     // 200 queries/second
        AvgHoldTime = 100ms      // 100ms average query
    )
    
    method ExecuteQuery() Bool {
        // Offered load = 200 × 0.1 = 20 Erlangs
        // With 30 servers, utilization = 20/30 = 67%
        // Low wait times expected
        
        let conn = self.pool.Acquire()
        if !conn {
            // All connections busy
            return false
        }
        
        // Model actual query time distribution
        let queryTime = sample dist {
            60 => 50ms,    // Fast queries
            30 => 100ms,   // Normal queries
            10 => 500ms    // Slow queries
        }
        
        delay(queryTime)
        return true
    }
}
```

### Example 2: API Rate Limiting

```sdl
component RateLimitedAPI {
    uses limiter ResourcePool(
        Size = 100,              // 100 concurrent requests
        ArrivalRate = 500.0,     // 500 requests/second peak
        AvgHoldTime = 150ms      // 150ms average response
    )
    
    method HandleRequest() Bool {
        // Offered load = 500 × 0.15 = 75 Erlangs
        // With 100 servers, utilization = 75%
        // Some queueing expected during peaks
        
        let token = self.limiter.Acquire()
        if !token {
            // Rate limit exceeded
            return false
        }
        
        // Process request
        delay(150ms)
        return true
    }
}
```

### Example 3: Microservice with Degradation

```sdl
component AdaptiveService {
    uses pool ResourcePool(Size = 10)
    
    param BaseArrivalRate Float = 50.0
    param CurrentLoad Float = 1.0  // Load multiplier
    
    method Process() Bool {
        // Actual arrival rate varies with load
        self.pool.ArrivalRate = self.BaseArrivalRate * self.CurrentLoad
        
        let conn = self.pool.Acquire()
        if !conn {
            return false
        }
        
        // Service time increases under load
        let utilization = self.pool.GetUtilization()
        let serviceTime = 20ms
        
        if utilization > 0.8 {
            // Degraded performance at high utilization
            serviceTime = 100ms
        } else if utilization > 0.6 {
            serviceTime = 50ms
        }
        
        delay(serviceTime)
        return true
    }
}
```

## Understanding Performance Cliffs

### The Non-Linear Nature of Queues

As utilization increases, wait times increase non-linearly:

```sdl
component PerformanceCliffDemo {
    uses pool ResourcePool(Size = 10, AvgHoldTime = 100ms)
    
    method DemoAt50PercentLoad() Bool {
        self.pool.ArrivalRate = 50.0  // 50% utilization
        // Average wait ≈ 11ms
        return self.pool.Acquire()
    }
    
    method DemoAt80PercentLoad() Bool {
        self.pool.ArrivalRate = 80.0  // 80% utilization
        // Average wait ≈ 44ms (4x increase!)
        return self.pool.Acquire()
    }
    
    method DemoAt90PercentLoad() Bool {
        self.pool.ArrivalRate = 90.0  // 90% utilization
        // Average wait ≈ 100ms (9x increase!)
        return self.pool.Acquire()
    }
    
    method DemoAt95PercentLoad() Bool {
        self.pool.ArrivalRate = 95.0  // 95% utilization
        // Average wait ≈ 211ms (19x increase!)
        return self.pool.Acquire()
    }
}
```

### Practical Implications

1. **Keep utilization below 80%** for predictable performance
2. **Monitor queue depths** not just utilization
3. **Plan capacity for peaks** not averages
4. **Add servers before hitting the cliff**

## Advanced Patterns

### 1. Hierarchical Queues

```sdl
component TieredService {
    uses fastPool ResourcePool(Size = 5, AvgHoldTime = 10ms)
    uses slowPool ResourcePool(Size = 20, AvgHoldTime = 100ms)
    
    method Handle() Bool {
        // Try fast pool first
        let fast = self.fastPool.Acquire()
        if fast {
            delay(10ms)
            return true
        }
        
        // Fall back to slow pool
        let slow = self.slowPool.Acquire()
        if slow {
            delay(100ms)
            return true
        }
        
        return false  // Both pools exhausted
    }
}
```

### 2. Queue with Timeouts

```sdl
component TimeoutQueue {
    uses pool ResourcePool(Size = 10)
    param Timeout Duration = 1s
    
    method HandleWithTimeout() Bool {
        let start = now()
        let acquired = self.pool.Acquire()
        
        if !acquired {
            return false
        }
        
        let waitTime = now() - start
        if waitTime > self.Timeout {
            // Waited too long, give up
            return false
        }
        
        // Process if we got here in time
        delay(50ms)
        return true
    }
}
```

### 3. Adaptive Capacity

```sdl
component AutoScalingService {
    uses pool ResourcePool(AvgHoldTime = 50ms)
    
    method AdjustCapacity() {
        let util = self.pool.GetUtilization()
        
        if util > 0.8 {
            // Scale up
            self.pool.Size = self.pool.Size * 1.5
        } else if util < 0.3 {
            // Scale down
            self.pool.Size = self.pool.Size * 0.8
        }
    }
}
```

## Measuring Queue Performance

SDL provides metrics for queue analysis:

```sdl
// In your simulation
sdl measure add queue_util server.pool utilization
sdl measure add queue_wait server.pool.WaitTime latency
sdl measure add queue_length server.pool.QueueLength count

// Run simulation
sdl gen add load server.Handle 100  // 100 RPS

// View results
sdl measure stats
```

## Common Pitfalls

### 1. Forgetting Little's Law

Little's Law: L = λ × W
- L = Average number in system
- λ = Arrival rate  
- W = Average time in system

```sdl
// If you have 100 requests/second spending 200ms each:
// Average number in system = 100 × 0.2 = 20
// You need at least 20 servers!
```

### 2. Ignoring Variance

Real systems have variable service times:

```sdl
// Bad: Assuming fixed service time
param ServiceTime = 50ms

// Good: Modeling variance
param ServiceTime = dist {
    70 => 30ms,   // Most requests are fast
    25 => 100ms,  // Some are slower
    5 => 500ms    // Few are very slow
}
```

### 3. Not Modeling Failures

Failures increase effective service time:

```sdl
component ServiceWithRetries {
    uses pool ResourcePool(Size = 10)
    param FailureRate Float = 0.05
    
    method Handle() Bool {
        let attempts = 0
        for attempts < 3 {
            let conn = self.pool.Acquire()
            if !conn {
                return false
            }
            
            delay(100ms)
            
            let success = sample dist {
                95 => true,
                5 => false
            }
            
            if success {
                return true
            }
            
            attempts = attempts + 1
        }
        
        // Effective service time is now higher due to retries!
        return false
    }
}
```

## Summary

Queueing theory in SDL helps you:

1. **Predict performance** before building
2. **Identify bottlenecks** through utilization metrics
3. **Size systems correctly** using mathematical models
4. **Understand performance cliffs** and avoid them
5. **Model real-world complexity** accurately

Remember: In distributed systems, the queue is often more important than the server. SDL's queueing models help you design systems that perform predictably under load.