import { LeetCoachDocument, DocumentSection, DocumentMetadata, TextDocumentSection, DrawingDocumentSection } from './types'; // Import the types

export const EMPTY = {
  "metadata": {
    "id": "doc-placeholder-uuid",
    "schemaVersion": "1.0",
    "lastSavedAt": "2025-04-17T23:55:05.731Z"
  },
  "title": "Design Bitly",
  "sections": [
    {
      "id": "section-1",
      "title": "Functional Requirements",
      "order": 1,
      "type": "text",
      "content": "<p>Hello this is the functional requirements for bitly</p>"
    },
    {
      "id": "section-3",
      "title": "System Components",
      "order": 2,
      "type": "drawing",
      "content": {
        "format": "placeholder_drawing",
        "data": {}
      }
    },
    {
      "id": "section-2",
      "title": "Scalability Analysis",
      "order": 3,
      "type": "plot",
      "content": {
        "format": "placeholder_plot",
        "data": {}
      }
    }
  ]
}

// Define Sample Document Data
export const DOCUMENT: LeetCoachDocument = {
    metadata: {
        id: "sample-doc-123",
        schemaVersion: "1.0",
        lastSavedAt: new Date(Date.now() - 3600 * 1000).toISOString() // Simulate saved an hour ago
    },
     title: "Loaded System Design Doc",
     sections: [
         {
             id: "section-1",
             type: "text",
             title: "Initial Requirements",
             order: 1,
             content: "<p>These are the <strong>loaded</strong> requirements.</p><ul><li>Req 1</li><li>Req 2</li></ul>"
         } as TextDocumentSection, // Type assertion for clarity
         {
             id: "section-3", // Intentionally out of order ID to test sorting/order prop
             type: "drawing",
             title: "High-Level Architecture",
             order: 2,
             content: { format: "placeholder_drawing", data: { info: "Loaded drawing data placeholder"} }
         } as DrawingDocumentSection,
         {
             id: "section-2",
             type: "text",
             title: "Data Model Ideas",
             order: 3,
             content: "<p>Some initial thoughts on the data model.</p>"
         } as TextDocumentSection,
     ]
 };
