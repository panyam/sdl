//go:build !wasm
// +build !wasm

package pages

import (
	"net/http"

	"github.com/panyam/sdl/services/fsbe"
	"github.com/panyam/templar"
)

// ViewContext holds shared context for rendering pages
type ViewContext struct {
	Templates     *templar.TemplateGroup
	CanvasService *fsbe.FSCanvasService
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
