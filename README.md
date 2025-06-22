# SDL (System Design Language)

A specialized language and toolchain for modeling, simulating, and analyzing the performance characteristics of distributed systems.

## What is SDL?

SDL enables rapid analysis of system designs through:
- **Performance modeling** with latency and availability distributions
- **Capacity analysis** using queuing theory (M/M/c models)
- **Bottleneck identification** under different loads
- **SLO evaluation** and performance exploration
- **Interactive analysis** with parameter modification
- **Real-time visualization** with web dashboard
- **RESTful API** for traffic generation and measurement management
- **Diagram generation** from system definitions

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd sdl

# Install prerequisites

#Install Buf if not already done.  This will make buf generally available
npm config set @buf:registry https://buf.build/gen/npm/v1/
npm install @connectrpc/connect @connectrpc/connect-web
npm install --save-dev -g @bufbuild/buf

# Install the es/connect generator plugins for proto and buf.  Needed for `buf generate`
npm install --save-dev -g @bufbuild/protoc-gen-es

# Build the CLI tool
make build

# Or install directly
go install ./cmd/sdl
```

### Your First SDL Model

Create a simple disk model (`mydisk.sdl`):

```sdl
// Define a disk component with realistic performance characteristics
component SimpleDisk {
    param ReadLatency = dist {
        90 => 5ms,    // 90% of reads: 5ms
         9 => 20ms,   // 9% of reads: 20ms  
         1 => 100ms   // 1% of reads: 100ms (outliers)
    }
    
    method Read() Bool {
        delay(sample self.ReadLatency)
        return sample dist {
            999 => true,   // 99.9% success rate
              1 => false   // 0.1% failure rate
        }
    }
}

system MySystem {
    use disk SimpleDisk
}
```

### Run Performance Analysis

```bash
# Validate your model
sdl validate mydisk.sdl

# Run 1000 simulations and analyze latency
sdl run mydisk.sdl MySystem disk.Read --count 1000 --output results.json

# Generate latency distribution plot
sdl plot results.json --type latency --output latency.png

# Create system architecture diagram
sdl diagram mydisk.sdl MySystem --type static --output architecture.svg
```

## 🏗️ Core Concepts

### Components and Systems

```sdl
// Define reusable components
component Database {
    param ConnectionPoolSize = 10
    param QueryLatency = 15ms
    
    uses cache Cache
    uses disk SimpleDisk
    
    method Query() Bool {
        // Try cache first
        let hit = self.cache.Read()
        if hit {
            return true
        }
        
        // Fall back to disk
        return self.disk.Read()
    }
}

// Compose into systems
system WebService {
    use db Database
    use loadBalancer LoadBalancer
}
```

### Probabilistic Modeling

```sdl
// Use distributions to model real-world variability
param ResponseTime = dist {
    50 => 10ms,     // Median case
    30 => 25ms,     // Slower responses
    15 => 50ms,     // Even slower
     4 => 200ms,    // Outliers
     1 => 1000ms    // Rare tail latencies
}

// Sample from distributions in your logic
method ProcessRequest() Bool {
    delay(sample self.ResponseTime)
    return true
}
```

### Capacity Modeling with ResourcePool

```sdl
import ResourcePool from "./common.sdl"

component DiskWithCapacity {
    // Model disk IOPS capacity
    uses pool ResourcePool(Size = 100)  // 100 IOPS capacity
    
    param ServiceTime = 10ms  // Time per I/O operation
    
    method Read() Bool {
        // Acquire capacity (may queue if busy)
        let acquired = self.pool.Acquire()
        if not acquired {
            return false  // Overloaded - request dropped
        }
        
        // Perform the actual I/O
        delay(self.ServiceTime)
        return true
    }
}
```

## 📊 Interactive Analysis

### Web Dashboard (New!)

Experience SDL through our powerful web interface with real-time visualization:

```bash
# Start the interactive web dashboard
./sdl serve --port 8080
```

Navigate to `http://localhost:8080` for the **"Incredible Machine"** experience:

- **2-Row Dynamic Layout:**
  - **Row 1**: System Architecture (left) + Traffic Generation & System Parameters (right)
  - **Row 2**: Live Metrics Grid with unlimited scrollable charts
- **Real-time Parameter Controls:** Sliders for instant system modification
- **Dynamic Metrics Visualization:** Charts auto-generated from `canvas.Measure()` calls
- **Live Performance Feedback:** WebSocket-powered real-time updates
- **Proper Panel Clipping:** All content contained within panel boundaries

### Canvas API for Advanced Analysis

SDL's Canvas API provides stateful, RESTful session management for complex analysis workflows:

#### RESTful Traffic Generation API
```bash
# Add traffic generators
curl -X POST http://localhost:8080/api/canvas/generators \
  -d '{"id":"load1", "name":"Peak Traffic", "target":"app.HandleRequest", "rate":100, "enabled":true}'

# Control generators  
curl -X POST http://localhost:8080/api/canvas/generators/start
curl -X POST http://localhost:8080/api/canvas/generators/load1/pause

# Get current state
curl http://localhost:8080/api/canvas/generators
```

#### Measurement Management
```bash
# Add measurements
curl -X POST http://localhost:8080/api/canvas/measurements \
  -d '{"id":"latency1", "name":"Response Latency", "metricType":"latency", "target":"app", "interval":1000, "enabled":true}'

# View measurements
curl http://localhost:8080/api/canvas/measurements
```

#### Canvas State Management
```bash
# Save current Canvas state
curl -X POST http://localhost:8080/api/canvas/state

# Restore previous state  
curl -X POST http://localhost:8080/api/canvas/state/restore \
  -d '{"loadedFiles":["app.sdl"], "activeSystem":"WebApp", "generators":{...}}'
```

#### CLI Integration
```bash
# Load and modify systems interactively
sdl execute analysis.recipe
```

Example recipe file (`analysis.recipe`):
```
load mydisk.sdl
use MySystem

# Test normal load
set disk.ReadLatency dist { 90 => 5ms, 10 => 20ms }
run normal_load disk.Read --count 1000
plot normal_load --type latency

# Test under high contention  
set disk.pool.ArrivalRate 50  # requests/second
set disk.pool.AvgHoldTime 20ms
run high_load disk.Read --count 1000
plot high_load --type latency

# Compare results
plot normal_load,high_load --type comparison
```

## 🔧 CLI Commands

### Core Workflow Commands

```bash
# Validate SDL syntax and semantics
sdl validate <file.sdl>

# Run simulations
sdl run <file.sdl> <SystemName> <method> [options]
  --count <n>        # Number of simulation runs
  --output <file>    # Save results to file
  --format json|csv  # Output format

# Generate plots from simulation results  
sdl plot <results.json> [options]
  --type latency|histogram|timeseries
  --output <file.png>
  --title "Custom Title"

# Create system diagrams
sdl diagram <file.sdl> <SystemName> [options]
  --type static|dynamic
  --format svg|png|excalidraw
  --output <file>

# Interactive analysis
sdl execute <recipe.file>

# Single execution trace (for debugging)
sdl trace <file.sdl> <SystemName> <method>
```

### Analysis and Inspection

```bash
# List components and systems
sdl list <file.sdl>

# Describe component details
sdl describe <file.sdl> <ComponentName>
```

## 📈 Capacity Analysis Example

Analyze how system performance degrades under increasing load:

```sdl
// High-capacity web server
component WebServer {
    uses pool ResourcePool(Size = 50)  // 50 concurrent requests
    param ProcessingTime = 100ms
    
    method HandleRequest() Bool {
        let acquired = self.pool.Acquire()
        if not acquired {
            return false  // Server overloaded
        }
        
        delay(self.ProcessingTime)
        return true
    }
}
```

Load testing recipe:
```
load webserver.sdl
use WebSystem

# Test increasing arrival rates
set server.pool.ArrivalRate 10
run load_10rps server.HandleRequest --count 1000

set server.pool.ArrivalRate 30  
run load_30rps server.HandleRequest --count 1000

set server.pool.ArrivalRate 60  # Above capacity!
run load_60rps server.HandleRequest --count 1000

# Compare latency distributions
plot load_10rps,load_30rps,load_60rps --type comparison
```

Expected results:
- **10 RPS**: ~100ms latency, 100% success
- **30 RPS**: ~120ms latency, 100% success  
- **60 RPS**: ~300ms+ latency, failures occur

## 🏗️ Architecture

SDL is built as a modular Go system:

- **`parser`**: Converts SDL text to Abstract Syntax Tree
- **`loader`**: Resolves imports and performs type checking
- **`runtime`**: Executes simulations and manages component instances
- **`components`**: Library of pre-built system components (disk, cache, etc.)
- **`console`**: Interactive analysis engine (Canvas API)
- **`viz`**: Plotting and diagram generation
- **`cmd/sdl`**: Command-line interface

## 📚 Examples

Explore the `examples/` directory for complete system models:

- **`examples/capacity.sdl`**: Capacity modeling with ResourcePool
- **`examples/bitly/`**: URL shortener service architecture
- **`examples/twitter/`**: Social media platform components
- **`examples/leetcode/`**: Algorithm and data structure performance

## 🌐 Web Dashboard

### Simple 2-Row Layout

The SDL web dashboard features a groundbreaking 2-row dynamic layout designed for system design interview coaching:

**Row 1 (50% height): System Architecture + Controls**
- **Enhanced System Architecture (70% width)**: Prominent visualization with detailed component metrics
- **Traffic Generation (30% × 48%)**: Dynamic traffic source management  
- **System Parameters (30% × 48%)**: Real-time parameter controls

**Row 2 (50% height): Dynamic Metrics Grid**
- **Unlimited Scrollable Charts**: Support for infinite metrics via `canvas.Measure()` calls
- **Color-Coded Visualization**: Red (latency), Green (QPS), Orange (errors), Purple (cache), Blue (utilization), Pink (memory)
- **Responsive Grid Layout**: Automatically adapts to screen size and content

### Key Features
- **Proper Panel Clipping**: All content contained within panel boundaries
- **Real-time Updates**: WebSocket-powered live chart updates
- **Enhanced System Visualization**: Supports complex enterprise-scale architectures
- **Professional Interface**: Clean, modern design suitable for conference presentations

### Getting Started
```bash
# Build and start the web dashboard
./sdl serve --port 8080
```

Visit `http://localhost:8080` to experience the interactive "Incredible Machine" interface.

For detailed information, see [WEB_DASHBOARD_GUIDE.md](WEB_DASHBOARD_GUIDE.md).

## 🛠️ Development

### Building from Source

```bash
# Install dependencies
go mod download

# Generate parser (requires goyacc)
make parser

# Build CLI
make build

# Run tests
go test ./...

# Run specific component tests
go test ./components -v
go test ./console -v

# Build and test web dashboard
cd web && npm install && npm run build && cd ..
```

### Project Structure

See `SUMMARY.md` files in each package for detailed technical documentation:

- **[Project Overview](SUMMARY.md)** - High-level architecture and status
- **[Core Package](core/SUMMARY.md)** - Probabilistic modeling primitives
- **[Components](components/SUMMARY.md)** - System component library
- **[Runtime](runtime/SUMMARY.md)** - Simulation execution engine
- **[Console](console/SUMMARY.md)** - Interactive analysis framework

## 🎯 Use Cases

### Performance Engineering
- Model latency distributions under different loads
- Identify bottlenecks before they hit production
- Compare architectural alternatives quantitatively

### Capacity Planning  
- Determine system limits using queuing theory
- Plan for traffic growth and resource scaling
- Optimize resource allocation across components

### SLO Validation
- Verify that designs meet latency and availability targets
- Test resilience under failure scenarios
- Generate evidence for capacity and performance claims

### Architecture Documentation
- Create executable specifications of system behavior
- Generate diagrams that stay synchronized with models
- Share performance assumptions across teams

## Technical Architecture

### Canvas API Design

SDL's Canvas API follows RESTful principles with WebSocket integration for optimal performance:

- **Control Plane (HTTP)**: All configuration operations use RESTful endpoints
  - Traffic generator management: `POST /api/canvas/generators`
  - Measurement configuration: `POST /api/canvas/measurements`  
  - Canvas state operations: `GET/POST /api/canvas/state`
- **Data Plane (WebSocket)**: Real-time updates only
  - Live metrics broadcasting
  - Parameter change notifications
  - Generator status updates

### Web Server Architecture

Consolidated web server implementation in `console/canvas_web.go`:

- **goutils WebSocket Integration**: Production-grade WebSocket handling with lifecycle hooks
- **Thread-safe Connection Management**: Concurrent client handling with proper synchronization
- **RESTful API Design**: Stateless operations with full Canvas state management
- **Single Binary Deployment**: Complete frontend bundled with Go backend

### Frontend Architecture

TypeScript + Tailwind frontend with professional 2-row layout:

- **Row 1**: System Architecture (70% width) + Traffic/Parameter Controls (30% width)
- **Row 2**: Dynamic metrics grid with unlimited scrollable charts
- **Real-time Updates**: WebSocket-powered live data visualization
- **Type-safe Communication**: Shared TypeScript interfaces with Go backend

## Roadmap

See [NEXTSTEPS.md](NEXTSTEPS.md) for detailed development plans:

- **COMPLETED**: RESTful Canvas API with traffic generation and measurement management
- **COMPLETED**: goutils WebSocket integration with production-grade connection handling
- **COMPLETED**: Consolidated web server architecture and 2-row dashboard layout
- **CURRENT**: Frontend integration with new RESTful Canvas API
- **NEXT**: Enhanced testing and documentation for production deployment

## 📄 License

[License information to be added]

## 🤝 Contributing

[Contributing guidelines to be added]

---

**SDL**: Making system performance modeling accessible, interactive, and actionable.
