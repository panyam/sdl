// ./components/documentTitle.ts

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
    private originalTitle: string = "Untitled System Design Document";

    constructor() {
        this.displayElement = document.getElementById('document-title-display') as HTMLDivElement;
        this.editElement = document.getElementById('document-title-edit');
        this.titleInput = document.getElementById('document-title-input') as HTMLInputElement;
        this.saveButton = document.getElementById('save-title-edit');
        this.cancelButton = document.getElementById('cancel-title-edit');
        this.titleText = this.displayElement?.querySelector('h1');

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

        const newTitle = this.titleInput.value.trim();
        
        // Only update if there's actual content
        if (newTitle) {
            this.titleText.textContent = newTitle;
            this.originalTitle = newTitle;
            
            // Update the document title too
            document.title = `${newTitle} - LeetCoach`;
            
            // Update timestamp if needed
            this.updateLastSavedTime();
        }

        this.showDisplayMode();
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
    updateLastSavedTime(): void {
        const timestampElement = document.getElementById('last-saved-time');
        if (timestampElement) {
            const now = new Date();
            const hours = now.getHours();
            const minutes = now.getMinutes();
            const ampm = hours >= 12 ? 'PM' : 'AM';
            const formattedHours = hours % 12 || 12; // Convert 0 to 12 for 12-hour format
            const formattedMinutes = minutes < 10 ? `0${minutes}` : minutes;
            
            timestampElement.textContent = `Last saved: Today at ${formattedHours}:${formattedMinutes} ${ampm}`;
        }
    }

     /**
      * Sets the document title programmatically.
      * Updates internal state and the displayed H1 element.
      */
     public setTitle(newTitle: string): void {
         if (newTitle && this.titleText) {
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


    /**
     * Initialize the component
     */
    public static init(): DocumentTitle {
        return new DocumentTitle();
    }
}
