package console

import (
	"log"
	"sync"
	"time"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
	protos "github.com/panyam/sdl/gen/go/sdl/v1"
	"github.com/panyam/sdl/runtime"
)

type GeneratorInfo struct {
	*protos.Generator

	stopped                   bool
	stopChan                  chan bool
	canvas                    *Canvas                     // Reference to parent canvas
	System                    *runtime.SystemInstance
	resolvedComponentInstance *runtime.ComponentInstance // Resolved from Component name
	resolvedMethodDecl        *runtime.MethodDecl       // Resolved method declaration

	// Virtual time management
	nextVirtualTime core.Duration
	timeMutex       sync.Mutex

	// For fractional rate handling
	eventAccumulator float64

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

	// Initialize GenFunc if not provided
	if g.GenFunc == nil {
		g.initializeGenFunc()
	}

	// Calculate interval based on rate
	interval := time.Second / time.Duration(g.Rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for i := 0; ; i++ {
		select {
		case <-g.stopChan:
			return
		case <-ticker.C:
			// Execute with virtual time
			g.GenFunc(i)
		}
	}
}

// initializeGenFunc sets up the actual eval-based generator function
func (g *GeneratorInfo) initializeGenFunc() {
	// Store reference to canvas for tracer access
	canvas := g.canvas

	g.GenFunc = func(iter int) {
		// Get next virtual time
		virtualTime := g.getNextVirtualTime()

		// Get tracer from canvas (may be nil if no metrics registered)
		var tracer runtime.Tracer
		if canvas != nil && canvas.metricTracer != nil {
			tracer = canvas.metricTracer
		}

		// Create evaluator with tracer
		eval := runtime.NewSimpleEval(g.System.File, tracer)

		// Create new environment for this execution
		env := g.System.Env.Push()
		currTime := virtualTime

		// Create call expression
		callExpr := &decl.CallExpr{
			Function: &decl.MemberAccessExpr{
				Receiver: &decl.IdentifierExpr{Value: g.Component},
				Member:   &decl.IdentifierExpr{Value: g.Method},
			},
			// TODO: Add support for arguments in the future
		}

		// Execute the call - all trace events will have virtual timestamps
		result, _ := eval.Eval(callExpr, env, &currTime)
		if eval.HasErrors() {
			log.Printf("Generator %s error during eval", g.Id)
		} else if result.IsNil() {
			log.Printf("Generator %s: method returned nil", g.Id)
		}
	}
}

// getNextVirtualTime returns the next virtual time and advances the clock
func (g *GeneratorInfo) getNextVirtualTime() core.Duration {
	g.timeMutex.Lock()
	defer g.timeMutex.Unlock()

	current := g.nextVirtualTime
	// Advance by 1/rate seconds
	g.nextVirtualTime += core.Duration(1.0 / float64(g.Rate))
	return current
}
