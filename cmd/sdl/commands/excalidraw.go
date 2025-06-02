package commands

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// Excalidraw structures - simplified
type ExcalidrawElement struct {
	ID             string  `json:"id"`
	Type           string  `json:"type"`
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
	Width          float64 `json:"width"`
	Height         float64 `json:"height"`
	Angle          float64 `json:"angle"`
	StrokeColor    string  `json:"strokeColor"`
	BackgroundColor string  `json:"backgroundColor"`
	FillStyle      string  `json:"fillStyle"`
	StrokeWidth    int     `json:"strokeWidth"`
	StrokeStyle    string  `json:"strokeStyle"`
	Roughness      int     `json:"roughness"`
	Opacity        int     `json:"opacity"`
	Seed           int64   `json:"seed"`
	Version        int     `json:"version"`
	VersionNonce   int64   `json:"versionNonce"`
	IsDeleted      bool    `json:"isDeleted,omitempty"`
	BoundElements  []*BoundElement `json:"boundElements,omitempty"`
	StartBinding   *Binding `json:"startBinding,omitempty"`
	EndBinding     *Binding `json:"endBinding,omitempty"`
	Points         [][]float64 `json:"points,omitempty"`      // For arrows
	Text           string      `json:"text,omitempty"`         // For text elements
	FontSize       float64     `json:"fontSize,omitempty"`
	FontFamily     int         `json:"fontFamily,omitempty"`
	TextAlign      string      `json:"textAlign,omitempty"`
	VerticalAlign  string      `json:"verticalAlign,omitempty"`
	Baseline       int         `json:"baseline,omitempty"`
	ContainerId    *string     `json:"containerId,omitempty"` // For text bound to shapes
	OriginalText   string      `json:"originalText,omitempty"` // For text elements
	StrokeSharpness string `json:"strokeSharpness,omitempty"` // For arrows
	StartArrowhead *string `json:"startArrowhead,omitempty"`
	EndArrowhead   *string `json:"endArrowhead,omitempty"`
}

type Binding struct {
	ElementID string  `json:"elementId"`
	Focus     float64 `json:"focus"` // 0 is edge, 0.5 is center. Not used by Excalidraw much for arrows.
	Gap       float64 `json:"gap"`   // Default is 1
}

type BoundElement struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ExcalidrawFile struct {
	Type          string               `json:"type"`
	Version       int                  `json:"version"`
	Source        string               `json:"source"`
	Elements      []*ExcalidrawElement `json:"elements"`
	AppState      map[string]interface{} `json:"appState"`
	Files         map[string]interface{} `json:"files,omitempty"`
}

func newRandomSeed() int64 {
	// Excalidraw seeds are positive integers
	return rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(2147483646) + 1
}

func generateExcalidrawOutput(systemName string, nodes []DiagramNode, edges []DiagramEdge) (string, error) {
	exFile := ExcalidrawFile{
		Type:    "excalidraw",
		Version: 2,
		Source:  "https://github.com/panyam/sdl",
		AppState: map[string]interface{}{
			"gridSize":          nil,
			"viewBackgroundColor": "#FFFFFF",
			"currentItemStrokeColor": "#1e1e1e",
		},
		Elements: make([]*ExcalidrawElement, 0),
	}

	nodeIDToExcalidrawID := make(map[string]string)

	// Layout parameters
	startX, startY := 50.0, 50.0
	currentX, currentY := startX, startY
	elementWidth, elementHeight := 200.0, 80.0
	textPadding := 10.0
	gapX, gapY := 60.0, 70.0
	elementsPerRow := 3
	countInRow := 0

	defaultFontSize := 16.0
	lineHeightFactor := 1.2 // Approximate factor for line height based on font size
	textHeightForTwoLines := defaultFontSize * lineHeightFactor * 2

	for _, node := range nodes {
		excalidrawRectID := "rect_" + node.ID + "_" + strconv.FormatInt(newRandomSeed(), 10)
		nodeIDToExcalidrawID[node.ID] = excalidrawRectID

		rect := &ExcalidrawElement{
			ID:              excalidrawRectID,
			Type:            "rectangle",
			X:               currentX,
			Y:               currentY,
			Width:           elementWidth,
			Height:          elementHeight,
			StrokeColor:     "#1e1e1e",
			BackgroundColor: "#f8f9fa", // Lighter gray
			FillStyle:       "solid",
			StrokeWidth:     1,
			StrokeStyle:     "solid",
			Roughness:       0,
			Opacity:         100,
			Seed:            newRandomSeed(),
			Version:         1,
			VersionNonce:    newRandomSeed(),
			StrokeSharpness: "round",
			BoundElements:   make([]*BoundElement, 0), 
		}
		exFile.Elements = append(exFile.Elements, rect)

		textLabelContent := fmt.Sprintf("%s\n(%s)", node.Name, node.Type)
		excalidrawTextID := "text_" + node.ID + "_" + strconv.FormatInt(newRandomSeed(), 10)
		textElement := &ExcalidrawElement{
			ID:              excalidrawTextID,
			Type:            "text",
			X:               currentX + textPadding, // X,Y relative to container if containerId is set
			Y:               currentY + (elementHeight-textHeightForTwoLines)/2, 
			Width:           elementWidth - (2 * textPadding),
			Height:          textHeightForTwoLines, 
			StrokeColor:     "#1e1e1e",
			BackgroundColor: "transparent",
			Text:            textLabelContent,
			OriginalText:    textLabelContent,
			FontSize:        defaultFontSize,
			FontFamily:      1, // 1 usually maps to a sans-serif font like Excalidraw's default
			TextAlign:       "center",
			VerticalAlign:   "middle",
			Baseline:        int(defaultFontSize * 0.8), // Approximation
			Seed:            newRandomSeed(),
			Version:         1,
			VersionNonce:    newRandomSeed(),
			ContainerId:     &excalidrawRectID, 
			BoundElements:   nil, 
		}
		exFile.Elements = append(exFile.Elements, textElement)
		rect.BoundElements = append(rect.BoundElements, &BoundElement{Type: "text", ID: excalidrawTextID})

		countInRow++
		if countInRow >= elementsPerRow {
			currentX = startX
			currentY += elementHeight + gapY
			countInRow = 0
		} else {
			currentX += elementWidth + gapX
		}
	}

	for _, edge := range edges {
		fromExcalID, fromOk := nodeIDToExcalidrawID[edge.FromID]
		toExcalID, toOk := nodeIDToExcalidrawID[edge.ToID]

		if fromOk && toOk {
			arrowID := "arrow_" + edge.FromID + "_to_" + edge.ToID + "_" + edge.Label + "_" + strconv.FormatInt(newRandomSeed(), 10)
			arrowHeadStyle := "arrow"
			arrow := &ExcalidrawElement{
				ID:              arrowID,
				Type:            "arrow",
				StrokeColor:     "#1e1e1e",
				BackgroundColor: "transparent",
				StrokeWidth:     1,
				StrokeStyle:     "solid",
				Roughness:       0,
				Opacity:         100,
				Seed:            newRandomSeed(),
				Version:         1,
				VersionNonce:    newRandomSeed(),
				StrokeSharpness: "round",
				StartBinding:    &Binding{ElementID: fromExcalID, Focus: 0.5, Gap: 1},
				EndBinding:      &Binding{ElementID: toExcalID, Focus: 0.5, Gap: 1},
				EndArrowhead:    &arrowHeadStyle,
				Points:          [][]float64{{0, 0}, {50, 50}}, // Placeholder, bindings override
				BoundElements:   make([]*BoundElement, 0),
			}
			exFile.Elements = append(exFile.Elements, arrow)

			if edge.Label != "" {
				labelID := "label_" + arrowID + "_" + strconv.FormatInt(newRandomSeed(), 10)
				// For arrow labels, X, Y, Width, Height are often less critical when containerId is used,
				// as Excalidraw positions the label along the arrow.
				labelElement := &ExcalidrawElement{
					ID:              labelID,
					Type:            "text",
					X:               0, // Position relative to arrow path
					Y:               0, // Position relative to arrow path
					Width:           float64(len(edge.Label))*defaultFontSize*0.6 + 10, // Estimate width
					Height:          defaultFontSize * lineHeightFactor,
					StrokeColor:     "#343a40", // Darker gray for label
					BackgroundColor: "transparent", 
					Text:            edge.Label,
					OriginalText:    edge.Label,
					FontSize:        defaultFontSize * 0.9, // Slightly smaller for edge labels
					FontFamily:      1, 
					TextAlign:       "center",
					VerticalAlign:   "middle",
					Baseline:        int(defaultFontSize * 0.9 * 0.8),
					Seed:            newRandomSeed(),
					Version:         1,
					VersionNonce:    newRandomSeed(),
					ContainerId:     &arrowID, 
					BoundElements:   nil,
				}
				exFile.Elements = append(exFile.Elements, labelElement)
				arrow.BoundElements = append(arrow.BoundElements, &BoundElement{Type: "text", ID: labelID})
			}
		}
	}

	jsonData, err := json.MarshalIndent(exFile, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling Excalidraw JSON: %w", err)
	}
	return string(jsonData), nil
}
