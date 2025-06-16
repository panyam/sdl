# SDL Web Dashboard Summary

**Version:** DockView Professional Layout (v2.0)

## 🎯 Purpose

The SDL Web Dashboard provides an interactive "Incredible Machine" style interface for system design interview coaching and distributed systems performance analysis. It transforms complex system modeling into an intuitive, visual experience.

## 🏗️ Architecture Overview

### Frontend Stack
- **Build System**: Vite for fast development and optimized production builds
- **Language**: TypeScript for type safety and better development experience
- **Styling**: Tailwind CSS for rapid, utility-first styling with custom DockView theme overrides
- **Layout System**: DockView for professional dockable panels with persistence
- **Charting**: Chart.js for real-time performance visualization
- **System Visualization**: Graphviz WASM for automatic system diagram layout
- **Communication**: Native WebSocket API for live updates

### Layout Architecture

#### DockView Professional Layout
The dashboard features a modern, professional dockable layout system using DockView that provides maximum flexibility and customization:

**4-Panel Dockable System**
- **System Architecture Panel**: Enhanced system visualization with Graphviz/dot rendering
  - Displays complete system structure from Canvas API
  - Shows all components with per-method traffic values
  - Automatic layout via dot engine for clean presentation
  - Dynamically updates based on loaded SDL file
- **Traffic Generation Panel**: Generator controls with Start/Stop functionality
  - Real-time traffic control and monitoring
  - Add/remove generators dynamically
- **Measurements Panel**: System measurement configuration
  - Add and configure custom metrics
  - Real-time measurement management
- **Live Metrics Panel**: Dynamic charts grid
  - Unlimited scrollable charts supporting infinite metrics via `canvas.Measure()` calls
  - 3-column responsive grid that adapts to screen size
  - Real-time chart updates with proper scaling

**DockView Features**
- **Fully Resizable**: Drag splitters to adjust panel sizes
- **Dockable Tabs**: Drag tabs to rearrange panel positions
- **Layout Persistence**: Custom layouts automatically saved to localStorage
- **Professional Styling**: Dark theme with blue highlights for active tabs
- **Reset Functionality**: One-click return to default 2x2 grid layout

### Key Components

#### Dashboard Class (`src/dashboard.ts`)
- **Main Application Controller**: Manages the entire dashboard state and rendering
- **WebSocket Integration**: Handles real-time communication with backend
- **Dynamic Chart Management**: Creates and updates charts based on server metrics
- **Parameter Control**: Manages real-time system parameter modification

#### Canvas API Client (`src/canvas-api.ts`)
- **RESTful API Integration**: Full implementation of Canvas REST endpoints
- **WebSocket Client**: Manages real-time bidirectional communication with REPL console synchronization
- **Enhanced State Management**: Handles complete Canvas state save/restore including system parameters and session recovery
- **Traffic Generation**: Complete generator lifecycle management
- **Measurement Control**: Dynamic chart creation from Canvas measurements
- **REPL Synchronization**: Real-time updates from console commands via WebSocket broadcasting
- **Error Handling**: Robust error management for all API operations

#### Type System (`src/types.ts`)
- **Shared Data Structures**: Type-safe communication with Go backend
- **API Interfaces**: Complete type coverage for all API operations
- **Chart Data Models**: Type definitions for dynamic chart management

## 🎨 Design Principles

### Visual Hierarchy
1. **System Architecture Prominence**: 70% width allocation emphasizes system visualization
2. **Organized Controls**: Separated panels prevent UI clutter
3. **Unlimited Metrics**: Scrollable grid accommodates any number of charts
4. **Professional Aesthetics**: Clean, modern interface suitable for presentations

### User Experience
1. **Immediate Feedback**: Parameter changes instantly update all relevant panels
2. **Visual Containment**: Proper clipping ensures clean boundaries
3. **Responsive Design**: Layout adapts to different screen sizes
4. **Intuitive Controls**: Clear, accessible interface for workshop audiences

### Technical Robustness
1. **Type Safety**: Full TypeScript coverage prevents runtime errors
2. **Real-time Updates**: WebSocket integration for live data synchronization
3. **Performance Optimization**: Efficient chart updates and memory management
4. **Error Resilience**: Graceful handling of connection issues and API errors

## 🔄 Data Flow

### Initialization Flow
1. **Load Application**: TypeScript bundle loads in browser
2. **Establish WebSocket**: Connect to backend for real-time updates and REPL synchronization
3. **Load Canvas State**: Fetch current Canvas state via REST API (includes REPL session state)
4. **Initialize System**: Display system architecture if Canvas has active system
5. **Setup Generators**: Load any existing traffic generators and measurements
6. **Sync REPL State**: Automatically sync with any active console session
7. **Render UI**: Display 2-row layout with populated or empty panels

### Parameter Modification Flow
1. **User Interaction**: Slider movement, checkbox toggle, or REPL console command
2. **State Update**: Update local dashboard state or receive from console
3. **API Call**: Send parameter change to backend via HTTP (or receive from REPL)
4. **Auto-Simulation**: Backend automatically runs new simulation
5. **WebSocket Broadcast**: Real-time results broadcast to all clients (including REPL)
6. **UI Refresh**: Update system architecture, metrics, and charts
7. **Cross-Session Sync**: Changes visible in both dashboard and console immediately

### Chart Management Flow
1. **Metric Registration**: Backend calls `canvas.Measure()` with new metrics
2. **WebSocket Notification**: Frontend receives new chart specification
3. **Dynamic Creation**: Create new Chart.js instance with appropriate colors
4. **Grid Integration**: Add chart to scrollable metrics grid
5. **Real-time Updates**: Continuous data updates via WebSocket

## 🧪 Testing Strategy

### Playwright Integration
- **Layout Validation**: Automated tests verify proper panel sizing and positioning
- **Clipping Verification**: Tests ensure content stays within panel boundaries
- **Responsive Testing**: Multi-viewport testing for different screen sizes
- **Interaction Testing**: Automated parameter modification and result verification

### Test Coverage
- **Grid Layout**: Verify 2-row structure and panel proportions
- **Scrolling Behavior**: Test unlimited metrics grid scrolling
- **Visual Clipping**: Ensure no content overflow beyond panel boundaries
- **Chart Creation**: Validate dynamic chart addition and color coding
- **Real-time Updates**: Test WebSocket communication and live updates

## 🚀 Production Features

### Deployment
- **Single Binary**: Complete frontend bundled with Go backend
- **Zero Configuration**: No external dependencies required
- **Development Mode**: Vite dev server for rapid iteration
- **Production Build**: Optimized, minified assets for conference deployment

### Performance
- **Fast Rendering**: Efficient TypeScript and Tailwind combination
- **Real-time Charts**: 60fps updates during live demonstrations
- **Memory Management**: Proper cleanup of chart instances and WebSocket connections
- **Network Efficiency**: Minimal payload sizes for real-time updates

## 🎯 Conference Workshop Success

### Educational Impact
- **Visual Learning**: Complex system concepts made immediately visible
- **Interactive Exploration**: Audience can see cause-and-effect relationships via both dashboard and console
- **Side-by-Side Demonstrations**: REPL console commands instantly update dashboard - perfect for workshops
- **Professional Presentation**: Suitable for large-screen conference demos
- **Enhanced Engagement**: "Incredible Machine" experience with real-time console synchronization captures attention
- **No curl Required**: Clean REPL interface eliminates clunky HTTP commands during presentations

### Technical Reliability
- **Production Ready**: Comprehensive testing ensures demo reliability
- **Cross-Platform**: Works on all modern browsers
- **Network Resilient**: Graceful handling of connection issues
- **Performance Optimized**: Smooth experience even under load

## 🔮 Future Enhancements

### Advanced Visualization
- **Drag-and-Drop**: Visual system composition interface
- **3D Architecture**: Enhanced system visualization capabilities
- **Animation Effects**: Smooth transitions for parameter changes
- **Presentation Mode**: Large-screen optimization features

### Collaboration Features
- **Multi-User Sessions**: Synchronized parameter changes across browsers
- **Presenter Controls**: Workshop leader override capabilities
- **Audience Interaction**: Real-time polling and feedback integration
- **Session Recording**: Capture and replay demonstration sessions

---

**The SDL Web Dashboard represents a breakthrough in system design education, providing an unparalleled interactive experience for understanding distributed systems performance.**
