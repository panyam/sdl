---
author: Sriram Panyam
title: Random ID Generator
date: 2020-01-30
description: A system for generating random IDs of particular lengths.
math: true
---

Random IDs (RID) are very important in a system.  They have several applications:

1. URL Shortening services where long URLs can be identified by short IDs (eg bitly etc).
2. As IDs for objects in a document store or file stores.
3. As IDs for tracking orders and transactions.
4. And more

They also have several advantages:

1. They can be scaled with minimal bottlenecks
2. They prevent "looping" attacks whereby clients can otherwise scrape all entities in a system by incrementally walking through linear IDs.
3. Like vanity license plats, random IDs customized (as long as they are not already allocated).

# Goals

* Design a robust and modular Random ID generator system.
* Nominate appropriate parameters to model different performance, reliability and cost characteristics.
* Enable reuse as a subsystem in other systems (discussed in this blog) with appropriate tradeoffs.

# Typical Scenarios

A RID generator can be used in the following ways:

* Directly via an API by a user (identified by a userID).
* A website (or UI) with a form for an authenticated user to make a post request to a service that returns an ID (or a failure response if any).
* Called by another service (via an API) for allocating IDs for entities it manages (similar to the direct-by-user case).

# Step 1 - Requirements

## Core Requirements (P0)

We will assume that calls to the RID Generator service are already authenticated and the clientID passed to this service contains information about the authenticated caller (user or a service).

Our service's API is modelled with the following abstract interface:

```
// Our ID Generator has the following parameters that can be configured
// for different scenarios
//
//  ID_SIZE    -   Size of the ID in # of characters.
//
// Also assume that IDs can contain upper case (latin) alphabet characters (A-Z) or digits (0-9).
system RIDGenerator<ID_SIZE> {
  service {
    // Create a new RID based on the create request
    Create(CreateRequest) returns RID {} 

    // Delete a given ID
    Delete(DeleteRequest)
  }

  message CreateRequest {
    // ID of the (authenticated) client making the request
    required string           clientID;

    // We can allow an ID to be reserved so that instead of a random one
    optional Char[ID_SIZE]    explicitRID;
  }

  message DeleteRequest {
    // ID of the client making the request (typically used to log/audit requests)
    int64           clientID;

    // Random ID being released/deleted
    Char[ID_SIZE]   rid;
  }

  message RID {
    // The created random ID
    Char[ID_SIZE]   id;

    // User/Client that owns this ID
    int64           clientID;

    // Create only field set by the service
    Timestamp       createdAt;
  }
}
```

Note that the expli

## Extended Requirements (P1/P2s)

We can explore further requirements if needed at a later stage.  Some examples are:

* Prevent user from adding spam/bad IDs into the system (Needs a spam/filtering system)
* Flag existing IDs, after they have been added, if needed. (Needs a system to classify IDs)
* Track callers of the service for monitoring, alerts, quotas etc. (Needs a monitoring system)
* IDs can have a time-out feature after which they are released back into the pool.  (Needs offline/async queues to track expirations and perform garbage collection).
* Variable Quality of Service – Different tiers of callers have different SLAs.
* Bulk creations

## Operational Requirements

Our system must provide some guarantees on operational requirements such as latency and availability:


* 90% of create requests must be succeed (or return a failure on duplicates) within 20ms
* System must be fault-tolerant and scalable (be able to serve traffic across the globe).

### Usage Estimates

```eqnml
* Creation QPS: ${10000}->$IDCreationQPS
* ID Retention Period: ${10}->$RetentionPeriod years
```

# Step 2 - High Level Design

An important step in all designs is gathering usage estimates and modelling the system's capacities.  However a very quick MVP of the kind of system that could support our functional requirements is desirable.  This can also guide our functional correctness and stand as a basis for identifying scalable bottlenecks and mitigations.

```
TODO - Show HLD
```

In our system the Create path could be (with minimal duplicate checking):

```
Create(req CreateRequest) returns CreateResponse:
  while true:
    rid = req.explicitID
    if rid == "":
      # Generate a 64bit random number and "truncate" it to ID_SIZE digit
      # number where each digit has 36 possibilities (A-Z + 0-9)
      randnum = rand64() % (36 ^ ID_SIZE)

      # b36encode converts a number into characters 
      # contain A-Z and 0-9  - similar to base64encoding
      rid = b36encode(randnum)

    newID = RID{
              id: rid,
              createdAt: now(),
              clientId: req.clientID
            }

    found_duplicates = insert newID into id_db

    if !found_duplicates:
      return newID

    if req.explicitID != "":
      return Error("Explicit ID already exists")
```


Our delete request is simple:

```
Delete(req DeleteRequest):
  id_db.DeleteById(req.rid)
```

# Step 3 - Does this scale?

## Capacity Estimates

Before we can evalute the system's scalability we need to understand the system's usage and capacity.

### Storage

```eqnml
$SizeOfID<"Size of ID Record">

Size per ID = ${25}->$SizeOfID bytes

Number of IDs created in 1 year:
    = $IDCreationQPS.Label * 86400 seconds per day * 365 days
    = ${IDCreationQPS.Label * 86400 * 365}->$IDSCreatedPerYear

Number of IDs created in 10 years:
    = $RetentionPeriod.Label * $IDSCreatedPerYear.Label
    = $RetentionPeriod * $IDSCreatedPerYear
    = ${RetentionPeriod * IDSCreatedPerYear}->$IDSCreatedIn10Years

Number of "digits" needed for $IDSCreatedIn10Years IDs, K 
    = Log($IDSCreatedIn10Years, Base = 36) 
    = ${Log(IDSCreatedIn10Years, Base = 36)}->$IDSize digits
```

### Network Bandwidth
```eqnml
= $IDCreationQPS.Label * $SizeOfID.Label
= $IDCreationQPS QPS * $SizeOfID bytes
= ${ToKBps(IDCreationQPS * SizeOfID)}->$NetworkBandwidth
```

```
TODO - Show modelling tool here?
```
