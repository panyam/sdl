# SDL WASM Support

This directory contains the WebAssembly implementation of SDL, enabling browser-based simulation without server infrastructure.

## Architecture

### Design Principles
- **Reuse existing codebase** - Avoid rewriting core logic
- **Server/WASM parity** - Same Canvas, Runtime, and SimpleEval
- **Clean separation** - Proto/gRPC only at service boundaries
- **Progressive enhancement** - Start simple, add features incrementally

### Key Components

1. **FileSystem Abstraction** (`loader/filesystem.go`)
   - Unified interface for file operations
   - Multiple backends: Local, HTTP, Memory, GitHub
   - Composite pattern for flexible mounting
   - Shared between server and WASM modes

2. **Canvas Refactoring**
   - Removing embedded proto types
   - Native Go types for Generator, Metric, SystemDiagram
   - Proto conversion in service layer only
   - Single Canvas implementation for both modes

3. **WASM API** (`wasm/cmd/main.go`)
   - Mirrors CLI commands in JavaScript
   - Canvas operations: load, use, info, reset
   - Generator management: add, update, remove, start, stop
   - File system access: read, write, list, mount

4. **Web Integration**
   - Extended dashboard with WASM mode toggle
   - File explorer for virtual filesystem
   - Monaco editor with SDL syntax highlighting
   - Reuses existing UI components

## Building

```bash
cd wasm
./build.sh
```

This creates:
- `web/sdl.wasm` - The WASM binary
- `web/wasm_exec.js` - Go's WASM support file

## Development Setup

1. **Start file server** (for development mode):
```bash
# Serves local SDL files for WASM to load
sdl serve-files --port 8081 --cors \
  --mount /examples=./examples \
  --mount /lib=./sdllib
```

2. **Run web dashboard**:
```bash
cd web
npm run dev
```

3. **Access WASM mode**:
```
http://localhost:5173/?wasm=true
```

## JavaScript API

```javascript
// Initialize WASM
const go = new Go();
const result = await WebAssembly.instantiateStreaming(
  fetch("sdl.wasm"), 
  go.importObject
);
go.run(result.instance);

// Use SDL API
SDL.canvas.load('/examples/uber.sdl');
SDL.canvas.use('UberMVP');
SDL.gen.add('api_load', 'api.handleRequest', 100);
SDL.run({ duration: '10s' });
```

## Current Status

### Working
- ✅ FileSystem abstraction with multiple backends
- ✅ WASM build infrastructure
- ✅ JavaScript API design
- ✅ Web component architecture

### In Progress
- 🚧 Canvas refactoring to remove proto dependencies
- 🚧 WASM module implementation
- 🚧 Dashboard integration

### TODO
- ⏳ SimpleEval WASM compatibility verification
- ⏳ Binary size optimization with TinyGo
- ⏳ Performance benchmarking
- ⏳ Example bundling strategy

## Limitations

1. **Binary Size**: Initial builds ~10-30MB (targeting <5MB)
2. **Performance**: 2-10x slower than native (acceptable for demos)
3. **Features**: No server persistence, metrics storage, or collaboration
4. **Memory**: Browser memory limits apply to large simulations

## Future Optimizations

1. **Rust Core**: Rewrite performance-critical parts in Rust
2. **Code Splitting**: Lazy load components as needed
3. **Web Workers**: Run simulations in background threads
4. **Caching**: Aggressive caching of compiled modules

## Testing

```bash
# Unit tests
go test ./wasm/...

# Integration tests
npm test

# Manual testing
1. Load example SDL file
2. Add generators
3. Run simulation
4. Verify results match server mode
```