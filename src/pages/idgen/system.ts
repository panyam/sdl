export const DB = {
  name: "DB",
  params: {
    maxRows: {
      desc: "Maximum storage capacity in the DB.",
    },
  },
  // UI related params - eg
  ui: {
    shape: "roundedrect",
  },
};

export const API = {
  name: "API",
  // A list of other components or Systems this component needs
  // "needs" is different from "contains".  needs relationships
  // refer to refernces where as "contains" relationships refer
  // to a system that hosts other components
  // For now this refers to components by the class names of concrete
  // types.  Later we can make this structural type based so that
  // references can be changed and swapped out (to demonstrate tight
  // contracts).
  needs: ["DB as db"],
  params: {},
  methods: {
    CreateID: [
      `Times(
          RetryCount,
          db.Get({ 404: db.Put({ 200: Ret(200) }) })
      )`,
      ".Ret(400)",
    ],
  },
};

export const IDGen = {
  name: "IDGen",
  params: {},
  contains: ["db: {Get, Put}", "api: {CreateID}"],
  methods: {
    // Simply proxy the Create to the API layer's create call
    Create: "api.CreateID",
  },
  // used to connect components here - only available for systems
  connections: ["api.DB = DB"],
};

// How can we use the above?
// We have a few options (expressed in an imaginary syntax:
//
//  sys = new IDGen()
//  sys.db = new DB()
//  sys.api = new API()
//  sys.connect();
//
// This would create the IDGen system allowing the user to set the DB and
// API components.
//
//
// However sometimes we want the system provider to have their own
// default implementation so we can do something like:
//
// sys = new IDGen()
//
// That's all that should be needed for the default behaviour.
//
// How about sending traffic to this?  We want "generators".
//
// eg
// cl1 = new Client()
// cl2 = new Client()
// cl3 = new Client()
//
// We have created 3 clients.  How do we represent the connection to the
// IDGen system?
//
// Eg:
//
// cl1.connect(sys, "Create")
// cl2.connect(sys, "Create")
// cl3.connect(sys, "Create")
//
// Alternatively we may want to take a "World centric" view.  This will
// allow us a couple of benefits:
//
// 1. A source does not need to keep track of lists of neighbors etc.
// 2. Components can be more easily wired - is this actually a benefit?
//
// Here we need the concept of a World object.  All objects in this world
// are one thing!
//
// Next thign we want to do is "render"  this.  ie visuall we want to
// see a system on the screen
//
// Here is where a World is useful as we can easily identify "root" objects to
// identify what a toplogical ordering of the nodes looks like.  Otherwise
// we would have to do:
//
// scene.render(cl1, cl2, cl3, sys)
//
// "sys" above is not required as it will be discovered since we are
// starting from the clients anyway
//
// A few helpers are needed to help the rendered:
//
// 1. A toplogical ordering of nodes (one heuristic is to layout all component
// "aligned on the same line" for items that topologically "equivalent" -
// eg cl1, cl2, cl3).
// 2. Provide node specific UI options (eg which images to use, where
// connectors are, what fonts to use etc).  This could also have defaults
// within the nodes itself.
// 3. Connector specific UI options.  This could be a little tricker.
// This could be specified during the "connect" call above, eg:
//
//          cl1.connect(sys, "Create", {options})
//
// This works for programmatic connections.  Alternatively at a node def
// can put this part of the connections.  At the node level we only do
// component level connection (eg api.DB = DB).  However it is even more
// interesting to be able to show which method in api calls which method
// in DB.   This is somethign we can get when doing a trace, however
//
// Formal definition of SystemView interface:
//
// SystemView {
//    /**
//     * Renders the entire system which is the closure of all the given components.
//     */
//    render(...components: Component[])
//
//    /**
//     * Highlight a particular end-to-end trace of an API call
//     * in a given path-style.
//     */
//    showCall(src, api, path-style)
//
//    /**
//     * Hide a particular call trace starting from a source.
//     */
//    hideCall(src, api)
// }
//
//
