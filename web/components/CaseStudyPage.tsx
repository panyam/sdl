import Split from 'split.js'
import SystemDrawing from "./SystemDrawing"
import TOCHighlighter from "./TOCHighlighter"

// Export the class for use in browser
// export { ExcalidrawWrapper, ExcalidrawToolbar };

// Also ensure it's correctly available as window.ExcalidrawWrapper
// This makes it consistently available regardless of webpack output settings
// (window as any).ExcalidrawWrapper = ExcalidrawWrapper;

// Overall CaseStudy page to also handle notes and TOC
class CaseStudyPage {
  caseStudyId: string
  constructor() {
    this.caseStudyId = (document.getElementById("caseStudyId") as HTMLInputElement).value

    // populate all drawings
    const drawings = document.querySelectorAll(".drawingContainer")
    for (const container of drawings) {
      const rootElem = container.querySelector(".systemDrawing") as HTMLDivElement;
      const tbElem= container.querySelector(".drawingToolbar") as HTMLDivElement;
      const sd = new SystemDrawing(this.caseStudyId, rootElem, tbElem)
    }
    // Get references to HTML elements

    // Split(["#outlinePanel", "#contentPanel", "#notesPanel"], { sizes: [15, 70, 15], direction: "horizontal", });
    Split(["#outlinePanel", "#contentPanel"], { sizes: [15, 85], direction: "horizontal", });

    // For testing only
    const contentPanel = document.getElementById('contentPanel') as HTMLDivElement
    const tocRoot = document.getElementById('table-of-contents') as HTMLDivElement;
    const toc = new TOCHighlighter(contentPanel, tocRoot)
  }
}

document.addEventListener("DOMContentLoaded", function() {
  const csp = new CaseStudyPage()
})
