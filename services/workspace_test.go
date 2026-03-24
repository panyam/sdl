package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseWorkspaceManifest verifies that sdl.json files parse directly
// into Workspace proto messages via protojson. This is the core manifest
// parsing that all workspace loading depends on.
func TestParseWorkspaceManifest(t *testing.T) {
	manifest := `{
		"name": "Test Workspace",
		"description": "A test workspace",
		"sources": {
			"stdlib": {"builtin": true},
			"patterns": {"github": "panyam/sdl-patterns", "ref": "v1.0.0", "path": "lib/"}
		},
		"designs": [
			{"name": "DesignA", "file": "a.sdl", "description": "First design"},
			{"name": "DesignB", "file": "b.sdl"}
		]
	}`

	ws, err := ParseWorkspaceManifest([]byte(manifest))
	require.NoError(t, err)

	assert.Equal(t, "Test Workspace", ws.Name)
	assert.Equal(t, "A test workspace", ws.Description)

	// Sources
	assert.Len(t, ws.Sources, 2)
	assert.True(t, ws.Sources["stdlib"].Builtin)
	assert.Equal(t, "panyam/sdl-patterns", ws.Sources["patterns"].Github)
	assert.Equal(t, "v1.0.0", ws.Sources["patterns"].Ref)
	assert.Equal(t, "lib/", ws.Sources["patterns"].Path)

	// Designs
	assert.Len(t, ws.Designs, 2)
	assert.Equal(t, "DesignA", ws.Designs[0].Name)
	assert.Equal(t, "a.sdl", ws.Designs[0].File)
	assert.Equal(t, "First design", ws.Designs[0].Description)
	assert.Equal(t, "DesignB", ws.Designs[1].Name)
	assert.Equal(t, "b.sdl", ws.Designs[1].File)
}

// TestLoadUberManifest verifies the real Uber example sdl.json parses correctly.
// This catches drift between the proto schema and the example manifests.
func TestLoadUberManifest(t *testing.T) {
	ws, err := LoadWorkspaceManifest("../examples/uber/sdl.json")
	require.NoError(t, err)

	assert.Equal(t, "Uber Architecture Evolution", ws.Name)
	assert.True(t, ws.Sources["stdlib"].Builtin)
	assert.Len(t, ws.Designs, 3)

	names := make([]string, len(ws.Designs))
	for i, d := range ws.Designs {
		names[i] = d.Name
	}
	assert.Contains(t, names, "UberMVP")
	assert.Contains(t, names, "UberIntermediate")
	assert.Contains(t, names, "UberModern")
}

// TestLoadBitlyManifest verifies the real Bitly example sdl.json parses correctly.
func TestLoadBitlyManifest(t *testing.T) {
	ws, err := LoadWorkspaceManifest("../examples/bitly/sdl.json")
	require.NoError(t, err)

	assert.Equal(t, "Bitly URL Shortener", ws.Name)
	assert.Len(t, ws.Designs, 1)
	assert.Equal(t, "Bitly", ws.Designs[0].Name)
}
