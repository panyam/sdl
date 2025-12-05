package runtime

import (
	"testing"

	"github.com/panyam/sdl/lib/loader"
)

func TestFlowEvalBasic(t *testing.T) {
	// Create a simple test system to validate FlowEval
	// We'll manually create the AST structures for testing

	// Create test context
	system := &SystemDecl{
		Name: &IdentifierExpr{Value: "TestSystem"},
	}

	context := NewFlowContext(system, map[string]interface{}{})

	// Since we don't have full loader integration yet, we'll test the core parsing logic
	// by directly testing the helper functions

	t.Run("parseCallTarget", func(t *testing.T) {
		testCases := []struct {
			input             string
			expectedComponent string
			expectedMethod    string
		}{
			{"db.LookupByPhone", "db", "LookupByPhone"},
			{"server.cache.Read", "server.cache", "Read"},
			{"pool.Acquire", "pool", "Acquire"},
		}

		for _, tc := range testCases {
			component, method := context.parseCallTarget(tc.input)
			if component != tc.expectedComponent {
				t.Errorf("parseCallTarget(%s): expected component %s, got %s",
					tc.input, tc.expectedComponent, component)
			}
			if method != tc.expectedMethod {
				t.Errorf("parseCallTarget(%s): expected method %s, got %s",
					tc.input, tc.expectedMethod, method)
			}
		}
	})

	t.Run("memberExpressionToString", func(t *testing.T) {
		// Test the member expression parsing
		memberExpr := &MemberAccessExpr{
			Receiver: &MemberAccessExpr{
				Receiver: &IdentifierExpr{Value: "self"},
				Member:   &IdentifierExpr{Value: "db"},
			},
			Member: &IdentifierExpr{Value: "LookupByPhone"},
		}

		result := context.memberExpressionToString(memberExpr)
		expected := "db.LookupByPhone"
		if result != expected {
			t.Errorf("memberExpressionToString: expected %s, got %s", expected, result)
		}
	})

	t.Run("evaluateConditionProbability", func(t *testing.T) {
		// Test with CacheHitRate parameter
		context.Parameters["CacheHitRate"] = 0.8

		prob := context.evaluateConditionProbability(nil) // Condition doesn't matter for this test
		expected := 0.8
		if prob != expected {
			t.Errorf("evaluateConditionProbability: expected %f, got %f", expected, prob)
		}

		// Test without parameter (should default to 0.5)
		context.Parameters = map[string]interface{}{}
		prob = context.evaluateConditionProbability(nil)
		expected = 0.5
		if prob != expected {
			t.Errorf("evaluateConditionProbability: expected %f, got %f", expected, prob)
		}
	})
}

func TestFlowEvalIntegration(t *testing.T) {
	// Integration test that loads the actual ContactsSystem and tests FlowEval
	// This test will help us identify what we need to fix in the implementation

	// Load the contacts system
	l := loader.NewLoader(nil, nil, 10)
	fileStatus, err := l.LoadFile("../examples/contacts/contacts.sdl", "", 0)
	if err != nil {
		t.Fatalf("Failed to load contacts.sdl: %v", err)
	}

	// Get the system declarations
	systemDecls, err := fileStatus.FileDecl.GetSystems()
	if err != nil {
		t.Fatalf("Failed to get system declarations: %v", err)
	}

	if len(systemDecls) == 0 {
		t.Fatalf("No system declarations found in contacts.sdl")
	}

	// Get the ContactsSystem specifically
	systemDecl := systemDecls["ContactsSystem"]
	if systemDecl == nil {
		t.Fatalf("ContactsSystem not found")
	}
	if systemDecl.Name.Value != "ContactsSystem" {
		t.Fatalf("Expected ContactsSystem, got %s", systemDecl.Name.Value)
	}

	// Create flow context
	parameters := map[string]interface{}{
		"CacheHitRate": 0.4, // 40% cache hit rate from ContactDatabase
	}
	context := NewFlowContext(systemDecl, parameters)

	// For now, we'll test the basic structure
	// TODO: Once we have proper component declaration access, we can test actual flow evaluation

	t.Logf("Successfully loaded ContactsSystem with %d body items", len(systemDecl.Body))
	for _, bodyItem := range systemDecl.Body {
		if instance, ok := bodyItem.(*InstanceDecl); ok {
			t.Logf("Instance: %s of type %s", instance.Name.Value, instance.ComponentName.Value)
		}
	}

	// Test basic FlowEval call (this will fail gracefully due to missing component declarations)
	flows := FlowEval("server", "HandleLookup", 10.0, context)
	t.Logf("FlowEval result (expected to be empty for now): %v", flows)

	// The test passes if we can load the system and call FlowEval without crashing
}
