import Split from 'split.js'
import SystemDrawing from "./SystemDrawing"
import PreviewManager from "./PreviewManager"
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
      const previewRoot = container.querySelector(".drawingPreviewContainer") as HTMLDivElement;
      const tbElem= container.querySelector(".drawingToolbar") as HTMLDivElement;
      const caseStudyId = (container.getAttribute("caseStudyId") || this.caseStudyId || "").trim()
      if (rootElem) {
        const sd = new SystemDrawing(caseStudyId, rootElem, tbElem)
      } else {
        // we have a "preview creator"
        const pm = new PreviewManager(caseStudyId, previewRoot)
      }
    }
    // Get references to HTML elements

    // Split(["#outlinePanel", "#contentPanel", "#notesPanel"], { sizes: [15, 70, 15], direction: "horizontal", });
    Split(["#outlinePanel", "#contentPanel"], { sizes: [15, 85], direction: "horizontal", });

    const contentPanel = document.getElementById('contentPanel') as HTMLDivElement
    const tocRoot = document.getElementById('table-of-contents') as HTMLDivElement;
    const toc = new TOCHighlighter(contentPanel, tocRoot)

    // For testing only
    const scrollToBottom = ((document.getElementById("scrollToBottom") as HTMLInputElement).value || "").trim()
    if (scrollToBottom.toLowerCase() == "true") {
      contentPanel.scrollTop = contentPanel.scrollHeight;
    }
  }
}

document.addEventListener("DOMContentLoaded", function() {
  const csp = new CaseStudyPage()
})
