import { BasePanel, PanelConfig } from './base-panel.js';
import { AppState } from '../core/app-state-manager.js';
import * as monaco from 'monaco-editor';

export interface SDLEditorPanelConfig extends PanelConfig {
  sdlContent: string;
  readOnly?: boolean;
  onChange?: (content: string) => void;
}

/**
 * Panel for editing SDL system definitions
 */
export class SDLEditorPanel extends BasePanel {
  private editor: monaco.editor.IStandaloneCodeEditor | null = null;
  private sdlContent: string;
  private readOnly: boolean;
  private onChange?: (content: string) => void;
  private resizeObserver: ResizeObserver | null = null;

  constructor(config: SDLEditorPanelConfig) {
    super({
      ...config,
      id: config.id || 'sdlEditor',
      title: config.title || 'System Design (SDL)'
    });
    
    this.sdlContent = config.sdlContent;
    this.readOnly = config.readOnly || false;
    this.onChange = config.onChange;
  }

  protected async onInitialize(): Promise<void> {
    // Register SDL language if not already registered
    if (!monaco.languages.getLanguages().find(lang => lang.id === 'sdl')) {
      monaco.languages.register({ id: 'sdl' });
      monaco.languages.setMonarchTokensProvider('sdl', {
        keywords: [
          'component', 'system', 'method', 'use', 'import', 'let', 'return',
          'if', 'else', 'switch', 'case', 'default', 'for', 'go', 'wait',
          'sample', 'distribute', 'enum', 'type', 'latency', 'capacity',
          'errorRate', 'self', 'config', 'true', 'false'
        ],
        typeKeywords: [
          'string', 'number', 'boolean', 'any', 'void'
        ],
        operators: [
          '=', '>', '<', '!', '~', '?', ':', '==', '<=', '>=', '!=',
          '&&', '||', '++', '--', '+', '-', '*', '/', '&', '|', '^', '%',
          '<<', '>>', '>>>', '+=', '-=', '*=', '/=', '&=', '|=', '^=',
          '%=', '<<=', '>>=', '>>>='
        ],
        tokenizer: {
          root: [
            [/\/\/.*$/, 'comment'],
            [/\/\*/, 'comment', '@comment'],
            [/[a-zA-Z_]\w*/, { 
              cases: { 
                '@keywords': 'keyword',
                '@typeKeywords': 'type',
                '@default': 'identifier' 
              } 
            }],
            [/:=/, 'operator'],
            [/->/, 'operator'],
            [/"[^"]*"/, 'string'],
            [/'[^']*'/, 'string'],
            [/\d+(\.\d+)?/, 'number'],
            [/{|}|\[|\]|\(|\)/, 'delimiter'],
            [/[<>]=?|[!=]=|&&|\|\||[+\-*\/%]/, 'operator']
          ],
          comment: [
            [/[^\/*]+/, 'comment'],
            [/\*\//, 'comment', '@pop'],
            [/[\/*]/, 'comment']
          ]
        }
      });
    }

    // Set up resize observer
    this.resizeObserver = new ResizeObserver(() => {
      this.editor?.layout();
    });
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
  }

  onStateChange(_state: AppState, _changedKeys: string[]): void {
    // SDL editor doesn't need to respond to state changes
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
          value: this.sdlContent,
          language: 'sdl',
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

  public updateContent(content: string): void {
    this.sdlContent = content;
    if (this.editor) {
      this.editor.setValue(content);
    }
  }

  public getContent(): string {
    return this.editor?.getValue() || this.sdlContent;
  }
}