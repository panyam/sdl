package console

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	protos "github.com/panyam/sdl/gen/go/sdl/v1"
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
	return
}

func (s *CanvasService) StartAllGenerators(ctx context.Context, req *protos.StartAllGeneratorsRequest) (resp *protos.StartAllGeneratorsResponse, err error) {
	slog.Info("StartAllGenerators Request", "req", req)
	resp = &protos.StartAllGeneratorsResponse{}
	return
}

func (s *CanvasService) StopAllGenerators(ctx context.Context, req *protos.StopAllGeneratorsRequest) (resp *protos.StopAllGeneratorsResponse, err error) {
	slog.Info("StopAllGenerators  Request", "req", req)
	resp = &protos.StopAllGeneratorsResponse{}
	return
}

func (s *CanvasService) ResumeGenerator(ctx context.Context, req *protos.ResumeGeneratorRequest) (resp *protos.ResumeGeneratorResponse, err error) {
	slog.Info("ResumeGenerator Request", "req", req)
	resp = &protos.ResumeGeneratorResponse{}
	return
}

func (s *CanvasService) PauseGenerator(ctx context.Context, req *protos.PauseGeneratorRequest) (resp *protos.PauseGeneratorResponse, err error) {
	slog.Info("PauseGenerator Request", "req", req)
	resp = &protos.PauseGeneratorResponse{}
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
