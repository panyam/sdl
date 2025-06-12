# SDL Examples Package Summary (`examples` package)

**Purpose:**

This package contains example system models with a primary focus on **workshop demonstrations and system design interview coaching**. The flagship Netflix streaming service demo (`netflix/`) provides a comprehensive scenario for teaching capacity modeling, queuing theory, and performance optimization. Legacy examples demonstrate the Go API usage, while new examples focus on creating compelling educational experiences for conference presentations and interview preparation.

**🎪 Workshop Demonstration Examples:**

1.  **`netflix/` - Netflix Streaming Service Demo (FLAGSHIP):**
    *   **Purpose:** Conference workshop demonstration of system design interview coaching with interactive simulations.
    *   **Components:** CDN with ResourcePool capacity modeling, VideoEncoder with queuing delays, VideoDatabase with connection pooling, LoadBalancer for traffic distribution.
    *   **Scenarios:** Complete traffic spike narrative from baseline performance → 4x load increase → cache optimization → capacity scaling → database bottlenecks → failure conditions.
    *   **Workshop Value:** Demonstrates M/M/c queuing theory, cache hit rate impact, capacity planning, and performance trade-offs through real-time parameter modification.
    *   **Files:** `netflix.sdl` (system model), `traffic_spike_demo_test.go` (Canvas-based test suite for validation).

**Additional Workshop Examples:**

2.  **`bitly/` - URL Shortening Service:**
    *   SDL model files for a Bitly-style URL shortening service.
    *   **Files:** `db.sdl`, `mvp.sdl` - Component and system definitions.
    *   **Workshop Potential:** Could be expanded for system design interview scenarios demonstrating database sharding, caching strategies, and global distribution.

3.  **`twitter/` - Social Media Platform:**
    *   SDL models for Twitter-style social media components.
    *   **Files:** `dbs.sdl`, `services.sdl` - Database and service architectures.
    *   **Workshop Potential:** Excellent for demonstrating timeline generation, fan-out patterns, and social graph challenges.

**Foundation SDL Files:**

*   **`common.sdl`**: Core native component signatures and types used across workshop examples. Defines components like `ResourcePool`, `Cache`, `HashIndex`, and enums like `HttpStatusCode`. Essential foundation for all workshop scenarios.
*   **`capacity.sdl`**: Demonstrates capacity modeling with ResourcePool components. Shows queuing delays and performance degradation patterns.
*   **`disk.sdl`**: Basic disk component model with latency distributions.
*   **`workshop.todo`**: Comprehensive development status and todo tracking for workshop preparation.

**Current Status:**

*   **Workshop-Ready:** Netflix demo provides complete conference presentation scenario with comprehensive test coverage.
*   **Educational Focus:** Examples designed specifically for system design interview coaching and interactive capacity modeling demonstrations.
*   **Foundation Complete:** Core SDL files (`common.sdl`, `capacity.sdl`) provide essential building blocks for workshop scenarios.
*   **Canvas Integration:** Netflix example fully integrated with Canvas API for real-time parameter modification and live demonstrations.
*   **Conference Preparation:** All examples validated for workshop use with edge case testing and rapid iteration support.

**Workshop Development Priorities:**

*   **Immediate:** Validate Netflix demo end-to-end, implement `sdl execute` and `sdl dashboard` commands for conference presentation.
*   **Next:** Expand Bitly and Twitter examples into full workshop scenarios with guided progression and interactive elements.
*   **Future:** Develop additional FAANG interview scenarios (Uber dispatch, Instagram feed, WhatsApp messaging) and automated workshop progression tools.
