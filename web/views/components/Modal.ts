// components/Modal.ts

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

  // Current modal data
  private currentTemplateId: string | null = null;
  private currentData: any = null;

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

    // Generic modal button handlers
    document.addEventListener('click', (e) => {
      const target = e.target as HTMLElement;
      
      // Handle cancel buttons by convention (id ending with "-cancel")
      if (target.id && target.id.endsWith('-cancel')) {
        this.hide();
      }
    });
  }

  /**
   * Check if the modal is currently visible
   */
  public isVisible(): boolean {
    return this.modalContainer ? !this.modalContainer.classList.contains('hidden') : false;
  }

  /**
   * Show a modal with content from the specified template
   * @param templateId ID of the template element
   * @param data Optional data to pass to the modal
   */
  public show(templateId: string, data: any = null): void {
    if (!this.modalContainer || !this.modalContent) return;

    console.log(`Looking for template: ${templateId}-template`);
    let templateElement = document.getElementById(`${templateId}-template`);
    if (!templateElement) {
      console.error(`Modal template not found: ${templateId}-template`);
      console.log(`Available templates:`, Array.from(document.querySelectorAll('[id$="-template"]')).map(el => el.id));
      // Check if we have a template registry to clone from
      const templateRegistry = document.getElementById('template-registry');
      if (templateRegistry) {
         const registeredTemplate = templateRegistry.querySelector(`[data-template-id="${templateId}"]`);
         if (registeredTemplate) {
           console.log(`Found template in registry: ${templateId}`);
           // Clone the template and add it to the body with the expected ID
           const clonedTemplate = registeredTemplate.cloneNode(true) as HTMLElement;
           clonedTemplate.id = `${templateId}-template`;
           clonedTemplate.classList.add('hidden');
           document.body.appendChild(clonedTemplate);

           // Now try to get the template again
           templateElement = document.getElementById(`${templateId}-template`);
         }
      }

      if (!templateElement) {
         return; // Still not found, exit
      }
    }

    // Store current modal info
    // this.currentTemplateId = templateId;
    this.currentTemplateId = templateId.replace('-template', '');
    this.currentData = data;

    // Clear existing content
    this.modalContent.innerHTML = '';

    // Clone template content
    const contentElement = templateElement.cloneNode(true) as HTMLElement;
    contentElement.classList.remove('hidden');
    contentElement.id = templateId; // Remove the "-template" suffix

    // Add content to modal
    this.modalContent.appendChild(contentElement);

    // Set data attributes for any data that needs to be accessed later
    if (data) {
      Object.entries(data).forEach(([key, value]) => {
        if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
          if (this.modalContent) this.modalContent.dataset[key] = String(value);
        }
      });
    }

    // Show modal
    this.modalContainer.classList.remove('hidden');
    
    // Trigger animations if needed
    setTimeout(() => {
      this.modalContainer.classList.add('modal-active');
    }, 10);
  }

  /**
   * Hide the modal
   */
  public hide(): void {
    if (!this.modalContainer) return;

    // Remove active class first (for animations)
    this.modalContainer.classList.remove('modal-active');
    
    // Hide after a short delay
    setTimeout(() => {
      this.modalContainer.classList.add('hidden');
      
      // Clear current modal info
      this.currentTemplateId = null;
      this.currentData = null;
    }, 200);
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
   * Update modal data
   */
  public updateData(newData: any): void {
    this.currentData = { ...this.currentData, ...newData };
    
    // Update data attributes
    if (this.modalContent && newData) {
      Object.entries(newData).forEach(([key, value]) => {
        if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
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

// Initialize the component when the DOM is fully loaded
// document.addEventListener('DOMContentLoaded', () => { Modal.init(); });
