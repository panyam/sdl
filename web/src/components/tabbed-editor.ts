import * as monaco from 'monaco-editor';
import { DockviewApi } from 'dockview-core';

export interface FileTab {
  path: string;
  content: string;
  modified: boolean;
  editor?: monaco.editor.IStandaloneCodeEditor;
  model?: monaco.editor.ITextModel;
  viewState?: monaco.editor.ICodeEditorViewState;
}

export class TabbedEditor {
  private container: HTMLElement;
  private dockview: DockviewApi | null = null;
  private tabs: Map<string, FileTab> = new Map();
  private activeTab: string | null = null;
  private onChange?: (path: string, content: string, modified: boolean) => void;
  
  constructor(container: HTMLElement, dockview: DockviewApi) {
    this.container = container;
    this.dockview = dockview;
    this.initializeMonaco();
    this.showWelcome();
  }

  private initializeMonaco() {
    // Register SDL language
    monaco.languages.register({ id: 'sdl' });

    // Define SDL syntax highlighting
    monaco.languages.setMonarchTokensProvider('sdl', {
      keywords: [
        'component', 'system', 'method', 'use', 'import', 'let', 'return',
        'if', 'else', 'switch', 'case', 'default', 'for', 'go', 'wait',
        'sample', 'distribute', 'enum', 'type', 'latency', 'capacity',
        'errorRate', 'self', 'config', 'true', 'false'
      ],
      
      operators: ['=', '.', ',', ':', ';', '(', ')', '{', '}', '[', ']'],
      
      symbols: /[=><!~?:&|+\-*\/\^%]+/,
      
      tokenizer: {
        root: [
          // Comments
          [/\/\/.*$/, 'comment'],
          [/\/\*/, 'comment', '@comment'],
          
          // Keywords
          [/\b(component|system|method|use|import|let|return|if|else|switch|case|default|for|go|wait|sample|distribute|enum|type)\b/, 'keyword'],
          
          // Properties
          [/\b(latency|capacity|errorRate|config)\b/, 'type'],
          
          // Identifiers
          [/[a-zA-Z_]\w*/, {
            cases: {
              '@keywords': 'keyword',
              '@default': 'identifier'
            }
          }],
          
          // Numbers
          [/\d+(\.\d+)?(ms|s|m|h)?/, 'number'],
          [/\d+%/, 'number'],
          
          // Strings
          [/"([^"\\]|\\.)*$/, 'string.invalid'],
          [/"/, 'string', '@string'],
          
          // Delimiters
          [/[{}()\[\]]/, 'delimiter'],
          [/[<>](?!@symbols)/, 'delimiter'],
          
          // Operators
          [/@symbols/, {
            cases: {
              '@operators': 'operator',
              '@default': ''
            }
          }],
        ],
        
        comment: [
          [/[^\/*]+/, 'comment'],
          [/\*\//, 'comment', '@pop'],
          [/[\/*]/, 'comment']
        ],
        
        string: [
          [/[^\\"]+/, 'string'],
          [/\\./, 'string.escape'],
          [/"/, 'string', '@pop']
        ]
      }
    });

    // Define SDL theme
    monaco.editor.defineTheme('sdl-dark', {
      base: 'vs-dark',
      inherit: true,
      rules: [
        { token: 'comment', foreground: '6A9955' },
        { token: 'keyword', foreground: '569CD6' },
        { token: 'type', foreground: '4EC9B0' },
        { token: 'string', foreground: 'CE9178' },
        { token: 'number', foreground: 'B5CEA8' },
        { token: 'operator', foreground: 'D4D4D4' }
      ],
      colors: {
        'editor.background': '#1e1e1e'
      }
    });
  }

  private showWelcome() {
    // Show welcome message in the container
    this.container.innerHTML = `
      <div class="flex items-center justify-center h-full text-gray-400">
        <div class="text-center">
          <div class="text-lg mb-2">SDL Editor</div>
          <div class="text-sm">Open a file from the file explorer to get started</div>
        </div>
      </div>
    `;
  }

  setChangeHandler(handler: (path: string, content: string, modified: boolean) => void) {
    this.onChange = handler;
  }

  async openFile(path: string, content: string, readOnly: boolean = false) {
    // Check if file is already open
    if (this.tabs.has(path)) {
      this.switchToTab(path);
      return;
    }

    // Create a new tab
    
    // Create editor container
    const editorContainer = document.createElement('div');
    editorContainer.className = 'h-full';
    
    // Create Monaco editor instance
    const editor = monaco.editor.create(editorContainer, {
      value: content,
      language: path.endsWith('.sdl') ? 'sdl' : 'plaintext',
      theme: 'sdl-dark',
      automaticLayout: true,
      minimap: { enabled: false },
      fontSize: 14,
      fontFamily: "'Monaco', 'Menlo', 'Ubuntu Mono', monospace",
      scrollBeyondLastLine: false,
      renderWhitespace: 'selection',
      lineNumbers: 'on',
      glyphMargin: false,
      folding: true,
      lineDecorationsWidth: 0,
      lineNumbersMinChars: 3,
      readOnly: readOnly
    });

    const model = editor.getModel()!;
    
    // Track changes
    editor.onDidChangeModelContent(() => {
      const tab = this.tabs.get(path);
      if (tab && !tab.modified) {
        tab.modified = true;
        this.updateTabTitle(path);
        if (this.onChange) {
          this.onChange(path, editor.getValue(), true);
        }
      }
    });

    // Add keyboard shortcuts
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      this.saveTab(path);
    });

    // Store tab info
    this.tabs.set(path, {
      path,
      content,
      modified: false,
      editor,
      model
    });

    // Add tab to dockview
    if (this.dockview) {
      const existingPanel = this.dockview.getPanel('codeEditor');
      if (existingPanel) {
        // Replace the welcome content with editor tabs
        if (this.tabs.size === 1) {
          this.container.innerHTML = '';
          this.container.appendChild(editorContainer);
        } else {
          // For multiple tabs, we need a more complex solution
          // For now, just replace the content
          this.container.innerHTML = '';
          this.container.appendChild(editorContainer);
        }
      }
    }

    this.activeTab = path;
    this.updateTabTitle(path);
  }

  private updateTabTitle(path: string) {
    const tab = this.tabs.get(path);
    if (!tab || !this.dockview) return;

    const panel = this.dockview.getPanel('codeEditor');
    if (panel) {
      const fileName = path.split('/').pop() || 'untitled';
      const title = tab.modified ? `${fileName} *` : fileName;
      panel.api.setTitle(title);
    }
  }

  switchToTab(path: string) {
    const tab = this.tabs.get(path);
    if (!tab) return;

    // Hide current editor
    if (this.activeTab && this.activeTab !== path) {
      const currentTab = this.tabs.get(this.activeTab);
      if (currentTab?.editor) {
        // Save view state
        currentTab.viewState = currentTab.editor.saveViewState() || undefined;
      }
    }

    // Show new editor
    if (tab.editor) {
      this.container.innerHTML = '';
      this.container.appendChild(tab.editor.getContainerDomNode());
      
      // Restore view state
      if (tab.viewState) {
        tab.editor.restoreViewState(tab.viewState);
      }
      
      tab.editor.focus();
    }

    this.activeTab = path;
    this.updateTabTitle(path);
  }

  closeTab(path: string) {
    const tab = this.tabs.get(path);
    if (!tab) return;

    // Check for unsaved changes
    if (tab.modified) {
      const save = confirm(`Save changes to ${path}?`);
      if (save) {
        this.saveTab(path);
      }
    }

    // Dispose editor
    if (tab.editor) {
      tab.editor.dispose();
    }
    if (tab.model) {
      tab.model.dispose();
    }

    this.tabs.delete(path);

    // Switch to another tab or show welcome
    if (this.tabs.size === 0) {
      this.showWelcome();
      this.activeTab = null;
    } else {
      const nextTab = this.tabs.keys().next().value;
      if (nextTab) {
        this.switchToTab(nextTab);
      }
    }
  }

  saveTab(path: string) {
    const tab = this.tabs.get(path);
    if (!tab || !tab.modified) return;

    tab.modified = false;
    this.updateTabTitle(path);
    
    if (this.onChange) {
      this.onChange(path, tab.editor?.getValue() || '', false);
    }
  }

  getActiveTab(): string | null {
    return this.activeTab;
  }

  getActiveContent(): string | null {
    if (!this.activeTab) return null;
    const tab = this.tabs.get(this.activeTab);
    return tab?.editor?.getValue() || null;
  }

  hasUnsavedChanges(): boolean {
    for (const tab of this.tabs.values()) {
      if (tab.modified) return true;
    }
    return false;
  }

  dispose() {
    for (const tab of this.tabs.values()) {
      if (tab.editor) {
        tab.editor.dispose();
      }
      if (tab.model) {
        tab.model.dispose();
      }
    }
    this.tabs.clear();
  }
}