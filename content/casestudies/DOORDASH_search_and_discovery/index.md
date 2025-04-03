---
title: "Design DoorDash's Search and Discovery"
productName: 'DoorDash Search and Discovery System'
date: 2025-02-07T11:29:10AM
tags: ['doordash', 'medium', 'search', 'discovery', 'yelp', 'location' ]
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: 'A search and discovery system for DoorDash'
scrollToBottom: true
---

# Design DoorDash's Search and Discovery System

For DoorDash’s Search and Discovery System, the key functional requirements should align with how users interact with the platform when searching for restaurants and food items.

---

### Functional Requirements (~2-3 minutes)

#### **1. Search for Restaurants & Menu Items**

* Users should be able to search for restaurants based on:  
  * Name (e.g., “McDonald’s”)  
  * Cuisine type (e.g., “Italian”)  
  * Specific dish or ingredient (e.g., “Spicy Chicken Sandwich”)  
* **Exclusion:** This does not include search optimization for non-food items (better suited for an **Inventory Management System**).

#### **2. Filtering and Sorting** (Make this 1.1 instead of 2)

* Users should be able to filter results by:  
  * Cuisine type (Mexican, Italian, Vegan, etc.)  
  * Price range (budget, mid-range, premium)  
  * Ratings (e.g., only show 4+ star restaurants)  
  * Delivery time (e.g., under 30 minutes)  
* Sorting options should include:  
  * Relevance (default)  
  * Fastest delivery  
  * Best rated  
  * Price (low to high, high to low)

#### **3. Geolocation-Based Results**

* Results should be **localized** based on:  
  * User’s current location  
  * Preferred delivery address  
  * Nearby trending restaurants  
* If a user searches in an unavailable location, suggest **nearby areas**.

#### **4. Handling Restaurant Availability**

* Display only open restaurants by default.  
* Option to show closed restaurants with expected open times.  
* Handle real-time inventory updates (e.g., sold-out items marked as unavailable).

---

### **OUT OF SCOPE**

#### **1. Multi-Platform Support**

* Ensure seamless experience across:  
  * Mobile app (iOS, Android)  
  * Web browser  
  * Voice search (Alexa, Google Assistant)

#### **2. Autocomplete and Typo Handling**

* Search should **auto-suggest** restaurant names, cuisines, and dishes.  
* Support fuzzy search for typos (e.g., “Chick Filay” → “Chick-fil-A”).

#### **3. Trending and Popular Suggestions (BECOMES ANALYTICS PROBLEM)**

* Display trending restaurants and dishes based on:  
  * Location-specific demand  
  * Time of day (e.g., breakfast, lunch, dinner)  
  * Seasonal trends (pumpkin spice latte in fall)  
* Highlight editor picks, top-rated local favorites.

#### **4. Promotions and Sponsored Listings (SIMILAR TO RANKING PROBLEM)**

* Support for:  
  * Sponsored restaurant placements  
  * Promotions and discounts affecting ranking

#### **2. Ranking and Personalization**

* Restaurants and menu items should be **ranked** based on:  
  * Distance and estimated delivery time  
  * Ratings and reviews  
  * User preferences (e.g., previous orders, favorite cuisines)  
  * Promotions or discounts  
* Personalization should be based on historical data and user behavior.

#### Other Features

| Feature | Better Suited System |
| ----- | ----- |
| Order placement & tracking | **Order Management System** |
| Real-time driver assignment | **Delivery Dispatch System** |
| Payment processing | **Payment Processing System** |
| Menu recommendations based on dietary restrictions | **Personalized Recommendation System** |
| Inventory tracking & stock updates | **Inventory Management System** |
| Fraud detection in reviews/ranking | **Review Moderation System** |

---

## **Key Entities**

The system needs to store and retrieve search-related data efficiently. Below are the primary entities:

### **1.1 Restaurant**

Represents a restaurant that users can search for.

| Field Name | Type | Description |
| ----- | ----- | ----- |
| `restaurant_id` | UUID | Unique identifier for the restaurant |
| `name` | String | Restaurant name |
| `location` | GeoPoint | Latitude/longitude of the restaurant |
| `cuisine` | Enum | Cuisine type (e.g., Mexican, Italian) |
| `rating` | Float | Average customer rating |
| `delivery_time` | Integer | Estimated delivery time (in minutes) |
| `is_open` | Boolean | Whether the restaurant is currently open |
| `menu_items` | List<MenuItem> | Menu items available |

### **1.2 Menu Item**

Represents a specific food item available at a restaurant.

| Field Name | Type | Description |
| ----- | ----- | ----- |
| `menu_item_id` | UUID | Unique identifier for the menu item |
| `restaurant_id` | UUID | Foreign key referencing the restaurant |
| `name` | String | Name of the dish |
| `category` | String | Category (e.g., Appetizer, Main Course) |
| `price` | Float | Price of the item |
| `availability` | Boolean | Whether the item is currently available |
| `tags` | List<String> | Keywords for search (e.g., "spicy", "vegan") |

### **1.3 User Search Query**

Represents the details of a user’s search request.

| Field Name | Type | Description |
| ----- | ----- | ----- |
| `query_id` | UUID | Unique identifier for search request |
| `user_id` | UUID | User performing the search |
| `query_text` | String | Raw search input |
| `filters` | JSON | Applied filters (e.g., rating > 4) |
| `location` | GeoPoint | User's location for geospatial search |
| `timestamp` | Timestamp | When the search was made |

---

## **Core APIs**

The system will expose RESTful (or GraphQL) APIs to interact with the search and discovery engine.

### **2.1 Search Restaurants API**

**Endpoint:** `GET /search/restaurants`

**Request Query Params:**

| Parameter | Type | Required | Description |
| ----- | ----- | ----- | ----- |
| `query` | String | Yes | Search string (e.g., "pizza") |
| `location` | GeoPoint | Yes | User’s location (lat/lon) |
| `radius_km` | Integer | No | Search radius in km (default: 10km) |
| `cuisine` | String | No | Filter by cuisine |
| `sort_by` | Enum | No | Sort by `rating`, `delivery_time`, `popularity` |

**Response:**

```
{  
  "restaurants": [  
    {  
      "restaurant_id": "123",  
      "name": "Pizza Palace",  
      "location": { "lat": 37.7749, "lon": -122.4194 },  
      "cuisine": "Italian",  
      "rating": 4.5,  
      "delivery_time": 30  
    }  
  ]  
}
```

---

### **2.2 Search Menu Items API**

**Endpoint:** `GET /search/menu-items`

**Request Query Params:**

| Parameter | Type | Required | Description |
| ----- | ----- | ----- | ----- |
| `query` | String | Yes | Search string (e.g., "spicy chicken") |
| `location` | GeoPoint | Yes | User’s location (lat/lon) |
| `radius_km` | Integer | No | Search radius in km (default: 10km) |
| `restaurant_id` | UUID | No | Filter by restaurant |
| `category` | String | No | Filter by category (e.g., "Desserts") |

**Response:**

```
{  
  "menu_items": [  
    {  
      "menu_item_id": "987",  
      "name": "Spicy Chicken Sandwich",  
      "restaurant_id": "123",  
      "restaurant_name": "Burger Barn",  
      "price": 8.99,  
      "availability": true  
    }  
  ]  
}
```

---

### **2.3 Get Popular & Trending Items**

**Endpoint:** `GET /search/trending`

**Request Query Params:**

| Parameter | Type | Required | Description |
| ----- | ----- | ----- | ----- |
| `location` | GeoPoint | Yes | User’s location (lat/lon) |
| `time_of_day` | String | No | Morning/Lunch/Dinner filtering |
| `season` | String | No | Seasonal trends (e.g., "winter") |

**Response:**

```
{  
  "trending_items": [  
    { "name": "Pumpkin Spice Latte", "restaurant": "Starbucks", "popularity_score": 95 }  
  ]  
}
```

---

## **3. High-Level System Design**

This is a **high-level architecture** for handling search queries efficiently.

### **3.1 Architecture Overview**

```
                           `+--------------------------+`  
                            `|      User Device        |`  
                            `|  (Mobile/Web/Voice)     |`  
                            `+--------------------------+`  
                                      `|`  
                                      `v`  
                            `+--------------------------+`  
                            `|   API Gateway (GraphQL)  |`  
                            `+--------------------------+`  
                                      `|`  
        `+----------------------------------------------------+`  
        `|                                                    |`  
        `v                                                    v`  
`+----------------------+                           +----------------------+`  
`| Search Service      |                           | Personalization Engine|`  
`| (Query Processing)  |                           | (Ranks & filters)     |`  
`+----------------------+                           +----------------------+`  
        `|                                                    |`  
        `v                                                    v`  
`+------------------------+                           +----------------------+`  
`| Elasticsearch Cluster  |                           | Recommendation System |`  
`| (Restaurant/Menu Index)|                           | (Trending, Favorites) |`  
`+------------------------+                           +----------------------+`  
        `|                                                    |`  
        `v                                                    v`  
`+------------------------+                           +----------------------+`  
`| PostgreSQL / DynamoDB  |                           | Event Stream (Kafka)  |`  
`| (Restaurant/Menu Data) |                           | (Search logs, Trends) |`  
`+------------------------+                           +----------------------+`
```

### **3.2 Components Explained**

[1](1). **API Gateway**: Exposes REST/GraphQL APIs and handles auth, rate-limiting.  
2. **Search Service**: Parses user queries, applies filters, and fetches relevant results.  
3. **Elasticsearch Cluster**: Stores **pre-indexed** restaurant and menu item data for **fast text search**.  
4. **Personalization Engine**: Ranks results based on user preferences, location, trending items.  
5. **Recommendation System**: Handles trending searches, personalized suggestions.  
6. **PostgreSQL / DynamoDB**: Stores structured restaurant and menu data.  
7. **Event Stream (Kafka)**: Captures search logs for trending analysis.


## Deep Dive 

### **1. Scaling Strategies**

To ensure **fast**, **reliable**, and **cost-efficient** search operations at DoorDash’s scale, we focus on:

#### **1.1 Scaling Reads Efficiently (Handling QPS in Millions)**

* **Elasticsearch Indexing for Fast Lookups**  
  * Restaurants & menu items are pre-indexed in **Elasticsearch** for **fast text searches**.  
  * **Partitioning by geographic region** to avoid global indexing issues.  
  * **Asynchronous index updates** when restaurants update their data.  
* **Distributed Cache for Hot Queries**  
  * **Redis / Aerospike** to cache frequently searched results.  
  * **Geohashed-based caching** (cache search results at the granularity of city/local region).  
  * **User-Specific Query Caching** (e.g., store results of past queries for repeat users).  
* **Load Balancing Across Search Nodes**  
  * **Multiple Elasticsearch clusters** in different geographic locations.  
  * Use **Geo-DNS routing** to direct users to the closest cluster.  
  * Deploy **multiple API Gateways with autoscaling** to balance traffic.

#### **1.2 Scaling Writes Efficiently (Handling Indexing Updates)**

* **Write-through Caching:** Updates to restaurant/menu data **invalidate cache** immediately.  
* **Event-Driven Data Pipeline**  
  * **Kafka / Pulsar** to queue incoming updates.  
  * **Stream Processing (Flink / Spark Streaming)** to index updates asynchronously.  
* **Bulk Indexing Strategy**  
  * Instead of indexing updates in real-time (high cost), **batch update the index every few minutes**.  
  * If critical (e.g., menu item sold out), **use incremental indexing**.

#### **1.3 Scaling the Ranking & Personalization System**

* **Real-Time Ranking Pipeline**  
  * Uses **Clickstream data** (e.g., search interactions, conversions).  
  * **Feature Store** (Feast) stores **ML-based ranking features**.  
* **ML-Based Re-Ranking with Vector Search**  
  * **Precompute restaurant embeddings** (e.g., based on similarity, trending score).  
  * Use **FAISS / Pinecone** for nearest-neighbor searches on embeddings.  
* **Distributed Key-Value Store for Personalization**  
  * **DynamoDB / Cassandra** stores **user preferences** and search history.

  ---

### **2. Failure Handling Strategies**

Ensuring high availability and resilience requires handling **service failures, network issues, and data inconsistencies**.

#### **2.1 API Layer Failure Handling**

| Failure Type | Mitigation Strategy |
| ----- | ----- |
| API Gateway Failure | Deploy multiple API gateways (AWS ALB, Nginx, Envoy) with **failover routing**. |
| Rate Limit Exceeded | Use **token bucket algorithm** to throttle excessive requests. |
| Query Timeout | Set a **hard timeout** (e.g., 200ms) for search queries. Return cached/stale results as fallback. |
| Elasticsearch Down | Route requests to **replicated clusters**, or fallback to **last known cache result**. |

#### **2.2 Data Consistency Failures**

| Failure Type | Mitigation Strategy |
| ----- | ----- |
| Inconsistent search results | Use **eventual consistency** via **Kafka stream updates**. |
| Stale menu data | Use **event-driven updates** + **TTL-based cache invalidation**. |
| Missing restaurant data | Perform **background re-indexing** of stale data every few hours. |

#### **2.3 Search Engine Failure Handling**

| Failure Type | Mitigation Strategy |
| ----- | ----- |
| Elasticsearch Node Failure | Auto-recovery using **Kubernetes self-healing**. |
| Query Performance Degradation | Use **query profiling** to detect slow queries and optimize indices. |
| Index Corruption | Keep **multiple backups of indices** and use **shadow indexing** for rollback. |

#### **2.4 Disaster Recovery Plan**

* **Multi-Region Deployment**  
  * Each region has an **active Elasticsearch cluster**.  
  * **Async replication** keeps data synced across regions.  
  * **Failover using DNS-based rerouting (Route53, Cloudflare)**.  
* **Snapshot Backups**  
  * Hourly **Elasticsearch snapshots** stored in **S3/GCS**.  
  * Restore from backup **within minutes** in case of catastrophic failure.

  ---

### **3. High Availability & Auto-Scaling**

#### **3.1 Search API Auto-Scaling**

* **Kubernetes Horizontal Pod Autoscaler (HPA)**  
  * Scale **API pods** dynamically based on **CPU/memory usage**.  
* **Elasticsearch Cluster Auto-Scaling**  
  * Scale **Elasticsearch nodes** based on **search query load**.  
  * Use **hot-warm-cold architecture** to optimize cost.

#### **3.2 Query Failover Strategy**

If **Elasticsearch cluster is down**, fall back to:

1. **Cached search results** (from Redis).  
2. **Backup search engine instance** (read-only).  
3. **Simplified keyword search** in **PostgreSQL** (lower fidelity but ensures partial results).  
   ---

### **4. Trade-Offs & Bottlenecks**

#### **4.1 Trade-Offs**

| Approach | Pros | Cons |
| ----- | ----- | ----- |
| Caching search results (Redis) | Reduces search latency | Can return **stale results** |
| Distributed indexing (Elasticsearch) | Fast search across millions of records | High infra cost |
| Precomputed rankings (ML-based) | Personalized and real-time | Requires model retraining |

#### **4.2 Bottlenecks**

| Bottleneck | Mitigation Strategy |
| ----- | ----- |
| High write load on search indices | **Batch updates** instead of real-time writes. |
| Hotspot queries in cache | **Sharded cache keys** by location/category. |
| High memory usage in Elasticsearch | **Index pruning** & **document compression**. |

