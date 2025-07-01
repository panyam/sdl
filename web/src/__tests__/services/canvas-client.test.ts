import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { CanvasClient } from '../../canvas-client';
import { createMockWebSocket } from '../mocks/canvas-api';

// Mock the ConnectRPC transport
vi.mock('@connectrpc/connect-web', () => ({
  createConnectTransport: vi.fn(() => ({}))
}));

// Mock the generated proto client
vi.mock('@proto/canvas_connect', () => ({
  createPromiseClient: vi.fn(() => ({
    getSystemConfig: vi.fn(),
    getSystemStatus: vi.fn(),
    readFile: vi.fn(),
    writeFile: vi.fn(),
    listFiles: vi.fn(),
    createDirectory: vi.fn(),
    deleteFile: vi.fn(),
    executeCommand: vi.fn()
  }))
}));

describe('CanvasClient', () => {
  let client: CanvasClient;
  let mockWebSocket: any;

  beforeEach(() => {
    mockWebSocket = createMockWebSocket();
    global.WebSocket = vi.fn(() => mockWebSocket) as any;
    client = new CanvasClient('http://localhost:8080');
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('System Operations', () => {
    it('should get system config', async () => {
      const mockConfig = {
        name: 'TestSystem',
        definition: 'system TestSystem { use app App }',
        trafficEnabled: true
      };
      
      client['client'].getSystemConfig = vi.fn().mockResolvedValue(mockConfig);
      
      const result = await client.getSystemConfig();
      
      expect(result).toEqual(mockConfig);
      expect(client['client'].getSystemConfig).toHaveBeenCalled();
    });

    it('should get system status', async () => {
      const mockStatus = {
        name: 'TestSystem',
        status: 'running',
        message: 'System is operational'
      };
      
      client['client'].getSystemStatus = vi.fn().mockResolvedValue(mockStatus);
      
      const result = await client.getSystemStatus();
      
      expect(result).toEqual(mockStatus);
      expect(client['client'].getSystemStatus).toHaveBeenCalled();
    });

    it('should handle system config errors', async () => {
      const error = new Error('Failed to get config');
      client['client'].getSystemConfig = vi.fn().mockRejectedValue(error);
      
      await expect(client.getSystemConfig()).rejects.toThrow('Failed to get config');
    });
  });

  describe('File Operations', () => {
    it('should read file', async () => {
      const mockResponse = {
        content: 'file content',
        error: ''
      };
      
      client['client'].readFile = vi.fn().mockResolvedValue(mockResponse);
      
      const result = await client.readFile({ path: '/test.sdl' });
      
      expect(result).toEqual(mockResponse);
      expect(client['client'].readFile).toHaveBeenCalledWith({ path: '/test.sdl' });
    });

    it('should write file', async () => {
      const mockResponse = {
        content: '',
        error: ''
      };
      
      client['client'].writeFile = vi.fn().mockResolvedValue(mockResponse);
      
      const result = await client.writeFile({ 
        path: '/test.sdl', 
        content: 'new content' 
      });
      
      expect(result).toEqual(mockResponse);
      expect(client['client'].writeFile).toHaveBeenCalledWith({
        path: '/test.sdl',
        content: 'new content'
      });
    });

    it('should list files', async () => {
      const mockResponse = {
        files: [
          { name: 'file1.sdl', isDirectory: false },
          { name: 'dir1', isDirectory: true }
        ],
        error: ''
      };
      
      client['client'].listFiles = vi.fn().mockResolvedValue(mockResponse);
      
      const result = await client.listFiles({ path: '/' });
      
      expect(result).toEqual(mockResponse);
      expect(client['client'].listFiles).toHaveBeenCalledWith({ path: '/' });
    });

    it('should create directory', async () => {
      const mockResponse = {
        content: '',
        error: ''
      };
      
      client['client'].createDirectory = vi.fn().mockResolvedValue(mockResponse);
      
      const result = await client.createDirectory({ path: '/newdir' });
      
      expect(result).toEqual(mockResponse);
      expect(client['client'].createDirectory).toHaveBeenCalledWith({ path: '/newdir' });
    });

    it('should delete file', async () => {
      const mockResponse = {
        content: '',
        error: ''
      };
      
      client['client'].deleteFile = vi.fn().mockResolvedValue(mockResponse);
      
      const result = await client.deleteFile({ path: '/test.sdl' });
      
      expect(result).toEqual(mockResponse);
      expect(client['client'].deleteFile).toHaveBeenCalledWith({ path: '/test.sdl' });
    });

    it('should handle file operation errors', async () => {
      const mockResponse = {
        content: '',
        error: 'File not found'
      };
      
      client['client'].readFile = vi.fn().mockResolvedValue(mockResponse);
      
      const result = await client.readFile({ path: '/nonexistent.sdl' });
      
      expect(result.error).toBe('File not found');
    });
  });

  describe('Command Execution', () => {
    it('should execute command', async () => {
      const mockResponse = {
        output: 'Command output',
        error: ''
      };
      
      client['client'].executeCommand = vi.fn().mockResolvedValue(mockResponse);
      
      const result = await client.executeCommand({ 
        command: 'echo test' 
      });
      
      expect(result).toEqual(mockResponse);
      expect(client['client'].executeCommand).toHaveBeenCalledWith({
        command: 'echo test'
      });
    });

    it('should handle command execution errors', async () => {
      const mockResponse = {
        output: '',
        error: 'Command failed'
      };
      
      client['client'].executeCommand = vi.fn().mockResolvedValue(mockResponse);
      
      const result = await client.executeCommand({ 
        command: 'invalid-command' 
      });
      
      expect(result.error).toBe('Command failed');
    });
  });

  describe('WebSocket Connection', () => {
    it('should connect WebSocket', () => {
      client.connectWebSocket();
      
      expect(global.WebSocket).toHaveBeenCalledWith('ws://localhost:8080/ws');
      expect(mockWebSocket.addEventListener).toHaveBeenCalledWith('open', expect.any(Function));
      expect(mockWebSocket.addEventListener).toHaveBeenCalledWith('message', expect.any(Function));
      expect(mockWebSocket.addEventListener).toHaveBeenCalledWith('error', expect.any(Function));
      expect(mockWebSocket.addEventListener).toHaveBeenCalledWith('close', expect.any(Function));
    });

    it('should handle WebSocket open event', () => {
      const openHandler = vi.fn();
      client.on('connected', openHandler);
      
      client.connectWebSocket();
      
      // Simulate open event
      const openEventHandler = mockWebSocket.addEventListener.mock.calls.find(
        call => call[0] === 'open'
      )[1];
      openEventHandler();
      
      expect(openHandler).toHaveBeenCalled();
    });

    it('should handle WebSocket message event', () => {
      const messageHandler = vi.fn();
      client.on('message', messageHandler);
      
      client.connectWebSocket();
      
      // Simulate message event
      const messageEventHandler = mockWebSocket.addEventListener.mock.calls.find(
        call => call[0] === 'message'
      )[1];
      
      const testMessage = { type: 'metrics', data: { value: 100 } };
      messageEventHandler({ data: JSON.stringify(testMessage) });
      
      expect(messageHandler).toHaveBeenCalledWith(testMessage);
    });

    it('should handle WebSocket error event', () => {
      const errorHandler = vi.fn();
      client.on('error', errorHandler);
      
      client.connectWebSocket();
      
      // Simulate error event
      const errorEventHandler = mockWebSocket.addEventListener.mock.calls.find(
        call => call[0] === 'error'
      )[1];
      
      const error = new Error('Connection failed');
      errorEventHandler(error);
      
      expect(errorHandler).toHaveBeenCalledWith(error);
    });

    it('should handle WebSocket close event', () => {
      const closeHandler = vi.fn();
      client.on('disconnected', closeHandler);
      
      client.connectWebSocket();
      
      // Simulate close event
      const closeEventHandler = mockWebSocket.addEventListener.mock.calls.find(
        call => call[0] === 'close'
      )[1];
      closeEventHandler();
      
      expect(closeHandler).toHaveBeenCalled();
    });

    it('should disconnect WebSocket', () => {
      client.connectWebSocket();
      client.disconnect();
      
      expect(mockWebSocket.close).toHaveBeenCalled();
    });

    it('should not error when disconnecting without connection', () => {
      expect(() => {
        client.disconnect();
      }).not.toThrow();
    });
  });

  describe('Metrics Subscription', () => {
    it('should subscribe to metrics', () => {
      client.connectWebSocket();
      client.subscribeToMetrics();
      
      expect(mockWebSocket.send).toHaveBeenCalledWith(
        JSON.stringify({ type: 'subscribe', channel: 'metrics' })
      );
    });

    it('should handle metrics data', () => {
      const metricsHandler = vi.fn();
      client.on('metrics', metricsHandler);
      
      client.connectWebSocket();
      
      // Simulate metrics message
      const messageEventHandler = mockWebSocket.addEventListener.mock.calls.find(
        call => call[0] === 'message'
      )[1];
      
      const metricsData = {
        type: 'metrics',
        data: {
          timestamp: Date.now(),
          metrics: [{ name: 'requests', value: 100 }]
        }
      };
      
      messageEventHandler({ data: JSON.stringify(metricsData) });
      
      expect(metricsHandler).toHaveBeenCalledWith(metricsData.data);
    });
  });

  describe('System Events Subscription', () => {
    it('should subscribe to system events', () => {
      client.connectWebSocket();
      client.subscribeToSystemEvents();
      
      expect(mockWebSocket.send).toHaveBeenCalledWith(
        JSON.stringify({ type: 'subscribe', channel: 'system' })
      );
    });

    it('should handle system event data', () => {
      const systemHandler = vi.fn();
      client.on('system-event', systemHandler);
      
      client.connectWebSocket();
      
      // Simulate system event message
      const messageEventHandler = mockWebSocket.addEventListener.mock.calls.find(
        call => call[0] === 'message'
      )[1];
      
      const systemEvent = {
        type: 'system-event',
        data: {
          event: 'status-changed',
          status: 'running'
        }
      };
      
      messageEventHandler({ data: JSON.stringify(systemEvent) });
      
      expect(systemHandler).toHaveBeenCalledWith(systemEvent.data);
    });
  });
});