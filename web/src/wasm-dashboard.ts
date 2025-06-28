import { Dashboard } from './dashboard.js';
import { WASMCanvasClient, createCanvasClient } from './wasm-integration.js';
import { FileExplorer } from './components/file-explorer.js';
import { CodeEditor, configureMonacoLoader } from './components/code-editor.js';
import { DockviewApi } from 'dockview-core';

/**
 * Extended Dashboard that supports both server and WASM modes
 */
export class WASMDashboard extends Dashboard {
  private isWASMMode: boolean = false;
  private fileExplorer: FileExplorer | null = null;
  private codeEditor: CodeEditor | null = null;
  private wasmClient: WASMCanvasClient | null = null;

  constructor(canvasId: string = 'default', useWASM: boolean = false) {
    // Check URL params for WASM mode
    const urlParams = new URLSearchParams(window.location.search);
    const wasmParam = urlParams.get('wasm');
    useWASM = wasmParam === 'true' || useWASM;

    super(canvasId);
    this.isWASMMode = useWASM;

    if (this.isWASMMode) {
      // Replace the API client with WASM client
      this.api = createCanvasClient(canvasId, true) as WASMCanvasClient;
      this.wasmClient = this.api as WASMCanvasClient;
      
      // Configure Monaco for code editor
      configureMonacoLoader();
    }
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

  protected initDockView() {
    const container = document.getElementById('dockview-container');
    if (!container) {
      console.error('❌ DockView container not found');
      return;
    }

    // Add mode indicator
    this.addModeToggle();

    // Call parent's initDockView with extended component factory
    const originalCreateComponent = this.createComponent.bind(this);
    
    this.dockview = new DockviewComponent(container, {
      createComponent: (options: any) => {
        // Handle WASM-specific components
        if (this.isWASMMode) {
          switch (options.name) {
            case 'fileExplorer':
              return this.createFileExplorerComponent();
            case 'codeEditor':
              return this.createCodeEditorComponent();
            case 'console':
              return this.createConsoleComponent();
          }
        }
        
        // Fall back to parent's component creation
        return originalCreateComponent(options);
      }
    }).api;

    // Load or create layout
    const savedLayout = this.loadLayoutConfig();
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
  }

  private addModeToggle() {
    const header = document.querySelector('.header-controls');
    if (header) {
      const toggle = document.createElement('div');
      toggle.className = 'mode-toggle';
      toggle.innerHTML = `
        <label class="switch">
          <input type="checkbox" id="wasmModeToggle" ${this.isWASMMode ? 'checked' : ''}>
          <span class="slider"></span>
          <span class="label">WASM Mode</span>
        </label>
      `;
      header.insertBefore(toggle, header.firstChild);

      // Add toggle handler
      const toggleInput = document.getElementById('wasmModeToggle') as HTMLInputElement;
      toggleInput.addEventListener('change', (e) => {
        const target = e.target as HTMLInputElement;
        const newUrl = new URL(window.location.href);
        newUrl.searchParams.set('wasm', target.checked.toString());
        window.location.href = newUrl.toString();
      });
    }
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
        }
      } catch (error) {
        console.error('Failed to load file:', error);
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
    element.className = 'h-full p-4 overflow-auto bg-gray-900 text-gray-300 font-mono text-sm';
    element.innerHTML = `
      <div id="console-output">
        <div class="text-green-400">SDL WASM Console Ready</div>
        <div class="text-gray-500">Use the editor above to write SDL code and click Load to run simulations.</div>
      </div>
    `;

    // Capture console output
    const originalLog = console.log;
    const originalError = console.error;
    
    console.log = (...args) => {
      originalLog(...args);
      this.appendToConsole(args.join(' '), 'log');
    };
    
    console.error = (...args) => {
      originalError(...args);
      this.appendToConsole(args.join(' '), 'error');
    };

    return {
      element,
      init: () => {},
      dispose: () => {
        // Restore original console methods
        console.log = originalLog;
        console.error = originalError;
      }
    };
  }

  private appendToConsole(message: string, type: 'log' | 'error' = 'log') {
    const output = document.getElementById('console-output');
    if (output) {
      const div = document.createElement('div');
      div.className = type === 'error' ? 'text-red-400' : 'text-gray-300';
      div.textContent = `[${new Date().toLocaleTimeString()}] ${message}`;
      output.appendChild(div);
      output.scrollTop = output.scrollHeight;
    }
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

  // Override load file to use current editor content in WASM mode
  async loadFile(filePath: string) {
    if (this.isWASMMode && this.codeEditor) {
      // Save current editor content as the file
      const content = this.codeEditor.getValue();
      await this.wasmClient!.writeFile(filePath, content);
    }
    
    await super.loadFile(filePath);
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