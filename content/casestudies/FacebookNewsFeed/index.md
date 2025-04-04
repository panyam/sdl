---
title: 'Design Facebook News Feed'
productName: 'FacebookNewsFeed'
date: 2025-02-13T11:29:10AM
tags: ['news feed', 'facebook', 'meta', 'posts', 'fan-out', 'medium' ]
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: FacebookNewsFeed
---

{{# include "DrawingView.html" #}}

{{ .FrontMatter.summary }}

TBR - To be rewritten

## Requirements

### Functional Requirements

* Users should be able to create posts.
* Users should be able to friend/follow people.
* Users should be able to view a feed of posts from people they follow, in chronological order.
* Users should be able to page through their feed.


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

* The system should be highly available (prioritizing availability over consistency). Tolerate up to 2 minutes for eventual consistency.
* Posting and viewing the feed should be fast, returning in < 500ms.
* The system should be able to handle a massive number of users (2B).
* Users should be able to follow an unlimited number of users, users should be able to be followed by an unlimited number of users.
* "Unlimited" followers (say in 100s of millions at P99, 10k at P50)

### Extensions (Out of scope)

* Users should be able to like and comment on posts.
* Posts can be private or have restricted visibility.

General:
* Authentication/User management
* Versioning, Scanning (malware, virsuses, csam etc)
* Storage limits, billing etc
* Analytics

## API/Interface/Entities


```
record Paginated<X> {
  Items []X
  Count int
  PageToken string
  NextPageToken string
}

record User {
  Id string
}

// Pkey = UserId + ":" + FollowerId
// G - Secondary index on UserId (for all my followers)
// On FollowerId - who I am following
record Follower {
  UserId string
  FollowerId string
}

record Post {
  Id string   /// Pkey - sharded
  CreatorId string    // Global Secondary index - CreatorId and CreatedAt (for getting all "my" posts by creation time)
  CreatedAt Time
  Content string
  PostType string
}
```

```
service FeedService {
  // Return the File with ID, CreatedAt, CreatorId
  CreatePost(post Post) Post {    // validates and sets post.Id, CreatedAt and CreatorId 
    POST /posts
    BODY: {post}
  }
  
  GetFeed()  Paginated<Post> {
    GET: "/feed"
  }
  
  // Listing a user's posts (not feed)
  ListPosts(userId)  Paginated<Post> {
    GET /users/{userId}/posts
  }
}

service FollowService {
  // Similar one for unfollow too
  Follow(userId string) Follow {    // logged in user follows a given user
    POST /users/{userId}/follow
  }
}
```

### High Level Design


{{ template "DrawingView" ( dict "caseStudyId" "FacebookNewsFeed" "id" "hld" ) }}

1. User is served ads (on feed etc)
2. User clicks on ad (user id may not be known - so some fingerprinting by device)
   etc).
3. Click event saved and processed - also user redirected (client vs server side - prefer server side)
4. Advertiser logs into see analytics

## How the feed works

```
PeopleIFollow = Select from Followers where followerId = <me>

Select from Posts WHERE UserId in PeopleIFollow and ORDER BY CreatedAt DESC
```

## Deep Dives - Scale on Post Reads

* Can be uneven (hot posts vs not)
* Edits infrequent
* Cache posts with TTL (LRU)
* Sharded/Distributed cache

## Deep Dives - Scale on Feed Reads

* Instead of querying each time materialize the feed into a FeedDB

```
record Feed {
  FollowerId UserId   // feed for a follower - can be Pkey
  PostId Time
  Posts [](PostId, Timestamp)      // room for 60k
}
```

* Update this on a post creation.   

```
def onNewPost(post):
  for followerId in self.followers:
    feed = getFeed(followerId)
    feed.Posts.append(post.id)    // trim and save etc
    feed.save()
```

For 10B posts a day - this is 10B * 1k followers updates = 10T updates on Feed DB (100M QPS)

* Can shard (by follower id)
* Can batch - leverage common followers across posts and aggregate

{{ template "DrawingView" ( dict "caseStudyId" "FacebookNewsFeed" "id" "feeddb" ) }}

## Deep Dives - Handling FanOut

Updating a million followers (on even more frequent one is very slow)

1. Have a cache for "PostsByPerson 
2. Only update followers if your follow count is < X
3. On GetFeed - get your own feed + those who are influencers by fanout and merge

{{ template "DrawingView" ( dict "caseStudyId" "FacebookNewsFeed" "id" "final" ) }}
