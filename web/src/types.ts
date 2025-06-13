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
  type: 'fileLoaded' | 'systemActivated' | 'parameterChanged' | 'simulationCompleted' | 'plotGenerated';
  [key: string]: any;
}

// Dashboard state
export interface DashboardState {
  currentFile?: string;
  currentSystem?: string;
  isConnected: boolean;
  simulationResults: Record<string, RunResult[]>;
  metrics: SystemMetrics;
}

export interface SystemMetrics {
  load: number;
  latency: number;
  successRate: number;
  serverUtilization: number;
  dbUtilization: number;
  cacheHitRate: number;
  dbConnections: string;
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