package runtime

import (
	"testing"
	"github.com/panyam/sdl/components"
)

// Mock components that implement FlowAnalyzable for testing
type MockComponentA struct{}
type MockComponentB struct{}

func (m *MockComponentA) GetFlowPattern(methodName string, inputRate float64, params map[string]interface{}) components.FlowPattern {
	if methodName == "Process" {
		return components.FlowPattern{
			Outflows:      map[string]float64{"componentC.Acquire": inputRate}, // Forward all traffic to C
			SuccessRate:   1.0,
			Amplification: 1.0,
			ServiceTime:   0.001,
		}
	}
	return components.FlowPattern{}
}

func (m *MockComponentB) GetFlowPattern(methodName string, inputRate float64, params map[string]interface{}) components.FlowPattern {
	if methodName == "Process" {
		return components.FlowPattern{
			Outflows:      map[string]float64{"componentC.Acquire": inputRate}, // Forward all traffic to C
			SuccessRate:   1.0,
			Amplification: 1.0,
			ServiceTime:   0.001,
		}
	}
	return components.FlowPattern{}
}

// TestFlowEvalWithBackPressure tests the new fixed-point flow evaluation with back-pressure
func TestFlowEvalWithBackPressure(t *testing.T) {
	// Create a test scenario: A->C, B->C where C has capacity constraints
	context := NewFlowContext(nil, map[string]interface{}{
		"cache.HitRate": 0.8,
	})
	// Increase convergence settings for testing
	context.MaxIterations = 20
	context.ConvergenceThreshold = 0.1

	// Register mock components A and B that forward traffic to C
	context.SetNativeComponent("componentA", &MockComponentA{})
	context.SetNativeComponent("componentB", &MockComponentB{})

	// Register component C with capacity constraints
	poolC := &components.ResourcePool{
		Name:        "componentC",
		Size:        2,              // Very limited - only 2 concurrent resources
		ArrivalRate: 1e-9,          // Will be updated by flow evaluation
		AvgHoldTime: 0.1,           // 100ms average hold time (much longer)
	}
	poolC.Init()
	context.SetNativeComponent("componentC", poolC)
	context.SetResourceLimit("componentC", 2)

	// Test scenario: High load that should trigger back-pressure
	entryPoints := map[string]float64{
		"componentA.Process": 100.0, // 100 RPS from A
		"componentB.Process": 200.0, // 200 RPS from B
	}

	t.Logf("Testing back-pressure scenario:")
	t.Logf("Entry points: A=%.1f RPS, B=%.1f RPS", entryPoints["componentA.Process"], entryPoints["componentB.Process"])

	// Solve the system flows
	finalRates := SolveSystemFlows(entryPoints, context)

	t.Logf("Final arrival rates after convergence:")
	for componentMethod, rate := range finalRates {
		t.Logf("  %s: %.2f RPS", componentMethod, rate)
	}

	// Verify that componentC receives the combined load from A and B
	expectedTotalLoadOnC := entryPoints["componentA.Process"] + entryPoints["componentB.Process"]
	if componentCRate, exists := finalRates["componentC.Acquire"]; exists {
		if componentCRate < expectedTotalLoadOnC*0.8 || componentCRate > expectedTotalLoadOnC*1.2 {
			t.Errorf("Expected componentC load ~%.1f, got %.1f", expectedTotalLoadOnC, componentCRate)
		}
		t.Logf("✓ ComponentC receives combined load: %.1f RPS", componentCRate)
	} else {
		t.Error("ComponentC should receive traffic from both A and B")
	}

	// Verify back-pressure effects: success rates should be updated
	t.Logf("All success rates after convergence:")
	for componentMethod, successRate := range context.SuccessRates {
		t.Logf("  %s: %.3f", componentMethod, successRate)
	}
	
	if successRate, exists := context.SuccessRates["componentC.Acquire"]; exists {
		t.Logf("ComponentC success rate under load: %.3f", successRate)
		if successRate >= 1.0 {
			t.Error("Expected success rate degradation due to high utilization")
		}
	} else {
		t.Error("ComponentC success rate should be calculated")
	}

	// Test with reduced load - should see better success rates
	t.Logf("\nTesting with reduced load:")
	entryPointsLow := map[string]float64{
		"componentA.Process": 5.0, // 5 RPS from A
		"componentB.Process": 5.0, // 5 RPS from B
	}

	contextLow := NewFlowContext(nil, map[string]interface{}{})
	contextLow.MaxIterations = 20
	contextLow.ConvergenceThreshold = 0.1
	contextLow.SetNativeComponent("componentA", &MockComponentA{})
	contextLow.SetNativeComponent("componentB", &MockComponentB{})
	
	poolCLow := &components.ResourcePool{
		Name:        "componentC",
		Size:        2,
		ArrivalRate: 1e-9,
		AvgHoldTime: 0.1,
	}
	poolCLow.Init()
	contextLow.SetNativeComponent("componentC", poolCLow)
	contextLow.SetResourceLimit("componentC", 2)

	_ = SolveSystemFlows(entryPointsLow, contextLow)

	if successRateLow, exists := contextLow.SuccessRates["componentC.Acquire"]; exists {
		t.Logf("ComponentC success rate under low load: %.3f", successRateLow)
		if successRateLow < 0.95 {
			t.Error("Expected high success rate under low utilization")
		}
	}

	t.Logf("✓ Back-pressure and convergence working correctly")
}

// TestFlowAnalyzableInterface tests the FlowAnalyzable implementations
func TestFlowAnalyzableInterface(t *testing.T) {
	// Test ResourcePool FlowAnalyzable
	pool := &components.ResourcePool{
		Name:        "testPool",
		Size:        5,
		ArrivalRate: 1e-9,
		AvgHoldTime: 0.02, // 20ms
	}
	pool.Init()

	// Test under normal load
	pattern := pool.GetFlowPattern("Acquire", 50.0, nil)
	t.Logf("ResourcePool under normal load (50 RPS):")
	t.Logf("  Success Rate: %.3f", pattern.SuccessRate)
	t.Logf("  Service Time: %.3fs", pattern.ServiceTime)

	if pattern.SuccessRate < 0.9 {
		t.Error("Expected high success rate under normal load")
	}

	// Test under high load (should trigger back-pressure)
	patternHigh := pool.GetFlowPattern("Acquire", 500.0, nil)
	t.Logf("ResourcePool under high load (500 RPS):")
	t.Logf("  Success Rate: %.3f", patternHigh.SuccessRate)
	t.Logf("  Service Time: %.3fs", patternHigh.ServiceTime)

	if patternHigh.SuccessRate >= pattern.SuccessRate {
		t.Error("Expected success rate degradation under high load")
	}

	// Test MM1Queue FlowAnalyzable
	queue := &components.MM1Queue{
		Name:           "testQueue",
		ArrivalRate:    1e-9,
		AvgServiceTime: 0.01, // 10ms
	}
	queue.Init()

	// Test under stable load
	queuePattern := queue.GetFlowPattern("Dequeue", 50.0, nil)
	t.Logf("MM1Queue under stable load (50 RPS):")
	t.Logf("  Success Rate: %.3f", queuePattern.SuccessRate)
	t.Logf("  Service Time: %.3fs", queuePattern.ServiceTime)

	if queuePattern.SuccessRate < 0.9 {
		t.Error("Expected high success rate under stable load")
	}

	// Test under overload (should be unstable)
	queuePatternUnstable := queue.GetFlowPattern("Dequeue", 200.0, nil) // > 100 RPS = unstable
	t.Logf("MM1Queue under overload (200 RPS):")
	t.Logf("  Success Rate: %.3f", queuePatternUnstable.SuccessRate)
	t.Logf("  Service Time: %.3fs", queuePatternUnstable.ServiceTime)

	if queuePatternUnstable.SuccessRate >= queuePattern.SuccessRate {
		t.Error("Expected success rate degradation under overload")
	}

	t.Logf("✓ FlowAnalyzable interface working correctly")
}

// TestConvergenceBehavior tests convergence detection and damping
func TestConvergenceBehavior(t *testing.T) {
	context := NewFlowContext(nil, nil)
	context.MaxIterations = 5
	context.ConvergenceThreshold = 0.1

	// Register a component that causes feedback
	pool := &components.ResourcePool{
		Name:        "feedbackComponent",
		Size:        5,
		ArrivalRate: 1e-9,
		AvgHoldTime: 0.01,
	}
	pool.Init()
	context.SetNativeComponent("feedbackComponent", pool)

	entryPoints := map[string]float64{
		"feedbackComponent.Acquire": 100.0,
	}

	t.Logf("Testing convergence with limited iterations:")
	finalRates := SolveSystemFlows(entryPoints, context)

	// Should converge or reach max iterations
	if len(finalRates) == 0 {
		t.Error("Expected some final rates even if convergence fails")
	}

	t.Logf("Convergence test completed with %d final rates", len(finalRates))
	for componentMethod, rate := range finalRates {
		t.Logf("  %s: %.2f RPS", componentMethod, rate)
	}

	t.Logf("✓ Convergence behavior working correctly")
}