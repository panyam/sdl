# SDL WASM Implementation

This directory contains the WebAssembly (WASM) implementation of SDL, allowing SDL simulations to run directly in web browsers without server-side infrastructure.

## Architecture

### FileSystem Abstraction
The WASM implementation uses the loader.FileSystem interface with WASM-specific implementations:

- **DevServerFS** - Fetches files from a development server using browser's fetch API
- **BundledFS** - Serves files embedded in the WASM binary at build time  
- **URLFetcherFS** - Fetches files from arbitrary URLs using fetch API
- **loader.MemoryFS** - In-memory filesystem for user edits (reused from loader package)
- **loader.CompositeFS** - Combines multiple filesystems with mount points (reused from loader package)

### Canvas Integration
The Canvas has been refactored to remove proto/gRPC dependencies:
- Native types created for Generator, Metric, SystemDiagram
- Canvas accepts a runtime parameter for dependency injection
- WASM creates custom runtime with FileSystem resolver
- Proto conversion happens only at the service boundary (server-side)

## Build Process

```bash
# Build WASM module
make build

# Output files:
# - web/sdl.wasm (28.6MB)
# - web/wasm_exec.js
```

## JavaScript API

The WASM module exposes a global `SDL` object that mirrors CLI commands:

```javascript
// Canvas operations
SDL.canvas.load(recipePath, canvasId)
SDL.canvas.use(systemName, canvasId)
SDL.canvas.info(canvasId)
SDL.canvas.list()
SDL.canvas.reset(canvasId)
SDL.canvas.remove(canvasId)

// Generator operations
SDL.gen.add(name, "component.method", rate, options)
SDL.gen.remove(name, options)
SDL.gen.update(name, rate, options)
SDL.gen.list(options)
SDL.gen.start(names, options)
SDL.gen.stop(names, options)

// File system operations
SDL.fs.readFile(path)
SDL.fs.writeFile(path, content)
SDL.fs.listFiles(directory)
SDL.fs.mount(prefix, url)

// Configuration
SDL.config.setDevMode(boolean)
```

## Development Setup

1. Start a development server for SDL files:
```bash
cd examples
python3 -m http.server 8081
```

2. Serve the web interface:
```bash
cd web
python3 -m http.server 8080
```

3. Open http://localhost:8080 in your browser

## Key Differences from Server Implementation

1. **No gRPC/Proto** - Uses native types throughout
2. **Browser FileSystem** - Uses fetch API instead of OS file operations
3. **Virtual Time** - Simulations run in virtual time, not real time
4. **Memory Constraints** - Limited by browser memory allocation
5. **No Server Metrics** - Metrics stored in-memory only

## Known Limitations

- Binary size: 28.6MB (working on optimization with TinyGo)
- No real-time streaming metrics (polling only)
- Limited to browser memory constraints
- No persistent storage (unless using IndexedDB)
- Cannot access local files directly (security restriction)

## Future Improvements

- TinyGo compilation for smaller binary size
- IndexedDB support for persistence
- Web Workers for background simulation
- Progressive loading of WASM chunks
- Service Worker for offline support
