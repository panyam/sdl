//go:build !wasm
// +build !wasm

package pages

import (
	"context"
	"net/http"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// CanvasViewerPage displays the canvas dashboard for viewing/editing
type CanvasViewerPage struct {
	BasePage
	CanvasId string
	Canvas   *protos.Canvas
	ReadOnly bool // If true, Monaco editor will be readonly
}

func (p *CanvasViewerPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.CanvasId = r.PathValue("canvasId")
	if p.CanvasId == "" {
		http.Error(w, "Canvas ID is required", http.StatusBadRequest)
		return nil, true
	}

	// Determine readonly mode from URL path
	// /canvases/{id}/view = readonly
	// /canvases/{id}/edit = editable
	p.ReadOnly = r.URL.Path[len(r.URL.Path)-4:] == "view"

	// Get the canvas from fsbe service
	resp, err := vc.CanvasService.GetCanvas(context.Background(), &protos.GetCanvasRequest{
		Id: p.CanvasId,
	})
	if err != nil {
		http.Error(w, "Canvas not found", http.StatusNotFound)
		return nil, true
	}

	p.Canvas = resp.Canvas

	if p.ReadOnly {
		p.Title = p.Canvas.Name
		p.PageType = "canvas-view"
	} else {
		p.Title = "Edit: " + p.Canvas.Name
		p.PageType = "canvas-edit"
	}

	return nil, false
}

func (p *CanvasViewerPage) Copy() View {
	return &CanvasViewerPage{}
}
