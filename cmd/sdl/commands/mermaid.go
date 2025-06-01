package commands

import (
	"bytes"
	"fmt"

	_ "github.com/panyam/sdl/decl"
)

func generateMermaidOutput(systemName string, nodes []DiagramNode, edges []DiagramEdge) string {
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
	return b.String()
}
