package console

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/panyam/sdl/decl"
	protos "github.com/panyam/sdl/gen/go/sdl/v1"
	"github.com/panyam/sdl/loader"
	"github.com/panyam/sdl/runtime"
	// Add if using SectionMetadataToProto helper:
	// tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// --- CanvasService struct holds configuration and state ---
type CanvasService struct {
	protos.UnimplementedCanvasServiceServer
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

	canvas := NewCanvas(canvasId)
	s.store[canvasId] = canvas
	resp = &protos.CreateCanvasResponse{Canvas: canvas.ToProto()}
	return resp, nil
}

func (c *Canvas) ToProto() *protos.Canvas {
	out := protos.Canvas{}
	out.Id = c.id
	if c.activeSystem != nil {
		out.ActiveSystem = c.activeSystem.System.Name.Value
	}
	return &out
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

func (s *CanvasService) AddGenerator(ctx context.Context, req *protos.AddGeneratorRequest) (resp *protos.AddGeneratorResponse, err error) {
	slog.Info("AddGenerator Request", "req", req)
	resp = &protos.AddGeneratorResponse{}
	s.withCanvas(req.Generator.CanvasId, func(canvas *Canvas) {
		gen := &GeneratorInfo{Generator: req.Generator}
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

func (s *CanvasService) ResumeGenerator(ctx context.Context, req *protos.ResumeGeneratorRequest) (resp *protos.ResumeGeneratorResponse, err error) {
	slog.Info("ResumeGenerator Request", "req", req)
	resp = &protos.ResumeGeneratorResponse{}
	s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		err = canvas.ResumeGenerator(req.GeneratorId)
	})
	return
}

func (s *CanvasService) PauseGenerator(ctx context.Context, req *protos.PauseGeneratorRequest) (resp *protos.PauseGeneratorResponse, err error) {
	slog.Info("PauseGenerator Request", "req", req)
	resp = &protos.PauseGeneratorResponse{}
	s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		err = canvas.PauseGenerator(req.GeneratorId)
	})
	return
}

func (s *CanvasService) UpdateGenerator(ctx context.Context, req *protos.UpdateGeneratorRequest) (resp *protos.UpdateGeneratorResponse, err error) {
	slog.Info("UpdateGenerator Request", "req", req)
	resp = &protos.UpdateGeneratorResponse{}
	return
}

func (s *CanvasService) AddMetric(ctx context.Context, req *protos.AddMetricRequest) (resp *protos.AddMetricResponse, err error) {
	slog.Info("AddMetric Request", "req", req)
	resp = &protos.AddMetricResponse{}
	s.withCanvas(req.Metric.CanvasId, func(canvas *Canvas) {
		err = canvas.metricTracer.AddMetricSpec(&MetricSpec{Metric: req.Metric})
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

// ListMetrics returns all available metrics
func (s *CanvasService) ListMetrics(ctx context.Context, req *protos.ListMetricsRequest) (resp *protos.ListMetricsResponse, err error) {
	slog.Info("ListMetrics Request", "req", req)
	resp = &protos.ListMetricsResponse{}

	err = s.withCanvas(req.CanvasId, func(canvas *Canvas) {
		if canvas.metricTracer == nil || canvas.metricTracer.store == nil {
			resp.Metrics = []*protos.MetricInfo{}
			return
		}

		// Get all metrics from the metric tracer
		metrics := canvas.metricTracer.ListMetrics()
		resp.Metrics = make([]*protos.MetricInfo, 0, len(metrics))

		for _, metric := range metrics {
			// Query the store to get stats about this metric
			queryOpts := QueryOptions{
				StartTime: time.Unix(0, 0),
				EndTime:   time.Now(),
				Limit:     1, // Just to get count
			}

			result, _ := canvas.metricTracer.store.Query(ctx, metric, queryOpts)

			info := &protos.MetricInfo{
				Id:         metric.Id,
				Component:  metric.Component,
				Method:     strings.Join(metric.Methods, ","),
				MetricType: metric.MetricType,
			}

			info.DataPoints = int64(len(result.Points))
			if len(result.Points) > 0 {
				info.OldestTimestamp = float64(result.Points[0].Timestamp.Unix())
				info.NewestTimestamp = float64(result.Points[len(result.Points)-1].Timestamp.Unix())
			}

			resp.Metrics = append(resp.Metrics, info)
		}
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
			return fmt.Sprintf("%d", value.FloatVal())
		} else if value.Type.Info == "Bool" {
			return fmt.Sprintf("%d", value.BoolVal())
		} else if value.Type.Info == "String" {
			return fmt.Sprintf("%d", value.StringVal())
		}
	}
	return value.String()
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

		// Canvas.GetSystemDiagram now returns proto directly
		resp.Diagram = diagram
	})

	return
}
