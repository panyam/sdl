---
title: 'Design a Key Value Store'
productName: 'KeyValueStore'
date: 2025-02-20T11:29:10AM
tags: ['facebook', 'meta', 'key value store', 'cache', 'lru', 'lfu' ]
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: KeyValueStore
scrollToBottom: true
---

{{# include "DrawingView.html" #}}

{{ .FrontMatter.summary }}

TBR - To be rewritten

If deep dive needed to scale immediately - jump here

https://www.hellointerview.com/learn/system-design/problem-breakdowns/distributed-cache#potential-deep-dives

## Requirements

### Functional Requirements

* Get, Set and Delete KV pairs
* Cache eviction timeout and policy (LRU - also be open to LFU)

### Scale Requirements:

* 1TB of data
* 1M QPS reads/writes at peak

### Non functional requirements

* SLO - 1MS
* Scalable to handle above scale
* Fault tolerant (part of scale)
* Eventual consistency ok 

### Extensions (Out of scope)

* Durability (no need to worry about restarts)
* Transactions
* Querying or complex data structures

General:
* Authentication/User management
* Versioning, Scanning (malware, virsuses, csam etc)
* Storage limits, billing etc
* Analytics

## API/Interface/Entities

```
// Assume it exists
service KVStore {
    Get(key string) Any {
    }
    
    Set(key string, value Any) {
    }
    
    Delete(key string) {
    }
}

```

### High Level Design


{{ template "DrawingView" ( dict "caseStudyId" "KeyValueStore" "id" "hld" ) }}

### Basic V1

```
class Cache:
    # Store in memory
    items = {}

    Get(self, key):
        return self.items[key]

    Set(self, key, value):
        self.items[key] = value

    Delete(self, key):
        delete self.items[key]
```

### V2 - With Expiration

```
class Cache:
    # Store in memory
    items = {}
    expiration = "1 minutes"

    Set(self, key, value):
        self.items[key] = (value, NOW())

    Get(self, key):
        val, time = self.items[key]
        if time > NOW() - expiration:
            delete self.items[key]
            return None
        self.items[key] = val, NOW()
        return val

    Delete(self, key):
        delete self.items[key]
```

### V3 - Max Capacity - Needs Eviction

* Use a Hashtable + DLL of nodes and values


```
class Node:
  key: string
  value: any
  last_accessed: Time
  next: Node
  prev: Node
```

```
class Cache:
  itemsById = Map<string, Node>
  itemHead, itemTail = (Node, Node)
```

### Deep Dive - Fault Tolerance and Availability - Multi node

* Assuming 10GB per node
* 1 TB = 100 Nodes

https://www.hellointerview.com/learn/system-design/problem-breakdowns/distributed-cache#potential-deep-dives

1. Replication

* Router writes to all replicas synchronosly
* Router writes to all replicas asyncsynchronosly - Good balance between simplicity and availability
* Let peers replicate among themselves (gossip) - but very complex

2. Scaling strategy

* Shard keys - N servers, distributed accordingly
* Also have multiple replicas per shard (make it variable and allow replication as above)
* Have ShardManager that manages shards and kicks off rebalancing as needed
* Replicate the shardmanager - should need fewer hosts.

On Throughput front:
* At 1MQPS - and each node handling say 20K qps - this is 50 nodes needed
* Assume double for extra room - so 100 nodes
* (easily manageble for book keeping)

On storage (mem) front:
* 32 or 64 GB of RAM (say 30% for bookkeeping) - Leaves 48GB of mem.
* At 1TB of data - Comes to 20 Nodes - Again double it for room

At throughput mem utilization will actually be lower - so 1TB / 100 = 10GB per node.

User smaller memory instances and have more!

Consistent hashing or Coordinator

3. Hot keys for Read

https://www.hellointerview.com/learn/system-design/problem-breakdowns/distributed-cache#4-what-happens-if-you-have-a-hot-key-that-is-being-read-from-a-lot

* Replicas of hot keys

4. Hot keys for Write

https://www.hellointerview.com/learn/system-design/problem-breakdowns/distributed-cache#5-what-happens-if-you-have-a-hot-key-that-is-being-written-to-a-lot

* Batching
* sharding with suffixes


5. Other things

* Connection Pooling
