export interface ToolbarButton {
  id: string;
  label: string;
  icon?: string;
  tooltip?: string;
  disabled?: boolean;
  onClick: () => void;
}

export class Toolbar {
  private element: HTMLElement;
  private buttons: ToolbarButton[] = [];
  private recipeControlsContainer: HTMLElement | null = null;

  constructor(container: HTMLElement) {
    this.element = container;
    this.render();
  }

  setButtons(buttons: ToolbarButton[]) {
    this.buttons = buttons;
    this.render();
  }

  updateButton(id: string, updates: Partial<ToolbarButton>) {
    const button = this.buttons.find(b => b.id === id);
    if (button) {
      Object.assign(button, updates);
      
      // Update just the specific button without re-rendering everything
      const buttonEl = this.element.querySelector(`#toolbar-${id}`) as HTMLButtonElement;
      if (buttonEl) {
        // Update disabled state
        if (updates.disabled !== undefined) {
          buttonEl.disabled = updates.disabled;
          if (updates.disabled) {
            buttonEl.className = 'toolbar-button px-3 py-1 rounded text-sm bg-gray-700 text-gray-500 cursor-not-allowed';
          } else {
            buttonEl.className = 'toolbar-button px-3 py-1 rounded text-sm bg-gray-700 hover:bg-gray-600 text-gray-200';
          }
        }
        
        // Update tooltip
        if (updates.tooltip) {
          buttonEl.title = updates.tooltip;
        }
        
        // Update label/icon if needed
        if (updates.label || updates.icon) {
          buttonEl.innerHTML = `
            ${button.icon ? `<span class="mr-1">${button.icon}</span>` : ''}
            ${button.label}
          `;
        }
      }
    }
  }

  private render() {
    this.element.innerHTML = `
      <div class="toolbar flex items-center gap-2 p-2 bg-gray-800 border-b border-gray-700">
        <div class="flex items-center gap-1">
          ${this.buttons.map(button => `
            <button 
              id="toolbar-${button.id}"
              class="toolbar-button px-3 py-1 rounded text-sm ${
                button.disabled 
                  ? 'bg-gray-700 text-gray-500 cursor-not-allowed' 
                  : 'bg-gray-700 hover:bg-gray-600 text-gray-200'
              }"
              ${button.disabled ? 'disabled' : ''}
              title="${button.tooltip || button.label}">
              ${button.icon ? `<span class="mr-1">${button.icon}</span>` : ''}
              ${button.label}
            </button>
          `).join('')}
        </div>
        <div class="mx-4 h-6 w-px bg-gray-700"></div>
        <div id="recipe-controls-container" class="flex items-center gap-2"></div>
        <div class="flex-1"></div>
        <div class="text-xs text-gray-400 mr-2">
          <span id="toolbar-status">Ready</span>
        </div>
      </div>
    `;

    // Store reference to recipe controls container
    this.recipeControlsContainer = this.element.querySelector('#recipe-controls-container');

    // Attach click handlers
    this.buttons.forEach(button => {
      const el = this.element.querySelector(`#toolbar-${button.id}`);
      if (el && !button.disabled) {
        el.addEventListener('click', button.onClick);
      }
    });
  }

  getRecipeControlsContainer(): HTMLElement | null {
    return this.recipeControlsContainer;
  }

  setStatus(message: string, type: 'info' | 'success' | 'error' = 'info') {
    const statusEl = this.element.querySelector('#toolbar-status');
    if (statusEl) {
      const colorClass = {
        info: 'text-gray-400',
        success: 'text-green-400',
        error: 'text-red-400'
      }[type];
      
      statusEl.className = `text-xs ${colorClass}`;
      statusEl.textContent = message;
    }
  }
}

// CSS
const style = document.createElement('style');
style.textContent = `
  .toolbar {
    height: 40px;
    flex-shrink: 0;
  }

  .toolbar-button {
    transition: all 0.2s;
    font-size: 13px;
    font-weight: 500;
  }

  .toolbar-button:not(:disabled):active {
    transform: scale(0.95);
  }

  .toolbar-btn {
    padding: 4px 8px;
    background: #374151;
    border: 1px solid #4b5563;
    border-radius: 4px;
    color: #d1d5db;
    cursor: pointer;
    transition: all 0.2s;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }

  .toolbar-btn:hover:not(:disabled) {
    background: #4b5563;
    border-color: #6b7280;
  }

  .toolbar-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .recipe-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
`;
document.head.appendChild(style);