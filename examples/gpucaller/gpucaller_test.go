// sdl/examples/gpucaller/gpucaller_test.go
package gpucaller

import (
	"fmt"
	"testing"

	"github.com/panyam/leetcoach/sdl/components"
	sdl "github.com/panyam/leetcoach/sdl/core"
)

// Helper function to set up the system
func setupGpuSystem(t *testing.T, gpuPoolSize int, appArrivalRate float64) *AppServer {
	const batchSize = 100 // Fixed batch size

	// 1. Define GPU Work Profile
	gpuWorkProfile := DefineGPUWorkProfile()
	if gpuWorkProfile == nil || gpuWorkProfile.Len() == 0 {
		t.Fatalf("Failed to define GPU work profile")
	}
	gpuWorkMeanLatency := sdl.MeanLatency(gpuWorkProfile) // Average time GPU is busy per batch

	// 2. Calculate Batch Arrival Rate at the Pool
	if appArrivalRate <= 0 {
		t.Logf("Warning: appArrivalRate is zero or negative, assuming very low batch rate.")
		appArrivalRate = 1e-9
	}
	batchArrivalRate := appArrivalRate / float64(batchSize) // Batches per second

	// 3. Create GPU Pool
	gpuPool := components.NewResourcePool(
		"GPU_Pool",
		uint(gpuPoolSize),
		batchArrivalRate,   // Lambda for M/M/c model (batches/sec)
		gpuWorkMeanLatency, // Ts for M/M/c model (avg hold time)
	)
	// We trust the Init function calculated them correctly for internal use.
	t.Logf("Setup: GPU Pool(Size=%d, BatchLambda=%.4f, Ts=%.6fs)", gpuPoolSize, batchArrivalRate, gpuWorkMeanLatency)

	// 4. Create GPU Batch Processor (Downstream for Batcher)
	gpuProcessor := (&GpuBatchProcessor{}).Init("GPU_Proc", gpuPool, gpuWorkProfile, batchArrivalRate)

	// 5. Create App Server Batcher
	appServerBatcher := components.NewBatcher(
		"App_Batcher",
		components.SizeBased,
		batchSize,
		0,              // Timeout not used for SizeBased
		appArrivalRate, // Lambda for Batcher (requests/sec)
		gpuProcessor,   // The downstream processor
	)
	t.Logf("Setup: App Batcher(Policy=SizeBased, N=%d, ReqLambda=%.2f)", batchSize, appArrivalRate)

	// 6. Create App Server
	appServer := (&AppServer{}).Init("AppServer_1", appServerBatcher)

	return appServer
}

// End-to-End Test
func TestGpuCaller_EndToEnd(t *testing.T) {
	// --- Simulation Parameters ---
	gpuPoolSize := 10       // Number of GPUs (N)
	appArrivalRate := 500.0 // Requests per second arriving at this app server (L_app)

	// --- Setup ---
	appServer := setupGpuSystem(t, gpuPoolSize, appArrivalRate)

	// --- Define Expectations ---
	// End-to-end SLO: P99 <= 500ms
	// Availability should be high, dominated by GPU work profile availability
	gpuWorkAvailability := sdl.Availability(DefineGPUWorkProfile())

	expectations := []sdl.Expectation{
		sdl.ExpectAvailability(sdl.GTE, gpuWorkAvailability*0.99), // Allow slight reduction
		sdl.ExpectP99(sdl.LT, sdl.Millis(500)),                    // The primary SLO
	}

	// --- Analyze ---
	analysisName := fmt.Sprintf("GPU Caller Infer (N=%d, Lambda=%.0f)", gpuPoolSize, appArrivalRate)
	analysisResult := sdl.Analyze(analysisName, func() *sdl.Outcomes[sdl.AccessResult] {
		return appServer.Infer()
	}, expectations...)

	// --- Assert ---
	analysisResult.Assert(t)

	// --- Optional: Log specific calculated values for context ---
	finalP99 := analysisResult.Metrics[sdl.P99LatencyMetric]
	slo := sdl.Millis(500)
	if finalP99 < slo {
		t.Logf("SLO PASSED: P99 Latency (%.6fs) is below target (%.3fs)", finalP99, slo)
	} else {
		t.Logf("SLO FAILED: P99 Latency (%.6fs) is >= target (%.3fs)", finalP99, slo)
	}
}

// --- Optional: Test with different parameters ---
func TestGpuCaller_HighLoad(t *testing.T) {
	gpuPoolSize := 5        // Fewer GPUs
	appArrivalRate := 800.0 // Higher arrival rate
	appServer := setupGpuSystem(t, gpuPoolSize, appArrivalRate)
	gpuWorkAvailability := sdl.Availability(DefineGPUWorkProfile())

	expectations := []sdl.Expectation{
		sdl.ExpectAvailability(sdl.GTE, gpuWorkAvailability*0.99),
		sdl.ExpectP99(sdl.LT, sdl.Millis(750)), // Might expect higher P99 under load
	}

	analysisName := fmt.Sprintf("GPU Caller Infer HighLoad (N=%d, Lambda=%.0f)", gpuPoolSize, appArrivalRate)
	analysisResult := sdl.Analyze(analysisName, func() *sdl.Outcomes[sdl.AccessResult] {
		return appServer.Infer()
	}, expectations...)

	analysisResult.Assert(t) // This might fail if P99 exceeds 750ms

	finalP99 := analysisResult.Metrics[sdl.P99LatencyMetric]
	slo := sdl.Millis(500)
	if finalP99 < slo {
		t.Logf("HighLoad: Primary SLO PASSED (P99 %.6fs < %.3fs)", finalP99, slo)
	} else {
		t.Logf("HighLoad: Primary SLO FAILED (P99 %.6fs >= %.3fs)", finalP99, slo)
	}

}
