---
title: 'Design AdClickAggregator'
productName: 'AdClickAggregator'
date: 2025-02-08T11:29:10AM
tags: ['ad', 'click', 'aggregator', 'flink', 'streaming', 'batch processing', 'hard' ]
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: AdClickAggregator
---

{{# include "DrawingView.html" #}}

{{ .FrontMatter.summary }}

TBR - To be rewritten

## Requirements

### Functional Requirements

* Users can click on an ad and be redirected to the advertiser's website
* Advertisers can query ad click metrics over time with a minimum granularity of 1 minute


### Scale Requirements:

* 10M ads a time
* 10k clicks per add at peak

### Non functional requirements

* Scalable to support a peak of 10k clicks per second
* Low latency analytics queries for advertisers (sub-second response time)
* Fault tolerant and accurate data collection. We should not lose any click data.
* As realtime as possible. Advertisers should be able to query data as soon as possible after the click.
* Idempotent click tracking. We should not count the same click multiple times.


### Extensions (Out of scope)

* Ad targeting
* Ad serving
* Cross device tracking
* Integration with offline marketing channels

General:
* Authentication/User management
* Versioning, Scanning (malware, virsuses, csam etc)
* Storage limits, billing etc
* Analytics

## API/Interface/Entities

```
record Ad {
  Id string
  RedirectURL URL
  // Other Metadata
}

record ClickEvent {
  AdId string
  UserId string
  EventId string    // <---- each click is its own event
  ClickedAt Timestamp
}
```

```
service AggregatorService {
  // Return the File with ID, CreatedAt, CreatorId
  ServeAds()  Paginated<Ad> {
    GET: "/ads"
  }
  
  ProcessClick(adId: string) Redirect {
    GET: "/ads/{adId}"
  }
}
```

### High Level Design


{{ template "DrawingView" ( dict "caseStudyId" "AdClickAggregator" "id" "hld" ) }}

1. User is served ads (on feed etc)
2. User clicks on ad (user id may not be known - so some fingerprinting by device)
   etc).
3. Click event saved and processed - also user redirected (client vs server side - prefer server side)
4. Advertiser logs into see analytics

## How to Query Metrics?

1. V1 - Basic SQL on analytics store

```sql
SELECT 
  COUNT(*) as TotalClicks, 
  COUNT(DISTINCT UserId) as UniqueUsers 
FROM ClicksDB
WHERE AdId = <...>
  AND Timestamp BETWEEN TimeRange
  GROUP BY AdId
```

## Preventing abuse and double counting

* Add an impression id (kind of like csrf)
* Make it signed (eg jwt) so user cannot fake it

* Ad Placement Service generates a unique impression ID for each ad instance shown to the user.
* The impression ID is signed with a secret key and sent to the browser along with the ad.
* When the user clicks on the ad, the browser sends the impression ID along with the click data.
* The Click Processor verifies the signature of the impression ID.
* The Click Processor checks if the impression ID exists in a cache. If it does, then it's a duplicate, and we ignore it. If it doesn't, then we put the click in the stream and add the impression ID to the cache.

https://www.hellointerview.com/learn/system-design/problem-breakdowns/ad-click-aggregator#potential-deep-dives

## Deep Dives - How to scale

* Click Processor Service: We can easily scale this service horizontally by adding more instances. Most modern cloud providers like AWS, Azure, and GCP provide managed services that automatically scale services based on CPU or memory usage. We'll need a load balancer in front of the service to distribute the load across instances.
* Stream: Both Kafka and Kinesis are distributed and can handle a large number of events per second but need to be properly configured. Kinesis, for example, has a limit of 1MB/s or 1000 records/s per shard, so we'll need to add some sharding. Sharding by AdId is a natural choice, this way, the stream processor can read from multiple shards in parallel since they will be independent of each other (all events for a given AdId will be in the same shard).
* Stream Processor: The stream processor, like Flink, can also be scaled horizontally by adding more tasks or jobs. We'll have a seperate Flink job reading from each shard doing the aggregation for the AdIds in that shard.
* OLAP Database: The OLAP database can be scaled horizontally by adding more nodes. While we could shard by AdId, we may also consider sharding by AdvertiserId instead. In doing so, all the data for a given advertiser will be on the same node, making queries for that advertiser's ads faster. This is in anticipation of advertisers querying for all of their active ads in a single view. Of course, it's important to monitor the database and query performance to ensure that it's meeting the SLAs and adapting the sharding strategy as needed.

