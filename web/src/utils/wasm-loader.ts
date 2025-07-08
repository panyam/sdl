/**
 * WASM Module Loader Utility
 * 
 * Provides utilities for loading page-specific WASM modules with caching,
 * error handling, and development-friendly cache busting.
 */

export interface WasmModuleInfo {
  name: string;
  path: string;
  loaded: boolean;
  instance: any;
}

export interface WasmModuleRegistry {
  [moduleName: string]: WasmModuleInfo;
}

declare global {
  interface Window {
    sdlWasmModules?: WasmModuleRegistry;
    Go?: any; // Go WASM runtime
  }
}

export class WasmLoader {
  private static loadingPromises: Map<string, Promise<any>> = new Map();

  /**
   * Load a WASM module by name
   * @param moduleName Name of the WASM module (e.g., 'systemdetail')
   * @param cacheBust Whether to add cache busting parameter
   * @returns Promise that resolves when WASM module is loaded and ready
   */
  static async loadModule(moduleName: string, cacheBust: boolean = true): Promise<any> {
    // Check if module is already loaded
    const moduleInfo = this.getModuleInfo(moduleName);
    if (!moduleInfo) {
      throw new Error(`WASM module '${moduleName}' not found in registry`);
    }

    if (moduleInfo.loaded && moduleInfo.instance) {
      return moduleInfo.instance;
    }

    // Check if already loading
    if (this.loadingPromises.has(moduleName)) {
      return this.loadingPromises.get(moduleName);
    }

    // Start loading
    const loadingPromise = this.doLoadModule(moduleName, cacheBust);
    this.loadingPromises.set(moduleName, loadingPromise);

    try {
      const result = await loadingPromise;
      this.loadingPromises.delete(moduleName);
      return result;
    } catch (error) {
      this.loadingPromises.delete(moduleName);
      throw error;
    }
  }

  /**
   * Get module info from registry
   */
  private static getModuleInfo(moduleName: string): WasmModuleInfo | undefined {
    return window.sdlWasmModules?.[moduleName];
  }

  /**
   * Actually load the WASM module
   */
  private static async doLoadModule(moduleName: string, cacheBust: boolean): Promise<any> {
    const moduleInfo = this.getModuleInfo(moduleName);
    if (!moduleInfo) {
      throw new Error(`Module ${moduleName} not found`);
    }

    console.log(`🔧 Loading WASM module: ${moduleName}`);

    try {
      // Ensure Go runtime is available
      if (!window.Go) {
        throw new Error('Go WASM runtime not available. Make sure wasm_exec.js is loaded.');
      }

      // Build WASM path with cache busting
      let wasmPath = moduleInfo.path;
      if (cacheBust) {
        const separator = wasmPath.includes('?') ? '&' : '?';
        wasmPath += `${separator}t=${Date.now()}`;
      }

      // Create Go instance
      const go = new window.Go();

      // Load and instantiate WASM
      const result = await WebAssembly.instantiateStreaming(
        fetch(wasmPath),
        go.importObject
      );

      // Run the Go program
      go.run(result.instance);

      // Wait for module initialization
      await this.waitForModuleReady(moduleName);

      // Mark as loaded
      moduleInfo.loaded = true;
      moduleInfo.instance = window; // WASM functions are exposed to global window

      console.log(`✅ WASM module loaded: ${moduleName}`);
      return moduleInfo.instance;

    } catch (error) {
      console.error(`❌ Failed to load WASM module ${moduleName}:`, error);
      throw error;
    }
  }

  /**
   * Wait for WASM module to be ready
   * Different modules may expose different initialization signals
   */
  private static async waitForModuleReady(moduleName: string): Promise<void> {
    const maxWaitTime = 5000; // 5 seconds
    const checkInterval = 50; // 50ms
    const startTime = Date.now();

    return new Promise((resolve, reject) => {
      const checkReady = () => {
        // Check if we've exceeded max wait time
        if (Date.now() - startTime > maxWaitTime) {
          reject(new Error(`WASM module ${moduleName} failed to initialize within ${maxWaitTime}ms`));
          return;
        }

        // Module-specific readiness checks
        switch (moduleName) {
          case 'systemdetail':
            // Check if SystemDetailTool functions are available
            if (typeof (window as any).newSystemDetailTool === 'function') {
              resolve();
              return;
            }
            break;
          
          default:
            // Generic check - just wait a bit for module to initialize
            if (Date.now() - startTime > 100) {
              resolve();
              return;
            }
        }

        // Not ready yet, check again
        setTimeout(checkReady, checkInterval);
      };

      checkReady();
    });
  }

  /**
   * Check if a module is loaded
   */
  static isModuleLoaded(moduleName: string): boolean {
    const moduleInfo = this.getModuleInfo(moduleName);
    return moduleInfo?.loaded || false;
  }

  /**
   * Get list of available modules
   */
  static getAvailableModules(): string[] {
    return Object.keys(window.sdlWasmModules || {});
  }

  /**
   * Preload modules (useful for background loading)
   */
  static async preloadModules(moduleNames: string[]): Promise<void> {
    const promises = moduleNames.map(name => 
      this.loadModule(name, true).catch(error => {
        console.warn(`Failed to preload WASM module ${name}:`, error);
      })
    );
    
    await Promise.allSettled(promises);
  }

  /**
   * Clear loading cache (useful for development)
   */
  static clearCache(): void {
    this.loadingPromises.clear();
    if (window.sdlWasmModules) {
      Object.values(window.sdlWasmModules).forEach(module => {
        module.loaded = false;
        module.instance = null;
      });
    }
  }
}

/**
 * Convenience function to load SystemDetailTool WASM module
 */
export async function loadSystemDetailTool(): Promise<any> {
  return WasmLoader.loadModule('systemdetail');
}

/**
 * Check if SystemDetailTool is available
 */
export function isSystemDetailToolAvailable(): boolean {
  return WasmLoader.isModuleLoaded('systemdetail');
}