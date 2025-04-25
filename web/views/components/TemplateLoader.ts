
// Helper to manage our templates
export class TemplateLoader {
  constructor(public registryName = "template-registry") {
  }

  load(templateId: string): HTMLElement | null {
    const templateRegistry = document.getElementById(this.registryName);
    if (!templateRegistry) {
      console.error("Template registry not found!");
      return null;
    }

    // TODO: Migrate registry to use <template> tag for better semantics and performance.
    // When migrating, find <template> and clone its '.content' DocumentFragment.
    const templateWrapper = templateRegistry.querySelector(`[data-template-id="${templateId}"]`);
    if (!templateWrapper) {
      console.error(`Template not found in registry: ${templateId}`);
      return null;
    }

    // Using hidden div: Clone the first child element which is the actual template root
    const templateRootElement = templateWrapper.cloneNode(true) as HTMLElement | null;
    if (!templateRootElement) {
      console.error(`Template content is empty for: ${templateId}`);
      return null;
    }
    return templateRootElement;
  }
}
