package runtime

import (
	"log"
)

// GeneratorEntryPointRuntime represents an entry point with runtime component instance
type GeneratorEntryPointRuntime struct {
	Component   *ComponentInstance // Component instance
	Method      string             // Method name
	Rate        float64            // requests per second
	GeneratorID string             // generator identifier
}

// FlowEvalRuntime computes outbound traffic rates from a given component instance and method
// Returns RateMap of downstream component instances and their methods with rates
func FlowEvalRuntime(component *ComponentInstance, method string, inputRate float64, scope *FlowScope) RateMap {
	if component == nil || method == "" || scope == nil {
		return NewRateMap()
	}
	
	outflows := NewRateMap()
	
	// Check for cycles using component instance
	if scope.IsInCallStack(component) {
		log.Printf("FlowEvalRuntime: Cycle detected for %s.%s, breaking recursion", component.ID(), method)
		return outflows
	}
	
	// Prevent infinite recursion
	if len(scope.CallStack) >= 20 {
		log.Printf("FlowEvalRuntime: Maximum call depth reached, stopping recursion")
		return outflows
	}
	
	// For native components, use GetFlowPattern directly
	if component.IsNative {
		pattern := component.GetFlowPattern(method, inputRate)
		// TODO: Convert pattern.Outflows (string-based) to ComponentInstance refs
		// For now, we'll need a resolver in scope to handle this
		for target, rate := range pattern.Outflows {
			targetComp, targetMethod := scope.ResolveTarget(target)
			if targetComp != nil {
				outflows.AddFlow(targetComp, targetMethod, rate)
			}
		}
		return outflows
	}
	
	// For SDL components, get the method declaration
	methodDecl, err := component.ComponentDecl.GetMethod(method)
	if err != nil || methodDecl == nil {
		log.Printf("FlowEvalRuntime: Method %s.%s not found", component.ID(), method)
		return outflows
	}
	
	// Push new scope for this method evaluation
	newScope := scope.Push(component, methodDecl)
	
	// Clear variable outcomes for this method analysis
	newScope.VariableOutcomes = make(map[string]float64)
	
	// Analyze method body for call patterns
	if methodDecl.Body != nil && methodDecl.Body.Statements != nil {
		analyzeStatementsRuntime(methodDecl.Body.Statements, inputRate, newScope, outflows)
	}
	
	return outflows
}

// analyzeStatementsRuntime processes a list of statements with runtime flow analysis
func analyzeStatementsRuntime(statements []Stmt, inputRate float64, scope *FlowScope, outflows RateMap) {
	for _, stmt := range statements {
		analyzeStatementRuntime(stmt, inputRate, scope, outflows)
	}
}

// analyzeStatementRuntime processes a single statement with runtime flow analysis
func analyzeStatementRuntime(stmt Stmt, inputRate float64, scope *FlowScope, outflows RateMap) {
	switch s := stmt.(type) {
	case *ExprStmt:
		analyzeExprStatementRuntime(s, inputRate, scope, outflows)
	case *IfStmt:
		analyzeIfStatementRuntime(s, inputRate, scope, outflows)
	case *AssignmentStmt:
		analyzeAssignmentStatementRuntime(s, inputRate, scope, outflows)
	case *ReturnStmt:
		analyzeReturnStatementRuntime(s, inputRate, scope, outflows)
	case *ForStmt:
		analyzeForStatementRuntime(s, inputRate, scope, outflows)
	case *BlockStmt:
		analyzeBlockStatementRuntime(s, inputRate, scope, outflows)
	case *LetStmt:
		analyzeLetStatementRuntime(s, inputRate, scope, outflows)
	default:
		// Other statement types don't generate flows
		log.Printf("analyzeStatementRuntime: Unhandled statement type: %T", stmt)
	}
}

// analyzeExprStatementRuntime processes expression statements that might contain calls
func analyzeExprStatementRuntime(stmt *ExprStmt, inputRate float64, scope *FlowScope, outflows RateMap) {
	// Check if the expression is a call
	if callExpr, ok := stmt.Expression.(*CallExpr); ok {
		analyzeCallExprRuntime(callExpr, inputRate, scope, outflows)
	}
}

// analyzeCallExprRuntime processes a call expression and adds it to outflows
func analyzeCallExprRuntime(callExpr *CallExpr, inputRate float64, scope *FlowScope, outflows RateMap) {
	// Extract the target component and method from the call
	targetComp, targetMethod := extractCallTargetRuntime(callExpr, scope)
	if targetComp != nil && targetMethod != "" {
		outflows.AddFlow(targetComp, targetMethod, inputRate)
		log.Printf("analyzeCallExprRuntime: Flow from %s.%s to %s.%s (%.2f RPS)",
			scope.CurrentComponent.ID(), scope.CurrentMethod.Name.Value,
			targetComp.ID(), targetMethod, inputRate)
	}
}

// extractCallTargetRuntime extracts the target component instance and method from a call expression
func extractCallTargetRuntime(call *CallExpr, scope *FlowScope) (*ComponentInstance, string) {
	if call == nil || call.Function == nil {
		return nil, ""
	}
	
	// Handle member access like self.db.LookupByPhone
	if memberExpr, ok := call.Function.(*MemberAccessExpr); ok {
		// Extract the dependency name and method
		depPath := extractMemberPath(memberExpr)
		if len(depPath) < 2 {
			return nil, ""
		}
		
		// Last element is the method name
		methodName := depPath[len(depPath)-1]
		
		// Everything before is the component path (e.g., "db" or "cache.inner")
		componentPath := depPath[:len(depPath)-1]
		
		// Resolve the component path to an actual instance
		targetComp := resolveComponentPath(componentPath, scope)
		return targetComp, methodName
	}
	
	// Handle simple function calls (might be local methods)
	if ident, ok := call.Function.(*IdentifierExpr); ok {
		// For now, assume it's a method on the current component
		return scope.CurrentComponent, ident.Value
	}
	
	return nil, ""
}

// extractMemberPath extracts the full path from a member expression
func extractMemberPath(expr *MemberAccessExpr) []string {
	var parts []string
	
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
	
	return parts
}

// resolveComponentPath resolves a component path to an actual ComponentInstance
func resolveComponentPath(path []string, scope *FlowScope) *ComponentInstance {
	if len(path) == 0 || scope.CurrentComponent == nil {
		return nil
	}
	
	// Start with the current component's environment
	currentEnv := scope.SysEnv
	
	// Look up the first element in the path
	depName := path[0]
	depValue, exists := currentEnv.Get(depName)
	if !exists {
		log.Printf("resolveComponentPath: Dependency '%s' not found in current scope", depName)
		return nil
	}
	
	// Extract the ComponentInstance from the Value
	if depValue.Value == nil {
		return nil
	}
	
	compInst, ok := depValue.Value.(*ComponentInstance)
	if !ok {
		log.Printf("resolveComponentPath: Value for '%s' is not a ComponentInstance", depName)
		return nil
	}
	
	// TODO: Handle nested paths (e.g., cache.inner) if needed
	if len(path) > 1 {
		log.Printf("resolveComponentPath: Nested paths not yet implemented: %v", path)
	}
	
	return compInst
}

// analyzeIfStatementRuntime handles conditional flows with runtime analysis
func analyzeIfStatementRuntime(stmt *IfStmt, inputRate float64, scope *FlowScope, outflows RateMap) {
	// Evaluate condition probability
	conditionProb := evaluateConditionProbabilityRuntime(stmt.Condition, scope)
	
	log.Printf("analyzeIfStatementRuntime: Processing if statement with condition probability %.2f", conditionProb)
	
	// Analyze then branch with probability-weighted rate
	if stmt.Then != nil {
		analyzeStatementRuntime(stmt.Then, inputRate*conditionProb, scope, outflows)
	}
	
	// Analyze else branch with inverted probability
	if stmt.Else != nil {
		analyzeStatementRuntime(stmt.Else, inputRate*(1.0-conditionProb), scope, outflows)
	}
}

// evaluateConditionProbabilityRuntime evaluates condition probability using runtime information
func evaluateConditionProbabilityRuntime(condition Expr, scope *FlowScope) float64 {
	// Check if the condition is a simple identifier (variable reference)
	if identExpr, ok := condition.(*IdentifierExpr); ok {
		varName := identExpr.Value
		
		// Check if we have tracked outcome for this variable
		if outcome, exists := scope.VariableOutcomes[varName]; exists {
			log.Printf("evaluateConditionProbabilityRuntime: Using tracked outcome for variable '%s': %.2f", varName, outcome)
			return outcome
		}
		
		// Try to look up the variable in the environment
		if value, exists := scope.GetVariable(varName); exists {
			// If it's a boolean with success rate info, use that
			// TODO: Extract success rate from Value if available
			_ = value
		}
	}
	
	// Default to 50% probability for unknown conditions
	log.Printf("evaluateConditionProbabilityRuntime: No tracked outcome for condition, using default 0.5")
	return 0.5
}

// analyzeAssignmentStatementRuntime handles assignments that might contain calls
func analyzeAssignmentStatementRuntime(stmt *AssignmentStmt, inputRate float64, scope *FlowScope, outflows RateMap) {
	// Check if the assigned expression contains a call
	if callExpr, ok := stmt.Value.(*CallExpr); ok {
		analyzeCallExprRuntime(callExpr, inputRate, scope, outflows)
	}
}

// analyzeReturnStatementRuntime handles return statements that might contain calls
func analyzeReturnStatementRuntime(stmt *ReturnStmt, inputRate float64, scope *FlowScope, outflows RateMap) {
	if stmt.ReturnValue != nil {
		if callExpr, ok := stmt.ReturnValue.(*CallExpr); ok {
			analyzeCallExprRuntime(callExpr, inputRate, scope, outflows)
		}
	}
}

// analyzeForStatementRuntime handles loops
func analyzeForStatementRuntime(stmt *ForStmt, inputRate float64, scope *FlowScope, outflows RateMap) {
	// For now, assume the loop body executes once per input
	// TODO: Implement proper loop analysis based on loop bounds
	if stmt.Body != nil {
		analyzeStatementRuntime(stmt.Body, inputRate, scope, outflows)
	}
}

// analyzeBlockStatementRuntime handles block statements
func analyzeBlockStatementRuntime(stmt *BlockStmt, inputRate float64, scope *FlowScope, outflows RateMap) {
	if stmt.Statements != nil {
		analyzeStatementsRuntime(stmt.Statements, inputRate, scope, outflows)
	}
}

// analyzeLetStatementRuntime handles let statements that might contain calls
func analyzeLetStatementRuntime(stmt *LetStmt, inputRate float64, scope *FlowScope, outflows RateMap) {
	// Check if the assigned expression contains a call
	if stmt.Value != nil {
		if callExpr, ok := stmt.Value.(*CallExpr); ok {
			// Analyze the call
			analyzeCallExprRuntime(callExpr, inputRate, scope, outflows)
			
			// Track the success rate of this method call for the variable
			if len(stmt.Variables) > 0 && stmt.Variables[0] != nil && stmt.Variables[0].Value != "" {
				targetComp, targetMethod := extractCallTargetRuntime(callExpr, scope)
				if targetComp != nil && targetMethod != "" {
					// Get the success rate of the called method
					successRate := getMethodSuccessRateRuntime(targetComp, targetMethod, scope)
					scope.TrackVariableOutcome(stmt.Variables[0].Value, successRate)
					log.Printf("analyzeLetStatementRuntime: Variable '%s' assigned from %s.%s with success rate %.2f",
						stmt.Variables[0].Value, targetComp.ID(), targetMethod, successRate)
				}
			}
		}
	}
}

// getMethodSuccessRateRuntime returns the success rate for a component method
func getMethodSuccessRateRuntime(component *ComponentInstance, method string, scope *FlowScope) float64 {
	// Check if we have a tracked success rate from the flow analysis
	if rate := scope.SuccessRates.GetRate(component, method); rate > 0 {
		return rate
	}
	
	// For native components, get it from GetFlowPattern
	if component.IsNative {
		pattern := component.GetFlowPattern(method, 1.0)
		return pattern.SuccessRate
	}
	
	// Default to 100% success rate
	return 1.0
}

// SolveSystemFlowsRuntime performs flow analysis using runtime component instances
// This implements a two-phase approach:
// 1. Flow propagation through the component graph
// 2. Iterative back-pressure adjustment until convergence
func SolveSystemFlowsRuntime(generators []GeneratorEntryPointRuntime, scope *FlowScope) RateMap {
	if scope == nil {
		log.Printf("SolveSystemFlowsRuntime: scope is nil")
		return NewRateMap()
	}
	
	// Initialize arrival rates with entry points
	for _, gen := range generators {
		if gen.Component != nil && gen.Method != "" {
			scope.ArrivalRates.SetRate(gen.Component, gen.Method, gen.Rate)
		}
	}
	
	log.Printf("SolveSystemFlowsRuntime: Starting fixed-point iteration with %d entry points", len(generators))
	
	// Configuration for convergence
	maxIterations := 10
	convergenceThreshold := 0.01 // Match string-based convergence threshold
	dampingFactor := 0.5
	
	// Iterate until convergence
	for iteration := 0; iteration < maxIterations; iteration++ {
		// Save old rates for convergence check
		oldRates := scope.ArrivalRates.Copy()
		
		// Phase 1: Recompute outgoing flows for each component
		newRates := NewRateMap()
		
		// Start with entry points
		for _, gen := range generators {
			if gen.Component != nil && gen.Method != "" {
				newRates.SetRate(gen.Component, gen.Method, gen.Rate)
			}
		}
		
		// Propagate flows through the system
		for component, methods := range scope.ArrivalRates {
			for method, inRate := range methods {
				if inRate > 1e-9 { // Only process non-trivial rates
					// Get outflows from this component.method
					outflows := FlowEvalRuntime(component, method, inRate, scope)
					
					// Add outflows to the new rates
					for targetComp, targetMethods := range outflows {
						for targetMethod, outRate := range targetMethods {
							newRates.AddFlow(targetComp, targetMethod, outRate)
						}
					}
				}
			}
		}
		
		// Phase 2: Update success rates based on new arrival rates
		updateSuccessRatesRuntime(newRates, scope)
		
		// Check convergence (all rates changed by < threshold)
		maxChange := computeMaxChange(oldRates, newRates)
		
		log.Printf("SolveSystemFlowsRuntime: Iteration %d, max change: %.6f", iteration, maxChange)
		
		if maxChange < convergenceThreshold {
			log.Printf("SolveSystemFlowsRuntime: Converged after %d iterations", iteration+1)
			scope.ArrivalRates = newRates
			return newRates
		}
		
		// Apply damping to prevent oscillation
		scope.ArrivalRates = applyDamping(oldRates, newRates, dampingFactor)
	}
	
	log.Printf("SolveSystemFlowsRuntime: Did not converge after %d iterations", maxIterations)
	return scope.ArrivalRates
}

// computeMaxChange calculates the maximum rate change between old and new rates
func computeMaxChange(oldRates, newRates RateMap) float64 {
	maxChange := 0.0
	
	// Check all components in both maps
	allComponents := make(map[*ComponentInstance]bool)
	for comp := range oldRates {
		allComponents[comp] = true
	}
	for comp := range newRates {
		allComponents[comp] = true
	}
	
	// Compare rates for each component and method
	for comp := range allComponents {
		// Get all methods for this component
		allMethods := make(map[string]bool)
		if oldRates[comp] != nil {
			for method := range oldRates[comp] {
				allMethods[method] = true
			}
		}
		if newRates[comp] != nil {
			for method := range newRates[comp] {
				allMethods[method] = true
			}
		}
		
		// Compare rates for each method
		for method := range allMethods {
			oldRate := oldRates.GetRate(comp, method)
			newRate := newRates.GetRate(comp, method)
			change := abs(newRate - oldRate)
			if change > maxChange {
				maxChange = change
			}
		}
	}
	
	return maxChange
}

// applyDamping applies damping factor to prevent oscillation
func applyDamping(oldRates, newRates RateMap, dampingFactor float64) RateMap {
	dampedRates := NewRateMap()
	
	// Get all components from both maps
	allComponents := make(map[*ComponentInstance]bool)
	for comp := range oldRates {
		allComponents[comp] = true
	}
	for comp := range newRates {
		allComponents[comp] = true
	}
	
	// Apply damping formula: dampedRate = oldRate + dampingFactor * (newRate - oldRate)
	for comp := range allComponents {
		// Get all methods for this component
		allMethods := make(map[string]bool)
		if oldRates[comp] != nil {
			for method := range oldRates[comp] {
				allMethods[method] = true
			}
		}
		if newRates[comp] != nil {
			for method := range newRates[comp] {
				allMethods[method] = true
			}
		}
		
		// Apply damping to each method
		for method := range allMethods {
			oldRate := oldRates.GetRate(comp, method)
			newRate := newRates.GetRate(comp, method)
			dampedRate := oldRate + dampingFactor*(newRate-oldRate)
			if dampedRate > 1e-9 { // Only store non-trivial rates
				dampedRates.SetRate(comp, method, dampedRate)
			}
		}
	}
	
	return dampedRates
}

// updateSuccessRatesRuntime updates success rates based on current arrival rates (back-pressure modeling)
func updateSuccessRatesRuntime(arrivalRates RateMap, scope *FlowScope) {
	for component, methods := range arrivalRates {
		for method, arrivalRate := range methods {
			// For native components, get success rate from GetFlowPattern
			if component.IsNative {
				pattern := component.GetFlowPattern(method, arrivalRate)
				scope.SuccessRates.SetRate(component, method, pattern.SuccessRate)
				continue
			}
			
			// For SDL components, use a simple back-pressure model
			// This is a placeholder - could be enhanced with specific component analysis
			successRate := 1.0
			
			// Simple back-pressure model: success rate degrades with high arrival rates
			// In a real system, this would consider component-specific capacity limits
			if arrivalRate > 100.0 { // Arbitrary threshold for demonstration
				// Success rate degrades as load increases
				utilization := arrivalRate / 100.0
				successRate = max(0.1, 1.0-utilization*0.5)
			}
			
			scope.SuccessRates.SetRate(component, method, successRate)
		}
	}
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}