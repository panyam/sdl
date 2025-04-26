// components/PlotSection.ts

import { BaseSection } from './BaseSection';
import { PlotContent, SectionData, SectionCallbacks } from './types'; // Or move interfaces

export class PlotSection extends BaseSection {
    protected static readonly ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-full h-full"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 006 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118 16.5h-2.25m-7.5 0h7.5m-7.5 0l-1 3m8.5-3l1 3m0 0l.5 1.5m-.5-1.5h-9.5m0 0l-.5 1.5M9 11.25v1.5M12 9v3.75m3-6v6" /></svg>`;

    // Placeholder for plot library instance or config data
    private plotConfig: object = {};
    private plotCanvas: HTMLCanvasElement | null = null; // If using canvas
    private plotContent: PlotContent;

    constructor(data: SectionData, element: HTMLElement, callbacks: SectionCallbacks = {}) {
        super(data, element, callbacks);
         // Ensure content is initialized as an object if not present
         if (typeof this.plotContent !== 'object' || this.plotContent === null) {
            this.plotContent = { format: 'placeholder_plot', data: {} };
        }
        this.allowEditOnClick = false;
        this.enableFullscreen();
    }

    /** Returns the (type) title for this section. */
    protected getSectionTypeTitle(): string {
      return "Plot Section"
    }


    /** Returns the svg content to show for this section. */
    protected getSectionIconSvg(): string {
      return PlotSection.ICON_SVG;
    }

    protected populateViewContent(): void {
        const previewContainer = this.contentContainer?.querySelector('.plot-preview-container');
        if (previewContainer) {
             const content = this.plotContent as PlotContent;
             previewContainer.innerHTML = ''; // Clear placeholder/previous plot

             console.log(`Placeholder: Render plot in section ${this.data.id}`);
              // **Placeholder:** Render the plot using content.data and a library (e.g., Chart.js, Plotly)
             // Example:
             // this.plotCanvas = document.createElement('canvas');
             // previewContainer.appendChild(this.plotCanvas);
             // new Chart(this.plotCanvas.getContext('2d'), content.data);
             if (content && Object.keys(content.data).length > 0) {
                 previewContainer.innerHTML = `<pre class="text-xs text-gray-600 dark:text-gray-400">${JSON.stringify(content.data, null, 2)}</pre>`;
            } else {
                 previewContainer.innerHTML = `<p class="text-gray-500 dark:text-gray-400 italic">No plot data. Click 'Edit' to configure.</p>`;
            }
        } else {
             console.warn(`View content area not found for plot section ${this.data.id}`);
        }
    }

    protected populateEditContent(): void {
        const editorContainer = this.contentContainer?.querySelector('.plot-editor-container');
        if (editorContainer instanceof HTMLElement) {
            const content = this.plotContent as PlotContent;
            this.plotConfig = content.data || {}; // Store current config

            // **Placeholder:** Create form fields or a JSON editor (like CodeMirror or Monaco)
            // to edit this.plotConfig based on content.format
            editorContainer.innerHTML = `
                <label for="plot-json-${this.data.id}" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Plot Configuration (JSON):</label>
                <textarea id="plot-json-${this.data.id}" rows="8" class="plot-config-textarea block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-white sm:text-sm">${JSON.stringify(this.plotConfig, null, 2)}</textarea>
            `;
            console.log(`Placeholder: Initialize plot config editor in section ${this.data.id}`);
        } else {
             console.warn(`Edit content area not found for plot section ${this.data.id}`);
        }
    }

    protected getContentFromEditMode(): PlotContent {
        console.log(`Placeholder: Get data from plot config editor for section ${this.data.id}`);
        let plotData: object = {};
        const textarea = this.contentContainer?.querySelector('.plot-config-textarea') as HTMLTextAreaElement | null;

        if (textarea) {
             try {
                 plotData = JSON.parse(textarea.value);
             } catch (e) {
                 console.error(`Invalid JSON in plot config for section ${this.data.id}:`, e);
                 // Optionally show an error to the user
                 alert("Error parsing plot configuration JSON. Please check the format.");
                 // Return the last known valid config or default
                 plotData = this.plotConfig;
             }
        } else {
             // Fallback if textarea not found (shouldn't happen)
             plotData = this.plotConfig;
        }


        // Return in the expected PlotContent format
        return {
            format: 'placeholder_plot', // Or the actual format used by your plot lib
            data: plotData
        };
    }

    /** Implement the abstract method from BaseSection */
    protected resizeContentForFullscreen(isEntering: boolean): void {
       // This method is crucial for plotting libraries that need explicit resize calls.
       console.log(`PlotSection ${this.data.id}: Resizing content for fullscreen=${isEntering}. Triggering plot library resize.`);

       // --- Placeholder for actual plot library integration ---
       // Example (Chart.js): if (this.plotInstance) { this.plotInstance.resize(); }
       // Example (Plotly): if (this.plotContainerElement) { Plotly.Plots.resize(this.plotContainerElement); }
    }
}
