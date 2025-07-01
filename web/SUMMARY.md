# SDL Web Dashboard Summary

**Version:** Unified Dashboard Layout with Recipe Integration (v3.2)

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
- **Communication**: 
  - Server Mode: gRPC-Web/Connect for API calls
  - WASM Mode: Direct WASM function calls
- **Code Editing**: Monaco Editor for SDL syntax highlighting
- **File Management**: FileClient interface for consistent file operations across modes

### Layout Architecture

#### Unified Dashboard Layout (v3.0)
The dashboard now features a unified layout that works seamlessly in both server and WASM modes:

**6-Panel Dockable System**
- **File Explorer Panel**: Browse and select SDL files
  - Server Mode: Limited file browsing (examples only)
  - WASM Mode: Full virtual filesystem with read/write capabilities
  - Tree view with folder expansion
  - Read-only file indicators
  
- **Code Editor Panel**: Monaco-powered SDL editing
  - Syntax highlighting for SDL language
  - Auto-save functionality in WASM mode
  - Read-only mode for example files
  
- **System Architecture Panel**: Enhanced system visualization with Graphviz/dot rendering
  - Displays complete system structure from Canvas API
  - Shows all components with per-method traffic values
  - Automatic layout via dot engine for clean presentation
  - Dynamically updates based on loaded SDL file
  
- **Traffic Generation Panel**: Generator controls with Start/Stop functionality
  - Real-time traffic control and monitoring
  - Add/remove generators dynamically
  - Rate adjustment with fine-grained control (0.1 RPS increments)
  
- **Console Panel**: System output and logs
  - Real-time console output capture
  - Error and warning highlighting
  - Command execution feedback
  
- **Live Metrics Panel**: Dynamic charts grid (tabbed with Console)
  - Unlimited scrollable charts supporting infinite metrics
  - 3-column responsive grid that adapts to screen size
  - Real-time chart updates with proper scaling

**DockView Features**
- **Fully Resizable**: Drag splitters to adjust panel sizes
- **Dockable Tabs**: Drag tabs to rearrange panel positions
- **Layout Persistence**: Custom layouts automatically saved to localStorage
- **Professional Styling**: Dark theme with blue highlights for active tabs
- **Reset Functionality**: One-click return to default 2x2 grid layout

### Recent Updates (v3.1)

#### Multi-Filesystem Support
- **MultiFSExplorer Component**: Manages multiple mounted filesystems
- **FileSystem Types**: Local (editable) and GitHub (read-only) 
- **Per-Filesystem Actions**: Add/Delete files, Refresh, with read-only indicators
- **Visual Hierarchy**: Collapsible file trees with folder/file icons

#### Tabbed Editor
- **TabbedEditor Component**: Replaces single-file editor
- **Multiple Files**: Open multiple files simultaneously in tabs
- **Modification Tracking**: Visual (*) indicator for unsaved changes
- **File Operations**: Save active tab, close with unsaved changes warning
- **Tab Uniqueness**: Composite keys (fsId:path) handle same-named files across filesystems
- **Filesystem Names**: Tab titles show filesystem:filename for clarity
- **Sync with Explorer**: Selecting tab highlights corresponding file in tree

#### Recipe Integration (v3.2)
- **Integrated Recipe Execution**: No separate Recipe Runner panel needed
- **Editor Toolbar**: Context-sensitive controls appear for .recipe files
- **Execution Controls**: Run, Step, Stop, Restart buttons in editor toolbar
- **Line Highlighting**: Current executing line highlighted with blue background
- **Tab Indicators**: Running recipes show ▶ in tab title
- **Console Output**: Recipe execution feedback in console panel
- **Debugger Experience**: Step through SDL commands with visual feedback

#### FileSystem Architecture (Completed)
- **FileSystemClient Interface**: Unified interface for all filesystem operations
- **Implementations**:
  - `LocalFileSystemClient`: Server-hosted filesystems via REST API
  - `GitHubFileSystemClient`: Read-only GitHub repository access
  - `IndexedDBFileSystemClient`: Browser-local storage using IndexedDB (planned)
- **Server Handler**: `/api/filesystems/{fsId}/{path}` with security and filtering
- **Security**: Path traversal protection, read-only enforcement
- **File Filtering**: Configurable extensions (`.sdl`, `.recipe`)

### Key Components

#### Dashboard Class (`src/dashboard.ts`)
- **Main Application Controller**: Manages the entire dashboard state and rendering
- **Unified Layout**: Single layout implementation for both server and WASM modes
- **FileClient Interface**: Uses abstraction for file operations
- **Dynamic Chart Management**: Creates and updates charts based on metrics
- **Component Creation**: Factory methods for all panel components
- **FileSystem Integration**: Uses FileSystemClient instances for all file operations

#### WASMDashboard Class (`src/wasm-dashboard.ts`)
- **WASM Extension**: Extends base Dashboard with WASM-specific features
- **WASMCanvasClient**: Replaces server API with WASM function calls
- **File Management**: Handles WASM virtual filesystem operations
- **Read-only Detection**: Special handling for bundled example files

#### CanvasClient (`src/canvas-client.ts`)
- **gRPC-Web Integration**: Uses Connect/gRPC-Web for server communication
- **FileClient Implementation**: Implements file operations for server mode
- **State Management**: Handles Canvas state and system operations
- **Generator Control**: Full generator lifecycle management
- **Metrics Streaming**: Real-time metric updates via streaming RPCs

#### WASMCanvasClient (`src/wasm-integration.ts`)
- **WASM Bridge**: Direct JavaScript-to-Go function calls
- **FileClient Implementation**: Virtual filesystem operations
- **Canvas Operations**: Load files, use systems, manage generators
- **Type Conversions**: Handles Go-to-JS type marshaling

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

## 🔒 Security Architecture

### FileSystem Security
- **Path Traversal Protection**: All file paths validated to prevent directory escape
- **Server-Side Control**: Server defines which directories are accessible
- **Read-Only Enforcement**: Filesystem-level write protection
- **File Type Filtering**: Only allowed extensions visible and accessible
- **Clean Separation**: Client doesn't know actual server paths, only filesystem IDs

### Technical Reliability
- **Production Ready**: Comprehensive testing ensures demo reliability
- **Cross-Platform**: Works on all modern browsers
- **Network Resilient**: Graceful handling of connection issues
- **Performance Optimized**: Smooth experience even under load

## 🌐 WASM Mode Features

### Browser-Based Runtime
- **Zero Server Cost**: Run simulations entirely in the browser
- **Instant Feedback**: No network latency for operations
- **Offline Capability**: Works without internet connection
- **Local File System**: Virtual filesystem for SDL projects

### Mode Selection
- **URL Parameter Based**: `?server=true` for server mode, default is WASM
- **Seamless Switching**: Same UI experience in both modes
- **Feature Parity**: Most features work identically
- **Performance Trade-offs**: WASM is slower but free

## 🔮 Future Enhancements

### WASM Optimization
- **Binary Size Reduction**: Target <10MB with TinyGo
- **Performance Improvements**: Web Workers for simulation
- **Progressive Loading**: Lazy load WASM modules
- **Caching Strategy**: Browser-based WASM caching

### Advanced Features
- **GitHub Integration**: Load examples directly from repos
- **Project Sharing**: Export/import SDL projects
- **Collaborative Editing**: Real-time multi-user support
- **Visual System Designer**: Drag-and-drop system composition

---

**The SDL Web Dashboard v3.0 with unified layout brings professional system modeling to everyone - free WASM mode for learning and experimentation, server mode for production workloads.**
