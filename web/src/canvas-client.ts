import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { CanvasService } from "./gen/sdl/v1/canvas_pb.ts";
import { create } from "@bufbuild/protobuf";
import { GeneratorSchema, MetricSchema, CanvasSchema } from "./gen/sdl/v1/models_pb.ts";
import type { Generator, Metric, Canvas } from "./gen/sdl/v1/models_pb.ts";
import { FileClient } from './types.js';

// Create transport with the Connect endpoint mounted at /api
const transport = createConnectTransport({
  baseUrl: `${window.location.origin}/api`,
  useBinaryFormat: false, // Use JSON for browser compatibility
});

// Create the Canvas service client
const client = createClient(CanvasService, transport);

// Default canvas ID
const DEFAULT_CANVAS_ID = "default";

export class CanvasClient implements FileClient {
  protected canvasId: string;

  constructor(canvasId: string = DEFAULT_CANVAS_ID) {
    this.canvasId = canvasId;
  }

  // Create a new canvas
  async createCanvas(): Promise<Canvas> {
    const canvas = create(CanvasSchema, {
      id: this.canvasId
    });
    
    const response = await client.createCanvas({
      canvas: canvas
    });
    
    if (!response.canvas) {
      throw new Error('Failed to create canvas');
    }
    
    return response.canvas;
  }

  // Ensure canvas exists (create if needed)
  async ensureCanvas(): Promise<Canvas> {
    let canvas = await this.getCanvas();
    if (!canvas) {
      console.log(`📦 Creating new canvas: ${this.canvasId}`);
      canvas = await this.createCanvas();
    }
    return canvas;
  }

  // Load a file into the canvas
  async loadFile(filePath: string) {
    const response = await client.loadFile({
      canvasId: this.canvasId,
      sdlFilePath: filePath
    });
    return {
      success: true,
      data: response
    };
  }

  // Use a specific system from loaded files
  async useSystem(systemName: string) {
    const response = await client.useSystem({
      canvasId: this.canvasId,
      systemName: systemName
    });
    return {
      success: true,
      data: response
    };
  }

  // Get canvas info
  async getCanvas(): Promise<Canvas | null> {
    try {
      const response = await client.getCanvas({
        id: this.canvasId
      });
      return response.canvas || null;
    } catch (error: any) {
      if (error.code === 'NOT_FOUND') {
        return null;
      }
      throw error;
    }
  }

  // Get current state (canvas info)
  async getState() {
    return this.getCanvas();
  }


  // FileClient interface implementation (placeholder for server mode)
  async listFiles(path: string): Promise<string[]> {
    try {
      // Fetch directory listing from server
      const response = await fetch(`/examples${path === '/' ? '' : path}/`);
      if (!response.ok) {
        throw new Error(`Failed to list files: ${response.statusText}`);
      }
      
      const html = await response.text();
      
      // Parse HTML directory listing
      const parser = new DOMParser();
      const doc = parser.parseFromString(html, 'text/html');
      const links = doc.querySelectorAll('a');
      
      const files: string[] = [];
      links.forEach(link => {
        const href = link.getAttribute('href');
        if (href && href !== '../') {
          // Convert relative paths to absolute paths
          const fullPath = path === '/' ? `/${href}` : `${path}/${href}`.replace(/\/+/g, '/');
          files.push(fullPath);
        }
      });
      
      return files;
    } catch (error) {
      console.error(`Failed to list files at ${path}:`, error);
      return [];
    }
  }

  async readFile(path: string): Promise<string> {
    try {
      const response = await fetch(`/examples${path}`);
      if (!response.ok) {
        throw new Error(`Failed to read file: ${response.statusText}`);
      }
      return await response.text();
    } catch (error) {
      console.error(`Failed to read file ${path}:`, error);
      throw error;
    }
  }

  async writeFile(path: string, content: string): Promise<void> {
    try {
      const response = await fetch(`/examples${path}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'text/plain',
        },
        body: content
      });
      
      if (!response.ok) {
        throw new Error(`Failed to write file: ${response.statusText}`);
      }
    } catch (error) {
      console.error(`Failed to write file ${path}:`, error);
      throw error;
    }
  }
  
  async deleteFile(path: string): Promise<void> {
    try {
      const response = await fetch(`/examples${path}`, {
        method: 'DELETE'
      });
      
      if (!response.ok) {
        throw new Error(`Failed to delete file: ${response.statusText}`);
      }
    } catch (error) {
      console.error(`Failed to delete file ${path}:`, error);
      throw error;
    }
  }

  // Set a parameter value
  async setParameter(path: string, value: any) {
    const response = await client.setParameter({
      canvasId: this.canvasId,
      path: path,
      newValue: value.toString()
    });
    return {
      success: response.success,
      data: response
    };
  }

  // Get parameters
  async getParameters(component?: string) {
    const response = await client.getParameters({
      canvasId: this.canvasId,
      path: component || ""
    });
    return {
      success: true,
      data: response.parameters
    };
  }

  // Generator operations
  async getGenerators() {
    const response = await client.listGenerators({
      canvasId: this.canvasId
    });
    
    // Convert array to object for compatibility
    const generatorsMap: Record<string, Generator> = {};
    response.generators.forEach((gen) => {
      generatorsMap[gen.id] = gen;
    });
    
    return {
      success: true,
      data: generatorsMap
    };
  }

  async addGenerator(name: string, component: string, method: string, rate: number) {
    const generator = create(GeneratorSchema, {
      canvasId: this.canvasId,
      name: name,
      component: component,
      method: method,
      rate: rate,
      enabled: false
    });
    
    const response = await client.addGenerator({
      generator: generator
    });
    return {
      success: true,
      data: response.generator
    };
  }

  async updateGeneratorRate(id: string, rate: number) {
    // Only update the rate field using field mask
    const generator = create(GeneratorSchema, {
      id: id,
      canvasId: this.canvasId,
      rate: rate
    });
    
    const response = await client.updateGenerator({
      generator: generator,
      updateMask: {
        paths: ["rate"]
      }
    });
    return {
      success: true,
      data: response.generator
    };
  }

  async deleteGenerator(id: string) {
    await client.deleteGenerator({
      canvasId: this.canvasId,
      generatorId: id
    });
    return { success: true };
  }

  async stopGenerator(id: string) {
    await client.stopGenerator({
      canvasId: this.canvasId,
      generatorId: id
    });
    return { success: true };
  }

  async startGenerator(id: string) {
    await client.startGenerator({
      canvasId: this.canvasId,
      generatorId: id
    });
    return { success: true };
  }

  async startAllGenerators() {
    await client.startAllGenerators({
      canvasId: this.canvasId
    });
    return { success: true };
  }

  async stopAllGenerators() {
    await client.stopAllGenerators({
      canvasId: this.canvasId
    });
    return { success: true };
  }

  // Metric operations
  async getMetrics() {
    const response = await client.listMetrics({
      canvasId: this.canvasId
    });
    
    // Convert array to object for compatibility
    const metricsMap: Record<string, Metric> = {};
    response.metrics.forEach((metric: any) => {
      metricsMap[metric.id] = metric;
    });
    
    return {
      success: true,
      data: metricsMap
    };
  }

  async addMetric(
    name: string,
    component: string,
    methods: string[],
    metricType: string,
    aggregation: string,
    aggregationWindow: number,
    matchResult?: string,
    matchResultType?: string
  ) {
    const metric = create(MetricSchema, {
      canvasId: this.canvasId,
      name: name,
      component: component,
      methods: methods,
      metricType: metricType,
      aggregation: aggregation,
      aggregationWindow: aggregationWindow,
      matchResult: matchResult,
      matchResultType: matchResultType,
      enabled: true
    });
    
    const response = await client.addMetric({
      metric: metric
    });
    return {
      success: true,
      data: response.metric
    };
  }

  async deleteMetric(id: string) {
    await client.deleteMetric({
      canvasId: this.canvasId,
      metricId: id
    });
    return { success: true };
  }

  async queryMetrics(metricId: string, startTime?: Date, endTime?: Date, limit?: number) {
    const response = await client.queryMetrics({
      canvasId: this.canvasId,
      metricId: metricId,
      startTime: startTime ? Math.floor(startTime.getTime() / 1000) : 0,
      endTime: endTime ? Math.floor(endTime.getTime() / 1000) : 0,
      limit: limit || 0
    });
    return {
      success: true,
      data: response.points
    };
  }

  // Execute trace for debugging
  async executeTrace(component: string, method: string) {
    const response = await client.executeTrace({
      canvasId: this.canvasId,
      component: component,
      method: method
    });
    return {
      success: true,
      data: response
    };
  }

  // Get system diagram
  async getSystemDiagram() {
    const response = await client.getSystemDiagram({
      canvasId: this.canvasId
    });
    return {
      success: true,
      data: response.diagram
    };
  }

  // Evaluate flows - equivalent to --apply-flows
  async evaluateFlows(strategy: string = "auto") {
    const response = await client.evaluateFlows({
      canvasId: this.canvasId,
      strategy: strategy
    });
    return {
      success: true,
      data: response
    };
  }

  // WebSocket connection for live metrics
  connectWebSocket(onMessage: (data: any) => void): WebSocket {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws`);
    
    ws.onopen = () => {
      console.log('WebSocket connected');
      // Send canvas ID to associate the connection
      ws.send(JSON.stringify({ type: 'subscribe', canvasId: this.canvasId }));
    };
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onMessage(data);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    ws.onclose = () => {
      console.log('WebSocket disconnected');
    };
    
    return ws;
  }

  // Stream real-time metric updates
  async *streamMetrics(metricIds: string[]) {
    const stream = client.streamMetrics({
      canvasId: this.canvasId,
      metricIds: metricIds
    });

    for await (const response of stream) {
      yield {
        success: true,
        data: response.updates || []
      };
    }
  }
}

// Export a default instance
export const canvasClient = new CanvasClient();
