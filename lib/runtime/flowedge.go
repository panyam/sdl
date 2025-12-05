package runtime

// FlowEdge represents a flow connection between two component methods
type FlowEdge struct {
	FromComponent *ComponentInstance
	FromMethod    string
	ToComponent   *ComponentInstance
	ToMethod      string
	Rate          float64
}

// FlowEdgeMap tracks all flow edges in the system
type FlowEdgeMap struct {
	edges []FlowEdge
}

// NewFlowEdgeMap creates a new empty FlowEdgeMap
func NewFlowEdgeMap() *FlowEdgeMap {
	return &FlowEdgeMap{
		edges: make([]FlowEdge, 0),
	}
}

// AddEdge adds a flow edge to the map
func (fem *FlowEdgeMap) AddEdge(fromComp *ComponentInstance, fromMethod string, toComp *ComponentInstance, toMethod string, rate float64) {
	if fromComp == nil || toComp == nil || fromMethod == "" || toMethod == "" || rate <= 0 {
		return
	}
	
	fem.edges = append(fem.edges, FlowEdge{
		FromComponent: fromComp,
		FromMethod:    fromMethod,
		ToComponent:   toComp,
		ToMethod:      toMethod,
		Rate:          rate,
	})
}

// GetEdges returns all flow edges
func (fem *FlowEdgeMap) GetEdges() []FlowEdge {
	return fem.edges
}

// Clear removes all edges
func (fem *FlowEdgeMap) Clear() {
	fem.edges = fem.edges[:0]
}