import { IPanelComponent, PanelConfig } from './base-panel.js';
import { SystemArchitecturePanel } from './system-architecture-panel.js';
import { TrafficGenerationPanel } from './traffic-generation-panel.js';
import { LiveMetricsPanel } from './live-metrics-panel.js';
import { IEventBus } from '../core/event-bus.js';
import { IAppStateManager } from '../core/app-state-manager.js';

export type PanelType = 'systemArchitecture' | 'trafficGeneration' | 'liveMetrics' | 'measurements';

export interface PanelFactoryConfig {
  eventBus: IEventBus;
  stateManager: IAppStateManager;
  graphviz?: any; // Optional shared Graphviz instance
}

/**
 * Factory for creating panel component instances
 */
export class PanelComponentFactory {
  private config: PanelFactoryConfig;

  constructor(config: PanelFactoryConfig) {
    this.config = config;
  }

  /**
   * Create a panel component by type
   */
  createPanel(type: PanelType, customConfig?: Partial<PanelConfig>): IPanelComponent | null {
    const baseConfig: PanelConfig = {
      id: type,
      title: this.getDefaultTitle(type),
      eventBus: this.config.eventBus,
      stateManager: this.config.stateManager,
      ...customConfig
    };

    switch (type) {
      case 'systemArchitecture':
        return new SystemArchitecturePanel({
          ...baseConfig,
          graphviz: this.config.graphviz
        });

      case 'trafficGeneration':
        return new TrafficGenerationPanel(baseConfig);

      case 'liveMetrics':
        return new LiveMetricsPanel(baseConfig);

      case 'measurements':
        // TODO: Implement MeasurementsPanel
        console.warn('MeasurementsPanel not yet implemented');
        return null;

      default:
        console.error(`Unknown panel type: ${type}`);
        return null;
    }
  }

  /**
   * Get default title for panel type
   */
  private getDefaultTitle(type: PanelType): string {
    const titles: Record<PanelType, string> = {
      systemArchitecture: 'System Architecture',
      trafficGeneration: 'Traffic Generation',
      liveMetrics: 'Live Metrics',
      measurements: 'Measurements'
    };
    
    return titles[type] || type;
  }

  /**
   * Create all default panels
   */
  createDefaultPanels(): Map<string, IPanelComponent> {
    const panels = new Map<string, IPanelComponent>();
    
    const defaultTypes: PanelType[] = [
      'systemArchitecture',
      'trafficGeneration',
      'liveMetrics'
    ];

    defaultTypes.forEach(type => {
      const panel = this.createPanel(type);
      if (panel) {
        panels.set(type, panel);
      }
    });

    return panels;
  }
}