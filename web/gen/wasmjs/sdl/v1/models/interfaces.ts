// Generated TypeScript interfaces from proto file
// DO NOT EDIT - This file is auto-generated

import { FieldMask, Timestamp } from "@bufbuild/protobuf/wkt";




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



export interface Canvas {
  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  /** Unique ID for the canvas */
  id: string;
  /** Human-readable name for the canvas */
  name: string;
  /** Description of what this canvas is for */
  description: string;
  /** The currently active system (from the systems defined in system_contents) */
  activeSystem: string;
  /** Contents of the .sdl file that defines one or more systems */
  systemContents: string;
  /** Recipe files for various scenarios (name -> contents map) */
  recipes: Record<string, string>;
  /** Registered generators for this canvas */
  generators?: Generator[];
  /** Registered live metrics for this canvas */
  metrics?: Metric[];
}



export interface File {
  /** Path relative to the canvas root */
  path: string;
  /** Contents of the file */
  contents: string;
}



export interface Generator {
  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  /** ID of the generator */
  id: string;
  /** Canvas this generator is sending traffic to */
  canvasId: string;
  /** A descriptive label */
  name: string;
  /** Name of the target component to generate traffic on. This component should be defined in the System,
 eg "server" */
  component: string;
  /** Method in the target component to generate traffic on. */
  method: string;
  /** Traffic rate in RPS (>= 1).  Does not support < 1 yet */
  rate: number;
  /** Duration in seconds over which the genarator is run. 0 => For ever */
  duration: number;
  /** whether it is enabled or not */
  enabled: boolean;
}



export interface Metric {
  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  id: string;
  canvasId: string;
  /** A descriptive label */
  name: string;
  /** Name of the target component to monitor
 eg "server" */
  component: string;
  /** Method in the target component to generate traffic on. */
  methods: string[];
  /** whether it is enabled or not */
  enabled: boolean;
  /** Type of metric "count" or "latency" */
  metricType: string;
  /** Type of aggregation on the metric */
  aggregation: string;
  /** Aggregation window (in seconds) to match on */
  aggregationWindow: number;
  /** Result value to match */
  matchResult: string;
  /** The result "type" if a matching result is provided
 This will be parsed into a type declaration so we know how to treat
 the match_result value provided */
  matchResultType: string;
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



export interface CreateCanvasRequest {
  canvas?: Canvas;
}



export interface CreateCanvasResponse {
  canvas?: Canvas;
  fieldErrors: Record<string, string>;
}



export interface UpdateCanvasRequest {
  canvas?: Canvas;
}



export interface UpdateCanvasResponse {
  canvas?: Canvas;
  updateMask?: FieldMask;
  deletedFiles: string[];
  updatedFiles: Record<string, File>;
}



export interface ListCanvasesRequest {
  pagination?: Pagination;
}



export interface ListCanvasesResponse {
  canvases?: Canvas[];
  pagination?: PaginationResponse;
}



export interface GetCanvasRequest {
  id: string;
}



export interface GetCanvasResponse {
  canvas?: Canvas;
}



export interface DeleteCanvasRequest {
  id: string;
}



export interface DeleteCanvasResponse {
}



export interface ResetCanvasRequest {
  canvasId: string;
}



export interface ResetCanvasResponse {
  success: boolean;
  message: string;
}



export interface LoadFileRequest {
  canvasId: string;
  sdlFilePath: string;
}



export interface LoadFileResponse {
}



export interface UseSystemRequest {
  canvasId: string;
  systemName: string;
}



export interface UseSystemResponse {
}



export interface AddGeneratorRequest {
  generator?: Generator;
  applyFlows: boolean;
}



export interface AddGeneratorResponse {
  generator?: Generator;
}



export interface ListGeneratorsRequest {
  canvasId: string;
}



export interface ListGeneratorsResponse {
  generators?: Generator[];
}



export interface GetGeneratorRequest {
  canvasId: string;
  generatorId: string;
}



export interface GetGeneratorResponse {
  generator?: Generator;
}



export interface UpdateGeneratorRequest {
  generator?: Generator;
  updateMask?: FieldMask;
  applyFlows: boolean;
}



export interface UpdateGeneratorResponse {
  generator?: Generator;
}



export interface StartGeneratorRequest {
  canvasId: string;
  generatorId: string;
}



export interface StartGeneratorResponse {
}



export interface StopGeneratorRequest {
  canvasId: string;
  generatorId: string;
}



export interface StopGeneratorResponse {
}



export interface DeleteGeneratorRequest {
  canvasId: string;
  generatorId: string;
  applyFlows: boolean;
}



export interface DeleteGeneratorResponse {
}



export interface StartAllGeneratorsRequest {
  canvasId: string;
}



export interface StartAllGeneratorsResponse {
  totalGenerators: number;
  startedCount: number;
  alreadyRunningCount: number;
  failedCount: number;
  failedIds: string[];
}



export interface StopAllGeneratorsRequest {
  canvasId: string;
}



export interface StopAllGeneratorsResponse {
  totalGenerators: number;
  stoppedCount: number;
  alreadyStoppedCount: number;
  failedCount: number;
  failedIds: string[];
}



export interface AddMetricRequest {
  metric?: Metric;
}



export interface AddMetricResponse {
  metric?: Metric;
}



export interface DeleteMetricRequest {
  canvasId: string;
  metricId: string;
}



export interface DeleteMetricResponse {
}



export interface ListMetricsRequest {
  canvasId: string;
}



export interface ListMetricsResponse {
  metrics?: Metric[];
}



export interface QueryMetricsRequest {
  canvasId: string;
  metricId: string;
  startTime: number;
  endTime: number;
  limit: number;
}



export interface QueryMetricsResponse {
  points?: MetricPoint[];
}



export interface AggregateMetricsRequest {
  canvasId: string;
  metricId: string;
  startTime: number;
  endTime: number;
  function: string;
  windowSize: number;
}



export interface AggregateMetricsResponse {
  results?: AggregateResult[];
}



export interface StreamMetricsRequest {
  canvasId: string;
  metricIds: string[];
}



export interface StreamMetricsResponse {
  updates?: MetricUpdate[];
}



export interface ExecuteTraceRequest {
  canvasId: string;
  component: string;
  method: string;
}



export interface ExecuteTraceResponse {
  traceData?: TraceData;
}



export interface TraceAllPathsRequest {
  canvasId: string;
  component: string;
  method: string;
  maxDepth: number;
}



export interface TraceAllPathsResponse {
  traceData?: AllPathsTraceData;
}



export interface SetParameterRequest {
  canvasId: string;
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
  canvasId: string;
  path: string;
}



export interface GetParametersResponse {
  parameters: Record<string, string>;
}



export interface BatchSetParametersRequest {
  canvasId: string;
  updates?: ParameterUpdate[];
}



export interface BatchSetParametersResponse {
  success: boolean;
  errorMessage: string;
  results?: ParameterUpdateResult[];
}



export interface EvaluateFlowsRequest {
  canvasId: string;
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
  canvasId: string;
}



export interface GetFlowStateResponse {
  state?: FlowState;
}



export interface GetSystemDiagramRequest {
  canvasId: string;
}



export interface GetSystemDiagramResponse {
  diagram?: SystemDiagram;
}



export interface GetUtilizationRequest {
  canvasId: string;
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
  canvasId: string;
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
  canvasId: string;
  availableSystems?: SystemInfo[];
}



export interface InitializePresenterRequest {
  canvasId: string;
}



export interface InitializePresenterResponse {
  success: boolean;
  error: string;
  canvasId: string;
  /** Available systems to choose from */
  availableSystems?: SystemInfo[];
  /** Initial state to render */
  diagram?: SystemDiagram;
  generators?: Generator[];
  metrics?: Metric[];
}



export interface ClientReadyRequest {
  canvasId: string;
}



export interface ClientReadyResponse {
  success: boolean;
  canvas?: Canvas;
}



export interface FileSelectedRequest {
  canvasId: string;
  filePath: string;
}



export interface FileSelectedResponse {
  success: boolean;
  content: string;
  error: string;
}



export interface FileSavedRequest {
  canvasId: string;
  filePath: string;
  content: string;
}



export interface FileSavedResponse {
  success: boolean;
  error: string;
}



export interface DiagramComponentClickedRequest {
  canvasId: string;
  componentName: string;
  methodName: string;
}



export interface DiagramComponentClickedResponse {
  success: boolean;
}



export interface DiagramComponentHoveredRequest {
  canvasId: string;
  componentName: string;
  methodName: string;
}



export interface DiagramComponentHoveredResponse {
  success: boolean;
}

