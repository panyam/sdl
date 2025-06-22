# Uber Architecture Evolution Demos

This directory contains SDL models and demo scripts showing how Uber's architecture evolved from a simple MVP to a globally distributed system.

## Overview

The demos illustrate three stages of architectural evolution:

1. **MVP (2009)** - A monolithic application with basic functionality
2. **Intermediate (2012)** - Service-oriented architecture with caching and better scaling
3. **Modern (2018+)** - Event-driven microservices with ML, real-time analytics, and global distribution

## Files

### SDL Models
- `mvp.sdl` - The original monolithic architecture
- `intermediate.sdl` - Service-oriented architecture with Redis caching
- `modern.sdl` - Modern microservices with event bus and ML

### Demo Scripts
- `mvp.recipe` - Interactive demo of the MVP system under load
- `intermediate.recipe` - Shows improvements from adding caching and services
- `modern.recipe` - Demonstrates modern architecture handling extreme scale
- `evolution.recipe` - Master script to run all demos

## Running the Demos

### Prerequisites
1. Start the SDL server:
   ```bash
   sdl serve
   ```

2. Open the dashboard in your browser:
   ```
   http://localhost:8080
   ```

### Individual Demos

Run any demo directly:
```bash
./examples/uber/mvp.recipe
./examples/uber/intermediate.recipe
./examples/uber/modern.recipe
```

### Complete Evolution Demo

For the full journey:
```bash
./examples/uber/evolution.recipe
```

### Side-by-Side Comparison

To see all three architectures simultaneously:

1. Open three terminals and run:
   ```bash
   # Terminal 1
   SDL_CANVAS_ID=uber-mvp ./examples/uber/mvp.recipe
   
   # Terminal 2
   SDL_CANVAS_ID=uber-v2 ./examples/uber/intermediate.recipe
   
   # Terminal 3
   SDL_CANVAS_ID=uber-v3 ./examples/uber/modern.recipe
   ```

2. Open three browser tabs:
   - http://localhost:8080/canvases/uber/
   - http://localhost:8080/canvases/uberv2/
   - http://localhost:8080/canvases/uberv3/

## Key Metrics Compared

| Architecture | Max Load | Latency | Cost/Month | Key Features |
|-------------|----------|---------|------------|--------------|
| MVP | 20 RPS | 350ms | $50 | Simple monolith, single DB |
| Intermediate | 300 RPS | 150ms | $205 | Services, Redis cache, async notifications |
| Modern | 5000+ RPS | 50ms | $15,300 | Microservices, ML, event-driven, global |

## Learning Points

### MVP Architecture
- **Strengths**: Simple, cheap, fast to build
- **Weaknesses**: No caching, synchronous operations, limited scale
- **Breaking point**: ~20 RPS due to database bottlenecks

### Intermediate Architecture
- **Improvements**: 
  - Redis caching (70-90% hit rates)
  - Service boundaries for isolation
  - Geo-indexing for efficient queries
  - Async notifications
- **Trade-offs**: More complexity, higher cost
- **Capacity**: 15x improvement over MVP

### Modern Architecture
- **Advanced Features**:
  - Event-driven with Kafka-like bus
  - ML-powered matching
  - Real-time analytics
  - Circuit breakers and rate limiting
  - Global distribution
- **Scale**: 250x the MVP capacity
- **Cost**: Significant investment but handles millions of rides

## Demo Tips

1. **Start with Context**: Explain this is based on public information about Uber's architecture evolution

2. **Interactive Elements**: 
   - Ask audience what they would optimize first
   - Show what happens when components fail
   - Discuss cost vs. scale trade-offs

3. **Key Moments**:
   - MVP breaking under morning rush
   - Cache importance in intermediate version
   - Circuit breakers in modern architecture

4. **Timing**:
   - MVP demo: ~10 minutes
   - Intermediate demo: ~10 minutes  
   - Modern demo: ~12 minutes
   - Full evolution: ~35 minutes

## Technical Details

### Components Used

**SDL Native Components**:
- `ResourcePool` - Connection pooling and capacity modeling
- `Database` - Data persistence with latency simulation
- `Cache` - Redis-like caching with hit rates
- `ExternalService` - Third-party APIs (Maps, SMS)
- `Queue` / `PriorityQueue` - Async message processing
- `CircuitBreaker` - Fault tolerance
- `RateLimiter` - Request throttling

**Custom Components** (defined in modern.sdl):
- `EventBus` - Kafka-like event streaming
- `GeoCache` - Specialized location caching
- `StreamProcessor` - Real-time analytics

### Evolution Patterns

1. **Caching Evolution**:
   - MVP: No caching
   - Intermediate: Basic Redis caching
   - Modern: Multi-layer with geo-caching

2. **Database Evolution**:
   - MVP: Single database, full table scans
   - Intermediate: Indexed queries, better pooling
   - Modern: Sharded, CQRS, time-series

3. **Service Communication**:
   - MVP: Direct method calls
   - Intermediate: Service boundaries
   - Modern: Event-driven + REST

4. **Resilience Evolution**:
   - MVP: Single point of failure
   - Intermediate: Some isolation
   - Modern: Circuit breakers, rate limiting, graceful degradation

## Extending the Demos

To add new scenarios:

1. Modify the SDL files to add new components or behaviors
2. Update the recipe files with new test scenarios
3. Add new metrics to track specific behaviors

Example additions:
- Fraud detection service
- Dynamic routing based on traffic
- Driver incentive calculations
- Multi-modal transportation (bikes, scooters)
