import * as React from "react"
import { ExprNode as Exp } from "../exprs"
import * as Styles from "../commonstyles"

// markup
const EstimatesSection = () => {
  return (
  <section style={Styles.sectionStyles}>
  <h2>Step 5 - Estimates and Parameters</h2>
    Before addressing scalability of a system, a system's limits, usage and capacity estimates are in order.   Our ID generation will be used as an independant service from a host system (eg TinyURL).   The host system may impose their own requirements.  Establishing this system's input and output parameters will inform the host system's performance characteristics.   Based on the sytem parameters, we can establish the following operational and performance requirements/parameters:

    <h3>Storage Estimates</h3>
    <p><span className = "param_title">Number of IDs created Daily, <span className="varname">IDsPerDay</span>:</span><br/>
      = CreationQPS * <Exp>SecondsPerDay</Exp>  (roughly 100k seconds per day)<br/>
      = <Exp setto="IDsPerDay" desc="Number of IDs created daily">CreationQPS  * SecondsPerDay</Exp> <br/>
    </p>

    <p><span className = "param_title">Number of IDs created Annually, IDsPerYear:</span><br/>
      = CreationQPS * 100,000 * <Exp>DaysPerYear</Exp> (roughly <Exp>DaysPerYear</Exp> days a year)<br/>
      = CreationQPS  * 4 * 10<sup>7</sup><br/>
      = <Exp setto="IDsPerYear" desc="Number of IDs created each year">IDsPerDay * DaysPerYear</Exp><br/>
    </p>

    <p><span className = "param_title">Annual Storage:</span><br/>
      = IDEntrySize * IDsPerYear<br/>
      = <Exp>IDEntrySize </Exp> Kb * <Exp>IDsPerYear</Exp><br/>
      = <Exp setto="AnnualStorage" desc="Storage required for IDs created in a year">IDEntrySize * IDsPerYear</Exp> KB<br/>
      = <Exp>AnnualStorage / (10 ** 9)</Exp> TB per year
    </p>

    <p><span className = "param_title">Total Storage over <Exp>RetentionPeriod</Exp> years:</span><br/>
      = RetentionPeriod * AnnualStorage<br/>
      = <Exp>RetentionPeriod</Exp> * <Exp>AnnualStorage / 10 ** 9</Exp><br/>
      = <Exp setto="TotalStorage" desc="Storage required for IDs created over the retention period">RetentionPeriod * AnnualStorage / 10 ** 9</Exp> TB<br/>
    </p>

    <h3>Bandwidth Estimates</h3>

    <p><span className = "param_title">Inbound Bandwidth:</span><br/>
      = CreationQPS * IDEntrySize<br/>
      = <Exp>CreationQPS</Exp> * <Exp>IDEntrySize</Exp> KBps<br/>
      = <Exp setto="InboundBandwidth" desc="Inbound network bandwidth into our system">CreationQPS * IDEntrySize</Exp> KBps<br/>
    </p>
  </section>
  )
}

export default EstimatesSection
