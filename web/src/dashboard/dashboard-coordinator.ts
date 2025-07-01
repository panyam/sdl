import { DockviewApi, DockviewComponent } from 'dockview-core';
import { globalEventBus, AppEvents } from '../core/event-bus.js';
import { appStateManager } from '../core/app-state-manager.js';
import { PanelComponentFactory } from '../panels/panel-factory.js';
import { CanvasService } from '../services/canvas-service.js';
import { IPanelComponent } from '../panels/base-panel.js';
import { Toolbar } from '../components/toolbar.js';
import { MultiFSExplorer } from '../components/multi-fs-explorer.js';
import { ConsolePanel } from '../components/console-panel.js';

/**
 * Slim coordinator that wires together all components
 */
export class DashboardCoordinator {
  private canvasId: string;
  private dockview: DockviewApi | null = null;
  private panels: Map<string, IPanelComponent> = new Map();
  private canvasService: CanvasService;
  private panelFactory: PanelComponentFactory;
  
  // Components (these will eventually be refactored too)
  private toolbar: Toolbar | null = null;
  private fileExplorer: MultiFSExplorer | null = null;

  constructor(canvasId: string = 'default') {
    this.canvasId = canvasId;
    this.canvasService = new CanvasService(canvasId);
    
    this.panelFactory = new PanelComponentFactory({
      eventBus: globalEventBus,
      stateManager: appStateManager
    });
  }

  async initialize(): Promise<void> {
    try {
      // Ensure canvas exists
      await this.canvasService.ensureCanvas();
      
      // Update document title
      document.title = `SDL Canvas - ${this.canvasId}`;
      
      // Render layout
      this.render();
      
      // Load initial state
      await this.loadCanvasState();
      
      // Set up event listeners
      this.setupEventListeners();
      
    } catch (error) {
      console.error('❌ Failed to initialize dashboard:', error);
      appStateManager.updateState({ 
        error: `Failed to initialize: ${error}` 
      });
    }
  }

  private render(): void {
    const app = document.getElementById('app');
    if (!app) return;

    app.innerHTML = `
      <div class="flex flex-col h-screen">
        <!-- Header with Toolbar -->
        <header class="bg-gray-800 border-b border-gray-700 px-4 py-2">
          <div class="flex items-center justify-between">
            <h1 class="text-xl font-bold">SDL Canvas: ${this.canvasId}</h1>
            <span class="text-xs text-gray-400">SDL Dashboard</span>
          </div>
        </header>
        
        <!-- Toolbar -->
        <div id="toolbar-container" class="bg-gray-800 border-b border-gray-700"></div>
        
        <!-- Main Layout -->
        <div id="dockview-container" class="flex-1"></div>
      </div>
    `;

    // Initialize toolbar
    this.initializeToolbar();
    
    // Initialize layout
    this.initializeLayout();
  }

  private initializeToolbar(): void {
    const container = document.getElementById('toolbar-container');
    if (!container) return;

    this.toolbar = new Toolbar(container);
    this.toolbar.setButtons([
      {
        id: 'load',
        label: 'Load',
        icon: '📂',
        tooltip: 'Load current file into Canvas',
        onClick: () => this.handleLoad()
      },
      {
        id: 'save',
        label: 'Save',
        icon: '💾',
        tooltip: 'Save current file',
        disabled: true,
        onClick: () => this.handleSave()
      },
      {
        id: 'run',
        label: 'Run',
        icon: '▶️',
        tooltip: 'Run simulation',
        disabled: true,
        onClick: () => this.handleRun()
      },
      {
        id: 'stop',
        label: 'Stop',
        icon: '⏹️',
        tooltip: 'Stop simulation',
        disabled: true,
        onClick: () => this.handleStop()
      }
    ]);
  }

  private initializeLayout(): void {
    const container = document.getElementById('dockview-container');
    if (!container) return;

    container.className = 'dockview-theme-dark flex-1';
    
    const dockviewComponent = new DockviewComponent(container, {
      createComponent: (options: any) => {
        // Handle panel components
        const panelTypes = ['systemArchitecture', 'trafficGeneration', 'liveMetrics'];
        if (panelTypes.includes(options.name)) {
          const panel = this.panelFactory.createPanel(options.name as any);
          if (panel) {
            this.panels.set(options.name, panel);
            const element = document.createElement('div');
            element.className = 'h-full p-4 overflow-auto';
            panel.initialize(element);
            
            return {
              element,
              init: () => {},
              dispose: () => {
                panel.dispose();
                this.panels.delete(options.name);
              }
            };
          }
        }
        
        // Handle other components (to be refactored)
        switch (options.name) {
          case 'fileExplorer':
            return this.createFileExplorerComponent();
          case 'codeEditor':
            return this.createCodeEditorComponent();
          case 'console':
            return this.createConsoleComponent();
        }
        
        // Fallback
        const element = document.createElement('div');
        element.innerHTML = `<div class="p-4">Unknown component: ${options.name}</div>`;
        return { element, init: () => {}, dispose: () => {} };
      }
    });

    this.dockview = dockviewComponent.api;
    this.createDefaultLayout();
  }

  private createDefaultLayout(): void {
    if (!this.dockview) return;

    // Add file explorer
    this.dockview.addPanel({
      id: 'fileExplorer',
      component: 'fileExplorer',
      title: 'Files',
      params: { width: 250 }
    });

    // Add code editor
    this.dockview.addPanel({
      id: 'codeEditor',
      component: 'codeEditor',
      title: 'SDL Editor',
      position: { direction: 'right' }
    });

    // Add system architecture
    this.dockview.addPanel({
      id: 'systemArchitecture',
      component: 'systemArchitecture',
      title: 'System Architecture',
      position: { direction: 'right' }
    });

    // Add traffic generation
    this.dockview.addPanel({
      id: 'trafficGeneration',
      component: 'trafficGeneration',
      title: 'Traffic Generation',
      position: { referencePanel: 'systemArchitecture', direction: 'below' }
    });

    // Add console
    this.dockview.addPanel({
      id: 'console',
      component: 'console',
      title: 'Console',
      position: { direction: 'below' }
    });

    // Add live metrics
    this.dockview.addPanel({
      id: 'liveMetrics',
      component: 'liveMetrics',
      title: 'Live Metrics',
      position: { referencePanel: 'console', direction: 'within' }
    });
  }

  private async loadCanvasState(): Promise<void> {
    appStateManager.updateState({ isLoading: true });
    
    try {
      const canvasState = await this.canvasService.getState();
      if (canvasState) {
        appStateManager.updateState({
          isConnected: true,
          currentFile: canvasState.loadedFiles?.[0],
          currentSystem: canvasState.activeSystem,
          isLoading: false
        });
        
        // Load generators
        const generators = await this.canvasService.getGenerators();
        const generateCalls = Object.values(generators).map((gen: any) => ({
          id: gen.id,
          name: gen.name,
          target: `${gen.component}.${gen.method}`,
          rate: gen.rate,
          enabled: gen.enabled
        }));
        
        appStateManager.updateState({ generateCalls });
        
        // Run initial simulation if system is loaded
        if (canvasState.activeSystem) {
          const result = await this.canvasService.runSimulation();
          if (result) {
            appStateManager.updateState({ simulationResults: result });
          }
        }
      }
    } catch (error) {
      console.error('Failed to load canvas state:', error);
      appStateManager.updateState({ 
        isLoading: false,
        error: `Failed to load state: ${error}`
      });
    }
  }

  private setupEventListeners(): void {
    // Generator events
    globalEventBus.on(AppEvents.GENERATOR_TOGGLED, async (data) => {
      const { generator, enabled } = data;
      try {
        if (enabled) {
          await this.canvasService.getClient().startGenerator(generator.name);
        } else {
          await this.canvasService.getClient().stopGenerator(generator.name);
        }
        
        // Refresh generators
        await this.refreshGenerators();
      } catch (error) {
        console.error('Failed to toggle generator:', error);
      }
    });

    globalEventBus.on(AppEvents.GENERATOR_UPDATED, async (data) => {
      const { generator, rate } = data;
      try {
        await this.canvasService.updateGeneratorRate(generator.name, rate);
        // State will be updated through event
      } catch (error) {
        console.error('Failed to update generator rate:', error);
      }
    });

    // Toolbar actions
    globalEventBus.on(AppEvents.TOOLBAR_ACTION, async (data) => {
      switch (data.action) {
        case 'toggleAllGenerators':
          await this.toggleAllGenerators();
          break;
        case 'addGenerator':
          // TODO: Show add generator dialog
          break;
        case 'removeGenerator':
          if (data.generatorId) {
            await this.removeGenerator(data.generatorId);
          }
          break;
      }
    });

    // Start metrics streaming
    this.canvasService.startMetricsStream((metrics) => {
      // Update state with new metrics
      const currentState = appStateManager.getState();
      const updatedCharts = { ...currentState.dynamicCharts };
      
      // Process metrics and update charts
      // (This is simplified - real implementation would be more complex)
      Object.entries(metrics).forEach(([key, value]) => {
        if (!updatedCharts[key]) {
          updatedCharts[key] = {
            chartName: key,
            metricName: key,
            title: key,
            data: [],
            labels: []
          };
        }
        
        // Add new data point
        updatedCharts[key].data.push(value as number);
        updatedCharts[key].labels.push(new Date().toLocaleTimeString());
        
        // Keep only last 20 points
        if (updatedCharts[key].data.length > 20) {
          updatedCharts[key].data.shift();
          updatedCharts[key].labels.shift();
        }
      });
      
      appStateManager.updateState({ dynamicCharts: updatedCharts });
    });
  }

  private async toggleAllGenerators(): Promise<void> {
    const state = appStateManager.getState();
    const hasEnabled = state.generateCalls.some(g => g.enabled);
    
    if (hasEnabled) {
      await this.canvasService.stopAllGenerators();
    } else {
      await this.canvasService.startAllGenerators();
    }
    
    await this.refreshGenerators();
  }

  private async removeGenerator(generatorId: string): Promise<void> {
    const state = appStateManager.getState();
    const generator = state.generateCalls.find(g => g.id === generatorId);
    
    if (generator) {
      try {
        await this.canvasService.deleteGenerator(generator.name);
        await this.refreshGenerators();
      } catch (error) {
        console.error('Failed to remove generator:', error);
      }
    }
  }

  private async refreshGenerators(): Promise<void> {
    try {
      const generators = await this.canvasService.getGenerators();
      const generateCalls = Object.values(generators).map((gen: any) => ({
        id: gen.id,
        name: gen.name,
        target: `${gen.component}.${gen.method}`,
        rate: gen.rate,
        enabled: gen.enabled
      }));
      
      appStateManager.updateState({ generateCalls });
    } catch (error) {
      console.error('Failed to refresh generators:', error);
    }
  }

  // Temporary methods - these components will be refactored next
  private createFileExplorerComponent(): any {
    const element = document.createElement('div');
    element.className = 'h-full overflow-auto';
    
    this.fileExplorer = new MultiFSExplorer(element);
    this.fileExplorer.setFileSelectHandler(async (path) => {
      // Handle file selection
      appStateManager.updateState({ currentFile: path });
    });
    
    // Initialize with dashboard reference (temporary)
    this.fileExplorer.initialize(this as any);
    
    return {
      element,
      init: () => {},
      dispose: () => {}
    };
  }

  private createCodeEditorComponent(): any {
    const element = document.createElement('div');
    element.className = 'h-full';
    
    // Code editor creation (simplified)
    element.innerHTML = '<div class="p-4 text-gray-400">Editor component</div>';
    
    return {
      element,
      init: () => {},
      dispose: () => {}
    };
  }

  private createConsoleComponent(): any {
    const element = document.createElement('div');
    element.className = 'h-full';
    
    new ConsolePanel(element);
    
    return {
      element,
      init: () => {},
      dispose: () => {}
    };
  }

  // Handler methods
  private async handleLoad(): Promise<void> {
    const state = appStateManager.getState();
    if (state.currentFile) {
      await this.canvasService.loadFile(state.currentFile);
    }
  }

  private async handleSave(): Promise<void> {
    // TODO: Implement save
  }

  private async handleRun(): Promise<void> {
    await this.canvasService.startAllGenerators();
    this.toolbar?.updateButton('run', { disabled: true });
    this.toolbar?.updateButton('stop', { disabled: false });
  }

  private async handleStop(): Promise<void> {
    await this.canvasService.stopAllGenerators();
    this.toolbar?.updateButton('run', { disabled: false });
    this.toolbar?.updateButton('stop', { disabled: true });
  }
}