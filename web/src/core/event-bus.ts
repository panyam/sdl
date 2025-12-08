/**
 * EventBus - Re-export from @panyam/tsappkit for consistency
 *
 * This module re-exports the EventBus from tsappkit and defines
 * application-specific event names.
 */

import { EventBus } from '@panyam/tsappkit';
import type { EventHandler } from '@panyam/tsappkit';

// Re-export types and class from tsappkit
export { EventBus };
export type { EventHandler };

/**
 * Interface for EventBus - maintained for backward compatibility
 */
export interface IEventBus {
  emit(event: string, data?: any): void;
  on(event: string, handler: EventHandler): void;
  off(event: string, handler: EventHandler): void;
  once(event: string, handler: EventHandler): void;
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
