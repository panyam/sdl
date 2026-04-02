package services

import (
	"context"
	"fmt"
	"strconv"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// WorkspacePresenter is the P in Model-View-Presenter for workspaces.
// It holds a DevEnv runtime instance and implements WorkspacePresenterServer
// (proto-generated). The presenter bridges user interactions (from the page)
// to the runtime (DevEnv), similar to lilbattle's GameViewPresenter.
//
// Not WASM-tagged — the presenter logic is the same for browser and CLI.
// WASM-specific page implementations (BrowserWorkspacePage) live in cmd/wasm/.
type WorkspacePresenter struct {
	DevEnv      *DevEnv
	initialized bool
}

// NewWorkspacePresenter creates a presenter wired to a DevEnv instance.
func NewWorkspacePresenter(devEnv *DevEnv) *WorkspacePresenter {
	return &WorkspacePresenter{DevEnv: devEnv}
}

// ============================================================================
// CanvasViewPresenterServer implementation — handle commands from browser
// ============================================================================

func (p *WorkspacePresenter) Initialize(ctx context.Context, req *protos.InitializePresenterRequest) (*protos.InitializePresenterResponse, error) {
	p.initialized = true
	return &protos.InitializePresenterResponse{
		Success:          true,
		WorkspaceId:         "default",
		AvailableSystems: systemNamesToInfos(p.DevEnv.AvailableSystems()),
	}, nil
}

func (p *WorkspacePresenter) ClientReady(ctx context.Context, req *protos.ClientReadyRequest) (*protos.ClientReadyResponse, error) {
	// Auto-select first system if none active
	systems := p.DevEnv.AvailableSystems()
	if p.DevEnv.ActiveSystem() == nil && len(systems) > 0 {
		if err := p.DevEnv.Use(systems[0]); err != nil {
			return nil, fmt.Errorf("failed to auto-select system: %w", err)
		}
	}

	// Build Canvas proto for the response (browser expects this format)
	canvas := &protos.Canvas{
		Id:                "default",
		LoadedSystemNames: systems,
	}
	if p.DevEnv.ActiveSystem() != nil && p.DevEnv.ActiveSystem().System != nil {
		canvas.ActiveSystem = p.DevEnv.ActiveSystem().System.Name.Value
	}

	return &protos.ClientReadyResponse{
		Success: true,
		Canvas:  canvas,
	}, nil
}

func (p *WorkspacePresenter) FileSelected(ctx context.Context, req *protos.FileSelectedRequest) (*protos.FileSelectedResponse, error) {
	if err := p.DevEnv.LoadFile(req.FilePath); err != nil {
		return nil, fmt.Errorf("failed to load file: %w", err)
	}

	// Auto-select first system if none active
	systems := p.DevEnv.AvailableSystems()
	if p.DevEnv.ActiveSystem() == nil && len(systems) > 0 {
		p.DevEnv.Use(systems[0])
	}

	return &protos.FileSelectedResponse{}, nil
}

func (p *WorkspacePresenter) FileSaved(ctx context.Context, req *protos.FileSavedRequest) (*protos.FileSavedResponse, error) {
	if err := p.DevEnv.LoadFile(req.FilePath); err != nil {
		return nil, fmt.Errorf("failed to reload file: %w", err)
	}
	return &protos.FileSavedResponse{}, nil
}

func (p *WorkspacePresenter) UseSystem(ctx context.Context, req *protos.UseSystemRequest) (*protos.UseSystemResponse, error) {
	if err := p.DevEnv.Use(req.SystemName); err != nil {
		return nil, err
	}
	return &protos.UseSystemResponse{}, nil
}

func (p *WorkspacePresenter) AddGenerator(ctx context.Context, req *protos.AddGeneratorRequest) (*protos.AddGeneratorResponse, error) {
	gen := req.Generator
	if gen == nil {
		return nil, fmt.Errorf("generator is required")
	}

	genInfo := &GeneratorInfo{Generator: gen}
	if err := p.DevEnv.AddGenerator(genInfo); err != nil {
		return nil, err
	}

	if req.ApplyFlows {
		p.DevEnv.EvaluateFlows("runtime")
	}

	return &protos.AddGeneratorResponse{Generator: gen}, nil
}

func (p *WorkspacePresenter) UpdateGenerator(ctx context.Context, req *protos.UpdateGeneratorRequest) (*protos.UpdateGeneratorResponse, error) {
	gen := req.Generator
	if gen == nil {
		return nil, fmt.Errorf("generator is required")
	}

	if err := p.DevEnv.UpdateGenerator(gen.Name, gen.Rate); err != nil {
		return nil, err
	}

	if req.ApplyFlows {
		p.DevEnv.EvaluateFlows("runtime")
	}

	return &protos.UpdateGeneratorResponse{Generator: gen}, nil
}

func (p *WorkspacePresenter) DeleteGenerator(ctx context.Context, req *protos.DeleteGeneratorRequest) (*protos.DeleteGeneratorResponse, error) {
	if err := p.DevEnv.RemoveGenerator(req.GeneratorId); err != nil {
		return nil, err
	}

	if req.ApplyFlows {
		p.DevEnv.EvaluateFlows("runtime")
	}

	return &protos.DeleteGeneratorResponse{}, nil
}

func (p *WorkspacePresenter) StartGenerator(ctx context.Context, req *protos.StartGeneratorRequest) (*protos.StartGeneratorResponse, error) {
	if err := p.DevEnv.StartGenerator(req.GeneratorId); err != nil {
		return nil, err
	}
	return &protos.StartGeneratorResponse{}, nil
}

func (p *WorkspacePresenter) StopGenerator(ctx context.Context, req *protos.StopGeneratorRequest) (*protos.StopGeneratorResponse, error) {
	if err := p.DevEnv.StopGenerator(req.GeneratorId); err != nil {
		return nil, err
	}
	return &protos.StopGeneratorResponse{}, nil
}

func (p *WorkspacePresenter) StartAllGenerators(ctx context.Context, req *protos.StartAllGeneratorsRequest) (*protos.StartAllGeneratorsResponse, error) {
	if err := p.DevEnv.StartAllGenerators(); err != nil {
		return nil, err
	}
	return &protos.StartAllGeneratorsResponse{}, nil
}

func (p *WorkspacePresenter) StopAllGenerators(ctx context.Context, req *protos.StopAllGeneratorsRequest) (*protos.StopAllGeneratorsResponse, error) {
	if err := p.DevEnv.StopAllGenerators(); err != nil {
		return nil, err
	}
	return &protos.StopAllGeneratorsResponse{}, nil
}

func (p *WorkspacePresenter) AddMetric(ctx context.Context, req *protos.AddMetricRequest) (*protos.AddMetricResponse, error) {
	m := req.Metric
	if m == nil {
		return nil, fmt.Errorf("metric is required")
	}

	spec := &MetricSpec{Metric: m}
	if err := p.DevEnv.AddMetric(spec); err != nil {
		return nil, err
	}

	return &protos.AddMetricResponse{Metric: m}, nil
}

func (p *WorkspacePresenter) DeleteMetric(ctx context.Context, req *protos.DeleteMetricRequest) (*protos.DeleteMetricResponse, error) {
	if err := p.DevEnv.RemoveMetric(req.MetricId); err != nil {
		return nil, err
	}
	return &protos.DeleteMetricResponse{}, nil
}

func (p *WorkspacePresenter) SetParameter(ctx context.Context, req *protos.SetParameterRequest) (*protos.SetParameterResponse, error) {
	value := parseParameterValue(req.NewValue)
	if err := p.DevEnv.SetParameter(req.Path, value); err != nil {
		return nil, err
	}

	// Re-evaluate flows after parameter change
	p.DevEnv.EvaluateFlows("runtime")

	return &protos.SetParameterResponse{}, nil
}

func (p *WorkspacePresenter) EvaluateFlows(ctx context.Context, req *protos.EvaluateFlowsRequest) (*protos.EvaluateFlowsResponse, error) {
	strategy := req.Strategy
	if strategy == "" {
		strategy = "runtime"
	}

	result, err := p.DevEnv.EvaluateFlows(strategy)
	if err != nil {
		return nil, err
	}

	return &protos.EvaluateFlowsResponse{
		Strategy:       strategy,
		Status:         "applied",
		ComponentRates: result.Flows.ComponentRates,
	}, nil
}

func (p *WorkspacePresenter) DiagramComponentClicked(ctx context.Context, req *protos.DiagramComponentClickedRequest) (*protos.DiagramComponentClickedResponse, error) {
	// Highlighting is a browser-local concern; no WASM round-trip needed.
	return &protos.DiagramComponentClickedResponse{Success: true}, nil
}

func (p *WorkspacePresenter) DiagramComponentHovered(ctx context.Context, req *protos.DiagramComponentHoveredRequest) (*protos.DiagramComponentHoveredResponse, error) {
	return &protos.DiagramComponentHoveredResponse{Success: true}, nil
}

// ============================================================================
// Helpers
// ============================================================================

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

// systemNamesToInfos converts a list of system names to SystemInfo protos.
func systemNamesToInfos(names []string) []*protos.SystemInfo {
	infos := make([]*protos.SystemInfo, len(names))
	for i, name := range names {
		infos[i] = &protos.SystemInfo{Id: name, Name: name}
	}
	return infos
}
