import { Dashboard } from './dashboard.js';
import { WASMCanvasClient } from './wasm-integration.js';
import { FileExplorer } from './components/file-explorer.js';
import { CodeEditor, configureMonacoLoader } from './components/code-editor.js';
import { DockviewComponent } from 'dockview-core';
import { Toolbar } from './components/toolbar.js';

/**
 * Extended Dashboard that supports both server and WASM modes
 */
export class WASMDashboard extends Dashboard {
  private isWASMMode: boolean = false;
  private fileExplorer: FileExplorer | null = null;
  private codeEditor: CodeEditor | null = null;
  private wasmClient: WASMCanvasClient | null = null;
  private toolbar: any = null;
  private consolePanel: any = null;
  private currentFile: string | null = null;

  constructor(canvasId: string = 'default', useWASM: boolean = false) {
    // Check URL params for WASM mode
    const urlParams = new URLSearchParams(window.location.search);
    const wasmParam = urlParams.get('wasm');
    useWASM = wasmParam === 'true' || useWASM;

    super(canvasId);
    this.isWASMMode = useWASM;

    if (this.isWASMMode) {
      // Create WASM client directly
      this.wasmClient = new WASMCanvasClient(canvasId);
      this.api = this.wasmClient as any; // Type assertion needed due to different interfaces
      
      // Configure Monaco for code editor
      configureMonacoLoader();
    }
  }

  protected override render() {
    const app = document.getElementById('app');
    if (!app) return;

    app.innerHTML = `
      <div class="flex flex-col h-screen">
        <!-- Header -->
        <header class="bg-gray-800 border-b border-gray-700 px-4 py-2">
          <div class="flex items-center justify-between">
            <h1 class="text-xl font-bold">SDL Canvas: ${this.canvasId}</h1>
            <span class="text-xs text-gray-400">WASM Mode</span>
          </div>
        </header>
        
        <!-- Toolbar -->
        <div id="toolbar-container"></div>
        
        <!-- Main Content with DockView -->
      <div id="dockview-container" class="flex-1"></div>
      </div>
    `;

    // Initialize toolbar
    this.initializeToolbar();
    
    // Initialize layout
    this.initializeLayout();
  }

  private initializeToolbar() {
    const container = document.getElementById('toolbar-container');
    if (!container) return;

    this.toolbar = new Toolbar(container);
    this.toolbar.setButtons([
      {
        id: 'load',
        label: 'Load',
        icon: '📂',
        tooltip: 'Load and compile SDL file',
        onClick: () => this.handleLoad()
      },
      {
        id: 'save',
        label: 'Save',
        icon: '💾',
        tooltip: 'Save current file',
        disabled: true,
        onClick: () => this.handleSave()
      },
      {
        id: 'run',
        label: 'Run',
        icon: '▶️',
        tooltip: 'Run simulation',
        disabled: true,
        onClick: () => this.handleRun()
      },
      {
        id: 'stop',
        label: 'Stop',
        icon: '⏹️',
        tooltip: 'Stop simulation',
        disabled: true,
        onClick: () => this.handleStop()
      }
    ]);
  }

  protected createDefaultLayout() {
    if (!this.dockview) return;

    if (this.isWASMMode) {
      // WASM mode layout with file explorer and editor
      this.dockview.addPanel({
        id: 'fileExplorer',
        component: 'fileExplorer',
        title: 'Files',
        position: { direction: 'left' },
        params: { width: 250 }
      });

      this.dockview.addPanel({
        id: 'codeEditor',
        component: 'codeEditor',
        title: 'SDL Editor'
      });

      this.dockview.addPanel({
        id: 'systemArchitecture',
        component: 'systemArchitecture',
        title: 'System Architecture',
        position: { direction: 'right' }
      });

      this.dockview.addPanel({
        id: 'trafficGeneration',
        component: 'trafficGeneration',
        title: 'Traffic Generation',
        position: { referencePanel: 'systemArchitecture', direction: 'below' }
      });

      this.dockview.addPanel({
        id: 'console',
        component: 'console',
        title: 'Console',
        position: { direction: 'below' }
      });
    } else {
      // Use parent's default layout for server mode
      super.createDefaultLayout();
    }
  }

  protected initializeLayout() {
    const container = document.getElementById('dockview-container');
    if (!container) {
      console.error('❌ DockView container not found');
      return;
    }

    // Apply dark theme to container
    container.className = 'dockview-theme-dark flex-1';
    
    // Load saved layout - but only if it's a WASM layout
    let savedLayout = this.loadLayoutConfig();
    
    // Check if saved layout has WASM components, otherwise ignore it
    /*
    if (savedLayout && (!savedLayout.panels || !savedLayout.panels.some((p: any) => 
      p.id === 'fileExplorer' || p.id === 'codeEditor' || p.id === 'console'))) {
      console.log('Ignoring non-WASM saved layout');
      savedLayout = null;
    }
   */
    
    // Create DockView component with component factory
    const dockviewComponent = new DockviewComponent(container, {
      createComponent: (options: any) => {
        // Handle WASM-specific components
        switch (options.name) {
          case 'fileExplorer':
            return this.createFileExplorerComponent();
          case 'codeEditor':
            return this.createCodeEditorComponent();
          case 'console':
            return this.createConsoleComponent();
          case 'systemArchitecture':
          case 'trafficGeneration':
          case 'liveMetrics':
            // Use parent's rendering for these
            const element = document.createElement('div');
            element.className = 'h-full p-4 overflow-auto';
            
            switch (options.name) {
              case 'systemArchitecture':
                element.innerHTML = this.renderSystemArchitectureOnly();
                break;
              case 'trafficGeneration':
                element.innerHTML = this.renderGenerateControls();
                break;
              case 'liveMetrics':
                element.innerHTML = `
                  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" style="grid-auto-rows: 200px;">
                    ${this.renderDynamicCharts()}
                  </div>
                `;
                break;
            }
            
            return {
              element,
              init: () => {},
              dispose: () => {}
            };
          default:
            // Unknown component
            const unknownElement = document.createElement('div');
            unknownElement.className = 'h-full p-4 overflow-auto';
            unknownElement.innerHTML = `<div>Unknown component: ${options.name}</div>`;
            return {
              element: unknownElement,
              init: () => {},
              dispose: () => {}
            };
        }
      }
    });

    this.dockview = dockviewComponent.api;

    // Listen for layout changes and save them
    this.dockview.onDidLayoutChange(() => {
      this.saveLayoutConfig();
    });

    // Load or create layout
    if (savedLayout) {
      try {
        this.dockview.fromJSON(savedLayout);
      } catch (error) {
        console.warn('Failed to restore layout:', error);
        this.createDefaultLayout();
      }
    } else {
      this.createDefaultLayout();
    }

    // Initialize WASM
    if (this.isWASMMode && this.wasmClient) {
      this.wasmClient.initialize().then(() => {
        console.log('✅ WASM initialized');
        this.refreshFileList();
      }).catch(error => {
        console.error('❌ Failed to initialize WASM:', error);
      });
    }

    // Setup interactivity after layout is initialized
    setTimeout(() => {
      this.setupInteractivity();
      this.initDynamicCharts();
    }, 100);
  }


  private createFileExplorerComponent() {
    const element = document.createElement('div');
    element.className = 'h-full overflow-auto';
    
    this.fileExplorer = new FileExplorer(element);
    
    // Set up handlers
    this.fileExplorer.setFileSelectHandler(async (path) => {
      try {
        const content = await this.wasmClient!.readFile(path);
        if (this.codeEditor) {
          this.codeEditor.loadFile(path, content);
          this.currentFile = path;
          // Enable save button
          this.toolbar?.updateButton('save', { disabled: false });
          this.consolePanel?.info(`Loaded file: ${path}`);
        }
      } catch (error) {
        console.error('Failed to load file:', error);
        this.consolePanel?.error(`Failed to load file: ${error}`);
      }
    });

    this.fileExplorer.setFileCreateHandler(async (path) => {
      try {
        await this.wasmClient!.writeFile(path, '// New SDL file\n');
        await this.refreshFileList();
        if (this.codeEditor) {
          this.codeEditor.loadFile(path, '// New SDL file\n');
        }
      } catch (error) {
        console.error('Failed to create file:', error);
      }
    });

    // Load initial files
    this.refreshFileList();

    return {
      element,
      init: () => {},
      dispose: () => {}
    };
  }

  private createCodeEditorComponent() {
    const element = document.createElement('div');
    element.className = 'h-full';
    
    // Wait for Monaco to load
    setTimeout(() => {
      this.codeEditor = new CodeEditor(element);
      
      this.codeEditor.setChangeHandler(async (content) => {
        // Auto-save to WASM filesystem
        const currentFile = '/workspace/current.sdl'; // TODO: track current file
        try {
          await this.wasmClient!.writeFile(currentFile, content);
          console.log('Auto-saved to WASM filesystem');
        } catch (error) {
          console.error('Failed to save:', error);
        }
      });
    }, 100);

    return {
      element,
      init: () => {},
      dispose: () => {
        if (this.codeEditor) {
          this.codeEditor.dispose();
        }
      }
    };
  }

  private createConsoleComponent() {
    const element = document.createElement('div');
    element.className = 'h-full';
    
    // Import and create console panel
    import('./components/console-panel.js').then(({ ConsolePanel, ConsoleInterceptor }) => {
      const consolePanel = new ConsolePanel(element);
      const interceptor = new ConsoleInterceptor();
      interceptor.attach(consolePanel);
      
      // Store reference to console panel
      this.consolePanel = consolePanel;
      
      // Store for cleanup
      (element as any)._consolePanel = consolePanel;
      (element as any)._interceptor = interceptor;
      
      // Initial messages
      consolePanel.success('SDL WASM Console Ready');
      consolePanel.info('Use the editor above to write SDL code and click Load to run simulations.');
    });

    return {
      element,
      init: () => {},
      dispose: () => {
        const panel = (element as any)._consolePanel;
        const interceptor = (element as any)._interceptor;
        if (interceptor) {
          interceptor.detach();
        }
        if (panel) {
          panel.dispose();
        }
      }
    };
  }


  private async refreshFileList() {
    if (!this.fileExplorer || !this.wasmClient) return;

    try {
      // Get files from various directories
      const allFiles: string[] = [];
      
      // Try to list common directories
      for (const dir of ['/examples', '/lib', '/workspace']) {
        try {
          const files = await this.wasmClient.listFiles(dir);
          allFiles.push(...files);
        } catch (error) {
          // Directory might not exist
        }
      }

      await this.fileExplorer.loadFiles(allFiles);
    } catch (error) {
      console.error('Failed to refresh file list:', error);
    }
  }

  // Toolbar handlers
  private async handleLoad() {
    if (!this.currentFile || !this.wasmClient) {
      this.consolePanel?.warning('Please select an SDL file to load');
      return;
    }

    this.toolbar?.setStatus('Loading...', 'info');
    this.consolePanel?.info(`Loading ${this.currentFile}...`);

    try {
      // Load the file into the canvas
      await this.wasmClient.loadFile(this.currentFile);
      
      // Get available systems
      const info = await this.wasmClient.getCanvas();
      
      if (info && info.systems && info.systems.length > 0) {
        // Use the first system
        const systemName = info.systems[0];
        await this.wasmClient.useSystem(systemName);
        
        this.consolePanel?.success(`Loaded system: ${systemName}`);
        this.toolbar?.setStatus('Ready', 'success');
        
        // Enable run button
        this.toolbar?.updateButton('run', { disabled: false });
        
        // Update system architecture
        this.updateAllPanels();
      } else {
        this.consolePanel?.warning('No systems found in SDL file');
        this.toolbar?.setStatus('No systems found', 'error');
      }
    } catch (error) {
      this.consolePanel?.error(`Failed to load: ${error}`);
      this.toolbar?.setStatus('Load failed', 'error');
    }
  }

  private async handleSave() {
    if (!this.currentFile || !this.codeEditor || !this.wasmClient) return;

    const content = this.codeEditor.getValue();
    
    try {
      // Check if file is readonly
      const result = await (window as any).SDL.fs.isReadOnly(this.currentFile);
      if (result.success && result.isReadOnly) {
        // Offer to save as a copy
        const newPath = prompt('This file is read-only. Save as:', `/workspace/${this.currentFile.split('/').pop()}`);
        if (newPath) {
          await this.wasmClient.writeFile(newPath, content);
          this.consolePanel?.success(`Saved as: ${newPath}`);
          await this.refreshFileList();
        }
      } else {
        await this.wasmClient.writeFile(this.currentFile, content);
        this.consolePanel?.success(`Saved: ${this.currentFile}`);
      }
    } catch (error) {
      this.consolePanel?.error(`Failed to save: ${error}`);
    }
  }

  private async handleRun() {
    if (!this.wasmClient) return;

    this.toolbar?.setStatus('Running simulation...', 'info');
    this.toolbar?.updateButton('run', { disabled: true });
    this.toolbar?.updateButton('stop', { disabled: false });

    try {
      // Start all generators
      await this.wasmClient.startGenerators();
      this.consolePanel?.success('Simulation started');
      
      // Update UI
      this.updateAllPanels();
    } catch (error) {
      this.consolePanel?.error(`Failed to start simulation: ${error}`);
      this.toolbar?.setStatus('Error', 'error');
    }
  }

  private async handleStop() {
    if (!this.wasmClient) return;

    try {
      // Stop all generators
      await this.wasmClient.stopGenerators();
      this.consolePanel?.info('Simulation stopped');
      
      this.toolbar?.updateButton('run', { disabled: false });
      this.toolbar?.updateButton('stop', { disabled: true });
      this.toolbar?.setStatus('Ready', 'info');
      
      // Update UI
      this.updateAllPanels();
    } catch (error) {
      this.consolePanel?.error(`Failed to stop simulation: ${error}`);
    }
  }

}

// Add WASM mode styles
const style = document.createElement('style');
style.textContent = `
  .mode-toggle {
    margin-right: 20px;
  }
  
  .switch {
    position: relative;
    display: inline-flex;
    align-items: center;
    cursor: pointer;
  }
  
  .switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }
  
  .slider {
    position: relative;
    display: inline-block;
    width: 44px;
    height: 24px;
    background-color: #555;
    border-radius: 24px;
    margin-right: 8px;
    transition: .4s;
  }
  
  .slider:before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    border-radius: 50%;
    transition: .4s;
  }
  
  input:checked + .slider {
    background-color: #0e639c;
  }
  
  input:checked + .slider:before {
    transform: translateX(20px);
  }
  
  .label {
    color: #cccccc;
    font-size: 14px;
  }
`;
document.head.appendChild(style);
