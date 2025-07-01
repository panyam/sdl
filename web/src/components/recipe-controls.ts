import { RecipeRunner } from './recipe-runner.js';
import type { CanvasClient } from '../canvas-client.js';
import type { TabbedEditor } from './tabbed-editor.js';

export interface RecipeControlsOptions {
  api: CanvasClient;
  tabbedEditor: TabbedEditor;
  onConsoleOutput?: (message: string, type: 'info' | 'success' | 'error' | 'warning') => void;
  onRecipeStateChange?: (isRunning: boolean, recipePath?: string, currentLine?: number) => void;
}

export class RecipeControls {
  private container: HTMLElement;
  private runner: RecipeRunner;
  private tabbedEditor: TabbedEditor;
  private currentRecipeTab: string | null = null;
  private isRunning: boolean = false;
  private onRecipeStateChange?: (isRunning: boolean, recipePath?: string, currentLine?: number) => void;

  constructor(container: HTMLElement, options: RecipeControlsOptions) {
    this.container = container;
    this.tabbedEditor = options.tabbedEditor;
    this.runner = new RecipeRunner(options.api);
    this.onRecipeStateChange = options.onRecipeStateChange;
    
    // Set up output handler
    if (options.onConsoleOutput) {
      this.runner.setOutputHandler(options.onConsoleOutput);
    }
    
    // Set up state change handler
    this.runner.setStateChangeHandler((state) => {
      this.isRunning = state.isRunning;
      
      // Update current line highlighting in editor
      if (this.currentRecipeTab && this.tabbedEditor) {
        let currentLineNumber: number | undefined;
        if (state.isRunning && state.currentStep < state.steps.length) {
          const currentStep = state.steps[state.currentStep];
          currentLineNumber = currentStep?.command?.lineNumber;
        }
        
        this.tabbedEditor.setRecipeRunning(
          this.currentRecipeTab,
          state.isRunning,
          currentLineNumber
        );
        
        // Notify about state change
        if (this.onRecipeStateChange) {
          const [_fsId, path] = this.currentRecipeTab.split(':');
          this.onRecipeStateChange(state.isRunning, path, currentLineNumber);
        }
      }
      
      if (!state.isRunning) {
        this.currentRecipeTab = null;
      }
      
      this.render();
    });
    
    this.render();
  }

  render() {
    const activeTab = this.tabbedEditor.getActiveTab();
    
    let isRecipeFile = false;
    let recipeErrors: string[] = [];
    if (activeTab) {
      const [_fsId, path] = activeTab.split(':');
      isRecipeFile = path?.endsWith('.recipe') || false;
      
      // Validate recipe content if it's a recipe file
      if (isRecipeFile && !this.isRunning) {
        const content = this.tabbedEditor.getActiveContent();
        if (content) {
          const validationErrors = this.runner.validateRecipe(content);
          recipeErrors = validationErrors.map(err => `Line ${err.lineNumber}: ${err.message}`);
        }
      }
    }
    
    const hasErrors = recipeErrors.length > 0;
    const canRun = isRecipeFile && !this.isRunning && !hasErrors;
    const canStop = this.isRunning;
    const canStep = this.isRunning;
    
    let recipeName = '';
    if (this.currentRecipeTab) {
      const [_fsId, path] = this.currentRecipeTab.split(':');
      recipeName = path.split('/').pop() || '';
    }
    
    // Build tooltip for run button
    let runTooltip = 'Run recipe';
    if (hasErrors) {
      runTooltip = `Recipe has errors:\n${recipeErrors.join('\n')}`;
    } else if (!isRecipeFile) {
      runTooltip = 'Open a .recipe file to run';
    } else if (this.isRunning) {
      runTooltip = 'A recipe is already running';
    }
    
    this.container.innerHTML = `
      <div class="recipe-controls flex items-center gap-2">
        ${this.isRunning ? `
          <span class="text-xs text-yellow-400 mr-2">
            <svg class="inline w-3 h-3 mr-1" viewBox="0 0 16 16" fill="currentColor">
              <path d="M5 3l8 5-8 5V3z"/>
            </svg>
            ${recipeName}
          </span>
        ` : ''}
        
        <button 
          class="toolbar-btn ${canRun ? '' : 'opacity-50 cursor-not-allowed'} ${hasErrors ? 'text-red-400' : ''}"
          onclick="window.recipeControls?.handleRun()"
          ${canRun ? '' : 'disabled'}
          title="${runTooltip}"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M5 3l8 5-8 5V3z"/>
          </svg>
        </button>
        
        <button 
          class="toolbar-btn ${canStep ? '' : 'opacity-50 cursor-not-allowed'}"
          onclick="window.recipeControls?.handleStep()"
          ${canStep ? '' : 'disabled'}
          title="Step through recipe"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M5 3l3 2.5L5 8V3zm4 0l3 2.5L9 8V3zm4 0h2v5h-2V3z"/>
          </svg>
        </button>
        
        <button 
          class="toolbar-btn ${canStop ? '' : 'opacity-50 cursor-not-allowed'}"
          onclick="window.recipeControls?.handleStop()"
          ${canStop ? '' : 'disabled'}
          title="Stop recipe"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M4 4h8v8H4V4z"/>
          </svg>
        </button>
      </div>
    `;
    
    // Make this instance globally accessible
    (window as any).recipeControls = this;
  }

  async handleRun() {
    const activeTab = this.tabbedEditor.getActiveTab();
    if (!activeTab || this.isRunning) return;
    
    const content = this.tabbedEditor.getActiveContent();
    if (!content) return;
    
    try {
      // Stop any currently running recipe
      if (this.currentRecipeTab && this.currentRecipeTab !== activeTab) {
        await this.runner.stop();
        this.tabbedEditor.setRecipeRunning(this.currentRecipeTab, false);
      }
      
      this.currentRecipeTab = activeTab;
      await this.runner.loadRecipe(activeTab.split(':')[1], content);
      await this.runner.start('step');
    } catch (error: any) {
      // Recipe failed to load - reset state
      this.currentRecipeTab = null;
      this.render();
      // Error already shown to console by runner
    }
  }

  async handleStep() {
    if (!this.isRunning) return;
    await this.runner.step();
  }

  async handleStop() {
    if (!this.isRunning || !this.currentRecipeTab) return;
    
    await this.runner.stop();
    this.tabbedEditor.setRecipeRunning(this.currentRecipeTab, false);
    this.currentRecipeTab = null;
  }

  updateForActiveTab() {
    // Re-render when active tab changes
    this.render();
    
    // Validate recipe if it's active
    const activeTab = this.tabbedEditor.getActiveTab();
    if (activeTab) {
      const [_fsId, path] = activeTab.split(':');
      if (path?.endsWith('.recipe')) {
        this.validateCurrentRecipe();
      } else {
        // Clear any existing errors
        this.tabbedEditor.clearErrorDecorations(activeTab);
      }
    }
  }

  validateCurrentRecipe() {
    const activeTab = this.tabbedEditor.getActiveTab();
    if (!activeTab) return;
    
    const [_fsId, path] = activeTab.split(':');
    if (!path?.endsWith('.recipe')) return;
    
    const content = this.tabbedEditor.getActiveContent();
    if (!content) return;
    
    const errors = this.runner.validateRecipe(content);
    this.tabbedEditor.updateErrorDecorations(activeTab, errors);
  }

  dispose() {
    this.runner.stop();
    // Don't clear the container on dispose - let the parent manage it
    // this.container.innerHTML = '';
  }
}