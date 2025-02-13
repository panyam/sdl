---
author: Sriram Panyam
title: (Almost) Incremental ID Generator
date: 2020-01-22
description: An ID generator module based on Twitter's Snowflake ID generator that generates (almost) incremental 64 bit IDs.
math: true
draft: true
---

Incremental IDs in a system have a nice property.   They are a pseudo indicator of the age of an entity (the ID is associated with).  There are two ways to obtain incremental IDs:

1. *Using a Database Auto-increment sequence*

Most databases can mint auto incrementing sequences for (usually) primary key fields.  At least most relational databases are equipped with this ability.  This ensures that the IDs that are generated are truly incremental.

The advantage of this scheme is the ubiquity of the feature (among database technologies) as well as guarantee of linearity.  However this suffers from the disadvantage that the single centralized database will quickly become a bottleneck.

2. *Distributed (Almost) sequential IDs*

Twitter in its [Snowflake](https://blog.twitter.com/engineering/en_us/a/2010/announcing-snowflake) paper proposed an ID generation algorithm that was horizontally scalable with guarantees of "almost" sequential 64 bit IDs.  This did however come at a price - namely increased complexity.  We shall go over this scheme in this post.

