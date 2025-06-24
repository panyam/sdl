# SDL Documentation

Welcome to the SDL (System Design Language) documentation. This guide will help you understand and effectively use SDL for modeling distributed systems.

## 📚 Documentation Structure

### Getting Started
1. **[SDL Language Reference](SDL_LANGUAGE_REFERENCE.md)** - Complete syntax and language features
2. **[Component Library](COMPONENT_LIBRARY.md)** - All built-in components with examples
3. **[Core Concepts](SDL_CORE_CONCEPTS.md)** - Fundamental ideas and design philosophy

### Understanding SDL
4. **[SDL vs Programming Languages](SDL_VS_PROGRAMMING.md)** - Key differences from traditional programming
5. **[Queueing Theory Guide](SDL_QUEUEING_THEORY.md)** - Mathematical foundations of SDL
6. **[System Design Guide](SDL_SYSTEM_DESIGN_GUIDE.md)** - Using SDL for architecture and interviews

### Using SDL Tools
7. **[Console Guide](console/)** - Interactive SDL console documentation
   - [Overview](console/01-OVERVIEW.md)
   - [Getting Started](console/02-GETTING-STARTED.md)
   - [Basic Commands](console/03-BASIC-COMMANDS.md)
   - [Traffic Generation](console/04-TRAFFIC-GENERATION.md)
   - [Measurements](console/05-MEASUREMENTS.md)
   - [Web Dashboard](console/06-WEB-DASHBOARD.md)
   - [Advanced Features](console/07-ADVANCED-FEATURES.md)
   - [Remote Access](console/08-REMOTE-ACCESS.md)
   - [Troubleshooting](console/09-TROUBLESHOOTING.md)

8. **[Web Dashboard Guide](WEB_DASHBOARD_GUIDE.md)** - Using the interactive dashboard
9. **[Measurement Viewing Guide](MEASUREMENT_VIEWING_GUIDE.md)** - Analyzing simulation results

## 🎯 Learning Path

### For Beginners
1. Start with [Core Concepts](SDL_CORE_CONCEPTS.md) to understand SDL's philosophy
2. Read [SDL vs Programming](SDL_VS_PROGRAMMING.md) to avoid common misconceptions
3. Work through the [Language Reference](SDL_LANGUAGE_REFERENCE.md) with examples
4. Try the [Console Getting Started](console/02-GETTING-STARTED.md) guide

### For System Designers
1. Jump to the [System Design Guide](SDL_SYSTEM_DESIGN_GUIDE.md)
2. Review the [Component Library](COMPONENT_LIBRARY.md) for modeling patterns
3. Understand [Queueing Theory](SDL_QUEUEING_THEORY.md) for capacity planning
4. Use the [Web Dashboard](WEB_DASHBOARD_GUIDE.md) for visualization

### For Performance Engineers
1. Deep dive into [Queueing Theory](SDL_QUEUEING_THEORY.md)
2. Master [Traffic Generation](console/04-TRAFFIC-GENERATION.md)
3. Learn [Measurements](console/05-MEASUREMENTS.md) and analysis
4. Explore [Advanced Features](console/07-ADVANCED-FEATURES.md)

## 🔑 Key Concepts

### What SDL IS:
- A **modeling language** for distributed systems
- A **simulation framework** for performance analysis
- A **capacity planning tool** based on queueing theory
- A **validation system** for architectural decisions

### What SDL IS NOT:
- Not a programming language
- Not a configuration tool
- Not a monitoring system
- Not an implementation framework

### Core Principles:
1. **Model behavior, not implementation**
2. **Embrace uncertainty with probabilistic modeling**
3. **Use virtual time for fast simulations**
4. **Focus on interactions between components**

## 💡 Quick Examples

### Simple Service Model
```sdl
component WebService {
    uses db Database
    uses cache Cache(HitRate = 0.8)
    
    method HandleRequest() Bool {
        // Check cache first
        if self.cache.Get() {
            return true
        }
        
        // Cache miss - query database
        return self.db.Query()
    }
}
```

### System with Load Balancing
```sdl
system ScalableAPI {
    use lb LoadBalancer
    use api1 APIServer(db = database)
    use api2 APIServer(db = database)
    use api3 APIServer(db = database)
    use database Database(ConnectionPoolSize = 100)
}
```

### Performance Testing
```bash
# Load the system
sdl load system.sdl

# Start traffic generation
sdl gen add normal api.HandleRequest 100  # 100 RPS

# Measure performance
sdl measure add latency api.HandleRequest latency
sdl measure add success api.HandleRequest count

# View results
sdl measure stats
```

## 🛠️ Common Use Cases

1. **Capacity Planning** - Determine how many servers you need
2. **Bottleneck Analysis** - Find system limits before building
3. **Architecture Validation** - Compare design alternatives
4. **Performance Prediction** - Estimate latencies under load
5. **Failure Analysis** - Understand cascading failures
6. **Cost Optimization** - Right-size infrastructure

## 📖 Example Systems

The SDL repository includes several complete examples:

- **Uber** - Ride-sharing platform evolution (MVP → Modern)
- **Netflix** - Video streaming with CDN
- **Twitter** - Social media timeline generation
- **Bitly** - URL shortener with caching
- **Contacts** - Simple microservice example

Find these in the `examples/` directory.

## 🚀 Next Steps

1. **Install SDL** - Follow the [README](../README.md) installation guide
2. **Try the Examples** - Run the demo systems in `examples/`
3. **Build Your Own** - Model a system you're familiar with
4. **Join the Community** - Share your models and learn from others

## 📝 Contributing

See our [Contributing Guide](../CONTRIBUTING.md) for:
- Documentation improvements
- New component implementations
- Example systems
- Bug fixes and features

Remember: SDL is about understanding systems, not building them. Focus on modeling what matters for performance and capacity.