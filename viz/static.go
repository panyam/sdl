package viz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"math/rand"
	"strconv"
	"time"
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

// --- SVG Generator ---

type SvgGenerator struct{}

func (g *SvgGenerator) Generate(systemName string, nodes []Node, edges []Edge) (string, error) {
	var svg bytes.Buffer
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
		titleX, titleY, html.EscapeString(systemName)))

	for i, node := range nodes {
		nodeX := currentX
		nodeY := currentY
		centerX := nodeX + rectWidth/2
		centerY := nodeY + rectHeight/2
		nodePositions[node.ID] = struct{ X, Y, CX, CY float64 }{nodeX, nodeY, centerX, centerY}

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
	for _, edge := range edges {
		fromPos, fromOk := nodePositions[edge.FromID]
		toPos, toOk := nodePositions[edge.ToID]

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

// --- Excalidraw Generator ---

type ExcalidrawGenerator struct{}

func (g *ExcalidrawGenerator) Generate(systemName string, nodes []Node, edges []Edge) (string, error) {
	scene := newExcalidrawScene()
	layoutState := struct {
		currentX, currentY, startX, startY, elementWidth, elementHeight, gapX, gapY float64
		elementsPerRow, countInRow                                                  int
	}{
		startX: 50.0, startY: 50.0, currentX: 50.0, currentY: 50.0,
		elementWidth: 200.0, elementHeight: 80.0,
		gapX: 100.0, gapY: 100.0,
		elementsPerRow: 3, countInRow: 0,
	}
	sdlNodeToExcalidrawRectID := make(map[string]string)

	for _, node := range nodes {
		labelText := fmt.Sprintf("%s\n(%s)", node.Name, node.Type)
		rect, _, err := scene.addRectangle(layoutState.currentX, layoutState.currentY, layoutState.elementWidth, layoutState.elementHeight, labelText, nil, nil)
		if err != nil {
			return "", fmt.Errorf("error adding SDL node %s to Excalidraw scene: %w", node.ID, err)
		}
		sdlNodeToExcalidrawRectID[node.ID] = rect.ID

		layoutState.countInRow++
		if layoutState.countInRow >= layoutState.elementsPerRow {
			layoutState.currentX = layoutState.startX
			layoutState.currentY += layoutState.elementHeight + layoutState.gapY
			layoutState.countInRow = 0
		} else {
			layoutState.currentX += layoutState.elementWidth + layoutState.gapX
		}
	}

	for _, edge := range edges {
		fromRectID, fromOk := sdlNodeToExcalidrawRectID[edge.FromID]
		toRectID, toOk := sdlNodeToExcalidrawRectID[edge.ToID]
		if !fromOk || !toOk {
			return "", fmt.Errorf("could not find Excalidraw ID for SDL node %s or %s for edge", edge.FromID, edge.ToID)
		}
		_, _, err := scene.addArrow(fromRectID, toRectID, edge.Label, nil, nil)
		if err != nil {
			return "", fmt.Errorf("error adding SDL edge from %s to %s (label: %s) to Excalidraw scene: %w", edge.FromID, edge.ToID, edge.Label, err)
		}
	}
	return scene.toJSON()
}

// ... (Excalidraw helper structs and methods) ...
type excalidrawElement struct {
	ID, Type, StrokeColor, BackgroundColor, FillStyle, StrokeStyle, StrokeSharpness string
	X, Y, Width, Height, Angle, FontSize, Baseline                                float64
	StrokeWidth, Roughness, Opacity, FontFamily                                   int
	Seed, VersionNonce                                                            int64
	Version                                                                       int
	IsDeleted                                                                     bool
	BoundElements                                                                 []*boundElement
	StartBinding, EndBinding                                                      *binding
	Points                                                                        [][]float64
	Text, OriginalText, ContainerId                                               *string
	VerticalAlign, TextAlign                                                      string
	StartArrowhead, EndArrowhead                                                  *string
}
type binding struct{ ElementID string; Focus, Gap float64 }
type boundElement struct{ Type, ID string }
type excalidrawFile struct {
	Type, Version, Source string
	Elements              []*excalidrawElement
	AppState              map[string]interface{}
	Files                 map[string]interface{}
}
type excalidrawScene struct {
	elements     []*excalidrawElement
	elementIDMap map[string]*excalidrawElement
	randSource   *rand.Rand
}
func newExcalidrawScene() *excalidrawScene {
	return &excalidrawScene{
		elements:     make([]*excalidrawElement, 0),
		elementIDMap: make(map[string]*excalidrawElement),
		randSource:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
func (s *excalidrawScene) newSeed() int64 { return s.randSource.Int63n(2147483646) + 1 }
func (s *excalidrawScene) newElementID(prefix string) string {
	return prefix + "_" + strconv.FormatInt(s.newSeed(), 36)
}
func (s *excalidrawScene) addElement(element *excalidrawElement) error {
	if element.ID == "" {
		element.ID = s.newElementID(element.Type)
	}
	s.elements = append(s.elements, element)
	s.elementIDMap[element.ID] = element
	return nil
}
func (s *excalidrawScene) getElement(id string) *excalidrawElement { return s.elementIDMap[id] }
func (s *excalidrawScene) addRectangle(x, y, w, h float64, label string, props, labelProps *excalidrawElement) (*excalidrawElement, *excalidrawElement, error) {
	rect := &excalidrawElement{
		Type: "rectangle", X: x, Y: y, Width: w, Height: h, StrokeColor: "#1e1e1e", BackgroundColor: "#f8f9fa",
		FillStyle: "solid", StrokeWidth: 1, StrokeStyle: "solid", Roughness: 1, StrokeSharpness: "round",
		Seed: s.newSeed(), Version: 2, VersionNonce: s.newSeed(), Opacity: 100,
	}
	s.addElement(rect)
	if label != "" {
		text, _, _ := s.addText(x+10, y+(h-24)/2, w-20, 24, label, &rect.ID, nil)
		rect.BoundElements = append(rect.BoundElements, &boundElement{Type: "text", ID: text.ID})
	}
	return rect, nil, nil
}
func (s *excalidrawScene) addText(x, y, w, h float64, text string, containerID *string, props *excalidrawElement) (*excalidrawElement, *excalidrawElement, error) {
	fs := 16.0
	bl := fs * 0.8
	textEl := &excalidrawElement{
		Type: "text", X: x, Y: y, Width: w, Height: h, Text: &text, OriginalText: &text, ContainerId: containerID,
		StrokeColor: "#1e1e1e", BackgroundColor: "transparent", FontSize: fs, FontFamily: 1, TextAlign: "center", VerticalAlign: "middle",
		Baseline: bl, Seed: s.newSeed(), Version: 2, VersionNonce: s.newSeed(), Opacity: 100,
	}
	s.addElement(textEl)
	return textEl, nil, nil
}
func (s *excalidrawScene) addArrow(from, to, label string, props, labelProps *excalidrawElement) (*excalidrawElement, *excalidrawElement, error) {
	source := s.getElement(from)
	target := s.getElement(to)
	ah := "arrow"
	arrow := &excalidrawElement{
		Type: "arrow", X: 0, Y: 0, Width: 0, Height: 0, EndArrowhead: &ah,
		StartBinding: &binding{ElementID: source.ID, Focus: 0.5, Gap: 1},
		EndBinding:   &binding{ElementID: target.ID, Focus: 0.5, Gap: 1},
		StrokeColor: "#1e1e1e", StrokeWidth: 1, StrokeStyle: "solid", Roughness: 0, StrokeSharpness: "round",
		Seed: s.newSeed(), Version: 2, VersionNonce: s.newSeed(), Opacity: 100,
	}
	s.addElement(arrow)
	if label != "" {
		text, _, _ := s.addText(0, 0, 0, 0, label, &arrow.ID, nil)
		arrow.BoundElements = append(arrow.BoundElements, &boundElement{Type: "text", ID: text.ID})
	}
	return arrow, nil, nil
}
func (s *excalidrawScene) toJSON() (string, error) {
	file := excalidrawFile{
		Type: "excalidraw", Version: 2, Source: "https://github.com/panyam/sdl",
		Elements: s.elements, AppState: map[string]interface{}{"viewBackgroundColor": "#FFFFFF"},
	}
	data, err := json.MarshalIndent(file, "", "  ")
	return string(data), err
}
