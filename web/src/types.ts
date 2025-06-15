// API Types matching Go backend

export interface RunResult {
  ts: number;        // Unix timestamp in milliseconds
  latency: number;   // Latency in milliseconds  
  result: string;    // String representation of result
  is_error: boolean; // Whether this run resulted in an error
  error?: string;    // Error message if is_error is true
}

export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface LoadRequest {
  filePath: string;
}

export interface UseRequest {
  systemName: string;
}

export interface SetRequest {
  path: string;
  value: any;
}

export interface RunRequest {
  varName: string;
  target: string;
  runs: number;
}

export interface PlotRequest {
  series: SeriesInfo[];
  outputFile: string;
  title: string;
}

export interface SeriesInfo {
  name: string;
  from: string;
}

// WebSocket message types
export interface WebSocketMessage {
  type: 'fileLoaded' | 'systemActivated' | 'parameterChanged' | 'simulationCompleted' | 'plotGenerated' |
        'generatorAdded' | 'generatorUpdated' | 'generatorRemoved' | 'generatorPaused' | 'generatorResumed' |
        'generatorsStarted' | 'generatorsStopped' | 'measurementAdded' | 'measurementUpdated' | 
        'measurementRemoved' | 'stateRestored' | 'ping' | 'pong' | 'connected';
  [key: string]: any;
}

// Dashboard state
export interface DashboardState {
  currentFile?: string;
  currentSystem?: string;
  isConnected: boolean;
  simulationResults: Record<string, RunResult[]>;
  metrics: SystemMetrics;
  dynamicCharts: Record<string, ChartData>;
  generateCalls: GenerateCall[];
}

// Dynamic chart support
export interface ChartData {
  chartName: string;
  metricName: string;
  target?: string;
  data: number[];
  labels: string[];
  title: string;
}

// Generate call controls
export interface GenerateCall {
  id: string;
  name: string;
  target: string;
  rate: number;
  enabled: boolean;
}

export interface SystemMetrics {
  load: number;
  latency: number;
  successRate: number;
  [key: string]: any; // Allow dynamic metrics
}

// Component parameter types
export interface ParameterConfig {
  name: string;
  path: string;
  type: 'number' | 'boolean' | 'string';
  min?: number;
  max?: number;
  step?: number;
  value: any;
}

// Canvas API types matching Go backend
export interface GeneratorConfig {
  id: string;
  name: string;
  target: string;
  rate: number;
  duration?: number; // in milliseconds
  enabled: boolean;
  options?: Record<string, any>;
}

export interface MeasurementConfig {
  id: string;
  name: string;
  metricType: string; // "latency", "throughput", "errors", etc.
  target: string;
  interval: number; // in milliseconds
  enabled: boolean;
  options?: Record<string, any>;
}

export interface MeasurementDataPoint {
  timestamp: number; // Timestamp in milliseconds
  value: number;
  target: string;
  run_id: string;
}

export interface CanvasState {
  loadedFiles: string[];
  activeFile: string;
  activeSystem: string;
  generators: Record<string, GeneratorConfig>;
  measurements: Record<string, MeasurementConfig>;
  sessionVars: Record<string, any>;
  lastRunResult?: any;
  systemParameters?: Record<string, any>;
  metricsHistory?: MetricSnapshot[];
}

export interface MetricSnapshot {
  timestamp: number;
  metricType: string;
  value: number;
  source: string;
}

// System diagram types (matching viz package)
export interface SystemDiagram {
  systemName: string;
  nodes: SystemNode[];
  edges: SystemEdge[];
}

export interface SystemNode {
  ID: string;
  Name: string;
  Type: string;
}

export interface SystemEdge {
  FromID: string;
  ToID: string;
  Label: string;
}