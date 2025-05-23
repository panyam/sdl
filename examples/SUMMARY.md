# SDL Examples Package Summary (`examples` package)

**Purpose:**

This package contains example system models built using the SDL library. These examples primarily demonstrate how to use the **Go API** of the `core` and `components` packages to model real-world scenarios and analyze their performance. They serve as integration tests and usage guides for the library's Go interface. The package also includes some sample `.sdl` files that can be used for testing the parser and loader.

**Key Examples (Go API):**

1.  **`bitly/` (in `examples/native/bitly`):
    *   Models a simplified URL shortening service like Bitly.
    *   Components: `IDGenerator`, `Cache` (`components.Cache`), `DatabaseComponent` (wraps `components.HashIndex`), `BitlyService` (orchestrator).
    *   Demonstrates composing cache reads/misses with database operations (`Redirect` operation).
    *   Demonstrates simple sequential operations (`ShortenURL` operation).
    *   Uses `core.Analyze` for testing performance against expectations.

2.  **`gpucaller/` (in `examples/native/gpucaller`):
    *   Models an application server making inference requests to a pool of GPUs, including batching.
    *   Components: `AppServer`, `Batcher` (`components.Batcher`), `GpuBatchProcessor` (implements `BatchProcessor` interface), stateless `ResourcePool` (`components.ResourcePool`).
    *   Defines a custom `gpuwork.go` profile for the batch processing time.
    *   Demonstrates using the stateless `ResourcePool` based on configured rates (`lambda`, `Ts`).
    *   Demonstrates using the `Batcher` component.
    *   Includes tests (`gpucaller_test.go`) that perform parameter sweeping (varying GPU pool size and QPS) using `core.Analyze` to evaluate SLOs under different loads.

3.  **`notifier/` (in `examples/native/notifier`):
    *   Models a notification system with asynchronous fan-out.
    *   Components: `NotifierService`, `MessageStore` (using `components.HashIndex`), `InboxStore` (using `components.LSMTree`), `AsyncProcessor`, CDC delay simulation.
    *   Highlights the challenge of modeling variable fan-out (message delivery to N recipients). The `AsyncProcessor` currently uses a manual expansion/approximation.
    *   Demonstrates combining synchronous (`SendMessage`) and asynchronous (CDC delay + `ProcessMessage`) paths for end-to-end analysis.

**Sample SDL Files:**

*   **`common.sdl`**: Defines a set of "native" component signatures (e.g., `NativeDisk`, `HashIndex`, `Cache`, `MM1Queue`) and a global enum `HttpStatusCode`. These native components are intended to be backed by Go implementations (like those in the `components` package).
*   **`bitly.sdl`**: An example SDL file that imports `common.sdl` and defines a `Disk` component (overriding/detailing a native one), a `Database` component, an `AppServer` component, and a `Bitly` system. It also shows an `AppServerWithCache` and `BitlyWithCache` system using an imported `Cache`.
    *   This file is used in `loader/loader_test.go` (`TestBitly`) to test parsing, loading, import resolution, and type inference.

**Current Status:**

*   Provides concrete usage examples for the **Go API** of the SDL library.
*   Demonstrates modeling different system patterns (caching, database interaction, resource pooling, batching, basic async flows) using the Go primitives.
*   Go API tests consistently use `core.Analyze` for verification against performance expectations.
*   The `.sdl` files serve as initial test cases for the DSL parser, loader, and the developing type inference system.

**Next Steps (for this package):**

*   Expand the Go API examples as new `core` or `components` features are added.
*   Develop more complex `.sdl` files to further test the DSL parser, loader, type inference, and eventually the DSL VM.
*   Potentially refactor the Go API examples into DSL equivalents once the DSL and VM are mature, to serve as benchmarks and demonstrate DSL capabilities.
