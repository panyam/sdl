import { CanvasAPI } from './canvas-api.js';
import { DashboardState, ParameterConfig, WebSocketMessage } from './types.js';
import { Chart, ChartConfiguration } from 'chart.js/auto';

export class Dashboard {
  private api: CanvasAPI;
  private state: DashboardState;
  private latencyChart: Chart | null = null;

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
      }
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

    this.updateChart();
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

  private updateChart() {
    if (!this.latencyChart) return;

    // Add new data point
    const now = Date.now();
    const data = this.latencyChart.data;
    
    if (data.datasets[0].data.length > 20) {
      data.labels?.shift();
      data.datasets[0].data.shift();
    }
    
    data.labels?.push(new Date(now).toLocaleTimeString());
    data.datasets[0].data.push(this.state.metrics.latency);
    
    this.latencyChart.update('none');
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
      <div class="min-h-screen p-4">
        <!-- Header -->
        <div class="mb-6">
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

        <!-- Main Grid Layout -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
          <!-- System Architecture Panel -->
          <div class="panel">
            <div class="panel-header">System Architecture</div>
            <div class="space-y-3">
              <div class="component-box">
                <div class="component-title">ContactAppServer</div>
                <div class="component-detail">Pool: ${this.parameters.find(p => p.path === 'server.pool.Size')?.value}/10</div>
                <div class="component-detail">Load: ${this.state.metrics.load.toFixed(1)} RPS</div>
              </div>
              <div class="text-center text-gray-500">↓</div>
              <div class="component-box">
                <div class="component-title">ContactDatabase</div>
                <div class="component-detail">Pool: ${this.parameters.find(p => p.path === 'server.db.pool.Size')?.value}/5</div>
                <div class="component-detail">Cache: ${(this.state.metrics.cacheHitRate * 100).toFixed(0)}%</div>
              </div>
              <div class="text-center text-gray-500">↓</div>
              <div class="component-box">
                <div class="component-title">HashIndex</div>
                <div class="component-detail">Lookup: ~8ms</div>
              </div>
            </div>
          </div>

          <!-- Current Metrics Panel -->
          <div class="panel">
            <div class="panel-header">Current Metrics</div>
            <div class="space-y-2">
              <div class="metric-item">
                <span class="metric-label">Load</span>
                <span class="metric-value">${this.state.metrics.load.toFixed(1)} RPS</span>
              </div>
              <div class="metric-item">
                <span class="metric-label">P95 Latency</span>
                <span class="metric-value ${this.getLatencyClass()}">${this.state.metrics.latency.toFixed(1)}ms</span>
              </div>
              <div class="metric-item">
                <span class="metric-label">Success Rate</span>
                <span class="metric-value ${this.getSuccessRateClass()}">${this.state.metrics.successRate.toFixed(1)}%</span>
              </div>
              <div class="metric-item">
                <span class="metric-label">Server Util</span>
                <span class="metric-value">${this.state.metrics.serverUtilization.toFixed(0)}%</span>
              </div>
              <div class="metric-item">
                <span class="metric-label">Cache Hit</span>
                <span class="metric-value">${(this.state.metrics.cacheHitRate * 100).toFixed(0)}%</span>
              </div>
            </div>
          </div>

          <!-- Parameter Controls Panel -->
          <div class="panel">
            <div class="panel-header">Parameter Controls</div>
            <div class="space-y-3">
              ${this.parameters.map(param => this.renderParameterControl(param)).join('')}
            </div>
          </div>
        </div>

        <!-- Performance Chart Panel -->
        <div class="panel">
          <div class="panel-header">Live Performance Chart</div>
          <div class="h-64">
            <canvas id="latency-chart"></canvas>
          </div>
        </div>
      </div>
    `;

    this.setupInteractivity();
    this.initChart();
  }

  private renderParameterControl(param: ParameterConfig): string {
    if (param.type === 'number') {
      return `
        <div class="space-y-1">
          <label class="block text-sm text-gray-300">${param.name}</label>
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
            <span class="text-sm text-gray-400 w-16">${param.value}</span>
          </div>
        </div>
      `;
    }
    return '';
  }

  private getLatencyClass(): string {
    if (this.state.metrics.latency < 20) return 'metric-good';
    if (this.state.metrics.latency < 50) return 'metric-warning';
    return 'metric-error';
  }

  private getSuccessRateClass(): string {
    if (this.state.metrics.successRate > 95) return 'metric-good';
    if (this.state.metrics.successRate > 80) return 'metric-warning';
    return 'metric-error';
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
  }

  private initChart() {
    const canvas = document.getElementById('latency-chart') as HTMLCanvasElement;
    if (!canvas) return;

    const config: ChartConfiguration = {
      type: 'line',
      data: {
        labels: [],
        datasets: [{
          label: 'P95 Latency (ms)',
          data: [],
          borderColor: 'rgb(59, 130, 246)',
          backgroundColor: 'rgba(59, 130, 246, 0.1)',
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
            ticks: { color: 'rgba(156, 163, 175, 1)' }
          },
          x: {
            grid: { color: 'rgba(55, 65, 81, 0.5)' },
            ticks: { color: 'rgba(156, 163, 175, 1)' }
          }
        },
        plugins: {
          legend: {
            labels: { color: 'rgba(156, 163, 175, 1)' }
          }
        }
      }
    };

    this.latencyChart = new Chart(canvas, config);
  }
}