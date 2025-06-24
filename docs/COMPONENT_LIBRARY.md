# SDL Component Library Reference

This document provides comprehensive documentation for all built-in SDL components. These components model common distributed system patterns and can be composed to create complex system architectures.

## Table of Contents

1. [Storage Components](#storage-components)
2. [Caching Components](#caching-components)
3. [Index Components](#index-components)
4. [Queue Components](#queue-components)
5. [Resource Management](#resource-management)
6. [Network Components](#network-components)
7. [Batch Processing](#batch-processing)
8. [Usage Patterns](#usage-patterns)

## Storage Components

### Disk

**Purpose**: Models storage device performance with realistic latency distributions for SSD and HDD profiles.

**Parameters**: None (uses predefined profiles)

**Methods**:
- `Read() Outcomes[Bool]` - Performs a disk read operation
- `Write() Outcomes[Bool]` - Performs a disk write operation

**Behavior**:
- SSD Profile (default): ~100μs latency with low variance
- HDD Profile: ~5ms latency with higher variance
- Models occasional slow operations (tail latency)

**Example**:
```sdl
component StorageService {
    uses disk Disk
    
    method SaveData() Bool {
        return self.disk.Write()
    }
}
```

### DiskWithContention

**Purpose**: Models disk I/O with queueing and contention under load. Essential for understanding storage bottlenecks.

**Parameters**:
- `ArrivalRate Float` - Request arrival rate (requests/second)
- `IOPS Float` - Maximum I/O operations per second
- `AvgServiceTime Duration` - Average time per I/O operation

**Methods**:
- `Read() Outcomes[Bool]` - Read with contention modeling
- `Write() Outcomes[Bool]` - Write with contention modeling

**Behavior**:
- Uses M/M/1 queueing model
- Latency increases as utilization approaches capacity
- Models both service time and queue wait time

**Example**:
```sdl
component DatabaseStorage {
    uses disk DiskWithContention(
        ArrivalRate = 1000.0,    // 1000 requests/sec
        IOPS = 5000.0,           // 5000 IOPS capacity
        AvgServiceTime = 0.2ms   // 200μs per operation
    )
    
    method Query() Bool {
        return self.disk.Read()
    }
}
```

### SortedFile

**Purpose**: Models sorted file storage with binary search capabilities. Ideal for modeling indexes and sorted data structures.

**Parameters**:
- `ArrivalRate Float` - Request arrival rate
- `NumRecords Int` - Number of records in the file
- `RecordSize Int` - Size of each record in bytes
- `PageSize Int` - I/O page size in bytes

**Methods**:
- `Find() Bool` - Binary search for a record
- `Scan() Bool` - Sequential scan through records
- `Delete() Bool` - Delete a record

**Performance Characteristics**:
- Find: O(log N) I/O operations
- Scan: O(N) I/O operations
- Delete: O(log N) for finding + rewrite cost

**Example**:
```sdl
component IndexedTable {
    uses index SortedFile(
        NumRecords = 1000000,    // 1M records
        RecordSize = 100,        // 100 bytes per record
        PageSize = 4096          // 4KB pages
    )
    
    method LookupById() Bool {
        return self.index.Find()
    }
}
```

### HeapFile

**Purpose**: Models unordered heap file storage. Suitable for append-only workloads and full table scans.

**Parameters**:
- `NumRecords Int` - Number of records
- `RecordSize Int` - Size of each record
- `PageSize Int` - I/O page size

**Methods**:
- `Scan() Bool` - Full table scan
- `Insert() Bool` - Append new record
- `Delete() Bool` - Delete record (marks as deleted)

**Performance Characteristics**:
- Scan: O(N) - must read all pages
- Insert: O(1) - append to end
- Delete: O(N) - must find record first

**Example**:
```sdl
component LogStorage {
    uses logs HeapFile(
        NumRecords = 10000000,   // 10M log entries
        RecordSize = 200,        // 200 bytes per entry
        PageSize = 8192          // 8KB pages
    )
    
    method AppendLog() Bool {
        return self.logs.Insert()
    }
}
```

### LSMTree

**Purpose**: Models Log-Structured Merge tree storage. Optimized for write-heavy workloads.

**Methods**:
- `Read() Bool` - Read operation
- `Write() Bool` - Write operation

**Behavior**:
- Writes are always fast (append to memory)
- Reads may require checking multiple levels
- Models compaction overhead implicitly

**Example**:
```sdl
component TimeSeriesDB {
    uses storage LSMTree
    
    method IngestMetric() Bool {
        return self.storage.Write()  // Fast append
    }
}
```

## Caching Components

### Cache

**Purpose**: Generic caching layer with configurable hit rate. Models cache effectiveness without implementation details.

**Parameters**:
- `HitRate Float` - Cache hit probability (0.0 to 1.0)
- `MaxThroughput Float` - Maximum operations per second

**Methods**:
- `Read() Bool` - Cache read attempt
- `Write() Bool` - Cache write operation

**Behavior**:
- Returns success based on HitRate probability
- Models zero latency for hits (in-memory)
- Caller must handle misses

**Example**:
```sdl
component CachedService {
    uses cache Cache(HitRate = 0.85)  // 85% hit rate
    uses db Database
    
    method GetData() Bool {
        let hit = self.cache.Read()
        if hit {
            return true
        }
        // Cache miss - go to database
        let data = self.db.Query()
        if data {
            self.cache.Write()
        }
        return data
    }
}
```

### CacheWithContention

**Purpose**: Cache model including connection pool contention and service time.

**Parameters**:
- `HitRate Float` - Cache hit probability
- `ArrivalRate Float` - Request arrival rate
- `ServiceTime Duration` - Time to service cache operation
- `MaxConnections Int` - Connection pool size

**Methods**:
- `Read() Bool` - Read with contention
- `Write() Bool` - Write with contention

**Behavior**:
- Models connection pool exhaustion
- Includes network round-trip time
- Realistic for distributed caches (Redis, Memcached)

**Example**:
```sdl
component APIService {
    uses redis CacheWithContention(
        HitRate = 0.9,
        ArrivalRate = 1000.0,
        ServiceTime = 1ms,       // Network + processing
        MaxConnections = 100     // Connection pool
    )
}
```

## Index Components

### HashIndex

**Purpose**: Hash-based index structure with O(1) average case performance.

**Methods**:
- `Find() Bool` - Point lookup
- `Insert() Bool` - Insert new entry
- `Delete() Bool` - Remove entry

**Behavior**:
- Constant time operations (average case)
- No range query support
- Models hash collisions implicitly

**Example**:
```sdl
component UserLookupService {
    uses userIndex HashIndex
    
    method FindUserById() Bool {
        return self.userIndex.Find()
    }
}
```

### BTreeIndex (BTree)

**Purpose**: B-tree index for ordered data with range query support.

**Methods**:
- `Find() Bool` - Point lookup
- `Range() Bool` - Range scan
- `Insert() Bool` - Insert maintaining order
- `Delete() Bool` - Delete maintaining balance

**Behavior**:
- O(log N) operations
- Efficient range queries
- Models page-based I/O

**Example**:
```sdl
component TimeRangeQuery {
    uses timeIndex BTree
    
    method GetEventsInRange() Bool {
        return self.timeIndex.Range()
    }
}
```

### BitmapIndex

**Purpose**: Bitmap index for set operations and analytics queries.

**Methods**:
- `Find() Bool` - Lookup operation
- `Insert() Bool` - Add to bitmap
- `Delete() Bool` - Remove from bitmap
- `Update() Bool` - Update bitmap entry

**Behavior**:
- Efficient for low-cardinality columns
- Fast boolean operations (AND, OR)
- High compression ratio

**Example**:
```sdl
component AnalyticsEngine {
    uses categoryIndex BitmapIndex
    
    method FilterByCategory() Bool {
        return self.categoryIndex.Find()
    }
}
```

## Queue Components

### ResourcePool

**Purpose**: Models a pool of limited resources using M/M/c queueing theory. Essential for connection pools, thread pools, and resource management.

**Parameters**:
- `Size Int` - Number of resources in pool
- `ArrivalRate Float` - Request arrival rate
- `AvgHoldTime Duration` - Average time resource is held

**Methods**:
- `Acquire() Bool` - Try to acquire a resource

**Behavior**:
- Uses Erlang-C formula for wait times
- Models both acquisition success and wait time
- Returns false when pool exhausted

**Utilization Tracking**:
- Provides real-time utilization metrics
- Identifies bottlenecks when utilization > 80%

**Example**:
```sdl
component DatabaseService {
    uses connPool ResourcePool(
        Size = 20,              // 20 connections
        ArrivalRate = 100.0,    // 100 requests/sec
        AvgHoldTime = 50ms      // 50ms per query
    )
    
    method ExecuteQuery() Bool {
        let conn = self.connPool.Acquire()
        if !conn {
            return false  // Pool exhausted
        }
        delay(50ms)  // Simulate query
        return true
    }
}
```

### MM1Queue

**Purpose**: Single-server queue (M/M/1) for modeling sequential processing.

**Methods**:
- `Enqueue() Bool` - Add item to queue
- `Dequeue() Bool` - Process item from queue

**Behavior**:
- Unlimited queue capacity
- FIFO processing
- Exponential service times

**Example**:
```sdl
component MessageProcessor {
    uses queue MM1Queue
    
    method ProcessMessage() Bool {
        return self.queue.Enqueue()
    }
}
```

### MMCKQueue

**Purpose**: Multi-server queue with capacity limit (M/M/c/K).

**Methods**:
- `Enqueue() Bool` - Add to queue (may reject)
- `Dequeue() Bool` - Process from queue

**Behavior**:
- Multiple parallel servers
- Finite queue capacity
- Rejects when full

**Example**:
```sdl
component LoadBalancer {
    uses servers MMCKQueue  // c servers, K capacity
    
    method HandleRequest() Bool {
        return self.servers.Enqueue()
    }
}
```

### Queue

**Purpose**: Generic queue with configurable parameters.

**Parameters**:
- `NumServers Int` - Number of parallel servers (c)
- `Capacity Int` - Maximum queue size (K)
- `ArrivalRate Float` - Request arrival rate
- `ServiceTime Duration` - Average processing time

**Methods**:
- `Enqueue() Bool` - Add to queue
- `Dequeue() Bool` - Remove from queue

**Example**:
```sdl
component TaskQueue {
    uses tasks Queue(
        NumServers = 4,
        Capacity = 100,
        ArrivalRate = 50.0,
        ServiceTime = 100ms
    )
}
```

## Network Components

### Link (NetworkLink)

**Purpose**: Models network connection with latency, jitter, and packet loss.

**Parameters**:
- `BaseLatency Duration` - Baseline network latency
- `MaxJitter Duration` - Maximum latency variation
- `PacketLossProb Float` - Packet loss probability (0.0-1.0)

**Methods**:
- `Transfer() Bool` - Transfer data over network

**Behavior**:
- Latency = BaseLatency + random(0, MaxJitter)
- Models packet loss probabilistically
- Suitable for WAN/Internet connections

**Example**:
```sdl
component CrossRegionService {
    uses wan Link(
        BaseLatency = 50ms,      // Cross-region baseline
        MaxJitter = 10ms,        // ±10ms variation
        PacketLossProb = 0.001   // 0.1% packet loss
    )
    
    method RemoteCall() Bool {
        return self.wan.Transfer()
    }
}
```

## Batch Processing

### Batcher

**Purpose**: Collects items and processes them in batches for efficiency.

**Parameters**:
- `Policy String` - "SizeBased" or "TimeBased"
- `BatchSize Int` - Target batch size
- `Timeout Duration` - Time window for collection
- `ArrivalRate Float` - Item arrival rate

**Methods**:
- `Submit() Bool` - Submit item for batching
- `ProcessBatch() Bool` - Process accumulated batch

**Behavior**:
- SizeBased: Processes when BatchSize reached
- TimeBased: Processes every Timeout interval
- Models batching efficiency gains

**Example**:
```sdl
component BulkProcessor {
    uses batcher Batcher(
        Policy = "SizeBased",
        BatchSize = 100,
        Timeout = 1s,
        ArrivalRate = 200.0
    )
    
    method SubmitItem() Bool {
        return self.batcher.Submit()
    }
}
```

## Usage Patterns

### Pattern 1: Service with Cache and Database

```sdl
component UserService {
    uses cache Cache(HitRate = 0.8)
    uses db Database
    uses pool ResourcePool(Size = 10, AvgHoldTime = 20ms)
    
    method GetUser() Bool {
        // Try cache first
        let cached = self.cache.Read()
        if cached {
            return true
        }
        
        // Get database connection
        let conn = self.pool.Acquire()
        if !conn {
            return false
        }
        
        // Query database
        let result = self.db.Query()
        if result {
            self.cache.Write()  // Update cache
        }
        
        return result
    }
}
```

### Pattern 2: Multi-Tier Storage

```sdl
component StorageHierarchy {
    uses memCache Cache(HitRate = 0.7)
    uses diskCache Cache(HitRate = 0.9)
    uses disk DiskWithContention(IOPS = 1000)
    
    method Read() Bool {
        // L1: Memory cache
        if self.memCache.Read() {
            return true
        }
        
        // L2: Disk cache
        if self.diskCache.Read() {
            self.memCache.Write()
            return true
        }
        
        // L3: Persistent storage
        let data = self.disk.Read()
        if data {
            self.diskCache.Write()
            self.memCache.Write()
        }
        return data
    }
}
```

### Pattern 3: Load Balanced Service

```sdl
component LoadBalancedAPI {
    uses loadBalancer Queue(NumServers = 5, ServiceTime = 10ms)
    uses connPool ResourcePool(Size = 50, AvgHoldTime = 10ms)
    
    method HandleRequest() Bool {
        // Load balancer queues request
        let queued = self.loadBalancer.Enqueue()
        if !queued {
            return false  // Rejected due to overload
        }
        
        // Get connection from pool
        let conn = self.connPool.Acquire()
        return conn
    }
}
```

## Best Practices

1. **Choose the Right Component**: 
   - Use Cache for simple hit/miss modeling
   - Use CacheWithContention for distributed caches
   - Use ResourcePool for connection pools

2. **Set Realistic Parameters**:
   - Base parameters on actual system measurements
   - Consider peak load scenarios
   - Account for network latencies

3. **Model Failure Modes**:
   - Set appropriate queue capacities
   - Model packet loss for network calls
   - Include timeout scenarios

4. **Compose Components**:
   - Build complex systems from simple components
   - Layer caches for realistic hierarchies
   - Chain queues for pipeline processing

5. **Monitor Utilization**:
   - ResourcePool and Queue components track utilization
   - High utilization (>80%) indicates bottlenecks
   - Use this data to identify scaling needs