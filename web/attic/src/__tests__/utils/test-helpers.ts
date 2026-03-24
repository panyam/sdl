import { vi } from 'vitest';
import type { Mock } from 'vitest';
import { EventBus } from '../../core/event-bus';
import { AppStateManager } from '../../core/app-state-manager';

/**
 * Wait for a condition to be true
 */
export async function waitFor(
  condition: () => boolean, 
  timeout: number = 1000,
  interval: number = 10
): Promise<void> {
  const startTime = Date.now();
  
  while (!condition()) {
    if (Date.now() - startTime > timeout) {
      throw new Error('Timeout waiting for condition');
    }
    await new Promise(resolve => setTimeout(resolve, interval));
  }
}

/**
 * Wait for an event to be emitted
 */
export async function waitForEvent<T = any>(
  eventBus: EventBus,
  eventName: string,
  timeout: number = 1000
): Promise<T> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error(`Timeout waiting for event: ${eventName}`));
    }, timeout);
    
    const handler = (data: T) => {
      clearTimeout(timer);
      resolve(data);
    };
    
    eventBus.once(eventName, handler);
  });
}

/**
 * Create a mock DOM element
 */
export function createMockElement(tag: string = 'div'): HTMLElement {
  const element = document.createElement(tag);
  element.getBoundingClientRect = vi.fn(() => ({
    x: 0,
    y: 0,
    width: 100,
    height: 100,
    top: 0,
    right: 100,
    bottom: 100,
    left: 0,
    toJSON: () => {}
  }));
  return element;
}

/**
 * Create a test EventBus instance
 */
export function createTestEventBus(): EventBus {
  return new EventBus();
}

/**
 * Create a test AppStateManager instance
 */
export function createTestStateManager(): AppStateManager {
  const eventBus = createTestEventBus();
  return new AppStateManager(eventBus);
}

/**
 * Spy on EventBus emissions
 */
export function spyOnEvents(eventBus: EventBus): Map<string, Mock> {
  const spies = new Map<string, Mock>();
  const originalEmit = eventBus.emit.bind(eventBus);
  
  eventBus.emit = vi.fn((event: string, ...args: any[]) => {
    if (!spies.has(event)) {
      spies.set(event, vi.fn());
    }
    spies.get(event)!(...args);
    return originalEmit(event, ...args);
  });
  
  return spies;
}

/**
 * Create a mock file structure
 */
export interface MockFile {
  name: string;
  content?: string;
  isDirectory?: boolean;
  children?: MockFile[];
}

export function createMockFileStructure(files: MockFile[]): Map<string, MockFile> {
  const fileMap = new Map<string, MockFile>();
  
  function addFiles(parentPath: string, files: MockFile[]) {
    files.forEach(file => {
      const path = parentPath ? `${parentPath}/${file.name}` : file.name;
      fileMap.set(path, file);
      
      if (file.children) {
        addFiles(path, file.children);
      }
    });
  }
  
  addFiles('', files);
  return fileMap;
}

/**
 * Custom matchers for Canvas-specific assertions
 */
export const customMatchers = {
  toHaveEmittedEvent(received: Map<string, Mock>, eventName: string) {
    const spy = received.get(eventName);
    const pass = spy && spy.mock.calls.length > 0;
    
    return {
      pass,
      message: () => pass
        ? `Expected not to have emitted event "${eventName}"`
        : `Expected to have emitted event "${eventName}"`
    };
  },
  
  toHaveEmittedEventWith(received: Map<string, Mock>, eventName: string, data: any) {
    const spy = received.get(eventName);
    const pass = spy && spy.mock.calls.some(call => 
      JSON.stringify(call[0]) === JSON.stringify(data)
    );
    
    return {
      pass,
      message: () => pass
        ? `Expected not to have emitted event "${eventName}" with data ${JSON.stringify(data)}`
        : `Expected to have emitted event "${eventName}" with data ${JSON.stringify(data)}`
    };
  }
};

// Extend Vitest matchers
declare module 'vitest' {
  interface Assertion<T = any> {
    toHaveEmittedEvent(eventName: string): void;
    toHaveEmittedEventWith(eventName: string, data: any): void;
  }
}

expect.extend(customMatchers);