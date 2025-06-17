package runtime

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/panyam/sdl/decl"
)

type SystemInstance struct {
	// Where the system is defined
	File   *FileInstance
	System *SystemDecl

	// Holds the component instances and parameters
	Env *Env[Value]
	
	// Flow state - system owns its complete flow state
	FlowContext *FlowContext `json:"flowContext"`
}

// Initializes a new runtime System instance and its root environment
func NewSystemInstance(file *FileInstance, system *SystemDecl) *SystemInstance {
	sysinst := &SystemInstance{
		File:   file, 
		System: system,
		// Initialize FlowContext with system and empty parameters
		FlowContext: NewFlowContext(system, make(map[string]interface{})),
	}
	
	// Register native components for flow analysis
	sysinst.RegisterNativeComponents()
	
	return sysinst
}

// A system declaration contains instantiations of components and other statemetns.
// Specifically in initializers it is important to not be bound by order.
// This method compiles the System into a set of statements that can be executed so that
// all components are intantiated first and then their properties/params are set.
func (s *SystemInstance) Initializer() (blockStmt *BlockStmt, err error) {
	// 1. a New Expression for constructing a componet
	// 2. a set expression for setting params/dependencies in a component - this way we can avoid the two pass approach?
	// This two passs approach needs to be repeated in every stage (eg type checking)
	var stmts []Stmt
	var instanceDecls []*InstanceDecl

	// Pass 1 - Create all the NewExprs and add the LetStmt to our list
	for _, item := range s.System.Body {
		switch it := item.(type) {
		case *InstanceDecl:
			// 1. Find the component definition in this file
			// 2. Instantiate it - it should be an expression evaluation (not overrides will be handled in pass 2)

			// if it is an instance declaration, find the component type from the environment
			// and create a new instance of the component type
			instanceDecls = append(instanceDecls, it)
			// Resolve the compdecl first
			compDecl, err := s.File.GetComponentDecl(it.ComponentName.Value)
			ensureNoErr(err)
			stmts = append(stmts, &decl.SetStmt{
				TargetExpr: it.Name,
				Value:      NewNewExpr(compDecl),
			})
		case *LetStmt:
			// Add this as is
			stmts = append(stmts, it)
		default:
			log.Fatal("Invalid type: ", it, reflect.TypeOf(it))
			// i.Errorf(item.Pos(), "type inference for system body item type %T not implemented yet", item)
		}
	}

	// Pass 2 - Add SetExpr statements to enable the overrides
	for _, it := range instanceDecls {
		for _, assign := range it.Overrides {
			stmts = append(stmts, &decl.SetStmt{
				TargetExpr: &MemberAccessExpr{
					Receiver: it.Name,
					Member:   assign.Var,
				},
				Value: assign.Value,
			})
		}
	}
	return &BlockStmt{Statements: stmts}, nil
}

type InitStmt struct {
	From     *InitStmt
	Pos      Location
	Attrib   string
	CompInst *ComponentInstance // this should be From.CompInst.Attrib.  If From == nil then this is a System level component
}

// Goes through all components and gets uninitialized components so user knows what/how to set them
// This is usually called after the Initializer expression is called but before any other expressions are called.
func (s *SystemInstance) GetUninitializedComponents(env *Env[Value]) (items []*InitStmt) {
	var visit func(i *InitStmt)
	visit = func(i *InitStmt) {
		compDecl := i.CompInst.ComponentDecl
		deps, _ := compDecl.Dependencies()
		for _, dep := range deps {
			depInst, ok := i.CompInst.Get(dep.Name.Value)
			if !ok || depInst.IsNil() {
				items = append(items, &InitStmt{
					From:   i,
					Pos:    compDecl.Pos(),
					Attrib: dep.Name.Value,
				})
			} else {
				visit(&InitStmt{
					From:     i,
					Pos:      compDecl.Pos(),
					Attrib:   dep.Name.Value,
					CompInst: depInst.Value.(*ComponentInstance),
				})
			}
		}
	}

	for _, item := range s.System.Body {
		it, ok := item.(*InstanceDecl)
		if !ok {
			continue
		}

		compValue, ok := env.Get(it.Name.Value)
		if !ok {
			items = append(items, &InitStmt{
				Pos:    item.Pos(),
				Attrib: it.Name.Value,
			})
			continue
		}

		compInst := compValue.Value.(*ComponentInstance)
		visit(&InitStmt{
			Pos:      item.Pos(),
			Attrib:   it.Name.Value,
			CompInst: compInst,
		})
	}
	return
}

// UpdateMethodArrivalRate updates the arrival rate for a specific method on a component
// and triggers FlowEval to recompute downstream effects.
func (si *SystemInstance) UpdateMethodArrivalRate(componentName, methodName string, rate float64) error {
	// Get component instance
	compVal, ok := si.Env.Get(componentName)
	if !ok {
		return fmt.Errorf("component %s not found", componentName)
	}
	
	comp, ok := compVal.Value.(*ComponentInstance)
	if !ok {
		return fmt.Errorf("%s is not a component", componentName)
	}
	
	// Set the arrival rate
	if err := comp.SetArrivalRate(methodName, rate); err != nil {
		return fmt.Errorf("failed to set arrival rate: %w", err)
	}
	
	// Trigger FlowEval to compute downstream effects
	return si.RecomputeFlows(componentName, methodName, rate)
}

// RecomputeFlows uses FlowEval to compute downstream traffic flows from a given entry point.
func (si *SystemInstance) RecomputeFlows(component, method string, inputRate float64) error {
	// Create flow context with current system parameters
	parameters := make(map[string]interface{})
	// TODO: Collect actual system parameters from si.Env
	
	context := NewFlowContext(si.System, parameters)
	flows := FlowEval(component, method, inputRate, context)
	
	// Apply computed flows to downstream components
	for target, rate := range flows {
		downstreamComp, downstreamMethod := si.parseFlowTarget(target)
		if downstreamComp != "" && downstreamMethod != "" {
			compVal, ok := si.Env.Get(downstreamComp)
			if ok {
				if comp, ok := compVal.Value.(*ComponentInstance); ok {
					if err := comp.SetArrivalRate(downstreamMethod, rate); err != nil {
						log.Printf("Warning: Failed to set arrival rate for %s.%s: %v", 
							downstreamComp, downstreamMethod, err)
					}
				}
			}
		}
	}
	return nil
}

// RecomputeAllFlows recalculates flows for all entry points in the system.
// Entry points are typically traffic generators or external interfaces.
func (si *SystemInstance) RecomputeAllFlows(entryPoints map[string]float64) error {
	// Safety check
	if si.Env == nil {
		return fmt.Errorf("system instance has no environment")
	}
	
	// Create flow context
	parameters := make(map[string]interface{})
	// TODO: Collect actual system parameters
	
	context := NewFlowContext(si.System, parameters)
	
	// Aggregate all flows
	allFlows := make(map[string]float64)
	
	// Compute flows from each entry point
	for target, rate := range entryPoints {
		component, method := si.parseFlowTarget(target)
		if component != "" && method != "" {
			flows := FlowEval(component, method, rate, context)
			for flowTarget, flowRate := range flows {
				allFlows[flowTarget] += flowRate
			}
			log.Printf("FlowEval: %s.%s @ %.1f RPS -> %v", component, method, rate, flows)
		}
	}
	
	// Apply all computed flows
	for target, rate := range allFlows {
		component, method := si.parseFlowTarget(target)
		if component != "" && method != "" {
			compVal, ok := si.Env.Get(component)
			if ok {
				if comp, ok := compVal.Value.(*ComponentInstance); ok {
					if err := comp.SetArrivalRate(method, rate); err != nil {
						log.Printf("Warning: Failed to set arrival rate for %s.%s: %v", 
							component, method, err)
					} else {
						log.Printf("Applied arrival rate: %s.%s = %.2f RPS", component, method, rate)
					}
				}
			}
		}
	}
	
	return nil
}

// parseFlowTarget splits "component.method" into component and method names.
func (si *SystemInstance) parseFlowTarget(target string) (component, method string) {
	parts := strings.Split(target, ".")
	if len(parts) >= 2 {
		method = parts[len(parts)-1]
		component = strings.Join(parts[:len(parts)-1], ".")
	}
	return component, method
}

// ===== NEW SYSTEM-LEVEL FLOW MANAGEMENT METHODS =====

// AddGenerator adds a new traffic generator to this system
func (si *SystemInstance) AddGenerator(id, name, target string, rate float64) error {
	if si.FlowContext == nil {
		si.FlowContext = NewFlowContext(si.System, make(map[string]interface{}))
	}
	return si.FlowContext.AddGenerator(id, name, target, rate)
}

// RemoveGenerator removes a traffic generator from this system
func (si *SystemInstance) RemoveGenerator(id string) error {
	if si.FlowContext == nil {
		return fmt.Errorf("no flow context available")
	}
	return si.FlowContext.RemoveGenerator(id)
}

// UpdateGenerator modifies an existing generator in this system
func (si *SystemInstance) UpdateGenerator(id string, rate float64, enabled bool) error {
	if si.FlowContext == nil {
		return fmt.Errorf("no flow context available")
	}
	return si.FlowContext.UpdateGenerator(id, rate, enabled)
}

// GetGenerator retrieves a generator by ID from this system
func (si *SystemInstance) GetGenerator(id string) (*GeneratorConfig, bool) {
	if si.FlowContext == nil {
		return nil, false
	}
	return si.FlowContext.GetGenerator(id)
}

// GetActiveGenerators returns all enabled generators in this system
func (si *SystemInstance) GetActiveGenerators() []*GeneratorConfig {
	if si.FlowContext == nil {
		return []*GeneratorConfig{}
	}
	return si.FlowContext.GetActiveGenerators()
}

// GetFlowMetrics returns flow statistics for this system
func (si *SystemInstance) GetFlowMetrics() FlowMetrics {
	if si.FlowContext == nil {
		return FlowMetrics{
			ComponentRPS: make(map[string]float64),
			GeneratorRPS: make(map[string]float64),
		}
	}
	return si.FlowContext.GetFlowMetrics()
}

// SetParameter updates a system parameter and marks flows as needing recalculation
func (si *SystemInstance) SetParameter(path string, value interface{}) error {
	if si.FlowContext == nil {
		si.FlowContext = NewFlowContext(si.System, make(map[string]interface{}))
	}
	
	// Update the parameter in the flow context
	if si.FlowContext.Parameters == nil {
		si.FlowContext.Parameters = make(map[string]interface{})
	}
	si.FlowContext.Parameters[path] = value
	
	// TODO: Apply parameter to actual system environment
	// TODO: Invalidate flows that depend on this parameter
	
	return nil
}

// GetParameter retrieves a system parameter value
func (si *SystemInstance) GetParameter(path string) (interface{}, bool) {
	if si.FlowContext == nil || si.FlowContext.Parameters == nil {
		return nil, false
	}
	
	value, exists := si.FlowContext.Parameters[path]
	return value, exists
}

// AnalyzeFlows performs complete flow analysis for all active generators in this system
func (si *SystemInstance) AnalyzeFlows() error {
	if si.FlowContext == nil {
		return fmt.Errorf("no flow context available")
	}
	
	// Get active generators from the flow context
	activeGenerators := si.FlowContext.GetActiveGenerators()
	if len(activeGenerators) == 0 {
		// No active generators - reset flows and return
		si.FlowContext.Reset()
		return nil
	}
	
	// Convert to GeneratorEntryPoint format for legacy analysis
	var generatorEntryPoints []GeneratorEntryPoint
	for _, gen := range activeGenerators {
		generatorEntryPoints = append(generatorEntryPoints, GeneratorEntryPoint{
			Target:      gen.Target,
			Rate:        gen.Rate,
			GeneratorID: gen.ID,
		})
	}
	
	// Run the existing flow analysis (this populates legacy FlowPaths)
	SolveSystemFlowsWithGenerators(generatorEntryPoints, si.FlowContext)
	
	// TODO: Convert legacy FlowPaths to new FlowPath structs in GeneratorFlows
	// TODO: Populate AggregatedFlows from analysis results
	
	return nil
}

// GetCurrentFlowRates returns the current flow rates for this system (legacy compatibility)
func (si *SystemInstance) GetCurrentFlowRates() map[string]float64 {
	if si.FlowContext == nil {
		return make(map[string]float64)
	}
	return si.FlowContext.ArrivalRates
}

// GetComponentTotalRPS calculates total RPS for a component by summing all its methods
func (si *SystemInstance) GetComponentTotalRPS(componentID string) float64 {
	rates := si.GetCurrentFlowRates()
	total := 0.0
	prefix := componentID + "."

	for target, rps := range rates {
		if strings.HasPrefix(target, prefix) {
			total += rps
		}
	}

	return total
}

// RegisterNativeComponents registers native component instances in the FlowContext
// This should be called after system initialization to enable flow analysis
func (si *SystemInstance) RegisterNativeComponents() {
	if si.FlowContext == nil {
		return
	}

	// Import the components package for FlowAnalyzable interface
	// Iterate through system instances and register native components
	for _, item := range si.System.Body {
		if instDecl, ok := item.(*InstanceDecl); ok {
			componentName := instDecl.ComponentName.Value
			instanceName := instDecl.Name.Value

			// Check if this is a known native component type and create an instance
			var nativeComponent interface{} // FlowAnalyzable interface would be imported here
			switch componentName {
			case "ResourcePool":
				// Native component creation would be implemented here
				// For now, we'll create a placeholder that can be extended
				log.Printf("SystemInstance: Would register ResourcePool '%s' for flow analysis", instanceName)
				
			case "HashIndex":
				log.Printf("SystemInstance: Would register HashIndex '%s' for flow analysis", instanceName)
				
			case "Cache":
				log.Printf("SystemInstance: Would register Cache '%s' for flow analysis", instanceName)
			}

			// Register the native component in the FlowContext when implemented
			if nativeComponent != nil {
				// si.FlowContext.SetNativeComponent(instanceName, nativeComponent)
				log.Printf("SystemInstance: Registered native component '%s' (type %s) for flow analysis", instanceName, componentName)
			}
		}
	}
}
