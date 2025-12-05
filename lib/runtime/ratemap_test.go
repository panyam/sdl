package runtime

import (
	"testing"
)

func TestRateMap(t *testing.T) {
	// Create mock component instances for testing
	comp1 := &ComponentInstance{id: "comp1"}
	comp2 := &ComponentInstance{id: "comp2"}
	
	t.Run("Basic Operations", func(t *testing.T) {
		rm := NewRateMap()
		
		// Test AddFlow
		rm.AddFlow(comp1, "method1", 10.0)
		rm.AddFlow(comp1, "method2", 20.0)
		rm.AddFlow(comp2, "method1", 30.0)
		
		// Test GetRate
		if rate := rm.GetRate(comp1, "method1"); rate != 10.0 {
			t.Errorf("Expected rate 10.0, got %f", rate)
		}
		if rate := rm.GetRate(comp1, "method2"); rate != 20.0 {
			t.Errorf("Expected rate 20.0, got %f", rate)
		}
		if rate := rm.GetRate(comp2, "method1"); rate != 30.0 {
			t.Errorf("Expected rate 30.0, got %f", rate)
		}
		
		// Test AddFlow accumulation
		rm.AddFlow(comp1, "method1", 5.0)
		if rate := rm.GetRate(comp1, "method1"); rate != 15.0 {
			t.Errorf("Expected accumulated rate 15.0, got %f", rate)
		}
		
		// Test SetRate (replaces value)
		rm.SetRate(comp1, "method1", 100.0)
		if rate := rm.GetRate(comp1, "method1"); rate != 100.0 {
			t.Errorf("Expected replaced rate 100.0, got %f", rate)
		}
	})
	
	t.Run("Nil and Empty Handling", func(t *testing.T) {
		rm := NewRateMap()
		
		// Test nil component
		rm.AddFlow(nil, "method", 10.0)
		if rate := rm.GetRate(nil, "method"); rate != 0.0 {
			t.Errorf("Expected 0.0 for nil component, got %f", rate)
		}
		
		// Test empty method
		rm.AddFlow(comp1, "", 10.0)
		if rate := rm.GetRate(comp1, ""); rate != 0.0 {
			t.Errorf("Expected 0.0 for empty method, got %f", rate)
		}
		
		// Test non-existent entries
		if rate := rm.GetRate(comp1, "nonexistent"); rate != 0.0 {
			t.Errorf("Expected 0.0 for non-existent entry, got %f", rate)
		}
	})
	
	t.Run("Aggregate Functions", func(t *testing.T) {
		rm := NewRateMap()
		rm.SetRate(comp1, "method1", 10.0)
		rm.SetRate(comp1, "method2", 20.0)
		rm.SetRate(comp2, "method1", 30.0)
		
		// Test GetTotalRate
		if total := rm.GetTotalRate(); total != 60.0 {
			t.Errorf("Expected total rate 60.0, got %f", total)
		}
		
		// Test GetComponentRate
		if compRate := rm.GetComponentRate(comp1); compRate != 30.0 {
			t.Errorf("Expected component rate 30.0, got %f", compRate)
		}
		if compRate := rm.GetComponentRate(comp2); compRate != 30.0 {
			t.Errorf("Expected component rate 30.0, got %f", compRate)
		}
	})
	
	t.Run("Copy and Clear", func(t *testing.T) {
		rm := NewRateMap()
		rm.SetRate(comp1, "method1", 10.0)
		rm.SetRate(comp2, "method1", 20.0)
		
		// Test Copy
		rmCopy := rm.Copy()
		if rate := rmCopy.GetRate(comp1, "method1"); rate != 10.0 {
			t.Errorf("Copy failed: expected rate 10.0, got %f", rate)
		}
		
		// Modify original, ensure copy is unchanged
		rm.SetRate(comp1, "method1", 50.0)
		if rate := rmCopy.GetRate(comp1, "method1"); rate != 10.0 {
			t.Errorf("Copy is not independent: expected rate 10.0, got %f", rate)
		}
		
		// Test Clear
		rm.Clear()
		if total := rm.GetTotalRate(); total != 0.0 {
			t.Errorf("Clear failed: expected total 0.0, got %f", total)
		}
		
		// Ensure copy is still intact
		if rate := rmCopy.GetRate(comp1, "method1"); rate != 10.0 {
			t.Errorf("Clear affected copy: expected rate 10.0, got %f", rate)
		}
	})
}