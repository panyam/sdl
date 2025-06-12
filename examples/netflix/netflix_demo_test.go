package workshop

import (
	"testing"

	"github.com/panyam/sdl/console"
)

// TestNetflixTrafficSpikeDemo validates all major features needed for the conference demo
// This serves as our comprehensive test bed for workshop functionality
func TestNetflixTrafficSpikeDemo(t *testing.T) {
	canvas := console.NewCanvas()

	// Load the Netflix scenario
	err := canvas.Load("netflix.sdl")
	if err != nil {
		t.Fatalf("Failed to load netflix.sdl: %v", err)
	}

	err = canvas.Use("NetflixSystem")
	if err != nil {
		t.Fatalf("Failed to use NetflixSystem: %v", err)
	}

	// === BASELINE TESTING ===
	t.Run("Baseline Performance", func(t *testing.T) {
		// Set normal traffic conditions
		err := canvas.Set("videoService.cdn.pool.ArrivalRate", 50.0)
		if err != nil {
			t.Fatalf("Failed to set CDN arrival rate: %v", err)
		}

		err = canvas.Set("videoService.cdn.pool.AvgHoldTime", "50ms")
		if err != nil {
			t.Fatalf("Failed to set CDN hold time: %v", err)
		}

		// Run baseline simulation
		err = canvas.Run("baseline", "videoService.StreamVideo", 1000)
		if err != nil {
			t.Fatalf("Failed to run baseline simulation: %v", err)
		}

		// Validate results exist and are reasonable
		results := canvas.GetResults("baseline")
		if len(results) == 0 {
			t.Fatal("No baseline results generated")
		}

		// Should have reasonable latency under normal load
		if results.P95Latency > 100.0 { // 100ms
			t.Errorf("Baseline P95 latency too high: %v ms", results.P95Latency)
		}

		// Should have high success rate
		if results.SuccessRate < 0.95 {
			t.Errorf("Baseline success rate too low: %v", results.SuccessRate)
		}
	})

	// === TRAFFIC SPIKE SIMULATION ===
	t.Run("Traffic Spike Impact", func(t *testing.T) {
		// Simulate "Stranger Things" season premiere - 4x traffic
		err := canvas.Set("videoService.cdn.pool.ArrivalRate", 200.0)
		if err != nil {
			t.Fatalf("Failed to set high arrival rate: %v", err)
		}

		err = canvas.Run("traffic_spike", "videoService.StreamVideo", 1000)
		if err != nil {
			t.Fatalf("Failed to run traffic spike simulation: %v", err)
		}

		baseline := canvas.GetResults("baseline")
		spike := canvas.GetResults("traffic_spike")

		// Traffic spike should significantly increase latency
		if spike.P95Latency <= baseline.P95Latency {
			t.Errorf("Traffic spike should increase latency: baseline=%v, spike=%v", 
				baseline.P95Latency, spike.P95Latency)
		}

		// Should show some failures under overload
		if spike.SuccessRate >= baseline.SuccessRate {
			t.Errorf("Traffic spike should reduce success rate: baseline=%v, spike=%v",
				baseline.SuccessRate, spike.SuccessRate)
		}
	})

	// === CACHE OPTIMIZATION ===
	t.Run("Cache Hit Rate Optimization", func(t *testing.T) {
		// Improve cache hit rate to 95%
		err := canvas.Set("videoService.cdn.CacheHitRate", 0.95)
		if err != nil {
			t.Fatalf("Failed to set cache hit rate: %v", err)
		}

		err = canvas.Run("optimized_cache", "videoService.StreamVideo", 1000)
		if err != nil {
			t.Fatalf("Failed to run optimized cache simulation: %v", err)
		}

		spike := canvas.GetResults("traffic_spike")
		optimized := canvas.GetResults("optimized_cache")

		// Better cache should improve performance even under load
		if optimized.P95Latency >= spike.P95Latency {
			t.Errorf("Cache optimization should improve latency: spike=%v, optimized=%v",
				spike.P95Latency, optimized.P95Latency)
		}
	})

	// === CAPACITY SCALING ===
	t.Run("CDN Capacity Scaling", func(t *testing.T) {
		// Double CDN capacity
		err := canvas.Set("videoService.cdn.pool.Size", 200)
		if err != nil {
			t.Fatalf("Failed to scale CDN capacity: %v", err)
		}

		err = canvas.Run("scaled_cdn", "videoService.StreamVideo", 1000)
		if err != nil {
			t.Fatalf("Failed to run scaled CDN simulation: %v", err)
		}

		scaled := canvas.GetResults("scaled_cdn")

		// Scaled capacity should handle the load well
		if scaled.SuccessRate < 0.98 {
			t.Errorf("Scaled CDN should have high success rate: %v", scaled.SuccessRate)
		}
	})

	// === DATABASE BOTTLENECK ===
	t.Run("Database Bottleneck", func(t *testing.T) {
		// Create database bottleneck
		err := canvas.Set("videoService.database.pool.ArrivalRate", 100.0)
		if err != nil {
			t.Fatalf("Failed to set database arrival rate: %v", err)
		}

		err = canvas.Run("db_bottleneck", "videoService.StreamVideo", 1000)
		if err != nil {
			t.Fatalf("Failed to run database bottleneck simulation: %v", err)
		}

		bottleneck := canvas.GetResults("db_bottleneck")

		// Database bottleneck should show up as degraded performance
		if bottleneck.SuccessRate > 0.90 {
			t.Logf("Database bottleneck may not be severe enough: success rate %v", 
				bottleneck.SuccessRate)
		}
	})

	// === ENCODING WORKLOAD ===
	t.Run("Video Encoding Performance", func(t *testing.T) {
		// Reset to reasonable settings for encoding test
		err := canvas.Set("videoService.encoder.pool.ArrivalRate", 5.0)
		if err != nil {
			t.Fatalf("Failed to set encoder arrival rate: %v", err)
		}

		err = canvas.Run("encoding_load", "videoService.UploadVideo", 100)
		if err != nil {
			t.Fatalf("Failed to run encoding simulation: %v", err)
		}

		encoding := canvas.GetResults("encoding_load")

		// Encoding should be much slower but reliable
		if encoding.P95Latency < 30000.0 { // 30 seconds
			t.Errorf("Encoding should take significant time: %v ms", encoding.P95Latency)
		}

		if encoding.SuccessRate < 0.95 {
			t.Errorf("Encoding should be reliable: %v", encoding.SuccessRate)
		}
	})

	// === VISUALIZATION TESTING ===
	t.Run("Visualization Generation", func(t *testing.T) {
		// Test that we can generate various visualizations
		tests := []struct {
			name     string
			plotType string
			results  []string
		}{
			{"Latency Plot", "latency", []string{"baseline"}},
			{"Comparison Plot", "comparison", []string{"baseline", "traffic_spike"}},
			{"Histogram", "histogram", []string{"scaled_cdn"}},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := canvas.Plot(test.results, test.plotType, "test_output")
				if err != nil {
					t.Errorf("Failed to generate %s: %v", test.name, err)
				}
			})
		}
	})

	// === ARCHITECTURE DIAGRAM ===
	t.Run("Architecture Diagram", func(t *testing.T) {
		err := canvas.Diagram("NetflixSystem", "static", "test_architecture")
		if err != nil {
			t.Errorf("Failed to generate architecture diagram: %v", err)
		}
	})
}

// TestWorkshopFeatureValidation tests specific features needed for the workshop
func TestWorkshopFeatureValidation(t *testing.T) {
	canvas := console.NewCanvas()

	err := canvas.Load("netflix.sdl")
	if err != nil {
		t.Fatalf("Failed to load netflix.sdl: %v", err)
	}

	err = canvas.Use("NetflixSystem")
	if err != nil {
		t.Fatalf("Failed to use NetflixSystem: %v", err)
	}

	// === PARAMETER MODIFICATION VALIDATION ===
	t.Run("Parameter Modification", func(t *testing.T) {
		// Test different parameter types
		paramTests := []struct {
			path  string
			value interface{}
		}{
			{"videoService.cdn.pool.Size", 150},
			{"videoService.cdn.pool.ArrivalRate", 75.5},
			{"videoService.cdn.pool.AvgHoldTime", "25ms"},
			{"videoService.cdn.CacheHitRate", 0.90},
		}

		for _, test := range paramTests {
			t.Run(test.path, func(t *testing.T) {
				err := canvas.Set(test.path, test.value)
				if err != nil {
					t.Errorf("Failed to set %s to %v: %v", test.path, test.value, err)
				}
			})
		}
	})

	// === RAPID ITERATION VALIDATION ===
	t.Run("Rapid Parameter Changes", func(t *testing.T) {
		// Simulate rapid parameter changes during live demo
		arrivalRates := []float64{25, 50, 100, 200, 150, 75}
		
		for i, rate := range arrivalRates {
			err := canvas.Set("videoService.cdn.pool.ArrivalRate", rate)
			if err != nil {
				t.Fatalf("Failed to set arrival rate to %v: %v", rate, err)
			}

			runName := fmt.Sprintf("rapid_test_%d", i)
			err = canvas.Run(runName, "videoService.StreamVideo", 500) // Smaller count for speed
			if err != nil {
				t.Fatalf("Failed to run simulation %s: %v", runName, err)
			}

			results := canvas.GetResults(runName)
			if len(results) == 0 {
				t.Fatalf("No results for simulation %s", runName)
			}
		}
	})

	// === EDGE CASE VALIDATION ===
	t.Run("Edge Cases", func(t *testing.T) {
		// Test extreme parameter values
		edgeTests := []struct {
			name  string
			path  string
			value interface{}
		}{
			{"Zero Capacity", "videoService.cdn.pool.Size", 0},
			{"Minimal Capacity", "videoService.cdn.pool.Size", 1},
			{"Very High Load", "videoService.cdn.pool.ArrivalRate", 1000.0},
			{"Perfect Cache", "videoService.cdn.CacheHitRate", 1.0},
			{"No Cache", "videoService.cdn.CacheHitRate", 0.0},
		}

		for _, test := range edgeTests {
			t.Run(test.name, func(t *testing.T) {
				err := canvas.Set(test.path, test.value)
				if err != nil {
					t.Errorf("Failed to set %s: %v", test.path, err)
				}

				err = canvas.Run(test.name, "videoService.StreamVideo", 100)
				// Don't fail on simulation errors for edge cases - they might legitimately fail
				if err != nil {
					t.Logf("Edge case %s resulted in error (may be expected): %v", test.name, err)
				}
			})
		}
	})
}

// Placeholder for results structure - this should match what Canvas.GetResults returns
type SimulationResults struct {
	P95Latency  float64
	P50Latency  float64
	SuccessRate float64
	Count       int
}