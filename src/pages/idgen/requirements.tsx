import * as React from "react"
import { VarNameDescNode as VND, VarNameNode as VN, ExprNode as Exp } from "../exprs"
import * as Styles from "../commonstyles"

// markup
const RequirementsSection = () => {
  return (
  <section style={Styles.sectionStyles}>
  <h2>Step 2 - Requirements</h2>
  Our goal is to create a modular component/system that can be used by any other system that needs ID generation as a seperate service.  The proposed ID generation service is a variation of the UUID generation.  Here it is distributed (while avoiding duplicates) and allows variable ID sizes for different use cases.  Some uses are in systems like <a href = "/pages/tinyurl/">TinyURL</a> and <a href = "/pages/pastebin/">PasteBin</a>.

  <h3>Functional Requirements</h3>
  <ul>
    <li>Generate a random K bit ID.  Why "bits" and not characters?  This is because the bits themselves can be encoded depending on the alphabets chosen by the application (eg base64, base32 etc).  Decoupled responsibilities makes our system cleaner.</li>
    <li>*Reserve* a custom K bit ID explicitly (eg as vanity IDs).</li>
    <li>Delete/Release a previously generated ID.</li>
    <li>Associate a payload (of "reasonable size") with the ID for reverse tracking (this is beginning to look like a simple key value store).</li>
  </ul>

  <h3>Extended Requirements</h3>
  <ul>
    <li>Expiration of IDs after a certain amount of time.</li>
    <li>Bulk Creations/Reservations</li>
  </ul>

  <h3>System Parameters</h3>
  Like any other component, our ID Generation service will have parameters that can be fine-tuned to allow certain tradeoffs between efficiency, cost and consistencies.

  <br/>
  The following parameters can control the cost and performance of our system (If the parameters are unclear, do not worry.  Read on and the semantics and impact of these parameters will become clearer):

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
    <li><Exp>CreationQPS</Exp> Qps for ID generation is *a lot*. For services that typically need random string IDs the traffic is much lower. eg for Bitly, the typical traffic is 600M bitlinks a month, resulting in 200QPS!!!   To put it in perspective, <Exp>CreationQPS</Exp> QPS would be needed if 1B Meta users shared 10 links each for URL shortening every day!
  </li>
    <li>With any QPS, we would not reach “full” storage requirements of the 10 year mark instantaneously. </li>
  </ul>


  If you are keen to understand how K should be calculated <a href="">click here</a>.

  <h3>Operational Requirements (SLOs)</h3>
  <ul>
    <li>IDs that are generated must be persisted since they can have (useful) metadata associated with them like creation timestamps, owner and some metadata.  A way to relax (parametrize) this constrained will be described in some of the extensions in the end.</li>
    <li>URL Creations within <Exp>P95CreationLatency</Exp>ms in 95th percentile</li>
    <li>Scalable, Highly Available</li>
    <li>Monitoring and Alerts available</li>
    <li>Geographically Distributed</li>
    <li>Expandable (in ID space)</li>
  </ul>
  </section>
  )
}

export default RequirementsSection;
