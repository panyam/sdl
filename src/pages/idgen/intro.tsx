import * as React from "react"
import * as Styles from "../commonstyles"

// markup
const IntroSection = () => {
  return (
  <section style={Styles.sectionStyles}>
  <h2 style={Styles.headingStyles}>Introduction</h2>
  ID Generation is a common concern for representing and identifying objects stored a database. IDs can be generated in several ways.  Each way has its advantages and disadvantages.

  <h2>Existing Schemes</h2>

  <h3>System Generated UUID</h3>
  <strong>Pros:</strong>
  <ul>
    <li>Easy to generate. Almost every framework/platform comes with a UUID implementation that can offer 128 bit unique keys.</li>
    <li>Random</li>
  </ul>

  <strong>Cons:</strong>
  <ul>
    <li>UUIDs are 128 bits in length and thus fixed.  This can be too large or too short for the use case at hand.</li>
    <li>When used in a distributed environment would require a central server (to coin UUIDs) which could become a bottleneck.</li>
    <li>Wont cut it if sequential IDs are required.</li>
  </ul>

  <h3>Using a Database Auto-increment sequence</h3>
  <strong>Pros:</strong>
  <ul>
    <li>Linear and auto-incrementing.</li>
    <li>Almost all databases have this feature and can be centralized.</li>
  </ul>

  <strong>Cons:</strong>
  <ul>
    <li>Central DB can become a bottleneck.</li>
    <li>No randomness if this is necessary.</li>
    <li>Cannot reserve string IDs if needed.</li>
  </ul>

  <h3>Distributed (Almost) sequential IDs (eg Twitter Snowflake)</h3>

  <strong>Pros:</strong>
  <ul>
    <li>Scalable and distributed</li>
    <li>“Almost” linear and monotonic.</li>
  </ul>

  <strong>Cons:</strong>
  <ul>
    <li>Not random enough.</li>
    <li>Cannot “reserve” IDs.</li>
    <li>Only numeric – cannot choose custom string IDs.</li>
  </ul>
    </section>
  )
}

export default IntroSection
