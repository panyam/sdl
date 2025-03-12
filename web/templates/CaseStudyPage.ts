
import { ExcalidrawWrapper, ExcalidrawToolbar } from './ExcalidrawWrapper';
import Split from 'split.js'

// Export the class for use in browser
export { ExcalidrawWrapper, ExcalidrawToolbar };

// Also ensure it's correctly available as window.ExcalidrawWrapper
// This makes it consistently available regardless of webpack output settings
(window as any).ExcalidrawWrapper = ExcalidrawWrapper;

class SystemDrawing {
  excalWrapper: ExcalidrawWrapper;
  excalToolbar: ExcalidrawToolbar;

  constructor(public readonly caseStudyId: string,
              public readonly container: HTMLDivElement,
              public readonly toolbarContainer: HTMLDivElement) {
    this.excalWrapper = new ExcalidrawWrapper(container, {
      uiOptions: {
        // libraryMenu: true,
        // canvasActions: true,
      }
    })
    if (toolbarContainer) {
      this.excalToolbar = new ExcalidrawToolbar(toolbarContainer, {
        vertical: true,
        excalidrawWrapper: this.excalWrapper,
      })
    }
  }
}

// Overall CaseStudy page to also handle notes and TOC
class CaseStudyPage {
  caseStudyId: string
  constructor() {
    this.caseStudyId = (document.getElementById("caseStudyId") as HTMLInputElement).value

    // populate all drawings
    const drawings = document.querySelectorAll(".systemDrawing")
    for (const drawing of drawings) {
      const rootElem = drawing as HTMLDivElement
      const tbElem = rootElem.querySelector("toolbar") as HTMLDivElement
      const sd = new SystemDrawing(this.caseStudyId, rootElem, tbElem)
    }
    // Get references to HTML elements

    Split(["#outlinePanel", "#contentPanel", "#notesPanel"], { sizes: [15, 70], direction: "horizontal", })
  }
}

document.addEventListener("DOMContentLoaded", function() {
  const csp = new CaseStudyPage()
})
