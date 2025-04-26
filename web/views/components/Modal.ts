// web/views/components/Modal.ts
import { TemplateLoader } from './TemplateLoader'; // Import TemplateLoader

/**
 * Modal manager for the application
 * Handles showing and hiding modals with different content
 */
export class Modal {
  private static instance: Modal | null = null;

  // Modal DOM elements
  private modalContainer: HTMLDivElement
  private modalBackdrop: HTMLElement | null;
  private modalPanel: HTMLElement | null;
  private modalContent: HTMLElement | null;
  private closeButton: HTMLElement | null;

  private templateLoader: TemplateLoader;

  // Current modal data - including optional callbacks
  private currentTemplateId: string | null = null;
  private currentData: any = null;
  private onSubmitCallback: ((modalData: any) => void) | null = null; // Store the callback
  private onApplyCallback: ((modalData: any) => void) | null = null; // Store the Apply callback


  /**
   * Private constructor for singleton pattern
   */
  private constructor() {
    // Get modal elements
    this.modalContainer = document.getElementById('modal-container') as HTMLDivElement;
    this.modalBackdrop = document.getElementById('modal-backdrop');
    this.modalPanel = document.getElementById('modal-panel');
    this.modalContent = document.getElementById('modal-content');
    this.closeButton = document.getElementById('modal-close');

    this.templateLoader = new TemplateLoader();

    this.bindEvents();
  }

  /**
   * Get the Modal instance (singleton)
   */
  public static getInstance(): Modal {
    if (!Modal.instance) {
      Modal.instance = new Modal();
    }
    return Modal.instance;
  }

  /**
   * Bind event listeners for modal interactions
   */
  private bindEvents(): void {
    // Close button click
    if (this.closeButton) {
      this.closeButton.addEventListener('click', () => this.hide());
    }

    // Click on backdrop to close
    if (this.modalBackdrop) {
      this.modalBackdrop.addEventListener('click', (e) => {
        // Only close if clicking directly on the backdrop
        if (e.target === this.modalBackdrop) {
          this.hide();
        }
      });
    }

    // Listen for Escape key
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && this.isVisible()) {
        this.hide();
      }
    });

    // --- Event Delegation for Modal Actions ---
    if (this.modalPanel) { // Listen on a persistent parent element
        this.modalPanel.addEventListener('click', (e: MouseEvent) => {
            const target = e.target as HTMLElement;

            // Handle generic close/cancel buttons
            const closeButton = target.closest('button[id$="-cancel"], button[id$="-close"]');
            if (closeButton) {
                console.log(`Modal cancel/close button clicked: ${closeButton.id}`);
                this.hide();
                return; // Stop further processing
            }

            // Handle specific actions like submit
            const actionButton = target.closest('button[data-modal-action]');
            if (actionButton) {
                const action = actionButton.getAttribute('data-modal-action');
                console.log(`Modal action button clicked: ${action}`);

                if (action === 'submit' && this.onSubmitCallback) {
                    this.onSubmitCallback(this.currentData); // Call the stored callback
                } else if (action === 'apply' && this.onApplyCallback) {
                    this.onApplyCallback(this.currentData); // Call the Apply callback
                    this.hide(); // Typically hide after applying
                }
                // Add more actions (e.g., 'apply', 'revise') later if needed
            }
        });
    }
  }

  /**
   * Check if the modal is currently visible
   */
  public isVisible(): boolean {
    return this.modalContainer ? !this.modalContainer.classList.contains('hidden') : false;
  }

  /**
   * Show a modal with content from the specified template ID.
   * Uses TemplateLoader to get the content element.
   * @param templateId ID used in `data-template-id` attribute in TemplateRegistry.html
   * @param data Optional data to pass to the modal. Can include callbacks like `onSubmit`.
   * @returns The root HTMLElement of the loaded content, or null if failed.
   */
  public show(templateId: string, data: any = null): HTMLElement | null {
    if (!this.modalContainer || !this.modalContent) {
        console.error("Modal container or content area not found.");
        return null;
    }

    // Use TemplateLoader to get the content element
    const contentElement = this.templateLoader.load(templateId);
    if (!contentElement) {
        console.error(`Modal content template not found or failed to load: ${templateId}`);
        // Optional: Show an error message in the modal itself
        this.modalContent.innerHTML = `<div class="p-6 text-red-600">Error: Could not load modal content ('${templateId}'). Check TemplateRegistry.html.</div>`;
        // Still show the modal container so the error is visible
        this.modalContainer.classList.remove('hidden');
        setTimeout(() => this.modalContainer.classList.add('modal-active'), 10);
        return null; // Indicate failure
    }

    // Store current modal info
    this.currentTemplateId = templateId;
    this.currentData = data || {}; // Ensure data is an object
    this.onSubmitCallback = data?.onSubmit || null; // Store the submit callback
    this.onApplyCallback = data?.onApply || null; // Store the apply callback

    // Clear existing content
    this.modalContent.innerHTML = '';

    // Add the loaded content element to the modal
    this.modalContent.appendChild(contentElement);

    // Set data attributes for non-function data
    if (data) {
      Object.entries(data).forEach(([key, value]) => {
        if (key !== 'onSubmit' && (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean')) {
          if (this.modalContent) this.modalContent.dataset[key] = String(value);
        }
      });
    }

    // Show modal container
    this.modalContainer.classList.remove('hidden');

    // Trigger animations if needed (add active class after a tick)
    setTimeout(() => {
      this.modalContainer.classList.add('modal-active');
    }, 10); // Small delay ensures transition applies correctly

    return contentElement; // Return the loaded element
  }

  /**
   * Hide the modal
   */
  public hide(): Promise<void> {
    return new Promise((resolve) => {
        if (!this.modalContainer) return;

        // Remove active class first (for animations)
        this.modalContainer.classList.remove('modal-active');

        // Hide after a short delay
        setTimeout(() => {
          this.modalContainer.classList.add('hidden');

          // Clear current modal info
          this.currentTemplateId = null;
          this.currentData = null;
          this.onSubmitCallback = null; // Clear callback
          this.onApplyCallback = null; // Clear apply callback
          if(this.modalContent) this.modalContent.innerHTML = ''; // Clear content
          resolve();
        }, 200); // Match typical transition duration
    });
  }

  /**
   * Get the current modal content element
   */
  public getContentElement(): HTMLElement | null {
    return this.modalContent;
  }

  /**
   * Get the current template ID
   */
  public getCurrentTemplate(): string | null {
    return this.currentTemplateId;
  }

  /**
   * Get the current modal data
   */
  public getCurrentData(): any {
    return this.currentData;
  }

  /**
   * Update modal data (excluding callbacks for now)
   */
  public updateData(newData: any): void {
    this.currentData = { ...this.currentData, ...newData };

    // Update data attributes for non-function data
    if (this.modalContent && newData) {
      Object.entries(newData).forEach(([key, value]) => { // Exclude callbacks when setting data attributes
         if (key !== 'onSubmit' && (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean')) {
          if (this.modalContent) this.modalContent.dataset[key] = String(value);
        }
      });
    }
  }

  /**
   * Initialize the modal component
   */
  public static init(): Modal {
    return Modal.getInstance();
  }
}
