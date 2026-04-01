import { BasePage, LCMComponent } from '@panyam/tsappkit';
import SdlBundle from '../../../gen/wasmjs';
import { DevEnvPageMethods } from '../../../gen/wasmjs/sdl/v1/services/devEnvPageClient';
import { CanvasViewPresenterClient } from '../../../gen/wasmjs/sdl/v1/services/canvasViewPresenterClient';
import { SingletonInitializerServiceClient } from '../../../gen/wasmjs/sdl/v1/services/singletonInitializerServiceClient';
import {
    UpdateDiagramRequest, UpdateDiagramResponse,
    LogMessageRequest, LogMessageResponse,
    UpdateFlowRatesRequest, UpdateFlowRatesResponse,
    DevEnvSystemChangedRequest, DevEnvSystemChangedResponse,
    DevEnvAvailableSystemsRequest, DevEnvAvailableSystemsResponse,
    DevEnvUpdateGeneratorRequest, DevEnvUpdateGeneratorResponse,
    DevEnvRemoveGeneratorRequest, DevEnvRemoveGeneratorResponse,
    DevEnvUpdateMetricRequest, DevEnvUpdateMetricResponse,
    DevEnvRemoveMetricRequest, DevEnvRemoveMetricResponse,
    SystemDiagram,
    Generator,
    Metric,
} from '../../../gen/wasmjs/sdl/v1/models/interfaces';

/**
 * Panel type identifiers for the workspace dashboard
 */
export type PanelId = 'diagram' | 'editor' | 'console' | 'generators' | 'metrics' | 'flow-rates';

/**
 * Abstract base class for Workspace Viewer Page implementations.
 *
 * This class contains all the core workspace/SDL logic (WASM, presenter, panels, events)
 * but delegates layout-specific concerns to child classes.
 *
 * Implements DevEnvPageMethods to receive push updates from the DevEnv via
 * the BrowserDevEnvPage forwarder in cmd/wasm/browser.go.
 */
export abstract class WorkspaceViewerPageBase extends BasePage implements LCMComponent, DevEnvPageMethods {
    // =========================================================================
    // Protected Fields - Available to child classes
    // =========================================================================
    protected wasmBundle: SdlBundle;
    protected canvasViewPresenterClient: CanvasViewPresenterClient;
    protected singletonInitializerClient: SingletonInitializerServiceClient;
    protected readonly: boolean = false;
    protected currentCanvasId: string | null;

    // Current state
    protected currentDiagram: SystemDiagram | null = null;
    protected currentGenerators: Generator[] = [];
    protected currentMetrics: Metric[] = [];
    protected flowRates: Record<string, number> = {};

    // =========================================================================
    // Abstract Methods - Must be implemented by child classes
    // =========================================================================

    /**
     * Initialize the layout system (DockView, CSS Grid, etc.)
     * Called during performLocalInit before components are created
     */
    protected abstract initializeLayout(): Promise<void>;

    /**
     * Create all panel instances and attach them to the DOM.
     * @returns Array of LCMComponent panels for lifecycle management
     */
    protected abstract createPanels(): LCMComponent[];

    /**
     * Get the container element where the diagram should render.
     */
    protected abstract getDiagramContainer(): HTMLElement;

    /**
     * Get the container element where the code editor should render.
     */
    protected abstract getEditorContainer(): HTMLElement;

    /**
     * Show/focus a specific panel
     */
    protected abstract showPanel(panelId: PanelId): void;

    /**
     * Update the diagram display with new data
     */
    protected abstract renderDiagram(diagram: SystemDiagram): void;

    /**
     * Update the generators panel with new data
     */
    protected abstract renderGenerators(generators: Generator[]): void;

    /**
     * Update the metrics panel with new data
     */
    protected abstract renderMetrics(metrics: Metric[]): void;

    /**
     * Update the flow rates display
     */
    protected abstract renderFlowRates(rates: Record<string, number>): void;

    /**
     * Highlight components in the diagram
     */
    protected abstract highlightDiagramComponents(componentIds: string[]): void;

    /**
     * Clear highlights from the diagram
     */
    protected abstract clearDiagramHighlights(): void;

    /**
     * Log a message to the console panel
     */
    protected abstract logToConsole(level: string, message: string): void;

    /**
     * Clear the console panel
     */
    protected abstract clearConsolePanel(): void;

    // =========================================================================
    // LCMComponent Interface Implementation
    // =========================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    async performLocalInit(): Promise<LCMComponent[]> {
        const pageData = (window as any).sdlPageData || {};
        this.currentCanvasId = pageData.canvasId || 'default';
        this.readonly = pageData.readonly || false;

        console.log(`[WorkspaceViewerPage] Initializing: ${this.currentCanvasId} (readonly: ${this.readonly})`);

        // Initialize layout system (DockView/Grid)
        await this.initializeLayout();

        // Create panels (implementation-specific)
        const panels = this.createPanels();

        // Kick off WASM loading
        await this.loadWASM();

        // Return child components for lifecycle management
        return panels.filter(c => c != null);
    }

    /**
     * Phase 2: Inject dependencies
     */
    setupDependencies(): void {
        // Child classes can override to inject dependencies into panels
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        this.bindCanvasEvents();

        // Register this page as browser service for WASM callbacks
        this.wasmBundle.registerBrowserService('DevEnvPage', this);

        // Initialize the presenter
        await this.initializePresenter();

        console.log('[WorkspaceViewerPage] Activated and ready');
    }

    // =========================================================================
    // Protected Helper Methods
    // =========================================================================

    /**
     * Load WASM bundle and initialize clients
     */
    protected async loadWASM(): Promise<void> {
        this.wasmBundle = new SdlBundle();
        this.canvasViewPresenterClient = new CanvasViewPresenterClient(this.wasmBundle);
        this.singletonInitializerClient = new SingletonInitializerServiceClient(this.wasmBundle);

        const wasmPath = (document.getElementById("wasmBundlePathField") as HTMLInputElement)?.value || '/wasm/sdl.wasm';
        await this.wasmBundle.loadWasm(wasmPath);
        await this.wasmBundle.waitUntilReady();

        console.log('[WorkspaceViewerPage] WASM bundle loaded');
    }

    /**
     * Initialize presenter with canvas data
     */
    protected async initializePresenter(): Promise<void> {
        const response = await this.canvasViewPresenterClient.initialize({
            canvasId: this.currentCanvasId || 'default',
        });

        if (!response.success) {
            throw new Error(`Presenter initialization failed: ${response.error}`);
        }

        console.log('[WorkspaceViewerPage] Presenter initialized:', response);
    }

    /**
     * Bind DOM events
     */
    protected bindCanvasEvents(): void {
        // File save shortcut (Ctrl+S / Cmd+S)
        document.addEventListener('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                this.handleSave();
            }
        });

        // Evaluate flows button
        const evalFlowsBtn = document.getElementById('evaluate-flows-btn');
        if (evalFlowsBtn) {
            evalFlowsBtn.addEventListener('click', () => {
                this.canvasViewPresenterClient.evaluateFlows({
                    canvasId: this.currentCanvasId || '',
                    strategy: 'runtime',
                });
            });
        }
    }

    /**
     * Handle file save
     */
    protected async handleSave(): Promise<void> {
        console.log('[WorkspaceViewerPage] Save requested');
    }

    /**
     * Notify presenter that a file was selected
     */
    protected async onFileSelected(filePath: string): Promise<void> {
        await this.canvasViewPresenterClient.fileSelected({
            canvasId: this.currentCanvasId || '',
            filePath,
        });
    }

    /**
     * Notify presenter that a file was saved
     */
    protected async onFileSaved(filePath: string, content: string): Promise<void> {
        await this.canvasViewPresenterClient.fileSaved({
            canvasId: this.currentCanvasId || '',
            filePath,
            content,
        });
    }

    /**
     * Notify presenter that a diagram component was clicked
     */
    protected async onDiagramComponentClicked(componentName: string, methodName?: string): Promise<void> {
        await this.canvasViewPresenterClient.diagramComponentClicked({
            canvasId: this.currentCanvasId || '',
            componentName,
            methodName: methodName || '',
        });
    }

    // =========================================================================
    // DevEnvPageMethods Interface - Browser RPC Methods
    // Called by DevEnv (via BrowserDevEnvPage) to push state updates
    // =========================================================================

    onSystemChanged(request: DevEnvSystemChangedRequest): DevEnvSystemChangedResponse {
        console.log('[WorkspaceViewerPage] onSystemChanged:', request.systemName);
        return {};
    }

    onAvailableSystemsChanged(request: DevEnvAvailableSystemsRequest): DevEnvAvailableSystemsResponse {
        console.log('[WorkspaceViewerPage] onAvailableSystemsChanged:', request.systemNames);
        return {};
    }

    updateDiagram(request: UpdateDiagramRequest): UpdateDiagramResponse {
        console.log('[WorkspaceViewerPage] updateDiagram');
        if (request.diagram) {
            this.currentDiagram = request.diagram;
            this.renderDiagram(request.diagram);
        }
        return {};
    }

    updateGenerator(request: DevEnvUpdateGeneratorRequest): DevEnvUpdateGeneratorResponse {
        console.log('[WorkspaceViewerPage] updateGenerator:', request.name);
        if (request.generator) {
            const idx = this.currentGenerators.findIndex(g => g.name === request.name);
            if (idx >= 0) {
                this.currentGenerators[idx] = request.generator;
            } else {
                this.currentGenerators.push(request.generator);
            }
            this.renderGenerators(this.currentGenerators);
        }
        return {};
    }

    removeGenerator(request: DevEnvRemoveGeneratorRequest): DevEnvRemoveGeneratorResponse {
        console.log('[WorkspaceViewerPage] removeGenerator:', request.name);
        this.currentGenerators = this.currentGenerators.filter(g => g.name !== request.name);
        this.renderGenerators(this.currentGenerators);
        return {};
    }

    updateMetric(request: DevEnvUpdateMetricRequest): DevEnvUpdateMetricResponse {
        console.log('[WorkspaceViewerPage] updateMetric:', request.name);
        if (request.metric) {
            const idx = this.currentMetrics.findIndex(m => m.name === request.name);
            if (idx >= 0) {
                this.currentMetrics[idx] = request.metric;
            } else {
                this.currentMetrics.push(request.metric);
            }
            this.renderMetrics(this.currentMetrics);
        }
        return {};
    }

    removeMetric(request: DevEnvRemoveMetricRequest): DevEnvRemoveMetricResponse {
        console.log('[WorkspaceViewerPage] removeMetric:', request.name);
        this.currentMetrics = this.currentMetrics.filter(m => m.name !== request.name);
        this.renderMetrics(this.currentMetrics);
        return {};
    }

    updateFlowRates(request: UpdateFlowRatesRequest): UpdateFlowRatesResponse {
        console.log('[WorkspaceViewerPage] updateFlowRates:', request.strategy);
        this.flowRates = request.rates || {};
        this.renderFlowRates(this.flowRates);
        return {};
    }

    logMessage(request: LogMessageRequest): LogMessageResponse {
        console.log('[WorkspaceViewerPage] logMessage:', request.level, request.message);
        this.logToConsole(request.level || 'info', request.message || '');
        return {};
    }
}
