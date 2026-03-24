
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
- In SDL system declaration you can declare the components in any order. There are no "set" statements. You pass the dependencies in the constructor of a "use" keyword.  For example:
```system Twitter {
    use app AppServer(db = database)
    use db Database
}```
- Here the AppServer component has a "db" dependency that is set by the "database" component declared in the next line.

## Available commands

- `buf generate`- To generate protos
- `make` - To generate all binaries
- `make dash` - To rebuild the web dashboard
- `make serve` - To start the server (go run cmd/sdl/main.go serve)
- `templar get` - Fetch vendored template dependencies (run from web/templates/)

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

## Stack Audit (March 2026)

GitHub issues created for stack alignment:
- #1: Migrate systems templates to namespace/extend pattern
- #2: Migrate systems listing to goapplib EntityListingData mixin
- #3: Use servicekit OutgoingMessage instead of custom CanvasWSMessage
- #4: Use ProtoJSON/TypedJSON codec for WebSocket
- #5: Use goapplib ViewContext instead of manual maps

## Architecture Direction: System/Canvas Consolidation

Analysis shows both systems detail page and canvas viewer page implement the same IDE-like dockview experience with different architectures:
- **Canvas viewer** uses the lilbattle-style MVP presenter pattern (Go presenter in WASM pushes UI updates via RPC)
- **System details** uses a frontend-orchestrated tool pattern (TS pulls data from WASM)
- Plan: consolidate to a single "Workspace" experience using the canvas viewer's presenter pattern
- See GitHub issues #1-#5 for incremental steps
