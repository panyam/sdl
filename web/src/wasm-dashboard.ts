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
      const allFiles: string[] = [];
      
      // WASM has specific directories
      for (const dir of ['/examples', '/lib', '/workspace']) {
        try {
          const files = await this.wasmClient.listFiles(dir);
          allFiles.push(...files);
        } catch (error) {
          // Directory might not exist
          console.debug(`Could not list ${dir}:`, error);
        }
      }

      await this.fileExplorer.loadFiles(allFiles);
    } catch (error) {
      console.error('Failed to refresh file list:', error);
    }
  }

  // Override handleSave to handle read-only files in WASM
  protected async handleSave() {
    if (!this.currentFile || !this.codeEditor || !this.wasmClient) return;

    const content = this.codeEditor.getValue();
    
    try {
      // Check if file is readonly in WASM
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
}