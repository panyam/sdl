package runtime

import (
	"github.com/panyam/sdl/lib/core"
)

// FlowNativeMethodInfo describes flow and timing characteristics of a native method
type FlowNativeMethodInfo struct {
	// Whether this method induces delay
	HasDelay bool

	// Function to extract delay from AST arguments during flow analysis
	// Returns 0 if delay cannot be determined statically
	ExtractDelay func(args []Expr) core.Duration

	// Whether this method makes external calls (affects flow)
	HasOutflows bool

	// Static outflows if known (e.g., a native logging service)
	// Map of "component.method" -> rate multiplier
	Outflows map[string]float64
}

// flowNativeMethods is the registry for flow analysis
// Separate from execution-time native methods for flexibility
var flowNativeMethods = map[string]*FlowNativeMethodInfo{
	"delay": {
		HasDelay: true,
		ExtractDelay: func(args []Expr) core.Duration {
			// Try to extract delay from first argument
			if len(args) > 0 {
				if lit, ok := args[0].(*LiteralExpr); ok {
					switch v := lit.Value.Value.(type) {
					case int64:
						return core.Duration(v)
					case core.Duration:
						return v
					}
				}
			}
			// Default to 0 if we can't determine statically
			return 0
		},
		HasOutflows: false,
	},
	"log": {
		// Logging typically has minimal delay
		HasDelay:    false,
		HasOutflows: false,
	},
}

// RegisterFlowNativeMethod registers a native method for flow analysis
func RegisterFlowNativeMethod(name string, info *FlowNativeMethodInfo) {
	flowNativeMethods[name] = info
}

// GetFlowNativeMethodInfo returns flow analysis info for a native method
// Returns nil if not registered (caller should use default behavior)
func GetFlowNativeMethodInfo(methodName string) *FlowNativeMethodInfo {
	return flowNativeMethods[methodName]
}

// analyzeNativeMethodCall analyzes a native method call during flow evaluation
func analyzeNativeMethodCall(methodName string, callExpr *CallExpr, inputRate float64, scope *FlowScope, outflows RateMap) core.Duration {
	info := GetFlowNativeMethodInfo(methodName)
	if info == nil {
		// Not registered for flow analysis - assume no delay, no outflows
		return 0
	}

	var totalDelay core.Duration

	// Extract delay if applicable
	if info.HasDelay && info.ExtractDelay != nil {
		totalDelay = info.ExtractDelay(callExpr.ArgList)
		if totalDelay > 0 {
			// log.Printf("analyzeNativeMethodCall: %s induces delay of %v", methodName, totalDelay)
		}
	}

	// Add any outflows
	if info.HasOutflows && info.Outflows != nil {
		for target, rateMultiplier := range info.Outflows {
			targetComp, targetMethod := scope.ResolveTarget(target)
			if targetComp != nil {
				outflows.AddFlow(targetComp, targetMethod, inputRate*rateMultiplier)

				// Record flow edge if tracking
				if scope.CurrentComponent != nil && scope.CurrentMethod != nil && scope.FlowEdges != nil {
					scope.FlowEdges.AddEdge(
						scope.CurrentComponent,
						scope.CurrentMethod.Name.Value,
						targetComp,
						targetMethod,
						inputRate*rateMultiplier,
					)
				}
			}
		}
	}

	return totalDelay
}
