package devenvbe

import (
	"context"
	"fmt"
	"strconv"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	protoservices "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/panyam/sdl/lib/loader"
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
	genInfo := &services.GeneratorInfo{Generator: gen}
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
	if err := s.DevEnv.RemoveGenerator(req.GeneratorId); err != nil {
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
	if err := s.DevEnv.StartGenerator(req.GeneratorId); err != nil {
		return nil, err
	}
	return &protos.StartGeneratorResponse{}, nil
}

func (s *WorkspaceService) StopGenerator(_ context.Context, req *protos.StopGeneratorRequest) (*protos.StopGeneratorResponse, error) {
	if err := s.DevEnv.StopGenerator(req.GeneratorId); err != nil {
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
	spec := &services.MetricSpec{Metric: m}
	if err := s.DevEnv.AddMetric(spec); err != nil {
		return nil, err
	}
	return &protos.AddMetricResponse{Metric: m}, nil
}

func (s *WorkspaceService) DeleteMetric(_ context.Context, req *protos.DeleteMetricRequest) (*protos.DeleteMetricResponse, error) {
	if err := s.DevEnv.RemoveMetric(req.MetricId); err != nil {
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
