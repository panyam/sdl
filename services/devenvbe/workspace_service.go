package devenvbe

import (
	"context"
	"fmt"
	"strconv"
	"time"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	protoservices "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/panyam/sdl/lib/loader"
	"github.com/panyam/sdl/lib/runtime"
	"github.com/panyam/sdl/services"
)

// WorkspaceService implements the generated WorkspaceServiceServer by wrapping DevEnv.
// This is the local-mode backend — no server needed. Follows the lilbattle
// fsbe.GamesService pattern where a backend wraps a local storage/runtime.
type WorkspaceService struct {
	protoservices.UnimplementedWorkspaceServiceServer
	DevEnv *services.DevEnv
}

// NewWorkspaceService creates a local workspace service backed by DevEnv.
func NewWorkspaceService(resolver loader.FileResolver) *WorkspaceService {
	return &WorkspaceService{
		DevEnv: services.NewDevEnv(resolver),
	}
}

// File and system management

func (s *WorkspaceService) LoadFile(_ context.Context, req *protos.LoadFileRequest) (*protos.LoadFileResponse, error) {
	if err := s.DevEnv.LoadFile(req.SdlFilePath); err != nil {
		return nil, err
	}
	return &protos.LoadFileResponse{}, nil
}

func (s *WorkspaceService) UseSystem(_ context.Context, req *protos.UseSystemRequest) (*protos.UseSystemResponse, error) {
	if err := s.DevEnv.Use(req.SystemName); err != nil {
		return nil, err
	}
	return &protos.UseSystemResponse{}, nil
}

// Generator management

func (s *WorkspaceService) AddGenerator(_ context.Context, req *protos.AddGeneratorRequest) (*protos.AddGeneratorResponse, error) {
	gen := req.Generator
	if gen == nil {
		return nil, fmt.Errorf("generator is required")
	}
	genInfo := &runtime.Generator{Generator: gen}
	if err := s.DevEnv.AddGenerator(genInfo); err != nil {
		return nil, err
	}
	if req.ApplyFlows {
		s.DevEnv.EvaluateFlows("runtime")
	}
	return &protos.AddGeneratorResponse{Generator: gen}, nil
}

func (s *WorkspaceService) UpdateGenerator(_ context.Context, req *protos.UpdateGeneratorRequest) (*protos.UpdateGeneratorResponse, error) {
	gen := req.Generator
	if gen == nil {
		return nil, fmt.Errorf("generator is required")
	}
	if err := s.DevEnv.UpdateGenerator(gen.Name, gen.Rate); err != nil {
		return nil, err
	}
	if req.ApplyFlows {
		s.DevEnv.EvaluateFlows("runtime")
	}
	return &protos.UpdateGeneratorResponse{Generator: gen}, nil
}

func (s *WorkspaceService) DeleteGenerator(_ context.Context, req *protos.DeleteGeneratorRequest) (*protos.DeleteGeneratorResponse, error) {
	if err := s.DevEnv.RemoveGenerator(req.GeneratorName); err != nil {
		return nil, err
	}
	if req.ApplyFlows {
		s.DevEnv.EvaluateFlows("runtime")
	}
	return &protos.DeleteGeneratorResponse{}, nil
}

func (s *WorkspaceService) ListGenerators(_ context.Context, _ *protos.ListGeneratorsRequest) (*protos.ListGeneratorsResponse, error) {
	return &protos.ListGeneratorsResponse{Generators: s.DevEnv.ListGenerators()}, nil
}

func (s *WorkspaceService) StartGenerator(_ context.Context, req *protos.StartGeneratorRequest) (*protos.StartGeneratorResponse, error) {
	if err := s.DevEnv.StartGenerator(req.GeneratorName); err != nil {
		return nil, err
	}
	return &protos.StartGeneratorResponse{}, nil
}

func (s *WorkspaceService) StopGenerator(_ context.Context, req *protos.StopGeneratorRequest) (*protos.StopGeneratorResponse, error) {
	if err := s.DevEnv.StopGenerator(req.GeneratorName); err != nil {
		return nil, err
	}
	return &protos.StopGeneratorResponse{}, nil
}

func (s *WorkspaceService) StartAllGenerators(_ context.Context, _ *protos.StartAllGeneratorsRequest) (*protos.StartAllGeneratorsResponse, error) {
	if err := s.DevEnv.StartAllGenerators(); err != nil {
		return nil, err
	}
	return &protos.StartAllGeneratorsResponse{}, nil
}

func (s *WorkspaceService) StopAllGenerators(_ context.Context, _ *protos.StopAllGeneratorsRequest) (*protos.StopAllGeneratorsResponse, error) {
	if err := s.DevEnv.StopAllGenerators(); err != nil {
		return nil, err
	}
	return &protos.StopAllGeneratorsResponse{}, nil
}

// Metric management

func (s *WorkspaceService) AddMetric(_ context.Context, req *protos.AddMetricRequest) (*protos.AddMetricResponse, error) {
	m := req.Metric
	if m == nil {
		return nil, fmt.Errorf("metric is required")
	}
	spec := &runtime.Metric{Metric: m}
	if err := s.DevEnv.AddMetric(spec); err != nil {
		return nil, err
	}
	return &protos.AddMetricResponse{Metric: m}, nil
}

func (s *WorkspaceService) DeleteMetric(_ context.Context, req *protos.DeleteMetricRequest) (*protos.DeleteMetricResponse, error) {
	if err := s.DevEnv.RemoveMetric(req.MetricName); err != nil {
		return nil, err
	}
	return &protos.DeleteMetricResponse{}, nil
}

func (s *WorkspaceService) ListMetrics(_ context.Context, _ *protos.ListMetricsRequest) (*protos.ListMetricsResponse, error) {
	return &protos.ListMetricsResponse{Metrics: s.DevEnv.ListMetrics()}, nil
}

// Parameters

func (s *WorkspaceService) SetParameter(_ context.Context, req *protos.SetParameterRequest) (*protos.SetParameterResponse, error) {
	value := parseParameterValue(req.NewValue)
	if err := s.DevEnv.SetParameter(req.Path, value); err != nil {
		return nil, err
	}
	s.DevEnv.EvaluateFlows("runtime")
	return &protos.SetParameterResponse{}, nil
}

func (s *WorkspaceService) GetParameters(_ context.Context, _ *protos.GetParametersRequest) (*protos.GetParametersResponse, error) {
	// DevEnv doesn't have a GetParameters equivalent yet
	return &protos.GetParametersResponse{}, nil
}

// Diagram and flow analysis

func (s *WorkspaceService) GetSystemDiagram(_ context.Context, _ *protos.GetSystemDiagramRequest) (*protos.GetSystemDiagramResponse, error) {
	diagram, err := s.DevEnv.GetSystemDiagram()
	if err != nil {
		return nil, err
	}
	return &protos.GetSystemDiagramResponse{
		Diagram: services.ToProtoSystemDiagram(diagram),
	}, nil
}

func (s *WorkspaceService) EvaluateFlows(_ context.Context, req *protos.EvaluateFlowsRequest) (*protos.EvaluateFlowsResponse, error) {
	strategy := req.Strategy
	if strategy == "" {
		strategy = "runtime"
	}
	result, err := s.DevEnv.EvaluateFlows(strategy)
	if err != nil {
		return nil, err
	}
	return &protos.EvaluateFlowsResponse{
		Strategy:       strategy,
		Status:         "applied",
		ComponentRates: result.Flows.ComponentRates,
	}, nil
}

func (s *WorkspaceService) GetFlowState(_ context.Context, _ *protos.GetFlowStateRequest) (*protos.GetFlowStateResponse, error) {
	rates, strategy := s.DevEnv.GetFlowState()
	return &protos.GetFlowStateResponse{
		State: &protos.FlowState{
			Strategy: strategy,
			Rates:    rates,
		},
	}, nil
}

func (s *WorkspaceService) BatchSetParameters(_ context.Context, req *protos.BatchSetParametersRequest) (*protos.BatchSetParametersResponse, error) {
	updates := make(map[string]any)
	for _, u := range req.Updates {
		updates[u.Path] = parseParameterValue(u.NewValue)
	}
	if err := s.DevEnv.BatchSetParameters(updates); err != nil {
		return &protos.BatchSetParametersResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	return &protos.BatchSetParametersResponse{Success: true}, nil
}

func (s *WorkspaceService) ExecuteTrace(_ context.Context, req *protos.ExecuteTraceRequest) (*protos.ExecuteTraceResponse, error) {
	traceData, err := s.DevEnv.ExecuteTrace(req.Component, req.Method)
	if err != nil {
		return nil, err
	}
	td := &protos.TraceData{
		System:     traceData.System,
		EntryPoint: traceData.EntryPoint,
	}
	for _, evt := range traceData.Events {
		te := &protos.TraceEvent{
			Kind:      string(evt.Kind),
			Timestamp: float64(evt.Timestamp),
			Duration:  float64(evt.Duration),
		}
		if evt.Component != nil {
			te.Component = evt.Component.ID()
		}
		if evt.Method != nil {
			te.Method = evt.Method.Name.Value
		}
		td.Events = append(td.Events, te)
	}
	return &protos.ExecuteTraceResponse{TraceData: td}, nil
}

func (s *WorkspaceService) TraceAllPaths(_ context.Context, req *protos.TraceAllPathsRequest) (*protos.TraceAllPathsResponse, error) {
	data, err := s.DevEnv.TraceAllPaths(req.Component, req.Method, req.MaxDepth)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return &protos.TraceAllPathsResponse{}, nil
	}
	// Convert runtime AllPathsTraceData to proto
	return &protos.TraceAllPathsResponse{
		TraceData: data.ToProto(),
	}, nil
}

func (s *WorkspaceService) GetUtilization(_ context.Context, _ *protos.GetUtilizationRequest) (*protos.GetUtilizationResponse, error) {
	utils := s.DevEnv.GetUtilization()
	resp := &protos.GetUtilizationResponse{}
	for _, u := range utils {
		for _, info := range u.Infos {
			resp.Utilizations = append(resp.Utilizations, &protos.UtilizationInfo{
				ComponentPath: u.Component,
				Utilization:  info.Utilization,
				Capacity:     float64(info.Capacity),
				IsBottleneck: info.IsBottleneck,
			})
		}
	}
	return resp, nil
}

func (s *WorkspaceService) QueryMetrics(_ context.Context, req *protos.QueryMetricsRequest) (*protos.QueryMetricsResponse, error) {
	opts := runtime.QueryOptions{
		StartTime: time.Unix(int64(req.StartTime), 0),
		EndTime:   time.Unix(int64(req.EndTime), 0),
		Limit:     int(req.Limit),
	}
	result, err := s.DevEnv.QueryMetrics(req.MetricName, opts)
	if err != nil {
		return nil, err
	}
	resp := &protos.QueryMetricsResponse{}
	for _, p := range result.Points {
		resp.Points = append(resp.Points, &protos.MetricPoint{
			Timestamp: float64(p.Timestamp.Unix()),
			Value:     p.Value,
		})
	}
	return resp, nil
}

// parseParameterValue converts a string value to the most appropriate Go type.
func parseParameterValue(s string) any {
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseBool(s); err == nil {
		return v
	}
	return s
}
