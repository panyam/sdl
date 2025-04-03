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

# {{.FrontMatter.title}}

{{ .FrontMatter.summary }}

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

### Barriers to Scale

* Billion documents in store
* DAU - 50M documents edited by upto 100 users
* Each user can make say 3 edits per second for an hour on a document

Number of edits = 3 edits per second * 100 users * 100k seconds * 50M active documents
= 1500 Trillion edits per day!

Ok looking at *per document* we have:
* 300 edits per second per document

If a document is say 10MB, it can end up with `300*100000` = `30M edits in a day`

That's a *lot* of storage for a single document.  Snapshotting is key so we dont store edits beyond a certain amount.

Edit Service would need to perform compaction/snapshotting periodically.   Can do this and send live version down to
EditStore along with version, timestamp etc so a reload only starts from here.

How about memory?

## Extended Challenges

#### **1\. Read-Only Mode: Efficiently Scaling to Millions of Readers**

##### **Challenges**

* Ensuring high availability and low latency for millions of concurrent readers.  
* Reducing load on the primary document editing service.  
* Providing real-time updates with minimal performance overhead.

##### **Solution**

* **Content Distribution via Caching/CDN**: Store read-only versions in distributed caches (e.g., Redis, Cloud CDN) to serve read-heavy workloads efficiently.  
* **Eventual Consistency with Replication**: Use a **leader-follower model** where the leader handles writes, and followers serve read requests with slight delays.  
* **Snapshot Mechanism**: Generate **read-optimized snapshots** of documents at specific intervals, allowing users to view historical versions without overwhelming storage.  
* **WebSocket or SSE (Server-Sent Events) for Updates**: Push incremental changes to read-only clients efficiently instead of refreshing the whole document.

##### **Trade-offs**

* **Latency vs. Freshness**: A strong caching mechanism can improve performance but might cause minor delays in content updates.  
* **Storage Cost**: Keeping multiple read-only replicas adds storage overhead but improves scalability.

---

#### **2\. Versioning: Extending Snapshot/Compaction Approach**

##### **Challenges**

* Efficiently storing and retrieving document versions.  
* Balancing performance with storage overhead.  
* Providing users with a way to restore previous versions quickly.

##### **Solution**

* **Delta-Based Versioning**: Store only **differential changes (deltas)** instead of full copies to reduce storage overhead.  
* **Compaction & Checkpointing**:  
  * Every **N** edits, create a full checkpoint.  
  * Store intermediate changes as diffs (delta encoding).  
* **Immutable Version Storage**:  
  * Keep a **log-structured merge tree (LSM-tree) or event sourcing model** for efficient retrieval.  
* **Storage Optimizations**:  
  * Use **S3 versioning, GCS Object Versioning, or a custom storage format** for efficient archival.  
  * Implement **garbage collection policies** to prune unnecessary versions after a retention period.

##### **Trade-offs**

* **Storage Cost vs. Performance**: Keeping too many snapshots increases storage, but aggressive compaction reduces retrieval efficiency.  
* **Granularity of Snapshots**: More frequent snapshots reduce loss risk but increase storage overhead.

---

#### **3\. Memory Optimization: Handling Large Documents Efficiently**

##### **Challenges**

* Large documents consuming excessive memory.  
* Increased pressure on real-time document rendering and collaboration features.

##### **Solution**

* **Lazy Loading and Virtualization**:  
  * Load only the **active portion** of the document in memory.  
  * Use **pagination and chunking** for rendering large documents efficiently.  
* **Efficient Data Structures**:  
  * Store documents as **compressed AST (Abstract Syntax Tree)** to optimize processing.  
  * Use **rope data structures** instead of plain text buffers for efficient text manipulation.  
* **Client-Side Rendering & Offloading**:  
  * Move some processing (e.g., rendering, formatting) to the **client-side**.  
  * Use **WebAssembly (Wasm) or workers** for offloading computations.  
* **Delta-Based Synchronization**:  
  * Instead of sending full documents, **only sync changes (CRDTs, OT algorithms)** to reduce memory footprint.

##### **Trade-offs**

* **Performance vs. Complexity**: Offloading computation to the client reduces server load but increases client-side requirements.  
* **Memory vs. Latency**: Keeping more in-memory improves performance but leads to higher resource consumption.

---

#### **4\. Offline Mode: Expanding Design for Disconnected Clients**

##### **Challenges**

* Users must be able to edit documents without an internet connection.  
* Syncing changes when reconnected should handle conflicts efficiently.  
* Ensuring minimal data loss when switching between offline and online modes.

##### **Solution**

* **Local Storage for Caching & Edits**:  
  * Use **IndexedDB, WebSQL, or file-based storage** to store documents locally.  
  * Store changes as **operation logs** instead of full document copies.  
* **Conflict Resolution via CRDT or OT**:  
  * **Conflict-free Replicated Data Types (CRDTs)**: Ensure edits merge automatically.  
  * **Operational Transformation (OT)**: Reorder and transform changes on sync.  
* **Background Sync & Delta-Based Updates**:  
  * When reconnected, **only send changed operations** rather than full documents.  
  * Use **conflict resolution policies** (e.g., last-writer-wins, manual merge).  
* **Progressive Web App (PWA) or Service Workers**:  
  * Enable seamless offline operation using **PWA APIs**.  
  * Prefetch documents for offline access.

##### **Trade-offs**

* **Complex Conflict Resolution**: More sophisticated algorithms (CRDT/OT) require careful implementation.  
* **Storage Constraints**: Storing large documents locally can be limited by browser storage quotas.

---

### **Final Thoughts**

* **Read-only mode** can be handled using caching, replication, and event-driven updates.  
* **Versioning** can be optimized using delta storage, compaction, and efficient checkpointing.  
* **Memory efficiency** requires lazy loading, efficient data structures, and delta-based synchronization.  
* **Offline mode** benefits from local caching, CRDT/OT-based conflict resolution, and seamless syncing.
