import { CanvasClient } from './canvas-client.js';
import { DashboardState, ParameterConfig, GenerateCall } from './types.js';
import type { SystemDiagram } from './gen/sdl/v1/canvas_pb.ts';
import { Chart, ChartConfiguration } from 'chart.js/auto';
import { Graphviz } from "@hpcc-js/wasm";
import { DockviewApi, DockviewComponent } from 'dockview-core';

export class Dashboard {
  private api: CanvasClient;
  private state: DashboardState;
  private charts: Record<string, Chart> = {};
  private systemDiagram: SystemDiagram | null = null;
  private chartUpdateInterval: number | null = null;
  private graphviz: any = null; // Will be initialized asynchronously
  private dockview: DockviewApi | null = null;
  private metricStreamController: AbortController | null = null;
  private generatorPollInterval: number | null = null;
  private canvasId: string;
  private isUpdatingGenerators: boolean = false; // Flag to prevent UI overwrites during updates
  private generatorUpdateTimeout: number | null = null; // Debounce timer for generator updates
  private layoutTopToBottom = false;

  // Parameter configurations - populated when a system is loaded
  private parameters: ParameterConfig[] = [];

  constructor(canvasId: string = 'default') {
    this.canvasId = canvasId;
    this.api = new CanvasClient(canvasId);
    this.state = {
      isConnected: false,
      simulationResults: {},
      metrics: {
        load: 0,
        latency: 0,
        successRate: 0
      },
      dynamicCharts: {},
      generateCalls: []
    };

    this.initializeGraphviz();
    this.initialize();
  }

  private async initializeGraphviz() {
    try {
      this.graphviz = await Graphviz.load();
      console.log('✅ Graphviz WASM loaded successfully');
    } catch (error) {
      console.error('❌ Failed to load Graphviz WASM:', error);
    }
  }

  private async initialize() {
    try {
      // Ensure the canvas exists (create if needed)
      await this.api.ensureCanvas();
      
      // Update document title
      document.title = `SDL Canvas - ${this.canvasId}`;
      
      // First render the layout structure
      this.render();
      
      // Then setup WebSocket after layout is ready
      setTimeout(() => {
        this.setupEventListeners();
        // Don't start chart updates here - they'll be started after metrics are loaded
      }, 100);
      
      // Load current Canvas state to see if there's an existing session
      await this.loadCanvasState();
    } catch (error) {
      console.error('❌ Failed to initialize dashboard:', error);
    }
  }

  private async loadCanvasState() {
    console.log('🔄 loadCanvasState() called - loading initial data');
    try {
      const stateResponse = await this.api.getState();
      if (stateResponse != null) {
        const canvasState = stateResponse;
        
        // Update dashboard state from Canvas state
        this.state.currentFile = canvasState.loadedFiles?.[0]; // Use first loaded file
        this.state.currentSystem = canvasState.activeSystem;
        
        // If there's an active system, load its diagram
        if (this.state.currentSystem) {
          this.loadSystemDiagram();
        }
        
        // Convert Canvas generators to dashboard generate calls
        this.state.generateCalls = (canvasState.generators || []).map(gen => ({
          id: gen.id,
          name: gen.name,
          target: `${gen.component}.${gen.method}`, // Combine component and method
          rate: gen.rate,
          enabled: gen.enabled
        }));
      }

      // Load generators from API
      console.log('🔄 Loading generators from API...');
      const generatorsResponse = await this.api.getGenerators();
      if (generatorsResponse.success && generatorsResponse.data) {
        console.log('✅ Generators loaded:', Object.keys(generatorsResponse.data).length);
        this.state.generateCalls = Object.values(generatorsResponse.data).map(gen => ({
          id: gen.id,
          name: gen.name,
          target: `${gen.component}.${gen.method}`, // Combine component and method
          rate: gen.rate,
          enabled: gen.enabled
        }));
      } else {
        console.log('⚠️ No generators found or failed to load');
      }

      // Load metrics and create dynamic charts
      console.log('🔄 Loading metrics from API...');
      const metricsResponse = await this.api.getMetrics();
      if (metricsResponse.success && metricsResponse.data) {
        console.log('✅ Metrics loaded:', Object.keys(metricsResponse.data).length);
        console.log('📊 Metrics data:', metricsResponse.data);
        // Convert metrics to dynamic charts - assume all metrics are enabled
        Object.values(metricsResponse.data).forEach((metric: any) => {
          console.log('🔍 Processing metric:', metric.id, 'full metric:', metric);
          // Handle methods as array from Metric proto
          const target = metric.methods && metric.methods.length > 0 
            ? `${metric.component}.${metric.methods[0]}` 
            : metric.component;
          
          // Build informative title with ID, metric type, and aggregation
          let title = `${metric.id}`;
          if (metric.metricType) {
            title += ` (${metric.metricType}`;
            if (metric.aggregation) {
              title += ` - ${metric.aggregation}`;
            }
            if (metric.aggregationWindow && metric.aggregationWindow > 0) {
              title += ` @ ${metric.aggregationWindow}s`;
            }
            title += ')';
          }
          
          this.state.dynamicCharts[metric.id] = {
            chartName: metric.id,
            metricName: metric.metricType || 'latency',
            target: target, // Store the actual target for API calls
            data: [],
            labels: [],
            title: title
          };
          console.log('📈 Added chart:', metric.id, 'to dynamicCharts');
        });
        console.log('📊 Final dynamicCharts state:', this.state.dynamicCharts);
      } else {
        console.log('⚠️ No metrics found or failed to load');
      }

      // Update UI panels AFTER loading all data
      console.log('🔄 Updating all panels with loaded data...');
      this.updateAllPanels();
      
      // Restart streaming with new metrics
      if (Object.keys(this.state.dynamicCharts).length > 0) {
        console.log('🚀 Restarting metric streaming after loading metrics');
        setTimeout(async () => {
          await this.startChartUpdates();
        }, 100); // Small delay to ensure charts are initialized
      }

    } catch (error) {
      console.error('❌ Failed to load Canvas state:', error);
    }
  }

  private setupEventListeners() {
    // We use Connect streaming for real-time metrics, not WebSockets
    // Disabled periodic generator polling - we now have manual refresh button
    if (false) {  // disabling generator polling for now but dont remove it
      this.startGeneratorPolling();
    }
  }
  
  private startGeneratorPolling() {
    // Clear any existing interval
    if (this.generatorPollInterval) {
      clearInterval(this.generatorPollInterval);
    }
    
    // Poll generators every 5 seconds
    this.generatorPollInterval = window.setInterval(() => {
      this.refreshGenerators();
    }, 5000);
    
    // Also do an immediate refresh
    this.refreshGenerators();
  }
  
  private stopGeneratorPolling() {
    if (this.generatorPollInterval) {
      clearInterval(this.generatorPollInterval);
      this.generatorPollInterval = null;
    }
  }

  // Removed WebSocket handler - we use Connect streaming instead
  /*
  private handleWebSocketMessage(message: WebSocketMessage) {
    console.log('📡 WebSocket message:', message);
    
    switch (message.type) {
      case 'connected':
        this.state.isConnected = true;
        break;
      case 'fileLoaded':
        this.state.currentFile = message.file;
        break;
      case 'systemActivated':
        this.state.currentSystem = message.system;
        this.state.isConnected = true;
        this.loadSystemDiagram();
        // Also refresh generators and measurements when system is activated
        this.refreshGenerators();
        // this.refreshMeasurements();
        this.updateAllPanels();
        break;
      case 'parameterChanged':
        // Update parameter value in UI
        const param = this.parameters.find(p => p.path === message.path);
        if (param) {
          param.value = message.value;
        }
        // Auto-trigger metrics update after parameter changes
        this.updateMetrics();
        break;
      case 'simulationCompleted':
        // Trigger metrics update
        this.updateMetrics();
        break;
      case 'generatorAdded':
      case 'generatorUpdated':
      case 'generatorRemoved':
      case 'generatorPaused':
      case 'generatorResumed':
      case 'generatorsStarted':
      case 'generatorsStopped':
        this.refreshGenerators();
        this.updateTrafficGenerationPanel();
        break;
      case 'measurementAdded':
      case 'measurementUpdated':
      case 'measurementRemoved':
        // this.refreshMeasurements();
        // this.updateMeasurementsPanel();
        this.updateLiveMetricsPanel();
        break;
      case 'plotGenerated':
        console.log('📊 Plot generated:', message.outputFile);
        break;
      case 'stateRestored':
        this.loadCanvasState();
        break;
    }

    // Don't update layout on WebSocket messages - layout is static once created
  }

  private async updateMetrics() {
    // Calculate metrics from the latest simulation
    // This is a simplified version - in a real implementation, 
    // we'd get this data from the API
    const param = this.parameters.find(p => p.path === 'server.pool.ArrivalRate');
    if (param) {
      this.state.metrics.load = param.value as number;
      this.state.metrics.latency = this.calculateEstimatedLatency(param.value as number);
      this.state.metrics.successRate = this.calculateEstimatedSuccessRate(param.value as number);
      this.state.metrics.serverUtilization = Math.min((param.value as number) / 10 * 100, 100);
    }

    const cacheParam = this.parameters.find(p => p.path === 'server.db.CacheHitRate');
    if (cacheParam) {
      this.state.metrics.cacheHitRate = cacheParam.value as number;
    }

    this.updateDynamicCharts();
  }

  private calculateEstimatedLatency(load: number): number {
    // Simple M/M/1 queue approximation for demonstration
    const serviceRate = 10; // Approximate service rate
    if (load >= serviceRate) return 999; // Overloaded
    return (1 / (serviceRate - load)) * 1000; // Convert to ms
  }

  private calculateEstimatedSuccessRate(load: number): number {
    const maxCapacity = 10;
    if (load <= maxCapacity) return 99.5;
    return Math.max(50, 99.5 - (load - maxCapacity) * 5);
  }
  */

  private async loadSystemDiagram() {
    try {
      // Fetch the system diagram from the server
      const response = await this.api.getSystemDiagram();
      if (response.success && response.data) {
        this.systemDiagram = response.data;
        console.log('📊 System diagram loaded:', this.systemDiagram);
        // Update only the system architecture panel
        this.updateSystemArchitecturePanel();
      } else {
        throw new Error('Failed to get system diagram');
      }
    } catch (error) {
      console.error('❌ Failed to load system diagram:', error);
      this.systemDiagram = null;
    }
  }

  private async setParameter(path: string, value: any) {
    try {
      const result = await this.api.setParameter(path, value);
      if (!result.success) {
        throw new Error(result.data?.errorMessage || 'Failed to set parameter');
      }
      console.log(`✅ Parameter ${path} set to ${value}`);
      
      // Parameter updated - metrics will be updated via WebSocket events
    } catch (error) {
      console.error(`❌ Failed to set parameter ${path}:`, error);
      this.showError(`Failed to set parameter: ${error}`);
    }
  }

  private async updateDynamicCharts() {
    // Update each chart with real measurement data
    for (const chartData of Object.values(this.state.dynamicCharts)) {
      const chart = this.charts[chartData.chartName];
      if (!chart) continue;

      try {
        // Fetch last 5 minutes of data for this measurement target
        const endTime = new Date().toISOString();
        const startTime = new Date(Date.now() - 5 * 60 * 1000).toISOString();
        
        // Use the target from the measurement configuration
        const target = chartData.target || chartData.metricName;
        
        // Use queryMetrics to get data for this metric
        const response = await this.api.queryMetrics(target, new Date(startTime), new Date(endTime), 100);
        
        if (response.success && response.data) {
          const dataPoints = response.data;
          
          // Clear existing chart data
          chart.data.labels = [];
          chart.data.datasets[0].data = [];
          
          // Add new data points (limit to last 20 points for performance)
          const recentPoints = dataPoints.slice(-20);
          
          recentPoints.forEach((point: any) => {
            const timestamp = new Date(point.timestamp).toLocaleTimeString();
            chart.data.labels?.push(timestamp);
            chart.data.datasets[0].data.push(point.value);
          });
          
          chart.update('none');
        } else {
          // If no data available, fall back to simulated data
          this.updateChartWithSimulatedData(chartData, chart);
        }
      } catch (error) {
        console.warn(`Failed to fetch data for ${chartData.chartName}:`, error);
        // Fall back to simulated data on error
        this.updateChartWithSimulatedData(chartData, chart);
      }
    }
  }

  private updateChartWithSimulatedData(chartData: any, chart: Chart) {
    const now = Date.now();
    const timestamp = new Date(now).toLocaleTimeString();

    // Simulate metric values based on chart type (fallback behavior)
    let value = 0;
    if (chartData.metricName.includes('p95Latency') || chartData.metricName.includes('p99Latency')) {
      value = this.state.metrics.latency + Math.random() * 10 - 5;
    } else if (chartData.metricName.includes('qps')) {
      value = this.state.metrics.load + Math.random() * 2 - 1;
    } else if (chartData.metricName.includes('errorRate')) {
      value = Math.max(0, (100 - this.state.metrics.successRate) + Math.random() * 2 - 1);
    } else if (chartData.metricName.includes('HitRate') || chartData.metricName.includes('utilization')) {
      value = Math.max(0, Math.min(100, this.state.metrics.cacheHitRate * 100 + Math.random() * 20 - 10));
    } else if (chartData.metricName.includes('memory')) {
      value = Math.max(0, 512 + Math.random() * 200 - 100); // Simulate heap usage in MB
    }

    // Update chart data
    const data = chart.data;
    if (data.datasets[0].data.length > 20) {
      data.labels?.shift();
      data.datasets[0].data.shift();
    }
    
    data.labels?.push(timestamp);
    data.datasets[0].data.push(value);
    
    chart.update('none');
  }

  private showError(message: string) {
    const errorDiv = document.getElementById('error-display');
    if (errorDiv) {
      errorDiv.textContent = message;
      errorDiv.classList.remove('hidden');
      setTimeout(() => errorDiv.classList.add('hidden'), 5000);
    }
  }


  private render() {
    const app = document.getElementById('app');
    if (!app) return;

    app.innerHTML = `
      <div class="h-screen flex flex-col">
        <!-- Header -->
        <div class="p-4 flex-shrink-0 bg-gray-900 border-b border-gray-700">
          <h1 class="text-xl font-bold text-blue-300 mb-2">SDL Canvas Dashboard - ${this.canvasId}</h1>
          <div class="flex items-center gap-4">
            <div class="flex items-center gap-2">
              <div class="w-2 h-2 rounded-full ${this.state.isConnected ? 'bg-green-400' : 'bg-red-400'}"></div>
              <span class="text-xs text-gray-400">${this.state.isConnected ? 'Connected' : 'Disconnected'}</span>
            </div>
            ${this.state.currentSystem ? `
              <div class="text-xs text-gray-300">
                <span class="text-gray-400">System:</span> ${this.state.currentSystem}
              </div>
            ` : ''}
            <button id="reset-layout-btn" class="text-xs px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-gray-300" title="Reset layout to default">
              ⚙️ Reset Layout
            </button>
          </div>
          <div id="error-display" class="hidden mt-2 p-2 bg-red-800 text-red-200 rounded text-sm"></div>
        </div>

        <!-- DockView Container -->
        <div id="dockview-container" class="flex-1"></div>
      </div>
    `;

    this.initializeLayout();
  }

  private initializeLayout() {
    // Destroy existing dockview if it exists
    if (this.dockview) {
      this.dockview.dispose();
      this.dockview = null;
    }

    const container = document.getElementById('dockview-container');
    if (!container) {
      console.error('❌ DockView container not found');
      return;
    }

    // Try to load saved layout configuration
    const savedLayout = this.loadLayoutConfig();
    
    // Apply dark theme to container
    container.className = 'dockview-theme-dark flex-1';
    
    // Create DockView component
    const dockviewComponent = new DockviewComponent(container, {
      createComponent: (options: any) => {
        const element = document.createElement('div');
        element.className = 'h-full p-4 overflow-auto';
        
        switch (options.name) {
          case 'systemArchitecture':
            element.innerHTML = this.renderSystemArchitectureOnly();
            break;
          case 'trafficGeneration':
            element.innerHTML = this.renderGenerateControls();
            break;
          // case 'measurements': element.innerHTML = this.renderMeasurements(); break;
          case 'liveMetrics':
            element.innerHTML = `
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" style="grid-auto-rows: 200px;">
                ${this.renderDynamicCharts()}
              </div>
            `;
            break;
        }
        
        return {
          element,
          init: () => {},
          dispose: () => {}
        };
      }
    });

    this.dockview = dockviewComponent.api;

    if (savedLayout) {
      // Restore saved layout
      try {
        this.dockview.fromJSON(savedLayout);
        console.log('📂 Layout restored from localStorage');
      } catch (error) {
        console.warn('Failed to restore saved layout, creating default:', error);
        localStorage.removeItem('sdl-dockview-layout');
        this.createDefaultLayout();
      }
    } else {
      // Create default layout
      this.createDefaultLayout();
    }

    // Listen for layout changes and save them
    this.dockview.onDidLayoutChange(() => {
      this.saveLayoutConfig();
    });

    console.log('🔧 DockView instance created');

    // Setup interactivity after layout is initialized
    setTimeout(() => {
      this.setupInteractivity();
      this.initDynamicCharts();
    }, 100);
  }

  private createDefaultLayout() {
    // Add panels to DockView in default configuration
    this.dockview!.addPanel({
      id: 'systemArchitecture',
      component: 'systemArchitecture',
      title: 'System Architecture'
    });

    this.dockview!.addPanel({
      id: 'trafficGeneration', 
      component: 'trafficGeneration',
      title: 'Traffic Generation',
      position: { direction: 'right' }
    });

    /*
    this.dockview!.addPanel({
      id: 'measurements',
      component: 'measurements', 
      title: 'Measurements',
      position: { direction: 'below' }
    });
   */

    this.dockview!.addPanel({
      id: 'liveMetrics',
      component: 'liveMetrics',
      title: 'Live Metrics', 
      position: { direction: 'below' }
    });
  }

  private updateSystemArchitecturePanel() {
    if (!this.dockview) return;
    
    // Update the system architecture panel content
    const panel = this.dockview.getPanel('systemArchitecture');
    if (panel) {
      const element = panel.view.content.element;
      element.innerHTML = this.renderSystemArchitectureOnly();
      
      // Re-setup interactivity for updated content
      setTimeout(() => {
        this.setupInteractivity();
      }, 10);
    }
  }

  private updateAllPanels() {
    if (!this.dockview) return;
    
    // Update all panel contents with current state
    this.updateTrafficGenerationPanel();
    // this.updateMeasurementsPanel();
    this.updateLiveMetricsPanel();
  }

  private updateTrafficGenerationPanel() {
    if (!this.dockview) return;
    
    const panel = this.dockview.getPanel('trafficGeneration');
    if (panel) {
      const element = panel.view.content.element;
      element.innerHTML = this.renderGenerateControls();
      
      // Re-setup interactivity
      setTimeout(() => {
        this.setupInteractivity();
      }, 10);
    }
  }

  private updateLiveMetricsPanel() {
    if (!this.dockview) return;
    
    const panel = this.dockview.getPanel('liveMetrics');
    if (panel) {
      const element = panel.view.content.element;
      element.innerHTML = `
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" style="grid-auto-rows: 200px;">
          ${this.renderDynamicCharts()}
        </div>
      `;
      
      // Re-initialize charts
      setTimeout(() => {
        this.initDynamicCharts();
      }, 10);
    }
  }


  private renderSystemArchitectureOnly(): string {
    if (!this.state.currentSystem || !this.systemDiagram) {
      return `
        <div class="flex items-center justify-center h-full">
          <div class="text-center text-gray-400">
            <div class="text-xl mb-2">No System Loaded</div>
            <div class="text-sm">Load an SDL file to view system architecture</div>
          </div>
        </div>
      `;
    }

    return `
      <div class="w-full h-full flex flex-col">
        <div class="text-center mb-4">
          <h3 class="text-lg font-semibold text-gray-300">${this.systemDiagram.systemName}</h3>
        </div>
        
        <!-- SVG System Architecture -->
        <div id="architecture-svg-container" class="flex-1 overflow-auto">
          ${this.renderSystemDiagramSVG()}
        </div>

      </div>
    `;
  }

  private generateDotFile(): string {
    if (!this.systemDiagram) return '';

    const systemName = this.systemDiagram.systemName || 'System';
    let dotContent = `digraph "${systemName}" {\n`;
    dotContent += `  rankdir=${this.layoutTopToBottom ? "TB": " LR"};\n`;
    dotContent += `  bgcolor="#1a1a1a";\n`;
    dotContent += `  node [fontname="Monaco,Menlo,monospace" fontcolor="white" style=filled];\n`;
    dotContent += `  edge [color="#9ca3af" arrowhead="normal" penwidth=2];\n`;
    dotContent += `  graph [ranksep=1.0 nodesep=0.8 pad=0.5];\n\n`;

    // Generate method-level nodes directly (nodes are already in "component:method" format)
    this.systemDiagram?.nodes?.forEach((node) => {
      // Each node represents a method with traffic rate
      const nodeId = node.id.replace(':', '_'); // Replace colon for DOT syntax
      
      // Get icon for the node
      const icon = this.getIconForNode(node);
      const displayLabel = `${icon} ${node.id}\\n${node.traffic}`;
      
      // Check if this is an internal component method
      const isInternal = node.type.includes('(internal)') || node.id.includes('.pool:') || node.id.includes('.driverTable:');
      
      dotContent += `  "${nodeId}" [label="${displayLabel}"`;
      dotContent += ` shape=box style="filled,rounded"`;
      dotContent += ` fillcolor="${isInternal ? '#312e81' : '#1f2937'}"`;
      dotContent += ` fontcolor="${isInternal ? '#c7d2fe' : '#a3e635'}"`;
      dotContent += ` fontsize=${isInternal ? '11' : '12'}`;
      dotContent += ` margin=0.15 penwidth=1];\n`;
    });

    dotContent += `\n`;

    // Generate edges between method nodes
    this.systemDiagram?.edges?.forEach(edge => {
      const fromNodeId = edge.fromId.replace(':', '_');
      const toNodeId = edge.toId.replace(':', '_');
      
      // Use generator-specific color if available, otherwise default
      const edgeColor = edge.color || "#9ca3af";
      let edgeStyle = ` color="${edgeColor}" fontcolor="${edgeColor}" fontsize=10`;
      
      // Add label if available
      const label = edge.label ? ` label="${edge.label}"` : '';
      
      dotContent += `  "${fromNodeId}" -> "${toNodeId}"[${label}${edgeStyle}];\n`;
    });

    dotContent += `}\n`;
    return dotContent;
  }

  private getIconForNode(node: any): string {
    // Use the icon field if available
    if (node.icon) {
      const iconMap: Record<string, string> = {
        'cache': '💾',
        'database': '🗄️',
        'service': '⚙️',
        'gateway': '🚪',
        'api': '🔌',
        'queue': '📋',
        'pool': '🏊',
        'network': '🌐',
        'storage': '💿',
        'index': '📇',
        'component': '📦'
      };
      return iconMap[node.icon] || '📦';
    }
    
    // Fallback to type-based icons for backward compatibility
    const type = node.type?.toLowerCase() || '';
    if (type.includes('cache')) return '💾';
    if (type.includes('database') || type.includes('db')) return '🗄️';
    if (type.includes('gateway')) return '🚪';
    if (type.includes('service')) return '⚙️';
    if (type.includes('queue')) return '📋';
    if (type.includes('pool')) return '🏊';
    if (type.includes('api')) return '🔌';
    
    return '📦'; // default component icon
  }

  private async convertDotToSVG(dotContent: string): Promise<string> {
    try {
      if (!this.graphviz) {
        console.warn('Graphviz not loaded yet, falling back to placeholder');
        return this.generateFallbackSVG();
      }

      // Use WASM Graphviz to convert dot to SVG
      const svg = this.graphviz.dot(dotContent);
      return svg;
    } catch (error) {
      console.warn('Dot conversion error:', error);
      return this.generateFallbackSVG();
    }
  }

  private generateFallbackSVG(): string {
    return `
      <svg width="300" height="200" viewBox="0 0 300 200" xmlns="http://www.w3.org/2000/svg">
        <rect width="100%" height="100%" fill="#1f2937" stroke="#4b5563"/>
        <text x="150" y="100" text-anchor="middle" fill="#9ca3af" font-family="monospace">
          Dot rendering unavailable
        </text>
      </svg>
    `;
  }

  private renderSystemDiagramSVG(): string {
    if (!this.systemDiagram) return '';

    // Generate dot file content
    const dotContent = this.generateDotFile();
    
    // Convert to SVG and render
    this.convertDotToSVG(dotContent).then(svg => {
      const container = document.getElementById('architecture-svg-container');
      if (container) {
        container.innerHTML = svg;
        const childSvg = container.querySelector("svg") as Element
        if (this.layoutTopToBottom) {
          childSvg.setAttribute("height", "100%")
          childSvg.removeAttribute("width")
          // childSvg.setAttribute("width", "100%")
        } else {
          childSvg.setAttribute("width", "100%")
          childSvg.removeAttribute("height")
          // childSvg.setAttribute("height", "100%")
        }
      }
    });
    
    // Return placeholder while conversion happens
    return `
      <div id="architecture-svg-container" style="width: 100%; height: 100%; display: flex; align-items: center; justify-content: center;">
        <div style="color: #9ca3af; font-family: monospace;">
          ${this.graphviz ? 'Rendering diagram...' : 'Loading Graphviz...'}
        </div>
      </div>
    `;
  }

  private renderGenerateControls(): string {
    const hasGenerators = this.state.generateCalls.length > 0;
    const hasEnabledGenerators = this.state.generateCalls.some(g => g.enabled);
    
    return `
      <div class="space-y-2">
        <div class="flex items-center justify-between mb-2">
          <div class="text-xs text-gray-400">Traffic Generation</div>
          <div class="flex items-center gap-2">
            <button id="refresh-generators" class="btn btn-outline text-xs px-2 py-1" title="Refresh generators">
              🔄
            </button>
            ${hasGenerators ? `
              <button id="toggle-all-generators" class="btn btn-outline text-xs px-2 py-1">
                ${hasEnabledGenerators ? '⏸️ Stop All' : '▶️ Start All'}
              </button>
            ` : ''}
          </div>
        </div>
        
        ${hasGenerators ? this.state.generateCalls.map(call => `
          <div class="bg-gray-800 border border-gray-600 rounded p-2">
            <div class="flex items-center justify-between mb-1">
              <span class="text-xs font-medium truncate">${call.name}</span>
              <div class="flex items-center gap-2">
                <div class="flex items-center border border-gray-600 rounded overflow-hidden">
                  <button class="bg-gray-700 hover:bg-gray-600 px-2 py-1 text-xs" 
                          data-rate-dec-id="${call.id}">−</button>
                  <input type="number" 
                         min="0" max="20" step="0.1" 
                         value="${call.rate.toFixed(1)}"
                         data-rate-input-id="${call.id}"
                         class="w-12 px-1 py-1 text-xs text-center bg-gray-800 border-0 outline-none [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none">
                  <button class="bg-gray-700 hover:bg-gray-600 px-2 py-1 text-xs" 
                          data-rate-inc-id="${call.id}">+</button>
                </div>
                <input type="checkbox" ${call.enabled ? 'checked' : ''} 
                       data-generate-id="${call.id}"
                       class="form-checkbox h-4 w-4 text-blue-600">
              </div>
            </div>
            
            <div class="text-xs text-gray-500">${call.target}</div>
          </div>
        `).join('') : `
          <div class="flex items-center justify-center py-4">
            <div class="text-center text-gray-500">
              <div class="text-xs">No Traffic Generators</div>
            </div>
          </div>
        `}
        
        <button id="add-generator" class="w-full btn btn-outline text-xs py-1" ${!this.state.currentSystem ? 'disabled' : ''}>
          + Add Generator
        </button>
      </div>
    `;
  }

  private renderDynamicCharts(): string {
    const charts = Object.values(this.state.dynamicCharts);
    
    if (charts.length === 0) {
      return `
        <div class="flex items-center justify-center h-full col-span-full">
          <div class="text-center text-gray-400">
            <div class="text-lg mb-2">No Metrics Available</div>
            <div class="text-sm">Load a system and add measurements to view live metrics</div>
          </div>
        </div>
      `;
    }
    
    return charts.map(chart => `
      <div class="bg-gray-800 border border-gray-600 rounded p-3 flex flex-col h-full">
        <h4 class="text-sm font-medium text-gray-300 mb-2 text-xs">${chart.title}</h4>
        <div class="flex-grow relative min-h-0">
          <canvas id="chart-${chart.chartName}" class="w-full h-full"></canvas>
        </div>
      </div>
    `).join('');
  }



  private setupInteractivity() {
    // Reset layout button
    const resetLayoutBtn = document.getElementById('reset-layout-btn');
    resetLayoutBtn?.addEventListener('click', () => {
      if (confirm('Reset layout to default? This will reload the page.')) {
        this.resetLayout();
      }
    });

    // Toggle all generators button
    const toggleAllBtn = document.getElementById('toggle-all-generators');
    toggleAllBtn?.addEventListener('click', () => this.toggleAllGenerators());
    
    // Refresh generators button
    const refreshBtn = document.getElementById('refresh-generators');
    refreshBtn?.addEventListener('click', () => {
      console.log('🔄 Manual generator refresh triggered');
      this.refreshGenerators();
    });

    // Parameter sliders
    this.parameters.forEach(param => {
      if (param.type === 'number') {
        const slider = document.getElementById(`slider-${param.path}`) as HTMLInputElement;
        slider?.addEventListener('input', (e) => {
          const value = parseFloat((e.target as HTMLInputElement).value);
          param.value = value;
          this.setParameter(param.path, value);
          // Parameter changes only affect the system parameters panel
        });
      }
    });

    // Generate controls
    this.state.generateCalls.forEach(call => {
      // Enable/disable checkbox
      const checkbox = document.querySelector(`[data-generate-id="${call.id}"]`) as HTMLInputElement;
      checkbox?.addEventListener('change', (e) => {
        call.enabled = (e.target as HTMLInputElement).checked;
        this.handleGenerateToggle(call);
      });

      // Rate input field
      const rateInput = document.querySelector(`[data-rate-input-id="${call.id}"]`) as HTMLInputElement;
      rateInput?.addEventListener('input', (e) => {
        const value = parseFloat((e.target as HTMLInputElement).value);
        if (!isNaN(value) && value >= 0 && value <= 20) {
          call.rate = value; // Support float values for fine-grained control
          this.handleGenerateRateChange(call);
        }
      });

      // Rate decrement button
      const rateDecBtn = document.querySelector(`[data-rate-dec-id="${call.id}"]`) as HTMLButtonElement;
      rateDecBtn?.addEventListener('click', () => {
        const newRate = Math.max(0, call.rate - 0.1);
        call.rate = Math.round(newRate * 10) / 10; // Round to 1 decimal place
        if (rateInput) rateInput.value = call.rate.toFixed(1);
        this.handleGenerateRateChange(call);
      });

      // Rate increment button
      const rateIncBtn = document.querySelector(`[data-rate-inc-id="${call.id}"]`) as HTMLButtonElement;
      rateIncBtn?.addEventListener('click', () => {
        const newRate = Math.min(20, call.rate + 0.1);
        call.rate = Math.round(newRate * 10) / 10; // Round to 1 decimal place
        if (rateInput) rateInput.value = call.rate.toFixed(1);
        this.handleGenerateRateChange(call);
      });
    });

    // Add generator button
    const addGeneratorBtn = document.getElementById('add-generator');
    addGeneratorBtn?.addEventListener('click', () => this.addNewGenerator());
  }

  private handleGenerateToggle(call: GenerateCall) {
    console.log(`${call.enabled ? 'Starting' : 'Stopping'} generator: ${call.name}`);
    // TODO: Implement actual traffic generation start/stop
    if (call.enabled) {
      this.startTrafficGeneration(call);
    } else {
      this.stopTrafficGeneration(call);
    }
  }

  private handleGenerateRateChange(call: GenerateCall) {
    console.log(`Changing rate for ${call.name} to ${call.rate} RPS`);
    
    // Clear any existing timeout
    if (this.generatorUpdateTimeout) {
      clearTimeout(this.generatorUpdateTimeout);
    }
    
    // Set a new timeout to debounce the API call
    this.generatorUpdateTimeout = window.setTimeout(() => {
      // Always update the rate, even if generator is not enabled
      // This enables real-time parameter tuning like in the "incredible machine"
      this.updateTrafficRate(call);
    }, 300); // 300ms debounce delay
  }

  private async startTrafficGeneration(call: GenerateCall) {
    try {
      const result = await this.api.startGenerator(call.id);
      if (!result.success) {
        throw new Error('Failed to resume generator');
      }
      console.log(`✅ Started traffic generation for ${call.target} at ${call.rate} RPS`);
    } catch (error) {
      console.error(`❌ Failed to start traffic generation:`, error);
      this.showError(`Failed to start traffic generation: ${error}`);
    }
  }

  private async stopTrafficGeneration(call: GenerateCall) {
    try {
      const result = await this.api.stopGenerator(call.id);
      if (!result.success) {
        throw new Error('Failed to pause generator');
      }
      console.log(`✅ Stopped traffic generation for ${call.target}`);
    } catch (error) {
      console.error(`❌ Failed to stop traffic generation:`, error);
      this.showError(`Failed to stop traffic generation: ${error}`);
    }
  }

  private async updateTrafficRate(call: GenerateCall) {
    try {
      // Set flag to prevent UI overwrites during update
      this.isUpdatingGenerators = true;
      
      // Update only the rate using the new updateGeneratorRate method
      const result = await this.api.updateGeneratorRate(call.id, call.rate);
      
      if (!result.success) {
        throw new Error('Failed to update generator rate');
      }
      
      // Apply flows to propagate the rate change throughout the system
      await this.api.evaluateFlows("auto");
      
      console.log(`✅ Updated traffic rate for ${call.target} to ${call.rate} RPS with flow evaluation`);
      
      // Refresh the system diagram to show updated flow rates
      setTimeout(() => {
        this.loadSystemDiagram();
        // Allow UI updates again after a delay
        this.isUpdatingGenerators = false;
      }, 1000); // Give time for the update to propagate
      
    } catch (error) {
      console.error(`❌ Failed to update traffic rate:`, error);
      this.showError(`Failed to update traffic rate: ${error}`);
      // Reset flag on error
      this.isUpdatingGenerators = false;
    }
  }

  private async addNewGenerator() {
    // TODO: Show a form/dialog to collect generator details
    // For now, just show a message that this feature needs implementation
    this.showError('Add Generator form not yet implemented. Use the Canvas API directly for now.');
  }

  private async toggleAllGenerators() {
    try {
      const hasEnabledGenerators = this.state.generateCalls.some(g => g.enabled);
      
      if (hasEnabledGenerators) {
        // Stop all generators
        const result = await this.api.stopAllGenerators();
        if (!result.success) {
          throw new Error('Failed to stop all generators');
        }
        console.log('✅ Stopped all generators');
      } else {
        // Start all generators
        const result = await this.api.startAllGenerators();
        if (!result.success) {
          throw new Error('Failed to start all generators');
        }
        console.log('✅ Started all generators');
      }
    } catch (error) {
      console.error('❌ Failed to toggle generators:', error);
      this.showError(`Failed to toggle generators: ${error}`);
    }
  }

  private async refreshGenerators() {
    // Skip refresh if we're updating generators to avoid UI overwrites
    if (this.isUpdatingGenerators) {
      return;
    }
    
    try {
      const response = await this.api.getGenerators();
      if (response.success && response.data) {
        const newGenerators = Object.values(response.data).map(gen => ({
          id: gen.id,
          name: gen.name,
          target: `${gen.component}.${gen.method}`, // Combine component and method
          rate: gen.rate,
          enabled: gen.enabled
        }));
        
        // Check if generators have changed
        const hasChanged = JSON.stringify(newGenerators) !== JSON.stringify(this.state.generateCalls);
        
        if (hasChanged) {
          this.state.generateCalls = newGenerators;
          this.updateTrafficGenerationPanel();
          // Also reload system diagram to show updated arrival rates
          this.loadSystemDiagram();
        }
      }
    } catch (error) {
      console.error('❌ Failed to refresh generators:', error);
    }
  }

  /*
  private async refreshMetrics() {
    try {
      const response = await this.api.getMetrics();
      if (response.success && response.data) {
        // Clear existing dynamic charts
        this.state.dynamicCharts = {};
        
        // Convert metrics to dynamic charts
        Object.values(response.data).forEach((metric: any) => {
          if (metric.enabled) {
            const target = metric.methods && metric.methods.length > 0 
              ? `${metric.component}.${metric.methods[0]}` 
              : metric.component;
            this.state.dynamicCharts[metric.id] = {
              chartName: metric.id,
              metricName: metric.metricType,
              target: target, // Store the actual target for API calls
              data: [],
              labels: [],
              title: metric.name || `${target} - ${metric.metricType}`
            };
          }
        });
        
        // Re-initialize charts after data changes
        this.initDynamicCharts();
      }
    } catch (error) {
      console.error('❌ Failed to refresh measurements:', error);
    }
  }
 */

  // Method to simulate server adding a new metric via canvas.Measure()
  public addMetricFromServer(metricName: string, chartTitle: string) {
    const chartName = metricName.replace(/[^a-zA-Z0-9]/g, '-').toLowerCase();
    
    this.state.dynamicCharts[metricName] = {
      chartName,
      metricName,
      data: [],
      labels: [],
      title: chartTitle
    };
    
    console.log(`📊 Server added new metric: ${metricName} -> "${chartTitle}"`);
    // Chart updates handled by real-time chart updates
  }

  private initDynamicCharts() {
    Object.values(this.state.dynamicCharts).forEach(chartData => {
      const canvas = document.getElementById(`chart-${chartData.chartName}`) as HTMLCanvasElement;
      if (!canvas) return;

      // Destroy existing chart if it exists
      if (this.charts[chartData.chartName]) {
        this.charts[chartData.chartName].destroy();
        delete this.charts[chartData.chartName];
      }

      // Determine chart color based on metric type
      let borderColor = 'rgb(59, 130, 246)'; // Default blue
      let backgroundColor = 'rgba(59, 130, 246, 0.1)';
      
      if (chartData.metricName.includes('p95Latency') || chartData.metricName.includes('p99Latency')) {
        borderColor = 'rgb(239, 68, 68)'; // Red for latency
        backgroundColor = 'rgba(239, 68, 68, 0.1)';
      } else if (chartData.metricName.includes('qps')) {
        borderColor = 'rgb(34, 197, 94)'; // Green for QPS
        backgroundColor = 'rgba(34, 197, 94, 0.1)';
      } else if (chartData.metricName.includes('errorRate')) {
        borderColor = 'rgb(245, 158, 11)'; // Orange for error rate
        backgroundColor = 'rgba(245, 158, 11, 0.1)';
      } else if (chartData.metricName.includes('HitRate')) {
        borderColor = 'rgb(168, 85, 247)'; // Purple for cache hit rate
        backgroundColor = 'rgba(168, 85, 247, 0.1)';
      } else if (chartData.metricName.includes('utilization')) {
        borderColor = 'rgb(14, 165, 233)'; // Sky blue for utilization
        backgroundColor = 'rgba(14, 165, 233, 0.1)';
      } else if (chartData.metricName.includes('memory')) {
        borderColor = 'rgb(236, 72, 153)'; // Pink for memory
        backgroundColor = 'rgba(236, 72, 153, 0.1)';
      }

      const config: ChartConfiguration = {
        type: 'line',
        data: {
          labels: [],
          datasets: [{
            label: chartData.title,
            data: [],
            borderColor,
            backgroundColor,
            borderWidth: 2,
            fill: true,
            tension: 0.1
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          scales: {
            y: {
              beginAtZero: true,
              grid: { color: 'rgba(55, 65, 81, 0.5)' },
              ticks: { color: 'rgba(156, 163, 175, 1)', font: { size: 10 } }
            },
            x: {
              grid: { color: 'rgba(55, 65, 81, 0.5)' },
              ticks: { color: 'rgba(156, 163, 175, 1)', font: { size: 10 } }
            }
          },
          plugins: {
            legend: {
              display: false // Hide legend to save space
            }
          }
        }
      };

      this.charts[chartData.chartName] = new Chart(canvas, config);
    });
  }

  private async startChartUpdates() {
    // Stop any existing polling
    this.stopChartUpdates();
    
    // Get all metric IDs from current charts - use the keys which are the actual metric IDs
    const metricIds = Object.keys(this.state.dynamicCharts);
    
    if (metricIds.length === 0) {
      console.log('📊 No metrics to stream');
      return;
    }
    
    // Create abort controller for clean shutdown
    this.metricStreamController = new AbortController();
    
    try {
      console.log('🔄 Starting metric stream for:', metricIds);
      
      // Start streaming metrics
      for await (const response of this.api.streamMetrics(metricIds)) {
        if (this.metricStreamController?.signal.aborted) {
          break;
        }
        
        if (response.success && response.data) {
          // Process each metric update
          for (const update of response.data) {
            console.log('📊 Metric update:', update.metricId, update.point);
            this.updateChartWithStreamData(update.metricId, update.point);
          }
        }
      }
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        console.log('📊 Metric stream stopped');
      } else {
        console.error('❌ Metric stream error:', error);
        // Fallback to polling on stream error
        this.startPollingFallback();
      }
    }
  }
  
  private updateChartWithStreamData(metricId: string, point: any) {
    // Find the chart for this metric - directly access by metric ID
    const chartData = this.state.dynamicCharts[metricId];
    
    if (!chartData) return;
    
    const chart = this.charts[chartData.chartName];
    if (!chart) return;
    
    // Add the new point to the chart
    // Convert Unix timestamp (seconds) to milliseconds for JavaScript Date
    const timestamp = new Date(point.timestamp * 1000).toLocaleTimeString();
    
    // Keep only last 20 points for performance
    if (chart.data.labels!.length >= 20) {
      chart.data.labels!.shift();
      chart.data.datasets[0].data.shift();
    }
    
    chart.data.labels!.push(timestamp);
    chart.data.datasets[0].data.push(point.value);
    
    // Update chart without animation for smooth real-time updates
    chart.update('none');
  }
  
  private startPollingFallback() {
    console.log('⚠️ Falling back to polling mode');
    // Update charts every 2 seconds using polling
    this.chartUpdateInterval = window.setInterval(() => {
      this.updateDynamicCharts();
    }, 2000);
  }

  private stopChartUpdates() {
    // Stop streaming
    if (this.metricStreamController) {
      this.metricStreamController.abort();
      this.metricStreamController = null;
    }
    
    // Stop polling
    if (this.chartUpdateInterval) {
      clearInterval(this.chartUpdateInterval);
      this.chartUpdateInterval = null;
    }
  }

  // Cleanup method for proper resource disposal
  public cleanup() {
    this.stopChartUpdates();
    this.stopGeneratorPolling();
    
    // Destroy all charts
    Object.values(this.charts).forEach(chart => {
      chart.destroy();
    });
    this.charts = {};

    // Cleanup DockView
    if (this.dockview) {
      this.dockview.dispose();
      this.dockview = null;
    }
  }

  private saveLayoutConfig() {
    if (!this.dockview) return;
    
    try {
      const layoutJson = this.dockview.toJSON();
      localStorage.setItem('sdl-dockview-layout', JSON.stringify(layoutJson));
      console.log('💾 DockView layout saved to localStorage');
    } catch (error) {
      console.warn('Failed to save DockView layout:', error);
    }
  }

  private loadLayoutConfig(): any | null {
    try {
      const saved = localStorage.getItem('sdl-dockview-layout');
      if (saved) {
        const layoutJson = JSON.parse(saved);
        console.log('📂 DockView layout loaded from localStorage');
        return layoutJson;
      }
    } catch (error) {
      console.warn('Failed to load DockView layout:', error);
      localStorage.removeItem('sdl-dockview-layout');
    }
    return null;
  }

  private resetLayout() {
    // Clear saved layout and recreate with default configuration
    localStorage.removeItem('sdl-dockview-layout');
    this.initializeLayout();
  }

  // NO LONGER SHOWING MEASUREMENTS PANEL - LiveMetrics panel is doing the same thing anyway
  /* 
  private updateMeasurementsPanel() {
    if (!this.dockview) return;
    
    const panel = this.dockview.getPanel('measurements');
    if (panel) {
      const element = panel.view.content.element;
      element.innerHTML = this.renderMeasurements();
    }
  }

  private renderMeasurements(): string {
    const measurements = Object.values(this.state.dynamicCharts);
    
    if (measurements.length === 0) {
      return `
        <div class="flex items-center justify-center h-full">
          <div class="text-center text-gray-400">
            <div class="text-sm">No Measurements</div>
            <div class="text-xs mt-1">Add measurements to track system metrics</div>
          </div>
        </div>
      `;
    }

    return `
      <div class="space-y-3">
        <div class="text-xs text-gray-400 mb-3">
          Active measurements for system monitoring
        </div>
        ${measurements.map(measurement => `
          <div class="bg-gray-800 border border-gray-600 rounded p-2">
            <div class="text-xs font-medium text-gray-300 mb-1">${measurement.title}</div>
            <div class="text-xs text-gray-500">Target: ${measurement.target || measurement.metricName}</div>
            <div class="text-xs text-green-400">Type: ${measurement.metricName}</div>
          </div>
        `).join('')}
        
        <button class="w-full btn btn-outline text-xs py-1" ${!this.state.currentSystem ? 'disabled' : ''}>
          + Add Measurement
        </button>
      </div>
    `;
  }

 */
}
