package decl

import (
	"fmt"
	"strings"
)

// OpNode represents a node in the evaluated operator tree.
// This tree represents the computation to be performed, before
// the actual probabilistic combination/reduction occurs.
type OpNode interface {
	opNode() // Marker method
	String() string
}

// --- Concrete OpNode Types ---

// LeafNode represents a terminal value in the tree, typically
// the result of evaluating a literal or a component call.
// It holds the dual-track VarState.
type LeafNode struct {
	State *VarState // Holds ValueOutcome + LatencyOutcome
}

func (n *LeafNode) opNode() {}
func (n *LeafNode) String() string {
	// Basic representation, can be enhanced
	valStr := "<nil>"
	latStr := "<nil>"
	if n.State != nil {
		if n.State.ValueOutcome != nil {
			// Try to represent the outcome simply for debugging
			if vo, ok := n.State.ValueOutcome.(fmt.Stringer); ok {
				valStr = vo.String()
			} else {
				valStr = fmt.Sprintf("%T", n.State.ValueOutcome)
			}
		}
		if n.State.LatencyOutcome != nil {
			if lo, ok := n.State.LatencyOutcome.(fmt.Stringer); ok {
				latStr = lo.String()
			} else {
				latStr = fmt.Sprintf("%T", n.State.LatencyOutcome)
			}
		}
	}
	return fmt.Sprintf("Leaf(Value: %s, Latency: %s)", valStr, latStr)
}

// NilNode represents an operation with no return value or side effect
// relevant to the final result (e.g., a 'let' statement).
type NilNode struct{}

func (n *NilNode) opNode()        {}
func (n *NilNode) String() string { return "Nil" }

// Singleton instance for NilNode
var theNilNode = &NilNode{}

// SequenceNode represents a series of operations executed sequentially.
// Only non-Nil results are typically included.
type SequenceNode struct {
	Steps []OpNode
}

func (n *SequenceNode) opNode() {}
func (n *SequenceNode) String() string {
	var sb strings.Builder
	sb.WriteString("Seq[")
	for i, step := range n.Steps {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(step.String())
	}
	sb.WriteString("]")
	return sb.String()
}

// BinaryOpNode represents a binary operation (e.g., +, &&, ==).
// The actual calculation is deferred until the Tree Evaluator stage.
type BinaryOpNode struct {
	Op    string // Operator symbol (e.g., "+", "&&", ">")
	Left  OpNode
	Right OpNode
}

func (n *BinaryOpNode) opNode() {}
func (n *BinaryOpNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Op, n.Right.String())
}

// IfChoiceNode represents a conditional branching point (if-then-else).
// The condition node will be evaluated first by the Tree Evaluator,
// and based on its boolean outcome (potentially probabilistic), the
// appropriate branch(es) will be evaluated.
type IfChoiceNode struct {
	Condition OpNode
	Then      OpNode
	Else      OpNode // Can be *NilNode if no else branch exists
}

func (n *IfChoiceNode) opNode() {}
func (n *IfChoiceNode) String() string {
	elseStr := "Nil"
	if n.Else != nil {
		elseStr = n.Else.String()
	}
	return fmt.Sprintf("If(%s) Then:{%s} Else:{%s}", n.Condition.String(), n.Then.String(), elseStr)
}

// TODO: Add UnaryOpNode, ChoiceNode, etc.
