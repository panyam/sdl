// sdl/examples/gpucaller/gpuwork.go
package gpucaller

import (
	sdl "github.com/panyam/leetcoach/sdl/core"
)

// DefineGPUWorkProfile creates an assumed Outcomes distribution representing the
// performance of processing one batch of 100 requests on a single GPU.
// It targets a P90 latency of <= 100ms.
//
// Limitation: The accuracy of the overall simulation depends significantly on how
// well this assumed profile matches the real-world performance characteristics
// (mean, variance, tail behavior, failure modes) of the GPU batch processing.
// Consider deriving this profile from actual measurements or using distribution
// helpers (when available) based on known percentiles for better accuracy.
func DefineGPUWorkProfile() *sdl.Outcomes[sdl.AccessResult] {
	profile := &sdl.Outcomes[sdl.AccessResult]{And: sdl.AndAccessResults}

	profile.Add(85, sdl.AccessResult{Success: true, Latency: sdl.Millis(75)}) // 85% fast path (75ms)
	profile.Add(5, sdl.AccessResult{Success: true, Latency: sdl.Millis(100)}) // 5% at P90 SLO (100ms)
	profile.Add(5, sdl.AccessResult{Success: true, Latency: sdl.Millis(150)}) // 5% tail (150ms)
	profile.Add(5, sdl.AccessResult{Success: false, Latency: sdl.Millis(50)}) // 5% failure (detected quickly)

	// Normalize weights just in case they don't sum perfectly (optional but good practice)
	// profile.Normalize() // Assuming Normalize exists or weights sum to 100

	return profile
}
