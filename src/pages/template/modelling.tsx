import * as React from "react"
import { CodeNode as Code, ExprNode as Exp } from "../exprs"
import * as Styles from "../commonstyles"

// markup
const ModellingSection = () => {
  return (<section style={Styles.sectionStyles}>
  <h2>Step 3 - APIs and Schemas</h2>

  <h3>Services and Entities</h3>

  Describe the service and entities here.  This should probably be its own DSL but for now we can just put it in code nodes.

  <Code>{`
    type Alias1 = RealType1;
    type Alias2 = RealType2;
    ...
    type AliasN = RealTypeN;

    service OurService {
        /**
         * Describe method 1.
         * @param param1    Describe param 1
         * @param param2    Describe param 2
         * ...
         * @param paramN    Describe param N
         *
         * @return  Describe return type
         */
        Method1(params) -> ReturnType

        /**
         * Describe method 2.
         *
         * @param param1    Describe param 1
         * @param param2    Describe param 2
         * ...
         * @param paramN    Describe param N
         *
         * @return  Describe return type
         */
        Method2(params) -> ReturnType

        /**
         * Describe method N.
         *
         * @param param1    Describe param 1
         * @param param2    Describe param 2
         * ...
         * @param paramN    Describe param N
         *
         * @return  Describe return type
         */
        MethodN(params) -> ReturnType
    }

    record RequestType1 {
      // Describe this
      field1: FieldType1;

      // Describe this
      field2: FieldType2;
    }

    // Describe any specific entity types
    record Entity1 {
      // Describe this
      field1: FieldType1;

      // Describe this
      field2: FieldType2;
    }
  `}</Code>

  <h3>Database Schemas</h3>
  Ideally what we want is to be able to generate DB schemas automatically so we dont have to type it here.  We want our Service and Entity definitions to be in a meta language.  Eg in our System.design "file" we could have have annotations that describe how the service or entity description looks like in different targets.

  But perhaps we can skip this section all together or just make it manual for now so we dont have to skip any SQL table schemas.

  </section>
  )
}

export default ModellingSection
