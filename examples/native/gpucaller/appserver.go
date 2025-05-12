// sdl/examples/gpucaller/appserver.go
package gpucaller

import (
	"fmt"

	"github.com/panyam/sdl/components"
	sdl "github.com/panyam/sdl/core"
)

// AppServer represents the application server handling inference requests.
type AppServer struct {
	Name    string
	Batcher *components.Batcher
}

// Init initializes the AppServer.
func (a *AppServer) Init(name string, batcher *components.Batcher) *AppServer {
	a.Name = name
	if batcher == nil {
		panic(fmt.Sprintf("AppServer '%s' initialized with nil Batcher", name))
	}
	a.Batcher = batcher
	return a
}

// Infer simulates handling a single inference request end-to-end from the user perspective.
func (a *AppServer) Infer() *sdl.Outcomes[sdl.AccessResult] {
	// The core logic is handled by submitting the request to the batcher.
	// The batcher manages waiting for batch formation and calling the downstream
	// GpuBatchProcessor, which handles GPU acquisition and work simulation.
	// The result includes batch wait time + GPU queue wait time + GPU processing time.
	return a.Batcher.Submit()
}
