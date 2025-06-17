package components

// FlowPattern describes the traffic behavior of a component method
type FlowPattern struct {
	// Output traffic to other components
	Outflows map[string]float64 // {component.method: outputRate}

	// Method behavior characteristics
	SuccessRate   float64 // 0.0 - 1.0
	Amplification float64 // outputRate / inputRate
	ServiceTime   float64 // Average service time in seconds

	// Optional: detailed conditional flows
	ConditionalFlows []ConditionalFlow
}

// ConditionalFlow represents conditional traffic patterns
type ConditionalFlow struct {
	Condition   string             // "cache_miss", "retry", "batch_processing"
	Probability float64            // 0.0 - 1.0
	Outflows    map[string]float64 // {component.method: outputRate}
}

// FlowAnalyzable interface for native components to report their flow patterns
type FlowAnalyzable interface {
	GetFlowPattern(methodName string, inputRate float64) FlowPattern
}
