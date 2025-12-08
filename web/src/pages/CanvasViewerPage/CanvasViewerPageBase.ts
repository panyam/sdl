import { BasePage, LCMComponent } from '@panyam/tsappkit';
import SdlBundle from '../../../gen/wasmjs';
import { CanvasServiceClient } from '../../../gen/wasmjs/sdl/v1/services/canvasServiceClient';
import { CanvasDashboardPageMethods } from '../../../gen/wasmjs/sdl/v1/services/canvasDashboardPageClient';
import { CanvasViewPresenterClient } from '../../../gen/wasmjs/sdl/v1/services/canvasViewPresenterClient';
import { SingletonInitializerServiceClient } from '../../../gen/wasmjs/sdl/v1/services/singletonInitializerServiceClient';
import {
    UpdateMetricRequest, UpdateMetricResponse,
    ClearMetricsRequest, ClearMetricsResponse,
    SetMetricsListRequest, SetMetricsListResponse,
    UpdateDiagramRequest, UpdateDiagramResponse,
    HighlightComponentsRequest, HighlightComponentsResponse,
    ClearHighlightsRequest, ClearHighlightsResponse,
    UpdateGeneratorStateRequest, UpdateGeneratorStateResponse,
    SetGeneratorListRequest, SetGeneratorListResponse,
    LogMessageRequest, LogMessageResponse,
    ClearConsoleRequest, ClearConsoleResponse,
    UpdateFlowRatesRequest, UpdateFlowRatesResponse,
    ShowFlowPathRequest, ShowFlowPathResponse,
    ClearFlowPathsRequest, ClearFlowPathsResponse,
    UpdateUtilizationRequest, UpdateUtilizationResponse,
    SystemDiagram,
    Generator,
    Metric,
} from '../../../gen/wasmjs/sdl/v1/models/interfaces';

/**
 * Panel type identifiers for the canvas dashboard
 */
export type PanelId = 'diagram' | 'editor' | 'console' | 'generators' | 'metrics' | 'flow-rates';

/**
 * Abstract base class for Canvas Viewer Page implementations.
 *
 * This class contains all the core canvas/SDL logic (WASM, presenter, panels, events)
 * but delegates layout-specific concerns to child classes.
 *
 * Implements CanvasDashboardPageMethods to receive callbacks from the WASM presenter.
 */
export abstract class CanvasViewerPageBase extends BasePage implements LCMComponent, CanvasDashboardPageMethods {
    // =========================================================================
    // Protected Fields - Available to child classes
    // =========================================================================
    protected wasmBundle: SdlBundle;
    protected canvasClient: CanvasServiceClient;
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
        // Load canvas config from page data
        const pageData = (window as any).sdlPageData || {};
        this.currentCanvasId = pageData.canvasId || 'default';
        this.readonly = pageData.readonly || false;

        console.log(`[CanvasViewerPage] Initializing canvas: ${this.currentCanvasId} (readonly: ${this.readonly})`);

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
        // Bind canvas-specific events
        this.bindCanvasEvents();

        // Register this page as browser service for WASM callbacks
        this.wasmBundle.registerBrowserService('CanvasDashboardPage', this);

        // Initialize the presenter
        await this.initializePresenter();

        console.log('[CanvasViewerPage] Activated and ready');
    }

    // =========================================================================
    // Protected Helper Methods
    // =========================================================================

    /**
     * Load WASM bundle and initialize clients
     */
    protected async loadWASM(): Promise<void> {
        this.wasmBundle = new SdlBundle();
        this.canvasClient = new CanvasServiceClient(this.wasmBundle);
        this.canvasViewPresenterClient = new CanvasViewPresenterClient(this.wasmBundle);
        this.singletonInitializerClient = new SingletonInitializerServiceClient(this.wasmBundle);

        // Get WASM path from page or use default
        const wasmPath = (document.getElementById("wasmBundlePathField") as HTMLInputElement)?.value || '/sdl.wasm';
        await this.wasmBundle.loadWasm(wasmPath);
        await this.wasmBundle.waitUntilReady();

        console.log('[CanvasViewerPage] WASM bundle loaded');
    }

    /**
     * Initialize presenter with canvas data
     */
    protected async initializePresenter(): Promise<void> {
        // Call presenter to initialize
        const response = await this.canvasViewPresenterClient.initialize({
            canvasId: this.currentCanvasId || 'default',
        });

        if (!response.success) {
            throw new Error(`Presenter initialization failed: ${response.error}`);
        }

        console.log('[CanvasViewerPage] Presenter initialized:', response);
    }

    /**
     * Bind canvas-specific DOM events
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
                    strategy: 'iterative',
                });
            });
        }
    }

    /**
     * Handle file save
     */
    protected async handleSave(): Promise<void> {
        // To be implemented by child classes
        console.log('[CanvasViewerPage] Save requested');
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
    // CanvasDashboardPageMethods Interface - Browser RPC Methods
    // Called by WASM presenter to update the UI
    // =========================================================================

    async updateMetric(request: UpdateMetricRequest): Promise<UpdateMetricResponse> {
        console.log('[CanvasViewerPage] updateMetric:', request);
        // Update specific metric - child class handles rendering
        return {};
    }

    async clearMetrics(_: ClearMetricsRequest): Promise<ClearMetricsResponse> {
        console.log('[CanvasViewerPage] clearMetrics');
        this.currentMetrics = [];
        this.renderMetrics([]);
        return {};
    }

    async setMetricsList(request: SetMetricsListRequest): Promise<SetMetricsListResponse> {
        console.log('[CanvasViewerPage] setMetricsList:', request.metrics?.length || 0, 'metrics');
        this.currentMetrics = request.metrics || [];
        this.renderMetrics(this.currentMetrics);
        return {};
    }

    async updateDiagram(request: UpdateDiagramRequest): Promise<UpdateDiagramResponse> {
        console.log('[CanvasViewerPage] updateDiagram');
        if (request.diagram) {
            this.currentDiagram = request.diagram;
            this.renderDiagram(request.diagram);
        }
        return {};
    }

    async highlightComponents(request: HighlightComponentsRequest): Promise<HighlightComponentsResponse> {
        console.log('[CanvasViewerPage] highlightComponents:', request.componentIds);
        this.highlightDiagramComponents(request.componentIds || []);
        return {};
    }

    async clearHighlights(_: ClearHighlightsRequest): Promise<ClearHighlightsResponse> {
        console.log('[CanvasViewerPage] clearHighlights');
        this.clearDiagramHighlights();
        return {};
    }

    async updateGeneratorState(request: UpdateGeneratorStateRequest): Promise<UpdateGeneratorStateResponse> {
        console.log('[CanvasViewerPage] updateGeneratorState:', request.generatorId, request.status);
        // Update specific generator state
        const idx = this.currentGenerators.findIndex(g => g.id === request.generatorId);
        if (idx >= 0) {
            this.currentGenerators[idx] = {
                ...this.currentGenerators[idx],
                enabled: request.enabled,
                rate: request.rate,
            };
            this.renderGenerators(this.currentGenerators);
        }
        return {};
    }

    async setGeneratorList(request: SetGeneratorListRequest): Promise<SetGeneratorListResponse> {
        console.log('[CanvasViewerPage] setGeneratorList:', request.generators?.length || 0, 'generators');
        this.currentGenerators = request.generators || [];
        this.renderGenerators(this.currentGenerators);
        return {};
    }

    async logMessage(request: LogMessageRequest): Promise<LogMessageResponse> {
        console.log('[CanvasViewerPage] logMessage:', request.level, request.message);
        this.logToConsole(request.level || 'info', request.message || '');
        return {};
    }

    async clearConsole(_: ClearConsoleRequest): Promise<ClearConsoleResponse> {
        console.log('[CanvasViewerPage] clearConsole');
        this.clearConsolePanel();
        return {};
    }

    async updateFlowRates(request: UpdateFlowRatesRequest): Promise<UpdateFlowRatesResponse> {
        console.log('[CanvasViewerPage] updateFlowRates:', request.strategy);
        this.flowRates = request.rates || {};
        this.renderFlowRates(this.flowRates);
        return {};
    }

    async showFlowPath(_: ShowFlowPathRequest): Promise<ShowFlowPathResponse> {
        console.log('[CanvasViewerPage] showFlowPath');
        // Child class can implement flow path visualization
        return {};
    }

    async clearFlowPaths(_: ClearFlowPathsRequest): Promise<ClearFlowPathsResponse> {
        console.log('[CanvasViewerPage] clearFlowPaths');
        // Child class can implement
        return {};
    }

    async updateUtilization(_: UpdateUtilizationRequest): Promise<UpdateUtilizationResponse> {
        console.log('[CanvasViewerPage] updateUtilization');
        // Child class can implement utilization display
        return {};
    }
}
