package services

import (
	"fmt"
	"os"
	"path/filepath"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"google.golang.org/protobuf/encoding/protojson"
)

// LoadWorkspaceManifest reads sdl.json and returns a Workspace proto directly.
// Uses protojson for zero-adapter deserialization.
func LoadWorkspaceManifest(path string) (*protos.Workspace, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	return ParseWorkspaceManifest(data)
}

// ParseWorkspaceManifest parses manifest JSON into a Workspace proto.
func ParseWorkspaceManifest(data []byte) (*protos.Workspace, error) {
	ws := &protos.Workspace{}
	if err := protojson.Unmarshal(data, ws); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	return ws, nil
}

// DiscoverWorkspaceManifest looks for sdl.json in the given directory.
func DiscoverWorkspaceManifest(dir string) (string, error) {
	for _, name := range []string{"sdl.json"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no sdl.json found in %s", dir)
}
