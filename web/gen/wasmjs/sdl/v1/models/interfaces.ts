// Generated TypeScript interfaces from proto file
// DO NOT EDIT - This file is auto-generated

import { Timestamp } from "@bufbuild/protobuf/wkt";




export interface Pagination {
  /** Instead of an offset an abstract  "page" key is provided that offers
 an opaque "pointer" into some offset in a result set. */
  pageKey: string;
  /** If a pagekey is not supported we can also support a direct integer offset
 for cases where it makes sense. */
  pageOffset: number;
  /** Number of results to return. */
  pageSize: number;
}



export interface PaginationResponse {
  /** The key/pointer string that subsequent List requests should pass to
 continue the pagination. */
  nextPageKey: string;
  /** Also support an integer offset if possible */
  nextPageOffset: number;
  /** Whether theere are more results. */
  hasMore: boolean;
  /** Total number of results. */
  totalResults: number;
}


/**
 * Workspace is a project container holding multiple designs (system architectures).
 Each design is backed by a Canvas for runtime simulation.
 The workspace also declares import sources for module resolution.
 */
export interface Workspace {
  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  id: string;
  name: string;
  description: string;
  /** Import sources for module resolution (name -> source config)
 e.g., "stdlib" -> { builtin: true }, "patterns" -> { github: "panyam/sdl-patterns" } */
  sources: Record<string, ImportSource>;
  /** Designs within this workspace (each is a system architecture) */
  designs?: WorkspaceDesign[];
  /** Currently active design name */
  activeDesign: string;
  /** Directory path (for file-based workspaces) */
  dir: string;
  /** Workspace-level metadata (for UI display) */
  tags: string[];
  difficulty: string;
  category: string;
}


/**
 * WorkspaceDesign represents one system architecture within a workspace.
 Maps to a `system` block in an SDL file, backed by a Canvas for runtime.
 */
export interface WorkspaceDesign {
  /** System name as declared in SDL (e.g., "UberMVP") */
  name: string;
  /** SDL file path relative to workspace root */
  file: string;
  /** Parent workspace ID */
  workspaceId: string;
  /** Brief description */
  description: string;
  /** Per-design metadata (for UI display) */
  tags: string[];
  difficulty: string;
  category: string;
}


/**
 * ImportSource declares where to resolve @name/ prefixed imports.
 The runtime configures the appropriate resolver based on context
 (CLI: disk/HTTP, WASM: memory/fetch, server: bundled/HTTP).
 */
export interface ImportSource {
  /** Built-in source (e.g., @stdlib/ — always available) */
  builtin: boolean;
  /** GitHub repository (e.g., "panyam/sdl-patterns") */
  github: string;
  /** HTTP(S) URL base */
  url: string;
  /** Subdirectory within the source */
  path: string;
  /** Version/branch/tag for pinning */
  ref: string;
}



export interface File {
  /** Path relative to the canvas root */
  path: string;
  /** Contents of the file */
  contents: string;
}



export interface Generator {
  /** Unique name within a system (e.g., "baseline", "health") */
  name: string;
  /** Target component path (e.g., "arch.webserver") */
  component: string;
  /** Target method name (e.g., "HandleRequest") */
  method: string;
  /** Traffic rate in RPS */
  rate: number;
  /** Duration in seconds (0 = forever) */
  duration: number;
  /** Whether the generator is active */
  enabled: boolean;
}



export interface Metric {
  /** Unique name within a system (e.g., "request_latency") */
  name: string;
  /** Target component path (e.g., "arch.webserver") */
  component: string;
  /** Target method names */
  methods: string[];
  /** Whether the metric is active */
  enabled: boolean;
  /** Type: "count", "latency", "utilization" */
  metricType: string;
  /** Aggregation function: "sum", "avg", "min", "max", "p50", "p90", "p95", "p99" */
  aggregation: string;
  /** Aggregation window in seconds */
  aggregationWindow: number;
  /** Result value to match (optional filter) */
  matchResult: string;
  /** Result type for match parsing */
  matchResultType: string;
  /** Statistics (populated by metric store) */
  oldestTimestamp: number;
  newestTimestamp: number;
  numDataPoints: number;
}



export interface MetricPoint {
  timestamp: number;
  value: number;
}


/**
 * Individual metric update
 */
export interface MetricUpdate {
  metricId: string;
  point?: MetricPoint;
}


/**
 * SystemDiagram represents the topology of a system
 */
export interface SystemDiagram {
  systemName: string;
  nodes?: DiagramNode[];
  edges?: DiagramEdge[];
}


/**
 * DiagramNode represents a component or instance in the system
 */
export interface DiagramNode {
  id: string;
  name: string;
  type: string;
  methods?: MethodInfo[];
  traffic: string;
  fullPath: string;
  icon: string;
}


/**
 * MethodInfo represents information about a component method
 */
export interface MethodInfo {
  name: string;
  returnType: string;
  traffic: number;
}


/**
 * DiagramEdge represents a connection between nodes
 */
export interface DiagramEdge {
  fromId: string;
  toId: string;
  fromMethod: string;
  toMethod: string;
  label: string;
  order: number;
  condition: string;
  probability: number;
  generatorId: string;
  color: string;
}


/**
 * Resource utilization information
 */
export interface UtilizationInfo {
  resourceName: string;
  componentPath: string;
  utilization: number;
  capacity: number;
  currentLoad: number;
  isBottleneck: boolean;
  warningThreshold: number;
  criticalThreshold: number;
}


/**
 * Represents a flow edge between components
 */
export interface FlowEdge {
  fromComponent: string;
  fromMethod: string;
  toComponent: string;
  toMethod: string;
  rate: number;
  condition: string;
}


/**
 * Current flow state
 */
export interface FlowState {
  strategy: string;
  rates: Record<string, number>;
  manualOverrides: Record<string, number>;
}


/**
 * TraceData matches the runtime.TraceData structure
 */
export interface TraceData {
  system: string;
  entryPoint: string;
  events?: TraceEvent[];
}


/**
 * TraceEvent matches the runtime.TraceEvent structure
 */
export interface TraceEvent {
  kind: string;
  id: number;
  parentId: number;
  timestamp: number;
  duration: number;
  component: string;
  method: string;
  args: string[];
  returnValue: string;
  errorMessage: string;
}


/**
 * Enhanced TraceData for all-paths traversal - represents the complete execution tree
 */
export interface AllPathsTraceData {
  traceId: string;
  /** The root TraceNode always starts from the <Component>.<Method> where we are kicking off the trace from */
  root?: TraceNode;
}


/**
 * TraceNode represents a single node in the execution tree
 */
export interface TraceNode {
  /** Name of the component and method in the form <Component>.<Method> we are starting the trace from */
  startingTarget: string;
  /** All edges in an ordered fashion */
  edges?: Edge[];
  /** Multiple groups for flexible labeling of sub-trees (loops, conditionals, etc.) */
  groups?: GroupInfo[];
}


/**
 * Edge represents a transition from one node to another in the execution tree
 */
export interface Edge {
  /** Unique Edge ID across the entire Trace */
  id: string;
  /** The next node this edge leads to */
  nextNode?: TraceNode;
  /** Label on the edge (if any) */
  label: string;
  /** Async edges denote Futures being sent without a return */
  isAsync: boolean;
  /** "Reverse" edges show a "wait" on a future */
  isReverse: boolean;
  /** This is optional but leaving it here just in case. */
  probability: string;
  /** Condition information for branching */
  condition: string;
  /** true if this edge represents a conditional branch */
  isConditional: boolean;
}


/**
 * GroupInfo allows flexible grouping of edges with labels
 */
export interface GroupInfo {
  /** Starting edge index */
  groupStart: number;
  /** Ending edge index (inclusive) */
  groupEnd: number;
  /** Generic label: "loop: 3x", "if cached", "switch: status" */
  groupLabel: string;
  /** Optional hint: "loop", "conditional", "switch" (for tooling) */
  groupType: string;
}


/**
 * Single parameter update
 */
export interface ParameterUpdate {
  path: string;
  newValue: string;
}


/**
 * Result for individual parameter update
 */
export interface ParameterUpdateResult {
  path: string;
  success: boolean;
  errorMessage: string;
  oldValue: string;
  newValue: string;
}



export interface AggregateResult {
  timestamp: number;
  value: number;
}



export interface LoadFileRequest {
  workspaceId: string;
  sdlFilePath: string;
}



export interface LoadFileResponse {
}



export interface UseSystemRequest {
  workspaceId: string;
  systemName: string;
}



export interface UseSystemResponse {
}



export interface AddGeneratorRequest {
  workspaceId: string;
  generator?: Generator;
  applyFlows: boolean;
}



export interface AddGeneratorResponse {
  generator?: Generator;
}



export interface ListGeneratorsRequest {
  workspaceId: string;
}



export interface ListGeneratorsResponse {
  generators?: Generator[];
}



export interface GetGeneratorRequest {
  workspaceId: string;
  generatorName: string;
}



export interface GetGeneratorResponse {
  generator?: Generator;
}



export interface UpdateGeneratorRequest {
  workspaceId: string;
  generator?: Generator;
  applyFlows: boolean;
}



export interface UpdateGeneratorResponse {
  generator?: Generator;
}



export interface StartGeneratorRequest {
  workspaceId: string;
  generatorName: string;
}



export interface StartGeneratorResponse {
}



export interface StopGeneratorRequest {
  workspaceId: string;
  generatorName: string;
}



export interface StopGeneratorResponse {
}



export interface DeleteGeneratorRequest {
  workspaceId: string;
  generatorName: string;
  applyFlows: boolean;
}



export interface DeleteGeneratorResponse {
}



export interface StartAllGeneratorsRequest {
  workspaceId: string;
}



export interface StartAllGeneratorsResponse {
  totalGenerators: number;
  startedCount: number;
  alreadyRunningCount: number;
  failedCount: number;
  failedIds: string[];
}



export interface StopAllGeneratorsRequest {
  workspaceId: string;
}



export interface StopAllGeneratorsResponse {
  totalGenerators: number;
  stoppedCount: number;
  alreadyStoppedCount: number;
  failedCount: number;
  failedIds: string[];
}



export interface AddMetricRequest {
  workspaceId: string;
  metric?: Metric;
}



export interface AddMetricResponse {
  metric?: Metric;
}



export interface DeleteMetricRequest {
  workspaceId: string;
  metricName: string;
}



export interface DeleteMetricResponse {
}



export interface ListMetricsRequest {
  workspaceId: string;
}



export interface ListMetricsResponse {
  metrics?: Metric[];
}



export interface QueryMetricsRequest {
  workspaceId: string;
  metricName: string;
  startTime: number;
  endTime: number;
  limit: number;
}



export interface QueryMetricsResponse {
  points?: MetricPoint[];
}



export interface AggregateMetricsRequest {
  workspaceId: string;
  metricName: string;
  startTime: number;
  endTime: number;
  function: string;
  windowSize: number;
}



export interface AggregateMetricsResponse {
  results?: AggregateResult[];
}



export interface StreamMetricsRequest {
  workspaceId: string;
  metricNames: string[];
}



export interface StreamMetricsResponse {
  updates?: MetricUpdate[];
}



export interface ExecuteTraceRequest {
  workspaceId: string;
  component: string;
  method: string;
}



export interface ExecuteTraceResponse {
  traceData?: TraceData;
}



export interface TraceAllPathsRequest {
  workspaceId: string;
  component: string;
  method: string;
  maxDepth: number;
}



export interface TraceAllPathsResponse {
  traceData?: AllPathsTraceData;
}



export interface SetParameterRequest {
  workspaceId: string;
  path: string;
  newValue: string;
}



export interface SetParameterResponse {
  success: boolean;
  errorMessage: string;
  newValue: string;
  oldValue: string;
}



export interface GetParametersRequest {
  workspaceId: string;
  path: string;
}



export interface GetParametersResponse {
  parameters: Record<string, string>;
}



export interface BatchSetParametersRequest {
  workspaceId: string;
  updates?: ParameterUpdate[];
}



export interface BatchSetParametersResponse {
  success: boolean;
  errorMessage: string;
  results?: ParameterUpdateResult[];
}



export interface EvaluateFlowsRequest {
  workspaceId: string;
  strategy: string;
}



export interface EvaluateFlowsResponse {
  strategy: string;
  status: string;
  iterations: number;
  warnings: string[];
  componentRates: Record<string, number>;
  flowEdges?: FlowEdge[];
}



export interface GetFlowStateRequest {
  workspaceId: string;
}



export interface GetFlowStateResponse {
  state?: FlowState;
}



export interface GetSystemDiagramRequest {
  workspaceId: string;
}



export interface GetSystemDiagramResponse {
  diagram?: SystemDiagram;
}



export interface GetUtilizationRequest {
  workspaceId: string;
  components: string[];
}



export interface GetUtilizationResponse {
  utilizations?: UtilizationInfo[];
}



export interface UpdateMetricRequest {
  metricId: string;
  point?: MetricPoint;
}



export interface UpdateMetricResponse {
}



export interface ClearMetricsRequest {
}



export interface ClearMetricsResponse {
}



export interface SetMetricsListRequest {
  metrics?: Metric[];
}



export interface SetMetricsListResponse {
}



export interface UpdateDiagramRequest {
  diagram?: SystemDiagram;
}



export interface UpdateDiagramResponse {
}



export interface HighlightComponentsRequest {
  componentIds: string[];
  highlightType: string;
  color: string;
}



export interface HighlightComponentsResponse {
}



export interface ClearHighlightsRequest {
  types: string[];
}



export interface ClearHighlightsResponse {
}



export interface UpdateGeneratorStateRequest {
  generatorId: string;
  enabled: boolean;
  rate: number;
  status: string;
}



export interface UpdateGeneratorStateResponse {
}



export interface SetGeneratorListRequest {
  generators?: Generator[];
}



export interface SetGeneratorListResponse {
}



export interface LogMessageRequest {
  level: string;
  message: string;
  source: string;
  timestamp: number;
}



export interface LogMessageResponse {
}



export interface ClearConsoleRequest {
}



export interface ClearConsoleResponse {
}



export interface UpdateFlowRatesRequest {
  rates: Record<string, number>;
  strategy: string;
}



export interface UpdateFlowRatesResponse {
}



export interface ShowFlowPathRequest {
  segments?: FlowPathSegment[];
  color: string;
  label: string;
}



export interface FlowPathSegment {
  fromComponent: string;
  fromMethod: string;
  toComponent: string;
  toMethod: string;
  rate: number;
}



export interface ShowFlowPathResponse {
}



export interface ClearFlowPathsRequest {
}



export interface ClearFlowPathsResponse {
}



export interface UpdateUtilizationRequest {
  utilizations?: UtilizationInfo[];
}



export interface UpdateUtilizationResponse {
}



export interface DevEnvSystemChangedRequest {
  systemName: string;
  /** Full list of available systems after the change */
  availableSystems: string[];
}



export interface DevEnvSystemChangedResponse {
}



export interface DevEnvAvailableSystemsRequest {
  systemNames: string[];
}



export interface DevEnvAvailableSystemsResponse {
}



export interface DevEnvUpdateGeneratorRequest {
  name: string;
  generator?: Generator;
}



export interface DevEnvUpdateGeneratorResponse {
}



export interface DevEnvRemoveGeneratorRequest {
  name: string;
}



export interface DevEnvRemoveGeneratorResponse {
}



export interface DevEnvUpdateMetricRequest {
  name: string;
  metric?: Metric;
}



export interface DevEnvUpdateMetricResponse {
}



export interface DevEnvRemoveMetricRequest {
  name: string;
}



export interface DevEnvRemoveMetricResponse {
}


/**
 * FileInfo represents information about a file or directory
 */
export interface FileInfo {
  name: string;
  path: string;
  isDirectory: boolean;
  size: number;
  modTime: string;
  mimeType: string;
}


/**
 * FilesystemInfo represents information about a mounted filesystem
 */
export interface FilesystemInfo {
  id: string;
  prefix: string;
  type: string;
  readOnly: boolean;
  basePath: string;
  extensions: string[];
}



export interface ListFilesystemsRequest {
}



export interface ListFilesystemsResponse {
  filesystems?: FilesystemInfo[];
}



export interface ListFilesRequest {
  filesystemId: string;
  path: string;
}



export interface ListFilesResponse {
  files?: FileInfo[];
}



export interface ReadFileRequest {
  filesystemId: string;
  path: string;
}



export interface ReadFileResponse {
  content: Uint8Array;
  fileInfo?: FileInfo;
}



export interface WriteFileRequest {
  filesystemId: string;
  path: string;
  content: Uint8Array;
}



export interface WriteFileResponse {
  fileInfo?: FileInfo;
}



export interface DeleteFileRequest {
  filesystemId: string;
  path: string;
}



export interface DeleteFileResponse {
  success: boolean;
}



export interface CreateDirectoryRequest {
  filesystemId: string;
  path: string;
}



export interface CreateDirectoryResponse {
  directoryInfo?: FileInfo;
}



export interface GetFileInfoRequest {
  filesystemId: string;
  path: string;
}



export interface GetFileInfoResponse {
  fileInfo?: FileInfo;
}


/**
 * SystemInfo represents a system in the catalog
 */
export interface SystemInfo {
  id: string;
  name: string;
  description: string;
  category: string;
  difficulty: string;
  tags: string[];
  icon: string;
  lastUpdated: string;
}


/**
 * SystemProject represents a full system project
 */
export interface SystemProject {
  id: string;
  name: string;
  description: string;
  category: string;
  difficulty: string;
  tags: string[];
  icon: string;
  versions: Record<string, SystemVersion>;
  defaultVersion: string;
  lastUpdated: string;
}


/**
 * SystemVersion represents a version of a system
 */
export interface SystemVersion {
  sdl: string;
  recipe: string;
  readme: string;
}



export interface ListSystemsRequest {
}



export interface ListSystemsResponse {
  systems?: SystemInfo[];
}



export interface GetSystemRequest {
  id: string;
  version: string;
}



export interface GetSystemResponse {
  system?: SystemProject;
}



export interface GetSystemContentRequest {
  id: string;
  version: string;
}



export interface GetSystemContentResponse {
  sdlContent: string;
  recipeContent: string;
  readmeContent: string;
}



export interface InitializeSingletonRequest {
  workspaceId: string;
  /** SDL content to load initially */
  sdlContent: string;
  /** System name to use after loading */
  systemName: string;
  /** JSON-encoded generator configs */
  generatorsData: string;
  /** JSON-encoded metric configs */
  metricsData: string;
}



export interface InitializeSingletonResponse {
  success: boolean;
  error: string;
  workspaceId: string;
  availableSystems?: SystemInfo[];
}



export interface InitializePresenterRequest {
  workspaceId: string;
}



export interface InitializePresenterResponse {
  success: boolean;
  error: string;
  workspaceId: string;
  /** Available systems to choose from */
  availableSystems?: SystemInfo[];
  /** Initial state to render */
  diagram?: SystemDiagram;
  generators?: Generator[];
  metrics?: Metric[];
}



export interface ClientReadyRequest {
  workspaceId: string;
}



export interface ClientReadyResponse {
  success: boolean;
  workspace?: Workspace;
}



export interface FileSelectedRequest {
  workspaceId: string;
  filePath: string;
}



export interface FileSelectedResponse {
  success: boolean;
  content: string;
  error: string;
}



export interface FileSavedRequest {
  workspaceId: string;
  filePath: string;
  content: string;
}



export interface FileSavedResponse {
  success: boolean;
  error: string;
}



export interface DiagramComponentClickedRequest {
  workspaceId: string;
  componentName: string;
  methodName: string;
}



export interface DiagramComponentClickedResponse {
  success: boolean;
}



export interface DiagramComponentHoveredRequest {
  workspaceId: string;
  componentName: string;
  methodName: string;
}



export interface DiagramComponentHoveredResponse {
  success: boolean;
}



export interface CreateWorkspaceRequest {
  workspace?: Workspace;
}



export interface CreateWorkspaceResponse {
  workspace?: Workspace;
}



export interface GetWorkspaceRequest {
  id: string;
}



export interface GetWorkspaceResponse {
  workspace?: Workspace;
}



export interface ListWorkspacesRequest {
}



export interface ListWorkspacesResponse {
  workspaces?: Workspace[];
}



export interface DeleteWorkspaceRequest {
  id: string;
}



export interface DeleteWorkspaceResponse {
}



export interface UpdateWorkspaceRequest {
  workspace?: Workspace;
}



export interface UpdateWorkspaceResponse {
  workspace?: Workspace;
}


/**
 * Get SDL content for a specific design within a workspace
 */
export interface GetDesignContentRequest {
  workspaceId: string;
  designName: string;
}



export interface GetDesignContentResponse {
  sdlContent: string;
  designName: string;
}


/**
 * Get all design contents for a workspace
 */
export interface GetAllDesignContentsRequest {
  workspaceId: string;
}



export interface GetAllDesignContentsResponse {
  /** design name -> SDL content */
  contents: Record<string, string>;
}

