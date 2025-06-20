package console

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
	protos "github.com/panyam/sdl/gen/go/sdl/v1"
	sdlruntime "github.com/panyam/sdl/runtime"
)

type GeneratorInfo struct {
	*protos.Generator

	stopped                   atomic.Bool
	stopChan                  chan bool
	canvas                    *Canvas // Reference to parent canvas
	System                    *sdlruntime.SystemInstance
	resolvedComponentInstance *sdlruntime.ComponentInstance // Resolved from Component name
	resolvedMethodDecl        *sdlruntime.MethodDecl        // Resolved method declaration

	// Virtual time management
	nextVirtualTime core.Duration
	timeMutex       sync.Mutex

	// For fractional rate handling
	eventAccumulator float64

	GenFunc func(iter int)
}

// Stops the generator
func (g *GeneratorInfo) Stop() {
	if g.stopped.Load() || g.stopChan == nil {
		return
	}
	log.Printf("Generator %s: Stopping...", g.Id)
	g.stopped.Store(true)
	close(g.stopChan)
}

// Starts a generator
func (g *GeneratorInfo) Start() {
	if g.Enabled {
		return
	}
	g.Enabled = true
	g.stopped.Store(false)
	g.stopChan = make(chan bool)

	go g.run()
}

func (g *GeneratorInfo) run() {
	genFuncMissing := g.GenFunc == nil
	defer func() {
		// Don't close stopChan here - it's closed by Stop()
		g.stopChan = nil
		g.Enabled = false
		if genFuncMissing {
			g.GenFunc = nil
		}
		log.Printf("Generator %s: Stopped", g.Id)
	}()

	// Initialize GenFunc if not provided
	if g.GenFunc == nil {
		g.initializeGenFunc()
	}

	// For high QPS, use batching
	if g.Rate > 100 {
		g.runBatched()
	} else {
		g.runSimple()
	}
}

// runSimple handles low rate generators (< 100 RPS)
func (g *GeneratorInfo) runSimple() {
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

// runBatched handles high rate generators (>= 100 RPS)
func (g *GeneratorInfo) runBatched() {
	// Batch interval: 10ms for consistent timing
	batchInterval := 10 * time.Millisecond
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	// Calculate events per batch
	eventsPerBatch := float64(g.Rate) * batchInterval.Seconds()

	log.Printf("Generator %s: Starting batched execution at %v RPS (%.2f events per %v batch)",
		g.Id, g.Rate, eventsPerBatch, batchInterval)

	// Bounded concurrency
	maxConcurrent := runtime.NumCPU() * 2
	sem := make(chan struct{}, maxConcurrent)

	batchCount := 0

	defer func() { log.Printf("Generator %s: Stopped after %d batches", g.Id, batchCount) }()

	for {
		select {
		case <-g.stopChan:
			return
		case <-ticker.C:
			// Calculate batch size with fractional accumulation
			g.eventAccumulator += eventsPerBatch
			batchSize := int(g.eventAccumulator)
			g.eventAccumulator -= float64(batchSize)

			if batchSize > 0 {
				batchCount++
				if batchCount%100 == 0 { // Log every 100 batches (1 second at 10ms intervals)
					log.Printf("Generator %s: Processed %d batches, current batch size: %d, Stopped: ", g.Id, batchCount, batchSize, g.stopped.Load())
				}
			}

			// Pre-calculate virtual times for this batch
			virtualTimes := make([]core.Duration, batchSize)
			for i := 0; i < batchSize; i++ {
				virtualTimes[i] = g.getNextVirtualTime()
			}

			// Execute batch with bounded concurrency
			for i := 0; i < batchSize; i++ {
				if g.stopped.Load() {
					return
				}
				vTime := virtualTimes[i]

				select {
				case <-g.stopChan:
					return
				case sem <- struct{}{}: // Acquire semaphore
					go func(virtualTime core.Duration) {
						defer func() { <-sem }() // Release semaphore
						g.executeAtVirtualTime(virtualTime)
					}(vTime)
				}
			}
		}
	}
}

// initializeGenFunc sets up the actual eval-based generator function
func (g *GeneratorInfo) initializeGenFunc() {
	g.GenFunc = func(iter int) {
		// Get next virtual time
		virtualTime := g.getNextVirtualTime()
		// Use the common execution method
		g.executeAtVirtualTime(virtualTime)
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

// executeAtVirtualTime executes a single eval at the given virtual time
func (g *GeneratorInfo) executeAtVirtualTime(virtualTime core.Duration) {
	// Get tracer from canvas (may be nil)
	var tracer sdlruntime.Tracer
	if g.canvas != nil && g.canvas.metricTracer != nil {
		tracer = g.canvas.metricTracer
	}

	// Create evaluator with tracer
	eval := sdlruntime.NewSimpleEval(g.System.File, tracer)

	// New environment for isolation
	env := g.System.Env.Push()
	currTime := virtualTime

	// Build method call expression
	callExpr := &decl.CallExpr{
		Function: &decl.MemberAccessExpr{
			Receiver: &decl.IdentifierExpr{Value: g.Component},
			Member:   &decl.IdentifierExpr{Value: g.Method},
		},
	}

	// Execute with virtual time
	result, _ := eval.Eval(callExpr, env, &currTime)
	if eval.HasErrors() {
		log.Printf("Generator %s error during eval", g.Id)
	} else if result.IsNil() {
		// This is normal for void methods
	}
}
