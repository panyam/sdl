import { LCMComponent } from '@panyam/tsappkit';
import { DockviewApi, DockviewComponent } from 'dockview-core';
import { Graphviz } from "@hpcc-js/wasm";
import { CanvasViewerPageBase, PanelId } from './CanvasViewerPageBase';
import { SystemDiagram, Generator, Metric } from '../../../gen/wasmjs/sdl/v1/models/interfaces';
import { TabbedEditor } from '../../components/tabbed-editor';
import { ConsolePanel } from '../../components/console-panel';

/**
 * DockView-based implementation of CanvasViewerPage
 *
 * This provides a flexible, resizable panel layout using DockView
 * with the SDL editor, diagram, console, and other panels.
 */
export class CanvasViewerPageDockView extends CanvasViewerPageBase {
    private dockview: DockviewApi | null = null;
    private graphviz: any = null;

    // Panel components
    private tabbedEditor: TabbedEditor | null = null;
    private consolePanel: ConsolePanel | null = null;

    // Panel containers
    private diagramContainer: HTMLElement | null = null;
    private editorContainer: HTMLElement | null = null;
    private generatorsContainer: HTMLElement | null = null;
    private metricsContainer: HTMLElement | null = null;
    private flowRatesContainer: HTMLElement | null = null;

    // =========================================================================
    // Abstract Method Implementations
    // =========================================================================

    protected async initializeLayout(): Promise<void> {
        const appElement = document.getElementById('app');
        if (!appElement) {
            throw new Error('App container not found');
        }

        // Initialize Graphviz for diagram rendering
        try {
            this.graphviz = await Graphviz.load();
            console.log('[CanvasViewerPageDockView] Graphviz loaded');
        } catch (error) {
            console.error('[CanvasViewerPageDockView] Failed to load Graphviz:', error);
        }

        // Create DockView container
        const dockviewContainer = document.createElement('div');
        dockviewContainer.id = 'dockview-container';
        dockviewContainer.style.cssText = 'width: 100%; height: 100%; position: absolute; top: 0; left: 0;';
        appElement.appendChild(dockviewContainer);

        // Initialize DockView
        const dockviewComponent = new DockviewComponent(dockviewContainer, {
            createComponent: (options) => {
                const result = this.createDockviewComponent(options);
                return {
                    element: result.element,
                    init: () => {},
                    dispose: result.dispose,
                };
            },
        });

        this.dockview = dockviewComponent.api;

        // Add panels in initial layout
        this.setupInitialLayout();
    }

    protected createPanels(): LCMComponent[] {
        const panels: LCMComponent[] = [];

        // Panels are created lazily by DockView when added
        // We'll initialize them in createDockviewComponent

        return panels;
    }

    protected getDiagramContainer(): HTMLElement {
        if (!this.diagramContainer) {
            this.diagramContainer = document.createElement('div');
            this.diagramContainer.className = 'diagram-container h-full w-full overflow-auto bg-gray-100 dark:bg-gray-800';
        }
        return this.diagramContainer;
    }

    protected getEditorContainer(): HTMLElement {
        if (!this.editorContainer) {
            this.editorContainer = document.createElement('div');
            this.editorContainer.className = 'editor-container h-full w-full';
        }
        return this.editorContainer;
    }

    protected showPanel(panelId: PanelId): void {
        if (!this.dockview) return;

        const panel = this.dockview.getPanel(panelId);
        if (panel) {
            panel.api.setActive();
        }
    }

    protected renderDiagram(diagram: SystemDiagram): void {
        if (!this.graphviz || !this.diagramContainer) return;

        try {
            // Convert diagram to DOT format
            const dot = this.diagramToDot(diagram);

            // Render with Graphviz
            const svg = this.graphviz.dot(dot, 'svg');
            this.diagramContainer.innerHTML = svg;

            // Add click handlers to nodes
            this.attachDiagramClickHandlers();
        } catch (error) {
            console.error('[CanvasViewerPageDockView] Failed to render diagram:', error);
            this.diagramContainer.innerHTML = `<div class="p-4 text-red-500">Failed to render diagram: ${error}</div>`;
        }
    }

    protected renderGenerators(generators: Generator[]): void {
        if (!this.generatorsContainer) return;

        const html = generators.length === 0
            ? '<div class="p-4 text-gray-500">No generators configured</div>'
            : generators.map(g => `
                <div class="generator-item p-2 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
                    <div>
                        <span class="font-medium">${g.id || 'Unnamed'}</span>
                        <span class="text-sm text-gray-500 ml-2">${g.rate || 0} req/s</span>
                    </div>
                    <span class="px-2 py-1 text-xs rounded ${g.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}">
                        ${g.enabled ? 'Running' : 'Stopped'}
                    </span>
                </div>
            `).join('');

        this.generatorsContainer.innerHTML = html;
    }

    protected renderMetrics(metrics: Metric[]): void {
        if (!this.metricsContainer) return;

        const html = metrics.length === 0
            ? '<div class="p-4 text-gray-500">No metrics configured</div>'
            : metrics.map(m => `
                <div class="metric-item p-2 border-b border-gray-200 dark:border-gray-700">
                    <span class="font-medium">${m.id || 'Unnamed'}</span>
                    <span class="text-sm text-gray-500 ml-2">${m.metricType || ''}</span>
                </div>
            `).join('');

        this.metricsContainer.innerHTML = html;
    }

    protected renderFlowRates(rates: Record<string, number>): void {
        if (!this.flowRatesContainer) return;

        const entries = Object.entries(rates);
        const html = entries.length === 0
            ? '<div class="p-4 text-gray-500">No flow rates calculated</div>'
            : entries.map(([component, rate]) => `
                <div class="flow-rate-item p-2 border-b border-gray-200 dark:border-gray-700 flex justify-between">
                    <span>${component}</span>
                    <span class="font-mono">${rate.toFixed(2)} req/s</span>
                </div>
            `).join('');

        this.flowRatesContainer.innerHTML = html;
    }

    protected highlightDiagramComponents(componentIds: string[]): void {
        if (!this.diagramContainer) return;

        // Clear existing highlights
        this.clearDiagramHighlights();

        // Highlight specified components
        componentIds.forEach(id => {
            const node = this.diagramContainer!.querySelector(`#${CSS.escape(id)}`);
            if (node) {
                node.classList.add('highlighted');
                // Add visual highlight style
                const ellipse = node.querySelector('ellipse, polygon');
                if (ellipse) {
                    (ellipse as SVGElement).style.stroke = '#3b82f6';
                    (ellipse as SVGElement).style.strokeWidth = '3';
                }
            }
        });
    }

    protected clearDiagramHighlights(): void {
        if (!this.diagramContainer) return;

        const highlighted = this.diagramContainer.querySelectorAll('.highlighted');
        highlighted.forEach(node => {
            node.classList.remove('highlighted');
            const ellipse = node.querySelector('ellipse, polygon');
            if (ellipse) {
                (ellipse as SVGElement).style.stroke = '';
                (ellipse as SVGElement).style.strokeWidth = '';
            }
        });
    }

    protected logToConsole(level: string, message: string): void {
        if (!this.consolePanel) return;

        switch (level.toLowerCase()) {
            case 'error':
                this.consolePanel.error(message);
                break;
            case 'warn':
            case 'warning':
                this.consolePanel.warning(message);
                break;
            case 'success':
                this.consolePanel.success(message);
                break;
            default:
                this.consolePanel.log(message);
        }
    }

    protected clearConsolePanel(): void {
        if (this.consolePanel) {
            this.consolePanel.clear();
        }
    }

    // =========================================================================
    // DockView Helpers
    // =========================================================================

    private setupInitialLayout(): void {
        if (!this.dockview) return;

        // Add editor panel (left/center)
        this.dockview.addPanel({
            id: 'editor',
            component: 'editor',
            title: 'Editor',
            position: { direction: 'left' },
        });

        // Add diagram panel (right of editor)
        this.dockview.addPanel({
            id: 'diagram',
            component: 'diagram',
            title: 'System Diagram',
            position: { direction: 'right', referencePanel: 'editor' },
        });

        // Add console panel (bottom)
        this.dockview.addPanel({
            id: 'console',
            component: 'console',
            title: 'Console',
            position: { direction: 'below', referencePanel: 'editor' },
        });

        // Add generators panel (tabbed with console)
        this.dockview.addPanel({
            id: 'generators',
            component: 'generators',
            title: 'Generators',
            position: { direction: 'within', referencePanel: 'console' },
        });

        // Add metrics panel (tabbed with console)
        this.dockview.addPanel({
            id: 'metrics',
            component: 'metrics',
            title: 'Metrics',
            position: { direction: 'within', referencePanel: 'console' },
        });

        // Add flow rates panel (tabbed with console)
        this.dockview.addPanel({
            id: 'flow-rates',
            component: 'flow-rates',
            title: 'Flow Rates',
            position: { direction: 'within', referencePanel: 'console' },
        });
    }

    private createDockviewComponent(options: any): { element: HTMLElement; dispose?: () => void } {
        const container = document.createElement('div');
        container.className = 'h-full w-full overflow-auto';

        switch (options.name) {
            case 'editor':
                this.editorContainer = container;
                this.tabbedEditor = new TabbedEditor(container, this.dockview!);
                break;

            case 'diagram':
                this.diagramContainer = container;
                container.innerHTML = '<div class="p-4 text-gray-500">Load an SDL file to see the system diagram</div>';
                break;

            case 'console':
                this.consolePanel = new ConsolePanel(container);
                break;

            case 'generators':
                this.generatorsContainer = container;
                container.innerHTML = '<div class="p-4 text-gray-500">No generators configured</div>';
                break;

            case 'metrics':
                this.metricsContainer = container;
                container.innerHTML = '<div class="p-4 text-gray-500">No metrics configured</div>';
                break;

            case 'flow-rates':
                this.flowRatesContainer = container;
                container.innerHTML = '<div class="p-4 text-gray-500">No flow rates calculated</div>';
                break;

            default:
                container.innerHTML = `<div class="p-4">Unknown panel: ${options.name}</div>`;
        }

        return { element: container };
    }

    // =========================================================================
    // Diagram Helpers
    // =========================================================================

    private diagramToDot(diagram: SystemDiagram): string {
        const lines: string[] = ['digraph G {'];
        lines.push('  rankdir=TB;');
        lines.push('  node [shape=box, style=filled, fillcolor="#e5e7eb"];');
        lines.push('  edge [color="#6b7280"];');

        // Add nodes
        if (diagram.nodes) {
            for (const node of diagram.nodes) {
                const label = node.name || node.id || 'Unknown';
                const color = node.type === 'service' ? '#dbeafe' : '#e5e7eb';
                lines.push(`  "${node.id}" [label="${label}", fillcolor="${color}"];`);
            }
        }

        // Add edges
        if (diagram.edges) {
            for (const edge of diagram.edges) {
                const label = edge.label ? ` [label="${edge.label}"]` : '';
                lines.push(`  "${edge.fromId}" -> "${edge.toId}"${label};`);
            }
        }

        lines.push('}');
        return lines.join('\n');
    }

    private attachDiagramClickHandlers(): void {
        if (!this.diagramContainer) return;

        const nodes = this.diagramContainer.querySelectorAll('.node');
        nodes.forEach(node => {
            node.addEventListener('click', () => {
                const nodeId = node.id;
                if (nodeId) {
                    this.onDiagramComponentClicked(nodeId);
                }
            });

            // Add hover effect
            node.addEventListener('mouseenter', () => {
                node.classList.add('hover');
            });
            node.addEventListener('mouseleave', () => {
                node.classList.remove('hover');
            });
        });
    }

    // =========================================================================
    // Override for file operations
    // =========================================================================

    protected async handleSave(): Promise<void> {
        if (!this.tabbedEditor || this.readonly) return;

        const activeTabKey = this.tabbedEditor.getActiveTab();
        const content = this.tabbedEditor.getActiveContent();

        if (!activeTabKey || !content) return;

        // Extract path from tabKey (format is "fsId:path")
        const colonIndex = activeTabKey.indexOf(':');
        const filePath = colonIndex > 0 ? activeTabKey.substring(colonIndex + 1) : activeTabKey;

        try {
            // Notify presenter about the save
            await this.onFileSaved(filePath, content);
            this.tabbedEditor.saveTab(activeTabKey);
            this.logToConsole('success', `Saved: ${filePath}`);
        } catch (error) {
            this.logToConsole('error', `Failed to save: ${error}`);
        }
    }
}
