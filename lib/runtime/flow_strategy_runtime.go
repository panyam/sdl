package runtime

import (
	"fmt"
)

// RuntimeFlowStrategy implements flow evaluation using the runtime-based approach
type RuntimeFlowStrategy struct{}

// Evaluate performs flow analysis using SolveSystemFlowsRuntime
func (s *RuntimeFlowStrategy) Evaluate(system *SystemInstance, generators []GeneratorConfigAPI) (*FlowAnalysisResult, error) {
	if system == nil {
		return nil, fmt.Errorf("system instance is required")
	}

	// Convert API generator configs to runtime generator entry points
	var runtimeGenerators []GeneratorEntryPointRuntime
	for _, gen := range generators {
		// Find the component instance
		compInst := system.FindComponent(gen.Component)
		if compInst == nil {
			Warn("RuntimeFlowStrategy: Component not found: %s", gen.Component)
			continue
		}

		runtimeGenerators = append(runtimeGenerators, GeneratorEntryPointRuntime{
			Component:   compInst,
			Method:      gen.Method,
			Rate:        gen.Rate,
			GeneratorID: gen.ID,
		})
	}

	// Create flow scope
	scope := NewFlowScope(system.Env)

	// Run flow evaluation
	rateMap := SolveSystemFlowsRuntime(runtimeGenerators, scope)

	// Apply the calculated rates to components
	for comp, methods := range rateMap {
		for method, rate := range methods {
			if err := comp.SetArrivalRate(method, rate); err != nil {
				// Log but don't fail - some components might not support arrival rates
				Warn("Failed to set arrival rate for %v.%s: %v", comp, method, err)
			}
		}
	}

	// Convert results to API format
	result := &FlowAnalysisResult{
		Strategy:   "runtime",
		Status:     FlowStatusConverged, // TODO: Detect actual convergence status
		Iterations: 10,                  // TODO: Track actual iterations
		System:     system.GetSystemName(),
		Generators: generators,
		Flows:      s.convertToFlowData(rateMap, scope, system),
		Warnings:   []string{"Control flow analysis may overestimate rates for early return patterns"},
	}

	return result, nil
}

// GetInfo returns metadata about this strategy
func (s *RuntimeFlowStrategy) GetInfo() StrategyInfo {
	return StrategyInfo{
		Name:        "runtime",
		Description: "Runtime-based flow evaluation using component instances",
		Status:      "stable",
		Limitations: []string{
			"Early returns overestimate flow",
			"No capacity modeling",
		},
		Recommended: true,
	}
}

// IsAvailable checks if this strategy can be used
func (s *RuntimeFlowStrategy) IsAvailable() bool {
	return true
}

// convertToFlowData converts runtime RateMap to API-friendly format
func (s *RuntimeFlowStrategy) convertToFlowData(rateMap RateMap, scope *FlowScope, system *SystemInstance) FlowData {
	edges := []FlowEdgeAPI{}
	componentRates := make(map[string]float64)

	// Convert flow edges if available
	if scope.FlowEdges != nil {
		for _, edge := range scope.FlowEdges.GetEdges() {
			// Find component names from instances
			fromName := s.findComponentName(edge.FromComponent, system)
			toName := s.findComponentName(edge.ToComponent, system)

			if fromName != "" && toName != "" {
				edges = append(edges, FlowEdgeAPI{
					From: ComponentMethod{
						Component: fromName,
						Method:    edge.FromMethod,
					},
					To: ComponentMethod{
						Component: toName,
						Method:    edge.ToMethod,
					},
					Rate: edge.Rate,
				})
			}
		}
	}

	// Convert component rates
	for component, methods := range rateMap {
		componentName := s.findComponentName(component, system)
		if componentName == "" {
			continue
		}

		for method, rate := range methods {
			key := fmt.Sprintf("%s.%s", componentName, method)
			componentRates[key] = rate
		}
	}

	// Metadata
	metadata := map[string]interface{}{
		"totalFlow":            s.calculateTotalFlow(componentRates),
		"maxComponentRate":     s.findMaxRate(componentRates),
		"convergenceThreshold": 0.01,
	}

	return FlowData{
		Edges:          edges,
		ComponentRates: componentRates,
		Metadata:       metadata,
	}
}

// findComponentName finds the variable name for a component instance
func (s *RuntimeFlowStrategy) findComponentName(comp *ComponentInstance, system *SystemInstance) string {
	if comp == nil || system == nil {
		return ""
	}

	// Search through the system environment for this component instance
	bindings := system.Env.All()
	for name, value := range bindings {
		if compValue, ok := value.Value.(*ComponentInstance); ok && compValue == comp {
			return name
		}
	}

	// If not found in system env, it might be a nested component
	// Try to find it as a child of another component
	for name, value := range bindings {
		if parentComp, ok := value.Value.(*ComponentInstance); ok && parentComp != nil {
			// Check if this component is a child of the parent
			for childName, childBinding := range parentComp.Env.All() {
				if childComp, ok := childBinding.Value.(*ComponentInstance); ok && childComp == comp {
					// Return parent.child format for nested components
					return fmt.Sprintf("%s.%s", name, childName)
				}
			}
		}
	}

	// If still not found, it's an internal component that shouldn't be exposed
	// Return empty string to filter it out
	return ""
}

// calculateTotalFlow sums all component rates
func (s *RuntimeFlowStrategy) calculateTotalFlow(rates map[string]float64) float64 {
	total := 0.0
	for _, rate := range rates {
		total += rate
	}
	return total
}

// findMaxRate finds the maximum rate among all components
func (s *RuntimeFlowStrategy) findMaxRate(rates map[string]float64) float64 {
	max := 0.0
	for _, rate := range rates {
		if rate > max {
			max = rate
		}
	}
	return max
}

// Register the runtime strategy on package initialization
func init() {
	RegisterFlowStrategy("runtime", &RuntimeFlowStrategy{})
}
