package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/panyam/sdl/services"
	"github.com/panyam/templar"
)

// WasmModule represents a WASM module to be loaded by the page
type WasmModule struct {
	Name string `json:"name"` // Module identifier (e.g., "systemdetail")
	Path string `json:"path"` // Path to .wasm file (e.g., "/dist/wasm/systemdetail.wasm")
}

// SystemsHandler handles system showcase pages
type SystemsHandler struct {
	templateGroup *templar.TemplateGroup
	catalog       *services.SystemCatalogService
}

// NewSystemsHandler creates a new systems handler
func NewSystemsHandler(templateGroup *templar.TemplateGroup) *SystemsHandler {
	return &SystemsHandler{
		templateGroup: templateGroup,
		catalog:       services.NewSystemCatalogService(),
	}
}

// Handler returns an HTTP handler for systems routes
func (h *SystemsHandler) Handler() http.Handler {
	mux := http.NewServeMux()

	// System listing page
	mux.HandleFunc("/systems", h.handleSystemListing)
	mux.HandleFunc("/systems/", h.handleSystemListing)

	// System details page
	mux.HandleFunc("/system/", h.handleSystemDetails)

	return mux
}

// handleSystemListing renders the system listing page
func (h *SystemsHandler) handleSystemListing(w http.ResponseWriter, r *http.Request) {
	// Get all systems from catalog
	systems := h.catalog.ListSystems()

	// Prepare template data
	data := map[string]interface{}{
		"Title":    "System Examples",
		"PageType": "system-listing",
		"Systems":  systems,
		"PageDataJSON": toJSON(map[string]interface{}{
			"pageType": "system-listing",
		}),
	}

	// Load and render template
	templates := h.templateGroup.MustLoad("systems/SystemsListingPage.html", "")

	// Render the template
	if err := h.templateGroup.RenderHtmlTemplate(w, templates[0], "SystemsListingPage", data, nil); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render page: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleSystemDetails redirects to workspace view (system details page moved to attic)
func (h *SystemsHandler) handleSystemDetails(w http.ResponseWriter, r *http.Request) {
	// Extract system ID from path: /system/bitly
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	systemID := parts[2]

	// Verify system exists
	system := h.catalog.GetSystem(systemID)
	if system == nil {
		http.NotFound(w, r)
		return
	}

	// Redirect to canvas viewer with this system's ID
	// Phase 2 will consolidate this under /workspaces/{id}
	http.Redirect(w, r, "/canvases/"+systemID+"/view", http.StatusFound)
}

// toJSON converts data to JSON string for template use
func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
