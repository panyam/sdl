package runtime

import (
	"testing"

	compdecl "github.com/panyam/sdl/lib/components/decl"
)

// TestFlowPropagation tests that flow evaluation correctly propagates traffic downstream
func TestFlowPropagation(t *testing.T) {
	t.Run("Simple Chain A->B->C", func(t *testing.T) {
		// Create three components in a chain: A calls B, B calls C
		mockA := compdecl.NewMockFlowComponent("compA", map[string]float64{
			"compB.Process": 1.0, // 100% of traffic goes to B
		})
		mockB := compdecl.NewMockFlowComponent("compB", map[string]float64{
			"compC.Process": 1.0, // 100% of traffic goes to C
		})
		mockC := compdecl.NewMockFlowComponent("compC", map[string]float64{}) // Leaf component

		// Create component instances
		compA := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockA,
			},
			id: "compA",
		}
		compB := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockB,
			},
			id: "compB",
		}
		compC := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockC,
			},
			id: "compC",
		}

		// Create scope and register components
		scope := NewFlowScope(nil)
		scope.SysEnv.Set("compA", Value{Value: compA})
		scope.SysEnv.Set("compB", Value{Value: compB})
		scope.SysEnv.Set("compC", Value{Value: compC})

		// Create generators starting at A
		generators := []GeneratorEntryPointRuntime{
			{
				Component:   compA,
				Method:      "Process",
				Rate:        100.0,
				GeneratorID: "gen1",
			},
		}

		// Solve the system
		result := SolveSystemFlowsRuntime(generators, scope)

		// Verify flow propagation
		rateA := result.GetRate(compA, "Process")
		rateB := result.GetRate(compB, "Process")
		rateC := result.GetRate(compC, "Process")

		t.Logf("Flow rates: A=%.2f, B=%.2f, C=%.2f", rateA, rateB, rateC)

		// All should receive 100 RPS (or close to it with convergence)
		if rateA < 99.0 || rateA > 101.0 {
			t.Errorf("Expected compA rate ~100.0, got %f", rateA)
		}
		if rateB < 99.0 || rateB > 101.0 {
			t.Errorf("Expected compB rate ~100.0 (from A), got %f", rateB)
		}
		if rateC < 99.0 || rateC > 101.0 {
			t.Errorf("Expected compC rate ~100.0 (from B), got %f", rateC)
		}
	})

	t.Run("Fan-out Pattern", func(t *testing.T) {
		// Create a fan-out pattern: A calls both B and C
		mockA := compdecl.NewMockFlowComponent("compA", map[string]float64{
			"compB.Process": 0.7, // 70% to B
			"compC.Process": 0.3, // 30% to C
		})
		mockB := compdecl.NewMockFlowComponent("compB", map[string]float64{})
		mockC := compdecl.NewMockFlowComponent("compC", map[string]float64{})

		// Create component instances
		compA := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockA,
			},
			id: "compA",
		}
		compB := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockB,
			},
			id: "compB",
		}
		compC := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockC,
			},
			id: "compC",
		}

		// Create scope and register components
		scope := NewFlowScope(nil)
		scope.SysEnv.Set("compA", Value{Value: compA})
		scope.SysEnv.Set("compB", Value{Value: compB})
		scope.SysEnv.Set("compC", Value{Value: compC})

		// Create generators
		generators := []GeneratorEntryPointRuntime{
			{
				Component:   compA,
				Method:      "Process",
				Rate:        100.0,
				GeneratorID: "gen1",
			},
		}

		// Solve the system
		result := SolveSystemFlowsRuntime(generators, scope)

		// Verify flow distribution
		rateA := result.GetRate(compA, "Process")
		rateB := result.GetRate(compB, "Process")
		rateC := result.GetRate(compC, "Process")

		t.Logf("Flow rates: A=%.2f, B=%.2f, C=%.2f", rateA, rateB, rateC)

		// Check expected rates
		if rateA < 99.0 || rateA > 101.0 {
			t.Errorf("Expected compA rate ~100.0, got %f", rateA)
		}
		if rateB < 69.0 || rateB > 71.0 {
			t.Errorf("Expected compB rate ~70.0 (70%% of A), got %f", rateB)
		}
		if rateC < 29.0 || rateC > 31.0 {
			t.Errorf("Expected compC rate ~30.0 (30%% of A), got %f", rateC)
		}
	})

	t.Run("Component Name Resolution", func(t *testing.T) {
		// Test that component names are resolved correctly
		// This simulates the contacts.sdl scenario where we have named components

		// Create components with realistic names
		mockServer := compdecl.NewMockFlowComponent("server", map[string]float64{
			"database.Query": 0.8,
			"cache.Get":      0.2,
		})
		mockDatabase := compdecl.NewMockFlowComponent("database", map[string]float64{})
		mockCache := compdecl.NewMockFlowComponent("cache", map[string]float64{})

		// Create component instances with IDs
		server := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockServer,
			},
			id: "comp:1", // Runtime ID
		}
		database := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockDatabase,
			},
			id: "comp:2", // Runtime ID
		}
		cache := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockCache,
			},
			id: "comp:3", // Runtime ID
		}

		// Create scope and register with variable names
		scope := NewFlowScope(nil)
		scope.SysEnv.Set("server", Value{Value: server})
		scope.SysEnv.Set("database", Value{Value: database})
		scope.SysEnv.Set("cache", Value{Value: cache})

		// Create generators
		generators := []GeneratorEntryPointRuntime{
			{
				Component:   server,
				Method:      "HandleRequest",
				Rate:        100.0,
				GeneratorID: "gen1",
			},
		}

		// Solve the system
		result := SolveSystemFlowsRuntime(generators, scope)

		// Log all rates for debugging
		t.Logf("Server (comp:1) rate: %.2f", result.GetRate(server, "HandleRequest"))
		t.Logf("Database (comp:2) rate: %.2f", result.GetRate(database, "Query"))
		t.Logf("Cache (comp:3) rate: %.2f", result.GetRate(cache, "Get"))

		// Verify that database and cache receive traffic
		dbRate := result.GetRate(database, "Query")
		cacheRate := result.GetRate(cache, "Get")

		if dbRate < 79.0 || dbRate > 81.0 {
			t.Errorf("Expected database rate ~80.0, got %f", dbRate)
		}
		if cacheRate < 19.0 || cacheRate > 21.0 {
			t.Errorf("Expected cache rate ~20.0, got %f", cacheRate)
		}
	})
}
