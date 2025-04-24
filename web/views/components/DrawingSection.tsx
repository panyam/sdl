// components/DrawingSection.ts

import { ContentApi } from './Api'; // Import API client
import { BaseSection } from './BaseSection';
import { SectionData, SectionCallbacks, DrawingContent, ExcalidrawSceneData } from './types';

// --- Excalidraw Integration ---
import React from 'react';
import ReactDOM from 'react-dom/client'; // Use client for React 18+
import {
  serializeAsJSON,
  loadFromBlob,
  exportToBlob,
  exportToSvg,
  convertToExcalidrawElements
} from "@excalidraw/excalidraw";
import {
  Excalidraw, MainMenu, Footer,
} from "@excalidraw/excalidraw";
// Optional: Import types if needed for stricter typing
// import { ExcalidrawElement, ExcalidrawImperativeAPI } from '@excalidraw/excalidraw/types/element/types';
// import { AppState } from '@excalidraw/excalidraw/types/types';

// Type alias for Excalidraw's API handle
type ExcalidrawApi = any; // Use ExcalidrawImperativeAPI if types are imported

export class DrawingSection extends BaseSection {
    protected static readonly ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-full h-full"><path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487zm0 0L19.5 7.125" /></svg>`;

    // Placeholder for drawing library instance
    // Store React root and Excalidraw API instance
    private reactRoot: ReactDOM.Root | null = null;
    private excalidrawAPI: ExcalidrawApi | null = null;
    private editorContainerElement: HTMLDivElement | null = null; // Store container reference
    private drawingContent: DrawingContent | null;
    private lightPreview = "";
    private darkPreview = "";

    constructor(data: SectionData, element: HTMLElement, callbacks: SectionCallbacks = {}) {
        super(data, element, callbacks);
        // Ensure content is initialized as an object if not present
        this.drawingContent = { format: 'excalidraw/json', data: { elements: [], appState: {} } };
        this.enableFullscreen();
    }

    /** Returns the (type) title for this section. */
    protected getSectionTypeTitle(): string {
      return "Drawing Section"
    }


    /** Returns the svg content to show for this section. */
    protected getSectionIconSvg(): string {
      return DrawingSection.ICON_SVG;
    }

    protected populateViewContent(): void {
        const isDarkMode = document.documentElement.classList.contains('dark');
        const previewContainer = this.contentContainer?.querySelector('.drawing-preview-container');
        const svg = isDarkMode ? this.darkPreview : this.lightPreview
        if (previewContainer) {
            const content = this.drawingContent
            // **Placeholder:** Render based on content.format and content.data
            // TODO: Implement a static SVG export/render for view mode later
            if (content && content.format === 'excalidraw/json' && (content.data as ExcalidrawSceneData)?.elements?.length > 0) {
                if (svg.length == 0) {
                 previewContainer.innerHTML = `<pre class="text-xs text-gray-600 dark:text-gray-400">${JSON.stringify(content.data, null, 2)}</pre>`;
                } else {
                 previewContainer.innerHTML = svg
                }
            } else {
                 previewContainer.innerHTML = `<p class="text-gray-500 dark:text-gray-400 italic">No drawing data. Click 'Edit' to start.</p>`;
            }
        } else {
             console.warn(`View content area not found for drawing section ${this.data.id}`);
        }
    }

    protected populateEditContent(): void {
        this.editorContainerElement = this.contentContainer?.querySelector('.drawing-editor-container') as HTMLDivElement;
        if (this.editorContainerElement instanceof HTMLElement) {
            this.editorContainerElement.innerHTML = ''; // Clear placeholder
            console.log(`Initializing Excalidraw in section ${this.data.id}`);

            // Ensure React root doesn't already exist (safety check)
            if (this.reactRoot) {
                console.warn(`DrawingSection ${this.data.id}: Attempting to populate edit content, but React root already exists. Unmounting first.`);
                this.reactRoot.unmount();
                this.reactRoot = null;
                this.excalidrawAPI = null;
            }

            this.reactRoot = ReactDOM.createRoot(this.editorContainerElement);
            const initialContent = this.drawingContent
            let initialElements: any[] = [];
            let initialAppState: any = {};

            if (true || (initialContent?.format === 'excalidraw/json' && typeof initialContent?.data === 'object')) {
                 // Ensure data conforms to ExcalidrawSceneData structure
                 const sceneData = initialContent!.data as ExcalidrawSceneData;
                 // initialElements = convertToExcalidrawElements(sceneData?.elements || []); // Use Excalidraw's conversion
                 initialElements = convertToExcalidrawElements([...(sceneData?.elements || [])]);
                 initialAppState = sceneData?.appState || {};
                 initialAppState.collaborators = []
            }

            // Determine theme
            const isDarkMode = document.documentElement.classList.contains('dark');

            this.reactRoot?.render(
              <Excalidraw 
                    theme = { isDarkMode ? 'dark' : 'light' }
                    excalidrawAPI ={ (api: ExcalidrawApi) => { this.excalidrawAPI = api; } }
                    onChange = { this.handleExcalidrawChange.bind(this) } // Debounced save or flag dirty state
                    initialData = { {
                        elements: initialElements,
                        appState: initialAppState,
                    } }
              >
                <MainMenu>
                  <MainMenu.DefaultItems.LoadScene />
                  <MainMenu.DefaultItems.SaveAsImage />
                  <MainMenu.DefaultItems.Export />
                  <MainMenu.DefaultItems.ToggleTheme />
                  <MainMenu.DefaultItems.ClearCanvas />
                  <MainMenu.DefaultItems.ChangeCanvasBackground/>
                  <MainMenu.DefaultItems.Help/>
                  </MainMenu>
                  </Excalidraw>
            );
        } else {
             console.warn(`Edit content area not found for drawing section ${this.data.id}`);
        }
    }

    /** Handles changes from Excalidraw - could be used for auto-save or marking dirty */
    private handleExcalidrawChange(elements: ReadonlyArray<any>, appState: any): void {
        // Store the latest data internally. Debounce saving or mark as dirty here.
        this.drawingContent = { format: 'excalidraw/json', data: {elements, appState} };
        // console.log(`Excalidraw changed in section ${this.data.id}. Elements: ${elements.length}`);
        // Example: Trigger debounced save after inactivity
        // this.debouncedSave();
    }

    protected getContentFromEditMode(): DrawingContent {
        console.log(`Getting data from Excalidraw instance for section ${this.data.id}`);

        let drawingData: ExcalidrawSceneData = { elements: [], appState: {} };

        // Try to get data from the stored state updated by onChange
        if (this.drawingContent) {
            drawingData = this.drawingContent.data as ExcalidrawSceneData;
        }
        // Fallback: Try to get directly from API if onChange didn't fire recently (less ideal)
        else if (this.excalidrawAPI) {
            drawingData = {
                 elements: this.excalidrawAPI.getSceneElements(),
                 appState: this.excalidrawAPI.getAppState(),
            };
            console.warn(`Getting Excalidraw data directly from API for ${this.data.id}. Should ideally use onChange state.`);
        } else {
             console.error(`Cannot get Excalidraw data: API instance not found for section ${this.data.id}. Returning empty.`);
        }

        // Return in the expected DrawingContent format
        return {
            format: 'excalidraw/json', // Or the actual format used by your lib
            data: drawingData
        };
    }
 
    // Override switchToViewMode to handle React unmounting
    public override switchToViewMode(saveChanges: boolean): void {
        if (this.mode === 'edit') {
            console.log(`DrawingSection ${this.data.id}: Unmounting React component on switch to view.`);
            if (this.reactRoot) {
                try {
                    this.reactRoot.unmount();
                } catch (error) {
                    console.error(`Error unmounting React root for section ${this.data.id}:`, error);
                } finally {
                    this.reactRoot = null;
                    this.excalidrawAPI = null;
                    this.editorContainerElement = null; // Clear container ref
                }
            }
        }
        // Call the base class method *after* cleanup
        super.switchToViewMode(saveChanges);
    }

    protected override async refreshContentFromServer() {
        this.drawingContent = { format: 'excalidraw/json', data: { elements: [], appState: {} } };
        try {
            const resp = await ContentApi.contentServiceGetContent({
              designId: this.designId,
              sectionId: this.sectionId,
              name: "main",
            })
            if (resp.contentBytes) {
              const json = atob(resp.contentBytes)
              if (json.trim().length > 0) {
                this.drawingContent = JSON.parse(json)
              }
            }
            // Now load dark and light svgs
            const respLight = await ContentApi.contentServiceGetContent({
              designId: this.designId,
              sectionId: this.sectionId,
              name: "light.svg",
            })
            if (respLight.contentBytes) this.lightPreview = atob(respLight.contentBytes)
            const respDark = await ContentApi.contentServiceGetContent({
              designId: this.designId,
              sectionId: this.sectionId,
              name: "dark.svg",
            })
            if (respDark.contentBytes) this.darkPreview = atob(respDark.contentBytes)
        } catch (err: any) {
          console.error("error loading: ", err)
        }
    }

    public async handleSaveClick(): Promise<void> {
        this.drawingContent = this.getContentFromEditMode()
        const sceneData = this.drawingContent!.data as ExcalidrawSceneData;
        const asSvgDark = await exportToSvg({
          elements: sceneData.elements,
          appState: {
            exportBackground: true,
            exportWithDarkMode: true,
          },
        } as any)
        this.darkPreview = asSvgDark.outerHTML

        const resp1 = await ContentApi.contentServiceSetContent({
          designId: this.designId,
          sectionId: this.sectionId,
          name: "dark.svg",
          contentBytes: btoa(this.darkPreview),
        })

        const asSvgLight = await exportToSvg({
          elements: sceneData.elements,
          appState: {
            exportBackground: true,
            exportWithDarkMode: false,
          },
        } as any)
        this.lightPreview = asSvgLight.outerHTML

        const resp2 = await ContentApi.contentServiceSetContent({
          designId: this.designId,
          sectionId: this.sectionId,
          name: "light.svg",
          contentBytes: btoa(this.lightPreview)
        })

        console.log(`Save button clicked or shortcut used for section ${this.data.id}.`);
        const resp = await ContentApi.contentServiceSetContent({
          designId: this.designId,
          sectionId: this.sectionId,
          name: "main",
          contentBytes: btoa(JSON.stringify(this.drawingContent)),
        })
        this.switchToViewMode(true);
    }

    /**
     * Handles theme changes specifically for the DrawingSection.
     * Re-renders the Excalidraw component if it's in edit mode.
     */
    public override handleThemeChange(): void {
         console.log(`DrawingSection ${this.data.id}: Handling theme change.`);
         if (this.mode !== 'edit' || !this.reactRoot || !this.editorContainerElement) {
             // Only need to act if the editor is currently active
             return;
         }

         console.log(`DrawingSection ${this.data.id}: Re-rendering Excalidraw for theme change.`);

         // Get current content before unmounting/re-rendering
         this.drawingContent = this.getContentFromEditMode();

         // Unmount the existing instance cleanly
         this.reactRoot.unmount();
         this.reactRoot = null;
         this.excalidrawAPI = null;

         // Re-populate content, which will re-mount Excalidraw with the correct theme
         this.populateEditContent();
     }

    /** Implement the abstract method from BaseSection */
    protected resizeContentForFullscreen(isEntering: boolean): void {
        console.log(`DrawingSection ${this.data.id}: Resizing content for fullscreen=${isEntering}. Excalidraw should resize automatically within container.`);
        // Excalidraw usually adapts to its container size.
        // Example: if (this.drawingEditorInstance && typeof this.drawingEditorInstance.resize === 'function') {
        //     this.drawingEditorInstance.resize(); // Call the library's specific resize/redraw method
        // }
    }
}
