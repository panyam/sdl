import { BasePanel, PanelConfig } from './base-panel.js';
import { AppState } from '../core/app-state-manager.js';
import { Chart, ChartConfiguration } from 'chart.js/auto';

/**
 * Panel for displaying live metrics charts
 */
export class LiveMetricsPanel extends BasePanel {
  private charts: Map<string, Chart> = new Map();
  private dynamicCharts: any = {};
  private resizeObserver: ResizeObserver | null = null;
  
  constructor(config: PanelConfig) {
    super({
      ...config,
      id: config.id || 'liveMetrics',
      title: config.title || 'Live Metrics'
    });
  }

  protected onInitialize(): void {
    // Set up resize observer for responsive charts
    this.resizeObserver = new ResizeObserver(() => {
      // Debounce chart updates on resize
      this.handleResize();
    });
  }

  protected onDispose(): void {
    // Dispose all charts
    this.charts.forEach(chart => chart.destroy());
    this.charts.clear();
    
    // Clean up resize observer
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
      this.resizeObserver = null;
    }
  }

  onStateChange(state: AppState, changedKeys: string[]): void {
    // Update charts if dynamic charts changed
    if (changedKeys.includes('dynamicCharts')) {
      this.dynamicCharts = state.dynamicCharts || {};
      this.render();
    }
    
    // Update metrics data
    if (changedKeys.includes('metrics') || changedKeys.includes('simulationResults')) {
      this.updateChartsData(state);
    }
  }

  protected render(): void {
    const content = this.renderDynamicCharts();
    this.setContent(content);
    
    // Initialize charts after render
    setTimeout(() => this.initializeCharts(), 10);
  }

  private renderDynamicCharts(): string {
    const chartIds = Object.keys(this.dynamicCharts);
    
    if (chartIds.length === 0) {
      return `
        <div class="h-full flex items-center justify-center">
          <div class="text-center text-gray-400">
            <div class="text-6xl mb-4">📊</div>
            <div class="text-lg">No Active Metrics</div>
            <div class="text-sm">Start a simulation to see live metrics</div>
          </div>
        </div>
      `;
    }

    return `
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-4" style="grid-auto-rows: 200px;">
        ${chartIds.map(chartId => this.renderChartContainer(chartId)).join('')}
      </div>
    `;
  }

  private renderChartContainer(chartId: string): string {
    const chartData = this.dynamicCharts[chartId];
    const title = chartData.chartName || chartId;
    
    return `
      <div class="bg-gray-800 rounded-lg p-4 flex flex-col">
        <h4 class="text-sm font-medium text-gray-300 mb-2">${title}</h4>
        <div class="flex-1 relative">
          <canvas id="chart-${chartId}" class="absolute inset-0 w-full h-full"></canvas>
        </div>
      </div>
    `;
  }

  private initializeCharts(): void {
    if (!this.container) return;

    // Clear existing charts
    this.charts.forEach(chart => chart.destroy());
    this.charts.clear();

    // Create new charts
    Object.keys(this.dynamicCharts).forEach(chartId => {
      const canvas = this.container?.querySelector(`#chart-${chartId}`) as HTMLCanvasElement;
      if (canvas) {
        const chart = this.createChart(canvas, chartId);
        if (chart) {
          this.charts.set(chartId, chart);
          
          // Observe canvas for resize
          if (this.resizeObserver) {
            this.resizeObserver.observe(canvas);
          }
        }
      }
    });
  }

  private createChart(canvas: HTMLCanvasElement, chartId: string): Chart | null {
    const chartData = this.dynamicCharts[chartId];
    if (!chartData) return null;

    const config: ChartConfiguration = {
      type: 'line',
      data: {
        labels: chartData.labels || [],
        datasets: [{
          label: chartData.metricName || 'Value',
          data: chartData.data || [],
          borderColor: this.getChartColor(chartId),
          backgroundColor: this.getChartColor(chartId, 0.1),
          borderWidth: 2,
          tension: 0.1,
          pointRadius: 0,
          pointHoverRadius: 3
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          mode: 'index',
          intersect: false,
        },
        plugins: {
          legend: {
            display: false
          },
          tooltip: {
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
            titleColor: '#fff',
            bodyColor: '#fff',
            borderColor: '#333',
            borderWidth: 1
          }
        },
        scales: {
          x: {
            display: true,
            grid: {
              color: 'rgba(255, 255, 255, 0.1)',
            },
            ticks: {
              color: '#9CA3AF',
              autoSkip: true,
              maxTicksLimit: 5
            }
          },
          y: {
            display: true,
            grid: {
              color: 'rgba(255, 255, 255, 0.1)',
            },
            ticks: {
              color: '#9CA3AF',
              callback: function(value) {
                return Number(value).toFixed(1);
              }
            }
          }
        }
      }
    };

    try {
      return new Chart(canvas, config);
    } catch (error) {
      console.error(`Failed to create chart ${chartId}:`, error);
      return null;
    }
  }

  private getChartColor(chartId: string, alpha: number = 1): string {
    // Generate consistent colors based on chart ID
    const colors = [
      `rgba(59, 130, 246, ${alpha})`, // Blue
      `rgba(16, 185, 129, ${alpha})`, // Green
      `rgba(251, 146, 60, ${alpha})`, // Orange
      `rgba(147, 51, 234, ${alpha})`, // Purple
      `rgba(236, 72, 153, ${alpha})`, // Pink
      `rgba(250, 204, 21, ${alpha})`, // Yellow
      `rgba(14, 165, 233, ${alpha})`, // Sky
      `rgba(34, 197, 94, ${alpha})`, // Emerald
    ];
    
    const index = chartId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) % colors.length;
    return colors[index];
  }

  private updateChartsData(_state: AppState): void {
    // Update each chart with new data
    this.charts.forEach((chart, chartId) => {
      const chartData = this.dynamicCharts[chartId];
      if (chartData && chart.data.datasets[0]) {
        chart.data.labels = chartData.labels || [];
        chart.data.datasets[0].data = chartData.data || [];
        chart.update('none'); // Update without animation for performance
      }
    });
  }

  private handleResize = (() => {
    let resizeTimer: number | null = null;
    
    return () => {
      if (resizeTimer) {
        clearTimeout(resizeTimer);
      }
      
      resizeTimer = window.setTimeout(() => {
        // Update all charts
        this.charts.forEach(chart => {
          chart.resize();
        });
      }, 250);
    };
  })();

  onResize(): void {
    this.handleResize();
  }
}