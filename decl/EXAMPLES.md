
# SDL DSL Examples

This document provides examples of how the System Design Language (SDL) DSL is intended to be used for modeling and analyzing system performance. These examples showcase the syntax for defining components, composing systems, specifying actions, and setting performance expectations.

**Note:** The DSL evaluator (`sdl/decl/eval.go`) is currently under development based on the "Operator Tree" strategy. These examples represent the target language structure and semantics, but the underlying execution logic is not yet complete.

## Example 1: Basic Disk I/O Analysis

This example defines a simple system with a single SSD Disk and analyzes the performance of its `Read` operation.

```dsl
// Define components (can be in separate files via import)
// Note: For simplicity here, we don't explicitly define methods like Read/Write
// in the component DSL, assuming the evaluator maps instance.Read() calls
// to the underlying Go component's Read() method later.
component Disk {
    param ProfileName: string = "SSD"; // Default parameter
    // Other params like PageSize, RecordSize could be added
}

// Define the system configuration
system BasicDiskSystem {
    // Instantiate a Disk component, named 'mySSD'.
    // It implicitly uses the default ProfileName="SSD".
    instance mySSD: Disk;

    // Analyze the performance of reading from the disk instance.
    analyze SSD_Read_Perf = mySSD.Read() expect {
        // Assertions about the expected performance metrics.
        // 'result' is the implicit name for the analysis outcome wrapper.
        result.Availability > 0.99;    // Expect > 99% availability
        result.MeanLatency < 0.2ms;   // Expect mean latency under 0.2 milliseconds
        result.P99 < 1ms;             // Expect 99th percentile latency under 1 millisecond
    }

    // Analyze the Write performance with different overrides
    instance myHDD: Disk = { ProfileName = "HDD"; }; // Override profile
    analyze HDD_Write_Perf = myHDD.Write() expect {
        result.Availability > 0.98;
        result.MeanLatency > 5ms;     // HDDs are slower
        result.P99 < 200ms;
    }
}
```

**Explanation:**

*   We declare a conceptual `Disk` component with a configurable `ProfileName`.
*   In `BasicDiskSystem`, we create two instances: `mySSD` (using defaults) and `myHDD` (overriding `ProfileName`).
*   Two `analyze` blocks trigger performance evaluations of `mySSD.Read()` and `myHDD.Write()`.
*   The `expect` blocks specify performance assertions using metrics like `Availability`, `MeanLatency`, `P99`. The evaluator (once built) will compare calculated metrics against these thresholds.

---

## Example 2: Component Interaction (Service using Disk)

This example shows a service component that depends on (`uses`) a Disk component.

```dsl
component Disk {
    param ProfileName: string = "SSD";
}

// A service that performs a write then a read on a disk.
component SimpleService {
    // Declares a dependency on a component instance of type Disk, named 'db'.
    // This 'db' name is used within the methods of SimpleService.
    uses db: Disk;

    // A method performing sequential operations.
    method WriteThenRead(data: string): bool {
        writeOutcome = db.Write(); // Call the underlying Write operation
        // The V4 evaluation model implicitly tracks latency accumulation.
        // We only need to handle the logical success/failure if necessary.
        if !writeOutcome.Success {
            log "Write failed, skipping read"; // Example logging
            return false;
        }
        // If write succeeds, proceed to read. Latency combines implicitly.
        readOutcome = db.Read();
        return readOutcome.Success; // Return success status of the read
    }
}

system ServiceSystem {
    // Instantiate the required components
    instance theDisk: Disk = { ProfileName = "SSD"; };

    // Instantiate SimpleService. The system needs to satisfy the 'uses db: Disk' dependency.
    // Explicitly wiring 'theDisk' instance to the 'db' dependency.
    // (Syntax for injection TBD, this is one possibility).
    instance theService: SimpleService = { db = theDisk; };

    // Analyze the combined operation
    analyze Service_WriteRead = theService.WriteThenRead("some_data") expect {
        // Expectations on the *combined* latency and availability.
        result.Availability > 0.97; // Slightly lower due to two operations
        result.MeanLatency < 0.5ms;
        result.P99 < 2ms;
    }
}
```

**Explanation:**

*   `SimpleService` declares `uses db: Disk`. This means any instance of `SimpleService` needs a `Disk` instance assigned to its internal `db` reference.
*   The `WriteThenRead` method calls `db.Write()` and `db.Read()`. The DSL evaluation model intends for the latency of these sequential operations to be automatically combined.
*   In `ServiceSystem`, `theDisk` is created first. Then, `theService` is instantiated, and `theDisk` is assigned to satisfy the `db` dependency (exact syntax for this wiring needs finalization).
*   The `analyze` block targets the service method, and the `expect` block asserts against the performance of the *entire* `WriteThenRead` operation.

---

## Example 3: Probabilistic Behavior (Cache)

This example models a typical cache pattern: check cache, on miss, fetch from source (disk) and populate cache.

```dsl
component Disk {
    param ProfileName: string = "SSD";
}
component Cache {
    param HitRate: float = 0.8; // 80% hit rate
    // Other params like HitLatency, MissLatency etc. exist in Go component
}

// A service using a cache and a disk
component DataService {
    uses c: Cache;
    uses db: Disk;

    // Reads data, utilizing the cache.
    // Returns true if data was successfully retrieved (hit or fetched).
    method GetData(key: string): bool {
        cacheRead = c.Read(); // Attempt cache read

        // If the cache operation succeeded AND it was a hit...
        // (Simplification: Assume .Success means HIT for this example DSL)
        if cacheRead.Success {
             // Cache hit! Latency is just the cache hit latency.
             log "Cache hit for key:", key;
             return true;
        } else {
             // Cache miss (or potentially cache failure). Fetch from DB.
             log "Cache miss for key:", key, "fetching from DB.";
             dbRead = db.Read();

             // If DB read succeeded, attempt to write back to cache (fire-and-forget perf).
             if dbRead.Success {
                 c.Write(); // Latency of Write() is implicitly tracked but doesn't block return
             }
             return dbRead.Success; // Return success status of the DB read
        }
    }
}

system CacheSystem {
    instance mainDB: Disk = { ProfileName = "SSD"; };
    instance mainCache: Cache = { HitRate = 0.90; }; // High hit rate
    instance dataSvc: DataService = { c = mainCache; db = mainDB; }; // Inject dependencies

    analyze Cache_Hit_Scenario = dataSvc.GetData("some_key") expect {
        // Because HitRate is high (90%), average latency should be closer
        // to cache latency than disk latency.
        result.Availability > 0.98;
        result.MeanLatency < 0.1ms; // Much lower than disk read mean
        result.P99 < 1ms;          // Tail latency dominated by cache misses + disk read
    }
}
```

**Explanation:**

*   `DataService` uses both a `Cache` and a `Disk`.
*   `GetData` checks the cache (`c.Read()`).
*   If it's a hit (`cacheRead.Success` in this simplified view), it returns quickly.
*   If it's a miss, it reads from the database (`db.Read()`) and attempts a cache write (`c.Write()`). The overall latency in the miss path includes the cache miss latency + DB read latency + cache write latency (all combined implicitly by the evaluator).
*   The `analyze` block evaluates the overall `GetData` performance. With a high `HitRate`, the average latency should be significantly lower than a direct database read.

---

## Example 4: Control Flow and Explicit Delay

This example demonstrates using `if` and the explicit `delay` statement.

```dsl
component Disk { /* ... */ }
component ProcessingService {
    uses storage: Disk;
    param MaxRetries: int = 3;
    param RetryDelay: duration = 10ms;

    method ProcessItem(item: string): bool {
        let attempts = 0;
        let success = false;

        // Loop equivalent (requires implementing loop construct later)
        // For now, simulate a single retry attempt logic
        opOutcome = storage.Write();
        attempts = attempts + 1; // Assume numeric ops work

        if !opOutcome.Success {
            log "Write failed on attempt", attempts, "delaying...";
            delay self.RetryDelay; // Explicitly add latency

            // Simulate one retry
            opOutcome = storage.Write();
            attempts = attempts + 1;

            if !opOutcome.Success {
                 log "Write failed after retry";
                 return false;
            } else {
                 log "Write succeeded on retry";
                 success = true;
            }
        } else {
            success = true;
        }
        return success;
    }
}

system RetrySystem {
    instance slowDisk: Disk = { ProfileName = "HDD"; }; // Use slower disk
    instance procSvc: ProcessingService = {
         storage = slowDisk;
         RetryDelay = 20ms; // Override default delay
    };

    analyze Process_With_Retry = procSvc.ProcessItem("job123") expect {
        // Availability might be higher due to retry, but latency increases on failure.
        result.Availability > 0.99; // Higher than single HDD write avail
        result.P99 > 20ms;         // P99 likely includes the retry delay + 2nd write
        result.MeanLatency > 10ms; // Mean latency higher than single HDD write
    }
}
```

**Explanation:**

*   `ProcessingService.ProcessItem` attempts a write.
*   If the first write fails (`!opOutcome.Success`), it explicitly introduces a `delay self.RetryDelay`. This adds to the latency *only* on the failure path.
*   It then retries the write. The total latency for the retry path includes the first failed write attempt + the delay + the second write attempt.
*   The `analyze` block tests this logic, expecting higher availability but also higher tail latency due to the potential delays.

---

## Example 5: Concurrency (`go` / `wait`)

This example shows fanning out writes to two disks in parallel.

```dsl
component Disk { /* ... */ }
component FanoutService {
    uses d1: Disk;
    uses d2: Disk;

    method WriteBoth(data: string): bool {
        // Start write operations concurrently using 'go'.
        // 'fut1' and 'fut2' represent the handles to these async operations.
        go fut1 = d1.Write();
        go fut2 = d2.Write();

        // Wait for both concurrent operations to complete.
        // The V4 model should calculate the combined latency based on the
        // *longer* of the two parallel operations, plus any sync overhead.
        wait fut1, fut2;

        // How to check success? Need to access future results.
        // Placeholder: Assume success if wait finishes without error (needs refinement).
        // Or maybe wait returns a combined outcome? TBD.
        log "Both writes completed";
        return true; // Simplified return
    }
}

system ConcurrentSystem {
    instance diskA: Disk = { ProfileName = "SSD"; };
    instance diskB: Disk = { ProfileName = "SSD"; };
    instance fanoutSvc: FanoutService = { d1 = diskA; d2 = diskB; };

    analyze Parallel_Writes = fanoutSvc.WriteBoth("payload") expect {
        // Latency should be driven by the max of the two parallel writes,
        // not the sum. Should be close to a single SSD write.
        result.Availability > 0.99; // Slightly lower than single if either fails
        result.MeanLatency < 0.3ms;  // Should be close to single write latency
        result.P99 < 1.5ms;         // Similar to single write P99
    }
}

```

**Explanation:**

*   `FanoutService.WriteBoth` uses `go fut1 = d1.Write()` and `go fut2 = d2.Write()` to initiate writes concurrently.
*   `wait fut1, fut2` blocks until *both* operations finish.
*   The core idea is that the total latency measured by the `analyze` block should reflect the time taken until the *slower* of the two parallel writes completes, not their sum.
*   The exact mechanism for retrieving success/failure status from waited-on futures needs definition in the evaluator.
