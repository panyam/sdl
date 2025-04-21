// components/Section.ts

import { Modal } from './Modal';

import { SectionType, DocumentSection, TextContent, DrawingContent, PlotContent, SectionData, SectionCallbacks } from './types'; // Import the types

/**
 * Section types
 */
// export type SectionType = 'text' | 'drawing' | 'plot';


/**
 * Manages an individual document section
 */
export class Section {
  private data: SectionData;
  private element: HTMLElement;
  private callbacks: SectionCallbacks;
  private modal: Modal;
  
  // DOM elements within the section
  private titleElement: HTMLElement | null;
  private contentElement: HTMLElement | null;
  private deleteButton: HTMLElement | null;
  private moveUpButton: HTMLElement | null;
  private moveDownButton: HTMLElement | null;
  private settingsButton: HTMLElement | null;
  private llmButton: HTMLElement | null;

  // Icon SVGs
  private static readonly TEXT_ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-full h-full"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25H12" /></svg>`;
  private static readonly DRAWING_ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-full h-full"><path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487zm0 0L19.5 7.125" /></svg>`;
  private static readonly PLOT_ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-full h-full"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 006 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118 16.5h-2.25m-7.5 0h7.5m-7.5 0l-1 3m8.5-3l1 3m0 0l.5 1.5m-.5-1.5h-9.5m0 0l-.5 1.5M9 11.25v1.5M12 9v3.75m3-6v6" /></svg>`;

  // Type-specific icon title attributes
  private static readonly TYPE_TITLES: Record<SectionType, string> = {
      text: "Text Section",
      drawing: "Drawing Section",
      plot: "Plot Section"
  }

  // DOM element for the type icon
  private typeIconElement: HTMLElement | null;

  
  /**
   * Create a new section
   */
  constructor(data: SectionData, element: HTMLElement, callbacks: SectionCallbacks = {}) {
    this.data = data;
    this.element = element;
    this.callbacks = callbacks;
    this.modal = Modal.getInstance();
    
    // Find DOM elements
    this.titleElement = element.querySelector('.section-title');
    this.contentElement = element.querySelector('.section-content');
    this.deleteButton = element.querySelector('.section-delete');
    this.moveUpButton = element.querySelector('.section-move-up');
    this.moveDownButton = element.querySelector('.section-move-down');
    this.settingsButton = element.querySelector('.section-settings');
    this.llmButton = element.querySelector('.section-ai');
    this.typeIconElement = element.querySelector('.section-type-icon');

    
    this.bindEvents();

    // Initialize the display title from the data
    if (this.titleElement) {
      this.titleElement.textContent = this.data.title;
    }

    // Initialize the type icon
    if (this.typeIconElement) {
        let iconSvg: string;
        switch (this.data.type) {
            case 'drawing': iconSvg = Section.DRAWING_ICON_SVG; break;
            case 'plot':    iconSvg = Section.PLOT_ICON_SVG; break;
            case 'text':
            default:        iconSvg = Section.TEXT_ICON_SVG; break;
        }
        this.typeIconElement.innerHTML = iconSvg;
        // Set title attribute for accessibility/tooltip
        this.typeIconElement.setAttribute('title', Section.TYPE_TITLES[this.data.type] || 'Section');
    }
    this.initializeContent();
  }
  
  /**
   * Bind event listeners for this section
   */
  private bindEvents(): void {
    // Title editing
    if (this.titleElement) {
      this.titleElement.addEventListener('click', this.startTitleEdit.bind(this));
    }
    
    // Delete button
    if (this.deleteButton) {
      this.deleteButton.addEventListener('click', () => {
        if (this.callbacks.onDelete) {
          this.callbacks.onDelete(this.data.id);
        }
      });
    }
    
    // Move up button
    if (this.moveUpButton) {
      this.moveUpButton.addEventListener('click', () => {
        if (this.callbacks.onMoveUp) {
          this.callbacks.onMoveUp(this.data.id);
        }
      });
    }
    
    // Move down button
    if (this.moveDownButton) {
      this.moveDownButton.addEventListener('click', () => {
        if (this.callbacks.onMoveDown) {
          this.callbacks.onMoveDown(this.data.id);
        }
      });
    }
    
    // Settings button
    if (this.settingsButton) {
      this.settingsButton.addEventListener('click', this.openSettings.bind(this));
    }
    
    // LLM button
    if (this.llmButton) {
      this.llmButton.addEventListener('click', this.openLlmDialog.bind(this));
    }
  }
  
  /**
   * Initialize content based on section type
   */
  private initializeContent(): void {
    if (!this.contentElement) return;
    
    // Clear current content
    const placeholder = this.contentElement.querySelector('.section-type-placeholder');
    if (placeholder) {
      placeholder.remove();
    }
    
    // Add type-specific content
    switch (this.data.type) {
      case 'text':
        this.initializeTextContent();
        break;
        
      case 'drawing':
        this.initializeDrawingContent();
        break;
        
      case 'plot':
        this.initializePlotContent();
        break;
    }
  }
  
  /**
   * Initialize content for text section type
   */
  private initializeTextContent(): void {
    if (!this.contentElement) return;
    
    // Create editable content area
     const contentDiv = document.createElement('div');
     contentDiv.className = 'min-h-[100px] border border-gray-200 dark:border-gray-700 rounded p-4 prose dark:prose-invert max-w-none';
     contentDiv.setAttribute('contenteditable', 'true');
     
     // Use loaded content if available (ensure it's treated as TextContent)
     const initialContent = (typeof this.data.content === 'string' && this.data.content.length > 0) ? this.data.content : '<p>Click to start editing...</p>';
     contentDiv.innerHTML = initialContent;
 
    // Add content change event
    contentDiv.addEventListener('blur', () => {
      const newContent = contentDiv.innerHTML;
      if (newContent !== this.data.content) {
        this.data.content = newContent;
        if (this.callbacks.onContentChange) {
          this.callbacks.onContentChange(this.data.id, newContent);
        }
      }
    });
    
    // Add to content element
    this.contentElement.appendChild(contentDiv);
  }
  
  /**
   * Initialize content for drawing section type
   */
  private initializeDrawingContent(): void {
    if (!this.contentElement) return;
    
    const content = this.data.content as DrawingContent; // Assume content matches type
    const drawingArea = document.createElement('div');
    drawingArea.className = 'min-h-[300px] border border-gray-200 dark:border-gray-700 rounded p-4 flex items-center justify-center';
    // Use loaded data if available, otherwise placeholder
    drawingArea.innerHTML = content?.data ? `<pre class="text-xs">${JSON.stringify(content.data, null, 2)}</pre>` : '<p class="text-gray-400 dark:text-gray-500">Drawing Area (Placeholder)</p>';
    this.contentElement.appendChild(drawingArea);
  }
  
  /**
   * Initialize content for plot section type
   */
  private initializePlotContent(): void {
    if (!this.contentElement) return;
    
    const content = this.data.content as PlotContent; // Assume content matches type
    const plotArea = document.createElement('div');
    plotArea.className = 'min-h-[300px] border border-gray-200 dark:border-gray-700 rounded p-4 flex items-center justify-center';
    // Use loaded data if available, otherwise placeholder
    plotArea.innerHTML = content?.data ? `<pre class="text-xs">${JSON.stringify(content.data, null, 2)}</pre>` : '<p class="text-gray-400 dark:text-gray-500">Plot Area (Placeholder)</p>';
    
    this.contentElement.appendChild(plotArea);
  }
  
  /**
   * Start title editing
   */
  private startTitleEdit(e: Event): void {
    e.preventDefault();
    e.stopPropagation();
    
    if (!this.titleElement) return;
    
    // Create an input element
    const input = document.createElement('input');
    input.type = 'text';
    input.className = 'px-2 py-1 w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 dark:text-white';
    input.value = this.data.title;
    
    // Replace title with input
    const originalText = this.titleElement.textContent;
    this.titleElement.textContent = '';
    this.titleElement.appendChild(input);
    
    // Focus input
    input.focus();
    input.select();
    
    // Handle input blur (save)
    const handleSave = () => {
      const newTitle = input.value.trim();
      
      // Only update if there's actual content
      if (newTitle) {
        this.titleElement!.textContent = newTitle;
        this.data.title = newTitle;
        
        if (this.callbacks.onTitleChange) {
          this.callbacks.onTitleChange(this.data.id, newTitle);
        }
      } else {
        // Restore original title if empty
        this.titleElement!.textContent = originalText;
      }
      
      // Remove events
      input.removeEventListener('blur', handleSave);
      input.removeEventListener('keydown', handleKeyDown);
    };
    
    // Handle keydown events
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        input.blur(); // Trigger save on Enter
      } else if (e.key === 'Escape') {
        e.preventDefault();
        this.titleElement!.textContent = originalText; // Restore original title
        
        // Remove events
        input.removeEventListener('blur', handleSave);
        input.removeEventListener('keydown', handleKeyDown);
      }
    };
    
    // Add event listeners
    input.addEventListener('blur', handleSave);
    input.addEventListener('keydown', handleKeyDown);
  }
  
  /**
   * Open section settings modal
   */
  private openSettings(): void {
    // This would show a settings modal when implemented
    console.log('Section settings clicked for', this.data.id);
    
    // Placeholder until settings modal is implemented
    alert(`Settings for section: ${this.data.title} (ID: ${this.data.id})`);
  }
  
  /**
   * Open LLM dialog for this section
   */
  private openLlmDialog(): void {
    this.modal.show('llm-dialog', {
      sectionId: this.data.id,
      sectionType: this.data.type,
      sectionTitle: this.data.title
    });
    
    // Update the current section display in the LLM dialog
    const currentSectionElement = document.getElementById('llm-current-section');
    if (currentSectionElement) {
      currentSectionElement.textContent = this.data.title;
    }
  }
  
  /**
   * Update the section number
   */
  public updateNumber(number: number): void {
    this.data.order = number;
    
    const numberElement = this.element.querySelector('.section-number');
    if (numberElement) {
      numberElement.textContent = `${number}.`;
    }
  }
  
  /**
   * Get the section data
   */
  public getData(): SectionData {
    return { ...this.data };
  }
  
  /**
   * Update the section data
   */
  public update(newData: Partial<SectionData>): void {
    // Update data object
    this.data = { ...this.data, ...newData };
    
    // Update title if needed
    if (newData.title && this.titleElement) {
      this.titleElement.textContent = newData.title;
    }
    
    // Update content if needed and different type
    if ((newData.type && newData.type !== this.data.type) || newData.content) {
      if (this.contentElement) {
        this.contentElement.innerHTML = '';
        this.initializeContent();
      }
    }
  }
  
  /**
   * Get the section data formatted for the document model.
   */
  public getDocumentData(): DocumentSection {
    const baseData = {
        id: this.data.id,
        title: this.data.title,
        order: this.data.order,
    };

    switch (this.data.type) {
        case 'text':
            // Ensure content is up-to-date if user didn't blur
            const textArea = this.contentElement?.querySelector('div[contenteditable="true"]'); // More specific selector if needed
            const currentContent = textArea ? textArea.innerHTML : this.data.content;
            return { ...baseData, type: 'text', content: currentContent as TextContent };
        case 'drawing':
            // Placeholder content for drawing
            const drawingContent: DrawingContent = { format: "placeholder_drawing", data: {} };
            return { ...baseData, type: 'drawing', content: drawingContent };
        case 'plot':
            // Placeholder content for plot
            const plotContent: PlotContent = { format: "placeholder_plot", data: {} };
            return { ...baseData, type: 'plot', content: plotContent };
    }
  }
}
