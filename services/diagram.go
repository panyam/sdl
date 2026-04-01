package services

import (
	"fmt"
	"log"
	"strings"

	"github.com/panyam/sdl/lib/runtime"
)

// BuildSystemDiagram creates a system topology diagram from the given system state.
// This is a standalone function so both Canvas and DevEnv can use it.
func BuildSystemDiagram(
	system *runtime.SystemInstance,
	generators map[string]*GeneratorInfo,
	flowScope *runtime.FlowScope,
	currentFlowRates map[string]float64,
) (*SystemDiagram, error) {
	if system == nil {
		return nil, fmt.Errorf("no active system set")
	}

	// Track component instances and their paths
	instancePaths := make(map[*runtime.ComponentInstance][]string)
	pathToInstance := make(map[string]*runtime.ComponentInstance)

	// Step 1: Build the instance and path map
	rootInstances := buildInstancePaths(system, instancePaths, pathToInstance)
	log.Println("Root Instances: ", rootInstances)
	for comp, ips := range instancePaths {
		log.Printf("FOUND INSTANCE PATHS: CompId: %p, Decl: %s, IPS: %s", comp, comp.ComponentDecl.Name.Value, strings.Join(ips, ", "))
	}

	var nodes []DiagramNode
	var edges []DiagramEdge

	// Track method nodes we've created
	methodNodes := make(map[string]*DiagramNode) // "component:method" -> node
	processedEdges := make(map[string]bool)      // track "from->to" to avoid duplicates

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
			Icon:     getComponentIcon(inst),
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
				getOrCreateMethodNode, instancePaths, currentFlowRates, &nodes, &edges, flowScope)
		}
	}

	// Also add any generator entry points that might not have been traversed
	for _, gen := range generators {
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
					Icon:     getComponentIcon(inst),
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
	if system.System != nil {
		systemName = system.System.Name.Value
	}

	return &SystemDiagram{
		SystemName: systemName,
		Nodes:      nodes,
		Edges:      edges,
	}, nil
}

// getComponentIcon determines the appropriate icon for a component based on its type and characteristics.
func getComponentIcon(inst *runtime.ComponentInstance) string {
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

// buildInstancePaths builds a map of component instances to their dotted paths,
// and returns the root-level instances.
func buildInstancePaths(system *runtime.SystemInstance, instancePaths map[*runtime.ComponentInstance][]string, pathToInstance map[string]*runtime.ComponentInstance) (rootInstances []*runtime.ComponentInstance) {
	type queueItem struct {
		instance *runtime.ComponentInstance
		path     string
	}

	var queue []queueItem
	systemEnv := system.Env

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
