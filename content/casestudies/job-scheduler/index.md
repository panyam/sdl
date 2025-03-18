---
title: 'Design a Job Scheduler'
productName: 'Job Schecduler'
date: 2025-02-03T11:29:10AM
tags: ['job scheduler', 'medium' ]
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: A job scheduler automatically schedules and executes jobs at times (or intervals) chosen by the caller.   They can be used to automate repetitive tasks (eg cron jobs, batch jobs, clean-up/maintainence jobs etc).
scrollToBottom: true
---

{{# include "DrawingView.html" #}}

{{ .FrontMatter.summary }}

## Requirements

### Functional Requirements

### Extensions (Out of scope)

### Non functional requirements

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

{{ template "DrawingView" ( dict "caseStudyId" "cleantemplate" "id" "casestudies/cleantemplate/hld" ) }}

---

## Deep Dives

