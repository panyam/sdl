import { FieldMask, Timestamp } from "@bufbuild/protobuf/wkt";


import { Pagination as PaginationInterface, PaginationResponse as PaginationResponseInterface, Canvas as CanvasInterface, File as FileInterface, Generator as GeneratorInterface, Metric as MetricInterface, MetricPoint as MetricPointInterface, MetricUpdate as MetricUpdateInterface, SystemDiagram as SystemDiagramInterface, DiagramNode as DiagramNodeInterface, MethodInfo as MethodInfoInterface, DiagramEdge as DiagramEdgeInterface, UtilizationInfo as UtilizationInfoInterface, FlowEdge as FlowEdgeInterface, FlowState as FlowStateInterface, TraceData as TraceDataInterface, TraceEvent as TraceEventInterface, AllPathsTraceData as AllPathsTraceDataInterface, TraceNode as TraceNodeInterface, Edge as EdgeInterface, GroupInfo as GroupInfoInterface, ParameterUpdate as ParameterUpdateInterface, ParameterUpdateResult as ParameterUpdateResultInterface, AggregateResult as AggregateResultInterface, CreateCanvasRequest as CreateCanvasRequestInterface, CreateCanvasResponse as CreateCanvasResponseInterface, UpdateCanvasRequest as UpdateCanvasRequestInterface, UpdateCanvasResponse as UpdateCanvasResponseInterface, ListCanvasesRequest as ListCanvasesRequestInterface, ListCanvasesResponse as ListCanvasesResponseInterface, GetCanvasRequest as GetCanvasRequestInterface, GetCanvasResponse as GetCanvasResponseInterface, DeleteCanvasRequest as DeleteCanvasRequestInterface, DeleteCanvasResponse as DeleteCanvasResponseInterface, ResetCanvasRequest as ResetCanvasRequestInterface, ResetCanvasResponse as ResetCanvasResponseInterface, LoadFileRequest as LoadFileRequestInterface, LoadFileResponse as LoadFileResponseInterface, UseSystemRequest as UseSystemRequestInterface, UseSystemResponse as UseSystemResponseInterface, AddGeneratorRequest as AddGeneratorRequestInterface, AddGeneratorResponse as AddGeneratorResponseInterface, ListGeneratorsRequest as ListGeneratorsRequestInterface, ListGeneratorsResponse as ListGeneratorsResponseInterface, GetGeneratorRequest as GetGeneratorRequestInterface, GetGeneratorResponse as GetGeneratorResponseInterface, UpdateGeneratorRequest as UpdateGeneratorRequestInterface, UpdateGeneratorResponse as UpdateGeneratorResponseInterface, StartGeneratorRequest as StartGeneratorRequestInterface, StartGeneratorResponse as StartGeneratorResponseInterface, StopGeneratorRequest as StopGeneratorRequestInterface, StopGeneratorResponse as StopGeneratorResponseInterface, DeleteGeneratorRequest as DeleteGeneratorRequestInterface, DeleteGeneratorResponse as DeleteGeneratorResponseInterface, StartAllGeneratorsRequest as StartAllGeneratorsRequestInterface, StartAllGeneratorsResponse as StartAllGeneratorsResponseInterface, StopAllGeneratorsRequest as StopAllGeneratorsRequestInterface, StopAllGeneratorsResponse as StopAllGeneratorsResponseInterface, AddMetricRequest as AddMetricRequestInterface, AddMetricResponse as AddMetricResponseInterface, DeleteMetricRequest as DeleteMetricRequestInterface, DeleteMetricResponse as DeleteMetricResponseInterface, ListMetricsRequest as ListMetricsRequestInterface, ListMetricsResponse as ListMetricsResponseInterface, QueryMetricsRequest as QueryMetricsRequestInterface, QueryMetricsResponse as QueryMetricsResponseInterface, AggregateMetricsRequest as AggregateMetricsRequestInterface, AggregateMetricsResponse as AggregateMetricsResponseInterface, StreamMetricsRequest as StreamMetricsRequestInterface, StreamMetricsResponse as StreamMetricsResponseInterface, ExecuteTraceRequest as ExecuteTraceRequestInterface, ExecuteTraceResponse as ExecuteTraceResponseInterface, TraceAllPathsRequest as TraceAllPathsRequestInterface, TraceAllPathsResponse as TraceAllPathsResponseInterface, SetParameterRequest as SetParameterRequestInterface, SetParameterResponse as SetParameterResponseInterface, GetParametersRequest as GetParametersRequestInterface, GetParametersResponse as GetParametersResponseInterface, BatchSetParametersRequest as BatchSetParametersRequestInterface, BatchSetParametersResponse as BatchSetParametersResponseInterface, EvaluateFlowsRequest as EvaluateFlowsRequestInterface, EvaluateFlowsResponse as EvaluateFlowsResponseInterface, GetFlowStateRequest as GetFlowStateRequestInterface, GetFlowStateResponse as GetFlowStateResponseInterface, GetSystemDiagramRequest as GetSystemDiagramRequestInterface, GetSystemDiagramResponse as GetSystemDiagramResponseInterface, GetUtilizationRequest as GetUtilizationRequestInterface, GetUtilizationResponse as GetUtilizationResponseInterface, UpdateMetricRequest as UpdateMetricRequestInterface, UpdateMetricResponse as UpdateMetricResponseInterface, ClearMetricsRequest as ClearMetricsRequestInterface, ClearMetricsResponse as ClearMetricsResponseInterface, SetMetricsListRequest as SetMetricsListRequestInterface, SetMetricsListResponse as SetMetricsListResponseInterface, UpdateDiagramRequest as UpdateDiagramRequestInterface, UpdateDiagramResponse as UpdateDiagramResponseInterface, HighlightComponentsRequest as HighlightComponentsRequestInterface, HighlightComponentsResponse as HighlightComponentsResponseInterface, ClearHighlightsRequest as ClearHighlightsRequestInterface, ClearHighlightsResponse as ClearHighlightsResponseInterface, UpdateGeneratorStateRequest as UpdateGeneratorStateRequestInterface, UpdateGeneratorStateResponse as UpdateGeneratorStateResponseInterface, SetGeneratorListRequest as SetGeneratorListRequestInterface, SetGeneratorListResponse as SetGeneratorListResponseInterface, LogMessageRequest as LogMessageRequestInterface, LogMessageResponse as LogMessageResponseInterface, ClearConsoleRequest as ClearConsoleRequestInterface, ClearConsoleResponse as ClearConsoleResponseInterface, UpdateFlowRatesRequest as UpdateFlowRatesRequestInterface, UpdateFlowRatesResponse as UpdateFlowRatesResponseInterface, ShowFlowPathRequest as ShowFlowPathRequestInterface, FlowPathSegment as FlowPathSegmentInterface, ShowFlowPathResponse as ShowFlowPathResponseInterface, ClearFlowPathsRequest as ClearFlowPathsRequestInterface, ClearFlowPathsResponse as ClearFlowPathsResponseInterface, UpdateUtilizationRequest as UpdateUtilizationRequestInterface, UpdateUtilizationResponse as UpdateUtilizationResponseInterface, FileInfo as FileInfoInterface, FilesystemInfo as FilesystemInfoInterface, ListFilesystemsRequest as ListFilesystemsRequestInterface, ListFilesystemsResponse as ListFilesystemsResponseInterface, ListFilesRequest as ListFilesRequestInterface, ListFilesResponse as ListFilesResponseInterface, ReadFileRequest as ReadFileRequestInterface, ReadFileResponse as ReadFileResponseInterface, WriteFileRequest as WriteFileRequestInterface, WriteFileResponse as WriteFileResponseInterface, DeleteFileRequest as DeleteFileRequestInterface, DeleteFileResponse as DeleteFileResponseInterface, CreateDirectoryRequest as CreateDirectoryRequestInterface, CreateDirectoryResponse as CreateDirectoryResponseInterface, GetFileInfoRequest as GetFileInfoRequestInterface, GetFileInfoResponse as GetFileInfoResponseInterface, SystemInfo as SystemInfoInterface, SystemProject as SystemProjectInterface, SystemVersion as SystemVersionInterface, ListSystemsRequest as ListSystemsRequestInterface, ListSystemsResponse as ListSystemsResponseInterface, GetSystemRequest as GetSystemRequestInterface, GetSystemResponse as GetSystemResponseInterface, GetSystemContentRequest as GetSystemContentRequestInterface, GetSystemContentResponse as GetSystemContentResponseInterface, InitializeSingletonRequest as InitializeSingletonRequestInterface, InitializeSingletonResponse as InitializeSingletonResponseInterface, InitializePresenterRequest as InitializePresenterRequestInterface, InitializePresenterResponse as InitializePresenterResponseInterface, ClientReadyRequest as ClientReadyRequestInterface, ClientReadyResponse as ClientReadyResponseInterface, FileSelectedRequest as FileSelectedRequestInterface, FileSelectedResponse as FileSelectedResponseInterface, FileSavedRequest as FileSavedRequestInterface, FileSavedResponse as FileSavedResponseInterface, DiagramComponentClickedRequest as DiagramComponentClickedRequestInterface, DiagramComponentClickedResponse as DiagramComponentClickedResponseInterface, DiagramComponentHoveredRequest as DiagramComponentHoveredRequestInterface, DiagramComponentHoveredResponse as DiagramComponentHoveredResponseInterface } from "./interfaces";





export class Pagination implements PaginationInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.Pagination";
  readonly __MESSAGE_TYPE = Pagination.MESSAGE_TYPE;

  /** Instead of an offset an abstract  "page" key is provided that offers
 an opaque "pointer" into some offset in a result set. */
  pageKey: string = "";
  /** If a pagekey is not supported we can also support a direct integer offset
 for cases where it makes sense. */
  pageOffset: number = 0;
  /** Number of results to return. */
  pageSize: number = 0;

  
}



export class PaginationResponse implements PaginationResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.PaginationResponse";
  readonly __MESSAGE_TYPE = PaginationResponse.MESSAGE_TYPE;

  /** The key/pointer string that subsequent List requests should pass to
 continue the pagination. */
  nextPageKey: string = "";
  /** Also support an integer offset if possible */
  nextPageOffset: number = 0;
  /** Whether theere are more results. */
  hasMore: boolean = false;
  /** Total number of results. */
  totalResults: number = 0;

  
}



export class Canvas implements CanvasInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.Canvas";
  readonly __MESSAGE_TYPE = Canvas.MESSAGE_TYPE;

  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  /** Unique ID for the canvas */
  id: string = "";
  /** Human-readable name for the canvas */
  name: string = "";
  /** Description of what this canvas is for */
  description: string = "";
  /** The currently active system (from the systems defined in system_contents) */
  activeSystem: string = "";
  /** Contents of the .sdl file that defines one or more systems */
  systemContents: string = "";
  /** Recipe files for various scenarios (name -> contents map) */
  recipes: Record<string, string> = {};
  /** Registered generators for this canvas */
  generators: Generator[] = [];
  /** Registered live metrics for this canvas */
  metrics: Metric[] = [];
  previewUrl: string = "";

  
}



export class File implements FileInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.File";
  readonly __MESSAGE_TYPE = File.MESSAGE_TYPE;

  /** Path relative to the canvas root */
  path: string = "";
  /** Contents of the file */
  contents: string = "";

  
}



export class Generator implements GeneratorInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.Generator";
  readonly __MESSAGE_TYPE = Generator.MESSAGE_TYPE;

  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  /** ID of the generator */
  id: string = "";
  /** Canvas this generator is sending traffic to */
  canvasId: string = "";
  /** A descriptive label */
  name: string = "";
  /** Name of the target component to generate traffic on. This component should be defined in the System,
 eg "server" */
  component: string = "";
  /** Method in the target component to generate traffic on. */
  method: string = "";
  /** Traffic rate in RPS (>= 1).  Does not support < 1 yet */
  rate: number = 0;
  /** Duration in seconds over which the genarator is run. 0 => For ever */
  duration: number = 0;
  /** whether it is enabled or not */
  enabled: boolean = false;

  
}



export class Metric implements MetricInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.Metric";
  readonly __MESSAGE_TYPE = Metric.MESSAGE_TYPE;

  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  id: string = "";
  canvasId: string = "";
  /** A descriptive label */
  name: string = "";
  /** Name of the target component to monitor
 eg "server" */
  component: string = "";
  /** Method in the target component to generate traffic on. */
  methods: string[] = [];
  /** whether it is enabled or not */
  enabled: boolean = false;
  /** Type of metric "count" or "latency" */
  metricType: string = "";
  /** Type of aggregation on the metric */
  aggregation: string = "";
  /** Aggregation window (in seconds) to match on */
  aggregationWindow: number = 0;
  /** Result value to match */
  matchResult: string = "";
  /** The result "type" if a matching result is provided
 This will be parsed into a type declaration so we know how to treat
 the match_result value provided */
  matchResultType: string = "";
  oldestTimestamp: number = 0;
  newestTimestamp: number = 0;
  numDataPoints: number = 0;

  
}



export class MetricPoint implements MetricPointInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.MetricPoint";
  readonly __MESSAGE_TYPE = MetricPoint.MESSAGE_TYPE;

  timestamp: number = 0;
  value: number = 0;

  
}


/**
 * Individual metric update
 */
export class MetricUpdate implements MetricUpdateInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.MetricUpdate";
  readonly __MESSAGE_TYPE = MetricUpdate.MESSAGE_TYPE;

  metricId: string = "";
  point?: MetricPoint;

  
}


/**
 * SystemDiagram represents the topology of a system
 */
export class SystemDiagram implements SystemDiagramInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.SystemDiagram";
  readonly __MESSAGE_TYPE = SystemDiagram.MESSAGE_TYPE;

  systemName: string = "";
  nodes: DiagramNode[] = [];
  edges: DiagramEdge[] = [];

  
}


/**
 * DiagramNode represents a component or instance in the system
 */
export class DiagramNode implements DiagramNodeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DiagramNode";
  readonly __MESSAGE_TYPE = DiagramNode.MESSAGE_TYPE;

  id: string = "";
  name: string = "";
  type: string = "";
  methods: MethodInfo[] = [];
  traffic: string = "";
  fullPath: string = "";
  icon: string = "";

  
}


/**
 * MethodInfo represents information about a component method
 */
export class MethodInfo implements MethodInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.MethodInfo";
  readonly __MESSAGE_TYPE = MethodInfo.MESSAGE_TYPE;

  name: string = "";
  returnType: string = "";
  traffic: number = 0;

  
}


/**
 * DiagramEdge represents a connection between nodes
 */
export class DiagramEdge implements DiagramEdgeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DiagramEdge";
  readonly __MESSAGE_TYPE = DiagramEdge.MESSAGE_TYPE;

  fromId: string = "";
  toId: string = "";
  fromMethod: string = "";
  toMethod: string = "";
  label: string = "";
  order: number = 0;
  condition: string = "";
  probability: number = 0;
  generatorId: string = "";
  color: string = "";

  
}


/**
 * Resource utilization information
 */
export class UtilizationInfo implements UtilizationInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UtilizationInfo";
  readonly __MESSAGE_TYPE = UtilizationInfo.MESSAGE_TYPE;

  resourceName: string = "";
  componentPath: string = "";
  utilization: number = 0;
  capacity: number = 0;
  currentLoad: number = 0;
  isBottleneck: boolean = false;
  warningThreshold: number = 0;
  criticalThreshold: number = 0;

  
}


/**
 * Represents a flow edge between components
 */
export class FlowEdge implements FlowEdgeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.FlowEdge";
  readonly __MESSAGE_TYPE = FlowEdge.MESSAGE_TYPE;

  fromComponent: string = "";
  fromMethod: string = "";
  toComponent: string = "";
  toMethod: string = "";
  rate: number = 0;
  condition: string = "";

  
}


/**
 * Current flow state
 */
export class FlowState implements FlowStateInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.FlowState";
  readonly __MESSAGE_TYPE = FlowState.MESSAGE_TYPE;

  strategy: string = "";
  rates: Record<string, number> = {};
  manualOverrides: Record<string, number> = {};

  
}


/**
 * TraceData matches the runtime.TraceData structure
 */
export class TraceData implements TraceDataInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.TraceData";
  readonly __MESSAGE_TYPE = TraceData.MESSAGE_TYPE;

  system: string = "";
  entryPoint: string = "";
  events: TraceEvent[] = [];

  
}


/**
 * TraceEvent matches the runtime.TraceEvent structure
 */
export class TraceEvent implements TraceEventInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.TraceEvent";
  readonly __MESSAGE_TYPE = TraceEvent.MESSAGE_TYPE;

  kind: string = "";
  id: number = 0;
  parentId: number = 0;
  timestamp: number = 0;
  duration: number = 0;
  component: string = "";
  method: string = "";
  args: string[] = [];
  returnValue: string = "";
  errorMessage: string = "";

  
}


/**
 * Enhanced TraceData for all-paths traversal - represents the complete execution tree
 */
export class AllPathsTraceData implements AllPathsTraceDataInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.AllPathsTraceData";
  readonly __MESSAGE_TYPE = AllPathsTraceData.MESSAGE_TYPE;

  traceId: string = "";
  /** The root TraceNode always starts from the <Component>.<Method> where we are kicking off the trace from */
  root?: TraceNode;

  
}


/**
 * TraceNode represents a single node in the execution tree
 */
export class TraceNode implements TraceNodeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.TraceNode";
  readonly __MESSAGE_TYPE = TraceNode.MESSAGE_TYPE;

  /** Name of the component and method in the form <Component>.<Method> we are starting the trace from */
  startingTarget: string = "";
  /** All edges in an ordered fashion */
  edges: Edge[] = [];
  /** Multiple groups for flexible labeling of sub-trees (loops, conditionals, etc.) */
  groups: GroupInfo[] = [];

  
}


/**
 * Edge represents a transition from one node to another in the execution tree
 */
export class Edge implements EdgeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.Edge";
  readonly __MESSAGE_TYPE = Edge.MESSAGE_TYPE;

  /** Unique Edge ID across the entire Trace */
  id: string = "";
  /** The next node this edge leads to */
  nextNode?: TraceNode;
  /** Label on the edge (if any) */
  label: string = "";
  /** Async edges denote Futures being sent without a return */
  isAsync: boolean = false;
  /** "Reverse" edges show a "wait" on a future */
  isReverse: boolean = false;
  /** This is optional but leaving it here just in case. */
  probability: string = "";
  /** Condition information for branching */
  condition: string = "";
  /** true if this edge represents a conditional branch */
  isConditional: boolean = false;

  
}


/**
 * GroupInfo allows flexible grouping of edges with labels
 */
export class GroupInfo implements GroupInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GroupInfo";
  readonly __MESSAGE_TYPE = GroupInfo.MESSAGE_TYPE;

  /** Starting edge index */
  groupStart: number = 0;
  /** Ending edge index (inclusive) */
  groupEnd: number = 0;
  /** Generic label: "loop: 3x", "if cached", "switch: status" */
  groupLabel: string = "";
  /** Optional hint: "loop", "conditional", "switch" (for tooling) */
  groupType: string = "";

  
}


/**
 * Single parameter update
 */
export class ParameterUpdate implements ParameterUpdateInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ParameterUpdate";
  readonly __MESSAGE_TYPE = ParameterUpdate.MESSAGE_TYPE;

  path: string = "";
  newValue: string = "";

  
}


/**
 * Result for individual parameter update
 */
export class ParameterUpdateResult implements ParameterUpdateResultInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ParameterUpdateResult";
  readonly __MESSAGE_TYPE = ParameterUpdateResult.MESSAGE_TYPE;

  path: string = "";
  success: boolean = false;
  errorMessage: string = "";
  oldValue: string = "";
  newValue: string = "";

  
}



export class AggregateResult implements AggregateResultInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.AggregateResult";
  readonly __MESSAGE_TYPE = AggregateResult.MESSAGE_TYPE;

  timestamp: number = 0;
  value: number = 0;

  
}



export class CreateCanvasRequest implements CreateCanvasRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.CreateCanvasRequest";
  readonly __MESSAGE_TYPE = CreateCanvasRequest.MESSAGE_TYPE;

  canvas?: Canvas;

  
}



export class CreateCanvasResponse implements CreateCanvasResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.CreateCanvasResponse";
  readonly __MESSAGE_TYPE = CreateCanvasResponse.MESSAGE_TYPE;

  canvas?: Canvas;
  fieldErrors: Record<string, string> = {};

  
}



export class UpdateCanvasRequest implements UpdateCanvasRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateCanvasRequest";
  readonly __MESSAGE_TYPE = UpdateCanvasRequest.MESSAGE_TYPE;

  canvas?: Canvas;

  
}



export class UpdateCanvasResponse implements UpdateCanvasResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateCanvasResponse";
  readonly __MESSAGE_TYPE = UpdateCanvasResponse.MESSAGE_TYPE;

  canvas?: Canvas;
  updateMask?: FieldMask;
  deletedFiles: string[] = [];
  updatedFiles: Record<string, File> = {};

  
}



export class ListCanvasesRequest implements ListCanvasesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListCanvasesRequest";
  readonly __MESSAGE_TYPE = ListCanvasesRequest.MESSAGE_TYPE;

  pagination?: Pagination;

  
}



export class ListCanvasesResponse implements ListCanvasesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListCanvasesResponse";
  readonly __MESSAGE_TYPE = ListCanvasesResponse.MESSAGE_TYPE;

  canvases: Canvas[] = [];
  pagination?: PaginationResponse;

  
}



export class GetCanvasRequest implements GetCanvasRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetCanvasRequest";
  readonly __MESSAGE_TYPE = GetCanvasRequest.MESSAGE_TYPE;

  id: string = "";

  
}



export class GetCanvasResponse implements GetCanvasResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetCanvasResponse";
  readonly __MESSAGE_TYPE = GetCanvasResponse.MESSAGE_TYPE;

  canvas?: Canvas;

  
}



export class DeleteCanvasRequest implements DeleteCanvasRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DeleteCanvasRequest";
  readonly __MESSAGE_TYPE = DeleteCanvasRequest.MESSAGE_TYPE;

  id: string = "";

  
}



export class DeleteCanvasResponse implements DeleteCanvasResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DeleteCanvasResponse";
  readonly __MESSAGE_TYPE = DeleteCanvasResponse.MESSAGE_TYPE;


  
}



export class ResetCanvasRequest implements ResetCanvasRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ResetCanvasRequest";
  readonly __MESSAGE_TYPE = ResetCanvasRequest.MESSAGE_TYPE;

  canvasId: string = "";

  
}



export class ResetCanvasResponse implements ResetCanvasResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ResetCanvasResponse";
  readonly __MESSAGE_TYPE = ResetCanvasResponse.MESSAGE_TYPE;

  success: boolean = false;
  message: string = "";

  
}



export class LoadFileRequest implements LoadFileRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.LoadFileRequest";
  readonly __MESSAGE_TYPE = LoadFileRequest.MESSAGE_TYPE;

  canvasId: string = "";
  sdlFilePath: string = "";

  
}



export class LoadFileResponse implements LoadFileResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.LoadFileResponse";
  readonly __MESSAGE_TYPE = LoadFileResponse.MESSAGE_TYPE;


  
}



export class UseSystemRequest implements UseSystemRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UseSystemRequest";
  readonly __MESSAGE_TYPE = UseSystemRequest.MESSAGE_TYPE;

  canvasId: string = "";
  systemName: string = "";

  
}



export class UseSystemResponse implements UseSystemResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UseSystemResponse";
  readonly __MESSAGE_TYPE = UseSystemResponse.MESSAGE_TYPE;


  
}



export class AddGeneratorRequest implements AddGeneratorRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.AddGeneratorRequest";
  readonly __MESSAGE_TYPE = AddGeneratorRequest.MESSAGE_TYPE;

  generator?: Generator;
  applyFlows: boolean = false;

  
}



export class AddGeneratorResponse implements AddGeneratorResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.AddGeneratorResponse";
  readonly __MESSAGE_TYPE = AddGeneratorResponse.MESSAGE_TYPE;

  generator?: Generator;

  
}



export class ListGeneratorsRequest implements ListGeneratorsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListGeneratorsRequest";
  readonly __MESSAGE_TYPE = ListGeneratorsRequest.MESSAGE_TYPE;

  canvasId: string = "";

  
}



export class ListGeneratorsResponse implements ListGeneratorsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListGeneratorsResponse";
  readonly __MESSAGE_TYPE = ListGeneratorsResponse.MESSAGE_TYPE;

  generators: Generator[] = [];

  
}



export class GetGeneratorRequest implements GetGeneratorRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetGeneratorRequest";
  readonly __MESSAGE_TYPE = GetGeneratorRequest.MESSAGE_TYPE;

  canvasId: string = "";
  generatorId: string = "";

  
}



export class GetGeneratorResponse implements GetGeneratorResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetGeneratorResponse";
  readonly __MESSAGE_TYPE = GetGeneratorResponse.MESSAGE_TYPE;

  generator?: Generator;

  
}



export class UpdateGeneratorRequest implements UpdateGeneratorRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateGeneratorRequest";
  readonly __MESSAGE_TYPE = UpdateGeneratorRequest.MESSAGE_TYPE;

  generator?: Generator;
  updateMask?: FieldMask;
  applyFlows: boolean = false;

  
}



export class UpdateGeneratorResponse implements UpdateGeneratorResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateGeneratorResponse";
  readonly __MESSAGE_TYPE = UpdateGeneratorResponse.MESSAGE_TYPE;

  generator?: Generator;

  
}



export class StartGeneratorRequest implements StartGeneratorRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.StartGeneratorRequest";
  readonly __MESSAGE_TYPE = StartGeneratorRequest.MESSAGE_TYPE;

  canvasId: string = "";
  generatorId: string = "";

  
}



export class StartGeneratorResponse implements StartGeneratorResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.StartGeneratorResponse";
  readonly __MESSAGE_TYPE = StartGeneratorResponse.MESSAGE_TYPE;


  
}



export class StopGeneratorRequest implements StopGeneratorRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.StopGeneratorRequest";
  readonly __MESSAGE_TYPE = StopGeneratorRequest.MESSAGE_TYPE;

  canvasId: string = "";
  generatorId: string = "";

  
}



export class StopGeneratorResponse implements StopGeneratorResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.StopGeneratorResponse";
  readonly __MESSAGE_TYPE = StopGeneratorResponse.MESSAGE_TYPE;


  
}



export class DeleteGeneratorRequest implements DeleteGeneratorRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DeleteGeneratorRequest";
  readonly __MESSAGE_TYPE = DeleteGeneratorRequest.MESSAGE_TYPE;

  canvasId: string = "";
  generatorId: string = "";
  applyFlows: boolean = false;

  
}



export class DeleteGeneratorResponse implements DeleteGeneratorResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DeleteGeneratorResponse";
  readonly __MESSAGE_TYPE = DeleteGeneratorResponse.MESSAGE_TYPE;


  
}



export class StartAllGeneratorsRequest implements StartAllGeneratorsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.StartAllGeneratorsRequest";
  readonly __MESSAGE_TYPE = StartAllGeneratorsRequest.MESSAGE_TYPE;

  canvasId: string = "";

  
}



export class StartAllGeneratorsResponse implements StartAllGeneratorsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.StartAllGeneratorsResponse";
  readonly __MESSAGE_TYPE = StartAllGeneratorsResponse.MESSAGE_TYPE;

  totalGenerators: number = 0;
  startedCount: number = 0;
  alreadyRunningCount: number = 0;
  failedCount: number = 0;
  failedIds: string[] = [];

  
}



export class StopAllGeneratorsRequest implements StopAllGeneratorsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.StopAllGeneratorsRequest";
  readonly __MESSAGE_TYPE = StopAllGeneratorsRequest.MESSAGE_TYPE;

  canvasId: string = "";

  
}



export class StopAllGeneratorsResponse implements StopAllGeneratorsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.StopAllGeneratorsResponse";
  readonly __MESSAGE_TYPE = StopAllGeneratorsResponse.MESSAGE_TYPE;

  totalGenerators: number = 0;
  stoppedCount: number = 0;
  alreadyStoppedCount: number = 0;
  failedCount: number = 0;
  failedIds: string[] = [];

  
}



export class AddMetricRequest implements AddMetricRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.AddMetricRequest";
  readonly __MESSAGE_TYPE = AddMetricRequest.MESSAGE_TYPE;

  metric?: Metric;

  
}



export class AddMetricResponse implements AddMetricResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.AddMetricResponse";
  readonly __MESSAGE_TYPE = AddMetricResponse.MESSAGE_TYPE;

  metric?: Metric;

  
}



export class DeleteMetricRequest implements DeleteMetricRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DeleteMetricRequest";
  readonly __MESSAGE_TYPE = DeleteMetricRequest.MESSAGE_TYPE;

  canvasId: string = "";
  metricId: string = "";

  
}



export class DeleteMetricResponse implements DeleteMetricResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DeleteMetricResponse";
  readonly __MESSAGE_TYPE = DeleteMetricResponse.MESSAGE_TYPE;


  
}



export class ListMetricsRequest implements ListMetricsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListMetricsRequest";
  readonly __MESSAGE_TYPE = ListMetricsRequest.MESSAGE_TYPE;

  canvasId: string = "";

  
}



export class ListMetricsResponse implements ListMetricsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListMetricsResponse";
  readonly __MESSAGE_TYPE = ListMetricsResponse.MESSAGE_TYPE;

  metrics: Metric[] = [];

  
}



export class QueryMetricsRequest implements QueryMetricsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.QueryMetricsRequest";
  readonly __MESSAGE_TYPE = QueryMetricsRequest.MESSAGE_TYPE;

  canvasId: string = "";
  metricId: string = "";
  startTime: number = 0;
  endTime: number = 0;
  limit: number = 0;

  
}



export class QueryMetricsResponse implements QueryMetricsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.QueryMetricsResponse";
  readonly __MESSAGE_TYPE = QueryMetricsResponse.MESSAGE_TYPE;

  points: MetricPoint[] = [];

  
}



export class AggregateMetricsRequest implements AggregateMetricsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.AggregateMetricsRequest";
  readonly __MESSAGE_TYPE = AggregateMetricsRequest.MESSAGE_TYPE;

  canvasId: string = "";
  metricId: string = "";
  startTime: number = 0;
  endTime: number = 0;
  function: string = "";
  windowSize: number = 0;

  
}



export class AggregateMetricsResponse implements AggregateMetricsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.AggregateMetricsResponse";
  readonly __MESSAGE_TYPE = AggregateMetricsResponse.MESSAGE_TYPE;

  results: AggregateResult[] = [];

  
}



export class StreamMetricsRequest implements StreamMetricsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.StreamMetricsRequest";
  readonly __MESSAGE_TYPE = StreamMetricsRequest.MESSAGE_TYPE;

  canvasId: string = "";
  metricIds: string[] = [];

  
}



export class StreamMetricsResponse implements StreamMetricsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.StreamMetricsResponse";
  readonly __MESSAGE_TYPE = StreamMetricsResponse.MESSAGE_TYPE;

  updates: MetricUpdate[] = [];

  
}



export class ExecuteTraceRequest implements ExecuteTraceRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ExecuteTraceRequest";
  readonly __MESSAGE_TYPE = ExecuteTraceRequest.MESSAGE_TYPE;

  canvasId: string = "";
  component: string = "";
  method: string = "";

  
}



export class ExecuteTraceResponse implements ExecuteTraceResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ExecuteTraceResponse";
  readonly __MESSAGE_TYPE = ExecuteTraceResponse.MESSAGE_TYPE;

  traceData?: TraceData;

  
}



export class TraceAllPathsRequest implements TraceAllPathsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.TraceAllPathsRequest";
  readonly __MESSAGE_TYPE = TraceAllPathsRequest.MESSAGE_TYPE;

  canvasId: string = "";
  component: string = "";
  method: string = "";
  maxDepth: number = 0;

  
}



export class TraceAllPathsResponse implements TraceAllPathsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.TraceAllPathsResponse";
  readonly __MESSAGE_TYPE = TraceAllPathsResponse.MESSAGE_TYPE;

  traceData?: AllPathsTraceData;

  
}



export class SetParameterRequest implements SetParameterRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.SetParameterRequest";
  readonly __MESSAGE_TYPE = SetParameterRequest.MESSAGE_TYPE;

  canvasId: string = "";
  path: string = "";
  newValue: string = "";

  
}



export class SetParameterResponse implements SetParameterResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.SetParameterResponse";
  readonly __MESSAGE_TYPE = SetParameterResponse.MESSAGE_TYPE;

  success: boolean = false;
  errorMessage: string = "";
  newValue: string = "";
  oldValue: string = "";

  
}



export class GetParametersRequest implements GetParametersRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetParametersRequest";
  readonly __MESSAGE_TYPE = GetParametersRequest.MESSAGE_TYPE;

  canvasId: string = "";
  path: string = "";

  
}



export class GetParametersResponse implements GetParametersResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetParametersResponse";
  readonly __MESSAGE_TYPE = GetParametersResponse.MESSAGE_TYPE;

  parameters: Record<string, string> = {};

  
}



export class BatchSetParametersRequest implements BatchSetParametersRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.BatchSetParametersRequest";
  readonly __MESSAGE_TYPE = BatchSetParametersRequest.MESSAGE_TYPE;

  canvasId: string = "";
  updates: ParameterUpdate[] = [];

  
}



export class BatchSetParametersResponse implements BatchSetParametersResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.BatchSetParametersResponse";
  readonly __MESSAGE_TYPE = BatchSetParametersResponse.MESSAGE_TYPE;

  success: boolean = false;
  errorMessage: string = "";
  results: ParameterUpdateResult[] = [];

  
}



export class EvaluateFlowsRequest implements EvaluateFlowsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.EvaluateFlowsRequest";
  readonly __MESSAGE_TYPE = EvaluateFlowsRequest.MESSAGE_TYPE;

  canvasId: string = "";
  strategy: string = "";

  
}



export class EvaluateFlowsResponse implements EvaluateFlowsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.EvaluateFlowsResponse";
  readonly __MESSAGE_TYPE = EvaluateFlowsResponse.MESSAGE_TYPE;

  strategy: string = "";
  status: string = "";
  iterations: number = 0;
  warnings: string[] = [];
  componentRates: Record<string, number> = {};
  flowEdges: FlowEdge[] = [];

  
}



export class GetFlowStateRequest implements GetFlowStateRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetFlowStateRequest";
  readonly __MESSAGE_TYPE = GetFlowStateRequest.MESSAGE_TYPE;

  canvasId: string = "";

  
}



export class GetFlowStateResponse implements GetFlowStateResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetFlowStateResponse";
  readonly __MESSAGE_TYPE = GetFlowStateResponse.MESSAGE_TYPE;

  state?: FlowState;

  
}



export class GetSystemDiagramRequest implements GetSystemDiagramRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetSystemDiagramRequest";
  readonly __MESSAGE_TYPE = GetSystemDiagramRequest.MESSAGE_TYPE;

  canvasId: string = "";

  
}



export class GetSystemDiagramResponse implements GetSystemDiagramResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetSystemDiagramResponse";
  readonly __MESSAGE_TYPE = GetSystemDiagramResponse.MESSAGE_TYPE;

  diagram?: SystemDiagram;

  
}



export class GetUtilizationRequest implements GetUtilizationRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetUtilizationRequest";
  readonly __MESSAGE_TYPE = GetUtilizationRequest.MESSAGE_TYPE;

  canvasId: string = "";
  components: string[] = [];

  
}



export class GetUtilizationResponse implements GetUtilizationResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetUtilizationResponse";
  readonly __MESSAGE_TYPE = GetUtilizationResponse.MESSAGE_TYPE;

  utilizations: UtilizationInfo[] = [];

  
}



export class UpdateMetricRequest implements UpdateMetricRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateMetricRequest";
  readonly __MESSAGE_TYPE = UpdateMetricRequest.MESSAGE_TYPE;

  metricId: string = "";
  point?: MetricPoint;

  
}



export class UpdateMetricResponse implements UpdateMetricResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateMetricResponse";
  readonly __MESSAGE_TYPE = UpdateMetricResponse.MESSAGE_TYPE;


  
}



export class ClearMetricsRequest implements ClearMetricsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ClearMetricsRequest";
  readonly __MESSAGE_TYPE = ClearMetricsRequest.MESSAGE_TYPE;


  
}



export class ClearMetricsResponse implements ClearMetricsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ClearMetricsResponse";
  readonly __MESSAGE_TYPE = ClearMetricsResponse.MESSAGE_TYPE;


  
}



export class SetMetricsListRequest implements SetMetricsListRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.SetMetricsListRequest";
  readonly __MESSAGE_TYPE = SetMetricsListRequest.MESSAGE_TYPE;

  metrics: Metric[] = [];

  
}



export class SetMetricsListResponse implements SetMetricsListResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.SetMetricsListResponse";
  readonly __MESSAGE_TYPE = SetMetricsListResponse.MESSAGE_TYPE;


  
}



export class UpdateDiagramRequest implements UpdateDiagramRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateDiagramRequest";
  readonly __MESSAGE_TYPE = UpdateDiagramRequest.MESSAGE_TYPE;

  diagram?: SystemDiagram;

  
}



export class UpdateDiagramResponse implements UpdateDiagramResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateDiagramResponse";
  readonly __MESSAGE_TYPE = UpdateDiagramResponse.MESSAGE_TYPE;


  
}



export class HighlightComponentsRequest implements HighlightComponentsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.HighlightComponentsRequest";
  readonly __MESSAGE_TYPE = HighlightComponentsRequest.MESSAGE_TYPE;

  componentIds: string[] = [];
  highlightType: string = "";
  color: string = "";

  
}



export class HighlightComponentsResponse implements HighlightComponentsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.HighlightComponentsResponse";
  readonly __MESSAGE_TYPE = HighlightComponentsResponse.MESSAGE_TYPE;


  
}



export class ClearHighlightsRequest implements ClearHighlightsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ClearHighlightsRequest";
  readonly __MESSAGE_TYPE = ClearHighlightsRequest.MESSAGE_TYPE;

  types: string[] = [];

  
}



export class ClearHighlightsResponse implements ClearHighlightsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ClearHighlightsResponse";
  readonly __MESSAGE_TYPE = ClearHighlightsResponse.MESSAGE_TYPE;


  
}



export class UpdateGeneratorStateRequest implements UpdateGeneratorStateRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateGeneratorStateRequest";
  readonly __MESSAGE_TYPE = UpdateGeneratorStateRequest.MESSAGE_TYPE;

  generatorId: string = "";
  enabled: boolean = false;
  rate: number = 0;
  status: string = "";

  
}



export class UpdateGeneratorStateResponse implements UpdateGeneratorStateResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateGeneratorStateResponse";
  readonly __MESSAGE_TYPE = UpdateGeneratorStateResponse.MESSAGE_TYPE;


  
}



export class SetGeneratorListRequest implements SetGeneratorListRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.SetGeneratorListRequest";
  readonly __MESSAGE_TYPE = SetGeneratorListRequest.MESSAGE_TYPE;

  generators: Generator[] = [];

  
}



export class SetGeneratorListResponse implements SetGeneratorListResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.SetGeneratorListResponse";
  readonly __MESSAGE_TYPE = SetGeneratorListResponse.MESSAGE_TYPE;


  
}



export class LogMessageRequest implements LogMessageRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.LogMessageRequest";
  readonly __MESSAGE_TYPE = LogMessageRequest.MESSAGE_TYPE;

  level: string = "";
  message: string = "";
  source: string = "";
  timestamp: number = 0;

  
}



export class LogMessageResponse implements LogMessageResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.LogMessageResponse";
  readonly __MESSAGE_TYPE = LogMessageResponse.MESSAGE_TYPE;


  
}



export class ClearConsoleRequest implements ClearConsoleRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ClearConsoleRequest";
  readonly __MESSAGE_TYPE = ClearConsoleRequest.MESSAGE_TYPE;


  
}



export class ClearConsoleResponse implements ClearConsoleResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ClearConsoleResponse";
  readonly __MESSAGE_TYPE = ClearConsoleResponse.MESSAGE_TYPE;


  
}



export class UpdateFlowRatesRequest implements UpdateFlowRatesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateFlowRatesRequest";
  readonly __MESSAGE_TYPE = UpdateFlowRatesRequest.MESSAGE_TYPE;

  rates: Record<string, number> = {};
  strategy: string = "";

  
}



export class UpdateFlowRatesResponse implements UpdateFlowRatesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateFlowRatesResponse";
  readonly __MESSAGE_TYPE = UpdateFlowRatesResponse.MESSAGE_TYPE;


  
}



export class ShowFlowPathRequest implements ShowFlowPathRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ShowFlowPathRequest";
  readonly __MESSAGE_TYPE = ShowFlowPathRequest.MESSAGE_TYPE;

  segments: FlowPathSegment[] = [];
  color: string = "";
  label: string = "";

  
}



export class FlowPathSegment implements FlowPathSegmentInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.FlowPathSegment";
  readonly __MESSAGE_TYPE = FlowPathSegment.MESSAGE_TYPE;

  fromComponent: string = "";
  fromMethod: string = "";
  toComponent: string = "";
  toMethod: string = "";
  rate: number = 0;

  
}



export class ShowFlowPathResponse implements ShowFlowPathResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ShowFlowPathResponse";
  readonly __MESSAGE_TYPE = ShowFlowPathResponse.MESSAGE_TYPE;


  
}



export class ClearFlowPathsRequest implements ClearFlowPathsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ClearFlowPathsRequest";
  readonly __MESSAGE_TYPE = ClearFlowPathsRequest.MESSAGE_TYPE;


  
}



export class ClearFlowPathsResponse implements ClearFlowPathsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ClearFlowPathsResponse";
  readonly __MESSAGE_TYPE = ClearFlowPathsResponse.MESSAGE_TYPE;


  
}



export class UpdateUtilizationRequest implements UpdateUtilizationRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateUtilizationRequest";
  readonly __MESSAGE_TYPE = UpdateUtilizationRequest.MESSAGE_TYPE;

  utilizations: UtilizationInfo[] = [];

  
}



export class UpdateUtilizationResponse implements UpdateUtilizationResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.UpdateUtilizationResponse";
  readonly __MESSAGE_TYPE = UpdateUtilizationResponse.MESSAGE_TYPE;


  
}


/**
 * FileInfo represents information about a file or directory
 */
export class FileInfo implements FileInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.FileInfo";
  readonly __MESSAGE_TYPE = FileInfo.MESSAGE_TYPE;

  name: string = "";
  path: string = "";
  isDirectory: boolean = false;
  size: number = 0;
  modTime: string = "";
  mimeType: string = "";

  
}


/**
 * FilesystemInfo represents information about a mounted filesystem
 */
export class FilesystemInfo implements FilesystemInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.FilesystemInfo";
  readonly __MESSAGE_TYPE = FilesystemInfo.MESSAGE_TYPE;

  id: string = "";
  prefix: string = "";
  type: string = "";
  readOnly: boolean = false;
  basePath: string = "";
  extensions: string[] = [];

  
}



export class ListFilesystemsRequest implements ListFilesystemsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListFilesystemsRequest";
  readonly __MESSAGE_TYPE = ListFilesystemsRequest.MESSAGE_TYPE;


  
}



export class ListFilesystemsResponse implements ListFilesystemsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListFilesystemsResponse";
  readonly __MESSAGE_TYPE = ListFilesystemsResponse.MESSAGE_TYPE;

  filesystems: FilesystemInfo[] = [];

  
}



export class ListFilesRequest implements ListFilesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListFilesRequest";
  readonly __MESSAGE_TYPE = ListFilesRequest.MESSAGE_TYPE;

  filesystemId: string = "";
  path: string = "";

  
}



export class ListFilesResponse implements ListFilesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListFilesResponse";
  readonly __MESSAGE_TYPE = ListFilesResponse.MESSAGE_TYPE;

  files: FileInfo[] = [];

  
}



export class ReadFileRequest implements ReadFileRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ReadFileRequest";
  readonly __MESSAGE_TYPE = ReadFileRequest.MESSAGE_TYPE;

  filesystemId: string = "";
  path: string = "";

  
}



export class ReadFileResponse implements ReadFileResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ReadFileResponse";
  readonly __MESSAGE_TYPE = ReadFileResponse.MESSAGE_TYPE;

  content: Uint8Array = new Uint8Array();
  fileInfo?: FileInfo;

  
}



export class WriteFileRequest implements WriteFileRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.WriteFileRequest";
  readonly __MESSAGE_TYPE = WriteFileRequest.MESSAGE_TYPE;

  filesystemId: string = "";
  path: string = "";
  content: Uint8Array = new Uint8Array();

  
}



export class WriteFileResponse implements WriteFileResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.WriteFileResponse";
  readonly __MESSAGE_TYPE = WriteFileResponse.MESSAGE_TYPE;

  fileInfo?: FileInfo;

  
}



export class DeleteFileRequest implements DeleteFileRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DeleteFileRequest";
  readonly __MESSAGE_TYPE = DeleteFileRequest.MESSAGE_TYPE;

  filesystemId: string = "";
  path: string = "";

  
}



export class DeleteFileResponse implements DeleteFileResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DeleteFileResponse";
  readonly __MESSAGE_TYPE = DeleteFileResponse.MESSAGE_TYPE;

  success: boolean = false;

  
}



export class CreateDirectoryRequest implements CreateDirectoryRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.CreateDirectoryRequest";
  readonly __MESSAGE_TYPE = CreateDirectoryRequest.MESSAGE_TYPE;

  filesystemId: string = "";
  path: string = "";

  
}



export class CreateDirectoryResponse implements CreateDirectoryResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.CreateDirectoryResponse";
  readonly __MESSAGE_TYPE = CreateDirectoryResponse.MESSAGE_TYPE;

  directoryInfo?: FileInfo;

  
}



export class GetFileInfoRequest implements GetFileInfoRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetFileInfoRequest";
  readonly __MESSAGE_TYPE = GetFileInfoRequest.MESSAGE_TYPE;

  filesystemId: string = "";
  path: string = "";

  
}



export class GetFileInfoResponse implements GetFileInfoResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetFileInfoResponse";
  readonly __MESSAGE_TYPE = GetFileInfoResponse.MESSAGE_TYPE;

  fileInfo?: FileInfo;

  
}


/**
 * SystemInfo represents a system in the catalog
 */
export class SystemInfo implements SystemInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.SystemInfo";
  readonly __MESSAGE_TYPE = SystemInfo.MESSAGE_TYPE;

  id: string = "";
  name: string = "";
  description: string = "";
  category: string = "";
  difficulty: string = "";
  tags: string[] = [];
  icon: string = "";
  lastUpdated: string = "";

  
}


/**
 * SystemProject represents a full system project
 */
export class SystemProject implements SystemProjectInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.SystemProject";
  readonly __MESSAGE_TYPE = SystemProject.MESSAGE_TYPE;

  id: string = "";
  name: string = "";
  description: string = "";
  category: string = "";
  difficulty: string = "";
  tags: string[] = [];
  icon: string = "";
  versions: Record<string, SystemVersion> = {};
  defaultVersion: string = "";
  lastUpdated: string = "";

  
}


/**
 * SystemVersion represents a version of a system
 */
export class SystemVersion implements SystemVersionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.SystemVersion";
  readonly __MESSAGE_TYPE = SystemVersion.MESSAGE_TYPE;

  sdl: string = "";
  recipe: string = "";
  readme: string = "";

  
}



export class ListSystemsRequest implements ListSystemsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListSystemsRequest";
  readonly __MESSAGE_TYPE = ListSystemsRequest.MESSAGE_TYPE;


  
}



export class ListSystemsResponse implements ListSystemsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ListSystemsResponse";
  readonly __MESSAGE_TYPE = ListSystemsResponse.MESSAGE_TYPE;

  systems: SystemInfo[] = [];

  
}



export class GetSystemRequest implements GetSystemRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetSystemRequest";
  readonly __MESSAGE_TYPE = GetSystemRequest.MESSAGE_TYPE;

  id: string = "";
  version: string = "";

  
}



export class GetSystemResponse implements GetSystemResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetSystemResponse";
  readonly __MESSAGE_TYPE = GetSystemResponse.MESSAGE_TYPE;

  system?: SystemProject;

  
}



export class GetSystemContentRequest implements GetSystemContentRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetSystemContentRequest";
  readonly __MESSAGE_TYPE = GetSystemContentRequest.MESSAGE_TYPE;

  id: string = "";
  version: string = "";

  
}



export class GetSystemContentResponse implements GetSystemContentResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.GetSystemContentResponse";
  readonly __MESSAGE_TYPE = GetSystemContentResponse.MESSAGE_TYPE;

  sdlContent: string = "";
  recipeContent: string = "";
  readmeContent: string = "";

  
}



export class InitializeSingletonRequest implements InitializeSingletonRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.InitializeSingletonRequest";
  readonly __MESSAGE_TYPE = InitializeSingletonRequest.MESSAGE_TYPE;

  canvasId: string = "";
  /** SDL content to load initially */
  sdlContent: string = "";
  /** System name to use after loading */
  systemName: string = "";
  /** JSON-encoded generator configs */
  generatorsData: string = "";
  /** JSON-encoded metric configs */
  metricsData: string = "";

  
}



export class InitializeSingletonResponse implements InitializeSingletonResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.InitializeSingletonResponse";
  readonly __MESSAGE_TYPE = InitializeSingletonResponse.MESSAGE_TYPE;

  success: boolean = false;
  error: string = "";
  canvasId: string = "";
  availableSystems: SystemInfo[] = [];

  
}



export class InitializePresenterRequest implements InitializePresenterRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.InitializePresenterRequest";
  readonly __MESSAGE_TYPE = InitializePresenterRequest.MESSAGE_TYPE;

  canvasId: string = "";

  
}



export class InitializePresenterResponse implements InitializePresenterResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.InitializePresenterResponse";
  readonly __MESSAGE_TYPE = InitializePresenterResponse.MESSAGE_TYPE;

  success: boolean = false;
  error: string = "";
  canvasId: string = "";
  /** Available systems to choose from */
  availableSystems: SystemInfo[] = [];
  /** Initial state to render */
  diagram?: SystemDiagram;
  generators: Generator[] = [];
  metrics: Metric[] = [];

  
}



export class ClientReadyRequest implements ClientReadyRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ClientReadyRequest";
  readonly __MESSAGE_TYPE = ClientReadyRequest.MESSAGE_TYPE;

  canvasId: string = "";

  
}



export class ClientReadyResponse implements ClientReadyResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.ClientReadyResponse";
  readonly __MESSAGE_TYPE = ClientReadyResponse.MESSAGE_TYPE;

  success: boolean = false;
  canvas?: Canvas;

  
}



export class FileSelectedRequest implements FileSelectedRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.FileSelectedRequest";
  readonly __MESSAGE_TYPE = FileSelectedRequest.MESSAGE_TYPE;

  canvasId: string = "";
  filePath: string = "";

  
}



export class FileSelectedResponse implements FileSelectedResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.FileSelectedResponse";
  readonly __MESSAGE_TYPE = FileSelectedResponse.MESSAGE_TYPE;

  success: boolean = false;
  content: string = "";
  error: string = "";

  
}



export class FileSavedRequest implements FileSavedRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.FileSavedRequest";
  readonly __MESSAGE_TYPE = FileSavedRequest.MESSAGE_TYPE;

  canvasId: string = "";
  filePath: string = "";
  content: string = "";

  
}



export class FileSavedResponse implements FileSavedResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.FileSavedResponse";
  readonly __MESSAGE_TYPE = FileSavedResponse.MESSAGE_TYPE;

  success: boolean = false;
  error: string = "";

  
}



export class DiagramComponentClickedRequest implements DiagramComponentClickedRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DiagramComponentClickedRequest";
  readonly __MESSAGE_TYPE = DiagramComponentClickedRequest.MESSAGE_TYPE;

  canvasId: string = "";
  componentName: string = "";
  methodName: string = "";

  
}



export class DiagramComponentClickedResponse implements DiagramComponentClickedResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DiagramComponentClickedResponse";
  readonly __MESSAGE_TYPE = DiagramComponentClickedResponse.MESSAGE_TYPE;

  success: boolean = false;

  
}



export class DiagramComponentHoveredRequest implements DiagramComponentHoveredRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DiagramComponentHoveredRequest";
  readonly __MESSAGE_TYPE = DiagramComponentHoveredRequest.MESSAGE_TYPE;

  canvasId: string = "";
  componentName: string = "";
  methodName: string = "";

  
}



export class DiagramComponentHoveredResponse implements DiagramComponentHoveredResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "sdl.v1.DiagramComponentHoveredResponse";
  readonly __MESSAGE_TYPE = DiagramComponentHoveredResponse.MESSAGE_TYPE;

  success: boolean = false;

  
}


