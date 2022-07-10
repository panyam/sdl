import * as React from "react"
import { CodeNode as Code, ExprNode as Exp } from "../exprs"
import * as Styles from "../commonstyles"

// markup
const HLDSection = () => {
  return (
  <section style={Styles.sectionStyles}>
  <h2>Step 4 - High Level Design</h2>

  Absolutely need a diagram here - Need a way to describe diagrams.

  <h3>Validating Functionality</h3>
  Once we show a HLD above, we want to talk about how each of the methods in our service is functionally possible in this architecture (without any scaling). Something liek:

  <h4>Service Method 1</h4>
  <Code>{`
    db = getDBClient()    # A client to our DB

    def Method1(method_1_params) -> Method1ReturnType:
      # Pseudo or Python code here...
  `}</Code>

  <h4>Service Method 2</h4>
  <Code>{`
    db = getDBClient()    # A client to our DB

    def Method2(method_2_params) -> Method2ReturnType:
      # Pseudo or Python code here...
  `}</Code>

  <h4>IDService.Release</h4>
  <Code>{`
    db = getDBClient()    # A client to our DB
    def Release(id: ID) -> IDEntry:
      entry = db.getID(id)
      if entry is not None:
        db.Delete(entry)
  `}</Code>
  
  <h3>Ensuring consistent writes</h3>
  Not sure if this section is required - our pseudo code above itself should contain the right locking primitives?  Or should we in every design we do go into locking.  I think pointing to a blog post about locks is good enough.
  </section>
  )
}

export default HLDSection
