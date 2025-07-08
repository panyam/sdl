package services

import (
	"fmt"
	"testing"

	"github.com/panyam/sdl/types"
	"github.com/stretchr/testify/assert"
)

func TestCapacityModeling(t *testing.T) {
	canvas := NewCanvas("test", nil)

	// 1. Load the SDL file with the capacity-aware component
	err := canvas.Load("../examples/disk.sdl")
	assert.NoError(t, err, "Loading capacity.sdl should succeed")

	// 2. Activate the system
	err = canvas.Use("TestCapacitySystem")
	assert.NoError(t, err, "Using TestCapacitySystem should succeed")

	// 3. Configure the model
	// The _rawReadLatency is: 90% @ 10ms, 9% @ 50ms, 1% @ 500ms
	// Mean = (0.90 * 10) + (0.09 * 50) + (0.01 * 500) = 9 + 4.5 + 5 = 18.5ms
	avgHoldTimeSeconds := 0.0185
	err = canvas.Set("MyDisk._pool.AvgHoldTime", avgHoldTimeSeconds)
	assert.NoError(t, err, "Setting AvgHoldTime on the disk's pool should succeed")

	// The theoretical max capacity (lambda_max) is c / Ts = 1 / 0.0185 ≈ 54.05 QPS
	// Let's test at 3 load levels.

	// --- Run 1: Low Load (10 QPS) ---
	// At this load, queueing delay should be minimal.
	lowLoadQPS := 10.0
	err = canvas.Set("MyDisk._pool.ArrivalRate", lowLoadQPS)
	assert.NoError(t, err)

	err = canvas.Run("low_load_results", "MyDisk.Read", WithRuns(2000))
	assert.NoError(t, err)

	// --- Run 2: High Load (50 QPS) ---
	// At this load (rho ≈ 50/54 = 0.925), queueing delay should be significant.
	highLoadQPS := 50.0
	err = canvas.Set("MyDisk._pool.ArrivalRate", highLoadQPS)
	assert.NoError(t, err)

	err = canvas.Run("high_load_results", "MyDisk.Read", WithRuns(2000))
	assert.NoError(t, err)

	// --- Run 3: Overload (60 QPS) ---
	// At this load (rho > 1), we expect most requests to fail.
	overloadQPS := 60.0
	err = canvas.Set("MyDisk._pool.ArrivalRate", overloadQPS)
	assert.NoError(t, err)

	err = canvas.Run("overload_results", "MyDisk.Read", WithRuns(2000))
	assert.NoError(t, err)

	// 4. Analyze and Assert Results
	lowLoadResults := canvas.sessionVars["low_load_results"].([]types.RunResult)
	highLoadResults := canvas.sessionVars["high_load_results"].([]types.RunResult)
	overloadResults := canvas.sessionVars["overload_results"].([]types.RunResult)

	avgLatencyLow := calculateAverageLatency(lowLoadResults)
	avgLatencyHigh := calculateAverageLatency(highLoadResults)
	failureRateOverload := calculateFailureRate(overloadResults)

	// Debug removed for cleaner output

	fmt.Println("--- Capacity Modeling Test Results ---")
	fmt.Printf("Raw Operation Latency (Avg): %.2f ms\n", avgHoldTimeSeconds*1000)
	fmt.Printf("Low Load (%.0f QPS) Avg Latency:  %.2f ms\n", lowLoadQPS, avgLatencyLow)
	fmt.Printf("High Load (%.0f QPS) Avg Latency: %.2f ms\n", highLoadQPS, avgLatencyHigh)
	fmt.Printf("Overload (%.0f QPS) Failure Rate: %.2f%%\n", overloadQPS, failureRateOverload*100)

	// Assertions
	// At low load, latency should be very close to the raw service time.
	assert.InDelta(t, avgLatencyLow, avgHoldTimeSeconds*1000, 5.0, "Low load latency should be close to raw latency")

	// At high load, latency should be significantly higher due to queuing.
	assert.Greater(t, avgLatencyHigh, avgLatencyLow*2, "High load latency should be much greater than low load latency")

	// At overload, most operations should fail.
	assert.Greater(t, failureRateOverload, 0.9, "Overload failure rate should be very high")
}

// Helper function to calculate average latency from run results.
func calculateAverageLatency(results []types.RunResult) float64 {
	var totalLatency float64
	count := 0
	for _, r := range results {
		// Only consider successful runs for latency calculation
		if r.ResultValue == "Val(Bool: true)" {
			totalLatency += r.Latency
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return totalLatency / float64(count)
}

// Helper function to calculate the failure rate.
func calculateFailureRate(results []types.RunResult) float64 {
	failures := 0
	for _, r := range results {
		// A failure is when the component returns 'false'.
		if r.ResultValue == "Val(Bool: false)" {
			failures++
		}
	}
	if len(results) == 0 {
		return 0
	}
	return float64(failures) / float64(len(results))
}
