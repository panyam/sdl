// web/views/components/BaseSection.ts

import { Modal } from './Modal';
import { SectionData, SectionType, SectionContent, DocumentSection, TextContent, DrawingContent, PlotContent, SectionCallbacks } from './types';
import { V1Section } from './apiclient'; // Import API Section type
import { TemplateLoader } from './TemplateLoader'; // Import the converter
import { ToastManager } from './ToastManager'; // Import the converter
import { LlmInteractionHandler } from './LlmInteractionHandler'; // Import the new handler
import { FullscreenHandler } from './FullscreenHandler'; // Import the Fullscreen handler

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
    protected addBeforeButton: HTMLElement | null;
    protected addAfterButton: HTMLElement | null;
    protected fullscreenButton: HTMLElement | null; // Moved from subclasses
    protected exitFullscreenButton: HTMLElement | null; // Found after view load

    protected toastManager: ToastManager; // Add toast manager
    protected templateLoader = new TemplateLoader()
    protected llmInteractionHandler: LlmInteractionHandler; // Add handler instance
    protected fullscreenHandler: FullscreenHandler | null = null; // Instance for fullscreen logic

    protected allowEditOnClick: boolean = true; // Default: Allow switching to edit on view click
 
    // Helper method to find elements within the section's root element
    protected _findElement<T extends HTMLElement>(selector: string, required: boolean = true): T | null {
        const element = this.element.querySelector<T>(selector);
        if (!element && required) {
            console.warn(`BaseSection ${this.sectionId}: Required element with selector "${selector}" not found.`);
        }
        return element;
    }

    constructor(data: SectionData, element: HTMLElement, callbacks: SectionCallbacks = {}) {
        this.data = data;
        this.element = element;
        this.callbacks = callbacks;
        this.modal = Modal.getInstance();
        this.toastManager = ToastManager.getInstance();
        this.llmInteractionHandler = new LlmInteractionHandler(this.modal, this.toastManager); // Instantiate handler

        // Find common structural elements
        this.sectionHeaderElement = this._findElement('.section-header');
        this.contentContainer = this._findElement('.section-content');
        this.titleElement = this._findElement('.section-title');
        this.typeIconElement = this._findElement('.section-type-icon');
        // Find toolbar buttons (make finding optional as they might not always be present depending on state/template)
        this.deleteButton = this._findElement('.section-delete', false);
        this.moveUpButton = this._findElement('.section-move-up', false);
        this.moveDownButton = this._findElement('.section-move-down', false);
        this.settingsButton = this._findElement('.section-settings', false);
        this.addBeforeButton = this._findElement('.section-add-before', false);
        this.addAfterButton = this._findElement('.section-add-after', false);
        this.fullscreenButton = this._findElement('.section-fullscreen', false);
        this.exitFullscreenButton = this._findElement('.section-exit-fullscreen', false);

        if (!this.contentContainer) {
            console.error(`Section content container not found for section ID: ${this.data.id}`);
            return;
        }

        this.updateDisplayTitle();
        this.updateTypeIcon();
        this._bindEvents(); // Consolidated event binding
        this.initViewAndLoadingState(); // Initialize view

        // Add instance to the element for easy access (e.g., retry button)
        (this.element as any).componentInstance = this;
    }

    /** Call this when the section element is removed from the DOM */
    public destroy(): void {
        this.fullscreenHandler?.destroy(); // Clean up fullscreen listeners
    }

    /** Initializes the section view and shows the loading state */
    protected initViewAndLoadingState(): void {
        this.mode = 'view'; // Start in view mode
        this._renderCurrentMode(true); // Render view mode with initial loading indicator
    }

    /** Renders the appropriate template and content based on the current mode */
    protected _renderCurrentMode(initialLoad: boolean = false): void {
        if (!this.contentContainer) {
            console.error(`BaseSection ${this.sectionId}: Cannot render mode, content container missing.`);
            return;
        }

        const success = this.loadTemplate(this.mode);
        if (success) {
            if (this.mode === 'view') {
                if (initialLoad) {
                    // Show loading message initially
                     const viewContentArea = this.contentContainer.querySelector('.section-view-content');
                     if (viewContentArea) {
                         viewContentArea.innerHTML = `<div class="p-4 text-center text-gray-500 dark:text-gray-400 italic">Loading content...</div>`;
                     } else {
                         this.contentContainer.innerHTML = `<div class="p-4 text-red-500">Error: View content area missing.</div>`;
                     }
                } else {
                    this.populateViewContent(); // Populate with actual content when not initial load
                }
                // No need for bindViewModeEvents with delegation covering triggers
            } else { // 'edit' mode
                this.populateEditContent();
            }
        }
        // Errors handled within loadTemplate
    }

    public get sectionId(): string {
       return this.data.id
    }

     public get designId(): string {
       return this.data.designId
     }

     /** Enables the fullscreen button functionality for this section */
     protected enableFullscreen(): void {
         if (this.fullscreenButton && this.exitFullscreenButton && this.contentContainer) {
            this.fullscreenButton.classList.remove('hidden'); // Make the button visible

            // Instantiate the handler if not already done
            if (!this.fullscreenHandler) {
                 this.fullscreenHandler = new FullscreenHandler(
                     { // Pass required elements
                         sectionElement: this.element,
                         contentContainer: this.contentContainer,
                         fullscreenButton: this.fullscreenButton,
                         exitFullscreenButton: this.exitFullscreenButton,
                         moveUpButton: this.moveUpButton,
                         moveDownButton: this.moveDownButton,
                         addBeforeButton: this.addBeforeButton,
                         addAfterButton: this.addAfterButton,
                         settingsButton: this.settingsButton,
                         deleteButton: this.deleteButton,
                     },
                     this.resizeContentForFullscreen.bind(this) // Pass the resize callback
                 );
            }
             console.log(`Fullscreen enabled for section ${this.data.id}`);
         } else {
             console.warn(`Attempted to enable fullscreen, but required elements (fullscreen/exit button or content container) not found for section ${this.data.id}`);
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

    /** Binds events using delegation and direct assignment */
    protected _bindEvents(): void {
        // Title editing
        if (this.titleElement) {
            // Ensure previous listener is removed if re-binding happens
            this.titleElement.removeEventListener('click', this.startTitleEdit);
            this.titleElement.addEventListener('click', this.startTitleEdit.bind(this));
        }

        // Use event delegation for button clicks within the section element
        // We attach the listener to the root element `this.element`
        this.element.onclick = (event: MouseEvent) => {
            const target = event.target as HTMLElement;

            // Find the closest button ancestor matching known action classes
            const actionButton = target.closest<HTMLButtonElement>(
                '.section-delete, .section-move-up, .section-move-down, .section-settings, .section-ai, .section-add-before, .section-add-after, .section-edit-trigger, .section-edit-save, .section-edit-cancel'
            );

            if (!actionButton) {
                // If the click wasn't on a button, check if it was on the view content area (excluding explicit triggers)
                if (this.mode === 'view' && target.closest('.section-view-content') && !target.closest('.section-edit-trigger')) {
                    this.handleViewClick(); // Trigger edit on general view content click
                }
                return; // Click wasn't on a recognized interactive element
            }

            // Handle actions based on button class
            if (actionButton.matches('.section-delete')) {
                this.callbacks.onDelete?.(this.data.id);
            } else if (actionButton.matches('.section-move-up')) {
                this.callbacks.onMoveUp?.(this.data.id);
            } else if (actionButton.matches('.section-move-down')) {
                this.callbacks.onMoveDown?.(this.data.id);
            } else if (actionButton.matches('.section-settings')) {
                this.openSettings();
            } else if (actionButton.matches('.section-ai')) {
                this.llmInteractionHandler.showLlmDialog(this.data, this.getApplyCallback());
            } else if (actionButton.matches('.section-add-before')) {
                this.callbacks.onAddSectionRequest?.(this.data.id, 'before');
            } else if (actionButton.matches('.section-add-after')) {
                this.callbacks.onAddSectionRequest?.(this.data.id, 'after');
            } else if (actionButton.matches('.section-edit-trigger') && this.mode === 'view') {
                 this.switchToEditMode();
            } else if (actionButton.matches('.section-edit-save') && this.mode === 'edit') {
                 this.handleSaveClick();
            } else if (actionButton.matches('.section-edit-cancel') && this.mode === 'edit') {
                 this.handleCancelClick();
            }
            // Note: Fullscreen button is handled separately by FullscreenHandler
        }
    }

    /**
     * Sets view state of the section as content loading.
     * By default shows a "Loading content..." message but sections can do other things.
     * @param isLoading Whether to show the loading or not-loading state.
     * @returns Whether view state succesfully changed.
     */
    protected setContentLoading(isLoading = true): boolean {
        if (isLoading) {
            if (this.isLoading || !this.contentContainer) {
                console.warn(`Section ${this.sectionId}: Load attempt while already loading or content container missing.`);
                return false
            }
            // Ensure we are in view mode visually, even if called unexpectedly
            if (this.mode !== 'view') {
                 console.warn(`Section ${this.sectionId}: loadContent called while not in view mode. Switching...`);
                 this.switchToViewMode(false); // Cancel any edit and switch to view
            }

            this.isLoading = true;
            // --- Loading state is now handled *within* the already loaded view template ---
            // Find the content area again (it might have been replaced by error previously)
            const viewContentArea = this.contentContainer.querySelector('.section-view-content');
            if (viewContentArea) {
                 viewContentArea.innerHTML = `
                     <div class="p-4 text-center text-gray-500 dark:text-gray-400 italic">
                         Loading content...
                     </div>`;
            } else {
                 console.warn(`Section ${this.sectionId}: '.section-view-content' not found during loadContent start.`);
                 // Show loading in the main container as fallback
                 this.contentContainer.innerHTML = `<div class="p-4 text-red-500">Error: View content area missing.</div>`;
                 return false
            }
            console.log(`Section ${this.sectionId}: Loading content for design ${this.designId}...`);
        } else {
            this.isLoading = false;
            console.log(`Section ${this.sectionId}: Loading finished.`);
        }
        return true
    }

    /**
     * Fetches the section's content from the API and populates the view.
     */
    public async loadContent(): Promise<void> {
        if (!this.setContentLoading(true)) return

        // Asks the section to load any content it might need from the server (if it thinks things are stale)
        await this.refreshContentFromServer();

        const viewContentArea = this.contentContainer!.querySelector('.section-view-content');
        try {
            // --- Content Population ---
            // The view template is *already loaded*. We just need to populate it.
            this.resetViewContent(viewContentArea as HTMLDivElement)
            this.populateViewContent(); // Subclass renders into the cleared viewContentArea
            console.log(`Section ${this.sectionId}: View template populated.`);

            // Re-bind events just in case populateViewContent overwrites something, though unlikely
            this.bindViewModeEvents();
        } catch (error: any) {
            console.error(`Section ${this.sectionId}: Failed to load content`, error);
            const errorMsg = error.message || (error.response ? await error.response.text() : 'Unknown error');

             // --- Error Display ---
             // Display error within the view content area if possible
            const targetErrorArea = viewContentArea || this.contentContainer; // Fallback to main container
            if (targetErrorArea) {
              targetErrorArea.innerHTML = `
                   <div class="p-4 text-red-500 dark:text-red-400 text-center">
                      Error loading content: ${errorMsg}
                      <button class="ml-2 text-blue-600 dark:text-blue-400 hover:underline" onclick="document.getElementById('${this.sectionId}')?.componentInstance?.loadContent()">Retry</button>
                   </div>`;
            }
        } finally {
            this.setContentLoading(false)
        }
    }

    /**
     * Sets the initial content provided (e.g., by the AddSection API response)
     * and renders the view mode directly, bypassing the loadContent fetch.
     * @param initialContent The content to render immediately.
     */
    public setInitialContentAndRender(): void {
         if (this.mode !== 'view') {
              console.warn(`Section ${this.sectionId}: setInitialContentAndRender called while not in view mode. Forcing view mode.`);
              this.mode = 'view'; // Ensure correct mode
               if (!this.loadTemplate('view')) {
                    console.error(`Section ${this.sectionId}: Failed to load view template in setInitialContentAndRender.`);
                     this.contentContainer!.innerHTML = `<div class="p-4 text-red-500">Error: Could not load view template.</div>`;
                    return;
               }
         }

         console.log(`Section ${this.sectionId}: Setting initial content and rendering.`);
         this.isLoading = false; // Ensure loading is false

        // Find the content area within the already loaded view template
         const viewContentArea = this.contentContainer?.querySelector('.section-view-content');
         if (viewContentArea) {
             this.resetViewContent(viewContentArea as HTMLDivElement)
             this.populateViewContent(); // Render the provided initial content
             this.bindViewModeEvents(); // Ensure events are bound
         } else {
              console.error(`Section ${this.sectionId}: '.section-view-content' not found during setInitialContentAndRender.`);
               if (this.contentContainer) {
                    this.contentContainer.innerHTML = `<div class="p-4 text-red-500">Error: View content area missing.</div>`;
               }
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


    /**
     * Switches the section to View mode.
     * Optionally saves changes from Edit mode before switching.
     * Loads the appropriate view template from the registry.
     */
    public switchToViewMode(saveChanges: boolean): void {
        let savePromise: Promise<void> = Promise.resolve();
        if (this.mode === 'edit' && saveChanges) {
            console.log(`switchToViewMode: Save requested for ${this.sectionId}.`);
            // Trigger save; subclasses handle actual API call via handleSaveClick
            savePromise = this.handleSaveClick();
        } else if (this.mode === 'edit' && !saveChanges) {
             console.log(`switchToViewMode: Cancelling edit for ${this.sectionId}.`);
             this.cleanupEditMode();
        }

        // Switch mode and render *after* save attempt completes (or if no save needed)
        savePromise.then(() => {
            this.mode = 'view';
            console.log(`Switching ${this.sectionId} to view mode.`);
            this._renderCurrentMode(); // Render the view template
        }).catch(err => {
            console.error(`Error during save for section ${this.sectionId}, staying in edit mode.`, err);
            this.toastManager.showToast("Save Failed", `Could not save: ${err.message || 'Unknown error'}`, "error");
        });
    }

    /** Cleans up resources used in edit mode (e.g., editors, listeners) */
    protected cleanupEditMode(): void {
        console.log(`BaseSection ${this.sectionId}: Cleaning up edit mode.`);
    }

    /**
     * Switches the section to Edit mode.
     * Loads the appropriate edit template from the registry.
     */
    public switchToEditMode(): void {
        if (this.mode === 'edit') {
            console.log(`Section ${this.sectionId}: Already in edit mode.`);
            return;
        }

        this.mode = 'edit';
        console.log(`Switching ${this.sectionId} to edit mode.`);
        this._renderCurrentMode(); // Render the edit template
    }

    /**
     * By default binds a "section-edit-trigger" button click handler to switch to edit mode.
     * This method is now less critical due to event delegation in _bindEvents.
     */
    protected bindViewModeEvents(): void {
    }

    /**
     * Called when the container is clicked in view mode.  By default switches to switch to edit mode.
     */
    protected handleViewClick(): void {
    }

    public async handleSaveClick(): Promise<void> {
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
        return this.templateLoader.loadInto(templateId, this.contentContainer); // Use loadInto
    }

    /** Returns the (type) title for this section. */
    protected getSectionTypeTitle(): string {
      return "Section"
    }


    /** Returns the svg content to show for this section. */
    protected getSectionIconSvg(): string {
      return ""
    }

    // Resets the view content to default state (uses the template's initial content as default)
    protected resetViewContent(viewContentArea: HTMLDivElement) {
      viewContentArea.innerHTML = ''; // Clear any loading message
      const mode = "view"
      const templateId = `${this.data.type}-section-${mode}`;
      const clonedHtml = this.templateLoader.loadHtml(templateId)
      if (clonedHtml) {
        viewContentArea.innerHTML = clonedHtml;
      }
    }

    /** Populates the loaded View mode template with the section's current data. */
    protected abstract populateViewContent(): void;

    /** Populates the loaded Edit mode template with the section's current data and initializes editors. */
    protected abstract populateEditContent(): void;

    /** Handles resizing the section's specific content (e.g., canvas, plot) when entering/exiting fullscreen or on window resize. */
    protected abstract resizeContentForFullscreen(isEntering: boolean): void;

    /** Retrieves the current content state from the Edit mode UI elements. */
    protected abstract getContentFromEditMode(): SectionContent;

    /** Reloads the preview content from the server. */
    protected async refreshContentFromServer(): Promise<void> {
      //
    }

     /** Updates the internal content state (e.g., this.textContent). To be implemented by subclasses. */
    protected abstract updateInternalContent(newContent: SectionContent): void;

    /**
     * Returns the callback function to apply LLM results to this specific section.
     * Base implementation returns undefined. Subclasses override this.
     */
    protected getApplyCallback(): ((generatedText: string) => void) | undefined {
        return undefined; // Base sections don't know how to apply text
    }

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
