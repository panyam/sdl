# SDL Release Notes

## Version 0.9.0 (Pre-Release) - June 2025

This pre-release represents a major architectural overhaul and feature expansion of the SDL platform. While not yet at 1.0, this version provides a solid foundation for system design modeling and simulation.

### Major Features

#### gRPC Architecture Migration
- Migrated from REST to gRPC with HTTP gateway support
- Dual protocol support: native gRPC (port 9090) and REST gateway (port 8080)
- Protobuf-based API definitions for type safety
- Connect-Web integration for browser streaming

#### Enhanced Simulation Engine
- Virtual time support for deterministic simulations
- High-throughput traffic generation (tested up to 1000+ RPS)
- M/M/c queuing theory implementation with utilization tracking
- Method-level execution tracing and analysis

#### Interactive Web Dashboard
- Professional 4-panel layout with drag-and-drop rearrangement
- Real-time system diagram with method-level visualization
- Live metrics streaming with Chart.js integration
- Traffic generator lifecycle management UI
- Persistent layout preferences

#### Flow Analysis System
- Automatic flow rate calculation based on generators
- Runtime-based flow evaluation with conditional paths
- Visual flow indicators with execution order
- Batch parameter updates for system-wide changes

#### Metrics Architecture
- Pluggable tracer interface design
- Ring buffer-based metric storage
- Pre-aggregated metrics at collection time
- Support for latency, count, and utilization metrics
- Streaming metrics API with time-window queries

### New Commands

#### Canvas Management
- `sdl canvas create <id>` - Create new simulation canvas
- `sdl canvas list` - List all available canvases
- `sdl canvas reset <id>` - Reset specific canvas (safety: requires ID)

#### Generator Commands
- `sdl gen add <name> <entrypoint> <rate>` - Add traffic generator
- `sdl gen update <name> <rate>` - Update generator rate efficiently
- `sdl gen start/stop/pause/resume` - Lifecycle management
- `--apply-flows` flag for automatic flow calculation

#### Analysis Commands
- `sdl trace <entrypoint>` - Human-readable execution trace
- `sdl paths <component.method>` - Static path analysis
- `sdl utilization` - Component utilization breakdown
- `sdl flows eval` - Evaluate system flow rates

### Example Systems

#### Uber Ride-Sharing Evolution
Complete three-stage demo showing architectural evolution:
- **MVP**: Monolithic design breaking at 20 RPS
- **Intermediate**: Service-oriented with caching, handles 300 RPS
- **Modern**: Event-driven microservices, 5000+ RPS capacity

#### Netflix Streaming Service
Comprehensive model including:
- CDN with geographic distribution
- Video encoding pipeline
- Database with caching layers
- Load balancing strategies

#### Additional Examples
- Contacts service (simple 2-tier)
- Bitly URL shortener
- Twitter timeline generation
- System delay demonstrations

### Developer Experience

#### Hot Reload Workflow
- Air configuration for automatic rebuilds
- Instant web dashboard updates
- Proto regeneration on change

#### Testing Infrastructure
- Comprehensive unit tests
- End-to-end test recipes
- Playwright browser testing
- Visual regression testing

### Breaking Changes

1. **API Migration**: All HTTP endpoints migrated to gRPC
   - Update clients to use new gRPC endpoints
   - REST gateway available for compatibility

2. **Canvas Reset**: Now requires explicit canvas ID
   - Change `sdl reset` to `sdl canvas reset <id>`

3. **Metric Types**: Removed `throughput` metric type
   - Use `count` metric with rate aggregation instead

### Known Limitations

1. **Platform Support**: Primarily tested on macOS
   - Linux support expected to work
   - Windows support untested

2. **Flow Analysis**: Conservative over-estimation
   - Early returns not fully modeled
   - Manual rate adjustments may be needed

3. **Installation**: Manual prerequisite installation required
   - No automated installer yet
   - Requires Go 1.24+, Node.js, goyacc

### Technical Improvements

- Thread-safe Canvas operations
- Efficient component instance resolution
- Memory-bounded metric storage
- Optimized flow solver with damping
- Proper error handling throughout

### Bug Fixes

- Fixed nil pointer in measurement endpoints
- Resolved chart destruction errors in dashboard
- Fixed component name resolution in nested systems
- Corrected time unit parsing (seconds as base)
- Fixed generator lifecycle race conditions

### Acknowledgments

This release represents months of architectural improvements and feature development. Special thanks to all contributors and early testers who provided valuable feedback.

### Upgrade Guide

For users upgrading from earlier versions:

1. Update CLI commands to use new syntax
2. Regenerate any custom proto definitions
3. Update dashboard bookmarks to include canvas ID
4. Review generator configurations for new features

### What's Next

See ROADMAP.md for upcoming features including:
- Temporal orchestration layer
- Enhanced component library
- Multi-user collaboration
- AI-assisted design