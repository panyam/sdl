//go:build !wasm
// +build !wasm

package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/services"
	"github.com/panyam/templar"
)

// ViewContext holds shared context for rendering pages
type ViewContext struct {
	Templates *templar.TemplateGroup
	ClientMgr *services.ClientMgr
}

// View interface that all pages must implement
type View interface {
	Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool)
}

// Copyable interface for page cloning
type Copyable interface {
	Copy() View
}

// Copier creates a ViewMaker from a Copyable
func Copier[V Copyable](v V) ViewMaker {
	return v.Copy
}

// ViewMaker is a function that creates a new View instance
type ViewMaker func() View

// BasePage contains common page properties
type BasePage struct {
	Title    string
	PageType string
}

// PagesHandler handles all page routes
type PagesHandler struct {
	mux     *http.ServeMux
	Context *ViewContext
}

// NewPagesHandler creates a new pages handler
func NewPagesHandler(templates *templar.TemplateGroup, clientMgr *services.ClientMgr) *PagesHandler {
	h := &PagesHandler{
		mux: http.NewServeMux(),
		Context: &ViewContext{
			Templates: templates,
			ClientMgr: clientMgr,
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
	h.mux.HandleFunc("/canvases/new", h.handleNewCanvas)
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

// handleNewCanvas handles GET (render form) and POST (create canvas)
func (h *PagesHandler) handleNewCanvas(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Render the create canvas form
		h.RenderView(&CreateCanvasPage{}, "CreateCanvasPage", r, w)
	case http.MethodPost:
		// Handle form submission
		h.createCanvasHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// createCanvasHandler handles the POST to create a new canvas
func (h *PagesHandler) createCanvasHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	// Create the canvas via gRPC
	resp, err := h.Context.ClientMgr.GetCanvasSvcClient().CreateCanvas(context.Background(), &protos.CreateCanvasRequest{
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
	_, err := h.Context.ClientMgr.GetCanvasSvcClient().DeleteCanvas(context.Background(), &protos.DeleteCanvasRequest{
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

// CanvasListingPage displays a list of all canvases
type CanvasListingPage struct {
	BasePage
	Canvases []*protos.Canvas
}

func (p *CanvasListingPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.Title = "My Canvases"
	p.PageType = "canvas-listing"

	// Get all canvases via gRPC
	resp, err := vc.ClientMgr.GetCanvasSvcClient().ListCanvases(context.Background(), &protos.ListCanvasesRequest{})
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

	// Get the canvas via gRPC
	resp, err := vc.ClientMgr.GetCanvasSvcClient().GetCanvas(context.Background(), &protos.GetCanvasRequest{
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

// CreateCanvasPage displays the form to create a new canvas
type CreateCanvasPage struct {
	BasePage
}

func (p *CreateCanvasPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.Title = "Create Canvas"
	p.PageType = "canvas-create"
	return nil, false
}

func (p *CreateCanvasPage) Copy() View {
	return &CreateCanvasPage{}
}
