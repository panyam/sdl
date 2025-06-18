package runtime

import (
	"testing"
	"time"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
)

func TestMetricStore(t *testing.T) {
	// Load and initialize the ContactsSystem
	system, env, currTime := loadSystem(t, "../examples/contacts/contacts.sdl", "ContactsSystem")

	// Create a runtime with metrics enabled
	rt := system.File.Runtime
	rt.EnableMetrics(1000)
	rt.metricStore.SetSystem(system)

	// Create tracer and evaluator with metrics enabled
	tracer := NewExecutionTracer()
	tracer.SetRuntime(rt)
	eval := NewSimpleEval(system.File, tracer)

	// Get the server component from the environment
	serverVal, ok := env.Get("server")
	if !ok {
		t.Fatal("Failed to get server component from environment")
	}
	serverComp := serverVal.Value.(*ComponentInstance)

	// Create a measurement for server lookup successes (since all are returning true)
	spec := MeasurementSpec{
		ID:          "server_lookups",
		Name:        "Server Lookups",
		Component:   "server", // The actual component type name
		Methods:     []string{"Lookup"},
		ResultValue: "Val(Bool: true)", // Look for successful lookups
		Metric:      MetricCount,
		Aggregation: AggRate,
		Window:      10 * time.Second,
	}

	err := rt.metricStore.AddMeasurement(spec)
	if err != nil {
		t.Fatalf("Failed to add measurement: %v", err)
	}

	// Create a method call expression for server.Lookup()
	serverExpr := &decl.IdentifierExpr{Value: "server"}
	serverExpr.SetInferredType(decl.ComponentType(serverComp.ComponentDecl))

	lookupExpr := &MemberAccessExpr{
		Receiver: serverExpr,
		Member:   &decl.IdentifierExpr{Value: "Lookup"},
	}

	callExpr := &CallExpr{
		Function: lookupExpr,
		ArgList:  []Expr{},
	}

	// Simulate 10 calls to server.Lookup()
	simTime := currTime // Continue from where initialization left off
	for i := 0; i < 10; i++ {
		_, _ = eval.Eval(callExpr, env, &simTime)
		simTime += 1e9 // Advance 1 second between calls
	}

	// Get the data
	points, err := rt.metricStore.GetMeasurementData("server_lookups", 100)
	if err != nil {
		t.Fatalf("Failed to get measurement data: %v", err)
	}

	// The server.Lookup() method should have been called 10 times
	if len(points) != 10 {
		t.Errorf("Expected 10 server lookup events, got %d", len(points))
	}

	// Test aggregation
	agg, err := rt.metricStore.GetAggregatedData("server_lookups")
	if err != nil {
		t.Fatalf("Failed to get aggregated data: %v", err)
	}

	// Check that we have a reasonable rate (10 events over ~9 seconds)
	expectedRate := 10.0 / 9.0
	if agg.Value < expectedRate-0.1 || agg.Value > expectedRate+0.1 {
		t.Errorf("Expected rate ~%f, got %f", expectedRate, agg.Value)
	}
}

func TestCircularBuffer(t *testing.T) {
	buffer := NewCircularBuffer(5)

	// Add more points than capacity
	for i := range 10 {
		buffer.Add(MetricPoint{
			Timestamp: core.Duration(i * 1e9), // i seconds in virtual time
			Value:     float64(i),
		})
	}

	// Should only have the last 5 points
	if buffer.Size() != 5 {
		t.Errorf("Expected size 5, got %d", buffer.Size())
	}

	// Get latest 3
	latest := buffer.GetLatest(3)
	if len(latest) != 3 {
		t.Errorf("Expected 3 points, got %d", len(latest))
	}

	// Check they are the most recent (7, 8, 9 seconds)
	for i, p := range latest {
		expected := core.Duration((7 + i) * 1e9) // Should be 7, 8, 9 seconds
		if p.Timestamp != expected {
			t.Errorf("Point %d: expected timestamp %v, got %v", i, expected, p.Timestamp)
		}
	}
}

func TestResultMatcher(t *testing.T) {
	tests := []struct {
		spec    string
		value   string
		matches bool
	}{
		{"*", "anything", true},
		{"*", "", true},
		{"true", "true", true},
		{"true", "false", false},
		{"!=false", "true", true},
		{"!=false", "false", false},
	}

	for _, test := range tests {
		matcher := CreateResultMatcher(test.spec)
		result := matcher.Matches(test.value)
		if result != test.matches {
			t.Errorf("Matcher %s with value %s: expected %v, got %v",
				test.spec, test.value, test.matches, result)
		}
	}
}

func TestAggregations(t *testing.T) {
	// Test count aggregations
	points := []MetricPoint{
		{Timestamp: core.Duration(1e9), Value: 1.0}, // 1 second
		{Timestamp: core.Duration(2e9), Value: 1.0}, // 2 seconds
		{Timestamp: core.Duration(3e9), Value: 1.0}, // 3 seconds
	}

	startTime := core.Duration(1e9)
	endTime := core.Duration(3e9)

	sum, err := aggregateCount(points, AggSum, startTime, endTime)
	if err != nil || sum != 3.0 {
		t.Errorf("Expected sum 3.0, got %f", sum)
	}

	rate, err := aggregateCount(points, AggRate, startTime, endTime)
	if err != nil || rate != 1.5 { // 3 events over 2 seconds
		t.Errorf("Expected rate 1.5, got %f", rate)
	}

	// Test latency aggregations
	latencyPoints := []MetricPoint{
		{Timestamp: core.Duration(1e9), Value: 10e6}, // 10ms in nanos
		{Timestamp: core.Duration(2e9), Value: 20e6}, // 20ms
		{Timestamp: core.Duration(3e9), Value: 30e6}, // 30ms
	}

	avg, err := aggregateLatency(latencyPoints, AggAvg)
	if err != nil || avg != 20.0 { // Average of 10, 20, 30 ms
		t.Errorf("Expected avg 20.0ms, got %f", avg)
	}

	p50, err := aggregateLatency(latencyPoints, AggP50)
	if err != nil || p50 != 20.0 {
		t.Errorf("Expected p50 20.0ms, got %f", p50)
	}
}
