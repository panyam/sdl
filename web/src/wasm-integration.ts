export interface WASMFileSystem {
  readFile(path: string): Promise<{ success: boolean; content?: string; error?: string }>;
  writeFile(path: string, content: string): Promise<{ success: boolean; error?: string }>;
  listFiles(dir: string): Promise<{ success: boolean; files?: string[]; error?: string }>;
  mount(prefix: string, url: string): Promise<{ success: boolean; error?: string }>;
}

export interface WASMSDL {
  version: string;
  canvas: {
    load: (path: string, canvasId?: string) => any;
    use: (systemName: string, canvasId?: string) => any;
    info: (canvasId?: string) => any;
    list: () => any;
    reset: (canvasId?: string) => any;
    remove: (canvasId: string) => any;
  };
  gen: {
    add: (name: string, target: string, rate: number, options?: any) => any;
    remove: (name: string, options?: any) => any;
    update: (name: string, rate: number, options?: any) => any;
    list: (options?: any) => any;
    start: (names?: string[], options?: any) => any;
    stop: (names?: string[], options?: any) => any;
  };
  metrics: {
    add: (name: string, target: string, type: string, aggregation: string, options?: any) => any;
    remove: (name: string, options?: any) => any;
    update: (name: string, options?: any) => any;
    list: (options?: any) => any;
    query: (metric: string, options?: any) => any;
  };
  run: (options?: any) => any;
  trace: (target: string, options?: any) => any;
  flows: (options?: any) => any;
  fs: WASMFileSystem;
  config: {
    setDevMode: (enabled: boolean) => any;
  };
}

/**
 * WASMCanvasClient provides a Canvas API interface that uses WASM
 * It doesn't extend CanvasClient but provides similar methods
 */
export class WASMCanvasClient {
  private wasm: WASMSDL | null = null;
  private wasmLoaded: boolean = false;
  protected canvasId: string;

  constructor(canvasId: string = 'default') {
    this.canvasId = canvasId;
  }

  async initialize(): Promise<void> {
    if (this.wasmLoaded) return;

    try {
      // Load WASM module
      const go = new (window as any).Go();
      const result = await WebAssembly.instantiateStreaming(
        fetch("sdl.wasm"),
        go.importObject
      );
      go.run(result.instance);

      // Wait for SDL to be available
      await new Promise(resolve => setTimeout(resolve, 100));

      if ((window as any).SDL) {
        this.wasm = (window as any).SDL;
        this.wasmLoaded = true;
        console.log('✅ SDL WASM loaded successfully');
      } else {
        throw new Error('SDL global not found');
      }
    } catch (error) {
      console.error('❌ Failed to load WASM:', error);
      throw error;
    }
  }

  // CanvasClient-compatible methods

  async ensureCanvas(): Promise<any> {
    await this.initialize();
    // WASM always has a canvas ready
    return { id: this.canvasId };
  }

  async loadFile(filePath: string): Promise<any> {
    await this.initialize();
    const result = this.wasm!.canvas.load(filePath, this.canvasId);
    if (!result.success) {
      throw new Error(result.error);
    }
    return { success: true, data: result };
  }

  async useSystem(systemName: string): Promise<any> {
    await this.initialize();
    const result = this.wasm!.canvas.use(systemName, this.canvasId);
    if (!result.success) {
      throw new Error(result.error);
    }
    return { success: true, data: result };
  }

  async getCanvas(): Promise<any> {
    await this.initialize();
    const info = this.wasm!.canvas.info(this.canvasId);
    return info.success ? { id: this.canvasId, ...info } : null;
  }

  async getState(): Promise<any> {
    await this.initialize();
    const info = this.wasm!.canvas.info(this.canvasId);
    
    // Get generators
    const genResult = this.wasm!.gen.list({ canvas: this.canvasId });
    const generators = genResult.success && genResult.generators ? genResult.generators : [];
    
    // Get metrics
    const metricsResult = this.wasm!.metrics.list({ canvas: this.canvasId });
    const metrics = metricsResult.success && metricsResult.metrics ? metricsResult.metrics : [];
    
    // Convert WASM response to match server format
    return {
      loadedFiles: [], // TODO: track loaded files
      activeSystem: info.activeSystem,
      generators: generators,
      metrics: metrics
    };
  }

  async getDiagram(): Promise<any> {
    // TODO: Generate diagram from WASM system
    return null;
  }

  async getGenerators(): Promise<any> {
    await this.initialize();
    const result = this.wasm!.gen.list({ canvas: this.canvasId });
    if (!result.success) {
      return { success: false, error: result.error };
    }
    
    // Convert array to object format expected by dashboard
    const generatorsObj: any = {};
    if (result.generators) {
      result.generators.forEach((gen: any) => {
        generatorsObj[gen.id] = gen;
      });
    }
    
    return { success: true, data: generatorsObj };
  }

  async getMetrics(): Promise<any> {
    await this.initialize();
    const result = this.wasm!.metrics.list({ canvas: this.canvasId });
    if (!result.success) {
      return { success: false, error: result.error };
    }
    
    // Convert array to object format expected by dashboard
    const metricsObj: any = {};
    if (result.metrics) {
      result.metrics.forEach((metric: any) => {
        metricsObj[metric.id] = metric;
      });
    }
    
    return { success: true, data: metricsObj };
  }

  async addGenerator(name: string, component: string, method: string, rate: number): Promise<any> {
    await this.initialize();
    const target = `${component}.${method}`;
    const result = this.wasm!.gen.add(name, target, rate, {
      canvas: this.canvasId,
      applyFlows: true
    });
    if (!result.success) {
      throw new Error(result.error);
    }
    return { success: true, data: result.generator };
  }

  async removeGenerator(name: string): Promise<void> {
    await this.initialize();
    const result = this.wasm!.gen.remove(name, { canvas: this.canvasId });
    if (!result.success) {
      throw new Error(result.error);
    }
  }

  async updateGenerator(name: string, rate: number): Promise<void> {
    await this.initialize();
    const result = this.wasm!.gen.update(name, rate, {
      canvas: this.canvasId,
      applyFlows: true
    });
    if (!result.success) {
      throw new Error(result.error);
    }
  }

  async startGenerators(names?: string[]): Promise<void> {
    await this.initialize();
    const result = this.wasm!.gen.start(names, { canvas: this.canvasId });
    if (!result.success) {
      throw new Error(result.error);
    }
  }

  async stopGenerators(names?: string[]): Promise<void> {
    await this.initialize();
    const result = this.wasm!.gen.stop(names, { canvas: this.canvasId });
    if (!result.success) {
      throw new Error(result.error);
    }
  }

  async runSimulation(duration: string = "10s"): Promise<any> {
    await this.initialize();
    const result = this.wasm!.run({
      canvas: this.canvasId,
      duration: duration
    });
    if (!result.success) {
      throw new Error(result.error);
    }
    return result;
  }

  // File system methods for WASM mode
  async readFile(path: string): Promise<string> {
    await this.initialize();
    const result = await this.wasm!.fs.readFile(path);
    if (!result.success) {
      throw new Error(result.error);
    }
    return result.content!;
  }

  async writeFile(path: string, content: string): Promise<void> {
    await this.initialize();
    const result = await this.wasm!.fs.writeFile(path, content);
    if (!result.success) {
      throw new Error(result.error);
    }
  }

  async listFiles(dir: string = "/"): Promise<string[]> {
    await this.initialize();
    const result = await this.wasm!.fs.listFiles(dir);
    if (!result.success) {
      throw new Error(result.error);
    }
    return result.files || [];
  }

  // Add stub methods that dashboard expects
  async streamMetrics(_metrics: string[], _onData: (data: any) => void, _signal?: AbortSignal): Promise<void> {
    // WASM doesn't support streaming yet
    console.warn('Metric streaming not supported in WASM mode');
  }

  async getFlowState(): Promise<any> {
    // TODO: Implement flow state in WASM
    return { success: true, data: {} };
  }

  async loadSystemDiagram(): Promise<any> {
    // TODO: Implement system diagram generation in WASM
    return { success: false, error: 'Not implemented in WASM mode' };
  }
}

/**
 * Factory function to create appropriate client based on mode
 */
export function createCanvasClient(canvasId: string, useWASM: boolean = false): any {
  if (useWASM) {
    return new WASMCanvasClient(canvasId);
  }
  // Dynamic import to avoid circular dependency
  return import('./canvas-client.js').then(module => new module.CanvasClient(canvasId));
}