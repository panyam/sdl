# ContactsSystem Cache Refactoring

## Overview
The ContactsSystem has been refactored to use an explicit Cache component with cache-aside pattern instead of simulating cache behavior with probability distributions.

## Previous Architecture
```
AppServer → Database (with internal cache simulation using sample dist)
              └→ HashIndex
```

The Database component previously simulated cache behavior internally:
```sdl
let cached = sample dist {
    40 => true,   // Cache hit  
    60 => false   // Cache miss
}
```

## New Architecture
```
AppServer → Cache (explicit component)
         └→ Database → HashIndex
```

### Cache-Aside Pattern Implementation
The AppServer now coordinates between Cache and Database:

1. **Check cache first**: `self.cache.Read()`
2. **On cache miss**: Query database `self.db.LookupByPhone()`
3. **Write-through**: Update cache with result `self.cache.Write()`

### Key Benefits
1. **Explicit modeling**: Cache is a first-class component, not a simulation
2. **Clear separation**: Database only handles DB operations, not cache logic
3. **Better metrics**: Can track cache-specific metrics separately
4. **More realistic**: Models actual cache behavior with proper hit/miss semantics
5. **Flexible configuration**: Cache parameters (hit rate, latency) are independent

### Component Changes

#### ContactDatabase
- Removed cache simulation logic
- Removed `CacheHitRate` parameter
- Now focuses solely on database operations

#### ContactAppServer  
- Added explicit `cache` dependency
- Implements cache-aside pattern in `HandleLookup()`
- Coordinates between cache and database

#### System Definition
- Added `contactCache` instance of Cache component
- Updated server dependencies to include cache

### FlowEval Integration
The Cache component implements the `FlowAnalyzable` interface:
- `Read()` method success rate equals cache hit rate
- Service time varies based on hit/miss
- No downstream calls (leaf component)

This ensures accurate traffic flow analysis through the cache-aside pattern.