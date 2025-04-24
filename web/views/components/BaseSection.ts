// components/BaseSection.ts

import { Modal } from './Modal';
import { SectionData, SectionType, DocumentSection, TextContent, DrawingContent, PlotContent, SectionCallbacks } from './types';
import { DesignApi } from './Api'; // Import API client
import { V1Section } from './apiclient'; // Import API Section type
import { extractContentFromApiSection } from './converters'; // Import the converter

/**
 * Abstract base class for all document sections.
 * Handles common functionality like title editing, controls (move, delete, LLM),
 * and switching between view and edit modes using templates from the registry.
 */
export abstract class BaseSection {
    protected data: SectionData;
    protected element: HTMLElement;
    protected callbacks: SectionCallbacks;
    protected modal: Modal;
    protected mode: 'view' | 'edit' = 'view';
    protected isLoading: boolean = false;

    // Common DOM elements
    protected sectionHeaderElement: HTMLElement | null;
    protected contentContainer: HTMLElement | null; // The .section-content div
    protected titleElement: HTMLElement | null;
    protected typeIconElement: HTMLElement | null;
    protected deleteButton: HTMLElement | null;
    protected moveUpButton: HTMLElement | null;
    protected moveDownButton: HTMLElement | null;
    protected settingsButton: HTMLElement | null;
    protected llmButton: HTMLElement | null;
    protected addBeforeButton: HTMLElement | null;
    protected addAfterButton: HTMLElement | null;
    protected fullscreenButton: HTMLElement | null; // Moved from subclasses
    protected exitFullscreenButton: HTMLElement | null = null; // Found after view load

    constructor(data: SectionData, element: HTMLElement, callbacks: SectionCallbacks = {}) {
        this.data = data; // May have null/empty content initially
        this.element = element;
        this.callbacks = callbacks;
        this.modal = Modal.getInstance();

        // Find common structural elements
        this.sectionHeaderElement = this.element.querySelector('.section-header');
        this.contentContainer = this.element.querySelector('.section-content');
        this.titleElement = this.element.querySelector('.section-title');
        this.typeIconElement = this.element.querySelector('.section-type-icon');
        this.deleteButton = this.element.querySelector('.section-delete');
        this.moveUpButton = this.element.querySelector('.section-move-up');
        this.moveDownButton = this.element.querySelector('.section-move-down');
        this.settingsButton = this.element.querySelector('.section-settings');
        this.llmButton = this.element.querySelector('.section-ai');
        this.addBeforeButton = this.element.querySelector('.section-add-before');
        this.addAfterButton = this.element.querySelector('.section-add-after');
        this.fullscreenButton = this.element.querySelector('.section-fullscreen');
        this.exitFullscreenButton = this.element.querySelector('.section-exit-fullscreen'); // Find it once in constructor

        if (!this.contentContainer) {
            console.error(`Section content container not found for section ID: ${this.data.id}`);
        }

        this.updateDisplayTitle();
        this.updateTypeIcon();
        this.bindCommonEvents();

        // Initial state: Show loading placeholder UNLESS initial content is already provided
        // (This check is a bit redundant now as loadContent/setInitialContentAndRender will overwrite)
        if (this.contentContainer && !this.data.content) { // Only show loading if no content passed initially
            this.contentContainer.innerHTML = `
                <div class="p-4 text-center text-gray-500 dark:text-gray-400 italic">
                    Loading content...
                </div>`;
        } else if (this.contentContainer && this.data.content) {
             // If content *was* passed in constructor (e.g., from AddSection response), render it immediately.
             // This avoids the loading flash for newly created sections.
            console.log(`Section ${this.sectionId}: Initial content provided in constructor. Rendering view.`);
            this.mode = 'view';
            if (this.loadTemplate('view')) {
                this.populateViewContent();
                this.bindViewModeEvents();
            } else {
                this.contentContainer.innerHTML = `<div class="p-4 text-red-500">Error loading view template.</div>`;
            }
        }

        // Add instance to the element for easy access (e.g., retry button)
         (this.element as any).componentInstance = this;

    }

     public get sectionId(): string {
       return this.data.id
     }

     public get designId(): string {
       return this.data.designId
     }
 
     /** Enables the fullscreen button functionality for this section */
     protected enableFullscreen(): void {
         if (this.fullscreenButton) {
             this.fullscreenButton.classList.remove('hidden');
             this.fullscreenButton.removeEventListener('click', this.enterFullscreen); // Prevent duplicates
             this.fullscreenButton.addEventListener('click', this.enterFullscreen.bind(this));
             console.log(`Fullscreen enabled for section ${this.data.id}`);
         } else {
             console.warn(`Attempted to enable fullscreen, but button not found for section ${this.data.id}`);
         }
    }

    /** Removes the section's root element from the DOM */
    public removeElement(): void {
      this.element.remove();
    }

    /** Updates the displayed title in the section header */
    protected updateDisplayTitle(): void {
        if (this.titleElement) {
            this.titleElement.textContent = this.data.title;
        }
    }

    /** Updates the type icon in the section header */
    protected updateTypeIcon(): void {
        if (this.typeIconElement) {
            this.typeIconElement.innerHTML = this.getSectionIconSvg();
            this.typeIconElement.setAttribute('title', this.getSectionTypeTitle())
        }
    }

    /** Binds events for common controls (delete, move, LLM, title editing etc.) */
    protected bindCommonEvents(): void {
        // Title editing
        if (this.titleElement) {
            // Ensure only one listener is attached if constructor is called multiple times (unlikely but safe)
            this.titleElement.removeEventListener('click', this.startTitleEdit);
            this.titleElement.addEventListener('click', this.startTitleEdit.bind(this));
        }

        // Delete button
        if (this.deleteButton) {
            this.deleteButton.addEventListener('click', () => {
                this.callbacks.onDelete?.(this.data.id);
            });
        }

        // Move up button
        if (this.moveUpButton) {
            this.moveUpButton.addEventListener('click', () => {
                this.callbacks.onMoveUp?.(this.data.id);
            });
        }

        // Move down button
        if (this.moveDownButton) {
            this.moveDownButton.addEventListener('click', () => {
                this.callbacks.onMoveDown?.(this.data.id);
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

        // Add Before button
        if (this.addBeforeButton) {
            this.addBeforeButton.addEventListener('click', () => {
                console.log(`Add Before requested for section ${this.data.id}`);
                this.callbacks.onAddSectionRequest?.(this.data.id, 'before');
            });
        }

        // Add After button
        if (this.addAfterButton) {
            this.addAfterButton.addEventListener('click', () => {
                console.log(`Add After requested for section ${this.data.id}`);
                this.callbacks.onAddSectionRequest?.(this.data.id, 'after');
            });
        }
    }

    /**
     * Fetches the section's content from the API and populates the view.
     */
    public async loadContent(): Promise<void> {
        // Prevent loading if content is already present (e.g., set by constructor/setInitial)
        // or if already loading.
        if (this.data.content || this.isLoading || !this.contentContainer) {
            console.warn(`Section ${this.sectionId}: Load skipped (already loaded, loading, or container missing). Has Content: ${!!this.data.content}, Is Loading: ${this.isLoading}`);
            // If content exists but view isn't rendered (rare edge case), render it now.
            if (this.data.content && this.contentContainer?.querySelector('.text-gray-500')) { // Check if placeholder is still there
                this.setInitialContentAndRender(this.data.content);
            }
            return;
        }

        this.isLoading = true;
        this.mode = 'view'; // Ensure mode is view during load

        this.contentContainer.innerHTML = `
            <div class="p-4 text-center text-gray-500 dark:text-gray-400 italic">
                Loading content...
            </div>`;
        console.log(`Section ${this.sectionId}: Loading content for design ${this.designId}...`);

        try {
            const apiSection: V1Section = await DesignApi.designServiceGetSection({
                designId: this.designId,
                sectionId: this.sectionId,
            });
            console.log(`Section ${this.sectionId}: API response received`, apiSection);
            this.data.content = extractContentFromApiSection(apiSection); // Store fetched content

            // Update title from API response if needed
            if (apiSection.title && apiSection.title !== this.data.title) {
                this.data.title = apiSection.title;
                this.updateDisplayTitle();
            }

            console.log(`Section ${this.sectionId}: Content extracted, rendering view.`);
            // Call internal render method
            this.renderViewMode();

        } catch (error: any) {
            console.error(`Section ${this.sectionId}: Failed to load content`, error);
            const errorMsg = error.message || (error.response ? await error.response.text() : 'Unknown error');
            this.contentContainer.innerHTML = `
                 <div class="p-4 text-red-500 dark:text-red-400 text-center">
                    Error loading content: ${errorMsg}
                    <button class="ml-2 text-blue-600 dark:text-blue-400 hover:underline" onclick="document.getElementById('${this.sectionId}')?.componentInstance?.loadContent()">Retry</button>
                 </div>`;
        } finally {
            this.isLoading = false;
            console.log(`Section ${this.sectionId}: Loading finished.`);
        }
    }

    /**
     * Sets the initial content provided (e.g., after API creation) and renders the view.
     * Bypasses the API fetch in loadContent.
     * @param initialContent The content to set.
     */
    public setInitialContentAndRender(initialContent: SectionData['content']): void {
        if (this.isLoading || !this.contentContainer) {
            console.warn(`Section ${this.sectionId}: Called setInitialContentAndRender while loading or missing container.`);
            return;
        }

        console.log(`Section ${this.sectionId}: Setting initial content and rendering view.`);
        this.data.content = initialContent; // Set the content directly
        this.mode = 'view'; // Ensure mode is view

        // Call internal render method
        this.renderViewMode();

         // Ensure loading state is off if it was somehow on
         this.isLoading = false;
    }

    /**
     * Internal helper to load the view template and populate it.
     * Assumes this.data.content is already set.
     */
    private renderViewMode(): void {
        if (!this.contentContainer) return;

        if (this.loadTemplate('view')) {
            this.populateViewContent(); // Subclass implements this to render the actual content
            this.bindViewModeEvents();
            console.log(`Section ${this.sectionId}: View template loaded and populated.`);
        } else {
            console.error(`Section ${this.sectionId}: Failed to load view template.`);
            this.contentContainer.innerHTML = `
                <div class="p-4 text-red-500 dark:text-red-400 text-center">
                    Error: Could not load view template.
                </div>`;
        }
    }

    /** Handles the logic for editing the section title */
    protected startTitleEdit(e: Event): void {
        e.preventDefault();
        e.stopPropagation();

        if (!this.titleElement) return;

        const currentTitle = this.data.title; // Store original title for cancel
        const input = document.createElement('input');
        input.type = 'text';
        // Apply similar styling as the title for consistency, adjust as needed
        input.className = 'text-lg font-medium border border-gray-300 dark:border-gray-600 dark:bg-gray-700 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 dark:text-white px-1 py-0.5';
        input.value = currentTitle;
        input.style.minWidth = '200px'; // Prevent input from being too small

        // Temporarily replace title span with input
        const titleSpan = this.titleElement; // Assuming titleElement is the span
        const parent = titleSpan.parentNode;
        if (!parent) return;

        parent.replaceChild(input, titleSpan);
        input.focus();
        input.select();

        const cleanup = () => {
            input.removeEventListener('blur', handleSave);
            input.removeEventListener('keydown', handleKeyDown);
            // Ensure the original title span is back if the input is removed
            if (!input.parentNode) { // If already removed by blur save/cancel
               return;
            }
             parent.replaceChild(titleSpan, input);
             titleSpan.textContent = this.data.title; // Ensure display is up-to-date
        };

        const handleSave = () => {
            const newTitle = input.value.trim();
            if (newTitle && newTitle !== this.data.title) {
                this.data.title = newTitle;
                this.callbacks.onTitleChange?.(this.data.id, newTitle);
            }
            // Restore title display (even if unchanged, removes input)
            parent.replaceChild(titleSpan, input);
            titleSpan.textContent = this.data.title; // Update displayed text
            cleanup();
        };

        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                input.blur(); // Trigger save
            } else if (e.key === 'Escape') {
                e.preventDefault();
                // Restore original title visually and remove input
                parent.replaceChild(titleSpan, input);
                titleSpan.textContent = currentTitle; // Revert display
                this.data.title = currentTitle; // Revert data if needed (though not saved yet)
                cleanup();
            }
        };

        input.addEventListener('blur', handleSave, { once: true }); // Auto-save on blur
        input.addEventListener('keydown', handleKeyDown);
    }


    /** Opens the section settings modal (Placeholder) */
    protected openSettings(): void {
        console.log('Section settings clicked for', this.data.id);
        alert(`Settings for section: ${this.data.title} (ID: ${this.data.id})`);
    }

    /** Opens the LLM dialog modal for this section */
    protected openLlmDialog(): void {
        this.modal.show('llm-dialog', {
            sectionId: this.data.id,
            sectionType: this.data.type,
            sectionTitle: this.data.title
        });

        // Update the current section display in the LLM dialog (if modal content is ready)
        // Note: This might need a slight delay or callback if modal content isn't immediately available
        setTimeout(() => {
            const currentSectionElement = document.getElementById('llm-current-section');
            if (currentSectionElement) {
                currentSectionElement.textContent = this.data.title;
            }
        }, 0);
    }

    /**
     * Switches the section to View mode.
     * Optionally saves changes from Edit mode before switching.
     * Loads the appropriate view template from the registry.
     */
    public switchToViewMode(saveChanges: boolean): void {
        // Step 5 will add the saving logic here using this.saveContent()
        if (this.mode === 'edit' && saveChanges) {
            console.log(`switchToViewMode: Save requested for ${this.sectionId}. Saving logic TBD in Step 5.`);
             // **** PLACEHOLDER for Step 5: Call this.saveContent() ****
             const newContent = this.getContentFromEditMode();
             this.data.content = newContent; // Update local data optimistically (will be confirmed/overwritten by saveContent)
             console.log(`Section ${this.sectionId} content updated locally (API save TBD).`);

            // Cleanup editor state *after* getting content, before rendering view
            this.cleanupEditMode();

        } else if (this.mode === 'edit' && !saveChanges) {
             // Cleanup editor state if needed
             this.cleanupEditMode();
        }

        this.mode = 'view';
        console.log(`Switching ${this.sectionId} to view mode.`);
        // Render the view using the current this.data.content
        this.renderViewMode();
    }

    // Add a cleanup method to be called before switching away from edit mode without saving
    protected cleanupEditMode(): void {
        // Base implementation does nothing. Subclasses (like DrawingSection)
        // will override this to unmount React components, etc.
        console.log(`BaseSection ${this.sectionId}: Cleaning up edit mode.`);
    }

    /**
     * Switches the section to Edit mode.
     * Loads the appropriate edit template from the registry.
     */
    public switchToEditMode(): void {
        if (this.mode === 'edit') {
            // If already in edit mode, maybe focus the editor?
            console.log(`Section ${this.sectionId}: Already in edit mode.`);
            return;
        }

        // Optional: Check if content has been loaded successfully before allowing edit?
        // if (this.data.content === undefined) { // Or check based on an error flag
        //    console.warn(`Section ${this.sectionId}: Cannot switch to edit, content not loaded.`);
        //    // Optionally trigger loadContent again or show a message
        //    // await this.loadContent(); // Maybe retry load?
        //    // if (this.data.content === undefined) return; // Exit if load failed again
        //    return;
        // }

        this.mode = 'edit';
        console.log(`Switching ${this.sectionId} to edit mode.`);
        if (this.loadTemplate('edit')) {
            this.populateEditContent(); // Initialize the editor with current this.data.content
            this.bindEditModeEvents();
        } else {
             console.error("Failed to load edit template for section", this.sectionId);
             this.mode = 'view'; // Revert mode if template fails
        }
    }

    /**
     * By default binds a "section-edit-trigger" button click handler to switch to edit mode.
     * Child sections can use other bindings.
     */
    protected bindViewModeEvents(): void {
        const editTrigger = this.contentContainer?.querySelector('.section-edit-trigger');
        if (editTrigger) {
             editTrigger.removeEventListener('click', this.handleViewClick); // Prevent multiple listeners
             editTrigger.addEventListener('click', this.handleViewClick.bind(this));
        }
    }

    /**
     * Called when the container is clicked in view mode.  By default switches to switch to edit mode.
     */
    protected handleViewClick(): void {
        this.switchToEditMode();
    }

    protected bindEditModeEvents(): void {
        const saveButton = this.contentContainer?.querySelector('.section-edit-save');
        const cancelButton = this.contentContainer?.querySelector('.section-edit-cancel');

        if (saveButton) {
            saveButton.removeEventListener('click', this.handleSaveClick);
            saveButton.addEventListener('click', this.handleSaveClick.bind(this));
        }
        if (cancelButton) {
            cancelButton.removeEventListener('click', this.handleCancelClick);
            cancelButton.addEventListener('click', this.handleCancelClick.bind(this));
        }
    }

    public handleSaveClick(): void {
        console.log(`Save button clicked or shortcut used for section ${this.data.id}.`);
        this.switchToViewMode(true);
    }

    private handleCancelClick(): void {
        this.switchToViewMode(false);
    }

    /**
     * Loads the specified template (view or edit) from the registry
     * and injects it into the section's content container.
     * Uses the hidden div approach for now.
     *
     * @returns True if the template was loaded successfully, false otherwise.
     */
    protected loadTemplate(mode: 'view' | 'edit'): boolean {
        if (!this.contentContainer) return false;

        const templateId = `${this.data.type}-section-${mode}`;
        const templateRegistry = document.getElementById('template-registry');
        if (!templateRegistry) {
            console.error("Template registry not found!");
            return false;
        }

        // TODO: Migrate registry to use <template> tag for better semantics and performance.
        // When migrating, find <template> and clone its '.content' DocumentFragment.
        const templateWrapper = templateRegistry.querySelector(`[data-template-id="${templateId}"]`);
        if (!templateWrapper) {
            console.error(`Template not found in registry: ${templateId}`);
            return false;
        }

        // Using hidden div: Clone the first child element which is the actual template root
        const templateRootElement = templateWrapper.firstElementChild?.cloneNode(true) as HTMLElement | null;
        if (!templateRootElement) {
             console.error(`Template content is empty for: ${templateId}`);
             return false;
        }

        // Clear previous content and append the new template
        this.contentContainer.innerHTML = '';
        this.contentContainer.appendChild(templateRootElement);

        return true;
    }

    /** Returns the (type) title for this section. */
    protected getSectionTypeTitle(): string {
      return "Section"
    }


    /** Returns the svg content to show for this section. */
    protected getSectionIconSvg(): string {
      return ""
    }

    // --- Abstract methods to be implemented by derived classes ---

    /** Populates the loaded View mode template with the section's current data. */
    protected abstract populateViewContent(): void;

    /** Populates the loaded Edit mode template with the section's current data and initializes editors. */
    protected abstract populateEditContent(): void;

    /** Handles resizing the section's specific content (e.g., canvas, plot) when entering/exiting fullscreen or on window resize. */
    protected abstract resizeContentForFullscreen(isEntering: boolean): void;


    /** Retrieves the current content state from the Edit mode UI elements. */
    protected abstract getContentFromEditMode(): SectionData['content'];


    // --- Public API ---

    /** Updates the displayed section number */
    public updateNumber(number: number): void {
        this.data.order = number;
        const numberElement = this.element.querySelector('.section-number');
        if (numberElement) {
            numberElement.textContent = `${number}.`;
        }
    }

    /** Returns a copy of the section's data */
    public getData(): SectionData {
        // Return a copy to prevent direct modification
        return JSON.parse(JSON.stringify(this.data));
    }

    /** Returns the section's ID */
    public getId(): string {
        return this.data.id;
    }

     /** Returns the section's current order */
     public getOrder(): number {
        return this.data.order;
    }

    /**
     * Gets the section data formatted for the document model.
     * Ensures content reflects the latest saved state.
     */
    public getDocumentData(): DocumentSection {
        // Important: Assumes this.data.content is kept up-to-date by switchToViewMode(true)
        // If called while in edit mode *before* saving, it returns the *last saved* content.
         const baseData = {
            id: this.data.id,
            title: this.data.title,
            order: this.data.order,
        };

        // Return type assertion based on the section's type
        switch (this.data.type) {
            case 'text':
                return { ...baseData, type: 'text', content: this.data.content as TextContent };
            case 'drawing':
                return { ...baseData, type: 'drawing', content: this.data.content as DrawingContent };
            case 'plot':
                return { ...baseData, type: 'plot', content: this.data.content as PlotContent };
            default:
                // Should not happen if types are handled correctly
                console.error(`Unknown section type in getDocumentData: ${this.data.type}`);
                // Fallback or throw error - returning as text for now
                return { ...baseData, type: 'text', content: String(this.data.content) };
        }
    }
 
     // --- Fullscreen State and Methods ---
     protected isFullscreen: boolean = false;
 
     protected enterFullscreen(): void {
         // Use this.element as the fullscreen target
         if (this.isFullscreen || !this.element || !this.exitFullscreenButton) return;
         console.log(`Entering fullscreen for section ${this.data.id}`);
         this.isFullscreen = true;
 
         // Add classes for styling
         // Use specific classes for easier removal and potential customization
         // Target the main element now
         this.element.classList.add('lc-section-fullscreen');
         this.element.classList.add('flex', 'flex-col'); // Ensure header/content stack vertically if needed
         this.element.classList.remove('mb-6'); // Ensure header/content stack vertically if needed
         this.contentContainer?.classList.add('flex-grow', 'overflow-auto'); // Make content area take remaining space and scroll internally
         document.body.classList.add('lc-fullscreen-active');
 
         // Selectively hide header controls
         this.moveUpButton?.classList.add('hidden');
         this.moveDownButton?.classList.add('hidden');
         this.addBeforeButton?.classList.add('hidden');
         this.addAfterButton?.classList.add('hidden');
         this.settingsButton?.classList.add('hidden');
         this.deleteButton?.classList.add('hidden');
         // Keep: Title, Number, Type Icon, Fullscreen, LLM
         this.exitFullscreenButton?.classList.remove('hidden'); // Show exit button
 
         // Bind exit listeners
         this.exitFullscreenButton?.addEventListener('click', this.exitFullscreen.bind(this), { once: true });
         window.addEventListener('resize', this._handleResize.bind(this));
 
         // Allow content to adjust size *after* container is fullscreen
         requestAnimationFrame(() => {
             this.resizeContentForFullscreen(true); // Notify subclass to resize content
         });
     }
 
    protected exitFullscreen(): void {
      if (!this.isFullscreen || !this.element || !this.exitFullscreenButton) return;
      console.log(`Exiting fullscreen for section ${this.data.id}`);
      this.isFullscreen = false;
 
      // Remove classes
      this.element.classList.remove('lc-section-fullscreen');
      this.element.classList.remove('flex', 'flex-col');
      this.element.classList.add('mb-6'); // Ensure header/content stack vertically if needed
      this.contentContainer?.classList.remove('flex-grow', 'overflow-auto');
      document.body.classList.remove('lc-fullscreen-active');
      this.exitFullscreenButton?.classList.add('hidden'); // Hide exit button

      // Restore header controls
      this.moveUpButton?.classList.remove('hidden');
      this.moveDownButton?.classList.remove('hidden');
      this.addBeforeButton?.classList.remove('hidden');
      this.addAfterButton?.classList.remove('hidden');
      this.settingsButton?.classList.remove('hidden');
      this.deleteButton?.classList.remove('hidden');
 
      // Unbind listeners
      window.removeEventListener('resize', this._handleResize.bind(this));
 
      // Allow content to adjust size *after* container is back to normal
      requestAnimationFrame(() => {
        this.resizeContentForFullscreen(false); // Notify subclass to resize content
      });
   }
 
     // Bound listener function to ensure 'this' context and allow removal
     private _boundHandleKeyDown = this._handleKeyDown.bind(this);
     private _handleKeyDown(event: KeyboardEvent): void {
         if (!this.isFullscreen) return; // Check state again just in case
         if (event.key === 'Escape') {
             this.exitFullscreen();
         }
     }
 
     // Bound listener function
     private _boundHandleResize = this._handleResize.bind(this);
     private _handleResize(): void {
         if (this.isFullscreen) this.resizeContentForFullscreen(true);
     }

    /**
     * Called when the application theme (light/dark) changes.
     * Subclasses can override this method to react to the theme change,
     * e.g., by re-rendering components or adjusting styles.
     * The default implementation does nothing.
     */
    public handleThemeChange(): void {
        // Default: Do nothing. Subclasses override if needed.
        // console.log(`BaseSection ${this.data.id}: Theme change notification received.`);
    }
}
