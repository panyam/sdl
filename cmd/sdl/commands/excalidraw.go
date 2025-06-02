package commands

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// ExcalidrawElement defines the structure for an individual Excalidraw element.
// It includes common fields and uses `omitempty` for optional ones.
type ExcalidrawElement struct {
	ID              string      `json:"id"`
	Type            string      `json:"type"`
	X               float64     `json:"x"`
	Y               float64     `json:"y"`
	Width           float64     `json:"width"`
	Height          float64     `json:"height"`
	Angle           float64     `json:"angle,omitempty"`
	StrokeColor     string      `json:"strokeColor,omitempty"`
	BackgroundColor string      `json:"backgroundColor,omitempty"`
	FillStyle       string      `json:"fillStyle,omitempty"`
	StrokeWidth     int         `json:"strokeWidth,omitempty"`
	StrokeStyle     string      `json:"strokeStyle,omitempty"`
	Roughness       int         `json:"roughness,omitempty"`
	Opacity         int         `json:"opacity,omitempty"` // Default is 100
	Seed            int64       `json:"seed"`
	Version         int         `json:"version"`
	VersionNonce    int64       `json:"versionNonce"`
	IsDeleted       bool        `json:"isDeleted,omitempty"`
	BoundElements   []*BoundElement `json:"boundElements,omitempty"`
	StartBinding    *Binding    `json:"startBinding,omitempty"`
	EndBinding      *Binding    `json:"endBinding,omitempty"`
	Points          [][]float64 `json:"points,omitempty"`
	Text            string      `json:"text,omitempty"`
	FontSize        float64     `json:"fontSize,omitempty"`
	FontFamily      int         `json:"fontFamily,omitempty"`
	TextAlign       string      `json:"textAlign,omitempty"`
	VerticalAlign   string      `json:"verticalAlign,omitempty"`
	Baseline        int         `json:"baseline,omitempty"`
	ContainerId     *string     `json:"containerId,omitempty"`
	OriginalText    string      `json:"originalText,omitempty"`
	StrokeSharpness string      `json:"strokeSharpness,omitempty"`
	StartArrowhead  *string     `json:"startArrowhead,omitempty"`
	EndArrowhead    *string     `json:"endArrowhead,omitempty"`
}

// Binding defines how an arrow binds to another element.
type Binding struct {
	ElementID string  `json:"elementId"`
	Focus     float64 `json:"focus,omitempty"` // How much into the shape, default usually 0.5 (center-ish)
	Gap       float64 `json:"gap,omitempty"`   // Default is 1 for arrows
}

// BoundElement links an element (like text) to its container or vice-versa.
type BoundElement struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// ExcalidrawFile is the top-level structure for an .excalidraw file.
type ExcalidrawFile struct {
	Type     string               `json:"type"`
	Version  int                  `json:"version"`
	Source   string               `json:"source,omitempty"`
	Elements []*ExcalidrawElement `json:"elements"`
	AppState map[string]interface{} `json:"appState,omitempty"`
	Files    map[string]interface{} `json:"files,omitempty"`
}

// ExcalidrawScene manages a collection of Excalidraw elements.
type ExcalidrawScene struct {
	elements     []*ExcalidrawElement
	elementIDMap map[string]*ExcalidrawElement // Maps Excalidraw-generated ID to element
	randSource   *rand.Rand
}

// NewExcalidrawScene creates a new, empty Excalidraw scene.
func NewExcalidrawScene() *ExcalidrawScene {
	return &ExcalidrawScene{
		elements:     make([]*ExcalidrawElement, 0),
		elementIDMap: make(map[string]*ExcalidrawElement),
		randSource:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *ExcalidrawScene) newSeed() int64 {
	return s.randSource.Int63n(2147483646) + 1
}

func (s *ExcalidrawScene) newElementID(prefix string) string {
	return prefix + "_" + strconv.FormatInt(s.newSeed(), 36) + "_" + strconv.FormatInt(s.randSource.Int63n(1000), 36)
}

// AddElement adds a pre-configured ExcalidrawElement to the scene.
// If element.ID is empty, a new one is generated.
func (s *ExcalidrawScene) AddElement(element *ExcalidrawElement) error {
	if element == nil {
		return fmt.Errorf("cannot add nil element")
	}
	if element.ID == "" {
		prefix := element.Type
		if prefix == "" {
			prefix = "el"
		}
		element.ID = s.newElementID(prefix)
	}
	if _, exists := s.elementIDMap[element.ID]; exists {
		return fmt.Errorf("element with ID '%s' already exists in the scene", element.ID)
	}

	// Ensure common defaults if not set
	if element.Seed == 0 {
		element.Seed = s.newSeed()
	}
	if element.Version == 0 {
		element.Version = 2 // Excalidraw default version for elements
	}
	if element.VersionNonce == 0 {
		element.VersionNonce = s.newSeed() // Or time based for more uniqueness
	}
	if element.Opacity == 0 {
        element.Opacity = 100 // Default to fully opaque if not specified
    }

	s.elements = append(s.elements, element)
	s.elementIDMap[element.ID] = element
	return nil
}

// AddRectangle creates a rectangle element and an optional bound text label.
func (s *ExcalidrawScene) AddRectangle(x, y, width, height float64, labelText string, rectProps *ExcalidrawElement, labelTextProps *ExcalidrawElement) (*ExcalidrawElement, *ExcalidrawElement, error) {
	rect := &ExcalidrawElement{
		Type:            "rectangle",
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		StrokeColor:     "#1e1e1e",
		BackgroundColor: "#f8f9fa",
		FillStyle:       "solid",
		StrokeWidth:     1,
		StrokeStyle:     "solid",
		Roughness:       0,
		StrokeSharpness: "round",
		BoundElements:   make([]*BoundElement, 0),
	}
	if rectProps != nil { // Apply overrides
		// TODO: selectively override fields, not just Type, X, Y, etc.
		if rectProps.StrokeColor != "" { rect.StrokeColor = rectProps.StrokeColor }
		// ... other properties
	}
	if err := s.AddElement(rect); err != nil {
		return nil, nil, err
	}

	var textLabel *ExcalidrawElement
	if labelText != "" {
		defaultFontSize := 16.0
		lineHeightFactor := 1.2
		textHeight := defaultFontSize * lineHeightFactor * float64(len(labelText)/20+1) // rough estimate

		textLabel = &ExcalidrawElement{
			Type:            "text",
			X:               x + 10, // Position will be relative to container anyway
			Y:               y + (height-textHeight)/2,
			Width:           width - 20,
			Height:          textHeight,
			StrokeColor:     "#1e1e1e",
			BackgroundColor: "transparent",
			Text:            labelText,
			OriginalText:    labelText,
			FontSize:        defaultFontSize,
			FontFamily:      1,
			TextAlign:       "center",
			VerticalAlign:   "middle",
			Baseline:        int(defaultFontSize * 0.8),
			ContainerId:     &rect.ID,
		}
		if labelTextProps != nil { // Apply overrides
			// TODO: selectively override fields
		}
		if err := s.AddElement(textLabel); err != nil {
			// Potentially remove 'rect' if adding its label fails? Or leave as is?
			return rect, nil, err
		}
		rect.BoundElements = append(rect.BoundElements, &BoundElement{Type: "text", ID: textLabel.ID})
	}
	return rect, textLabel, nil
}

// AddArrow creates an arrow between two elements, with an optional label.
func (s *ExcalidrawScene) AddArrow(sourceElementID, targetElementID string, labelText string, arrowProps *ExcalidrawElement, labelTextProps *ExcalidrawElement) (*ExcalidrawElement, *ExcalidrawElement, error) {
	// Validate source and target elements exist (optional, bindings might handle it)
	// if _, ok := s.elementIdMap[sourceElementID]; !ok {
	// 	return nil, nil, fmt.Errorf("source element ID '%s' for arrow not found in scene", sourceElementID)
	// }
	// if _, ok := s.elementIdMap[targetElementID]; !ok {
	// 	return nil, nil, fmt.Errorf("target element ID '%s' for arrow not found in scene", targetElementID)
	// }

	arrowHeadStyle := "arrow"
	arrow := &ExcalidrawElement{
		Type:            "arrow",
		X:               0,
		Y:               0,
		Width:           0,
		Height:          0,
		StrokeColor:     "#1e1e1e",
		BackgroundColor: "transparent",
		StrokeWidth:     1,
		StrokeStyle:     "solid",
		Roughness:       0,
		StrokeSharpness: "round",
		Points:          [][]float64{{0, 0}, {1, 1}},
		EndArrowhead:    &arrowHeadStyle,
		StartBinding:    &Binding{ElementID: sourceElementID, Focus: 0.5, Gap: 1},
		EndBinding:      &Binding{ElementID: targetElementID, Focus: 0.5, Gap: 1},
		BoundElements:   make([]*BoundElement, 0),
	}
	if arrowProps != nil { // Apply overrides
		// TODO: selectively override fields
	}
	if err := s.AddElement(arrow); err != nil {
		return nil, nil, err
	}

	var textLabel *ExcalidrawElement
	if labelText != "" {
		defaultFontSize := 14.0
		lineHeightFactor := 1.2
		textLabel = &ExcalidrawElement{
			Type:            "text",
			X:               0,
			Y:               0,
			Width:           float64(len(labelText))*defaultFontSize*0.6 + 10,
			Height:          defaultFontSize * lineHeightFactor,
			StrokeColor:     "#343a40",
			BackgroundColor: "transparent",
			Text:            labelText,
			OriginalText:    labelText,
			FontSize:        defaultFontSize,
			FontFamily:      1,
			TextAlign:       "center",
			VerticalAlign:   "middle",
			Baseline:        int(defaultFontSize * 0.8),
			ContainerId:     &arrow.ID,
		}
		if labelTextProps != nil { // Apply overrides
			// TODO: selectively override fields
		}
		if err := s.AddElement(textLabel); err != nil {
			return arrow, nil, err
		}
		arrow.BoundElements = append(arrow.BoundElements, &BoundElement{Type: "text", ID: textLabel.ID})
	}

	return arrow, textLabel, nil
}

// ToJSON marshals the scene into an ExcalidrawFile JSON string.
func (s *ExcalidrawScene) ToJSON() (string, error) {
	exFile := ExcalidrawFile{
		Type:    "excalidraw",
		Version: 2,
		Source:  "https://github.com/panyam/sdl", // Or make configurable
		AppState: map[string]interface{}{
			"gridSize":          nil,
			"viewBackgroundColor": "#FFFFFF",
			"currentItemStrokeColor": "#1e1e1e", // Default stroke color for new items in UI
		},
		Elements: s.elements,
	}
	jsonData, err := json.MarshalIndent(exFile, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling Excalidraw JSON: %w", err)
	}
	return string(jsonData), nil
}

// generateExcalidrawOutput is now the adapter function that uses the generic ExcalidrawScene.
// This function will typically live in a different file (e.g., diagram.go or sdl_to_excalidraw_adapter.go)
// but is kept here for now as we are modifying excalidraw.go directly.
func generateExcalidrawOutput(systemName string, sdlNodes []DiagramNode, sdlEdges []DiagramEdge) (string, error) {
	scene := NewExcalidrawScene()
	
	// Layout parameters for SDL nodes
	layoutState := struct {
		currentX      float64
		currentY      float64
		startX        float64
		startY        float64
		elementWidth  float64
		elementHeight float64
		gapX          float64
		gapY          float64
		elementsPerRow int
		countInRow    int
	}{
		startX:        50.0,
		startY:        50.0,
		currentX:      50.0,
		currentY:      50.0,
		elementWidth:  200.0,
		elementHeight: 80.0,
		gapX:          60.0,
		gapY:          70.0,
		elementsPerRow: 3,
		countInRow:    0,
	}

	sdlNodeToExcalidrawRectID := make(map[string]string)

	for _, node := range sdlNodes {
		labelText := fmt.Sprintf("%s\n(%s)", node.Name, node.Type)
		rect, _, err := scene.AddRectangle(layoutState.currentX, layoutState.currentY, layoutState.elementWidth, layoutState.elementHeight, labelText, nil, nil)
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

	for _, edge := range sdlEdges {
		fromRectID, fromOk := sdlNodeToExcalidrawRectID[edge.FromID]
		toRectID, toOk := sdlNodeToExcalidrawRectID[edge.ToID]
		if !fromOk || !toOk {
			return "", fmt.Errorf("could not find Excalidraw ID for SDL node %s or %s for edge", edge.FromID, edge.ToID)
		}
		_, _, err := scene.AddArrow(fromRectID, toRectID, edge.Label, nil, nil)
		if err != nil {
			return "", fmt.Errorf("error adding SDL edge from %s to %s (label: %s) to Excalidraw scene: %w", edge.FromID, edge.ToID, edge.Label, err)
		}
	}

	return scene.ToJSON()
}
