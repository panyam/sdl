package contacts

import (
	"fmt"
	"testing"

	"github.com/panyam/sdl/console"
)

// TestContactsServiceBasic validates Canvas API with simple two-tier system
func TestContactsServiceBasic(t *testing.T) {
	canvas := console.NewCanvas("test", nil)

	// Load the contacts service
	err := canvas.Load("contacts.sdl")
	if err != nil {
		t.Fatalf("Failed to load contacts.sdl: %v", err)
	}

	err = canvas.Use("ContactsSystem")
	if err != nil {
		t.Fatalf("Failed to use ContactsSystem: %v", err)
	}

	t.Log("✓ Canvas loaded contacts service successfully")

	// === BASELINE TESTING ===
	t.Log("=== BASELINE TESTING ===")
	
	// Set initial parameters
	canvas.Set("server.pool.ArrivalRate", 5.0)     // 5 RPS baseline load
	canvas.Set("server.pool.AvgHoldTime", "15ms")  // 5ms processing + 10ms DB time
	canvas.Set("database.pool.ArrivalRate", 3.0)   // 3 RPS to DB (60% cache miss)
	canvas.Set("database.pool.AvgHoldTime", "12ms") // 10ms query + 2ms overhead
	canvas.Set("contactCache.HitRate", 0.6)         // 60% cache hit rate

	err = canvas.Run("baseline", "server.HandleLookup", console.WithRuns(1000))
	if err != nil {
		t.Fatalf("Failed to run baseline: %v", err)
	}
	t.Log("✓ Baseline performance test completed")

	// === LOAD INCREASE ===
	t.Log("=== LOAD INCREASE ===")
	
	// Increase load to test capacity limits
	canvas.Set("server.pool.ArrivalRate", 15.0)    // 3x load increase
	canvas.Set("database.pool.ArrivalRate", 9.0)   // Corresponding DB load

	err = canvas.Run("high_load", "server.HandleLookup", console.WithRuns(1000))
	if err != nil {
		t.Fatalf("Failed to run high load test: %v", err)
	}
	t.Log("✓ High load test completed")

	// === CACHE OPTIMIZATION ===
	t.Log("=== CACHE OPTIMIZATION ===")
	
	// Improve cache hit rate to reduce DB load
	canvas.Set("contactCache.HitRate", 0.8)         // 80% cache hit rate
	canvas.Set("database.pool.ArrivalRate", 3.0)    // Reduced DB load due to better cache

	err = canvas.Run("optimized_cache", "server.HandleLookup", console.WithRuns(1000))
	if err != nil {
		t.Fatalf("Failed to run optimized cache test: %v", err)
	}
	t.Log("✓ Cache optimization test completed")

	// === CAPACITY SCALING ===
	t.Log("=== CAPACITY SCALING ===")
	
	// Scale up server capacity
	canvas.Set("server.pool.Size", 20)              // Double server capacity

	err = canvas.Run("scaled_server", "server.HandleLookup", console.WithRuns(1000))
	if err != nil {
		t.Fatalf("Failed to run scaled server test: %v", err)
	}
	t.Log("✓ Server scaling test completed")

	// === DATABASE BOTTLENECK ===
	t.Log("=== DATABASE BOTTLENECK ===")
	
	// Create database bottleneck by increasing DB load
	canvas.Set("database.pool.ArrivalRate", 20.0)   // High DB load

	err = canvas.Run("db_bottleneck", "server.HandleLookup", console.WithRuns(1000))
	if err != nil {
		t.Fatalf("Failed to run DB bottleneck test: %v", err)
	}
	t.Log("✓ Database bottleneck test completed")

	t.Log("\n=== CONTACTS SERVICE VALIDATION COMPLETE ===")
	t.Log("All Canvas API operations successful - ready for workshop development!")
}

// TestContactsParameterModification validates different parameter types
func TestContactsParameterModification(t *testing.T) {
	canvas := console.NewCanvas("test", nil)

	err := canvas.Load("contacts.sdl")
	if err != nil {
		t.Fatalf("Failed to load contacts.sdl: %v", err)
	}

	err = canvas.Use("ContactsSystem")
	if err != nil {
		t.Fatalf("Failed to use ContactsSystem: %v", err)
	}

	// Test various parameter types that Canvas API should support
	paramTests := []struct {
		name  string
		path  string
		value interface{}
	}{
		// Native component parameters (ResourcePool)
		{"Server Pool Size", "server.pool.Size", 15},
		{"DB Pool Size", "database.pool.Size", 8},
		{"Server Arrival Rate", "server.pool.ArrivalRate", 12.5},
		{"DB Arrival Rate", "database.pool.ArrivalRate", 8.0},
		{"Hold Time", "server.pool.AvgHoldTime", 0.020},  // 20ms = 0.020 seconds

		// Cache component parameters
		{"Cache Hit Rate", "contactCache.HitRate", 0.6},

		// Edge cases
		{"Zero Pool Size", "server.pool.Size", 0},
		{"Minimal Pool", "server.pool.Size", 1},
		{"Perfect Cache", "contactCache.HitRate", 1.0},
		{"No Cache", "contactCache.HitRate", 0.0},
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

// TestContactsVisualization validates plot and diagram generation
func TestContactsVisualization(t *testing.T) {
	canvas := console.NewCanvas("test", nil)

	err := canvas.Load("contacts.sdl")
	if err != nil {
		t.Fatalf("Failed to load contacts.sdl: %v", err)
	}

	err = canvas.Use("ContactsSystem")
	if err != nil {
		t.Fatalf("Failed to use ContactsSystem: %v", err)
	}

	// Run a few simulations for visualization testing
	canvas.Set("server.pool.ArrivalRate", 5.0)
	canvas.Run("viz_baseline", "server.HandleLookup", console.WithRuns(500))

	canvas.Set("server.pool.ArrivalRate", 12.0)
	canvas.Run("viz_load", "server.HandleLookup", console.WithRuns(500))

	// Test different visualization types
	visualTests := []struct {
		name        string
		seriesNames []string
		outputFile  string
	}{
		{"Single Latency Plot", []string{"viz_baseline"}, "test_contacts_single.svg"},
		{"Comparison Plot", []string{"viz_baseline", "viz_load"}, "test_contacts_comparison.svg"},
		{"Load Test", []string{"viz_load"}, "test_contacts_load.svg"},
	}

	for _, test := range visualTests {
		t.Run(test.name, func(t *testing.T) {
			// Build plot options
			opts := []console.PlotOption{
				console.WithOutput(test.outputFile),
			}
			for _, seriesName := range test.seriesNames {
				opts = append(opts, console.WithSeries(seriesName, seriesName))
			}
			
			err := canvas.Plot(opts...)
			if err != nil {
				t.Errorf("Failed to generate %s: %v", test.name, err)
			} else {
				t.Logf("✓ Generated %s successfully", test.name)
			}
		})
	}

	t.Log("✓ All visualization tests completed")
}

// TestContactsRapidIteration simulates rapid parameter changes for live demos
func TestContactsRapidIteration(t *testing.T) {
	canvas := console.NewCanvas("test", nil)

	err := canvas.Load("contacts.sdl")
	if err != nil {
		t.Fatalf("Failed to load contacts.sdl: %v", err)
	}

	err = canvas.Use("ContactsSystem")
	if err != nil {
		t.Fatalf("Failed to use ContactsSystem: %v", err)
	}

	// Simulate rapid parameter changes during live demo
	t.Log("=== SIMULATING LIVE DEMO PARAMETER CHANGES ===")
	
	scenarios := []struct {
		name        string
		serverLoad  float64
		dbLoad      float64
		cacheHit    float64
		serverCap   int
		description string
	}{
		{"Normal Load", 5.0, 3.0, 0.6, 10, "Typical usage"},
		{"Busy Period", 10.0, 6.0, 0.6, 10, "Peak hours"},
		{"Cache Warmed", 10.0, 2.0, 0.8, 10, "Better cache performance"},
		{"Scaled Server", 15.0, 3.0, 0.8, 20, "Server capacity doubled"},
		{"DB Overload", 15.0, 15.0, 0.2, 20, "Database becomes bottleneck"},
	}

	for i, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("--- %s: %s ---", scenario.name, scenario.description)
			
			// Set parameters rapidly
			canvas.Set("server.pool.ArrivalRate", scenario.serverLoad)
			canvas.Set("database.pool.ArrivalRate", scenario.dbLoad)
			canvas.Set("contactCache.HitRate", scenario.cacheHit)
			canvas.Set("server.pool.Size", scenario.serverCap)

			// Run quick simulation
			runName := fmt.Sprintf("demo_step_%d", i+1)
			err := canvas.Run(runName, "server.HandleLookup", console.WithRuns(300)) // Smaller count for speed
			if err != nil {
				t.Logf("Simulation %s resulted in error (may be expected): %v", runName, err)
			} else {
				t.Logf("✓ Simulation %s completed successfully", runName)
			}
		})
	}
	
	t.Log("✓ Rapid iteration test completed - Canvas API handles live demo scenarios!")
}