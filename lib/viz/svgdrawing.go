package viz

import (
	"bytes"
	"fmt"
	"html"
	"math"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// --- SVG Generator ---

type SvgGenerator struct{}

func (g *SvgGenerator) Generate(diagram *protos.SystemDiagram) (string, error) {
	var svg bytes.Buffer
	// ... (SVG generation code remains the same as before)
	canvasWidth := 1024
	canvasHeight := 768
	padding := 20.0
	rectWidth := 180.0
	rectHeight := 70.0
	gapX := 60.0
	gapY := 70.0
	fontSize := 14.0
	lineHeight := fontSize * 1.2
	startX := padding
	startY := padding + 30
	currentX := startX
	currentY := startY
	elementsPerRow := 4
	maxRowWidth := padding
	nodePositions := make(map[string]struct{ X, Y, CX, CY float64 })

	svg.WriteString(fmt.Sprintf("<svg width=\"%d\" height=\"%d\" xmlns=\"http://www.w3.org/2000/svg\">\n", canvasWidth, canvasHeight))
	svg.WriteString("  <style>\n")
	svg.WriteString("    .node-rect { stroke: #333; stroke-width: 1.5px; fill: #f8f9fa; }\n")
	svg.WriteString(fmt.Sprintf("    .node-text { font-family: Arial, sans-serif; font-size: %.1fpx; fill: #212529; text-anchor: middle; }\n", fontSize))
	svg.WriteString("    .edge-line { stroke: #333; stroke-width: 1.5px; fill: none; }\n")
	svg.WriteString(fmt.Sprintf("    .edge-label { font-family: Arial, sans-serif; font-size: %.1fpx; fill: #495057; text-anchor: middle; }\n", fontSize*0.9))
	svg.WriteString(fmt.Sprintf("    .diagram-title { font-family: Arial, sans-serif; font-size: %.1fpx; font-weight: bold; text-anchor: middle; }\n", fontSize*1.2))
	svg.WriteString("  </style>\n")
	svg.WriteString("  <defs>\n")
	svg.WriteString("    <marker id=\"arrowhead\" markerWidth=\"10\" markerHeight=\"7\" refX=\"0\" refY=\"3.5\" orient=\"auto\">\n")
	svg.WriteString("      <polygon points=\"0 0, 10 3.5, 0 7\" fill=\"#333\" />\n")
	svg.WriteString("    </marker>\n")
	svg.WriteString("  </defs>\n")
	titleX := float64(canvasWidth) / 2.0
	titleY := padding + fontSize
	svg.WriteString(fmt.Sprintf("  <text x=\"%.1f\" y=\"%.1f\" class=\"diagram-title\">Static Diagram for System: %s</text>\n",
		titleX, titleY, html.EscapeString(diagram.SystemName)))

	for i, node := range diagram.Nodes {
		nodeX := currentX
		nodeY := currentY
		centerX := nodeX + rectWidth/2
		centerY := nodeY + rectHeight/2
		nodePositions[node.Id] = struct{ X, Y, CX, CY float64 }{nodeX, nodeY, centerX, centerY}

		svg.WriteString(fmt.Sprintf("  <rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\" rx=\"5\" ry=\"5\" class=\"node-rect\" />\n",
			nodeX, nodeY, rectWidth, rectHeight))
		textY1 := centerY - lineHeight/2 + fontSize/2.5
		textY2 := centerY + lineHeight/2 + fontSize/2.5
		svg.WriteString(fmt.Sprintf("  <text x=\"%.1f\" y=\"%.1f\" class=\"node-text\">%s</text>\n",
			centerX, textY1, html.EscapeString(node.Name)))
		svg.WriteString(fmt.Sprintf("  <text x=\"%.1f\" y=\"%.1f\" class=\"node-text\" style=\"font-size:%.1fpx; fill: #555;\">%s</text>\n",
			centerX, textY2, fontSize*0.85, html.EscapeString("("+node.Type+")")))

		currentX += rectWidth + gapX
		if (i+1)%elementsPerRow == 0 {
			if currentX > maxRowWidth {
				maxRowWidth = currentX - gapX
			}
			currentX = startX
			currentY += rectHeight + gapY
		}
	}
	for _, edge := range diagram.Edges {
		fromPos, fromOk := nodePositions[edge.FromId]
		toPos, toOk := nodePositions[edge.ToId]

		if fromOk && toOk {
			dx := toPos.CX - fromPos.CX
			dy := toPos.CY - fromPos.CY
			dist := math.Sqrt(dx*dx + dy*dy)
			arrowHeadOffset := 7.0
			endX := toPos.CX
			endY := toPos.CY

			if dist > arrowHeadOffset {
				endX = toPos.CX - (dx/dist)*arrowHeadOffset
				endY = toPos.CY - (dy/dist)*arrowHeadOffset
			}

			svg.WriteString(fmt.Sprintf("  <line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"edge-line\" marker-end=\"url(#arrowhead)\" />\n",
				fromPos.CX, fromPos.CY, endX, endY))
			if edge.Label != "" {
				labelX := (fromPos.CX + toPos.CX) / 2
				labelY := (fromPos.CY+toPos.CY)/2 - 5
				svg.WriteString(fmt.Sprintf("  <text x=\"%.1f\" y=\"%.1f\" class=\"edge-label\" style=\"fill: #007bff;\">%s</text>\n",
					labelX, labelY, html.EscapeString(edge.Label)))
			}
		}
	}
	svg.WriteString("</svg>")

	return svg.String(), nil
}
