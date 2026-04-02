# Constraints

> Architectural rules for this project. Validated by `/stack-audit`.

## Constraints

### Panel/Page Interfaces Over WASM Types
**Rule**: Service-layer code must reference pages via Go interfaces (WorkspacePage), never via WASM-generated client types (WorkspacePageClient). Concrete WASM implementations (BrowserWorkspacePage) belong in `cmd/wasm/browser.go`. This enables alternative implementations (ConsoleWorkspacePage, test mocks) without WASM build tags.
**Why**: Coupling services to WASM types makes them untestable with `go test` and prevents reuse in non-browser contexts (CLI, server-side). The lilbattle project uses the same pattern.
**Verify**: `grep -rn 'wasmservices\.\|WorkspacePageClient\|DashboardPageClient' services/ --include='*.go' | grep -v '_test.go'`
**Scope**: services/

### No Canvas Types in Active Code
**Rule**: Canvas, CanvasService, CanvasViewPresenter are removed. All active code uses WorkspaceService, WorkspacePresenter, DevEnv. Canvas files are in `web/attic/`.
**Why**: Canvas was a pass-through to SystemInstance. DevEnv replaced it (#34). WorkspaceService is the unified service interface (#40).
**Verify**: `grep -rn 'CanvasService\|NewCanvas\b' services/ cmd/ web/server/ --include='*.go' | grep -v attic | grep -v gen/`
**Scope**: project-wide

### Service Backends Follow Lilbattle Pattern
**Rule**: Service backends live in sub-packages named by storage type: `devenvbe/` (local DevEnv), `connectclient/` (remote gRPC), `inmem/` (in-memory CRUD). All use proto request/response types.
**Why**: Consistent with lilbattle's `fsbe/`, `gormbe/`, `connectclient/` pattern. Enables shared CLI lib extraction.
**Verify**: manual
**Scope**: services/

### No Workarounds Without Root Cause
**Rule**: Always find the root cause of an issue before proposing a fix. Never create workarounds without asking.
**Why**: Workarounds accumulate and hide the real problem.
**Verify**: manual
**Scope**: project-wide
