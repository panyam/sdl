package sdl

import (
	"testing"
	// Ensure metrics helpers are accessible
)

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

	// --- Test Find ---
	findOutcomes := bmi.Find()
	if findOutcomes == nil || findOutcomes.Len() == 0 {
		t.Fatal("Find() returned nil or empty outcomes")
	}
	findAvail := Availability(findOutcomes)
	findMean := MeanLatency(findOutcomes)
	findP99 := PercentileLatency(findOutcomes, 0.99)
	t.Logf("Bitmap Find: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", findAvail, findMean, findP99, findOutcomes.Len())

	if findAvail < 0.99 { // Find should generally be reliable
		t.Errorf("Bitmap Find availability (%.4f) seems too low", findAvail)
	}

	// --- Test Insert (representative of modify) ---
	insertOutcomes := bmi.Insert()
	if insertOutcomes == nil || insertOutcomes.Len() == 0 {
		t.Fatal("Insert() returned nil or empty outcomes")
	}
	insertAvail := Availability(insertOutcomes)
	insertMean := MeanLatency(insertOutcomes)
	insertP99 := PercentileLatency(insertOutcomes, 0.99)
	t.Logf("Bitmap Insert: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", insertAvail, insertMean, insertP99, insertOutcomes.Len())

	if insertAvail < 0.99 { // Modify operations depend on disk reliability
		t.Errorf("Bitmap Insert availability (%.4f) seems too low", insertAvail)
	}

	// --- Compare Find vs Insert ---
	// Expect Find to be significantly faster than Insert/Update/Delete on average
	if findMean >= insertMean {
		t.Errorf("Bitmap Find Mean Latency (%.6fs) should be less than Insert Mean (%.6fs)", findMean, insertMean)
	}
	// P99 might be complex, but likely find is faster there too
	if findP99 >= insertP99 {
		t.Errorf("Bitmap Find P99 Latency (%.6fs) should likely be less than Insert P99 (%.6fs)", findP99, insertP99)
	}
}
