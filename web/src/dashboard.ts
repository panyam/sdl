import { CanvasAPI } from './canvas-api.js';
import { DashboardState, ParameterConfig, WebSocketMessage, GenerateCall } from './types.js';
import { Chart, ChartConfiguration } from 'chart.js/auto';

export class Dashboard {
  private api: CanvasAPI;
  private state: DashboardState;
  private charts: Record<string, Chart> = {};

  // Parameter configurations for the contacts service
  private parameters: ParameterConfig[] = [
    { name: 'Server Arrival Rate', path: 'server.pool.ArrivalRate', type: 'number', min: 0, max: 50, step: 0.5, value: 5.0 },
    { name: 'Server Pool Size', path: 'server.pool.Size', type: 'number', min: 1, max: 50, step: 1, value: 10 },
    { name: 'DB Arrival Rate', path: 'server.db.pool.ArrivalRate', type: 'number', min: 0, max: 30, step: 0.5, value: 3.0 },
    { name: 'DB Pool Size', path: 'server.db.pool.Size', type: 'number', min: 1, max: 20, step: 1, value: 5 },
    { name: 'Cache Hit Rate', path: 'server.db.CacheHitRate', type: 'number', min: 0, max: 1, step: 0.01, value: 0.4 },
  ];

  constructor() {
    this.api = new CanvasAPI();
    this.state = {
      isConnected: false,
      simulationResults: {},
      metrics: {
        load: 0,
        latency: 0,
        successRate: 0,
        serverUtilization: 0,
        dbUtilization: 0,
        cacheHitRate: 0.4,
        dbConnections: '0/5'
      },
      dynamicCharts: {
        'server.HandleLookup.p95Latency': {
          chartName: 'server-latency',
          metricName: 'server.HandleLookup.p95Latency',
          data: [],
          labels: [],
          title: 'Server P95 Latency'
        },
        'server.HandleLookup.qps': {
          chartName: 'server-qps',
          metricName: 'server.HandleLookup.qps',
          data: [],
          labels: [],
          title: 'Server QPS'
        },
        'database.QueryContact.p95Latency': {
          chartName: 'db-latency',
          metricName: 'database.QueryContact.p95Latency',
          data: [],
          labels: [],
          title: 'Database P95 Latency'
        },
        'server.HandleLookup.errorRate': {
          chartName: 'server-errors',
          metricName: 'server.HandleLookup.errorRate',
          data: [],
          labels: [],
          title: 'Server Error Rate'
        },
        'database.QueryContact.qps': {
          chartName: 'db-qps',
          metricName: 'database.QueryContact.qps',
          data: [],
          labels: [],
          title: 'Database QPS'
        },
        'cache.HitRate.percentage': {
          chartName: 'cache-hit',
          metricName: 'cache.HitRate.percentage',
          data: [],
          labels: [],
          title: 'Cache Hit Rate %'
        },
        'server.HandleLookup.p99Latency': {
          chartName: 'server-p99',
          metricName: 'server.HandleLookup.p99Latency',
          data: [],
          labels: [],
          title: 'Server P99 Latency'
        },
        'network.bandwidth.utilization': {
          chartName: 'network-util',
          metricName: 'network.bandwidth.utilization',
          data: [],
          labels: [],
          title: 'Network Utilization %'
        },
        'memory.heap.usage': {
          chartName: 'memory-heap',
          metricName: 'memory.heap.usage',
          data: [],
          labels: [],
          title: 'Heap Memory Usage MB'
        }
      },
      generateCalls: [
        {
          id: 'lookup-traffic',
          name: 'Contact Lookup Traffic',
          target: 'server.HandleLookup',
          rate: 5.0,
          enabled: true
        },
        {
          id: 'bulk-traffic',
          name: 'Bulk Load Traffic',
          target: 'server.HandleBulk',
          rate: 1.0,
          enabled: false
        }
      ]
    };

    this.setupEventListeners();
    this.render();
  }

  private setupEventListeners() {
    this.api.onMessage((message: WebSocketMessage) => {
      this.handleWebSocketMessage(message);
    });
  }

  private handleWebSocketMessage(message: WebSocketMessage) {
    console.log('📡 WebSocket message:', message);
    
    switch (message.type) {
      case 'fileLoaded':
        this.state.currentFile = message.file;
        break;
      case 'systemActivated':
        this.state.currentSystem = message.system;
        this.state.isConnected = true;
        break;
      case 'parameterChanged':
        // Update parameter value in UI
        const param = this.parameters.find(p => p.path === message.path);
        if (param) {
          param.value = message.value;
        }
        break;
      case 'simulationCompleted':
        // Trigger metrics update
        this.updateMetrics();
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

  private async loadContactsService() {
    try {
      const loadResult = await this.api.load('examples/contacts/contacts.sdl');
      if (!loadResult.success) {
        throw new Error(loadResult.error);
      }

      const useResult = await this.api.use('ContactsSystem');
      if (!useResult.success) {
        throw new Error(useResult.error);
      }

      console.log('✅ Contacts service loaded successfully');
    } catch (error) {
      console.error('❌ Failed to load contacts service:', error);
      this.showError(`Failed to load contacts service: ${error}`);
    }
  }

  private async runSimulation() {
    try {
      const result = await this.api.run('latest', 'server.HandleLookup', 1000);
      if (!result.success) {
        throw new Error(result.error);
      }
      console.log('✅ Simulation completed');
    } catch (error) {
      console.error('❌ Simulation failed:', error);
      this.showError(`Simulation failed: ${error}`);
    }
  }

  private async setParameter(path: string, value: any) {
    try {
      const result = await this.api.set(path, value);
      if (!result.success) {
        throw new Error(result.error);
      }
      console.log(`✅ Parameter ${path} set to ${value}`);
      
      // Auto-run simulation after parameter change
      await this.runSimulation();
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
    // Simple error display - in a real app, you'd want a proper notification system
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
            <button id="load-btn" class="btn btn-primary">Load Contacts Service</button>
            <button id="run-btn" class="btn btn-secondary">Run Simulation</button>
            <div class="flex items-center gap-2">
              <div class="w-2 h-2 rounded-full ${this.state.isConnected ? 'bg-green-400' : 'bg-red-400'}"></div>
              <span class="text-sm text-gray-400">${this.state.isConnected ? 'Connected' : 'Disconnected'}</span>
            </div>
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
    return `
      <div class="space-y-6">
        <!-- Enhanced System Architecture with more space -->
        <div class="component-box bg-blue-900/30 border-blue-600">
          <div class="component-title text-xl font-bold">ContactAppServer</div>
          <div class="grid grid-cols-2 gap-3 mt-3">
            <div class="component-detail">Pool: ${this.parameters.find(p => p.path === 'server.pool.Size')?.value}/10</div>
            <div class="component-detail">Load: ${this.state.metrics.load.toFixed(1)} RPS</div>
            <div class="component-detail">Utilization: ${this.state.metrics.serverUtilization.toFixed(0)}%</div>
            <div class="component-detail">Success: ${this.state.metrics.successRate.toFixed(1)}%</div>
          </div>
        </div>
        
        <div class="text-center text-gray-400 text-2xl font-bold">↓</div>
        
        <div class="component-box bg-green-900/30 border-green-600">
          <div class="component-title text-xl font-bold">ContactDatabase</div>
          <div class="grid grid-cols-2 gap-3 mt-3">
            <div class="component-detail">Pool: ${this.parameters.find(p => p.path === 'server.db.pool.Size')?.value}/5</div>
            <div class="component-detail">Cache: ${(this.state.metrics.cacheHitRate * 100).toFixed(0)}%</div>
            <div class="component-detail">Connections: ${this.state.metrics.dbConnections}</div>
            <div class="component-detail">DB Util: ${this.state.metrics.dbUtilization.toFixed(0)}%</div>
          </div>
        </div>
        
        <div class="text-center text-gray-400 text-2xl font-bold">↓</div>
        
        <div class="component-box bg-purple-900/30 border-purple-600">
          <div class="component-title text-xl font-bold">HashIndex</div>
          <div class="grid grid-cols-2 gap-3 mt-3">
            <div class="component-detail">Lookup: ~8ms</div>
            <div class="component-detail">Type: In-Memory</div>
            <div class="component-detail">Size: 10K entries</div>
            <div class="component-detail">Hit Rate: 98%</div>
          </div>
        </div>

        <!-- Additional system info can go here with more space -->
        <div class="mt-6 p-4 bg-gray-800/50 rounded border border-gray-600">
          <h4 class="text-lg font-semibold text-gray-300 mb-2">System Health</h4>
          <div class="grid grid-cols-3 gap-4">
            <div class="text-center">
              <div class="text-2xl font-bold text-green-400">${this.state.metrics.successRate.toFixed(1)}%</div>
              <div class="text-xs text-gray-400">Success Rate</div>
            </div>
            <div class="text-center">
              <div class="text-2xl font-bold text-blue-400">${this.state.metrics.latency.toFixed(0)}ms</div>
              <div class="text-xs text-gray-400">Avg Latency</div>
            </div>
            <div class="text-center">
              <div class="text-2xl font-bold text-yellow-400">${this.state.metrics.load.toFixed(1)}</div>
              <div class="text-xs text-gray-400">Current Load</div>
            </div>
          </div>
        </div>
      </div>
    `;
  }

  private renderSystemParameters(): string {
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
    return `
      <div class="space-y-2">
        <div class="text-xs text-gray-400 mb-2">
          Add/remove traffic to test system behavior
        </div>
        
        ${this.state.generateCalls.map(call => `
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
        `).join('')}
        
        <button id="add-generator" class="w-full btn btn-outline text-xs py-1">
          + Add
        </button>
      </div>
    `;
  }

  private renderDynamicCharts(): string {
    const charts = Object.values(this.state.dynamicCharts);
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
    // Load button
    const loadBtn = document.getElementById('load-btn');
    loadBtn?.addEventListener('click', () => this.loadContactsService());

    // Run button
    const runBtn = document.getElementById('run-btn');
    runBtn?.addEventListener('click', () => this.runSimulation());

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

  private startTrafficGeneration(call: GenerateCall) {
    // TODO: Call Canvas API to start traffic generation
    console.log(`Starting traffic generation for ${call.target} at ${call.rate} RPS`);
  }

  private stopTrafficGeneration(call: GenerateCall) {
    // TODO: Call Canvas API to stop traffic generation
    console.log(`Stopping traffic generation for ${call.target}`);
  }

  private updateTrafficRate(call: GenerateCall) {
    // TODO: Call Canvas API to update traffic rate
    console.log(`Updating traffic rate for ${call.target} to ${call.rate} RPS`);
  }

  private addNewGenerator() {
    const newId = `generator-${Date.now()}`;
    const newGenerator: GenerateCall = {
      id: newId,
      name: 'Custom Traffic',
      target: 'server.HandleCustom',
      rate: 1.0,
      enabled: false
    };
    
    this.state.generateCalls.push(newGenerator);
    this.render();
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