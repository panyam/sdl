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

// UtilizationProvider is an interface for components that can report their utilization.
// This allows hierarchical components to expose utilization of their bottleneck resources.
type UtilizationProvider interface {
	// GetUtilizationInfo returns information about resource utilization.
	// Components can return multiple utilization metrics if they have multiple resources.
	GetUtilizationInfo() []UtilizationInfo
}

// UtilizationInfo represents utilization information for a resource.
type UtilizationInfo struct {
	// ResourceName identifies the resource (e.g., "pool", "disk", "cpu")
	ResourceName string
	
	// ComponentPath is the path to the component (e.g., "database.pool", "database.driverTable.disk")
	ComponentPath string
	
	// Utilization is the current utilization (0.0 to 1.0)
	Utilization float64
	
	// Capacity is the maximum capacity (e.g., pool size, queue size)
	Capacity float64
	
	// CurrentLoad is the current load (e.g., arrival rate)
	CurrentLoad float64
	
	// IsBottleneck indicates if this is likely the bottleneck resource
	IsBottleneck bool
	
	// WarningThreshold is the utilization level that triggers a warning (e.g., 0.8)
	WarningThreshold float64
	
	// CriticalThreshold is the utilization level that triggers critical alert (e.g., 0.95)
	CriticalThreshold float64
}

// GetBottleneckUtilization returns the highest utilization among all resources.
// This is a helper function for components with multiple resources.
func GetBottleneckUtilization(infos []UtilizationInfo) *UtilizationInfo {
	if len(infos) == 0 {
		return nil
	}
	
	var bottleneck *UtilizationInfo
	maxUtil := 0.0
	
	for i := range infos {
		if infos[i].Utilization > maxUtil {
			maxUtil = infos[i].Utilization
			bottleneck = &infos[i]
		}
	}
	
	if bottleneck != nil {
		bottleneck.IsBottleneck = true
	}
	
	return bottleneck
}
