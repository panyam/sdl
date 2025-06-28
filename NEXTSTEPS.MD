# SDL Project Next Steps

## Immediate Priority: WASM Support (In Progress)

### 1. Canvas Refactoring (Current Focus)
- [ ] Define native types for Generator, Metric, SystemDiagram in console package
- [ ] Update GeneratorInfo to use native fields instead of embedded proto
- [ ] Update MetricSpec to use native fields instead of embedded proto
- [ ] Create conversion utilities (toProto/fromProto) in service.go
- [ ] Update Canvas methods to use native types
- [ ] Test Canvas works without proto dependencies

### 2. WASM Runtime Compatibility
- [ ] Verify SimpleEval works in WASM (no file I/O, no OS-specific calls)
- [ ] Update Canvas to use FileSystem interface for all file operations
- [ ] Build minimal WASM binary with just runtime components
- [ ] Test basic SDL evaluation in browser

### 3. Web Integration
- [ ] Complete file explorer integration with WASM filesystem
- [ ] Wire up Monaco editor with SDL syntax highlighting
- [ ] Connect load/run buttons to WASM API
- [ ] Add simulation results visualization
- [ ] Implement mode toggle (server vs WASM)

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

## Blocked/Decisions Needed
- Proto dependency in Canvas blocking WASM build
- Need to decide on bundled examples strategy
- TinyGo compatibility unknowns
- Binary size targets for free tier

## Notes
- Focus on developer experience for WASM mode
- Keep server mode as premium offering
- Ensure feature parity where reasonable
- Document limitations clearly
- Plan for graceful degradation