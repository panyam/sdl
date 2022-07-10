import * as React from "react"
import { CodeNode as Code, ExprNode as Exp } from "../exprs"
import * as Styles from "../commonstyles"
import { Diagram, SimpleIDG } from "./diags";

// markup
const HLDSection = () => {
  return (
  <section style={Styles.sectionStyles}>
  <h2>Step 4 - High Level Design</h2>

  <Diagram src={SimpleIDG()}></Diagram>
  <Code>{`
    component API(db: DB) {
      Create(req) -> IDEntry {
        RetryCount * {
          db.GetID {
            400 => db.Put() {
              200 => return 200
            }
          }
        }
        return 400    // Too many collissions due to retries
      }
    }

    component DB {
      GetID() {
        //
        // Busy and return takes a dist and performs the following:
        // 1. Samples the entry from the distribution
        // 2. Applies a latency of the value from the sampled value
        // 3. return the code from the sampled value
        BusyAndReturn(Dist {
          90 => (404, 10:ms),
          10 => (200, 10:ms),
          10 => (200, 100:ms),
        })
      }

      PutID() {
        // Doing Dists as 2D instead of tuples
        code,dist <- Dist {
          90 => (200, {90 => 10:ms, 10 => 100:ms)    // No duplicates found
          10 => (400, 10:ms)    // Found a duplicate
        }
        latency <- dist
        Busy(latency)
        return code
      }
    }
  `}
  </Code>
Absolutely need a diagram here

  <h3>Validating Functionality</h3>
  How does our architecture ensure our functional requirements are met?  

  <h4>IDService.Create</h4>
  <Code>{`
    db = getDBClient()    # A client to our DB

    def Create(req: CreateRequest) -> IDEntry:
      if req.customID is not None:
        # See if custom ID already exists
        entry = db.getID(req.customID)
        if entry is not None:
          error "Custom id already exists"
        else:
          return db.put(IDEntry{
            id: req.customID,
            creator: req.creator
          })
      else:
        while i < retryCount:   # Allow few retries
          id = generateKBitRandomID(K)
          # See if it already exists
          entry = db.getID(req.customID)
          if entry is None:
            return db.put(IDEntry{
              id: id,
              creator: req.creator
            })
  `}</Code>

  The python-esque pseudo code is fairly self explanatory:

  <ul>
    <li>First we generate a random K bit ID (if a custom id is not specified). This could be as simple as calling a random number generator for a random value between 0 and 2<sup>K</sup></li>
    <li>We check if this ID already exists.</li>
    <li>If the ID does not exist it is saved into the DB</li>
    <li>If the ID exists the process retried (upto a limit)</li>
  </ul>

  <h4>IDService.Release</h4>
  <Code>{`
    db = getDBClient()    # A client to our DB
    def Release(id: ID) -> IDEntry:
      entry = db.getID(id)
      if entry is not None:
        db.Delete(entry)
  `}</Code>
  
  <h3>Ensuring consistent writes</h3>
  There is a small cause for concern in the Create method.   When the entry is written to the DB, how can we ensure that an entry that was missing (with the get call) is not created by another concurrent call to this method with the same ID (eg the same custom ID was submitted)?   Locking (and specifically read-modify-write patterns) are described in detail here and apply commonly to several other systems too.  So we will not go into detail here.  Rest assured we can assume that locking primitives are available in the datastores we pick.
  </section>
  )
}

export default HLDSection
