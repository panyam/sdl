package components

import (
	"testing"

	sc "github.com/panyam/leetcoach/sdl/core"
	// Ensure metrics helpers are accessible
)

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

	// --- Test Find ---
	findOutcomes := bt.Find()
	if findOutcomes == nil || findOutcomes.Len() == 0 {
		t.Fatal("Find() returned nil or empty outcomes")
	}
	findAvail := sc.Availability(findOutcomes)
	findMean := sc.MeanLatency(findOutcomes)
	findP99 := sc.PercentileLatency(findOutcomes, 0.99)
	t.Logf("BTree Find : Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", findAvail, findMean, findP99, findOutcomes.Len())

	// Plausibility: Find involves ~Height disk reads + CPU. Should be slower than single read.
	singleReadMean := sc.MeanLatency(bt.Disk.Read())
	if findMean < singleReadMean*float64(height)*0.8 { // Allow some leeway
		t.Errorf("BTree Find mean latency (%.6fs) seems too low compared to ~%d disk reads (%.6fs)", findMean, height, singleReadMean)
	}
	if findAvail < 0.99 { // Should be reliable if disk is reliable
		t.Errorf("BTree Find availability (%.4f) seems too low", findAvail)
	}

	// --- Test Insert ---
	insertOutcomes := bt.Insert()
	if insertOutcomes == nil || insertOutcomes.Len() == 0 {
		t.Fatal("Insert() returned nil or empty outcomes")
	}
	insertAvail := sc.Availability(insertOutcomes)
	insertMean := sc.MeanLatency(insertOutcomes)
	insertP99 := sc.PercentileLatency(insertOutcomes, 0.99)
	t.Logf("BTree Insert: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", insertAvail, insertMean, insertP99, insertOutcomes.Len())

	// Plausibility: Insert = Find + Modify + Write + Propagation. Should be slower than Find.
	if insertMean <= findMean {
		t.Errorf("BTree Insert mean latency (%.6fs) should be greater than Find mean (%.6fs)", insertMean, findMean)
	}

	// --- Test Delete ---
	deleteOutcomes := bt.Delete()
	if deleteOutcomes == nil || deleteOutcomes.Len() == 0 {
		t.Fatal("Delete() returned nil or empty outcomes")
	}
	deleteAvail := sc.Availability(deleteOutcomes)
	deleteMean := sc.MeanLatency(deleteOutcomes)
	deleteP99 := sc.PercentileLatency(deleteOutcomes, 0.99)
	t.Logf("BTree Delete: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", deleteAvail, deleteMean, deleteP99, deleteOutcomes.Len())

	// Plausibility: Delete cost should be similar to Insert cost in this model.
	if !approxEqualTest(insertMean, deleteMean, insertMean*0.2) { // Allow 20% difference
		t.Logf("Warning: BTree Insert (%.6fs) and Delete (%.6fs) mean latencies differ significantly", insertMean, deleteMean)
	}

	// Instead, check if Insert/Delete availability makes sense relative to Find * Write * Prop
	expectedInsertDeleteAvail := sc.Availability(findOutcomes) * sc.Availability(bt.Disk.Write()) * sc.Availability(sc.And(bt.Disk.Read(), bt.Disk.Write(), sc.AndAccessResults))
	t.Logf("Expected Insert/Delete Avail Approx: %.4f", expectedInsertDeleteAvail)
	if !approxEqualTest(insertAvail, expectedInsertDeleteAvail, 0.002) { // Allow slightly larger tolerance due to multiple steps/reductions
		t.Errorf("Insert availability (%.4f) differs significantly from expected (%.4f)", insertAvail, expectedInsertDeleteAvail)
	}
	if !approxEqualTest(deleteAvail, expectedInsertDeleteAvail, 0.002) {
		t.Errorf("Delete availability (%.4f) differs significantly from expected (%.4f)", deleteAvail, expectedInsertDeleteAvail)
	}
}
