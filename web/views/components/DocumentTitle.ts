// ./components/documentTitle.ts

import { DesignApi } from './Api'; // Import API client
import { ToastManager } from './ToastManager'; // Import ToastManager

/**
 * DocumentTitle component - Handles the editing of the document title
 */
export class DocumentTitle {
    private displayElement: HTMLElement | null;
    private editElement: HTMLElement | null;
    private titleInput: HTMLInputElement | null;
    private saveButton: HTMLElement | null;
    private cancelButton: HTMLElement | null;
    private titleText: HTMLElement | null;
    private lastSavedTimeElement: HTMLElement | null;
    private originalTitle: string = "Untitled System Design Document";
    private toastManager: ToastManager; // Add ToastManager instance

    constructor(public currentDesignId: string | null) {
        this.displayElement = document.getElementById('document-title-display') as HTMLDivElement;
        this.editElement = document.getElementById('document-title-edit');
        this.titleInput = document.getElementById('document-title-input') as HTMLInputElement;
        this.saveButton = document.getElementById('save-title-edit');
        this.cancelButton = document.getElementById('cancel-title-edit');
        this.titleText = this.displayElement?.querySelector('h1');
        this.lastSavedTimeElement = document.getElementById('last-saved-time');
        this.toastManager = ToastManager.getInstance(); // Initialize ToastManager

        // Store original title if it exists in the DOM
        if (this.titleText) {
             this.originalTitle = this.titleText.textContent?.trim() || this.originalTitle;
        }

        this.bindEvents();
    }

    /**
     * Bind all event listeners for the component
     */
    private bindEvents(): void {
        // Click on title display to show edit mode
        if (this.displayElement) {
            this.displayElement.addEventListener('click', this.showEditMode.bind(this));
        }

        // Save button click
        if (this.saveButton) {
            this.saveButton.addEventListener('click', this.saveTitle.bind(this));
        }

        // Cancel button click
        if (this.cancelButton) {
            this.cancelButton.addEventListener('click', this.cancelEdit.bind(this));
        }

        // Listen for Enter key in input
        if (this.titleInput) {
            this.titleInput.addEventListener('keydown', (e: KeyboardEvent) => {
                if (e.key === 'Enter') {
                    this.saveTitle();
                } else if (e.key === 'Escape') {
                    this.cancelEdit();
                }
            });
        }
    }

    /**
     * Show the edit mode for the title
     */
    private showEditMode(): void {
        if (!this.displayElement || !this.editElement || !this.titleInput) return;

        // Hide display mode
        this.displayElement.classList.add('hidden');
        
        // Show edit mode
        this.editElement.classList.remove('hidden');
        
        // Set the input value to current title
        this.titleInput.value = this.originalTitle;
        
        // Focus the input field
        this.titleInput.focus();
        this.titleInput.select();
    }

    /**
     * Save the new title and return to display mode
     */
    private saveTitle(): void {
        if (!this.displayElement || !this.editElement || !this.titleInput || !this.titleText) return;
        if (!this.currentDesignId) {
             console.error("Cannot save title: Design ID is missing.");
             this.toastManager.showToast("Save Error", "Cannot save title, Design ID not found.", "error");
             this.cancelEdit(); // Revert UI
             return;
        }

        const newTitle = this.titleInput.value.trim();
        
        if (newTitle) {
            this.titleText.textContent = newTitle;
            this.originalTitle = newTitle;

            // Update the document title too
            document.title = `${newTitle} - LeetCoach`;
            
            // Call API to save
            this.saveTitleToApi(this.currentDesignId, newTitle);

        }

        this.showDisplayMode();
    }

    /**
     * NEW: Calls the API to update the design name.
     */
    private async saveTitleToApi(designId: string, newTitle: string): Promise<void> {
        console.log(`Attempting to save title "${newTitle}" for design ID ${designId}`);
        try {
            await DesignApi.designServiceUpdateDesign({
                designId: designId, // Pass the actual design ID
                body: {
                    design: { name: newTitle }, // Only send the name in the design object
                    // updateMask: { paths: ["name"] } // Specify only 'name' in the mask
                    updateMask: "name",
                }
            });
            this.toastManager.showToast("Title Saved", `Document title updated to "${newTitle}".`, "success", 2000);
            // Update the 'last saved' timestamp on successful save
            this.updateLastSavedTime(new Date()); // Pass current time

        } catch (error: any) {
            console.error("Error saving document title:", error);
            this.toastManager.showToast("Save Error", `Failed to save title: ${error.message || 'Server error'}`, "error");
            // Optionally revert the title in the UI if save fails?
            // this.titleText!.textContent = this.originalTitle; // Revert visual
        }
    }

    /**
     * Cancel editing and return to display mode
     */
    private cancelEdit(): void {
        this.showDisplayMode();
    }

    /**
     * Show the display mode
     */
    private showDisplayMode(): void {
        if (!this.displayElement || !this.editElement) return;
        
        // Show display mode
        this.displayElement.classList.remove('hidden');
        
        // Hide edit mode
        this.editElement.classList.add('hidden');
    }

    /**
     * Update the last saved timestamp
     */
    updateLastSavedTime(date?: Date): void {
        if (this.lastSavedTimeElement) {
            const now = date || new Date(); // Use provided date or current time
            const hours = now.getHours();
            const minutes = now.getMinutes();
            const ampm = hours >= 12 ? 'PM' : 'AM';
            const formattedHours = hours % 12 || 12; // Convert 0 to 12 for 12-hour format
            const formattedMinutes = minutes < 10 ? `0${minutes}` : minutes;
            
            // Simple date check for "Today" - could be enhanced
            const isToday = now.toDateString() === new Date().toDateString();
            const prefix = isToday ? "Today" : now.toLocaleDateString();

            this.lastSavedTimeElement.textContent = `Last saved: ${prefix} at ${formattedHours}:${formattedMinutes} ${ampm}`;
        }
    }

     /**
      * Sets the document title programmatically.
      * Updates internal state and the displayed H1 element.
      */
     public setTitle(newTitle: string): void {
        if (newTitle && this.titleText) {
        if (!this.currentDesignId) {
             console.error("Cannot save title: Design ID is missing.");
             this.toastManager.showToast("Save Error", "Cannot save title, Design ID not found.", "error");
             this.cancelEdit(); // Revert UI
             return;
        }
             this.titleText.textContent = newTitle;
             this.originalTitle = newTitle;
             document.title = `${newTitle} - LeetCoach`; // Update browser tab title
             this.updateLastSavedTime(); // Also update timestamp when loading
         }
     }
 
    /**
     * Get the current title text.
     */
    public getTitle(): string {
        return this.originalTitle; // originalTitle is updated on save
    }
}
