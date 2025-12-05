package runtime

import (
	"log/slog"
	"os"
	"testing"

	compdecl "github.com/panyam/sdl/lib/components/decl"
)

// TestFlowEvalWithDelaySDL tests flow evaluation with SDL files that have delays
func TestFlowEvalWithDelaySDL(t *testing.T) {
	t.Run("Queue buildup example", func(t *testing.T) {
		sys, _ := loadSystem(t, "../examples/delays/queue_buildup.sdl", "QueueBuildupDemo")
		env := sys.Env

		// Get processor component
		processorVal, _ := env.Get("processor")
		processor := processorVal.Value.(*ComponentInstance)

		// Create flow scope
		scope := NewFlowScope(env)

		// Test at different load levels
		testCases := []struct {
			name     string
			rate     float64
			expected string
		}{
			{"Light load (50 RPS)", 50.0, "Should handle fine"},
			{"At capacity (100 RPS)", 100.0, "At limit"},
			{"Overloaded (200 RPS)", 200.0, "Should see degradation"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				generators := []GeneratorEntryPointRuntime{{
					Component:   processor,
					Method:      "Process",
					Rate:        tc.rate,
					GeneratorID: "test",
				}}

				result := SolveSystemFlowsRuntime(generators, scope)
				rate := result.GetRate(processor, "Process")

				t.Logf("%s: Arrival rate %.0f RPS", tc.name, rate)

				// Currently we just verify flow propagates
				// In real system, we'd check success rates drop at high load
				if rate < tc.rate*0.9 || rate > tc.rate*1.1 {
					t.Errorf("Expected rate ~%.0f, got %.2f", tc.rate, rate)
				}
			})
		}
	})

	t.Run("Cascading delays example", func(t *testing.T) {
		sys, _ := loadSystem(t, "../examples/delays/cascading_delays.sdl", "CascadingDelayDemo")
		env := sys.Env

		// Get components
		apiVal, _ := env.Get("api")
		api := apiVal.Value.(*ComponentInstance)

		backendVal, _ := env.Get("backend")
		backend := backendVal.Value.(*ComponentInstance)

		dbVal, _ := env.Get("db")
		db := dbVal.Value.(*ComponentInstance)

		cacheVal, _ := env.Get("cache")
		cache := cacheVal.Value.(*ComponentInstance)

		// Debug: Check cache configuration
		t.Logf("Cache IsNative: %v, Type: %T", cache.IsNative, cache.NativeInstance)
		if cache.IsNative {
			// Import the decl package to access wrapped components
			if declCache, ok := cache.NativeInstance.(*compdecl.Cache); ok {
				t.Logf("Cache HitRate: %.2f", declCache.Wrapped.HitRate)
			} else {
				t.Logf("Cache is not *compdecl.Cache, it's %T", cache.NativeInstance)
			}
		}

		// Create flow scope
		scope := NewFlowScope(env)

		// Test flow propagation
		generators := []GeneratorEntryPointRuntime{{
			Component:   api,
			Method:      "HandleRequest",
			Rate:        100.0,
			GeneratorID: "test",
		}}

		// Enable debug logging for this test
		oldLogLevel := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
		defer slog.SetDefault(oldLogLevel)

		result := SolveSystemFlowsRuntime(generators, scope)

		// Check flow rates
		apiRate := result.GetRate(api, "HandleRequest")
		backendRate := result.GetRate(backend, "Process")
		dbRate := result.GetRate(db, "Query")
		cacheReadRate := result.GetRate(cache, "Read")
		cacheWriteRate := result.GetRate(cache, "Write")

		t.Logf("Flow rates:")
		t.Logf("  API: %.2f RPS", apiRate)
		t.Logf("  Backend: %.2f RPS", backendRate)
		t.Logf("  Database: %.2f RPS (should be ~20%% of traffic)", dbRate)
		t.Logf("  Cache Read: %.2f RPS", cacheReadRate)
		t.Logf("  Cache Write: %.2f RPS", cacheWriteRate)

		// Verify cascade
		if backendRate < 95 || backendRate > 105 {
			t.Errorf("Expected backend rate ~100, got %.2f", backendRate)
		}

		// With 80% cache hit rate, DB should see ~20% of traffic
		if dbRate < 15 || dbRate > 25 {
			t.Errorf("Expected DB rate ~20 (20%% of 100), got %.2f", dbRate)
		}

		// Count delay calls during flow analysis
		delays := 0
		for _, edge := range scope.FlowEdges.GetEdges() {
			t.Logf("Flow edge: %s.%s -> %s.%s (%.2f RPS)",
				edge.FromComponent.ID(), edge.FromMethod,
				edge.ToComponent.ID(), edge.ToMethod,
				edge.Rate)
			delays++
		}
	})
}
