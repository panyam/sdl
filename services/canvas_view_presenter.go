//go:build js && wasm
// +build js,wasm

package services

import (
	"context"
	"fmt"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	wasmservices "github.com/panyam/sdl/gen/wasm/go/sdl/v1/services"
)

// CanvasViewPresenter is the P in Model-View-Presenter.
// It orchestrates between the browser views (CanvasDashboardPage) and
// the backend services (CanvasService, SystemsService).
//
// The presenter handles UI logic, dispatches to services, and calls
// back to the browser to update the view.
type CanvasViewPresenter struct {
	// Service dependencies
	CanvasService  wasmservices.CanvasServiceServer
	SystemsService wasmservices.SystemsServiceServer

	// Browser-provided page client (for callbacks to browser)
	DashboardPage *wasmservices.CanvasDashboardPageClient

	// State
	initialized bool
	canvasID    string
}

// NewCanvasViewPresenter creates a new presenter instance.
func NewCanvasViewPresenter() *CanvasViewPresenter {
	return &CanvasViewPresenter{
		canvasID: "default",
	}
}

// === Initialization ===

// Initialize is called when the dashboard is first loaded.
func (p *CanvasViewPresenter) Initialize(ctx context.Context, req *protos.InitializePresenterRequest) (*protos.InitializePresenterResponse, error) {
	p.initialized = true

	if req.CanvasId != "" {
		p.canvasID = req.CanvasId
	}

	// Get list of available systems
	systemsResp, err := p.SystemsService.ListSystems(ctx, &protos.ListSystemsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}

	return &protos.InitializePresenterResponse{
		Success:          true,
		CanvasId:         p.canvasID,
		AvailableSystems: systemsResp.Systems,
	}, nil
}

// ClientReady is called by the browser after the UI is fully ready for updates.
func (p *CanvasViewPresenter) ClientReady(ctx context.Context, req *protos.ClientReadyRequest) (*protos.ClientReadyResponse, error) {
	if req.CanvasId != "" {
		p.canvasID = req.CanvasId
	}

	// Get current canvas state and push to browser
	canvasResp, err := p.CanvasService.GetCanvas(ctx, &protos.GetCanvasRequest{
		Id: p.canvasID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get canvas: %w", err)
	}

	// Update the browser with initial diagram
	diagramResp, err := p.CanvasService.GetSystemDiagram(ctx, &protos.GetSystemDiagramRequest{
		CanvasId: p.canvasID,
	})
	if err == nil && diagramResp.Diagram != nil && p.DashboardPage != nil {
		p.DashboardPage.UpdateDiagram(ctx, &protos.UpdateDiagramRequest{
			Diagram: diagramResp.Diagram,
		})
	}

	// Get generators and update browser
	gensResp, err := p.CanvasService.ListGenerators(ctx, &protos.ListGeneratorsRequest{
		CanvasId: p.canvasID,
	})
	if err == nil && p.DashboardPage != nil {
		p.DashboardPage.SetGeneratorList(ctx, &protos.SetGeneratorListRequest{
			Generators: gensResp.Generators,
		})
	}

	// Get metrics and update browser
	metricsResp, err := p.CanvasService.ListMetrics(ctx, &protos.ListMetricsRequest{
		CanvasId: p.canvasID,
	})
	if err == nil && p.DashboardPage != nil {
		p.DashboardPage.SetMetricsList(ctx, &protos.SetMetricsListRequest{
			Metrics: metricsResp.Metrics,
		})
	}

	return &protos.ClientReadyResponse{
		Success: true,
		Canvas:  canvasResp.Canvas,
	}, nil
}

// === File Operations ===

// FileSelected is called when the user selects a file in the explorer.
func (p *CanvasViewPresenter) FileSelected(ctx context.Context, req *protos.FileSelectedRequest) (*protos.FileSelectedResponse, error) {
	// Load the selected SDL file
	_, err := p.CanvasService.LoadFile(ctx, &protos.LoadFileRequest{
		CanvasId:    p.canvasID,
		SdlFilePath: req.FilePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load file: %w", err)
	}

	// Refresh the diagram
	p.refreshDiagram(ctx)

	return &protos.FileSelectedResponse{}, nil
}

// FileSaved is called when the user saves a file.
func (p *CanvasViewPresenter) FileSaved(ctx context.Context, req *protos.FileSavedRequest) (*protos.FileSavedResponse, error) {
	// Reload the file if it's the current one
	_, err := p.CanvasService.LoadFile(ctx, &protos.LoadFileRequest{
		CanvasId:    p.canvasID,
		SdlFilePath: req.FilePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to reload file: %w", err)
	}

	// Refresh the diagram
	p.refreshDiagram(ctx)

	return &protos.FileSavedResponse{}, nil
}

// === System Operations ===

// UseSystem is called when the user wants to load/use a system.
func (p *CanvasViewPresenter) UseSystem(ctx context.Context, req *protos.UseSystemRequest) (*protos.UseSystemResponse, error) {
	resp, err := p.CanvasService.UseSystem(ctx, req)
	if err != nil {
		return nil, err
	}

	// Refresh UI after system change
	p.refreshDiagram(ctx)
	p.refreshGenerators(ctx)

	return resp, nil
}

// === Generator Operations ===

// AddGenerator is called when the user adds a generator.
func (p *CanvasViewPresenter) AddGenerator(ctx context.Context, req *protos.AddGeneratorRequest) (*protos.AddGeneratorResponse, error) {
	resp, err := p.CanvasService.AddGenerator(ctx, req)
	if err != nil {
		return nil, err
	}

	// Update browser with new generator state
	if p.DashboardPage != nil && resp.Generator != nil {
		status := "stopped"
		if resp.Generator.Enabled {
			status = "running"
		}
		p.DashboardPage.UpdateGeneratorState(ctx, &protos.UpdateGeneratorStateRequest{
			GeneratorId: resp.Generator.Id,
			Enabled:     resp.Generator.Enabled,
			Rate:        resp.Generator.Rate,
			Status:      status,
		})
	}

	// Re-evaluate flows if requested
	if req.ApplyFlows {
		p.evaluateAndUpdateFlows(ctx)
	}

	return resp, nil
}

// DeleteGenerator is called when the user removes a generator.
func (p *CanvasViewPresenter) DeleteGenerator(ctx context.Context, req *protos.DeleteGeneratorRequest) (*protos.DeleteGeneratorResponse, error) {
	resp, err := p.CanvasService.DeleteGenerator(ctx, req)
	if err != nil {
		return nil, err
	}

	// Refresh generators list
	p.refreshGenerators(ctx)

	// Re-evaluate flows if requested
	if req.ApplyFlows {
		p.evaluateAndUpdateFlows(ctx)
	}

	return resp, nil
}

// UpdateGenerator is called when the user updates a generator rate.
func (p *CanvasViewPresenter) UpdateGenerator(ctx context.Context, req *protos.UpdateGeneratorRequest) (*protos.UpdateGeneratorResponse, error) {
	resp, err := p.CanvasService.UpdateGenerator(ctx, req)
	if err != nil {
		return nil, err
	}

	// Update browser with modified generator state
	if p.DashboardPage != nil && resp.Generator != nil {
		status := "stopped"
		if resp.Generator.Enabled {
			status = "running"
		}
		p.DashboardPage.UpdateGeneratorState(ctx, &protos.UpdateGeneratorStateRequest{
			GeneratorId: resp.Generator.Id,
			Enabled:     resp.Generator.Enabled,
			Rate:        resp.Generator.Rate,
			Status:      status,
		})
	}

	// Re-evaluate flows if requested
	if req.ApplyFlows {
		p.evaluateAndUpdateFlows(ctx)
	}

	return resp, nil
}

// StartGenerator is called when the user starts a generator.
func (p *CanvasViewPresenter) StartGenerator(ctx context.Context, req *protos.StartGeneratorRequest) (*protos.StartGeneratorResponse, error) {
	req.CanvasId = p.canvasID
	resp, err := p.CanvasService.StartGenerator(ctx, req)
	if err != nil {
		return nil, err
	}

	// Refresh the generator state
	p.refreshGenerators(ctx)

	return resp, nil
}

// StopGenerator is called when the user stops a generator.
func (p *CanvasViewPresenter) StopGenerator(ctx context.Context, req *protos.StopGeneratorRequest) (*protos.StopGeneratorResponse, error) {
	req.CanvasId = p.canvasID
	resp, err := p.CanvasService.StopGenerator(ctx, req)
	if err != nil {
		return nil, err
	}

	// Refresh the generator state
	p.refreshGenerators(ctx)

	return resp, nil
}

// StartAllGenerators is called when the user starts all generators.
func (p *CanvasViewPresenter) StartAllGenerators(ctx context.Context, req *protos.StartAllGeneratorsRequest) (*protos.StartAllGeneratorsResponse, error) {
	req.CanvasId = p.canvasID
	resp, err := p.CanvasService.StartAllGenerators(ctx, req)
	if err != nil {
		return nil, err
	}

	// Refresh all generator states
	p.refreshGenerators(ctx)

	return resp, nil
}

// StopAllGenerators is called when the user stops all generators.
func (p *CanvasViewPresenter) StopAllGenerators(ctx context.Context, req *protos.StopAllGeneratorsRequest) (*protos.StopAllGeneratorsResponse, error) {
	req.CanvasId = p.canvasID
	resp, err := p.CanvasService.StopAllGenerators(ctx, req)
	if err != nil {
		return nil, err
	}

	// Refresh all generator states
	p.refreshGenerators(ctx)

	return resp, nil
}

// === Metric Operations ===

// AddMetric is called when the user adds a metric to track.
func (p *CanvasViewPresenter) AddMetric(ctx context.Context, req *protos.AddMetricRequest) (*protos.AddMetricResponse, error) {
	resp, err := p.CanvasService.AddMetric(ctx, req)
	if err != nil {
		return nil, err
	}

	// Refresh metrics list
	p.refreshMetrics(ctx)

	return resp, nil
}

// DeleteMetric is called when the user removes a metric.
func (p *CanvasViewPresenter) DeleteMetric(ctx context.Context, req *protos.DeleteMetricRequest) (*protos.DeleteMetricResponse, error) {
	resp, err := p.CanvasService.DeleteMetric(ctx, req)
	if err != nil {
		return nil, err
	}

	// Refresh metrics list
	p.refreshMetrics(ctx)

	return resp, nil
}

// === Parameter Operations ===

// SetParameter is called when the user changes a component parameter.
func (p *CanvasViewPresenter) SetParameter(ctx context.Context, req *protos.SetParameterRequest) (*protos.SetParameterResponse, error) {
	req.CanvasId = p.canvasID
	resp, err := p.CanvasService.SetParameter(ctx, req)
	if err != nil {
		return nil, err
	}

	// Re-evaluate flows after parameter change
	p.evaluateAndUpdateFlows(ctx)

	return resp, nil
}

// === Flow Operations ===

// EvaluateFlows is called when the user wants to evaluate flows.
func (p *CanvasViewPresenter) EvaluateFlows(ctx context.Context, req *protos.EvaluateFlowsRequest) (*protos.EvaluateFlowsResponse, error) {
	req.CanvasId = p.canvasID
	resp, err := p.CanvasService.EvaluateFlows(ctx, req)
	if err != nil {
		return nil, err
	}

	// Update browser with flow rates
	if p.DashboardPage != nil {
		p.DashboardPage.UpdateFlowRates(ctx, &protos.UpdateFlowRatesRequest{
			Rates:    resp.ComponentRates,
			Strategy: resp.Strategy,
		})
	}

	return resp, nil
}

// === Diagram Interactions ===

// DiagramComponentClicked is called when the user clicks on a component.
func (p *CanvasViewPresenter) DiagramComponentClicked(ctx context.Context, req *protos.DiagramComponentClickedRequest) (*protos.DiagramComponentClickedResponse, error) {
	// Highlight the clicked component
	if p.DashboardPage != nil {
		p.DashboardPage.HighlightComponents(ctx, &protos.HighlightComponentsRequest{
			ComponentIds: []string{req.ComponentName},
		})
	}

	return &protos.DiagramComponentClickedResponse{
		Success: true,
	}, nil
}

// DiagramComponentHovered is called when the user hovers over a component.
func (p *CanvasViewPresenter) DiagramComponentHovered(ctx context.Context, req *protos.DiagramComponentHoveredRequest) (*protos.DiagramComponentHoveredResponse, error) {
	// Could show tooltips or highlight connections
	return &protos.DiagramComponentHoveredResponse{
		Success: true,
	}, nil
}

// === Helper methods ===

func (p *CanvasViewPresenter) refreshDiagram(ctx context.Context) {
	if p.DashboardPage == nil || p.CanvasService == nil {
		return
	}

	diagramResp, err := p.CanvasService.GetSystemDiagram(ctx, &protos.GetSystemDiagramRequest{
		CanvasId: p.canvasID,
	})
	if err == nil && diagramResp.Diagram != nil {
		p.DashboardPage.UpdateDiagram(ctx, &protos.UpdateDiagramRequest{
			Diagram: diagramResp.Diagram,
		})
	}
}

func (p *CanvasViewPresenter) refreshGenerators(ctx context.Context) {
	if p.DashboardPage == nil || p.CanvasService == nil {
		return
	}

	gensResp, err := p.CanvasService.ListGenerators(ctx, &protos.ListGeneratorsRequest{
		CanvasId: p.canvasID,
	})
	if err == nil {
		p.DashboardPage.SetGeneratorList(ctx, &protos.SetGeneratorListRequest{
			Generators: gensResp.Generators,
		})
	}
}

func (p *CanvasViewPresenter) refreshMetrics(ctx context.Context) {
	if p.DashboardPage == nil || p.CanvasService == nil {
		return
	}

	metricsResp, err := p.CanvasService.ListMetrics(ctx, &protos.ListMetricsRequest{
		CanvasId: p.canvasID,
	})
	if err == nil {
		p.DashboardPage.SetMetricsList(ctx, &protos.SetMetricsListRequest{
			Metrics: metricsResp.Metrics,
		})
	}
}

func (p *CanvasViewPresenter) evaluateAndUpdateFlows(ctx context.Context) {
	if p.DashboardPage == nil || p.CanvasService == nil {
		return
	}

	resp, err := p.CanvasService.EvaluateFlows(ctx, &protos.EvaluateFlowsRequest{
		CanvasId: p.canvasID,
		Strategy: "iterative",
	})
	if err == nil {
		p.DashboardPage.UpdateFlowRates(ctx, &protos.UpdateFlowRatesRequest{
			Rates:    resp.ComponentRates,
			Strategy: resp.Strategy,
		})
	}
}
