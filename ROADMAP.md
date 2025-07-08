# SDL Project Roadmap

This document outlines the long-term vision and development roadmap for SDL (System Design Language), an open-source platform for modeling and simulating distributed systems.

## Project Vision

SDL aims to become the industry standard for system design education, capacity planning, and performance modeling. Our goal is to make complex distributed systems behavior understandable and predictable through interactive simulations.

## Current State (July 2025)

### Core Platform
- **Language**: SDL modeling language with components, systems, and probabilistic outcomes
- **Runtime**: High-performance simulation engine with virtual time support
- **Architecture**: gRPC-based microservices with HTTP gateway
- **Metrics**: Real-time streaming metrics with aggregation
- **Visualization**: Interactive web dashboard with system diagrams
- **File Management**: Multi-filesystem support with security and filtering
- **Dashboard Architecture**: Event-driven components with centralized state management

### Key Features
- Traffic generation up to 1000+ RPS
- Flow analysis with automatic rate calculation
- M/M/c queuing theory implementation
- Method-level system visualization
- Multi-canvas isolation for concurrent simulations
- Secure filesystem access with extension filtering
- Tabbed editor with real-time file operations
- Integrated recipe execution with visual debugging
- Multi-filesystem support with unified interface

### Recent Achievements (v4.0)
- **FileSystem Architecture**: Complete abstraction layer for file operations
- **Security Implementation**: Path traversal protection and read-only enforcement
- **File Filtering**: Server-side extension filtering (.sdl, .recipe)
- **Multi-FileSystem Support**: Local and GitHub filesystem clients
- **Recipe Integration**: In-editor execution controls for SDL command sequences
- **Tab Management**: Composite keys for unique file identification across filesystems
- **Visual Debugging**: Line highlighting and execution tracking for recipes
- **Architecture Refactoring**: Event-driven dashboard with 80% code reduction
- **Component Isolation**: Self-contained panels with lifecycle management
- **Service Layer**: Clean API abstraction with event emission
- **State Management**: Centralized AppStateManager with observer pattern
- **Recipe Controls**: Singleton pattern with global toolbar integration
- **UI Stability**: Fixed toolbar render issues preserving child components

### Latest Achievements (v4.1 - December 2024)
- **Minitools Architecture**: Page-specific WASM modules replacing monolithic approach
- **SystemDetailTool**: Complete Go implementation with @stdlib import support and recipe parsing
- **WASM Integration**: Dedicated 27MB WASM module with JavaScript bindings and TypeScript wrapper
- **Enhanced Build System**: Multiple WASM module discovery, cataloging, and conditional loading
- **Security Model**: Import validation preventing local imports, shell syntax protection
- **Template Infrastructure**: SystemDetailTool serves as template for creating additional focused tools
- **Environment Agnostic**: Same tool works across CLI, server, and WASM contexts
- **Performance Optimization**: Lazy loading of WASM modules with cache busting for development

## Roadmap Phases

### Phase 1: Production Readiness (Q3 2025) - In Progress
**Goal**: Make SDL accessible to the broader developer community

- **Architecture Modernization** (NEW)
  - ✅ Event-driven component architecture
  - ✅ Decoupled dashboard components
  - ⏳ Complete migration from legacy dashboard
  - ⏳ WASM mode compatibility with new architecture

- **Documentation**
  - Comprehensive SDL language reference
  - Component library documentation
  - API reference generation
  - Tutorial series expansion

- **Developer Experience**
  - One-command installation
  - Cross-platform support (Windows, Linux, macOS)
  - IDE extensions (VS Code, IntelliJ)
  - Improved error messages

- **Community Building**
  - Open source governance model
  - Contribution guidelines
  - Community forums/Discord
  - Regular release cadence

### Phase 2: Enhanced Modeling Capabilities (Q4 2025)
**Goal**: Model more complex distributed system patterns

- **Temporal Orchestration Layer**
  - Discrete event simulation engine
  - State machine modeling
  - Circuit breakers and retry logic
  - Rate limiting and backpressure
  - Chaos engineering scenarios

- **Advanced Components**
  - Message queues (Kafka, RabbitMQ patterns)
  - Service mesh behaviors
  - CDN and edge computing
  - Blockchain and consensus algorithms
  - ML inference pipelines

- **Network Modeling**
  - Latency distributions by geography
  - Bandwidth constraints
  - Packet loss simulation
  - Network partition scenarios

### Phase 3: Enterprise Features (Q1 2026)
**Goal**: Enable SDL adoption in production environments

- **Observability Integration**
  - OpenTelemetry export
  - Prometheus metrics
  - Grafana dashboards
  - APM tool integration

- **Cloud Cost Modeling**
  - AWS/GCP/Azure pricing integration
  - Resource cost estimation
  - Cost optimization recommendations
  - Multi-region deployment costs

- **Capacity Planning**
  - What-if scenario analysis
  - Load testing integration
  - Performance regression detection
  - SLA compliance checking

### Phase 4: Collaborative Platform (Q2 2026)
**Goal**: Transform SDL into a collaborative system design platform

- **Multi-User Features**
  - Real-time collaborative editing
  - Design review workflows
  - Version control integration
  - Team workspaces

- **Visual System Builder**
  - Drag-and-drop component composition
  - Visual parameter tuning
  - Live system modification
  - Component marketplace

- **Knowledge Sharing**
  - Public system design library
  - Industry-specific templates
  - Best practices repository
  - Case study database

### Phase 5: AI-Assisted Design (Q3 2026)
**Goal**: Leverage AI to accelerate system design

- **AI Features**
  - Natural language to SDL conversion
  - Automatic bottleneck detection
  - Performance optimization suggestions
  - Anomaly detection in designs

- **Learning Platform**
  - Interactive system design courses
  - Automated grading and feedback
  - Personalized learning paths
  - Certification programs

## Technical Roadmap

### Performance Improvements
- GPU acceleration for large-scale simulations
- Distributed simulation across multiple nodes
- Incremental compilation for faster iteration
- Memory optimization for million-component systems

### Language Evolution
- Type inference improvements
- Module system for component libraries
- Generics for reusable components
- Macro system for common patterns

### Platform Expansion
- WebAssembly runtime for browser execution
- Mobile applications for system monitoring
- CLI plugins architecture
- REST/GraphQL API alternatives

## Research Directions

### Academic Partnerships
- Queuing theory advancement
- New simulation algorithms
- Formal verification integration
- Performance prediction ML models

### Industry Collaboration
- Real-world system case studies
- Production trace replay
- Vendor-specific component libraries
- Industry standard development

## Success Metrics

### Adoption Goals
- 10,000+ active users by end of 2025
- 100+ contributed components
- 50+ enterprise adoptions
- 20+ university courses using SDL

### Technical Goals
- Sub-millisecond simulation latency
- Million operations per second
- 99.9% simulation accuracy
- Support for 100k+ component systems

## Contributing Areas

We welcome contributions in:
- Component library expansion
- Language feature development
- Documentation and tutorials
- Visualization improvements
- Testing and validation
- Performance optimization
- Platform integrations

## Long-term Vision (2027+)

SDL aims to become:
- The de facto standard for system design education
- A critical tool in production capacity planning
- The foundation for next-generation system optimization
- A bridge between design and implementation

Through open collaboration and continuous innovation, SDL will democratize distributed systems expertise and make complex system behavior predictable and manageable for everyone.