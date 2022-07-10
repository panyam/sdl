import * as React from "react"
import { CodeNode as Code, ExprNode as Exp } from "../exprs"
import * as Styles from "../commonstyles"

// markup
const ModellingSection = () => {
  return (<section style={Styles.sectionStyles}>
  <h2>Step 3 - APIs and Schemas</h2>

  <h3>Services and Entities</h3>

  Our TinyURL service exposes the following methods:

  <Code>{`
    type ID = bytes[32]

    service IDService {
      /**
       * Create/Reserve an ID entry.
       */
      Create(CreateRequest) -> IDEntry;

      /**
       * Releases an ID back into the pool (if it exists).
       */
      Release(ID) -> void;
    }

    record CreateRequest {
      // Creator of this ID for tracking ownership
      creator: UserID;

      // An optional custom ID to explicitly reserve
      customID: string?
    }

    // ID Entries that are saved (if we expect persisted IDs)
    record IDEntry {
      // Approx 32 bytes - can store
      id: ID

      // "entity" that created this ID
      creator: string

      // Creation time stamp of the TinyURL
      // Will be set by service
      createdAt: Timestamp

      // custom data user may want to store
      customData: bytes[512]
    }
  `}</Code>

  <h3>Database Schemas</h3>
  <p>
  Our schema definition would depend on the choice of databases.  However we can generally consider our schema definition if a SQL or NO-SQL datastore were to be used.
  </p>

  <p>
  Typical NoSQL datastores are schema-less and can store data close to a json-like hierarchical format.  So if a NoSQL datastore is used the above IDEntry is very identical to its JSON counter part.
  </p>

  <p> We can also specify a schema if we decide to store our ID entries in a SQL DB: </p>

  <Code>{`
      CREATE TABLE IDEntry (
        id          varchar(32)     PRIMARY KEY,
        creator_id  varchar(16)
        created_at  DateTime        NOT NULL,
        custom_data: varbinary[512]
      )
  `}</Code>

  </section>
  )
}

export default ModellingSection
