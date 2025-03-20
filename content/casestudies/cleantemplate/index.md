---
title: 'Design ProductName'
date: 2024-05-28T11:29:10AM
tags: ['template', 'easy']
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: Description of the product being designed
scrollToBottom: true
---

{{# include "DrawingView.html" #}}

# Design ProductName

## Requirements

### Functional Requirements

### Extensions (Out of scope)

### Non functional requirements

* Availability - Describe why availability is important - eg if user cannot access site they cannot warn dangers
  * Be ready to calculate error rates
* Consistency - eg if user bought something they want to see it in their orders otherwise it is problematic
* Freshness - how fresh is the data user is seeing (ie if user is seeing prices from a 3P system, what SLO do we want on the freshness - would affect their purchase decisions)
  * For both above freshness SLOs would be key.
  * Also call out RYW guarantees for user created entities (tweets, posts, trades etc)
* Scalable - Address scalability for the scale requirement numbers and SLOs on CRUD operations (again why it affects product/platform/user experience goodness)
* Geo capabilities - Part of scalability is address how users in various regions are affected etc.
  * Also has impact on freshness/consistency

**Latency Targets:** What are your expectations for the response times for price updates and trade creation? How critical is near-real-time performance here?

**Throughput and Scalability:** With the estimated load, what kind of throughput numbers are you targeting, and how would you design the system to scale effectively?

**Availability and Fault Tolerance:** What measures would you implement to ensure the system remains highly available and can gracefully handle failures?


## API/Interface/Entities

### Entities

```
record Entity1 {
}

record Entity2 {
}

record Entity3 {
}
```

### Services/APIs

```
service ThingService {
  CreateThing(x Thing) Thing {
    method: POST
    url: "/things"
  }
}
```

## High Level Design

{{ template "DrawingView" ( dict "caseStudyId" "cleantemplate" "id" "hld" ) }}

---

## Deep Dives

