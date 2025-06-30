/**
 * Console Panel Component
 * Displays SDL compilation results, validation errors, and runtime logs
 */
export class ConsolePanel {
  private element: HTMLElement;
  private outputElement: HTMLElement | null = null;
  private maxMessages = 1000;
  // private messageTypes = new Set(['log', 'error', 'warning', 'info', 'success']);
  private activeFilters = new Set(['log', 'error', 'warning', 'info', 'success']);

  constructor(container: HTMLElement) {
    this.element = container;
    this.render();
    this.setupEventListeners();
  }

  private render() {
    this.element.innerHTML = `
      <div class="console-panel flex flex-col h-full bg-gray-900">
        <div class="console-header flex items-center justify-between p-2 border-b border-gray-700">
          <div class="flex items-center gap-2">
            <span class="text-gray-400 text-sm font-medium">Console</span>
            <div class="flex gap-1">
              ${this.renderFilterButtons()}
            </div>
          </div>
          <button class="clear-button p-1 hover:bg-gray-800 rounded" title="Clear console">
            <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
            </svg>
          </button>
        </div>
        <div class="console-output flex-1 overflow-auto p-2 font-mono text-sm" id="console-output">
          <div class="text-gray-500">SDL Console Ready</div>
        </div>
      </div>
    `;

    this.outputElement = this.element.querySelector('#console-output');
  }

  private renderFilterButtons(): string {
    const filters = [
      { type: 'log', color: 'gray', icon: '📋' },
      { type: 'error', color: 'red', icon: '❌' },
      { type: 'warning', color: 'yellow', icon: '⚠️' },
      { type: 'info', color: 'blue', icon: 'ℹ️' },
      { type: 'success', color: 'green', icon: '✅' }
    ];

    return filters.map(f => `
      <button class="filter-button p-1 text-xs rounded ${this.activeFilters.has(f.type) ? `bg-${f.color}-900 text-${f.color}-400` : 'text-gray-600'}" 
              data-type="${f.type}" title="Toggle ${f.type} messages">
        <span>${f.icon}</span>
      </button>
    `).join('');
  }

  private setupEventListeners() {
    // Clear button
    const clearButton = this.element.querySelector('.clear-button');
    clearButton?.addEventListener('click', () => this.clear());

    // Filter buttons
    this.element.querySelectorAll('.filter-button').forEach(button => {
      button.addEventListener('click', (e) => {
        const type = (e.currentTarget as HTMLElement).dataset.type;
        if (type) {
          this.toggleFilter(type);
        }
      });
    });
  }

  private toggleFilter(type: string) {
    if (this.activeFilters.has(type)) {
      this.activeFilters.delete(type);
    } else {
      this.activeFilters.add(type);
    }
    this.updateFilterButtons();
    this.updateMessageVisibility();
  }

  private updateFilterButtons() {
    this.element.querySelectorAll('.filter-button').forEach(button => {
      const type = (button as HTMLElement).dataset.type;
      if (type && this.activeFilters.has(type)) {
        button.classList.add('active');
      } else {
        button.classList.remove('active');
      }
    });
  }

  private updateMessageVisibility() {
    if (!this.outputElement) return;
    
    this.outputElement.querySelectorAll('.console-message').forEach(message => {
      const msgType = (message as HTMLElement).dataset.type || 'log';
      if (this.activeFilters.has(msgType)) {
        (message as HTMLElement).style.display = 'block';
      } else {
        (message as HTMLElement).style.display = 'none';
      }
    });
  }

  public log(message: string, type: 'log' | 'error' | 'warning' | 'info' | 'success' = 'log') {
    if (!this.outputElement) return;

    const timestamp = new Date().toLocaleTimeString();
    const messageDiv = document.createElement('div');
    messageDiv.className = `console-message console-${type}`;
    messageDiv.dataset.type = type;
    
    const colorClass = {
      log: 'text-gray-300',
      error: 'text-red-400',
      warning: 'text-yellow-400',
      info: 'text-blue-400',
      success: 'text-green-400'
    }[type];

    messageDiv.innerHTML = `
      <span class="text-gray-600">[${timestamp}]</span>
      <span class="${colorClass} ml-2">${this.escapeHtml(message)}</span>
    `;

    this.outputElement.appendChild(messageDiv);
    
    // Remove old messages if exceeding limit
    const messages = this.outputElement.querySelectorAll('.console-message');
    if (messages.length > this.maxMessages) {
      messages[0].remove();
    }

    // Auto-scroll to bottom
    this.outputElement.scrollTop = this.outputElement.scrollHeight;

    // Apply filter visibility
    if (!this.activeFilters.has(type)) {
      messageDiv.style.display = 'none';
    }
  }

  public error(message: string) {
    this.log(message, 'error');
  }

  public warning(message: string) {
    this.log(message, 'warning');
  }

  public info(message: string) {
    this.log(message, 'info');
  }

  public success(message: string) {
    this.log(message, 'success');
  }

  public clear() {
    if (this.outputElement) {
      this.outputElement.innerHTML = '<div class="text-gray-500">Console cleared</div>';
    }
  }

  private escapeHtml(text: string): string {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  public dispose() {
    // Cleanup if needed
  }
}

// Global console interceptor for WASM mode
export class ConsoleInterceptor {
  private consolePanel: ConsolePanel | null = null;
  private originalLog: typeof console.log;
  private originalError: typeof console.error;
  private originalWarn: typeof console.warn;
  private originalInfo: typeof console.info;

  constructor() {
    this.originalLog = console.log;
    this.originalError = console.error;
    this.originalWarn = console.warn;
    this.originalInfo = console.info;
  }

  public attach(consolePanel: ConsolePanel) {
    this.consolePanel = consolePanel;

    // Intercept console methods
    console.log = (...args) => {
      this.originalLog(...args);
      if (this.consolePanel) {
        this.consolePanel.log(args.map(a => String(a)).join(' '), 'log');
      }
    };

    console.error = (...args) => {
      this.originalError(...args);
      if (this.consolePanel) {
        this.consolePanel.error(args.map(a => String(a)).join(' '));
      }
    };

    console.warn = (...args) => {
      this.originalWarn(...args);
      if (this.consolePanel) {
        this.consolePanel.warning(args.map(a => String(a)).join(' '));
      }
    };

    console.info = (...args) => {
      this.originalInfo(...args);
      if (this.consolePanel) {
        this.consolePanel.info(args.map(a => String(a)).join(' '));
      }
    };
  }

  public detach() {
    console.log = this.originalLog;
    console.error = this.originalError;
    console.warn = this.originalWarn;
    console.info = this.originalInfo;
    this.consolePanel = null;
  }
}

// CSS styles for console panel
const style = document.createElement('style');
style.textContent = `
  .console-panel {
    background: #1a1a1a;
  }

  .console-output {
    font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
    line-height: 1.4;
  }

  .console-message {
    padding: 2px 0;
    white-space: pre-wrap;
    word-wrap: break-word;
  }

  .filter-button {
    transition: all 0.2s;
  }

  .filter-button.active {
    opacity: 1;
  }

  .filter-button:not(.active) {
    opacity: 0.5;
  }

  .console-output::-webkit-scrollbar {
    width: 8px;
  }

  .console-output::-webkit-scrollbar-track {
    background: #2a2a2a;
  }

  .console-output::-webkit-scrollbar-thumb {
    background: #4a4a4a;
    border-radius: 4px;
  }

  .console-output::-webkit-scrollbar-thumb:hover {
    background: #5a5a5a;
  }
`;
document.head.appendChild(style);
