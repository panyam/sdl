// components/bitmap_test.go
package components

import (
	// Added
	"testing"

	sc "github.com/panyam/leetcoach/sdl/core"
)

// Test Init remains the same...
func TestBitmapIndex_Init(t *testing.T) {
	bmi := NewBitmapIndex()
	if bmi.Disk.ReadOutcomes == nil {
		t.Fatal("BitmapIndex Disk not initialized")
	}
	if bmi.Cardinality == 0 {
		t.Error("Default Cardinality should be > 0")
	}
	// Add more checks for defaults
}

func TestBitmapIndex_Find_Vs_Modify_Metrics(t *testing.T) {
	bmi := NewBitmapIndex()
	// Use SSD for faster base operations
	bmi.Disk.Init()
	bmi.MaxOutcomeLen = 15

	// Set some specific parameters for testing
	bmi.NumRecords = 1000000     // 1 million rows
	bmi.QuerySelectivity = 0.001 // 0.1% selectivity -> 1000 result rows
	bmi.Cardinality = 50         // Example: 50 distinct values

	// --- Test Find ---
	findOutcomes := bmi.Find()
	if findOutcomes == nil || findOutcomes.Len() == 0 {
		t.Fatal("Find() returned nil or empty outcomes")
	}
	// Manual
	findAvail := sc.Availability(findOutcomes)
	findMean := sc.MeanLatency(findOutcomes)
	findP99 := sc.PercentileLatency(findOutcomes, 0.99)
	t.Logf("Manual Log - Bitmap Find: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", findAvail, findMean, findP99, findOutcomes.Len())
	// Analyze
	findExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, 0.99),
		sc.ExpectMeanLatency(sc.LT, sc.Millis(10)), // Find should be fast
	}
	findAnalysis := sc.Analyze("Bitmap Find", func() *sc.Outcomes[sc.AccessResult] { return findOutcomes }, findExpectations...)
	findAnalysis.Assert(t)

	// Plausibility check (manual)
	if findAvail < 0.99 {
		t.Errorf("Manual Check - Bitmap Find availability (%.4f) seems too low", findAvail)
	}

	// --- Test Insert (representative of modify) ---
	insertOutcomes := bmi.Insert()
	if insertOutcomes == nil || insertOutcomes.Len() == 0 {
		t.Fatal("Insert() returned nil or empty outcomes")
	}
	// Manual
	insertAvail := sc.Availability(insertOutcomes)
	insertMean := sc.MeanLatency(insertOutcomes)
	insertP99 := sc.PercentileLatency(insertOutcomes, 0.99)
	t.Logf("Manual Log - Bitmap Insert: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", insertAvail, insertMean, insertP99, insertOutcomes.Len())
	// Analyze
	insertExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, 0.98),    // Mod slightly less avail
		sc.ExpectMeanLatency(sc.GTE, findMean), // Insert > Find
	}
	insertAnalysis := sc.Analyze("Bitmap Insert", func() *sc.Outcomes[sc.AccessResult] { return insertOutcomes }, insertExpectations...)
	insertAnalysis.Assert(t)

	// Plausibility checks (manual)
	if insertAvail < 0.99 {
		t.Errorf("Manual Check - Bitmap Insert availability (%.4f) seems too low", insertAvail)
	}
	if findMean >= insertMean {
		t.Errorf("Manual Check - Bitmap Find Mean Latency (%.6fs) should be less than Insert Mean (%.6fs)", findMean, insertMean)
	}
	if findP99 >= insertP99 {
		t.Errorf("Manual Check - Bitmap Find P99 Latency (%.6fs) should likely be less than Insert P99 (%.6fs)", findP99, insertP99)
	}
}
