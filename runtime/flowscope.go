package runtime

import "strings"

// FlowScope manages the runtime context for flow evaluation
type FlowScope struct {
	Outer            *FlowScope
	Evaluator        *FlowEvaluator
	SysEnv           *Env[Value]
	CurrentComponent *ComponentInstance
	CurrentMethod    *MethodDecl
	ArrivalRates     RateMap
	SuccessRates     RateMap
	CallStack        []*ComponentInstance
	
	// Variable outcome tracking for conditional flow analysis
	VariableOutcomes map[string]float64
}

// NewFlowScope creates a new root flow scope
func NewFlowScope(evaluator *FlowEvaluator, sysEnv *Env[Value]) *FlowScope {
	return &FlowScope{
		Outer:            nil,
		Evaluator:        evaluator,
		SysEnv:           sysEnv,
		ArrivalRates:     NewRateMap(),
		SuccessRates:     NewRateMap(),
		CallStack:        make([]*ComponentInstance, 0),
		VariableOutcomes: make(map[string]float64),
	}
}

// Push creates a new nested scope for entering a component method
func (fs *FlowScope) Push(component *ComponentInstance, method *MethodDecl) *FlowScope {
	return &FlowScope{
		Outer:            fs,
		Evaluator:        fs.Evaluator,
		SysEnv:           fs.SysEnv.Push(), // Push new env scope
		CurrentComponent: component,
		CurrentMethod:    method,
		ArrivalRates:     fs.ArrivalRates, // Share rate maps with parent
		SuccessRates:     fs.SuccessRates,
		CallStack:        append(fs.CallStack, component),
		VariableOutcomes: make(map[string]float64), // Fresh variable tracking per method
	}
}

// Pop returns the parent scope
func (fs *FlowScope) Pop() *FlowScope {
	if fs.Outer != nil {
		return fs.Outer
	}
	return fs
}

// IsInCallStack checks if a component is already in the call stack (cycle detection)
func (fs *FlowScope) IsInCallStack(component *ComponentInstance) bool {
	for _, comp := range fs.CallStack {
		if comp == component {
			return true
		}
	}
	return false
}

// ResolveTarget resolves a target string like "db.Query" to component instance and method
// This will use the current environment to look up component instances
func (fs *FlowScope) ResolveTarget(target string) (*ComponentInstance, string) {
	// Split target into component and method parts
	parts := strings.Split(target, ".")
	if len(parts) < 2 {
		return nil, ""
	}
	
	// Last part is the method
	method := parts[len(parts)-1]
	
	// Everything before is the component path
	componentPath := strings.Join(parts[:len(parts)-1], ".")
	
	// Try to find the component in the environment
	if value, exists := fs.SysEnv.Get(componentPath); exists {
		if value.Value != nil {
			if comp, ok := value.Value.(*ComponentInstance); ok {
				return comp, method
			}
		}
	}
	
	// If not found directly, try just the first part (common case)
	if len(parts) > 2 {
		if value, exists := fs.SysEnv.Get(parts[0]); exists {
			if value.Value != nil {
				if comp, ok := value.Value.(*ComponentInstance); ok {
					// For now, return the component with the full remaining path as method
					// This is a simplification - in a real system we'd traverse nested components
					return comp, strings.Join(parts[1:], ".")
				}
			}
		}
	}
	
	return nil, ""
}

// GetVariable retrieves a variable from the current or outer scopes
func (fs *FlowScope) GetVariable(name string) (Value, bool) {
	return fs.SysEnv.Get(name)
}

// SetVariable sets a variable in the current scope
func (fs *FlowScope) SetVariable(name string, value Value) {
	fs.SysEnv.Set(name, value)
}

// TrackVariableOutcome records the success rate of a variable (for conditional evaluation)
func (fs *FlowScope) TrackVariableOutcome(varName string, successRate float64) {
	fs.VariableOutcomes[varName] = successRate
}

// GetVariableOutcome retrieves the tracked success rate for a variable
func (fs *FlowScope) GetVariableOutcome(varName string) (float64, bool) {
	rate, exists := fs.VariableOutcomes[varName]
	return rate, exists
}