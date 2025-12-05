//go:build !wasm
// +build !wasm

package pages

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/services/fsbe"
	"github.com/panyam/templar"
)

// PagesHandler handles all page routes
type PagesHandler struct {
	mux     *http.ServeMux
	Context *ViewContext
}

// NewPagesHandler creates a new pages handler
func NewPagesHandler(templates *templar.TemplateGroup, canvasService *fsbe.FSCanvasService) *PagesHandler {
	h := &PagesHandler{
		mux: http.NewServeMux(),
		Context: &ViewContext{
			Templates:     templates,
			CanvasService: canvasService,
		},
	}
	h.setupRoutes()
	return h
}

func (h *PagesHandler) Handler() http.Handler {
	return h.mux
}

func (h *PagesHandler) setupRoutes() {
	// Canvas routes
	h.mux.HandleFunc("/canvases/", h.handleCanvasesRoot)
	h.mux.HandleFunc("/canvases/new", h.handleCreateCanvas)
	h.mux.HandleFunc("/canvases/{canvasId}/view", h.ViewRenderer(Copier(&CanvasViewerPage{}), "CanvasViewerPage"))
	h.mux.HandleFunc("/canvases/{canvasId}/edit", h.ViewRenderer(Copier(&CanvasViewerPage{}), "CanvasViewerPage"))
	h.mux.HandleFunc("/canvases/{canvasId}", h.handleCanvasActions)
}

// handleCanvasesRoot handles the /canvases/ listing page
func (h *PagesHandler) handleCanvasesRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/canvases/" {
		http.NotFound(w, r)
		return
	}
	h.RenderView(&CanvasListingPage{}, "CanvasListingPage", r, w)
}

// handleCreateCanvas handles creating a new canvas
func (h *PagesHandler) handleCreateCanvas(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Show the create form - just redirect to listing for now
		// TODO: Create a proper CanvasCreatePage
		http.Redirect(w, r, "/canvases/", http.StatusFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	// Create the canvas
	resp, err := h.Context.CanvasService.CreateCanvas(context.Background(), &protos.CreateCanvasRequest{
		Canvas: &protos.Canvas{
			Name:        name,
			Description: description,
		},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create canvas: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to the edit page
	http.Redirect(w, r, fmt.Sprintf("/canvases/%s/edit", resp.Canvas.Id), http.StatusFound)
}

// handleCanvasActions handles different HTTP methods for canvas operations
func (h *PagesHandler) handleCanvasActions(w http.ResponseWriter, r *http.Request) {
	canvasId := r.PathValue("canvasId")
	if canvasId == "" {
		http.Error(w, "Canvas ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		h.deleteCanvas(w, r, canvasId)
	default:
		// Default to redirecting to view
		http.Redirect(w, r, fmt.Sprintf("/canvases/%s/view", canvasId), http.StatusFound)
	}
}

// deleteCanvas handles deleting a canvas
func (h *PagesHandler) deleteCanvas(w http.ResponseWriter, r *http.Request, canvasId string) {
	_, err := h.Context.CanvasService.DeleteCanvas(context.Background(), &protos.DeleteCanvasRequest{
		Id: canvasId,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete canvas: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect back to listing
	http.Redirect(w, r, "/canvases/", http.StatusFound)
}

// ViewRenderer creates an HTTP handler for a view
func (h *PagesHandler) ViewRenderer(view ViewMaker, templateName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.RenderView(view(), templateName, r, w)
	}
}

// RenderView renders a view with its template
func (h *PagesHandler) RenderView(view View, templateName string, r *http.Request, w http.ResponseWriter) {
	if templateName == "" {
		t := reflect.TypeOf(view)
		e := t.Elem()
		templateName = e.Name()
	}

	err, finished := view.Load(r, w, h.Context)
	if finished {
		return
	}

	if err != nil {
		log.Println("Error loading view:", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}

	templateFile := "canvases/" + templateName + ".html"
	tmpl, err := h.Context.Templates.Loader.Load(templateFile, "")
	if err != nil {
		log.Printf("Template load error for %s: %v", templateFile, err)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	if err := h.Context.Templates.RenderHtmlTemplate(w, tmpl[0], templateName, view, nil); err != nil {
		log.Printf("Template render error: %v", err)
		http.Error(w, "Template render error", http.StatusInternalServerError)
	}
}
