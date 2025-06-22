import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { CanvasService } from "./gen/sdl/v1/canvas_pb.ts";
import { create } from "@bufbuild/protobuf";
import { GeneratorSchema, MetricSchema } from "./gen/sdl/v1/models_pb.ts";
import type { Generator, Metric } from "./gen/sdl/v1/models_pb.ts";

// Create transport with the gRPC-gateway endpoint
const transport = createConnectTransport({
  baseUrl: window.location.origin,
  useBinaryFormat: false, // Use JSON for browser compatibility
});

// Create the Canvas service client
const client = createClient(CanvasService, transport);

// Default canvas ID
const DEFAULT_CANVAS_ID = "default";

export class CanvasClient {
  private canvasId: string;

  constructor(canvasId: string = DEFAULT_CANVAS_ID) {
    this.canvasId = canvasId;
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

  // Get canvas information
  async getCanvas() {
    const response = await client.getCanvas({
      id: this.canvasId
    });
    return {
      success: true,
      data: response.canvas
    };
  }

  // Get current state (canvas info)
  async getState() {
    return this.getCanvas();
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

  /*
  async updateGenerator(id: string, updates: Partial<Generator>) {
    const generator = create(GeneratorSchema, {
      ...updates,
      id: id,
      canvasId: this.canvasId
    });
    
    const response = await client.updateGenerator({
      generator: generator
    });
    return {
      success: true,
      data: response.generator
    };
  }
 */

  async deleteGenerator(id: string) {
    await client.deleteGenerator({
      canvasId: this.canvasId,
      generatorId: id
    });
    return { success: true };
  }

  async pauseGenerator(id: string) {
    await client.pauseGenerator({
      canvasId: this.canvasId,
      generatorId: id
    });
    return { success: true };
  }

  async resumeGenerator(id: string) {
    await client.resumeGenerator({
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
}

// Export a default instance
export const canvasClient = new CanvasClient();
