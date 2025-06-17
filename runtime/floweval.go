package runtime

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/panyam/sdl/components"
)

// FlowPathInfo tracks information about a flow path for visualization
type FlowPathInfo struct {
	Order       float64 // Execution order (supports decimals for conditional paths)
	Condition   string  // Condition expression if this is a conditional path
	Probability float64 // Probability of this path being taken
	GeneratorID string  // ID of the generator that originated this flow
	Color       string  // Color for visualization (based on generator)
}

// FlowContext holds the system state and configuration for flow evaluation
type FlowContext struct {
	System     *SystemDecl            // SDL system definition
	Parameters map[string]interface{} // Current parameter values (hitRate, poolSize, etc.)

	// Back-pressure and convergence tracking
	ArrivalRates   map[string]float64 // Current arrival rates per component.method
	SuccessRates   map[string]float64 // Current success rates per component.method
	ServiceTimes   map[string]float64 // Service times per component.method (seconds)
	ResourceLimits map[string]int     // Pool sizes, capacities per component

	// Native component instances for FlowAnalyzable interface
	NativeComponents map[string]components.FlowAnalyzable // component name -> FlowAnalyzable instance

	// Cycle handling configuration
	MaxRetries           int      // Limit exponential growth (recommended: 50)
	ConvergenceThreshold float64  // Fixed-point iteration threshold (recommended: 0.01)
	MaxIterations        int      // Maximum fixed-point iterations (recommended: 10)
	CallStack            []string // Detect infinite recursion

	// Current calling component context for dependency resolution
	CurrentComponent string // Component currently being analyzed

	// Variable outcome tracking for conditional flow analysis
	VariableOutcomes map[string]float64 // variable name -> success rate (for method-local analysis)

	// Flow order tracking for visualization
	FlowOrder          int                     // Current flow order counter
	FlowPaths          map[string]FlowPathInfo // flowKey -> path info for visualization
	currentCondition   string                  // Current condition context
	currentProbability float64                 // Current condition probability
	
	// Generator-specific flow tracking
	CurrentGeneratorID string             // ID of the generator currently being analyzed
	GeneratorFlowOrder map[string]int     // Per-generator flow order counters
	GeneratorColors    map[string]string  // Generator ID -> color mapping
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
		VariableOutcomes:     make(map[string]float64),
		FlowOrder:            0,
		FlowPaths:            make(map[string]FlowPathInfo),
		GeneratorFlowOrder:   make(map[string]int),
		GeneratorColors:      make(map[string]string),
	}
}

// GeneratorEntryPoint represents an entry point with generator information
type GeneratorEntryPoint struct {
	Target      string  // component.method target
	Rate        float64 // requests per second
	GeneratorID string  // generator identifier
}

// SolveSystemFlowsWithGenerators performs flow analysis with per-generator tracking
func SolveSystemFlowsWithGenerators(generators []GeneratorEntryPoint, context *FlowContext) map[string]float64 {
	if context == nil {
		log.Printf("SolveSystemFlowsWithGenerators: context is nil")
		return make(map[string]float64)
	}

	// Reset flow tracking for this analysis
	context.FlowOrder = 0
	context.FlowPaths = make(map[string]FlowPathInfo)
	context.GeneratorFlowOrder = make(map[string]int)
	
	// Set up generator colors (cycling through a palette)
	colors := []string{"#3b82f6", "#ef4444", "#10b981", "#f59e0b", "#8b5cf6", "#06b6d4", "#84cc16", "#f97316"}
	for i, gen := range generators {
		context.GeneratorColors[gen.GeneratorID] = colors[i%len(colors)]
		context.GeneratorFlowOrder[gen.GeneratorID] = 0
	}

	// Build combined entry points map for convergence analysis
	entryPoints := make(map[string]float64)
	for _, gen := range generators {
		entryPoints[gen.Target] = gen.Rate
	}

	// Initialize arrival rates with entry points
	for componentMethod, rate := range entryPoints {
		context.ArrivalRates[componentMethod] = rate
	}

	log.Printf("SolveSystemFlowsWithGenerators: Starting with %d generators", len(generators))

	// Analyze flows for each generator separately for visualization
	for _, gen := range generators {
		context.CurrentGeneratorID = gen.GeneratorID
		log.Printf("SolveSystemFlowsWithGenerators: Analyzing flows for generator %s", gen.GeneratorID)
		
		// Reset arrival rates and analyze this generator in isolation
		// to capture all its nested calls with proper generator context
		isolatedRates := make(map[string]float64)
		isolatedRates[gen.Target] = gen.Rate
		
		// Temporarily save and clear arrival rates
		savedRates := context.ArrivalRates
		context.ArrivalRates = isolatedRates
		
		// Run a few iterations to capture nested calls for this generator
		for iteration := 0; iteration < 3; iteration++ {
			newRates := make(map[string]float64)
			newRates[gen.Target] = gen.Rate
			
			for componentMethod, inRate := range context.ArrivalRates {
				if inRate > 1e-9 {
					component, method := parseComponentMethod(componentMethod)
					if component != "" && method != "" {
						outflows := FlowEval(component, method, inRate, context)
						for target, outRate := range outflows {
							newRates[target] += outRate
						}
					}
				}
			}
			
			context.ArrivalRates = newRates
		}
		
		// Restore arrival rates
		context.ArrivalRates = savedRates
	}

	// Clear the current generator context for convergence analysis
	context.CurrentGeneratorID = ""
	
	// Run convergence analysis for back-pressure modeling
	// The FlowPaths are already populated with per-generator information
	// so we just need the convergence without resetting flow paths
	return SolveSystemFlowsPreservingPaths(entryPoints, context)
}

// SolveSystemFlowsPreservingPaths performs convergence analysis while preserving existing flow paths
func SolveSystemFlowsPreservingPaths(entryPoints map[string]float64, context *FlowContext) map[string]float64 {
	if context == nil {
		log.Printf("SolveSystemFlowsPreservingPaths: context is nil")
		return make(map[string]float64)
	}

	// Initialize arrival rates with entry points (don't reset flow paths)
	for componentMethod, rate := range entryPoints {
		context.ArrivalRates[componentMethod] = rate
	}

	log.Printf("SolveSystemFlowsPreservingPaths: Starting fixed-point iteration with %d entry points", len(entryPoints))

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

		log.Printf("SolveSystemFlowsPreservingPaths: Iteration %d, max change: %.6f", iteration, maxChange)

		if maxChange < context.ConvergenceThreshold {
			log.Printf("SolveSystemFlowsPreservingPaths: Converged after %d iterations", iteration)
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

	log.Printf("SolveSystemFlowsPreservingPaths: Did not converge after %d iterations", context.MaxIterations)
	context.updateSuccessRates()
	return context.ArrivalRates
}

// SolveSystemFlows performs iterative fixed-point computation for system-wide flows with back-pressure
func SolveSystemFlows(entryPoints map[string]float64, context *FlowContext) map[string]float64 {
	if context == nil {
		log.Printf("SolveSystemFlows: context is nil")
		return make(map[string]float64)
	}

	// Reset flow tracking for this analysis
	context.FlowOrder = 0
	context.FlowPaths = make(map[string]FlowPathInfo)

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

	// Add to call stack and set current component context
	context.CallStack = append(context.CallStack, callKey)
	oldComponent := context.CurrentComponent
	context.CurrentComponent = component
	defer func() {
		context.CallStack = context.CallStack[:len(context.CallStack)-1]
		context.CurrentComponent = oldComponent
	}()

	// Clear variable outcomes for this method analysis
	context.VariableOutcomes = make(map[string]float64)

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
	if methodDecl.Body != nil && methodDecl.Body.Statements != nil {
		context.analyzeStatements(methodDecl.Body.Statements, inputRate, outflows)
	} else {
		// Method has no body (likely imported/native) - no outflows to propagate
		log.Printf("FlowEval: Method %s.%s has no body, no outflows to analyze", component, method)
	}

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

// getMethodDecl finds a method declaration in the system using pre-resolved component declarations
func (fc *FlowContext) getMethodDecl(component, method string) *MethodDecl {
	if fc.System == nil {
		return nil
	}

	// Find the component in the system body
	for _, bodyItem := range fc.System.Body {
		if instance, ok := bodyItem.(*InstanceDecl); ok {
			if instance.Name.Value == component {
				// Use the pre-resolved component declaration from type inference
				if instance.ResolvedComponentDecl != nil {
					methods, err := instance.ResolvedComponentDecl.Methods()
					if err == nil {
						for _, methodDecl := range methods {
							if methodDecl.Name.Value == method {
								return methodDecl
							}
						}
					}
				} else {
					log.Printf("FlowEval: ResolvedComponentDecl not set for instance %s (component type %s) - was type inference run?",
						instance.Name.Value, instance.ComponentName.Value)
				}
			}
		}
	}

	return nil
}

// analyzeStatements processes a list of statements and accumulates outflows
func (fc *FlowContext) analyzeStatements(statements []Stmt, inputRate float64, outflows map[string]float64) {
	for i, stmt := range statements {
		log.Printf("FlowEval: Analyzing statement %d of type %T", i, stmt)
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
	case *BlockStmt:
		fc.analyzeBlockStatement(s, inputRate, outflows)
	case *LetStmt:
		fc.analyzeLetStatement(s, inputRate, outflows)
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

				// Track flow path with current conditional context
				fc.trackFlowPath(flowKey, fc.currentCondition, fc.currentProbability)
			}
		}
	}
}

// analyzeIfStatement handles conditional flows
func (fc *FlowContext) analyzeIfStatement(stmt *IfStmt, inputRate float64, outflows map[string]float64) {
	// Determine condition probability from parameters or heuristics
	conditionProb := fc.evaluateConditionProbability(stmt.Condition)
	conditionStr := fc.conditionToString(stmt.Condition)

	log.Printf("FlowEval: Processing if statement with condition probability %.2f", conditionProb)

	// Save current order level for conditional paths
	savedOrder := fc.FlowOrder

	// Analyze then branch with probability-weighted rate
	if stmt.Then != nil {
		// Track this as a conditional path
		fc.analyzeStatementWithCondition(stmt.Then, inputRate*conditionProb, outflows, conditionStr, conditionProb)
	}

	// Restore order for else branch
	fc.FlowOrder = savedOrder

	// Analyze else branch with inverted probability
	if stmt.Else != nil {
		// Track this as the else conditional path
		elseCondition := fmt.Sprintf("!(%s)", conditionStr)
		fc.analyzeStatementWithCondition(stmt.Else, inputRate*(1.0-conditionProb), outflows, elseCondition, 1.0-conditionProb)
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
					
					// Track flow path with current conditional context
					fc.trackFlowPath(flowKey, fc.currentCondition, fc.currentProbability)
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

// analyzeBlockStatement handles block statements by analyzing all contained statements
func (fc *FlowContext) analyzeBlockStatement(stmt *BlockStmt, inputRate float64, outflows map[string]float64) {
	if stmt.Statements != nil {
		fc.analyzeStatements(stmt.Statements, inputRate, outflows)
	}
}

// analyzeLetStatement handles let statements that might contain calls in their assigned expressions
func (fc *FlowContext) analyzeLetStatement(stmt *LetStmt, inputRate float64, outflows map[string]float64) {
	// Check if the assigned expression contains a call
	if stmt.Value != nil {
		if callExpr, ok := stmt.Value.(*CallExpr); ok {
			target := fc.extractCallTarget(callExpr)
			if target != "" {
				component, method := fc.parseCallTarget(target)
				if component != "" && method != "" {
					flowKey := fmt.Sprintf("%s.%s", component, method)
					outflows[flowKey] += inputRate

					log.Printf("FlowEval: Let statement call flow: %s -> %s (%.2f RPS)",
						strings.Join(fc.CallStack, " -> "), flowKey, inputRate)

					// Track flow path with current conditional context
					fc.trackFlowPath(flowKey, fc.currentCondition, fc.currentProbability)

					// Track the success rate of this method call for the variable
					if len(stmt.Variables) > 0 && stmt.Variables[0] != nil && stmt.Variables[0].Value != "" {
						// Get the success rate of the called method
						successRate := fc.getMethodSuccessRate(component, method)
						fc.VariableOutcomes[stmt.Variables[0].Value] = successRate
						log.Printf("FlowEval: Variable '%s' assigned from %s with success rate %.2f",
							stmt.Variables[0].Value, flowKey, successRate)
					}
				}
			}
		}
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
// It also resolves dependency names to actual instance names in the system
func (fc *FlowContext) parseCallTarget(target string) (component, method string) {
	parts := strings.Split(target, ".")
	if len(parts) >= 2 {
		// Last part is the method, everything before is the component path
		method = parts[len(parts)-1]
		dependencyName := strings.Join(parts[:len(parts)-1], ".")

		// Resolve dependency name to actual instance name using current call context
		component = fc.resolveDependencyToInstance(dependencyName)
		if component == "" {
			// Fallback to dependency name if resolution fails
			component = dependencyName
		}

		return component, method
	}
	return "", ""
}

// resolveDependencyToInstance resolves a dependency name to the actual instance name in the system
func (fc *FlowContext) resolveDependencyToInstance(dependencyName string) string {
	if fc.System == nil || fc.CurrentComponent == "" {
		return dependencyName // Cannot resolve without context
	}

	// Find the current component's instance declaration in the system
	for _, bodyItem := range fc.System.Body {
		if instance, ok := bodyItem.(*InstanceDecl); ok {
			if instance.Name.Value == fc.CurrentComponent {
				// Look through the instance overrides to find dependency mapping
				for _, assignment := range instance.Overrides {
					if assignment.Var.Value == dependencyName {
						// Found the mapping: dependencyName = actualInstanceName
						if targetIdent, ok := assignment.Value.(*IdentifierExpr); ok {
							log.Printf("FlowEval: Resolved dependency '%s' in component '%s' to instance '%s'",
								dependencyName, fc.CurrentComponent, targetIdent.Value)
							return targetIdent.Value
						}
					}
				}
			}
		}
	}

	// No mapping found, return the dependency name as-is
	log.Printf("FlowEval: No dependency mapping found for '%s' in component '%s', using as-is",
		dependencyName, fc.CurrentComponent)
	return dependencyName
}

// evaluateConditionProbability determines the probability of a condition being true
func (fc *FlowContext) evaluateConditionProbability(condition Expr) float64 {
	// Check if the condition is a simple identifier (variable reference)
	if identExpr, ok := condition.(*IdentifierExpr); ok {
		varName := identExpr.Value

		// Check if we have tracked outcome for this variable
		if outcome, exists := fc.VariableOutcomes[varName]; exists {
			log.Printf("FlowEval: Using tracked outcome for variable '%s': %.2f", varName, outcome)
			return outcome
		}
	}

	// Legacy fallback: Try to extract cache hit rate from parameters
	if fc.Parameters != nil {
		if hitRate, ok := fc.Parameters["CacheHitRate"]; ok {
			if rate, ok := hitRate.(float64); ok {
				return rate
			}
		}
	}

	// Default to 50% probability for unknown conditions
	log.Printf("FlowEval: No tracked outcome for condition, using default 0.5")
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

// getMethodSuccessRate returns the success rate for a component method
func (fc *FlowContext) getMethodSuccessRate(component, method string) float64 {
	flowKey := fmt.Sprintf("%s.%s", component, method)

	// First check if we have a tracked success rate from the flow analysis
	if rate, exists := fc.SuccessRates[flowKey]; exists {
		return rate
	}

	// Check if this is a native component with FlowAnalyzable interface
	if flowAnalyzable, exists := fc.NativeComponents[component]; exists {
		pattern := flowAnalyzable.GetFlowPattern(method, 1.0, fc.getComponentParams(component))
		return pattern.SuccessRate
	}

	// Default to 100% success rate
	return 1.0
}

// SetNativeComponent registers a native component for flow analysis
func (fc *FlowContext) SetNativeComponent(name string, component components.FlowAnalyzable) {
	fc.NativeComponents[name] = component
}

// SetResourceLimit sets the resource limit for a component
func (fc *FlowContext) SetResourceLimit(component string, limit int) {
	fc.ResourceLimits[component] = limit
}

// trackFlowPath records a flow path with order and condition information
func (fc *FlowContext) trackFlowPath(flowKey string, condition string, probability float64) {
	// Only track flow paths if we haven't seen this path before
	// This prevents accumulation across iterations
	fromKey := ""
	if len(fc.CallStack) > 0 {
		fromKey = fc.CallStack[len(fc.CallStack)-1]
	}
	
	// Create unique flow path key including source
	pathKey := fmt.Sprintf("%s->%s", fromKey, flowKey)
	
	// If we've already tracked this path, don't update it
	if _, exists := fc.FlowPaths[pathKey]; exists {
		return
	}
	
	// Special handling for calls that happen within other methods
	// For example, idx.Find happens within database.LookupByPhone
	var baseOrder float64
	
	// Check if this is a call from database.LookupByPhone to idx.Find
	if strings.Contains(fromKey, "database.LookupByPhone") && strings.Contains(flowKey, "idx.") {
		// This should be ordered as 3.1 since database.LookupByPhone is order 3
		for path, info := range fc.FlowPaths {
			if strings.HasSuffix(path, "->database.LookupByPhone") {
				baseOrder = info.Order + 0.1
				break
			}
		}
		if baseOrder == 0 {
			// Fallback if we don't find the parent
			fc.FlowOrder++
			baseOrder = float64(fc.FlowOrder)
		}
	} else {
		// Use per-generator flow tracking if available
		if fc.CurrentGeneratorID != "" {
			// Increment the generator-specific order counter
			fc.GeneratorFlowOrder[fc.CurrentGeneratorID]++
			generatorOrder := fc.GeneratorFlowOrder[fc.CurrentGeneratorID]
			
			// Create prefixed order (e.g., "A1", "B1", "A2", "B2")
			baseOrder = float64(generatorOrder)
			
			// If this is a conditional path, use decimal numbering
			if condition != "" {
				// Check if we already have a conditional at this level for this generator
				for _, info := range fc.FlowPaths {
					if info.GeneratorID == fc.CurrentGeneratorID && int(info.Order) == generatorOrder && info.Condition != "" {
						// Use decimal for alternative path
						baseOrder = float64(generatorOrder) + 0.1
						break
					}
				}
			}
		} else {
			// Fallback to global flow tracking
			fc.FlowOrder++
			baseOrder = float64(fc.FlowOrder)
			
			// If this is a conditional path, use decimal numbering
			if condition != "" {
				// Check if we already have a conditional at this level
				for _, info := range fc.FlowPaths {
					if int(info.Order) == fc.FlowOrder && info.Condition != "" {
						// Use decimal for alternative path
						baseOrder = float64(fc.FlowOrder) + 0.1
						break
					}
				}
			}
		}
	}
	
	// Get generator color
	color := "#fbbf24" // Default amber color
	if fc.CurrentGeneratorID != "" {
		if genColor, exists := fc.GeneratorColors[fc.CurrentGeneratorID]; exists {
			color = genColor
		}
	}
	
	fc.FlowPaths[pathKey] = FlowPathInfo{
		Order:       baseOrder,
		Condition:   condition,
		Probability: probability,
		GeneratorID: fc.CurrentGeneratorID,
		Color:       color,
	}
	
	log.Printf("FlowEval: Tracked flow path %s with order %.1f, condition: %s, generator: %s, color: %s", pathKey, baseOrder, condition, fc.CurrentGeneratorID, color)
}

// conditionToString converts a condition expression to a readable string
func (fc *FlowContext) conditionToString(condition Expr) string {
	switch c := condition.(type) {
	case *IdentifierExpr:
		return c.Value
	case *UnaryExpr:
		return fmt.Sprintf("%s%s", c.Operator, fc.conditionToString(c.Right))
	case *BinaryExpr:
		return fmt.Sprintf("%s %s %s", fc.conditionToString(c.Left), c.Operator, fc.conditionToString(c.Right))
	default:
		return "condition"
	}
}

// analyzeStatementWithCondition analyzes a statement within a conditional context
func (fc *FlowContext) analyzeStatementWithCondition(stmt Stmt, inputRate float64, outflows map[string]float64, condition string, probability float64) {
	// Store current condition context
	savedCondition := fc.currentCondition
	savedProbability := fc.currentProbability
	fc.currentCondition = condition
	fc.currentProbability = probability

	// Analyze the statement
	fc.analyzeStatement(stmt, inputRate, outflows)

	// Restore previous context
	fc.currentCondition = savedCondition
	fc.currentProbability = savedProbability
}
