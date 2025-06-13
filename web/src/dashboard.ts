import { CanvasAPI } from './canvas-api.js';
import { DashboardState, ParameterConfig, WebSocketMessage, GenerateCall, SystemDiagram } from './types.js';
import { Chart, ChartConfiguration } from 'chart.js/auto';

export class Dashboard {
  private api: CanvasAPI;
  private state: DashboardState;
  private charts: Record<string, Chart> = {};
  private systemDiagram: SystemDiagram | null = null;

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

    this.setupEventListeners();
    this.initialize();
  }

  private async initialize() {
    try {
      // Load current Canvas state to see if there's an existing session
      await this.loadCanvasState();
      
      this.render();
    } catch (error) {
      console.error('❌ Failed to initialize dashboard:', error);
      this.render(); // Render anyway with empty state
    }
  }

  private async loadCanvasState() {
    try {
      const stateResponse = await this.api.getState();
      if (stateResponse.success && stateResponse.data) {
        const canvasState = stateResponse.data;
        
        // Update dashboard state from Canvas state
        this.state.currentFile = canvasState.activeFile;
        this.state.currentSystem = canvasState.activeSystem;
        
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
      const generatorsResponse = await this.api.getGenerators();
      if (generatorsResponse.success && generatorsResponse.data) {
        this.state.generateCalls = Object.values(generatorsResponse.data).map(gen => ({
          id: gen.id,
          name: gen.name,
          target: gen.target,
          rate: gen.rate,
          enabled: gen.enabled
        }));
      }

      // Load measurements and create dynamic charts
      const measurementsResponse = await this.api.getMeasurements();
      if (measurementsResponse.success && measurementsResponse.data) {
        // Convert measurements to dynamic charts
        Object.values(measurementsResponse.data).forEach(measurement => {
          if (measurement.enabled) {
            this.state.dynamicCharts[measurement.metricType] = {
              chartName: measurement.id,
              metricName: measurement.metricType,
              data: [],
              labels: [],
              title: measurement.name
            };
          }
        });
      }

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
        break;
      case 'measurementAdded':
      case 'measurementUpdated':
      case 'measurementRemoved':
        this.refreshMeasurements();
        break;
      case 'plotGenerated':
        console.log('📊 Plot generated:', message.outputFile);
        break;
      case 'stateRestored':
        this.loadCanvasState();
        break;
    }

    this.render();
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
        // Re-render to show the diagram
        this.render();
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

  private updateDynamicCharts() {
    const now = Date.now();
    const timestamp = new Date(now).toLocaleTimeString();

    // Update each chart with simulated data
    Object.values(this.state.dynamicCharts).forEach(chartData => {
      const chart = this.charts[chartData.chartName];
      if (!chart) return;

      // Simulate metric values based on chart type
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
    });
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
      <div class="h-screen flex flex-col p-4">
        <!-- Header -->
        <div class="mb-4 flex-shrink-0">
          <h1 class="text-2xl font-bold text-blue-300 mb-2">SDL Canvas Dashboard</h1>
          <div class="flex items-center gap-4">
            <div class="flex items-center gap-2">
              <div class="w-2 h-2 rounded-full ${this.state.isConnected ? 'bg-green-400' : 'bg-red-400'}"></div>
              <span class="text-sm text-gray-400">${this.state.isConnected ? 'Connected' : 'Disconnected'}</span>
            </div>
            ${this.state.currentSystem ? `
              <div class="text-sm text-gray-300">
                <span class="text-gray-400">System:</span> ${this.state.currentSystem}
              </div>
            ` : ''}
          </div>
          <div id="error-display" class="hidden mt-2 p-2 bg-red-800 text-red-200 rounded text-sm"></div>
        </div>

        <!-- Row 1: System Architecture + Right Side Panels (50% height) -->
        <div class="flex gap-4 mb-4" style="height: 45%;">
          <!-- System Architecture Panel (70% width) -->
          <div class="panel overflow-hidden" style="flex: 0 0 70%;">
            <div class="panel-header">System Architecture</div>
            <div class="h-full overflow-y-auto p-4 space-y-4">
              ${this.renderSystemArchitectureOnly()}
            </div>
          </div>

          <!-- Right Side: Traffic Generation + System Parameters (30% width) -->
          <div class="flex flex-col gap-4" style="flex: 0 0 28%;">
            <!-- Traffic Generation Panel (Top 50%) -->
            <div class="panel overflow-hidden" style="height: 48%;">
              <div class="panel-header">Traffic Generation</div>
              <div class="p-3 h-full overflow-y-auto">
                ${this.renderGenerateControls()}
              </div>
            </div>

            <!-- System Parameters Panel (Bottom 50%) -->
            <div class="panel overflow-hidden" style="height: 48%;">
              <div class="panel-header">System Parameters</div>
              <div class="p-3 h-full overflow-y-auto">
                ${this.renderSystemParameters()}
              </div>
            </div>
          </div>
        </div>

        <!-- Row 2: Dynamic Metrics Grid (50% height) -->
        <div class="panel overflow-hidden" style="height: 45%;">
          <div class="panel-header">Live Metrics</div>
          <div class="p-4 h-full overflow-y-auto">
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" style="grid-auto-rows: 200px;">
              ${this.renderDynamicCharts()}
            </div>
          </div>
        </div>
      </div>
    `;

    this.setupInteractivity();
    this.initDynamicCharts();
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

    const getNodeColor = (type: string) => {
      const colors: Record<string, string> = {
        'ContactAppServer': 'bg-blue-900/30 border-blue-600',
        'ContactDatabase': 'bg-green-900/30 border-green-600', 
        'HashIndex': 'bg-purple-900/30 border-purple-600',
        'ResourcePool': 'bg-yellow-900/30 border-yellow-600',
        'Cache': 'bg-orange-900/30 border-orange-600'
      };
      return colors[type] || 'bg-gray-900/30 border-gray-600';
    };

    return `
      <div class="space-y-4">
        <div class="text-center mb-4">
          <h3 class="text-lg font-semibold text-gray-300">${this.systemDiagram.systemName}</h3>
        </div>
        
        <!-- Component Topology -->
        <div class="space-y-3">
          ${this.systemDiagram.nodes.map((node, index) => `
            <div class="component-box ${getNodeColor(node.Type)}">
              <div class="component-title text-sm font-bold">${node.Name}</div>
              <div class="text-xs text-gray-400 mt-1">${node.Type}</div>
            </div>
            ${index < this.systemDiagram!.nodes.length - 1 ? `
              <div class="text-center text-gray-400 text-lg">↓</div>
            ` : ''}
          `).join('')}
        </div>

        <!-- System Health Summary -->
        <div class="mt-6 p-3 bg-gray-800/50 rounded border border-gray-600">
          <h4 class="text-sm font-semibold text-gray-300 mb-2">System Health</h4>
          <div class="grid grid-cols-3 gap-2 text-xs">
            <div class="text-center">
              <div class="text-lg font-bold text-green-400">${this.state.metrics.successRate.toFixed(1)}%</div>
              <div class="text-xs text-gray-400">Success</div>
            </div>
            <div class="text-center">
              <div class="text-lg font-bold text-blue-400">${this.state.metrics.latency.toFixed(0)}ms</div>
              <div class="text-xs text-gray-400">Latency</div>
            </div>
            <div class="text-center">
              <div class="text-lg font-bold text-yellow-400">${this.state.metrics.load.toFixed(1)}</div>
              <div class="text-xs text-gray-400">Load</div>
            </div>
          </div>
        </div>
      </div>
    `;
  }

  private renderSystemParameters(): string {
    if (this.parameters.length === 0) {
      return `
        <div class="flex items-center justify-center h-full">
          <div class="text-center text-gray-400">
            <div class="text-sm">No Parameters Available</div>
            <div class="text-xs mt-1">Load a system to configure parameters</div>
          </div>
        </div>
      `;
    }

    return `
      <div class="space-y-3">
        <div class="text-xs text-gray-400 mb-3">
          Adjust system configuration parameters in real-time
        </div>
        ${this.parameters.map(param => this.renderParameterControl(param)).join('')}
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

  private renderParameterControl(param: ParameterConfig): string {
    if (param.type === 'number') {
      return `
        <div class="space-y-1">
          <label class="block text-xs text-gray-300">${param.name}</label>
          <div class="flex items-center gap-2">
            <input 
              type="range" 
              id="slider-${param.path}" 
              min="${param.min}" 
              max="${param.max}" 
              step="${param.step}"
              value="${param.value}"
              class="flex-1 h-1 bg-gray-600 rounded appearance-none slider"
            >
            <span class="text-xs text-gray-400 w-12 text-right">${param.value}</span>
          </div>
        </div>
      `;
    }
    return '';
  }


  private setupInteractivity() {
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
          this.render(); // Re-render to update display
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
        call.rate = value;
        this.handleGenerateRateChange(call);
        this.render(); // Re-render to update display
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
        rate: call.rate,
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
            this.state.dynamicCharts[measurement.metricType] = {
              chartName: measurement.id,
              metricName: measurement.metricType,
              data: [],
              labels: [],
              title: measurement.name
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
    this.render();
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
}