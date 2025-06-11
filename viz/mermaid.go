package viz

import (
	"bytes"
	"fmt"
)

// --- Mermaid Static Generator ---

type MermaidStaticGenerator struct{}

func (g *MermaidStaticGenerator) Generate(systemName string, nodes []Node, edges []Edge) (string, error) {
	var b bytes.Buffer
	b.WriteString("graph TD;\n")
	b.WriteString(fmt.Sprintf("  subgraph System %s\n", systemName))

	for _, node := range nodes {
		b.WriteString(fmt.Sprintf("    %s[\"%s (%s)\"];\n", node.ID, node.Name, node.Type))
	}

	for _, edge := range edges {
		b.WriteString(fmt.Sprintf("    %s -- \"%s\" --> %s;\n", edge.FromID, edge.Label, edge.ToID))
	}
	b.WriteString("  end\n")
	return b.String(), nil
}
