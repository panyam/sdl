package runtime

import (
	"fmt"
	"strings"

	"github.com/panyam/sdl/lib/decl"
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

// FindComponent resolves a dotted path like "a.b.c" to a nested ComponentInstance.
//
// Resolution order:
//  1. Try the path directly in the system environment (exact match).
//  2. If not found, try prepending each system parameter name. This provides
//     implicit scoping so that "webserver.RequestRide" resolves to
//     "arch.webserver.RequestRide" when the system has a single parameter "arch".
//     For multi-parameter systems, each parameter is tried in declaration order.
func (s *SystemInstance) FindComponent(fqn string) (out *ComponentInstance) {
	// Try exact path first
	if comp := s.findComponentExact(fqn); comp != nil {
		return comp
	}

	// Implicit parameter scoping: try prepending each system parameter name
	if s.System != nil {
		for _, param := range s.System.Parameters {
			prefixed := param.Name.Value + "." + fqn
			if comp := s.findComponentExact(prefixed); comp != nil {
				return comp
			}
		}
	}

	return nil
}

// findComponentExact resolves a dotted path without implicit scoping.
func (s *SystemInstance) findComponentExact(fqn string) (out *ComponentInstance) {
	parts := strings.Split(fqn, ".")

	currentEnv := s.Env
	var currentComponent *ComponentInstance

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

		if i < len(parts)-1 {
			currentEnv = comp.Env
		}
	}

	return currentComponent
}

// Initializer compiles the system into initialization statements.
//
// For parameterized systems (system Name(p1 Type1, p2 Type2) { ... }), each
// parameter creates a component instance of the declared type. These are the
// top-level entry points for the simulation.
//
// For legacy systems (system Name { ... }) with no parameters, the body may
// contain LetStmt declarations. InstanceDecl ('use') is no longer supported.
func (s *SystemInstance) Initializer() (blockStmt *BlockStmt, err error) {
	var stmts []Stmt

	// Create component instances for each system parameter
	for _, param := range s.System.Parameters {
		compDecl, err := s.File.GetComponentDecl(param.TypeDecl.Name)
		ensureNoErr(err)
		stmts = append(stmts, &decl.SetStmt{
			TargetExpr: param.Name,
			Value:      NewNewExpr(compDecl),
		})
	}

	// Process body items (LetStmt, OptionsDecl — no more InstanceDecl)
	for _, item := range s.System.Body {
		switch it := item.(type) {
		case *LetStmt:
			stmts = append(stmts, it)
		default:
			Error("Invalid system body item type: %T", it)
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

// GetUninitializedComponents walks system parameters and their dependency
// trees to find any uninitialized components. Called after initialization
// to detect missing wiring.
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

	// Walk system parameters instead of Body InstanceDecls
	for _, param := range s.System.Parameters {
		compValue, ok := env.Get(param.Name.Value)
		if !ok {
			items = append(items, &InitStmt{
				Pos:    param.Pos(),
				Attrib: param.Name.Value,
			})
			continue
		}

		compInst := compValue.Value.(*ComponentInstance)
		visit(&InitStmt{
			Pos:      param.Pos(),
			Attrib:   param.Name.Value,
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
