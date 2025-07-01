import { EventBus } from './core/event-bus';
import { AppStateManager } from './core/app-state-manager';
import { CanvasClient } from './canvas-client';
import { WasmManager } from './wasm-integration';
import { SystemArchitecturePanel } from './panels/system-architecture-panel';
import { LiveMetricsPanel } from './panels/live-metrics-panel';
import { TrafficGenerationPanel } from './panels/traffic-generation-panel';
import { RecipeRunner } from './components/recipe-runner';
import * as monaco from 'monaco-editor';

interface SystemPageData {
  systemId: string;
  systemName: string;
  systemDescription: string;
  sdlContent: string;
  recipeContent: string;
  mode?: 'wasm' | 'server';
}

export class SystemDetailsPage {
  private container: HTMLElement;
  private eventBus: EventBus;
  private stateManager: AppStateManager;
  private pageData: SystemPageData;
  private editor?: monaco.editor.IStandaloneCodeEditor;
  private recipeRunner?: RecipeRunner;
  private currentTab: 'sdl' | 'recipe' = 'sdl';
  
  // Panels
  private architecturePanel?: SystemArchitecturePanel;
  private metricsPanel?: LiveMetricsPanel;
  private trafficPanel?: TrafficGenerationPanel;

  constructor(pageData: SystemPageData) {
    this.container = document.getElementById('app') || document.body;
    this.pageData = pageData;
    this.eventBus = new EventBus();
    this.stateManager = new AppStateManager(this.eventBus);
  }

  async initialize(): Promise<void> {
    this.initializeEditor();
    this.initializePanels();
    this.attachEventListeners();
  }

  private initializeEditor(): void {
    const container = document.getElementById('editor-container');
    if (!container) return;

    // Register SDL language if not already registered
    if (!monaco.languages.getLanguages().find(lang => lang.id === 'sdl')) {
      monaco.languages.register({ id: 'sdl' });
      monaco.languages.setMonarchTokensProvider('sdl', {
        tokenizer: {
          root: [
            [/\/\/.*$/, 'comment'],
            [/import/, 'keyword'],
            [/system|service|database|cache/, 'keyword'],
            [/:=/, 'operator'],
            [/->/, 'operator'],
            [/"[^"]*"/, 'string'],
            [/\d+/, 'number'],
            [/{|}|\[|\]/, 'delimiter']
          ]
        }
      });
    }

    // Determine theme based on current page theme
    const isDarkMode = document.documentElement.classList.contains('dark');
    
    this.editor = monaco.editor.create(container, {
      value: this.pageData.sdlContent,
      language: 'sdl',
      theme: isDarkMode ? 'vs-dark' : 'vs',
      automaticLayout: true,
      minimap: { enabled: false },
      fontSize: 14,
      lineNumbers: 'on',
      wordWrap: 'on',
      scrollBeyondLastLine: false
    });
  }

  private initializePanels(): void {
    // Initialize architecture panel
    const archContainer = document.getElementById('architecture-panel');
    if (archContainer) {
      // TODO: Properly initialize services
      this.architecturePanel = new SystemArchitecturePanel(
        archContainer,
        this.eventBus,
        {} as any
      );
    }

    // Initialize metrics panel
    const metricsContainer = document.getElementById('metrics-panel');
    if (metricsContainer) {
      this.metricsPanel = new LiveMetricsPanel(
        metricsContainer,
        this.eventBus,
        {} as any
      );
    }

    // Initialize traffic panel
    const trafficContainer = document.getElementById('traffic-panel');
    if (trafficContainer) {
      this.trafficPanel = new TrafficGenerationPanel(
        trafficContainer,
        this.eventBus,
        {} as any
      );
    }
  }

  private attachEventListeners(): void {
    // Tab switching
    document.querySelectorAll('.tab-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const tab = (e.target as HTMLElement).dataset.tab as 'sdl' | 'recipe';
        if (tab) this.switchTab(tab);
      });
    });

    // Run/Stop buttons
    document.getElementById('run-btn')?.addEventListener('click', () => {
      this.runSystem();
    });

    document.getElementById('stop-btn')?.addEventListener('click', () => {
      this.stopSystem();
    });

    // Share button
    document.getElementById('share-btn')?.addEventListener('click', () => {
      this.shareSystem();
    });

    // Listen for theme changes
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.type === 'attributes' && mutation.attributeName === 'class') {
          const isDarkMode = document.documentElement.classList.contains('dark');
          monaco.editor.setTheme(isDarkMode ? 'vs-dark' : 'vs');
        }
      });
    });
    
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    });
  }

  private switchTab(tab: 'sdl' | 'recipe'): void {
    this.currentTab = tab;
    
    // Update tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
      btn.classList.toggle('active', btn.dataset.tab === tab);
    });

    // Update editor content
    if (this.editor) {
      if (tab === 'sdl') {
        this.editor.setValue(this.pageData.sdlContent);
        monaco.editor.setModelLanguage(this.editor.getModel()!, 'sdl');
      } else {
        this.editor.setValue(this.pageData.recipeContent);
        monaco.editor.setModelLanguage(this.editor.getModel()!, 'shell');
      }
    }
  }

  private async runSystem(): Promise<void> {
    const runBtn = document.getElementById('run-btn') as HTMLElement;
    const stopBtn = document.getElementById('stop-btn') as HTMLElement;
    const outputSection = document.getElementById('output-section') as HTMLElement;
    const outputContent = document.getElementById('output-content') as HTMLElement;

    // Update UI
    runBtn.style.display = 'none';
    stopBtn.style.display = 'block';
    outputSection.style.display = 'block';

    // Show panels
    const metricsPanel = document.getElementById('metrics-panel');
    const trafficPanel = document.getElementById('traffic-panel');
    if (metricsPanel) metricsPanel.style.display = 'block';
    if (trafficPanel) trafficPanel.style.display = 'block';

    // Get current SDL content
    const sdlContent = this.currentTab === 'sdl' ? this.editor?.getValue() : this.pageData.sdlContent;

    // Run system based on mode
    if (this.pageData.mode === 'wasm') {
      outputContent.innerHTML = '<div class="info">Initializing WASM runtime...</div>';
      // TODO: Initialize WASM canvas and run
    } else {
      outputContent.innerHTML = '<div class="info">Starting system on server...</div>';
      // TODO: Send to server to run
    }

    // Run recipe if on recipe tab
    if (this.currentTab === 'recipe') {
      this.runRecipe(outputContent);
    }
  }

  private runRecipe(outputElement: HTMLElement): void {
    const recipeContent = this.currentTab === 'recipe' ? 
      this.editor?.getValue() || this.pageData.recipeContent : 
      this.pageData.recipeContent;

    this.recipeRunner = new RecipeRunner(
      outputElement,
      (step, command) => {
        console.log(`Step ${step}: ${command}`);
      },
      () => {
        console.log('Recipe completed');
      }
    );

    this.recipeRunner.runRecipe(recipeContent);
  }

  private stopSystem(): void {
    const runBtn = document.getElementById('run-btn') as HTMLElement;
    const stopBtn = document.getElementById('stop-btn') as HTMLElement;

    // Update UI
    runBtn.style.display = 'block';
    stopBtn.style.display = 'none';

    // Stop recipe if running
    if (this.recipeRunner) {
      this.recipeRunner.stop();
    }

    // TODO: Stop actual system
  }

  private async shareSystem(): Promise<void> {
    const url = window.location.href;
    
    try {
      await navigator.clipboard.writeText(url);
      
      // Show feedback
      const shareBtn = document.getElementById('share-btn') as HTMLElement;
      const originalText = shareBtn.textContent;
      shareBtn.textContent = 'Copied!';
      setTimeout(() => {
        shareBtn.textContent = originalText;
      }, 2000);
    } catch (err) {
      console.error('Failed to copy URL:', err);
    }
  }

  destroy(): void {
    this.editor?.dispose();
    this.architecturePanel?.destroy();
    this.metricsPanel?.destroy();
    this.trafficPanel?.destroy();
  }
}