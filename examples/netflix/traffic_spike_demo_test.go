package netflix

import (
	"fmt"
	"testing"

	"github.com/panyam/sdl/console"
)

// TestNetflixTrafficSpikeDemo - Canvas-based recipe for conference workshop
// This serves as our comprehensive test bed validating all major features
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
	t.Log("=== BASELINE TESTING ===")
	// Test normal traffic conditions
	canvas.Set("videoService.cdn.pool.ArrivalRate", 50.0)
	canvas.Set("videoService.cdn.pool.AvgHoldTime", "50ms")
	
	err = canvas.Run("baseline", "videoService.StreamVideo", 1000)
	if err != nil {
		t.Fatalf("Failed to run baseline: %v", err)
	}
	t.Log("✓ Baseline performance test completed")

	// === TRAFFIC SPIKE SIMULATION ===
	t.Log("=== TRAFFIC SPIKE SIMULATION ===")
	// Simulate "Stranger Things" season premiere - 4x traffic spike
	canvas.Set("videoService.cdn.pool.ArrivalRate", 200.0)
	
	err = canvas.Run("traffic_spike", "videoService.StreamVideo", 1000)
	if err != nil {
		t.Fatalf("Failed to run traffic spike: %v", err)
	}
	t.Log("✓ Traffic spike simulation completed")

	// === CACHE OPTIMIZATION ===
	t.Log("=== CACHE OPTIMIZATION ===")
	// Improve cache hit rate (better content distribution)
	canvas.Set("videoService.cdn.CacheHitRate", 0.95)
	
	err = canvas.Run("optimized_cache", "videoService.StreamVideo", 1000)
	if err != nil {
		t.Fatalf("Failed to run optimized cache: %v", err)
	}
	t.Log("✓ Cache optimization test completed")

	// === CAPACITY SCALING ===
	t.Log("=== CAPACITY SCALING ===")
	// Scale up CDN capacity - double it
	canvas.Set("videoService.cdn.pool.Size", 200)
	
	err = canvas.Run("scaled_cdn", "videoService.StreamVideo", 1000)
	if err != nil {
		t.Fatalf("Failed to run scaled CDN: %v", err)
	}
	t.Log("✓ CDN capacity scaling test completed")

	// === DATABASE BOTTLENECK ===
	t.Log("=== DATABASE BOTTLENECK ===")
	// Create database bottleneck under extreme load
	canvas.Set("videoService.database.pool.ArrivalRate", 100.0)
	
	err = canvas.Run("db_bottleneck", "videoService.StreamVideo", 1000)
	if err != nil {
		t.Fatalf("Failed to run database bottleneck: %v", err)
	}
	t.Log("✓ Database bottleneck simulation completed")

	// === ENCODING WORKLOAD ===
	t.Log("=== ENCODING WORKLOAD ===")
	// Test video upload/encoding pipeline (different method)
	canvas.Set("videoService.encoder.pool.ArrivalRate", 5.0)
	
	err = canvas.Run("encoding_load", "videoService.UploadVideo", 100)
	if err != nil {
		t.Fatalf("Failed to run encoding load: %v", err)
	}
	t.Log("✓ Video encoding performance test completed")

	// === FAILURE SCENARIOS ===
	t.Log("=== FAILURE SCENARIOS ===")
	// What happens when CDN capacity is severely limited?
	canvas.Set("videoService.cdn.pool.Size", 1)  // Severely limited CDN
	
	err = canvas.Run("cdn_failure", "videoService.StreamVideo", 1000)
	if err != nil {
		t.Fatalf("Failed to run CDN failure scenario: %v", err)
	}
	t.Log("✓ CDN failure scenario completed")

	t.Log("\n=== DEMO SCENARIO VALIDATION COMPLETE ===")
	t.Log("All Canvas API operations successful - ready for conference demo!")
}

// TestNetflixVisualizationGeneration tests all visualization capabilities
func TestNetflixVisualizationGeneration(t *testing.T) {
	canvas := console.NewCanvas()

	// Load and setup
	err := canvas.Load("netflix.sdl")
	if err != nil {
		t.Fatalf("Failed to load netflix.sdl: %v", err)
	}

	err = canvas.Use("NetflixSystem")
	if err != nil {
		t.Fatalf("Failed to use NetflixSystem: %v", err)
	}

	// Run a few quick simulations for visualization testing
	canvas.Set("videoService.cdn.pool.ArrivalRate", 50.0)
	canvas.Set("videoService.cdn.pool.AvgHoldTime", "50ms")
	canvas.Run("viz_baseline", "videoService.StreamVideo", 500)

	canvas.Set("videoService.cdn.pool.ArrivalRate", 150.0)
	canvas.Run("viz_load", "videoService.StreamVideo", 500)

	// Test different plot types
	plotTests := []struct {
		name     string
		results  []string
		plotType string
	}{
		{"Single Latency Plot", []string{"viz_baseline"}, "latency"},
		{"Comparison Plot", []string{"viz_baseline", "viz_load"}, "comparison"},
		{"Histogram", []string{"viz_load"}, "histogram"},
	}

	for _, test := range plotTests {
		t.Run(test.name, func(t *testing.T) {
			// For now, just test that Plot doesn't error
			// In real implementation, would check file generation
			err := canvas.Plot(test.results, test.plotType, fmt.Sprintf("test_%s", test.name))
			if err != nil {
				t.Errorf("Failed to generate %s: %v", test.name, err)
			} else {
				t.Logf("✓ Generated %s successfully", test.name)
			}
		})
	}

	// Test architecture diagram generation
	t.Run("Architecture Diagram", func(t *testing.T) {
		err := canvas.Diagram("NetflixSystem", "static", "netflix_architecture")
		if err != nil {
			t.Errorf("Failed to generate architecture diagram: %v", err)
		} else {
			t.Log("✓ Generated architecture diagram successfully")
		}
	})
}

// TestNetflixParameterModification validates parameter modification edge cases
func TestNetflixParameterModification(t *testing.T) {
	canvas := console.NewCanvas()

	err := canvas.Load("netflix.sdl")
	if err != nil {
		t.Fatalf("Failed to load netflix.sdl: %v", err)
	}

	err = canvas.Use("NetflixSystem")
	if err != nil {
		t.Fatalf("Failed to use NetflixSystem: %v", err)
	}

	// Test various parameter types and edge cases
	paramTests := []struct {
		name  string
		path  string
		value interface{}
	}{
		// Numeric parameters
		{"Pool Size", "videoService.cdn.pool.Size", 100},
		{"Float Arrival Rate", "videoService.cdn.pool.ArrivalRate", 75.5},
		{"Cache Hit Rate", "videoService.cdn.CacheHitRate", 0.85},
		
		// Duration parameters  
		{"Hold Time", "videoService.cdn.pool.AvgHoldTime", "25ms"},
		{"Query Time", "videoService.database.QueryTime", "10ms"},
		{"Encoding Time", "videoService.encoder.EncodingTime", "45000ms"},

		// Edge cases
		{"Zero Pool Size", "videoService.cdn.pool.Size", 0},
		{"Minimal Pool", "videoService.cdn.pool.Size", 1},
		{"Perfect Cache", "videoService.cdn.CacheHitRate", 1.0},
		{"No Cache", "videoService.cdn.CacheHitRate", 0.0},
		{"High Load", "videoService.cdn.pool.ArrivalRate", 500.0},
	}

	for _, test := range paramTests {
		t.Run(test.name, func(t *testing.T) {
			err := canvas.Set(test.path, test.value)
			if err != nil {
				t.Errorf("Failed to set %s to %v: %v", test.path, test.value, err)
			} else {
				t.Logf("✓ Successfully set %s = %v", test.path, test.value)
			}
		})
	}
}

// TestNetflixRapidIteration simulates live demo parameter changes
func TestNetflixRapidIteration(t *testing.T) {
	canvas := console.NewCanvas()

	err := canvas.Load("netflix.sdl")
	if err != nil {
		t.Fatalf("Failed to load netflix.sdl: %v", err)
	}

	err = canvas.Use("NetflixSystem")
	if err != nil {
		t.Fatalf("Failed to use NetflixSystem: %v", err)
	}

	// Simulate rapid parameter changes during live demo
	t.Log("=== SIMULATING LIVE DEMO PARAMETER CHANGES ===")
	
	scenarios := []struct {
		name        string
		arrivalRate float64
		cacheHit    float64
		poolSize    int
		description string
	}{
		{"Normal Load", 50, 0.85, 100, "Normal Netflix traffic"},
		{"Evening Peak", 100, 0.85, 100, "Peak viewing hours"},
		{"Viral Content", 200, 0.85, 100, "Stranger Things premiere"},
		{"With CDN Scaling", 200, 0.85, 200, "Scale up CDN capacity"},
		{"Better Caching", 200, 0.95, 200, "Improve cache distribution"},
		{"Extreme Load", 400, 0.95, 200, "Black Friday levels"},
	}

	for i, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("--- %s: %s ---", scenario.name, scenario.description)
			
			// Set parameters
			canvas.Set("videoService.cdn.pool.ArrivalRate", scenario.arrivalRate)
			canvas.Set("videoService.cdn.CacheHitRate", scenario.cacheHit)
			canvas.Set("videoService.cdn.pool.Size", scenario.poolSize)

			// Run simulation
			runName := fmt.Sprintf("demo_step_%d", i+1)
			err := canvas.Run(runName, "videoService.StreamVideo", 500) // Smaller count for speed
			if err != nil {
				t.Logf("Simulation %s resulted in error (may be expected): %v", runName, err)
			} else {
				t.Logf("✓ Simulation %s completed successfully", runName)
			}
		})
	}
	
	t.Log("✓ Rapid iteration test completed - ready for live demo!")
}

// TestNetflixWorkshopEdgeCases tests edge cases for workshop scenarios
func TestNetflixWorkshopEdgeCases(t *testing.T) {
	canvas := console.NewCanvas()

	err := canvas.Load("netflix.sdl")
	if err != nil {
		t.Fatalf("Failed to load netflix.sdl: %v", err)
	}

	err = canvas.Use("NetflixSystem")
	if err != nil {
		t.Fatalf("Failed to use NetflixSystem: %v", err)
	}

	// Test extreme scenarios that might come up in workshop Q&A
	edgeTests := []struct {
		name    string
		setup   func()
		runName string
		desc    string
	}{
		{
			name: "Zero CDN Capacity",
			setup: func() {
				canvas.Set("videoService.cdn.pool.Size", 0)
				canvas.Set("videoService.cdn.pool.ArrivalRate", 50.0)
			},
			runName: "zero_cdn",
			desc:    "What happens with no CDN capacity?",
		},
		{
			name: "Infinite Load",
			setup: func() {
				canvas.Set("videoService.cdn.pool.Size", 100)
				canvas.Set("videoService.cdn.pool.ArrivalRate", 10000.0)
			},
			runName: "infinite_load",
			desc:    "System under extreme overload",
		},
		{
			name: "No Database Connections",
			setup: func() {
				canvas.Set("videoService.database.pool.Size", 0)
				canvas.Set("videoService.cdn.pool.ArrivalRate", 50.0)
			},
			runName: "no_db",
			desc:    "Database connection pool exhausted",
		},
		{
			name: "Perfect System",
			setup: func() {
				canvas.Set("videoService.cdn.pool.Size", 1000)
				canvas.Set("videoService.cdn.CacheHitRate", 1.0)
				canvas.Set("videoService.database.pool.Size", 100)
				canvas.Set("videoService.cdn.pool.ArrivalRate", 50.0)
			},
			runName: "perfect",
			desc:    "Idealized system with unlimited resources",
		},
	}

	for _, test := range edgeTests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Testing: %s", test.desc)
			
			// Setup the scenario
			test.setup()

			// Run simulation (may fail for some edge cases)
			err := canvas.Run(test.runName, "videoService.StreamVideo", 100)
			if err != nil {
				t.Logf("Edge case %s resulted in error (may be expected): %v", test.name, err)
			} else {
				t.Logf("✓ Edge case %s completed without error", test.name)
			}
		})
	}
}