import { 
  APIResponse, 
  LoadRequest, 
  UseRequest, 
  SetRequest, 
  RunRequest, 
  PlotRequest,
  WebSocketMessage 
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

  // HTTP utility
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
}