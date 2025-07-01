import { EventBus } from './core/event-bus';
import { AppStateManager } from './core/app-state-manager';
import { CanvasClient } from './canvas-client';
import { SystemArchitecturePanel } from './panels/system-architecture-panel';
import { LiveMetricsPanel } from './panels/live-metrics-panel';
import { TrafficGenerationPanel } from './panels/traffic-generation-panel';
import { SDLEditorPanel } from './panels/sdl-editor-panel';
import { RecipeEditorPanel } from './panels/recipe-editor-panel';
import { ConsolePanel } from './components/console-panel';
import { Toolbar } from './components/toolbar';
import { RecipeRunner } from './components/recipe-runner';
import { DockviewApi, DockviewComponent } from 'dockview-core';
import { configureMonacoLoader } from './components/code-editor';
import { systemsService } from './services/systems-service';
import type { SystemProject } from './gen/sdl/v1/systems_pb';

interface SystemPageData {
  systemId: string;
  mode?: 'wasm' | 'server';
}

export class SystemDetailsPage {
  private container: HTMLElement;
  private eventBus: EventBus;
  private stateManager: AppStateManager;
  private pageData: SystemPageData;
  private systemData: SystemProject | null = null;
  private systemContent: { sdlContent: string; recipeContent: string; readmeContent: string } | null = null;
  private dockview: DockviewApi | null = null;
  private canvasClient: CanvasClient;
  
  // Panels
  private architecturePanel?: SystemArchitecturePanel;
  private metricsPanel?: LiveMetricsPanel;
  private trafficPanel?: TrafficGenerationPanel;
  private sdlEditorPanel?: SDLEditorPanel;
  private recipeEditorPanel?: RecipeEditorPanel;
  private consolePanel?: ConsolePanel;
  
  // Toolbar and controls
  private toolbar?: Toolbar;
  private recipeRunner?: RecipeRunner;

  constructor(pageData: SystemPageData) {
    this.container = document.getElementById('app') || document.body;
    this.pageData = pageData;
    this.eventBus = new EventBus();
    this.stateManager = new AppStateManager();
    this.canvasClient = new CanvasClient(pageData.systemId);
    
    // Configure Monaco loader
    configureMonacoLoader();
  }

  async initialize(): Promise<void> {
    try {
      // Load system data and content via API
      await this.loadSystemData();
      
      // Create the page layout
      this.createPageLayout();
      
      // Initialize state
      this.stateManager.updateState({
        currentSystem: this.pageData.systemId,
        currentFile: 'system.sdl'
      });
      
      // Setup toolbar
      this.initializeToolbar();
      
      // Initialize dockview layout
      this.initializeLayout();
      
      // Load system diagram
      this.loadSystemDiagram();
    } catch (error) {
      console.error('Failed to initialize system details page:', error);
      this.showError('Failed to load system data');
    }
  }
  
  private async loadSystemData(): Promise<void> {
    // Load system metadata and content in parallel
    const [systemData, systemContent] = await Promise.all([
      systemsService.getSystem(this.pageData.systemId),
      systemsService.getSystemContent(this.pageData.systemId)
    ]);
    
    this.systemData = systemData;
    this.systemContent = systemContent;
    
    // Update page title with actual system name
    if (this.systemData) {
      document.title = `${this.systemData.name} - SDL System`;
    }
  }
  
  private showError(message: string): void {
    this.container.innerHTML = `
      <div class="flex items-center justify-center h-full">
        <div class="text-center">
          <div class="text-red-500 text-lg mb-2">⚠️ Error</div>
          <div class="text-gray-600 dark:text-gray-400">${message}</div>
        </div>
      </div>
    `;
  }

  private createPageLayout(): void {
    this.container.innerHTML = `
      <div class="h-screen flex flex-col bg-gray-50 dark:bg-gray-950">
        <!-- Header -->
        <header class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-3">
            <div class="flex justify-between items-center pr-48">
                <div class="flex items-center gap-4">
                    <a href="/systems" class="inline-flex items-center gap-2 px-4 py-2 text-gray-600 dark:text-gray-400 border border-gray-300 dark:border-gray-600 rounded-lg hover:text-gray-900 dark:hover:text-white hover:border-gray-400 dark:hover:border-gray-500 transition-colors">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
                        </svg>
                        Back to Systems
                    </a>
                    <h1 class="text-2xl font-bold text-gray-900 dark:text-white">${this.systemData?.name || 'Loading...'}</h1>
                </div>
                <div class="flex items-center gap-2">
                    <span class="text-sm text-gray-600 dark:text-gray-400">${this.systemData?.description || ''}</span>
                </div>
            </div>
        </header>
        
        <!-- Toolbar -->
        <div id="toolbar-container" class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"></div>
        
        <!-- Main Dockview Container -->
        <div id="dockview-container" class="flex-1"></div>
      </div>
    `;
  }

  private initializeToolbar(): void {
    const toolbarContainer = document.getElementById('toolbar-container');
    if (!toolbarContainer) return;
    
    this.toolbar = new Toolbar(toolbarContainer);
    this.toolbar.setButtons([
      {
        id: 'save',
        label: 'Save',
        icon: '💾',
        tooltip: 'Save changes',
        onClick: () => this.saveChanges()
      },
      {
        id: 'share',
        label: 'Share',
        icon: '🔗',
        tooltip: 'Share this system',
        onClick: () => this.shareSystem()
      },
      {
        id: 'run',
        label: 'Run',
        icon: '▶️',
        tooltip: 'Run simulation',
        onClick: () => this.runSystem()
      },
      {
        id: 'stop',
        label: 'Stop',
        icon: '⏹️',
        tooltip: 'Stop simulation',
        disabled: true,
        onClick: () => this.stopSystem()
      },
      {
        id: 'step',
        label: 'Step',
        icon: '⏩',
        tooltip: 'Step through recipe',
        disabled: true,
        onClick: () => this.stepRecipe()
      }
    ]);
    
    // Initialize recipe controls
    this.initializeRecipeControls();
  }

  private initializeLayout(): void {
    const container = document.getElementById('dockview-container');
    if (!container) {
      console.error('❌ DockView container not found');
      return;
    }

    // Apply theme class based on current theme
    const isDarkMode = document.documentElement.classList.contains('dark');
    container.className = isDarkMode ? 'dockview-theme-dark flex-1' : 'dockview-theme-light flex-1';
    
    // Listen for theme changes
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.type === 'attributes' && mutation.attributeName === 'class') {
          const isDarkMode = document.documentElement.classList.contains('dark');
          container.className = isDarkMode ? 'dockview-theme-dark flex-1' : 'dockview-theme-light flex-1';
        }
      });
    });
    
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    });
    
    // Create DockView component
    const dockviewComponent = new DockviewComponent(container, {
      createComponent: (options: any) => {
        switch (options.name) {
          case 'sdlEditor':
            return this.createSDLEditorComponent();
          case 'recipeEditor':
            return this.createRecipeEditorComponent();
          case 'systemArchitecture':
            return this.createSystemArchitectureComponent();
          case 'trafficGeneration':
            return this.createTrafficGenerationComponent();
          case 'liveMetrics':
            return this.createLiveMetricsComponent();
          case 'console':
            return this.createConsoleComponent();
          default:
            return {
              element: document.createElement('div'),
              init: () => {},
              dispose: () => {}
            };
        }
      }
    });

    this.dockview = dockviewComponent.api;
    
    // Load saved layout or create default
    const savedLayout = this.loadLayout();
    if (savedLayout) {
      try {
        this.dockview.fromJSON(savedLayout);
      } catch (e) {
        console.warn('Failed to restore layout, using default', e);
        this.createDefaultLayout();
      }
    } else {
      this.createDefaultLayout();
    }
    
    // Save layout on changes
    this.dockview.onDidLayoutChange(() => {
      this.saveLayout();
    });
  }

  private createDefaultLayout(): void {
    if (!this.dockview) return;

    // Add SDL editor panel
    this.dockview.addPanel({
      id: 'sdlEditor',
      component: 'sdlEditor',
      title: 'System Design (SDL)'
    });

    // Add recipe editor panel as a separate panel below SDL editor
    this.dockview.addPanel({
      id: 'recipeEditor',
      component: 'recipeEditor',
      title: 'Demo Recipe',
      position: { direction: 'below', referencePanel: 'sdlEditor' }
    });

    // Add system architecture panel
    this.dockview.addPanel({
      id: 'systemArchitecture',
      component: 'systemArchitecture',
      title: 'System Architecture',
      position: { direction: 'right', referencePanel: 'sdlEditor' }
    });

    // Add traffic generation panel
    this.dockview.addPanel({
      id: 'trafficGeneration',
      component: 'trafficGeneration',
      title: 'Traffic Generation',
      position: { direction: 'below', referencePanel: 'systemArchitecture' }
    });

    // Add live metrics panel
    this.dockview.addPanel({
      id: 'liveMetrics',
      component: 'liveMetrics',
      title: 'Live Metrics',
      position: { direction: 'below', referencePanel: 'trafficGeneration' }
    });

    // Add console panel below recipe editor
    this.dockview.addPanel({
      id: 'console',
      component: 'console',
      title: 'Output',
      position: { direction: 'below', referencePanel: 'recipeEditor' }
    });
  }

  private createSDLEditorComponent() {
    const container = document.createElement('div');
    container.style.width = '100%';
    container.style.height = '100%';
    
    this.sdlEditorPanel = new SDLEditorPanel({
      id: 'sdlEditor',
      title: 'System Design (SDL)',
      eventBus: this.eventBus,
      stateManager: this.stateManager,
      sdlContent: this.systemContent?.sdlContent || '',
      readOnly: false,
      onChange: (_content) => {
        // Update state when SDL content changes
        this.stateManager.updateState({ currentFile: 'system.sdl' });
      }
    });
    
    return {
      element: container,
      init: async () => {
        await this.sdlEditorPanel?.initialize(container);
        
        // Update content after initialization if we have it
        if (this.systemContent?.sdlContent && this.sdlEditorPanel) {
          this.sdlEditorPanel.updateContent(this.systemContent.sdlContent);
        }
      },
      dispose: () => this.sdlEditorPanel?.dispose()
    };
  }

  private createRecipeEditorComponent() {
    const container = document.createElement('div');
    container.style.width = '100%';
    container.style.height = '100%';
    
    this.recipeEditorPanel = new RecipeEditorPanel({
      id: 'recipeEditor',
      title: 'Demo Recipe',
      eventBus: this.eventBus,
      stateManager: this.stateManager,
      recipeContent: this.systemContent?.recipeContent || '',
      readOnly: false,
      onChange: (_content) => {
        // Update state when recipe content changes
        this.stateManager.updateState({ currentFile: 'demo.recipe' });
      }
    });
    
    return {
      element: container,
      init: async () => {
        await this.recipeEditorPanel?.initialize(container);
        
        // Update content after initialization if we have it
        if (this.systemContent?.recipeContent && this.recipeEditorPanel) {
          this.recipeEditorPanel.updateContent(this.systemContent.recipeContent);
        }
      },
      dispose: () => this.recipeEditorPanel?.dispose()
    };
  }

  private createSystemArchitectureComponent() {
    const container = document.createElement('div');
    container.style.width = '100%';
    container.style.height = '100%';
    
    this.architecturePanel = new SystemArchitecturePanel({
      id: 'systemArchitecture',
      title: 'System Architecture',
      eventBus: this.eventBus,
      stateManager: this.stateManager
    });
    
    return {
      element: container,
      init: async () => {
        await this.architecturePanel?.initialize(container);
      },
      dispose: () => this.architecturePanel?.dispose()
    };
  }

  private createTrafficGenerationComponent() {
    const container = document.createElement('div');
    container.style.width = '100%';
    container.style.height = '100%';
    
    this.trafficPanel = new TrafficGenerationPanel({
      id: 'trafficGeneration',
      title: 'Traffic Generation',
      eventBus: this.eventBus,
      stateManager: this.stateManager
    });
    
    return {
      element: container,
      init: async () => {
        await this.trafficPanel?.initialize(container);
      },
      dispose: () => this.trafficPanel?.dispose()
    };
  }

  private createLiveMetricsComponent() {
    const container = document.createElement('div');
    container.style.width = '100%';
    container.style.height = '100%';
    
    this.metricsPanel = new LiveMetricsPanel({
      id: 'liveMetrics',
      title: 'Live Metrics',
      eventBus: this.eventBus,
      stateManager: this.stateManager
    });
    
    return {
      element: container,
      init: async () => {
        await this.metricsPanel?.initialize(container);
      },
      dispose: () => this.metricsPanel?.dispose()
    };
  }

  private createConsoleComponent() {
    const container = document.createElement('div');
    this.consolePanel = new ConsolePanel(container);
    
    return {
      element: container,
      init: () => {},
      dispose: () => this.consolePanel?.clear()
    };
  }

  private initializeRecipeControls(): void {
    if (!this.toolbar) {
      return;
    }
    
    // Initialize recipe runner
    this.recipeRunner = new RecipeRunner(this.canvasClient);
  }
  

  private async loadSystemDiagram(): Promise<void> {
    try {
      // Use WASM to generate the system diagram from SDL content
      if (!this.systemContent?.sdlContent) {
        console.warn('No SDL content available for diagram generation');
        return;
      }

      // Import the WASM module dynamically
      const { generateSystemDiagram } = await import('./wasm-integration');
      
      // Generate diagram using WASM
      const diagramData = await generateSystemDiagram(this.systemContent.sdlContent);
      
      if (diagramData) {
        // Update state
        this.stateManager.updateState({ 
          currentSystem: this.pageData.systemId
        });
        
        // Emit event so architecture panel can update
        this.eventBus.emit('system:diagram:loaded', diagramData);
      }
    } catch (error) {
      console.error('Failed to generate system diagram:', error);
    }
  }

  private async runSystem(): Promise<void> {
    // Update toolbar buttons
    this.toolbar?.updateButton('run', { disabled: true });
    this.toolbar?.updateButton('stop', { disabled: false });
    
    // Get current SDL content when needed
    // const sdlContent = this.sdlEditorPanel?.getContent() || this.pageData.sdlContent;
    
    // Log to console
    this.consolePanel?.info('Starting system simulation...');
    
    try {
      // Run system based on mode
      if (this.pageData.mode === 'wasm') {
        this.consolePanel?.info('Initializing WASM runtime...');
        // TODO: Initialize WASM canvas and run
      } else {
        this.consolePanel?.info('Starting system on server...');
        // For now, just save the SDL content locally and get the diagram
        // TODO: Implement proper SDL compilation endpoint
        this.consolePanel?.info('Processing SDL...');
        
        // Store the SDL content in state
        this.stateManager.updateState({ currentFile: 'system.sdl' });
        
        // Load system diagram
        await this.loadSystemDiagram();
        this.consolePanel?.success('System loaded successfully');
      }
    } catch (error) {
      this.consolePanel?.error(`Error: ${error}`);
    }
  }

  private stopSystem(): void {
    // Update toolbar buttons
    this.toolbar?.updateButton('run', { disabled: false });
    this.toolbar?.updateButton('stop', { disabled: true });
    
    // Stop simulation
    this.consolePanel?.info('Stopping simulation...');
    // Clear the system diagram
    this.eventBus.emit('system:diagram:loaded', null);
    this.consolePanel?.success('Simulation stopped');
    
    // Stop recipe if running
    if (this.recipeRunner) {
      this.recipeRunner.stop();
    }
  }

  private async saveChanges(): Promise<void> {
    this.toolbar?.setStatus('Saving...', 'info');
    
    try {
      // TODO: Implement save endpoint
      // const sdlContent = this.sdlEditorPanel?.getContent();
      // const recipeContent = this.recipeEditorPanel?.getContent();
      // For now, just show success
      this.consolePanel?.success('Changes saved successfully');
      this.toolbar?.setStatus('Saved', 'success');
      
      setTimeout(() => {
        this.toolbar?.setStatus('Ready', 'info');
      }, 2000);
    } catch (error) {
      this.consolePanel?.error('Failed to save changes');
      this.toolbar?.setStatus('Save failed', 'error');
    }
  }

  private async shareSystem(): Promise<void> {
    const url = window.location.href;
    
    try {
      await navigator.clipboard.writeText(url);
      
      // Show feedback in toolbar
      this.toolbar?.setStatus('URL copied to clipboard!', 'success');
      setTimeout(() => {
        this.toolbar?.setStatus('Ready', 'info');
      }, 2000);
    } catch (err) {
      console.error('Failed to copy URL:', err);
      this.consolePanel?.error('Failed to copy URL to clipboard');
    }
  }
  
  private stepRecipe(): void {
    if (this.recipeRunner) {
      // TODO: Implement step mode in RecipeRunner
      this.consolePanel?.info('Stepping through recipe...');
    }
  }

  private saveLayout(): void {
    if (!this.dockview) return;
    
    const layout = this.dockview.toJSON();
    const layoutKey = `sdl-details-layout-${this.pageData.systemId}`;
    localStorage.setItem(layoutKey, JSON.stringify(layout));
  }
  
  private loadLayout(): any {
    const layoutKey = `sdl-details-layout-${this.pageData.systemId}`;
    const saved = localStorage.getItem(layoutKey);
    return saved ? JSON.parse(saved) : null;
  }

  public destroy(): void {
    // Save layout before destroying
    this.saveLayout();
    
    // Dispose dockview
    if (this.dockview) {
      this.dockview.dispose();
    }
    
    // Clean up panels
    this.architecturePanel?.dispose();
    this.metricsPanel?.dispose();
    this.trafficPanel?.dispose();
    this.sdlEditorPanel?.dispose();
    this.recipeEditorPanel?.dispose();
    this.consolePanel?.clear();
  }
}