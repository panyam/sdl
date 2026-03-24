import { vi } from 'vitest';
import type { 
  SystemConfig, 
  SystemStatus, 
  FileRequest, 
  FileResponse,
  MetricsData,
  ListFileRequest,
  ListFileResponse,
  CommandRequest,
  CommandResponse
} from '@proto/canvas_pb';

export const mockSystemConfig: SystemConfig = {
  name: 'test-system',
  definition: 'system TestSystem { use app App }',
  trafficEnabled: false
} as SystemConfig;

export const mockSystemStatus: SystemStatus = {
  name: 'test-system',
  status: 'running',
  message: 'System is operational'
} as SystemStatus;

export const mockFileResponse: FileResponse = {
  content: 'test file content',
  error: ''
} as FileResponse;

export const mockListFileResponse: ListFileResponse = {
  files: [
    { name: 'test.sdl', isDirectory: false },
    { name: 'examples', isDirectory: true }
  ],
  error: ''
} as ListFileResponse;

export const mockMetricsData: MetricsData = {
  timestamp: Date.now(),
  metrics: [
    { name: 'requests', value: 100, unit: 'count' },
    { name: 'latency', value: 25.5, unit: 'ms' }
  ]
} as MetricsData;

export const mockCommandResponse: CommandResponse = {
  output: 'Command executed successfully',
  error: ''
} as CommandResponse;

export const createMockCanvasClient = () => ({
  getSystemConfig: vi.fn().mockResolvedValue(mockSystemConfig),
  getSystemStatus: vi.fn().mockResolvedValue(mockSystemStatus),
  readFile: vi.fn().mockResolvedValue(mockFileResponse),
  writeFile: vi.fn().mockResolvedValue(mockFileResponse),
  listFiles: vi.fn().mockResolvedValue(mockListFileResponse),
  createDirectory: vi.fn().mockResolvedValue(mockFileResponse),
  deleteFile: vi.fn().mockResolvedValue(mockFileResponse),
  executeCommand: vi.fn().mockResolvedValue(mockCommandResponse),
  
  // WebSocket methods
  connectWebSocket: vi.fn(),
  subscribeToMetrics: vi.fn(),
  subscribeToSystemEvents: vi.fn(),
  disconnect: vi.fn()
});

export const createMockWebSocket = () => {
  const listeners = new Map<string, Set<Function>>();
  
  return {
    send: vi.fn(),
    close: vi.fn(),
    addEventListener: vi.fn((event: string, handler: Function) => {
      if (!listeners.has(event)) {
        listeners.set(event, new Set());
      }
      listeners.get(event)!.add(handler);
    }),
    removeEventListener: vi.fn((event: string, handler: Function) => {
      listeners.get(event)?.delete(handler);
    }),
    dispatchEvent: vi.fn((event: string, data?: any) => {
      listeners.get(event)?.forEach(handler => handler(data));
    }),
    readyState: WebSocket.OPEN,
    CONNECTING: WebSocket.CONNECTING,
    OPEN: WebSocket.OPEN,
    CLOSING: WebSocket.CLOSING,
    CLOSED: WebSocket.CLOSED
  };
};