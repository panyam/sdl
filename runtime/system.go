package runtime

import (
	"fmt"
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
}

// Initializes a new runtime System instance and its root environment
func NewSystemInstance(file *FileInstance, system *SystemDecl) *SystemInstance {
	sysinst := &SystemInstance{File: file, System: system}
	return sysinst
}

// GetSystemName returns the name of the system
func (s *SystemInstance) GetSystemName() string {
	if s.System != nil && s.System.Name != nil {
		return s.System.Name.Value
	}
	return ""
}

// Finds a nested component a.b.c starting at the root of a system
func (s *SystemInstance) FindComponent(fqn string) (out *ComponentInstance) {
	parts := strings.Split(fqn, ".")

	// Start from the system's environment
	currentEnv := s.Env
	var currentComponent *ComponentInstance

	// Navigate through nested components
	for i, part := range parts {
		value, ok := currentEnv.Get(part)
		if !ok {
			return nil
		}

		comp, ok := value.Value.(*ComponentInstance)
		if !ok {
			return nil
		}

		currentComponent = comp

		// If this is not the last part, move to the component's environment
		if i < len(parts)-1 {
			currentEnv = comp.InitialEnv
		}
	}

	return currentComponent
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
			Error("Invalid type: %v %v", it, reflect.TypeOf(it))
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
						Warn("Failed to set arrival rate for %s.%s: %v",
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
			Debug("FlowEval: %s.%s @ %.1f RPS -> %v", component, method, rate, flows)
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
						Warn("Failed to set arrival rate for %s.%s: %v",
							component, method, err)
					} else {
						Debug("Applied arrival rate: %s.%s = %.2f RPS", component, method, rate)
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

// AllComponents returns all component instances in the system
func (si *SystemInstance) AllComponents() []*ComponentInstance {
	if si.Env == nil {
		return nil
	}
	
	var components []*ComponentInstance
	
	// Iterate through the system environment to find all components
	bindings := si.Env.All()
	for _, binding := range bindings {
		if comp, ok := binding.Value.(*ComponentInstance); ok && comp != nil {
			components = append(components, comp)
		}
	}
	
	return components
}
