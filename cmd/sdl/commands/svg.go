package commands

import (
	"bytes"
	"fmt"
	"html"
	"math"
)

// Helper to escape strings for SVG text content
func escapeSVGText(s string) string {
	return html.EscapeString(s)
}

func generateSvgOutput(systemName string, nodes []DiagramNode, edges []DiagramEdge) (string, error) {
	var svg bytes.Buffer

	// SVG parameters (can be made configurable later)
	canvasWidth := 1024
	canvasHeight := 768 // Adjusted dynamically later if needed
	padding := 20.0
	rectWidth := 180.0
	rectHeight := 70.0
	gapX := 60.0
	gapY := 70.0
	fontSize := 14.0
	lineHeight := fontSize * 1.2

	// Layout calculation (simple row-based)
	startX := padding
	startY := padding + 30 // Extra space for title
	currentX := startX
	currentY := startY
	elementsPerRow := 4
	maxRowWidth := padding
	nodePositions := make(map[string]struct{ X, Y, CX, CY float64 })

	svg.WriteString(fmt.Sprintf("<svg width=\"%d\" height=\"%d\" xmlns=\"http://www.w3.org/2000/svg\">\n", canvasWidth, canvasHeight))
	// Define a style block (optional, can use inline styles too)
	svg.WriteString("  <style>\n")
	svg.WriteString("    .node-rect { stroke: #333; stroke-width: 1.5px; fill: #f8f9fa; }\n")
	svg.WriteString(fmt.Sprintf("    .node-text { font-family: Arial, sans-serif; font-size: %.1fpx; fill: #212529; text-anchor: middle; }\n", fontSize))
	svg.WriteString("    .edge-line { stroke: #333; stroke-width: 1.5px; fill: none; }\n")
	svg.WriteString(fmt.Sprintf("    .edge-label { font-family: Arial, sans-serif; font-size: %.1fpx; fill: #495057; text-anchor: middle; }\n", fontSize*0.9))
	svg.WriteString(fmt.Sprintf("    .diagram-title { font-family: Arial, sans-serif; font-size: %.1fpx; font-weight: bold; text-anchor: middle; }\n", fontSize*1.2))
	svg.WriteString("  </style>\n")

	// Define marker for arrowhead
	svg.WriteString("  <defs>\n")
	svg.WriteString("    <marker id=\"arrowhead\" markerWidth=\"10\" markerHeight=\"7\" refX=\"0\" refY=\"3.5\" orient=\"auto\">\n")
	svg.WriteString("      <polygon points=\"0 0, 10 3.5, 0 7\" fill=\"#333\" />\n")
	svg.WriteString("    </marker>\n")
	svg.WriteString("  </defs>\n")

	// Diagram Title
	titleX := float64(canvasWidth) / 2.0
	titleY := padding + fontSize
	svg.WriteString(fmt.Sprintf("  <text x=\"%.1f\" y=\"%.1f\" class=\"diagram-title\">Static Diagram for System: %s</text>\n",
		titleX, titleY, escapeSVGText(systemName)))

	// First pass: Draw nodes and store their center positions
	for i, node := range nodes {
		nodeX := currentX
		nodeY := currentY
		centerX := nodeX + rectWidth/2
		centerY := nodeY + rectHeight/2
		nodePositions[node.ID] = struct{ X, Y, CX, CY float64 }{nodeX, nodeY, centerX, centerY}

		svg.WriteString(fmt.Sprintf("  <rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\" rx=\"5\" ry=\"5\" class=\"node-rect\" />\n",
			nodeX, nodeY, rectWidth, rectHeight))

		// Text for node (name and type on two lines)
		textY1 := centerY - lineHeight/2 + fontSize/2.5 // Adjust for baseline of first line
		textY2 := centerY + lineHeight/2 + fontSize/2.5 // Adjust for baseline of second line

		svg.WriteString(fmt.Sprintf("  <text x=\"%.1f\" y=\"%.1f\" class=\"node-text\">%s</text>\n",
			centerX, textY1, escapeSVGText(node.Name)))
		svg.WriteString(fmt.Sprintf("  <text x=\"%.1f\" y=\"%.1f\" class=\"node-text\" style=\"font-size:%.1fpx; fill: #555;\">%s</text>\n",
			centerX, textY2, fontSize*0.85, escapeSVGText("("+node.Type+")")))

		currentX += rectWidth + gapX
		if (i+1)%elementsPerRow == 0 {
			if currentX > maxRowWidth {
				maxRowWidth = currentX - gapX
			} // Store max width of this row
			currentX = startX
			currentY += rectHeight + gapY
		}
	}

	// Adjust canvas height if needed (simple approximation)
	actualHeight := currentY
	if countInRow := len(nodes) % elementsPerRow; countInRow != 0 || len(nodes) == 0 {
		actualHeight += rectHeight + padding
	} else {
		actualHeight += padding // Only add padding if last row was full or no nodes
	}
	if actualHeight < startY+rectHeight+padding { // ensure min height for one row
		actualHeight = startY + rectHeight + padding
	}

	// Update canvasWidth based on maxRowWidth
	if maxRowWidth > float64(canvasWidth) {
		canvasWidth = int(maxRowWidth + padding)
	}
	// We need to rewrite the SVG tag if height/width changed, or use JS to update it.
	// For simplicity, we will just ensure the initial canvasHeight is large enough for typical cases.
	// A more robust solution would be to generate elements, then determine bounds, then write SVG tag.

	// Second pass: Draw edges
	for _, edge := range edges {
		fromPos, fromOk := nodePositions[edge.FromID]
		toPos, toOk := nodePositions[edge.ToID]

		if fromOk && toOk {
			// Simple line from center to center for now
			// For arrowhead, line needs to end slightly before the target's center
			dx := toPos.CX - fromPos.CX
			dy := toPos.CY - fromPos.CY
			dist := math.Sqrt(dx*dx + dy*dy)
			arrowHeadOffset := 7.0 // Approximate offset for arrowhead to not overlap center point
			endX := toPos.CX
			endY := toPos.CY

			if dist > arrowHeadOffset {
				endX = toPos.CX - (dx/dist)*arrowHeadOffset
				endY = toPos.CY - (dy/dist)*arrowHeadOffset
			}

			svg.WriteString(fmt.Sprintf("  <line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"edge-line\" marker-end=\"url(#arrowhead)\" />\n",
				fromPos.CX, fromPos.CY, endX, endY))

			// Edge label
			if edge.Label != "" {
				labelX := (fromPos.CX + toPos.CX) / 2
				labelY := (fromPos.CY+toPos.CY)/2 - 5 // Offset slightly above the line
				svg.WriteString(fmt.Sprintf("  <text x=\"%.1f\" y=\"%.1f\" class=\"edge-label\" style=\"fill: #007bff;\">%s</text>\n",
					labelX, labelY, escapeSVGText(edge.Label)))
			}
		}
	}

	svg.WriteString("</svg>")
	return svg.String(), nil
}
