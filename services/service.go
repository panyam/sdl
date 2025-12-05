package services

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	protoservices "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/panyam/sdl/lib/components"
	"github.com/panyam/sdl/lib/decl"
	"github.com/panyam/sdl/lib/loader"
	"github.com/panyam/sdl/lib/runtime"
	// Add if using SectionMetadataToProto helper:
	// tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// --- CanvasService struct holds configuration and state ---
type CanvasService struct {
	protoservices.UnimplementedCanvasServiceServer
	store      map[string]*Canvas
	storeMutex sync.RWMutex
}

// --- NewCanvasService Constructor ---
func NewCanvasService() *CanvasService {
	out := &CanvasService{
		store: map[string]*Canvas{}, // Assign store
	}
	out.CreateCanvas(context.Background(), &protos.CreateCanvasRequest{
		Canvas: &protos.Canvas{
			Id: "default",
		},
	})
	return out
}

// GetDefaultCanvas returns the default canvas
func (s *CanvasService) GetDefaultCanvas() *Canvas {
	s.storeMutex.RLock()
	defer s.storeMutex.RUnlock()
	return s.store["default"]
}

func (s *CanvasService) CreateCanvas(ctx context.Context, req *protos.CreateCanvasRequest) (resp *protos.CreateCanvasResponse, err error) {
	slog.Info("CreateCanvas Request", "req", req)
	s.storeMutex.Lock()
	defer s.storeMutex.Unlock()

	canvasProto := req.Canvas
	if canvasProto == nil {
		return nil, status.Error(codes.InvalidArgument, "Canvas payload cannot be nil")
	}

	canvasId := strings.TrimSpace(canvasProto.Id)
	if canvasId == "" {
		slog.Error("No canvasID provided")
		return nil, status.Error(codes.InvalidArgument, "Provide a canvas id")
	}
	if s.store[canvasId] != nil {
		slog.Error("Canvas ID already exists", "id", canvasId)
		return nil, status.Error(codes.AlreadyExists, "Canvas ID already taken")
	}
	slog.Debug("Creating Canvas: ", "id", canvasId)

	canvas := NewCanvas(canvasId, nil)
	s.store[canvasId] = canvas
	resp = &protos.CreateCanvasResponse{Canvas: canvas.ToProto()}
	return resp, nil
}

func (s *CanvasService) ListCanvases(ctx context.Context, req *protos.ListCanvasesRequest) (resp *protos.ListCanvasesResponse, err error) {
	slog.Info("TODO: ListCanvases Request", "req", req)
	resp = &protos.ListCanvasesResponse{}
	s.storeMutex.RLock()
	defer s.storeMutex.RUnlock()
	for _, v := range s.store {
		resp.Canvases = append(resp.Canvases, v.ToProto())
	}
	return
}

func (s *CanvasService) DeleteCanvas(ctx context.Context, req *protos.DeleteCanvasRequest) (resp *protos.DeleteCanvasResponse, err error) {
	slog.Info("TODO: DeleteCanvas Request", "req", req)
	resp = &protos.DeleteCanvasResponse{}
	return
}

func (s *CanvasService) ResetCanvas(ctx context.Context, req *protos.ResetCanvasRequest) (resp *protos.ResetCanvasResponse, err error) {
	slog.Info("ResetCanvas Request", "req", req)
	resp = &protos.ResetCanvasResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		resetErr := canvas.Reset()
		if resetErr != nil {
			resp.Success = false
			resp.Message = fmt.Sprintf("Failed to reset canvas: %v", resetErr)
			err = resetErr
		} else {
			resp.Success = true
			resp.Message = "Canvas reset successfully - all generators stopped, metrics cleared, and state reset"
		}
	})

	return
}

func (s *CanvasService) withCanvas(canvasId string, callback func(*Canvas)) (err error) {
	s.storeMutex.Lock()
	defer s.storeMutex.Unlock()
	if c, ok := s.store[canvasId]; c == nil || !ok {
		slog.Error("Canvas ID already exists", "id", canvasId)
		return status.Error(codes.NotFound, "Canvas not found for id")
	} else if callback != nil {
		callback(c)
	}
	return nil
}

func (s *CanvasService) GetCanvas(ctx context.Context, req *protos.GetCanvasRequest) (resp *protos.GetCanvasResponse, err error) {
	slog.Info("GetCanvas Request", "req", req)
	resp = &protos.GetCanvasResponse{}
	s.withCanvas(req.Id, func(canvas *Canvas) {
		resp.Canvas = canvas.ToProto()
	})
	return
}

func (s *CanvasService) LoadFile(ctx context.Context, req *protos.LoadFileRequest) (resp *protos.LoadFileResponse, err error) {
	slog.Info("LoadFile Request", "req", req)
	resp = &protos.LoadFileResponse{}
	s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		err = canvas.Load(req.SdlFilePath)
	})
	return
}

func (s *CanvasService) UseSystem(ctx context.Context, req *protos.UseSystemRequest) (resp *protos.UseSystemResponse, err error) {
	slog.Info("UseSystem Request", "req", req)
	resp = &protos.UseSystemResponse{}
	s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		err = canvas.Use(req.SystemName)
	})
	return
}

func (s *CanvasService) ListGenerators(ctx context.Context, req *protos.ListGeneratorsRequest) (resp *protos.ListGeneratorsResponse, err error) {
	slog.Info("ListGenerators Request", "req", req)
	resp = &protos.ListGeneratorsResponse{}
	s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		gens := canvas.ListGenerators()
		for _, v := range gens {
			resp.Generators = append(resp.Generators, v.Generator)
		}
	})
	return
}

// GetGenerator returns a specific generator
func (s *CanvasService) GetGenerator(ctx context.Context, req *protos.GetGeneratorRequest) (resp *protos.GetGeneratorResponse, err error) {
	slog.Info("GetGenerator Request", "req", req)
	resp = &protos.GetGeneratorResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		generator := canvas.GetGenerator(req.GeneratorId)
		// Convert to proto format
		resp.Generator = generator.Generator
	})

	return
}

func (s *CanvasService) AddGenerator(ctx context.Context, req *protos.AddGeneratorRequest) (resp *protos.AddGeneratorResponse, err error) {
	slog.Info("AddGenerator Request", "req", req)
	resp = &protos.AddGeneratorResponse{}
	s.withCanvas(req.Generator.CanvasId, func(canvas *Canvas) {
		nativeGen := req.Generator
		gen := &GeneratorInfo{Generator: nativeGen}
		err = canvas.AddGenerator(gen)
	})
	return
}

func (s *CanvasService) DeleteGenerator(ctx context.Context, req *protos.DeleteGeneratorRequest) (resp *protos.DeleteGeneratorResponse, err error) {
	slog.Info("DeleteGenerator Request", "req", req)
	resp = &protos.DeleteGeneratorResponse{}
	s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		err = canvas.RemoveGenerator(req.GeneratorId)
	})
	return
}

func (s *CanvasService) StartAllGenerators(ctx context.Context, req *protos.StartAllGeneratorsRequest) (resp *protos.StartAllGeneratorsResponse, err error) {
	slog.Info("StartAllGenerators Request", "req", req)
	resp = &protos.StartAllGeneratorsResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		result, startErr := canvas.StartAllGenerators()
		if result != nil {
			resp.TotalGenerators = int32(result.TotalGenerators)
			resp.StartedCount = int32(result.ProcessedCount)
			resp.AlreadyRunningCount = int32(result.AlreadyInStateCount)
			resp.FailedCount = int32(result.FailedCount)
			resp.FailedIds = result.FailedIDs
		}
		if startErr != nil {
			err = startErr
		}
	})
	return
}

func (s *CanvasService) StopAllGenerators(ctx context.Context, req *protos.StopAllGeneratorsRequest) (resp *protos.StopAllGeneratorsResponse, err error) {
	slog.Info("StopAllGenerators  Request", "req", req)
	resp = &protos.StopAllGeneratorsResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		result, stopErr := canvas.StopAllGenerators()
		if result != nil {
			resp.TotalGenerators = int32(result.TotalGenerators)
			resp.StoppedCount = int32(result.ProcessedCount)
			resp.AlreadyStoppedCount = int32(result.AlreadyInStateCount)
			resp.FailedCount = int32(result.FailedCount)
			resp.FailedIds = result.FailedIDs
		}
		if stopErr != nil {
			err = stopErr
		}
	})
	return
}

func (s *CanvasService) StartGenerator(ctx context.Context, req *protos.StartGeneratorRequest) (resp *protos.StartGeneratorResponse, err error) {
	slog.Info("StartGenerator Request", "req", req)
	resp = &protos.StartGeneratorResponse{}
	s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		err = canvas.StartGenerator(req.GeneratorId)
	})
	return
}

func (s *CanvasService) StopGenerator(ctx context.Context, req *protos.StopGeneratorRequest) (resp *protos.StopGeneratorResponse, err error) {
	slog.Info("StopGenerator Request", "req", req)
	resp = &protos.StopGeneratorResponse{}
	s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		err = canvas.StopGenerator(req.GeneratorId)
	})
	return
}

func (s *CanvasService) UpdateGenerator(ctx context.Context, req *protos.UpdateGeneratorRequest) (resp *protos.UpdateGeneratorResponse, err error) {
	slog.Info("UpdateGenerator Request", "req", req)
	resp = &protos.UpdateGeneratorResponse{}

	err = s.withCanvas(req.Generator.CanvasId, func(canvas *Canvas) {
		// Update the generator
		nativeGen := req.Generator
		err = canvas.UpdateGenerator(nativeGen)
		if err != nil {
			return
		}
		resp.Generator = req.Generator
	})

	return
}

func (s *CanvasService) AddMetric(ctx context.Context, req *protos.AddMetricRequest) (resp *protos.AddMetricResponse, err error) {
	slog.Info("AddMetric Request", "req", req)
	resp = &protos.AddMetricResponse{}
	s.withCanvas(req.Metric.CanvasId, func(canvas *Canvas) {
		nativeMetric := req.Metric
		err = canvas.metricTracer.AddMetricSpec(&MetricSpec{Metric: nativeMetric})
		resp.Metric = req.Metric
	})
	return
}

func (s *CanvasService) DeleteMetric(ctx context.Context, req *protos.DeleteMetricRequest) (resp *protos.DeleteMetricResponse, err error) {
	slog.Info("DeleteMetric Request", "req", req)
	resp = &protos.DeleteMetricResponse{}
	s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		canvas.metricTracer.RemoveMetricSpec(req.MetricId)
	})
	return
}

func (s *CanvasService) ExecuteTrace(ctx context.Context, req *protos.ExecuteTraceRequest) (resp *protos.ExecuteTraceResponse, err error) {
	slog.Info("ExecuteTrace Request", "req", req)
	resp = &protos.ExecuteTraceResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		// Execute trace on the canvas
		traceData, traceErr := canvas.ExecuteTrace(req.Component, req.Method)
		if traceErr != nil {
			err = traceErr
			return
		}

		// Convert runtime.TraceData to proto.TraceData
		resp.TraceData = convertTraceDataToProto(traceData)
	})
	return
}

func (s *CanvasService) TraceAllPaths(ctx context.Context, req *protos.TraceAllPathsRequest) (resp *protos.TraceAllPathsResponse, err error) {
	slog.Info("TraceAllPaths Request", "req", req)
	resp = &protos.TraceAllPathsResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		// Execute path traversal on the canvas
		allPathsData, traceErr := canvas.TraceAllPaths(req.Component, req.Method, req.MaxDepth)
		if traceErr != nil {
			err = traceErr
			return
		}

		// Convert runtime.AllPathsTraceData to proto.AllPathsTraceData
		resp.TraceData = convertAllPathsTraceDataToProto(allPathsData)
	})
	return
}

// Helper to convert runtime.TraceData to proto.TraceData
func convertTraceDataToProto(td *runtime.TraceData) *protos.TraceData {
	if td == nil {
		return nil
	}

	protoEvents := make([]*protos.TraceEvent, len(td.Events))
	for i, event := range td.Events {
		protoEvents[i] = &protos.TraceEvent{
			Kind:         string(event.Kind),
			Id:           event.ID,
			ParentId:     event.ParentID,
			Timestamp:    float64(event.Timestamp),
			Duration:     float64(event.Duration),
			Component:    event.ComponentName,
			Method:       event.MethodName,
			Args:         event.Arguments,
			ReturnValue:  event.ReturnValue,
			ErrorMessage: event.ErrorMessage,
		}
	}

	return &protos.TraceData{
		System:     td.System,
		EntryPoint: td.EntryPoint,
		Events:     protoEvents,
	}
}

// Helper to convert runtime.AllPathsTraceData to proto.AllPathsTraceData
func convertAllPathsTraceDataToProto(aptd *runtime.AllPathsTraceData) *protos.AllPathsTraceData {
	if aptd == nil {
		return nil
	}

	return &protos.AllPathsTraceData{
		TraceId: aptd.TraceID,
		Root:    convertTraceNodeToProto(&aptd.Root),
	}
}

// Helper to convert runtime.TraceNode to proto.TraceNode
func convertTraceNodeToProto(tn *runtime.TraceNode) *protos.TraceNode {
	if tn == nil {
		return nil
	}

	protoEdges := make([]*protos.Edge, len(tn.Edges))
	for i, edge := range tn.Edges {
		protoEdges[i] = convertEdgeToProto(&edge)
	}

	protoGroups := make([]*protos.GroupInfo, len(tn.Groups))
	for i, group := range tn.Groups {
		protoGroups[i] = &protos.GroupInfo{
			GroupStart: group.GroupStart,
			GroupEnd:   group.GroupEnd,
			GroupLabel: group.GroupLabel,
			GroupType:  group.GroupType,
		}
	}

	return &protos.TraceNode{
		StartingTarget: tn.StartingTarget,
		Edges:          protoEdges,
		Groups:         protoGroups,
	}
}

// Helper to convert runtime.Edge to proto.Edge
func convertEdgeToProto(e *runtime.Edge) *protos.Edge {
	if e == nil {
		return nil
	}

	return &protos.Edge{
		Id:            e.ID,
		NextNode:      convertTraceNodeToProto(&e.NextNode),
		Label:         e.Label,
		IsAsync:       e.IsAsync,
		IsReverse:     e.IsReverse,
		Probability:   e.Probability,
		Condition:     e.Condition,
		IsConditional: e.IsConditional,
	}
}

// ListMetrics returns all available metrics
func (s *CanvasService) ListMetrics(ctx context.Context, req *protos.ListMetricsRequest) (resp *protos.ListMetricsResponse, err error) {
	slog.Info("ListMetrics Request", "req", req)
	resp = &protos.ListMetricsResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		if canvas.metricTracer == nil {
			resp.Metrics = []*protos.Metric{}
			return
		}

		// Get all metrics from the metric tracer
		metrics := canvas.metricTracer.ListMetrics()
		protoMetrics := make([]*protos.Metric, len(metrics))
		for i, m := range metrics {
			protoMetrics[i] = m
		}
		resp.Metrics = protoMetrics
	})

	return
}

// QueryMetrics returns raw metric data points
func (s *CanvasService) QueryMetrics(ctx context.Context, req *protos.QueryMetricsRequest) (resp *protos.QueryMetricsResponse, err error) {
	slog.Info("QueryMetrics Request", "req", req)
	resp = &protos.QueryMetricsResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		if canvas.metricTracer == nil || canvas.metricTracer.store == nil {
			err = status.Error(codes.FailedPrecondition, "no metric store available")
			return
		}

		// Find the metric by ID
		metric := canvas.metricTracer.GetMetricByID(req.MetricId)
		if metric == nil {
			err = status.Errorf(codes.NotFound, "metric %s not found", req.MetricId)
			return
		}

		// Query the store
		queryOpts := QueryOptions{
			StartTime: time.Unix(int64(req.StartTime), 0),
			EndTime:   time.Unix(int64(req.EndTime), 0),
			Limit:     int(req.Limit),
		}

		result, queryErr := canvas.metricTracer.store.Query(ctx, metric, queryOpts)
		if queryErr != nil {
			err = status.Errorf(codes.Internal, "failed to query metrics: %v", queryErr)
			return
		}

		// Convert to proto format
		resp.Points = make([]*protos.MetricPoint, len(result.Points))
		for i, point := range result.Points {
			resp.Points[i] = &protos.MetricPoint{
				Timestamp: float64(point.Timestamp.Unix()),
				Value:     point.Value,
			}
		}
	})

	return
}

// AggregateMetrics returns aggregated metric data
func (s *CanvasService) AggregateMetrics(ctx context.Context, req *protos.AggregateMetricsRequest) (resp *protos.AggregateMetricsResponse, err error) {
	slog.Info("AggregateMetrics Request", "req", req)
	resp = &protos.AggregateMetricsResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		if canvas.metricTracer == nil || canvas.metricTracer.store == nil {
			err = status.Error(codes.FailedPrecondition, "no metric store available")
			return
		}

		// Find the metric by ID
		metric := canvas.metricTracer.GetMetricByID(req.MetricId)
		if metric == nil {
			err = status.Errorf(codes.NotFound, "metric %s not found", req.MetricId)
			return
		}

		// Map string to AggregateFunc
		var aggFunc AggregateFunc
		switch req.Function {
		case "count":
			aggFunc = AggCount
		case "sum":
			aggFunc = AggSum
		case "avg":
			aggFunc = AggAvg
		case "min":
			aggFunc = AggMin
		case "max":
			aggFunc = AggMax
		case "p50":
			aggFunc = AggP50
		case "p90":
			aggFunc = AggP90
		case "p95":
			aggFunc = AggP95
		case "p99":
			aggFunc = AggP99
		default:
			err = status.Errorf(codes.InvalidArgument, "unknown aggregation function: %s", req.Function)
			return
		}

		// Aggregate the data
		aggOpts := AggregateOptions{
			StartTime: time.Unix(int64(req.StartTime), 0),
			EndTime:   time.Unix(int64(req.EndTime), 0),
			Functions: []AggregateFunc{aggFunc},
			Window:    time.Duration(req.WindowSize * float64(time.Second)),
		}

		result, aggErr := canvas.metricTracer.store.Aggregate(ctx, metric, aggOpts)
		if aggErr != nil {
			err = status.Errorf(codes.Internal, "failed to aggregate metrics: %v", aggErr)
			return
		}

		// Convert to proto format
		resp.Results = make([]*protos.AggregateResult, len(result.Buckets))
		for i, bucket := range result.Buckets {
			// Get the value for our requested function
			value := bucket.Values[aggFunc]
			resp.Results[i] = &protos.AggregateResult{
				Timestamp: float64(bucket.Time.Unix()),
				Value:     value,
			}
		}
	})

	return
}

// SetParameter sets a component parameter value
func (s *CanvasService) SetParameter(ctx context.Context, req *protos.SetParameterRequest) (resp *protos.SetParameterResponse, err error) {
	slog.Info("SetParameter Request", "req", req)
	resp = &protos.SetParameterResponse{Success: true}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		if canvas.activeSystem == nil {
			err = status.Error(codes.FailedPrecondition, "no active system")
			resp.Success = false
			resp.ErrorMessage = "no active system"
			return
		}

		// Parse the value expression using SDL parser helper
		expr, parseErr := loader.ParseExpresssion(req.NewValue)
		if parseErr != nil {
			err = status.Errorf(codes.InvalidArgument, "failed to parse value expression: %v", parseErr)
			resp.Success = false
			resp.ErrorMessage = parseErr.Error()
			return
		}

		// Evaluate the expression to get the actual decl.Value
		eval := runtime.NewSimpleEval(canvas.activeSystem.File, nil)
		result, _ := eval.Eval(expr, canvas.activeSystem.Env, nil)
		if eval.HasErrors() {
			evalErr := eval.Errors[0]
			err = status.Errorf(codes.InvalidArgument, "failed to evaluate expression: %v", evalErr)
			resp.Success = false
			resp.ErrorMessage = evalErr.Error()
			return
		}

		// Use runtime.SetParam to set the parameter

		oldValue, setErr := canvas.runtime.SetParam(canvas.activeSystem, req.Path, result)
		if setErr != nil {
			err = status.Errorf(codes.Internal, "failed to set parameter: %v", setErr)
			resp.Success = false
			resp.ErrorMessage = setErr.Error()
			return
		}

		// Convert decl.Value to Go value for storage
		resp.OldValue = oldValue.String()
		resp.NewValue = result.String()
	})

	return
}

// GetParameters gets parameter values
func (s *CanvasService) GetParameters(ctx context.Context, req *protos.GetParametersRequest) (resp *protos.GetParametersResponse, err error) {
	slog.Info("GetParameters Request", "req", req)
	resp = &protos.GetParametersResponse{
		Parameters: make(map[string]string),
	}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		if canvas.activeSystem == nil {
			err = status.Error(codes.FailedPrecondition, "no active system")
			return
		}

		if req.Path == "" {
			err = status.Error(codes.FailedPrecondition, "Provide one or more parameter paths")
			return
		}

		// Get specific parameter using runtime.GetParam
		paramValue, getErr := canvas.runtime.GetParam(canvas.activeSystem, req.Path)
		if getErr != nil {
			err = status.Errorf(codes.NotFound, "failed to get parameter: %v", getErr)
			return
		}

		// Format the value as SDL expression
		resp.Parameters[req.Path] = formatDeclValueAsSDL(paramValue)
	})

	return
}

// formatValueAsSDL converts a Go value to SDL expression string
func formatValueAsSDL(value any) string {
	switch v := value.(type) {
	case float64:
		return fmt.Sprintf("%g", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case string:
		// If it's already an SDL expression, return as-is
		if strings.HasPrefix(v, "'") || strings.HasPrefix(v, "\"") {
			return v
		}
		// Otherwise quote it
		return fmt.Sprintf("'%s'", v)
	default:
		// For complex types, use string representation
		return fmt.Sprintf("%v", v)
	}
}

// formatDeclValueAsSDL converts a decl.Value to SDL expression string
func formatDeclValueAsSDL(value decl.Value) string {
	switch value.Type.Tag {
	case decl.TypeTagSimple:
		if value.Type.Info == "Int" {
			return fmt.Sprintf("%d", value.IntVal())
		} else if value.Type.Info == "Float" {
			return fmt.Sprintf("%g", value.FloatVal())
		} else if value.Type.Info == "Bool" {
			return fmt.Sprintf("%t", value.BoolVal())
		} else if value.Type.Info == "String" {
			return fmt.Sprintf("'%s'", value.StringVal())
		}
	}
	return value.String()
}

// parseSDLExpression parses an SDL expression string and returns a Go value
func parseSDLExpression(expr string) (any, error) {
	// Try to parse as common types first for efficiency
	if expr == "true" || expr == "false" {
		return expr == "true", nil
	}

	// Try to parse as integer
	var intVal int64
	if _, err := fmt.Sscanf(expr, "%d", &intVal); err == nil {
		return intVal, nil
	}

	// Try to parse as float
	var floatVal float64
	if _, err := fmt.Sscanf(expr, "%f", &floatVal); err == nil {
		return floatVal, nil
	}

	// If it's a quoted string, extract the content
	if (strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'")) ||
		(strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"")) {
		return expr[1 : len(expr)-1], nil
	}

	// For more complex expressions, fall back to string
	return expr, nil
}

// GetSystemDiagram returns the system topology for visualization
func (s *CanvasService) GetSystemDiagram(ctx context.Context, req *protos.GetSystemDiagramRequest) (resp *protos.GetSystemDiagramResponse, err error) {
	slog.Info("GetSystemDiagram Request", "req", req)
	resp = &protos.GetSystemDiagramResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		// Get the system diagram from the canvas
		diagram, diagErr := canvas.GetSystemDiagram()
		if diagErr != nil {
			err = status.Errorf(codes.Internal, "failed to get system diagram: %v", diagErr)
			return
		}

		// Convert native diagram to proto
		resp.Diagram = ToProtoSystemDiagram(diagram)
	})

	return
}

// BatchSetParameters sets multiple parameters atomically
func (s *CanvasService) BatchSetParameters(ctx context.Context, req *protos.BatchSetParametersRequest) (resp *protos.BatchSetParametersResponse, err error) {
	slog.Info("BatchSetParameters Request", "req", req)
	resp = &protos.BatchSetParametersResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		// Convert proto updates to map
		updates := make(map[string]any)
		for _, update := range req.Updates {
			// Parse SDL expression to Go value
			value, parseErr := parseSDLExpression(update.NewValue)
			if parseErr != nil {
				err = status.Errorf(codes.InvalidArgument, "failed to parse value for %s: %v", update.Path, parseErr)
				return
			}
			updates[update.Path] = value
		}

		// Apply batch updates
		results, batchErr := canvas.BatchSetParameters(updates)
		if batchErr != nil && results == nil {
			log.Println("Result: ", results)
			log.Println("Err: ", batchErr)
			err = status.Errorf(codes.Internal, "failed to batch set parameters: %v", batchErr)
			return
		}

		// Convert results to proto format
		resp.Results = make([]*protos.ParameterUpdateResult, 0, len(results))
		resp.Success = true
		for path, result := range results {
			protoResult := &protos.ParameterUpdateResult{
				Path:     path,
				OldValue: result.String(),
			}
			resp.Results = append(resp.Results, protoResult)
		}

		if batchErr != nil {
			resp.ErrorMessage = batchErr.Error()
			resp.Success = false
		}
	})

	return
}

// EvaluateFlows evaluates system flows using specified strategy
func (s *CanvasService) EvaluateFlows(ctx context.Context, req *protos.EvaluateFlowsRequest) (resp *protos.EvaluateFlowsResponse, err error) {
	slog.Info("EvaluateFlows Request", "req", req)
	resp = &protos.EvaluateFlowsResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		// Evaluate flows with the specified strategy
		result, evalErr := canvas.EvaluateFlowWithStrategy(req.Strategy)
		if evalErr != nil {
			err = status.Errorf(codes.Internal, "failed to evaluate flows: %v", evalErr)
			return
		}

		// Convert result to proto format
		resp.Strategy = result.Strategy
		resp.Status = string(result.Status)
		resp.Iterations = int32(result.Iterations)
		resp.Warnings = result.Warnings

		// Convert component rates
		resp.ComponentRates = result.Flows.ComponentRates

		// Convert flow edges
		resp.FlowEdges = make([]*protos.FlowEdge, len(result.Flows.Edges))
		for i, edge := range result.Flows.Edges {
			resp.FlowEdges[i] = &protos.FlowEdge{
				FromComponent: edge.From.Component,
				FromMethod:    edge.From.Method,
				ToComponent:   edge.To.Component,
				ToMethod:      edge.To.Method,
				Rate:          edge.Rate,
			}
		}
	})

	return
}

// GetFlowState returns the current flow state
func (s *CanvasService) GetFlowState(ctx context.Context, req *protos.GetFlowStateRequest) (resp *protos.GetFlowStateResponse, err error) {
	slog.Info("GetFlowState Request", "req", req)
	resp = &protos.GetFlowStateResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		// Get current flow state from canvas
		flowState := canvas.GetCurrentFlowState()

		// Convert to proto format
		resp.State = &protos.FlowState{
			Strategy:        flowState.Strategy,
			Rates:           flowState.Rates,
			ManualOverrides: flowState.ManualOverrides,
		}
	})

	return
}

// StreamMetrics streams real-time metric updates
func (s *CanvasService) StreamMetrics(req *protos.StreamMetricsRequest, stream protoservices.CanvasService_StreamMetricsServer) error {
	slog.Info("StreamMetrics Request", "req", req)

	// Get the canvas
	s.storeMutex.RLock()
	canvas := s.store[req.CanvasId]
	s.storeMutex.RUnlock()

	if canvas == nil {
		return status.Error(codes.NotFound, "canvas not found")
	}

	// Get the metric store
	if canvas.metricTracer == nil || canvas.metricTracer.store == nil {
		return status.Error(codes.FailedPrecondition, "no metric store available")
	}

	// Subscribe to metric updates
	ctx := stream.Context()
	updateChan, err := canvas.metricTracer.store.Subscribe(ctx, req.MetricIds)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to subscribe to metrics: %v", err)
	}

	// Stream updates until context is cancelled
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case batch := <-updateChan:
			if batch == nil {
				// Channel closed
				return nil
			}

			// Convert to proto format
			response := &protos.StreamMetricsResponse{
				Updates: make([]*protos.MetricUpdate, len(batch.Updates)),
			}

			for i, update := range batch.Updates {
				response.Updates[i] = &protos.MetricUpdate{
					MetricId: update.MetricID,
					Point: &protos.MetricPoint{
						Timestamp: float64(update.Point.Timestamp.Unix()),
						Value:     update.Point.Value,
					},
				}
			}

			// Send the update
			if err := stream.Send(response); err != nil {
				return err
			}
		}
	}
}

// GetUtilization returns resource utilization information
func (s *CanvasService) GetUtilization(ctx context.Context, req *protos.GetUtilizationRequest) (resp *protos.GetUtilizationResponse, err error) {
	slog.Info("GetUtilization Request", "req", req)
	resp = &protos.GetUtilizationResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		// Get the system instance
		system := canvas.activeSystem
		if system == nil {
			return
		}

		// Collect utilization info from all components
		var allInfos []components.UtilizationInfo

		// If specific components are requested, only get those
		if len(req.Components) > 0 {
			for _, compName := range req.Components {
				if comp := system.FindComponent(compName); comp != nil {
					infos := comp.GetUtilizationInfo()
					allInfos = append(allInfos, infos...)
				}
			}
		} else {
			// Get all components
			for _, comp := range system.AllComponents() {
				infos := comp.GetUtilizationInfo()
				allInfos = append(allInfos, infos...)
			}
		}

		// Convert to proto format
		resp.Utilizations = make([]*protos.UtilizationInfo, len(allInfos))
		for i, info := range allInfos {
			resp.Utilizations[i] = &protos.UtilizationInfo{
				ResourceName:      info.ResourceName,
				ComponentPath:     info.ComponentPath,
				Utilization:       info.Utilization,
				Capacity:          info.Capacity,
				CurrentLoad:       info.CurrentLoad,
				IsBottleneck:      info.IsBottleneck,
				WarningThreshold:  info.WarningThreshold,
				CriticalThreshold: info.CriticalThreshold,
			}
		}
	})

	return
}
