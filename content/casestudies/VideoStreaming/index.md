---
title: 'Design a Video Streaming'
productName: 'VideoStreaming'
date: 2025-02-06T11:29:10AM
tags: ['netflix', 'youtube', 'medium', 'file storage']
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: A video [streaming](streaming) service.  It allows members to shows and movies on an internet-connected device available on various platforms like IOS, Android, Web etc.
---

{{# include "DrawingView.html" #}}

{{ .FrontMatter.summary }}

## Requirements

### Functional Requirements

The functional requirements are:

1. Users can watch and share shows
2. Admin - Can upload/delete shows
3. Comment/Like videos

### Non functional requirements

* High scale and low-enough latency
* Reliable/Durable uploads
* Efficiency needs.
* Availability over consistency 

### Extensions (Out of scope)

* Authentication/User management
* Versioning, Scanning (malware, virsuses, csam etc)
* Storage limits, billing etc
* Analytics
* Searching of shows
* Geo blocking
* Pause/Resume (assume video always downloaded - but can get around - becomes chunking/DBox problem)

### Scale/Numbers

* Read heavy system
* Num users - 1B with 200M DAU
* 5 shows per user per day on average = 1B watched shows.
* Say 1:20 read:write ratio (in case of VideoStreaming it will be larger shows but with fewer uploads, with Youtube - smaller
  but more uploads - eg shorts etc) - ie 50M videos uploaded daily
* 

#### Storage Requirements:

* Youtube style - 1 Min of video is about 100MB of storage - so assuming average video is say 10Minutes  about 1GB per
  upload per user.  So 50M uploaders per day * 400 days per year * 1Gb per upload (100MB per min for 10 mins) =~ 20000 ExaBytes of video storage per year
* VideoStreaming style - Fewer - say 100 new shows a month - but each show being about say 1 hr long - could be 100 hours per
  month - 100 * 60 * 0.1GB = 600GB

## API/Interface/Entities

```
record user {
  Id string
}

record Video {
  Id string
  Name string
  CreatorId UserId
  CreatedAt Timestamp
  UpdatedAt Timestamp
  ContentURI URI          // Where stored in S3
  DownloadURL URL         // URL to download the file from - this can keep changing and may even be generated on the fly
  
  // Show info - may be searchable
  Metadata any
}

record Comments {
  VideoId string
  UserId string
  CreatedAt Timestamp
  Comment text
}
```

```
CRUD on Videos
Upload/Download on Video - Similar to Dropbox
```

### High Level Design


{{ template "DrawingView" ( dict "caseStudyId" "VideoStreaming" "id" "hld" ) }}

1. User "uploads" a file - the client is really calling a CreateFile API and then calling the UploadFile API.
2. On CreateFile, File entry is created and contentURI is also created (or we can create buckets with naming convention
   etc).
3. UploadFile fetches a signed/authed upload URL (ephemaral).
4. Client POSTs on this URL directly to the content store instead of going through DB
5. Content store on upload can send to CDNs for caching, notifies FileService to update metadata (like hashes,
   last udpated etc).  We can keep metadata seperately or as part of the content store itself.
6. Sharing etc straight forward.
   * Need File -> UserID (needs stronger consistency)
   * Secondary index - UserId -> File (ok to be eventual - "my files")

## Deep Dives - Chunking and Parallel Uploads

1. How do ensure resumability? ie dont want to file 90% of the way in only to start all over again.  Browsers may not
   even allow it (not to mention VPNs, Firewalls etc). Not to mention this is too serial!


* Instead of a file - chunk into say 10-20MB chunks (depending on file size - 1GB = 50 chunks)
* Upload by chunk - eg `POST files/{fileId}/chunks/{chunkId}` - each chunk would return its own upload URL

```
record File {
  ...
  NumChunks int
}

record FileChunk {
  FileId string
  ChunkSize int
  ChunkIndex int
  ChunkURI  URI   // uri in the content store
  CreatedAt Timestamp
  LastUpdated Timestamp
  status string   // Uploading, Done  etc
}
```

{{ template "DrawingView" ( dict "caseStudyId" "VideoStreaming" "id" "final" ) }}
