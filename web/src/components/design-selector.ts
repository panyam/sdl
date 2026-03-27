/**
 * DesignSelector — a dropdown for switching between system designs
 * within a workspace. Rendered in the workspace toolbar above the
 * dockview panels.
 *
 * Populated after the presenter initializes and returns the list of
 * loaded system names from the Canvas proto.
 */

export class DesignSelector {
    private container: HTMLElement;
    private select: HTMLSelectElement;
    private label: HTMLElement;
    private onSelect: (systemName: string) => void;

    constructor(parent: HTMLElement, onSelect: (systemName: string) => void) {
        this.container = document.createElement('div');
        this.container.className = 'flex items-center gap-2';
        this.onSelect = onSelect;

        this.label = document.createElement('span');
        this.label.className = 'text-sm font-medium text-gray-600 dark:text-gray-300';
        this.label.textContent = 'Design:';

        this.select = document.createElement('select');
        this.select.className = 'text-sm bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1.5 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500 focus:border-transparent';
        this.select.disabled = true;

        // Placeholder while loading
        const placeholder = document.createElement('option');
        placeholder.textContent = 'Loading designs...';
        placeholder.disabled = true;
        placeholder.selected = true;
        this.select.appendChild(placeholder);

        this.select.addEventListener('change', () => {
            const value = this.select.value;
            if (value) {
                this.onSelect(value);
            }
        });

        this.container.appendChild(this.label);
        this.container.appendChild(this.select);
        parent.appendChild(this.container);
    }

    /** Update the dropdown with available designs. Selects activeDesign if provided. */
    setDesigns(designs: string[], activeDesign?: string): void {
        // Clear existing options using DOM methods (not innerHTML)
        while (this.select.firstChild) {
            this.select.removeChild(this.select.firstChild);
        }
        this.select.disabled = designs.length === 0;

        if (designs.length === 0) {
            const opt = document.createElement('option');
            opt.textContent = 'No designs loaded';
            opt.disabled = true;
            opt.selected = true;
            this.select.appendChild(opt);
            return;
        }

        for (const name of designs) {
            const opt = document.createElement('option');
            opt.value = name;
            opt.textContent = name;
            if (name === activeDesign) {
                opt.selected = true;
            }
            this.select.appendChild(opt);
        }
    }

    /** Get the currently selected design name. */
    getSelected(): string | null {
        return this.select.value || null;
    }
}
