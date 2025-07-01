import { RecipeRunner, RecipeState, RecipeStep } from './recipe-runner.js';
import './recipe-panel.css';

export class RecipePanel {
  private container: HTMLElement;
  private runner: RecipeRunner;
  private state: RecipeState | null = null;
  private consoleOutput?: (message: string, type: 'info' | 'success' | 'error' | 'warning') => void;

  constructor(container: HTMLElement, runner: RecipeRunner) {
    this.container = container;
    this.runner = runner;
    
    // Set up event handlers
    this.runner.setStateChangeHandler((state) => this.onStateChange(state));
    this.runner.setOutputHandler((msg, type) => this.onOutput(msg, type));
    
    this.render();
  }

  setConsoleOutput(handler: (message: string, type: 'info' | 'success' | 'error' | 'warning') => void) {
    this.consoleOutput = handler;
  }

  private onStateChange(state: RecipeState) {
    this.state = state;
    this.render();
    
    // Auto-scroll to current step
    if (state.currentStep > 0) {
      const currentElement = this.container.querySelector(`[data-step="${state.currentStep - 1}"]`);
      if (currentElement) {
        currentElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
    }
  }

  private onOutput(message: string, type: 'info' | 'success' | 'error' | 'warning') {
    if (this.consoleOutput) {
      this.consoleOutput(message, type);
    }
  }

  private render() {
    if (!this.state) {
      this.container.innerHTML = `
        <div class="recipe-panel h-full flex items-center justify-center">
          <div class="text-center text-gray-400">
            <div class="text-lg mb-2">No Recipe Loaded</div>
            <div class="text-sm">Open a .recipe file to start</div>
          </div>
        </div>
      `;
      return;
    }

    const { mode, currentStep, steps, fileName, autoDelay } = this.state;
    const progress = steps.length > 0 ? (currentStep / steps.length) * 100 : 0;

    this.container.innerHTML = `
      <div class="recipe-panel h-full flex flex-col">
        <!-- Header -->
        <div class="recipe-header p-3 bg-gray-800 border-b border-gray-700">
          <div class="flex items-center justify-between mb-2">
            <h3 class="text-sm font-semibold">${fileName}</h3>
            <span class="text-xs text-gray-400">${currentStep}/${steps.length} steps</span>
          </div>
          
          <!-- Progress bar -->
          <div class="w-full bg-gray-700 rounded h-1 mb-3">
            <div class="bg-blue-500 h-1 rounded transition-all duration-300" style="width: ${progress}%"></div>
          </div>
          
          <!-- Controls -->
          <div class="flex items-center gap-2">
            ${this.renderControlButtons()}
            
            <div class="flex-1"></div>
            
            <!-- Mode selector -->
            <select id="recipe-mode" class="text-xs bg-gray-700 text-gray-200 px-2 py-1 rounded">
              <option value="step" ${mode === 'step' ? 'selected' : ''}>Step Mode</option>
              <option value="auto" ${mode === 'auto' ? 'selected' : ''}>Auto Mode</option>
            </select>
            
            <!-- Auto delay (only shown in auto mode) -->
            ${mode === 'auto' ? `
              <div class="flex items-center gap-1 text-xs">
                <label for="auto-delay">Delay:</label>
                <input type="number" id="auto-delay" value="${autoDelay}" min="100" max="10000" step="100" 
                       class="w-16 bg-gray-700 text-gray-200 px-1 py-0.5 rounded">
                <span class="text-gray-400">ms</span>
              </div>
            ` : ''}
          </div>
        </div>
        
        <!-- Steps list -->
        <div class="recipe-steps flex-1 overflow-y-auto p-3">
          ${steps.map((step, index) => this.renderStep(step, index)).join('')}
        </div>
      </div>
    `;

    // Attach event handlers
    this.attachEventHandlers();
  }

  private renderControlButtons(): string {
    if (!this.state) return '';
    
    const { isRunning, isPaused, mode, currentStep, steps } = this.state;

    if (!isRunning) {
      // Not running - show start button
      return `
        <button id="btn-start" class="btn-control" title="Start execution">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M5 3l8 5-8 5V3z"/>
          </svg>
        </button>
        <button id="btn-restart" class="btn-control" title="Restart" ${steps.some(s => s.status !== 'pending') ? '' : 'disabled'}>
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M13.451 5.609l-.579-.939-1.068.812-.076.094c-.335.415-.927 1.341-1.124 2.876l-.021.165.033.163c.071.363.224.694.456.97l.087.102c.25.282.554.514.897.683l.123.061c.404.182.852.279 1.312.279.51 0 1.003-.12 1.444-.349l.105-.059c.435-.255.785-.618 1.014-1.051l.063-.119c.185-.38.283-.8.283-1.228 0-.347-.063-.684-.183-1.003l-.056-.147-.098-.245zm-3.177 3.342c-.169 0-.331-.037-.48-.109l-.044-.023c-.122-.061-.227-.145-.313-.249l-.032-.04c-.084-.106-.144-.227-.176-.361l-.012-.056c-.03-.137-.037-.283-.01-.428l.008-.059c.088-.987.373-1.76.603-2.122.183.338.276.735.276 1.142 0 .168-.02.332-.06.491l-.023.079c-.082.268-.225.51-.417.703l-.037.035c-.189.186-.423.325-.689.413l-.064.021c-.14.042-.288.063-.44.063zm1.373-4.326l2.255-1.718 1.017 1.647-2.351 1.79-.921-1.719zm-10.296.577l1.017-1.647 2.255 1.718-.921 1.719-2.351-1.79z"/>
          </svg>
        </button>
      `;
    }

    // Running - show pause/resume, step, and stop buttons
    if (isPaused || mode === 'step') {
      return `
        ${isPaused ? `
          <button id="btn-resume" class="btn-control" title="Resume">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M5 3l8 5-8 5V3z"/>
            </svg>
          </button>
        ` : mode === 'step' ? `
          <button id="btn-step" class="btn-control" title="Execute step" ${currentStep >= steps.length ? 'disabled' : ''}>
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M5 3l3 2.5L5 8V3zm4 0l3 2.5L9 8V3zm4 0h2v5h-2V3z"/>
            </svg>
          </button>
        ` : `
          <button id="btn-pause" class="btn-control" title="Pause">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M5 3h2v10H5V3zm4 0h2v10H9V3z"/>
            </svg>
          </button>
        `}
        <button id="btn-stop" class="btn-control btn-danger" title="Stop execution">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M4 4h8v8H4V4z"/>
          </svg>
        </button>
      `;
    }

    // Auto mode running - show pause and stop
    return `
      <button id="btn-pause" class="btn-control" title="Pause">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
          <path d="M5 3h2v10H5V3zm4 0h2v10H9V3z"/>
        </svg>
      </button>
      <button id="btn-stop" class="btn-control btn-danger" title="Stop execution">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
          <path d="M4 4h8v8H4V4z"/>
        </svg>
      </button>
    `;
  }

  private renderStep(step: RecipeStep, index: number): string {
    const { command, status, output, error } = step;
    const isCurrent = this.state?.currentStep === index;
    
    let statusIcon = '';
    let statusClass = '';
    
    switch (status) {
      case 'pending':
        statusIcon = '○';
        statusClass = 'text-gray-500';
        break;
      case 'running':
        statusIcon = '►';
        statusClass = 'text-blue-400 animate-pulse';
        break;
      case 'completed':
        statusIcon = '✓';
        statusClass = 'text-green-400';
        break;
      case 'failed':
        statusIcon = '✗';
        statusClass = 'text-red-400';
        break;
      case 'skipped':
        statusIcon = '⊘';
        statusClass = 'text-gray-600';
        break;
    }

    const typeClass = {
      'command': 'recipe-command',
      'echo': 'recipe-echo',
      'comment': 'recipe-comment',
      'pause': 'recipe-pause',
      'empty': 'recipe-empty'
    }[command.type];

    return `
      <div class="recipe-step ${typeClass} ${isCurrent ? 'current' : ''}" data-step="${index}">
        <div class="flex items-start gap-2">
          <span class="step-status ${statusClass} text-xs mt-0.5">${statusIcon}</span>
          <div class="flex-1">
            ${this.renderStepContent(command)}
            ${error ? `<div class="text-xs text-red-400 mt-1">Error: ${error}</div>` : ''}
            ${output && command.type === 'command' ? `<div class="text-xs text-gray-500 mt-1">${output}</div>` : ''}
          </div>
          <span class="text-xs text-gray-600">${command.lineNumber}</span>
        </div>
      </div>
    `;
  }

  private renderStepContent(command: any): string {
    switch (command.type) {
      case 'command':
        const args = command.args || [];
        return `<code class="text-xs">${command.command} ${args.join(' ')}</code>`;
      
      case 'echo':
        return `<div class="text-sm text-gray-300">${command.description}</div>`;
      
      case 'comment':
        return `<div class="text-xs text-gray-500 italic">${command.description}</div>`;
      
      case 'pause':
        return `<div class="text-sm text-yellow-400">⏸ ${command.description}</div>`;
      
      case 'empty':
        return '';
      
      default:
        return `<div class="text-xs text-gray-600">${command.rawLine}</div>`;
    }
  }

  private attachEventHandlers() {
    // Control buttons
    const btnStart = this.container.querySelector('#btn-start');
    const btnStep = this.container.querySelector('#btn-step');
    const btnPause = this.container.querySelector('#btn-pause');
    const btnResume = this.container.querySelector('#btn-resume');
    const btnStop = this.container.querySelector('#btn-stop');
    const btnRestart = this.container.querySelector('#btn-restart');
    
    // Mode selector
    const modeSelect = this.container.querySelector('#recipe-mode') as HTMLSelectElement;
    const delayInput = this.container.querySelector('#auto-delay') as HTMLInputElement;

    // Button handlers
    btnStart?.addEventListener('click', () => {
      const mode = modeSelect?.value as 'step' | 'auto' || 'step';
      this.runner.start(mode);
    });

    btnStep?.addEventListener('click', () => {
      this.runner.step();
    });

    btnPause?.addEventListener('click', () => {
      this.runner.pause();
    });

    btnResume?.addEventListener('click', () => {
      this.runner.resume();
    });

    btnStop?.addEventListener('click', () => {
      this.runner.stop();
    });

    btnRestart?.addEventListener('click', () => {
      this.runner.restart();
    });

    // Mode change handler
    modeSelect?.addEventListener('change', () => {
      if (this.state && !this.state.isRunning) {
        this.state.mode = modeSelect.value as 'step' | 'auto';
        this.render();
      }
    });

    // Delay change handler
    delayInput?.addEventListener('change', () => {
      const delay = parseInt(delayInput.value);
      if (!isNaN(delay)) {
        this.runner.setAutoDelay(delay);
      }
    });
  }

  async loadRecipe(filePath: string, content: string) {
    await this.runner.loadRecipe(filePath, content);
  }

  dispose() {
    // Clean up if needed
  }
}