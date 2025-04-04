---
title: 'Design DropBox'
productName: 'DropBox'
date: 2024-06-28T11:29:10AM
tags: ['dropbox', 'easy', 'dropbox']
draft: false
authors: ['Sri Panyam']
template: "CaseStudyPage.html/CaseStudyPage"
summary: Dropbox is a cloud-based file storage service that for users to store and share files in a secure and reliable waywhile accessing them from anywhere on any device.
---

{{# include "DrawingView.html" #}}

{{ .FrontMatter.summary }}

## Requirements

### Functional Requirements

The functional requirements are:

1.  Upload/Download files
2.  Syncfiles to devices
3.  Share files with friends and view/downloads downloaded files.

Scale Requirements:

* 1B users registered (DB has 700M I think)
* User uploads 10 files a month and shares with say 20 friends, Average size 100MB, P90 1GB
* Total files per month = 10B files ~ 250B files a year
* Storage = 10% at 1GB, rest averaging 100MB = 1B at 1GB and 9B at 100MB ~ 2B GB per month ~ 25B GB per year
* Sharing - 250B * 20 = 5Trillion shares a year

### Non functional requirements

* Highly available - Users should be able to upload and download their files (even if sharing is degraded).
* Large file support - 100GB
* Security/Reliability/Durability - Files should not be lost or corrupted
* Fast uploads/downloads/syncs

### Extensions (Out of scope)

* Authentication/User management
* Versioning, Scanning (malware, virsuses, csam etc)
* Storage limits, billing etc
* Analytics

## API/Interface/Entities

```
record user {
  Id string
}

record File {
  Id string
  Name string
  CreatorId UserId
  CreatedAt Timestamp
  UpdatedAt Timestamp
  ContentURI URI          // Where stored in S3
  DownloadURL URL         // URL to download the file from - this can keep changing and may even be generated on the fly
  ContentType string
  Status string
  SharedWith UserId[]     // Option1 if small # users eg < 100, otherwise see Option 2
}

// Option 2 for share
record FileShare {
  FileId string
  UserId string
  Permissions []string
}

record FileContents {
  FileId string
  ContentURI string
  ContentHash string
  Blob []byte
}
```

```
service FileService {
  // Return the File with ID, CreatedAt, CreatorId
  CreateFile(file: File) File {
    POST: "/files"
    BODY: {file}
  }
  
  UpdateFileShares(file: File) File {
    PATCH: "/files/{file.Id}"
    BODY: {file.SharedWith}
  }
  
  UploadFile(FileContents) {
    POST /files/{fileId}
  }
  
  DownloadFile(fileId string) FileContents {
    GET /files/{fileId}
  }
}
```

### High Level Design


{{ template "DrawingView" ( dict "caseStudyId" "DropBox" "id" "hld" ) }}


## Deep Dives

Here we will be asked what are the things to dive on and explain how our design addresses requirements.

1. How do we ensure each shortened URL is short and unique

### Option 1: Setup a global counter and increment transactionally on each call

### Option 2: Generate a random K char ID

```
1. set id = alias if one provided
2. if id == "": id = generateShortId()
3. entity := {Id: id, LongUrl: longUrl, Expiration....}
4. db.InsertIfNotExists(entity)
5. if failed: id = "" and goto step   // TODO - Provide # retries
6. return entity


// 6 alnum char gets us 2B IDs
generateShortId(KChars = 6) {
  out = ""
  for range(KChars) {
  out += randomAlphaOrDigit()
  }
  return out
}
```

Writing into this needs a write into the main index (by shortUrl - could be a hash index) followed by a write to a second index by longUrl (a btree index).
Given 10k iops on SSDs, writes could be 50-100ms for both indexes.   May or may not be transactional to tradeoff consistency with Availability.

2. How do you ensure redirects are fast

Main index is hash-index - reads are 1-10 ms.   To get it faster introduce a cache infront of the Database.

3. Further high scale

Imagine load from all over the world.   Here it would be useful to replicate our store to other regions and have traffic from users go to regions closest to them for reads.  Writes would still go to a single region for consistency (otherwise managing global master-master configs are more complicated and expensive).

