//go:build !wasm
// +build !wasm

package fsbe

import (
	"path/filepath"

	"github.com/panyam/goutils/utils"
)

const SDL_DATA_ROOT = "~/dev-app-data/sdl"

// DevDataPath returns a path relative to the SDL data root
func DevDataPath(path string) string {
	return filepath.Join(utils.ExpandUserPath(SDL_DATA_ROOT), path)
}
