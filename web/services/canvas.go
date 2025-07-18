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

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
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

// NewCanvas creates a new interactive canvas session with a custom runtime.
// If runtime is nil, a default runtime with no resolvers will be created.
func NewCanvas(id string, r *runtime.Runtime) *Canvas {
	if r == nil {
		l := loader.NewLoader(nil, nil, 10)
		r = runtime.NewRuntime(l)
	}
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

// GetRuntime returns the runtime instance for this canvas
func (c *Canvas) GetRuntime() *runtime.Runtime {
	return c.runtime
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

		err := gen.Stop(true)
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
	if c.generators[gen.ID] != nil {
		return status.Error(codes.AlreadyExists, "Generator with name already exists")
	}
	c.generators[gen.ID] = gen
	gen.Start()
	return c.recomputeSystemFlows()
}

// UpdateGenerator updates an existing traffic generator configuration
func (c *Canvas) UpdateGenerator(gen *Generator) error {
	c.generatorsLock.Lock()
	defer c.generatorsLock.Unlock()

	existing, exists := c.generators[gen.ID]
	if !exists {
		return status.Error(codes.NotFound, fmt.Sprintf("generator '%s' not found", gen.ID))
	}

	// Stop the existing generator if it's running
	wasRunning := existing.IsRunning()
	if wasRunning {
		existing.Stop(true)
	}

	// Only update rate and name - component and method should not change
	existing.Name = gen.Name
	existing.Rate = gen.Rate
	// existing.Enabled = gen.Enabled

	// Update the proto representation with the new values
	existing.Generator.Name = gen.Name
	existing.Generator.Rate = gen.Rate
	// existing.Generator.Enabled = gen.Enabled

	// Restart the generator if it was running and is still enabled
	if wasRunning /* && existing.Enabled */ {
		existing.Start()
	}

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
	gen.Stop(true)
	return c.recomputeSystemFlows()
}

func (c *Canvas) StopGenerator(genId string) error {
	c.generatorsLock.Lock()
	defer c.generatorsLock.Unlock()

	if c.generators[genId] == nil {
		return status.Error(codes.NotFound, "Generator not found")
	}
	c.generators[genId].Stop(true)
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

func (c *Canvas) GetGenerator(id string) *GeneratorInfo {
	c.generatorsLock.RLock()
	defer c.generatorsLock.RUnlock()
	return c.generators[id]
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

// GetSystemDiagram returns the system topology with method-level nodes and edges
func (c *Canvas) GetSystemDiagram() (*SystemDiagram, error) {
	if c.activeSystem == nil {
		return nil, fmt.Errorf("no active system set. Use Load() and Use() commands first")
	}

	// Track component instances and their paths
	instancePaths := make(map[*runtime.ComponentInstance][]string)
	pathToInstance := make(map[string]*runtime.ComponentInstance)

	// Step 1: Build the instance and path map
	rootInstances := c.buildInstancePaths(instancePaths, pathToInstance)
	log.Println("Root Instances: ", rootInstances)
	for comp, ips := range instancePaths {
		log.Printf("FOUND INSTANCE PATHS: CompId: %p, Decl: %s, IPS: %s", comp, comp.ComponentDecl.Name.Value, strings.Join(ips, ", "))
	}

	var nodes []DiagramNode
	var edges []DiagramEdge

	// Track method nodes we've created
	methodNodes := make(map[string]*DiagramNode) // "component:method" -> node
	processedEdges := make(map[string]bool)      // track "from->to" to avoid duplicates

	// Get the current flow rates for traffic labels
	currentFlowRates := c.GetCurrentFlowRates()

	// Helper to find the primary path for an instance
	getPrimaryPath := func(inst *runtime.ComponentInstance) string {
		if paths, ok := instancePaths[inst]; ok && len(paths) > 0 {
			return paths[0]
		}
		return ""
	}

	// Helper to create or get a method node
	getOrCreateMethodNode := func(inst *runtime.ComponentInstance, methodName string) *DiagramNode {
		path := getPrimaryPath(inst)
		if path == "" {
			return nil
		}

		nodeId := fmt.Sprintf("%s:%s", path, methodName)
		if node, exists := methodNodes[nodeId]; exists {
			return node
		}

		// Get rate for this method if available
		rateKey := fmt.Sprintf("%s.%s", path, methodName)
		rate := currentFlowRates[rateKey]

		node := &DiagramNode{
			ID:   nodeId,
			Name: nodeId,
			Type: inst.ComponentDecl.Name.Value,
			Methods: []MethodInfo{{
				Name:    methodName,
				Traffic: rate,
			}},
			FullPath: path,
			Icon:     c.getComponentIcon(inst),
			Traffic:  "",
		}
		if rate > 0 {
			node.Traffic = fmt.Sprintf("%.1f rps", rate)
		}
		methodNodes[nodeId] = node
		nodes = append(nodes, *node)
		return node
	}

	// Recursive helper to process a specific method and its calls
	var processMethodCalls func(inst *runtime.ComponentInstance, methodName string,
		processedMethods map[string]bool,
		getOrCreateMethodNode func(*runtime.ComponentInstance, string) *DiagramNode,
		instancePaths map[*runtime.ComponentInstance][]string,
		currentFlowRates map[string]float64,
		nodes *[]DiagramNode,
		edges *[]DiagramEdge,
		flowScope *runtime.FlowScope)

	processMethodCalls = func(inst *runtime.ComponentInstance, methodName string,
		processedMethods map[string]bool,
		getOrCreateMethodNode func(*runtime.ComponentInstance, string) *DiagramNode,
		instancePaths map[*runtime.ComponentInstance][]string,
		currentFlowRates map[string]float64,
		nodes *[]DiagramNode,
		edges *[]DiagramEdge,
		flowScope *runtime.FlowScope) {

		// Skip if we've already processed this method
		path := getPrimaryPath(inst)
		if path == "" {
			return
		}

		methodKey := fmt.Sprintf("%s:%s", path, methodName)
		if processedMethods[methodKey] {
			return
		}
		processedMethods[methodKey] = true

		// Find all calls this specific method makes
		neighbors := inst.NeighborsFromMethod(methodName)
		for _, neighbor := range neighbors {
			// Create node for the called method
			toNode := getOrCreateMethodNode(neighbor.Component, neighbor.MethodName)
			if toNode == nil {
				continue
			}

			// Create edge
			fromNode := getOrCreateMethodNode(inst, methodName)
			if fromNode != nil {
				edgeKey := fmt.Sprintf("%s->%s", fromNode.ID, toNode.ID)
				if !processedEdges[edgeKey] {
					processedEdges[edgeKey] = true

					// Get rate if available from flow analysis
					rate := 0.0
					if flowScope != nil && flowScope.FlowEdges != nil {
						for _, flowEdge := range flowScope.FlowEdges.GetEdges() {
							if flowEdge.FromComponent == inst &&
								flowEdge.FromMethod == methodName &&
								flowEdge.ToComponent == neighbor.Component &&
								flowEdge.ToMethod == neighbor.MethodName {
								rate = flowEdge.Rate
								break
							}
						}
					}

					newedge := &DiagramEdge{
						FromID:     fromNode.ID,
						ToID:       toNode.ID,
						FromMethod: methodName,
						ToMethod:   neighbor.MethodName,
						Label:      "",
					}
					if rate > 0 {
						newedge.Label = fmt.Sprintf("%.1f rps", rate)
					}
					*edges = append(*edges, *newedge)
				}
			}

			// Recursively process the called method
			processMethodCalls(neighbor.Component, neighbor.MethodName, processedMethods,
				getOrCreateMethodNode, instancePaths, currentFlowRates, nodes, edges, flowScope)
		}
	}

	// Track which methods have been processed to avoid duplicates
	processedMethods := make(map[string]bool)

	// Process all root instances and traverse their methods
	for _, rootInst := range rootInstances {
		// Process each method in the root instance
		methods, _ := rootInst.ComponentDecl.Methods()
		for _, method := range methods {
			methodName := method.Name.Value

			// Create node for this root method
			fromNode := getOrCreateMethodNode(rootInst, methodName)
			if fromNode == nil {
				continue
			}

			// Process all calls from this method recursively
			processMethodCalls(rootInst, methodName, processedMethods,
				getOrCreateMethodNode, instancePaths, currentFlowRates, &nodes, &edges, c.currentFlowScope)
		}
	}

	// Also add any generator entry points that might not have been traversed
	for _, gen := range c.generators {
		if gen.IsRunning() && gen.Rate > 0 {
			if inst, ok := pathToInstance[gen.Component]; ok {
				getOrCreateMethodNode(inst, gen.Method)
			}
		}
	}

	// Track component-only nodes (for components without methods)
	componentNodes := make(map[string]*DiagramNode) // "component" -> node

	// Create component-only nodes for components that have no methods with traffic
	// This ensures we still show the system structure even without traffic
	for inst, paths := range instancePaths {
		if len(paths) > 0 {
			primaryPath := paths[0]
			hasMethodNode := false

			// Check if any method node exists for this component
			for nodeId := range methodNodes {
				if strings.HasPrefix(nodeId, primaryPath+":") {
					hasMethodNode = true
					break
				}
			}

			// If no method nodes exist, create a component-only node
			if !hasMethodNode {
				node := &DiagramNode{
					ID:       primaryPath,
					Name:     primaryPath,
					Type:     inst.ComponentDecl.Name.Value,
					Methods:  []MethodInfo{},
					Traffic:  "-",
					FullPath: primaryPath,
					Icon:     c.getComponentIcon(inst),
				}
				componentNodes[primaryPath] = node
				nodes = append(nodes, *node)
			}
		}
	}

	// Add structural edges between component-only nodes (parent-child relationships)
	for _, node := range componentNodes {
		componentPath := node.ID

		// Check if this component has a parent
		lastDot := strings.LastIndex(componentPath, ".")
		if lastDot > 0 {
			parentPath := componentPath[:lastDot]

			// Only add edge if parent also has a component-only node
			if _, hasParentNode := componentNodes[parentPath]; hasParentNode {
				edges = append(edges, DiagramEdge{
					FromID: parentPath,
					ToID:   componentPath,
					Label:  "",
				})
			}
		}
	}

	systemName := ""
	if c.activeSystem.System != nil {
		systemName = c.activeSystem.System.Name.Value
	}

	return &SystemDiagram{
		SystemName: systemName,
		Nodes:      nodes,
		Edges:      edges,
	}, nil
}

// getComponentIcon determines the appropriate icon for a component based on its type and characteristics
func (c *Canvas) getComponentIcon(inst *runtime.ComponentInstance) string {
	if inst == nil || inst.ComponentDecl == nil {
		return "component" // default icon
	}

	compType := inst.ComponentDecl.Name.Value
	compTypeLower := strings.ToLower(compType)

	// Check native component types first
	switch compType {
	case "Cache", "CacheWithContention":
		return "cache"
	case "Database":
		return "database"
	case "ResourcePool":
		return "pool"
	case "MM1Queue", "MMCKQueue":
		return "queue"
	case "Link":
		return "network"
	case "HashIndex", "BTreeIndex", "BitmapIndex":
		return "index"
	case "SortedFile", "HeapFile", "LSMTree":
		return "storage"
	}

	// Check by naming patterns
	if strings.Contains(compTypeLower, "service") {
		return "service"
	}
	if strings.Contains(compTypeLower, "gateway") {
		return "gateway"
	}
	if strings.Contains(compTypeLower, "api") {
		return "api"
	}
	if strings.Contains(compTypeLower, "cache") {
		return "cache"
	}
	if strings.Contains(compTypeLower, "database") || strings.Contains(compTypeLower, "db") {
		return "database"
	}
	if strings.Contains(compTypeLower, "queue") {
		return "queue"
	}
	if strings.Contains(compTypeLower, "pool") {
		return "pool"
	}

	// Check by dependencies - if it has certain types of dependencies, infer its role
	deps, _ := inst.ComponentDecl.Dependencies()
	hasDatabaseDep := false
	hasCacheDep := false
	hasPoolDep := false

	for _, dep := range deps {
		if dep.ResolvedComponent != nil {
			depType := dep.ResolvedComponent.Name.Value
			if depType == "Database" || strings.Contains(strings.ToLower(depType), "db") {
				hasDatabaseDep = true
			}
			if depType == "Cache" || strings.Contains(strings.ToLower(depType), "cache") {
				hasCacheDep = true
			}
			if depType == "ResourcePool" || strings.Contains(strings.ToLower(depType), "pool") {
				hasPoolDep = true
			}
		}
	}

	// Infer based on dependencies
	if hasDatabaseDep && hasCacheDep {
		return "service" // Likely a service that uses both cache and database
	}
	if hasPoolDep {
		return "service" // Components with pools are typically services
	}

	// Default icon
	return "component"
}

// Helper to build instance path map
// This will help us get all the unique component instances in a system (even nested ones)
func (c *Canvas) buildInstancePaths(instancePaths map[*runtime.ComponentInstance][]string, pathToInstance map[string]*runtime.ComponentInstance) (rootInstances []*runtime.ComponentInstance) {
	type queueItem struct {
		instance *runtime.ComponentInstance
		path     string
	}

	var queue []queueItem
	systemEnv := c.activeSystem.Env

	// Start with top-level instances
	for varName, value := range systemEnv.All() {
		if varName == "self" {
			continue
		}
		if compInst, ok := value.Value.(*runtime.ComponentInstance); ok {
			queue = append(queue, queueItem{instance: compInst, path: varName})
			instancePaths[compInst] = append(instancePaths[compInst], varName)
			pathToInstance[varName] = compInst
		}
	}

	// BFS to find all instances and their paths
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.instance.Env != nil {
			for varName, value := range item.instance.Env.All() {
				if varName == "self" {
					continue
				}
				if subInst, ok := value.Value.(*runtime.ComponentInstance); ok {
					subPath := fmt.Sprintf("%s.%s", item.path, varName)
					queue = append(queue, queueItem{instance: subInst, path: subPath})
					instancePaths[subInst] = append(instancePaths[subInst], subPath)
					pathToInstance[subPath] = subInst
				}
			}
		}
	}

	for _, v := range instancePaths {
		if len(v) == 1 {
			rootInstances = append(rootInstances, pathToInstance[v[0]])
		}
	}
	return
}

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
				ID:        gen.ID,
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

	// Also populate the FlowScope's ArrivalRates so ApplyToComponents works
	c.proposedFlowScope.ArrivalRates = c.proposedFlowRates

	// Populate FlowEdges from the result
	if c.proposedFlowScope.FlowEdges != nil {
		c.proposedFlowScope.FlowEdges.Clear()
		for _, edge := range result.Flows.Edges {
			// Find component instances from names
			fromComp := c.activeSystem.FindComponent(edge.From.Component)
			toComp := c.activeSystem.FindComponent(edge.To.Component)
			if fromComp != nil && toComp != nil {
				c.proposedFlowScope.FlowEdges.AddEdge(
					fromComp,
					edge.From.Method,
					toComp,
					edge.To.Method,
					edge.Rate,
				)
			}
		}
	}

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
// and applies the arrival rates to the actual components
func (c *Canvas) applyProposedFlows() {
	if c.proposedFlowScope != nil {
		c.currentFlowScope = c.proposedFlowScope
		c.proposedFlowScope = nil
		c.currentFlowRates = c.proposedFlowRates
		c.proposedFlowRates = nil

		// Apply the calculated arrival rates to the actual components
		if c.currentFlowScope != nil {
			if err := c.currentFlowScope.ApplyToComponents(); err != nil {
				// Log the error but don't fail - some components might not support arrival rates
				slog.Warn("Failed to apply some arrival rates", "error", err)
			}
		}
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
		gen.Stop(true)
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

// GetSimulationTime returns the current simulation time in seconds
// For now, we return 0 as we don't have virtual time tracking yet
func (c *Canvas) GetSimulationTime() float64 {
	// TODO: Implement proper virtual time tracking
	// This would track the virtual time based on generator events
	return 0
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
				ID:        gen.ID,
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

// BatchSetParameters sets multiple parameters atomically
func (c *Canvas) BatchSetParameters(updates map[string]any) (map[string]decl.Value, error) {
	if c.activeSystem == nil || c.activeSystem.Env == nil {
		return nil, fmt.Errorf("no active system. Call Use() before BatchSetParameters()")
	}

	var newValues []decl.Value
	var paramPaths []string

	// First pass: validate all parameters
	for path, value := range updates {
		paramPaths = append(paramPaths, path)

		// Convert value to decl.Value
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
			err = fmt.Errorf("param: %s, unsupported value type: %T", path, value)
		}

		if err != nil {
			return nil, err
		}
		newValues = append(newValues, newValue)
	}

	// Second pass: apply all validated updates
	// Now call the BatchSet in the runtime
	oldValues := map[string]decl.Value{}
	err := c.runtime.BatchSetParams(c.activeSystem, paramPaths, newValues, oldValues)
	if err != nil {
		return nil, err
	}

	// Recompute flows after parameter changes
	err = c.recomputeSystemFlows()
	if err != nil {
		err = fmt.Errorf("parameters set but flow recomputation failed: %w", err)
	}

	return oldValues, nil
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
