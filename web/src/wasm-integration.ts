import { FileClient } from './types.js';
import { create } from '@bufbuild/protobuf';
import { SystemDiagramSchema, DiagramNodeSchema, DiagramEdgeSchema } from './gen/sdl/v1/canvas_pb.js';

export interface WASMFileSystem {
  readFile(path: string): Promise<{ success: boolean; content?: string; error?: string }>;
  writeFile(path: string, content: string): Promise<{ success: boolean; error?: string }>;
  listFiles(dir: string): Promise<{ success: boolean; files?: string[]; error?: string }>;
  mount(prefix: string, url: string): Promise<{ success: boolean; error?: string }>;
  isReadOnly(path: string): Promise<{ success: boolean; isReadOnly?: boolean; error?: string }>;
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
export class WASMCanvasClient implements FileClient {
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
      const go = new (globalThis as any).Go();
      const result = await WebAssembly.instantiateStreaming(
        fetch(`/sdl.wasm?t=${Date.now()}`),
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
    if (info.success) {
      return { 
        id: this.canvasId, 
        hasActiveSystem: info.hasActiveSystem,
        activeSystem: info.activeSystem,
        systems: info.systems || [],
        generators: info.generators || 0
      };
    }
    return null;
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
  
  async deleteFile(_path: string): Promise<void> {
    // WASM filesystem doesn't support delete yet
    throw new Error('Delete not supported in WASM mode');
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

/**
 * Generate a system diagram from SDL content using WASM
 * This parses the SDL and creates a diagram structure similar to what the server would return
 */
export async function generateSystemDiagram(sdlContent: string): Promise<any> {
  try {
    // Initialize WASM if not already loaded
    const wasm = (window as any).SDL;
    if (!wasm) {
      // Try to load WASM
      const go = new (globalThis as any).Go();
      const result = await WebAssembly.instantiateStreaming(
        fetch(`/sdl.wasm?t=${Date.now()}`),
        go.importObject
      );
      go.run(result.instance);
      
      // Wait for SDL to be available
      await new Promise(resolve => setTimeout(resolve, 100));
      
      if (!(window as any).SDL) {
        throw new Error('Failed to load SDL WASM');
      }
    }
    
    // Use WASM to parse SDL content
    const sdl = (window as any).SDL;
    
    // Create a temporary canvas to load the SDL
    const tempCanvasId = 'temp-diagram-' + Date.now();
    
    // Save SDL content to virtual filesystem
    const tempPath = '/tmp/system.sdl';
    await sdl.fs.writeFile(tempPath, sdlContent);
    
    // Load the SDL file into the canvas
    const loadResult = sdl.canvas.load(tempPath, tempCanvasId);
    if (!loadResult.success) {
      throw new Error(loadResult.error || 'Failed to load SDL');
    }
    
    // Get canvas info which includes the system information
    const info = sdl.canvas.info(tempCanvasId);
    if (!info.success) {
      throw new Error(info.error || 'Failed to get canvas info');
    }
    
    // Extract system name from the loaded systems
    const systemName = info.systems && info.systems.length > 0 ? info.systems[0] : 'System';
    
    // Parse SDL to extract components and their methods
    const components = parseSDLComponents(sdlContent);
    
    // Create nodes for each component method
    const nodes: any[] = [];
    const edges: any[] = [];
    
    // Generate nodes
    components.forEach(component => {
      component.methods.forEach(method => {
        const nodeId = `${component.name}:${method}`;
        nodes.push(create(DiagramNodeSchema, {
          id: nodeId,
          name: method,
          type: component.type,
          traffic: '0 req/s', // Initial traffic
          fullPath: component.varName,
          icon: component.type,
          methods: [] // Will be populated later if needed
        }));
      });
    });
    
    // Generate edges based on dependencies
    components.forEach(component => {
      component.dependencies.forEach(dep => {
        // Find the dependency component
        const depComponent = components.find(c => c.varName === dep.varName);
        if (depComponent) {
          // Create edges from each method of the component to methods of the dependency
          component.methods.forEach(method => {
            depComponent.methods.forEach(depMethod => {
              edges.push(create(DiagramEdgeSchema, {
                fromId: `${component.name}:${method}`,
                toId: `${depComponent.name}:${depMethod}`,
                fromMethod: method,
                toMethod: depMethod,
                label: '',
                color: '#9ca3af',
                order: 0,
                condition: '',
                probability: 0,
                generatorId: ''
              }));
            });
          });
        }
      });
    });
    
    // Clean up temporary canvas
    sdl.canvas.remove(tempCanvasId);
    
    // Create and return the system diagram
    const diagram = create(SystemDiagramSchema, {
      systemName: systemName,
      nodes: nodes,
      edges: edges
    });
    
    return diagram;
  } catch (error) {
    console.error('Failed to generate system diagram:', error);
    throw error;
  }
}

/**
 * Parse SDL content to extract components and their relationships
 */
function parseSDLComponents(sdlContent: string): Array<{
  name: string;
  varName: string;
  type: string;
  methods: string[];
  dependencies: Array<{varName: string; paramName: string}>;
}> {
  const components: Array<any> = [];
  
  // Simple SDL parser - this is a basic implementation
  // In a real implementation, you'd want to use a proper SDL parser
  
  // Match system declaration
  const systemMatch = sdlContent.match(/system\s+(\w+)\s*{([^}]+)}/s);
  if (!systemMatch) {
    return components;
  }
  
  const systemBody = systemMatch[2];
  
  // Match component declarations (use statements)
  const useRegex = /use\s+(\w+)\s+(\w+)(?:\s*\(([^)]+)\))?/g;
  let match;
  
  while ((match = useRegex.exec(systemBody)) !== null) {
    const varName = match[1];
    const componentType = match[2];
    const paramsStr = match[3] || '';
    
    // Parse dependencies from parameters
    const dependencies: Array<{varName: string; paramName: string}> = [];
    if (paramsStr) {
      const depRegex = /(\w+)\s*=\s*(\w+)/g;
      let depMatch;
      while ((depMatch = depRegex.exec(paramsStr)) !== null) {
        dependencies.push({
          paramName: depMatch[1],
          varName: depMatch[2]
        });
      }
    }
    
    // Determine component type and default methods
    let methods: string[] = [];
    let type = componentType;
    
    // Common SDL component types and their typical methods
    if (componentType === 'AppServer' || componentType === 'Server') {
      methods = ['handle', 'process'];
      type = 'server';
    } else if (componentType === 'Database' || componentType === 'DB') {
      methods = ['query', 'insert', 'update', 'delete'];
      type = 'database';
    } else if (componentType === 'Cache') {
      methods = ['get', 'set', 'delete'];
      type = 'cache';
    } else if (componentType === 'Gateway' || componentType === 'APIGateway') {
      methods = ['route', 'authenticate'];
      type = 'gateway';
    } else if (componentType === 'Queue' || componentType === 'MessageQueue') {
      methods = ['publish', 'consume'];
      type = 'queue';
    } else {
      // Default methods for unknown components
      methods = ['process'];
      type = 'service';
    }
    
    components.push({
      name: componentType,
      varName: varName,
      type: type,
      methods: methods,
      dependencies: dependencies
    });
  }
  
  return components;
}
