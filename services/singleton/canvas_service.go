//go:build js && wasm
// +build js,wasm

package singleton

import (
	"context"
	"fmt"
	"sync"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	wasmservices "github.com/panyam/sdl/gen/wasm/go/sdl/v1/services"
	"github.com/panyam/sdl/lib/loader"
	"github.com/panyam/sdl/lib/runtime"
	"github.com/panyam/sdl/services"
)

// SingletonCanvasService provides a single in-memory canvas for WASM mode.
// Unlike the server-side CanvasService which manages multiple canvases,
// this service manages a single canvas instance initialized from the browser.
type SingletonCanvasService struct {
	canvas     *services.Canvas
	canvasLock sync.RWMutex
	fileSystem loader.FileSystem
}

// NewSingletonCanvasService creates a new singleton canvas service
func NewSingletonCanvasService(fs loader.FileSystem) *SingletonCanvasService {
	svc := &SingletonCanvasService{
		fileSystem: fs,
	}
	// Initialize a default canvas
	svc.initializeCanvas("default")
	return svc
}

func (s *SingletonCanvasService) initializeCanvas(id string) {
	fsResolver := loader.NewFileSystemResolver(s.fileSystem)
	sdlLoader := loader.NewLoader(nil, fsResolver, 10)
	r := runtime.NewRuntime(sdlLoader)
	s.canvas = services.NewCanvas(id, r)
}

// GetInternalCanvas returns the internal singleton canvas (for internal use)
func (s *SingletonCanvasService) GetInternalCanvas() *services.Canvas {
	s.canvasLock.RLock()
	defer s.canvasLock.RUnlock()
	return s.canvas
}

// CreateCanvas creates a new canvas (replaces the singleton)
func (s *SingletonCanvasService) CreateCanvas(ctx context.Context, req *protos.CreateCanvasRequest) (*protos.CreateCanvasResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	canvasID := "default"
	if req.Canvas != nil && req.Canvas.Id != "" {
		canvasID = req.Canvas.Id
	}

	s.initializeCanvas(canvasID)

	return &protos.CreateCanvasResponse{
		Canvas: s.canvas.ToProto(),
	}, nil
}

func (s *SingletonCanvasService) UpdateCanvas(ctx context.Context, req *protos.UpdateCanvasRequest) (*protos.UpdateCanvasResponse, error) {
	return nil, nil
}

// ListCanvases returns the single canvas
func (s *SingletonCanvasService) ListCanvases(ctx context.Context, req *protos.ListCanvasesRequest) (*protos.ListCanvasesResponse, error) {
	s.canvasLock.RLock()
	defer s.canvasLock.RUnlock()

	return &protos.ListCanvasesResponse{
		Canvases: []*protos.Canvas{s.canvas.ToProto()},
	}, nil
}

// GetCanvas returns the singleton canvas details (implements CanvasServiceServer interface)
func (s *SingletonCanvasService) GetCanvas(ctx context.Context, req *protos.GetCanvasRequest) (*protos.GetCanvasResponse, error) {
	s.canvasLock.RLock()
	defer s.canvasLock.RUnlock()

	return &protos.GetCanvasResponse{
		Canvas: s.canvas.ToProto(),
	}, nil
}

// LoadFile loads an SDL file into the canvas
func (s *SingletonCanvasService) LoadFile(ctx context.Context, req *protos.LoadFileRequest) (*protos.LoadFileResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	err := s.canvas.Load(req.SdlFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load file: %w", err)
	}

	return &protos.LoadFileResponse{}, nil
}

// UseSystem activates a system in the canvas
func (s *SingletonCanvasService) UseSystem(ctx context.Context, req *protos.UseSystemRequest) (*protos.UseSystemResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	err := s.canvas.Use(req.SystemName)
	if err != nil {
		return nil, fmt.Errorf("failed to use system: %w", err)
	}

	return &protos.UseSystemResponse{}, nil
}

// DeleteCanvas resets the singleton canvas
func (s *SingletonCanvasService) DeleteCanvas(ctx context.Context, req *protos.DeleteCanvasRequest) (*protos.DeleteCanvasResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	s.initializeCanvas("default")
	return &protos.DeleteCanvasResponse{}, nil
}

// ResetCanvas resets the canvas state
func (s *SingletonCanvasService) ResetCanvas(ctx context.Context, req *protos.ResetCanvasRequest) (*protos.ResetCanvasResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	err := s.canvas.Reset()
	if err != nil {
		return &protos.ResetCanvasResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to reset canvas: %v", err),
		}, nil
	}

	return &protos.ResetCanvasResponse{
		Success: true,
		Message: "Canvas reset successfully",
	}, nil
}

// AddGenerator adds a generator to the canvas
func (s *SingletonCanvasService) AddGenerator(ctx context.Context, req *protos.AddGeneratorRequest) (*protos.AddGeneratorResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	gen := &services.GeneratorInfo{Generator: req.Generator}
	err := s.canvas.AddGenerator(gen)
	if err != nil {
		return nil, fmt.Errorf("failed to add generator: %w", err)
	}

	return &protos.AddGeneratorResponse{
		Generator: req.Generator,
	}, nil
}

// ListGenerators lists all generators
func (s *SingletonCanvasService) ListGenerators(ctx context.Context, req *protos.ListGeneratorsRequest) (*protos.ListGeneratorsResponse, error) {
	s.canvasLock.RLock()
	defer s.canvasLock.RUnlock()

	gens := s.canvas.ListGenerators()
	protoGens := make([]*protos.Generator, 0, len(gens))
	for _, g := range gens {
		if g != nil && g.Generator != nil {
			protoGens = append(protoGens, g.Generator)
		}
	}

	return &protos.ListGeneratorsResponse{
		Generators: protoGens,
	}, nil
}

// GetGenerator gets a specific generator
func (s *SingletonCanvasService) GetGenerator(ctx context.Context, req *protos.GetGeneratorRequest) (*protos.GetGeneratorResponse, error) {
	s.canvasLock.RLock()
	defer s.canvasLock.RUnlock()

	gen := s.canvas.GetGenerator(req.GeneratorId)
	if gen == nil {
		return nil, fmt.Errorf("generator not found: %s", req.GeneratorId)
	}

	return &protos.GetGeneratorResponse{
		Generator: gen.Generator,
	}, nil
}

// UpdateGenerator updates a generator
func (s *SingletonCanvasService) UpdateGenerator(ctx context.Context, req *protos.UpdateGeneratorRequest) (*protos.UpdateGeneratorResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	err := s.canvas.UpdateGenerator(req.Generator)
	if err != nil {
		return nil, fmt.Errorf("failed to update generator: %w", err)
	}

	return &protos.UpdateGeneratorResponse{
		Generator: req.Generator,
	}, nil
}

// StartGenerator starts a generator
func (s *SingletonCanvasService) StartGenerator(ctx context.Context, req *protos.StartGeneratorRequest) (*protos.StartGeneratorResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	err := s.canvas.StartGenerator(req.GeneratorId)
	if err != nil {
		return nil, fmt.Errorf("failed to start generator: %w", err)
	}

	return &protos.StartGeneratorResponse{}, nil
}

// StopGenerator stops a generator
func (s *SingletonCanvasService) StopGenerator(ctx context.Context, req *protos.StopGeneratorRequest) (*protos.StopGeneratorResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	err := s.canvas.StopGenerator(req.GeneratorId)
	if err != nil {
		return nil, fmt.Errorf("failed to stop generator: %w", err)
	}

	return &protos.StopGeneratorResponse{}, nil
}

// DeleteGenerator deletes a generator
func (s *SingletonCanvasService) DeleteGenerator(ctx context.Context, req *protos.DeleteGeneratorRequest) (*protos.DeleteGeneratorResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	err := s.canvas.RemoveGenerator(req.GeneratorId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete generator: %w", err)
	}

	return &protos.DeleteGeneratorResponse{}, nil
}

// StartAllGenerators starts all generators
func (s *SingletonCanvasService) StartAllGenerators(ctx context.Context, req *protos.StartAllGeneratorsRequest) (*protos.StartAllGeneratorsResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	result, err := s.canvas.StartAllGenerators()
	if err != nil {
		return nil, fmt.Errorf("failed to start all generators: %w", err)
	}

	return &protos.StartAllGeneratorsResponse{
		TotalGenerators:     int32(result.TotalGenerators),
		StartedCount:        int32(result.ProcessedCount),
		AlreadyRunningCount: int32(result.AlreadyInStateCount),
		FailedCount:         int32(result.FailedCount),
		FailedIds:           result.FailedIDs,
	}, nil
}

// StopAllGenerators stops all generators
func (s *SingletonCanvasService) StopAllGenerators(ctx context.Context, req *protos.StopAllGeneratorsRequest) (*protos.StopAllGeneratorsResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	result, err := s.canvas.StopAllGenerators()
	if err != nil {
		return nil, fmt.Errorf("failed to stop all generators: %w", err)
	}

	return &protos.StopAllGeneratorsResponse{
		TotalGenerators:     int32(result.TotalGenerators),
		StoppedCount:        int32(result.ProcessedCount),
		AlreadyStoppedCount: int32(result.AlreadyInStateCount),
		FailedCount:         int32(result.FailedCount),
		FailedIds:           result.FailedIDs,
	}, nil
}

// GetSystemDiagram returns the system diagram
func (s *SingletonCanvasService) GetSystemDiagram(ctx context.Context, req *protos.GetSystemDiagramRequest) (*protos.GetSystemDiagramResponse, error) {
	s.canvasLock.RLock()
	defer s.canvasLock.RUnlock()

	diagram, err := s.canvas.GetSystemDiagram()
	if err != nil {
		return nil, fmt.Errorf("failed to get system diagram: %w", err)
	}

	return &protos.GetSystemDiagramResponse{
		Diagram: services.ToProtoSystemDiagram(diagram),
	}, nil
}

// EvaluateFlows evaluates flow analysis
func (s *SingletonCanvasService) EvaluateFlows(ctx context.Context, req *protos.EvaluateFlowsRequest) (*protos.EvaluateFlowsResponse, error) {
	s.canvasLock.Lock()
	defer s.canvasLock.Unlock()

	result, err := s.canvas.EvaluateFlowWithStrategy(req.Strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate flows: %w", err)
	}

	flowEdges := make([]*protos.FlowEdge, len(result.Flows.Edges))
	for i, edge := range result.Flows.Edges {
		flowEdges[i] = &protos.FlowEdge{
			FromComponent: edge.From.Component,
			FromMethod:    edge.From.Method,
			ToComponent:   edge.To.Component,
			ToMethod:      edge.To.Method,
			Rate:          edge.Rate,
		}
	}

	return &protos.EvaluateFlowsResponse{
		Strategy:       result.Strategy,
		Status:         string(result.Status),
		Iterations:     int32(result.Iterations),
		Warnings:       result.Warnings,
		ComponentRates: result.Flows.ComponentRates,
		FlowEdges:      flowEdges,
	}, nil
}

// Stub implementations for remaining methods

func (s *SingletonCanvasService) ExecuteTrace(ctx context.Context, req *protos.ExecuteTraceRequest) (*protos.ExecuteTraceResponse, error) {
	return nil, fmt.Errorf("ExecuteTrace not implemented in WASM")
}

func (s *SingletonCanvasService) TraceAllPaths(ctx context.Context, req *protos.TraceAllPathsRequest) (*protos.TraceAllPathsResponse, error) {
	return nil, fmt.Errorf("TraceAllPaths not implemented in WASM")
}

func (s *SingletonCanvasService) SetParameter(ctx context.Context, req *protos.SetParameterRequest) (*protos.SetParameterResponse, error) {
	return nil, fmt.Errorf("SetParameter not implemented in WASM")
}

func (s *SingletonCanvasService) GetParameters(ctx context.Context, req *protos.GetParametersRequest) (*protos.GetParametersResponse, error) {
	return nil, fmt.Errorf("GetParameters not implemented in WASM")
}

func (s *SingletonCanvasService) BatchSetParameters(ctx context.Context, req *protos.BatchSetParametersRequest) (*protos.BatchSetParametersResponse, error) {
	return nil, fmt.Errorf("BatchSetParameters not implemented in WASM")
}

func (s *SingletonCanvasService) GetFlowState(ctx context.Context, req *protos.GetFlowStateRequest) (*protos.GetFlowStateResponse, error) {
	return nil, fmt.Errorf("GetFlowState not implemented in WASM")
}

func (s *SingletonCanvasService) AddMetric(ctx context.Context, req *protos.AddMetricRequest) (*protos.AddMetricResponse, error) {
	return nil, fmt.Errorf("AddMetric not implemented in WASM")
}

func (s *SingletonCanvasService) DeleteMetric(ctx context.Context, req *protos.DeleteMetricRequest) (*protos.DeleteMetricResponse, error) {
	return nil, fmt.Errorf("DeleteMetric not implemented in WASM")
}

func (s *SingletonCanvasService) ListMetrics(ctx context.Context, req *protos.ListMetricsRequest) (*protos.ListMetricsResponse, error) {
	return nil, fmt.Errorf("ListMetrics not implemented in WASM")
}

func (s *SingletonCanvasService) QueryMetrics(ctx context.Context, req *protos.QueryMetricsRequest) (*protos.QueryMetricsResponse, error) {
	return nil, fmt.Errorf("QueryMetrics not implemented in WASM")
}

func (s *SingletonCanvasService) StreamMetrics(req *protos.StreamMetricsRequest, stream wasmservices.StreamMetrics_ServerStream) error {
	return fmt.Errorf("StreamMetrics not implemented in WASM")
}

func (s *SingletonCanvasService) GetUtilization(ctx context.Context, req *protos.GetUtilizationRequest) (*protos.GetUtilizationResponse, error) {
	return nil, fmt.Errorf("GetUtilization not implemented in WASM")
}
