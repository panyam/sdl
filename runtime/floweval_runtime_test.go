package runtime

import (
	"testing"

	compdecl "github.com/panyam/sdl/components/decl"
)

func TestFlowEvalRuntime(t *testing.T) {
	t.Run("Native Component", func(t *testing.T) {
		// Create a flow scope
		scope := NewFlowScope(nil)

		// Create a native component instance (ResourcePool)
		pool := compdecl.NewResourcePool("testpool")
		pool.Wrapped.Size = 5
		pool.Wrapped.ArrivalRate = 1e-9
		pool.Wrapped.AvgHoldTime = 0.02
		pool.Wrapped.Init()

		// Create a ComponentInstance wrapper
		compInst := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: pool, // ResourcePool implements components.FlowAnalyzable
			},
			id: "testPool",
		}

		// Test FlowEvalRuntime
		inputRate := 50.0
		outflows := FlowEvalRuntime(compInst, "Acquire", inputRate, scope)

		// For ResourcePool, we expect no outflows (it's a leaf component)
		if outflows.GetTotalRate() != 0 {
			t.Errorf("Expected no outflows from ResourcePool, got total rate: %f", outflows.GetTotalRate())
		}

		// Verify the component was called via GetFlowPattern
		// We can check this indirectly by looking at the pattern it would return
		pattern := compInst.GetFlowPattern("Acquire", inputRate)
		if pattern.SuccessRate <= 0 {
			t.Errorf("Expected positive success rate, got: %f", pattern.SuccessRate)
		}
	})

	t.Run("Nil Handling", func(t *testing.T) {
		scope := NewFlowScope(nil)

		// Test with nil component
		outflows := FlowEvalRuntime(nil, "method", 10.0, scope)
		if outflows.GetTotalRate() != 0 {
			t.Errorf("Expected empty RateMap for nil component")
		}

		// Test with empty method
		compInst := &ComponentInstance{id: "test"}
		outflows = FlowEvalRuntime(compInst, "", 10.0, scope)
		if outflows.GetTotalRate() != 0 {
			t.Errorf("Expected empty RateMap for empty method")
		}

		// Test with nil scope
		outflows = FlowEvalRuntime(compInst, "method", 10.0, nil)
		if outflows.GetTotalRate() != 0 {
			t.Errorf("Expected empty RateMap for nil scope")
		}
	})

	t.Run("Cycle Detection", func(t *testing.T) {
		scope := NewFlowScope(nil)

		// Create a component and add it to call stack
		compInst := &ComponentInstance{id: "comp1"}
		scope.CallStack = append(scope.CallStack, compInst)

		// Try to evaluate the same component (should detect cycle)
		outflows := FlowEvalRuntime(compInst, "method", 10.0, scope)
		if outflows.GetTotalRate() != 0 {
			t.Errorf("Expected empty RateMap due to cycle detection")
		}
	})
}

// FlowEvaluator is a placeholder for now
type FlowEvaluator struct {
	// TODO: Add fields as needed
}

func TestSolveSystemFlowsRuntime(t *testing.T) {
	t.Run("Basic Entry Point", func(t *testing.T) {
		// Create environment and scope
		scope := NewFlowScope(nil)

		// Create a native component
		pool := compdecl.NewResourcePool("testpool")
		pool.Wrapped.Size = 5
		pool.Wrapped.Init()

		compInst := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: pool,
			},
			id: "testPool",
		}

		// Create generators
		generators := []GeneratorEntryPointRuntime{
			{
				Component:   compInst,
				Method:      "Acquire",
				Rate:        10.0,
				GeneratorID: "gen1",
			},
		}

		// Solve the system
		result := SolveSystemFlowsRuntime(generators, scope)

		// Verify the entry point rate was set
		if rate := result.GetRate(compInst, "Acquire"); rate != 10.0 {
			t.Errorf("Expected rate 10.0, got %f", rate)
		}

		// Verify total system rate
		if total := result.GetTotalRate(); total != 10.0 {
			t.Errorf("Expected total rate 10.0, got %f", total)
		}
	})

	t.Run("Multiple Components with Flow", func(t *testing.T) {
		// This test simulates A -> B flow pattern
		scope := NewFlowScope(nil)

		// Create component A that flows to B
		mockA := compdecl.NewMockFlowComponent("compA", map[string]float64{
			"compB.Process": 1.0, // 100% of traffic goes to B
		})

		compA := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockA,
			},
			id: "compA",
		}

		// Create component B (leaf component)
		mockB := compdecl.NewMockFlowComponent("compB", map[string]float64{}) // No outflows

		compB := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockB,
			},
			id: "compB",
		}

		// Register component B in the scope so it can be resolved
		// In a real system, this would be done by the environment setup
		scope.SysEnv.Set("compB", Value{Value: compB})

		// Create generators starting at A
		generators := []GeneratorEntryPointRuntime{
			{
				Component:   compA,
				Method:      "Process",
				Rate:        50.0,
				GeneratorID: "gen1",
			},
		}

		// Solve the system
		result := SolveSystemFlowsRuntime(generators, scope)

		// Verify A receives the entry point rate
		if rate := result.GetRate(compA, "Process"); rate != 50.0 {
			t.Errorf("Expected compA rate 50.0, got %f", rate)
		}

		// Verify B receives flow from A (with tolerance for convergence)
		rateB := result.GetRate(compB, "Process")
		if rateB < 49.9 || rateB > 50.1 {
			t.Errorf("Expected compB rate ~50.0 (from A), got %f", rateB)
		}

		// Verify total system rate (with tolerance)
		total := result.GetTotalRate()
		if total < 99.9 || total > 100.1 {
			t.Errorf("Expected total rate ~100.0, got %f", total)
		}
	})
}
