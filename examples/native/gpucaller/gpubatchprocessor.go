// sdl/examples/gpucaller/gpubatchprocessor.go
package gpucaller

import (
	"fmt"
	// Required for Ceil
	"github.com/panyam/leetcoach/sdl/components"
	sdl "github.com/panyam/leetcoach/sdl/core"
)

// GpuBatchProcessor implements components.BatchProcessor
// It acquires a GPU, simulates processing, and (implicitly) releases the GPU.
type GpuBatchProcessor struct {
	Name             string
	GpuPool          *components.ResourcePool
	GpuWorkProfile   *sdl.Outcomes[sdl.AccessResult]                                       // The performance profile of the GPU work itself
	BatchArrivalRate float64                                                               // Lambda for batches arriving at the POOL
	ReductionTrigger int                                                                   // Configurable outcome reduction trigger length
	ReductionTarget  int                                                                   // Configurable outcome reduction target length
	ReductionTrimmer func(*sdl.Outcomes[sdl.AccessResult]) *sdl.Outcomes[sdl.AccessResult] // Cache trimmer
}

// Init initializes the processor.
func (p *GpuBatchProcessor) Init(name string, pool *components.ResourcePool, workProfile *sdl.Outcomes[sdl.AccessResult], batchLambda float64) *GpuBatchProcessor {
	p.Name = name
	if pool == nil || workProfile == nil {
		panic(fmt.Sprintf("GpuBatchProcessor '%s' initialized with nil pool or workProfile", name))
	}
	p.GpuPool = pool
	p.GpuWorkProfile = workProfile
	p.BatchArrivalRate = batchLambda                                           // Rate at which BATCHES hit the pool
	p.ReductionTrigger = 100                                                   // Default trigger length for reduction
	p.ReductionTarget = 15                                                     // Default target length after reduction
	p.ReductionTrimmer = sdl.TrimToSize(p.ReductionTrigger, p.ReductionTarget) // Cache the trimmer
	return p
}

// ProcessBatch defines how a batch is handled.
func (p *GpuBatchProcessor) ProcessBatch(batchSize uint) *sdl.Outcomes[sdl.AccessResult] {
	// 1. Acquire GPU resource - includes potential queueing time at the pool
	//    We pass the BatchArrivalRate (lambda for the pool) to Acquire's model.
	acqOutcome := p.GpuPool.Acquire() // Acquire now gets lambda from pool's internal state

	// 2. Get the GPU work outcome (pre-defined profile for the batch)
	workOutcome := p.GpuWorkProfile

	// 3. Combine Acquire + Work sequentially
	//    The latency adds up. Success depends on both steps.
	combined := sdl.And(acqOutcome, workOutcome,
		func(acqRes, workRes sdl.AccessResult) sdl.AccessResult {
			// If acquire failed (e.g., infinite queue), the operation fails.
			// If acquire succeeded but work failed, the operation fails.
			// Latency is Acquire Latency (wait time) + Work Latency.
			return sdl.AccessResult{
				Success: acqRes.Success && workRes.Success,
				Latency: acqRes.Latency + workRes.Latency,
			}
		})

	// 4. Apply reduction to manage outcome complexity
	//    Split -> Trim -> Append pattern
	successes, failures := combined.Split(sdl.AccessResult.IsSuccess)
	// Ensure ReductionTrimmer is initialized (defensive)
	trimmer := p.ReductionTrimmer
	if trimmer == nil {
		trimmer = sdl.TrimToSize(p.ReductionTrigger, p.ReductionTarget)
	}
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures) // Apply to failures too
	finalOutcome := (&sdl.Outcomes[sdl.AccessResult]{And: combined.And}).Append(trimmedSuccesses, trimmedFailures)

	// Note: ResourcePool.Release is not explicitly called here. The analytical model
	// uses the AvgHoldTime (derived from workOutcome mean) to calculate contention.
	// Note: Network latency between AppServer and GPU Pool is not explicitly
	// included here. It's assumed to be negligible or part of the GpuWorkProfile.
	// For higher fidelity, add explicit NetworkLink steps before/after workOutcome.
	return finalOutcome
}
