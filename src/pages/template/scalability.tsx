
import * as React from "react"
import { ExprNode as Exp } from "../exprs"
import * as Styles from "../commonstyles"

// markup
const ScalabilitySection = () => {
  return (
  <section style={Styles.sectionStyles}>
  <h2>Step 6 - Scalability Barriers</h2>
  With our system parameters and estimates above we can now identify scalability barriers and bottlenecks for each of our API methods.  This can drive the design choices we can make towards a scalable and reliable system (to fit without our operational requirements).

  <p>Systems need to be scaled either horizontally (adding more nodes) and/or vertically (making nodes beefier) if the available nodes are constrained in one or more dimensions.  Let us look at the dimensions in turn</p>

  <h3>Bandwidth constraints</h3>
  <p>
  Our system's network (inbound and outbound) bandwidth is dictated by the CreationQPS at <Exp>InboundBandwidth</Exp> KBps.
  Since most servers (even commodity ones) are equipped with 100MBps or 1GBps network interface cards (NICs), handling our inbound bandwidth does not impose any serious bottlenecks with respect to network bandwidth.
  </p>

  <h3>Storage constraints</h3>
  Persistent disks from your <a href="https://aws.amazon.com/ebs/features/">popular</a> <a href="https://cloud.google.com/compute/docs/disks#:~:text=Each%20persistent%20disk%20can%20be,to%20create%20large%20logical%20volumes.">cloud</a> <a href="https://docs.microsoft.com/en-us/azure/virtual-machines/disks-types#ultra-disk-size">providers</a> allow disks from 1GB to 64 TB.

  Even with a disk with the maximu size, we would need at least
  <Exp>ceil(TotalStorage / 64)</Exp> disks.

  <h3>Storage IO Constraints</h3>
  <h3>Compute Constraints</h3>
  </section>
  )
}

export default ScalabilitySection
