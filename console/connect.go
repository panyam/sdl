package console

import (
	"context"

	"connectrpc.com/connect"
	v1 "github.com/panyam/sdl/gen/go/sdl/v1"
)

// ConnectCanvasServiceAdapter adapts the gRPC CanvasService to Connect's interface
type ConnectCanvasServiceAdapter struct {
	svc *CanvasService
}

func NewConnectCanvasServiceAdapter(svc *CanvasService) *ConnectCanvasServiceAdapter {
	return &ConnectCanvasServiceAdapter{svc: svc}
}

func (a *ConnectCanvasServiceAdapter) CreateCanvas(ctx context.Context, req *connect.Request[v1.CreateCanvasRequest]) (*connect.Response[v1.CreateCanvasResponse], error) {
	resp, err := a.svc.CreateCanvas(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) ListCanvases(ctx context.Context, req *connect.Request[v1.ListCanvasesRequest]) (*connect.Response[v1.ListCanvasesResponse], error) {
	resp, err := a.svc.ListCanvases(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) GetCanvas(ctx context.Context, req *connect.Request[v1.GetCanvasRequest]) (*connect.Response[v1.GetCanvasResponse], error) {
	resp, err := a.svc.GetCanvas(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) LoadFile(ctx context.Context, req *connect.Request[v1.LoadFileRequest]) (*connect.Response[v1.LoadFileResponse], error) {
	resp, err := a.svc.LoadFile(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) UseSystem(ctx context.Context, req *connect.Request[v1.UseSystemRequest]) (*connect.Response[v1.UseSystemResponse], error) {
	resp, err := a.svc.UseSystem(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) DeleteCanvas(ctx context.Context, req *connect.Request[v1.DeleteCanvasRequest]) (*connect.Response[v1.DeleteCanvasResponse], error) {
	resp, err := a.svc.DeleteCanvas(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) AddGenerator(ctx context.Context, req *connect.Request[v1.AddGeneratorRequest]) (*connect.Response[v1.AddGeneratorResponse], error) {
	resp, err := a.svc.AddGenerator(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) StartAllGenerators(ctx context.Context, req *connect.Request[v1.StartAllGeneratorsRequest]) (*connect.Response[v1.StartAllGeneratorsResponse], error) {
	resp, err := a.svc.StartAllGenerators(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) StopAllGenerators(ctx context.Context, req *connect.Request[v1.StopAllGeneratorsRequest]) (*connect.Response[v1.StopAllGeneratorsResponse], error) {
	resp, err := a.svc.StopAllGenerators(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) ListGenerators(ctx context.Context, req *connect.Request[v1.ListGeneratorsRequest]) (*connect.Response[v1.ListGeneratorsResponse], error) {
	resp, err := a.svc.ListGenerators(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) GetGenerator(ctx context.Context, req *connect.Request[v1.GetGeneratorRequest]) (*connect.Response[v1.GetGeneratorResponse], error) {
	resp, err := a.svc.GetGenerator(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) UpdateGenerator(ctx context.Context, req *connect.Request[v1.UpdateGeneratorRequest]) (*connect.Response[v1.UpdateGeneratorResponse], error) {
	resp, err := a.svc.UpdateGenerator(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) DeleteGenerator(ctx context.Context, req *connect.Request[v1.DeleteGeneratorRequest]) (*connect.Response[v1.DeleteGeneratorResponse], error) {
	resp, err := a.svc.DeleteGenerator(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) StartGenerator(ctx context.Context, req *connect.Request[v1.StartGeneratorRequest]) (*connect.Response[v1.StartGeneratorResponse], error) {
	resp, err := a.svc.StartGenerator(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) StopGenerator(ctx context.Context, req *connect.Request[v1.StopGeneratorRequest]) (*connect.Response[v1.StopGeneratorResponse], error) {
	resp, err := a.svc.StopGenerator(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) AddMetric(ctx context.Context, req *connect.Request[v1.AddMetricRequest]) (*connect.Response[v1.AddMetricResponse], error) {
	resp, err := a.svc.AddMetric(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) ListMetrics(ctx context.Context, req *connect.Request[v1.ListMetricsRequest]) (*connect.Response[v1.ListMetricsResponse], error) {
	resp, err := a.svc.ListMetrics(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) DeleteMetric(ctx context.Context, req *connect.Request[v1.DeleteMetricRequest]) (*connect.Response[v1.DeleteMetricResponse], error) {
	resp, err := a.svc.DeleteMetric(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) QueryMetrics(ctx context.Context, req *connect.Request[v1.QueryMetricsRequest]) (*connect.Response[v1.QueryMetricsResponse], error) {
	resp, err := a.svc.QueryMetrics(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) GetParameters(ctx context.Context, req *connect.Request[v1.GetParametersRequest]) (*connect.Response[v1.GetParametersResponse], error) {
	resp, err := a.svc.GetParameters(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) SetParameter(ctx context.Context, req *connect.Request[v1.SetParameterRequest]) (*connect.Response[v1.SetParameterResponse], error) {
	resp, err := a.svc.SetParameter(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) ExecuteTrace(ctx context.Context, req *connect.Request[v1.ExecuteTraceRequest]) (*connect.Response[v1.ExecuteTraceResponse], error) {
	resp, err := a.svc.ExecuteTrace(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectCanvasServiceAdapter) GetSystemDiagram(ctx context.Context, req *connect.Request[v1.GetSystemDiagramRequest]) (*connect.Response[v1.GetSystemDiagramResponse], error) {
	resp, err := a.svc.GetSystemDiagram(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}
