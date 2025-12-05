package viz

import (
	"bytes"
	"fmt"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// --- Mermaid Static Generator ---

type MermaidStaticGenerator struct{}

func (g *MermaidStaticGenerator) Generate(diagram *protos.SystemDiagram) (string, error) {
	var b bytes.Buffer
	b.WriteString("graph TD;\n")
	b.WriteString(fmt.Sprintf("  subgraph System %s\n", diagram.SystemName))

	for _, node := range diagram.Nodes {
		b.WriteString(fmt.Sprintf("    %s[\"%s (%s)\"];\n", node.Id, node.Name, node.Type))
	}

	for _, edge := range diagram.Edges {
		b.WriteString(fmt.Sprintf("    %s -- \"%s\" --> %s;\n", edge.FromId, edge.Label, edge.ToId))
	}
	b.WriteString("  end\n")
	return b.String(), nil
}
