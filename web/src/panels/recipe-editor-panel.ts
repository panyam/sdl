import { BasePanel, PanelConfig } from './base-panel.js';
import { AppState } from '../core/app-state-manager.js';
import { AppEvents } from '../core/event-bus.js';
import * as monaco from 'monaco-editor';

export interface RecipeEditorPanelConfig extends PanelConfig {
  recipeContent: string;
  readOnly?: boolean;
  onChange?: (content: string) => void;
  onRunningStateChange?: (isRunning: boolean, currentLine?: number) => void;
}

/**
 * Panel for editing and displaying recipe files
 */
export class RecipeEditorPanel extends BasePanel {
  private editor: monaco.editor.IStandaloneCodeEditor | null = null;
  private recipeContent: string;
  private readOnly: boolean;
  private onChange?: (content: string) => void;
  private onRunningStateChange?: (isRunning: boolean, currentLine?: number) => void;
  private resizeObserver: ResizeObserver | null = null;
  private executionDecorations: string[] = [];
  private errorDecorations: string[] = [];
  private _isRunning: boolean = false;
  private _currentLine?: number;

  constructor(config: RecipeEditorPanelConfig) {
    super({
      ...config,
      id: config.id || 'recipeEditor',
      title: config.title || 'Demo Recipe'
    });
    
    this.recipeContent = config.recipeContent;
    this.readOnly = config.readOnly || false;
    this.onChange = config.onChange;
    this.onRunningStateChange = config.onRunningStateChange;
  }

  protected async onInitialize(): Promise<void> {
    // Set up resize observer
    this.resizeObserver = new ResizeObserver(() => {
      this.editor?.layout();
    });

    // Listen for recipe execution events
    this.eventBus.on(AppEvents.RECIPE_STARTED, this.handleRecipeStarted);
    this.eventBus.on(AppEvents.RECIPE_STEP_EXECUTED, this.handleRecipeStepExecuted);
    this.eventBus.on(AppEvents.RECIPE_COMPLETED, this.handleRecipeCompleted);
    this.eventBus.on(AppEvents.RECIPE_ERROR, this.handleRecipeError);
  }

  protected onDispose(): void {
    if (this.editor) {
      this.editor.dispose();
      this.editor = null;
    }
    
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
      this.resizeObserver = null;
    }

    // Remove event listeners
    this.eventBus.off(AppEvents.RECIPE_STARTED, this.handleRecipeStarted);
    this.eventBus.off(AppEvents.RECIPE_STEP_EXECUTED, this.handleRecipeStepExecuted);
    this.eventBus.off(AppEvents.RECIPE_COMPLETED, this.handleRecipeCompleted);
    this.eventBus.off(AppEvents.RECIPE_ERROR, this.handleRecipeError);
  }

  onStateChange(_state: AppState, _changedKeys: string[]): void {
    // Recipe editor doesn't need to respond to state changes
    // Content is managed directly through config
  }

  protected render(): void {
    const isDarkMode = document.documentElement.classList.contains('dark');
    
    const content = `
      <div class="h-full w-full flex flex-col bg-gray-50 dark:bg-gray-900">
        <div id="${this.id}-editor" class="flex-1"></div>
      </div>
    `;
    
    this.setContent(content);
    
    // Initialize Monaco editor after content is rendered
    setTimeout(() => {
      const editorContainer = document.getElementById(`${this.id}-editor`);
      if (editorContainer && !this.editor) {
        this.editor = monaco.editor.create(editorContainer, {
          value: this.recipeContent,
          language: 'shell',
          theme: isDarkMode ? 'vs-dark' : 'vs',
          automaticLayout: false,
          minimap: { enabled: false },
          fontSize: 14,
          lineNumbers: 'on',
          wordWrap: 'on',
          scrollBeyondLastLine: false,
          readOnly: this.readOnly
        });

        // Set up change handler
        if (this.onChange && !this.readOnly) {
          this.editor.onDidChangeModelContent(() => {
            const content = this.editor?.getValue() || '';
            this.onChange!(content);
          });
        }

        // Observe resize
        if (this.resizeObserver && editorContainer) {
          this.resizeObserver.observe(editorContainer);
        }

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
    }, 100);
  }

  private handleRecipeStarted = () => {
    this.setRunning(true);
  };

  private handleRecipeStepExecuted = (event: any) => {
    if (event.lineNumber) {
      this.highlightLine(event.lineNumber);
    }
  };

  private handleRecipeCompleted = () => {
    this.setRunning(false);
    this.clearDecorations();
  };

  private handleRecipeError = (event: any) => {
    this.setRunning(false);
    if (event.lineNumber) {
      this.highlightError(event.lineNumber);
    }
  };

  public setRunning(isRunning: boolean, currentLine?: number): void {
    this._isRunning = isRunning;
    this._currentLine = currentLine;
    
    if (isRunning && currentLine) {
      this.highlightLine(currentLine);
    } else if (!isRunning) {
      this.clearDecorations();
    }

    this.onRunningStateChange?.(isRunning, currentLine);
  }

  private highlightLine(lineNumber: number): void {
    if (!this.editor) return;

    // Clear previous execution decorations
    this.executionDecorations = this.editor.deltaDecorations(
      this.executionDecorations,
      [{
        range: new monaco.Range(lineNumber, 1, lineNumber, 1),
        options: {
          isWholeLine: true,
          className: 'bg-blue-500 bg-opacity-20',
          glyphMarginClassName: 'codicon codicon-arrow-right text-blue-500'
        }
      }]
    );

    // Scroll to line
    this.editor.revealLineInCenter(lineNumber);
  }

  private highlightError(lineNumber: number): void {
    if (!this.editor) return;

    // Clear execution decorations
    this.executionDecorations = this.editor.deltaDecorations(this.executionDecorations, []);

    // Add error decoration
    this.errorDecorations = this.editor.deltaDecorations(
      this.errorDecorations,
      [{
        range: new monaco.Range(lineNumber, 1, lineNumber, 1),
        options: {
          isWholeLine: true,
          className: 'bg-red-500 bg-opacity-20',
          glyphMarginClassName: 'codicon codicon-error text-red-500'
        }
      }]
    );
  }

  private clearDecorations(): void {
    if (!this.editor) return;

    this.executionDecorations = this.editor.deltaDecorations(this.executionDecorations, []);
    this.errorDecorations = this.editor.deltaDecorations(this.errorDecorations, []);
  }

  public updateContent(content: string): void {
    this.recipeContent = content;
    if (this.editor) {
      this.editor.setValue(content);
    }
  }

  public getContent(): string {
    return this.editor?.getValue() || this.recipeContent;
  }
}