# SDL Project Next Steps

## Immediate Priority: WASM Support (In Progress)

### 1. Canvas Refactoring ✅ COMPLETED
- [x] Define native types for Generator, Metric, SystemDiagram in console package
- [x] Update GeneratorInfo to use native fields instead of embedded proto
- [x] Update MetricSpec to use native fields instead of embedded proto
- [x] Create conversion utilities (toProto/fromProto) in service.go and conversions.go
- [x] Update Canvas methods to use native types
- [x] Test Canvas works without proto dependencies
- [x] Remove DuckDB dependency (unused, was blocking WASM)

### 2. WASM Runtime Compatibility ✅ COMPLETED
- [x] Verify FileSystem abstraction works in WASM
- [x] Update Canvas to accept custom runtime with FileSystem resolver
- [x] Build minimal WASM binary with just runtime components (28.6MB)
- [x] Test basic SDL loading in browser
- [x] Fix Go WASM slice marshaling issue ([]string to []interface{})
- [x] Clean up duplicate FileSystem implementations (removed ~150 lines)
- [ ] Optimize binary size (current: 28.6MB, target: <10MB with TinyGo)

### 3. Web Integration ✅ COMPLETED
- [x] Complete file explorer integration with WASM filesystem
- [x] Wire up Monaco editor with SDL syntax highlighting
- [x] Connect load/run buttons to WASM API
- [x] Unified dashboard layout for both server and WASM modes
- [x] Implement mode toggle (server vs WASM) - URL parameter based (?server=true)
- [x] FileClient interface adopted for consistent file operations
- [ ] Add simulation results visualization (metrics streaming)

### 4. WASM Optimization
- [ ] Measure initial binary size with standard Go
- [ ] Try TinyGo build and document compatibility issues
- [ ] Implement lazy loading for WASM modules
- [ ] Add progress indicators for long simulations
- [ ] Cache compiled WASM binaries

## Medium Term Goals

### 5. Enhanced WASM Features
- [ ] Import/export SDL projects as ZIP
- [ ] GitHub integration for loading examples
- [ ] Shareable playground URLs
- [ ] Offline mode with service worker
- [ ] Performance profiling in browser

### 6. Demo Content
- [ ] Create interactive tutorials
- [ ] Port all examples to WASM-compatible format
- [ ] Add guided walkthroughs
- [ ] Create performance comparison demos
- [ ] Build showcase gallery

### 7. Production WASM Deployment
- [ ] Set up CDN for WASM binary
- [ ] Implement versioning strategy
- [ ] Add telemetry for usage analytics
- [ ] Create deployment pipeline
- [ ] Document WASM limitations clearly

## Long Term Vision

### 8. Rust Core Investigation
- [ ] Prototype SimpleEval in Rust
- [ ] Benchmark Rust vs Go performance
- [ ] Design WASM interface for hybrid approach
- [ ] Evaluate binary size improvements
- [ ] Create migration plan if beneficial

### 9. Advanced Features
- [ ] Multi-canvas support in WASM
- [ ] Collaborative editing via WebRTC
- [ ] Visual system designer
- [ ] AI-assisted system modeling
- [ ] Performance optimization suggestions

### 10. Enterprise Features
- [ ] Private WASM deployments
- [ ] Custom component libraries
- [ ] Integration with monitoring tools
- [ ] Export to Terraform/K8s
- [ ] Compliance and audit trails

## Technical Debt

### Code Quality
- [ ] Add comprehensive WASM tests
- [ ] Document WASM architecture
- [ ] Create WASM development guide
- [ ] Standardize error handling
- [ ] Improve type safety

### Performance
- [ ] Profile WASM execution
- [ ] Optimize hot paths
- [ ] Implement web workers for simulation
- [ ] Add simulation progress API
- [ ] Memory usage optimization

## Completed Recently
- ✅ Created FileSystem abstraction in loader package
- ✅ Designed WASM module structure
- ✅ Created web components (file explorer, code editor)
- ✅ Established WASM build process
- ✅ Identified Canvas refactoring approach
- ✅ Refactored Canvas to use native types instead of proto
- ✅ Created proto/native conversion functions at service boundaries
- ✅ Removed DuckDB dependency completely
- ✅ Successfully built WASM binary (28.6MB)
- ✅ Updated MetricTracer to use native types
- ✅ Fixed Go WASM []string marshaling issue (must convert to []interface{})
- ✅ Modified Canvas to accept runtime parameter for dependency injection
- ✅ Integrated WASM FileSystem with Canvas via custom runtime
- ✅ Created working test page demonstrating SDL loading in browser
- ✅ Cleaned up duplicate FileSystem implementations in WASM:
  - Removed duplicate FileSystem interface, CompositeFS, and MemoryFS
  - Kept WASM-specific implementations (DevServerFS, URLFetcherFS, BundledFS)
  - Removed redundant WASMHTTPFileSystem
  - All WASM code now uses loader.FileSystem interface
- ✅ Unified dashboard layout implementation (June 2025):
  - Refactored base Dashboard class to support both server and WASM modes
  - Implemented FileClient interface for consistent file operations
  - WASMDashboard now extends base Dashboard with minimal overrides
  - Mode detection via URL parameters (?server=true for server mode)
  - Unified panel layout: File Explorer | Code Editor | System Architecture | Traffic Generation | Console/Metrics

## Blocked/Decisions Needed
- ~~Proto dependency in Canvas blocking WASM build~~ ✅ RESOLVED
- Need to decide on bundled examples strategy
- TinyGo compatibility unknowns
- Binary size targets for free tier (current: 28.6MB)

## Notes
- Focus on developer experience for WASM mode
- Keep server mode as premium offering
- Ensure feature parity where reasonable
- Document limitations clearly
- Plan for graceful degradation