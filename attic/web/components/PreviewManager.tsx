

import { FONT_FAMILY } from "@excalidraw/excalidraw";

// import { ExcalidrawWrapper, ExcalidrawToolbar } from './ExcalidrawWrapper';
import DrawingBase from "./DrawingBase";

import React, { useState, createRef, createContext, useContext } from 'react';
import ReactDOM from 'react-dom/client';
import {
  Excalidraw, MainMenu, Footer,
  serializeAsJSON,
  exportToSvg,
  convertToExcalidrawElements
} from "@excalidraw/excalidraw";
import type { ExcalidrawImperativeAPI } from "@excalidraw/excalidraw/types/types";
// import type { ExcalidrawElement } from "@excalidraw/excalidraw/types/types";

type ExcalidrawElement = any;
type AppState = any;

export default class PreviewManager extends DrawingBase {
  private imageElement: HTMLImageElement;
  private buttonElement: HTMLButtonElement;

  // excalWrapper: ExcalidrawWrapper;
  // excalToolbar: ExcalidrawToolbar;

  constructor(public readonly caseStudyId: string,
              public readonly drawingRoot: HTMLDivElement) {
    super(caseStudyId, drawingRoot)

    // The image where we will show updated image
    this.initHandlers();
  }

  private async loadFromServer(): Promise<any> {
    const response = await fetch(this.drawingUrl);
    const parsedData = await response.json()
    // const parsedData = JSON.parse(this.initialData || "{}");
    return {
      "elements": parsedData.elements || [],
      "appState": parsedData.appState || {
          theme: "light",
          viewBackgroundColor: "#ffffff"
      } as any,
    }
  }

  protected async onClick(evt: Event) {
    // first load the image
    const drawingData = await this.loadFromServer()
    this.openEditor(drawingData)
  }

  protected obtainedExcalidrawAPI(api: ExcalidrawImperativeAPI) {
    this.excalidrawInstance = api;
    setTimeout(() => { api.scrollToContent(this.elements, { fitToContent: true, animate: false, }) }, 200)
  }

  async closeEditor() {
    const divB = this.excalidrawRoot;
    if (!divB) {
      return
    }

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

    // TODO - see if a save is also needed

    // update image
    this.drawingRoot.innerHTML = `
      <a href="javascript:void(0)" class = "drawingEditorLink" id = "drawingPreviewImage_${this.caseStudyId}_{$.drawingId}">
        <img class = "drawingPreviewImage" src='${this.drawingPreviewUrl}' /
      </a>
    `

    divB.innerHTML = "";
    divB.parentElement?.removeChild(divB)
    this.excalidrawRoot = null;

    this.initHandlers()
  }

  private initHandlers() {
    this.imageElement = this.drawingRoot.querySelector('img') as HTMLImageElement;
    this.buttonElement = this.drawingRoot.querySelector('button') as HTMLButtonElement;
    if (this.imageElement) {
      this.imageElement.addEventListener('click', this.onClick.bind(this));
    } else {
      this.buttonElement.addEventListener('click', this.onClick.bind(this));
    }
  }

  openEditor(drawingData: any) {
    if (this.excalidrawRoot) {
      alert("Cannot be in this state")
      return
    }

    const divB = this.excalidrawRoot = document.createElement("div");
    document.body.appendChild(this.excalidrawRoot);

    // This will only happen when edited or clicked
    ReactDOM.createRoot(this.excalidrawRoot).render(
      <Excalidraw excalidrawAPI = {this.obtainedExcalidrawAPI.bind(this)} initialData = {drawingData} onChange={this.onChange.bind(this)}>
        <MainMenu>
          <MainMenu.Item onSelect={() => this.saveToServer()}> Save to Server </MainMenu.Item>
          <MainMenu.DefaultItems.LoadScene />
          <MainMenu.DefaultItems.SaveAsImage />
          <MainMenu.DefaultItems.Export />
          <MainMenu.DefaultItems.ToggleTheme />
          <MainMenu.DefaultItems.ClearCanvas />
          <MainMenu.DefaultItems.ChangeCanvasBackground/>
          <MainMenu.DefaultItems.Help/>
        </MainMenu>
      </Excalidraw>
    );

    const self = this;
    this.scrollYPosition = window.scrollY;

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

    const closer = (e: any) => {
      if (e.key === 'Escape') {
        if (confirm("Do you want to go finish editing?")) {
          document.removeEventListener('keydown', closer)
          this.closeEditor()
        }
      }
    }

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
      closeBtn.onclick = () => { this.closeEditor() };
      divB.appendChild(closeBtn);
    }

    // Add ESC key listener to exit fullscreen
    document.addEventListener('keydown', closer)
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
      
      console.log(`${Date.now()} - Drawing saved to API with ID: ${this.drawingId}`);
      
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

  get drawingPreviewUrl(): string {
    return `/api/drawings/${this.caseStudyId}/${this.drawingId}?format=light.svg`
  }

  get drawingUrl(): string {
    return `/api/drawings/${this.caseStudyId}/${this.drawingId}`
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
}
