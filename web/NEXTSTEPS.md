# SDL Web Dashboard - Next Steps

## Completed (March 2026)

### Build Infrastructure
- Fixed `make dash` build (tsappkit dist/, update-template-assets.js paths)
- Unified dist directory — all outputs to `<project>/dist/`
- Fixed Makefile build order: parser -> WASM -> dash -> binary -> run
- Fixed Tailwind content paths (`./templates/` not `../templates/`)
- Fixed vitest config path (`.__tests__` not `__tests__`)
- Added `make webtest` target and pre-push hook
- Replaced update-template-assets.js with Vite manifest + viteJS/viteCSS template functions

### Phase 1: Clean Foundation (Issue #6, PR #12)
- Moved dead code to `web/attic/`
- Fixed DockView theme, layout (flex fill not absolute), CSS variables
- 9 Phase 1 verification tests

### Phase 2: Route Consolidation (Issue #7, PR #13)
- Routes unified under `/workspaces/`
- Nav: "Workspaces" + "Examples"

### Phase 3: Unified Landing Page (Issue #8, PR #14)
- Examples + user workspaces on one page
- "Open" button auto-creates canvas from workspace

### Phase 4+5: Workspace Proto + Manifests (Issue #9/#10, PR #15)
- Workspace, WorkspaceDesign, ImportSource proto messages
- sdl.json manifest format (protojson)
- SystemCatalogService replaced by WorkspaceService (lilbattle pattern)
- All Uber imports standardized to @stdlib/

### Phase 4+5 continued: Design Selector (Issue #16, PR #17)
- WorkspaceService: interface + BackendWorkspaceService + inmem storage
- Canvas struct embeds Canvas proto directly (no duplicated fields)
- ScriptTagFS: WASM resolver that reads SDL from DOM textareas
- CompositeFS prefix stripping fix for @stdlib/ resolution
- Design selector dropdown in workspace toolbar
- Auto-loads SDL from embedded textareas on page open
- Diagram renders for selected design

### Known Issues
- Monaco editor theme toggle not working at runtime
- Editor panel doesn't show SDL source for selected design
- Go test suite has pre-existing failures
- Console/Generators/Metrics panels empty (no generators configured from UI yet)

## Next

### CompilationUnit (#19, high priority)
- Clean separation: Workspace (source) → CompilationUnit (AST) → Canvas (runtime)
- Canvas receives compiled artifact, never touches files/imports
- Supersedes #18 (design-canvas decoupling)

### Phase 6: Diagram Upgrade (#11)
- Replace Graphviz with Vis.js Network (interactive)

### Remaining polish
- Wire editor panel to show SDL source for selected design
- Fix Monaco theme toggle
- Add generator/metrics config from UI (make SDL declarative, no recipes needed)
