package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

// MockFlowComponent is a mock component for testing flow patterns
type MockFlowComponent struct {
	NWBase[*MockFlowComponentImpl]
}

// MockFlowComponentImpl is the underlying implementation
type MockFlowComponentImpl struct {
	Name     string
	Outflows map[string]float64
}

func (m *MockFlowComponentImpl) Init() {
	// No initialization needed for mock
}

func (m *MockFlowComponentImpl) GetFlowPattern(method string, inputRate float64) components.FlowPattern {
	// Scale outflows by input rate
	scaledOutflows := make(map[string]float64)
	for target, multiplier := range m.Outflows {
		scaledOutflows[target] = inputRate * multiplier
	}
	
	return components.FlowPattern{
		Outflows:    scaledOutflows,
		SuccessRate: 1.0,
		ServiceTime: 0.001,
	}
}

// NewMockFlowComponent creates a new mock flow component
func NewMockFlowComponent(name string, outflows map[string]float64) *MockFlowComponent {
	impl := &MockFlowComponentImpl{
		Name:     name,
		Outflows: outflows,
	}
	return &MockFlowComponent{NWBase: NewNWBase(name, impl)}
}

// Process is a mock method that can be called
func (m *MockFlowComponent) Process() decl.Value {
	// Return a simple boolean success
	return decl.BoolValue(true)
}