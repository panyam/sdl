
## Understand the Project First
- I am continuing with a previous project.  You will find the summaries in SUMMARY.md files located in the top level as various sub folders.  NEXTSTEPS.md is used to note what has been completed and what are next steps in our roadmap.  Thorougly understand it and give me a recap so we can continue where we left off.

## Coding Style and Conservativeness
- Be conservative on how many comments are you are adding or modifying unless it is absolutely necessary (for example a comment could be contradicting what is going on - in which case it is prudent to modify it).
- When modifying files just focus on areas where the change is required instead of diving into a full fledged refactor.
- Make sure you ignore 'gen' and 'node_modules' as it has a lot of files you wont need for most things and are either auto generated or just package dependencies
- When updating .md files and in commit messages use emojis and flowerly languages sparingly.  We dont want to be too grandios or overpromising.
- Make sure the playwright tool is setup so you can inspect the browser when we are implementing and testing the Dashboard features.
- Do not refer to claude or anthropic or gemini in your commit messages
- Do not rebuild the server - it will be continuosly be rebuilt and run by the air configs.  Output of the server will be written to /tmp/sdlserver.log.  Build errors will also be shown in this log file.
- Find the root cause of an issue before figuring out a solution.  Fix problems.
- Do not create workarounds for issues without asking.  Always find the root cause of an issue and fix it.
- Do not use `as any` casts to suppress TypeScript errors.  Find and fix the root cause.
- The web module automatically builds when files are changed - DO NOT run npm build or npm run build commands.
- Proto files are automatically regenerated when changed - DO NOT run buf generate commands.

## WASM Build Memory
- You can build the wasm binary by simply doing `make wasmbin` in the top level directory but this is being built as files are changed.

## Continuous Builds

Builds for frontend, wasm, backend are all running continuously and can be queried using the `devloop` cli tool.   You have the following devloop commands:
- `devloop config` - Get configuration from running devloop server
- `devloop paths` - List all file patterns being watched
- `devloop trigger <rulename>` - Trigger execution of a specific rule
- `devloop logs <rulename>`  - Stream logs from running devloop server
- `devloop status <rulename>` - Get status of rules from running devloop server

## Summary instructions

- When you are using compact, please focus on test output and code changes
- For the ROADMAP.md always use the top-level ./ROADMAP.md so we have a global view of the roadmap instead of being fragemented in various folders.

## SDL Demo Guidelines
- Make sure when you create SDL demos they are not as markdown but as .recipe files that are executable with pause points that print out what is going to be come next before the SDL command is executed.

## Session Workflow Memories
- When you checkpoint update all relevant .md files with our latest understanding, statuses and progress in the current session and then commit.

## SDL System Declaration Notes
- Systems declare typed parameters referencing component types. Components handle all composition via `uses`. The `use` keyword has been removed from the language.
- Example:
```
component TwitterArch {
    uses app AppServer(db = database)
    uses database Database()
}

system Twitter(arch TwitterArch) {
}
```
- Here `TwitterArch` is a component that composes `AppServer` and `Database`. The system takes it as a parameter.
- `uses x Foo()` (with empty parens) creates a default instance. `uses x Foo(dep = y)` wires dependencies.
- `uses x Foo` (no parens) means the dependency must be provided by a parent component.

## Available commands

- `cd protos && make buf` - To generate protos (or `make buf` from top level)
- `make` or `make all` - Build everything (order: parser -> WASM -> dash -> binary -> run)
- `make dash` - To rebuild the web dashboard
- `make serve` - To start the server (go run cmd/sdl/main.go serve)
- `make webtest` - Run web unit tests (vitest)
- `make wasmbin` - Build WASM binaries
- `templar get` - Fetch vendored template dependencies (run from web/templates/)

## Build System Gotchas

- **Build order matters**: `make all` runs parser -> wasmbin -> dash -> binary -> run. WASM must build before dash so `update-template-assets.js` can discover `.wasm` files in `dist/wasm/`.
- **Unified dist/**: All build outputs go to `<project>/dist/` (not `web/dist/`). Vite outputs to `../dist`, WASM builds to `dist/wasm/`, server serves from `./dist/`.
- **Asset cache busting**: Vite generates `.vite/manifest.json` in dist/. The Go server reads it at startup and provides `viteJS`/`viteCSS` template functions. No post-build script needed — templates use `{{ viteJS "index.html" }}` and `{{ viteCSS "index.html" }}`.
- **Tailwind content paths**: `web/tailwind.config.js` must scan `./templates/**/*.html` (not `../templates/`). Wrong paths cause Tailwind to purge all utility classes used in Go templates.
- **Test directory**: Tests live in `web/src/.__tests__/` (with dot prefix). Vitest config must reference `.__tests__` not `__tests__`.
- **Pre-push hook**: `.git/hooks/pre-push` runs web tests before push. Go tests have pre-existing failures so are not included yet.

## Template System (Templar)

- Templates use templar vendoring with `@source/` prefix syntax for external dependencies
- Config: `web/templates/templar.yaml` defines sources (currently goapplib)
- Vendored templates: `web/templates/templar_modules/` (checked into git, like go vendor/)
- To update vendored templates: `cd web/templates && templar get`
- Template references must use `@goapplib/...` prefix (not bare `goapplib/...`)
- Go server uses `tmplr.NewSourceLoaderFromConfig()` to resolve `@source/` paths
- SDL BasePage.html extends GoalBase:BasePage — all pages must define `PostBodySection` block (even if empty)

## Template Block Names (SDL BasePage)

The SDL BasePage extends goapplib BasePage with these mappings:
- `GoalBase:CSSSection` -> `SDLCSSSection`
- `GoalBase:HeaderSection` -> `SDLHeaderSection`
- `GoalBase:BodySection` -> `BodySection`
- `GoalBase:AppContainerSection` -> `SDLAppContainerSection`
- `GoalBase:AppScriptSection` -> `SDLAppScriptSection`
- `GoalBase:PostScriptsSection` -> `PostBodySection`

Pages that include BasePage.html must define: `BodySection`, `PostBodySection`, optionally `ExtraHeadSection`.

## tsappkit (TypeScript dependency)

- Located at `~/newstack/goapplib/main/tsappkit/`
- Linked via pnpm in web/package.json as `@panyam/tsappkit`
- Symlinked through node_modules to the goapplib repo
- Must be built (`cd ~/newstack/goapplib/main/tsappkit && pnpm build`) if dist/ is missing
- Provides: EventBus, BasePage, LCMComponent, ThemeManager, LifecycleController

## EntityListingData URL Formats

- goapplib EntityListingData uses `fmt.Sprintf(ViewUrlFormat, id)` for links
- URL format strings must contain `%s` placeholder for the entity ID
- Example: `/canvases/%s/view` not `/canvases`

## Value Type System (lib/decl)

- Type names are lowercase: `"bool"`, `"int"`, `"float"`, `"string"`, `"nil"`
- `Value.String()` format: `RV(type: value)` — e.g., `RV(int: 10)`
- List type string uses brackets: `List[int]`, `Outcomes[string]`
- `NilType` is a `SimpleType` (tag `TypeTagSimple`), not `TypeTagNil` — check with `r.Type == NilType`, not `r.Type.Tag == TypeTagNil`
- `Value.Deref()` returns `(r, error)` not `(nil, error)` for nil values — prevents nil pointer panics in callers

## Flow Solver Convergence

- Both string-based and runtime-based solvers use fixed-point iteration with damping
- Parameters: 30 max iterations, 0.01 convergence threshold, 0.3 damping factor (runtime) / 0.7 update rate (string-based)
- Known limitation: flow evaluator doesn't apply native component outcomes (e.g., cache hit rate) to branch weighting in if/else

## Test Status (lib/)

- `lib/decl`, `lib/runtime`, `lib/parser`, `lib/loader` — all passing
- `lib/components` — pre-existing failures (HDD/SSD disk profile differentiation)
- `lib/core` — pre-existing failure (distribution edge case)
- Runtime tests use `sys.FindComponent("arch.X")` to access components inside system parameters

## Stack Audit (March 2026)

GitHub issues created for stack alignment:
- #1: Migrate systems templates to namespace/extend pattern
- #2: Migrate systems listing to goapplib EntityListingData mixin
- #3: Use servicekit OutgoingMessage instead of custom CanvasWSMessage
- #4: Use ProtoJSON/TypedJSON codec for WebSocket
- #5: Use goapplib ViewContext instead of manual maps

## Dead Code and Attic

- Dead/superseded code is moved to `web/attic/` (not deleted) for reference
- Includes: dashboard.ts monolith, system-details-page, old panel system, dashboard-coordinator, app-state-manager
- Dead tests moved to `web/attic/src/__tests__/`
- When removing code, always move to attic especially if it existed in previously working versions

## DockView Theming

- Container must have `dockview-theme-dark` or `dockview-theme-light` class
- Use MutationObserver on `document.documentElement` class changes to toggle theme
- CSS variables defined in `web/src/style.css` (not inline in templates)
- Panel content containers need explicit `bg-white dark:bg-gray-900` (dockview only themes chrome, not content)
- For pages with header: use flex fill pattern (`flex: 1 1 0%; min-height: 0`) not `position: absolute` — absolute covers the header
- See goapplib/tsappkit `BESTPRACTICES.md` for the canonical dockview pattern

## Monaco Editor Theming

- Monaco theme is global (`monaco.editor.setTheme()`), not per-editor
- Define light+dark variants for custom languages: `sdl-dark`/`sdl-light`, `recipe-dark`/`recipe-light`
- Don't hardcode `editor.background` in custom themes — inherit from `vs`/`vs-dark` base
- Known issue: theme toggle doesn't update Monaco editors (needs investigation)

## Architecture: Workspace vs Canvas vs Design

- **Workspace** = project/repo (e.g., "Uber") — source files, designs, import sources, manifest (`sdl.json`)
- **Design** = one system architecture (e.g., "UberMVP") — a `system` block in an SDL file
- **Canvas** = runtime/VM that executes a design — generators, metrics, flow analysis
- Analogy: Workspace = git repo, Design = source file, Canvas = running process
- Proto: `Workspace` (project metadata), `Canvas` (runtime state), `WorkspaceDesign` (design entry)
- Future: `CompilationUnit` artifact between Workspace and Canvas (see #19)

## WorkspaceService (lilbattle pattern)

- Interface: `services/workspace_service.go` — CRUD for workspaces + design content
- Backend: `services/backend_workspace_service.go` — wraps `WorkspaceStorageProvider`
- In-memory storage: `services/inmem/workspace_storage.go` — seeded from examples/
- Manifest: `services/workspace.go` — `LoadWorkspaceManifest()` parses `sdl.json` via protojson
- Example manifests: `examples/uber/sdl.json`, `examples/bitly/sdl.json`
- Services layer works with protos directly — no custom Go types duplicating proto fields

## ScriptTagFS (SDL embedding in pages)

- Server embeds SDL files in hidden `<textarea class="sdl-design-source">` elements
- WASM `ScriptTagFS` (cmd/wasm/filesystem.go) reads from DOM via `querySelector`
- Mounted at `/designs/` in WASM CompositeFS
- Uses textareas not script tags — Go templates HTML-escape script tag content
- Flow: page load → discover textareas → `fileSelected("/designs/X.sdl")` → ScriptTagFS reads DOM → parser runs

## CompositeFS Prefix Stripping

- `CompositeFS.findFS()` strips mount prefix before passing to underlying FS
- `@stdlib/common.sdl` → MemoryFS receives `common.sdl`
- Exception: URL-based mounts (`://`, `.com/`) keep full path for their FS implementations
- This is critical for `@stdlib/` imports to resolve correctly

## Import Standardization

- All SDL files should use `@stdlib/common.sdl` (not relative `../common.sdl`)
- Named source prefixes (`@name/`) preferred over relative imports
- Relative imports not forbidden, but discouraged — makes modules harder to relocate

## Architecture Direction: Workspace Consolidation

GitHub issues (Phases 1-3 complete, 4+5 in PR #17):
- #6: Phase 1 — Clean foundation (PR #12, merged)
- #7: Phase 2 — Route consolidation (PR #13, merged)
- #8: Phase 3 — Unified landing page (PR #14, merged)
- #9/#10: Phase 4+5 — Workspace proto + manifests (PR #15, merged)
- #11: Phase 6 — Diagram upgrade (Vis.js)
- #16: Design selector UI (PR #17, in progress)
- #19: CompilationUnit — clean Workspace→AST→Canvas separation (high priority, future)

## Conference Demos

- Uber architecture evolution: `examples/uber/{mvp,intermediate,modern}.{sdl,recipe}`
- Run via CLI: `SDL_CANVAS_ID=ubermvp sh ./examples/uber/mvp.recipe`
- Three terminals + three browser tabs for side-by-side comparison
- Last known working commit for demo: `13a91ef` (before goapplib migration)
