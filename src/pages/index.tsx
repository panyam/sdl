import * as React from "react"

// styles
const pageStyles = {
  color: "#232129",
  padding: 96,
  fontFamily: "-apple-system, Roboto, sans-serif, serif",
}
const headingStyles = {
  marginTop: 0,
  marginBottom: 64,
}
const headingAccentStyles = {
  color: "#663399",
}
const paragraphStyles = {
  marginBottom: 48,
}
const codeStyles = {
  color: "#8A6534",
  padding: 4,
  backgroundColor: "#FFF4DB",
  fontSize: "1.25rem",
  borderRadius: 4,
}
const listStyles = {
  marginBottom: 96,
  paddingLeft: 0,
}
const listItemStyles = {
  fontWeight: 300,
  fontSize: 24,
  maxWidth: 560,
  marginBottom: 30,
}

const linkStyle = {
  color: "#8954A8",
  fontWeight: "bold",
  fontSize: 16,
  verticalAlign: "5%",
}

const docLinkStyle = {
  ...linkStyle,
  listStyleType: "none",
  marginBottom: 24,
}

const descriptionStyle = {
  color: "#232129",
  fontSize: 14,
  marginTop: 10,
  marginBottom: 0,
  lineHeight: 1.25,
}

const badgeStyle = {
  color: "#fff",
  backgroundColor: "#088413",
  border: "1px solid #088413",
  fontSize: 11,
  fontWeight: "bold",
  letterSpacing: 1,
  borderRadius: 4,
  padding: "4px 6px",
  display: "inline-block",
  position: "relative" as "relative",
  top: -2,
  marginLeft: 10,
  lineHeight: 1,
}

// data
const basicComponents = [
  {
    text: "Template",
    draft: true,
    badge: false,
    url: "/template/",
    description:
      "A simple template for creating our system design entries",
    color: "#E95800",
  }
  ,{
    text: "Indexes",
    badge: true,
    url: "/indexes/",
    description:
      "Key indexes that can be used as components in datastores.  Some examples include Hash Indexes, BTree Indexes, Logfiles, SSTables.",
    color: "#E95800",
  }
  ,{
    text: "ID Generator",
    badge: true,
    url: "/idgen/",
    description:
      "A scalable and tunable ID Generator great for generating random string IDs of customizable lengths.",
    color: "#E95800",
  }
  ,{
    text: "Transaction Managers",
    url: "/transactions/",
    description:
      "Understanding of Transactions and systems to perform/enact transactions over distributed components in a pluggable and modular way.",
    color: "#E95800",
  }
  ,{
    text: "Persistent Connection Managers",
    url: "/pcms/",
    description:
      "Clients can use persistent connections to a system to receive data (events or alerts) without a request.  Persistent Connection Managers are a building block in enabling this.",
    color: "#E95800",
  }
  ,{
    text: "ZooKeeper",
    url: "/zookeeper/",
    description:
      "How is ZooKeeper designed?",
    color: "#E95800",
  }
  ,{
    text: "Distributed Counters",
    url: "/distcounts/",
    description:
      "Counting items at scale in a distributed systems have interesting challenges.   Learn about them here.",
    color: "#E95800",
  }
  ,{
    text: "CRDT",
    url: "/crdts/",
    description:
      "https://hal.inria.fr/file/index/docid/555588/filename/techreport.pdf",
    color: "#E95800",
  }
]

const miscSystems = [
  {
    text: "Airplane Booking System",
    url: "/airplanebooking/",
    description:
      "Design an Airplane ticket booking system in a globally available and scalable way.",
    color: "#E95800",
  }
  ,{
    text: "Event/Ticket Booking System (Ticketmaster)",
    url: "/ticketmaster/",
    description:
      "Design a ticket booking system for events and venues (similar to an Airplane breservation system).",
    color: "#E95800",
  }
  ,{
    text: "Web crawlers",
    url: "/webcrawlers/",
    description:
      "Design a webcrawler and planet scale",
    color: "#E95800",
  }
  ,{
    text: "Rate Limiters",
    url: "/ratelimiters/",
    description:
      "Design a ratelimiter for handing several scenarios",
    color: "#E95800",
  }
  ,{
    text: "Rate Limiters",
    url: "/ratelimiters/",
    description:
      "Design a ratelimiter for handing several scenarios",
    color: "#E95800",
  }
  ,{
    text: "AD Platforms",
    url: "/adplatforms/",
    description:
      "Design a two sided market place for ads with consumers and campain managers",
    color: "#E95800",
  }
]

const searchSystems = [
  {
    text: "Full Text Search",
    url: "/ftsearch/",
    description:
      "Explore and understand how a full-text search platform can be built",
    color: "#E95800",
  }
  ,{
    text: "FB Post Search",
    url: "/fbpostsearch/",
    description:
      "Design a system to power Facebook's post search",
    color: "#E95800",
  }
  ,{
    text: "Typeahead Search",
    url: "/typeahead/",
    description:
      "Design a Typeahead search systems for power auto-complete use cases",
    color: "#E95800",
  }
  ,{
    text: "Relevance/Ranking Systems",
    url: "/relandranking/",
    description:
      "Design systems to apply relevance and ranking to a users's feed (or other lists) in a more dynamic way possibly even applied at serve time",
    color: "#E95800",
  }
];

const mlServices = [
  {
    text: "Analytic Systems (Pinot, Druid etc)",
    url: "/analytics/",
    description:
      "Design a general purpose analytics system that processes events and allows slicing/dicing on several dimensions for the users",
    color: "#E95800",
  }
  ,{
    text: "Experimentation Platforms",
    url: "/experimentation/",
    description:
      "Experimentation allows a product developer to both control who a feature is presented to as well as understand the impact of these features on the overall business metrics",
    color: "#E95800",
  }
  ,{
    text: "ML Training Platform",
    url: "/mltraining/",
    description:
      "Build a large scale ML Training platform",
    color: "#E95800",
  }
  ,{
    text: "ML Model Serving Platform",
    url: "/mlserving/",
    description:
      "Once ML models are trained they need to be served at scale.",
    color: "#E95800",
  }
]

const gamingServices = [
  {
    text: "Gaming services",
    url: "/gaming/",
    description:
      "Understand the landscape around what systems are specific to different kinds of games or which tradeoffs in other (discussed) systems specifically apply to games",
    color: "#E95800",
  }
]

const ecommerceSystems = [
  {
    text: "Order Delivery Service (Doordash, Instacart, UberEats etc)",
    url: "/orderdeliversvc/",
    description:
      "Design a system for initiating and tracking orders for pickup and delivery used in several scenarios (eg Doordash, Instacart, UberEats etc)",
    color: "#E95800",
  }
  ,{
    text: "Place Booking (AirBnb)",
    url: "/placebooking/",
    description:
      "Design a system for booking places at particular times (eg AirBnb)",
    color: "#E95800",
  }
  ,{
    text: "Supply/Demand Systems (or 2 sided market places?)",
    url: "/supplyanddemand/",
    description:
      "TBD",
    color: "#E95800",
  }
]

const cloudInfra = [
  {
    text: "Deployment Service",
    url: "/deployments/",
    description:
      "Design a platform for deploying a release/binary/service from compiled artifacts to systems globally (or even galacticaly) at scale.",
    color: "#E95800",
  }
  ,{
    text: "Metrics and Monitoring",
    url: "/monitoring/",
    description:
      "Design a system that can collect metrics from a system and different levels (CPU, Host, Process etc) and enable analysis of this for different types of users (Developer, SREs, Business etc)",
    color: "#E95800",
  }
  ,{
    text: "Cluster Manager",
    url: "/clustermanager/",
    description:
      "Design a system for provisioning VMs and other resources on the cloud (eg MIGs or AutoScaling groups)",
    color: "#E95800",
  }
  ,{
    text: "Reactive Autoscaler",
    url: "/reactiveautoscaler/",
    description:
      "Design a system for listening to systems via monitoring/metrics and horizontally scaling a cluster by adding/removing machines to match capacity",
    color: "#E95800",
  }
  ,{
    text: "Migration Strategies",
    url: "/migrations/",
    description:
      "Migrations are hard - Explore strategies, costs and consistency tradeoffs (eg dual reads, dual writes etc)",
    color: "#E95800",
  }
  ,{
    text: "High Availability Strategies",
    url: "/highavailability/",
    description:
      "Eg - HA with Regional PDs - https://cloud.google.com/compute/docs/disks/high-availability-regional-persistent-disk",
    color: "#E95800",
  }
  ,{
    text: "Disaster Recovery Systems",
    url: "/drstory/",
    description:
      "Design a system that can enable backups and restore of systems with variable RTO/RPO targets",
    color: "#E95800",
  }
  ,{
    text: "Application Control Plane",
    url: "/acp/",
    description:
      "Control planes power modern application development and deployment.  They ensure the right resources are provisioned, correctly setup for monitoring health and even enable self healing of a cluster.  Here we design parts of such a system",
    color: "#E95800",
  }
  ,{
    text: "Design a serverless engine (Cloud Functions, AWS Lambda)",
    url: "/lambda/",
    description:
      "Building a serverless platform like Cloud Functions or AWS Lambda with other primitive cloud infrastructure components so far",
    color: "#E95800",
  }
]

const locationServices = [
  {
    text: "Driver Location Service",
    url: "/driverlocsvc/",
    description:
      "Design a system for tracking the location of millions of mobile drivers world-wide in an (almost) realtime manner.  These systems can power Uber (for showing driver locations), Deliver systems (eg Fedex, Amazon for showing package locations etc).",
    color: "#E95800",
  }
  ,{
    text: "Static Place Services (Yelp, Google Maps)",
    url: "/placessvc/",
    description:
      "Design a services where (static) objects/entities on a map can be represented and searched for by users (eg finding places near me).",
    color: "#E95800",
  }
  ,{
    text: "Routing Services (Yelp, Google Maps)",
    url: "/routingsvc/",
    description:
      "Design a service for showing routes between any two locations on a map for helpgin drivers navigate to their destinations.",
    color: "#E95800",
  }
]

const socialNetworks = [
  {
    text: "Twitter",
    url: "/twitter/",
    description:
      "Exploring the design of Twitter based on some of the components built so far",
    color: "#E95800",
  }
  ,{
    text: "Instagram",
    url: "/instagram/",
    description:
      "Design Instagram",
    color: "#E95800",
  }
  ,{
    text: "Facebook Feed",
    url: "/fbfeed/",
    description:
      "Explore and understand the design of the Facebook feed with different post types",
    color: "#E95800",
  }
  ,{
    text: "TikTok Feed",
    url: "/tiktop/",
    description:
      "Design TikTok",
    color: "#E95800",
  }
  ,{
    text: "FB Messenger",
    url: "/fbmessenger/",
    description:
      "Design Facebook Messenger",
    color: "#E95800",
  }
  ,{
    text: "Whatsapp",
    url: "/whatsapp/",
    description:
      "Design Whatsapp",
    color: "#E95800",
  }
]

const contentSystems = [
  {
    text: "Blobstore",
    url: "/blobstore/",
    description:
      "Design a blobstore for storing blobs of arbitrary (but reasonable) sizes used to power systems like Pastebin",
    color: "#E95800",
  }
  ,{
    text: "Filestore",
    url: "/filestore/",
    description:
      "Design a filestore - Unlike blobstores - which are wroteonly, would we expect filestores to provide mutability in random order by multiple people - like Dropbox and Google Docs?",
    color: "#E95800",
  }
  ,{
    text: "Pastebin",
    url: "/pastebin/",
    description:
      "Design a pastebin service where users can share snippets of data such as code samples or log files with limited visibility",
    color: "#E95800",
  }
  ,{
    text: "Photo Sharing",
    url: "/photosharing/",
    description:
      "Design a photo sharing services where users can upload photos to be shared by different audiences (private, friends, public) consumable from a variety of media globally.",
    color: "#E95800",
  }
  ,{
    text: "Video Sharing (Youtube, Vimeo)",
    url: "/videosharing/",
    description:
      "Design a video sharing services where users can upload movies to be shared by different audiences (private, friends, public) consumable from a variety of media globally.",
    color: "#E95800",
  }
  ,{
    text: "Collaborative File Editing (Google Docs, Jamboard, Sheets etc)",
    url: "/videosharing/",
    description:
      "Design a collaborating editing system for a group of users to edit reasonably sized files (order of 10s of MBs)",
    color: "#E95800",
  }
]

// markup
const IndexPage = () => {
  return (
    <main style={pageStyles}>
      <title>Home Page</title>
      <h1 style={headingStyles}>
        Welcome to LeetCoach
        <br />
        - Modular and Quantitative System Design
      </h1>
      <h2>Basic Components</h2>
      <ul style={listStyles}> {basicComponents.map(createSystemTile)} </ul>

      <h2>Cloud Infrastructure</h2>
      <ul style={listStyles}> {cloudInfra.map(createSystemTile)} </ul>

      <h2>Social Networks/Systems</h2>
      <ul style={listStyles}> {socialNetworks.map(createSystemTile)} </ul>

      <h2>Miscellaneous Systems</h2>
      <ul style={listStyles}> {miscSystems.map(createSystemTile)} </ul>

      <h2>Content Sharing Systems</h2>
      <ul style={listStyles}> {contentSystems.map(createSystemTile)} </ul>

      <h2>eCommerce Systems</h2>
      <ul style={listStyles}> {ecommerceSystems.map(createSystemTile)} </ul>

      <h2>Location based Services</h2>
      <ul style={listStyles}> {locationServices.map(createSystemTile)} </ul>

      <h2>Gaming Services</h2>
      <ul style={listStyles}> {gamingServices.map(createSystemTile)} </ul>

      <h2>Search Systems</h2>
      <ul style={listStyles}> {searchSystems.map(createSystemTile)} </ul>

      <h2>Analytic/ML Services</h2>
      <ul style={listStyles}> {mlServices.map(createSystemTile)} </ul>
    </main>
  )
}

function createSystemTile(link: any) {
return (<li key={link.url} style={{ ...listItemStyles, color: link.color }}>
            <span>
              <a
                style={linkStyle}
                href={`${link.url}?utm_source=starter&utm_medium=start-page&utm_campaign=minimal-starter-ts`}
              >
                {link.text}
              </a>
              {link.badge && (
                <span style={badgeStyle} aria-label="New Badge">
                  NEW!
                </span>
              )}
              <p style={descriptionStyle}>{link.description}</p>
            </span>
          </li>
        )
}

export default IndexPage
