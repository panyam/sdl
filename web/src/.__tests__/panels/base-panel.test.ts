import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BasePanel } from '../../panels/base-panel';
import { EventBus } from '../../core/event-bus';

// Create a concrete implementation for testing
class TestPanel extends BasePanel {
  constructor(container: HTMLElement, eventBus: EventBus) {
    super(container, eventBus, 'test-panel');
  }
  
  render(): void {
    this.container.innerHTML = '<div>Test Panel Content</div>';
  }
}

describe('BasePanel', () => {
  let container: HTMLElement;
  let eventBus: EventBus;
  let panel: TestPanel;

  beforeEach(() => {
    container = document.createElement('div');
    eventBus = new EventBus();
    panel = new TestPanel(container, eventBus);
  });

  describe('Initialization', () => {
    it('should initialize with correct properties', () => {
      expect(panel['panelId']).toBe('test-panel');
      expect(panel['container']).toBe(container);
      expect(panel['eventBus']).toBe(eventBus);
    });

    it('should add panel class to container', () => {
      expect(container.classList.contains('panel')).toBe(true);
    });

    it('should set panel id on container', () => {
      expect(container.id).toBe('test-panel');
    });
  });

  describe('Visibility', () => {
    it('should show panel', () => {
      panel.show();
      
      expect(container.style.display).not.toBe('none');
      expect(container.classList.contains('panel-visible')).toBe(true);
    });

    it('should hide panel', () => {
      panel.hide();
      
      expect(container.style.display).toBe('none');
      expect(container.classList.contains('panel-visible')).toBe(false);
    });

    it('should toggle visibility', () => {
      panel.toggle();
      expect(container.style.display).toBe('none');
      
      panel.toggle();
      expect(container.style.display).not.toBe('none');
    });

    it('should emit visibility events', () => {
      const visibilityHandler = vi.fn();
      eventBus.on('panel:visibility:changed', visibilityHandler);
      
      panel.show();
      expect(visibilityHandler).toHaveBeenCalledWith({
        panelId: 'test-panel',
        visible: true
      });
      
      panel.hide();
      expect(visibilityHandler).toHaveBeenCalledWith({
        panelId: 'test-panel',
        visible: false
      });
    });
  });

  describe('Event Subscription', () => {
    it('should subscribe to events', () => {
      const handler = vi.fn();
      panel.subscribeToEvent('test-event', handler);
      
      eventBus.emit('test-event', { data: 'test' });
      
      expect(handler).toHaveBeenCalledWith({ data: 'test' });
    });

    it('should track subscriptions for cleanup', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      
      panel.subscribeToEvent('event1', handler1);
      panel.subscribeToEvent('event2', handler2);
      
      expect(panel['subscriptions']).toHaveLength(2);
    });
  });

  describe('Event Emission', () => {
    it('should emit events through event bus', () => {
      const handler = vi.fn();
      eventBus.on('custom-event', handler);
      
      panel.emitEvent('custom-event', { value: 123 });
      
      expect(handler).toHaveBeenCalledWith({ value: 123 });
    });
  });

  describe('Cleanup', () => {
    it('should cleanup subscriptions on destroy', () => {
      const handler = vi.fn();
      panel.subscribeToEvent('test-event', handler);
      
      panel.destroy();
      
      eventBus.emit('test-event', 'data');
      expect(handler).not.toHaveBeenCalled();
    });

    it('should clear container on destroy', () => {
      panel.render();
      expect(container.innerHTML).not.toBe('');
      
      panel.destroy();
      expect(container.innerHTML).toBe('');
    });

    it('should call onDestroy hook', () => {
      const onDestroySpy = vi.spyOn(panel, 'onDestroy');
      
      panel.destroy();
      
      expect(onDestroySpy).toHaveBeenCalled();
    });
  });

  describe('Update', () => {
    it('should call onUpdate when update is called', () => {
      const onUpdateSpy = vi.spyOn(panel, 'onUpdate');
      const data = { test: 'data' };
      
      panel.update(data);
      
      expect(onUpdateSpy).toHaveBeenCalledWith(data);
    });
  });

  describe('Render', () => {
    it('should render content to container', () => {
      panel.render();
      
      expect(container.innerHTML).toContain('Test Panel Content');
    });
  });
});