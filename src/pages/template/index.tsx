import * as React from "react"
import IntroSection from "./intro"
import HLDSection from "./hld"
import EstimatesSection from "./estimates"
import RequirementsSection from "./requirements"
import ModellingSection from "./modelling"
import ScalabilitySection from "./scalability"
import * as Styles from "../commonstyles"

// markup
const IndexPage = () => {
  return (
    <main style={Styles.pageStyles}>
      <title>###System Title Here###</title>
      <IntroSection/>
      <RequirementsSection/>
      <ModellingSection/>
      <HLDSection/>
      <EstimatesSection/>
      <ScalabilitySection />
    </main>
  )
}

export default IndexPage
