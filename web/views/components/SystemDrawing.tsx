
import { FONT_FAMILY } from "@excalidraw/excalidraw";

// import { ExcalidrawWrapper, ExcalidrawToolbar } from './ExcalidrawWrapper';
import React, { useState, createRef, createContext, useContext } from 'react';
import ReactDOM from 'react-dom/client';
import {
  Excalidraw, MainMenu, Footer,
  serializeAsJSON,
  loadFromBlob,
  exportToBlob,
  exportToSvg,
  convertToExcalidrawElements
} from "@excalidraw/excalidraw";
import type { ExcalidrawImperativeAPI } from "@excalidraw/excalidraw/types/types";
// import type { ExcalidrawElement } from "@excalidraw/excalidraw/types/types";

type ExcalidrawElement = any;
type AppState = any;

export default class SystemDrawing {
  private excalidrawInstance: ExcalidrawImperativeAPI | null = null;
  private elements: readonly ExcalidrawElement[] = [];
  private appState: Partial<AppState> = {};
  private isDarkTheme: boolean = false;
  private isReadOnly: boolean = false;
  private initialData: string | null = null;
  private preElement: HTMLPreElement;
  private excalidrawRoot: HTMLDivElement;
  private toggleFSButton: HTMLButtonElement;
  private lastUpdatedAt = 0
  private lastSavedAt = -1
  private saveSentAt = -1
  private saverIntervalId = null as any

  private uiOptions = {
    libraryMenu: true,   // Left sidebar toggle
    canvasActions: true, // Bottom toolbar toggle
    defaultSidebarOpen: false
  };

  // excalWrapper: ExcalidrawWrapper;
  // excalToolbar: ExcalidrawToolbar;

  constructor(public readonly designId: string,
              public readonly sectionId: string,
              public readonly drawingRoot: HTMLDivElement,
              options?: {
                initialData?: string;
              }) {
    if (!drawingRoot) {
      throw new Error("Container element is required");
    }
    this.excalidrawRoot = drawingRoot.querySelector(".excalidrawRoot") as HTMLDivElement;
    if (!this.excalidrawRoot) {
      this.excalidrawRoot = document.createElement("div");
      this.excalidrawRoot.style.width = "100%";
      this.excalidrawRoot.style.height = "100%";
      drawingRoot.appendChild(this.excalidrawRoot);
      //this.container.style.position = "relative";
    // Create a root element for Excalidraw
    // Make sure container is positioned correctly for notifications
    }

    
    // Check if there's a pre element with initial drawing data
    this.preElement = drawingRoot.querySelector('pre') as HTMLPreElement;
    const preElementData = this.preElement ? this.preElement.textContent : null;
    if (this.preElement) {
      this.preElement.style.display = 'none';
    }
    
    // Determine initial data (prioritize explicitly passed data)
    this.initialData = options?.initialData || preElementData || null;
    const parsedData = JSON.parse(this.initialData || "{}");

    // TODO - see how to save app state
    const initialData = {
      "elements": parsedData.elements || [],
      "appState": parsedData.appState || {
          theme: "light",
          viewBackgroundColor: "#ffffff"
      } as any,
    }

    const ref = createRef()

    ReactDOM.createRoot(this.excalidrawRoot).render(
      <Excalidraw excalidrawAPI = {this.obtainedExcalidrawAPI.bind(this)} initialData = {initialData} onChange={this.onChange.bind(this)}>
        <MainMenu>
          <MainMenu.Item onSelect={() => this.saveToServer()}> Save to Server </MainMenu.Item>
          <MainMenu.Item onSelect={() => this.reloadFromServer()}> Reload from Server </MainMenu.Item>
          <MainMenu.DefaultItems.LoadScene />
          <MainMenu.DefaultItems.SaveAsImage />
          <MainMenu.DefaultItems.Export />
          <MainMenu.DefaultItems.ToggleTheme />
          <MainMenu.DefaultItems.ClearCanvas />
          {true && <MainMenu.Item onSelect={() => this.toggleFullScreen()}> Toggle Full Screen </MainMenu.Item> }
          <MainMenu.DefaultItems.ChangeCanvasBackground/>
          <MainMenu.DefaultItems.Help/>
        </MainMenu>
      </Excalidraw>
    );
  }

  private onChange(elements: readonly ExcalidrawElement[], state: AppState, files: any) {
    this.elements = elements;
    this.appState = state;
    this.lastUpdatedAt = Date.now()
  }

  private obtainedExcalidrawAPI(api: ExcalidrawImperativeAPI) {
    this.excalidrawInstance = api;
    setTimeout(() => { api.scrollToContent(this.elements, { fitToContent: true, animate: true, }) }, 200)
  }

  scrollYPosition = 0
  setFullscreen(fs = true) {
    const divB = this.excalidrawRoot;
    const self = this;
    /**
     * Handler for ESC key to exit fullscreen mode
     */
    function escKeyHandler(e: any) {
      if (e.key === 'Escape') {
        self.setFullscreen(false);
      }
    }

    if (fs) {
        this.scrollYPosition = window.scrollY;

        // Move divB to body for fullscreen modal display
        document.body.appendChild(divB);

        // Apply fullscreen styles
        divB.style.position = 'fixed';
        divB.style.top = '0';
        divB.style.left = '0';
        divB.style.width = '100%';
        divB.style.height = '100%';
        divB.style.zIndex = '9999';
        divB.style.background = 'white';
        divB.style.overflow = 'auto';
        divB.style.padding = '40px';
        divB.style.boxSizing = 'border-box';

        // Prevent scrolling on the body
        document.body.style.overflow = 'hidden';

        // Add a semi-transparent overlay
        const overlay = document.createElement('div');
        overlay.id = 'fullscreen-overlay';
        overlay.style.position = 'fixed';
        overlay.style.top = '0';
        overlay.style.left = '0';
        overlay.style.width = '100%';
        overlay.style.height = '100%';
        overlay.style.backgroundColor = 'rgba(0, 0, 0, 0.5)';
        overlay.style.zIndex = '9998';
        document.body.insertBefore(overlay, divB);

        // Add close button to fullscreen mode
        if (!document.getElementById('close-fullscreen-btn')) {
          const closeBtn = document.createElement('button');
          closeBtn.id = 'close-fullscreen-btn';
          closeBtn.innerHTML = '<img src="/static/images/close_fullscreen.svg" width="32px"/>'; // Close Fullscreen';
          closeBtn.classList.add('close-btn');
          closeBtn.style.position = 'fixed';
          closeBtn.style.top = '20px';
          closeBtn.style.right = '20px';
          closeBtn.style.zIndex = '10000';
          closeBtn.onclick = () => {
            this.setFullscreen(false);
          };
          divB.appendChild(closeBtn);
        }

        // Add ESC key listener to exit fullscreen
        document.addEventListener('keydown', escKeyHandler);
    } else {
        // Remove the overlay if it exists
        const overlay = document.getElementById('fullscreen-overlay');
        if (overlay) {
          overlay.remove();
        }

        // Remove close button
        const closeBtn = document.getElementById('close-fullscreen-btn');
        if (closeBtn) {
          closeBtn.remove();
        }

        // Move divB back to its original parent
        this.drawingRoot.appendChild(divB);

        // Restore original styles
        divB.style.position = "";
        divB.style.padding = '0px';
        /*
        divB.style.top = divB.originalPosition.top;
        divB.style.left = divB.originalPosition.left;
        divB.style.width = divB.originalPosition.width;
        divB.style.height = divB.originalPosition.height;
        divB.style.zIndex = divB.originalPosition.zIndex;
        divB.style.padding = '20px';
       */
        divB.style.background = '#e0f7fa';
        divB.style.overflow = '';

        // Restore body scrolling
        document.body.style.overflow = '';

        // Restore scroll position
        if (this.scrollYPosition !== undefined) {
          window.scrollTo(0, this.scrollYPosition);
        }

        // Remove ESC key listener
        document.removeEventListener('keydown', escKeyHandler);
    }
  }

  toggleFullScreen() {
    if (this.excalidrawRoot.parentElement == this.drawingRoot) {
      this.setFullscreen(true)
    } else {
      this.setFullscreen(false)
    }
  }

  async saveToServer(showNotification = true) {
    const jsonData = this.getAsJSON();
    const asSvgDark = await exportToSvg({
      elements: this.elements,
      appState: {
        exportBackground: true,
        exportWithDarkMode: true,
      },
    } as any)

    const asSvgLight = await exportToSvg({
      elements: this.elements,
      appState: {
        exportBackground: true,
        exportWithDarkMode: false,
      },
    } as any)

    const payload = JSON.stringify({
      "formats": {
        "json": jsonData,
        "dark.svg": asSvgDark.outerHTML,
        "light.svg": asSvgLight.outerHTML,
      },
    })

    console.log("Saved: ", jsonData)
    this.saveSentAt = Date.now()
    try {
      const response = await fetch(this.drawingUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: payload,
      });
      
      if (!response.ok) {
        throw new Error(`API error: ${response.status}`);
      }
      
      console.log(`${Date.now()} - Drawing saved to API with ID: ${this.sectionId}`);

      // 1. Update the Pre element if it exists
      if (this.preElement) {
        this.preElement.textContent = jsonData;
      }
      
      // Show notification via wrapper
      if (showNotification) this.showNotification("Drawing saved successfully!");
      this.lastSavedAt = Date.now()
    } catch (error) {
      console.error("Failed to save drawing to API:", error);
      this.showNotification("Error saving drawing to API", true);
    } finally {
      this.saveSentAt = -1;
    }
  }

  get drawingUrl(): string {
    return `/api/drawings/${this.sectionId}`
  }

  async reloadFromServer() {
    const response = await fetch(this.drawingUrl);
    const data = await response.json()
    this.loadFromJSON(JSON.stringify(data))
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
        elements: convertToExcalidrawElements(elements),
        // elements: (elements),
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
    this.excalidrawRoot.appendChild(notification);
    
    // Fade in
    setTimeout(() => {
      notification.style.opacity = "1";
    }, 10);
    
    // Remove after delay
    setTimeout(() => {
      notification.style.opacity = "0";
      setTimeout(() => {
        if (notification.parentNode === this.excalidrawRoot) {
          this.excalidrawRoot.removeChild(notification);
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
