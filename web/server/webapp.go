//go:build !wasm
// +build !wasm

package server

import (
	"html/template"
	"log"
	"net/http"

	goal "github.com/panyam/goapplib"
	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/services"
	gotl "github.com/panyam/goutils/template"
)

const TEMPLATES_FOLDER = "web/templates"

// BasePage provides common page properties.
// All pages should embed this.
type BasePage struct {
	goal.BasePage
	Title     string
	PageType  string
	ActiveTab string
}

// SdlApp is the pure application context.
// It holds all app-specific state without knowing about goapplib.
// Views access this via app.Context in goal.App[*SdlApp].
type SdlApp struct {
	// Services
	ClientMgr *services.ClientMgr

	// Canvas groups (will reference weewarApp and goalApp)
	CanvasesGroup *CanvasesGroup

	// Legacy handlers (API, WebSocket, Systems) - kept for now
	api            *SDLApi
	wsHandler      *CanvasWSHandler
	systemsHandler *SystemsHandler

	mux     *http.ServeMux
	BaseUrl string
}

// NewSdlApp creates a new SdlApp and its associated goal.App.
// Returns the SdlApp and the goal.App wrapper.
func NewSdlApp(grpcAddress string) (sdlApp *SdlApp, goalApp *goal.App[*SdlApp], err error) {
	// Create client manager for gRPC calls
	clientMgr := services.NewClientMgr(grpcAddress)

	// Create SdlApp (pure app context)
	sdlApp = &SdlApp{
		ClientMgr: clientMgr,
	}

	// Setup templates with app-specific FuncMap additions
	// goapplib templates are available via symlink at templates/goapplib/
	templates := goal.SetupTemplates(TEMPLATES_FOLDER)
	// Add goutils template functions (Ago, etc.)
	templates.AddFuncs(gotl.DefaultFuncMap())
	templates.AddFuncs(template.FuncMap{
		// Ctx provides access to the SdlApp context in templates
		"Ctx": func() *SdlApp { return sdlApp },
	})

	// Create goal.App wrapper
	goalApp = goal.NewApp(sdlApp, templates)

	// Initialize API (canvasService is nil - Connect handler will be skipped)
	api := NewSDLApi(grpcAddress, nil)
	sdlApp.api = api

	// Initialize WebSocket handler
	sdlApp.wsHandler = &CanvasWSHandler{
		clients: make(map[string]*CanvasWSConn),
	}

	// Initialize systems handler (still uses old pattern for now)
	templateGroup, err := SetupTemplates(TEMPLATES_FOLDER)
	if err != nil {
		log.Printf("Failed to setup templates for systems: %v", err)
	}
	sdlApp.systemsHandler = NewSystemsHandler(templateGroup)

	// Create CanvasesGroup
	sdlApp.CanvasesGroup = &CanvasesGroup{
		sdlApp:  sdlApp,
		goalApp: goalApp,
	}

	return
}

// Handler returns a configured HTTP handler with all routes.
func (a *SdlApp) Handler() http.Handler {
	r := http.NewServeMux()

	// API routes
	r.Handle("/api/", http.StripPrefix("/api", a.api.Handler()))

	// Serve examples directory for WASM demos
	r.Handle("/examples/", http.StripPrefix("/examples", http.FileServer(http.Dir("./examples/"))))

	// System showcase routes (server-rendered, using old handler for now)
	if a.systemsHandler != nil {
		r.Handle("/systems", a.systemsHandler.Handler())
		r.Handle("/systems/", a.systemsHandler.Handler())
		r.Handle("/system/", a.systemsHandler.Handler())
	}

	// Canvas pages (using new goapplib pattern)
	r.Handle("/canvases/", http.StripPrefix("/canvases", a.CanvasesGroup.Handler()))

	// Root redirect to systems listing
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			http.Redirect(w, req, "/systems", http.StatusFound)
			return
		}
		// Serve static files for other root-level paths
		http.FileServer(http.Dir("./web/dist/")).ServeHTTP(w, req)
	})

	return r
}

// CanvasesGroup implements goal.PageGroup for /canvases routes.
type CanvasesGroup struct {
	sdlApp  *SdlApp
	goalApp *goal.App[*SdlApp]
}

// Handler returns the configured HTTP handler for canvas routes.
func (g *CanvasesGroup) Handler() http.Handler {
	return g.RegisterRoutes(g.goalApp)
}

// RegisterRoutes registers all canvas-related routes using goal.Register.
func (g *CanvasesGroup) RegisterRoutes(app *goal.App[*SdlApp]) *http.ServeMux {
	mux := http.NewServeMux()

	// Register pages using goal's generic registration
	// Template spec format: "path/file" auto-derives block name from filename
	goal.Register[*CanvasListingPage](app, mux, "/",
		goal.WithTemplate("canvases/CanvasListingPage"))
	goal.Register[*CreateCanvasPage](app, mux, "GET /new",
		goal.WithTemplate("canvases/CreateCanvasPage"))
	goal.Register[*CanvasViewerPage](app, mux, "GET /{canvasId}/view",
		goal.WithTemplate("canvases/CanvasViewerPage"))
	goal.Register[*CanvasViewerPage](app, mux, "GET /{canvasId}/edit",
		goal.WithTemplate("canvases/CanvasViewerPage"))

	// Custom handlers for POST/DELETE actions
	mux.HandleFunc("POST /new", g.createCanvasHandler)
	// Handle all methods for /{canvasId} - DELETE is handled in canvasActionsHandler
	mux.HandleFunc("/{canvasId}", g.canvasActionsHandler)

	return mux
}

// createCanvasHandler handles POST to create a new canvas
func (g *CanvasesGroup) createCanvasHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	resp, err := g.sdlApp.ClientMgr.GetCanvasSvcClient().CreateCanvas(r.Context(), &protos.CreateCanvasRequest{
		Canvas: &protos.Canvas{
			Name:        name,
			Description: description,
		},
	})
	if err != nil {
		http.Error(w, "Failed to create canvas", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/canvases/"+resp.Canvas.Id+"/edit", http.StatusFound)
}

// deleteCanvasHandler handles DELETE to remove a canvas
func (g *CanvasesGroup) deleteCanvasHandler(w http.ResponseWriter, r *http.Request) {
	canvasId := r.PathValue("canvasId")
	if canvasId == "" {
		http.Error(w, "Canvas ID is required", http.StatusBadRequest)
		return
	}

	_, err := g.sdlApp.ClientMgr.GetCanvasSvcClient().DeleteCanvas(r.Context(), &protos.DeleteCanvasRequest{
		Id: canvasId,
	})
	if err != nil {
		http.Error(w, "Failed to delete canvas", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/canvases/", http.StatusFound)
}

// canvasActionsHandler handles default canvas actions
func (g *CanvasesGroup) canvasActionsHandler(w http.ResponseWriter, r *http.Request) {
	canvasId := r.PathValue("canvasId")
	if canvasId == "" {
		http.Error(w, "Canvas ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		g.deleteCanvasHandler(w, r)
	default:
		// Redirect to view
		http.Redirect(w, r, "/canvases/"+canvasId+"/view", http.StatusFound)
	}
}

// Page structs - implementing goal.View[*SdlApp]

// CanvasListingPage displays a list of all canvases
type CanvasListingPage struct {
	BasePage
	Header      Header
	ListingData *goal.EntityListingData[*protos.Canvas]
}

func (p *CanvasListingPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*SdlApp]) (err error, finished bool) {
	p.Title = "My Canvases"
	p.PageType = "canvas-listing"
	p.ActiveTab = "canvases"

	// Load header
	p.Header.Load(r, w, app)

	// Get all canvases via gRPC
	resp, err := app.Context.ClientMgr.GetCanvasSvcClient().ListCanvases(r.Context(), &protos.ListCanvasesRequest{})
	if err != nil {
		http.Error(w, "Failed to list canvases", http.StatusInternalServerError)
		return err, true
	}

	// Build listing data for EntityListing template
	p.ListingData = goal.NewEntityListingData[*protos.Canvas]("My Canvases", "/canvases").
		WithCreate("javascript:createNewCanvas()", "New Canvas").
		WithDelete("/canvases")
	p.ListingData.Items = resp.Canvases

	return nil, false
}

// CanvasViewerPage displays the canvas dashboard for viewing/editing
type CanvasViewerPage struct {
	BasePage
	Header   Header
	CanvasId string
	Canvas   *protos.Canvas
	ReadOnly bool
}

func (p *CanvasViewerPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*SdlApp]) (err error, finished bool) {
	p.CanvasId = r.PathValue("canvasId")
	p.ActiveTab = "canvases"

	// Load header
	p.Header.Load(r, w, app)

	if p.CanvasId == "" {
		http.Error(w, "Canvas ID is required", http.StatusBadRequest)
		return nil, true
	}

	// Determine readonly mode from URL path
	p.ReadOnly = len(r.URL.Path) >= 4 && r.URL.Path[len(r.URL.Path)-4:] == "view"

	// Get the canvas via gRPC
	resp, err := app.Context.ClientMgr.GetCanvasSvcClient().GetCanvas(r.Context(), &protos.GetCanvasRequest{
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

// CreateCanvasPage displays the form to create a new canvas
type CreateCanvasPage struct {
	BasePage
	Header Header
}

func (p *CreateCanvasPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*SdlApp]) (err error, finished bool) {
	p.Title = "Create Canvas"
	p.PageType = "canvas-create"
	p.ActiveTab = "canvases"

	// Load header
	p.Header.Load(r, w, app)

	return nil, false
}

// Header is a reusable header component
type Header struct {
	AppName   string
	ActiveTab string
}

func (h *Header) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*SdlApp]) (err error, finished bool) {
	h.AppName = "SDL"
	return nil, false
}
