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
}

export class TabbedEditor {
  private container: HTMLElement;
  private dockview: DockviewApi | null = null;
  private tabs: Map<string, FileTab> = new Map(); // Key is now fsId:path
  private activeTab: string | null = null;
  private onChange?: (path: string, content: string, modified: boolean, fsId?: string) => void;
  private onTabSwitch?: (path: string, fsId: string) => void;
  private onRecipeAction?: (action: 'run' | 'stop' | 'step' | 'restart', content: string) => void;
  private tabBar: HTMLElement | null = null;
  private editorContainer: HTMLElement | null = null;
  private recipeToolbar: HTMLElement | null = null;
  
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
        <div class="recipe-toolbar hidden bg-gray-800 border-b border-gray-700 p-2" style="flex-shrink: 0;"></div>
        <div class="editor-content flex-1 overflow-hidden"></div>
      </div>
    `;
    
    this.tabBar = this.container.querySelector('.tab-bar');
    this.editorContainer = this.container.querySelector('.editor-content');
    this.recipeToolbar = this.container.querySelector('.recipe-toolbar');
    
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

  setRecipeActionHandler(handler: (action: 'run' | 'stop' | 'step' | 'restart', content: string) => void) {
    this.onRecipeAction = handler;
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
      const tab = this.tabs.get(tabKey);
      if (tab && !tab.modified) {
        tab.modified = true;
        this.updateTabTitle(tabKey);
        if (this.onChange) {
          this.onChange(path, editor.getValue(), true, fsId);
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
    
    // Update recipe toolbar visibility
    this.updateRecipeToolbar();
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

  private updateRecipeToolbar() {
    if (!this.recipeToolbar || !this.activeTab) return;
    
    const tab = this.tabs.get(this.activeTab);
    if (!tab || !tab.isRecipe) {
      this.recipeToolbar.classList.add('hidden');
      return;
    }
    
    // Show recipe toolbar
    this.recipeToolbar.classList.remove('hidden');
    
    // Update toolbar content based on running state
    if (tab.isRunning) {
      this.recipeToolbar.innerHTML = `
        <div class="flex items-center gap-2">
          <button class="recipe-btn" onclick="window.tabbedEditor?.handleRecipeAction('step')">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M5 3l3 2.5L5 8V3zm4 0l3 2.5L9 8V3zm4 0h2v5h-2V3z"/>
            </svg>
            Step
          </button>
          <button class="recipe-btn recipe-btn-danger" onclick="window.tabbedEditor?.handleRecipeAction('stop')">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M4 4h8v8H4V4z"/>
            </svg>
            Stop
          </button>
          <button class="recipe-btn" onclick="window.tabbedEditor?.handleRecipeAction('restart')">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M13.451 5.609l-.579-.939-1.068.812-.076.094c-.335.415-.927 1.341-1.124 2.876l-.021.165.033.163c.071.363.224.694.456.97l.087.102c.25.282.554.514.897.683l.123.061c.404.182.852.279 1.312.279.51 0 1.003-.12 1.444-.349l.105-.059c.435-.255.785-.618 1.014-1.051l.063-.119c.185-.38.283-.8.283-1.228 0-.347-.063-.684-.183-1.003l-.056-.147-.098-.245zm-3.177 3.342c-.169 0-.331-.037-.48-.109l-.044-.023c-.122-.061-.227-.145-.313-.249l-.032-.04c-.084-.106-.144-.227-.176-.361l-.012-.056c-.03-.137-.037-.283-.01-.428l.008-.059c.088-.987.373-1.76.603-2.122.183.338.276.735.276 1.142 0 .168-.02.332-.06.491l-.023.079c-.082.268-.225.51-.417.703l-.037.035c-.189.186-.423.325-.689.413l-.064.021c-.14.042-.288.063-.44.063zm1.373-4.326l2.255-1.718 1.017 1.647-2.351 1.79-.921-1.719zm-10.296.577l1.017-1.647 2.255 1.718-.921 1.719-2.351-1.79z"/>
            </svg>
            Restart
          </button>
          <div class="flex-1"></div>
          <span class="text-xs text-gray-400">Line ${tab.currentLine || 0}</span>
        </div>
      `;
    } else {
      this.recipeToolbar.innerHTML = `
        <div class="flex items-center gap-2">
          <button class="recipe-btn recipe-btn-primary" onclick="window.tabbedEditor?.handleRecipeAction('run')">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M5 3l8 5-8 5V3z"/>
            </svg>
            Run
          </button>
          <div class="flex-1"></div>
          <span class="text-xs text-gray-400">SDL Recipe</span>
        </div>
      `;
    }
    
    // Make this instance globally accessible for button handlers
    (window as any).tabbedEditor = this;
  }

  handleRecipeAction(action: 'run' | 'stop' | 'step' | 'restart') {
    if (!this.activeTab || !this.onRecipeAction) return;
    
    const tab = this.tabs.get(this.activeTab);
    if (!tab || !tab.editor) return;
    
    const content = tab.editor.getValue();
    this.onRecipeAction(action, content);
  }

  setRecipeRunning(tabKey: string, isRunning: boolean, currentLine?: number) {
    const tab = this.tabs.get(tabKey);
    if (!tab) return;
    
    tab.isRunning = isRunning;
    tab.currentLine = currentLine;
    
    // Update tab title
    this.updateTabTitle(tabKey);
    
    // Update toolbar if this is the active tab
    if (tabKey === this.activeTab) {
      this.updateRecipeToolbar();
    }
    
    // Update line highlighting
    if (tab.editor) {
      // Clear previous decorations
      if (tab.decorations) {
        tab.editor.deltaDecorations(tab.decorations, []);
      }
      
      // Add new decoration for current line
      if (isRunning && currentLine) {
        tab.decorations = tab.editor.deltaDecorations([], [
          {
            range: new monaco.Range(currentLine, 1, currentLine, 1),
            options: {
              isWholeLine: true,
              className: 'recipe-current-line',
              glyphMarginClassName: 'recipe-current-line-glyph'
            }
          }
        ]);
        
        // Scroll to line
        tab.editor.revealLineInCenter(currentLine);
      }
    }
  }

  getActiveTabKey(): string | null {
    return this.activeTab;
  }
}