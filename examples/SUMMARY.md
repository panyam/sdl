# SDL Examples Package Summary (`sdl/examples`)

**Purpose:**

This package contains example system models built using the SDL library. These examples demonstrate how to use the core primitives and components **via the Go API** to model real-world scenarios and analyze their performance. They serve as integration tests and usage guides for the library's Go interface.

**Key Examples:**

1.  **`bitly/`:**
    *   Models a simplified URL shortening service like Bitly.
    *   Components: `IDGenerator`, `Cache` (`components.Cache`), `DatabaseComponent` (wraps `components.HashIndex`), `BitlyService` (orchestrator).
    *   Demonstrates composing cache reads/misses with database operations (`Redirect` operation).
    *   Demonstrates simple sequential operations (`ShortenURL` operation).
    *   Uses `core.Analyze` for testing performance against expectations.

2.  **`gpucaller/`:**
    *   Models an application server making inference requests to a pool of GPUs, including batching.
    *   Components: `AppServer`, `Batcher` (`components.Batcher`), `GpuBatchProcessor` (implements `BatchProcessor` interface), stateless `ResourcePool` (`components.ResourcePool`).
    *   Defines a custom `gpuwork.go` profile for the batch processing time.
    *   Demonstrates using the stateless `ResourcePool` based on configured rates (`lambda`, `Ts`).
    *   Demonstrates using the `Batcher` component.
    *   Includes tests (`gpucaller_test.go`) that perform parameter sweeping (varying GPU pool size and QPS) using `core.Analyze` to evaluate SLOs under different loads.

3.  **`notifier/`:**
    *   Models a notification system with asynchronous fan-out.
    *   Components: `NotifierService`, `MessageStore` (using `HashIndex`), `InboxStore` (using `LSMTree`), `AsyncProcessor`, CDC delay simulation.
    *   Highlights the challenge of modeling variable fan-out (message delivery to N recipients). The `AsyncProcessor` currently uses a manual expansion/approximation.
    *   Demonstrates combining synchronous (`SendMessage`) and asynchronous (CDC delay + `ProcessMessage`) paths for end-to-end analysis.

**Current Status:**

*   Provides concrete usage examples for the **Go API** of the SDL library.
*   Demonstrates modeling different system patterns (caching, database interaction, resource pooling, batching, basic async flows).
*   Tests consistently use `core.Analyze` for verification against performance expectations.
*   These examples serve as valuable test cases for the `core` and `components` packages.
*   They could potentially be rewritten using the DSL once the parser and interpreter are complete, serving as target use cases for the DSL implementation.
