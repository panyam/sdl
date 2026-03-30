package runtime

import (
	"testing"
)

// TestFlowEvalWithContactsSDL tests flow evaluation with the actual contacts.sdl file
func TestFlowEvalWithContactsSDL(t *testing.T) {
	sys, _ := loadSystem(t, "../../examples/contacts/contacts.sdl", "ContactsSystem")
	env := sys.Env

	// Components are inside the arch parameter
	server := sys.FindComponent("arch.server")
	if server == nil {
		t.Fatal("Server component not found at arch.server")
	}

	cache := sys.FindComponent("arch.contactCache")
	if cache == nil {
		t.Fatal("Cache component not found at arch.contactCache")
	}

	db := sys.FindComponent("arch.database")
	if db == nil {
		t.Fatal("Database component not found at arch.database")
	}

	idx := sys.FindComponent("arch.idx")
	if idx == nil {
		t.Fatal("Index component not found at arch.idx")
	}

	// Create flow scope with the system environment
	scope := NewFlowScope(env)

	// Create generator entry point for server.Lookup at 10 RPS
	generators := []GeneratorEntryPointRuntime{
		{
			Component:   server,
			Method:      "Lookup",
			Rate:        10.0,
			GeneratorID: "test-gen",
		},
	}

	// Solve the system flows
	result := SolveSystemFlowsRuntime(generators, scope)

	// Validate the flow rates
	t.Run("Server rates", func(t *testing.T) {
		rate := result.GetRate(server, "Lookup")
		if rate < 9.9 || rate > 10.1 {
			t.Errorf("Expected server.Lookup rate ~10.0, got %f", rate)
		}
	})

	t.Run("Cache rates", func(t *testing.T) {
		rate := result.GetRate(cache, "Read")
		if rate < 9.9 || rate > 10.1 {
			t.Errorf("Expected cache.Read rate ~10.0 (100%% of lookups), got %f", rate)
		}

		writeRate := result.GetRate(cache, "Write")
		t.Logf("Cache.Write rate: %f", writeRate)
		if writeRate < 0 || writeRate > 10 {
			t.Errorf("Cache.Write rate out of expected range: %f", writeRate)
		}
	})

	t.Run("Database rates", func(t *testing.T) {
		rate := result.GetRate(db, "LookupByPhone")
		t.Logf("Database.LookupByPhone rate: %f", rate)
		if rate < 0 || rate > 10 {
			t.Errorf("Database.LookupByPhone rate out of expected range: %f", rate)
		}
	})

	t.Run("Index rates", func(t *testing.T) {
		rate := result.GetRate(idx, "Find")
		t.Logf("Index.Find rate: %f", rate)
		dbRate := result.GetRate(db, "LookupByPhone")
		if rate < dbRate*0.9 || rate > dbRate*1.1 {
			t.Errorf("Expected idx.Find rate ~%f (same as DB lookups), got %f", dbRate, rate)
		}
	})

	// Test Insert flow as well
	t.Run("Insert flow", func(t *testing.T) {
		insertGenerators := []GeneratorEntryPointRuntime{
			{
				Component:   server,
				Method:      "Insert",
				Rate:        1.0,
				GeneratorID: "test-gen-insert",
			},
		}

		insertScope := NewFlowScope(env)
		insertResult := SolveSystemFlowsRuntime(insertGenerators, insertScope)

		serverInsertRate := insertResult.GetRate(server, "Insert")
		if serverInsertRate < 0.9 || serverInsertRate > 1.1 {
			t.Errorf("Expected server.Insert rate ~1.0, got %f", serverInsertRate)
		}

		dbInsertRate := insertResult.GetRate(db, "Insert")
		if dbInsertRate < 0.9 || dbInsertRate > 1.1 {
			t.Errorf("Expected database.Insert rate ~1.0, got %f", dbInsertRate)
		}

		idxInsertRate := insertResult.GetRate(idx, "Insert")
		t.Logf("Index.Insert rate: %f", idxInsertRate)
		if idxInsertRate < 0 || idxInsertRate > 2 {
			t.Errorf("Index.Insert rate out of expected range: %f", idxInsertRate)
		}
	})

	// Log all rates for debugging
	t.Log("=== All flow rates ===")
	for comp, methods := range result {
		for method, rate := range methods {
			if rate > 0.01 {
				t.Logf("%s.%s: %.2f RPS", comp.ID(), method, rate)
			}
		}
	}
}
