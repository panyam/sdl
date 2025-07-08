/**
 * EventBus provides a decoupled communication mechanism between components
 * using the publish-subscribe pattern.
 */
export interface IEventBus {
  emit(event: string, data?: any): void;
  on(event: string, handler: EventHandler): void;
  off(event: string, handler: EventHandler): void;
  once(event: string, handler: EventHandler): void;
}

export type EventHandler = (data?: any) => void;

export class EventBus implements IEventBus {
  private events: Map<string, Set<EventHandler>> = new Map();
  private onceHandlers: Map<string, Set<EventHandler>> = new Map();

  /**
   * Emit an event with optional data
   */
  emit(event: string, data?: any): void {
    // Regular handlers
    const handlers = this.events.get(event);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          console.error(`Error in event handler for '${event}':`, error);
        }
      });
    }

    // One-time handlers
    const onceHandlers = this.onceHandlers.get(event);
    if (onceHandlers) {
      onceHandlers.forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          console.error(`Error in once handler for '${event}':`, error);
        }
      });
      // Clear once handlers after execution
      this.onceHandlers.delete(event);
    }
  }

  /**
   * Subscribe to an event
   */
  on(event: string, handler: EventHandler): void {
    if (!this.events.has(event)) {
      this.events.set(event, new Set());
    }
    this.events.get(event)!.add(handler);
  }

  /**
   * Unsubscribe from an event
   */
  off(event: string, handler: EventHandler): void {
    const handlers = this.events.get(event);
    if (handlers) {
      handlers.delete(handler);
      if (handlers.size === 0) {
        this.events.delete(event);
      }
    }

    // Also check once handlers
    const onceHandlers = this.onceHandlers.get(event);
    if (onceHandlers) {
      onceHandlers.delete(handler);
      if (onceHandlers.size === 0) {
        this.onceHandlers.delete(event);
      }
    }
  }

  /**
   * Subscribe to an event for one-time execution
   */
  once(event: string, handler: EventHandler): void {
    if (!this.onceHandlers.has(event)) {
      this.onceHandlers.set(event, new Set());
    }
    this.onceHandlers.get(event)!.add(handler);
  }

  /**
   * Clear all event handlers
   */
  clear(): void {
    this.events.clear();
    this.onceHandlers.clear();
  }
}

// Common event names used across the application
export const AppEvents = {
  // File events
  FILE_SELECTED: 'file:selected',
  FILE_CREATED: 'file:created',
  FILE_SAVED: 'file:saved',
  FILE_DELETED: 'file:deleted',
  FILE_MODIFIED: 'file:modified',
  
  // System events
  SYSTEM_LOADED: 'system:loaded',
  SYSTEM_CLEARED: 'system:cleared',
  SIMULATION_STARTED: 'simulation:started',
  SIMULATION_STOPPED: 'simulation:stopped',
  
  // Generator events
  GENERATOR_ADDED: 'generator:added',
  GENERATOR_REMOVED: 'generator:removed',
  GENERATOR_UPDATED: 'generator:updated',
  GENERATOR_TOGGLED: 'generator:toggled',
  
  // Metrics events
  METRICS_UPDATED: 'metrics:updated',
  CHART_CREATED: 'chart:created',
  CHART_REMOVED: 'chart:removed',
  
  // Recipe events
  RECIPE_STARTED: 'recipe:started',
  RECIPE_STOPPED: 'recipe:stopped',
  RECIPE_STEP: 'recipe:step',
  RECIPE_COMPLETED: 'recipe:completed',
  RECIPE_STEP_EXECUTED: 'recipe:step:executed',
  RECIPE_ERROR: 'recipe:error',
  
  // UI events
  PANEL_RESIZED: 'ui:panel:resized',
  LAYOUT_CHANGED: 'ui:layout:changed',
  TOOLBAR_ACTION: 'ui:toolbar:action',
  
  // Error events
  ERROR_OCCURRED: 'error:occurred',
  WARNING_ISSUED: 'warning:issued',
} as const;

// Singleton instance
export const globalEventBus = new EventBus();