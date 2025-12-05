package runtime

// FlowState represents the current flow analysis state
type FlowState struct {
	Strategy         string             `json:"strategy"`
	AppliedAt        string             `json:"appliedAt,omitempty"`
	Rates            map[string]float64 `json:"rates"`
	ManualOverrides  map[string]float64 `json:"manualOverrides,omitempty"`
}

// ArrivalRateUpdate represents a request to update a component method's arrival rate
type ArrivalRateUpdate struct {
	Component string  `json:"component"`
	Method    string  `json:"method"`
	Rate      float64 `json:"rate"`
}

// FlowComparisonResult represents a comparison between two flow analyses
type FlowComparisonResult struct {
	Strategy1   string                 `json:"strategy1"`
	Strategy2   string                 `json:"strategy2"`
	Differences []FlowRateDifference   `json:"differences"`
	Similarity  float64                `json:"similarity"` // 0.0 to 1.0
}

// FlowRateDifference represents a difference in flow rates between strategies
type FlowRateDifference struct {
	Component string  `json:"component"`
	Method    string  `json:"method"`
	Rate1     float64 `json:"rate1"`
	Rate2     float64 `json:"rate2"`
	Delta     float64 `json:"delta"`
	Percent   float64 `json:"percent"` // Percentage difference
}