import { FONT_FAMILY } from "@excalidraw/excalidraw";

// import { ExcalidrawWrapper, ExcalidrawToolbar } from './ExcalidrawWrapper';
import Split from 'split.js'
import React, { useState, createContext, useContext } from 'react';
import ReactDOM from 'react-dom/client';
import {
  Excalidraw, MainMenu, Footer , exportToBlob,
  serializeAsJSON,
  loadFromBlob,
  convertToExcalidrawElements
} from "@excalidraw/excalidraw";
import type { ExcalidrawImperativeAPI } from "@excalidraw/excalidraw/types/types";

type ExcalidrawElement = any;
type AppState = any;

// Export the class for use in browser
// export { ExcalidrawWrapper, ExcalidrawToolbar };

// Also ensure it's correctly available as window.ExcalidrawWrapper
// This makes it consistently available regardless of webpack output settings
// (window as any).ExcalidrawWrapper = ExcalidrawWrapper;

class SystemDrawing {
  private excalidrawInstance: ExcalidrawImperativeAPI | null = null;
  private elements: readonly ExcalidrawElement[] = [];
  private appState: Partial<AppState> = {};
  private isDarkTheme: boolean = false;
  private isReadOnly: boolean = false;
  private initialData: string | null = null;
  private preElement: HTMLPreElement;
  private excalidrawRoot = document.createElement("div");

  private uiOptions = {
    libraryMenu: true,   // Left sidebar toggle
    canvasActions: true, // Bottom toolbar toggle
    defaultSidebarOpen: false
  };

  // excalWrapper: ExcalidrawWrapper;
  // excalToolbar: ExcalidrawToolbar;

  constructor(public readonly caseStudyId: string,
              public readonly container: HTMLDivElement,
              public readonly toolbarContainer: HTMLDivElement,
              options?: {
                initialData?: string;
              }) {
    if (!container) {
      throw new Error("Container element is required");
    }
    ReactDOM.createRoot(container).render(
      <Excalidraw excalidrawAPI = {this.obtainedExcalidrawAPI.bind(this)}>
        <MainMenu>
          <MainMenu.DefaultItems.LoadScene />
          <MainMenu.DefaultItems.SaveAsImage />
          <MainMenu.DefaultItems.Export />
          <MainMenu.DefaultItems.ToggleTheme />
          <MainMenu.DefaultItems.ClearCanvas />
          <MainMenu.DefaultItems.ChangeCanvasBackground/>
          <MainMenu.DefaultItems.Help/>
          <MainMenu.Item onSelect={() => this.saveToServer()}> Save </MainMenu.Item>
          <MainMenu.Item onSelect={() => this.reloadFromServer()}> Reload </MainMenu.Item>
        </MainMenu>
      </Excalidraw>
    );
    
    // Check if there's a pre element with initial drawing data
    this.preElement = container.querySelector('pre') as HTMLPreElement;
    const preElementData = this.preElement ? this.preElement.textContent : null;
    if (this.preElement) {
      this.preElement.style.display = 'none';
    }
    
    // Determine initial data (prioritize explicitly passed data)
    this.initialData = options?.initialData || preElementData || null;
  }

  /**
   * Initialize the Excalidraw instance
   * @param initialData Optional JSON string containing initial drawing data
   */
  private async initialize(): Promise<void> {
    // Create a root element for Excalidraw
    // this.excalidrawRoot.style.width = "100%";
    // this.excalidrawRoot.style.height = "100%";
    // this.container.appendChild(excalidrawRoot);
    // Make sure container is positioned correctly for notifications
    // this.container.style.position = "relative";

    // If a pre element exists in the container, hide it
  }

  private obtainedExcalidrawAPI(api: ExcalidrawImperativeAPI) {
    const wrapperSelf = this
    if (api) {
      wrapperSelf.excalidrawInstance = api;
      
      // Apply default settings
      wrapperSelf.excalidrawInstance.updateScene({
        elements: wrapperSelf.elements,
        appState: {
          theme: "light",
          viewBackgroundColor: "#ffffff"
        },
      });
      
      // If there's initial data, load it now that we have the instance
      if (this.initialData && wrapperSelf.excalidrawInstance) {
        // Small delay to ensure everything is ready
        setTimeout(() => {
          try {
            if (this.initialData && wrapperSelf.excalidrawInstance) {
              wrapperSelf.loadFromJSON(this.initialData);
            }
          } catch (error) {
            console.error("Failed to load initial drawing data:", error);
          }
        }, 300);
      }
    } 
  }

  async saveToServer() {
    // Get the current drawing as JSON
    const jsonData = this.getAsJSON();
    console.log("Saved: ", jsonData)
    
    // 1. Update the Pre element if it exists
    const container = this.container;
    const preElement = container.querySelector('pre');
    if (preElement) {
      preElement.textContent = jsonData;
    }
    
    // 2. Save to API if an ID is provided in the container's data attribute
    const drawingId = container.dataset.drawingId;
    if (drawingId) {
      try {
        const response = await fetch(`/drawings/${drawingId}/`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: jsonData
        });
        
        if (!response.ok) {
          throw new Error(`API error: ${response.status}`);
        }
        
        console.log(`Drawing saved to API with ID: ${drawingId}`);
        
        // Show notification via wrapper
        this.showNotification("Drawing saved successfully!");
      } catch (error) {
        console.error("Failed to save drawing to API:", error);
        this.showNotification("Error saving drawing to API", true);
      }
    } else {
      // If no API saving is configured, fall back to download
      const blob = new Blob([jsonData], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "excalidraw-drawing.json";
      a.click();
      URL.revokeObjectURL(url);
    }
  }

  async reloadFromServer() {
    const fileInput = document.createElement("input");
    fileInput.type = "file";
    fileInput.accept = "application/json";
    fileInput.style.display = "none";
    document.body.appendChild(fileInput);
    
    fileInput.addEventListener("change", async (e: Event) => {
      const target = e.target as HTMLInputElement;
      if (target.files && target.files[0]) {
        const file = target.files[0];
        await this.loadFromBlob(file);
      }
      document.body.removeChild(fileInput);
    }, { once: true });
    
    fileInput.click();
  }

  /**
   * Load content from a JSON string
   * @param jsonData The JSON string to load
   */
  public loadFromJSON(jsonData: string): void {
    try {
      if (!this.excalidrawInstance) {
        console.error("Excalidraw instance not initialized");
        return;
      }
      
      const parsedData = JSON.parse(jsonData);
      const elements = parsedData.elements || [];
      
      this.excalidrawInstance.updateScene({
        // elements: convertToExcalidrawElements(elements),
        elements: (elements),
        appState: parsedData.appState || {},
      });
    } catch (error) {
      console.error("Failed to load data from JSON:", error);
    }
  }

  /**
   * Load content from a blob (like a file)
   * @param blob The blob containing Excalidraw data
   */
  public async loadFromBlob(blob: Blob): Promise<void> {
    try {
      if (!this.excalidrawInstance) {
        console.error("Excalidraw instance not initialized");
        return;
      }
      
      const { elements, appState, files } = await loadFromBlob(blob, null, null);
      
      // The updateScene function doesn't directly support files
      // We use a type assertion to handle this
      this.excalidrawInstance.updateScene({
        elements: elements || [],
        appState: (appState as any) || {}
      });
    } catch (error) {
      console.error("Failed to load data from blob:", error);
    }
  }

  /**
   * Export the current drawing as a blob
   * @param mimeType The mime type of the exported blob (default: "image/png")
   * @returns A Promise resolving to a Blob
   */
  public async exportToBlob(mimeType: string = "image/png"): Promise<Blob> {
    if (!this.excalidrawInstance) {
      throw new Error("Excalidraw instance not initialized");
    }
    
    return exportToBlob({
      elements: this.elements,
      appState: this.appState as AppState,
      files: {},
      mimeType
    });
  }

  /**
   * Get the current drawing as a JSON string
   * @returns A JSON string representing the current drawing
   */
  public getAsJSON(): string {
    if (!this.excalidrawInstance) {
      throw new Error("Excalidraw instance not initialized");
    }
    
    return serializeAsJSON(this.elements, this.appState as AppState, {}, "local");
  }

  /**
   * Set the theme for the editor
   * @param theme The theme to set ("light" or "dark")
   */
  public setTheme(theme: "light" | "dark"): void {
    if (!this.excalidrawInstance) {
      console.error("Excalidraw instance not initialized");
      return;
    }
    
    this.isDarkTheme = theme === "dark";
    this.appState = {
      ...this.appState,
      theme
    };
    
    this.excalidrawInstance.updateScene({
      elements: this.elements,
      appState: { theme }
    });
  }
  
  /**
   * Toggle between light and dark themes
   */
  public toggleTheme(): void {
    this.isDarkTheme = !this.isDarkTheme;
    this.setTheme(this.isDarkTheme ? "dark" : "light");
  }

  /**
   * Clear the drawing
   */
  public clearDrawing(): void {
    if (!this.excalidrawInstance) {
      console.error("Excalidraw instance not initialized");
      return;
    }
    
    this.excalidrawInstance.updateScene({
      elements: []
    });
  }

  /**
   * Set the drawing to read-only mode
   * @param readOnly Whether the drawing should be read-only
   */
  public setReadOnly(readOnly: boolean): void {
    if (!this.excalidrawInstance) {
      console.error("Excalidraw instance not initialized");
      return;
    }
    
    this.isReadOnly = readOnly;
    this.appState = {
      ...this.appState,
      viewModeEnabled: readOnly
    };
    
    this.excalidrawInstance.updateScene({
      elements: this.elements,
      appState: { viewModeEnabled: readOnly }
    });
  }
  
  /**
   * Toggle read-only mode
   */
  public toggleReadOnly(): void {
    this.isReadOnly = !this.isReadOnly;
    this.setReadOnly(this.isReadOnly);
  }

  /**
   * Set the visibility of various UI elements
   * @param options Object containing visibility options
   */
  public setUIOptions(options: {
    libraryMenu?: boolean;
    canvasActions?: boolean;
    defaultSidebarOpen?: boolean;
  }): void {
    // Update UI options
    if (options.libraryMenu !== undefined) {
      this.uiOptions.libraryMenu = options.libraryMenu;
    }
    
    if (options.canvasActions !== undefined) {
      this.uiOptions.canvasActions = options.canvasActions;
    }
    
    if (options.defaultSidebarOpen !== undefined) {
      this.uiOptions.defaultSidebarOpen = options.defaultSidebarOpen;
    }
    
    // Update Excalidraw UI if instance exists
    if (this.excalidrawInstance) {
      // For Excalidraw's built-in UI elements, we need to update the scene with new UI options
      // Note: This won't take effect immediately for all UI elements as Excalidraw has limitations
      // on dynamically updating some UI visibility options after initialization
      this.excalidrawInstance.updateScene({
        appState: {
          ...this.appState,
          openSidebar: this.uiOptions.defaultSidebarOpen ? { name: 'library' } : null
        }
      });
    }
  }
  
  /**
   * Shows a temporary notification message
   * @param message The message to display
   * @param isError Whether this is an error notification
   */
  public showNotification(message: string, isError: boolean = false): void {
    // Create notification element
    const notification = document.createElement("div");
    notification.textContent = message;
    notification.style.position = "absolute";
    notification.style.top = "10px";
    notification.style.left = "50%";
    notification.style.transform = "translateX(-50%)";
    notification.style.padding = "8px 16px";
    notification.style.borderRadius = "4px";
    notification.style.backgroundColor = isError ? "#ff4d4f" : "#52c41a";
    notification.style.color = "white";
    notification.style.boxShadow = "0 2px 8px rgba(0, 0, 0, 0.15)";
    notification.style.zIndex = "1000";
    notification.style.transition = "opacity 0.5s ease-in-out";
    notification.style.opacity = "0";
    
    // Add to container
    this.container.appendChild(notification);
    
    // Fade in
    setTimeout(() => {
      notification.style.opacity = "1";
    }, 10);
    
    // Remove after delay
    setTimeout(() => {
      notification.style.opacity = "0";
      setTimeout(() => {
        if (notification.parentNode === this.container) {
          this.container.removeChild(notification);
        }
      }, 500);
    }, 3000);
  }

  /**
   * Show all UI elements
   */
  public showUI(): void {
    this.setUIOptions({
      libraryMenu: true,
      canvasActions: true
    });
  }

  /**
   * Hide all UI elements for a clean canvas view
   */
  public hideUI(): void {
    this.setUIOptions({
      libraryMenu: false,
      canvasActions: false
    });
  }

  /**
   * Toggle the visibility of the library menu (left sidebar)
   * @param visible Whether the library menu should be visible
   */
  public toggleLibraryMenu(visible?: boolean): void {
    const newState = visible !== undefined ? visible : !this.uiOptions.libraryMenu;
    this.setUIOptions({ libraryMenu: newState });
  }

  /**
   * Toggle the visibility of canvas actions (bottom toolbar)
   * @param visible Whether the canvas actions should be visible
   */
  public toggleCanvasActions(visible?: boolean): void {
    const newState = visible !== undefined ? visible : !this.uiOptions.canvasActions;
    this.setUIOptions({ canvasActions: newState });
  }
}

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

    Split(["#outlinePanel", "#contentPanel", "#notesPanel"], { sizes: [15, 70, 15], direction: "horizontal", });

    // For testing only
    const contentPanel = document.getElementById('contentPanel') as HTMLDivElement
    // contentPanel.focus();
    contentPanel.scrollTop = contentPanel.scrollHeight;
  }
}

document.addEventListener("DOMContentLoaded", function() {
  const csp = new CaseStudyPage()
})
