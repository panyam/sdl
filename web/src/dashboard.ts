import { CanvasAPI } from './canvas-api.js';
import { DashboardState, ParameterConfig, WebSocketMessage, GenerateCall, SystemDiagram, MeasurementDataPoint } from './types.js';
import { Chart, ChartConfiguration } from 'chart.js/auto';
import { Graphviz } from "@hpcc-js/wasm";
import { DockviewApi, DockviewComponent } from 'dockview-core';

export class Dashboard {
  private api: CanvasAPI;
  private state: DashboardState;
  private charts: Record<string, Chart> = {};
  private systemDiagram: SystemDiagram | null = null;
  private chartUpdateInterval: number | null = null;
  private graphviz: any = null; // Will be initialized asynchronously
  private dockview: DockviewApi | null = null;

  // Parameter configurations - populated when a system is loaded
  private parameters: ParameterConfig[] = [];

  constructor() {
    this.api = new CanvasAPI();
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
      // First render the layout structure
      this.render();
      
      // Then setup WebSocket after layout is ready
      setTimeout(() => {
        this.setupEventListeners();
        this.startChartUpdates();
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
      if (stateResponse.success && stateResponse.data) {
        const canvasState = stateResponse.data;
        
        // Update dashboard state from Canvas state
        this.state.currentFile = canvasState.activeFile;
        this.state.currentSystem = canvasState.activeSystem;
        
        // If there's an active system, load its diagram
        if (this.state.currentSystem) {
          this.loadSystemDiagram();
        }
        
        // Convert Canvas generators to dashboard generate calls
        this.state.generateCalls = Object.values(canvasState.generators || {}).map(gen => ({
          id: gen.id,
          name: gen.name,
          target: gen.target,
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
          target: gen.target,
          rate: gen.rate,
          enabled: gen.enabled
        }));
      } else {
        console.log('⚠️ No generators found or failed to load');
      }

      // Load measurements and create dynamic charts
      console.log('🔄 Loading measurements from API...');
      const measurementsResponse = await this.api.getMeasurements();
      if (measurementsResponse.success && measurementsResponse.data) {
        console.log('✅ Measurements loaded:', Object.keys(measurementsResponse.data).length);
        // Convert measurements to dynamic charts
        Object.values(measurementsResponse.data).forEach(measurement => {
          if (measurement.enabled) {
            this.state.dynamicCharts[measurement.id] = {
              chartName: measurement.id,
              metricName: measurement.metricType,
              target: measurement.target, // Store the actual target for API calls
              data: [],
              labels: [],
              title: measurement.name || `${measurement.target} - ${measurement.metricType}`
            };
          }
        });
      } else {
        console.log('⚠️ No measurements found or failed to load');
      }

      // Update UI panels AFTER loading all data
      console.log('🔄 Updating all panels with loaded data...');
      this.updateAllPanels();

    } catch (error) {
      console.error('❌ Failed to load Canvas state:', error);
    }
  }

  private setupEventListeners() {
    this.api.onMessage((message: WebSocketMessage) => {
      this.handleWebSocketMessage(message);
    });
  }

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
        this.refreshMeasurements();
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
        this.refreshMeasurements();
        this.updateMeasurementsPanel();
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

  private async loadSystemDiagram() {
    try {
      const response = await this.api.getDiagram();
      if (response.success && response.data) {
        this.systemDiagram = response.data;
        console.log('📊 System diagram loaded:', this.systemDiagram);
        // Update only the system architecture panel
        this.updateSystemArchitecturePanel();
      } else {
        console.warn('Failed to load system diagram:', response.error);
        this.systemDiagram = null;
      }
    } catch (error) {
      console.error('❌ Failed to load system diagram:', error);
      this.systemDiagram = null;
    }
  }

  private async setParameter(path: string, value: any) {
    try {
      const result = await this.api.set(path, value);
      if (!result.success) {
        throw new Error(result.error);
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
        
        const response = await this.api.getMeasurementData(target, startTime, endTime);
        
        if (response.success && response.data && response.data.dataPoints) {
          const dataPoints = response.data.dataPoints;
          
          // Clear existing chart data
          chart.data.labels = [];
          chart.data.datasets[0].data = [];
          
          // Add new data points (limit to last 20 points for performance)
          const recentPoints = dataPoints.slice(-20);
          
          recentPoints.forEach((point: MeasurementDataPoint) => {
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
          <h1 class="text-xl font-bold text-blue-300 mb-2">SDL Canvas Dashboard</h1>
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
          case 'measurements':
            element.innerHTML = this.renderMeasurements();
            break;
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

    this.dockview!.addPanel({
      id: 'measurements',
      component: 'measurements', 
      title: 'Measurements',
      position: { direction: 'below' }
    });

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
    this.updateMeasurementsPanel();
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

  private updateMeasurementsPanel() {
    if (!this.dockview) return;
    
    const panel = this.dockview.getPanel('measurements');
    if (panel) {
      const element = panel.view.content.element;
      element.innerHTML = this.renderMeasurements();
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
    dotContent += `  rankdir=LR;\n`;
    dotContent += `  bgcolor="#1a1a1a";\n`;
    dotContent += `  compound=true;\n`;
    dotContent += `  node [fontname="Monaco,Menlo,monospace" fontcolor="white" style=filled];\n`;
    dotContent += `  edge [color="#9ca3af" arrowhead="normal" penwidth=2];\n`;
    dotContent += `  graph [ranksep=1.0 nodesep=0.8 pad=0.5];\n\n`;

    // Create a map to track method nodes and their connections
    const methodNodes: string[] = [];
    const edges: string[] = [];

    // Generate clusters (components) with method nodes inside
    this.systemDiagram.nodes.forEach((node) => {
      const clusterName = `cluster_${node.ID}`;
      const methods = node.Methods || [];
      
      dotContent += `  subgraph ${clusterName} {\n`;
      dotContent += `    label="${node.Name}\\n(${node.Type})";\n`;
      dotContent += `    style="filled,rounded";\n`;
      dotContent += `    fillcolor="#1f2937";\n`;
      dotContent += `    fontcolor="#60a5fa";\n`;
      dotContent += `    fontsize=16;\n`;
      dotContent += `    fontname="Monaco,Menlo,monospace";\n`;
      dotContent += `    color="#4b5563";\n`;
      dotContent += `    penwidth=2;\n`;
      dotContent += `    margin=16;\n\n`;

      if (methods.length > 0) {
        // Create method nodes inside the cluster
        methods.forEach((method) => {
          const methodNodeId = `${node.ID}_${method.Name}`;
          const traffic = this.getMethodTraffic(node.ID, method.Name);
          
          // Disabling return type in label for now
          // dotContent += `    ${methodNodeId} [label="${method.Name}() → ${method.ReturnType}\\n \\n ${traffic} rps"`;
          dotContent += `    ${methodNodeId} [label="${method.Name}()\\n \\n ${traffic} rps"`;
          dotContent += ` shape=box style="filled,rounded" fillcolor="#2d3748" fontcolor="#a3e635"`;
          dotContent += ` fontsize=12 fontname="Monaco,Menlo,monospace" margin=0.3 penwidth=1];\n`;
          
          methodNodes.push(methodNodeId);
        });
      } else {
        // Component with no methods - create a simple node
        const nodeId = `${node.ID}_component`;
        dotContent += `    ${nodeId} [label="No Methods\\n🔄 0 rps"`;
        dotContent += ` shape=box style="filled,rounded" fillcolor="#374151" fontcolor="#9ca3af"`;
        dotContent += ` fontsize=11 fontname="Monaco,Menlo,monospace" margin=0.2];\n`;
        methodNodes.push(nodeId);
      }
      
      dotContent += `  }\n\n`;
    });

    // Generate edges between method nodes based on system dependencies
    // For now, connect sequential nodes (this could be enhanced with actual dependency data)
    for (let i = 0; i < this.systemDiagram.nodes.length - 1; i++) {
      const currentNode = this.systemDiagram.nodes[i];
      const nextNode = this.systemDiagram.nodes[i + 1];
      
      const currentMethods = currentNode.Methods || [];
      const nextMethods = nextNode.Methods || [];
      
      if (currentMethods.length > 0 && nextMethods.length > 0) {
        // Connect first method of current to first method of next
        const fromMethod = `${currentNode.ID}_${currentMethods[0].Name}`;
        const toMethod = `${nextNode.ID}_${nextMethods[0].Name}`;
        edges.push(`  ${fromMethod} -> ${toMethod};`);
      }
    }

    // Add all edges
    edges.forEach(edge => {
      dotContent += `${edge}\n`;
    });

    dotContent += `}\n`;
    return dotContent;
  }

  private getMethodTraffic(componentId: string, methodName: string): number {
    // Get traffic for specific method from generator data
    const generator = this.state.generateCalls.find(g => 
      g.target.includes(componentId) && g.target.includes(methodName)
    );
    return generator?.enabled ? generator.rate : 0;
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


  private renderGenerateControls(): string {
    const hasGenerators = this.state.generateCalls.length > 0;
    const hasEnabledGenerators = this.state.generateCalls.some(g => g.enabled);
    
    return `
      <div class="space-y-2">
        <div class="flex items-center justify-between mb-2">
          <div class="text-xs text-gray-400">Traffic Generation</div>
          ${hasGenerators ? `
            <button id="toggle-all-generators" class="btn btn-outline text-xs px-2 py-1">
              ${hasEnabledGenerators ? '⏸️ Stop All' : '▶️ Start All'}
            </button>
          ` : ''}
        </div>
        
        ${hasGenerators ? this.state.generateCalls.map(call => `
          <div class="bg-gray-800 border border-gray-600 rounded p-2">
            <div class="flex items-center justify-between mb-1">
              <span class="text-xs font-medium truncate">${call.name}</span>
              <input type="checkbox" ${call.enabled ? 'checked' : ''} 
                     data-generate-id="${call.id}"
                     class="form-checkbox h-3 w-3 text-blue-600">
            </div>
            
            <div class="text-xs text-gray-500 mb-1">${call.target}</div>
            
            <div class="flex items-center gap-1">
              <input type="range" 
                     min="0" max="20" step="0.5" 
                     value="${call.rate}"
                     data-rate-id="${call.id}"
                     class="flex-1 h-1 bg-gray-600 rounded appearance-none slider">
              <span class="text-xs text-gray-300 w-10 text-right">${call.rate}</span>
            </div>
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

      // Rate slider
      const rateSlider = document.querySelector(`[data-rate-id="${call.id}"]`) as HTMLInputElement;
      rateSlider?.addEventListener('input', (e) => {
        const value = parseFloat((e.target as HTMLInputElement).value);
        call.rate = Math.round(value);
        this.handleGenerateRateChange(call);
        // Chart updates handled by real-time chart updates // Update content without recreating layout
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
    // TODO: Implement actual rate change
    if (call.enabled) {
      this.updateTrafficRate(call);
    }
  }

  private async startTrafficGeneration(call: GenerateCall) {
    try {
      const result = await this.api.resumeGenerator(call.id);
      if (!result.success) {
        throw new Error(result.error);
      }
      console.log(`✅ Started traffic generation for ${call.target} at ${call.rate} RPS`);
    } catch (error) {
      console.error(`❌ Failed to start traffic generation:`, error);
      this.showError(`Failed to start traffic generation: ${error}`);
    }
  }

  private async stopTrafficGeneration(call: GenerateCall) {
    try {
      const result = await this.api.pauseGenerator(call.id);
      if (!result.success) {
        throw new Error(result.error);
      }
      console.log(`✅ Stopped traffic generation for ${call.target}`);
    } catch (error) {
      console.error(`❌ Failed to stop traffic generation:`, error);
      this.showError(`Failed to stop traffic generation: ${error}`);
    }
  }

  private async updateTrafficRate(call: GenerateCall) {
    try {
      const config = {
        id: call.id,
        name: call.name,
        target: call.target,
        rate: Math.round(call.rate),
        enabled: call.enabled
      };
      const result = await this.api.updateGenerator(call.id, config);
      if (!result.success) {
        throw new Error(result.error);
      }
      console.log(`✅ Updated traffic rate for ${call.target} to ${call.rate} RPS`);
    } catch (error) {
      console.error(`❌ Failed to update traffic rate:`, error);
      this.showError(`Failed to update traffic rate: ${error}`);
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
        const result = await this.api.stopGenerators();
        if (!result.success) {
          throw new Error(result.error);
        }
        console.log('✅ Stopped all generators');
      } else {
        // Start all generators
        const result = await this.api.startGenerators();
        if (!result.success) {
          throw new Error(result.error);
        }
        console.log('✅ Started all generators');
      }
    } catch (error) {
      console.error('❌ Failed to toggle generators:', error);
      this.showError(`Failed to toggle generators: ${error}`);
    }
  }

  private async refreshGenerators() {
    try {
      const response = await this.api.getGenerators();
      if (response.success && response.data) {
        this.state.generateCalls = Object.values(response.data).map(gen => ({
          id: gen.id,
          name: gen.name,
          target: gen.target,
          rate: gen.rate,
          enabled: gen.enabled
        }));
      }
    } catch (error) {
      console.error('❌ Failed to refresh generators:', error);
    }
  }

  private async refreshMeasurements() {
    try {
      const response = await this.api.getMeasurements();
      if (response.success && response.data) {
        // Clear existing dynamic charts
        this.state.dynamicCharts = {};
        
        // Convert measurements to dynamic charts
        Object.values(response.data).forEach(measurement => {
          if (measurement.enabled) {
            this.state.dynamicCharts[measurement.id] = {
              chartName: measurement.id,
              metricName: measurement.metricType,
              target: measurement.target, // Store the actual target for API calls
              data: [],
              labels: [],
              title: measurement.name || `${measurement.target} - ${measurement.metricType}`
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

  private startChartUpdates() {
    // Update charts every 2 seconds
    this.chartUpdateInterval = window.setInterval(() => {
      this.updateDynamicCharts();
    }, 2000);
  }

  private stopChartUpdates() {
    if (this.chartUpdateInterval) {
      clearInterval(this.chartUpdateInterval);
      this.chartUpdateInterval = null;
    }
  }

  // Cleanup method for proper resource disposal
  public cleanup() {
    this.stopChartUpdates();
    this.api.disconnect();
    
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
}
