---
title: 'Design Live Commenting System'
productName: 'LiveComments'
date: 2025-02-16T11:29:10AM
tags: ['comments', 'facebook', 'meta', 'live', 'fan-out', 'medium' ]
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: LiveComments
scrollToBottom: true
---

{{# include "DrawingView.html" #}}

{{ .FrontMatter.summary }}

TBR - To be rewritten

## Requirements

### Functional Requirements

* Viewers can post comments on a Live video feed.
* Viewers can see new comments being posted while they are watching the live video.
* Viewers can see comments made before they joined the live feed.

### Scale Requirements:

* 1B DAU
* User follows about 200 people
* Has about 1k followers p50, Million in P99
* User Posts about 10 posts a day - 10B posts per day - 4T posts per year
* See feed about 10 times a day - Care about say top 100 posts on feed (even with scrolling this is no more than a day
  or two of your followees)
  - Assume your feed can go back to a month
  - About 30 days * 200 followees * 10 posts per day = 60k posts

### Non functional requirements

* The system should scale to support millions of concurrent videos and thousands of comments per second per live video.
* The system should prioritize availability over consistency, eventual consistency is fine.
* The system should have low latency, broadcasting comments to viewers in near-real time (< 200ms end-to-end latency under typical network conditions)
* Pluggable - Commenting system and Video team can be different teams

### Extensions (Out of scope)

* Viewers can reply to comments
* Viewers can react to comments

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

record LiveVideo {
  Id string
  CreatorId string
  CreatedAt Time
  
  // Other metadata
}

record Comment {
  Id string         // PKEY  - can even have entityID as part of the ID for easier classification
  EntityId string   // shard key - EntityId
  CreatedAt Time    // secondary index - CreatedAt
  CreatorId string
  Text string
}
```

```
// Assume it exists
service VideoService {
    // ... CRUD
}

// Could be diff team and assumes authenticated
service CommentService {
  CreateComment(comment Comment) Comment {    // validates and sets comment.Id, CreatedAt and CreatorId 
    POST /comments
    BODY: {comment}
  }
  
  ListComments(entityId string) Paginated<Comment> {
    GET /comments/?entityId={entityId}
    
    // Alternatively - if enitty service wants to expose an endpoint (but needs teach specific change)
    // GET /entitiType/{entityId}/comments
  }
}

```

### High Level Design


{{ template "DrawingView" ( dict "caseStudyId" "LiveComments" "id" "hld" ) }}

### Deep Dive - New comments

* Can Poll - All of its problems
* Can be notified - SSE or WS
* SSE can have limits (if commenting on multiple videos and listening) but is simpler
* WS no limits - but increased complexity due to bidrectionality

{{ template "DrawingView" ( dict "caseStudyId" "LiveComments" "id" "final" ) }}

### Deep Dive - Scaling

* Shard by EntityID so comments for a particular entity can be fetched from a single shard
* Cache comments by CommentID

or CommentID -> Secondary Index -> Comments

{{ template "DrawingView" ( dict "caseStudyId" "LiveComments" "id" "channels" ) }}

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

{{ template "DrawingView" ( dict "caseStudyId" "LiveComments" "id" "coordinator" ) }}
