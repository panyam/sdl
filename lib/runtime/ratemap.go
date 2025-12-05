package runtime

// RateMap tracks flow rates per component instance and method
type RateMap map[*ComponentInstance]map[string]float64

// NewRateMap creates a new empty RateMap
func NewRateMap() RateMap {
	return make(RateMap)
}

// AddFlow adds or updates the flow rate for a component method
func (rm RateMap) AddFlow(component *ComponentInstance, method string, rate float64) {
	if component == nil || method == "" {
		return
	}
	
	if rm[component] == nil {
		rm[component] = make(map[string]float64)
	}
	rm[component][method] += rate
}

// SetRate sets the exact rate for a component method (replacing any existing value)
func (rm RateMap) SetRate(component *ComponentInstance, method string, rate float64) {
	if component == nil || method == "" {
		return
	}
	
	if rm[component] == nil {
		rm[component] = make(map[string]float64)
	}
	rm[component][method] = rate
}

// GetRate returns the flow rate for a component method
func (rm RateMap) GetRate(component *ComponentInstance, method string) float64 {
	if component == nil || method == "" || rm[component] == nil {
		return 0.0
	}
	return rm[component][method]
}

// Clear removes all entries from the RateMap
func (rm RateMap) Clear() {
	for k := range rm {
		delete(rm, k)
	}
}

// Copy creates a deep copy of the RateMap
func (rm RateMap) Copy() RateMap {
	result := NewRateMap()
	for comp, methods := range rm {
		result[comp] = make(map[string]float64)
		for method, rate := range methods {
			result[comp][method] = rate
		}
	}
	return result
}

// GetTotalRate returns the sum of all rates in the map
func (rm RateMap) GetTotalRate() float64 {
	total := 0.0
	for _, methods := range rm {
		for _, rate := range methods {
			total += rate
		}
	}
	return total
}

// GetComponentRate returns the sum of all method rates for a specific component
func (rm RateMap) GetComponentRate(component *ComponentInstance) float64 {
	if component == nil || rm[component] == nil {
		return 0.0
	}
	
	total := 0.0
	for _, rate := range rm[component] {
		total += rate
	}
	return total
}