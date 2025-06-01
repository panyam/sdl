package commands

// DiagramNode represents an instance in the static diagram
type DiagramNode struct {
	ID   string // Instance name
	Name string // Instance name for display
	Type string // Component type for display
}

// DiagramEdge represents a connection between instances
type DiagramEdge struct {
	FromID string
	ToID   string
	Label  string
}
