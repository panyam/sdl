// sdl/examples/gpucaller/gpucaller_test.go
package gpucaller

import (
	"fmt"
	"testing"

	"github.com/panyam/leetcoach/sdl/components"
	sdl "github.com/panyam/leetcoach/sdl/core"
)

// Helper function to set up the system
// setupGpuSystem wires together the components for a single AppServer instance
// interacting with a GPU pool.
//
// Limitations:
//   - Single App Server: Models the performance path for one server instance.
//     Total system throughput requires external extrapolation (NumServers * PerServerCapacity).
//   - Steady State: Represents average performance after warm-up, not initial requests.
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

// analyzeSystem performs the setup, analysis, and assertion for a given config
// Note: The results reflect steady-state analytical approximations.
func analyzeSystemAndGetResult(t *testing.T, gpuPoolSize int, appArrivalRate float64, p99SLOMillis float64) sdl.AnalysisResult[sdl.AccessResult] {
	t.Helper()
	// --- Setup ---
	appServer := setupGpuSystem(t, gpuPoolSize, appArrivalRate)

	// --- Define Expectations ---
	// gpuWorkAvailability := sdl.Availability(DefineGPUWorkProfile())
	targetP99SLO := sdl.Millis(p99SLOMillis)
	expectations := []sdl.Expectation{
		// sdl.ExpectAvailability(sdl.GTE, gpuWorkAvailability*0.99), // Allow slight reduction
		sdl.ExpectAvailability(sdl.GTE, 0.92), // Expect >= 0.92 after reduction
		sdl.ExpectP99(sdl.LT, targetP99SLO),   // Check against the parameter
	}

	// --- Analyze ---
	analysisName := fmt.Sprintf("GPU Caller Infer (N=%d, Lambda=%.0f)", gpuPoolSize, appArrivalRate)
	analysisResult := sdl.Analyze(analysisName, func() *sdl.Outcomes[sdl.AccessResult] {
		return appServer.Infer()
	}, expectations...)

	return analysisResult // Return the result struct
}

// --- Test Scenarios ---
// These scenarios test the system model under different configurations.
// The accuracy depends on the analytical models used (M/M/c for pool,
// average wait for batcher) and the realism of the defined GPU work profile.
func TestGpuCaller_Scenarios(t *testing.T) {
	// Base SLO
	const p99SLO = 500.0 // ms

	// Define scenarios to test
	scenarios := []struct {
		name        string
		gpuPoolSize int
		appQPS      float64
		expectFail  bool
	}{
		{"Baseline_10GPU_5kQPS", 10, 5000.0, false},
		{"Baseline_20GPU_5kQPS", 20, 5000.0, false},
		{"HighLoad_10GPU_10kQPS", 10, 10000.0, false},
		{"HighLoad_20GPU_10kQPS", 20, 10000.0, false},
		{"TargetLoad_20GPU_20kQPS", 20, 20000.0, false}, // Target QPS
		{"TargetLoad_30GPU_20kQPS", 30, 20000.0, false}, // More GPUs for Target QPS
		{"StressLoad_30GPU_25kQPS", 30, 25000.0, false},
		// Add a scenario likely to saturate the pool and miss SLO
		{"Saturate_10GPU_20kQPS", 10, 20000.0, true},
	}

	// Run analysis for each scenario as a subtest
	for _, sc := range scenarios {
		// Capture scenario variable for closure
		scenario := sc
		t.Run(scenario.name, func(subT *testing.T) {
			// Call the analysis function (which no longer asserts internally)
			result := analyzeSystemAndGetResult(subT, scenario.gpuPoolSize, scenario.appQPS, p99SLO)

			// Always log the detailed results
			result.LogResults(subT)

			// Conditional Assertion: Fail the test only if the pass/fail status is unexpected
			if scenario.expectFail {
				// We expected one or more expectations within Analyze to fail.
				// The test run PASSES if Analyze reported a failure (result.AllPassed == false).
				if result.AllPassed {
					subT.Errorf("Test Logic FAIL: Expected expectation failures for this scenario, but all passed.")
				} else {
					// This is the desired outcome for this scenario. Log appropriately maybe.
					subT.Logf("Test Logic PASS: Scenario correctly resulted in expectation failures as expected.")
				}
			} else {
				// We expected all expectations within Analyze to pass.
				// The test run PASSES if Analyze reported success (result.AllPassed == true).
				if !result.AllPassed {
					subT.Errorf("Test Logic FAIL: Expected all expectations to pass, but some failed.")
				} else {
					// This is the desired outcome.
					subT.Logf("Test Logic PASS: All expectations met as expected.")
				}
			}
		})
	}
}
