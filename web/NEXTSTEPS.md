# SDL Web Dashboard - Next Steps

## Completed (March 2026)

### Build Infrastructure
- Fixed `make dash` build (tsappkit dist/, update-template-assets.js paths)
- Unified dist directory — all outputs to `<project>/dist/`
- Fixed Makefile build order: parser -> WASM -> dash -> binary -> run
- Fixed Tailwind content paths (`./templates/` not `../templates/`)
- Fixed vitest config path (`.__tests__` not `__tests__`)
- Added `make webtest` target and pre-push hook

### Templar Vendoring Migration
- Replaced goapplib symlink with templar vendoring (`@goapplib/` prefix)
- Go server uses `tmplr.NewSourceLoaderFromConfig()`

### Phase 1: Clean Foundation (Issue #6, PR #12)
- Moved 15 dead source files + 6 dead test files to `web/attic/`
- Fixed DockView: theme class, CSS variables in style.css, flex fill layout (not absolute)
- Added MutationObserver for theme toggle on dockview container
- Added light/dark theme variants for Monaco editor (sdl-light, recipe-light)
- System details page redirects to canvas viewer (will 404 until Phase 3)
- Nil check on Canvas response to prevent panic
- 9 Phase 1 verification tests + 14 EventBus tests (all passing)

### Known Issues
- Monaco editor theme toggle not working at runtime (global setTheme called but doesn't update)
- System detail redirect shows 404 (no canvas exists for system IDs — needs Phase 3 Fork)
- Go test suite has pre-existing failures (not from our changes)

### Phase 2: Route Consolidation (Issue #7, PR #13)
- Routes unified under `/workspaces/` (listing, view, edit, create)
- `/canvases/*` redirects to `/workspaces/*`
- `/` redirects to `/workspaces/`
- Nav: "Workspaces" + "Examples"
- Replaced update-template-assets.js with Vite manifest + viteJS/viteCSS template functions
- Proto stays `Canvas` — workspace is UI naming only

## Current: Phase 3 — Unified Landing Page (Issue #8)
- Single listing: examples + user workspaces
- "Fork" button creates workspace from example

### Phase 4: Multi-Design UI (Issue #9)
- Design selector within workspace (uber-mvp, uber-v2, uber-modern)
- Backend already supports multiple systems per canvas

### Phase 5: Module/Import System (Issue #10)
- `sdl.yaml` workspace manifest with versioned sources
- Build on existing CompositeFS mount architecture

### Phase 6: Diagram Upgrade (Issue #11)
- Replace Graphviz with Vis.js Network (interactive)
- Phases 4-6 can be done in parallel
