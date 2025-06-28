package console

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
	sdlruntime "github.com/panyam/sdl/runtime"
)

type GeneratorInfo struct {
	*Generator // Use native type instead of proto

	stopped                   atomic.Bool
	stopChan                  chan bool
	canvas                    *Canvas // Reference to parent canvas
	System                    *sdlruntime.SystemInstance
	resolvedComponentInstance *sdlruntime.ComponentInstance // Resolved from Component name
	resolvedMethodDecl        *sdlruntime.MethodDecl        // Resolved method declaration

	// Virtual time management
	nextVirtualTime core.Duration
	timeMutex       sync.Mutex
	stopNotifyChan  chan bool

	// For fractional rate handling
	eventAccumulator float64

	GenFunc func(iter int)
}

// IsRunning returns true if the generator is currently running
func (g *GeneratorInfo) IsRunning() bool {
	return g.Enabled && !g.stopped.Load()
}

// Stop stops the generator
func (g *GeneratorInfo) Stop(wait bool) error {
	if g.stopped.Load() || g.stopChan == nil {
		return nil
	}
	if wait {
		g.stopNotifyChan = make(chan bool, 1)
		defer func() {
			close(g.stopNotifyChan)
			g.stopNotifyChan = nil
		}()
	}
	log.Printf("Generator %s: Stopping...", g.ID)
	g.stopped.Store(true)
	close(g.stopChan)
	g.Enabled = false
	<-g.stopNotifyChan
	return nil
}

// Start starts a generator
func (g *GeneratorInfo) Start() error {
	log.Printf("Is it enabled???: ", g.Enabled)
	if g.Enabled {
		return nil
	}
	g.Enabled = true
	g.stopped.Store(false)
	g.stopChan = make(chan bool)

	go g.run()
	return nil
}

func (g *GeneratorInfo) run() {
	defer func() {
		// Don't close stopChan here - it's closed by Stop()
		g.stopChan = nil
		g.Enabled = false
		log.Printf("Generator %s: Stopped", g.ID)
		log.Println("notifying....", g.stopNotifyChan)
		if g.stopNotifyChan != nil {
			g.stopNotifyChan <- true
		}
	}()

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

	log.Printf("Generator %s: Starting Simple execution at %v RPS", g.ID, g.Rate)

	if g.GenFunc == nil {
		// Initialize GenFunc if not provided
		g.initializeGenFunc()
		defer func() {
			g.GenFunc = nil
		}()
	}

	for i := 0; ; i++ {
		// Log every 10 or 20 iterations since this is low qps
		if i%100 == 0 { // Log every 100 batches (1 second at 10ms intervals)
			log.Printf("Low RPS Generator %s: Processed %d iterations, Stopped: %t, Rate: %f, interval: %f", g.ID, i, g.stopped.Load(), g.Rate, interval)
		}
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
		g.ID, g.Rate, eventsPerBatch, batchInterval)

	// Bounded concurrency
	maxConcurrent := runtime.NumCPU() * 2
	sem := make(chan struct{}, maxConcurrent)

	batchCount := 0

	defer func() { log.Printf("Generator %s: Stopped after %d batches", g.ID, batchCount) }()

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
					log.Printf("Generator %s: Processed %d batches, current batch size: %d, Stopped: ", g.ID, batchCount, batchSize, g.stopped.Load())
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

// initializeGenFunc sets up the actual eval-based generator function for low QPS generations
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
	// Create evaluator with tracer - ok to panic if nil as it shouldnt be
	eval := sdlruntime.NewSimpleEval(g.System.File, g.canvas.metricTracer)

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
		log.Printf("Generator %s error during eval", g.ID)
	} else if result.IsNil() {
		// This is normal for void methods
	}
}
