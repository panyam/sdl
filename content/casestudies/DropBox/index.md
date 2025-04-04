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

{{ template "DrawingView" ( dict "caseStudyId" "DropBox" "id" "final" ) }}
