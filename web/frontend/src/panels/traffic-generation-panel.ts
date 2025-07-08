import { BasePanel, PanelConfig } from './base-panel.js';
import { AppState } from '../core/app-state-manager.js';
import { AppEvents } from '../core/event-bus.js';
import { GenerateCall } from '../types.js';

/**
 * Panel for managing traffic generators
 */
export class TrafficGenerationPanel extends BasePanel {
  private generateCalls: GenerateCall[] = [];
  private isUpdatingGenerators = false;
  
  constructor(config: PanelConfig) {
    super({
      ...config,
      id: config.id || 'trafficGeneration',
      title: config.title || 'Traffic Generation'
    });
  }

  protected onInitialize(): void {
    // Listen for generator-specific events
    this.eventBus.on(AppEvents.GENERATOR_TOGGLED, this.handleGeneratorToggled);
  }

  protected onDispose(): void {
    this.eventBus.off(AppEvents.GENERATOR_TOGGLED, this.handleGeneratorToggled);
  }

  onStateChange(state: AppState, changedKeys: string[]): void {
    // Update generators if they changed
    if (changedKeys.includes('generateCalls')) {
      // Skip update if we're in the middle of updating to avoid UI overwrites
      if (!this.isUpdatingGenerators) {
        this.generateCalls = state.generateCalls || [];
        this.render();
      }
    }
  }

  protected render(): void {
    const content = this.renderGenerateControls();
    this.setContent(content);
    
    // Attach event listeners after render
    setTimeout(() => this.attachEventListeners(), 10);
  }

  private renderGenerateControls(): string {
    if (!this.generateCalls || this.generateCalls.length === 0) {
      return `
        <div class="h-full flex items-center justify-center bg-gray-50 dark:bg-gray-900">
          <div class="text-center text-gray-600 dark:text-gray-400">
            <div class="text-6xl mb-4">🚦</div>
            <div class="text-lg">No Traffic Generators</div>
            <div class="text-sm">Load a system to configure traffic generation</div>
          </div>
        </div>
      `;
    }

    const allEnabled = this.generateCalls.every(call => call.enabled);
    const toggleButtonText = allEnabled ? '⏹️ Stop All' : '▶️ Start All';
    
    return `
      <div class="space-y-4 p-4 bg-gray-50 dark:bg-gray-900">
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Traffic Generators</h3>
          <button id="toggle-all-generators" 
                  class="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm transition-colors">
            ${toggleButtonText}
          </button>
        </div>
        
        ${this.generateCalls.map(call => this.renderGeneratorControl(call)).join('')}
        
        <div class="mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
          <button id="add-generator" 
                  class="w-full py-2 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-900 dark:text-white rounded text-sm transition-colors">
            ➕ Add Generator
          </button>
        </div>
      </div>
    `;
  }

  private renderGeneratorControl(call: GenerateCall): string {
    const isEnabled = call.enabled;
    const toggleClass = isEnabled ? 'bg-green-600' : 'bg-gray-600';
    const toggleText = isEnabled ? 'ON' : 'OFF';
    
    return `
      <div class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700" data-generator-id="${call.id}">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center space-x-3">
            <button class="toggle-generator ${toggleClass} px-3 py-1 rounded text-xs font-semibold text-white transition-colors"
                    data-id="${call.id}">
              ${toggleText}
            </button>
            <span class="font-medium text-gray-900 dark:text-white">${call.name}</span>
          </div>
          <button class="remove-generator text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 transition-colors"
                  data-id="${call.id}">
            ✕
          </button>
        </div>
        
        <div class="text-sm text-gray-600 dark:text-gray-400 mb-2">Target: ${call.target}</div>
        
        <div class="flex items-center space-x-2">
          <label class="text-sm text-gray-700 dark:text-gray-300">Rate (RPS):</label>
          <input type="range" 
                 class="rate-slider flex-1" 
                 min="0" 
                 max="100" 
                 step="0.1"
                 value="${call.rate}"
                 data-id="${call.id}"
                 ${!isEnabled ? 'disabled' : ''}>
          <input type="number" 
                 class="rate-input w-20 px-2 py-1 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white"
                 min="0" 
                 max="1000" 
                 step="0.1"
                 value="${call.rate}"
                 data-id="${call.id}"
                 ${!isEnabled ? 'disabled' : ''}>
        </div>
      </div>
    `;
  }

  private attachEventListeners(): void {
    if (!this.container) return;

    // Toggle all generators
    const toggleAllBtn = this.container.querySelector('#toggle-all-generators');
    toggleAllBtn?.addEventListener('click', () => {
      this.eventBus.emit(AppEvents.TOOLBAR_ACTION, { action: 'toggleAllGenerators' });
    });

    // Add generator
    const addGeneratorBtn = this.container.querySelector('#add-generator');
    addGeneratorBtn?.addEventListener('click', () => {
      this.eventBus.emit(AppEvents.TOOLBAR_ACTION, { action: 'addGenerator' });
    });

    // Individual generator controls
    this.container.querySelectorAll('.toggle-generator').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const id = (e.target as HTMLElement).dataset.id;
        const generator = this.generateCalls.find(g => g.id === id);
        if (generator) {
          this.handleGenerateToggle(generator);
        }
      });
    });

    // Remove generator buttons
    this.container.querySelectorAll('.remove-generator').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const id = (e.target as HTMLElement).dataset.id;
        this.eventBus.emit(AppEvents.TOOLBAR_ACTION, { 
          action: 'removeGenerator', 
          generatorId: id 
        });
      });
    });

    // Rate sliders and inputs
    this.container.querySelectorAll('.rate-slider').forEach(slider => {
      slider.addEventListener('input', (e) => {
        const input = e.target as HTMLInputElement;
        const id = input.dataset.id;
        const value = parseFloat(input.value);
        
        // Update corresponding number input
        const numberInput = this.container?.querySelector(`.rate-input[data-id="${id}"]`) as HTMLInputElement;
        if (numberInput) {
          numberInput.value = value.toString();
        }
        
        // Debounce the actual update
        this.debounceRateUpdate(id!, value);
      });
    });

    this.container.querySelectorAll('.rate-input').forEach(input => {
      input.addEventListener('change', (e) => {
        const target = e.target as HTMLInputElement;
        const id = target.dataset.id;
        const value = parseFloat(target.value);
        
        // Update corresponding slider
        const slider = this.container?.querySelector(`.rate-slider[data-id="${id}"]`) as HTMLInputElement;
        if (slider) {
          slider.value = value.toString();
        }
        
        const generator = this.generateCalls.find(g => g.id === id);
        if (generator) {
          this.handleGenerateRateChange(generator, value);
        }
      });
    });
  }

  private rateUpdateTimer: { [id: string]: number } = {};
  
  private debounceRateUpdate(id: string, value: number): void {
    // Clear existing timer
    if (this.rateUpdateTimer[id]) {
      clearTimeout(this.rateUpdateTimer[id]);
    }
    
    // Set new timer
    this.rateUpdateTimer[id] = window.setTimeout(() => {
      const generator = this.generateCalls.find(g => g.id === id);
      if (generator) {
        this.handleGenerateRateChange(generator, value);
      }
    }, 300); // 300ms debounce
  }

  private handleGenerateToggle(generator: GenerateCall): void {
    this.isUpdatingGenerators = true;
    
    // Emit event for dashboard to handle
    this.eventBus.emit(AppEvents.GENERATOR_TOGGLED, {
      generator,
      enabled: !generator.enabled
    });
    
    // Update local state optimistically
    const index = this.generateCalls.findIndex(g => g.id === generator.id);
    if (index !== -1) {
      this.generateCalls[index] = { ...generator, enabled: !generator.enabled };
      this.render();
    }
    
    // Reset flag after a short delay
    setTimeout(() => {
      this.isUpdatingGenerators = false;
    }, 100);
  }

  private handleGenerateRateChange(generator: GenerateCall, newRate: number): void {
    this.isUpdatingGenerators = true;
    
    // Emit event for dashboard to handle
    this.eventBus.emit(AppEvents.GENERATOR_UPDATED, {
      generator,
      rate: newRate
    });
    
    // Reset flag after a short delay
    setTimeout(() => {
      this.isUpdatingGenerators = false;
    }, 100);
  }

  private handleGeneratorToggled = (_data: any) => {
    // This is called when generator toggle is confirmed
    // The state will be updated through AppStateManager
  };
}