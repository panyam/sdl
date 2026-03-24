import { DashboardState } from '../types.js';
import { globalEventBus, AppEvents } from './event-bus.js';

export interface AppState extends DashboardState {
  // Additional state properties
  activePanel?: string;
  layoutConfig?: any;
  isLoading: boolean;
  error?: string;
}

export type StateListener = (state: AppState, changedKeys: string[]) => void;

export interface IAppStateManager {
  getState(): Readonly<AppState>;
  updateState(updates: Partial<AppState>): void;
  subscribe(listener: StateListener): () => void;
  reset(): void;
}

/**
 * Centralized state management with immutable updates and change notifications
 */
export class AppStateManager implements IAppStateManager {
  private state: AppState;
  private listeners: Set<StateListener> = new Set();
  private updateTimer: number | null = null;
  private pendingUpdates: Partial<AppState> = {};
  private changedKeys: Set<string> = new Set();

  constructor(initialState?: Partial<AppState>) {
    this.state = {
      isConnected: false,
      isLoading: false,
      simulationResults: {},
      metrics: {
        load: 0,
        latency: 0,
        successRate: 0
      },
      dynamicCharts: {},
      generateCalls: [],
      ...initialState
    };
  }

  /**
   * Get a read-only copy of the current state
   */
  getState(): Readonly<AppState> {
    return Object.freeze({ ...this.state });
  }

  /**
   * Update state with partial updates
   * Batches updates to avoid excessive re-renders
   */
  updateState(updates: Partial<AppState>): void {
    // Track which keys are being updated
    Object.keys(updates).forEach(key => this.changedKeys.add(key));
    
    // Merge updates
    this.pendingUpdates = {
      ...this.pendingUpdates,
      ...updates
    };

    // Batch updates using microtask
    if (!this.updateTimer) {
      this.updateTimer = window.setTimeout(() => {
        this.applyUpdates();
      }, 0);
    }
  }

  /**
   * Apply pending updates and notify listeners
   */
  private applyUpdates(): void {
    const changedKeys = Array.from(this.changedKeys);
    const oldState = this.state;
    
    // Create new state with updates
    this.state = {
      ...this.state,
      ...this.pendingUpdates
    };

    // Clear pending updates
    this.pendingUpdates = {};
    this.changedKeys.clear();
    this.updateTimer = null;

    // Notify listeners
    this.listeners.forEach(listener => {
      try {
        listener(this.getState(), changedKeys);
      } catch (error) {
        console.error('Error in state listener:', error);
      }
    });

    // Emit relevant events based on what changed
    this.emitStateChangeEvents(oldState, this.state, changedKeys);
  }

  /**
   * Subscribe to state changes
   * Returns unsubscribe function
   */
  subscribe(listener: StateListener): () => void {
    this.listeners.add(listener);
    
    // Return unsubscribe function
    return () => {
      this.listeners.delete(listener);
    };
  }

  /**
   * Reset state to initial values
   */
  reset(): void {
    this.updateState({
      isConnected: false,
      isLoading: false,
      currentFile: undefined,
      currentSystem: undefined,
      simulationResults: {},
      metrics: {
        load: 0,
        latency: 0,
        successRate: 0
      },
      dynamicCharts: {},
      generateCalls: [],
      error: undefined
    });
  }

  /**
   * Emit events based on state changes
   */
  private emitStateChangeEvents(oldState: AppState, newState: AppState, changedKeys: string[]): void {
    // System changes
    if (changedKeys.includes('currentSystem')) {
      if (newState.currentSystem && !oldState.currentSystem) {
        globalEventBus.emit(AppEvents.SYSTEM_LOADED, newState.currentSystem);
      } else if (!newState.currentSystem && oldState.currentSystem) {
        globalEventBus.emit(AppEvents.SYSTEM_CLEARED);
      }
    }

    // File changes
    if (changedKeys.includes('currentFile') && newState.currentFile !== oldState.currentFile) {
      globalEventBus.emit(AppEvents.FILE_SELECTED, {
        path: newState.currentFile,
        previousPath: oldState.currentFile
      });
    }

    // Generator changes
    if (changedKeys.includes('generateCalls')) {
      // Detect added/removed/updated generators
      const oldGenerators = new Map(oldState.generateCalls.map(g => [g.id, g]));
      const newGenerators = new Map(newState.generateCalls.map(g => [g.id, g]));

      // Check for additions
      newGenerators.forEach((gen, id) => {
        if (!oldGenerators.has(id)) {
          globalEventBus.emit(AppEvents.GENERATOR_ADDED, gen);
        } else {
          const oldGen = oldGenerators.get(id)!;
          if (JSON.stringify(oldGen) !== JSON.stringify(gen)) {
            globalEventBus.emit(AppEvents.GENERATOR_UPDATED, gen);
          }
        }
      });

      // Check for removals
      oldGenerators.forEach((gen, id) => {
        if (!newGenerators.has(id)) {
          globalEventBus.emit(AppEvents.GENERATOR_REMOVED, gen);
        }
      });
    }

    // Metrics changes
    if (changedKeys.includes('metrics') || changedKeys.includes('dynamicCharts')) {
      globalEventBus.emit(AppEvents.METRICS_UPDATED, {
        metrics: newState.metrics,
        charts: newState.dynamicCharts
      });
    }

    // Error changes
    if (changedKeys.includes('error') && newState.error) {
      globalEventBus.emit(AppEvents.ERROR_OCCURRED, newState.error);
    }
  }
}

// Singleton instance
export const appStateManager = new AppStateManager();