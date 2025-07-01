import { describe, it, expect, vi, beforeEach } from 'vitest';
import { AppStateManager } from '../../core/app-state-manager';
import { EventBus } from '../../core/event-bus';
import { spyOnEvents } from '../utils/test-helpers';

describe('AppStateManager', () => {
  let eventBus: EventBus;
  let stateManager: AppStateManager;
  let eventSpies: Map<string, any>;

  beforeEach(() => {
    eventBus = new EventBus();
    stateManager = new AppStateManager(eventBus);
    eventSpies = spyOnEvents(eventBus);
  });

  describe('SystemState', () => {
    it('should initialize with default system state', () => {
      const state = stateManager.getSystemState();
      
      expect(state).toEqual({
        config: null,
        status: null,
        error: null
      });
    });

    it('should update system config and emit event', () => {
      const config = {
        name: 'TestSystem',
        definition: 'system TestSystem { use app App }',
        trafficEnabled: true
      };
      
      stateManager.setSystemConfig(config);
      
      expect(stateManager.getSystemState().config).toEqual(config);
      expect(eventSpies).toHaveEmittedEventWith('system:config:changed', config);
    });

    it('should update system status and emit event', () => {
      const status = {
        name: 'TestSystem',
        status: 'running',
        message: 'System is operational'
      };
      
      stateManager.setSystemStatus(status);
      
      expect(stateManager.getSystemState().status).toEqual(status);
      expect(eventSpies).toHaveEmittedEventWith('system:status:changed', status);
    });

    it('should set system error and emit event', () => {
      const error = 'Connection failed';
      
      stateManager.setSystemError(error);
      
      expect(stateManager.getSystemState().error).toBe(error);
      expect(eventSpies).toHaveEmittedEventWith('system:error', error);
    });

    it('should clear system error', () => {
      stateManager.setSystemError('Some error');
      stateManager.setSystemError(null);
      
      expect(stateManager.getSystemState().error).toBeNull();
    });
  });

  describe('FileSystemState', () => {
    it('should update current path and emit event', () => {
      const path = '/test/path';
      
      stateManager.setCurrentPath(path);
      
      expect(stateManager.getFileSystemState().currentPath).toBe(path);
      expect(eventSpies).toHaveEmittedEventWith('filesystem:path:changed', path);
    });

    it('should update selected file and emit event', () => {
      const file = '/test/file.sdl';
      
      stateManager.setSelectedFile(file);
      
      expect(stateManager.getFileSystemState().selectedFile).toBe(file);
      expect(eventSpies).toHaveEmittedEventWith('filesystem:file:selected', file);
    });

    it('should add open file and emit event', () => {
      const file = '/test/file1.sdl';
      
      stateManager.addOpenFile(file);
      
      expect(stateManager.getFileSystemState().openFiles).toContain(file);
      expect(eventSpies).toHaveEmittedEventWith('filesystem:file:opened', file);
    });

    it('should not add duplicate open files', () => {
      const file = '/test/file.sdl';
      
      stateManager.addOpenFile(file);
      stateManager.addOpenFile(file);
      
      const openFiles = stateManager.getFileSystemState().openFiles;
      expect(openFiles.filter(f => f === file).length).toBe(1);
    });

    it('should remove open file and emit event', () => {
      const file = '/test/file.sdl';
      
      stateManager.addOpenFile(file);
      stateManager.removeOpenFile(file);
      
      expect(stateManager.getFileSystemState().openFiles).not.toContain(file);
      expect(eventSpies).toHaveEmittedEventWith('filesystem:file:closed', file);
    });
  });

  describe('MetricsState', () => {
    it('should update metrics data and emit event', () => {
      const metricsData = {
        timestamp: Date.now(),
        metrics: [
          { name: 'requests', value: 100, unit: 'count' },
          { name: 'latency', value: 25.5, unit: 'ms' }
        ]
      };
      
      stateManager.updateMetrics(metricsData);
      
      expect(stateManager.getMetricsState().data).toEqual([metricsData]);
      expect(eventSpies).toHaveEmittedEventWith('metrics:updated', metricsData);
    });

    it('should maintain metrics history up to max size', () => {
      const maxSize = 100; // Default max history size
      
      // Add more than max size
      for (let i = 0; i < maxSize + 10; i++) {
        stateManager.updateMetrics({
          timestamp: i,
          metrics: [{ name: 'test', value: i, unit: 'count' }]
        });
      }
      
      const history = stateManager.getMetricsState().data;
      expect(history.length).toBe(maxSize);
      expect(history[0].timestamp).toBe(10); // Oldest should be removed
    });

    it('should toggle metrics streaming and emit event', () => {
      stateManager.setMetricsStreaming(true);
      
      expect(stateManager.getMetricsState().isStreaming).toBe(true);
      expect(eventSpies).toHaveEmittedEventWith('metrics:streaming:changed', true);
      
      stateManager.setMetricsStreaming(false);
      
      expect(stateManager.getMetricsState().isStreaming).toBe(false);
      expect(eventSpies).toHaveEmittedEventWith('metrics:streaming:changed', false);
    });

    it('should clear metrics history', () => {
      stateManager.updateMetrics({
        timestamp: Date.now(),
        metrics: [{ name: 'test', value: 1, unit: 'count' }]
      });
      
      stateManager.clearMetrics();
      
      expect(stateManager.getMetricsState().data).toEqual([]);
    });
  });

  describe('UIState', () => {
    it('should update panel visibility and emit event', () => {
      stateManager.setPanelVisibility('metrics', true);
      
      expect(stateManager.getUIState().panelVisibility.metrics).toBe(true);
      expect(eventSpies).toHaveEmittedEventWith('ui:panel:visibility:changed', {
        panel: 'metrics',
        visible: true
      });
    });

    it('should update active panel and emit event', () => {
      stateManager.setActivePanel('architecture');
      
      expect(stateManager.getUIState().activePanel).toBe('architecture');
      expect(eventSpies).toHaveEmittedEventWith('ui:panel:active:changed', 'architecture');
    });

    it('should toggle theme and emit event', () => {
      stateManager.setTheme('light');
      
      expect(stateManager.getUIState().theme).toBe('light');
      expect(eventSpies).toHaveEmittedEventWith('ui:theme:changed', 'light');
    });
  });

  describe('State Persistence', () => {
    it('should save state to localStorage', () => {
      const mockLocalStorage = {
        setItem: vi.fn()
      };
      global.localStorage = mockLocalStorage as any;
      
      stateManager.setCurrentPath('/test');
      stateManager.saveState();
      
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
        'sdl-canvas-state',
        expect.any(String)
      );
    });

    it('should load state from localStorage', () => {
      const savedState = {
        filesystem: {
          currentPath: '/saved/path',
          selectedFile: null,
          openFiles: ['/file1.sdl']
        }
      };
      
      const mockLocalStorage = {
        getItem: vi.fn(() => JSON.stringify(savedState))
      };
      global.localStorage = mockLocalStorage as any;
      
      stateManager.loadState();
      
      expect(stateManager.getFileSystemState().currentPath).toBe('/saved/path');
      expect(stateManager.getFileSystemState().openFiles).toEqual(['/file1.sdl']);
    });

    it('should handle invalid saved state gracefully', () => {
      const mockLocalStorage = {
        getItem: vi.fn(() => 'invalid json')
      };
      global.localStorage = mockLocalStorage as any;
      
      expect(() => {
        stateManager.loadState();
      }).not.toThrow();
    });
  });

  describe('State Subscriptions', () => {
    it('should notify subscribers on state changes', () => {
      const subscriber = vi.fn();
      
      stateManager.subscribe('system', subscriber);
      stateManager.setSystemConfig({ name: 'Test' } as any);
      
      expect(subscriber).toHaveBeenCalledWith(stateManager.getSystemState());
    });

    it('should handle multiple subscribers', () => {
      const subscriber1 = vi.fn();
      const subscriber2 = vi.fn();
      
      stateManager.subscribe('filesystem', subscriber1);
      stateManager.subscribe('filesystem', subscriber2);
      
      stateManager.setCurrentPath('/test');
      
      expect(subscriber1).toHaveBeenCalled();
      expect(subscriber2).toHaveBeenCalled();
    });

    it('should unsubscribe correctly', () => {
      const subscriber = vi.fn();
      
      const unsubscribe = stateManager.subscribe('ui', subscriber);
      unsubscribe();
      
      stateManager.setTheme('light');
      
      expect(subscriber).not.toHaveBeenCalled();
    });
  });
});