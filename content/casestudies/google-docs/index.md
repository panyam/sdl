---
title: 'Design Google Docs'
date: 2025-02-06T11:29:10AM
tags: ['google docs', 'hard', 'websockets', 'concurrency']
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: Design Google Docs
scrollToBottom: true
---

{{# include "DrawingView.html" #}}

## Requirements

### Functional Requirements

* CRUD on Documenst
* Live Editing of documents - Get live updates on docs and see each other's cursors

### Extensions (Out of scope)

Functional:
1) Auditability/Compliance - especially for user generated entities (trades - are they legal, posts - CSAM?, blogs - copyright etc)
2) Analytics - For the entities a user creates what kind of analytics can be done.  Typical are - top N kind of problems.
3) Authentication/Access Control - user management
4) History and versioning

Non Functional:
1) Observability - So we can ensure reliability, uptime, attacks etc
2) Disaster recovery - comes under reliabiliability and is a great way to go into which parts are "effectful" and which arent.
3) CI/CD, Deployability, Experimentation

### Non functional requirements

* Eventual consistency OK
* Low (ish) # concurrent writes ~ 100
* Low update latency - say 100ms
* Users can be distributed globally editing a single doc!
* Durability of documents
* Billions of documents


## API/Interface/Entities

### Entities

```

record Message {
  union {
    Inserted,
    Deleted,
    Updated,
    CursorUpdated,
  }
}

record Cursor {
  docId string
  userId string
  position int
}

record Editor {
  cursor Cursor
}

record Document {
  id string
  creatorId string
  createdAt Timestamp
  updatedAt Timestamp
  
  // history, permissions, other info
  Metadata any
  
  // Where the actual document content is stored
  blobId string
}

record DocumentBlob {
  id string
  data []byte
}
```

### Services/APIs

```
service DocumentService {
  CRUD on Document
  
  Connect(docId string) Stream<Message> {
    WS /docs/{docId}
  }
}
```

## High Level Design

{{ template "DrawingView" ( dict "caseStudyId" "google-docs" "id" "casestudies/google-docs/hld" ) }}

1. User can CRUD document metadata (not including editing) which writes our our MD store (for now KV Store like dynamo
   or postgres is fine)


#### Option 1: Send entire document

Users live edit by sending updates to the document - V1 can just be each user uploads entire document, EditService
transactionally updates (with OCC).

Issues:

1. Inefficient - each user key stroke could be an entire upload
2. Locking for each key stroke will grind it to a halt.
3. Could batch this but clunky.

#### Option 2: Send diffs - eg "inserted", "updated" etc.

Pros: lower volume of data
Cons: concurrency issues

Cannot just handle an "insert/edit" etc.  It is based on the version of the document on which the edit was performed.

1. Use operational transformations (OT) or CRDT.
2. Can also use WebRTC (for P2P comms).  Active research in both.

For now we use OT.

{{ template "DrawingView" ( dict "caseStudyId" "google-docs" "id" "casestudies/google-docs/editstore" ) }}

---

## Deep Dives

