//go:build !wasm
// +build !wasm

package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	goal "github.com/panyam/goapplib"
	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/services"
	gotl "github.com/panyam/goutils/template"
	gohttp "github.com/panyam/servicekit/http"
	tmplr "github.com/panyam/templar"
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

// ViteManifestEntry represents one entry in Vite's .vite/manifest.json
type ViteManifestEntry struct {
	File    string   `json:"file"`
	CSS     []string `json:"css"`
	IsEntry bool     `json:"isEntry"`
	Src     string   `json:"src"`
}

// SdlApp is the pure application context.
// It holds all app-specific state without knowing about goapplib.
// Views access this via app.Context in goal.App[*SdlApp].
type SdlApp struct {
	// Services
	ClientMgr *services.ClientMgr

	// Workspace group (routes at /workspaces, backed by Canvas proto)
	WorkspacesGroup *WorkspacesGroup

	// Legacy handlers (API, WebSocket, Systems) - kept for now
	api            *SDLApi
	wsHandler      *CanvasWSHandler
	systemsHandler *SystemsHandler

	// Vite manifest for cache-busted asset URLs
	ViteManifest map[string]ViteManifestEntry

	mux     *http.ServeMux
	BaseUrl string
}

// LoadViteManifest reads dist/.vite/manifest.json and populates ViteManifest.
func (a *SdlApp) LoadViteManifest(distDir string) {
	a.ViteManifest = make(map[string]ViteManifestEntry)
	manifestPath := filepath.Join(distDir, ".vite", "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Printf("Warning: could not read Vite manifest at %s: %v (using fallback paths)", manifestPath, err)
		return
	}
	if err := json.Unmarshal(data, &a.ViteManifest); err != nil {
		log.Printf("Warning: invalid Vite manifest: %v", err)
		return
	}
	for src, entry := range a.ViteManifest {
		if entry.IsEntry {
			log.Printf("[VITE] %s -> /%s (css: %v)", src, entry.File, entry.CSS)
		}
	}
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

	// Setup templates with SourceLoader for @goapplib/ vendored dependencies
	templates := tmplr.NewTemplateGroup()
	configPath := filepath.Join(TEMPLATES_FOLDER, "templar.yaml")
	sourceLoader, err := tmplr.NewSourceLoaderFromConfig(configPath)
	if err != nil {
		log.Printf("Warning: Could not load templar.yaml: %v. Falling back to basic loader.", err)
		templates.Loader = tmplr.NewFileSystemLoader(TEMPLATES_FOLDER)
	} else {
		templates.Loader = sourceLoader
	}
	templates.AddFuncs(goal.DefaultFuncMap())
	templates.AddFuncs(gotl.DefaultFuncMap())
	// Load Vite manifest for cache-busted asset URLs
	sdlApp.LoadViteManifest("dist")

	templates.AddFuncs(template.FuncMap{
		// Ctx provides access to the SdlApp context in templates
		"Ctx": func() *SdlApp { return sdlApp },
		// viteJS returns the hashed JS path for a Vite entry point
		// Usage: {{ viteJS "index.html" }} -> /assets/index-abc123.js
		"viteJS": func(entry string) string {
			if e, ok := sdlApp.ViteManifest[entry]; ok {
				return "/" + e.File
			}
			return "/assets/index.js" // fallback
		},
		// viteCSS returns the hashed CSS path for a Vite entry point
		// Usage: {{ viteCSS "index.html" }} -> /assets/index-abc123.css
		"viteCSS": func(entry string) string {
			if e, ok := sdlApp.ViteManifest[entry]; ok && len(e.CSS) > 0 {
				return "/" + e.CSS[0]
			}
			return "/assets/index.css" // fallback
		},
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

	// Initialize systems handler using the same SourceLoader-backed templates
	sdlApp.systemsHandler = NewSystemsHandler(templates)

	// Create WorkspacesGroup
	sdlApp.WorkspacesGroup = &WorkspacesGroup{
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

	// WebSocket endpoint for Canvas real-time updates
	r.HandleFunc("/ws/canvas", gohttp.WSServe(a.wsHandler, nil))

	// Serve examples directory for WASM demos
	r.Handle("/examples/", http.StripPrefix("/examples", http.FileServer(http.Dir("./examples/"))))

	// Workspace pages (unified view — replaces old /canvases and /systems)
	r.Handle("/workspaces/", http.StripPrefix("/workspaces", a.WorkspacesGroup.Handler()))

	// Backward-compat redirects for old routes
	r.HandleFunc("/systems", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/workspaces/", http.StatusFound)
	})
	r.HandleFunc("/systems/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/workspaces/", http.StatusFound)
	})
	r.HandleFunc("/system/", func(w http.ResponseWriter, req *http.Request) {
		// /system/bitly -> /workspaces/ (example fork will be done from the listing)
		http.Redirect(w, req, "/workspaces/", http.StatusFound)
	})
	r.HandleFunc("/canvases/", func(w http.ResponseWriter, req *http.Request) {
		newPath := "/workspaces" + req.URL.Path[len("/canvases"):]
		http.Redirect(w, req, newPath, http.StatusFound)
	})

	// Root redirect to workspaces
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			http.Redirect(w, req, "/workspaces/", http.StatusFound)
			return
		}
		// Serve static files for other root-level paths
		http.FileServer(http.Dir("./dist/")).ServeHTTP(w, req)
	})

	return r
}

// WorkspacesGroup implements goal.PageGroup for /workspaces routes.
// Backed by Canvas proto — "workspace" is UI naming only.
type WorkspacesGroup struct {
	sdlApp  *SdlApp
	goalApp *goal.App[*SdlApp]
}

// Handler returns the configured HTTP handler for workspace routes.
func (g *WorkspacesGroup) Handler() http.Handler {
	return g.RegisterRoutes(g.goalApp)
}

// RegisterRoutes registers all workspace-related routes using goal.Register.
func (g *WorkspacesGroup) RegisterRoutes(app *goal.App[*SdlApp]) *http.ServeMux {
	mux := http.NewServeMux()

	goal.Register[*WorkspaceListingPage](app, mux, "/",
		goal.WithTemplate("workspaces/WorkspaceListingPage"))
	goal.Register[*CreateWorkspacePage](app, mux, "GET /new",
		goal.WithTemplate("workspaces/CreateWorkspacePage"))
	goal.Register[*WorkspacePage](app, mux, "GET /{workspaceId}/view",
		goal.WithTemplate("workspaces/WorkspacePage"))
	goal.Register[*WorkspacePage](app, mux, "GET /{workspaceId}/edit",
		goal.WithTemplate("workspaces/WorkspacePage"))

	// Custom handlers for POST/DELETE/fork actions
	mux.HandleFunc("POST /new", g.createWorkspaceHandler)
	mux.HandleFunc("POST /fork", g.forkExampleHandler)
	mux.HandleFunc("/{workspaceId}", g.workspaceActionsHandler)

	return mux
}

// createWorkspaceHandler handles POST to create a new workspace (backed by Canvas)
func (g *WorkspacesGroup) createWorkspaceHandler(w http.ResponseWriter, r *http.Request) {
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

	http.Redirect(w, r, "/workspaces/"+resp.Canvas.Id+"/edit", http.StatusFound)
}

// forkExampleHandler opens an example as a workspace, creating one if needed.
// Uses the example ID as the canvas ID so repeated clicks reuse the same workspace.
func (g *WorkspacesGroup) forkExampleHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	exampleId := r.FormValue("exampleId")
	if exampleId == "" {
		http.Error(w, "Example ID is required", http.StatusBadRequest)
		return
	}

	// Look up the example system
	system := g.sdlApp.systemsHandler.catalog.GetSystem(exampleId)
	if system == nil {
		http.Error(w, "Example not found", http.StatusNotFound)
		return
	}

	// Try to open existing workspace for this example
	existing, _ := g.sdlApp.ClientMgr.GetCanvasSvcClient().GetCanvas(r.Context(), &protos.GetCanvasRequest{
		Id: exampleId,
	})
	if existing != nil && existing.Canvas != nil {
		http.Redirect(w, r, "/workspaces/"+exampleId+"/view", http.StatusFound)
		return
	}

	// Create a new workspace with the example's SDL pre-loaded
	version := system.Versions[system.DefaultVersion]
	resp, err := g.sdlApp.ClientMgr.GetCanvasSvcClient().CreateCanvas(r.Context(), &protos.CreateCanvasRequest{
		Canvas: &protos.Canvas{
			Id:             exampleId,
			Name:           system.Name,
			Description:    system.Description,
			SystemContents: version.SDL,
		},
	})
	if err != nil {
		http.Error(w, "Failed to create workspace", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/workspaces/"+resp.Canvas.Id+"/view", http.StatusFound)
}

// deleteWorkspaceHandler handles DELETE to remove a workspace
func (g *WorkspacesGroup) deleteWorkspaceHandler(w http.ResponseWriter, r *http.Request) {
	workspaceId := r.PathValue("workspaceId")
	if workspaceId == "" {
		http.Error(w, "Workspace ID is required", http.StatusBadRequest)
		return
	}

	_, err := g.sdlApp.ClientMgr.GetCanvasSvcClient().DeleteCanvas(r.Context(), &protos.DeleteCanvasRequest{
		Id: workspaceId,
	})
	if err != nil {
		http.Error(w, "Failed to delete workspace", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/workspaces/", http.StatusFound)
}

// workspaceActionsHandler handles default workspace actions
func (g *WorkspacesGroup) workspaceActionsHandler(w http.ResponseWriter, r *http.Request) {
	workspaceId := r.PathValue("workspaceId")
	if workspaceId == "" {
		http.Error(w, "Workspace ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		g.deleteWorkspaceHandler(w, r)
	default:
		http.Redirect(w, r, "/workspaces/"+workspaceId+"/view", http.StatusFound)
	}
}

// Page structs - implementing goal.View[*SdlApp]

// WorkspaceListingPage displays examples + user workspaces on one page
type WorkspaceListingPage struct {
	BasePage
	Header      Header
	Examples    []services.SystemInfo
	ListingData *goal.EntityListingData[*protos.Canvas]
}

func (p *WorkspaceListingPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*SdlApp]) (err error, finished bool) {
	p.Title = "Workspaces"
	p.PageType = "workspace-listing"
	p.ActiveTab = "workspaces"

	// Load header
	p.Header.Load(r, w, app)

	// Load examples from system catalog
	if app.Context.systemsHandler != nil {
		p.Examples = app.Context.systemsHandler.catalog.ListSystems()
	}

	// Get user workspaces via gRPC (Canvas service)
	resp, err := app.Context.ClientMgr.GetCanvasSvcClient().ListCanvases(r.Context(), &protos.ListCanvasesRequest{})
	if err != nil {
		// Non-fatal — show examples even if workspace service is down
		resp = &protos.ListCanvasesResponse{}
	}

	// Build listing data for EntityListing template
	p.ListingData = goal.NewEntityListingData[*protos.Canvas]("My Workspaces", "/workspaces/%s/view").
		WithCreate("javascript:createNewWorkspace()", "New Workspace").
		WithDelete("/workspaces/%s")
	p.ListingData.Items = resp.Canvases

	return nil, false
}

// WorkspacePage displays the workspace IDE for viewing/editing
type WorkspacePage struct {
	BasePage
	Header      Header
	WorkspaceId string
	Canvas      *protos.Canvas // proto stays Canvas
	ReadOnly    bool
}

func (p *WorkspacePage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*SdlApp]) (err error, finished bool) {
	p.WorkspaceId = r.PathValue("workspaceId")
	p.ActiveTab = "workspaces"

	// Load header
	p.Header.Load(r, w, app)

	if p.WorkspaceId == "" {
		http.Error(w, "Workspace ID is required", http.StatusBadRequest)
		return nil, true
	}

	// Determine readonly mode from URL path
	p.ReadOnly = len(r.URL.Path) >= 4 && r.URL.Path[len(r.URL.Path)-4:] == "view"

	// Get the workspace via gRPC (Canvas service)
	resp, err := app.Context.ClientMgr.GetCanvasSvcClient().GetCanvas(r.Context(), &protos.GetCanvasRequest{
		Id: p.WorkspaceId,
	})
	if err != nil {
		http.Error(w, "Workspace not found", http.StatusNotFound)
		return nil, true
	}

	p.Canvas = resp.Canvas
	if p.Canvas == nil {
		http.Error(w, "Workspace not found", http.StatusNotFound)
		return nil, true
	}

	p.PageType = "canvas-dashboard" // main.ts handles this page type
	if p.ReadOnly {
		p.Title = p.Canvas.Name
	} else {
		p.Title = "Edit: " + p.Canvas.Name
	}

	return nil, false
}

// CreateWorkspacePage displays the form to create a new workspace
type CreateWorkspacePage struct {
	BasePage
	Header Header
}

func (p *CreateWorkspacePage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*SdlApp]) (err error, finished bool) {
	p.Title = "Create Workspace"
	p.PageType = "workspace-create"
	p.ActiveTab = "workspaces"

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
