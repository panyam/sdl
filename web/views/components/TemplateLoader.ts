
// Helper to manage our templates
export class TemplateLoader {
  constructor(public registryName = "template-registry") {
  }


  /**
   * Finds the specified template wrapper element in the registry.
   * @param templateId The data-template-id of the wrapper element.
   * @returns The wrapper HTMLElement or null if not found.
   */
  private _findTemplateWrapper(templateId: string): HTMLElement | null {
      const templateRegistry = document.getElementById(this.registryName);
      if (!templateRegistry) {
          console.error(`Template registry '#${this.registryName}' not found!`);
          return null;
      }
      const templateWrapper = templateRegistry.querySelector<HTMLElement>(`[data-template-id="${templateId}"]`);
      if (!templateWrapper) {
          console.error(`Template with ID "${templateId}" not found in registry '#${this.registryName}'.`);
          return null;
      }
      return templateWrapper;
  }

  /**
   * Loads and returns the inner HTML of a template definition.
   * @param templateId The data-template-id of the wrapper element.
   * @returns The innerHTML content as string if it exists otherwise null.
   */
  loadHtml(templateId: string): string | null {
    const templateWrapper = this._findTemplateWrapper(templateId);
    if (!templateWrapper) {
        return null;
    }
    return templateWrapper.innerHTML
  }

  /**
   * Loads and clones the child elements of a template definition.
   * @param templateId The data-template-id of the wrapper element.
   * @returns An array of cloned HTMLElement children, or an empty array if not found or has no children.
   */
  load(templateId: string): HTMLElement[] {
    const templateWrapper = this._findTemplateWrapper(templateId);
    if (!templateWrapper) {
        return [];
    }

    // Using hidden div: Clone the first child element which is the actual template root
    const templateRootElement = templateWrapper.cloneNode(true) as HTMLElement | null;
    if (!templateRootElement) {
      console.error(`Template content is empty for: ${templateId}`);
      return [];
    }
    // deep clone AND return th children
    return Array.from(templateRootElement.children) as HTMLElement[];

    // clond the children instead
    // Clone each child element individually and return as an array
    // const childElements = Array.from(templateWrapper.children) as HTMLElement[];
    // return childElements.map(child => child.cloneNode(true) as HTMLElement);
  }

   /**
   * Loads a template's content, clears the target element, and appends the cloned content into it.
   * @param templateId The data-template-id of the wrapper element to load.
   * @param targetElement The HTMLElement where the cloned content should be placed.
   * @returns True if the operation was successful (template found and content appended, even if content was empty), false otherwise.
   */
  public loadInto(templateId: string, targetElement: HTMLElement | null): boolean {
    if (!targetElement) {
      console.error(`Cannot load template "${templateId}": Target element is null.`);
      return false;
    }

    const templateWrapper = this._findTemplateWrapper(templateId);
    if (!templateWrapper) {
       // Error logged in _findTemplateWrapper
       targetElement.innerHTML = `<div class="p-4 text-red-500">Error loading template '${templateId}' (Not Found)</div>`; // Provide feedback in target
       return false;
    }

    // Clear the target element first
    targetElement.innerHTML = '';

    // Clone and append children
    const childElements = Array.from(templateWrapper.children) as HTMLElement[];
    if (childElements.length === 0) {
      console.warn(`Template "${templateId}" has no child elements to load.`);
      // Still considered successful, just loaded nothing.
    } else {
      childElements.forEach(child => {
        targetElement.appendChild(child.cloneNode(true) as HTMLElement);
      });
    }

    return true; // Operation successful
  }
}
