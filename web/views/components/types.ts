// components/types.ts

export type SectionType = 'text' | 'drawing' | 'plot';

export interface DocumentMetadata {
  id: string; // UUID - Placeholder for now
  schemaVersion: string;
  // createdAt: string; // ISO 8601 timestamp - Omitted for simplicity for now
  lastSavedAt: string; // ISO 8601 timestamp
}

export type TextContent = string; // HTML content

// More specific structure for Excalidraw data
export interface ExcalidrawSceneData {
    elements: ReadonlyArray<any>; // Use ExcalidrawElement type if importing Excalidraw types
    appState: any; // Use ExcalidrawAppState type if importing
}


export interface DrawingContent {
  format: "excalidraw/json" | string; // Be specific for Excalidraw
  // Use the specific scene data type or fallback to object/string
  // Use object for Excalidraw as we store { elements, appState }
  data: ExcalidrawSceneData | object | string;
}

export interface PlotContent {
  format: string; // e.g., "chartjs_config", "plotly_json", etc.
  data: object; // The configuration or data object for the plot
}

export interface BaseDocumentSection {
  id: string;
  title: string;
  order: number;
}

export interface TextDocumentSection extends BaseDocumentSection {
  type: 'text';
  content: TextContent;
}

export interface DrawingDocumentSection extends BaseDocumentSection {
  type: 'drawing';
  content: DrawingContent; // Placeholder content for now
}

export interface PlotDocumentSection extends BaseDocumentSection {
  type: 'plot';
  content: PlotContent; // Placeholder content for now
}


export interface SectionData {
  id: string;
  designId: string; // <-- ADD THIS LINE
  type: SectionType;
  title: string;
  content: TextContent | DrawingContent | PlotContent;
  order: number;
}

export interface SectionCallbacks {
  onDelete?: (sectionId: string) => void;
  onMoveUp?: (sectionId: string) => void;
  onMoveDown?: (sectionId: string) => void;
  onTitleChange?: (sectionId: string, newTitle: string) => void;
  // Ensure content type matches SectionData['content']
  onContentChange?: (sectionId: string, newContent: SectionData['content']) => void;
  // Callback for requesting section addition
  onAddSectionRequest?: (relativeToId: string, position: 'before' | 'after') => void;
}

export type DocumentSection = TextDocumentSection | DrawingDocumentSection | PlotDocumentSection;

export interface LeetCoachDocument {
  metadata: DocumentMetadata;
  title: string;
  sections: DocumentSection[];
}
