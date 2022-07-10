import * as React from "react"
import * as Styles from "../commonstyles"

// markup
const IntroSection = () => {
  return (
  <section style={Styles.sectionStyles}>
  <h2 style={Styles.headingStyles}>Introduction</h2>
  Introduction goes here.  Some things to add:

  <ul>
    <li>What is the business problem (eg for a single company or as it applies to several companies - eg Doordash or Delivery as it applies to Doordash, Instacart etc).</li>
    <li>Why the business exists and what value it adds to the users.</li>
    <li>What are the challenges?</li>
    <li>What is the system being tackled here for the specific challenge.</li>
  </ul>

  <h2>Existing Schemes</h2>

  Talk about any existing schemes if any
    </section>
  )
}

export default IntroSection
