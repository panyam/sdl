package viz

import (
	"bytes"
	"fmt"
	"strings"

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
		label := fmt.Sprintf("%s\\n(%s)", node.Name, node.Type)
		if node.Traffic != "" && node.Traffic != "0 rps" {
			// Add traffic info with proper escaping for newlines
			trafficInfo := node.Traffic
			// Replace newlines with \n for DOT format
			trafficInfo = strings.ReplaceAll(trafficInfo, "\n", "\\n")
			label += fmt.Sprintf("\\n%s", trafficInfo)
		}
		b.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\"];\n", node.Id, label))
	}

	for _, edge := range diagram.Edges {
		b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", edge.FromId, edge.ToId, edge.Label))
	}
	b.WriteString("}\n")
	return b.String(), nil
}
