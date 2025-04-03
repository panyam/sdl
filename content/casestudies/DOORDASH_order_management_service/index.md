---
title: "Design DoorDash's Order Management System"
productName: 'DoorDash Order Management System'
date: 2025-02-07T11:29:10AM
tags: ['doordash', 'medium', 'orders', 'order management', 'ubereats', 'location' ]
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: 'An order management system for DoorDash'
---

# Design DoorDash's Order Management System (OMS)

The Order Management System (OMS) is responsible for handling the entire lifecycle of an order, from creation to fulfillment. 

---






---

<D-c>

### Functional Requirements (~2-3 minutes)

Here are the core functional requirements a user (customer, restaurant, Dasher, support agent) would expect:

The functional requirements are:

1. Users should place orders (order info below)
2. Track their orders live (ish)
3. Merchants should see queued up orders and update order status (accept, reject, cancel, ready for pickup)
4. Order status by merchants and dashers across various states
5. Handle failure cases (reasonably)
6. Assume: all orders are from a single store
   - For order that spans multiple stores, we can treat it as seperate orders or have each Order being part of a large
     "Pickup" and each order can have its own status etc.  This is more complicated but doable.

Scale Requirements:

Usage Stats: https://backlinko.com/doordash-users#doordash-usage-statistics

* 3B orders a years (250M Orders a month ~ 8M orders a day)
* 600k businesses
* 50M MAU
* 7M Active Dashers (may be upto 10M now)
* On average - 500 orders per merchant per day (250M / 600k) ~ 50 per hour (assuming 10 hours of operation)
* Peak may be 100 in an hour (note this would need a merchant to have 50 chefs - so this is extreme)

### Extensions (Out of scope)

Functional:

1) Restaurant menu management/search - Assume restaurant onboarding, and menu manageemnt is taken care off.
2) Payment Management - Assume payments, refunds are all taken care off
3) Notification systems (tracking will be live on the page - but email/sms etc not in scope)
4) [Order dispatch system](../delivery_dispatch_system/index.md) - we assume a dasher is assigned automatically (can be designed otherwise)
5) Authentication - we assume a user is logged in (or in the case of guest checkout cookie/session based)
6) Order mutability - Assume orders once created are it (or they can be cancelled in some cases - but we wont deal with
   extra charges on this etc)

Non Functional:
1) Observability - So we can ensure reliability, uptime, attacks etc
2) Disaster recovery - comes under reliabiliability and is a great way to go into which parts are "effectful" and which arent.
3) CI/CD, Deployability, Experimentation

Both:
1) Tiered SLOs (usage/pricing/billing etc)

### Non functional requirements (~2-3 minutes)

* Fault Tolerance, Availability and Durability:
  - System should be available so users can place orders without seeing failures.
  - Ok for orders to take a slightly bit longer but should be durable
  - Once placed orders cannot be lost
* Consistency:
  - Eventual consistency is ok for tracking,
  - Hard consistency needed when updating states
* Scalable to our needs with good UX.   Order creations and updates should happen in 200ms (not including payment or external systems).
* Update tracking should be near real-time - Want freshness to be say within 1-5s of updates.

---

## API/Interface/Entities (2-3 minutes)

### Entities
   
```
// User and Dasher are just general entities
record User { id string }
record Dasher {
  Id string   // ID of the dasher 
}

record Order {
  Id string
  CreatedAt TimeStmp
  UpdatedAt TimeStmp
  CustomerId string
  RestaurantId string
  DasherId string
  Status enum { PLACED, CONFIRMED, COOKING, READY_FOR_PICKUP, DELIVERING, DELIVERED, CANCELLED }
  Items List<OrderItem>
  
  Priority int
  PickupAt LatLong
  DropoffAt LatLong
  
  // Only keep the total price
  TotalPrice double
  
  // We may keep payment details here or else where depending on what is visible etc
  PaymentDetails any
  
  // version etc for OCC
  Version int
}

record OrderItem {
  Id string         // item id
  StoreId string    // which store is this item from
  Name string
  Quantity string
  Price string      // how much the user is paying for it
  Customizations Json   // eg extra onions etc
}
```

### Services/APIs (~5 minutes)

Now for the APIs/Endpoints needed.

```
service OrdersService {
  CreateOrder(order Order) Order {    // with Id, Status, CreatedAt, CreatedUser set
    POST "/orders"
    body: {order}
  }
  GetOrder(orderId string) Order {    // Get details of an order
    GET /orders/{orderId}
  }
  UpdateOrder(orderId string, patch Patch<Order>) Order {    // Updated order
    PUT /orders/{orderId}
    body: {patch}
  }
  ListMerchantOrders(merchantId string) Order[] {    // Updated order
    PUT /orders/?merchantId={merchantId}
  }
  ListUserOrders(userId string) Order[] {    // Updated order
    PUT /orders/?userId={userId}
  }
  ListDasherOrders(dasherId string) Order[] {    // Updated order
    PUT /orders/?dasherId={dasherId}
  }
}
```

---

## High Level Design (~ 5-10 minutes)

<Drawing id="/hld" preview="./hld.png" width="800px" >High Level Design</Drawing>

Key flows:

1. User places an order - CreateOrder is called (assuming payment successful)
2. CreateOrder creates an Order entry in the DB with status = PLACED
3. Merchant can either poll or subscribe to see new orders.  Merchant can decline - (but a capacity management system
   can be in front of this).  Acceptance is more interesting. 
4. Status is now CONFIRMED
5. At some point merchant starts COOKING and marks it as COOKING - also not interesting as they are just on it.
6. Once done status = READY_FOR_PICKUP
7. Order management system now gets this order and performs matching and dasher picks up. 
8. status = DELIVERING and now it is upto the driver to set status to DELIVERED when done.

Basic "series of handoffs" problem.

While delivering the app could send status updates on where the driver is at - but that can either be here or on a
driver location updates problem.

---

## Deep Dives

Now time to go through the deep dive and address concerns.

* We have 8M orders placed a day - so about 80 orders per second created.
* An order may take upto 30 minutes to be fully delivered (from say all the states).
* During its lifecycle an order may be updated say once a minute (by merchant and dasher combined as it goes through
  various states.
* But will be read more frequently - say every 5s (as user is tracking it).  So during 30mins - that is about 400 reads
  (30mins * 60s / 5 = 360 ~ 400)
* Our read qps on an Order (to track status) is about 80 * 400 = 32K QPS
* On the listing side - merchant and dasher may say poll once a minute to see new orders (we will also consider pushes)
* If they poll we are looking at 7M + 600k order "searches".

A few assumptions and tradeoffs.

1. We can assume that a merchant's overwhelming time is in cooking so checking for new orders in real time is not
   realistic - they would look at it when say a "chef" is free - assuming once a minute.  So a real time push is not
   really needed yet - atleast hard to justify complexity.
2. When would subscription be feasible? If time to ack needs to be fast (user doesnt want to wait for a minute to know
   if you can even get to my order) but time to cook can be relaxed.   Say even here - if a 5s delay is feasible
   would be polling be ok?
3. On the dasher side - polling may again not be a bad option again since a dasher is driving mostly and order matching
   would be more nuanced.

So assuming polling we are looking at:

- Number of merchants = 600k
- ListMerchantOrders QPS = 600k / 5 ~ 120k QPS
- Number of dashers = 7M
- ListDasherOrders QPS - 7M / 5 = 1.4M QPS

Also note dasher and merchant connectivity may not be as "solid" - bad network, bad machines etc

So to scale this we have a few options.

1. Storage

* 80 M orders per day
* 10kb per order
* 800GB per day 
* 320 TB per year  (you dont even need to keep orders more than a day in "hot" storage and the rest in cold or archival mode for auditing if needed)
  
Assuming a day of storage for "live working" (users' orders can be in another user specific index, orders > 1d could be
in warm/cold tiers) we are looking at a TB of data.

Total Write QPS:
1. Each order = 1 Write (for creation) + 10 updates (during lifecycle going throuhg various states)
2. Total write QPS per day = 80M ~ 800 writes per second.
3. Note the real reads are coming from the ByUser, ByMerchants and ByDasher indexes (and by status for the order management
   system)
4. So our writes also need to update these DBs - however we only need RYW consistency for ByUser index - rest can be
   eventual (as global secondary indexes with partiiton key being user, merchant and dasher respectively)

In this system we have:

<Drawing id="/with_secondary_indexes" preview="./secondary_indexes.png" width="800px" >With Secondary Indexes</Drawing>
![alt System with Secondary Indexes](./secondary_indexes.png "With Secondary Indexes")

In this scheme we have:

1. User submits an order and it is created in the Orders DB
2. Since we want RYW for the user to see their own order, we can do either:
   a) transactional writes on the OrderDB and the ByUser index.
   b) Or since a user will be more interested in a "particular" order instead of a full listing - the order ID would go
      to the orders DB anyway.  By the time a listing is performed we can assume it is eventually materialized in the
      ByUser index.
   c) For something truly scalable and consistent we an have the ByUser as our primary table, have a global sharding by
      userId and then have the order matching system with multiple consumers per shard.
3. The ByMerchant and ByDriver indexes can be built asynchronously for eventually consistency
4. Merchants and Dashers can poll their respective indexes.

Since we are looking at about a terrabyte of data per day - we are jsut better off using in memory caches (either
distributed and/or replicated).   2M reads on this cache is reasonably fast - memory optimized ec2 instances can do 100k
qps.  So a distributed cache of 20 machines (40 with replication) will do.  Note write qps is low so evictions can be
LRU (and/or on cancelled or ready_for_pickup or delivered)

If a truly shareded was TRULY required:

1. ByX stores can be in cassandra or GSI scheme on something like DynamoDB will do.
