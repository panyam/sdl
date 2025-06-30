import * as monaco from 'monaco-editor';

export class CodeEditor {
  private container: HTMLElement;
  private editor: monaco.editor.IStandaloneCodeEditor | null = null;
  private currentFile: string | null = null;
  private onChange?: (content: string) => void;
  private modified: boolean = false;
  private isReadOnly: boolean = false;

  constructor(container: HTMLElement) {
    this.container = container;
    this.initializeEditor();
  }

  private initializeEditor() {
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

    // Define SDL theme (optional - uses VS Code dark theme by default)
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

    // Create editor
    this.editor = monaco.editor.create(this.container, {
      value: '// Welcome to SDL Editor\n// Load a file from the file explorer or start typing...\n',
      language: 'sdl',
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
      lineNumbersMinChars: 3
    });

    // Track changes
    this.editor.onDidChangeModelContent(() => {
      this.modified = true;
      if (this.onChange) {
        this.onChange(this.editor!.getValue());
      }
    });

    // Add keyboard shortcuts
    this.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      this.save();
    });
  }

  setChangeHandler(handler: (content: string) => void) {
    this.onChange = handler;
  }

  async loadFile(path: string, content: string, readOnly?: boolean) {
    this.currentFile = path;
    this.modified = false;
    
    // Check if file is readonly
    if (readOnly === undefined && (window as any).SDL && (window as any).SDL.fs) {
      try {
        const result = await (window as any).SDL.fs.isReadOnly(path);
        this.isReadOnly = result.success && result.isReadOnly;
      } catch (err) {
        // Default check by prefix
        this.isReadOnly = path.startsWith('/examples/') || 
                         path.startsWith('/lib/') || 
                         path.startsWith('/demos/');
      }
    } else {
      this.isReadOnly = readOnly || false;
    }
    
    if (this.editor) {
      this.editor.setValue(content);
      this.editor.setPosition({ lineNumber: 1, column: 1 });
      this.editor.updateOptions({ readOnly: this.isReadOnly });
      
      // Update editor title/status
      this.updateStatus();
    }
  }

  getValue(): string {
    return this.editor?.getValue() || '';
  }

  setValue(content: string) {
    if (this.editor) {
      this.editor.setValue(content);
    }
  }

  save() {
    if (this.isReadOnly) {
      console.warn('Cannot save readonly file:', this.currentFile);
      // Could show a dialog to save as a copy
      return;
    }
    
    if (this.modified && this.onChange) {
      this.onChange(this.getValue());
      this.modified = false;
      this.updateStatus();
    }
  }

  private updateStatus() {
    // Could update a status bar or indicator
    const fileName = this.currentFile ? this.currentFile.split('/').pop() : 'Untitled';
    const dirtyIndicator = this.modified ? ' ●' : '';
    const readOnlyIndicator = this.isReadOnly ? ' [Read Only]' : '';
    console.log(`Editing: ${fileName}${dirtyIndicator}${readOnlyIndicator}`);
  }

  dispose() {
    if (this.editor) {
      this.editor.dispose();
    }
  }

  // Utility methods
  format() {
    if (this.editor) {
      this.editor.getAction('editor.action.formatDocument')?.run();
    }
  }

  undo() {
    if (this.editor) {
      this.editor.trigger('', 'undo', null);
    }
  }

  redo() {
    if (this.editor) {
      this.editor.trigger('', 'redo', null);
    }
  }

  findReplace() {
    if (this.editor) {
      this.editor.getAction('editor.action.startFindReplaceAction')?.run();
    }
  }
}

// Export helper to configure Monaco loader
export function configureMonacoLoader() {
  // Configure Monaco loader to use CDN
  (window as any).require = { paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs' } };
}