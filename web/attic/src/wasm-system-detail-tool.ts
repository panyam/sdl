/**
 * TypeScript wrapper for SystemDetailTool WASM module
 * Provides a clean interface matching the Go tool's test structure
 */

import { loadSystemDetailTool } from './utils/wasm-loader';

export interface SystemInfo {
  compiled: boolean;
  systemCount: number;
  systems: string[];
  activeSystem?: string;
  sdlSize: number;
  recipeSize: number;
}

export interface CompileResult {
  success: boolean;
  errors: string[];
  systems: string[];
  diagram?: any;
}

export interface RecipeStep {
  index: number;
  lineNumber: number;
  command: string;
  args: string[];
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
  output?: string;
  error?: string;
  startTime?: number;
  endTime?: number;
}

export interface RecipeExecState {
  isRunning: boolean;
  currentStep: number;
  totalSteps: number;
  steps: RecipeStep[];
  mode: 'step' | 'auto';
  fileName: string;
}

export interface SystemDetailToolCallbacks {
  onError?: (message: string) => void;
  onInfo?: (message: string) => void;
  onSuccess?: (message: string) => void;
}

export class WASMSystemDetailTool {
  private wasmModule: any;
  private isLoaded = false;
  private callbacks: SystemDetailToolCallbacks = {};

  constructor() {
    this.loadWASM();
  }

  private async loadWASM(): Promise<void> {
    try {
      // Use the centralized WASM loader
      this.wasmModule = await loadSystemDetailTool();
      this.isLoaded = true;
      
      // Create a new tool instance
      this.wasmModule.newSystemDetailTool();
      
      console.log('✅ SystemDetailTool WASM loaded successfully');
    } catch (error) {
      console.error('❌ Failed to load SystemDetailTool WASM:', error);
      throw error;
    }
  }

  async waitForReady(): Promise<void> {
    while (!this.isLoaded) {
      await new Promise(resolve => setTimeout(resolve, 10));
    }
  }

  setCallbacks(callbacks: SystemDetailToolCallbacks): void {
    this.callbacks = callbacks;
    console.log("New Callbacks: ", this.callbacks)
    
    if (this.isLoaded) {
      const result = this.wasmModule.setCallbacks({
        onError: callbacks.onError || (() => {}),
        onInfo: callbacks.onInfo || (() => {}),
        onSuccess: callbacks.onSuccess || (() => {}),
      });
      
      if (!result.success) {
        throw new Error(`Failed to set callbacks: ${result.error}`);
      }
    }
  }

  async initialize(systemId: string, sdlContent: string, recipeContent: string): Promise<void> {
    await this.waitForReady();
    
    const result = this.wasmModule.initialize(systemId, sdlContent, recipeContent);
    if (!result.success) {
      throw new Error(`Failed to initialize: ${result.error}`);
    }
  }

  async setSDLContent(content: string): Promise<void> {
    await this.waitForReady();
    
    const result = this.wasmModule.setSDLContent(content);
    if (!result.success) {
      throw new Error(`Failed to set SDL content: ${result.error}`);
    }
  }

  async setRecipeContent(content: string): Promise<void> {
    await this.waitForReady();
    
    const result = this.wasmModule.setRecipeContent(content);
    if (!result.success) {
      throw new Error(`Failed to set recipe content: ${result.error}`);
    }
  }

  async getSystemInfo(): Promise<SystemInfo> {
    await this.waitForReady();
    
    const result = this.wasmModule.getSystemInfo();
    if (!result.success) {
      throw new Error(`Failed to get system info: ${result.error}`);
    }
    
    return result.data as SystemInfo;
  }

  async getCompileResult(): Promise<CompileResult | null> {
    await this.waitForReady();
    
    const result = this.wasmModule.getCompileResult();
    if (!result.success) {
      // Return null if no compilation result available (not an error)
      if (result.error && result.error.includes('No compilation result available')) {
        return null;
      }
      throw new Error(`Failed to get compile result: ${result.error}`);
    }
    
    return result.data as CompileResult;
  }

  async getExecState(): Promise<RecipeExecState | null> {
    await this.waitForReady();
    
    const result = this.wasmModule.getExecState();
    if (!result.success) {
      // Return null if no execution state available (not an error)
      if (result.error && result.error.includes('No execution state available')) {
        return null;
      }
      throw new Error(`Failed to get exec state: ${result.error}`);
    }
    
    return result.data as RecipeExecState;
  }

  async useSystem(systemName: string): Promise<void> {
    await this.waitForReady();
    
    const result = this.wasmModule.useSystem(systemName);
    if (!result.success) {
      throw new Error(`Failed to use system: ${result.error}`);
    }
  }

  async getSystemID(): Promise<string> {
    await this.waitForReady();
    
    const result = this.wasmModule.getSystemID();
    if (!result.success) {
      throw new Error(`Failed to get system ID: ${result.error}`);
    }
    
    return result.data as string;
  }

  async getSDLContent(): Promise<string> {
    await this.waitForReady();
    
    const result = this.wasmModule.getSDLContent();
    if (!result.success) {
      throw new Error(`Failed to get SDL content: ${result.error}`);
    }
    
    return result.data as string;
  }

  async getRecipeContent(): Promise<string> {
    await this.waitForReady();
    
    const result = this.wasmModule.getRecipeContent();
    if (!result.success) {
      throw new Error(`Failed to get recipe content: ${result.error}`);
    }
    
    return result.data as string;
  }

  // Utility methods for common operations
  async compileSDL(sdlContent: string): Promise<CompileResult> {
    await this.setSDLContent(sdlContent);
    const result = await this.getCompileResult();
    if (!result) {
      throw new Error('Compilation failed - no result available');
    }
    return result;
  }

  async parseRecipe(recipeContent: string): Promise<RecipeExecState> {
    await this.setRecipeContent(recipeContent);
    const result = await this.getExecState();
    if (!result) {
      throw new Error('Recipe parsing failed - no execution state available');
    }
    return result;
  }

  async isCompiled(): Promise<boolean> {
    const info = await this.getSystemInfo();
    return info.compiled;
  }

  async getSystems(): Promise<string[]> {
    const info = await this.getSystemInfo();
    return info.systems;
  }

  async getActiveSystem(): Promise<string | undefined> {
    const info = await this.getSystemInfo();
    return info.activeSystem;
  }
}
