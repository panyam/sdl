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
      this.render();
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
        <div class="flex-1"></div>
        <div class="text-xs text-gray-400 mr-2">
          <span id="toolbar-status">Ready</span>
        </div>
      </div>
    `;

    // Attach click handlers
    this.buttons.forEach(button => {
      const el = this.element.querySelector(`#toolbar-${button.id}`);
      if (el && !button.disabled) {
        el.addEventListener('click', button.onClick);
      }
    });
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
`;
document.head.appendChild(style);