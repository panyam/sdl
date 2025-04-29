// components/btree_test.go
package components

import (
	// Added
	"testing"

	sc "github.com/panyam/leetcoach/sdl/core"
)

// Tests for Init, Height remain the same...
func TestBTreeIndex_Init(t *testing.T) {
	bt := (&BTreeIndex{}).Init() // Use constructor style matching BTreeIndex.Init()

	if bt.Disk.ReadOutcomes == nil {
		t.Fatal("BTreeIndex Disk not initialized")
	}
	if bt.NodeFanout <= 1 {
		t.Error("Default NodeFanout should be > 1")
	}
	if bt.Occupancy <= 0 || bt.Occupancy > 1.0 {
		t.Error("Default Occupancy invalid")
	}
	if bt.AvgSplitPropCost < 0 || bt.AvgMergePropCost < 0 {
		t.Error("Propagation costs cannot be negative")
	}
	// Add more checks for defaults
}

func TestBTreeIndex_Height(t *testing.T) {
	bt := (&BTreeIndex{}).Init()
	bt.PageSize = 4096
	bt.RecordSize = 100 // ~40 records/page
	bt.NodeFanout = 50  // Lower fanout for testing height changes

	// Test case 1: Small number of records (should fit in root/leaf)
	bt.NumRecords = 100
	h1 := bt.Height()
	t.Logf("Height (100 records, Fanout 50): %d", h1)
	if h1 < 1 || h1 > 2 { // Expecting 1 (root only) or 2 (root + leaf)
		t.Errorf("Unexpected height %d for small record count", h1)
	}

	// Test case 2: More records, requiring multiple levels
	// NumPages = ceil(1M * 100 / 4096) = ceil(24414) = 24415
	// Height = ceil(log50(24415)) + 1 = ceil(4.38 / 1.70) + 1 = ceil(2.58) + 1 = 3 + 1 = 4
	bt.NumRecords = 1000000
	h2 := bt.Height()
	t.Logf("Height (1M records, Fanout 50): %d", h2)
	if h2 < 3 || h2 > 5 { // Expecting height around 4, allow some flexibility
		t.Errorf("Unexpected height %d for 1M records, Fanout 50", h2)
	}

	// Test case 3: Higher fanout should reduce height
	bt.NodeFanout = 200
	// Height = ceil(log200(24415)) + 1 = ceil(4.38 / 2.30) + 1 = ceil(1.90) + 1 = 2 + 1 = 3
	h3 := bt.Height()
	t.Logf("Height (1M records, Fanout 200): %d", h3)
	if h3 >= h2 {
		t.Errorf("Height should decrease with higher fanout (got %d, previous %d)", h3, h2)
	}
	if h3 < 2 || h3 > 4 { // Expecting height around 3
		t.Errorf("Unexpected height %d for 1M records, Fanout 200", h3)
	}
}

func TestBTreeIndex_Operations_Metrics(t *testing.T) {
	bt := (&BTreeIndex{}).Init()
	// Use SSD for faster base operations
	bt.Disk.Init()
	bt.MaxOutcomeLen = 15
	bt.NumRecords = 5000000 // Larger number of records
	bt.NodeFanout = 150     // Realistic fanout

	height := bt.Height()
	t.Logf("BTree Test Setup: NumRecords=%.0f, Fanout=%d, Height=%d", float64(bt.NumRecords), bt.NodeFanout, height)

	// Base SSD read mean for comparison
	singleReadMean := sc.MeanLatency(bt.Disk.Read())

	// --- Test Find ---
	findOutcomes := bt.Find()
	if findOutcomes == nil || findOutcomes.Len() == 0 {
		t.Fatal("Find() returned nil or empty outcomes")
	}
	// Manual
	findAvail := sc.Availability(findOutcomes)
	findMean := sc.MeanLatency(findOutcomes)
	findP99 := sc.PercentileLatency(findOutcomes, 0.99)
	t.Logf("Manual Log - BTree Find : Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", findAvail, findMean, findP99, findOutcomes.Len())
	// Analyze
	findExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, 0.99),
		sc.ExpectMeanLatency(sc.GTE, singleReadMean*float64(height)*0.5), // Expect mean > half base reads
		sc.ExpectMeanLatency(sc.LT, singleReadMean*float64(height)*2.0),  // Expect mean < double base reads
	}
	findAnalysis := sc.Analyze("BTree Find", func() *sc.Outcomes[sc.AccessResult] { return findOutcomes }, findExpectations...)
	findAnalysis.LogResults(t)

	// Plausibility check (manual)
	if findMean < singleReadMean*float64(height)*0.8 {
		t.Errorf("Manual Check - BTree Find mean latency (%.6fs) seems too low compared to ~%d disk reads (%.6fs)", findMean, height, singleReadMean)
	}
	if findAvail < 0.99 {
		t.Errorf("Manual Check - BTree Find availability (%.4f) seems too low", findAvail)
	}

	// --- Test Insert ---
	insertOutcomes := bt.Insert()
	if insertOutcomes == nil || insertOutcomes.Len() == 0 {
		t.Fatal("Insert() returned nil or empty outcomes")
	}
	// Manual
	insertAvail := sc.Availability(insertOutcomes)
	insertMean := sc.MeanLatency(insertOutcomes)
	insertP99 := sc.PercentileLatency(insertOutcomes, 0.99)
	t.Logf("Manual Log - BTree Insert: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", insertAvail, insertMean, insertP99, insertOutcomes.Len())
	// Analyze
	insertExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, 0.98),    // Slightly lower avail due to more ops
		sc.ExpectMeanLatency(sc.GTE, findMean), // Insert >= Find
	}
	insertAnalysis := sc.Analyze("BTree Insert", func() *sc.Outcomes[sc.AccessResult] { return insertOutcomes }, insertExpectations...)
	insertAnalysis.LogResults(t)

	// Plausibility check (manual)
	if insertMean <= findMean {
		t.Errorf("Manual Check - BTree Insert mean latency (%.6fs) should be greater than Find mean (%.6fs)", insertMean, findMean)
	}

	// --- Test Delete ---
	deleteOutcomes := bt.Delete()
	if deleteOutcomes == nil || deleteOutcomes.Len() == 0 {
		t.Fatal("Delete() returned nil or empty outcomes")
	}
	// Manual
	deleteAvail := sc.Availability(deleteOutcomes)
	deleteMean := sc.MeanLatency(deleteOutcomes)
	deleteP99 := sc.PercentileLatency(deleteOutcomes, 0.99)
	t.Logf("Manual Log - BTree Delete: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", deleteAvail, deleteMean, deleteP99, deleteOutcomes.Len())
	// Analyze
	deleteExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, 0.98),
		sc.ExpectMeanLatency(sc.GTE, findMean), // Delete >= Find
	}
	deleteAnalysis := sc.Analyze("BTree Delete", func() *sc.Outcomes[sc.AccessResult] { return deleteOutcomes }, deleteExpectations...)
	deleteAnalysis.LogResults(t)

	// Plausibility checks (manual)
	if !approxEqualTest(insertMean, deleteMean, insertMean*0.2) {
		t.Logf("Manual Check Warning: BTree Insert (%.6fs) and Delete (%.6fs) mean latencies differ significantly", insertMean, deleteMean)
	}
}
