import { BasePanel, PanelConfig } from './base-panel.js';
import { AppState } from '../core/app-state-manager.js';
import { AppEvents, EventHandler } from '../core/event-bus.js';
import { Graphviz } from "@hpcc-js/wasm";
import { SystemDiagram } from '../gen/sdl/v1/canvas_pb.js';

export interface SystemArchitecturePanelConfig extends PanelConfig {
  graphviz?: any; // Graphviz instance (optional, will load if not provided)
}

/**
 * Panel for displaying system architecture diagrams
 */
export class SystemArchitecturePanel extends BasePanel {
  private graphviz: any = null;
  private systemDiagram: SystemDiagram | null = null;
  private layoutTopToBottom = false;
  private diagramZoom = 1.0;
  private diagramLoadedHandler?: EventHandler;

  constructor(config: SystemArchitecturePanelConfig) {
    super({
      ...config,
      id: config.id || 'systemArchitecture',
      title: config.title || 'System Architecture'
    });
    
    this.graphviz = config.graphviz;
  }

  protected async onInitialize(): Promise<void> {
    // Load Graphviz if not provided
    if (!this.graphviz) {
      try {
        this.graphviz = await Graphviz.load();
        console.log('✅ Graphviz WASM loaded for SystemArchitecturePanel');
      } catch (error) {
        console.error('❌ Failed to load Graphviz WASM:', error);
      }
    }

    // Listen for layout toggle events
    this.eventBus.on(AppEvents.TOOLBAR_ACTION, this.handleToolbarAction);
    
    // Listen for system diagram loaded events
    this.diagramLoadedHandler = (diagram: any) => {
      this.systemDiagram = diagram;
      this.render();
    };
    this.eventBus.on('system:diagram:loaded', this.diagramLoadedHandler);
  }

  protected onDispose(): void {
    this.eventBus.off(AppEvents.TOOLBAR_ACTION, this.handleToolbarAction);
    if (this.diagramLoadedHandler) {
      this.eventBus.off('system:diagram:loaded', this.diagramLoadedHandler);
    }
  }

  onStateChange(state: AppState, changedKeys: string[]): void {
    // Check if system diagram needs update
    if (changedKeys.includes('simulationResults') || changedKeys.includes('currentSystem')) {
      const systemDiagram = state.simulationResults?.systemDiagram;
      if (systemDiagram) {
        this.systemDiagram = systemDiagram as any;
        this.render();
      } else if (!systemDiagram && this.systemDiagram) {
        this.systemDiagram = null;
        this.render();
      }
    }
  }

  protected render(): void {
    if (!this.systemDiagram) {
      this.showEmpty('No System Loaded', 'Load an SDL file to view system architecture');
      return;
    }

    this.setContent(`
      <div class="w-full h-full flex flex-col bg-gray-50 dark:bg-gray-900 p-4">
        <div class="text-center mb-4">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-300">${this.systemDiagram.systemName}</h3>
        </div>
        
        <!-- SVG System Architecture -->
        <div id="architecture-svg-container" class="flex-1 overflow-auto bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
          ${this.renderSystemDiagramSVG()}
        </div>
      </div>
    `);

    // Setup interactions after render
    setTimeout(() => this.setupDiagramInteractions(), 10);
  }

  private renderSystemDiagramSVG(): string {
    if (!this.systemDiagram) return '';

    const dotContent = this.generateDotFile();
    const svgContent = this.convertDotToSVG(dotContent);

    return `
      <div id="svg-zoom-wrapper" class="h-full overflow-auto cursor-grab active:cursor-grabbing">
        <div id="svg-transform-wrapper" style="transform: scale(${this.diagramZoom}); transform-origin: top left; transition: transform 0.1s ease;">
          ${svgContent}
        </div>
      </div>
      <div class="absolute bottom-2 right-2 text-xs text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
        Zoom: ${Math.round(this.diagramZoom * 100)}% | Use Ctrl+Scroll to zoom | Double-click to reset
      </div>
    `;
  }

  private generateDotFile(): string {
    if (!this.systemDiagram) return '';

    const systemName = this.systemDiagram.systemName || 'System';
    let dotContent = `digraph "${systemName}" {\n`;
    dotContent += `  rankdir=${this.layoutTopToBottom ? "TB" : "LR"};\n`;
    dotContent += `  bgcolor="#1a1a1a";\n`;
    dotContent += `  node [fontname="Monaco,Menlo,monospace" fontcolor="white" style=filled];\n`;
    dotContent += `  edge [color="#9ca3af" arrowhead="normal" penwidth=2];\n`;
    dotContent += `  graph [ranksep=1.0 nodesep=0.8 pad=0.5];\n`;
    dotContent += `  compound=true;\n\n`;

    // Group nodes by component
    const componentGroups = new Map<string, any[]>();
    
    this.systemDiagram?.nodes?.forEach((node) => {
      const [componentName, methodName] = node.id.split(':');
      if (!componentGroups.has(componentName)) {
        componentGroups.set(componentName, []);
      }
      componentGroups.get(componentName)!.push({...node, methodName});
    });

    // Generate subgraph clusters for each component
    let clusterIndex = 0;
    componentGroups.forEach((methods, componentName) => {
      const hasInternalMethods = methods.some(m => 
        m.type.includes('(internal)') || 
        m.id.includes('.pool:') || 
        m.id.includes('.driverTable:')
      );
      
      const componentIcon = this.getIconForNode(methods[0]);
      
      dotContent += `  subgraph cluster_${clusterIndex} {\n`;
      dotContent += `    label="${componentIcon} ${componentName}";\n`;
      dotContent += `    style="filled,rounded";\n`;
      dotContent += `    fillcolor="${hasInternalMethods ? '#1e1b4b' : '#111827'}";\n`;
      dotContent += `    fontcolor="#e5e7eb";\n`;
      dotContent += `    fontsize=14;\n`;
      dotContent += `    margin=12;\n`;
      dotContent += `    penwidth=2;\n`;
      dotContent += `    color="${hasInternalMethods ? '#4c1d95' : '#374151'}";\n\n`;
      
      methods.forEach((node) => {
        const nodeId = node.id.replace(':', '_');
        const methodLabel = `${node.methodName}\\n${node.traffic}`;
        
        const isInternal = node.type.includes('(internal)') || 
                          node.id.includes('.pool:') || 
                          node.id.includes('.driverTable:');
        
        dotContent += `    "${nodeId}" [label="${methodLabel}"`;
        dotContent += ` shape=box style="filled,rounded"`;
        dotContent += ` fillcolor="${isInternal ? '#4c1d95' : '#1f2937'}"`;
        dotContent += ` fontcolor="${isInternal ? '#e9d5ff' : '#a3e635'}"`;
        dotContent += ` fontsize=${isInternal ? '10' : '11'}`;
        dotContent += ` margin=0.1 penwidth=1];\n`;
      });
      
      dotContent += `  }\n\n`;
      clusterIndex++;
    });

    // Generate edges
    this.systemDiagram?.edges?.forEach(edge => {
      const fromNodeId = edge.fromId.replace(':', '_');
      const toNodeId = edge.toId.replace(':', '_');
      const edgeColor = edge.color || "#9ca3af";
      let edgeStyle = ` color="${edgeColor}" fontcolor="${edgeColor}" fontsize=10`;
      const label = edge.label ? ` label="${edge.label}"` : '';
      
      dotContent += `  "${fromNodeId}" -> "${toNodeId}"[${label}${edgeStyle}];\n`;
    });

    dotContent += `}\n`;
    return dotContent;
  }

  private convertDotToSVG(dotContent: string): string {
    try {
      if (!this.graphviz) {
        console.warn('Graphviz not loaded yet');
        return this.generateFallbackSVG();
      }

      const svg = this.graphviz.dot(dotContent);
      
      // Clean up the SVG
      let cleanedSvg = svg
        .replace(/width="[^"]*"/, 'width="100%"')
        .replace(/height="[^"]*"/, 'height="100%"')
        .replace(/<title>.*?<\/title>/g, '');

      return cleanedSvg;
    } catch (error) {
      console.error('Error converting DOT to SVG:', error);
      return this.generateFallbackSVG();
    }
  }

  private generateFallbackSVG(): string {
    return `
      <div class="flex items-center justify-center h-full text-gray-400">
        <div class="text-center">
          <div class="text-6xl mb-4">📊</div>
          <div class="text-lg">System Visualization</div>
          <div class="text-sm mt-2">Dot rendering unavailable</div>
        </div>
      </div>
    `;
  }

  private getIconForNode(node: any): string {
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
    
    const type = node.type?.toLowerCase() || '';
    if (type.includes('cache')) return '💾';
    if (type.includes('database') || type.includes('db')) return '🗄️';
    if (type.includes('gateway')) return '🚪';
    if (type.includes('service')) return '⚙️';
    if (type.includes('queue')) return '📋';
    if (type.includes('pool')) return '🏊';
    if (type.includes('api')) return '🔌';
    
    return '📦';
  }

  private setupDiagramInteractions(): void {
    const container = this.container?.querySelector('#architecture-svg-container');
    if (!container) return;

    const wrapper = container.querySelector('#svg-zoom-wrapper') as HTMLElement;
    const transformWrapper = container.querySelector('#svg-transform-wrapper') as HTMLElement;
    
    if (!wrapper || !transformWrapper) return;

    // Mouse wheel zoom
    wrapper.addEventListener('wheel', (e: WheelEvent) => {
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault();
        
        const scrollLeftPercent = wrapper.scrollLeft / (wrapper.scrollWidth - wrapper.clientWidth);
        const scrollTopPercent = wrapper.scrollTop / (wrapper.scrollHeight - wrapper.clientHeight);
        
        const delta = e.deltaY > 0 ? 0.9 : 1.1;
        const newZoom = Math.max(0.1, Math.min(5, this.diagramZoom * delta));
        
        this.diagramZoom = newZoom;
        this.updateDiagramTransform();
        
        setTimeout(() => {
          wrapper.scrollLeft = scrollLeftPercent * (wrapper.scrollWidth - wrapper.clientWidth);
          wrapper.scrollTop = scrollTopPercent * (wrapper.scrollHeight - wrapper.clientHeight);
        }, 10);
      }
    });

    // Reset zoom on double click
    wrapper.addEventListener('dblclick', () => {
      this.diagramZoom = 1.0;
      this.updateDiagramTransform();
    });
  }

  private updateDiagramTransform(): void {
    const transformWrapper = this.container?.querySelector('#svg-transform-wrapper') as HTMLElement;
    const zoomIndicator = this.container?.querySelector('#architecture-svg-container > div:last-child') as HTMLElement;
    
    if (transformWrapper) {
      transformWrapper.style.transform = `scale(${this.diagramZoom})`;
    }
    
    if (zoomIndicator) {
      zoomIndicator.innerHTML = `Zoom: ${Math.round(this.diagramZoom * 100)}% | Use Ctrl+Scroll to zoom | Double-click to reset`;
    }
  }

  private handleToolbarAction = (data: any) => {
    if (data.action === 'toggleLayout') {
      this.layoutTopToBottom = !this.layoutTopToBottom;
      this.render();
    }
  };

  onResize(): void {
    // Re-render to adjust SVG if needed
    if (this.systemDiagram) {
      this.render();
    }
  }
}