import { describe, it, expect, vi, beforeEach } from 'vitest';
import { EventBus } from '../../core/event-bus';

describe('EventBus', () => {
  let eventBus: EventBus;

  beforeEach(() => {
    eventBus = new EventBus();
  });

  describe('on', () => {
    it('should register event listeners', () => {
      const handler = vi.fn();
      eventBus.on('test-event', handler);
      
      eventBus.emit('test-event', { data: 'test' });
      
      expect(handler).toHaveBeenCalledWith({ data: 'test' });
      expect(handler).toHaveBeenCalledTimes(1);
    });

    it('should allow multiple listeners for the same event', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      
      eventBus.on('test-event', handler1);
      eventBus.on('test-event', handler2);
      
      eventBus.emit('test-event', 'data');
      
      expect(handler1).toHaveBeenCalledWith('data');
      expect(handler2).toHaveBeenCalledWith('data');
    });

    it('should handle listeners for different events independently', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      
      eventBus.on('event1', handler1);
      eventBus.on('event2', handler2);
      
      eventBus.emit('event1', 'data1');
      
      expect(handler1).toHaveBeenCalledWith('data1');
      expect(handler2).not.toHaveBeenCalled();
    });
  });

  describe('once', () => {
    it('should register a listener that only fires once', () => {
      const handler = vi.fn();
      eventBus.once('test-event', handler);
      
      eventBus.emit('test-event', 'first');
      eventBus.emit('test-event', 'second');
      
      expect(handler).toHaveBeenCalledTimes(1);
      expect(handler).toHaveBeenCalledWith('first');
    });

    it('should remove once listener after first emit', () => {
      const handler = vi.fn();
      eventBus.once('test-event', handler);
      
      eventBus.emit('test-event', 'data');
      eventBus.emit('test-event', 'data2');
      
      expect(handler).toHaveBeenCalledTimes(1);
      expect(handler).toHaveBeenCalledWith('data');
    });
  });

  describe('off', () => {
    it('should remove specific event listener', () => {
      const handler = vi.fn();
      eventBus.on('test-event', handler);
      
      eventBus.off('test-event', handler);
      eventBus.emit('test-event', 'data');
      
      expect(handler).not.toHaveBeenCalled();
    });

    it('should only remove the specified listener', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      
      eventBus.on('test-event', handler1);
      eventBus.on('test-event', handler2);
      
      eventBus.off('test-event', handler1);
      eventBus.emit('test-event', 'data');
      
      expect(handler1).not.toHaveBeenCalled();
      expect(handler2).toHaveBeenCalledWith('data');
    });

    it('should handle removing non-existent listener gracefully', () => {
      const handler = vi.fn();
      
      expect(() => {
        eventBus.off('test-event', handler);
      }).not.toThrow();
    });

    it('should only remove handlers that were added', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      const handler3 = vi.fn();
      
      eventBus.on('test-event', handler1);
      eventBus.on('test-event', handler2);
      
      // Try to remove a handler that was never added
      eventBus.off('test-event', handler3);
      
      eventBus.emit('test-event', 'data');
      
      expect(handler1).toHaveBeenCalledWith('data');
      expect(handler2).toHaveBeenCalledWith('data');
      expect(handler3).not.toHaveBeenCalled();
    });

    it('should remove once handlers when using off', () => {
      const handler = vi.fn();
      
      eventBus.once('test-event', handler);
      eventBus.off('test-event', handler);
      
      eventBus.emit('test-event', 'data');
      
      expect(handler).not.toHaveBeenCalled();
    });
  });

  describe('emit', () => {
    it('should call all registered listeners', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      
      eventBus.on('test-event', handler1);
      eventBus.on('test-event', handler2);
      
      eventBus.emit('test-event', 'data');
      
      expect(handler1).toHaveBeenCalledWith('data');
      expect(handler2).toHaveBeenCalledWith('data');
    });

    it('should not throw when no listeners exist', () => {
      expect(() => {
        eventBus.emit('test-event', 'data');
      }).not.toThrow();
    });

    it('should pass data to listeners', () => {
      const handler = vi.fn();
      eventBus.on('test-event', handler);
      
      const testData = { arg1: 'value1', arg2: 123 };
      eventBus.emit('test-event', testData);
      
      expect(handler).toHaveBeenCalledWith(testData);
    });

    it('should handle errors in listeners without affecting other listeners', () => {
      const errorHandler = vi.fn(() => {
        throw new Error('Handler error');
      });
      const successHandler = vi.fn();
      
      eventBus.on('test-event', errorHandler);
      eventBus.on('test-event', successHandler);
      
      expect(() => {
        eventBus.emit('test-event', 'data');
      }).not.toThrow();
      
      expect(errorHandler).toHaveBeenCalled();
      expect(successHandler).toHaveBeenCalled();
    });
  });

  describe('clear', () => {
    it('should remove all event handlers', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      const onceHandler = vi.fn();
      
      eventBus.on('event1', handler1);
      eventBus.on('event2', handler2);
      eventBus.once('event3', onceHandler);
      
      eventBus.clear();
      
      eventBus.emit('event1', 'data');
      eventBus.emit('event2', 'data');
      eventBus.emit('event3', 'data');
      
      expect(handler1).not.toHaveBeenCalled();
      expect(handler2).not.toHaveBeenCalled();
      expect(onceHandler).not.toHaveBeenCalled();
    });
  });
});