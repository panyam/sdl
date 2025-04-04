---
title: 'Design Permissions Management System'
productName: 'Permissions'
date: 2025-02-20T11:29:10AM
tags: ['comments', 'facebook', 'meta', 'live', 'fan-out', 'medium' ]
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: Permissions
scrollToBottom: true
---

{{# include "DrawingView.html" #}}

{{ .FrontMatter.summary }}

TBR - To be rewritten

## Requirements

### Functional Requirements

* Users can add/remove access to a resource for other users
* System must validate whether a particular access by a user is allowed on a particular resource

### Non functional requirements

* Scalable for reads
* Accurate
* Strong consistency
* Writes must be durable - consistency over availability?
* Integrated by other parties (like live comments)

### Scale Requirements:

* 1B DAU
* User follows about 200 people
* Has about 1k followers p50, Million in P99
* User Posts about 10 posts a day - 10B posts per day - 4T posts per year
* See feed about 10 times a day - Care about say top 100 posts on feed (even with scrolling this is no more than a day
  or two of your followees)
  - Assume your feed can go back to a month
  - About 30 days * 200 followees * 10 posts per day = 60k posts

### Extensions (Out of scope)

* Transactional access

General:
* Authentication/User management
* Versioning, Scanning (malware, virsuses, csam etc)
* Storage limits, billing etc
* Analytics

## API/Interface/Entities


```
record Paginated<X, TokenType=string> {
  Items []X
  Count int
  PageToken TokenType
  NextPageToken TokenType
}

record User {
  Id string
}

record Permission {
  EntityId string             // PKEY  - EntityId + UserId
  UserId string
  Permissions []string
  
  UpdatedAt Time        // secondary index - CreatedAt
  // Other metadata
}
```

```
// Assume it exists
service PermissionService {
    UpdatePermissions(userId string, entityId string, permissionsi []string) Permission {
      POST /permissions
      BODY: {...}
    }
    
    CanSee(entityId string, userId string) bool {
      GET /permissions/?entityId,userId
    }
}

```

### High Level Design


{{ template "DrawingView" ( dict "caseStudyId" "Permissions" "id" "hld" ) }}

### Deep Dive - New comments

* Can Poll - All of its problems
* Can be notified - SSE or WS
* SSE can have limits (if commenting on multiple videos and listening) but is simpler
* WS no limits - but increased complexity due to bidrectionality

{{ template "DrawingView" ( dict "caseStudyId" "Permissions" "id" "final" ) }}

### Deep Dive - Scaling

* Shard by EntityID so comments for a particular entity can be fetched from a single shard
* Cache comments by CommentID

or CommentID -> Secondary Index -> Comments

{{ template "DrawingView" ( dict "caseStudyId" "Permissions" "id" "channels" ) }}

### Deep Dive - Scaling Channel Services

Usual API

```
CONNECT /comments/?entityId = entity
```

* And we have a connection between userId -> entityId
* But we need a quick way of finding *all* connections for a given entityId so we can push a comment back out.

Assume:

* 100K connections per host possible, Say an event can have 1M subscribers (P99 and say 1k at p50) - so need to push out to 1M connections
  quickly
* A "hot" event will need 10 servers
* *But* we dont want to have 10 servers fully saturated - so we want to increase number of servers and distribute.
* We want to keep a list of "servers" per event in some distributed store

1. User a coordinate service to assign servers to entities
2. Also manages entiites and spills over if need be
3. Use L7 LB so connection terminates at LB and LB takes care of routing/rebalancing etc

* CS has a mapping of entity <--> server - Acts as a DNS
* Keep replicas on standby so it can failover
* Scaling can be autoscaling and registration with Router or Router monitors health and coordinates autoscaling.

{{ template "DrawingView" ( dict "caseStudyId" "Permissions" "id" "coordinator" ) }}

### Deep Dive - Regionalization

Advantage of above is a sharded on entity - we can also have duplicates.  Have "local" channel servers but global
routers.  Routers can be in master master (with local replicas).   When a comment in Zone/Region A gets a comment - it
sends it to the Router in Zone/Region A, which forwards to channel services in this region along with all other regions
it is connected to.  Even with say 10 regions worldwide - connections minimised.  Locally have replicas so that it picks
up.
