# Constraints

> Architectural rules for this project. Validated by `/stack-audit`.

## Constraints

### Panel/Page Interfaces Over WASM Types
**Rule**: Service-layer code must reference panels and pages via Go interfaces, never via WASM-generated client types. Concrete WASM implementations (forwarding interface calls to browser clients) belong in a `browser` module (`cmd/wasm/browser.go`). This enables alternative implementations (CLI console, test mocks, headless) without WASM build tags.
**Why**: Coupling services to WASM types makes them untestable with `go test` and prevents reuse in non-browser contexts (CLI, server-side). The lilbattle project hit this and adopted the same pattern.
**Verify**: `grep -rn 'wasmservices\.\|DevEnvPageClient\|DashboardPageClient' services/ --include='*.go' | grep -v '_test.go'`
**Scope**: services/

### No Canvas/CanvasService in New Code
**Rule**: New code must use DevEnv, not Canvas or SingletonCanvasService. Canvas remains for existing code paths but should not gain new callers.
**Why**: DevEnv replaces Canvas as the simulation coordinator (#34). Canvas is a pass-through to SystemInstance that adds unnecessary indirection.
**Verify**: manual
**Scope**: project-wide (new code only)

### No Workarounds Without Root Cause
**Rule**: Always find the root cause of an issue before proposing a fix. Never create workarounds without asking.
**Why**: Workarounds accumulate and hide the real problem.
**Verify**: manual
**Scope**: project-wide
