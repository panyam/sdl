package console

import (
	"testing"
)

func TestFlowEvalIntegration(t *testing.T) {
	// Test FlowEval integration with Canvas traffic generators
	canvas := NewCanvas("test", nil)

	// Load the contacts system
	err := canvas.Load("../examples/contacts/contacts.sdl")
	if err != nil {
		t.Fatalf("Failed to load contacts.sdl: %v", err)
	}

	// Use the ContactsSystem
	err = canvas.Use("ContactsSystem")
	if err != nil {
		t.Fatalf("Failed to use ContactsSystem: %v", err)
	}

	// Add a traffic generator
	genConfig := &GeneratorConfig{
		ID:      "test-gen",
		Name:    "Test Generator",
		Target:  "server.Lookup",
		Rate:    15, // 15 RPS
		Enabled: true,
	}

	err = canvas.AddGenerator(genConfig)
	if err != nil {
		t.Fatalf("Failed to add generator: %v", err)
	}

	// Update the generator (this should trigger FlowEval)
	genConfig.Rate = 20
	err = canvas.UpdateGenerator(genConfig)
	if err != nil {
		t.Fatalf("Failed to update generator: %v", err)
	}

	// Test passed if we get here without crashing
	t.Logf("FlowEval integration test completed successfully")
	
	// Print generator info
	generators := canvas.GetGenerators()
	for id, gen := range generators {
		t.Logf("Generator %s: %s @ %d RPS", id, gen.Target, gen.Rate)
	}
}