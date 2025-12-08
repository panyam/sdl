
// Generated TypeScript schemas from proto file
// DO NOT EDIT - This file is auto-generated

import { FieldType, FieldSchema, MessageSchema, BaseSchemaRegistry } from "@protoc-gen-go-wasmjs/runtime";


/**
 * Schema for Pagination message
 */
export const PaginationSchema: MessageSchema = {
  name: "Pagination",
  fields: [
    {
      name: "pageKey",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "pageOffset",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "pageSize",
      type: FieldType.NUMBER,
      id: 3,
    },
  ],
};


/**
 * Schema for PaginationResponse message
 */
export const PaginationResponseSchema: MessageSchema = {
  name: "PaginationResponse",
  fields: [
    {
      name: "nextPageKey",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "nextPageOffset",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "hasMore",
      type: FieldType.BOOLEAN,
      id: 4,
    },
    {
      name: "totalResults",
      type: FieldType.NUMBER,
      id: 5,
    },
  ],
};


/**
 * Schema for Canvas message
 */
export const CanvasSchema: MessageSchema = {
  name: "Canvas",
  fields: [
    {
      name: "createdAt",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "updatedAt",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "id",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "activeSystem",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "systemContents",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "recipes",
      type: FieldType.STRING,
      id: 8,
    },
    {
      name: "generators",
      type: FieldType.MESSAGE,
      id: 9,
      messageType: "sdl.v1.Generator",
      repeated: true,
    },
    {
      name: "metrics",
      type: FieldType.MESSAGE,
      id: 10,
      messageType: "sdl.v1.Metric",
      repeated: true,
    },
    {
      name: "previewUrl",
      type: FieldType.STRING,
      id: 11,
    },
  ],
};


/**
 * Schema for File message
 */
export const FileSchema: MessageSchema = {
  name: "File",
  fields: [
    {
      name: "path",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "contents",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for Generator message
 */
export const GeneratorSchema: MessageSchema = {
  name: "Generator",
  fields: [
    {
      name: "createdAt",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "updatedAt",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "id",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "component",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "method",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "rate",
      type: FieldType.NUMBER,
      id: 8,
    },
    {
      name: "duration",
      type: FieldType.NUMBER,
      id: 9,
    },
    {
      name: "enabled",
      type: FieldType.BOOLEAN,
      id: 10,
    },
  ],
};


/**
 * Schema for Metric message
 */
export const MetricSchema: MessageSchema = {
  name: "Metric",
  fields: [
    {
      name: "createdAt",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "updatedAt",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "id",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "component",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "methods",
      type: FieldType.REPEATED,
      id: 7,
      repeated: true,
    },
    {
      name: "enabled",
      type: FieldType.BOOLEAN,
      id: 8,
    },
    {
      name: "metricType",
      type: FieldType.STRING,
      id: 9,
    },
    {
      name: "aggregation",
      type: FieldType.STRING,
      id: 10,
    },
    {
      name: "aggregationWindow",
      type: FieldType.NUMBER,
      id: 11,
    },
    {
      name: "matchResult",
      type: FieldType.STRING,
      id: 12,
    },
    {
      name: "matchResultType",
      type: FieldType.STRING,
      id: 13,
    },
    {
      name: "oldestTimestamp",
      type: FieldType.NUMBER,
      id: 14,
    },
    {
      name: "newestTimestamp",
      type: FieldType.NUMBER,
      id: 15,
    },
    {
      name: "numDataPoints",
      type: FieldType.NUMBER,
      id: 16,
    },
  ],
};


/**
 * Schema for MetricPoint message
 */
export const MetricPointSchema: MessageSchema = {
  name: "MetricPoint",
  fields: [
    {
      name: "timestamp",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "value",
      type: FieldType.NUMBER,
      id: 2,
    },
  ],
};


/**
 * Schema for MetricUpdate message
 */
export const MetricUpdateSchema: MessageSchema = {
  name: "MetricUpdate",
  fields: [
    {
      name: "metricId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "point",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "sdl.v1.MetricPoint",
    },
  ],
};


/**
 * Schema for SystemDiagram message
 */
export const SystemDiagramSchema: MessageSchema = {
  name: "SystemDiagram",
  fields: [
    {
      name: "systemName",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "nodes",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "sdl.v1.DiagramNode",
      repeated: true,
    },
    {
      name: "edges",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "sdl.v1.DiagramEdge",
      repeated: true,
    },
  ],
};


/**
 * Schema for DiagramNode message
 */
export const DiagramNodeSchema: MessageSchema = {
  name: "DiagramNode",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "type",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "methods",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "sdl.v1.MethodInfo",
      repeated: true,
    },
    {
      name: "traffic",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "fullPath",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "icon",
      type: FieldType.STRING,
      id: 7,
    },
  ],
};


/**
 * Schema for MethodInfo message
 */
export const MethodInfoSchema: MessageSchema = {
  name: "MethodInfo",
  fields: [
    {
      name: "name",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "returnType",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "traffic",
      type: FieldType.NUMBER,
      id: 3,
    },
  ],
};


/**
 * Schema for DiagramEdge message
 */
export const DiagramEdgeSchema: MessageSchema = {
  name: "DiagramEdge",
  fields: [
    {
      name: "fromId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "toId",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "fromMethod",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "toMethod",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "label",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "order",
      type: FieldType.NUMBER,
      id: 6,
    },
    {
      name: "condition",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "probability",
      type: FieldType.NUMBER,
      id: 8,
    },
    {
      name: "generatorId",
      type: FieldType.STRING,
      id: 9,
    },
    {
      name: "color",
      type: FieldType.STRING,
      id: 10,
    },
  ],
};


/**
 * Schema for UtilizationInfo message
 */
export const UtilizationInfoSchema: MessageSchema = {
  name: "UtilizationInfo",
  fields: [
    {
      name: "resourceName",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "componentPath",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "utilization",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "capacity",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "currentLoad",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "isBottleneck",
      type: FieldType.BOOLEAN,
      id: 6,
    },
    {
      name: "warningThreshold",
      type: FieldType.NUMBER,
      id: 7,
    },
    {
      name: "criticalThreshold",
      type: FieldType.NUMBER,
      id: 8,
    },
  ],
};


/**
 * Schema for FlowEdge message
 */
export const FlowEdgeSchema: MessageSchema = {
  name: "FlowEdge",
  fields: [
    {
      name: "fromComponent",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "fromMethod",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "toComponent",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "toMethod",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "rate",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "condition",
      type: FieldType.STRING,
      id: 6,
    },
  ],
};


/**
 * Schema for FlowState message
 */
export const FlowStateSchema: MessageSchema = {
  name: "FlowState",
  fields: [
    {
      name: "strategy",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "rates",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "manualOverrides",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for TraceData message
 */
export const TraceDataSchema: MessageSchema = {
  name: "TraceData",
  fields: [
    {
      name: "system",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "entryPoint",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "events",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "sdl.v1.TraceEvent",
      repeated: true,
    },
  ],
};


/**
 * Schema for TraceEvent message
 */
export const TraceEventSchema: MessageSchema = {
  name: "TraceEvent",
  fields: [
    {
      name: "kind",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "id",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "parentId",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "timestamp",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "duration",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "component",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "method",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "args",
      type: FieldType.REPEATED,
      id: 8,
      repeated: true,
    },
    {
      name: "returnValue",
      type: FieldType.STRING,
      id: 9,
    },
    {
      name: "errorMessage",
      type: FieldType.STRING,
      id: 10,
    },
  ],
};


/**
 * Schema for AllPathsTraceData message
 */
export const AllPathsTraceDataSchema: MessageSchema = {
  name: "AllPathsTraceData",
  fields: [
    {
      name: "traceId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "root",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "sdl.v1.TraceNode",
    },
  ],
};


/**
 * Schema for TraceNode message
 */
export const TraceNodeSchema: MessageSchema = {
  name: "TraceNode",
  fields: [
    {
      name: "startingTarget",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "edges",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "sdl.v1.Edge",
      repeated: true,
    },
    {
      name: "groups",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "sdl.v1.GroupInfo",
      repeated: true,
    },
  ],
};


/**
 * Schema for Edge message
 */
export const EdgeSchema: MessageSchema = {
  name: "Edge",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "nextNode",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "sdl.v1.TraceNode",
    },
    {
      name: "label",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "isAsync",
      type: FieldType.BOOLEAN,
      id: 4,
    },
    {
      name: "isReverse",
      type: FieldType.BOOLEAN,
      id: 5,
    },
    {
      name: "probability",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "condition",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "isConditional",
      type: FieldType.BOOLEAN,
      id: 8,
    },
  ],
};


/**
 * Schema for GroupInfo message
 */
export const GroupInfoSchema: MessageSchema = {
  name: "GroupInfo",
  fields: [
    {
      name: "groupStart",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "groupEnd",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "groupLabel",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "groupType",
      type: FieldType.STRING,
      id: 4,
    },
  ],
};


/**
 * Schema for ParameterUpdate message
 */
export const ParameterUpdateSchema: MessageSchema = {
  name: "ParameterUpdate",
  fields: [
    {
      name: "path",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "newValue",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for ParameterUpdateResult message
 */
export const ParameterUpdateResultSchema: MessageSchema = {
  name: "ParameterUpdateResult",
  fields: [
    {
      name: "path",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 2,
    },
    {
      name: "errorMessage",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "oldValue",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "newValue",
      type: FieldType.STRING,
      id: 5,
    },
  ],
};


/**
 * Schema for AggregateResult message
 */
export const AggregateResultSchema: MessageSchema = {
  name: "AggregateResult",
  fields: [
    {
      name: "timestamp",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "value",
      type: FieldType.NUMBER,
      id: 2,
    },
  ],
};


/**
 * Schema for CreateCanvasRequest message
 */
export const CreateCanvasRequestSchema: MessageSchema = {
  name: "CreateCanvasRequest",
  fields: [
    {
      name: "canvas",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Canvas",
    },
  ],
};


/**
 * Schema for CreateCanvasResponse message
 */
export const CreateCanvasResponseSchema: MessageSchema = {
  name: "CreateCanvasResponse",
  fields: [
    {
      name: "canvas",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Canvas",
    },
    {
      name: "fieldErrors",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for UpdateCanvasRequest message
 */
export const UpdateCanvasRequestSchema: MessageSchema = {
  name: "UpdateCanvasRequest",
  fields: [
    {
      name: "canvas",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Canvas",
    },
  ],
};


/**
 * Schema for UpdateCanvasResponse message
 */
export const UpdateCanvasResponseSchema: MessageSchema = {
  name: "UpdateCanvasResponse",
  fields: [
    {
      name: "canvas",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Canvas",
    },
    {
      name: "updateMask",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.FieldMask",
    },
    {
      name: "deletedFiles",
      type: FieldType.REPEATED,
      id: 3,
      repeated: true,
    },
    {
      name: "updatedFiles",
      type: FieldType.STRING,
      id: 4,
    },
  ],
};


/**
 * Schema for ListCanvasesRequest message
 */
export const ListCanvasesRequestSchema: MessageSchema = {
  name: "ListCanvasesRequest",
  fields: [
    {
      name: "pagination",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Pagination",
    },
  ],
};


/**
 * Schema for ListCanvasesResponse message
 */
export const ListCanvasesResponseSchema: MessageSchema = {
  name: "ListCanvasesResponse",
  fields: [
    {
      name: "canvases",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Canvas",
      repeated: true,
    },
    {
      name: "pagination",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "sdl.v1.PaginationResponse",
    },
  ],
};


/**
 * Schema for GetCanvasRequest message
 */
export const GetCanvasRequestSchema: MessageSchema = {
  name: "GetCanvasRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for GetCanvasResponse message
 */
export const GetCanvasResponseSchema: MessageSchema = {
  name: "GetCanvasResponse",
  fields: [
    {
      name: "canvas",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Canvas",
    },
  ],
};


/**
 * Schema for DeleteCanvasRequest message
 */
export const DeleteCanvasRequestSchema: MessageSchema = {
  name: "DeleteCanvasRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for DeleteCanvasResponse message
 */
export const DeleteCanvasResponseSchema: MessageSchema = {
  name: "DeleteCanvasResponse",
  fields: [
  ],
};


/**
 * Schema for ResetCanvasRequest message
 */
export const ResetCanvasRequestSchema: MessageSchema = {
  name: "ResetCanvasRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for ResetCanvasResponse message
 */
export const ResetCanvasResponseSchema: MessageSchema = {
  name: "ResetCanvasResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "message",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for LoadFileRequest message
 */
export const LoadFileRequestSchema: MessageSchema = {
  name: "LoadFileRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "sdlFilePath",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for LoadFileResponse message
 */
export const LoadFileResponseSchema: MessageSchema = {
  name: "LoadFileResponse",
  fields: [
  ],
};


/**
 * Schema for UseSystemRequest message
 */
export const UseSystemRequestSchema: MessageSchema = {
  name: "UseSystemRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "systemName",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for UseSystemResponse message
 */
export const UseSystemResponseSchema: MessageSchema = {
  name: "UseSystemResponse",
  fields: [
  ],
};


/**
 * Schema for AddGeneratorRequest message
 */
export const AddGeneratorRequestSchema: MessageSchema = {
  name: "AddGeneratorRequest",
  fields: [
    {
      name: "generator",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Generator",
    },
    {
      name: "applyFlows",
      type: FieldType.BOOLEAN,
      id: 2,
    },
  ],
};


/**
 * Schema for AddGeneratorResponse message
 */
export const AddGeneratorResponseSchema: MessageSchema = {
  name: "AddGeneratorResponse",
  fields: [
    {
      name: "generator",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Generator",
    },
  ],
};


/**
 * Schema for ListGeneratorsRequest message
 */
export const ListGeneratorsRequestSchema: MessageSchema = {
  name: "ListGeneratorsRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for ListGeneratorsResponse message
 */
export const ListGeneratorsResponseSchema: MessageSchema = {
  name: "ListGeneratorsResponse",
  fields: [
    {
      name: "generators",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Generator",
      repeated: true,
    },
  ],
};


/**
 * Schema for GetGeneratorRequest message
 */
export const GetGeneratorRequestSchema: MessageSchema = {
  name: "GetGeneratorRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "generatorId",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for GetGeneratorResponse message
 */
export const GetGeneratorResponseSchema: MessageSchema = {
  name: "GetGeneratorResponse",
  fields: [
    {
      name: "generator",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Generator",
    },
  ],
};


/**
 * Schema for UpdateGeneratorRequest message
 */
export const UpdateGeneratorRequestSchema: MessageSchema = {
  name: "UpdateGeneratorRequest",
  fields: [
    {
      name: "generator",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Generator",
    },
    {
      name: "updateMask",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.FieldMask",
    },
    {
      name: "applyFlows",
      type: FieldType.BOOLEAN,
      id: 3,
    },
  ],
};


/**
 * Schema for UpdateGeneratorResponse message
 */
export const UpdateGeneratorResponseSchema: MessageSchema = {
  name: "UpdateGeneratorResponse",
  fields: [
    {
      name: "generator",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Generator",
    },
  ],
};


/**
 * Schema for StartGeneratorRequest message
 */
export const StartGeneratorRequestSchema: MessageSchema = {
  name: "StartGeneratorRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "generatorId",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for StartGeneratorResponse message
 */
export const StartGeneratorResponseSchema: MessageSchema = {
  name: "StartGeneratorResponse",
  fields: [
  ],
};


/**
 * Schema for StopGeneratorRequest message
 */
export const StopGeneratorRequestSchema: MessageSchema = {
  name: "StopGeneratorRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "generatorId",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for StopGeneratorResponse message
 */
export const StopGeneratorResponseSchema: MessageSchema = {
  name: "StopGeneratorResponse",
  fields: [
  ],
};


/**
 * Schema for DeleteGeneratorRequest message
 */
export const DeleteGeneratorRequestSchema: MessageSchema = {
  name: "DeleteGeneratorRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "generatorId",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "applyFlows",
      type: FieldType.BOOLEAN,
      id: 3,
    },
  ],
};


/**
 * Schema for DeleteGeneratorResponse message
 */
export const DeleteGeneratorResponseSchema: MessageSchema = {
  name: "DeleteGeneratorResponse",
  fields: [
  ],
};


/**
 * Schema for StartAllGeneratorsRequest message
 */
export const StartAllGeneratorsRequestSchema: MessageSchema = {
  name: "StartAllGeneratorsRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for StartAllGeneratorsResponse message
 */
export const StartAllGeneratorsResponseSchema: MessageSchema = {
  name: "StartAllGeneratorsResponse",
  fields: [
    {
      name: "totalGenerators",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "startedCount",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "alreadyRunningCount",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "failedCount",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "failedIds",
      type: FieldType.REPEATED,
      id: 5,
      repeated: true,
    },
  ],
};


/**
 * Schema for StopAllGeneratorsRequest message
 */
export const StopAllGeneratorsRequestSchema: MessageSchema = {
  name: "StopAllGeneratorsRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for StopAllGeneratorsResponse message
 */
export const StopAllGeneratorsResponseSchema: MessageSchema = {
  name: "StopAllGeneratorsResponse",
  fields: [
    {
      name: "totalGenerators",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "stoppedCount",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "alreadyStoppedCount",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "failedCount",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "failedIds",
      type: FieldType.REPEATED,
      id: 5,
      repeated: true,
    },
  ],
};


/**
 * Schema for AddMetricRequest message
 */
export const AddMetricRequestSchema: MessageSchema = {
  name: "AddMetricRequest",
  fields: [
    {
      name: "metric",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Metric",
    },
  ],
};


/**
 * Schema for AddMetricResponse message
 */
export const AddMetricResponseSchema: MessageSchema = {
  name: "AddMetricResponse",
  fields: [
    {
      name: "metric",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Metric",
    },
  ],
};


/**
 * Schema for DeleteMetricRequest message
 */
export const DeleteMetricRequestSchema: MessageSchema = {
  name: "DeleteMetricRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "metricId",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for DeleteMetricResponse message
 */
export const DeleteMetricResponseSchema: MessageSchema = {
  name: "DeleteMetricResponse",
  fields: [
  ],
};


/**
 * Schema for ListMetricsRequest message
 */
export const ListMetricsRequestSchema: MessageSchema = {
  name: "ListMetricsRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for ListMetricsResponse message
 */
export const ListMetricsResponseSchema: MessageSchema = {
  name: "ListMetricsResponse",
  fields: [
    {
      name: "metrics",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Metric",
      repeated: true,
    },
  ],
};


/**
 * Schema for QueryMetricsRequest message
 */
export const QueryMetricsRequestSchema: MessageSchema = {
  name: "QueryMetricsRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "metricId",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "startTime",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "endTime",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "limit",
      type: FieldType.NUMBER,
      id: 5,
    },
  ],
};


/**
 * Schema for QueryMetricsResponse message
 */
export const QueryMetricsResponseSchema: MessageSchema = {
  name: "QueryMetricsResponse",
  fields: [
    {
      name: "points",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.MetricPoint",
      repeated: true,
    },
  ],
};


/**
 * Schema for AggregateMetricsRequest message
 */
export const AggregateMetricsRequestSchema: MessageSchema = {
  name: "AggregateMetricsRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "metricId",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "startTime",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "endTime",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "function",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "windowSize",
      type: FieldType.NUMBER,
      id: 6,
    },
  ],
};


/**
 * Schema for AggregateMetricsResponse message
 */
export const AggregateMetricsResponseSchema: MessageSchema = {
  name: "AggregateMetricsResponse",
  fields: [
    {
      name: "results",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.AggregateResult",
      repeated: true,
    },
  ],
};


/**
 * Schema for StreamMetricsRequest message
 */
export const StreamMetricsRequestSchema: MessageSchema = {
  name: "StreamMetricsRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "metricIds",
      type: FieldType.REPEATED,
      id: 2,
      repeated: true,
    },
  ],
};


/**
 * Schema for StreamMetricsResponse message
 */
export const StreamMetricsResponseSchema: MessageSchema = {
  name: "StreamMetricsResponse",
  fields: [
    {
      name: "updates",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.MetricUpdate",
      repeated: true,
    },
  ],
};


/**
 * Schema for ExecuteTraceRequest message
 */
export const ExecuteTraceRequestSchema: MessageSchema = {
  name: "ExecuteTraceRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "component",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "method",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for ExecuteTraceResponse message
 */
export const ExecuteTraceResponseSchema: MessageSchema = {
  name: "ExecuteTraceResponse",
  fields: [
    {
      name: "traceData",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.TraceData",
    },
  ],
};


/**
 * Schema for TraceAllPathsRequest message
 */
export const TraceAllPathsRequestSchema: MessageSchema = {
  name: "TraceAllPathsRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "component",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "method",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "maxDepth",
      type: FieldType.NUMBER,
      id: 4,
    },
  ],
};


/**
 * Schema for TraceAllPathsResponse message
 */
export const TraceAllPathsResponseSchema: MessageSchema = {
  name: "TraceAllPathsResponse",
  fields: [
    {
      name: "traceData",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.AllPathsTraceData",
    },
  ],
};


/**
 * Schema for SetParameterRequest message
 */
export const SetParameterRequestSchema: MessageSchema = {
  name: "SetParameterRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "path",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "newValue",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for SetParameterResponse message
 */
export const SetParameterResponseSchema: MessageSchema = {
  name: "SetParameterResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "errorMessage",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "newValue",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "oldValue",
      type: FieldType.STRING,
      id: 4,
    },
  ],
};


/**
 * Schema for GetParametersRequest message
 */
export const GetParametersRequestSchema: MessageSchema = {
  name: "GetParametersRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "path",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for GetParametersResponse message
 */
export const GetParametersResponseSchema: MessageSchema = {
  name: "GetParametersResponse",
  fields: [
    {
      name: "parameters",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for BatchSetParametersRequest message
 */
export const BatchSetParametersRequestSchema: MessageSchema = {
  name: "BatchSetParametersRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "updates",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "sdl.v1.ParameterUpdate",
      repeated: true,
    },
  ],
};


/**
 * Schema for BatchSetParametersResponse message
 */
export const BatchSetParametersResponseSchema: MessageSchema = {
  name: "BatchSetParametersResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "errorMessage",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "results",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "sdl.v1.ParameterUpdateResult",
      repeated: true,
    },
  ],
};


/**
 * Schema for EvaluateFlowsRequest message
 */
export const EvaluateFlowsRequestSchema: MessageSchema = {
  name: "EvaluateFlowsRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "strategy",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for EvaluateFlowsResponse message
 */
export const EvaluateFlowsResponseSchema: MessageSchema = {
  name: "EvaluateFlowsResponse",
  fields: [
    {
      name: "strategy",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "status",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "iterations",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "warnings",
      type: FieldType.REPEATED,
      id: 4,
      repeated: true,
    },
    {
      name: "componentRates",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "flowEdges",
      type: FieldType.MESSAGE,
      id: 6,
      messageType: "sdl.v1.FlowEdge",
      repeated: true,
    },
  ],
};


/**
 * Schema for GetFlowStateRequest message
 */
export const GetFlowStateRequestSchema: MessageSchema = {
  name: "GetFlowStateRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for GetFlowStateResponse message
 */
export const GetFlowStateResponseSchema: MessageSchema = {
  name: "GetFlowStateResponse",
  fields: [
    {
      name: "state",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.FlowState",
    },
  ],
};


/**
 * Schema for GetSystemDiagramRequest message
 */
export const GetSystemDiagramRequestSchema: MessageSchema = {
  name: "GetSystemDiagramRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for GetSystemDiagramResponse message
 */
export const GetSystemDiagramResponseSchema: MessageSchema = {
  name: "GetSystemDiagramResponse",
  fields: [
    {
      name: "diagram",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.SystemDiagram",
    },
  ],
};


/**
 * Schema for GetUtilizationRequest message
 */
export const GetUtilizationRequestSchema: MessageSchema = {
  name: "GetUtilizationRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "components",
      type: FieldType.REPEATED,
      id: 2,
      repeated: true,
    },
  ],
};


/**
 * Schema for GetUtilizationResponse message
 */
export const GetUtilizationResponseSchema: MessageSchema = {
  name: "GetUtilizationResponse",
  fields: [
    {
      name: "utilizations",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.UtilizationInfo",
      repeated: true,
    },
  ],
};


/**
 * Schema for UpdateMetricRequest message
 */
export const UpdateMetricRequestSchema: MessageSchema = {
  name: "UpdateMetricRequest",
  fields: [
    {
      name: "metricId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "point",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "sdl.v1.MetricPoint",
    },
  ],
};


/**
 * Schema for UpdateMetricResponse message
 */
export const UpdateMetricResponseSchema: MessageSchema = {
  name: "UpdateMetricResponse",
  fields: [
  ],
};


/**
 * Schema for ClearMetricsRequest message
 */
export const ClearMetricsRequestSchema: MessageSchema = {
  name: "ClearMetricsRequest",
  fields: [
  ],
};


/**
 * Schema for ClearMetricsResponse message
 */
export const ClearMetricsResponseSchema: MessageSchema = {
  name: "ClearMetricsResponse",
  fields: [
  ],
};


/**
 * Schema for SetMetricsListRequest message
 */
export const SetMetricsListRequestSchema: MessageSchema = {
  name: "SetMetricsListRequest",
  fields: [
    {
      name: "metrics",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Metric",
      repeated: true,
    },
  ],
};


/**
 * Schema for SetMetricsListResponse message
 */
export const SetMetricsListResponseSchema: MessageSchema = {
  name: "SetMetricsListResponse",
  fields: [
  ],
};


/**
 * Schema for UpdateDiagramRequest message
 */
export const UpdateDiagramRequestSchema: MessageSchema = {
  name: "UpdateDiagramRequest",
  fields: [
    {
      name: "diagram",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.SystemDiagram",
    },
  ],
};


/**
 * Schema for UpdateDiagramResponse message
 */
export const UpdateDiagramResponseSchema: MessageSchema = {
  name: "UpdateDiagramResponse",
  fields: [
  ],
};


/**
 * Schema for HighlightComponentsRequest message
 */
export const HighlightComponentsRequestSchema: MessageSchema = {
  name: "HighlightComponentsRequest",
  fields: [
    {
      name: "componentIds",
      type: FieldType.REPEATED,
      id: 1,
      repeated: true,
    },
    {
      name: "highlightType",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "color",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for HighlightComponentsResponse message
 */
export const HighlightComponentsResponseSchema: MessageSchema = {
  name: "HighlightComponentsResponse",
  fields: [
  ],
};


/**
 * Schema for ClearHighlightsRequest message
 */
export const ClearHighlightsRequestSchema: MessageSchema = {
  name: "ClearHighlightsRequest",
  fields: [
    {
      name: "types",
      type: FieldType.REPEATED,
      id: 1,
      repeated: true,
    },
  ],
};


/**
 * Schema for ClearHighlightsResponse message
 */
export const ClearHighlightsResponseSchema: MessageSchema = {
  name: "ClearHighlightsResponse",
  fields: [
  ],
};


/**
 * Schema for UpdateGeneratorStateRequest message
 */
export const UpdateGeneratorStateRequestSchema: MessageSchema = {
  name: "UpdateGeneratorStateRequest",
  fields: [
    {
      name: "generatorId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "enabled",
      type: FieldType.BOOLEAN,
      id: 2,
    },
    {
      name: "rate",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "status",
      type: FieldType.STRING,
      id: 4,
    },
  ],
};


/**
 * Schema for UpdateGeneratorStateResponse message
 */
export const UpdateGeneratorStateResponseSchema: MessageSchema = {
  name: "UpdateGeneratorStateResponse",
  fields: [
  ],
};


/**
 * Schema for SetGeneratorListRequest message
 */
export const SetGeneratorListRequestSchema: MessageSchema = {
  name: "SetGeneratorListRequest",
  fields: [
    {
      name: "generators",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.Generator",
      repeated: true,
    },
  ],
};


/**
 * Schema for SetGeneratorListResponse message
 */
export const SetGeneratorListResponseSchema: MessageSchema = {
  name: "SetGeneratorListResponse",
  fields: [
  ],
};


/**
 * Schema for LogMessageRequest message
 */
export const LogMessageRequestSchema: MessageSchema = {
  name: "LogMessageRequest",
  fields: [
    {
      name: "level",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "message",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "source",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "timestamp",
      type: FieldType.NUMBER,
      id: 4,
    },
  ],
};


/**
 * Schema for LogMessageResponse message
 */
export const LogMessageResponseSchema: MessageSchema = {
  name: "LogMessageResponse",
  fields: [
  ],
};


/**
 * Schema for ClearConsoleRequest message
 */
export const ClearConsoleRequestSchema: MessageSchema = {
  name: "ClearConsoleRequest",
  fields: [
  ],
};


/**
 * Schema for ClearConsoleResponse message
 */
export const ClearConsoleResponseSchema: MessageSchema = {
  name: "ClearConsoleResponse",
  fields: [
  ],
};


/**
 * Schema for UpdateFlowRatesRequest message
 */
export const UpdateFlowRatesRequestSchema: MessageSchema = {
  name: "UpdateFlowRatesRequest",
  fields: [
    {
      name: "rates",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "strategy",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for UpdateFlowRatesResponse message
 */
export const UpdateFlowRatesResponseSchema: MessageSchema = {
  name: "UpdateFlowRatesResponse",
  fields: [
  ],
};


/**
 * Schema for ShowFlowPathRequest message
 */
export const ShowFlowPathRequestSchema: MessageSchema = {
  name: "ShowFlowPathRequest",
  fields: [
    {
      name: "segments",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.FlowPathSegment",
      repeated: true,
    },
    {
      name: "color",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "label",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for FlowPathSegment message
 */
export const FlowPathSegmentSchema: MessageSchema = {
  name: "FlowPathSegment",
  fields: [
    {
      name: "fromComponent",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "fromMethod",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "toComponent",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "toMethod",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "rate",
      type: FieldType.NUMBER,
      id: 5,
    },
  ],
};


/**
 * Schema for ShowFlowPathResponse message
 */
export const ShowFlowPathResponseSchema: MessageSchema = {
  name: "ShowFlowPathResponse",
  fields: [
  ],
};


/**
 * Schema for ClearFlowPathsRequest message
 */
export const ClearFlowPathsRequestSchema: MessageSchema = {
  name: "ClearFlowPathsRequest",
  fields: [
  ],
};


/**
 * Schema for ClearFlowPathsResponse message
 */
export const ClearFlowPathsResponseSchema: MessageSchema = {
  name: "ClearFlowPathsResponse",
  fields: [
  ],
};


/**
 * Schema for UpdateUtilizationRequest message
 */
export const UpdateUtilizationRequestSchema: MessageSchema = {
  name: "UpdateUtilizationRequest",
  fields: [
    {
      name: "utilizations",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.UtilizationInfo",
      repeated: true,
    },
  ],
};


/**
 * Schema for UpdateUtilizationResponse message
 */
export const UpdateUtilizationResponseSchema: MessageSchema = {
  name: "UpdateUtilizationResponse",
  fields: [
  ],
};


/**
 * Schema for FileInfo message
 */
export const FileInfoSchema: MessageSchema = {
  name: "FileInfo",
  fields: [
    {
      name: "name",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "path",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "isDirectory",
      type: FieldType.BOOLEAN,
      id: 3,
    },
    {
      name: "size",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "modTime",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "mimeType",
      type: FieldType.STRING,
      id: 6,
    },
  ],
};


/**
 * Schema for FilesystemInfo message
 */
export const FilesystemInfoSchema: MessageSchema = {
  name: "FilesystemInfo",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "prefix",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "type",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "readOnly",
      type: FieldType.BOOLEAN,
      id: 4,
    },
    {
      name: "basePath",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "extensions",
      type: FieldType.REPEATED,
      id: 6,
      repeated: true,
    },
  ],
};


/**
 * Schema for ListFilesystemsRequest message
 */
export const ListFilesystemsRequestSchema: MessageSchema = {
  name: "ListFilesystemsRequest",
  fields: [
  ],
};


/**
 * Schema for ListFilesystemsResponse message
 */
export const ListFilesystemsResponseSchema: MessageSchema = {
  name: "ListFilesystemsResponse",
  fields: [
    {
      name: "filesystems",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.FilesystemInfo",
      repeated: true,
    },
  ],
};


/**
 * Schema for ListFilesRequest message
 */
export const ListFilesRequestSchema: MessageSchema = {
  name: "ListFilesRequest",
  fields: [
    {
      name: "filesystemId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "path",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for ListFilesResponse message
 */
export const ListFilesResponseSchema: MessageSchema = {
  name: "ListFilesResponse",
  fields: [
    {
      name: "files",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.FileInfo",
      repeated: true,
    },
  ],
};


/**
 * Schema for ReadFileRequest message
 */
export const ReadFileRequestSchema: MessageSchema = {
  name: "ReadFileRequest",
  fields: [
    {
      name: "filesystemId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "path",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for ReadFileResponse message
 */
export const ReadFileResponseSchema: MessageSchema = {
  name: "ReadFileResponse",
  fields: [
    {
      name: "content",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "fileInfo",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "sdl.v1.FileInfo",
    },
  ],
};


/**
 * Schema for WriteFileRequest message
 */
export const WriteFileRequestSchema: MessageSchema = {
  name: "WriteFileRequest",
  fields: [
    {
      name: "filesystemId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "path",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "content",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for WriteFileResponse message
 */
export const WriteFileResponseSchema: MessageSchema = {
  name: "WriteFileResponse",
  fields: [
    {
      name: "fileInfo",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.FileInfo",
    },
  ],
};


/**
 * Schema for DeleteFileRequest message
 */
export const DeleteFileRequestSchema: MessageSchema = {
  name: "DeleteFileRequest",
  fields: [
    {
      name: "filesystemId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "path",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for DeleteFileResponse message
 */
export const DeleteFileResponseSchema: MessageSchema = {
  name: "DeleteFileResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
  ],
};


/**
 * Schema for CreateDirectoryRequest message
 */
export const CreateDirectoryRequestSchema: MessageSchema = {
  name: "CreateDirectoryRequest",
  fields: [
    {
      name: "filesystemId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "path",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for CreateDirectoryResponse message
 */
export const CreateDirectoryResponseSchema: MessageSchema = {
  name: "CreateDirectoryResponse",
  fields: [
    {
      name: "directoryInfo",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.FileInfo",
    },
  ],
};


/**
 * Schema for GetFileInfoRequest message
 */
export const GetFileInfoRequestSchema: MessageSchema = {
  name: "GetFileInfoRequest",
  fields: [
    {
      name: "filesystemId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "path",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for GetFileInfoResponse message
 */
export const GetFileInfoResponseSchema: MessageSchema = {
  name: "GetFileInfoResponse",
  fields: [
    {
      name: "fileInfo",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.FileInfo",
    },
  ],
};


/**
 * Schema for SystemInfo message
 */
export const SystemInfoSchema: MessageSchema = {
  name: "SystemInfo",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "category",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "difficulty",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "tags",
      type: FieldType.REPEATED,
      id: 6,
      repeated: true,
    },
    {
      name: "icon",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "lastUpdated",
      type: FieldType.STRING,
      id: 8,
    },
  ],
};


/**
 * Schema for SystemProject message
 */
export const SystemProjectSchema: MessageSchema = {
  name: "SystemProject",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "category",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "difficulty",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "tags",
      type: FieldType.REPEATED,
      id: 6,
      repeated: true,
    },
    {
      name: "icon",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "versions",
      type: FieldType.STRING,
      id: 8,
    },
    {
      name: "defaultVersion",
      type: FieldType.STRING,
      id: 9,
    },
    {
      name: "lastUpdated",
      type: FieldType.STRING,
      id: 10,
    },
  ],
};


/**
 * Schema for SystemVersion message
 */
export const SystemVersionSchema: MessageSchema = {
  name: "SystemVersion",
  fields: [
    {
      name: "sdl",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "recipe",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "readme",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for ListSystemsRequest message
 */
export const ListSystemsRequestSchema: MessageSchema = {
  name: "ListSystemsRequest",
  fields: [
  ],
};


/**
 * Schema for ListSystemsResponse message
 */
export const ListSystemsResponseSchema: MessageSchema = {
  name: "ListSystemsResponse",
  fields: [
    {
      name: "systems",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.SystemInfo",
      repeated: true,
    },
  ],
};


/**
 * Schema for GetSystemRequest message
 */
export const GetSystemRequestSchema: MessageSchema = {
  name: "GetSystemRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "version",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for GetSystemResponse message
 */
export const GetSystemResponseSchema: MessageSchema = {
  name: "GetSystemResponse",
  fields: [
    {
      name: "system",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "sdl.v1.SystemProject",
    },
  ],
};


/**
 * Schema for GetSystemContentRequest message
 */
export const GetSystemContentRequestSchema: MessageSchema = {
  name: "GetSystemContentRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "version",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for GetSystemContentResponse message
 */
export const GetSystemContentResponseSchema: MessageSchema = {
  name: "GetSystemContentResponse",
  fields: [
    {
      name: "sdlContent",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "recipeContent",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "readmeContent",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for InitializeSingletonRequest message
 */
export const InitializeSingletonRequestSchema: MessageSchema = {
  name: "InitializeSingletonRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "sdlContent",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "systemName",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "generatorsData",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "metricsData",
      type: FieldType.STRING,
      id: 5,
    },
  ],
};


/**
 * Schema for InitializeSingletonResponse message
 */
export const InitializeSingletonResponseSchema: MessageSchema = {
  name: "InitializeSingletonResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "error",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "availableSystems",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "sdl.v1.SystemInfo",
      repeated: true,
    },
  ],
};


/**
 * Schema for InitializePresenterRequest message
 */
export const InitializePresenterRequestSchema: MessageSchema = {
  name: "InitializePresenterRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for InitializePresenterResponse message
 */
export const InitializePresenterResponseSchema: MessageSchema = {
  name: "InitializePresenterResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "error",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "availableSystems",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "sdl.v1.SystemInfo",
      repeated: true,
    },
    {
      name: "diagram",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "sdl.v1.SystemDiagram",
    },
    {
      name: "generators",
      type: FieldType.MESSAGE,
      id: 6,
      messageType: "sdl.v1.Generator",
      repeated: true,
    },
    {
      name: "metrics",
      type: FieldType.MESSAGE,
      id: 7,
      messageType: "sdl.v1.Metric",
      repeated: true,
    },
  ],
};


/**
 * Schema for ClientReadyRequest message
 */
export const ClientReadyRequestSchema: MessageSchema = {
  name: "ClientReadyRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for ClientReadyResponse message
 */
export const ClientReadyResponseSchema: MessageSchema = {
  name: "ClientReadyResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "canvas",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "sdl.v1.Canvas",
    },
  ],
};


/**
 * Schema for FileSelectedRequest message
 */
export const FileSelectedRequestSchema: MessageSchema = {
  name: "FileSelectedRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "filePath",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for FileSelectedResponse message
 */
export const FileSelectedResponseSchema: MessageSchema = {
  name: "FileSelectedResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "content",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "error",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for FileSavedRequest message
 */
export const FileSavedRequestSchema: MessageSchema = {
  name: "FileSavedRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "filePath",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "content",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for FileSavedResponse message
 */
export const FileSavedResponseSchema: MessageSchema = {
  name: "FileSavedResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "error",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for DiagramComponentClickedRequest message
 */
export const DiagramComponentClickedRequestSchema: MessageSchema = {
  name: "DiagramComponentClickedRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "componentName",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "methodName",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for DiagramComponentClickedResponse message
 */
export const DiagramComponentClickedResponseSchema: MessageSchema = {
  name: "DiagramComponentClickedResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
  ],
};


/**
 * Schema for DiagramComponentHoveredRequest message
 */
export const DiagramComponentHoveredRequestSchema: MessageSchema = {
  name: "DiagramComponentHoveredRequest",
  fields: [
    {
      name: "canvasId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "componentName",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "methodName",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for DiagramComponentHoveredResponse message
 */
export const DiagramComponentHoveredResponseSchema: MessageSchema = {
  name: "DiagramComponentHoveredResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
  ],
};



/**
 * Package-scoped schema registry for sdl.v1
 */
export const sdl_v1SchemaRegistry: Record<string, MessageSchema> = {
  "sdl.v1.Pagination": PaginationSchema,
  "sdl.v1.PaginationResponse": PaginationResponseSchema,
  "sdl.v1.Canvas": CanvasSchema,
  "sdl.v1.File": FileSchema,
  "sdl.v1.Generator": GeneratorSchema,
  "sdl.v1.Metric": MetricSchema,
  "sdl.v1.MetricPoint": MetricPointSchema,
  "sdl.v1.MetricUpdate": MetricUpdateSchema,
  "sdl.v1.SystemDiagram": SystemDiagramSchema,
  "sdl.v1.DiagramNode": DiagramNodeSchema,
  "sdl.v1.MethodInfo": MethodInfoSchema,
  "sdl.v1.DiagramEdge": DiagramEdgeSchema,
  "sdl.v1.UtilizationInfo": UtilizationInfoSchema,
  "sdl.v1.FlowEdge": FlowEdgeSchema,
  "sdl.v1.FlowState": FlowStateSchema,
  "sdl.v1.TraceData": TraceDataSchema,
  "sdl.v1.TraceEvent": TraceEventSchema,
  "sdl.v1.AllPathsTraceData": AllPathsTraceDataSchema,
  "sdl.v1.TraceNode": TraceNodeSchema,
  "sdl.v1.Edge": EdgeSchema,
  "sdl.v1.GroupInfo": GroupInfoSchema,
  "sdl.v1.ParameterUpdate": ParameterUpdateSchema,
  "sdl.v1.ParameterUpdateResult": ParameterUpdateResultSchema,
  "sdl.v1.AggregateResult": AggregateResultSchema,
  "sdl.v1.CreateCanvasRequest": CreateCanvasRequestSchema,
  "sdl.v1.CreateCanvasResponse": CreateCanvasResponseSchema,
  "sdl.v1.UpdateCanvasRequest": UpdateCanvasRequestSchema,
  "sdl.v1.UpdateCanvasResponse": UpdateCanvasResponseSchema,
  "sdl.v1.ListCanvasesRequest": ListCanvasesRequestSchema,
  "sdl.v1.ListCanvasesResponse": ListCanvasesResponseSchema,
  "sdl.v1.GetCanvasRequest": GetCanvasRequestSchema,
  "sdl.v1.GetCanvasResponse": GetCanvasResponseSchema,
  "sdl.v1.DeleteCanvasRequest": DeleteCanvasRequestSchema,
  "sdl.v1.DeleteCanvasResponse": DeleteCanvasResponseSchema,
  "sdl.v1.ResetCanvasRequest": ResetCanvasRequestSchema,
  "sdl.v1.ResetCanvasResponse": ResetCanvasResponseSchema,
  "sdl.v1.LoadFileRequest": LoadFileRequestSchema,
  "sdl.v1.LoadFileResponse": LoadFileResponseSchema,
  "sdl.v1.UseSystemRequest": UseSystemRequestSchema,
  "sdl.v1.UseSystemResponse": UseSystemResponseSchema,
  "sdl.v1.AddGeneratorRequest": AddGeneratorRequestSchema,
  "sdl.v1.AddGeneratorResponse": AddGeneratorResponseSchema,
  "sdl.v1.ListGeneratorsRequest": ListGeneratorsRequestSchema,
  "sdl.v1.ListGeneratorsResponse": ListGeneratorsResponseSchema,
  "sdl.v1.GetGeneratorRequest": GetGeneratorRequestSchema,
  "sdl.v1.GetGeneratorResponse": GetGeneratorResponseSchema,
  "sdl.v1.UpdateGeneratorRequest": UpdateGeneratorRequestSchema,
  "sdl.v1.UpdateGeneratorResponse": UpdateGeneratorResponseSchema,
  "sdl.v1.StartGeneratorRequest": StartGeneratorRequestSchema,
  "sdl.v1.StartGeneratorResponse": StartGeneratorResponseSchema,
  "sdl.v1.StopGeneratorRequest": StopGeneratorRequestSchema,
  "sdl.v1.StopGeneratorResponse": StopGeneratorResponseSchema,
  "sdl.v1.DeleteGeneratorRequest": DeleteGeneratorRequestSchema,
  "sdl.v1.DeleteGeneratorResponse": DeleteGeneratorResponseSchema,
  "sdl.v1.StartAllGeneratorsRequest": StartAllGeneratorsRequestSchema,
  "sdl.v1.StartAllGeneratorsResponse": StartAllGeneratorsResponseSchema,
  "sdl.v1.StopAllGeneratorsRequest": StopAllGeneratorsRequestSchema,
  "sdl.v1.StopAllGeneratorsResponse": StopAllGeneratorsResponseSchema,
  "sdl.v1.AddMetricRequest": AddMetricRequestSchema,
  "sdl.v1.AddMetricResponse": AddMetricResponseSchema,
  "sdl.v1.DeleteMetricRequest": DeleteMetricRequestSchema,
  "sdl.v1.DeleteMetricResponse": DeleteMetricResponseSchema,
  "sdl.v1.ListMetricsRequest": ListMetricsRequestSchema,
  "sdl.v1.ListMetricsResponse": ListMetricsResponseSchema,
  "sdl.v1.QueryMetricsRequest": QueryMetricsRequestSchema,
  "sdl.v1.QueryMetricsResponse": QueryMetricsResponseSchema,
  "sdl.v1.AggregateMetricsRequest": AggregateMetricsRequestSchema,
  "sdl.v1.AggregateMetricsResponse": AggregateMetricsResponseSchema,
  "sdl.v1.StreamMetricsRequest": StreamMetricsRequestSchema,
  "sdl.v1.StreamMetricsResponse": StreamMetricsResponseSchema,
  "sdl.v1.ExecuteTraceRequest": ExecuteTraceRequestSchema,
  "sdl.v1.ExecuteTraceResponse": ExecuteTraceResponseSchema,
  "sdl.v1.TraceAllPathsRequest": TraceAllPathsRequestSchema,
  "sdl.v1.TraceAllPathsResponse": TraceAllPathsResponseSchema,
  "sdl.v1.SetParameterRequest": SetParameterRequestSchema,
  "sdl.v1.SetParameterResponse": SetParameterResponseSchema,
  "sdl.v1.GetParametersRequest": GetParametersRequestSchema,
  "sdl.v1.GetParametersResponse": GetParametersResponseSchema,
  "sdl.v1.BatchSetParametersRequest": BatchSetParametersRequestSchema,
  "sdl.v1.BatchSetParametersResponse": BatchSetParametersResponseSchema,
  "sdl.v1.EvaluateFlowsRequest": EvaluateFlowsRequestSchema,
  "sdl.v1.EvaluateFlowsResponse": EvaluateFlowsResponseSchema,
  "sdl.v1.GetFlowStateRequest": GetFlowStateRequestSchema,
  "sdl.v1.GetFlowStateResponse": GetFlowStateResponseSchema,
  "sdl.v1.GetSystemDiagramRequest": GetSystemDiagramRequestSchema,
  "sdl.v1.GetSystemDiagramResponse": GetSystemDiagramResponseSchema,
  "sdl.v1.GetUtilizationRequest": GetUtilizationRequestSchema,
  "sdl.v1.GetUtilizationResponse": GetUtilizationResponseSchema,
  "sdl.v1.UpdateMetricRequest": UpdateMetricRequestSchema,
  "sdl.v1.UpdateMetricResponse": UpdateMetricResponseSchema,
  "sdl.v1.ClearMetricsRequest": ClearMetricsRequestSchema,
  "sdl.v1.ClearMetricsResponse": ClearMetricsResponseSchema,
  "sdl.v1.SetMetricsListRequest": SetMetricsListRequestSchema,
  "sdl.v1.SetMetricsListResponse": SetMetricsListResponseSchema,
  "sdl.v1.UpdateDiagramRequest": UpdateDiagramRequestSchema,
  "sdl.v1.UpdateDiagramResponse": UpdateDiagramResponseSchema,
  "sdl.v1.HighlightComponentsRequest": HighlightComponentsRequestSchema,
  "sdl.v1.HighlightComponentsResponse": HighlightComponentsResponseSchema,
  "sdl.v1.ClearHighlightsRequest": ClearHighlightsRequestSchema,
  "sdl.v1.ClearHighlightsResponse": ClearHighlightsResponseSchema,
  "sdl.v1.UpdateGeneratorStateRequest": UpdateGeneratorStateRequestSchema,
  "sdl.v1.UpdateGeneratorStateResponse": UpdateGeneratorStateResponseSchema,
  "sdl.v1.SetGeneratorListRequest": SetGeneratorListRequestSchema,
  "sdl.v1.SetGeneratorListResponse": SetGeneratorListResponseSchema,
  "sdl.v1.LogMessageRequest": LogMessageRequestSchema,
  "sdl.v1.LogMessageResponse": LogMessageResponseSchema,
  "sdl.v1.ClearConsoleRequest": ClearConsoleRequestSchema,
  "sdl.v1.ClearConsoleResponse": ClearConsoleResponseSchema,
  "sdl.v1.UpdateFlowRatesRequest": UpdateFlowRatesRequestSchema,
  "sdl.v1.UpdateFlowRatesResponse": UpdateFlowRatesResponseSchema,
  "sdl.v1.ShowFlowPathRequest": ShowFlowPathRequestSchema,
  "sdl.v1.FlowPathSegment": FlowPathSegmentSchema,
  "sdl.v1.ShowFlowPathResponse": ShowFlowPathResponseSchema,
  "sdl.v1.ClearFlowPathsRequest": ClearFlowPathsRequestSchema,
  "sdl.v1.ClearFlowPathsResponse": ClearFlowPathsResponseSchema,
  "sdl.v1.UpdateUtilizationRequest": UpdateUtilizationRequestSchema,
  "sdl.v1.UpdateUtilizationResponse": UpdateUtilizationResponseSchema,
  "sdl.v1.FileInfo": FileInfoSchema,
  "sdl.v1.FilesystemInfo": FilesystemInfoSchema,
  "sdl.v1.ListFilesystemsRequest": ListFilesystemsRequestSchema,
  "sdl.v1.ListFilesystemsResponse": ListFilesystemsResponseSchema,
  "sdl.v1.ListFilesRequest": ListFilesRequestSchema,
  "sdl.v1.ListFilesResponse": ListFilesResponseSchema,
  "sdl.v1.ReadFileRequest": ReadFileRequestSchema,
  "sdl.v1.ReadFileResponse": ReadFileResponseSchema,
  "sdl.v1.WriteFileRequest": WriteFileRequestSchema,
  "sdl.v1.WriteFileResponse": WriteFileResponseSchema,
  "sdl.v1.DeleteFileRequest": DeleteFileRequestSchema,
  "sdl.v1.DeleteFileResponse": DeleteFileResponseSchema,
  "sdl.v1.CreateDirectoryRequest": CreateDirectoryRequestSchema,
  "sdl.v1.CreateDirectoryResponse": CreateDirectoryResponseSchema,
  "sdl.v1.GetFileInfoRequest": GetFileInfoRequestSchema,
  "sdl.v1.GetFileInfoResponse": GetFileInfoResponseSchema,
  "sdl.v1.SystemInfo": SystemInfoSchema,
  "sdl.v1.SystemProject": SystemProjectSchema,
  "sdl.v1.SystemVersion": SystemVersionSchema,
  "sdl.v1.ListSystemsRequest": ListSystemsRequestSchema,
  "sdl.v1.ListSystemsResponse": ListSystemsResponseSchema,
  "sdl.v1.GetSystemRequest": GetSystemRequestSchema,
  "sdl.v1.GetSystemResponse": GetSystemResponseSchema,
  "sdl.v1.GetSystemContentRequest": GetSystemContentRequestSchema,
  "sdl.v1.GetSystemContentResponse": GetSystemContentResponseSchema,
  "sdl.v1.InitializeSingletonRequest": InitializeSingletonRequestSchema,
  "sdl.v1.InitializeSingletonResponse": InitializeSingletonResponseSchema,
  "sdl.v1.InitializePresenterRequest": InitializePresenterRequestSchema,
  "sdl.v1.InitializePresenterResponse": InitializePresenterResponseSchema,
  "sdl.v1.ClientReadyRequest": ClientReadyRequestSchema,
  "sdl.v1.ClientReadyResponse": ClientReadyResponseSchema,
  "sdl.v1.FileSelectedRequest": FileSelectedRequestSchema,
  "sdl.v1.FileSelectedResponse": FileSelectedResponseSchema,
  "sdl.v1.FileSavedRequest": FileSavedRequestSchema,
  "sdl.v1.FileSavedResponse": FileSavedResponseSchema,
  "sdl.v1.DiagramComponentClickedRequest": DiagramComponentClickedRequestSchema,
  "sdl.v1.DiagramComponentClickedResponse": DiagramComponentClickedResponseSchema,
  "sdl.v1.DiagramComponentHoveredRequest": DiagramComponentHoveredRequestSchema,
  "sdl.v1.DiagramComponentHoveredResponse": DiagramComponentHoveredResponseSchema,
};

/**
 * Schema registry instance for sdl.v1 package with utility methods
 * Extends BaseSchemaRegistry with package-specific schema data
 */
// Schema utility functions (now inherited from BaseSchemaRegistry in runtime package)
// Creating instance with package-specific schema registry
const registryInstance = new BaseSchemaRegistry(sdl_v1SchemaRegistry);

export const getSchema = registryInstance.getSchema.bind(registryInstance);
export const getFieldSchema = registryInstance.getFieldSchema.bind(registryInstance);
export const getFieldSchemaById = registryInstance.getFieldSchemaById.bind(registryInstance);
export const isOneofField = registryInstance.isOneofField.bind(registryInstance);
export const getOneofFields = registryInstance.getOneofFields.bind(registryInstance);