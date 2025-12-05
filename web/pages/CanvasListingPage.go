//go:build !wasm
// +build !wasm

package pages

import (
	"context"
	"net/http"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// CanvasListingPage displays a list of all canvases
type CanvasListingPage struct {
	BasePage
	Canvases []*protos.Canvas
}

func (p *CanvasListingPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.Title = "My Canvases"
	p.PageType = "canvas-listing"

	// Get all canvases from the fsbe service
	resp, err := vc.CanvasService.ListCanvases(context.Background(), &protos.ListCanvasesRequest{})
	if err != nil {
		http.Error(w, "Failed to list canvases", http.StatusInternalServerError)
		return err, true
	}

	p.Canvases = resp.Canvases
	return nil, false
}

func (p *CanvasListingPage) Copy() View {
	return &CanvasListingPage{}
}
