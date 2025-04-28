package sdl

import (
	"testing"
	// Ensure metrics helpers are accessible
)

func TestHashIndex_Init(t *testing.T) {
	hi := NewHashIndex()
	if hi.Disk.ReadOutcomes == nil {
		t.Fatal("HashIndex Disk not initialized")
	}
	if hi.LoadFactorThreshold <= 0 || hi.LoadFactorThreshold > 1.0 {
		t.Error("Default LoadFactorThreshold invalid")
	}
	// Add more checks
}

func TestHashIndex_Probabilities(t *testing.T) {
	hi := NewHashIndex()

	// Use different base params for clarity
	hi.PageSize = 8192
	hi.RecordSize = 200 // ~30 records/page at LF=0.75
	hi.LoadFactorThreshold = 0.75
	hi.AvgOverflowReads = 0.25 // Make collision effect slightly higher

	// Low load
	hi.NumRecords = 100000
	lf1 := hi.currentLoadFactor()
	pc1 := hi.collisionProbability()
	pr1 := hi.resizeProbability()
	t.Logf("Load Factor (Low: %.0f recs): %.3f, Collision Prob: %.4f, Resize Prob: %.4f", float64(hi.NumRecords), lf1, pc1, pr1)
	if pr1 > 0 {
		t.Errorf("Resize prob should be 0 for low load factor %.3f", lf1)
	}

	// Medium load (near threshold)
	hi.NumRecords = uint(float64(hi.NumRecords) * 2.5) // Target LF around 0.75
	lf2 := hi.currentLoadFactor()
	pc2 := hi.collisionProbability()
	pr2 := hi.resizeProbability()
	t.Logf("Load Factor (Med: %.0f recs): %.3f, Collision Prob: %.4f, Resize Prob: %.4f", float64(hi.NumRecords), lf2, pc2, pr2)
	if !(lf2 > lf1) {
		t.Error("Load factor should increase")
	}
	if !(pc2 > pc1) {
		t.Error("Collision probability should increase")
	}
	// Resize prob might be exactly 0 or slightly positive if LF is just past threshold
	if pr2 < 0 || pr2 > 0.05 {
		t.Errorf("Resize prob %.4f unexpected near threshold LF %.3f", pr2, lf2)
	}

	// High load (well past threshold)
	hi.NumRecords = uint(float64(hi.NumRecords) * 2.0) // Target LF around 1.5 (but calc is relative to initial capacity)
	lf3 := hi.currentLoadFactor()
	pc3 := hi.collisionProbability()
	pr3 := hi.resizeProbability()
	t.Logf("Load Factor (High: %.0f recs): %.3f, Collision Prob: %.4f, Resize Prob: %.4f", float64(hi.NumRecords), lf3, pc3, pr3)
	if !(lf3 > lf2) {
		t.Error("Load factor should increase further")
	}
	if !(pc3 > pc2) {
		t.Error("Collision probability should increase further")
	}
	if !(pr3 > pr2) {
		t.Error("Resize probability should increase past threshold")
	}
	if pr3 <= 0 {
		t.Error("Resize probability should be > 0 well past threshold")
	}
}

func TestHashIndex_Find_Vs_Modify_Metrics(t *testing.T) {
	hi := NewHashIndex()
	hi.Disk.Init()
	hi.MaxOutcomeLen = 15
	hi.NumRecords = 800000 // Reasonably high load

	// --- Test Find ---
	findOutcomes := hi.Find()
	if findOutcomes == nil || findOutcomes.Len() == 0 {
		t.Fatal("Find() returned nil or empty outcomes")
	}
	findAvail := Availability(findOutcomes)
	findMean := MeanLatency(findOutcomes)
	findP99 := PercentileLatency(findOutcomes, 0.99)
	t.Logf("Hash Find: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", findAvail, findMean, findP99, findOutcomes.Len())

	// --- Test Insert ---
	insertOutcomes := hi.Insert()
	if insertOutcomes == nil || insertOutcomes.Len() == 0 {
		t.Fatal("Insert() returned nil or empty outcomes")
	}
	insertAvail := Availability(insertOutcomes)
	insertMean := MeanLatency(insertOutcomes)
	insertP99 := PercentileLatency(insertOutcomes, 0.99)
	t.Logf("Hash Insert: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", insertAvail, insertMean, insertP99, insertOutcomes.Len())

	// --- Test Delete ---
	deleteOutcomes := hi.Delete()
	if deleteOutcomes == nil || deleteOutcomes.Len() == 0 {
		t.Fatal("Delete() returned nil or empty outcomes")
	}
	deleteAvail := Availability(deleteOutcomes)
	deleteMean := MeanLatency(deleteOutcomes)
	deleteP99 := PercentileLatency(deleteOutcomes, 0.99)
	t.Logf("Hash Delete: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", deleteAvail, deleteMean, deleteP99, deleteOutcomes.Len())

	// --- Compare Metrics ---
	// Find should generally be fastest on average (no writes, no resize checks)
	if findMean >= deleteMean {
		t.Errorf("Hash Find Mean Latency (%.6fs) should typically be less than Delete Mean (%.6fs)", findMean, deleteMean)
	}
	// Insert can be very slow if resize occurs
	// Delete should be somewhere between Find and Insert (no resize)
	if deleteMean >= insertMean && hi.resizeProbability() > 0.01 { // Only expect Insert to be slower if resize is likely
		t.Logf("Info: Insert mean (%.6fs) slower than delete mean (%.6fs) as expected with potential resize", insertMean, deleteMean)
	} else if deleteMean >= insertMean {
		t.Logf("Warning: Delete mean (%.6fs) unexpectedly higher than Insert mean (%.6fs) when resize prob is low", deleteMean, insertMean)
	}

	// Availability should be similar if disk profiles are reliable
	if !approxEqualTest(findAvail, deleteAvail, 0.005) {
		t.Errorf("Find and Delete availability should be similar")
	}
}
