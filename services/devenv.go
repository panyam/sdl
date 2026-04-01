package services

import (
	"fmt"
	"log"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/lib/decl"
	"github.com/panyam/sdl/lib/loader"
	"github.com/panyam/sdl/lib/runtime"
)

// DevEnv is the primary simulation coordinator, replacing Canvas + CanvasViewPresenter.
// It is constructed with a FileResolver, owns a Runtime internally, and pushes typed
// updates to an attached WorkspacePage.
type DevEnv struct {
	runtime       *runtime.Runtime
	activeSystem  *runtime.SystemInstance
	loadedSystems map[string]*runtime.SystemInstance

	// Generator management
	generators     map[string]*GeneratorInfo
	generatorsLock sync.RWMutex

	// Metrics
	metricTracer *MetricTracer

	// Flow analysis
	currentFlowScope    *runtime.FlowScope
	currentFlowRates    runtime.RateMap
	currentFlowStrategy string
	manualRateOverrides map[string]float64

	// Simulation time
	simulationStartTime time.Time
	simulationStarted   bool

	// Page handler (single panel endpoint, like CanvasDashboardPage)
	page     WorkspacePage
	pageLock sync.RWMutex
}

// NewDevEnv creates a new DevEnv with the given file resolver.
func NewDevEnv(resolver loader.FileResolver) *DevEnv {
	sdlLoader := loader.NewLoader(nil, resolver, 10)
	rt := runtime.NewRuntime(sdlLoader)
	return &DevEnv{
		runtime:             rt,
		loadedSystems:       make(map[string]*runtime.SystemInstance),
		generators:          make(map[string]*GeneratorInfo),
		manualRateOverrides: make(map[string]float64),
	}
}

// SimulationContext implementation

func (d *DevEnv) GetTracer() runtime.Tracer          { return d.metricTracer }
func (d *DevEnv) GetSimulationStartTime() time.Time   { return d.simulationStartTime }
func (d *DevEnv) IsSimulationStarted() bool            { return d.simulationStarted }
func (d *DevEnv) GetSimulationTime() float64           { return 0 } // TODO: virtual time tracking

// Page handler management

// SetPage attaches a WorkspacePage that will receive all panel updates.
func (d *DevEnv) SetPage(page WorkspacePage) {
	d.pageLock.Lock()
	defer d.pageLock.Unlock()
	d.page = page
}

// ClearPage removes the current page handler.
func (d *DevEnv) ClearPage() {
	d.pageLock.Lock()
	defer d.pageLock.Unlock()
	d.page = nil
}

func (d *DevEnv) getPage() WorkspacePage {
	d.pageLock.RLock()
	defer d.pageLock.RUnlock()
	return d.page
}

// Core API

// LoadFile parses an SDL file and makes its systems available.
func (d *DevEnv) LoadFile(filePath string) error {
	_, err := d.runtime.LoadFile(filePath)
	if err != nil {
		return err
	}
	if page := d.getPage(); page != nil {
		page.OnAvailableSystemsChanged(d.AvailableSystems())
	}
	return nil
}

// AvailableSystems returns the names of all systems discovered across loaded files.
func (d *DevEnv) AvailableSystems() []string {
	systems := d.runtime.AvailableSystems()
	return slices.Collect(maps.Keys(systems))
}

// ActiveSystem returns the currently active system instance, or nil.
func (d *DevEnv) ActiveSystem() *runtime.SystemInstance {
	return d.activeSystem
}

// GetActiveSystemName returns the name of the active system, or empty string.
func (d *DevEnv) GetActiveSystemName() string {
	if d.activeSystem == nil || d.activeSystem.System == nil {
		return ""
	}
	return d.activeSystem.System.Name.Value
}

// ListGenerators returns proto Generator copies for all registered generators.
func (d *DevEnv) ListGenerators() []*protos.Generator {
	d.generatorsLock.RLock()
	defer d.generatorsLock.RUnlock()
	result := make([]*protos.Generator, 0, len(d.generators))
	for _, gen := range d.generators {
		result = append(result, gen.Generator)
	}
	return result
}

// ListMetrics returns proto Metric copies for all tracked metrics.
func (d *DevEnv) ListMetrics() []*protos.Metric {
	if d.metricTracer == nil {
		return nil
	}
	return d.metricTracer.ListMetrics()
}

// Use activates a system by name. Creates the SystemInstance if needed,
// wires up declared generators and metrics, and notifies the page handler.
func (d *DevEnv) Use(systemName string) error {
	if d.loadedSystems[systemName] == nil {
		system, err := d.runtime.NewSystem(systemName)
		if err != nil {
			return err
		}
		d.loadedSystems[systemName] = system
	}

	// Stop existing generators before switching
	d.stopAllGeneratorsInternal()

	d.activeSystem = d.loadedSystems[systemName]

	// Reset metric tracer
	if d.metricTracer != nil {
		d.metricTracer.Clear()
	}
	d.metricTracer = NewMetricTracer(d.activeSystem, d)

	// Reset simulation time
	d.simulationStarted = false

	// Initialize flow contexts
	d.initializeFlowContexts()

	// Auto-create generators and metrics declared in the system block
	if err := d.createDeclaredGenerators(); err != nil {
		return err
	}
	if err := d.createDeclaredMetrics(); err != nil {
		return err
	}

	// Notify page handler
	if page := d.getPage(); page != nil {
		page.OnSystemChanged(systemName, d.AvailableSystems())

		// Push diagram
		if diagram, err := d.GetSystemDiagram(); err == nil {
			page.UpdateDiagram(diagram)
		}

		// Push generators
		d.generatorsLock.RLock()
		for name, gen := range d.generators {
			page.UpdateGenerator(name, gen.Generator)
		}
		d.generatorsLock.RUnlock()

		// Push metrics
		if d.metricTracer != nil {
			for _, m := range d.metricTracer.ListMetrics() {
				page.UpdateMetric(m.Name, m)
			}
		}
	}

	return nil
}

// Generator management

// AddGenerator adds and starts a new generator.
func (d *DevEnv) AddGenerator(gen *GeneratorInfo) error {
	if gen.Id == "" {
		return fmt.Errorf("generator ID cannot be empty")
	}
	if d.activeSystem == nil {
		return fmt.Errorf("no active system")
	}

	gen.simCtx = d
	gen.System = d.activeSystem

	// Resolve component and method if not already resolved
	if gen.resolvedComponentInstance == nil && gen.Component != "" {
		gen.resolvedComponentInstance = d.activeSystem.FindComponent(gen.Component)
	}

	d.generatorsLock.Lock()
	if d.generators[gen.Id] != nil {
		d.generatorsLock.Unlock()
		return fmt.Errorf("generator '%s' already exists", gen.Id)
	}
	d.generators[gen.Id] = gen
	d.generatorsLock.Unlock()

	gen.Start()

	if page := d.getPage(); page != nil {
		page.UpdateGenerator(gen.Name, gen.Generator)
	}

	return nil
}

// UpdateGenerator updates an existing generator's rate.
func (d *DevEnv) UpdateGenerator(name string, rate float64) error {
	d.generatorsLock.Lock()
	gen, ok := d.generators[name]
	if !ok {
		d.generatorsLock.Unlock()
		return fmt.Errorf("generator '%s' not found", name)
	}
	gen.Rate = rate
	d.generatorsLock.Unlock()

	if page := d.getPage(); page != nil {
		page.UpdateGenerator(name, gen.Generator)
	}
	return nil
}

// RemoveGenerator stops and removes a generator by name.
func (d *DevEnv) RemoveGenerator(name string) error {
	d.generatorsLock.Lock()
	gen, ok := d.generators[name]
	if !ok {
		d.generatorsLock.Unlock()
		return fmt.Errorf("generator '%s' not found", name)
	}
	delete(d.generators, name)
	d.generatorsLock.Unlock()

	gen.Stop(false)

	if page := d.getPage(); page != nil {
		page.RemoveGenerator(name)
	}
	return nil
}

// StartGenerator starts a generator by name.
func (d *DevEnv) StartGenerator(name string) error {
	d.generatorsLock.RLock()
	gen, ok := d.generators[name]
	d.generatorsLock.RUnlock()
	if !ok {
		return fmt.Errorf("generator '%s' not found", name)
	}

	if err := gen.Start(); err != nil {
		return err
	}

	// Mark simulation as started on first generator start
	if !d.simulationStarted {
		d.simulationStarted = true
		d.simulationStartTime = time.Now()
	}

	if page := d.getPage(); page != nil {
		page.UpdateGenerator(name, gen.Generator)
	}
	return nil
}

// StopGenerator stops a generator by name.
func (d *DevEnv) StopGenerator(name string) error {
	d.generatorsLock.RLock()
	gen, ok := d.generators[name]
	d.generatorsLock.RUnlock()
	if !ok {
		return fmt.Errorf("generator '%s' not found", name)
	}

	if err := gen.Stop(false); err != nil {
		return err
	}

	if page := d.getPage(); page != nil {
		page.UpdateGenerator(name, gen.Generator)
	}
	return nil
}

// StartAllGenerators starts all registered generators.
func (d *DevEnv) StartAllGenerators() error {
	d.generatorsLock.RLock()
	gens := make([]*GeneratorInfo, 0, len(d.generators))
	for _, gen := range d.generators {
		gens = append(gens, gen)
	}
	d.generatorsLock.RUnlock()

	if !d.simulationStarted && len(gens) > 0 {
		d.simulationStarted = true
		d.simulationStartTime = time.Now()
	}

	for _, gen := range gens {
		gen.Start()
	}
	return nil
}

// StopAllGenerators stops all registered generators.
func (d *DevEnv) StopAllGenerators() error {
	d.stopAllGeneratorsInternal()
	return nil
}

func (d *DevEnv) stopAllGeneratorsInternal() {
	d.generatorsLock.RLock()
	gens := make([]*GeneratorInfo, 0, len(d.generators))
	for _, gen := range d.generators {
		gens = append(gens, gen)
	}
	d.generatorsLock.RUnlock()

	for _, gen := range gens {
		gen.Stop(false)
	}
}

// Metric management

// AddMetric adds a new metric to the tracer and notifies the page.
func (d *DevEnv) AddMetric(spec *MetricSpec) error {
	if d.metricTracer == nil {
		return fmt.Errorf("no active system")
	}
	if err := d.metricTracer.AddMetricSpec(spec); err != nil {
		return err
	}
	if page := d.getPage(); page != nil {
		page.UpdateMetric(spec.Name, spec.Metric)
	}
	return nil
}

// RemoveMetric removes a metric by ID and notifies the page.
func (d *DevEnv) RemoveMetric(id string) error {
	if d.metricTracer == nil {
		return fmt.Errorf("no active system")
	}
	metric := d.metricTracer.GetMetricByID(id)
	d.metricTracer.RemoveMetricSpec(id)
	if page := d.getPage(); page != nil && metric != nil {
		page.RemoveMetric(metric.Name)
	}
	return nil
}

// Parameter management

// SetParameter modifies a component parameter at runtime.
func (d *DevEnv) SetParameter(path string, value any) error {
	if d.activeSystem == nil || d.activeSystem.Env == nil {
		return fmt.Errorf("no active system")
	}

	parts := strings.Split(path, ".")
	componentPath, paramName := strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]
	componentInstance := d.activeSystem.FindComponent(componentPath)
	if componentInstance == nil {
		return fmt.Errorf("component '%s' not found", componentPath)
	}

	var newValue decl.Value
	var err error
	switch v := value.(type) {
	case int:
		newValue, err = decl.NewValue(decl.IntType, int64(v))
	case int64:
		newValue, err = decl.NewValue(decl.IntType, v)
	case float64:
		newValue, err = decl.NewValue(decl.FloatType, v)
	case bool:
		newValue, err = decl.NewValue(decl.BoolType, v)
	case string:
		newValue, err = decl.NewValue(decl.StrType, v)
	default:
		err = fmt.Errorf("unsupported value type: %T", value)
	}
	if err != nil {
		return err
	}

	return componentInstance.Set(paramName, newValue)
}

// Diagram

// GetSystemDiagram builds and returns the current system topology.
func (d *DevEnv) GetSystemDiagram() (*SystemDiagram, error) {
	if d.activeSystem == nil {
		return nil, fmt.Errorf("no active system")
	}
	return BuildSystemDiagram(d.activeSystem, d.generators, d.currentFlowScope, d.getCurrentFlowRates())
}

// Flow analysis

// EvaluateFlows evaluates and applies flow rates using the given strategy,
// then notifies the page handler.
func (d *DevEnv) EvaluateFlows(strategy string) (*runtime.FlowAnalysisResult, error) {
	if d.activeSystem == nil {
		return nil, fmt.Errorf("no active system")
	}

	// Build generator configs
	var generators []runtime.GeneratorConfigAPI
	d.generatorsLock.RLock()
	for _, gen := range d.generators {
		if gen.Enabled {
			generators = append(generators, runtime.GeneratorConfigAPI{
				ID:        gen.Id,
				Component: gen.Component,
				Method:    gen.Method,
				Rate:      float64(gen.Rate),
			})
		}
	}
	d.generatorsLock.RUnlock()

	result, err := runtime.EvaluateFlowStrategy(strategy, d.activeSystem, generators)
	if err != nil {
		return nil, err
	}

	// Apply results
	d.currentFlowScope = runtime.NewFlowScope(d.activeSystem.Env)
	d.currentFlowRates = d.convertFlowResultToRateMap(result)
	d.currentFlowScope.ArrivalRates = d.currentFlowRates
	d.currentFlowStrategy = strategy

	// Populate FlowEdges
	if d.currentFlowScope.FlowEdges != nil {
		d.currentFlowScope.FlowEdges.Clear()
		for _, edge := range result.Flows.Edges {
			fromComp := d.activeSystem.FindComponent(edge.From.Component)
			toComp := d.activeSystem.FindComponent(edge.To.Component)
			if fromComp != nil && toComp != nil {
				d.currentFlowScope.FlowEdges.AddEdge(fromComp, edge.From.Method, toComp, edge.To.Method, edge.Rate)
			}
		}
	}

	// Apply arrival rates to components
	if err := d.currentFlowScope.ApplyToComponents(); err != nil {
		slog.Warn("Failed to apply some arrival rates", "error", err)
	}

	// Notify page
	if page := d.getPage(); page != nil {
		page.UpdateFlowRates(d.getCurrentFlowRates(), strategy)
	}

	return result, nil
}

// Close stops all generators, clears metrics, and releases resources.
func (d *DevEnv) Close() error {
	d.stopAllGeneratorsInternal()
	if d.metricTracer != nil {
		d.metricTracer.Clear()
		d.metricTracer = nil
	}
	return nil
}

// Internal helpers

func (d *DevEnv) createDeclaredGenerators() error {
	if d.activeSystem == nil {
		return nil
	}

	// Clear existing generators
	d.generatorsLock.Lock()
	d.generators = make(map[string]*GeneratorInfo)
	d.generatorsLock.Unlock()

	for _, gen := range d.activeSystem.Generators {
		genInfo := &GeneratorInfo{
			Generator: &protos.Generator{
				Id:        gen.Name,
				Name:      gen.Name,
				Component: gen.ComponentPath,
				Method:    gen.MethodName,
				Rate:      gen.RPS(),
				Duration:  gen.Duration,
				Enabled:   true,
			},
		}
		genInfo.resolvedComponentInstance = gen.ResolvedComponent
		genInfo.resolvedMethodDecl = gen.ResolvedMethod
		genInfo.System = d.activeSystem
		genInfo.simCtx = d

		d.generatorsLock.Lock()
		d.generators[gen.Name] = genInfo
		d.generatorsLock.Unlock()
	}

	// Recompute flows if generators exist
	if len(d.activeSystem.Generators) > 0 {
		return d.recomputeSystemFlows()
	}
	return nil
}

func (d *DevEnv) createDeclaredMetrics() error {
	if d.activeSystem == nil || d.metricTracer == nil {
		return nil
	}
	for _, m := range d.activeSystem.Metrics {
		methods := []string{}
		if m.MethodName != "" {
			methods = []string{m.MethodName}
		}
		metricSpec := &MetricSpec{
			Metric: &protos.Metric{
				Id:                m.Name,
				Name:              m.Name,
				Component:         m.ComponentPath,
				Methods:           methods,
				MetricType:        m.MetricType,
				Aggregation:       m.Aggregation,
				AggregationWindow: m.Window,
				Enabled:           true,
			},
		}
		if err := d.metricTracer.AddMetricSpec(metricSpec); err != nil {
			log.Printf("Warning: failed to create declared metric '%s': %v", m.Name, err)
		}
	}
	return nil
}

func (d *DevEnv) initializeFlowContexts() {
	if d.activeSystem == nil {
		return
	}
	d.currentFlowScope = runtime.NewFlowScope(d.activeSystem.Env)
	d.currentFlowRates = runtime.NewRateMap()
}

func (d *DevEnv) recomputeSystemFlows() error {
	if d.activeSystem == nil {
		return nil
	}

	var generators []runtime.GeneratorConfigAPI
	d.generatorsLock.RLock()
	for _, gen := range d.generators {
		if gen.Enabled {
			generators = append(generators, runtime.GeneratorConfigAPI{
				ID:        gen.Id,
				Component: gen.Component,
				Method:    gen.Method,
				Rate:      float64(gen.Rate),
			})
		}
	}
	d.generatorsLock.RUnlock()

	strategy := d.currentFlowStrategy
	if strategy == "" {
		strategy = "runtime"
	}

	result, err := runtime.EvaluateFlowStrategy(strategy, d.activeSystem, generators)
	if err != nil {
		return err
	}

	d.currentFlowScope = runtime.NewFlowScope(d.activeSystem.Env)
	d.currentFlowRates = d.convertFlowResultToRateMap(result)
	d.currentFlowScope.ArrivalRates = d.currentFlowRates

	if d.currentFlowScope.FlowEdges != nil {
		d.currentFlowScope.FlowEdges.Clear()
		for _, edge := range result.Flows.Edges {
			fromComp := d.activeSystem.FindComponent(edge.From.Component)
			toComp := d.activeSystem.FindComponent(edge.To.Component)
			if fromComp != nil && toComp != nil {
				d.currentFlowScope.FlowEdges.AddEdge(fromComp, edge.From.Method, toComp, edge.To.Method, edge.Rate)
			}
		}
	}

	if err := d.currentFlowScope.ApplyToComponents(); err != nil {
		slog.Warn("Failed to apply some arrival rates", "error", err)
	}

	return nil
}

func (d *DevEnv) getCurrentFlowRates() map[string]float64 {
	if d.currentFlowRates == nil {
		return make(map[string]float64)
	}
	return d.rateMapToStringMap(d.currentFlowRates)
}

func (d *DevEnv) rateMapToStringMap(rateMap runtime.RateMap) map[string]float64 {
	result := make(map[string]float64)

	instanceToName := make(map[*runtime.ComponentInstance]string)
	if d.activeSystem != nil && d.activeSystem.System != nil {
		for _, param := range d.activeSystem.System.Parameters {
			instanceName := param.Name.Value
			if comp := d.activeSystem.FindComponent(instanceName); comp != nil {
				instanceToName[comp] = instanceName
			}
		}
	}

	for component, methods := range rateMap {
		if component != nil {
			componentName := component.ID()
			if name, found := instanceToName[component]; found {
				componentName = name
			}
			for method, rate := range methods {
				key := fmt.Sprintf("%s.%s", componentName, method)
				result[key] = rate
			}
		}
	}

	return result
}

func (d *DevEnv) convertFlowResultToRateMap(result *runtime.FlowAnalysisResult) runtime.RateMap {
	rateMap := runtime.NewRateMap()

	for compMethod, rate := range result.Flows.ComponentRates {
		parts := strings.Split(compMethod, ".")
		if len(parts) >= 2 {
			componentName := parts[0]
			methodName := strings.Join(parts[1:], ".")
			compInst := d.activeSystem.FindComponent(componentName)
			if compInst != nil {
				rateMap.SetRate(compInst, methodName, rate)
			}
		}
	}

	for compMethod, rate := range d.manualRateOverrides {
		parts := strings.Split(compMethod, ".")
		if len(parts) >= 2 {
			componentName := parts[0]
			methodName := strings.Join(parts[1:], ".")
			compInst := d.activeSystem.FindComponent(componentName)
			if compInst != nil {
				rateMap.SetRate(compInst, methodName, rate)
			}
		}
	}

	return rateMap
}
