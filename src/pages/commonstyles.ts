import { defineCustomElements as deckDeckGoHighlightElement } from "@deckdeckgo/highlight-code/dist/loader";
deckDeckGoHighlightElement();

export const pageStyles = {
  color: "#232129",
  padding: 96,
  fontFamily: "-apple-system, Roboto, sans-serif, serif",
};

// styles
export const sectionStyles = {
  // color: "#232129",
  // fontFamily: "-apple-system, Roboto, sans-serif, serif",
  color: "rgba(61,61,78,var(--tw-text-opacity))",
  lineHeight: "1.7",
  outline: "none",
  fontFamily: "Droid Serif, Georgia, serif",
  fontSize: "18px",
  overflowWrap: "break-word",
  padding: "50px",
  "border-block-start": "solid 50px lightblue",
};
export const headingStyles = {
  marginTop: 0,
};
export const headingAccentStyles = {
  color: "#663399",
};
export const codeStyles = {
  color: "#8A6534",
  padding: 4,
  backgroundColor: "#FFF4DB",
  fontSize: "1.25rem",
  borderRadius: 4,
};
export const paragraphStyles = {
  marginBottom: 48,
};
export const listStyles = {
  marginBottom: 96,
  paddingLeft: 0,
};
export const listItemStyles = {
  fontWeight: 300,
  fontSize: 24,
  maxWidth: 560,
  marginBottom: 30,
};

export const linkStyle = {
  color: "#8954A8",
  fontWeight: "bold",
  fontSize: 16,
  verticalAlign: "5%",
};

export const docLinkStyle = {
  ...linkStyle,
  listStyleType: "none",
  marginBottom: 24,
};

export const descriptionStyle = {
  color: "#232129",
  fontSize: 14,
  marginTop: 10,
  marginBottom: 0,
  lineHeight: 1.25,
};

export const docLink = {
  text: "TypeScript Documentation",
  url: "https://www.gatsbyjs.com/docs/how-to/custom-configuration/typescript/",
  color: "#8954A8",
};

export const badgeStyle = {
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
};
