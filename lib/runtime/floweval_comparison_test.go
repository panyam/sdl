package runtime

import (
	"math"
	"testing"

	compdecl "github.com/panyam/sdl/lib/components/decl"
	"github.com/panyam/sdl/lib/loader"
)

// TestFlowEvalComparison runs the same scenarios through both string-based and runtime-based
// FlowEval to ensure they produce identical results
func TestFlowEvalComparison(t *testing.T) {
	t.Run("Simple Native Component", func(t *testing.T) {
		// Create a ResourcePool component
		pool := compdecl.NewResourcePool("testpool")
		pool.Wrapped.Size = 5
		pool.Wrapped.ArrivalRate = 1e-9
		pool.Wrapped.AvgHoldTime = 0.02
		pool.Wrapped.Init()

		// String-based eval setup
		system := &SystemDecl{
			Name: &IdentifierExpr{Value: "TestSystem"},
		}
		stringContext := NewFlowContext(system, map[string]interface{}{})
		stringContext.NativeComponents["testpool"] = pool

		// Runtime-based eval setup
		runtimeScope := NewFlowScope(nil)
		compInst := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: pool,
			},
			id: "testpool",
		}

		// Test with various rates
		testRates := []float64{10.0, 50.0, 100.0, 200.0}
		for _, rate := range testRates {
			// String-based eval
			stringFlows := FlowEval("testpool", "Acquire", rate, stringContext)

			// Runtime-based eval
			runtimeFlows := FlowEvalRuntime(compInst, "Acquire", rate, runtimeScope)

			// Compare results - both should return empty maps for ResourcePool
			stringTotal := sumMapValues(stringFlows)
			runtimeTotal := runtimeFlows.GetTotalRate()

			if !floatEquals(stringTotal, runtimeTotal) {
				t.Errorf("Rate %f: String total=%f, Runtime total=%f",
					rate, stringTotal, runtimeTotal)
			}
		}
	})

	t.Run("Component with Outflows", func(t *testing.T) {
		// Create a mock component that has outflows
		mockFlows := map[string]float64{
			"db.Query":  0.8,
			"cache.Get": 0.2,
		}
		mockComp := compdecl.NewMockFlowComponent("service", mockFlows)

		// String-based setup
		system := &SystemDecl{
			Name: &IdentifierExpr{Value: "TestSystem"},
		}
		stringContext := NewFlowContext(system, map[string]interface{}{})
		stringContext.NativeComponents["service"] = mockComp

		// Runtime-based setup
		runtimeScope := NewFlowScope(nil)
		compInst := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: mockComp,
			},
			id: "service",
		}

		// Create mock target components for the runtime resolver
		dbComp := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: compdecl.NewMockFlowComponent("db", map[string]float64{}),
			},
			id: "db",
		}
		cacheComp := &ComponentInstance{
			ObjectInstance: ObjectInstance{
				IsNative:       true,
				NativeInstance: compdecl.NewMockFlowComponent("cache", map[string]float64{}),
			},
			id: "cache",
		}

		// Register components in scope so they can be resolved
		runtimeScope.SysEnv.Set("db", Value{Value: dbComp})
		runtimeScope.SysEnv.Set("cache", Value{Value: cacheComp})

		// Test
		inputRate := 100.0
		stringFlows := FlowEval("service", "Process", inputRate, stringContext)
		runtimeFlows := FlowEvalRuntime(compInst, "Process", inputRate, runtimeScope)

		// Compare total rates
		stringTotal := sumMapValues(stringFlows)
		runtimeTotal := runtimeFlows.GetTotalRate()

		if !floatEquals(stringTotal, runtimeTotal) {
			t.Errorf("Total rates differ: String=%f, Runtime=%f",
				stringTotal, runtimeTotal)
		}

		// Compare individual flows
		for target, expectedRate := range mockFlows {
			stringRate := stringFlows[target]
			// For runtime, we need to check the actual flow pattern
			pattern := compInst.GetFlowPattern("Process", inputRate)
			runtimeRate := pattern.Outflows[target]

			expectedActualRate := inputRate * expectedRate
			if !floatEquals(stringRate, expectedActualRate) {
				t.Errorf("String flow to %s: expected %f, got %f",
					target, expectedActualRate, stringRate)
			}
			if !floatEquals(runtimeRate, expectedActualRate) {
				t.Errorf("Runtime flow to %s: expected %f, got %f",
					target, expectedActualRate, runtimeRate)
			}
		}
	})

	t.Run("System Flow Solve Comparison", func(t *testing.T) {
		// Create a simple A -> B system
		mockA := compdecl.NewMockFlowComponent("compA", map[string]float64{
			"compB.Process": 1.0,
		})
		mockB := compdecl.NewMockFlowComponent("compB", map[string]float64{})

		// String-based setup
		system := &SystemDecl{
			Name: &IdentifierExpr{Value: "TestSystem"},
		}
		stringContext := NewFlowContext(system, map[string]interface{}{})
		stringContext.NativeComponents["compA"] = mockA
		stringContext.NativeComponents["compB"] = mockB

		// String-based solve
		stringEntryPoints := map[string]float64{
			"compA.Process": 50.0,
		}
		stringResult := SolveSystemFlows(stringEntryPoints, stringContext)

		// Runtime-based setup
		runtimeScope := NewFlowScope(nil)
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

		// Register components in scope
		runtimeScope.SysEnv.Set("compA", Value{Value: compA})
		runtimeScope.SysEnv.Set("compB", Value{Value: compB})

		// Runtime-based solve
		runtimeGenerators := []GeneratorEntryPointRuntime{
			{
				Component:   compA,
				Method:      "Process",
				Rate:        50.0,
				GeneratorID: "gen1",
			},
		}
		runtimeResult := SolveSystemFlowsRuntime(runtimeGenerators, runtimeScope)

		// Compare results
		// String-based uses "component.method" keys
		stringRateA := stringResult["compA.Process"]
		stringRateB := stringResult["compB.Process"]

		// Runtime-based uses ComponentInstance
		runtimeRateA := runtimeResult.GetRate(compA, "Process")
		runtimeRateB := runtimeResult.GetRate(compB, "Process")

		if !floatEquals(stringRateA, runtimeRateA) {
			t.Errorf("CompA rates differ: String=%f, Runtime=%f", stringRateA, runtimeRateA)
		}
		if !floatEquals(stringRateB, runtimeRateB) {
			t.Errorf("CompB rates differ: String=%f, Runtime=%f", stringRateB, runtimeRateB)
		}

		// Both should show ~50.0 for each component (allowing for convergence differences)
		if math.Abs(runtimeRateA-50.0) > 0.1 {
			t.Errorf("CompA rate should be ~50.0, got %f", runtimeRateA)
		}
		if math.Abs(runtimeRateB-50.0) > 0.1 {
			t.Errorf("CompB rate should be ~50.0, got %f", runtimeRateB)
		}
	})
}

// TestContactsSystemComparison tests the actual contacts.sdl example with both implementations
func TestContactsSystemComparison(t *testing.T) {
	// Load the contacts system
	l := loader.NewLoader(nil, nil, 10)
	fileStatus, err := l.LoadFile("../examples/contacts/contacts.sdl", "", 0)
	if err != nil {
		t.Fatalf("Failed to load contacts.sdl: %v", err)
	}

	systemDecls, err := fileStatus.FileDecl.GetSystems()
	if err != nil {
		t.Fatalf("Failed to get system declarations: %v", err)
	}

	systemDecl := systemDecls["ContactsSystem"]
	if systemDecl == nil {
		t.Fatalf("ContactsSystem not found")
	}

	// For now, we'll skip the actual comparison since we need to complete
	// the SDL component instantiation logic first
	t.Skip("ContactsSystem comparison pending SDL component support")
}

// Helper functions
func floatEquals(a, b float64) bool {
	const epsilon = 1e-9
	return math.Abs(a-b) < epsilon
}

func sumMapValues(m map[string]float64) float64 {
	total := 0.0
	for _, v := range m {
		total += v
	}
	return total
}
