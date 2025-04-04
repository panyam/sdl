---
title: "Design DoorDash's Inventory Management System"
productName: 'Inventory Management System'
date: 2024-05-28T11:29:10AM
tags: ['inventory management', 'inventory', 'medium', 'DoorDash']
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: This system will manage the availability and stock levels of items across virtual stores, ghost kitchens, grocery vendors, and partner restaurants, ensuring that inventory updates happen in real-time to prevent overselling and optimize logistics.
---

{{ .FrontMatter.summary }}

## Functional Requirements

#### **1\. Real-Time Stock Updates**

* Ensure inventory reflects the latest stock levels **in real-time** across multiple vendors.  
* Automatic synchronization when:  
  * An order is placed.  
  * A vendor manually updates stock.  
  * A delivery issue (e.g., spoiled/damaged items) reduces available stock.

#### **2\. Multi-Vendor Inventory Tracking**

* Support multiple inventory sources (restaurants, grocery stores, warehouses).  
* Each vendor should be able to:  
  * Define stock levels per location.  
  * Set restocking frequency.  
  * Specify perishable vs. non-perishable items.

#### **3\. Low-Stock and Out-of-Stock Alerts** (P2)

* Notify vendors when stock levels are running low.  
* Notify users when an item they previously ordered is no longer available.  
* Option for vendors to enable automatic restocking if supported.

#### **4\. Perishable and Non-Perishable Item Handling**

* Support different inventory models for perishable (fresh food, groceries) vs. non-perishable (packaged items).  
* Allow vendors to set **expiration dates** and **automatic removal** for perishable goods.

#### **5\. Regional Availability and Inventory Partitioning**

* Inventory should be **location-aware** (i.e., users should only see items available in their delivery region).  
* Handle stock distribution across **multiple fulfillment centers, restaurants, and warehouses**.

#### **6\. Real-Time Availability for Users**

* The system should prevent users from adding out-of-stock items to their cart.  
* Items should be marked as “Limited Stock” when nearing depletion.  
* If an item goes out of stock after order placement, suggest alternatives before checkout.

#### **7\. Vendor API for Inventory Management**

* Vendors should be able to integrate their **own POS (Point of Sale) systems** with DoorDash via APIs.  
* APIs should allow:  
  * Bulk stock updates.  
  * Individual item-level stock modifications.  
  * Scheduled inventory synchronization.

#### **8\. Bulk and Scheduled Inventory Updates**

* Support vendors updating inventory in bulk (e.g., CSV uploads or API requests).  
* Scheduled restocking updates (e.g., “Restock every morning at 6 AM”).

#### **9\. Fraud and Error Detection**

* Detect anomalies in stock levels (e.g., sudden depletion due to incorrect reporting).  
* Prevent overselling due to race conditions (e.g., concurrent users adding limited-stock items to cart).

#### **10\. Inventory Synchronization with Marketplace Partners**

* Allow direct integration with **third-party grocery stores and warehouses**.  
* Sync inventory data between **DoorDash and partners like Walmart, Safeway, or Costco**.

---

### **Requirements That Are Out of Scope for This Problem**

**1\. Order Matching & Dispatch System**

* This is a separate problem that handles **assigning drivers to deliveries**, optimizing routes, and ensuring timely fulfillment.  
* See: **Real-Time Delivery Dispatch System Design**.

**2\. Recommendation and Search Ranking for Items**

* Suggesting popular or trending inventory items belongs in a **Discovery/Search system**.  
* See: **DoorDash’s Search and Discovery System**.

**3\. Demand Prediction and Surge Pricing**

* Predicting which inventory items are likely to go out of stock due to demand spikes is a **Machine Learning problem**.  
* See: **Real-Time Demand Prediction System for Surge Pricing**.

## **1\. Key Entities**

These are the primary entities involved in managing real-time inventory for vendors (restaurants, grocery stores, virtual kitchens, etc.).

### **Core Entities**

#### **1\. Vendor**

* Represents a restaurant, grocery store, or fulfillment center managing inventory.  
* Attributes:  
  * `vendor_id` (UUID) 🔑  
  * `name`  
  * `location_id` (Geospatial reference)  
  * `inventory_sync_enabled` (boolean)  
  * `sync_frequency` (e.g., every 5 minutes)

#### **2\. Inventory Item**

* Represents a unique item managed in the inventory.  
* Attributes:  
  * `item_id` (UUID) 🔑  
  * `vendor_id` (FK)  
  * `name`  
  * `category` (e.g., Grocery, Fast Food, Beverage)  
  * `price`  
  * `stock_quantity`  
  * `low_stock_threshold`  
  * `is_perishable` (boolean)  
  * `expiry_date` (optional, for perishable items)

#### **3\. Stock Transaction**

* Represents an event that modifies inventory (purchase, restocking, spoilage, etc.).  
* Attributes:  
  * `transaction_id` (UUID) 🔑  
  * `item_id` (FK)  
  * `vendor_id` (FK)  
  * `change_type` (enum: PURCHASE, RESTOCK, SPOILAGE, ADJUSTMENT)  
  * `quantity_change` (int)  
  * `timestamp`

#### **4\. Inventory Event Queue**

* An internal mechanism for real-time inventory updates.  
* Attributes:  
  * `event_id` (UUID) 🔑  
  * `vendor_id` (FK)  
  * `item_id` (FK)  
  * `event_type` (enum: STOCK\_UPDATE, LOW\_STOCK\_ALERT, OUT\_OF\_STOCK)  
  * `previous_stock`  
  * `new_stock`  
  * `timestamp`

  ---

## **2\. Key APIs**

These APIs allow vendors to manage inventory and ensure real-time synchronization.

### **1\. Vendor Inventory Management APIs**

#### **➤ Get Inventory for a Vendor**

```
GET /vendors/<vendor_id>/inventory
```

**Purpose**: Fetch the current inventory of a vendor.

**Response**:

```
{

  "vendor_id": "123e4567-e89b-12d3-a456-426614174000",

  "items": [

    {

      "item_id": "abc123",

      "name": "Cheeseburger",

      "stock_quantity": 25,

      "low_stock_threshold": 5,

      "is_perishable": true,

      "expiry_date": "2025-03-10"

    }

  ]

}
```

#### **➤ Update Stock for an Item**

```
POST /vendors/<vendor_id>/inventory<item_id>/update
```

**Request Body**:

```
{

  "change_type": "PURCHASE",

  "quantity_change": -1

}
```

**Response**:

```
`{`

  `"item_id": "abc123",`

  `"previous_stock": 25,`

  `"new_stock": 24`

`}`
```

#### **➤ Bulk Inventory Update (CSV or JSON Upload)**

```
POST /vendors/<vendor_id>/inventory/bulk_update
```

**Request Body (JSON Example)**:

```
`{`

  `"items": [`

    `{ "item_id": "xyz456", "stock_quantity": 100 },`

    `{ "item_id": "abc123", "stock_quantity": 50 }`

  `]`

`}`
```

**Response**:

```
`{` `"updated_items": 2` `}`
```

---

### **2\. Inventory Event APIs**

#### **➤ Subscribe to Inventory Updates (Webhook for Real-Time Sync)**

```
POST /vendors/<vendor_id>/inventory/subscribe
```

**Request Body**:

```
`{`
  `"callback_url": "https://example.com/webhook",`
  `"event_types": ["LOW_STOCK_ALERT", "OUT_OF_STOCK"]`
`}`
```

**Response**:

```
`{ "subscription_id": "sub-123456" }`
```

#### **➤ Get Real-Time Stock Events**

```
GET /inventory/events?vendor_id={vendor_id}
```

**Response**:

```
{
  "events": [
    {
      "event_id": "evt-789",
      "item_id": "abc123",
      "event_type": "LOW_STOCK_ALERT",
      "previous_stock": 6,
      "new_stock": 5
    }
  ]
}
```

---

## **3\. High-Level System Design**

This is a **high-level architecture** for handling inventory updates, ensuring real-time synchronization, and avoiding stock inconsistencies.

### **Architecture Components**

1. **API Gateway**  
   * Routes all vendor inventory API requests.  
   * Manages rate limiting, authentication, and security.  
2. **Inventory Service**  
   * Core logic for fetching, updating, and validating inventory data.  
   * Manages stock changes and triggers events for critical thresholds.  
3. **Database (PostgreSQL \+ Redis)**  
   * **PostgreSQL (or DynamoDB)** stores vendor and inventory data.  
   * **Redis** caches frequently accessed inventory to improve read performance.  
4. **Event-Driven System (Kafka / Pub/Sub / Kinesis)**  
   * **Why?** Ensures stock updates propagate across services with low latency.  
   * **Producers**: Inventory service publishes stock change events.  
   * **Consumers**: Downstream services (search, recommendations) consume updates.  
5. **Low-Stock Alert Processor**  
   * **Detects** when stock falls below the threshold.  
   * **Publishes alerts** to vendors via webhooks or messages.  
6. **Inventory Sync Job**  
   * **Scheduled Jobs** sync inventory updates from external vendors.  
   * Supports **bulk imports from large retailers (e.g., Walmart, Safeway).**


```
           ┌──────────────────────────┐
           │      API Gateway         │
           └─────────▲────────────────┘
                     │
     ┌───────────────┼─────────────────────────┐
     │               │                         │
┌────▼──────┐  ┌────▼──────┐            ┌──────▼──────────┐
│Inventory  │  │ Event Bus │            │ Low-Stock       │
│Service    │  │ (Kafka)   │            │ Alert Processor │
└───────────┘  └────▲──────┘            └───────▲─────────┘
                    │                           │
            ┌───────┴─────────┐                 │
            │  Search Cache   │   ┌─────────────┘
            │  (Redis)        │   │
            └───────▲─────────┘   │
                    │             │
            ┌───────┴────────┐    │
            │ Postgres/NoSQL │    │
            │ (Inventory DB) │    │
            └────────────────┘    │
                                  │
     ┌────────────────────────────┘
     │
┌────▼──────┐
│ Vendor API│
│  Clients  │
└───────────┘
```

## **Step-by-Step Request Lifecycle**

From the user's perspective there are 2 flows:

1. Get the availability and details of an item

```
GET /items/{itemid}
```

This is a straight foward read from the items DB

2. Place an order for N items

```
POST /order/
BODY: List of [(item, count, price)]

def placeOrder(items):
  for item in items:
    result = update_inventory(item.id, -item.count)     # this takes care of sending low stock alerts
    if result = "out of stock":
      kick off refund flow or "find replacement" or call user flow
```

### **Step 1: API Gateway (Entry Point)**

🔹 **Component:** API Gateway (NGINX, Kong, Apigee)  
🔹 **Purpose:**

* Handles authentication and rate limiting.  
* Ensures the request is **well-formed** and **authorized**.  
* Routes the request to the appropriate backend service.

🔹 **Failure Scenarios & Mitigation:**

| Failure | Mitigation |
| ----- | ----- |
| **Invalid API key** | Return 401 Unauthorized |
| **Too many requests (rate limit exceeded)** | Return 429 Too Many Requests |

---

### **Step 2: Inventory Service (Processing Logic)**

🔹 **Component:** Inventory Service (Go, Node.js, Java)  
🔹 **Purpose:**

* Validates request payload.  
* Ensures **idempotency** using a unique `transaction_id`.  
* Reads **current stock level** from the database.  
* Checks if **stock is sufficient** before allowing deduction.

🔹 **Failure Scenarios & Mitigation:**

| Failure | Mitigation |
| ----- | ----- |
| **Stock level is already 0** | Return 400 Bad Request (Out of Stock) |
| **Vendor ID or Item ID not found** | Return 404 Not Found |
| **Duplicate request (e.g., retries from client)** | Deduplicate using `transaction_id` |

---

### **Step 3: Database Write Operation**

🔹 **Component:** Primary Database (PostgreSQL, DynamoDB)  
🔹 **Purpose:**

* Executes an **atomic update** (e.g., `UPDATE inventory SET stock_quantity = stock_quantity - 1`).  
* Ensures consistency using **transactions**.

🔹 **Failure Scenarios & Mitigation:**

| Failure | Mitigation |
| ----- | ----- |
| **Database deadlock** | Retry using exponential backoff |
| **Primary DB failure** | Failover to replica, queue update for retry |
| **High write contention (multiple updates at once)** | Use optimistic locking or distributed transactions |

---

### **Step 4: Event-Driven Stock Update**

🔹 **Component:** **Kafka / Pub/Sub / Kinesis (Event Bus)**  
🔹 **Purpose:**

* Publishes an **inventory update event** (`ITEM_STOCK_UPDATED`).  
* Ensures eventual consistency across all services.  
* Allows downstream services (search, recommendations) to react.

🔹 **Example Event:**

```
{

  "event_id": "evt-789",

  "vendor_id": "123",

  "item_id": "abc123",

  "old_stock": 25,

  "new_stock": 24,

  "timestamp": "2025-02-26T12:34:56Z"

}
```

🔹 **Failure Scenarios & Mitigation:**

| Failure | Mitigation |
| ----- | ----- |
| **Message bus failure** | Store in Dead Letter Queue (DLQ) for retries |
| **Consumer failure (downstream services not consuming messages)** | Implement retry logic, monitor queue depth |

---

### **Step 5: Cache Invalidation & Updates**

🔹 **Component:** **Redis / Memcached (Cache Layer)**  
🔹 **Purpose:**

* Updates cached inventory data to reflect the latest stock level.  
* Ensures **fast reads** for API queries.

🔹 **Cache Update Strategy:**

1. **Write-Through**: Update cache synchronously after DB write.  
2. **Write-Back**: Update DB and cache asynchronously for performance.  
3. **TTL-Based Expiry**: Set a **1–5 second expiry** to force fresh DB reads.

🔹 **Failure Scenarios & Mitigation:**

| Failure | Mitigation |
| ----- | ----- |
| **Cache inconsistency (DB and cache mismatch)** | Implement cache invalidation strategy |
| **Cache eviction (entry removed due to memory pressure)** | Fallback to DB reads |

---

### **Step 6: Vendor & User Notifications**

🔹 **Component:** **Low-Stock Alert Processor**  
🔹 **Purpose:**

* If stock drops below a **low-stock threshold**, send alerts to the vendor.  
* If item is **out-of-stock**, remove it from user-facing search results.

🔹 **Failure Scenarios & Mitigation:**

| Failure | Mitigation |
| ----- | ----- |
| **Notification service failure** | Retry sending notifications using job queue |
| **User adds item to cart but it goes out of stock before checkout** | Show "out of stock" message at checkout |

---

### **Step 7: Search & Discovery System Sync**

🔹 **Component:** **Search Index Updater**  
🔹 **Purpose:**

* Updates search database to **reflect real-time inventory changes**.  
* If an item goes out of stock, **temporarily remove it from search results**.

🔹 **Failure Scenarios & Mitigation:**

| Failure | Mitigation |
| ----- | ----- |
| **Search index update delay** | Periodic batch reindexing |

## Scaling and Fault Tolerance

### **1\. Scaling Strategies**

Scaling strategies ensure the system can handle a high volume of inventory updates from thousands of vendors while providing real-time availability to users.

#### **1.1 Scaling the Read Path (Fetching Inventory Data)**

* **Use Caching (Redis / Memcached) for Fast Reads**  
  * Frequently accessed inventory data should be **cached** at the API layer.  
  * Example: Store ```vendor_id → {items[], stock levels}``` in Redis.  
  * **Trade-off**: Slightly stale data (\~1–2 sec lag) in exchange for better performance.  
* **Read Replicas for Load Distribution**  
  * Scale **PostgreSQL / DynamoDB** with **read replicas** to handle high QPS.  
  * **Trade-off**: Eventual consistency between the primary DB and replicas.  
* **Partitioning Inventory by Vendor or Region**  
  * Instead of a single **monolithic table**, **shard data** based on:  
    * **Vendor ID** → Useful for independent vendor scaling.  
    * **Region/City** → Efficient for location-based filtering.  
  * **Trade-off**: Requires careful routing logic in API Gateway.

---

#### **1.2 Scaling the Write Path (Handling Stock Updates)**

* **Eventual Consistency for Inventory Updates**  
  * Instead of **synchronous writes** to the database, use **event queues (Kafka/Pub/Sub)** to handle stock updates **asynchronously**.  
  * Vendors receive **near real-time updates (\~100ms delay)** but benefit from better throughput.  
  * **Trade-off**: Risk of temporary stale inventory data before events propagate.  
* **Batch Processing for Bulk Stock Updates**  
  * Large vendors may push **bulk inventory updates** via CSV uploads or APIs.  
  * Instead of processing synchronously, use **batch processing** (e.g., scheduled jobs every 5 minutes).  
  * **Trade-off**: Not truly real-time but necessary for scalability.  
* **Idempotency to Prevent Duplicate Stock Updates**  
  * Vendors may **retry API calls** in case of failures.  
  * Each update request should include a **unique transaction ID** to ensure idempotency.

---

### **2\. Fault Tolerance Mechanisms**

Fault tolerance ensures that the system remains operational even when components fail.

#### **2.1 Handling Database Failures**

* **Primary-Replica Failover**  
  * If the primary DB fails, **automatically promote a read replica**.  
  * Use **hot standbys** (e.g., AWS RDS Multi-AZ, DynamoDB global tables).  
* **Write-Ahead Logs for Data Recovery**  
  * Maintain an append-only **transaction log** to recover lost writes.  
* **Circuit Breaker for Slow Queries**  
  * If DB queries start timing out, **serve stale data from cache** instead.

---

#### **2.2 Handling Message Queue Failures**

* **Kafka with Retry Mechanisms**  
  * Ensure **at-least-once delivery** of inventory events.  
  * Messages should be retried with **exponential backoff** to prevent overload.  
* **Dead Letter Queue (DLQ) for Failed Events**  
  * If an inventory event fails to process, send it to a **DLQ** for manual review.

---

#### **2.3 Handling API Failures**

* **Rate Limiting and API Throttling**  
  * Prevent a **single vendor** from overloading the system with excessive updates.  
  * Use **token bucket rate limiting** (e.g., `X updates per second per vendor`).  
* **Graceful Degradation with Fallback Mechanisms**  
  * If real-time stock data is **unavailable**, fallback to the **last known stock** in Redis.

---

### **3\. Trade-Offs**

Every design decision comes with **trade-offs**. Here are some key ones to consider.

| Decision | Pros | Cons |
| ----- | ----- | ----- |
| **Caching inventory in Redis** | Reduces DB load, speeds up API reads | Data might be slightly stale (1-2s delay) |
| **Asynchronous inventory updates with Kafka** | Handles high volume, improves system responsiveness | Inventory updates are eventual, not instant |
| **Partitioning inventory by vendor ID** | Improves scalability, parallel processing | More complex query logic, cross-vendor joins are harder |
| **Using read replicas for load balancing** | Allows horizontal scaling for high QPS | Introduces eventual consistency issues |
| **Dead Letter Queue for failed inventory updates** | Prevents dropped updates | Requires manual intervention for failed updates |
| **Batch processing bulk updates** | Prevents system overload | Not real-time (delays up to a few minutes) |

---

### **4\. Scaling Inventory Data Beyond a Single Region**

If DoorDash expands globally, how do we handle **inventory across multiple geographies**?

* **Use Multi-Region Database Replication**  
  * Each region has its own **inventory database** (e.g., **US-East, US-West, Europe**).  
  * Global services (like **search**) fetch inventory from the closest region.  
* **Geo-Partitioned Inventory Stores**  
  * Inventory is **sharded per region** to reduce latency.  
  * Example: **Restaurants in California query West Coast DB**, those in New York query East Coast DB.  
* **Distributed Event Processing**  
  * Kafka topics are **partitioned per region**.  
  * Ensures **localized inventory updates** without cross-region lag.

---

### **5\. Monitoring & Observability**

Observability is crucial to detect failures in real-time.

* **Metrics to Monitor**  
  * Inventory update latency (P99)  
  * Event processing failure rate  
  * DB query performance (slow queries)  
  * Cache hit/miss ratio  
* **Tools for Observability**  
  * **Distributed Tracing**: OpenTelemetry for tracking API requests.  
  * **Log Aggregation**: ELK Stack, Splunk, or Datadog for logging.  
  * **Metrics & Alerts**: Prometheus \+ Grafana for monitoring trends.
