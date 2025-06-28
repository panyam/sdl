import { CanvasClient } from './canvas-client.js';

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
 * WASMCanvasClient implements the same interface as CanvasClient but uses WASM
 */
export class WASMCanvasClient extends CanvasClient {
  private wasm: WASMSDL | null = null;
  private wasmLoaded: boolean = false;
  private canvasId: string;

  constructor(canvasId: string = 'default') {
    super(canvasId);
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

  async ensureCanvas(): Promise<void> {
    await this.initialize();
    // WASM always has a canvas ready
  }

  async loadFile(filePath: string): Promise<void> {
    await this.initialize();
    const result = this.wasm!.canvas.load(filePath, this.canvasId);
    if (!result.success) {
      throw new Error(result.error);
    }
  }

  async useSystem(systemName: string): Promise<void> {
    await this.initialize();
    const result = this.wasm!.canvas.use(systemName, this.canvasId);
    if (!result.success) {
      throw new Error(result.error);
    }
  }

  async getState(): Promise<any> {
    await this.initialize();
    const info = this.wasm!.canvas.info(this.canvasId);
    
    // Convert WASM response to match server format
    return {
      loadedFiles: [], // TODO: track loaded files
      activeSystem: info.activeSystem,
      generators: [], // TODO: get generators from WASM
      metrics: [] // TODO: get metrics from WASM
    };
  }

  async getDiagram(): Promise<any> {
    // TODO: Generate diagram from WASM system
    return null;
  }

  async addGenerator(name: string, component: string, method: string, rate: number): Promise<void> {
    await this.initialize();
    const target = `${component}.${method}`;
    const result = this.wasm!.gen.add(name, target, rate, {
      canvas: this.canvasId,
      applyFlows: true
    });
    if (!result.success) {
      throw new Error(result.error);
    }
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
    return result.files!;
  }
}

/**
 * Factory function to create appropriate client based on mode
 */
export function createCanvasClient(canvasId: string, useWASM: boolean = false): CanvasClient {
  if (useWASM) {
    return new WASMCanvasClient(canvasId);
  }
  return new CanvasClient(canvasId);
}