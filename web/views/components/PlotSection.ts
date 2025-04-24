// components/PlotSection.ts

import { BaseSection } from './BaseSection';
import { PlotContent, SectionData, SectionCallbacks } from './types'; // Or move interfaces

export class PlotSection extends BaseSection {

    // Placeholder for plot library instance or config data
    private plotConfig: object = {};
    private plotCanvas: HTMLCanvasElement | null = null; // If using canvas

    constructor(data: SectionData, element: HTMLElement, callbacks: SectionCallbacks = {}) {
        super(data, element, callbacks);
         // Ensure content is initialized as an object if not present
         if (typeof this.data.content !== 'object' || this.data.content === null) {
            this.data.content = { format: 'placeholder_plot', data: {} };
        }
        this.enableFullscreen();
    }

    protected populateViewContent(): void {
        const previewContainer = this.contentContainer?.querySelector('.plot-preview-container');
        if (previewContainer) {
             const content = this.data.content as PlotContent;
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
            const content = this.data.content as PlotContent;
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
