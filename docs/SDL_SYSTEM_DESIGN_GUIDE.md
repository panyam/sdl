# Using SDL for System Design Interviews

## Introduction

System design interviews test your ability to architect large-scale distributed systems. SDL helps you validate your designs with actual performance models, moving beyond hand-wavy explanations to concrete analysis.

## Why SDL for System Design?

Traditional system design interviews involve:
- Drawing boxes and arrows
- Making assumptions about scale
- Guessing at performance characteristics
- Hand-waving about bottlenecks

SDL enables you to:
- Model your design with real performance characteristics
- Validate assumptions with simulations
- Identify bottlenecks before building
- Demonstrate deep understanding of system behavior

## Common System Design Patterns in SDL

### 1. URL Shortener (e.g., bit.ly)

**Key Requirements:**
- Shorten long URLs to short codes
- Redirect short URLs to long URLs
- Handle 100M URLs, 10:1 read/write ratio

**SDL Model:**
```sdl
// Data storage component
component URLDatabase {
    uses pool ResourcePool(
        Size = 100,           // 100 connections
        AvgHoldTime = 5ms     // Fast key-value lookups
    )
    
    param NumShards Int = 10
    param CacheHitRate Float = 0.9  // 90% of reads hit cache
    
    method WriteURL() Bool {
        let shard = sample dist {
            10 => 0, 10 => 1, 10 => 2, 10 => 3, 10 => 4,
            10 => 5, 10 => 6, 10 => 7, 10 => 8, 10 => 9
        }
        
        return self.pool.Acquire()
    }
    
    method ReadURL() Bool {
        // Check cache first
        let cacheHit = sample dist {
            90 => true,
            10 => false
        }
        
        if cacheHit {
            delay(0.1ms)  // Memory access
            return true
        }
        
        // Cache miss - go to database
        return self.pool.Acquire()
    }
}

// API service
component URLShortenerAPI {
    uses db URLDatabase
    uses rateLimiter ResourcePool(Size = 1000)  // 1000 concurrent requests
    
    method ShortenURL() Bool {
        // Rate limiting
        if !self.rateLimiter.Acquire() {
            return false
        }
        
        // Generate short code (CPU bound)
        delay(1ms)
        
        // Store in database
        return self.db.WriteURL()
    }
    
    method RedirectURL() Bool {
        // Rate limiting
        if !self.rateLimiter.Acquire() {
            return false
        }
        
        // Lookup URL
        return self.db.ReadURL()
    }
}

// Complete system
system URLShortener {
    use api URLShortenerAPI
    use cache Cache(HitRate = 0.9)
    use db URLDatabase
}
```

**Analysis Questions:**
1. What's the maximum throughput?
2. Where are the bottlenecks?
3. How does cache hit rate affect performance?

### 2. Social Media Feed (e.g., Twitter Timeline)

**Key Requirements:**
- Users follow other users
- Generate home timeline from followed users
- Handle celebrity users with millions of followers

**SDL Model:**
```sdl
// User types with different characteristics
enum UserType {
    Regular,    // 100 followers average
    Popular,    // 10K followers average  
    Celebrity   // 1M+ followers
}

component TweetStorage {
    uses pool ResourcePool(Size = 200, AvgHoldTime = 10ms)
    
    method WriteTweet() Bool {
        return self.pool.Acquire()
    }
    
    method ReadTweets() Bool {
        // Batch read of recent tweets
        delay(20ms)
        return self.pool.Acquire()
    }
}

component TimelineService {
    uses tweetStore TweetStorage
    uses cache Cache(HitRate = 0.7)  // Timeline cache
    uses fanoutQueue Queue(NumServers = 50)
    
    method PostTweet(userType UserType) Bool {
        // Write tweet
        if !self.tweetStore.WriteTweet() {
            return false
        }
        
        // Fan-out based on user type
        switch userType {
            Regular => {
                // Push to followers' timelines
                gobatch 100 {
                    self.fanoutQueue.Enqueue()
                }
            }
            Popular => {
                // Hybrid approach - push to active users only
                gobatch 1000 {
                    self.fanoutQueue.Enqueue()
                }
            }
            Celebrity => {
                // Pull model - don't push
                return true
            }
        }
        
        return true
    }
    
    method GetTimeline(userType UserType) Bool {
        // Check cache first
        if self.cache.Read() {
            return true
        }
        
        switch userType {
            Regular => {
                // Read pre-computed timeline
                delay(5ms)
                return true
            }
            Popular => {
                // Merge pre-computed + recent
                delay(20ms)
                return true
            }
            Celebrity => {
                // Pull from followed users
                gobatch 100 {
                    self.tweetStore.ReadTweets()
                }
                delay(50ms)  // Merge time
                return true
            }
        }
    }
}
```

### 3. Ride Sharing Service (e.g., Uber)

**Key Requirements:**
- Match riders with drivers
- Real-time location updates
- Surge pricing under high demand

**SDL Model:**
```sdl
component LocationService {
    uses geoIndex BTreeIndex  // Spatial index
    uses pool ResourcePool(Size = 50, AvgHoldTime = 5ms)
    
    method UpdateLocation() Bool {
        // Update driver location
        return self.geoIndex.Insert()
    }
    
    method FindNearbyDrivers() Bool {
        // Range query on geo index
        let result = self.geoIndex.Range()
        delay(10ms)  // Process results
        return result
    }
}

component MatchingService {
    uses location LocationService
    uses pricing PricingService
    
    param SurgeThreshold Float = 0.8  // 80% utilization triggers surge
    
    method RequestRide() Bool {
        // Find nearby drivers
        let drivers = self.location.FindNearbyDrivers()
        if !drivers {
            return false  // No drivers available
        }
        
        // Calculate price
        let utilization = self.GetUtilization()
        if utilization > self.SurgeThreshold {
            self.pricing.CalculateSurgePrice()
        } else {
            self.pricing.CalculateNormalPrice()
        }
        
        // Match with driver
        delay(sample dist {
            70 => 100ms,  // Quick match
            25 => 500ms,  // Multiple attempts
            5 => 2s       // Hard to match
        })
        
        return true
    }
}

component RideTrackingService {
    uses pubsub EventBus
    
    method TrackRide() Bool {
        // Publish location updates every 5 seconds
        for true {
            self.pubsub.Publish("location_update")
            delay(5s)
        }
    }
}

system RideSharingPlatform {
    use matching MatchingService
    use location LocationService  
    use tracking RideTrackingService
    use notifications NotificationService
}
```

### 4. Video Streaming (e.g., Netflix)

**Key Requirements:**
- Stream video content globally
- Adaptive bitrate based on bandwidth
- CDN for content delivery

**SDL Model:**
```sdl
component CDNNode {
    param CacheHitRate Float = 0.85
    param Capacity Int = 1000  // Concurrent streams
    
    uses pool ResourcePool(Size = self.Capacity)
    
    method StreamVideo() Bool {
        let cacheHit = sample dist {
            85 => true,
            15 => false
        }
        
        if !cacheHit {
            // Fetch from origin
            delay(100ms)
        }
        
        // Stream to user
        return self.pool.Acquire()
    }
}

component VideoStreamingService {
    uses cdn List[CDNNode]
    uses origin OriginServer
    uses analytics AnalyticsService
    
    method StartStream() Outcomes[String] {
        // Select CDN node based on geography
        let nodeIndex = sample dist {
            40 => 0,  // US East
            30 => 1,  // US West
            20 => 2,  // Europe
            10 => 3   // Asia
        }
        
        let node = self.cdn[nodeIndex]
        
        // Try to start stream
        if !node.StreamVideo() {
            // CDN node full, try origin
            if !self.origin.StreamVideo() {
                return "error"
            }
        }
        
        // Track analytics asynchronously
        go self.analytics.TrackView()
        
        return sample dist {
            70 => "hd",      // HD quality
            25 => "sd",      // SD quality  
            5 => "low"       // Low quality
        }
    }
}
```

### 5. Distributed Cache (e.g., Redis)

**Key Requirements:**
- High-performance key-value store
- Replication for reliability
- Consistent hashing for distribution

**SDL Model:**
```sdl
component CacheShard {
    param Capacity Int = 10000
    param EvictionRate Float = 0.1
    
    uses memory ResourcePool(
        Size = 1000,  // Concurrent operations
        AvgHoldTime = 0.1ms
    )
    
    method Get() Outcomes[Bool] {
        let found = sample dist {
            80 => true,   // Key exists
            20 => false   // Key missing
        }
        
        if self.memory.Acquire() {
            delay(0.1ms)  // Memory access
            return found
        }
        
        return false  // Overloaded
    }
    
    method Set() Bool {
        // Check if eviction needed
        let evict = sample dist {
            10 => true,
            90 => false
        }
        
        if evict {
            delay(1ms)  // Eviction overhead
        }
        
        return self.memory.Acquire()
    }
}

component DistributedCache {
    uses shards List[CacheShard]
    uses hashRing ConsistentHash
    
    param NumShards Int = 10
    param ReplicationFactor Int = 3
    
    method Get(key String) Bool {
        // Determine shard using consistent hashing
        let shardId = self.hashRing.GetShard(key)
        
        // Try primary
        let primary = self.shards[shardId]
        let result = primary.Get()
        
        if !result {
            // Try replicas
            for i < self.ReplicationFactor - 1 {
                let replicaId = (shardId + i + 1) % self.NumShards
                let replica = self.shards[replicaId]
                if replica.Get() {
                    return true
                }
            }
        }
        
        return result
    }
}
```

## Performance Analysis Techniques

### 1. Load Testing

```sdl
// Test increasing load
sdl gen add load_10 api.endpoint 10
sdl gen add load_100 api.endpoint 100  
sdl gen add load_1000 api.endpoint 1000
sdl gen add load_10000 api.endpoint 10000

// Measure response times at each level
sdl measure add latency_10 api.endpoint latency
```

### 2. Bottleneck Identification

```sdl
// Measure component utilization
sdl utilization

// Output shows:
// database.pool: 95% utilized (CRITICAL)
// cache.connections: 45% utilized (OK)
// api.rateLimiter: 60% utilized (OK)
```

### 3. Failure Scenario Testing

```sdl
// Simulate cache failure
sdl set cache.HitRate 0.0  // 0% hit rate

// Simulate database slowdown
sdl set db.pool.AvgHoldTime 100ms  // 10x slower

// Simulate network issues
sdl set network.PacketLossRate 0.05  // 5% packet loss
```

## Interview Strategy

### 1. Start with Requirements

```sdl
// Document scale requirements as parameters
param DailyActiveUsers = 100_000_000
param PeakQPS = 10_000
param AvgResponseTime = 100ms
param P99ResponseTime = 500ms
```

### 2. Model Core Components

```sdl
// Start simple
component BasicAPI {
    uses db Database
    
    method Handle() Bool {
        return self.db.Query()
    }
}

// Add complexity incrementally
component ImprovedAPI {
    uses cache Cache
    uses db Database
    
    method Handle() Bool {
        if self.cache.Get() {
            return true
        }
        return self.db.Query()
    }
}
```

### 3. Validate Design Decisions

```sdl
// Test: Does caching help?
sdl set cache.HitRate 0.0   // No cache
sdl measure add no_cache api.Handle latency

sdl set cache.HitRate 0.8   // 80% cache hit
sdl measure add with_cache api.Handle latency

// Compare results to validate caching benefit
```

### 4. Explore Trade-offs

```sdl
// Option 1: More database connections
sdl set db.pool.Size 100

// Option 2: Better caching
sdl set cache.HitRate 0.95

// Option 3: Add read replicas
component DatabaseWithReplicas {
    uses primary Database
    uses replicas List[Database]
    
    method Read() Bool {
        // Route reads to replicas
        let replica = sample dist {
            25 => 0, 25 => 1, 25 => 2, 25 => 3
        }
        return self.replicas[replica].Query()
    }
}
```

## Common Interview Topics

### 1. Scalability

Show how system scales:
```sdl
// Horizontal scaling
param NumAPIServers = 10
param NumDatabaseShards = 5

// Vertical scaling  
param ServerCPUs = 16
param ServerMemoryGB = 64
```

### 2. Reliability

Model failure scenarios:
```sdl
component ReliableService {
    uses primary Service
    uses backup Service
    
    method Handle() Bool {
        let result = self.primary.Process()
        if !result {
            // Failover to backup
            return self.backup.Process()
        }
        return result
    }
}
```

### 3. Consistency

Model consistency trade-offs:
```sdl
component EventuallyConsistentStore {
    param ReplicationLag = dist {
        70 => 10ms,
        25 => 100ms,
        5 => 1s
    }
    
    method Read() Bool {
        // May read stale data
        delay(sample self.ReplicationLag)
        return true
    }
}
```

## Tips for Success

1. **Start Simple**: Begin with basic components and add complexity
2. **Use Real Numbers**: Base parameters on actual system metrics
3. **Model the Critical Path**: Focus on performance-critical operations
4. **Consider Failure Modes**: Show how system handles failures
5. **Validate Assumptions**: Use simulations to verify design decisions
6. **Show Trade-offs**: Demonstrate understanding of design alternatives

## Summary

SDL transforms system design interviews from theoretical discussions to data-driven analysis. By modeling your designs, you can:

- Prove your architecture meets requirements
- Identify bottlenecks before they occur
- Make informed trade-off decisions
- Demonstrate deep system understanding

Remember: The goal isn't to build the system, but to understand how it will behave under various conditions. SDL gives you the tools to do this with confidence.