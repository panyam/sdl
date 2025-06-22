package viz

import (
	"bytes"
	"fmt"

	protos "github.com/panyam/sdl/gen/go/sdl/v1"
)

// --- DOT Generator ---

type DotGenerator struct{}

func (g *DotGenerator) Generate(diagram *protos.SystemDiagram) (string, error) {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("digraph \"%s\" {\n", diagram.SystemName))
	b.WriteString("  rankdir=LR;\n")
	b.WriteString(fmt.Sprintf("  label=\"Static Diagram for System: %s\";\n", diagram.SystemName))
	b.WriteString("  node [shape=record];\n")

	for _, node := range diagram.Nodes {
		b.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n(%s)\"];\n", node.Id, node.Name, node.Type))
	}

	for _, edge := range diagram.Edges {
		b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", edge.FromId, edge.ToId, edge.Label))
	}
	b.WriteString("}\n")
	return b.String(), nil
}
