# SDL Web Dashboard - Next Steps

## Completed (March 2026 Session)

### Build Fixes
- Fixed `make dash` build: tsappkit dist/ was missing, built from source
- Fixed `update-template-assets.js` path resolution (was `../../` should be `../`)
- Fixed stale `loadedFiles` reference on Canvas proto (field doesn't exist)

### Templar Vendoring Migration
- Replaced stale `goapplib` symlink with proper templar vendoring
- Created `web/templates/templar.yaml` with goapplib source config
- Ran `templar get` to fetch templates into `templar_modules/`
- Updated template references: `goapplib/...` -> `@goapplib/...`
- Updated Go server to use `tmplr.NewSourceLoaderFromConfig()` instead of `goal.SetupTemplates()`
- Systems handler now shares the same SourceLoader-backed template group

### Template Fixes
- Fixed `safeHTMLAttr` in goapplib to accept `any` (was crashing on nil values)
- Fixed `EntityListingData` URL formats: added `%s` placeholder for canvas IDs
- Added missing `PostBodySection` block to SystemDetailsPage and CanvasViewerPage templates
- Fixed CanvasViewerPage template block names to match SDL BasePage mappings

### Stack Audit
- Created 5 GitHub issues (#1-#5) for stack alignment improvements
- Identified system/canvas duplication and consolidation path

## In Progress

### Get Pages Fully Rendering
- Systems listing page: rendering
- System details page: needs PostBodySection fix verification
- Canvas listing page: rendering, links now work
- Canvas viewer page: needs PostBodySection fix verification

## Next Priorities

### 1. System/Canvas Consolidation Planning
- Both use dockview + Monaco + Graphviz but with different architectures
- Canvas viewer has the better pattern (MVP presenter via WASM, like lilbattle)
- System details page uses frontend-orchestrated tool pattern
- Consolidate to single "Workspace" experience

### 2. Stack Alignment (GitHub Issues #1-#5)
- #1: Systems templates to namespace/extend
- #2: Systems listing to EntityListingData mixin
- #3: ServiceKit OutgoingMessage
- #4: WebSocket codec upgrade
- #5: ViewContext for systems handler

### 3. Complete WASM Integration
- Verify canvas viewer WASM mode works end-to-end
- System details WASM mode with SystemDetailTool
- Recipe execution in browser

### 4. Legacy Code Cleanup
- Remove `dashboard.ts` monolithic controller (replaced by DashboardCoordinator)
- Remove `dashboard-coordinator.ts` (replaced by canvas viewer presenter)
- Clean up stale `CanvasState` interface in `types.ts`
