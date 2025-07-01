import * as monaco from 'monaco-editor';
import { DockviewApi } from 'dockview-core';
import './tabbed-editor.css';

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
  private tabBar: HTMLElement | null = null;
  private editorContainer: HTMLElement | null = null;
  
  constructor(container: HTMLElement, dockview: DockviewApi) {
    this.container = container;
    this.dockview = dockview;
    this.initializeMonaco();
    this.initializeLayout();
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

  private initializeLayout() {
    // Create layout structure
    this.container.innerHTML = `
      <div class="editor-layout flex flex-col h-full">
        <div class="tab-bar flex items-center bg-gray-800 border-b border-gray-700 overflow-x-auto" style="height: 35px; flex-shrink: 0;"></div>
        <div class="editor-content flex-1 overflow-hidden"></div>
      </div>
    `;
    
    this.tabBar = this.container.querySelector('.tab-bar');
    this.editorContainer = this.container.querySelector('.editor-content');
    
    this.showWelcome();
  }
  
  private showWelcome() {
    if (!this.editorContainer) return;
    
    // Show welcome message in the editor container
    this.editorContainer.innerHTML = `
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
    this.createTabElement(path);
    
    // Create editor container
    const editorContainer = document.createElement('div');
    editorContainer.className = 'h-full';
    editorContainer.style.display = 'none'; // Initially hidden
    
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

    // Add editor to container
    if (this.editorContainer) {
      // Clear welcome message if this is the first tab
      if (this.tabs.size === 1) {
        this.editorContainer.innerHTML = '';
      }
      this.editorContainer.appendChild(editorContainer);
    }

    // Switch to the new tab
    this.switchToTab(path);
  }

  private createTabElement(path: string) {
    if (!this.tabBar) return;
    
    const fileName = path.split('/').pop() || 'untitled';
    const tabElement = document.createElement('div');
    tabElement.className = 'tab';
    tabElement.dataset.path = path;
    
    tabElement.innerHTML = `
      <span class="tab-title text-sm text-gray-200">${fileName}</span>
      <button class="tab-close ml-2 text-gray-400 hover:text-gray-200" title="Close">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
          <path d="M8 8.707l3.646 3.647.708-.707L8.707 8l3.647-3.646-.707-.708L8 7.293 4.354 3.646l-.707.708L7.293 8l-3.647 3.646.708.708L8 8.707z"/>
        </svg>
      </button>
    `;
    
    // Tab click handler
    tabElement.addEventListener('click', (e) => {
      if (!(e.target as HTMLElement).closest('.tab-close')) {
        this.switchToTab(path);
      }
    });
    
    // Close button handler
    const closeBtn = tabElement.querySelector('.tab-close');
    closeBtn?.addEventListener('click', (e) => {
      e.stopPropagation();
      this.closeTab(path);
    });
    
    this.tabBar.appendChild(tabElement);
  }
  
  private updateTabTitle(path: string) {
    const tab = this.tabs.get(path);
    if (!tab) return;

    // Update tab element
    const tabElement = this.tabBar?.querySelector(`[data-path="${CSS.escape(path)}"]`);
    if (tabElement) {
      const titleElement = tabElement.querySelector('.tab-title');
      if (titleElement) {
        const fileName = path.split('/').pop() || 'untitled';
        titleElement.textContent = tab.modified ? `${fileName} *` : fileName;
      }
    }
    
    // Update dockview panel title if this is the active tab
    if (this.activeTab === path && this.dockview) {
      const panel = this.dockview.getPanel('codeEditor');
      if (panel) {
        const fileName = path.split('/').pop() || 'untitled';
        const title = tab.modified ? `${fileName} *` : fileName;
        panel.api.setTitle(title);
      }
    }
  }

  switchToTab(path: string) {
    const tab = this.tabs.get(path);
    if (!tab || !this.editorContainer) return;

    // Hide current editor
    if (this.activeTab && this.activeTab !== path) {
      const currentTab = this.tabs.get(this.activeTab);
      if (currentTab?.editor) {
        // Save view state
        currentTab.viewState = currentTab.editor.saveViewState() || undefined;
        // Hide the editor container
        const currentContainer = currentTab.editor.getContainerDomNode();
        if (currentContainer) {
          currentContainer.style.display = 'none';
        }
      }
      
      // Update tab appearance
      const currentTabElement = this.tabBar?.querySelector(`[data-path="${CSS.escape(this.activeTab)}"]`);
      if (currentTabElement) {
        currentTabElement.classList.remove('active');
      }
    }

    // Show new editor
    if (tab.editor) {
      const editorContainer = tab.editor.getContainerDomNode();
      if (editorContainer) {
        editorContainer.style.display = 'block';
      }
      
      // Restore view state
      if (tab.viewState) {
        tab.editor.restoreViewState(tab.viewState);
      }
      
      tab.editor.layout();
      tab.editor.focus();
    }
    
    // Update tab appearance
    const newTabElement = this.tabBar?.querySelector(`[data-path="${CSS.escape(path)}"]`);
    if (newTabElement) {
      newTabElement.classList.add('active');
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

    // Remove tab element
    const tabElement = this.tabBar?.querySelector(`[data-path="${CSS.escape(path)}"]`);
    if (tabElement) {
      tabElement.remove();
    }

    // Remove editor container
    if (tab.editor) {
      const editorContainer = tab.editor.getContainerDomNode();
      if (editorContainer && editorContainer.parentNode) {
        editorContainer.parentNode.removeChild(editorContainer);
      }
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
      
      // Update panel title
      if (this.dockview) {
        const panel = this.dockview.getPanel('codeEditor');
        if (panel) {
          panel.api.setTitle('Code Editor');
        }
      }
    } else {
      // Switch to the next tab (prefer tabs to the right)
      let nextTab: string | null = null;
      const tabPaths = Array.from(this.tabs.keys());
      const currentIndex = tabPaths.indexOf(path);
      
      if (currentIndex > 0) {
        nextTab = tabPaths[currentIndex - 1];
      } else if (tabPaths.length > 0) {
        nextTab = tabPaths[0];
      }
      
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