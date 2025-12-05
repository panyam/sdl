package runtime

import (
	"fmt"
	"sync"
)

// FlowStrategy defines the interface for different flow evaluation strategies
type FlowStrategy interface {
	// Evaluate performs flow analysis given system and generators
	Evaluate(system *SystemInstance, generators []GeneratorConfigAPI) (*FlowAnalysisResult, error)
	
	// GetInfo returns metadata about this strategy
	GetInfo() StrategyInfo
	
	// IsAvailable checks if this strategy can be used
	IsAvailable() bool
}

// GeneratorConfigAPI represents a traffic generator configuration for the Flow API
// This is the API-friendly version (uses component names, not instances)
type GeneratorConfigAPI struct {
	ID         string  `json:"id"`
	Component  string  `json:"component"`
	Method     string  `json:"method"`
	Rate       float64 `json:"rate"`
}

// FlowAnalysisResult contains the results of flow analysis
type FlowAnalysisResult struct {
	Strategy    string                    `json:"strategy"`
	Status      FlowStatus                `json:"status"`
	Iterations  int                       `json:"iterations,omitempty"`
	System      string                    `json:"system"`
	Generators  []GeneratorConfigAPI      `json:"generators"`
	Flows       FlowData                  `json:"flows"`
	Warnings    []string                  `json:"warnings,omitempty"`
}

// FlowStatus indicates the status of flow analysis
type FlowStatus string

const (
	FlowStatusConverged FlowStatus = "converged"
	FlowStatusPartial   FlowStatus = "partial"
	FlowStatusFailed    FlowStatus = "failed"
)

// FlowData contains the flow analysis data
type FlowData struct {
	Edges          []FlowEdgeAPI         `json:"edges"`
	ComponentRates map[string]float64    `json:"componentRates"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// FlowEdgeAPI represents a flow between two component methods (API version)
type FlowEdgeAPI struct {
	From ComponentMethod `json:"from"`
	To   ComponentMethod `json:"to"`
	Rate float64         `json:"rate"`
}

// ComponentMethod identifies a component and method
type ComponentMethod struct {
	Component string `json:"component"`
	Method    string `json:"method"`
}

// StrategyInfo provides metadata about a flow strategy
type StrategyInfo struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Status       string   `json:"status"`
	Limitations  []string `json:"limitations"`
	Recommended  bool     `json:"recommended"`
}

// Global strategy registry
var (
	flowStrategies = make(map[string]FlowStrategy)
	flowStrategyMutex  sync.RWMutex
	defaultFlowStrategy = "runtime"
)

// RegisterFlowStrategy registers a new flow evaluation strategy
func RegisterFlowStrategy(name string, strategy FlowStrategy) error {
	flowStrategyMutex.Lock()
	defer flowStrategyMutex.Unlock()
	
	if _, exists := flowStrategies[name]; exists {
		return fmt.Errorf("flow strategy '%s' already registered", name)
	}
	
	flowStrategies[name] = strategy
	return nil
}

// GetFlowStrategy retrieves a registered flow strategy
func GetFlowStrategy(name string) (FlowStrategy, error) {
	flowStrategyMutex.RLock()
	defer flowStrategyMutex.RUnlock()
	
	strategy, exists := flowStrategies[name]
	if !exists {
		return nil, fmt.Errorf("flow strategy '%s' not found", name)
	}
	
	return strategy, nil
}

// ListFlowStrategies returns all registered flow strategies
func ListFlowStrategies() map[string]StrategyInfo {
	flowStrategyMutex.RLock()
	defer flowStrategyMutex.RUnlock()
	
	result := make(map[string]StrategyInfo)
	for name, strategy := range flowStrategies {
		result[name] = strategy.GetInfo()
	}
	
	return result
}

// EvaluateFlowStrategy runs flow analysis with the specified strategy
func EvaluateFlowStrategy(strategyName string, system *SystemInstance, generators []GeneratorConfigAPI) (*FlowAnalysisResult, error) {
	strategy, err := GetFlowStrategy(strategyName)
	if err != nil {
		return nil, err
	}
	
	if !strategy.IsAvailable() {
		return nil, fmt.Errorf("flow strategy '%s' is not available", strategyName)
	}
	
	return strategy.Evaluate(system, generators)
}

// GetDefaultFlowStrategy returns the name of the default flow strategy
func GetDefaultFlowStrategy() string {
	return "runtime"
}