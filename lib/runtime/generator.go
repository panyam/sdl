package runtime

import (
	"log"
	goruntime "runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panyam/sdl/lib/core"
	"github.com/panyam/sdl/lib/decl"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// Generator represents a traffic generator bound to a system.
// Embeds the proto Generator for transport and adds runtime execution state.
// This consolidates the old runtime.Generator + services.GeneratorInfo into one type.
type Generator struct {
	*protos.Generator // Embed proto (Id, Name, Component, Method, Rate, Duration, Enabled)

	// Rate configuration not in proto
	RateInterval core.Duration // Interval in seconds (default 1.0 = per second)

	// Resolved references (populated during system init)
	ResolvedComponent *ComponentInstance
	ResolvedMethod    *MethodDecl

	// Runtime execution state
	stopped          atomic.Bool
	stopChan         chan bool
	SimCtx           SimulationContext
	System           *SystemInstance
	nextVirtualTime  core.Duration
	timeMutex        sync.Mutex
	stopNotifyChan   chan bool
	eventAccumulator float64
	GenFunc          func(iter int)
}

// RPS returns the effective requests per second.
func (g *Generator) RPS() float64 {
	if g.RateInterval <= 0 {
		return g.Rate // assume per second
	}
	return g.Rate / g.RateInterval
}

// NewGeneratorFromSpec creates a Generator from a compile-time GeneratorSpec.
func NewGeneratorFromSpec(spec *GeneratorSpec) *Generator {
	return &Generator{
		Generator: &protos.Generator{
			Id:       spec.Name,
			Name:     spec.Name,
			Component: spec.ComponentPath,
			Method:   spec.MethodName,
			Rate:     spec.Rate,
			Duration: spec.Duration,
			Enabled:  true,
		},
		RateInterval: spec.RateInterval,
	}
}

// IsRunning returns true if the generator is currently running.
func (g *Generator) IsRunning() bool {
	return g.Enabled && !g.stopped.Load()
}

// Stop stops the generator.
func (g *Generator) Stop(wait bool) error {
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
	log.Printf("Generator %s: Stopping...", g.Id)
	g.stopped.Store(true)
	close(g.stopChan)
	g.Enabled = false
	<-g.stopNotifyChan
	return nil
}

// Start starts a generator.
func (g *Generator) Start() error {
	if g.Enabled {
		return nil
	}
	g.Enabled = true
	g.stopped.Store(false)
	g.stopChan = make(chan bool)
	go g.run()
	return nil
}

func (g *Generator) run() {
	defer func() {
		g.stopChan = nil
		g.Enabled = false
		log.Printf("Generator %s: Stopped", g.Id)
		if g.stopNotifyChan != nil {
			g.stopNotifyChan <- true
		}
	}()

	if g.Rate > 100 {
		g.runBatched()
	} else {
		g.runSimple()
	}
}

func (g *Generator) runSimple() {
	interval := time.Second / time.Duration(g.Rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Generator %s: Starting Simple execution at %v RPS", g.Id, g.Rate)

	if g.GenFunc == nil {
		g.initializeGenFunc()
		defer func() { g.GenFunc = nil }()
	}

	for i := 0; ; i++ {
		if i%100 == 0 {
			log.Printf("Low RPS Generator %s: Processed %d iterations, Stopped: %t, Rate: %f, interval: %v", g.Id, i, g.stopped.Load(), g.Rate, interval)
		}
		select {
		case <-g.stopChan:
			return
		case <-ticker.C:
			g.GenFunc(i)
		}
	}
}

func (g *Generator) runBatched() {
	batchInterval := 10 * time.Millisecond
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	eventsPerBatch := float64(g.Rate) * batchInterval.Seconds()

	log.Printf("Generator %s: Starting batched execution at %v RPS (%.2f events per %v batch)",
		g.Id, g.Rate, eventsPerBatch, batchInterval)

	maxConcurrent := goruntime.NumCPU() * 2
	sem := make(chan struct{}, maxConcurrent)
	batchCount := 0

	defer func() { log.Printf("Generator %s: Stopped after %d batches", g.Id, batchCount) }()

	for {
		select {
		case <-g.stopChan:
			return
		case <-ticker.C:
			g.eventAccumulator += eventsPerBatch
			batchSize := int(g.eventAccumulator)
			g.eventAccumulator -= float64(batchSize)

			if batchSize > 0 {
				batchCount++
				if batchCount%100 == 0 {
					log.Printf("Generator %s: Processed %d batches, current batch size: %d, Stopped: %v", g.Id, batchCount, batchSize, g.stopped.Load())
				}
			}

			virtualTimes := make([]core.Duration, batchSize)
			for i := range batchSize {
				virtualTimes[i] = g.getNextVirtualTime()
			}

			for i := range batchSize {
				if g.stopped.Load() {
					return
				}
				vTime := virtualTimes[i]

				select {
				case <-g.stopChan:
					return
				case sem <- struct{}{}:
					go func(virtualTime core.Duration) {
						defer func() { <-sem }()
						g.executeAtVirtualTime(virtualTime)
					}(vTime)
				}
			}
		}
	}
}

func (g *Generator) initializeGenFunc() {
	g.GenFunc = func(iter int) {
		virtualTime := g.getNextVirtualTime()
		g.executeAtVirtualTime(virtualTime)
	}
}

func (g *Generator) getNextVirtualTime() core.Duration {
	g.timeMutex.Lock()
	defer g.timeMutex.Unlock()
	current := g.nextVirtualTime
	g.nextVirtualTime += core.Duration(1.0 / float64(g.Rate))
	return current
}

func (g *Generator) executeAtVirtualTime(virtualTime core.Duration) {
	eval := NewSimpleEval(g.System.File, g.SimCtx.GetTracer())
	env := g.System.Env.Push()
	currTime := virtualTime

	callExpr := &decl.CallExpr{
		Function: &decl.MemberAccessExpr{
			Receiver: &decl.IdentifierExpr{Value: g.Component},
			Member:   &decl.IdentifierExpr{Value: g.Method},
		},
	}

	result, _ := eval.Eval(callExpr, env, &currTime)
	if eval.HasErrors() {
		log.Printf("Generator %s error during eval", g.Id)
	} else if result.IsNil() {
		// Normal for void methods
	}
}
