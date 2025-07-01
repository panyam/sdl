import { IEventBus } from '../core/event-bus.js';
import { IAppStateManager, AppState } from '../core/app-state-manager.js';

/**
 * Base interface for all panel components
 */
export interface IPanelComponent {
  readonly id: string;
  readonly title: string;
  
  initialize(container: HTMLElement): void;
  dispose(): void;
  onStateChange(state: AppState, changedKeys: string[]): void;
  onResize?(): void;
}

/**
 * Configuration for panel components
 */
export interface PanelConfig {
  id: string;
  title: string;
  eventBus: IEventBus;
  stateManager: IAppStateManager;
}

/**
 * Abstract base class for panel components
 */
export abstract class BasePanel implements IPanelComponent {
  public readonly id: string;
  public readonly title: string;
  
  protected container: HTMLElement | null = null;
  protected eventBus: IEventBus;
  protected stateManager: IAppStateManager;
  protected stateUnsubscribe?: () => void;
  
  constructor(config: PanelConfig) {
    this.id = config.id;
    this.title = config.title;
    this.eventBus = config.eventBus;
    this.stateManager = config.stateManager;
  }

  /**
   * Initialize the panel with its container
   */
  initialize(container: HTMLElement): void {
    this.container = container;
    
    // Subscribe to state changes
    this.stateUnsubscribe = this.stateManager.subscribe((state, changedKeys) => {
      this.onStateChange(state, changedKeys);
    });
    
    // Initial render
    this.render();
    
    // Panel-specific initialization
    this.onInitialize();
  }

  /**
   * Clean up resources
   */
  dispose(): void {
    // Unsubscribe from state
    if (this.stateUnsubscribe) {
      this.stateUnsubscribe();
      this.stateUnsubscribe = undefined;
    }
    
    // Panel-specific cleanup
    this.onDispose();
    
    // Clear container
    if (this.container) {
      this.container.innerHTML = '';
      this.container = null;
    }
  }

  /**
   * Handle state changes
   */
  abstract onStateChange(state: AppState, changedKeys: string[]): void;

  /**
   * Render the panel content
   */
  protected abstract render(): void;

  /**
   * Panel-specific initialization
   */
  protected abstract onInitialize(): void;

  /**
   * Panel-specific cleanup
   */
  protected abstract onDispose(): void;

  /**
   * Handle resize events
   */
  onResize(): void {
    // Override in subclasses if needed
  }

  /**
   * Helper to safely update container HTML
   */
  protected setContent(html: string): void {
    if (this.container) {
      this.container.innerHTML = html;
    }
  }

  /**
   * Helper to show loading state
   */
  protected showLoading(message: string = 'Loading...'): void {
    this.setContent(`
      <div class="flex items-center justify-center h-full">
        <div class="text-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <div class="text-gray-400">${message}</div>
        </div>
      </div>
    `);
  }

  /**
   * Helper to show error state
   */
  protected showError(error: string): void {
    this.setContent(`
      <div class="flex items-center justify-center h-full">
        <div class="text-center">
          <div class="text-red-500 text-xl mb-2">⚠️ Error</div>
          <div class="text-gray-400">${error}</div>
        </div>
      </div>
    `);
  }

  /**
   * Helper to show empty state
   */
  protected showEmpty(message: string, subMessage?: string): void {
    this.setContent(`
      <div class="flex items-center justify-center h-full">
        <div class="text-center text-gray-400">
          <div class="text-xl mb-2">${message}</div>
          ${subMessage ? `<div class="text-sm">${subMessage}</div>` : ''}
        </div>
      </div>
    `);
  }
}