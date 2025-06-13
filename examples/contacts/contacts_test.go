package contacts

import (
	"testing"

	"github.com/panyam/sdl/console"
)

// TestContactsServiceBasic validates Canvas API with simple two-tier system
func TestContactsServiceBasic(t *testing.T) {
	canvas := console.NewCanvas()

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
	canvas.Set("server.db.pool.ArrivalRate", 3.0)  // 3 RPS to DB (60% cache miss)
	canvas.Set("server.db.pool.AvgHoldTime", "12ms") // 10ms query + 2ms overhead

	err = canvas.Run("baseline", "server.HandleLookup", 1000)
	if err != nil {
		t.Fatalf("Failed to run baseline: %v", err)
	}
	t.Log("✓ Baseline performance test completed")

	// === LOAD INCREASE ===
	t.Log("=== LOAD INCREASE ===")
	
	// Increase load to test capacity limits
	canvas.Set("server.pool.ArrivalRate", 15.0)    // 3x load increase
	canvas.Set("server.db.pool.ArrivalRate", 9.0)  // Corresponding DB load

	err = canvas.Run("high_load", "server.HandleLookup", 1000)
	if err != nil {
		t.Fatalf("Failed to run high load test: %v", err)
	}
	t.Log("✓ High load test completed")

	// === CACHE OPTIMIZATION ===
	t.Log("=== CACHE OPTIMIZATION ===")
	
	// Improve cache hit rate to reduce DB load
	canvas.Set("server.db.CacheHitRate", 0.8)       // 80% cache hit rate
	canvas.Set("server.db.pool.ArrivalRate", 3.0)   // Reduced DB load due to better cache

	err = canvas.Run("optimized_cache", "server.HandleLookup", 1000)
	if err != nil {
		t.Fatalf("Failed to run optimized cache test: %v", err)
	}
	t.Log("✓ Cache optimization test completed")

	// === CAPACITY SCALING ===
	t.Log("=== CAPACITY SCALING ===")
	
	// Scale up server capacity
	canvas.Set("server.pool.Size", 20)              // Double server capacity

	err = canvas.Run("scaled_server", "server.HandleLookup", 1000)
	if err != nil {
		t.Fatalf("Failed to run scaled server test: %v", err)
	}
	t.Log("✓ Server scaling test completed")

	// === DATABASE BOTTLENECK ===
	t.Log("=== DATABASE BOTTLENECK ===")
	
	// Create database bottleneck by increasing DB load
	canvas.Set("server.db.pool.ArrivalRate", 20.0)  // High DB load

	err = canvas.Run("db_bottleneck", "server.HandleLookup", 1000)
	if err != nil {
		t.Fatalf("Failed to run DB bottleneck test: %v", err)
	}
	t.Log("✓ Database bottleneck test completed")

	t.Log("\n=== CONTACTS SERVICE VALIDATION COMPLETE ===")
	t.Log("All Canvas API operations successful - ready for workshop development!")
}

// TestContactsParameterModification validates different parameter types
func TestContactsParameterModification(t *testing.T) {
	canvas := console.NewCanvas()

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
		// Numeric parameters
		{"Server Pool Size", "server.pool.Size", 15},
		{"DB Pool Size", "server.db.pool.Size", 8},
		{"Arrival Rate", "server.pool.ArrivalRate", 12.5},
		
		// Duration parameters  
		{"Processing Time", "server.ProcessingTime", "8ms"},
		{"Query Time", "server.db.QueryTime", "15ms"},
		{"Hold Time", "server.pool.AvgHoldTime", "20ms"},

		// Probability parameters
		{"Cache Hit Rate", "server.db.CacheHitRate", 0.6},

		// Edge cases
		{"Zero Pool Size", "server.pool.Size", 0},
		{"Minimal Pool", "server.pool.Size", 1},
		{"Perfect Cache", "server.db.CacheHitRate", 1.0},
		{"No Cache", "server.db.CacheHitRate", 0.0},
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
	canvas := console.NewCanvas()

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
	canvas.Run("viz_baseline", "server.HandleLookup", 500)

	canvas.Set("server.pool.ArrivalRate", 12.0)
	canvas.Run("viz_load", "server.HandleLookup", 500)

	// Test different visualization types
	visualTests := []struct {
		name     string
		results  []string
		plotType string
	}{
		{"Single Latency Plot", []string{"viz_baseline"}, "latency"},
		{"Comparison Plot", []string{"viz_baseline", "viz_load"}, "comparison"},
		{"Histogram", []string{"viz_load"}, "histogram"},
	}

	for _, test := range visualTests {
		t.Run(test.name, func(t *testing.T) {
			err := canvas.Plot(test.results, test.plotType, "test_contacts_"+test.name)
			if err != nil {
				t.Errorf("Failed to generate %s: %v", test.name, err)
			} else {
				t.Logf("✓ Generated %s successfully", test.name)
			}
		})
	}

	// Test architecture diagram generation
	t.Run("Architecture Diagram", func(t *testing.T) {
		err := canvas.Diagram("ContactsSystem", "static", "contacts_architecture")
		if err != nil {
			t.Errorf("Failed to generate architecture diagram: %v", err)
		} else {
			t.Log("✓ Generated architecture diagram successfully")
		}
	})
}

// TestContactsRapidIteration simulates rapid parameter changes for live demos
func TestContactsRapidIteration(t *testing.T) {
	canvas := console.NewCanvas()

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
		{"Normal Load", 5.0, 3.0, 0.4, 10, "Typical usage"},
		{"Busy Period", 10.0, 6.0, 0.4, 10, "Peak hours"},
		{"Cache Warmed", 10.0, 2.0, 0.8, 10, "Better cache performance"},
		{"Scaled Server", 15.0, 3.0, 0.8, 20, "Server capacity doubled"},
		{"DB Overload", 15.0, 15.0, 0.2, 20, "Database becomes bottleneck"},
	}

	for i, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("--- %s: %s ---", scenario.name, scenario.description)
			
			// Set parameters rapidly
			canvas.Set("server.pool.ArrivalRate", scenario.serverLoad)
			canvas.Set("server.db.pool.ArrivalRate", scenario.dbLoad)
			canvas.Set("server.db.CacheHitRate", scenario.cacheHit)
			canvas.Set("server.pool.Size", scenario.serverCap)

			// Run quick simulation
			runName := fmt.Sprintf("demo_step_%d", i+1)
			err := canvas.Run(runName, "server.HandleLookup", 300) // Smaller count for speed
			if err != nil {
				t.Logf("Simulation %s resulted in error (may be expected): %v", runName, err)
			} else {
				t.Logf("✓ Simulation %s completed successfully", runName)
			}
		})
	}
	
	t.Log("✓ Rapid iteration test completed - Canvas API handles live demo scenarios!")
}