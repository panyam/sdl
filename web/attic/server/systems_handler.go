package server

import (
	"encoding/json"

	"github.com/panyam/sdl/web/attic/services"
)

// SystemsHandler holds the system catalog for workspace discovery.
// All /systems routes now redirect to /workspaces — this only provides
// the catalog data to other handlers.
type SystemsHandler struct {
	catalog *services.SystemCatalogService
}

func NewSystemsHandler() *SystemsHandler {
	return &SystemsHandler{
		catalog: services.NewSystemCatalogService(),
	}
}

// toJSON converts data to JSON string for template use
func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
