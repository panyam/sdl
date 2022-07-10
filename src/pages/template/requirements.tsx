import * as React from "react"
import { VarNameDescNode as VND, VarNameNode as VN, ExprNode as Exp } from "../exprs"
import * as Styles from "../commonstyles"

// markup
const HLDSection = () => {
  return (
  <section style={Styles.sectionStyles}>
  <h2>Step 2 - Requirements</h2>
  Our goal is to create a modular component/system to solve this particular system.  Ideally what we want is this system can be "pluggable" else where by simply parametrizing what ever it is we are parametrizing so it can be used somewhere else.  Eg if it is an IDGenerator, we can parametrize it over say # IDs, QPS, Error tolerance etc and "instantiate" it in the host system.

  <h3>Functional Requirements</h3>

  Talk about function requirements here.  These are absolute core things we should solve in this design (it can actually be more than what one expects in an interview and we can leave it upto the interviewer to "prune" things for time management.
  <ul>
    <li>Core Req 1 </li>
    <li>Core Req 2 </li>
    <li>Core Req 3 </li>
    <li>...</li>
    <li>Core Req n </li>
  </ul>

  <h3>Extended Requirements</h3>

  Talk about extended requirements.  These are typically requirements that do not offer direct "feature" value or are really V2/V3 requirements.  Beyond V2/V3 we want to add "accounting/logistic" requirements here too.

  <ul>
    <li>Extended Func Req 1</li>
    <li>Extended Func Req 2</li>
    <li>...</li>
    <li>Extended Func Req N</li>
  </ul>

  Other extended requirements that are typically "ancilliary" - ie needed by company but may not be directly a user facing requirement:

  <ul>
    <li>Monetization (eg Ads) and/or Billing</li>
    <li>Analytics</li>
    <li>Abuse, Fraud, Spam detection/filtering (if applicable)</li>
    <li>Legal and Compliance</li>
    <li>Think of more and add here as a common template</li>
  </ul>

  <h3>Operational Requirements and SLOs</h3>
  Talk about the operational requirements that are about Availability, Scalabiilty and Reliability.  This may be a place to setup some metrics like P95 latencies and/or availability numbers (eg 5 nines).

  <ul>
    <li>Scalable, Highly Available</li>
    <li>Logging, Metrics, Monitoring and Alerts available</li>
    <li>Geographically Distributed (with Disaster Recovery enabled)</li>
    <li>Expandable (in some domain pertinent to the system being designed)</li>
  </ul>

  <h3>System Parameters</h3>

  The golden bit about our blog is we are able to parametrize and modularize a system so we can start composing systems!  With this we want some way of doing a system DSL.  Something like (very grpc like):

  <pre><code>{`
    system IDGen {
      @param(desc = "Number of IDs generated per second", default = 10000)
      CreationQPS: number

      @param(desc = "Size of ID entry (including optional 'extra' payloads) (KB)", default = 1)
      IDEntrySize: number
    }
  `}</code></pre>
  

  <br/>
  If we do not yet do a DSL for describing the system we want to create variables (and constants) here which form the "input parameters" of our system.  (if we want input parameters can be detected as those that do not depend on other parameters).  Some examples are:

  <Exp setto="SecondsPerDay" desc="Rough number of seconds per day" hidden="true">100000</Exp>
  <Exp setto="DaysPerYear" desc="Rough number of days per year" hidden="true">400</Exp>

  <ul>
    <li>
      <Exp setto="CreationQPS" desc="Number of IDs generation per second" hidden="true">10000</Exp>
      <VND var="CreationQPS"/> QPS
    </li>
    <li>
      <Exp setto="IDEntrySize" desc="Size of ID entry (including optional 'extra' payloads) (KB)" hidden="true">1</Exp>
      <VND var="IDEntrySize"/> KB
    </li>
    <li>
      <Exp setto="D" desc="Disk access time (in ms) for one IO on storage nodes (if needed)" hidden="true">1</Exp>
      <VND var="D"/> ms
    </li>
    <li>
      <Exp setto="RetentionPeriod" desc="Number of years to keep IDs around for" hidden="true">10</Exp>
      <VND var="RetentionPeriod"/>
    </li>

    <li>
      <Exp setto="P95CreationLatency" desc="P95 creation latency in ms" hidden="true">10</Exp>
      <VND var="P95CreationLatency"/>
    </li>

    <li>
      <Exp setto="CollissionRate" desc="Maximum collission rate when generation random IDs" hidden="true">0.01</Exp>
      <VND var="CollissionRate"/>
    </li>
  </ul>

  <b>Notes:</b>
  <ul>
      <li>Any notes we want can go here</li>
  </ul>
  </section>
  )
}

export default HLDSection
