import {
  Excalidraw,
  exportToBlob,
  serializeAsJSON,
  loadFromBlob,
  convertToExcalidrawElements
} from "@excalidraw/excalidraw";

// Import types correctly
import type { ExcalidrawImperativeAPI } from "@excalidraw/excalidraw/types/types";
// Define element types based on library
type ExcalidrawElement = any;
type AppState = any;

// Define our toolbar button properties
export interface ToolbarButton {
  label: string;
  onClick: () => void;
  className?: string;
}

/**
 * Independent toolbar for Excalidraw that can be placed anywhere
 */
export class ExcalidrawToolbar {
  private toolbarElement: HTMLDivElement;
  private buttons: ToolbarButton[] = [];
  private excalidrawWrapper: ExcalidrawWrapper | null = null;
  
  /**
   * Creates a toolbar for Excalidraw
   * @param container The HTML element to mount the toolbar in
   * @param options Configuration options
   */
  constructor(
    container: HTMLElement, 
    options?: {
      vertical?: boolean;
      excalidrawWrapper?: ExcalidrawWrapper;
    }
  ) {
    // Create toolbar element
    this.toolbarElement = document.createElement("div");
    this.toolbarElement.className = "excalidraw-toolbar";
    this.toolbarElement.style.display = "flex";
    this.toolbarElement.style.flexDirection = options?.vertical ? "column" : "row";
    this.toolbarElement.style.gap = "5px";
    this.toolbarElement.style.padding = "5px";
    
    // Set wrapper if provided
    if (options?.excalidrawWrapper) {
      this.setExcalidrawWrapper(options.excalidrawWrapper);
    }
    
    // Append to container
    container.appendChild(this.toolbarElement);
  }
  
  /**
   * Associate an ExcalidrawWrapper with this toolbar
   * @param wrapper The ExcalidrawWrapper instance
   */
  public setExcalidrawWrapper(wrapper: ExcalidrawWrapper): void {
    this.excalidrawWrapper = wrapper;
    
    // Add default buttons if wrapper is provided
    this.addDefaultButtons();
  }
  
  /**
   * Add a button to the toolbar
   * @param button The button configuration
   */
  public addButton(button: ToolbarButton): void {
    const buttonEl = document.createElement("button");
    buttonEl.textContent = button.label;
    buttonEl.style.padding = "6px 12px";
    buttonEl.style.cursor = "pointer";
    buttonEl.style.border = "1px solid #ccc";
    buttonEl.style.borderRadius = "4px";
    buttonEl.style.backgroundColor = "white";
    buttonEl.style.transition = "background-color 0.2s";
    
    if (button.className) {
      buttonEl.className = button.className;
    }
    
    buttonEl.addEventListener("click", button.onClick);
    
    // Add hover effect
    buttonEl.addEventListener("mouseover", () => {
      buttonEl.style.backgroundColor = "#f0f0f0";
    });
    
    buttonEl.addEventListener("mouseout", () => {
      buttonEl.style.backgroundColor = "white";
    });
    
    this.toolbarElement.appendChild(buttonEl);
    this.buttons.push(button);
  }
  
  /**
   * Add default buttons for controlling Excalidraw
   */
  private addDefaultButtons(): void {
    if (!this.excalidrawWrapper) {
      return;
    }
    
    // Clear existing buttons
    while (this.toolbarElement.firstChild) {
      this.toolbarElement.removeChild(this.toolbarElement.firstChild);
    }
    this.buttons = [];
    
    // Add save button
    this.addButton({
      label: "Save",
      onClick: async () => {
        if (this.excalidrawWrapper) {
          // Get the current drawing as JSON
          const jsonData = this.excalidrawWrapper.getAsJSON();
          
          // 1. Update the Pre element if it exists
          const container = this.excalidrawWrapper.getContainer();
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
              this.excalidrawWrapper.showNotification("Drawing saved successfully!");
            } catch (error) {
              console.error("Failed to save drawing to API:", error);
              this.excalidrawWrapper.showNotification("Error saving drawing to API", true);
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
      }
    });
    
    // Add load button
    this.addButton({
      label: "Load",
      onClick: () => {
        if (this.excalidrawWrapper) {
          const fileInput = document.createElement("input");
          fileInput.type = "file";
          fileInput.accept = "application/json";
          fileInput.style.display = "none";
          document.body.appendChild(fileInput);
          
          fileInput.addEventListener("change", async (e: Event) => {
            const target = e.target as HTMLInputElement;
            if (target.files && target.files[0] && this.excalidrawWrapper) {
              const file = target.files[0];
              await this.excalidrawWrapper.loadFromBlob(file);
            }
            document.body.removeChild(fileInput);
          }, { once: true });
          
          fileInput.click();
        }
      }
    });
    
    // Add clear button
    this.addButton({
      label: "Clear",
      onClick: () => {
        if (this.excalidrawWrapper) {
          this.excalidrawWrapper.clearDrawing();
        }
      }
    });
    
    // Add theme toggle button
    this.addButton({
      label: "Toggle Theme",
      onClick: () => {
        if (this.excalidrawWrapper) {
          this.excalidrawWrapper.toggleTheme();
        }
      }
    });
    
    // Add readonly toggle button
    this.addButton({
      label: "Toggle Read-Only",
      onClick: () => {
        if (this.excalidrawWrapper) {
          this.excalidrawWrapper.toggleReadOnly();
        }
      }
    });
  }
  
  /**
   * Get the toolbar element
   * @returns The HTML element containing the toolbar
   */
  public getElement(): HTMLDivElement {
    return this.toolbarElement;
  }
}

/**
 * A vanilla TypeScript wrapper for Excalidraw that doesn't require React
 */
export class ExcalidrawWrapper {
  private container: HTMLDivElement;
  private excalidrawInstance: ExcalidrawImperativeAPI | null = null;
  private elements: readonly ExcalidrawElement[] = [];
  private appState: Partial<AppState> = {};
  private isDarkTheme: boolean = false;
  private isReadOnly: boolean = false;
  
  // UI visibility state
  private uiOptions = {
    libraryMenu: true,   // Left sidebar toggle
    canvasActions: true, // Bottom toolbar toggle
    defaultSidebarOpen: false
  };

  /**
   * Creates an instance of ExcalidrawWrapper.
   * @param container The HTML div element to mount Excalidraw on
   * @param options Configuration options for the Excalidraw instance
   */
  constructor(
    container: HTMLDivElement,
    options?: {
      initialData?: string;
      uiOptions?: {
        libraryMenu?: boolean;
        canvasActions?: boolean;
        defaultSidebarOpen?: boolean;
      }
    }
  ) {
    if (!container) {
      throw new Error("Container element is required");
    }
    this.container = container;
    
    // Check if there's a pre element with initial drawing data
    const preElement = container.querySelector('pre');
    const preElementData = preElement ? preElement.textContent : null;
    
    // Apply UI options if provided
    if (options?.uiOptions) {
      if (options.uiOptions.libraryMenu !== undefined) {
        this.uiOptions.libraryMenu = options.uiOptions.libraryMenu;
      }
      if (options.uiOptions.canvasActions !== undefined) {
        this.uiOptions.canvasActions = options.uiOptions.canvasActions;
      }
      if (options.uiOptions.defaultSidebarOpen !== undefined) {
        this.uiOptions.defaultSidebarOpen = options.uiOptions.defaultSidebarOpen;
      }
    }
    
    // Determine initial data (prioritize explicitly passed data)
    const initialData = options?.initialData || preElementData;
    
    // Initialize the drawing component
    this.initialize(initialData);
  }
  
  /**
   * Get the container element
   * @returns The container element
   */
  public getContainer(): HTMLDivElement {
    return this.container;
  }

  /**
   * Initialize the Excalidraw instance
   * @param initialData Optional JSON string containing initial drawing data
   */
  private async initialize(initialData: string | null = null): Promise<void> {
    try {
      // Create a root element for Excalidraw
      const excalidrawRoot = document.createElement("div");
      excalidrawRoot.style.width = "100%";
      excalidrawRoot.style.height = "100%";
      
      // If a pre element exists in the container, hide it
      const preElement = this.container.querySelector('pre');
      if (preElement) {
        preElement.style.display = 'none';
      }
      
      this.container.appendChild(excalidrawRoot);
      
      // Make sure container is positioned correctly for notifications
      this.container.style.position = "relative";

      // We need to create a React element and render it
      // This is internal to the wrapper - the consumer doesn't need to use React
      const excalidrawModule = await import("@excalidraw/excalidraw");
      const React = await import("react");
      const ReactDOM = await import("react-dom");
      
      // Create a React component for Excalidraw
      interface ExcalidrawComponentProps {
        onChange: (elements: readonly ExcalidrawElement[], appState: Partial<AppState>) => void;
      }
      
      // Render the component
      const self = this;

      class ExcalidrawComponent extends React.Component<ExcalidrawComponentProps> {
        excalidrawRef: React.RefObject<ExcalidrawImperativeAPI>;
        
        constructor(props: ExcalidrawComponentProps) {
          super(props);
          this.excalidrawRef = React.createRef<ExcalidrawImperativeAPI>();
        }

        private obtainedExcalidrawAPI(api: ExcalidrawImperativeAPI) {
            // Pass the instance back to our wrapper
          if (api) {
            const wrapperSelf = self; // Reference to the wrapper
            wrapperSelf.excalidrawInstance = api;
            
            // Apply default settings
            wrapperSelf.excalidrawInstance.updateScene({
              elements: wrapperSelf.elements,
              appState: {
                theme: "light",
                viewBackgroundColor: "#ffffff"
              }
            });
            
            // If there's initial data, load it now that we have the instance
            if (initialData && wrapperSelf.excalidrawInstance) {
              // Small delay to ensure everything is ready
              setTimeout(() => {
                try {
                  if (wrapperSelf.excalidrawInstance) {
                    wrapperSelf.loadFromJSON(initialData);
                  }
                } catch (error) {
                  console.error("Failed to load initial drawing data:", error);
                }
              }, 300);
            }
          } 
        }
        
        componentDidMount() {
          if (this.excalidrawRef.current) {
            this.obtainedExcalidrawAPI(this.excalidrawRef.current)
          }
        }
        
        render() {
          // Create Excalidraw component with correct properties
          const props: any = {
            onChange: (elements: readonly ExcalidrawElement[], appState: Partial<AppState>) => {
              this.props.onChange(elements, appState);
            }
          };
          
          // Properly configure UI options
          if (self.uiOptions) {
            props.UIOptions = {
              tools: self.uiOptions.canvasActions,
              menu: self.uiOptions.canvasActions
            };
            
            // Configure canvas actions
            if (self.uiOptions.canvasActions !== undefined) {
              props.UIOptions.canvasActions = {};
              
              // Set each canvas action to either an empty object (enabled) or undefined (disabled)
              if (self.uiOptions.canvasActions) {
                props.UIOptions.canvasActions.export = {};
                props.UIOptions.canvasActions.saveAsScene = {};
                props.UIOptions.canvasActions.clearCanvas = {};
                props.UIOptions.canvasActions.loadScene = {};
                props.UIOptions.canvasActions.saveToActiveFile = {};
                props.UIOptions.canvasActions.saveFileToDisk = {};
              } else {
                props.UIOptions.canvasActions.export = undefined;
                props.UIOptions.canvasActions.saveAsScene = undefined;
                props.UIOptions.canvasActions.clearCanvas = undefined;
                props.UIOptions.canvasActions.loadScene = undefined;
                props.UIOptions.canvasActions.saveToActiveFile = undefined;
                props.UIOptions.canvasActions.saveFileToDisk = undefined;
              }
            }
            
            // Set sidebar preferences
            if (self.uiOptions.defaultSidebarOpen !== undefined) {
              props.UIOptions.defaultSidebarDockedPreference = self.uiOptions.defaultSidebarOpen;
            }
          }
          
          // Use createRef via props instead of direct ref assignment
          props.excalidrawRef = this.excalidrawRef;
          props.excalidrawAPI = (api: ExcalidrawImperativeAPI) => {
            console.log("Ok we got here???")
            this.obtainedExcalidrawAPI(api)
          }
          
          return React.createElement(excalidrawModule.Excalidraw, props);
        }
      }
      
      // Simply use the legacy React 17 render method to avoid issues
      // The newer React 18 createRoot API would require importing from react-dom/client
      // @ts-ignore - Ignoring TypeScript errors since ReactDOM.render is technically deprecated
      ReactDOM.render(
        React.createElement(ExcalidrawComponent, {
          onChange: (elements: readonly ExcalidrawElement[], appState: Partial<AppState>) => {
            self.elements = elements;
            self.appState = appState;
            // You can add custom onChange handling here
          }
        }),
        excalidrawRoot
      );
    } catch (error) {
      console.error("Failed to initialize Excalidraw:", error);
    }
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
        appState: parsedData.appState || {}
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
        appState: appState || {}
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
