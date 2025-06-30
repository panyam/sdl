import { Dashboard } from './dashboard.js';
import { WASMCanvasClient } from './wasm-integration.js';
import { configureMonacoLoader } from './components/code-editor.js';

/**
 * WASM-specific Dashboard that uses WASMCanvasClient
 */
export class WASMDashboard extends Dashboard {
  private wasmClient: WASMCanvasClient | null = null;

  constructor(canvasId: string = 'default') {
    super(canvasId);
    
    // Create WASM client and replace the API
    this.wasmClient = new WASMCanvasClient(canvasId);
    this.api = this.wasmClient as any; // WASMCanvasClient implements FileClient
    
    // Configure Monaco for code editor
    configureMonacoLoader();
  }

  public async initialize() {
    // Initialize WASM first
    if (this.wasmClient) {
      try {
        await this.wasmClient.initialize();
        console.log('✅ WASM initialized');
      } catch (error) {
        console.error('❌ Failed to initialize WASM:', error);
      }
    }

    // Then call parent initialization
    await super.initialize();
  }

  // Override refreshFileList to handle WASM-specific file system
  protected async refreshFileList() {
    if (!this.fileExplorer || !this.wasmClient) return;

    try {
      // Refresh the 'examples' filesystem (in WASM mode, this will load from virtual FS)
      await this.fileExplorer.refreshFileSystem('examples');
    } catch (error) {
      console.error('Failed to refresh file list:', error);
    }
  }

  // Override handleSave to handle read-only files in WASM
  protected async handleSave() {
    if (!this.wasmClient) return;

    if (!this.tabbedEditor) return;
    
    const activeTab = this.tabbedEditor.getActiveTab();
    const content = this.tabbedEditor.getActiveContent();
    
    if (!activeTab || !content) return;
    
    try {
      // Check if file is readonly in WASM
      const result = await (window as any).SDL.fs.isReadOnly(activeTab);
      if (result.success && result.isReadOnly) {
        // Offer to save as a copy
        const newPath = prompt('This file is read-only. Save as:', `/workspace/${activeTab.split('/').pop()}`);
        if (newPath) {
          await this.wasmClient.writeFile(newPath, content);
          this.consolePanel?.success(`Saved as: ${newPath}`);
          await this.refreshFileList();
        }
      } else {
        await this.wasmClient.writeFile(activeTab, content);
        this.consolePanel?.success(`Saved: ${activeTab}`);
        this.tabbedEditor.saveTab(activeTab);
      }
    } catch (error) {
      this.consolePanel?.error(`Failed to save: ${error}`);
    }
  }
}