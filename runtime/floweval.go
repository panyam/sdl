package runtime

import (
	"fmt"
	"log"
	"math"
	"strings"
	
	"github.com/panyam/sdl/components"
)

// FlowContext holds the system state and configuration for flow evaluation
type FlowContext struct {
	System     *SystemDecl                  // SDL system definition
	Parameters map[string]interface{}       // Current parameter values (hitRate, poolSize, etc.)
	
	// Back-pressure and convergence tracking
	ArrivalRates   map[string]float64        // Current arrival rates per component.method
	SuccessRates   map[string]float64        // Current success rates per component.method
	ServiceTimes   map[string]float64        // Service times per component.method (seconds)
	ResourceLimits map[string]int            // Pool sizes, capacities per component
	
	// Native component instances for FlowAnalyzable interface
	NativeComponents map[string]components.FlowAnalyzable // component name -> FlowAnalyzable instance
	
	// Cycle handling configuration
	MaxRetries           int     // Limit exponential growth (recommended: 50)
	ConvergenceThreshold float64 // Fixed-point iteration threshold (recommended: 0.01)
	MaxIterations        int     // Maximum fixed-point iterations (recommended: 10)
	CallStack            []string // Detect infinite recursion
}

// NewFlowContext creates a new FlowContext with sensible defaults
func NewFlowContext(system *SystemDecl, parameters map[string]interface{}) *FlowContext {
	return &FlowContext{
		System:               system,
		Parameters:           parameters,
		ArrivalRates:         make(map[string]float64),
		SuccessRates:         make(map[string]float64),
		ServiceTimes:         make(map[string]float64),
		ResourceLimits:       make(map[string]int),
		NativeComponents:     make(map[string]components.FlowAnalyzable),
		MaxRetries:           50,
		ConvergenceThreshold: 0.01,
		MaxIterations:        10,
		CallStack:            make([]string, 0),
	}
}

// SolveSystemFlows performs iterative fixed-point computation for system-wide flows with back-pressure
func SolveSystemFlows(entryPoints map[string]float64, context *FlowContext) map[string]float64 {
	if context == nil {
		log.Printf("SolveSystemFlows: context is nil")
		return make(map[string]float64)
	}

	// Initialize arrival rates with entry points
	for componentMethod, rate := range entryPoints {
		context.ArrivalRates[componentMethod] = rate
	}

	log.Printf("SolveSystemFlows: Starting fixed-point iteration with %d entry points", len(entryPoints))

	// Iterate until convergence
	for iteration := 0; iteration < context.MaxIterations; iteration++ {
		// Save old rates for convergence check
		oldRates := make(map[string]float64)
		for k, v := range context.ArrivalRates {
			oldRates[k] = v
		}

		// Recompute outgoing flows for each component with current load
		newRates := make(map[string]float64)
		
		// Start with entry points
		for componentMethod, rate := range entryPoints {
			newRates[componentMethod] = rate
		}

		// Propagate flows through the system
		for componentMethod, inRate := range context.ArrivalRates {
			if inRate > 1e-9 { // Only process non-trivial rates
				component, method := parseComponentMethod(componentMethod)
				if component != "" && method != "" {
					outflows := FlowEval(component, method, inRate, context)
					for target, outRate := range outflows {
						newRates[target] += outRate
					}
				}
			}
		}

		// Check convergence (all rates changed by < threshold)
		maxChange := 0.0
		for componentMethod := range newRates {
			oldRate := oldRates[componentMethod]
			newRate := newRates[componentMethod]
			change := math.Abs(newRate - oldRate)
			if change > maxChange {
				maxChange = change
			}
		}

		log.Printf("SolveSystemFlows: Iteration %d, max change: %.6f", iteration, maxChange)

		if maxChange < context.ConvergenceThreshold {
			log.Printf("SolveSystemFlows: Converged after %d iterations", iteration)
			// Update final arrival rates and success rates
			for k, v := range newRates {
				context.ArrivalRates[k] = v
			}
			context.updateSuccessRates()
			return newRates
		}

		// Apply damping to prevent oscillation (0.5 damping factor)
		for componentMethod := range newRates {
			oldRate := oldRates[componentMethod]
			newRate := newRates[componentMethod]
			context.ArrivalRates[componentMethod] = oldRate + 0.5*(newRate-oldRate)
		}
	}

	log.Printf("SolveSystemFlows: Did not converge after %d iterations", context.MaxIterations)
	// Update success rates based on final arrival rates
	context.updateSuccessRates()
	return context.ArrivalRates
}

// FlowEval computes outbound traffic rates from a given component.method at inputRate
// Returns map of {downstreamComponent.method: outputRate}
func FlowEval(component, method string, inputRate float64, context *FlowContext) map[string]float64 {
	if context == nil {
		log.Printf("FlowEval: context is nil, returning empty flows")
		return map[string]float64{}
	}

	callKey := fmt.Sprintf("%s.%s", component, method)
	
	// Detect cycles using call stack
	if context.isInCallStack(callKey) {
		log.Printf("FlowEval: Cycle detected for %s, breaking recursion", callKey)
		return map[string]float64{}
	}

	// Prevent infinite recursion
	if len(context.CallStack) >= 20 {
		log.Printf("FlowEval: Maximum call depth reached, stopping recursion")
		return map[string]float64{}
	}

	// Add to call stack
	context.CallStack = append(context.CallStack, callKey)
	defer func() { 
		context.CallStack = context.CallStack[:len(context.CallStack)-1] 
	}()

	// Check if this is a native component first
	if flowAnalyzable, exists := context.NativeComponents[component]; exists {
		pattern := flowAnalyzable.GetFlowPattern(method, inputRate, context.getComponentParams(component))
		return pattern.Outflows
	}

	// Get method definition from system
	methodDecl := context.getMethodDecl(component, method)
	if methodDecl == nil {
		log.Printf("FlowEval: Method %s.%s not found", component, method)
		return map[string]float64{}
	}

	// Analyze method body for call patterns
	outflows := make(map[string]float64)
	context.analyzeStatements(methodDecl.Body.Statements, inputRate, outflows)

	return outflows
}

// isInCallStack checks if a call key is already in the call stack
func (fc *FlowContext) isInCallStack(callKey string) bool {
	for _, existing := range fc.CallStack {
		if existing == callKey {
			return true
		}
	}
	return false
}

// getMethodDecl finds a method declaration in the system
func (fc *FlowContext) getMethodDecl(component, method string) *MethodDecl {
	if fc.System == nil {
		return nil
	}

	// Find the component in the system body
	for _, bodyItem := range fc.System.Body {
		if instance, ok := bodyItem.(*InstanceDecl); ok {
			if instance.Name.Value == component {
				// Get the component type declaration
				if compDecl := fc.getComponentDecl(instance.ComponentName.Value); compDecl != nil {
					// Find the method in the component
					methods, err := compDecl.Methods()
					if err == nil {
						for _, methodDecl := range methods {
							if methodDecl.Name.Value == method {
								return methodDecl
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// getComponentDecl finds a component declaration (this is a placeholder - needs proper loader integration)
func (fc *FlowContext) getComponentDecl(componentType string) *ComponentDecl {
	// TODO: This needs to be integrated with the loader to get the actual component declaration
	// For now, return nil - we'll need to pass component declarations through the context
	log.Printf("FlowEval: getComponentDecl not yet implemented for %s", componentType)
	return nil
}

// analyzeStatements processes a list of statements and accumulates outflows
func (fc *FlowContext) analyzeStatements(statements []Stmt, inputRate float64, outflows map[string]float64) {
	for _, stmt := range statements {
		fc.analyzeStatement(stmt, inputRate, outflows)
	}
}

// analyzeStatement processes a single statement and updates outflows
func (fc *FlowContext) analyzeStatement(stmt Stmt, inputRate float64, outflows map[string]float64) {
	switch s := stmt.(type) {
	case *ExprStmt:
		fc.analyzeExprStatement(s, inputRate, outflows)
	case *IfStmt:
		fc.analyzeIfStatement(s, inputRate, outflows)
	case *AssignmentStmt:
		fc.analyzeAssignStatement(s, inputRate, outflows)
	case *ReturnStmt:
		fc.analyzeReturnStatement(s, inputRate, outflows)
	case *ForStmt:
		fc.analyzeForStatement(s, inputRate, outflows)
	default:
		// Other statement types don't generate flows
		log.Printf("FlowEval: Unhandled statement type: %T", stmt)
	}
}

// analyzeExprStatement processes expression statements that might contain calls
func (fc *FlowContext) analyzeExprStatement(stmt *ExprStmt, inputRate float64, outflows map[string]float64) {
	// Check if the expression is a call
	if callExpr, ok := stmt.Expression.(*CallExpr); ok {
		target := fc.extractCallTarget(callExpr)
		if target != "" {
			component, method := fc.parseCallTarget(target)
			if component != "" && method != "" {
				flowKey := fmt.Sprintf("%s.%s", component, method)
				outflows[flowKey] += inputRate
				
				log.Printf("FlowEval: Expression call flow: %s -> %s (%.2f RPS)", 
					strings.Join(fc.CallStack, " -> "), flowKey, inputRate)
			}
		}
	}
}

// analyzeIfStatement handles conditional flows
func (fc *FlowContext) analyzeIfStatement(stmt *IfStmt, inputRate float64, outflows map[string]float64) {
	// Determine condition probability from parameters or heuristics
	conditionProb := fc.evaluateConditionProbability(stmt.Condition)
	
	// Analyze then branch with probability-weighted rate
	if stmt.Then != nil {
		fc.analyzeStatement(stmt.Then, inputRate*conditionProb, outflows)
	}
	
	// Analyze else branch with inverted probability
	if stmt.Else != nil {
		fc.analyzeStatement(stmt.Else, inputRate*(1.0-conditionProb), outflows)
	}
}

// analyzeAssignStatement handles assignments that might contain calls
func (fc *FlowContext) analyzeAssignStatement(stmt *AssignmentStmt, inputRate float64, outflows map[string]float64) {
	// Check if the assigned expression contains a call
	if callExpr, ok := stmt.Value.(*CallExpr); ok {
		target := fc.extractCallTarget(callExpr)
		if target != "" {
			component, method := fc.parseCallTarget(target)
			if component != "" && method != "" {
				flowKey := fmt.Sprintf("%s.%s", component, method)
				outflows[flowKey] += inputRate
				
				log.Printf("FlowEval: Assignment call flow: %s -> %s (%.2f RPS)", 
					strings.Join(fc.CallStack, " -> "), flowKey, inputRate)
			}
		}
	}
}

// analyzeReturnStatement handles return statements that might contain calls
func (fc *FlowContext) analyzeReturnStatement(stmt *ReturnStmt, inputRate float64, outflows map[string]float64) {
	if stmt.ReturnValue != nil {
		if callExpr, ok := stmt.ReturnValue.(*CallExpr); ok {
			target := fc.extractCallTarget(callExpr)
			if target != "" {
				component, method := fc.parseCallTarget(target)
				if component != "" && method != "" {
					flowKey := fmt.Sprintf("%s.%s", component, method)
					outflows[flowKey] += inputRate
					
					log.Printf("FlowEval: Return call flow: %s -> %s (%.2f RPS)", 
						strings.Join(fc.CallStack, " -> "), flowKey, inputRate)
				}
			}
		}
	}
}

// analyzeForStatement handles loops (basic implementation)
func (fc *FlowContext) analyzeForStatement(stmt *ForStmt, inputRate float64, outflows map[string]float64) {
	// For now, assume the loop body executes once per input
	// TODO: Implement proper loop analysis based on loop bounds
	if stmt.Body != nil {
		fc.analyzeStatement(stmt.Body, inputRate, outflows)
	}
}

// extractCallTarget extracts the target string from a call expression
func (fc *FlowContext) extractCallTarget(call *CallExpr) string {
	if call == nil || call.Function == nil {
		return ""
	}

	// Handle member access like self.db.LookupByPhone
	if memberExpr, ok := call.Function.(*MemberAccessExpr); ok {
		return fc.memberExpressionToString(memberExpr)
	}

	// Handle simple function calls
	if ident, ok := call.Function.(*IdentifierExpr); ok {
		return ident.Value
	}

	return ""
}

// memberExpressionToString converts a member expression to a string
func (fc *FlowContext) memberExpressionToString(expr *MemberAccessExpr) string {
	var parts []string
	
	// Recursively build the member access chain
	current := expr
	for current != nil {
		if current.Member != nil {
			parts = append([]string{current.Member.Value}, parts...)
		}
		
		if memberExpr, ok := current.Receiver.(*MemberAccessExpr); ok {
			current = memberExpr
		} else if ident, ok := current.Receiver.(*IdentifierExpr); ok {
			if ident.Value != "self" { // Skip "self" prefix
				parts = append([]string{ident.Value}, parts...)
			}
			break
		} else {
			break
		}
	}
	
	return strings.Join(parts, ".")
}

// parseCallTarget parses "db.LookupByPhone" into component "db" and method "LookupByPhone"
func (fc *FlowContext) parseCallTarget(target string) (component, method string) {
	parts := strings.Split(target, ".")
	if len(parts) >= 2 {
		// Last part is the method, everything before is the component path
		method = parts[len(parts)-1]
		component = strings.Join(parts[:len(parts)-1], ".")
		return component, method
	}
	return "", ""
}

// evaluateConditionProbability determines the probability of a condition being true
func (fc *FlowContext) evaluateConditionProbability(condition Expr) float64 {
	// For now, use simple heuristics based on parameter values
	// TODO: Implement sophisticated condition analysis
	
	// Try to extract cache hit rate from parameters
	if fc.Parameters != nil {
		if hitRate, ok := fc.Parameters["CacheHitRate"]; ok {
			if rate, ok := hitRate.(float64); ok {
				return rate
			}
		}
	}
	
	// Default to 50% probability for unknown conditions
	return 0.5
}

// parseComponentMethod parses "component.method" into separate parts
func parseComponentMethod(componentMethod string) (component, method string) {
	parts := strings.Split(componentMethod, ".")
	if len(parts) >= 2 {
		// Last part is the method, everything before is the component path
		method = parts[len(parts)-1]
		component = strings.Join(parts[:len(parts)-1], ".")
		return component, method
	}
	return "", ""
}

// updateSuccessRates updates success rates based on current arrival rates (back-pressure modeling)
func (fc *FlowContext) updateSuccessRates() {
	for componentMethod, arrivalRate := range fc.ArrivalRates {
		component, method := parseComponentMethod(componentMethod)
		if component == "" || method == "" {
			continue
		}

		// Check if this is a native component with FlowAnalyzable interface
		if flowAnalyzable, exists := fc.NativeComponents[component]; exists {
			pattern := flowAnalyzable.GetFlowPattern(method, arrivalRate, fc.getComponentParams(component))
			fc.SuccessRates[componentMethod] = pattern.SuccessRate
			fc.ServiceTimes[componentMethod] = pattern.ServiceTime
			continue
		}

		// Default success rate modeling for SDL components
		// This is a placeholder - could be enhanced with specific component analysis
		successRate := 1.0
		
		// Simple back-pressure model: success rate degrades with high arrival rates
		if limit, hasLimit := fc.ResourceLimits[component]; hasLimit && limit > 0 {
			utilization := arrivalRate * fc.getServiceTime(componentMethod) / float64(limit)
			if utilization > 0.8 {
				// Success rate degrades as utilization approaches 1.0
				successRate = math.Max(0.1, 1.0-utilization*0.5)
			}
		}
		
		fc.SuccessRates[componentMethod] = successRate
	}
}

// getComponentParams extracts parameters for a specific component
func (fc *FlowContext) getComponentParams(component string) map[string]interface{} {
	params := make(map[string]interface{})
	if fc.Parameters != nil {
		// Look for component-specific parameters with prefix "component."
		prefix := component + "."
		for key, value := range fc.Parameters {
			if strings.HasPrefix(key, prefix) {
				paramName := strings.TrimPrefix(key, prefix)
				params[paramName] = value
			}
		}
	}
	return params
}

// getServiceTime gets the service time for a component.method
func (fc *FlowContext) getServiceTime(componentMethod string) float64 {
	if serviceTime, exists := fc.ServiceTimes[componentMethod]; exists {
		return serviceTime
	}
	// Default service time
	return 0.001 // 1ms default
}

// SetNativeComponent registers a native component for flow analysis
func (fc *FlowContext) SetNativeComponent(name string, component components.FlowAnalyzable) {
	fc.NativeComponents[name] = component
}

// SetResourceLimit sets the resource limit for a component
func (fc *FlowContext) SetResourceLimit(component string, limit int) {
	fc.ResourceLimits[component] = limit
}