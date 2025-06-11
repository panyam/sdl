package viz

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

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

// --- Excalidraw Helper Structs and Methods ---

type ExcalidrawElement struct {
	ID              string          `json:"id"`
	Type            string          `json:"type"`
	X               float64         `json:"x"`
	Y               float64         `json:"y"`
	Width           float64         `json:"width"`
	Height          float64         `json:"height"`
	Angle           float64         `json:"angle,omitempty"`
	StrokeColor     string          `json:"strokeColor"`
	BackgroundColor string          `json:"backgroundColor"`
	FillStyle       string          `json:"fillStyle"`
	StrokeWidth     int             `json:"strokeWidth"`
	StrokeStyle     string          `json:"strokeStyle"`
	Roughness       int             `json:"roughness"`
	Opacity         int             `json:"opacity"`
	Seed            int64           `json:"seed"`
	Version         int             `json:"version"`
	VersionNonce    int64           `json:"versionNonce"`
	IsDeleted       bool            `json:"isDeleted,omitempty"`
	BoundElements   []*BoundElement `json:"boundElements,omitempty"`
	StartBinding    *Binding        `json:"startBinding,omitempty"`
	EndBinding      *Binding        `json:"endBinding,omitempty"`
	Points          [][]float64     `json:"points,omitempty"`
	Text            string          `json:"text,omitempty"`
	FontSize        float64         `json:"fontSize,omitempty"`
	FontFamily      int             `json:"fontFamily,omitempty"`
	TextAlign       string          `json:"textAlign,omitempty"`
	VerticalAlign   string          `json:"verticalAlign,omitempty"`
	Baseline        int             `json:"baseline,omitempty"`
	ContainerId     *string         `json:"containerId,omitempty"`
	OriginalText    string          `json:"originalText,omitempty"`
	StrokeSharpness string          `json:"strokeSharpness,omitempty"`
	StartArrowhead  *string         `json:"startArrowhead,omitempty"`
	EndArrowhead    *string         `json:"endArrowhead,omitempty"`
}

type Binding struct {
	ElementID string  `json:"elementId"`
	Focus     float64 `json:"focus,omitempty"`
	Gap       float64 `json:"gap,omitempty"`
}

type BoundElement struct{ Type, ID string }
type ExcalidrawFile struct {
	Type, Version, Source string
	Elements              []*ExcalidrawElement
	AppState              map[string]interface{}
	Files                 map[string]interface{}
}
type excalidrawScene struct {
	elements     []*ExcalidrawElement
	elementIDMap map[string]*ExcalidrawElement
	randSource   *rand.Rand
}

func newExcalidrawScene() *excalidrawScene {
	return &excalidrawScene{
		elements:     make([]*ExcalidrawElement, 0),
		elementIDMap: make(map[string]*ExcalidrawElement),
		randSource:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
func (s *excalidrawScene) newSeed() int64 { return s.randSource.Int63n(2147483646) + 1 }
func (s *excalidrawScene) newElementID(prefix string) string {
	return prefix + "_" + strconv.FormatInt(s.newSeed(), 36)
}
func (s *excalidrawScene) addElement(element *ExcalidrawElement) error {
	if element.ID == "" {
		element.ID = s.newElementID(element.Type)
	}
	s.elements = append(s.elements, element)
	s.elementIDMap[element.ID] = element
	return nil
}
func (s *excalidrawScene) getElement(id string) *ExcalidrawElement { return s.elementIDMap[id] }
func (s *excalidrawScene) addRectangle(x, y, w, h float64, label string, props, labelProps *ExcalidrawElement) (*ExcalidrawElement, *ExcalidrawElement, error) {
	rect := &ExcalidrawElement{
		Type: "rectangle", X: x, Y: y, Width: w, Height: h, StrokeColor: "#1e1e1e", BackgroundColor: "#f8f9fa",
		FillStyle: "solid", StrokeWidth: 1, StrokeStyle: "solid", Roughness: 1, StrokeSharpness: "round",
		Seed: s.newSeed(), Version: 2, VersionNonce: s.newSeed(), Opacity: 100,
	}
	s.addElement(rect)
	if label != "" {
		text, _, _ := s.addText(x+10, y+(h-24)/2, w-20, 24, label, &rect.ID, nil)
		rect.BoundElements = append(rect.BoundElements, &BoundElement{Type: "text", ID: text.ID})
	}
	return rect, nil, nil
}
func (s *excalidrawScene) addText(x, y, w, h float64, text string, containerID *string, props *ExcalidrawElement) (*ExcalidrawElement, *ExcalidrawElement, error) {
	fs := 16.0
	bl := int(fs * 0.8)
	textEl := &ExcalidrawElement{
		Type: "text", X: x, Y: y, Width: w, Height: h, Text: text, OriginalText: text, ContainerId: containerID,
		StrokeColor: "#1e1e1e", BackgroundColor: "transparent", FontSize: fs, FontFamily: 1, TextAlign: "center", VerticalAlign: "middle",
		Baseline: bl, Seed: s.newSeed(), Version: 2, VersionNonce: s.newSeed(), Opacity: 100,
	}
	s.addElement(textEl)
	return textEl, nil, nil
}
func (s *excalidrawScene) addArrow(from, to, label string, props, labelProps *ExcalidrawElement) (*ExcalidrawElement, *ExcalidrawElement, error) {
	source := s.getElement(from)
	target := s.getElement(to)
	ah := "arrow"
	arrow := &ExcalidrawElement{
		Type: "arrow", StartArrowhead: nil, EndArrowhead: &ah,
		StartBinding: &Binding{ElementID: source.ID, Focus: 0.5, Gap: 1},
		EndBinding:   &Binding{ElementID: target.ID, Focus: 0.5, Gap: 1},
		StrokeColor:  "#1e1e1e", StrokeWidth: 1, StrokeStyle: "solid", Roughness: 0, StrokeSharpness: "round",
		Seed: s.newSeed(), Version: 2, VersionNonce: s.newSeed(), Opacity: 100,
	}
	s.addElement(arrow)
	if label != "" {
		text, _, _ := s.addText(0, 0, 0, 0, label, &arrow.ID, nil)
		arrow.BoundElements = append(arrow.BoundElements, &BoundElement{Type: "text", ID: text.ID})
	}
	return arrow, nil, nil
}
func (s *excalidrawScene) toJSON() (string, error) {
	file := ExcalidrawFile{
		Type: "excalidraw", Version: "2", Source: "https://github.com/panyam/sdl",
		Elements: s.elements, AppState: map[string]interface{}{"viewBackgroundColor": "#FFFFFF"},
	}
	data, err := json.MarshalIndent(file, "", "  ")
	return string(data), err
}
