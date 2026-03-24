import type { LCMComponent, EventHandler } from '@panyam/tsappkit';
import type { IEventBus } from '../core/event-bus.js';
import type { IAppStateManager, AppState } from '../core/app-state-manager.js';

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
 * Abstract base class for panel components.
 *
 * Implements LCMComponent from tsappkit while bridging DockView's
 * initialize(container) pattern. Keeps SDL-specific stateManager integration
 * which will eventually be replaced by a presenter pattern.
 */
export abstract class BasePanel implements IPanelComponent, LCMComponent {
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

  // ============== LCMComponent Lifecycle ==============

  performLocalInit(): LCMComponent[] | Promise<LCMComponent[]> {
    return [];
  }

  setupDependencies(): void | Promise<void> {}

  activate(): void | Promise<void> {
    this.stateUnsubscribe = this.stateManager.subscribe((state, changedKeys) => {
      this.onStateChange(state, changedKeys);
    });
  }

  deactivate(): void | Promise<void> {
    if (this.stateUnsubscribe) {
      this.stateUnsubscribe();
      this.stateUnsubscribe = undefined;
    }
    this.onDispose();
    if (this.container) {
      this.container.innerHTML = '';
      this.container = null;
    }
  }

  // ============== DockView Integration ==============

  initialize(container: HTMLElement): void {
    this.container = container;
    const children = this.performLocalInit();
    if (children instanceof Promise) {
      children.then(() => {
        this.setupDependencies();
        this.activate();
        this.render();
        this.onInitialize();
      });
    } else {
      this.setupDependencies();
      this.activate();
      this.render();
      this.onInitialize();
    }
  }

  dispose(): void {
    this.deactivate();
  }

  // ============== Abstract Methods ==============

  abstract onStateChange(state: AppState, changedKeys: string[]): void;
  protected abstract render(): void;
  protected abstract onInitialize(): void;
  protected abstract onDispose(): void;

  onResize(): void {}

  // ============== Event Helpers ==============

  protected on(eventType: string, handler: EventHandler): void {
    this.eventBus.on(eventType, handler);
  }

  protected off(eventType: string, handler: EventHandler): void {
    this.eventBus.off(eventType, handler);
  }

  protected emit(eventType: string, data?: any): void {
    this.eventBus.emit(eventType, data);
  }

  // ============== DOM Helpers ==============

  protected setContent(html: string): void {
    if (this.container) {
      this.container.innerHTML = html;
    }
  }

  protected findElement<T extends HTMLElement = HTMLElement>(selector: string): T | null {
    return this.container?.querySelector<T>(selector) ?? null;
  }

  protected findElements<T extends HTMLElement = HTMLElement>(selector: string): T[] {
    return this.container ? Array.from(this.container.querySelectorAll<T>(selector)) : [];
  }

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

  protected showError(error: string): void {
    this.setContent(`
      <div class="flex items-center justify-center h-full">
        <div class="text-center">
          <div class="text-red-500 text-xl mb-2">Error</div>
          <div class="text-gray-400">${error}</div>
        </div>
      </div>
    `);
  }

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
