import { CanvasClient } from '../canvas-client.js';
import { globalEventBus, AppEvents } from '../core/event-bus.js';
import { 
  CanvasState,
  RunResult
} from '../types.js';

export interface ICanvasService {
  loadFile(path: string): Promise<void>;
  useSystem(systemName: string): Promise<void>;
  createCanvas(): Promise<void>;
  getState(): Promise<CanvasState | null>;
  startAllGenerators(): Promise<void>;
  stopAllGenerators(): Promise<void>;
  addGenerator(name: string, component: string, method: string, rate: number): Promise<void>;
  updateGeneratorRate(name: string, rate: number): Promise<void>;
  deleteGenerator(name: string): Promise<void>;
  getGenerators(): Promise<Record<string, any>>;
  runSimulation(): Promise<Record<string, RunResult[]> | null>;
}

/**
 * Service for Canvas API operations
 */
export class CanvasService implements ICanvasService {
  private client: CanvasClient;
  
  constructor(canvasId: string = 'default') {
    this.client = new CanvasClient(canvasId);
  }

  async ensureCanvas(): Promise<void> {
    await this.client.ensureCanvas();
  }

  async loadFile(path: string): Promise<void> {
    try {
      globalEventBus.emit(AppEvents.FILE_SELECTED, { path });
      const response = await this.client.loadFile(path);
      
      if (response.success) {
        globalEventBus.emit(AppEvents.SYSTEM_LOADED, response.data);
      } else {
        throw new Error('Failed to load file');
      }
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to load file: ${error}`);
      throw error;
    }
  }

  async useSystem(systemName: string): Promise<void> {
    try {
      const response = await this.client.useSystem(systemName);
      
      if (response.success) {
        globalEventBus.emit(AppEvents.SYSTEM_LOADED, systemName);
      } else {
        throw new Error('Failed to use system');
      }
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to use system: ${error}`);
      throw error;
    }
  }

  async createCanvas(): Promise<void> {
    try {
      await this.client.createCanvas();
      
      // Canvas created successfully
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to create canvas: ${error}`);
      throw error;
    }
  }

  async getState(): Promise<CanvasState | null> {
    try {
      // TODO: Properly convert Canvas to CanvasState
      const canvas = await this.client.getState();
      return canvas as any;
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to get state: ${error}`);
      return null;
    }
  }

  async startAllGenerators(): Promise<void> {
    try {
      const response = await this.client.startAllGenerators();
      
      if (response.success) {
        globalEventBus.emit(AppEvents.SIMULATION_STARTED);
      } else {
        throw new Error('Failed to start generators');
      }
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to start generators: ${error}`);
      throw error;
    }
  }

  async stopAllGenerators(): Promise<void> {
    try {
      const response = await this.client.stopAllGenerators();
      
      if (response.success) {
        globalEventBus.emit(AppEvents.SIMULATION_STOPPED);
      } else {
        throw new Error('Failed to stop generators');
      }
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to stop generators: ${error}`);
      throw error;
    }
  }

  async addGenerator(name: string, component: string, method: string, rate: number): Promise<void> {
    try {
      const response = await this.client.addGenerator(name, component, method, rate);
      
      if (response.success) {
        globalEventBus.emit(AppEvents.GENERATOR_ADDED, { name, component, method, rate });
      } else {
        throw new Error('Failed to add generator');
      }
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to add generator: ${error}`);
      throw error;
    }
  }

  async updateGeneratorRate(name: string, rate: number): Promise<void> {
    try {
      const response = await this.client.updateGeneratorRate(name, rate);
      
      if (response.success) {
        globalEventBus.emit(AppEvents.GENERATOR_UPDATED, { name, rate });
      } else {
        throw new Error('Failed to update generator');
      }
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to update generator: ${error}`);
      throw error;
    }
  }

  async deleteGenerator(name: string): Promise<void> {
    try {
      const response = await this.client.deleteGenerator(name);
      
      if (response.success) {
        globalEventBus.emit(AppEvents.GENERATOR_REMOVED, { name });
      } else {
        throw new Error('Failed to delete generator');
      }
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to delete generator: ${error}`);
      throw error;
    }
  }

  async getGenerators(): Promise<Record<string, any>> {
    try {
      const response = await this.client.getGenerators();
      
      if (response.success && response.data) {
        return response.data;
      } else {
        throw new Error('Failed to get generators');
      }
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to get generators: ${error}`);
      throw error;
    }
  }

  async runSimulation(): Promise<Record<string, RunResult[]> | null> {
    try {
      // TODO: Implement runSimulation in CanvasClient
      return null;
    } catch (error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, `Failed to run simulation: ${error}`);
      return null;
    }
  }

  /**
   * Start metrics streaming
   */
  startMetricsStream(onMetrics: (metrics: any) => void): () => void {
    // TODO: Implement proper metrics streaming
    const controller = new AbortController();
    
    (async () => {
      try {
        const stream = this.client.streamMetrics([]);
        for await (const update of stream) {
          if (controller.signal.aborted) break;
          if (update.data) {
            onMetrics(update.data);
          }
        }
      } catch (error) {
        console.error('Metrics stream error:', error);
      }
    })();
    
    return () => controller.abort();
  }

  /**
   * Get the underlying client (for advanced usage)
   */
  getClient(): CanvasClient {
    return this.client;
  }
}