import { 
  APIResponse, 
  LoadRequest, 
  UseRequest, 
  SetRequest, 
  RunRequest, 
  PlotRequest,
  WebSocketMessage,
  GeneratorConfig,
  MeasurementConfig,
  CanvasState
} from './types.js';

const API_BASE = '/api';

export class CanvasAPI {
  private ws: WebSocket | null = null;
  private wsListeners: ((message: WebSocketMessage) => void)[] = [];

  constructor() {
    this.connectWebSocket();
  }

  // HTTP API methods
  async load(filePath: string): Promise<APIResponse> {
    const request: LoadRequest = { filePath };
    return this.post('/load', request);
  }

  async use(systemName: string): Promise<APIResponse> {
    const request: UseRequest = { systemName };
    return this.post('/use', request);
  }

  async set(path: string, value: any): Promise<APIResponse> {
    const request: SetRequest = { path, value };
    return this.post('/set', request);
  }

  async run(varName: string, target: string, runs: number = 1000): Promise<APIResponse> {
    const request: RunRequest = { varName, target, runs };
    return this.post('/run', request);
  }

  async plot(series: { name: string; from: string }[], outputFile: string, title: string = ''): Promise<APIResponse> {
    const request: PlotRequest = { series, outputFile, title };
    return this.post('/plot', request);
  }

  // RESTful Canvas API methods

  // Canvas state management
  async getState(): Promise<APIResponse<CanvasState>> {
    return this.get('/canvas/state');
  }

  async saveState(): Promise<APIResponse<CanvasState>> {
    return this.post('/canvas/state', {});
  }

  async restoreState(state: CanvasState): Promise<APIResponse> {
    return this.post('/canvas/state/restore', state);
  }

  // Generator management
  async getGenerators(): Promise<APIResponse<Record<string, GeneratorConfig>>> {
    return this.get('/canvas/generators');
  }

  async addGenerator(config: GeneratorConfig): Promise<APIResponse> {
    return this.post('/canvas/generators', config);
  }

  async getGenerator(id: string): Promise<APIResponse<GeneratorConfig>> {
    return this.get(`/canvas/generators/${id}`);
  }

  async updateGenerator(id: string, config: GeneratorConfig): Promise<APIResponse> {
    return this.put(`/canvas/generators/${id}`, config);
  }

  async removeGenerator(id: string): Promise<APIResponse> {
    return this.delete(`/canvas/generators/${id}`);
  }

  async pauseGenerator(id: string): Promise<APIResponse> {
    return this.post(`/canvas/generators/${id}/pause`, {});
  }

  async resumeGenerator(id: string): Promise<APIResponse> {
    return this.post(`/canvas/generators/${id}/resume`, {});
  }

  async startGenerators(): Promise<APIResponse> {
    return this.post('/canvas/generators/start', {});
  }

  async stopGenerators(): Promise<APIResponse> {
    return this.post('/canvas/generators/stop', {});
  }

  // Measurement management
  async getMeasurements(): Promise<APIResponse<Record<string, MeasurementConfig>>> {
    return this.get('/canvas/measurements');
  }

  async addMeasurement(config: MeasurementConfig): Promise<APIResponse> {
    return this.post('/canvas/measurements', config);
  }

  async getMeasurement(id: string): Promise<APIResponse<MeasurementConfig>> {
    return this.get(`/canvas/measurements/${id}`);
  }

  async updateMeasurement(id: string, config: MeasurementConfig): Promise<APIResponse> {
    return this.put(`/canvas/measurements/${id}`, config);
  }

  async removeMeasurement(id: string): Promise<APIResponse> {
    return this.delete(`/canvas/measurements/${id}`);
  }

  // WebSocket connection
  private connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/live`;
    
    this.ws = new WebSocket(wsUrl);
    
    this.ws.onopen = () => {
      console.log('🔌 WebSocket connected');
      this.notifyListeners({ type: 'systemActivated', connected: true } as WebSocketMessage);
    };
    
    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.notifyListeners(message);
      } catch (error) {
        console.error('❌ Failed to parse WebSocket message:', error);
      }
    };
    
    this.ws.onclose = () => {
      console.log('🔌 WebSocket disconnected');
      // Attempt to reconnect after 3 seconds
      setTimeout(() => this.connectWebSocket(), 3000);
    };
    
    this.ws.onerror = (error) => {
      console.error('❌ WebSocket error:', error);
    };
  }

  // WebSocket event listening
  onMessage(listener: (message: WebSocketMessage) => void) {
    this.wsListeners.push(listener);
  }

  private notifyListeners(message: WebSocketMessage) {
    this.wsListeners.forEach(listener => listener(message));
  }

  // HTTP utilities
  private async get(endpoint: string): Promise<APIResponse> {
    try {
      const response = await fetch(`${API_BASE}${endpoint}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      return await response.json();
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  private async post(endpoint: string, data: any): Promise<APIResponse> {
    try {
      const response = await fetch(`${API_BASE}${endpoint}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });

      return await response.json();
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  private async put(endpoint: string, data: any): Promise<APIResponse> {
    try {
      const response = await fetch(`${API_BASE}${endpoint}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });

      return await response.json();
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  private async delete(endpoint: string): Promise<APIResponse> {
    try {
      const response = await fetch(`${API_BASE}${endpoint}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      return await response.json();
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }
}