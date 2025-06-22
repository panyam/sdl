package console

import (
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
	protos "github.com/panyam/sdl/gen/go/sdl/v1"
	"github.com/panyam/sdl/loader"
	"github.com/panyam/sdl/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Canvas struct {
	id             string
	runtime        *runtime.Runtime
	activeSystem   *runtime.SystemInstance
	loadedSystems  map[string]*runtime.SystemInstance
	generators     map[string]*GeneratorInfo
	generatorsLock sync.RWMutex

	metricTracer *MetricTracer

	currentFlowScope    *runtime.FlowScope // Current flow state (applied/active)
	proposedFlowScope   *runtime.FlowScope // Proposed flow state (being evaluated)
	currentFlowRates    runtime.RateMap    // Current flow rates (runtime-based)
	proposedFlowRates   runtime.RateMap    // Proposed flow rates (runtime-based)
	currentFlowStrategy string             // Strategy used for current flow rates
	manualRateOverrides map[string]float64 // Manual arrival rate overrides

	// Simulation time tracking
	simulationStartTime time.Time
	simulationStarted   bool
}

// NewCanvas creates a new interactive canvas session.
func NewCanvas(id string) *Canvas {
	l := loader.NewLoader(nil, nil, 10)
	r := runtime.NewRuntime(l)
	return &Canvas{
		id:                  id,
		runtime:             r,
		loadedSystems:       map[string]*runtime.SystemInstance{},
		generators:          map[string]*GeneratorInfo{},
		manualRateOverrides: make(map[string]float64),
		currentFlowStrategy: runtime.GetDefaultFlowStrategy(),
	}
}

// Load parses, validates, and loads an SDL file and all its imports
// into the session. If the file is already loaded, it will be re-loaded
// and re-validated to ensure freshness.
func (c *Canvas) Load(filePath string) error {
	// TODO: Handle hot-reloading. For now, we load once.
	_, err := c.runtime.LoadFile(filePath)
	if err != nil {
		return fmt.Errorf("error loading file '%s': %w", filePath, err)
	}

	// Invalidate flow contexts since file changed
	// c.invalidateFlowContexts()

	return nil
}

// CurrentSystem returns the currently active system
func (c *Canvas) CurrentSystem() *runtime.SystemInstance {
	return c.activeSystem
}

// Use uses a system in a given file as the one being tested
func (c *Canvas) Use(systemName string) error {
	if c.loadedSystems[systemName] == nil {
		system, err := c.runtime.NewSystem(systemName)
		if err != nil {
			return err
		}
		c.loadedSystems[systemName] = system
	}
	log.Println("DId we come here??")
	c.activeSystem = c.loadedSystems[systemName]
	// Always create a new metric tracer when switching systems
	if c.metricTracer != nil {
		c.metricTracer.Clear()
	}
	c.metricTracer = NewMetricTracer(c.activeSystem, c)

	// Reset simulation time tracking
	c.simulationStarted = false

	// Initialize flow contexts for the new system
	c.initializeFlowContexts()

	return nil
}

// Set modifies a component parameter at runtime.
// The path is a dot-separated string, e.g., "app.cache.HitRate".
func (c *Canvas) Set(path string, value any) error {
	if c.activeSystem == nil || c.activeSystem.Env == nil {
		return fmt.Errorf("no active system. Call Use() before Set()")
	}
	parts := strings.Split(path, ".")
	componentPath, paramName := strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]
	componentInstance := c.activeSystem.FindComponent(componentPath)

	// For user-defined components, set the parameter in their runtime environment.
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
		err = fmt.Errorf("unsupported value type for Set on user-defined component: %T", value)
	}
	if err != nil {
		err = componentInstance.Set(paramName, newValue)
	}
	return err
}

// GetAvailableSystemNames returns all system names from loaded SDL files
func (c *Canvas) GetAvailableSystemNames() []string {
	systems := c.runtime.AvailableSystems()
	return slices.Collect(maps.Keys(systems))
}

// Starts/stops all generators
// BatchGeneratorResult holds the results of batch generator operations
type BatchGeneratorResult struct {
	TotalGenerators     int
	ProcessedCount      int // started or stopped
	AlreadyInStateCount int // already running/stopped
	FailedCount         int
	FailedIDs           []string
}

// StartAllGenerators starts all generators and returns detailed results
func (c *Canvas) StartAllGenerators() (*BatchGeneratorResult, error) {
	c.generatorsLock.Lock()
	defer c.generatorsLock.Unlock()

	result := &BatchGeneratorResult{
		TotalGenerators: len(c.generators),
		FailedIDs:       []string{},
	}

	for id, gen := range c.generators {
		if gen.IsRunning() {
			result.AlreadyInStateCount++
			continue
		}

		err := gen.Start()
		if err != nil {
			result.FailedCount++
			result.FailedIDs = append(result.FailedIDs, id)
		} else {
			result.ProcessedCount++
		}
	}

	// Recompute flows after starting generators
	if err := c.recomputeSystemFlows(); err != nil {
		return result, fmt.Errorf("failed to recompute flows: %w", err)
	}

	return result, nil
}

// StopAllGenerators stops all generators and returns detailed results
func (c *Canvas) StopAllGenerators() (*BatchGeneratorResult, error) {
	c.generatorsLock.Lock()
	defer c.generatorsLock.Unlock()

	result := &BatchGeneratorResult{
		TotalGenerators: len(c.generators),
		FailedIDs:       []string{},
	}

	for id, gen := range c.generators {
		if !gen.IsRunning() {
			result.AlreadyInStateCount++
			continue
		}

		err := gen.Stop()
		if err != nil {
			result.FailedCount++
			result.FailedIDs = append(result.FailedIDs, id)
		} else {
			result.ProcessedCount++
		}
	}

	// Recompute flows after stopping generators
	if err := c.recomputeSystemFlows(); err != nil {
		return result, fmt.Errorf("failed to recompute flows: %w", err)
	}

	return result, nil
}

// ToggleAllGenerators toggles all generators (deprecated, use Start/StopAllGenerators)
func (c *Canvas) ToggleAllGenerators(start bool) error {
	if start {
		_, err := c.StartAllGenerators()
		return err
	} else {
		_, err := c.StopAllGenerators()
		return err
	}
}

// AddGenerator adds a new traffic generator configuration
func (c *Canvas) AddGenerator(gen *GeneratorInfo) error {
	gen.System = c.activeSystem
	if c.activeSystem == nil || c.activeSystem.Env == nil {
		return status.Error(codes.InvalidArgument, "System has not been loaded or initialized")
	}
	c.generatorsLock.Lock()
	defer c.generatorsLock.Unlock()

	// Only process exit events with method calls
	if gen.Component == "" || gen.Method == "" {
		return status.Error(codes.InvalidArgument, "Component and Method cannot be empty")
	}

	// Resolve the component instance from the system
	system := c.activeSystem
	gen.resolvedComponentInstance = system.FindComponent(gen.Component)
	if gen.resolvedComponentInstance == nil {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("component '%s' not found in system when adding generator", gen.Component))
	}

	// Check method match
	methodDecl, err := gen.resolvedComponentInstance.ComponentDecl.GetMethod(gen.Method)
	if err != nil || methodDecl == nil {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("'%s' is not a method in the component", gen.Method))
	}
	gen.resolvedMethodDecl = methodDecl
	gen.canvas = c

	gen.System = c.activeSystem
	if c.generators[gen.Id] != nil {
		return status.Error(codes.AlreadyExists, "Generator with name already exists")
	}
	c.generators[gen.Id] = gen
	gen.Start()
	return c.recomputeSystemFlows()
}

func (c *Canvas) RemoveGenerator(genId string) error {
	c.generatorsLock.Lock()
	defer c.generatorsLock.Unlock()

	gen, exists := c.generators[genId]
	if !exists {
		return status.Error(codes.NotFound, "Generator with name not found")
	}
	delete(c.generators, genId)
	gen.Stop()
	return c.recomputeSystemFlows()
}

func (c *Canvas) StopGenerator(genId string) error {
	c.generatorsLock.Lock()
	defer c.generatorsLock.Unlock()

	if c.generators[genId] == nil {
		return status.Error(codes.NotFound, "Generator not found")
	}
	c.generators[genId].Stop()
	return c.recomputeSystemFlows()
}

func (c *Canvas) StartGenerator(genId string) error {
	c.generatorsLock.Lock()
	defer c.generatorsLock.Unlock()

	if c.generators[genId] == nil {
		return status.Error(codes.NotFound, "Generator not found")
	}

	// Track simulation start time when first generator starts
	if !c.simulationStarted {
		c.simulationStartTime = time.Now()
		c.simulationStarted = true
	}

	c.generators[genId].Start()
	return c.recomputeSystemFlows()
}

// ListGenerators returns all generator configurations
func (c *Canvas) ListGenerators() map[string]*GeneratorInfo {
	c.generatorsLock.RLock()
	defer c.generatorsLock.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*GeneratorInfo)
	for k, v := range c.generators {
		result[k] = v
	}
	return result
}

// --- Option types for Run/Plot ---

// GetSystemDiagram returns the topology of the currently active system
func (c *Canvas) GetSystemDiagram() (*protos.SystemDiagram, error) {
	if c.activeSystem == nil {
		return nil, fmt.Errorf("no active system set. Use Load() and Use() commands first")
	}

	systemName := c.activeSystem.System.Name.Value

	// Get current flow rates (no need to recalculate - they are kept up to date)
	currentFlowRates := c.GetCurrentFlowRates()

	// Extract nodes and edges from the system declaration
	var nodes []*protos.DiagramNode
	var edges []*protos.DiagramEdge
	instanceNameToID := make(map[string]string)

	// Build nodes from instance declarations
	fileInstance := c.activeSystem.File
	for _, item := range c.activeSystem.System.Body {
		if instDecl, ok := item.(*decl.InstanceDecl); ok {
			nodeID := instDecl.Name.Value
			instanceNameToID[nodeID] = nodeID

			// Get component methods using runtime system
			var methods []*protos.MethodInfo
			if c.runtime != nil {
				// Use runtime system to resolve component - it has already resolved imports
				component, err := fileInstance.GetComponentDecl(instDecl.ComponentName.Value)
				if err == nil && component != nil {
					componentMethods, _ := component.Methods()
					for methodName, methodDecl := range componentMethods {
						returnType := "void"
						if methodDecl.ReturnType != nil {
							returnType = methodDecl.ReturnType.Name
						}

						// Get traffic rate for this method from current flow rates
						methodTarget := fmt.Sprintf("%s.%s", nodeID, methodName)
						methodTraffic := currentFlowRates[methodTarget]

						// Only include methods with non-zero traffic to reduce clutter
						if methodTraffic > 0 {
							methods = append(methods, &protos.MethodInfo{
								Name:       methodName,
								ReturnType: returnType,
								Traffic:    methodTraffic,
							})
						}
					}
				}
			}

			// Calculate total traffic for this component
			componentTotalRPS := c.GetComponentTotalRPS(nodeID)
			var trafficDisplay string
			if componentTotalRPS > 0 {
				trafficDisplay = fmt.Sprintf("%.1f rps", componentTotalRPS)
			} else {
				trafficDisplay = "0 rps"
			}

			nodes = append(nodes, &protos.DiagramNode{
				Id:      nodeID,
				Name:    instDecl.Name.Value,
				Type:    instDecl.ComponentName.Value,
				Methods: methods,
				Traffic: trafficDisplay,
			})
		}
	}

	// Build edges from instance overrides/dependencies
	for _, item := range c.activeSystem.System.Body {
		if instDecl, ok := item.(*decl.InstanceDecl); ok {
			fromNodeID := instanceNameToID[instDecl.Name.Value]

			// Add edges from instance overrides (system-level dependencies)
			for _, assignment := range instDecl.Overrides {
				if targetIdent, okIdent := assignment.Value.(*decl.IdentifierExpr); okIdent {
					if toNodeID, isInstance := instanceNameToID[targetIdent.Value]; isInstance {
						edges = append(edges, &protos.DiagramEdge{
							FromId: fromNodeID,
							ToId:   toNodeID,
							Label:  assignment.Var.Value,
						})
					}
				}
			}
		}
	}

	// TODO: Add flow-based edges from runtime flow analysis
	// The runtime-based flow evaluation doesn't track flow paths in the same way
	// as the string-based implementation. This will need to be reimplemented
	// if flow visualization is needed.
	/*
		if c.currentFlowScope != nil && c.currentFlowRates != nil {
			// Future implementation of flow path visualization
		}
	*/

	// Note: Flow path visualization needs to be reimplemented for runtime-based evaluation
	// The runtime approach doesn't track individual flow paths like the string-based version

	return &protos.SystemDiagram{
		SystemName: systemName,
		Nodes:      nodes,
		Edges:      edges,
	}, nil
}

// / ----
// / ----
// / ----
// / ----
// / ----
// / ----
// / ----
// / ----
// / ----
// / ----

// evaluateProposedFlows calculates what the system flows would be with current generator settings
func (c *Canvas) evaluateProposedFlows() error {
	return c.evaluateProposedFlowsWithStrategy(c.currentFlowStrategy)
}

// evaluateProposedFlowsWithStrategy calculates flows using specified strategy
func (c *Canvas) evaluateProposedFlowsWithStrategy(strategy string) error {
	if c.activeSystem == nil {
		return fmt.Errorf("no active system available for flow calculation")
	}

	// Build API generator configs
	var generators []runtime.GeneratorConfigAPI
	for _, gen := range c.generators {
		if gen.Enabled {
			generators = append(generators, runtime.GeneratorConfigAPI{
				ID:        gen.Id,
				Component: gen.Component,
				Method:    gen.Method,
				Rate:      float64(gen.Rate),
			})
		}
	}

	// Evaluate flows using the strategy
	result, err := runtime.EvaluateFlowStrategy(strategy, c.activeSystem, generators)
	if err != nil {
		return err
	}

	// Convert results back to runtime format for compatibility
	// This is temporary until we fully migrate to the new API
	c.proposedFlowScope = runtime.NewFlowScope(c.activeSystem.Env)
	c.proposedFlowRates = c.convertFlowResultToRateMap(result)

	return nil
}

// convertFlowResultToRateMap converts API flow result back to RateMap
// This is a temporary compatibility layer
func (c *Canvas) convertFlowResultToRateMap(result *runtime.FlowAnalysisResult) runtime.RateMap {
	rateMap := runtime.NewRateMap()

	for compMethod, rate := range result.Flows.ComponentRates {
		parts := strings.Split(compMethod, ".")
		if len(parts) >= 2 {
			componentName := parts[0]
			methodName := strings.Join(parts[1:], ".")

			// Find the component instance
			compInst := c.activeSystem.FindComponent(componentName)
			if compInst != nil {
				rateMap.SetRate(compInst, methodName, rate)
			}
		}
	}

	// Apply manual overrides
	for compMethod, rate := range c.manualRateOverrides {
		parts := strings.Split(compMethod, ".")
		if len(parts) >= 2 {
			componentName := parts[0]
			methodName := strings.Join(parts[1:], ".")

			compInst := c.activeSystem.FindComponent(componentName)
			if compInst != nil {
				rateMap.SetRate(compInst, methodName, rate)
			}
		}
	}

	return rateMap
}

// applyProposedFlows moves the proposed flow state to current (accepting the new flow state)
func (c *Canvas) applyProposedFlows() {
	if c.proposedFlowScope != nil {
		c.currentFlowScope = c.proposedFlowScope
		c.proposedFlowScope = nil
		c.currentFlowRates = c.proposedFlowRates
		c.proposedFlowRates = nil
	}
}

// invalidateFlowContexts clears flow contexts when system state changes
func (c *Canvas) invalidateFlowContexts() {
	c.currentFlowScope = nil
	c.proposedFlowScope = nil
	c.currentFlowRates = nil
	c.proposedFlowRates = nil
}

// initializeFlowContexts sets up initial flow contexts for a new system
func (c *Canvas) initializeFlowContexts() {
	if c.activeSystem == nil {
		return
	}

	// Initialize current scope with system environment
	c.currentFlowScope = runtime.NewFlowScope(c.activeSystem.Env)
	c.currentFlowRates = runtime.NewRateMap()

	// Clear proposed context
	c.proposedFlowScope = nil
	c.proposedFlowRates = nil
}

// GetCurrentFlowRates returns the current (applied) flow rates
func (c *Canvas) GetCurrentFlowRates() map[string]float64 {
	if c.currentFlowRates == nil {
		return make(map[string]float64)
	}
	// Convert RateMap to map[string]float64 for API compatibility
	return c.rateMapToStringMap(c.currentFlowRates)
}

// GetProposedFlowRates returns the proposed flow rates (being evaluated)
func (c *Canvas) GetProposedFlowRates() map[string]float64 {
	if c.proposedFlowRates == nil {
		return make(map[string]float64)
	}
	// Convert RateMap to map[string]float64 for API compatibility
	return c.rateMapToStringMap(c.proposedFlowRates)
}

// GetComponentTotalRPS calculates total RPS for a component by summing all its methods
func (c *Canvas) GetComponentTotalRPS(componentID string) float64 {
	rates := c.GetCurrentFlowRates()
	total := 0.0
	prefix := componentID + "."

	for target, rps := range rates {
		if strings.HasPrefix(target, prefix) {
			total += rps
		}
	}

	return total
}

// Close closes the Canvas and cleans up resources
func (c *Canvas) Close() error {
	if c.metricTracer != nil {
		c.metricTracer.Clear()
		c.metricTracer = nil
	}
	return nil
}

// Reset clears the canvas completely - stops all generators, removes metrics, and resets state
func (c *Canvas) Reset() error {
	// Stop all generators
	c.generatorsLock.Lock()
	for _, gen := range c.generators {
		gen.Stop()
	}
	c.generators = make(map[string]*GeneratorInfo)
	c.generatorsLock.Unlock()

	// Clear metrics
	if c.metricTracer != nil {
		c.metricTracer.Clear()
	}

	// Reset system state
	c.activeSystem = nil
	c.loadedSystems = make(map[string]*runtime.SystemInstance)

	// Reset flow state
	c.currentFlowScope = nil
	c.proposedFlowScope = nil
	c.currentFlowRates = nil
	c.proposedFlowRates = nil
	c.currentFlowStrategy = runtime.GetDefaultFlowStrategy()
	c.manualRateOverrides = make(map[string]float64)

	// Reset simulation time
	c.simulationStarted = false

	// Also reset the loader and runtime as we want a clean slate.
	// Perhaps later we can see hwo to reuse them or only clear based on req flags.
	loader := loader.NewLoader(nil, nil, 10)
	c.runtime = runtime.NewRuntime(loader)

	return nil
}

// ExecuteTrace runs a single method call and returns detailed trace data
func (c *Canvas) ExecuteTrace(componentName, methodName string) (*runtime.TraceData, error) {
	if c.activeSystem == nil {
		return nil, status.Error(codes.FailedPrecondition, "no active system. Call Use() before executing trace")
	}

	// Find the component instance
	compInst := c.activeSystem.FindComponent(componentName)
	if compInst == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("component '%s' not found in system when executing trace", componentName))
	}

	// Check method exists
	methodDecl, err := compInst.ComponentDecl.GetMethod(methodName)
	if err != nil || methodDecl == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("method '%s' not found in component '%s'", methodName, componentName))
	}

	// Create execution tracer
	tracer := runtime.NewExecutionTracer()
	tracer.SetRuntime(c.runtime)

	// Create evaluator with the tracer
	eval := runtime.NewSimpleEval(c.activeSystem.File, tracer)

	// Execute the method call
	env := c.activeSystem.Env.Push()
	var currTime core.Duration = 0

	// Create a call expression to invoke the method
	callExpr := &decl.CallExpr{
		Function: &decl.MemberAccessExpr{
			Receiver: &decl.IdentifierExpr{Value: componentName},
			Member:   &decl.IdentifierExpr{Value: methodName},
		},
	}

	// Execute the call
	_, _ = eval.Eval(callExpr, env, &currTime)

	// Build trace data
	traceData := &runtime.TraceData{
		System:     c.activeSystem.System.Name.Value,
		EntryPoint: fmt.Sprintf("%s.%s", componentName, methodName),
		Events:     tracer.Events,
	}

	return traceData, nil
}

// TraceAllPaths performs breadth-first traversal to discover all possible execution paths
func (c *Canvas) TraceAllPaths(componentName, methodName string, maxDepth int32) (*runtime.AllPathsTraceData, error) {
	if c.activeSystem == nil {
		return nil, status.Error(codes.FailedPrecondition, "no active system. Call Use() before tracing paths")
	}

	// Find the component instance to get its type
	compInst := c.activeSystem.FindComponent(componentName)
	if compInst == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("component '%s' not found in system", componentName))
	}

	// Create path traversal engine with the loader
	pathTraversal := runtime.NewPathTraversal(c.runtime.Loader)

	// Perform the traversal using the component type
	traceData, err := pathTraversal.TraceAllPaths(componentName, compInst.ComponentDecl, methodName, maxDepth)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to trace all paths: %v", err))
	}

	return traceData, nil
}

// rateMapToStringMap converts a runtime RateMap to string-based map for API compatibility
func (c *Canvas) rateMapToStringMap(rateMap runtime.RateMap) map[string]float64 {
	result := make(map[string]float64)

	// Build a reverse map from component instance to variable name
	instanceToName := make(map[*runtime.ComponentInstance]string)
	if c.activeSystem != nil && c.activeSystem.System != nil {
		// Look through the system declaration to find instance names
		for _, item := range c.activeSystem.System.Body {
			if instDecl, ok := item.(*decl.InstanceDecl); ok {
				instanceName := instDecl.Name.Value
				// Try to find the actual component instance
				if comp := c.activeSystem.FindComponent(instanceName); comp != nil {
					instanceToName[comp] = instanceName
				}
			}
		}
	}

	// Convert each component.method entry to "name.method" string key
	for component, methods := range rateMap {
		if component != nil {
			// Use variable name if available, otherwise fall back to ID
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

// Flow Strategy API Methods

// GetFlowStrategies returns all available flow strategies
func (c *Canvas) GetFlowStrategies() map[string]runtime.StrategyInfo {
	return runtime.ListFlowStrategies()
}

// EvaluateFlowWithStrategy evaluates flows using specified strategy
func (c *Canvas) EvaluateFlowWithStrategy(strategy string) (*runtime.FlowAnalysisResult, error) {
	if c.activeSystem == nil {
		return nil, fmt.Errorf("no active system")
	}

	// Build API generator configs
	var generators []runtime.GeneratorConfigAPI
	for _, gen := range c.generators {
		if gen.Enabled {
			generators = append(generators, runtime.GeneratorConfigAPI{
				ID:        gen.Id,
				Component: gen.Component,
				Method:    gen.Method,
				Rate:      float64(gen.Rate),
			})
		}
	}

	return runtime.EvaluateFlowStrategy(strategy, c.activeSystem, generators)
}

// ApplyFlowStrategy applies flows from specified strategy as current arrival rates
func (c *Canvas) ApplyFlowStrategy(strategy string) error {
	// Evaluate flows with the strategy
	err := c.evaluateProposedFlowsWithStrategy(strategy)
	if err != nil {
		return err
	}

	// Apply the proposed flows
	c.applyProposedFlows()
	c.currentFlowStrategy = strategy

	return nil
}

// GetCurrentFlowState returns the current flow state
func (c *Canvas) GetCurrentFlowState() *runtime.FlowState {
	rates := make(map[string]float64)

	// Convert current flow rates to API format
	if c.currentFlowRates != nil {
		for comp, methods := range c.currentFlowRates {
			// Find component name
			compName := ""
			for name, value := range c.activeSystem.Env.All() {
				if compInst, ok := value.Value.(*runtime.ComponentInstance); ok && compInst == comp {
					compName = name
					break
				}
			}

			if compName != "" {
				for method, rate := range methods {
					key := fmt.Sprintf("%s.%s", compName, method)
					rates[key] = rate
				}
			}
		}
	}

	return &runtime.FlowState{
		Strategy:        c.currentFlowStrategy,
		Rates:           rates,
		ManualOverrides: c.manualRateOverrides,
	}
}

// SetComponentArrivalRate sets a manual arrival rate override
func (c *Canvas) SetComponentArrivalRate(component, method string, rate float64) error {
	if c.activeSystem == nil {
		return fmt.Errorf("no active system")
	}

	// Verify component exists
	compInst := c.activeSystem.FindComponent(component)
	if compInst == nil {
		return fmt.Errorf("component '%s' not found", component)
	}

	// Store the override
	key := fmt.Sprintf("%s.%s", component, method)
	c.manualRateOverrides[key] = rate

	// Recompute flows to apply the override
	return c.recomputeSystemFlows()
}

// recomputeSystemFlows evaluates proposed flows and automatically applies them
func (c *Canvas) recomputeSystemFlows() error {
	// Evaluate what the new flows would be
	err := c.evaluateProposedFlows()
	if err != nil {
		return err
	}

	// Automatically apply the proposed flows
	// In the future, we could add confirmation logic here
	c.applyProposedFlows()

	return nil
}
