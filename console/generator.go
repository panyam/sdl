package console

import (
	"log"
	"time"

	protos "github.com/panyam/sdl/gen/go/sdl/v1"
	"github.com/panyam/sdl/runtime"
)

type GeneratorInfo struct {
	*protos.Generator

	stopped                   bool
	stopChan                  chan bool
	System                    *runtime.SystemInstance
	resolvedComponentInstance *runtime.ComponentInstance // Resolved from Component name

	GenFunc func(iter int)
}

// Stops the generator
func (g *GeneratorInfo) Stop() {
	if !g.Enabled || g.stopChan == nil {
		return
	}
	g.stopped = true
	g.stopChan <- true
}

// Starts a generator
func (g *GeneratorInfo) Start() {
	if g.Enabled {
		return
	}
	g.Enabled = true
	g.stopped = false
	g.stopChan = make(chan bool)

	go g.run()
}

func (g *GeneratorInfo) run() {
	genFuncMissing := g.GenFunc == nil
	defer func() {
		close(g.stopChan)
		g.stopChan = nil
		g.Enabled = false
		if genFuncMissing {
			g.GenFunc = nil
		}
	}()

	if g.GenFunc == nil {
		g.GenFunc = func(iter int) {
			log.Println("Generating Dummy Logs: ", iter)
		}
	}

	ticker := time.NewTicker(1000 * time.Millisecond)
	for i := 0; ; i++ {
		select {
		case <-g.stopChan:
			return
		case <-ticker.C:
			// Generate it here
			g.GenFunc(i)
		}
	}
}
