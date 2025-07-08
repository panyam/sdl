import * as monaco from 'monaco-editor';
import { DockviewApi } from 'dockview-core';
import './tabbed-editor.css';

export interface FileTab {
  path: string;
  fsId: string;
  fsName?: string;
  content: string;
  modified: boolean;
  isRecipe?: boolean;
  isRunning?: boolean;
  currentLine?: number;
  editor?: monaco.editor.IStandaloneCodeEditor;
  model?: monaco.editor.ITextModel;
  viewState?: monaco.editor.ICodeEditorViewState;
  decorations?: string[];
  executionDecorations?: string[];
  errorDecorations?: string[];
}

export class TabbedEditor {
  private container: HTMLElement;
  private dockview: DockviewApi | null = null;
  private tabs: Map<string, FileTab> = new Map(); // Key is now fsId:path
  private activeTab: string | null = null;
  private onChange?: (path: string, content: string, modified: boolean, fsId?: string) => void;
  private onTabSwitch?: (path: string, fsId: string) => void;
  private onRecipeContentChange?: (tabKey: string) => void;
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

    // Register Recipe language
    monaco.languages.register({ id: 'recipe' });

    // Define Recipe syntax highlighting
    monaco.languages.setMonarchTokensProvider('recipe', {
      keywords: ['sdl', 'echo', 'read'],
      
      tokenizer: {
        root: [
          // Comments
          [/^#.*$/, 'comment'],
          
          // SDL commands
          [/^sdl\s+(load|use|gen|metrics|set|canvas)/, 'keyword'],
          
          // Echo statements
          [/^echo\s+/, 'keyword'],
          
          // Read (pause) statements
          [/^read\s*$/, 'keyword'],
          
          // Strings in quotes
          [/"([^"\\]|\\.)*"/, 'string'],
          [/'([^'\\]|\\.)*'/, 'string'],
          
          // Numbers
          [/\d+(\.\d+)?/, 'number'],
          
          // Variables (shown as errors)
          [/\$\w+/, 'invalid'],
          [/\$\{[^}]+\}/, 'invalid'],
          
          // Unsupported shell constructs
          [/^\s*(if|for|while|case|function)\s+/, 'invalid'],
          [/\|/, 'invalid'],
          [/>>?/, 'invalid'],
          [/</, 'invalid'],
          [/&\s*$/, 'invalid'],
          
          // Component.method patterns
          [/\b\w+\.\w+\b/, 'type'],
          
          // Flags
          [/--\w+/, 'attribute'],
          
          // Default
          [/.*/, 'text']
        ]
      }
    });

    // Define Recipe theme
    monaco.editor.defineTheme('recipe-dark', {
      base: 'vs-dark',
      inherit: true,
      rules: [
        { token: 'comment', foreground: '6A9955' },
        { token: 'keyword', foreground: '569CD6' },
        { token: 'string', foreground: 'CE9178' },
        { token: 'number', foreground: 'B5CEA8' },
        { token: 'type', foreground: '4EC9B0' },
        { token: 'attribute', foreground: '9CDCFE' },
        { token: 'invalid', foreground: 'FF6B6B', fontStyle: 'bold' },
        { token: 'text', foreground: 'D4D4D4' }
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

  setChangeHandler(handler: (path: string, content: string, modified: boolean, fsId?: string) => void) {
    this.onChange = handler;
  }

  setTabSwitchHandler(handler: (path: string, fsId: string) => void) {
    this.onTabSwitch = handler;
  }

  setRecipeContentChangeHandler(handler: (tabKey: string) => void) {
    this.onRecipeContentChange = handler;
  }


  async openFile(path: string, content: string, readOnly: boolean = false, fsId: string = 'local', fsName?: string) {
    const tabKey = `${fsId}:${path}`;
    
    // Check if file is already open
    if (this.tabs.has(tabKey)) {
      this.switchToTab(tabKey);
      return;
    }

    // Create a new tab
    this.createTabElement(tabKey, path, fsId, fsName);
    
    // Create editor container
    const editorContainer = document.createElement('div');
    editorContainer.className = 'h-full';
    editorContainer.style.display = 'none'; // Initially hidden
    
    // Determine language and theme based on file extension
    let language = 'plaintext';
    let theme = 'vs-dark';
    
    if (path.endsWith('.sdl')) {
      language = 'sdl';
      theme = 'sdl-dark';
    } else if (path.endsWith('.recipe')) {
      language = 'recipe';
      theme = 'recipe-dark';
    }
    
    // Create Monaco editor instance
    const editor = monaco.editor.create(editorContainer, {
      value: content,
      language: language,
      theme: theme,
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
      const tab = this.tabs.get(tabKey);
      if (tab) {
        if (!tab.modified) {
          tab.modified = true;
          this.updateTabTitle(tabKey);
          if (this.onChange) {
            this.onChange(path, editor.getValue(), true, fsId);
          }
        }
        
        // Notify about recipe content changes for validation
        if (path.endsWith('.recipe') && this.onRecipeContentChange) {
          this.onRecipeContentChange(tabKey);
        }
      }
    });

    // Add keyboard shortcuts
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      this.saveTab(tabKey);
    });

    // Store tab info
    const isRecipe = path.endsWith('.recipe');
    this.tabs.set(tabKey, {
      path,
      fsId,
      fsName,
      content,
      modified: false,
      isRecipe,
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
    this.switchToTab(tabKey);
  }

  private createTabElement(tabKey: string, path: string, fsId: string, fsName?: string) {
    if (!this.tabBar) return;
    
    const fileName = path.split('/').pop() || 'untitled';
    const tabTitle = fsName ? `${fsName}:${fileName}` : fileName;
    const tabElement = document.createElement('div');
    tabElement.className = 'tab';
    tabElement.dataset.tabKey = tabKey;
    tabElement.dataset.path = path;
    tabElement.dataset.fsId = fsId;
    
    tabElement.innerHTML = `
      <span class="tab-title text-sm text-gray-200">${tabTitle}</span>
      <button class="tab-close ml-2 text-gray-400 hover:text-gray-200" title="Close">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
          <path d="M8 8.707l3.646 3.647.708-.707L8.707 8l3.647-3.646-.707-.708L8 7.293 4.354 3.646l-.707.708L7.293 8l-3.647 3.646.708.708L8 8.707z"/>
        </svg>
      </button>
    `;
    
    // Tab click handler
    tabElement.addEventListener('click', (e) => {
      if (!(e.target as HTMLElement).closest('.tab-close')) {
        this.switchToTab(tabKey);
      }
    });
    
    // Close button handler
    const closeBtn = tabElement.querySelector('.tab-close');
    closeBtn?.addEventListener('click', (e) => {
      e.stopPropagation();
      this.closeTab(tabKey);
    });
    
    this.tabBar.appendChild(tabElement);
  }
  
  private updateTabTitle(tabKey: string) {
    const tab = this.tabs.get(tabKey);
    if (!tab) return;

    // Update tab element
    const tabElement = this.tabBar?.querySelector(`[data-tab-key="${CSS.escape(tabKey)}"]`);
    if (tabElement) {
      const titleElement = tabElement.querySelector('.tab-title');
      if (titleElement) {
        const fileName = tab.path.split('/').pop() || 'untitled';
        const baseTitle = tab.fsName ? `${tab.fsName}:${fileName}` : fileName;
        let displayTitle = baseTitle;
        
        // Add indicators
        if (tab.isRunning) {
          displayTitle = `▶ ${displayTitle}`;
        }
        if (tab.modified) {
          displayTitle = `${displayTitle} *`;
        }
        
        titleElement.textContent = displayTitle;
      }
    }
    
    // Update dockview panel title if this is the active tab
    if (this.activeTab === tabKey && this.dockview) {
      const panel = this.dockview.getPanel('codeEditor');
      if (panel) {
        const fileName = tab.path.split('/').pop() || 'untitled';
        const baseTitle = tab.fsName ? `${tab.fsName}:${fileName}` : fileName;
        let title = baseTitle;
        
        if (tab.isRunning) {
          title = `▶ ${title}`;
        }
        if (tab.modified) {
          title = `${title} *`;
        }
        
        panel.api.setTitle(title);
      }
    }
  }

  switchToTab(tabKey: string) {
    const tab = this.tabs.get(tabKey);
    if (!tab || !this.editorContainer) return;

    // Hide current editor
    if (this.activeTab && this.activeTab !== tabKey) {
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
      const currentTabElement = this.tabBar?.querySelector(`[data-tab-key="${CSS.escape(this.activeTab)}"]`);
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
    const newTabElement = this.tabBar?.querySelector(`[data-tab-key="${CSS.escape(tabKey)}"]`);
    if (newTabElement) {
      newTabElement.classList.add('active');
    }

    this.activeTab = tabKey;
    this.updateTabTitle(tabKey);
    
    // Notify about tab switch
    if (this.onTabSwitch && tab) {
      this.onTabSwitch(tab.path, tab.fsId);
    }
  }

  closeTab(tabKey: string) {
    const tab = this.tabs.get(tabKey);
    if (!tab) return;

    // Check for unsaved changes
    if (tab.modified) {
      const save = confirm(`Save changes to ${tab.path}?`);
      if (save) {
        this.saveTab(tabKey);
      }
    }

    // Remove tab element
    const tabElement = this.tabBar?.querySelector(`[data-tab-key="${CSS.escape(tabKey)}"]`);
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

    this.tabs.delete(tabKey);

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
      const tabKeys = Array.from(this.tabs.keys());
      const currentIndex = tabKeys.indexOf(tabKey);
      
      if (currentIndex > 0) {
        nextTab = tabKeys[currentIndex - 1];
      } else if (tabKeys.length > 0) {
        nextTab = tabKeys[0];
      }
      
      if (nextTab) {
        this.switchToTab(nextTab);
      }
    }
  }

  saveTab(tabKey: string) {
    const tab = this.tabs.get(tabKey);
    if (!tab || !tab.modified) return;

    tab.modified = false;
    this.updateTabTitle(tabKey);
    
    if (this.onChange) {
      this.onChange(tab.path, tab.editor?.getValue() || '', false, tab.fsId);
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

  activeTabHasUnsavedChanges(): boolean {
    if (!this.activeTab) return false;
    const tab = this.tabs.get(this.activeTab);
    return tab?.modified || false;
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

  setRecipeRunning(tabKey: string, isRunning: boolean, currentLine?: number) {
    const tab = this.tabs.get(tabKey);
    if (!tab) return;
    
    console.log(`setRecipeRunning called - tabKey: ${tabKey}, isRunning: ${isRunning}, currentLine: ${currentLine}`);
    
    tab.isRunning = isRunning;
    tab.currentLine = currentLine;
    
    // Update tab title
    this.updateTabTitle(tabKey);
    
    // Update line highlighting
    if (tab.editor) {
      // Clear previous execution decorations
      if (tab.executionDecorations) {
        tab.editor.deltaDecorations(tab.executionDecorations, []);
        tab.executionDecorations = undefined;
      }
      
      // Add new decoration for current line
      if (isRunning && currentLine) {
        console.log(`Setting recipe line decoration for line ${currentLine}`);
        const newDecorations = [{
          range: new monaco.Range(currentLine, 1, currentLine, 1),
          options: {
            isWholeLine: true,
            className: 'recipe-current-line',
            glyphMarginClassName: 'recipe-current-line-glyph'
          }
        }];
        
        tab.executionDecorations = tab.editor.deltaDecorations([], newDecorations);
        console.log(`Decorations set: ${tab.executionDecorations}`);
        
        // Scroll to line
        tab.editor.revealLineInCenter(currentLine);
      }
    } else {
      console.log('No editor found for tab');
    }
  }

  getActiveTabKey(): string | null {
    return this.activeTab;
  }

  updateErrorDecorations(tabKey: string, errors: Array<{lineNumber: number, message: string, severity: 'error' | 'warning'}>) {
    const tab = this.tabs.get(tabKey);
    if (!tab || !tab.editor) return;
    
    // Clear existing error decorations
    if (tab.errorDecorations) {
      tab.editor.deltaDecorations(tab.errorDecorations, []);
    }
    
    // Create model markers for errors
    const model = tab.editor.getModel();
    if (!model) return;
    
    // Convert errors to Monaco markers
    const markers = errors.map(error => ({
      severity: error.severity === 'error' ? monaco.MarkerSeverity.Error : monaco.MarkerSeverity.Warning,
      startLineNumber: error.lineNumber,
      startColumn: 1,
      endLineNumber: error.lineNumber,
      endColumn: model.getLineMaxColumn(error.lineNumber),
      message: error.message,
      source: 'Recipe Validator'
    }));
    
    // Set markers on the model
    monaco.editor.setModelMarkers(model, 'recipe-validator', markers);
    
    // Also add decorations for visual feedback
    const decorations = errors.map(error => ({
      range: new monaco.Range(error.lineNumber, 1, error.lineNumber, 1),
      options: {
        isWholeLine: true,
        className: error.severity === 'error' ? 'recipe-error-line' : 'recipe-warning-line',
        glyphMarginClassName: error.severity === 'error' ? 'recipe-error-glyph' : 'recipe-warning-glyph',
        glyphMarginHoverMessage: { value: error.message }
      }
    }));
    
    tab.errorDecorations = tab.editor.deltaDecorations([], decorations);
  }

  clearErrorDecorations(tabKey: string) {
    const tab = this.tabs.get(tabKey);
    if (!tab || !tab.editor) return;
    
    // Clear decorations
    if (tab.errorDecorations) {
      tab.editor.deltaDecorations(tab.errorDecorations, []);
      tab.errorDecorations = undefined;
    }
    
    // Clear markers
    const model = tab.editor.getModel();
    if (model) {
      monaco.editor.setModelMarkers(model, 'recipe-validator', []);
    }
  }
}