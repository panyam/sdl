package viz

import (
	"bytes"
	"fmt"
)

// --- DOT Generator ---

type DotGenerator struct{}

func (g *DotGenerator) Generate(systemName string, nodes []Node, edges []Edge) (string, error) {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("digraph \"%s\" {\n", systemName))
	b.WriteString("  rankdir=LR;\n")
	b.WriteString(fmt.Sprintf("  label=\"Static Diagram for System: %s\";\n", systemName))
	b.WriteString("  node [shape=record];\n")

	for _, node := range nodes {
		b.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n(%s)\"];\n", node.ID, node.Name, node.Type))
	}

	for _, edge := range edges {
		b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", edge.FromID, edge.ToID, edge.Label))
	}
	b.WriteString("}\n")
	return b.String(), nil
}
